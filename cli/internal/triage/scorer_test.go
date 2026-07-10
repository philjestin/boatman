package triage

import (
	"context"
	"encoding/json"
	"testing"

	agentruntime "github.com/philjestin/boatman-ecosystem/shared/agentruntime"
	"github.com/philjestin/boatmanmode/internal/config"
	"github.com/philjestin/boatmanmode/internal/cost"
)

type scorerFakeProvider struct{}

func (scorerFakeProvider) Name() string {
	return "fake"
}

func (scorerFakeProvider) Capabilities(context.Context) (agentruntime.Capabilities, error) {
	return agentruntime.Capabilities{}, nil
}

func (scorerFakeProvider) StartRun(context.Context, agentruntime.RunRequest) (agentruntime.EventStream, error) {
	return nil, nil
}

func (scorerFakeProvider) ResumeRun(context.Context, string, agentruntime.RunInput) (agentruntime.EventStream, error) {
	return nil, nil
}

func (scorerFakeProvider) CancelRun(context.Context, string) error {
	return nil
}

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
			Labels:                     []string{"frontend", "bug"},
			MentionsFiles:              []string{"next/packages/ui/Button.tsx"},
			Domains:                    []string{"frontend"},
			Dependencies:               []string{"FE-100"},
			AcceptanceCriteriaPresent:  true,
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

func TestScorerOutputSchemaIsValidJSON(t *testing.T) {
	schema := scorerOutputSchema()
	if schema == nil {
		t.Fatal("schema should not be nil")
	}
	if !schema.Strict {
		t.Fatal("schema should be strict")
	}
	if !json.Valid(schema.Schema) {
		t.Fatalf("schema is not valid JSON: %s", schema.Schema)
	}
	if err := agentruntime.ValidateOutputSchema(schema); err != nil {
		t.Fatalf("schema should pass runtime validation: %v", err)
	}
}

func TestScoreUsesRuntimeProviderRequest(t *testing.T) {
	cfg := &config.Config{
		Claude: config.ClaudeConfig{
			Models: config.ModelConfig{Scorer: "score-model"},
		},
	}
	scorer := newScorerWithProvider(cfg, scorerFakeProvider{})

	var captured agentruntime.RunRequest
	scorer.runText = func(_ context.Context, provider agentruntime.Provider, req agentruntime.RunRequest) (string, *cost.Usage, error) {
		captured = req
		if provider.Name() != "fake" {
			t.Fatalf("provider = %q, want fake", provider.Name())
		}
		return `{"clarity": 4, "codeLocality": 3, "patternMatch": 2, "validationStrength": 1, "dependencyRisk": 0, "productAmbiguity": 1, "blastRadius": 2, "uncertainAxes": [], "reasons": ["ok"]}`, &cost.Usage{InputTokens: 3}, nil
	}

	response, usage, err := scorer.Score(context.Background(), NormalizedTicket{
		TicketID: "ENG-9",
		Title:    "Runtime scorer",
	})
	if err != nil {
		t.Fatalf("Score error: %v", err)
	}
	if response.Clarity != 4 {
		t.Fatalf("Clarity = %d, want 4", response.Clarity)
	}
	if usage == nil || usage.InputTokens != 3 {
		t.Fatalf("usage = %#v, want input tokens", usage)
	}
	if captured.Role != agentruntime.RoleScorer {
		t.Fatalf("Role = %q, want %q", captured.Role, agentruntime.RoleScorer)
	}
	if captured.Profile != "triage-scorer" {
		t.Fatalf("Profile = %q, want triage-scorer", captured.Profile)
	}
	if captured.Model != "score-model" {
		t.Fatalf("Model = %q, want score-model", captured.Model)
	}
	if captured.OutputSchema == nil || captured.OutputSchema.Name != "triage_scorer_response" {
		t.Fatalf("OutputSchema = %#v, want triage scorer schema", captured.OutputSchema)
	}
	if len(captured.Messages) != 1 || !containsStr(captured.Messages[0].Content, "ENG-9") {
		t.Fatalf("Messages = %#v, want ticket prompt", captured.Messages)
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
