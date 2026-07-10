// Package events provides JSON event emission for boatmanapp integration.
// Events are emitted to stdout as newline-delimited JSON for the desktop app to parse.
package events

import (
	"encoding/json"
	"fmt"
	"os"

	agentruntime "github.com/philjestin/boatman-ecosystem/shared/agentruntime"
)

// Event represents a structured event emitted during workflow execution.
type Event struct {
	Type        string         `json:"type"`
	ID          string         `json:"id,omitempty"`
	Name        string         `json:"name,omitempty"`
	Description string         `json:"description,omitempty"`
	Status      string         `json:"status,omitempty"`
	Message     string         `json:"message,omitempty"`
	Data        map[string]any `json:"data,omitempty"`
}

// Emit writes a JSON event to stdout.
func Emit(event Event) {
	json, _ := json.Marshal(event)
	fmt.Fprintln(os.Stdout, string(json))
	if runtimeEventsEnabled() {
		EmitRuntime(Normalize(event))
	}
}

// EmitRuntime writes a provider-neutral runtime event to stdout.
func EmitRuntime(event agentruntime.Event) {
	json, _ := json.Marshal(event)
	fmt.Fprintln(os.Stdout, string(json))
}

// Normalize converts a legacy Boatman event into the provider-neutral runtime
// event stream. It is intentionally lossy only for fields that never existed in
// the legacy protocol.
func Normalize(event Event) agentruntime.Event {
	runtimeEvent := agentruntime.NewEvent(legacyRuntimeType(event.Type))
	runtimeEvent.PhaseID = event.ID
	runtimeEvent.TaskID = event.ID
	runtimeEvent.Name = event.Name
	runtimeEvent.Description = event.Description
	runtimeEvent.Message = event.Message
	runtimeEvent.Status = legacyRuntimeStatus(event.Status)
	runtimeEvent.Data = event.Data

	if event.Type == "claude_stream" {
		runtimeEvent.Provider = "claude-cli"
		runtimeEvent.Type = agentruntime.EventProviderRaw
		raw := json.RawMessage(event.Message)
		if json.Valid(raw) {
			runtimeEvent.Raw = raw
			runtimeEvent.Message = ""
		}
	}

	return runtimeEvent
}

func runtimeEventsEnabled() bool {
	return os.Getenv("BOATMAN_RUNTIME_EVENTS") == "1"
}

func legacyRuntimeType(eventType string) agentruntime.EventType {
	switch eventType {
	case "agent_started":
		return agentruntime.EventPhaseStarted
	case "agent_completed":
		return agentruntime.EventPhaseCompleted
	case "task_created":
		return agentruntime.EventTaskCreated
	case "task_updated":
		return agentruntime.EventTaskUpdated
	case "progress":
		return agentruntime.EventLogMessage
	default:
		return agentruntime.EventLogMessage
	}
}

func legacyRuntimeStatus(status string) agentruntime.Status {
	switch status {
	case "success", "succeeded", "passed":
		return agentruntime.StatusSucceeded
	case "failed", "error":
		return agentruntime.StatusFailed
	case "completed", "done":
		return agentruntime.StatusCompleted
	case "in_progress", "running":
		return agentruntime.StatusInProgress
	case "skipped":
		return agentruntime.StatusSkipped
	default:
		return agentruntime.Status(status)
	}
}

// AgentStarted emits an event when an agent begins execution.
func AgentStarted(id, name, description string) {
	Emit(Event{
		Type:        "agent_started",
		ID:          id,
		Name:        name,
		Description: description,
	})
}

// AgentCompleted emits an event when an agent finishes execution.
func AgentCompleted(id, name, status string) {
	Emit(Event{
		Type:   "agent_completed",
		ID:     id,
		Name:   name,
		Status: status,
	})
}

// AgentCompletedWithData emits an event when an agent finishes execution with additional metadata.
func AgentCompletedWithData(id, name, status string, data map[string]any) {
	Emit(Event{
		Type:   "agent_completed",
		ID:     id,
		Name:   name,
		Status: status,
		Data:   data,
	})
}

// TaskCreated emits an event when a task is created.
func TaskCreated(id, name, description string) {
	Emit(Event{
		Type:        "task_created",
		ID:          id,
		Name:        name,
		Description: description,
	})
}

// TaskUpdated emits an event when a task's status changes.
func TaskUpdated(id, status string) {
	Emit(Event{
		Type:   "task_updated",
		ID:     id,
		Status: status,
	})
}

// Progress emits a general progress message.
func Progress(message string) {
	Emit(Event{
		Type:    "progress",
		Message: message,
	})
}

// ClaudeStream emits a raw stream-json line from Claude for a given phase.
// The desktop app uses this to show Claude's streaming activity in the UI.
func ClaudeStream(phaseID string, rawLine string) {
	Emit(Event{
		Type:    "claude_stream",
		ID:      phaseID,
		Message: rawLine,
	})
}

// ProviderRaw emits a provider-neutral raw stream payload. New provider-backed
// callers should use this instead of the legacy claude_stream event.
func ProviderRaw(phaseID string, provider string, rawLine string) {
	if provider == "" {
		provider = "claude-cli"
	}
	event := agentruntime.NewEvent(agentruntime.EventProviderRaw)
	event.PhaseID = phaseID
	event.Provider = provider
	raw := json.RawMessage(rawLine)
	if json.Valid(raw) {
		event.Raw = raw
	} else {
		event.Message = rawLine
	}
	EmitRuntime(event)
}
