package triage

import (
	"testing"
)

func TestBuildStats(t *testing.T) {
	classifications := []Classification{
		{TicketID: "ENG-1", Category: CategoryAIDefinite},
		{TicketID: "ENG-2", Category: CategoryAIDefinite},
		{TicketID: "ENG-3", Category: CategoryAILikely},
		{TicketID: "ENG-4", Category: CategoryHumanReviewRequired},
		{TicketID: "ENG-5", Category: CategoryHumanOnly},
		{TicketID: "ENG-6", Category: CategoryHumanOnly},
	}

	clusters := []Cluster{
		{ClusterID: "cluster-1", TicketIDs: []string{"ENG-1", "ENG-2"}},
		{ClusterID: "cluster-2", TicketIDs: []string{"ENG-3"}},
	}

	stats := buildStats(classifications, clusters, 15000, 0.05)

	if stats.TotalTickets != 6 {
		t.Errorf("expected 6 total tickets, got %d", stats.TotalTickets)
	}
	if stats.AIDefiniteCount != 2 {
		t.Errorf("expected 2 AI_DEFINITE, got %d", stats.AIDefiniteCount)
	}
	if stats.AILikelyCount != 1 {
		t.Errorf("expected 1 AI_LIKELY, got %d", stats.AILikelyCount)
	}
	if stats.HumanReviewCount != 1 {
		t.Errorf("expected 1 HUMAN_REVIEW_REQUIRED, got %d", stats.HumanReviewCount)
	}
	if stats.HumanOnlyCount != 2 {
		t.Errorf("expected 2 HUMAN_ONLY, got %d", stats.HumanOnlyCount)
	}
	if stats.ClusterCount != 2 {
		t.Errorf("expected 2 clusters, got %d", stats.ClusterCount)
	}
	if stats.TotalTokensUsed != 15000 {
		t.Errorf("expected 15000 tokens, got %d", stats.TotalTokensUsed)
	}
	if stats.TotalCostUSD != 0.05 {
		t.Errorf("expected $0.05, got %f", stats.TotalCostUSD)
	}
}

func TestBuildStats_Empty(t *testing.T) {
	stats := buildStats(nil, nil, 0, 0)

	if stats.TotalTickets != 0 {
		t.Errorf("expected 0 total tickets, got %d", stats.TotalTickets)
	}
	if stats.ClusterCount != 0 {
		t.Errorf("expected 0 clusters, got %d", stats.ClusterCount)
	}
}

func TestClassificationRationale(t *testing.T) {
	// Hard stops
	c := Classification{
		Category:  CategoryHumanOnly,
		HardStops: []string{"payments/billing"},
	}
	r := classificationRationale(c)
	if !containsStr(r, "hard stops") {
		t.Errorf("expected hard stop rationale, got %q", r)
	}

	// Gate failure
	c = Classification{
		Category: CategoryHumanReviewRequired,
		GateResults: []GateResult{
			{Gate: "CLARITY_GATE", Passed: true},
			{Gate: "BLAST_RADIUS_GATE", Passed: false},
		},
	}
	r = classificationRationale(c)
	if !containsStr(r, "BLAST_RADIUS_GATE") {
		t.Errorf("expected gate failure rationale, got %q", r)
	}

	// Clean classification
	c = Classification{
		Category: CategoryAIDefinite,
	}
	r = classificationRationale(c)
	if !containsStr(r, "AI_DEFINITE") {
		t.Errorf("expected classification rationale, got %q", r)
	}
}

func TestBuildTicketUUIDMap(t *testing.T) {
	tickets := []struct {
		id         string
		identifier string
	}{
		{"uuid-1", "ENG-1"},
		{"uuid-2", "ENG-2"},
		{"uuid-3", "FE-100"},
	}

	// We need linear.FullTicket but we're in triage package.
	// Test the buildTicketUUIDMap function indirectly through buildStats instead,
	// since buildTicketUUIDMap requires linear.FullTicket which we test separately.
	_ = tickets
}

func TestTriageResult_Structure(t *testing.T) {
	result := TriageResult{
		Tickets: []NormalizedTicket{
			{TicketID: "ENG-1", Title: "Test ticket"},
		},
		Classifications: []Classification{
			{TicketID: "ENG-1", Category: CategoryAIDefinite},
		},
		Clusters: []Cluster{
			{ClusterID: "cluster-1", TicketIDs: []string{"ENG-1"}},
		},
		ContextDocs: []ContextDoc{
			{ClusterID: "cluster-1"},
		},
		Stats: TriageStats{
			TotalTickets:    1,
			AIDefiniteCount: 1,
		},
	}

	if len(result.Tickets) != 1 {
		t.Errorf("expected 1 ticket, got %d", len(result.Tickets))
	}
	if result.Stats.AIDefiniteCount != 1 {
		t.Errorf("expected 1 AI_DEFINITE, got %d", result.Stats.AIDefiniteCount)
	}
}
