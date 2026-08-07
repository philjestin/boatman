package frontendqueue

import "testing"

func TestClassifyTicketCopyOnlyAutoPR(t *testing.T) {
	estimate := 1.0
	plan := ClassifyTicket(Ticket{
		ID:          "OPX-933",
		Title:       "Add Icon to Pause Menu Item",
		Description: "Figma: https://www.figma.com/design/example?node-id=7700-11168",
		Labels:      []string{"Projects"},
		Estimate:    &estimate,
		StatusType:  "backlog",
	}, PlanOptions{})

	if plan.Type != TicketStyleToken {
		t.Fatalf("Type = %s, want %s", plan.Type, TicketStyleToken)
	}
	if plan.Policy != PolicyAutoPR {
		t.Fatalf("Policy = %s, want %s", plan.Policy, PolicyAutoPR)
	}
	if plan.PassKCandidates != 3 {
		t.Fatalf("PassKCandidates = %d, want 3", plan.PassKCandidates)
	}
	if len(plan.FigmaRefs) != 1 {
		t.Fatalf("FigmaRefs len = %d, want 1", len(plan.FigmaRefs))
	}
}

func TestClassifyTicketComponentRestyleDraftPR(t *testing.T) {
	plan := ClassifyTicket(Ticket{
		ID:          "OPX-963",
		Title:       "Centre Table v2 pagination to the table",
		Description: "`src/components/hai-ui/TableV2.tsx` L174 currently uses justify-between.\n\n## Acceptance criteria\n- [ ] Pagination is centred to the table",
		Labels:      []string{"UI Fix", "Rosettafy"},
		StatusType:  "unstarted",
	}, PlanOptions{})

	if plan.Type != TicketComponentRestyle {
		t.Fatalf("Type = %s, want %s", plan.Type, TicketComponentRestyle)
	}
	if plan.Policy != PolicyDraftPR {
		t.Fatalf("Policy = %s, want %s", plan.Policy, PolicyDraftPR)
	}
	if plan.TargetKey != "src/components/hai-ui/tablev2.tsx" {
		t.Fatalf("TargetKey = %s, want src/components/hai-ui/tablev2.tsx", plan.TargetKey)
	}
	if len(plan.Validation) == 0 {
		t.Fatalf("expected validation gates")
	}
}

func TestClassifyTicketAmbiguousPlanOnly(t *testing.T) {
	plan := ClassifyTicket(Ticket{
		ID:          "OPX-876",
		Title:       "Reconsider Add step Modal",
		Description: "Modal or full-page? Rosetta-fy. Fix copy.",
		StatusType:  "backlog",
	}, PlanOptions{})

	if plan.Type != TicketAmbiguousDesign {
		t.Fatalf("Type = %s, want %s", plan.Type, TicketAmbiguousDesign)
	}
	if plan.Policy != PolicyPlanOnly {
		t.Fatalf("Policy = %s, want %s", plan.Policy, PolicyPlanOnly)
	}
}

func TestPlanTicketsBuildsConflictAwareBatches(t *testing.T) {
	tickets := []Ticket{
		{
			ID:          "OPX-963",
			Title:       "Centre Table v2 pagination to the table",
			Description: "`src/components/hai-ui/TableV2.tsx` L174",
			StatusType:  "unstarted",
		},
		{
			ID:          "OPX-964",
			Title:       "Restyle Table v2 empty state",
			Description: "`src/components/hai-ui/TableV2.tsx` L1531",
			StatusType:  "unstarted",
		},
		{
			ID:          "RLE-112",
			Title:       "Clean up Preview mode metadata",
			Description: "Update text and metadata copy.",
			StatusType:  "backlog",
		},
	}

	plan := PlanTickets(tickets, PlanOptions{MaxParallel: 3})

	if plan.Stats.TotalTickets != 3 {
		t.Fatalf("TotalTickets = %d, want 3", plan.Stats.TotalTickets)
	}
	if len(plan.Batches) != 2 {
		t.Fatalf("Batches len = %d, want 2", len(plan.Batches))
	}
	first := plan.Batches[0]
	if len(first.TicketIDs) != 2 {
		t.Fatalf("first batch ticket count = %d, want 2", len(first.TicketIDs))
	}
	if containsBoth(first.TicketIDs, "OPX-963", "OPX-964") {
		t.Fatalf("TableV2 tickets with same target should not share a batch: %#v", first.TicketIDs)
	}
}

func TestClassifyTicketCanceledBlocked(t *testing.T) {
	plan := ClassifyTicket(Ticket{
		ID:         "OPX-882",
		Title:      "Reduce Persona title to 15pt",
		StatusType: "canceled",
	}, PlanOptions{})

	if plan.Policy != PolicyBlocked {
		t.Fatalf("Policy = %s, want %s", plan.Policy, PolicyBlocked)
	}
	if len(plan.Blockers) == 0 {
		t.Fatalf("expected blocker")
	}
}

func containsBoth(values []string, left, right string) bool {
	hasLeft := false
	hasRight := false
	for _, value := range values {
		if value == left {
			hasLeft = true
		}
		if value == right {
			hasRight = true
		}
	}
	return hasLeft && hasRight
}
