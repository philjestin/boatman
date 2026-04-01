package triage

import (
	"testing"
)

func TestParseScoreResponse_RawJSON(t *testing.T) {
	response := `{"clarity": 4, "codeLocality": 3, "patternMatch": 5, "validationStrength": 2, "dependencyRisk": 1, "productAmbiguity": 0, "blastRadius": 2, "uncertainAxes": ["validationStrength"], "reasons": ["clear requirements", "single module"]}`

	scored, err := parseScoreResponse(response)
	if err != nil {
		t.Fatalf("parseScoreResponse failed: %v", err)
	}

	if scored.Clarity != 4 {
		t.Errorf("expected clarity 4, got %d", scored.Clarity)
	}
	if scored.CodeLocality != 3 {
		t.Errorf("expected codeLocality 3, got %d", scored.CodeLocality)
	}
	if scored.PatternMatch != 5 {
		t.Errorf("expected patternMatch 5, got %d", scored.PatternMatch)
	}
	if scored.ValidationStrength != 2 {
		t.Errorf("expected validationStrength 2, got %d", scored.ValidationStrength)
	}
	if scored.DependencyRisk != 1 {
		t.Errorf("expected dependencyRisk 1, got %d", scored.DependencyRisk)
	}
	if scored.ProductAmbiguity != 0 {
		t.Errorf("expected productAmbiguity 0, got %d", scored.ProductAmbiguity)
	}
	if scored.BlastRadius != 2 {
		t.Errorf("expected blastRadius 2, got %d", scored.BlastRadius)
	}
	if len(scored.UncertainAxes) != 1 || scored.UncertainAxes[0] != "validationStrength" {
		t.Errorf("expected uncertainAxes [validationStrength], got %v", scored.UncertainAxes)
	}
	if len(scored.Reasons) != 2 {
		t.Errorf("expected 2 reasons, got %d", len(scored.Reasons))
	}
}

func TestParseScoreResponse_FencedCodeBlock(t *testing.T) {
	response := "Here is my analysis:\n```json\n{\"clarity\": 5, \"codeLocality\": 4, \"patternMatch\": 3, \"validationStrength\": 4, \"dependencyRisk\": 1, \"productAmbiguity\": 0, \"blastRadius\": 0, \"uncertainAxes\": [], \"reasons\": [\"test\"]}\n```\nThat's my evaluation."

	scored, err := parseScoreResponse(response)
	if err != nil {
		t.Fatalf("parseScoreResponse failed: %v", err)
	}

	if scored.Clarity != 5 {
		t.Errorf("expected clarity 5, got %d", scored.Clarity)
	}
	if scored.BlastRadius != 0 {
		t.Errorf("expected blastRadius 0, got %d", scored.BlastRadius)
	}
}

func TestParseScoreResponse_FencedNoLanguage(t *testing.T) {
	response := "```\n{\"clarity\": 3, \"codeLocality\": 3, \"patternMatch\": 3, \"validationStrength\": 3, \"dependencyRisk\": 1, \"productAmbiguity\": 1, \"blastRadius\": 1, \"uncertainAxes\": [], \"reasons\": []}\n```"

	scored, err := parseScoreResponse(response)
	if err != nil {
		t.Fatalf("parseScoreResponse failed: %v", err)
	}

	if scored.Clarity != 3 {
		t.Errorf("expected clarity 3, got %d", scored.Clarity)
	}
}

func TestParseScoreResponse_EmbeddedInText(t *testing.T) {
	response := "Based on my analysis, I score this ticket as follows: {\"clarity\": 2, \"codeLocality\": 1, \"patternMatch\": 0, \"validationStrength\": 1, \"dependencyRisk\": 4, \"productAmbiguity\": 3, \"blastRadius\": 5, \"uncertainAxes\": [\"clarity\", \"patternMatch\"], \"reasons\": [\"vague\", \"novel\"]} I hope this helps!"

	scored, err := parseScoreResponse(response)
	if err != nil {
		t.Fatalf("parseScoreResponse failed: %v", err)
	}

	if scored.Clarity != 2 {
		t.Errorf("expected clarity 2, got %d", scored.Clarity)
	}
	if scored.BlastRadius != 5 {
		t.Errorf("expected blastRadius 5, got %d", scored.BlastRadius)
	}
}

