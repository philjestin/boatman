package triage

import (
	"testing"
)

func TestClassify_AIDefinite(t *testing.T) {
	ticket := &NormalizedTicket{
		TicketID: "ENG-100",
		Title:    "Fix button color",
		Signals: Signals{
			AcceptanceCriteriaPresent: true,
			HasDesignSpec:             true,
		},
	}

	scores := RubricScores{
		Clarity:            5,
		CodeLocality:       5,
		PatternMatch:       4,
		ValidationStrength: 4,
		DependencyRisk:     0,
		ProductAmbiguity:   0,
		BlastRadius:        0,
	}

	c := Classify(ticket, scores, nil, nil)

	if c.Category != CategoryAIDefinite {
		t.Errorf("expected AI_DEFINITE, got %s", c.Category)
	}
	if c.TicketID != "ENG-100" {
		t.Errorf("expected ticket ID ENG-100, got %s", c.TicketID)
	}
}

func TestClassify_AILikely(t *testing.T) {
	ticket := &NormalizedTicket{
		TicketID: "ENG-101",
		Title:    "Update form validation",
		Signals: Signals{
			AcceptanceCriteriaPresent: true,
		},
	}

	scores := RubricScores{
		Clarity:            3,
		CodeLocality:       3,
		PatternMatch:       2,
		ValidationStrength: 2,
		DependencyRisk:     2,
		ProductAmbiguity:   2,
		BlastRadius:        2,
	}

	c := Classify(ticket, scores, []string{"patternMatch"}, []string{"moderate confidence"})

	if c.Category != CategoryAILikely {
		t.Errorf("expected AI_LIKELY, got %s", c.Category)
	}
	if len(c.UncertainAxes) != 1 {
		t.Errorf("expected 1 uncertain axis, got %d", len(c.UncertainAxes))
	}
}

func TestClassify_HumanReviewRequired_GateFail(t *testing.T) {
	ticket := &NormalizedTicket{
		TicketID: "ENG-102",
		Title:    "Refactor module",
		Signals: Signals{
			AcceptanceCriteriaPresent: true,
		},
	}

	scores := RubricScores{
		Clarity:            1, // Below gate threshold of 2
		CodeLocality:       4,
		PatternMatch:       3,
		ValidationStrength: 3,
		DependencyRisk:     1,
		ProductAmbiguity:   1,
		BlastRadius:        1,
	}

	c := Classify(ticket, scores, nil, nil)

	if c.Category != CategoryHumanReviewRequired {
		t.Errorf("expected HUMAN_REVIEW_REQUIRED, got %s", c.Category)
	}

	// Should have gate results with clarity failing
	foundFailed := false
	for _, g := range c.GateResults {
		if g.Gate == "CLARITY_GATE" && !g.Passed {
			foundFailed = true
		}
	}
	if !foundFailed {
		t.Error("expected CLARITY_GATE to fail")
	}
}

func TestClassify_HumanReviewRequired_HighBlastRadius(t *testing.T) {
	ticket := &NormalizedTicket{
		TicketID: "ENG-103",
		Title:    "Update config",
		Signals: Signals{
			AcceptanceCriteriaPresent: true,
		},
	}

	scores := RubricScores{
		Clarity:            5,
		CodeLocality:       5,
		PatternMatch:       5,
		ValidationStrength: 5,
		DependencyRisk:     0,
		ProductAmbiguity:   0,
		BlastRadius:        4, // Above gate threshold of 3
	}

	c := Classify(ticket, scores, nil, nil)

	if c.Category != CategoryHumanReviewRequired {
		t.Errorf("expected HUMAN_REVIEW_REQUIRED, got %s", c.Category)
	}
}

