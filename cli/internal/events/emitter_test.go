package events

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	agentruntime "github.com/philjestin/boatman-ecosystem/shared/agentruntime"
)

func TestAgentStarted(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	AgentStarted("test-123", "Test Agent", "Testing agent events")

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("Failed to read from pipe: %v", err)
	}

	var event Event
	if err := json.Unmarshal(buf.Bytes(), &event); err != nil {
		t.Fatalf("Failed to parse event JSON: %v", err)
	}

	if event.Type != "agent_started" {
		t.Errorf("Expected type 'agent_started', got '%s'", event.Type)
	}
	if event.ID != "test-123" {
		t.Errorf("Expected ID 'test-123', got '%s'", event.ID)
	}
	if event.Name != "Test Agent" {
		t.Errorf("Expected name 'Test Agent', got '%s'", event.Name)
	}
	if event.Description != "Testing agent events" {
		t.Errorf("Expected description 'Testing agent events', got '%s'", event.Description)
	}
}

func TestAgentCompleted(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	AgentCompleted("test-123", "Test Agent", "success")

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("Failed to read from pipe: %v", err)
	}

	var event Event
	if err := json.Unmarshal(buf.Bytes(), &event); err != nil {
		t.Fatalf("Failed to parse event JSON: %v", err)
	}

	if event.Type != "agent_completed" {
		t.Errorf("Expected type 'agent_completed', got '%s'", event.Type)
	}
	if event.ID != "test-123" {
		t.Errorf("Expected ID 'test-123', got '%s'", event.ID)
	}
	if event.Status != "success" {
		t.Errorf("Expected status 'success', got '%s'", event.Status)
	}
}

func TestTaskCreated(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	TaskCreated("task-1", "Implement feature", "Add new API endpoint")

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("Failed to read from pipe: %v", err)
	}

	var event Event
	if err := json.Unmarshal(buf.Bytes(), &event); err != nil {
		t.Fatalf("Failed to parse event JSON: %v", err)
	}

	if event.Type != "task_created" {
		t.Errorf("Expected type 'task_created', got '%s'", event.Type)
	}
	if event.ID != "task-1" {
		t.Errorf("Expected ID 'task-1', got '%s'", event.ID)
	}
}

func TestTaskUpdated(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	TaskUpdated("task-1", "in_progress")

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("Failed to read from pipe: %v", err)
	}

	var event Event
	if err := json.Unmarshal(buf.Bytes(), &event); err != nil {
		t.Fatalf("Failed to parse event JSON: %v", err)
	}

	if event.Type != "task_updated" {
		t.Errorf("Expected type 'task_updated', got '%s'", event.Type)
	}
	if event.Status != "in_progress" {
		t.Errorf("Expected status 'in_progress', got '%s'", event.Status)
	}
}

func TestProgress(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	Progress("Running tests...")

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("Failed to read from pipe: %v", err)
	}

	var event Event
	if err := json.Unmarshal(buf.Bytes(), &event); err != nil {
		t.Fatalf("Failed to parse event JSON: %v", err)
	}

	if event.Type != "progress" {
		t.Errorf("Expected type 'progress', got '%s'", event.Type)
	}
	if event.Message != "Running tests..." {
		t.Errorf("Expected message 'Running tests...', got '%s'", event.Message)
	}
}

func TestRuntimeEventBridgeOptIn(t *testing.T) {
	t.Setenv("BOATMAN_RUNTIME_EVENTS", "1")

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	AgentCompleted("execute-123", "Execution", "success")

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("Failed to read from pipe: %v", err)
	}

	dec := json.NewDecoder(&buf)
	var legacy Event
	if err := dec.Decode(&legacy); err != nil {
		t.Fatalf("Failed to parse legacy event JSON: %v", err)
	}
	var runtimeEvent agentruntime.Event
	if err := dec.Decode(&runtimeEvent); err != nil {
		t.Fatalf("Failed to parse runtime event JSON: %v", err)
	}

	if legacy.Type != "agent_completed" {
		t.Fatalf("legacy.Type = %q, want agent_completed", legacy.Type)
	}
	if runtimeEvent.Type != agentruntime.EventPhaseCompleted {
		t.Fatalf("runtimeEvent.Type = %q, want %q", runtimeEvent.Type, agentruntime.EventPhaseCompleted)
	}
	if runtimeEvent.Status != agentruntime.StatusSucceeded {
		t.Fatalf("runtimeEvent.Status = %q, want %q", runtimeEvent.Status, agentruntime.StatusSucceeded)
	}
	if runtimeEvent.PhaseID != "execute-123" {
		t.Fatalf("runtimeEvent.PhaseID = %q, want execute-123", runtimeEvent.PhaseID)
	}
}

func TestNormalizeClaudeStream(t *testing.T) {
	event := Event{
		Type:    "claude_stream",
		ID:      "executor",
		Message: `{"type":"content_block_delta","delta":{"text":"hello"}}`,
	}

	runtimeEvent := Normalize(event)

	if runtimeEvent.Type != agentruntime.EventProviderRaw {
		t.Fatalf("Type = %q, want %q", runtimeEvent.Type, agentruntime.EventProviderRaw)
	}
	if runtimeEvent.Provider != "claude-cli" {
		t.Fatalf("Provider = %q, want claude-cli", runtimeEvent.Provider)
	}
	if string(runtimeEvent.Raw) != event.Message {
		t.Fatalf("Raw = %s, want %s", runtimeEvent.Raw, event.Message)
	}
}

func TestProviderRawEmitsRuntimeEvent(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	ProviderRaw("executor", "", `{"type":"content_block_delta","delta":{"text":"hello"}}`)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("Failed to read from pipe: %v", err)
	}

	var event agentruntime.Event
	if err := json.Unmarshal(buf.Bytes(), &event); err != nil {
		t.Fatalf("Failed to parse runtime event JSON: %v", err)
	}
	if event.Type != agentruntime.EventProviderRaw {
		t.Fatalf("Type = %q, want %q", event.Type, agentruntime.EventProviderRaw)
	}
	if event.PhaseID != "executor" {
		t.Fatalf("PhaseID = %q, want executor", event.PhaseID)
	}
	if event.Provider != "claude-cli" {
		t.Fatalf("Provider = %q, want claude-cli", event.Provider)
	}
	if string(event.Raw) != `{"type":"content_block_delta","delta":{"text":"hello"}}` {
		t.Fatalf("Raw = %s", event.Raw)
	}
}
