package plan

import (
	"strings"
	"testing"

	"github.com/philjestin/boatmanmode/internal/triage"
)

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
