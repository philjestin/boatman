package triage

import (
	"testing"
)

func TestOverlapScore(t *testing.T) {
	tests := []struct {
		name string
		a, b NormalizedTicket
		min  float64
	}{
		{
			name: "shared files",
			a: NormalizedTicket{
				Signals: Signals{MentionsFiles: []string{"packs/auth/app/models/user.rb"}},
			},
			b: NormalizedTicket{
				Signals: Signals{MentionsFiles: []string{"packs/auth/app/models/user.rb"}},
			},
			min: 3.0,
		},
		{
			name: "file prefix match",
			a: NormalizedTicket{
				Signals: Signals{MentionsFiles: []string{"packs/auth/app/models/user.rb"}},
			},
			b: NormalizedTicket{
				Signals: Signals{MentionsFiles: []string{"packs/auth/app/models/session.rb"}},
			},
			// "packs/auth/app/models/session.rb" has prefix "packs/auth/app/models/user.rb"? No.
			// But "packs/auth/app/models/user.rb" does NOT have prefix of the other.
			// These don't share a prefix match in either direction.
			min: 0.0,
		},
		{
			name: "shared domains",
			a: NormalizedTicket{
				Signals: Signals{Domains: []string{"frontend", "graphql"}},
			},
			b: NormalizedTicket{
				Signals: Signals{Domains: []string{"frontend", "backend"}},
			},
			min: 1.0, // "frontend" shared = 1.0
		},
		{
			name: "shared labels",
			a: NormalizedTicket{
				Signals: Signals{Labels: []string{"bug", "P1"}},
			},
			b: NormalizedTicket{
				Signals: Signals{Labels: []string{"bug", "enhancement"}},
			},
			min: 0.5, // "bug" shared = 0.5
		},
		{
			name: "shared dependencies",
			a: NormalizedTicket{
				Signals: Signals{Dependencies: []string{"ENG-100", "ENG-200"}},
			},
			b: NormalizedTicket{
				Signals: Signals{Dependencies: []string{"ENG-100"}},
			},
			min: 1.5, // "ENG-100" shared = 1.5
		},
		{
			name: "no overlap",
			a: NormalizedTicket{
				Signals: Signals{
					Domains: []string{"frontend"},
					Labels:  []string{"P1"},
				},
			},
			b: NormalizedTicket{
				Signals: Signals{
					Domains: []string{"backend"},
					Labels:  []string{"P3"},
				},
			},
			min: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := overlapScore(&tt.a, &tt.b)
			if score < tt.min {
				t.Errorf("expected score >= %.1f, got %.1f", tt.min, score)
			}
		})
	}
}

func TestClusterTickets_SingleTicket(t *testing.T) {
	tickets := []NormalizedTicket{
		{TicketID: "ENG-1", Signals: Signals{Domains: []string{"frontend"}}},
	}
	classifications := []Classification{
		{TicketID: "ENG-1", Category: CategoryAIDefinite},
	}

	clusters, docs := ClusterTickets(tickets, classifications)

	if len(clusters) != 1 {
		t.Fatalf("expected 1 cluster, got %d", len(clusters))
	}
	if len(clusters[0].TicketIDs) != 1 {
		t.Errorf("expected 1 ticket in cluster, got %d", len(clusters[0].TicketIDs))
	}
	if len(docs) != 1 {
		t.Fatalf("expected 1 context doc, got %d", len(docs))
	}
}

