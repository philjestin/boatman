package team

import (
	"context"

	"github.com/philjestin/boatman-ecosystem/harness/cost"
)

// Handler executes work assigned to an agent.
type Handler interface {
	Handle(ctx context.Context, task *Task) (*Result, error)
}

// HandlerFunc adapts a plain function to the Handler interface.
type HandlerFunc func(ctx context.Context, task *Task) (*Result, error)

// Handle calls f(ctx, task).
func (f HandlerFunc) Handle(ctx context.Context, task *Task) (*Result, error) {
	return f(ctx, task)
}

// Agent is a named, described unit within a team.
// Description is critical for LLM-based routing.
type Agent struct {
	Name        string
	Description string
	Handler     Handler
}

// NewAgent creates an Agent with the given name, description, and handler.
func NewAgent(name, description string, h Handler) Agent {
	return Agent{
		Name:        name,
		Description: description,
		Handler:     h,
	}
}

// Task is the input to an agent.
type Task struct {
	ID          string
	Description string
	Context     string         // Upstream context (from handoff, parent).
	Input       map[string]any // Arbitrary structured data.
	Constraints []string
	ParentID    string // Links to parent task when decomposed.
}

// Result is the output from an agent.
type Result struct {
	AgentName    string
	Output       string
	Data         map[string]any // Structured output (e.g., "passed", "issues", "steps").
	Usage        cost.Usage
	FilesChanged []string
	Diff         string
	Children     []Result // Sub-team results (for nesting).
	Error        error
}
