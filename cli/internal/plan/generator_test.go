package plan

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	agentruntime "github.com/philjestin/boatman-ecosystem/shared/agentruntime"
	"github.com/philjestin/boatmanmode/internal/config"
	"github.com/philjestin/boatmanmode/internal/cost"
	"github.com/philjestin/boatmanmode/internal/triage"
)

type generatorFakeProvider struct{}

func (generatorFakeProvider) Name() string {
	return "fake"
}

func (generatorFakeProvider) Capabilities(context.Context) (agentruntime.Capabilities, error) {
	return agentruntime.Capabilities{}, nil
}

func (generatorFakeProvider) StartRun(context.Context, agentruntime.RunRequest) (agentruntime.EventStream, error) {
	return nil, nil
}

func (generatorFakeProvider) ResumeRun(context.Context, string, agentruntime.RunInput) (agentruntime.EventStream, error) {
	return nil, nil
}

func (generatorFakeProvider) CancelRun(context.Context, string) error {
	return nil
}

func TestParsePlanResponse_JSONBlock(t *testing.T) {
	response := "Here is the plan:\n\n```json\n" + `{
  "approach": "Update the component",
  "candidateFiles": ["src/Foo.tsx"],
  "newFiles": [],
  "deletedFiles": [],
  "validation": ["yarn test"],
  "rollback": "Revert PR",
  "stopConditions": ["If tests fail"],
  "uncertainties": []
}` + "\n```\n\nDone."

	plan, err := parsePlanResponse(response)
	if err != nil {
		t.Fatalf("parsePlanResponse failed: %v", err)
	}

	if plan.Approach != "Update the component" {
		t.Errorf("unexpected approach: %s", plan.Approach)
	}
	if len(plan.CandidateFiles) != 1 || plan.CandidateFiles[0] != "src/Foo.tsx" {
		t.Errorf("unexpected candidate files: %v", plan.CandidateFiles)
	}
	if plan.Rollback != "Revert PR" {
		t.Errorf("unexpected rollback: %s", plan.Rollback)
	}
	if len(plan.StopConditions) != 1 {
		t.Errorf("expected 1 stop condition, got %d", len(plan.StopConditions))
	}
}

func TestParsePlanResponse_PlainJSON(t *testing.T) {
	// No markdown fences, just raw JSON.
	response := `{"approach":"direct json","candidateFiles":["a.ts"],"newFiles":[],"deletedFiles":[],"validation":["yarn test"],"rollback":"revert","stopConditions":["stop"],"uncertainties":[]}`

	plan, err := parsePlanResponse(response)
	if err != nil {
		t.Fatalf("parsePlanResponse failed: %v", err)
	}

	if plan.Approach != "direct json" {
		t.Errorf("unexpected approach: %s", plan.Approach)
	}
}

func TestParsePlanResponse_NoJSON(t *testing.T) {
	response := "I couldn't generate a plan because the ticket is too vague."

	_, err := parsePlanResponse(response)
	if err == nil {
		t.Error("expected error for response with no JSON")
	}
}