func TestClusterTickets_GroupsRelated(t *testing.T) {
	tickets := []NormalizedTicket{
		{
			TicketID: "ENG-1",
			Signals: Signals{
				MentionsFiles: []string{"packs/auth/app/models/user.rb"},
				Domains:       []string{"backend"},
			},
		},
		{
			TicketID: "ENG-2",
			Signals: Signals{
				MentionsFiles: []string{"packs/auth/app/models/user.rb"},
				Domains:       []string{"backend"},
			},
		},
		{
			TicketID: "ENG-3",
			Signals: Signals{
				Domains: []string{"frontend"},
			},
		},
	}
	classifications := []Classification{
		{TicketID: "ENG-1", Category: CategoryAILikely},
		{TicketID: "ENG-2", Category: CategoryAILikely},
		{TicketID: "ENG-3", Category: CategoryAIDefinite},
	}

	clusters, _ := ClusterTickets(tickets, classifications)

	// ENG-1 and ENG-2 share a file (+3.0) and domain (+1.0) = 4.0 >= 2.0 threshold
	// ENG-3 shares nothing with them
	if len(clusters) < 2 {
		t.Errorf("expected at least 2 clusters (related pair + singleton), got %d", len(clusters))
	}

	// Find the cluster with 2 tickets
	foundPair := false
	for _, cl := range clusters {
		if len(cl.TicketIDs) == 2 {
			foundPair = true
			hasEng1, hasEng2 := false, false
			for _, id := range cl.TicketIDs {
				if id == "ENG-1" {
					hasEng1 = true
				}
				if id == "ENG-2" {
					hasEng2 = true
				}
			}
			if !hasEng1 || !hasEng2 {
				t.Errorf("expected ENG-1 and ENG-2 in same cluster, got %v", cl.TicketIDs)
			}
		}
	}
	if !foundPair {
		t.Error("expected a cluster with 2 tickets")
	}
}

func TestClusterTickets_NoOverlap(t *testing.T) {
	tickets := []NormalizedTicket{
		{TicketID: "ENG-1", Signals: Signals{Domains: []string{"frontend"}}},
		{TicketID: "ENG-2", Signals: Signals{Domains: []string{"backend"}}},
		{TicketID: "ENG-3", Signals: Signals{Domains: []string{"graphql"}}},
	}
	classifications := []Classification{
		{TicketID: "ENG-1", Category: CategoryAIDefinite},
		{TicketID: "ENG-2", Category: CategoryAIDefinite},
		{TicketID: "ENG-3", Category: CategoryAIDefinite},
	}

	clusters, _ := ClusterTickets(tickets, classifications)

	// Each domain is different, overlap is only 1.0 per shared domain = below 2.0 threshold
	// So each should be its own cluster
	if len(clusters) != 3 {
		t.Errorf("expected 3 singleton clusters, got %d", len(clusters))
	}
}

