package boatmanmode

import (
	"encoding/json"
	"testing"
)

func TestIsRuntimeEventType(t *testing.T) {
	tests := []struct {
		eventType string
		want      bool
	}{
		{"phase.started", true},
		{"provider.raw", true},
		{"agent_started", false},
		{"claude_stream", false},
	}

	for _, tt := range tests {
		t.Run(tt.eventType, func(t *testing.T) {
			if got := isRuntimeEventType(tt.eventType); got != tt.want {
				t.Fatalf("isRuntimeEventType(%q) = %v, want %v", tt.eventType, got, tt.want)
			}
		})
	}
}

func TestRuntimeRawPayloadPrefersRaw(t *testing.T) {
	event := BoatmanEvent{
		Type:    "provider.raw",
		Message: "fallback",
		Raw:     json.RawMessage(`{"type":"content_block_delta"}`),
	}

	if got := runtimeRawPayload(event); got != `{"type":"content_block_delta"}` {
		t.Fatalf("runtimeRawPayload = %q, want raw JSON", got)
	}
}

func TestRuntimeRawPayloadFallsBackToMessage(t *testing.T) {
	event := BoatmanEvent{
		Type:    "provider.raw",
		Message: "not json",
	}

	if got := runtimeRawPayload(event); got != "not json" {
		t.Fatalf("runtimeRawPayload = %q, want message", got)
	}
}

func TestRuntimeEnvUsesProjectBoatmanDirs(t *testing.T) {
	t.Setenv("BOATMAN_RUNTIME_STORE", "")
	t.Setenv("BOATMAN_RUNTIME_STORE_DIR", "")
	t.Setenv("BOATMAN_MEMORY_DIR", "")
	integration := &Integration{repoPath: "/repo"}

	got := integration.runtimeEnv()
	want := []string{
		"BOATMAN_RUNTIME_EVENTS=1",
		"BOATMAN_RUNTIME_STORE_DIR=/repo/.boatman/runs",
		"BOATMAN_MEMORY_DIR=/repo/.boatman/memory",
	}
	if len(got) != len(want) {
		t.Fatalf("runtimeEnv = %#v, want %#v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("runtimeEnv = %#v, want %#v", got, want)
		}
	}
}

func TestRuntimeEnvHonorsOverridesAndOptOut(t *testing.T) {
	t.Setenv("BOATMAN_RUNTIME_STORE_DIR", "/tmp/runs")
	t.Setenv("BOATMAN_MEMORY_DIR", "/tmp/memory")
	integration := &Integration{repoPath: "/repo"}

	got := integration.runtimeEnv()
	want := []string{
		"BOATMAN_RUNTIME_EVENTS=1",
		"BOATMAN_RUNTIME_STORE_DIR=/tmp/runs",
		"BOATMAN_MEMORY_DIR=/tmp/memory",
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("runtimeEnv = %#v, want %#v", got, want)
		}
	}

	t.Setenv("BOATMAN_RUNTIME_STORE", "0")
	got = integration.runtimeEnv()
	if len(got) != 2 || got[1] != "BOATMAN_MEMORY_DIR=/tmp/memory" {
		t.Fatalf("runtimeEnv with opt out = %#v, want no runtime store dir", got)
	}
}
