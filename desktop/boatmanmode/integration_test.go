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