func TestClassify_HumanReviewRequired_NoACNoSpec(t *testing.T) {
	ticket := &NormalizedTicket{
		TicketID: "ENG-104",
		Title:    "Improve performance",
		Signals: Signals{
			AcceptanceCriteriaPresent: false,
			HasDesignSpec:             false,
		},
	}

	scores := RubricScores{
		Clarity:            5,
		CodeLocality:       5,
		PatternMatch:       5,
		ValidationStrength: 5,
		DependencyRisk:     0,
		ProductAmbiguity:   0,
		BlastRadius:        0,
	}

	c := Classify(ticket, scores, nil, nil)

	if c.Category != CategoryHumanReviewRequired {
		t.Errorf("expected HUMAN_REVIEW_REQUIRED, got %s", c.Category)
	}
	if len(c.HardStops) == 0 {
		t.Error("expected hard stop reason for missing AC/spec")
	}
}

func TestClassify_HumanOnly_HardStop(t *testing.T) {
	tests := []struct {
		name  string
		title string
	}{
		{"payments", "Fix payments processing bug"},
		{"billing", "Update billing logic"},
		{"authentication", "Refactor authentication flow"},
		{"migration", "Run data migration for users"},
		{"incident", "Handle sev1 incident"},
		{"compliance", "GDPR compliance update"},
		{"multi-repo", "Multi-repo dependency update"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ticket := &NormalizedTicket{
				TicketID: "ENG-200",
				Title:    tt.title,
				Signals: Signals{
					AcceptanceCriteriaPresent: true,
					HasDesignSpec:             true,
				},
			}

			scores := RubricScores{
				Clarity:            5,
				CodeLocality:       5,
				PatternMatch:       5,
				ValidationStrength: 5,
				DependencyRisk:     0,
				ProductAmbiguity:   0,
				BlastRadius:        0,
			}

			c := Classify(ticket, scores, nil, nil)

			if c.Category != CategoryHumanOnly {
				t.Errorf("expected HUMAN_ONLY for %q, got %s", tt.title, c.Category)
			}
			if len(c.HardStops) == 0 {
				t.Errorf("expected hard stop reasons for %q", tt.title)
			}
		})
	}
}

func TestClassify_HumanReviewRequired_SoftStop(t *testing.T) {
	tests := []struct {
		name  string
		title string
	}{
		{"feature flag", "Add feature flag for new rollout"},
		{"staged rollout", "Staged rollout of search v2"},
		{"deploy coordination", "Deploy coordination for release train"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ticket := &NormalizedTicket{
				TicketID: "ENG-201",
				Title:    tt.title,
				Signals: Signals{
					AcceptanceCriteriaPresent: true,
					HasDesignSpec:             true,
				},
			}

			scores := RubricScores{
				Clarity:            5,
				CodeLocality:       5,
				PatternMatch:       5,
				ValidationStrength: 5,
				DependencyRisk:     0,
				ProductAmbiguity:   0,
				BlastRadius:        0,
			}

			c := Classify(ticket, scores, nil, nil)

			if c.Category != CategoryHumanReviewRequired {
				t.Errorf("expected HUMAN_REVIEW_REQUIRED for soft stop %q, got %s", tt.title, c.Category)
			}
		})
	}
}

func TestClassify_HardStopFromDescription(t *testing.T) {
	ticket := &NormalizedTicket{
		TicketID:    "ENG-300",
		Title:       "Update user profile page",
		Description: "This change touches the authentication flow and needs OAuth support",
		Signals: Signals{
			AcceptanceCriteriaPresent: true,
			HasDesignSpec:             true,
		},
	}

	scores := RubricScores{
		Clarity:            5,
		CodeLocality:       5,
		PatternMatch:       5,
		ValidationStrength: 5,
		DependencyRisk:     0,
		ProductAmbiguity:   0,
		BlastRadius:        0,
	}

	c := Classify(ticket, scores, nil, nil)

	if c.Category != CategoryHumanOnly {
		t.Errorf("expected HUMAN_ONLY from description keywords, got %s", c.Category)
	}
}