func TestParsePlanResponse_InvalidJSON(t *testing.T) {
	response := "```json\n{not valid json}\n```"

	_, err := parsePlanResponse(response)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestParsePlanResponse_JSONWithoutLangTag(t *testing.T) {
	response := "```\n" + `{"approach":"no lang tag","candidateFiles":[],"newFiles":[],"deletedFiles":[],"validation":["yarn test"],"rollback":"revert","stopConditions":["stop"],"uncertainties":[]}` + "\n```"

	plan, err := parsePlanResponse(response)
	if err != nil {
		t.Fatalf("parsePlanResponse failed: %v", err)
	}

	if plan.Approach != "no lang tag" {
		t.Errorf("unexpected approach: %s", plan.Approach)
	}
}

func TestBuildPlannerPrompt_BasicTicket(t *testing.T) {
	ticket := triage.NormalizedTicket{
		TicketID:    "ENG-42",
		Title:       "Fix button color",
		Description: "The button should be blue",
		Signals: triage.Signals{
			MentionsFiles: []string{"src/Button.tsx"},
			Domains:       []string{"frontend"},
		},
	}

	classification := triage.Classification{
		TicketID: "ENG-42",
		Category: triage.CategoryAIDefinite,
		Rubric: triage.RubricScores{
			Clarity:      5,
			CodeLocality: 4,
			PatternMatch: 3,
		},
		Reasons: []string{"clear AC", "single file change"},
	}

	prompt := buildPlannerPrompt(ticket, classification, nil)

	if !strings.Contains(prompt, "ENG-42") {
		t.Error("prompt should contain ticket ID")
	}
	if !strings.Contains(prompt, "Fix button color") {
		t.Error("prompt should contain title")
	}
	if !strings.Contains(prompt, "The button should be blue") {
		t.Error("prompt should contain description")
	}
	if !strings.Contains(prompt, "src/Button.tsx") {
		t.Error("prompt should contain mentioned files")
	}
	if !strings.Contains(prompt, "frontend") {
		t.Error("prompt should contain domains")
	}
	if !strings.Contains(prompt, "AI_DEFINITE") {
		t.Error("prompt should contain classification category")
	}
	if !strings.Contains(prompt, "clear AC") {
		t.Error("prompt should contain scoring reasons")
	}
}

func TestBuildPlannerPrompt_WithContextDoc(t *testing.T) {
	ticket := triage.NormalizedTicket{
		TicketID: "ENG-42",
		Title:    "Fix button",
	}

	classification := triage.Classification{
		TicketID: "ENG-42",
		Category: triage.CategoryAIDefinite,
	}

	contextDoc := &triage.ContextDoc{
		ClusterID:      "cluster-frontend",
		Rationale:      "Related UI components",
		RepoAreas:      []string{"src/components/", "src/styles/"},
		KnownPatterns:  []string{"Use Rosetta components"},
		ValidationPlan: []string{"yarn test", "yarn check-types"},
		Risks:          []string{"May affect other buttons"},
		CostCeiling: triage.CostCeiling{
			MaxTokensPerTicket:       100000,
			MaxAgentMinutesPerTicket: 10,
		},
	}

	prompt := buildPlannerPrompt(ticket, classification, contextDoc)

	if !strings.Contains(prompt, "cluster-frontend") {
		t.Error("prompt should contain cluster ID")
	}
	if !strings.Contains(prompt, "src/components/") {
		t.Error("prompt should contain repo areas")
	}
	if !strings.Contains(prompt, "Use Rosetta components") {
		t.Error("prompt should contain known patterns")
	}
	if !strings.Contains(prompt, "yarn test") {
		t.Error("prompt should contain validation commands")
	}
	if !strings.Contains(prompt, "May affect other buttons") {
		t.Error("prompt should contain risks")
	}
	if !strings.Contains(prompt, "100K tokens") {
		t.Error("prompt should contain cost ceiling")
	}
}

func TestBuildPlannerPrompt_TruncatesLongDescription(t *testing.T) {
	longDesc := strings.Repeat("x", 5000)
	ticket := triage.NormalizedTicket{
		TicketID:    "ENG-42",
		Title:       "Fix thing",
		Description: longDesc,
	}

	classification := triage.Classification{
		TicketID: "ENG-42",
		Category: triage.CategoryAIDefinite,
	}

	prompt := buildPlannerPrompt(ticket, classification, nil)

	// Description should be truncated to 3000 chars + truncation marker.
	if len(prompt) >= len(longDesc) {
		t.Error("prompt should truncate long descriptions")
	}
	if !strings.Contains(prompt, "...(truncated)") {
		t.Error("prompt should contain truncation marker")
	}
}

func TestBuildPlannerPrompt_NoDescription(t *testing.T) {
	ticket := triage.NormalizedTicket{
		TicketID: "ENG-42",
		Title:    "Fix thing",
	}

	classification := triage.Classification{
		TicketID: "ENG-42",
		Category: triage.CategoryAIDefinite,
	}

	prompt := buildPlannerPrompt(ticket, classification, nil)

	if strings.Contains(prompt, "## Description") {
		t.Error("prompt should not contain description section when description is empty")
	}
}

func TestPlannerOutputSchemaIsValidJSON(t *testing.T) {
	schema := plannerOutputSchema()
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

func TestGeneratePlanUsesRuntimeProviderRequest(t *testing.T) {
	cfg := &config.Config{
		Claude: config.ClaudeConfig{
			Effort: "high",
			Models: config.ModelConfig{
				Planner: "planner-model",
			},
		},
	}
	generator := newGeneratorWithProvider(cfg, "/repo", generatorFakeProvider{})

	var captured agentruntime.RunRequest
	generator.runText = func(_ context.Context, provider agentruntime.Provider, req agentruntime.RunRequest, onEvent func(agentruntime.Event)) (string, *cost.Usage, error) {
		captured = req
		if provider.Name() != "fake" {
			t.Fatalf("provider = %q, want fake", provider.Name())
		}
		if onEvent == nil {
			t.Fatal("expected provider raw event observer")
		}
		return `{"approach":"do it","candidateFiles":["src/Foo.tsx"],"newFiles":[],"deletedFiles":[],"validation":["yarn test"],"rollback":"revert","stopConditions":["stop"],"uncertainties":[]}`, &cost.Usage{InputTokens: 7}, nil
	}

	plan, usage, err := generator.GeneratePlan(context.Background(),
		triage.NormalizedTicket{TicketID: "ENG-7", Title: "Plan me"},
		triage.Classification{TicketID: "ENG-7", Category: triage.CategoryAIDefinite},
		nil,
	)
	if err != nil {
		t.Fatalf("GeneratePlan error: %v", err)
	}
	if plan.TicketID != "ENG-7" {
		t.Fatalf("TicketID = %q, want ENG-7", plan.TicketID)
	}
	if usage == nil || usage.InputTokens != 7 {
		t.Fatalf("usage = %#v, want input tokens", usage)
	}
	if captured.Role != agentruntime.RolePlanner {
		t.Fatalf("Role = %q, want %q", captured.Role, agentruntime.RolePlanner)
	}
	if captured.Profile != "triage-planner" {
		t.Fatalf("Profile = %q, want triage-planner", captured.Profile)
	}
	if captured.Model != "planner-model" {
		t.Fatalf("Model = %q, want planner-model", captured.Model)
	}
	if captured.WorkDir != "/repo" {
		t.Fatalf("WorkDir = %q, want /repo", captured.WorkDir)
	}
	if captured.ApprovalPolicy != agentruntime.ApprovalFullAuto {
		t.Fatalf("ApprovalPolicy = %q, want %q", captured.ApprovalPolicy, agentruntime.ApprovalFullAuto)
	}
	if captured.Reasoning == nil || captured.Reasoning.Effort != "high" {
		t.Fatalf("Reasoning = %#v, want high effort", captured.Reasoning)
	}
	if len(captured.Tools) != 3 {
		t.Fatalf("Tools = %#v, want 3 planner tools", captured.Tools)
	}
	if captured.OutputSchema == nil || captured.OutputSchema.Name != "triage_ticket_plan" {
		t.Fatalf("OutputSchema = %#v, want planner schema", captured.OutputSchema)
	}
	if captured.Metadata["phaseId"] != "triage-planner" {
		t.Fatalf("phaseId = %q, want triage-planner", captured.Metadata["phaseId"])
	}
}
