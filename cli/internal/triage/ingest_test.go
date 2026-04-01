package triage

import (
	"testing"
	"time"

	"github.com/philjestin/boatmanmode/internal/linear"
)

func TestNormalize(t *testing.T) {
	now := time.Now()
	ticket := &linear.FullTicket{
		Ticket: linear.Ticket{
			Identifier:  "ENG-42",
			Title:       "Add search bar to dashboard",
			Description: "We need a search bar in packs/dashboard/app/views/search.tsx. Acceptance criteria:\n- [ ] Search bar renders\n- [ ] Results filter on input",
			Labels:      []string{"frontend", "enhancement"},
		},
		Comments: []linear.Comment{
			{Body: "Looks good", CreatedAt: now, UserName: "Alice"},
			{Body: "Agreed", CreatedAt: now, UserName: "Bob"},
		},
		Team:        "ENG",
		ProjectName: "Dashboard Revamp",
		UpdatedAt:   now,
	}

	n := Normalize(ticket, 168)

	if n.TicketID != "ENG-42" {
		t.Errorf("expected TicketID ENG-42, got %s", n.TicketID)
	}
	if n.Title != "Add search bar to dashboard" {
		t.Errorf("unexpected title: %s", n.Title)
	}
	if n.Signals.TeamKey != "ENG" {
		t.Errorf("expected team ENG, got %s", n.Signals.TeamKey)
	}
	if n.Signals.ProjectName != "Dashboard Revamp" {
		t.Errorf("expected project Dashboard Revamp, got %s", n.Signals.ProjectName)
	}
	if n.Signals.CommentCount != 2 {
		t.Errorf("expected 2 comments, got %d", n.Signals.CommentCount)
	}
	if !n.Signals.AcceptanceCriteriaPresent {
		t.Error("expected acceptance criteria present")
	}
	if !n.Signals.AcceptanceCriteriaExplicit {
		t.Error("expected acceptance criteria explicit (checkboxes found)")
	}

	// Staleness
	expectedStale := n.IngestedAt.Add(168 * time.Hour)
	if n.StaleAfter.Sub(expectedStale) > time.Second {
		t.Errorf("unexpected stale time: %v", n.StaleAfter)
	}
}

func TestNormalize_SummaryTruncation(t *testing.T) {
	longDesc := ""
	for i := 0; i < 300; i++ {
		longDesc += "x"
	}
	ticket := &linear.FullTicket{
		Ticket: linear.Ticket{
			Identifier:  "ENG-1",
			Description: longDesc,
		},
	}

	n := Normalize(ticket, 24)

	if len(n.Summary) != 200 {
		t.Errorf("expected summary length 200, got %d", len(n.Summary))
	}
	// Full description should be preserved
	if len(n.Description) != 300 {
		t.Errorf("expected description length 300, got %d", len(n.Description))
	}
}

func TestNormalizeBatch(t *testing.T) {
	tickets := []linear.FullTicket{
		{Ticket: linear.Ticket{Identifier: "ENG-1", Title: "First"}},
		{Ticket: linear.Ticket{Identifier: "ENG-2", Title: "Second"}},
		{Ticket: linear.Ticket{Identifier: "ENG-3", Title: "Third"}},
	}

	result := NormalizeBatch(tickets, 24)

	if len(result) != 3 {
		t.Fatalf("expected 3 results, got %d", len(result))
	}
	if result[0].TicketID != "ENG-1" {
		t.Errorf("expected first ticket ENG-1, got %s", result[0].TicketID)
	}
	if result[2].TicketID != "ENG-3" {
		t.Errorf("expected third ticket ENG-3, got %s", result[2].TicketID)
	}
}

func TestExtractFilePaths(t *testing.T) {
	tests := []struct {
		name  string
		text  string
		count int
		want  []string
	}{
		{
			name:  "single path",
			text:  "Check packs/dashboard/app/models/user.rb for the issue",
			count: 1,
			want:  []string{"packs/dashboard/app/models/user.rb"},
		},
		{
			name:  "multiple paths",
			text:  "Files: next/packages/ui/Button.tsx and packs/auth/app/services/login.rb",
			count: 2,
			want:  []string{"next/packages/ui/Button.tsx", "packs/auth/app/services/login.rb"},
		},
		{
			name:  "no paths",
			text:  "Just a regular description with no file references",
			count: 0,
		},
		{
			name:  "deduplication",
			text:  "See packs/foo/bar.rb and also packs/foo/bar.rb for context",
			count: 1,
			want:  []string{"packs/foo/bar.rb"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractFilePaths(tt.text)
			if len(result) != tt.count {
				t.Errorf("expected %d paths, got %d: %v", tt.count, len(result), result)
			}
			for i, want := range tt.want {
				if i < len(result) && result[i] != want {
					t.Errorf("path[%d]: expected %q, got %q", i, want, result[i])
				}
			}
		})
	}
}