func TestDeriveClusterID(t *testing.T) {
	tests := []struct {
		name    string
		index   int
		tickets []NormalizedTicket
		want    string
	}{
		{
			name:  "with domain",
			index: 0,
			tickets: []NormalizedTicket{
				{Signals: Signals{Domains: []string{"frontend"}}},
				{Signals: Signals{Domains: []string{"frontend"}}},
			},
			want: "cluster-frontend-1",
		},
		{
			name:  "no domain",
			index: 2,
			tickets: []NormalizedTicket{
				{Signals: Signals{}},
			},
			want: "cluster-3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := deriveClusterID(tt.index, tt.tickets)
			if got != tt.want {
				t.Errorf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestBuildClusterRationale(t *testing.T) {
	// Standalone
	tickets := []NormalizedTicket{
		{TicketID: "ENG-1"},
	}
	r := buildClusterRationale(tickets)
	if r != "Standalone ticket: ENG-1" {
		t.Errorf("unexpected standalone rationale: %s", r)
	}

	// Multiple tickets with domains
	tickets = []NormalizedTicket{
		{TicketID: "ENG-1", Signals: Signals{Domains: []string{"frontend"}}},
		{TicketID: "ENG-2", Signals: Signals{Domains: []string{"frontend"}}},
	}
	r = buildClusterRationale(tickets)
	if r == "" {
		t.Error("expected non-empty rationale")
	}
}

func TestIntersect(t *testing.T) {
	tests := []struct {
		a, b []string
		want int
	}{
		{[]string{"a", "b", "c"}, []string{"b", "c", "d"}, 2},
		{[]string{"a"}, []string{"b"}, 0},
		{nil, []string{"a"}, 0},
		{[]string{}, []string{}, 0},
		{[]string{"x", "y"}, []string{"x", "y"}, 2},
	}

	for _, tt := range tests {
		result := intersect(tt.a, tt.b)
		if len(result) != tt.want {
			t.Errorf("intersect(%v, %v) = %v (len %d), want len %d", tt.a, tt.b, result, len(result), tt.want)
		}
	}
}

func TestSortedKeys(t *testing.T) {
	m := map[string]bool{"cherry": true, "apple": true, "banana": true}
	keys := sortedKeys(m)

	if len(keys) != 3 {
		t.Fatalf("expected 3 keys, got %d", len(keys))
	}
	if keys[0] != "apple" || keys[1] != "banana" || keys[2] != "cherry" {
		t.Errorf("expected sorted keys, got %v", keys)
	}
}

func TestUnionMentionsFiles(t *testing.T) {
	tickets := []NormalizedTicket{
		{Signals: Signals{MentionsFiles: []string{"a.rb", "b.rb"}}},
		{Signals: Signals{MentionsFiles: []string{"b.rb", "c.rb"}}},
	}

	result := unionMentionsFiles(tickets)

	if len(result) != 3 {
		t.Errorf("expected 3 unique files, got %d: %v", len(result), result)
	}
}

func TestDeriveCostCeiling(t *testing.T) {
	// AI_DEFINITE present -> smaller budget
	definite := deriveCostCeiling([]Classification{
		{Category: CategoryAIDefinite},
		{Category: CategoryAILikely},
	})
	if definite.MaxTokensPerTicket != 500000 {
		t.Errorf("expected 500K tokens for AI_DEFINITE, got %d", definite.MaxTokensPerTicket)
	}
	if definite.MaxAgentMinutesPerTicket != 30 {
		t.Errorf("expected 30 min for AI_DEFINITE, got %d", definite.MaxAgentMinutesPerTicket)
	}

	// No AI_DEFINITE -> larger budget
	likely := deriveCostCeiling([]Classification{
		{Category: CategoryAILikely},
		{Category: CategoryHumanReviewRequired},
	})
	if likely.MaxTokensPerTicket != 1000000 {
		t.Errorf("expected 1M tokens for AI_LIKELY, got %d", likely.MaxTokensPerTicket)
	}
	if likely.MaxAgentMinutesPerTicket != 60 {
		t.Errorf("expected 60 min for AI_LIKELY, got %d", likely.MaxAgentMinutesPerTicket)
	}

	// Empty classifications
	empty := deriveCostCeiling(nil)
	if empty.MaxTokensPerTicket != 1000000 {
		t.Errorf("expected 1M tokens for empty, got %d", empty.MaxTokensPerTicket)
	}
}

func TestGenerateContextDoc(t *testing.T) {
	cluster := Cluster{
		ClusterID: "cluster-frontend-1",
		Rationale: "2 tickets grouped by shared domains: frontend",
		TicketIDs: []string{"ENG-1", "ENG-2"},
		RepoAreas: []string{"next/packages/ui/Button.tsx"},
	}

	tickets := []NormalizedTicket{
		{
			TicketID: "ENG-1",
			Signals:  Signals{Domains: []string{"frontend"}},
		},
		{
			TicketID: "ENG-2",
			Signals:  Signals{Domains: []string{"frontend", "graphql"}},
		},
	}

	classifications := []Classification{
		{TicketID: "ENG-1", Category: CategoryAIDefinite, UncertainAxes: []string{"patternMatch"}},
		{TicketID: "ENG-2", Category: CategoryAILikely, UncertainAxes: []string{"blastRadius"}},
	}

	doc := generateContextDoc(cluster, tickets, classifications)

	if doc.ClusterID != "cluster-frontend-1" {
		t.Errorf("expected cluster ID cluster-frontend-1, got %s", doc.ClusterID)
	}
	if len(doc.TicketIDs) != 2 {
		t.Errorf("expected 2 ticket IDs, got %d", len(doc.TicketIDs))
	}
	if len(doc.KnownPatterns) == 0 {
		t.Error("expected known patterns from frontend/graphql domains")
	}
	if len(doc.Risks) != 2 {
		t.Errorf("expected 2 risks, got %d", len(doc.Risks))
	}
	// AI_DEFINITE present -> 500K budget
	if doc.CostCeiling.MaxTokensPerTicket != 500000 {
		t.Errorf("expected 500K tokens, got %d", doc.CostCeiling.MaxTokensPerTicket)
	}
}
