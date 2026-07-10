package claudecli

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	agentruntime "github.com/philjestin/boatman-ecosystem/shared/agentruntime"
	"github.com/philjestin/boatman-ecosystem/shared/agentruntime/conformance"
	"github.com/philjestin/boatmanmode/internal/cost"
)

func TestProviderImplementsRuntimeProvider(t *testing.T) {
	var _ agentruntime.Provider = New()
}

func TestWithCommand(t *testing.T) {
	provider := New(WithCommand("custom-claude"))
	if provider.command != "custom-claude" {
		t.Fatalf("command = %q, want custom-claude", provider.command)
	}
}

func TestCapabilities(t *testing.T) {
	provider := New()

	caps, err := provider.Capabilities(context.Background())
	if err != nil {
		t.Fatalf("Capabilities error: %v", err)
	}
	if caps.Provider != providerName {
		t.Fatalf("Provider = %q, want %q", caps.Provider, providerName)
	}
	if !caps.SupportsStreaming {
		t.Fatal("SupportsStreaming should be true")
	}
	if !caps.SupportsMCP {
		t.Fatal("SupportsMCP should be true")
	}
}

func TestProviderRunConformance(t *testing.T) {
	provider := New()
	provider.invoke = func(_ context.Context, _ agentruntime.RunRequest, forward func(string)) (string, *cost.Usage, error) {
		forward(`{"type":"content_block_delta","delta":{"text":"hello"}}`)
		return "done", &cost.Usage{InputTokens: 10, OutputTokens: 5, TotalCostUSD: 0.01}, nil
	}

	conformance.AssertProviderRun(t, provider, agentruntime.RunRequest{
		RunID:    "run-conformance",
		Role:     agentruntime.RoleExecutor,
		Model:    "sonnet",
		WorkDir:  "/repo",
		Messages: []agentruntime.Message{{Role: "user", Content: "build it"}},
	}, conformance.RunOptions{
		RequireMessage:     true,
		RequireUsage:       true,
		RequireProviderRaw: true,
	})
}

func TestStartRunSuccessEmitsNormalizedEvents(t *testing.T) {
	provider := New()
	provider.invoke = func(_ context.Context, req agentruntime.RunRequest, forward func(string)) (string, *cost.Usage, error) {
		if req.Role != agentruntime.RoleExecutor {
			t.Fatalf("Role = %q, want %q", req.Role, agentruntime.RoleExecutor)
		}
		forward(`{"type":"content_block_delta","delta":{"text":"hello"}}`)
		return "done", &cost.Usage{InputTokens: 10, OutputTokens: 5, TotalCostUSD: 0.01}, nil
	}

	stream, err := provider.StartRun(context.Background(), agentruntime.RunRequest{
		RunID:    "run-1",
		Role:     agentruntime.RoleExecutor,
		Model:    "sonnet",
		WorkDir:  "/repo",
		Messages: []agentruntime.Message{{Role: "user", Content: "build it"}},
	})
	if err != nil {
		t.Fatalf("StartRun error: %v", err)
	}

	events := collect(stream)
	assertEventTypes(t, events,
		agentruntime.EventRunStarted,
		agentruntime.EventPhaseStarted,
		agentruntime.EventProviderRaw,
		agentruntime.EventUsageUpdated,
		agentruntime.EventMessageCompleted,
		agentruntime.EventPhaseCompleted,
		agentruntime.EventRunCompleted,
	)

	if string(events[2].Raw) == "" {
		t.Fatal("provider raw event should preserve raw JSON")
	}
	if events[3].Usage == nil || events[3].Usage.InputTokens != 10 {
		t.Fatalf("usage event = %#v, want input tokens", events[3].Usage)
	}
	if events[4].Message != "done" {
		t.Fatalf("message = %q, want done", events[4].Message)
	}
	if events[len(events)-1].Status != agentruntime.StatusSucceeded {
		t.Fatalf("final status = %q, want %q", events[len(events)-1].Status, agentruntime.StatusSucceeded)
	}
}

func TestStartRunFailureEmitsFailedEvents(t *testing.T) {
	provider := New()
	provider.invoke = func(context.Context, agentruntime.RunRequest, func(string)) (string, *cost.Usage, error) {
		return "", nil, errors.New("boom")
	}

	stream, err := provider.StartRun(context.Background(), agentruntime.RunRequest{
		RunID: "run-2",
		Role:  agentruntime.RoleReviewer,
	})
	if err != nil {
		t.Fatalf("StartRun error: %v", err)
	}

	events := collect(stream)
	assertEventTypes(t, events,
		agentruntime.EventRunStarted,
		agentruntime.EventPhaseStarted,
		agentruntime.EventPhaseCompleted,
		agentruntime.EventRunFailed,
	)
	if events[2].Status != agentruntime.StatusFailed {
		t.Fatalf("phase status = %q, want failed", events[2].Status)
	}
	if events[3].Message != "boom" {
		t.Fatalf("failure message = %q, want boom", events[3].Message)
	}
}

func TestStartRunRejectsInvalidOutputSchema(t *testing.T) {
	provider := New()
	provider.invoke = func(context.Context, agentruntime.RunRequest, func(string)) (string, *cost.Usage, error) {
		t.Fatal("invoke should not be called for invalid schema")
		return "", nil, nil
	}

	_, err := provider.StartRun(context.Background(), agentruntime.RunRequest{
		OutputSchema: &agentruntime.OutputSchema{
			Name:   "bad",
			Schema: json.RawMessage(`{"properties":{}}`),
		},
	})
	if err == nil {
		t.Fatal("expected invalid schema error")
	}
}

func TestPromptFromMessages(t *testing.T) {
	prompt := promptFromMessages([]agentruntime.Message{
		{Role: "user", Content: "first"},
		{Role: "assistant", Content: "second"},
	})

	want := "user: first\n\nassistant: second"
	if prompt != want {
		t.Fatalf("prompt = %q, want %q", prompt, want)
	}
}

func collect(stream agentruntime.EventStream) []agentruntime.Event {
	var events []agentruntime.Event
	for event := range stream {
		events = append(events, event)
	}
	return events
}

func assertEventTypes(t *testing.T, events []agentruntime.Event, want ...agentruntime.EventType) {
	t.Helper()
	if len(events) != len(want) {
		t.Fatalf("event count = %d, want %d: %#v", len(events), len(want), events)
	}
	for i := range want {
		if events[i].Type != want[i] {
			t.Fatalf("events[%d].Type = %q, want %q", i, events[i].Type, want[i])
		}
	}
}
