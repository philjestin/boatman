package plan

import (
	"fmt"

	"github.com/philjestin/boatmanmode/internal/events"
)

const (
	EventPlanStarted         = "plan_started"
	EventPlanTicketPlanning  = "plan_ticket_planning"
	EventPlanTicketPlanned   = "plan_ticket_planned"
	EventPlanTicketValidated = "plan_ticket_validated"
	EventPlanComplete        = "plan_complete"
	EventPlanError           = "plan_error"
)

// EmitPlanStarted emits a plan_started event.
func EmitPlanStarted(count int) {
	events.Emit(events.Event{
		Type:    EventPlanStarted,
		Message: fmt.Sprintf("Starting plan generation for %d tickets", count),
		Data: map[string]any{
			"ticketCount": count,
		},
	})
}

// EmitTicketPlanning emits a plan_ticket_planning event.
func EmitTicketPlanning(ticketID string, index, total int) {
	events.Emit(events.Event{
		Type:    EventPlanTicketPlanning,
		ID:      ticketID,
		Message: fmt.Sprintf("Planning %s (%d/%d)", ticketID, index+1, total),
		Data: map[string]any{
			"index": index,
			"total": total,
		},
	})
}

// EmitTicketPlanned emits a plan_ticket_planned event.
func EmitTicketPlanned(result PlanResult, index, total int) {
	data := map[string]any{
		"index": index,
		"total": total,
	}

	msg := fmt.Sprintf("Planned %s (%d/%d)", result.TicketID, index+1, total)
	status := ""

	if result.Error != "" {
		msg = fmt.Sprintf("FAILED planning %s (%d/%d): %s", result.TicketID, index+1, total, result.Error)
		status = "error"
		data["error"] = result.Error
	} else if result.Plan != nil {
		data["candidateFileCount"] = len(result.Plan.CandidateFiles)
		data["stopConditionCount"] = len(result.Plan.StopConditions)
	}

	if result.Usage != nil {
		data["inputTokens"] = result.Usage.InputTokens
		data["outputTokens"] = result.Usage.OutputTokens
	}

	events.Emit(events.Event{
		Type:    EventPlanTicketPlanned,
		ID:      result.TicketID,
		Status:  status,
		Message: msg,
		Data:    data,
	})
}

// EmitTicketValidated emits a plan_ticket_validated event.
func EmitTicketValidated(ticketID string, validation *PlanValidation) {
	status := "passed"
	msg := fmt.Sprintf("Plan for %s passed all gates", ticketID)
	if !validation.Passed {
		status = "failed"
		// Find first failing gate for message
		for _, g := range validation.GateResults {
			if !g.Passed {
				msg = fmt.Sprintf("Plan for %s failed gate: %s — %s", ticketID, g.Gate, g.Reason)
				break
			}
		}
	}

	data := map[string]any{
		"passed":          validation.Passed,
		"validatedFiles":  len(validation.ValidatedFiles),
		"missingFiles":    len(validation.MissingFiles),
		"outOfScopeFiles": len(validation.OutOfScopeFiles),
	}

	events.Emit(events.Event{
		Type:    EventPlanTicketValidated,
		ID:      ticketID,
		Status:  status,
		Message: msg,
		Data:    data,
	})
}

// EmitPlanComplete emits a plan_complete event with plan results and stats.
func EmitPlanComplete(results []PlanResult, stats PlanStats) {
	events.Emit(events.Event{
		Type:    EventPlanComplete,
		Message: fmt.Sprintf("Plan generation complete: %d passed, %d failed out of %d", stats.Passed, stats.Failed, stats.Total),
		Data: map[string]any{
			"results": results,
			"stats":   stats,
		},
	})
}

// EmitPlanError emits a plan_error event.
func EmitPlanError(err error) {
	events.Emit(events.Event{
		Type:    EventPlanError,
		Status:  "error",
		Message: fmt.Sprintf("Plan generation error: %v", err),
		Data: map[string]any{
			"error": err.Error(),
		},
	})
}
