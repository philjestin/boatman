package runprep

import (
	"context"
	"testing"

	agentruntime "github.com/philjestin/boatman-ecosystem/shared/agentruntime"
)

type fakeProvider struct {
	events []agentruntime.Event
}

func (p fakeProvider) Name() string { return "fake" }
func (p fakeProvider) Capabilities(context.Context) (agentruntime.Capabilities, error) {
	return agentruntime.Capabilities{Provider: "fake"}, nil
}
func (p fakeProvider) StartRun(context.Context, agentruntime.RunRequest) (agentruntime.EventStream, error) {
	ch := make(chan agentruntime.Event, len(p.events))
	for _, event := range p.events {
		ch <- event
	}
	close(ch)
	return ch, nil
}
func (p fakeProvider) ResumeRun(context.Context, string, agentruntime.RunInput) (agentruntime.EventStream, error) {
	return nil, nil
}
func (p fakeProvider) CancelRun(context.Context, string) error { return nil }

func TestInitialEventsProviderEmitsAfterFirstEvent(t *testing.T) {
	started := agentruntime.NewEvent(agentruntime.EventRunStarted)
	memory := agentruntime.NewEvent(agentruntime.EventMemoryLoaded)
	completed := agentruntime.NewEvent(agentruntime.EventRunCompleted)

	provider := NewInitialEventsProvider(fakeProvider{
		events: []agentruntime.Event{started, completed},
	}, []agentruntime.Event{memory})

	stream, err := provider.StartRun(context.Background(), agentruntime.RunRequest{})
	if err != nil {
		t.Fatalf("StartRun error: %v", err)
	}
	var got []agentruntime.EventType
	for event := range stream {
		got = append(got, event.Type)
	}
	want := []agentruntime.EventType{
		agentruntime.EventRunStarted,
		agentruntime.EventMemoryLoaded,
		agentruntime.EventRunCompleted,
	}
	if len(got) != len(want) {
		t.Fatalf("events = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("events = %v, want %v", got, want)
		}
	}
}