func TestParseScoreResponse_NoJSON(t *testing.T) {
	response := "I cannot evaluate this ticket because the description is empty."

	_, err := parseScoreResponse(response)
	if err == nil {
		t.Error("expected error for response with no JSON")
	}
}

func TestParseScoreResponse_InvalidJSON(t *testing.T) {
	response := "{clarity: 5, this is not valid json}"

	_, err := parseScoreResponse(response)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestClampScore(t *testing.T) {
	tests := []struct {
		input int
		want  int
	}{
		{-1, 0},
		{-100, 0},
		{0, 0},
		{1, 1},
		{3, 3},
		{5, 5},
		{6, 5},
		{100, 5},
	}

	for _, tt := range tests {
		got := clampScore(tt.input)
		if got != tt.want {
			t.Errorf("clampScore(%d) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestParseScoreResponse_ClampsOutOfRange(t *testing.T) {
	response := `{"clarity": 10, "codeLocality": -3, "patternMatch": 5, "validationStrength": 0, "dependencyRisk": 7, "productAmbiguity": 5, "blastRadius": 0, "uncertainAxes": [], "reasons": []}`

	scored, err := parseScoreResponse(response)
	if err != nil {
		t.Fatalf("parseScoreResponse failed: %v", err)
	}

	if scored.Clarity != 5 {
		t.Errorf("expected clarity clamped to 5, got %d", scored.Clarity)
	}
	if scored.CodeLocality != 0 {
		t.Errorf("expected codeLocality clamped to 0, got %d", scored.CodeLocality)
	}
	if scored.DependencyRisk != 5 {
		t.Errorf("expected dependencyRisk clamped to 5, got %d", scored.DependencyRisk)
	}
}

func TestBuildUserPrompt(t *testing.T) {
	ticket := NormalizedTicket{
		TicketID:    "ENG-42",
		Title:       "Fix button alignment",
		Description: "The submit button is misaligned on mobile viewports",
		Signals: Signals{
			Labels:                    []string{"frontend", "bug"},
			MentionsFiles:             []string{"next/packages/ui/Button.tsx"},
			Domains:                   []string{"frontend"},
			Dependencies:              []string{"FE-100"},
			AcceptanceCriteriaPresent: true,
			AcceptanceCriteriaExplicit: false,
			HasDesignSpec:              true,
		},
	}

	prompt := buildUserPrompt(ticket)

	// Should contain ticket ID and title
	if !containsStr(prompt, "ENG-42") {
		t.Error("expected ticket ID in prompt")
	}
	if !containsStr(prompt, "Fix button alignment") {
		t.Error("expected title in prompt")
	}

	// Should contain labels
	if !containsStr(prompt, "frontend, bug") {
		t.Error("expected labels in prompt")
	}

	// Should contain signals
	if !containsStr(prompt, "next/packages/ui/Button.tsx") {
		t.Error("expected mentioned files in prompt")
	}
	if !containsStr(prompt, "FE-100") {
		t.Error("expected dependencies in prompt")
	}
}

func TestBuildUserPrompt_TruncatesLongDescription(t *testing.T) {
	longDesc := ""
	for i := 0; i < 5000; i++ {
		longDesc += "x"
	}

	ticket := NormalizedTicket{
		TicketID:    "ENG-1",
		Title:       "Test",
		Description: longDesc,
	}

	prompt := buildUserPrompt(ticket)

	// The description in the prompt should be truncated to 3000 chars
	// The total prompt will be longer due to template text
	if len(prompt) > 4000 {
		t.Errorf("prompt too long, expected truncation: len=%d", len(prompt))
	}
}

func TestBuildUserPrompt_EmptySignals(t *testing.T) {
	ticket := NormalizedTicket{
		TicketID: "ENG-1",
		Title:    "Test",
	}

	prompt := buildUserPrompt(ticket)

	if !containsStr(prompt, "none") {
		t.Error("expected 'none' for empty labels")
	}
	if !containsStr(prompt, "none detected") {
		t.Error("expected 'none detected' for empty signals")
	}
}

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && findSubstr(s, substr))
}

func findSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