func TestExtractDomains(t *testing.T) {
	tests := []struct {
		name   string
		labels []string
		files  []string
		text   string
		want   map[string]bool
	}{
		{
			name:   "from labels",
			labels: []string{"frontend", "bug"},
			want:   map[string]bool{"frontend": true},
		},
		{
			name:  "from file extensions",
			files: []string{"app/models/user.rb", "next/Button.tsx"},
			want:  map[string]bool{"backend": true, "frontend": true},
		},
		{
			name: "from description text",
			text: "Update the GraphQL resolver for the mutation",
			want: map[string]bool{"graphql": true},
		},
		{
			name:   "combined sources",
			labels: []string{"api"},
			files:  []string{"spec/models/user_spec.rb"},
			text:   "Add a REST endpoint with rspec tests",
			want:   map[string]bool{"api": true, "backend": true, "testing": true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractDomains(tt.labels, tt.files, tt.text)
			got := make(map[string]bool)
			for _, d := range result {
				got[d] = true
			}
			for domain := range tt.want {
				if !got[domain] {
					t.Errorf("expected domain %q in result %v", domain, result)
				}
			}
		})
	}
}

func TestExtractDependencies(t *testing.T) {
	tests := []struct {
		name  string
		text  string
		count int
		want  []string
	}{
		{
			name:  "single reference",
			text:  "Blocked by ENG-456",
			count: 1,
			want:  []string{"ENG-456"},
		},
		{
			name:  "multiple references",
			text:  "Depends on ENG-100, FE-200, and API-300",
			count: 3,
		},
		{
			name:  "no references",
			text:  "No ticket references here",
			count: 0,
		},
		{
			name:  "deduplication",
			text:  "See ENG-123 and also ENG-123",
			count: 1,
			want:  []string{"ENG-123"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractDependencies(tt.text)
			if len(result) != tt.count {
				t.Errorf("expected %d deps, got %d: %v", tt.count, len(result), result)
			}
			for i, want := range tt.want {
				if i < len(result) && result[i] != want {
					t.Errorf("dep[%d]: expected %q, got %q", i, want, result[i])
				}
			}
		})
	}
}

func TestDetectAcceptanceCriteria(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		present  bool
		explicit bool
	}{
		{
			name:     "checkboxes",
			text:     "Requirements:\n- [ ] Item 1\n- [ ] Item 2",
			present:  true,
			explicit: true,
		},
		{
			name:     "checked boxes",
			text:     "Done:\n- [x] Item 1\n- [ ] Item 2",
			present:  true,
			explicit: true,
		},
		{
			name:     "AC header",
			text:     "## Acceptance Criteria\nThe button should be blue",
			present:  true,
			explicit: true,
		},
		{
			name:     "numbered list only",
			text:     "Steps:\n1. Do this\n2. Do that\n3. Verify result",
			present:  true,
			explicit: false,
		},
		{
			name:     "no criteria",
			text:     "Just a vague description with no structure",
			present:  false,
			explicit: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			present, explicit := detectAcceptanceCriteria(tt.text)
			if present != tt.present {
				t.Errorf("present: expected %v, got %v", tt.present, present)
			}
			if explicit != tt.explicit {
				t.Errorf("explicit: expected %v, got %v", tt.explicit, explicit)
			}
		})
	}
}

func TestDetectDesignSpec(t *testing.T) {
	tests := []struct {
		text string
		want bool
	}{
		{"See the Figma link for design", true},
		{"Check the design spec attached", true},
		{"Based on the mockup from product", true},
		{"Wireframe is in the doc", true},
		{"View the prototype here", true},
		{"Just fix the bug please", false},
		{"No design artifacts", false},
	}

	for _, tt := range tests {
		got := detectDesignSpec(tt.text)
		if got != tt.want {
			t.Errorf("detectDesignSpec(%q) = %v, want %v", tt.text, got, tt.want)
		}
	}
}
