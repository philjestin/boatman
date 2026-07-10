package runstore

import (
	"context"
	"fmt"
	"time"

	agentruntime "github.com/philjestin/boatman-ecosystem/shared/agentruntime"
)

// RecordingProvider wraps a provider and tees its runtime events into a Store.
type RecordingProvider struct {
	Provider agentruntime.Provider
	Store    Store
}

// NewRecordingProvider returns provider wrapped with event recording. Nil inputs
// return the original provider so call sites can compose it conditionally.
func NewRecordingProvider(provider agentruntime.Provider, store Store) agentruntime.Provider {
	if provider == nil || store == nil {
		return provider
	}
	return RecordingProvider{Provider: provider, Store: store}
}

func (p RecordingProvider) Name() string {
	return p.Provider.Name()
}

func (p RecordingProvider) Capabilities(ctx context.Context) (agentruntime.Capabilities, error) {
	return p.Provider.Capabilities(ctx)
}

func (p RecordingProvider) StartRun(ctx context.Context, req agentruntime.RunRequest) (agentruntime.EventStream, error) {
	if req.RunID == "" {
		req.RunID = fmt.Sprintf("%s-%d", p.Provider.Name(), time.Now().UnixNano())
	}
	if req.Provider == "" {
		req.Provider = p.Provider.Name()
	}
	if err := p.Store.StartRun(ctx, req); err != nil {
		return nil, fmt.Errorf("start runtime run store: %w", err)
	}
	stream, err := p.Provider.StartRun(ctx, req)
	if err != nil {
		failed := agentruntime.NewEvent(agentruntime.EventRunFailed)
		failed.RunID = req.RunID
		failed.Provider = req.Provider
		failed.Model = req.Model
		failed.Role = req.Role
		failed.Status = agentruntime.StatusFailed
		failed.Message = err.Error()
		_ = p.Store.Append(ctx, failed)
		return nil, err
	}

	out := make(chan agentruntime.Event, 32)
	go func() {
		defer close(out)
		for event := range stream {
			if event.RunID == "" {
				event.RunID = req.RunID
			}
			if event.Provider == "" {
				event.Provider = p.Provider.Name()
			}
			if err := p.Store.Append(ctx, event); err != nil {
				failed := agentruntime.NewEvent(agentruntime.EventRunFailed)
				failed.RunID = event.RunID
				failed.Provider = event.Provider
				failed.Model = event.Model
				failed.Role = event.Role
				failed.Status = agentruntime.StatusFailed
				failed.Message = fmt.Sprintf("record runtime event: %v", err)
				select {
				case <-ctx.Done():
				case out <- failed:
				}
				return
			}
			select {
			case <-ctx.Done():
				return
			case out <- event:
			}
		}
	}()
	return out, nil
}

func (p RecordingProvider) ResumeRun(ctx context.Context, runID string, input agentruntime.RunInput) (agentruntime.EventStream, error) {
	stream, err := p.Provider.ResumeRun(ctx, runID, input)
	if err != nil {
		return nil, err
	}
	out := make(chan agentruntime.Event, 32)
	go func() {
		defer close(out)
		for event := range stream {
			if event.RunID == "" {
				event.RunID = runID
			}
			if event.Provider == "" {
				event.Provider = p.Provider.Name()
			}
			if err := p.Store.Append(ctx, event); err != nil {
				return
			}
			select {
			case <-ctx.Done():
				return
			case out <- event:
			}
		}
	}()
	return out, nil
}

func (p RecordingProvider) CancelRun(ctx context.Context, runID string) error {
	return p.Provider.CancelRun(ctx, runID)
}
