package triage

import (
	"encoding/json"
	"testing"
	"time"
)

func TestCategory_Constants(t *testing.T) {
	tests := []struct {
		cat  Category
		want string
	}{
		{CategoryAIDefinite, "AI_DEFINITE"},
		{CategoryAILikely, "AI_LIKELY"},
		{CategoryHumanReviewRequired, "HUMAN_REVIEW_REQUIRED"},
		{CategoryHumanOnly, "HUMAN_ONLY"},
	}

	for _, tt := range tests {
		if string(tt.cat) != tt.want {
			t.Errorf("expected %q, got %q", tt.want, string(tt.cat))
		}
	}
}

func TestStage_Constants(t *testing.T) {
	tests := []struct {
		stage Stage
		want  string
	}{
		{StageIngest, "ingest"},
		{StageScore, "score"},
		{StageCluster, "cluster"},
	}

	for _, tt := range tests {
		if string(tt.stage) != tt.want {
			t.Errorf("expected %q, got %q", tt.want, string(tt.stage))
		}
	}
}

func TestNormalizedTicket_JSON(t *testing.T) {
	ticket := NormalizedTicket{
		TicketID:    "ENG-42",
		Title:       "Test ticket",
		Summary:     "A test",
		Description: "Full description",
		IngestedAt:  time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		StaleAfter:  time.Date(2026, 1, 8, 0, 0, 0, 0, time.UTC),
		Signals: Signals{
			MentionsFiles:             []string{"foo.rb"},
			Domains:                   []string{"backend"},
			Labels:                    []string{"bug"},
			AcceptanceCriteriaPresent: true,
			CommentCount:              3,
			TeamKey:                   "ENG",
		},
	}

	data, err := json.Marshal(ticket)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var parsed NormalizedTicket
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if parsed.TicketID != "ENG-42" {
		t.Errorf("expected ENG-42, got %s", parsed.TicketID)
	}
	if parsed.Signals.TeamKey != "ENG" {
		t.Errorf("expected team ENG, got %s", parsed.Signals.TeamKey)
	}
	if !parsed.Signals.AcceptanceCriteriaPresent {
		t.Error("expected AcceptanceCriteriaPresent to be true")
	}
}

func TestClassification_JSON(t *testing.T) {
	c := Classification{
		TicketID: "ENG-1",
		Category: CategoryAIDefinite,
		Rubric: RubricScores{
			Clarity:            5,
			CodeLocality:       4,
			PatternMatch:       3,
			ValidationStrength: 4,
			DependencyRisk:     1,
			ProductAmbiguity:   0,
			BlastRadius:        0,
		},
		UncertainAxes: []string{"patternMatch"},
		Reasons:       []string{"clear AC"},
		HardStops:     nil,
		GateResults: []GateResult{
			{Gate: "CLARITY_GATE", Passed: true},
		},
	}

	data, err := json.Marshal(c)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var parsed Classification
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if parsed.Category != CategoryAIDefinite {
		t.Errorf("expected AI_DEFINITE, got %s", parsed.Category)
	}
	if parsed.Rubric.Clarity != 5 {
		t.Errorf("expected clarity 5, got %d", parsed.Rubric.Clarity)
	}
	if len(parsed.GateResults) != 1 {
		t.Errorf("expected 1 gate result, got %d", len(parsed.GateResults))
	}
}

func TestDecisionLogEntry_JSON(t *testing.T) {
	entry := DecisionLogEntry{
		TicketID:   "ENG-1",
		Stage:      StageScore,
		Verdict:    "AI_DEFINITE",
		Agent:      "triage-pipeline",
		Rationale:  "all gates passed",
		Timestamp:  time.Date(2026, 3, 31, 12, 0, 0, 0, time.UTC),
		TokensUsed: 5000,
		CostUSD:    0.01,
		Model:      "claude-sonnet-4-6",
		Details:    json.RawMessage(`{"clarity":5}`),
	}

	data, err := json.Marshal(entry)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var parsed DecisionLogEntry
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if parsed.TicketID != "ENG-1" {
		t.Errorf("expected ENG-1, got %s", parsed.TicketID)
	}
	if parsed.Stage != StageScore {
		t.Errorf("expected score stage, got %s", parsed.Stage)
	}
	if parsed.TokensUsed != 5000 {
		t.Errorf("expected 5000 tokens, got %d", parsed.TokensUsed)
	}
}

func TestScorerResponse_EmbedsRubricScores(t *testing.T) {
	resp := ScorerResponse{
		RubricScores: RubricScores{
			Clarity:    4,
			BlastRadius: 1,
		},
		UncertainAxes: []string{"clarity"},
		Reasons:       []string{"moderate confidence"},
	}

	// Verify embedding works - can access RubricScores fields directly
	if resp.Clarity != 4 {
		t.Errorf("expected clarity 4, got %d", resp.Clarity)
	}
	if resp.BlastRadius != 1 {
		t.Errorf("expected blastRadius 1, got %d", resp.BlastRadius)
	}
}

func TestTriageStats_JSON(t *testing.T) {
	stats := TriageStats{
		TotalTickets:    10,
		AIDefiniteCount: 3,
		AILikelyCount:   4,
		HumanReviewCount: 2,
		HumanOnlyCount:  1,
		ClusterCount:    3,
		TotalTokensUsed: 50000,
		TotalCostUSD:    0.15,
	}

	data, err := json.Marshal(stats)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var parsed TriageStats
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if parsed.TotalTickets != 10 {
		t.Errorf("expected 10 total, got %d", parsed.TotalTickets)
	}
	if parsed.AIDefiniteCount+parsed.AILikelyCount+parsed.HumanReviewCount+parsed.HumanOnlyCount != 10 {
		t.Error("category counts should sum to total")
	}
}
