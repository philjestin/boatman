// Package triage implements the ADR-004 backlog triage pipeline.
// Phase 1: Ingest, score, classify, and cluster Linear tickets
// to determine AI execution suitability.
package triage

import (
	"encoding/json"
	"time"
)

// Category represents the AI solvability classification of a ticket.
type Category string

const (
	CategoryAIDefinite          Category = "AI_DEFINITE"
	CategoryAILikely            Category = "AI_LIKELY"
	CategoryHumanReviewRequired Category = "HUMAN_REVIEW_REQUIRED"
	CategoryHumanOnly           Category = "HUMAN_ONLY"
)

// Stage represents a pipeline stage for decision logging.
type Stage string

const (
	StageIngest  Stage = "ingest"
	StageScore   Stage = "score"
	StageCluster Stage = "cluster"
	StagePlan    Stage = "plan"
)

// NormalizedTicket is the Stage 1 output — a Linear ticket with extracted signals
// and a staleness TTL.
type NormalizedTicket struct {
	TicketID    string    `json:"ticketId"`
	Title       string    `json:"title"`
	Summary     string    `json:"summary"`
	Description string    `json:"description"`
	IngestedAt  time.Time `json:"ingestedAt"`
	StaleAfter  time.Time `json:"staleAfter"`
	Signals     Signals   `json:"signals"`
}

// Signals are structured data extracted from the ticket during ingestion.
type Signals struct {
	MentionsFiles              []string  `json:"mentionsFiles"`
	Domains                    []string  `json:"domains"`
	Dependencies               []string  `json:"dependencies"`
	Labels                     []string  `json:"labels"`
	AcceptanceCriteriaPresent  bool      `json:"acceptanceCriteriaPresent"`
	AcceptanceCriteriaExplicit bool      `json:"acceptanceCriteriaExplicit"`
	HasDesignSpec              bool      `json:"hasDesignSpec"`
	CommentCount               int       `json:"commentCount"`
	LastUpdated                time.Time `json:"lastUpdated"`
	TeamKey                    string    `json:"teamKey"`
	ProjectName                string    `json:"projectName"`
	Estimate                   *float64  `json:"estimate,omitempty"`
}

// RubricScores holds the 7-dimension rubric evaluation from the scorer.
// Positive dimensions (higher is better): clarity, codeLocality, patternMatch, validationStrength.
// Negative dimensions (higher is worse): dependencyRisk, productAmbiguity, blastRadius.
// All scores are 0-5.
type RubricScores struct {
	Clarity            int `json:"clarity"`
	CodeLocality       int `json:"codeLocality"`
	PatternMatch       int `json:"patternMatch"`
	ValidationStrength int `json:"validationStrength"`
	DependencyRisk     int `json:"dependencyRisk"`
	ProductAmbiguity   int `json:"productAmbiguity"`
	BlastRadius        int `json:"blastRadius"`
}

// ScorerResponse is the JSON structure returned by Claude when scoring a ticket.
type ScorerResponse struct {
	RubricScores
	UncertainAxes []string `json:"uncertainAxes"`
	Reasons       []string `json:"reasons"`
}

// GateResult records whether a single classification gate passed or failed.
type GateResult struct {
	Gate   string `json:"gate"`
	Passed bool   `json:"passed"`
	Reason string `json:"reason,omitempty"`
}

// Classification is the Stage 2 output — rubric scores, category assignment,
// and the full gate/hard-stop audit trail.
type Classification struct {
	TicketID      string       `json:"ticketId"`
	Category      Category     `json:"category"`
	Rubric        RubricScores `json:"rubric"`
	UncertainAxes []string     `json:"uncertainAxes"`
	Reasons       []string     `json:"reasons"`
	HardStops     []string     `json:"hardStops"`
	GateResults   []GateResult `json:"gateResults"`
}

// Cluster groups related tickets that share code areas or patterns.
type Cluster struct {
	ClusterID string   `json:"clusterId"`
	Rationale string   `json:"rationale"`
	TicketIDs []string `json:"tickets"`
	RepoAreas []string `json:"repoAreas"`
}

// CostCeiling defines token and time budgets for execution (used in future phases).
type CostCeiling struct {
	MaxTokensPerTicket       int `json:"maxTokensPerTicket"`
	MaxAgentMinutesPerTicket int `json:"maxAgentMinutesPerTicket"`
}

// ContextDoc is the Stage 3 output — a shared context artifact for a cluster
// of related tickets.
type ContextDoc struct {
	ClusterID      string      `json:"clusterId"`
	Rationale      string      `json:"rationale"`
	TicketIDs      []string    `json:"tickets"`
	RepoAreas      []string    `json:"repoAreas"`
	KnownPatterns  []string    `json:"knownPatterns"`
	ValidationPlan []string    `json:"validationPlan"`
	Risks          []string    `json:"risks"`
	CostCeiling    CostCeiling `json:"costCeiling"`
}

// DecisionLogEntry records a single pipeline decision for audit and tuning.
type DecisionLogEntry struct {
	TicketID   string          `json:"ticketId"`
	Stage      Stage           `json:"stage"`
	Verdict    string          `json:"verdict"`
	Agent      string          `json:"agent"`
	Rationale  string          `json:"rationale"`
	Timestamp  time.Time       `json:"timestamp"`
	TokensUsed int             `json:"tokensUsed,omitempty"`
	CostUSD    float64         `json:"costUsd,omitempty"`
	Model      string          `json:"model,omitempty"`
	Details    json.RawMessage `json:"details,omitempty"`
}

// TriageResult is the aggregate output of a complete pipeline run.
type TriageResult struct {
	Tickets         []NormalizedTicket `json:"tickets"`
	Classifications []Classification  `json:"classifications"`
	Clusters        []Cluster         `json:"clusters"`
	ContextDocs     []ContextDoc      `json:"contextDocs"`
	Stats           TriageStats       `json:"stats"`
	// Plans holds Stage 4 plan results (json.RawMessage to avoid circular import with plan package).
	Plans     json.RawMessage `json:"plans,omitempty"`
	PlanStats json.RawMessage `json:"planStats,omitempty"`
}

// TriageStats summarizes a triage run.
type TriageStats struct {
	TotalTickets      int     `json:"totalTickets"`
	AIDefiniteCount   int     `json:"aiDefiniteCount"`
	AILikelyCount     int     `json:"aiLikelyCount"`
	HumanReviewCount  int     `json:"humanReviewCount"`
	HumanOnlyCount    int     `json:"humanOnlyCount"`
	ClusterCount      int     `json:"clusterCount"`
	TotalTokensUsed   int     `json:"totalTokensUsed"`
	TotalCostUSD      float64 `json:"totalCostUsd"`
}
