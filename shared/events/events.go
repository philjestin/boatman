// Package events defines the event types shared between CLI and desktop.
// These events form the integration contract between the two components.
package events

// Event represents a structured event emitted during workflow execution.
// Events are emitted by the CLI and consumed by the desktop app.
type Event struct {
	Type        string         `json:"type"`
	ID          string         `json:"id,omitempty"`
	Name        string         `json:"name,omitempty"`
	Description string         `json:"description,omitempty"`
	Status      string         `json:"status,omitempty"`
	Message     string         `json:"message,omitempty"`
	Data        map[string]any `json:"data,omitempty"`
}

// Event type constants
const (
	// Agent lifecycle events
	EventAgentStarted   = "agent_started"
	EventAgentCompleted = "agent_completed"

	// Task events
	EventTaskCreated = "task_created"
	EventTaskUpdated = "task_updated"

	// Progress events
	EventProgress = "progress"
)

// Status constants
const (
	StatusSuccess = "success"
	StatusFailed  = "failed"
)

// AgentMetadata contains phase-specific data attached to agent completion events.
type AgentMetadata struct {
	// Diff contains the git diff of changes made
	Diff string `json:"diff,omitempty"`

	// Plan contains the planning analysis text
	Plan string `json:"plan,omitempty"`

	// Feedback contains code review feedback
	Feedback string `json:"feedback,omitempty"`

	// Issues contains review issues found
	Issues []ReviewIssue `json:"issues,omitempty"`

	// RefactorDiff contains the diff after refactoring
	RefactorDiff string `json:"refactor_diff,omitempty"`
}

// ReviewIssue represents a code review issue.
type ReviewIssue struct {
	File     string `json:"file"`
	Line     int    `json:"line,omitempty"`
	Severity string `json:"severity"` // "error", "warning", "info"
	Message  string `json:"message"`
	Code     string `json:"code,omitempty"` // Code snippet
}