func TestClassify_HardStopFromLabels(t *testing.T) {
	ticket := &NormalizedTicket{
		TicketID: "ENG-301",
		Title:    "Update profile page",
		Signals: Signals{
			Labels:                    []string{"billing", "frontend"},
			AcceptanceCriteriaPresent: true,
			HasDesignSpec:             true,
		},
	}

	scores := RubricScores{
		Clarity:            5,
		CodeLocality:       5,
		PatternMatch:       5,
		ValidationStrength: 5,
		DependencyRisk:     0,
		ProductAmbiguity:   0,
		BlastRadius:        0,
	}

	c := Classify(ticket, scores, nil, nil)

	if c.Category != CategoryHumanOnly {
		t.Errorf("expected HUMAN_ONLY from label 'billing', got %s", c.Category)
	}
}

func TestClassify_DefaultToHumanReview(t *testing.T) {
	// Scores that don't meet AI_DEFINITE or AI_LIKELY thresholds
	ticket := &NormalizedTicket{
		TicketID: "ENG-400",
		Title:    "Vague improvement task",
		Signals: Signals{
			AcceptanceCriteriaPresent: true,
		},
	}

	scores := RubricScores{
		Clarity:            2,
		CodeLocality:       2,
		PatternMatch:       1,
		ValidationStrength: 1,
		DependencyRisk:     2,
		ProductAmbiguity:   2,
		BlastRadius:        2,
	}

	c := Classify(ticket, scores, nil, nil)

	if c.Category != CategoryHumanReviewRequired {
		t.Errorf("expected HUMAN_REVIEW_REQUIRED for borderline scores, got %s", c.Category)
	}
}

func TestEvaluateGates(t *testing.T) {
	tests := []struct {
		name    string
		scores  RubricScores
		allPass bool
	}{
		{
			name: "all pass",
			scores: RubricScores{
				Clarity: 3, BlastRadius: 2, ProductAmbiguity: 2, DependencyRisk: 2,
			},
			allPass: true,
		},
		{
			name: "clarity fails",
			scores: RubricScores{
				Clarity: 1, BlastRadius: 2, ProductAmbiguity: 2, DependencyRisk: 2,
			},
			allPass: false,
		},
		{
			name: "blast radius fails",
			scores: RubricScores{
				Clarity: 3, BlastRadius: 4, ProductAmbiguity: 2, DependencyRisk: 2,
			},
			allPass: false,
		},
		{
			name: "product ambiguity fails",
			scores: RubricScores{
				Clarity: 3, BlastRadius: 2, ProductAmbiguity: 4, DependencyRisk: 2,
			},
			allPass: false,
		},
		{
			name: "dependency risk fails",
			scores: RubricScores{
				Clarity: 3, BlastRadius: 2, ProductAmbiguity: 2, DependencyRisk: 4,
			},
			allPass: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gates := evaluateGates(tt.scores)
			allPassed := true
			for _, g := range gates {
				if !g.Passed {
					allPassed = false
					break
				}
			}
			if allPassed != tt.allPass {
				t.Errorf("expected allPass=%v, got %v", tt.allPass, allPassed)
			}
		})
	}
}

func TestContainsAny(t *testing.T) {
	tests := []struct {
		text     string
		keywords []string
		want     bool
	}{
		{"fix the payments flow", []string{"payments", "billing"}, true},
		{"update the button color", []string{"payments", "billing"}, false},
		{"handle stripe webhook", []string{"stripe", "invoice"}, true},
		{"", []string{"anything"}, false},
		{"some text", []string{}, false},
	}

	for _, tt := range tests {
		got := containsAny(tt.text, tt.keywords)
		if got != tt.want {
			t.Errorf("containsAny(%q, %v) = %v, want %v", tt.text, tt.keywords, got, tt.want)
		}
	}
}

func TestBuildSearchText(t *testing.T) {
	ticket := &NormalizedTicket{
		Title:       "Fix Payment Bug",
		Description: "The STRIPE integration is broken",
		Signals: Signals{
			Labels: []string{"Backend", "Urgent"},
		},
	}

	text := buildSearchText(ticket)

	// Should be lowercase
	if text != "fix payment bug the stripe integration is broken backend urgent" {
		t.Errorf("unexpected search text: %s", text)
	}
}
