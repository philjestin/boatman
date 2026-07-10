// Package conformance contains reusable checks for agentruntime providers.
package conformance

import (
	"context"
	"testing"

	agentruntime "github.com/philjestin/boatman-ecosystem/shared/agentruntime"
)

// RunOptions controls optional provider run assertions.
type RunOptions struct {
	RequireMessage     bool
	RequireUsage       bool
	RequireProviderRaw bool
}

// AssertProviderRun starts a provider run and verifies that it emits a
// well-formed successful runtime lifecycle. It returns the collected events so
// adapter-specific tests can add deeper assertions.
func AssertProviderRun(
	t testing.TB,
	provider agentruntime.Provider,
	req agentruntime.RunRequest,
	opts RunOptions,
) []agentruntime.Event {
	t.Helper()

	if provider == nil {
		t.Fatal("provider must not be nil")
	}
	if req.RunID == "" {
		req.RunID = "conformance-run"
	}
	if req.Role == "" {
		req.Role = agentruntime.RoleChat
	}
	if req.Provider == "" {
		req.Provider = provider.Name()
	}

	caps, err := provider.Capabilities(context.Background())
	if err != nil {
		t.Fatalf("Capabilities error: %v", err)
	}
	if caps.Provider != "" && caps.Provider != provider.Name() {
		t.Fatalf("capabilities provider = %q, want %q", caps.Provider, provider.Name())
	}

	stream, err := provider.StartRun(context.Background(), req)
	if err != nil {
		t.Fatalf("StartRun error: %v", err)
	}

	var events []agentruntime.Event
	for event := range stream {
		events = append(events, event)
		if event.Version != agentruntime.ProtocolVersion {
			t.Fatalf("event %q version = %d, want %d", event.Type, event.Version, agentruntime.ProtocolVersion)
		}
		if event.Type == "" {
			t.Fatal("event type must not be empty")
		}
		if event.Timestamp.IsZero() {
			t.Fatalf("event %q timestamp should be set", event.Type)
		}
		if event.RunID == "" {
			t.Fatalf("event %q run ID should be set", event.Type)
		}
		if event.Provider != "" && event.Provider != provider.Name() {
			t.Fatalf("event %q provider = %q, want %q", event.Type, event.Provider, provider.Name())
		}
	}

	if len(events) == 0 {
		t.Fatal("provider emitted no events")
	}
	if events[0].Type != agentruntime.EventRunStarted {
		t.Fatalf("first event = %q, want %q", events[0].Type, agentruntime.EventRunStarted)
	}
	if !hasEvent(events, agentruntime.EventPhaseStarted) {
		t.Fatalf("events should include %q", agentruntime.EventPhaseStarted)
	}
	if !hasEvent(events, agentruntime.EventPhaseCompleted) {
		t.Fatalf("events should include %q", agentruntime.EventPhaseCompleted)
	}
	if !hasEvent(events, agentruntime.EventRunCompleted) {
		t.Fatalf("events should include successful %q", agentruntime.EventRunCompleted)
	}
	if hasEvent(events, agentruntime.EventRunFailed) {
		t.Fatalf("successful conformance run should not emit %q", agentruntime.EventRunFailed)
	}
	if opts.RequireMessage && !hasEvent(events, agentruntime.EventMessageCompleted) {
		t.Fatalf("events should include %q", agentruntime.EventMessageCompleted)
	}
	if opts.RequireUsage && !hasEvent(events, agentruntime.EventUsageUpdated) {
		t.Fatalf("events should include %q", agentruntime.EventUsageUpdated)
	}
	if opts.RequireProviderRaw && !hasEvent(events, agentruntime.EventProviderRaw) {
		t.Fatalf("events should include %q", agentruntime.EventProviderRaw)
	}

	return events
}

func hasEvent(events []agentruntime.Event, eventType agentruntime.EventType) bool {
	for _, event := range events {
		if event.Type == eventType {
			return true
		}
	}
	return false
}
