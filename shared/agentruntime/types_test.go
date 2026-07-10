package agentruntime

import (
	"encoding/json"
	"testing"
)

func TestNewEventSetsProtocolMetadata(t *testing.T) {
	event := NewEvent(EventPhaseStarted)

	if event.Version != ProtocolVersion {
		t.Fatalf("Version = %d, want %d", event.Version, ProtocolVersion)
	}
	if event.Type != EventPhaseStarted {
		t.Fatalf("Type = %q, want %q", event.Type, EventPhaseStarted)
	}
	if event.Timestamp.IsZero() {
		t.Fatal("Timestamp should be set")
	}
}

func TestEventMarshalWithRawProviderPayload(t *testing.T) {
	event := NewEvent(EventProviderRaw)
	event.Provider = "claude-cli"
	event.PhaseID = "executor"
	event.Raw = json.RawMessage(`{"type":"content_block_delta","delta":{"text":"hi"}}`)

	payload, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded Event
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if decoded.Provider != "claude-cli" {
		t.Fatalf("Provider = %q, want claude-cli", decoded.Provider)
	}
	if string(decoded.Raw) != `{"type":"content_block_delta","delta":{"text":"hi"}}` {
		t.Fatalf("Raw = %s", decoded.Raw)
	}
}
