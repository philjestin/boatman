// Package plan implements Stage 4 of the triage pipeline: generating and
// validating execution plans for AI_DEFINITE tickets before they advance
// to autonomous execution.
package plan

import "github.com/philjestin/boatmanmode/internal/cost"

// TicketPlan is the structured execution plan for a single ticket.
type TicketPlan struct {
	TicketID       string   `json:"ticketId"`
	Approach       string   `json:"approach"`
	CandidateFiles []string `json:"candidateFiles"`
	NewFiles       []string `json:"newFiles"`
	DeletedFiles   []string `json:"deletedFiles"`
	Validation     []string `json:"validation"`
	Rollback       string   `json:"rollback"`
	StopConditions []string `json:"stopConditions"`
	Uncertainties  []string `json:"uncertainties"`
}

// PlanGateResult records whether a single validation gate passed or failed.
type PlanGateResult struct {
	Gate   string `json:"gate"`
	Passed bool   `json:"passed"`
	Reason string `json:"reason,omitempty"`
}

// PlanValidation holds the results of all gate checks for a plan.
type PlanValidation struct {
	Passed          bool             `json:"passed"`
	GateResults     []PlanGateResult `json:"gateResults"`
	ValidatedFiles  []string         `json:"validatedFiles"`
	MissingFiles    []string         `json:"missingFiles"`
	OutOfScopeFiles []string         `json:"outOfScopeFiles"`
}

// PlanResult holds the output of plan generation + validation for one ticket.
type PlanResult struct {
	TicketID   string          `json:"ticketId"`
	Plan       *TicketPlan     `json:"plan,omitempty"`
	Validation *PlanValidation `json:"validation,omitempty"`
	Usage      *cost.Usage     `json:"usage,omitempty"`
	Error      string          `json:"error,omitempty"`
}

// PlanStats is the aggregate output of a plan generation batch.
type PlanStats struct {
	Total           int     `json:"total"`
	Passed          int     `json:"passed"`
	Failed          int     `json:"failed"`
	TotalTokensUsed int     `json:"totalTokensUsed"`
	TotalCostUSD    float64 `json:"totalCostUsd"`
}
