package triage

import (
	"encoding/json"
	"fmt"

	"github.com/philjestin/boatmanmode/internal/events"
)

// Triage event types emitted during pipeline execution.
const (
	EventTriageStarted        = "triage_started"
	EventTriageFetchComplete  = "triage_fetch_complete"
	EventTriageScoringStarted = "triage_scoring_started"
	EventTriageTicketScoring  = "triage_ticket_scoring"
	EventTriageTicketScored   = "triage_ticket_scored"
	EventTriageScoringComplete = "triage_scoring_complete"
	EventTriageClassifying    = "triage_classifying"
	EventTriageClustering     = "triage_clustering"
	EventTriageComplete       = "triage_complete"
	EventTriageError          = "triage_error"
)

// emitTriageStarted emits an event when the triage pipeline begins.
func emitTriageStarted(ticketCount int, teams []string) {
	events.Emit(events.Event{
		Type:    EventTriageStarted,
		Message: fmt.Sprintf("Starting triage for %d tickets", ticketCount),
		Data: map[string]any{
			"teams": teams,
		},
	})
}

// emitFetchComplete emits an event when ticket fetching finishes.
func emitFetchComplete(count int) {
	events.Emit(events.Event{
		Type:    EventTriageFetchComplete,
		Message: fmt.Sprintf("Fetched %d tickets from Linear", count),
		Data: map[string]any{
			"ticketCount": count,
		},
	})
}

// emitScoringStarted emits an event when Claude scoring begins.
func emitScoringStarted(count, concurrency int) {
	events.Emit(events.Event{
		Type:    EventTriageScoringStarted,
		Message: fmt.Sprintf("Scoring %d tickets (concurrency: %d)", count, concurrency),
		Data: map[string]any{
			"ticketCount": count,
			"concurrency": concurrency,
		},
	})
}

// emitTicketScoring emits an event when a ticket begins scoring.
func emitTicketScoring(ticketID, title string, index, total int) {
	events.Emit(events.Event{
		Type:    EventTriageTicketScoring,
		ID:      ticketID,
		Name:    title,
		Message: fmt.Sprintf("Scoring %s (%d/%d)", ticketID, index+1, total),
		Data: map[string]any{
			"index": index,
			"total": total,
		},
	})
}

// emitTicketScored emits an event when a ticket finishes scoring.
func emitTicketScored(st ScoredTicket, index, total int) {
	data := map[string]any{
		"ticketID": st.Ticket.TicketID,
		"index":    index,
		"total":    total,
	}

	status := "scored"
	msg := fmt.Sprintf("Scored %s (%d/%d)", st.Ticket.TicketID, index+1, total)

	if st.Err != nil {
		data["error"] = st.Err.Error()
		status = "error"
		msg = fmt.Sprintf("FAILED %s (%d/%d): %s", st.Ticket.TicketID, index+1, total, st.Err.Error())
	} else {
		data["clarity"] = st.Response.Clarity
		data["codeLocality"] = st.Response.CodeLocality
		data["patternMatch"] = st.Response.PatternMatch
		data["validationStrength"] = st.Response.ValidationStrength
		data["dependencyRisk"] = st.Response.DependencyRisk
		data["productAmbiguity"] = st.Response.ProductAmbiguity
		data["blastRadius"] = st.Response.BlastRadius
	}

	events.Emit(events.Event{
		Type:    EventTriageTicketScored,
		ID:      st.Ticket.TicketID,
		Name:    st.Ticket.Title,
		Status:  status,
		Message: msg,
		Data:    data,
	})
}

// emitScoringComplete emits an event when all scoring finishes.
func emitScoringComplete(scored, failed int) {
	events.Emit(events.Event{
		Type:    EventTriageScoringComplete,
		Message: fmt.Sprintf("Scoring complete: %d scored, %d failed", scored, failed),
		Data: map[string]any{
			"scored": scored,
			"failed": failed,
		},
	})
}

// emitClassifying emits an event when classification begins.
func emitClassifying(count int) {
	events.Emit(events.Event{
		Type:    EventTriageClassifying,
		Message: fmt.Sprintf("Classifying %d tickets", count),
		Data: map[string]any{
			"ticketCount": count,
		},
	})
}

// emitClustering emits an event when clustering begins.
func emitClustering(count int) {
	events.Emit(events.Event{
		Type:    EventTriageClustering,
		Message: fmt.Sprintf("Clustering %d tickets", count),
		Data: map[string]any{
			"ticketCount": count,
		},
	})
}

// emitTriageComplete emits an event with the full triage result.
func emitTriageComplete(result *TriageResult) {
	resultJSON, _ := json.Marshal(result)
	events.Emit(events.Event{
		Type:    EventTriageComplete,
		Status:  "completed",
		Message: fmt.Sprintf("Triage complete: %d tickets", result.Stats.TotalTickets),
		Data: map[string]any{
			"result": json.RawMessage(resultJSON),
			"stats":  result.Stats,
		},
	})
}

// emitTriageError emits an error event.
func emitTriageError(err error) {
	events.Emit(events.Event{
		Type:    EventTriageError,
		Status:  "error",
		Message: err.Error(),
	})
}
