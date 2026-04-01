package triage

import (
	"strings"
	"testing"
)

func TestFormatTriageComment_AIDefinite(t *testing.T) {
	c := Classification{
		TicketID: "ENG-100",
		Category: CategoryAIDefinite,
		Rubric: RubricScores{
			Clarity:            5,
			CodeLocality:       4,
			PatternMatch:       4,
			ValidationStrength: 3,
			DependencyRisk:     1,
			ProductAmbiguity:   0,
			BlastRadius:        0,
		},
		GateResults: []GateResult{
			{Gate: "CLARITY_GATE", Passed: true},
			{Gate: "BLAST_RADIUS_GATE", Passed: true},
		},
		Reasons:       []string{"Clear requirements", "Single file change"},
		UncertainAxes: []string{"patternMatch"},
	}

	comment := FormatTriageComment(c)

	// Check header
	if !strings.Contains(comment, "## AI Triage: AI_DEFINITE") {
		t.Error("expected AI_DEFINITE header")
	}

	// Check rubric table
	if !strings.Contains(comment, "| Clarity | 5/5 |") {
		t.Error("expected clarity score in table")
	}
	if !strings.Contains(comment, "| Blast Radius | 0/5 |") {
		t.Error("expected blast radius score in table")
	}

	// Check positive markers
	if !strings.Contains(comment, "(good)") {
		t.Error("expected (good) marker for high positive scores")
	}

	// Check negative markers (low risk for low negative scores)
	if !strings.Contains(comment, "(low risk)") {
		t.Error("expected (low risk) marker for low negative scores")
	}

	// Check gate results
	if !strings.Contains(comment, "CLARITY_GATE: pass") {
		t.Error("expected passing gate result")
	}

	// Check reasons
	if !strings.Contains(comment, "Clear requirements") {
		t.Error("expected reasons section")
	}

	// Check uncertain axes
	if !strings.Contains(comment, "patternMatch") {
		t.Error("expected uncertain axes section")
	}

	// Check timestamp
	if !strings.Contains(comment, "Triaged by Boatman at") {
		t.Error("expected timestamp footer")
	}
}

func TestFormatTriageComment_WithHardStops(t *testing.T) {
	c := Classification{
		TicketID:  "ENG-200",
		Category:  CategoryHumanOnly,
		Rubric:    RubricScores{Clarity: 5},
		HardStops: []string{"payments/billing", "authentication/authorization"},
	}

	comment := FormatTriageComment(c)

	if !strings.Contains(comment, "## AI Triage: HUMAN_ONLY") {
		t.Error("expected HUMAN_ONLY header")
	}
	if !strings.Contains(comment, "### Hard Stops") {
		t.Error("expected hard stops section")
	}
	if !strings.Contains(comment, "payments/billing") {
		t.Error("expected payments hard stop")
	}
	if !strings.Contains(comment, "authentication/authorization") {
		t.Error("expected auth hard stop")
	}
}

func TestFormatTriageComment_FailedGates(t *testing.T) {
	c := Classification{
		TicketID: "ENG-300",
		Category: CategoryHumanReviewRequired,
		Rubric:   RubricScores{Clarity: 1, BlastRadius: 4},
		GateResults: []GateResult{
			{Gate: "CLARITY_GATE", Passed: false, Reason: "clarity < 2 triggers human review"},
			{Gate: "BLAST_RADIUS_GATE", Passed: false, Reason: "blastRadius > 3 triggers human review"},
		},
	}

	comment := FormatTriageComment(c)

	if !strings.Contains(comment, "CLARITY_GATE: FAIL") {
		t.Error("expected failed clarity gate")
	}
	if !strings.Contains(comment, "BLAST_RADIUS_GATE: FAIL") {
		t.Error("expected failed blast radius gate")
	}
}

func TestFormatTriageComment_MinimalFields(t *testing.T) {
	c := Classification{
		TicketID: "ENG-400",
		Category: CategoryAILikely,
		Rubric:   RubricScores{},
	}

	comment := FormatTriageComment(c)

	// Should still produce valid output without optional sections
	if !strings.Contains(comment, "## AI Triage: AI_LIKELY") {
		t.Error("expected AI_LIKELY header")
	}
	// No reasons/uncertainAxes/hardStops sections
	if strings.Contains(comment, "### Reasons") {
		t.Error("should not have reasons section when empty")
	}
	if strings.Contains(comment, "### Uncertain Axes") {
		t.Error("should not have uncertain axes section when empty")
	}
	if strings.Contains(comment, "### Hard Stops") {
		t.Error("should not have hard stops section when empty")
	}
}

func TestPositiveMarker(t *testing.T) {
	tests := []struct {
		score int
		want  string
	}{
		{0, ""},
		{1, ""},
		{2, ""},
		{3, "(good)"},
		{4, "(good)"},
		{5, "(good)"},
	}

	for _, tt := range tests {
		got := positiveMarker(tt.score)
		if got != tt.want {
			t.Errorf("positiveMarker(%d) = %q, want %q", tt.score, got, tt.want)
		}
	}
}

func TestNegativeMarker(t *testing.T) {
	tests := []struct {
		score int
		want  string
	}{
		{0, "(low risk)"},
		{1, "(low risk)"},
		{2, "(low risk)"},
		{3, ""},
		{4, ""},
		{5, ""},
	}

	for _, tt := range tests {
		got := negativeMarker(tt.score)
		if got != tt.want {
			t.Errorf("negativeMarker(%d) = %q, want %q", tt.score, got, tt.want)
		}
	}
}
