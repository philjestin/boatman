package runner

import (
	"context"
	"time"
)

// Observer provides structured observation of runner lifecycle events.
// It is a richer alternative to Hooks that receives typed context about each phase.
type Observer interface {
	OnRunStart(ctx context.Context, req *Request)
	OnRunComplete(ctx context.Context, result *Result)
	OnStepStart(ctx context.Context, step string)
	OnStepComplete(ctx context.Context, step string, duration time.Duration, err error)
}

// NopObserver is a no-op Observer that ignores all events.
type NopObserver struct{}

func (NopObserver) OnRunStart(_ context.Context, _ *Request)                            {}
func (NopObserver) OnRunComplete(_ context.Context, _ *Result)                           {}
func (NopObserver) OnStepStart(_ context.Context, _ string)                              {}
func (NopObserver) OnStepComplete(_ context.Context, _ string, _ time.Duration, _ error) {}

// WithObserver attaches an Observer to the Runner.
func WithObserver(o Observer) Option {
	return func(r *Runner) { r.observer = o }
}
