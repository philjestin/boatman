package team

import (
	"github.com/philjestin/boatman-ecosystem/harness/cost"
)

// Strategy determines how selected agents are executed.
type Strategy int

const (
	// Sequential executes agents one after another.
	Sequential Strategy = iota
	// Parallel executes agents concurrently.
	Parallel
)

// ErrorPolicy determines how errors from individual agents are handled.
type ErrorPolicy int

const (
	// FailFast aborts execution on the first agent error.
	FailFast ErrorPolicy = iota
	// CollectErrors continues execution and collects all errors.
	CollectErrors
)

// Option configures a Team.
type Option func(*Team)

// WithAgents adds agents to the team.
func WithAgents(agents ...Agent) Option {
	return func(t *Team) { t.agents = append(t.agents, agents...) }
}

// WithRouter sets the routing strategy for agent selection.
func WithRouter(r Router) Option {
	return func(t *Team) { t.router = r }
}

// WithStrategy sets the execution strategy (Sequential or Parallel).
func WithStrategy(s Strategy) Option {
	return func(t *Team) { t.strategy = s }
}

// WithAggregator sets the result aggregation strategy.
func WithAggregator(a Aggregator) Option {
	return func(t *Team) { t.aggregator = a }
}

// WithTeamObserver sets the observer for team lifecycle events.
func WithTeamObserver(o TeamObserver) Option {
	return func(t *Team) { t.observer = o }
}

// WithTeamGuard sets the guard for agent execution gating.
func WithTeamGuard(g TeamGuard) Option {
	return func(t *Team) { t.guard = g }
}

// WithCostTracker sets the cost tracker for usage accounting.
func WithCostTracker(ct *cost.Tracker) Option {
	return func(t *Team) { t.costTracker = ct }
}

// WithErrorPolicy sets how agent errors are handled.
func WithErrorPolicy(p ErrorPolicy) Option {
	return func(t *Team) { t.errorPolicy = p }
}

// WithDescription sets the team's description, used for LLM-based routing
// when the team is nested as an agent.
func WithDescription(desc string) Option {
	return func(t *Team) { t.description = desc }
}
