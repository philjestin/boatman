package providers

import (
	"context"
	"errors"
	"testing"

	agentruntime "github.com/philjestin/boatman-ecosystem/shared/agentruntime"
	"github.com/philjestin/boatman-ecosystem/shared/agentruntime/runstore"
)

type streamProvider struct {
	events []agentruntime.Event
	err    error
}

func (p streamProvider) Name() string {
	return "stream"
}

func (p streamProvider) Capabilities(context.Context) (agentruntime.Capabilities, error) {
	return agentruntime.Capabilities{}, nil
}

func (p streamProvider) StartRun(context.Context, agentruntime.RunRequest) (agentruntime.EventStream, error) {
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

func (p streamProvider) ResumeRun(context.Context, string, agentruntime.RunInput) (agentruntime.EventStream, error) {
	return nil, nil
}

func (p streamProvider) CancelRun(context.Context, string) error {
	return nil
}

func TestRunTextCollectsResponseAndUsage(t *testing.T) {
	message := agentruntime.NewEvent(agentruntime.EventMessageCompleted)
	message.Message = "hello"
	usageEvent := agentruntime.NewEvent(agentruntime.EventUsageUpdated)
	usageEvent.Usage = &agentruntime.Usage{InputTokens: 4, OutputTokens: 2, TotalCostUSD: 0.01}

	response, usage, err := RunText(context.Background(), streamProvider{
		events: []agentruntime.Event{message, usageEvent},
	}, agentruntime.RunRequest{})
	if err != nil {
		t.Fatalf("RunText error: %v", err)
	}
	if response != "hello" {
		t.Fatalf("response = %q, want hello", response)
	}
	if usage == nil || usage.InputTokens != 4 || usage.OutputTokens != 2 {
		t.Fatalf("usage = %#v, want converted tokens", usage)
	}
}

func TestRunTextReturnsProviderStartError(t *testing.T) {
	_, _, err := RunText(context.Background(), streamProvider{err: errors.New("start failed")}, agentruntime.RunRequest{})
	if err == nil || err.Error() != "start failed" {
		t.Fatalf("err = %v, want start failed", err)
	}
}

func TestRunTextReturnsRunFailedEvent(t *testing.T) {
	failed := agentruntime.NewEvent(agentruntime.EventRunFailed)
	failed.Message = "model exploded"

	_, _, err := RunText(context.Background(), streamProvider{
		events: []agentruntime.Event{failed},
	}, agentruntime.RunRequest{})
	if err == nil || err.Error() != "model exploded" {
		t.Fatalf("err = %v, want model exploded", err)
	}
}

func TestRunTextWithEventsObservesStream(t *testing.T) {
	message := agentruntime.NewEvent(agentruntime.EventMessageCompleted)
	message.Message = "hello"
	raw := agentruntime.NewEvent(agentruntime.EventProviderRaw)
	raw.Message = "raw"

	var seen []agentruntime.EventType
	response, _, err := RunTextWithEvents(context.Background(), streamProvider{
		events: []agentruntime.Event{raw, message},
	}, agentruntime.RunRequest{}, func(event agentruntime.Event) {
		seen = append(seen, event.Type)
	})
	if err != nil {
		t.Fatalf("RunTextWithEvents error: %v", err)
	}
	if response != "hello" {
		t.Fatalf("response = %q, want hello", response)
	}
	if len(seen) != 2 || seen[0] != agentruntime.EventProviderRaw || seen[1] != agentruntime.EventMessageCompleted {
		t.Fatalf("seen = %v, want provider raw then message completed", seen)
	}
}

func TestRunTextRecordsRuntimeStoreWhenConfigured(t *testing.T) {
	storeDir := t.TempDir()
	message := agentruntime.NewEvent(agentruntime.EventMessageCompleted)
	message.Message = "stored"
	completed := agentruntime.NewEvent(agentruntime.EventRunCompleted)
	completed.Status = agentruntime.StatusSucceeded

	response, _, err := RunText(context.Background(), streamProvider{
		events: []agentruntime.Event{message, completed},
	}, agentruntime.RunRequest{
		RunID:    "stored-run",
		Role:     agentruntime.RoleScorer,
		Provider: "stream",
		Metadata: map[string]string{
			"runStoreDir": storeDir,
		},
	})
	if err != nil {
		t.Fatalf("RunText error: %v", err)
	}
	if response != "stored" {
		t.Fatalf("response = %q, want stored", response)
	}

	metadata, events, err := runstore.NewFileStore(storeDir).LoadRun(context.Background(), "stored-run")
	if err != nil {
		t.Fatalf("LoadRun error: %v", err)
	}
	if metadata.Status != agentruntime.StatusSucceeded || metadata.EventCount != 2 {
		t.Fatalf("metadata = %#v, want succeeded with 2 events", metadata)
	}
	if len(events) != 2 || events[0].RunID != "stored-run" || events[0].Message != "stored" {
		t.Fatalf("events = %#v, want recorded response event", events)
	}
}

func TestRunTextDefaultsRuntimeStoreForWorkDir(t *testing.T) {
	workDir := t.TempDir()
	message := agentruntime.NewEvent(agentruntime.EventMessageCompleted)
	message.Message = "stored by default"
	completed := agentruntime.NewEvent(agentruntime.EventRunCompleted)
	completed.Status = agentruntime.StatusSucceeded

	response, _, err := RunText(context.Background(), streamProvider{
		events: []agentruntime.Event{message, completed},
	}, agentruntime.RunRequest{
		RunID:    "default-store-run",
		Role:     agentruntime.RolePlanner,
		Provider: "stream",
		WorkDir:  workDir,
	})
	if err != nil {
		t.Fatalf("RunText error: %v", err)
	}
	if response != "stored by default" {
		t.Fatalf("response = %q, want stored by default", response)
	}

	metadata, events, err := runstore.NewFileStore(workDir+"/.boatman/runs").LoadRun(context.Background(), "default-store-run")
	if err != nil {
		t.Fatalf("LoadRun error: %v", err)
	}
	if metadata.Status != agentruntime.StatusSucceeded || len(events) != 2 {
		t.Fatalf("metadata/events = %#v/%#v, want recorded default store", metadata, events)
	}
}

func TestRunTextRecordsProviderStartFailure(t *testing.T) {
	storeDir := t.TempDir()

	_, _, err := RunText(context.Background(), streamProvider{err: errors.New("start failed")}, agentruntime.RunRequest{
		RunID:    "failed-run",
		Provider: "stream",
		Metadata: map[string]string{
			"runStoreDir": storeDir,
		},
	})
	if err == nil || err.Error() != "start failed" {
		t.Fatalf("err = %v, want start failed", err)
	}

	metadata, events, err := runstore.NewFileStore(storeDir).LoadRun(context.Background(), "failed-run")
	if err != nil {
		t.Fatalf("LoadRun error: %v", err)
	}
	if metadata.Status != agentruntime.StatusFailed {
		t.Fatalf("metadata status = %q, want failed", metadata.Status)
	}
	if len(events) != 1 || events[0].Type != agentruntime.EventRunFailed || events[0].Message != "start failed" {
		t.Fatalf("events = %#v, want start failure event", events)
	}
}
