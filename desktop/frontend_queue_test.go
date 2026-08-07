package main

import (
	"strings"
	"testing"

	agentruntime "github.com/philjestin/boatman-ecosystem/shared/agentruntime"
	"github.com/philjestin/boatman-ecosystem/shared/agentruntime/frontendqueue"
)

func TestNormalizeLinearProjectInput(t *testing.T) {
	got := normalizeLinearProjectInput("https://linear.app/joinhandshake/project/design-operator-workspace-improvements-d40e69c1364b/issues")
	want := "design-operator-workspace-improvements-d40e69c1364b"
	if got != want {
		t.Fatalf("normalizeLinearProjectInput() = %q, want %q", got, want)
	}
}

func TestPlanFrontendTicketQueueWithProvidedTickets(t *testing.T) {
	app := NewApp()
	result, err := app.PlanFrontendTicketQueue(FrontendTicketQueueRequest{
		ProjectPath: "/workspace/app",
		MaxParallel: 2,
		Tickets: []frontendqueue.Ticket{
			{
				ID:          "OPX-963",
				Title:       "Centre Table v2 pagination to the table",
				Description: "`src/components/hai-ui/TableV2.tsx` L174",
				StatusType:  "unstarted",
			},
			{
				ID:         "RLE-112",
				Title:      "Clean up Preview mode metadata",
				StatusType: "backlog",
			},
		},
	})
	if err != nil {
		t.Fatalf("PlanFrontendTicketQueue() error = %v", err)
	}
	if result.Source != "request" {
		t.Fatalf("Source = %q, want request", result.Source)
	}
	if result.Plan.Stats.TotalTickets != 2 {
		t.Fatalf("TotalTickets = %d, want 2", result.Plan.Stats.TotalTickets)
	}
	if len(result.Plan.Batches) == 0 {
		t.Fatalf("expected at least one batch")
	}
}

func TestSelectFrontendQueueRunPlansUsesBatch(t *testing.T) {
	plan := frontendqueue.Plan{
		Tickets: []frontendqueue.TicketPlan{
			{
				Ticket: frontendqueue.Ticket{ID: "OPX-1", Title: "Copy tweak"},
				Policy: frontendqueue.PolicyAutoPR,
			},
			{
				Ticket: frontendqueue.Ticket{ID: "OPX-2", Title: "Needs decision"},
				Policy: frontendqueue.PolicyPlanOnly,
			},
		},
		Batches: []frontendqueue.Batch{
			{ID: "batch-01", TicketIDs: []string{"OPX-2", "OPX-1"}},
		},
	}

	input := normalizeFrontendQueueRunRequest(FrontendTicketQueueRunRequest{
		ProjectPath: "/workspace/app",
		Plan:        plan,
		BatchID:     "batch-01",
	})
	selected, batchID := selectFrontendQueueRunPlans(input)
	if batchID != "batch-01" {
		t.Fatalf("batchID = %q, want batch-01", batchID)
	}
	if len(selected) != 2 {
		t.Fatalf("selected len = %d, want 2", len(selected))
	}
	if selected[0].Ticket.ID != "OPX-1" || selected[1].Ticket.ID != "OPX-2" {
		t.Fatalf("selected preserves plan order = %#v", selected)
	}
	if !frontendQueuePolicyRunnable(selected[0].Policy) {
		t.Fatalf("auto-pr ticket should be runnable")
	}
	if frontendQueuePolicyRunnable(selected[1].Policy) {
		t.Fatalf("plan-only ticket should not be runnable")
	}
}

func TestFrontendQueueExecutionPromptIncludesAutomationContract(t *testing.T) {
	plan := frontendqueue.TicketPlan{
		Ticket:       frontendqueue.Ticket{ID: "OPX-933", Title: "Update warning icon"},
		Policy:       frontendqueue.PolicyAutoPR,
		TargetKey:    "component:badge",
		WorkerPrompt: "Implement frontend ticket OPX-933.",
		Validation: []frontendqueue.ValidationStep{
			frontendqueue.ValidationVisualStep("Capture before/after screenshots."),
		},
		PassKCandidates: 3,
	}
	input := normalizeFrontendQueueRunRequest(FrontendTicketQueueRunRequest{
		LinearProject: "design-operator-workspace-improvements-d40e69c1364b",
		BaseBranch:    "main",
	})
	models := agentruntime.ModelProfile{Plan: "opus", Implementation: "sonnet", Skills: "haiku"}
	branch := frontendQueueBranchName("frontend-queue-20260804-120000", plan)
	prompt := frontendQueueExecutionPrompt(plan, input, "frontend-queue-20260804-120000", branch, models)

	for _, want := range []string{
		"Start by using /plan",
		"fresh git worktree",
		"Chrome MCP or Playwright",
		"peer-review and lydia-code-review",
		"Pass@K: 3 candidates",
		"Planning model: opus",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("prompt missing %q:\n%s", want, prompt)
		}
	}
	if !strings.HasPrefix(branch, "frontend-queue/opx-933/") {
		t.Fatalf("branch = %q, want frontend-queue/opx-933 prefix", branch)
	}
}
