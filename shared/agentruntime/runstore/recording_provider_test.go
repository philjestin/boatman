package runstore

import (
	"context"
	"errors"
	"testing"

	agentruntime "github.com/philjestin/boatman-ecosystem/shared/agentruntime"
)

type recordingProviderFake struct {
	events []agentruntime.Event
	err    error
}

func (p recordingProviderFake) Name() string {
	return "fake"
}

func (p recordingProviderFake) Capabilities(context.Context) (agentruntime.Capabilities, error) {
	return agentruntime.Capabilities{Provider: "fake"}, nil
}

func (p recordingProviderFake) StartRun(context.Context, agentruntime.RunRequest) (agentruntime.EventStream, error) {
	if p.err != nil {
		return nil, p.err
	}
	ch := make(chan agentruntime.Event, len(p.events))
	for _, event := range p.events {
		ch <- event
	}
	close(ch)
	return ch, nil
}

func (p recordingProviderFake) ResumeRun(context.Context, string, agentruntime.RunInput) (agentruntime.EventStream, error) {
	return nil, nil
}

func (p recordingProviderFake) CancelRun(context.Context, string) error {
	return nil
}

func TestRecordingProviderRecordsEvents(t *testing.T) {
	store := NewFileStore(t.TempDir())
	message := agentruntime.NewEvent(agentruntime.EventMessageCompleted)
	message.Message = "hello"
	completed := agentruntime.NewEvent(agentruntime.EventRunCompleted)
	completed.Status = agentruntime.StatusSucceeded

	provider := NewRecordingProvider(recordingProviderFake{
		events: []agentruntime.Event{message, completed},
	}, store)
	stream, err := provider.StartRun(context.Background(), agentruntime.RunRequest{
		RunID: "recorded",
		Role:  agentruntime.RoleChat,
	})
	if err != nil {
		t.Fatalf("StartRun error: %v", err)
	}
	for range stream {
	}

	metadata, events, err := store.LoadRun(context.Background(), "recorded")
	if err != nil {
		t.Fatalf("LoadRun error: %v", err)
	}
	if metadata.Status != agentruntime.StatusSucceeded || metadata.EventCount != 2 {
		t.Fatalf("metadata = %#v, want succeeded with two events", metadata)
	}
	if len(events) != 2 || events[0].RunID != "recorded" || events[0].Provider != "fake" {
		t.Fatalf("events = %#v, want run/provider filled", events)
	}
}

func TestRecordingProviderRecordsStartFailure(t *testing.T) {
	store := NewFileStore(t.TempDir())
	provider := NewRecordingProvider(recordingProviderFake{err: errors.New("boom")}, store)

	_, err := provider.StartRun(context.Background(), agentruntime.RunRequest{RunID: "failed"})
	if err == nil || err.Error() != "boom" {
		t.Fatalf("err = %v, want boom", err)
	}

	metadata, events, err := store.LoadRun(context.Background(), "failed")
	if err != nil {
		t.Fatalf("LoadRun error: %v", err)
	}
	if metadata.Status != agentruntime.StatusFailed {
		t.Fatalf("status = %q, want failed", metadata.Status)
	}
	if len(events) != 1 || events[0].Type != agentruntime.EventRunFailed {
		t.Fatalf("events = %#v, want run.failed", events)
	}
}
