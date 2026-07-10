package runprep

import (
	"context"

	agentruntime "github.com/philjestin/boatman-ecosystem/shared/agentruntime"
)

// NewInitialEventsProvider wraps provider so preparation events are emitted
// immediately after the provider's first lifecycle event. That keeps run.started
// first while still making memory/configuration work visible in the stream.
func NewInitialEventsProvider(provider agentruntime.Provider, events []agentruntime.Event) agentruntime.Provider {
	if provider == nil || len(events) == 0 {
		return provider
	}
	copied := make([]agentruntime.Event, len(events))
	copy(copied, events)
	return initialEventsProvider{provider: provider, events: copied}
}

type initialEventsProvider struct {
	provider agentruntime.Provider
	events   []agentruntime.Event
}

func (p initialEventsProvider) Name() string {
	return p.provider.Name()
}

func (p initialEventsProvider) Capabilities(ctx context.Context) (agentruntime.Capabilities, error) {
	return p.provider.Capabilities(ctx)
}

func (p initialEventsProvider) StartRun(ctx context.Context, req agentruntime.RunRequest) (agentruntime.EventStream, error) {
	stream, err := p.provider.StartRun(ctx, req)
	if err != nil {
		return nil, err
	}
	out := make(chan agentruntime.Event, 32)
	go func() {
		defer close(out)
		emittedInitial := false
		emitInitial := func() bool {
			if emittedInitial {
				return true
			}
			emittedInitial = true
			for _, event := range p.events {
				select {
				case <-ctx.Done():
					return false
				case out <- event:
				}
			}
			return true
		}
		for event := range stream {
			select {
			case <-ctx.Done():
				return
			case out <- event:
			}
			if !emitInitial() {
				return
			}
		}
		if !emittedInitial {
			emitInitial()
		}
	}()
	return out, nil
}

func (p initialEventsProvider) ResumeRun(ctx context.Context, runID string, input agentruntime.RunInput) (agentruntime.EventStream, error) {
	return p.provider.ResumeRun(ctx, runID, input)
}

func (p initialEventsProvider) CancelRun(ctx context.Context, runID string) error {
	return p.provider.CancelRun(ctx, runID)
}
