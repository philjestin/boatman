package runner

import (
	"context"
	"time"
)

// Guard is consulted before each step execution. If AllowStep returns an error,
// the step is skipped and the error is propagated.
type Guard interface {
	AllowStep(ctx context.Context, step string, state *GuardState) error
}

// GuardState provides the guard with current run metrics for informed decisions.
type GuardState struct {
	Iterations   int
	ElapsedTime  time.Duration
	TotalCostUSD float64
	FilesChanged int
}

// NopGuard always allows every step.
type NopGuard struct{}

// AllowStep always returns nil, allowing the step to proceed.
func (NopGuard) AllowStep(_ context.Context, _ string, _ *GuardState) error { return nil }

// WithGuard attaches a Guard to the Runner.
func WithGuard(g Guard) Option {
	return func(r *Runner) { r.guard = g }
}
