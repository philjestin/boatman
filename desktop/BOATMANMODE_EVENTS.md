# BoatmanMode Event System

This document explains how boatmanapp tracks boatmanmode's internal agents and tasks in the UI.

## Overview

When boatmanmode executes a ticket, it orchestrates multiple agents (planning, implementation, testing, peer review, etc.) in tmux sessions. To surface this activity in boatmanapp's UI, boatmanmode needs to emit structured JSON events to stdout.

## Event Flow

```
┌─────────────────┐
│  boatmanmode    │
│  CLI Process    │
│                 │
│  Emits JSON     │
│  events to      │
│  stdout         │
└────────┬────────┘
         │
         │ {"type": "agent_started", "id": "plan-123", ...}
         │
         ▼
┌─────────────────────────────────────┐
│  boatmanapp Integration             │
│  (boatmanmode/integration.go)       │
│                                     │
│  Parses JSON events                 │
│  Emits Wails events                 │
└────────┬────────────────────────────┘
         │
         │ EventsEmit("boatmanmode:event")
         │
         ▼
┌─────────────────────────────────────┐
│  Frontend (TypeScript)              │
│                                     │
│  Listens to boatmanmode:event       │
│  Calls HandleBoatmanModeEvent()     │
└────────┬────────────────────────────┘
         │
         ▼
┌─────────────────────────────────────┐
│  Backend (app.go)                   │
│                                     │
│  Creates/updates tasks in session   │
└────────┬────────────────────────────┘
         │
         ▼
┌─────────────────────────────────────┐
│  UI Task List                       │
│                                     │
│  Displays agents/tasks              │
└─────────────────────────────────────┘
```

## Event Format

Boatmanmode should output JSON events to stdout, one per line:

```json
{"type": "agent_started", "id": "plan-123", "name": "Planning Implementation", "description": "Creating implementation plan"}
{"type": "progress", "message": "Analyzing codebase structure..."}
{"type": "agent_completed", "id": "plan-123", "name": "Planning Implementation", "status": "success"}
{"type": "task_created", "id": "task-1", "name": "Implement feature X", "description": "Add new API endpoint"}
{"type": "task_updated", "id": "task-1", "status": "in_progress"}
```

## Event Types

### 1. `agent_started`

Emitted when an agent begins execution (e.g., planning agent, implementation agent, peer review agent).

**Fields:**
- `type`: `"agent_started"` (required)
- `id`: Unique agent identifier (required) - e.g., `"plan-123"`, `"impl-456"`
- `name`: Human-readable agent name (required) - e.g., `"Planning Implementation"`
- `description`: What the agent is doing (optional) - e.g., `"Creating implementation plan for ticket TICKET-123"`

**Example:**
```json
{
  "type": "agent_started",
  "id": "plan-a7b3c",
  "name": "Planning Implementation",
  "description": "Analyzing codebase and creating implementation plan"
}
```

**UI Behavior:**
- Creates a new task in the session's task list
- Status: `in_progress`
- Shows in Tasks tab with 🤖 icon

### 2. `agent_completed`

Emitted when an agent finishes execution.

**Fields:**
- `type`: `"agent_completed"` (required)
- `id`: Agent identifier (must match agent_started id) (required)
- `name`: Agent name (optional, for display)
- `status`: `"success"` or `"failed"` (required)
- `message`: Additional context (optional) - e.g., `"Plan validated and approved"`

**Example:**
```json
{
  "type": "agent_completed",
  "id": "plan-a7b3c",
  "name": "Planning Implementation",
  "status": "success"
}
```

**UI Behavior:**
- Updates existing task status to `completed` (if success) or `failed` (if failed)
- Shows checkmark ✅ or error ❌ in task list

### 3. `task_created`

Emitted when boatmanmode's internal task system creates a task (e.g., from Claude CLI task tools).

**Fields:**
- `type`: `"task_created"` (required)
- `id`: Task identifier (required) - e.g., `"task-1"`, `"task-2"`
- `name`: Task subject (required) - e.g., `"Implement API endpoint"`
- `description`: Detailed task description (optional)
- `status`: Initial status (optional, defaults to `"pending"`)

**Example:**
```json
{
  "type": "task_created",
  "id": "task-1",
  "name": "Implement /api/users endpoint",
  "description": "Add GET and POST handlers for user management",
  "status": "pending"
}
```

**UI Behavior:**
- Creates a new task in the session's task list
- Status: `pending` (or specified status)
- Shows in Tasks tab with 📋 icon

### 4. `task_updated`

Emitted when a task's status changes.

**Fields:**
- `type`: `"task_updated"` (required)
- `id`: Task identifier (must match task_created id) (required)
- `name`: Task name (optional, can update name if provided)
- `status`: New status (required) - `"pending"`, `"in_progress"`, `"completed"`, `"failed"`

**Example:**
```json
{
  "type": "task_updated",
  "id": "task-1",
  "status": "in_progress"
}
```

**UI Behavior:**
- Updates existing task status
- Shows appropriate icon based on status

### 5. `progress`

General progress message (not tied to specific agent/task).

**Fields:**
- `type`: `"progress"` (required)
- `message`: Progress message (required) - e.g., `"Running tests..."`

**Example:**
```json
{
  "type": "progress",
  "message": "Running unit tests..."
}
```

**UI Behavior:**
- Displays in output stream
- Shows with ⏳ icon
- Does NOT create a task

## Implementation in BoatmanMode CLI

To add event emission to boatmanmode:

### 1. Create Event Emitter

```go
// pkg/events/emitter.go
package events

import (
	"encoding/json"
	"fmt"
	"os"
)

type Event struct {
	Type        string                 `json:"type"`
	ID          string                 `json:"id,omitempty"`
	Name        string                 `json:"name,omitempty"`
	Description string                 `json:"description,omitempty"`
	Status      string                 `json:"status,omitempty"`
	Message     string                 `json:"message,omitempty"`
	Data        map[string]interface{} `json:"data,omitempty"`
}

func Emit(event Event) {
	json, _ := json.Marshal(event)
	fmt.Fprintln(os.Stdout, string(json))
}

func AgentStarted(id, name, description string) {
	Emit(Event{
		Type:        "agent_started",
		ID:          id,
		Name:        name,
		Description: description,
	})
}

func AgentCompleted(id, name, status string) {
	Emit(Event{
		Type:   "agent_completed",
		ID:     id,
		Name:   name,
		Status: status,
	})
}

func TaskCreated(id, name, description string) {
	Emit(Event{
		Type:        "task_created",
		ID:          id,
		Name:        name,
		Description: description,
	})
}

func TaskUpdated(id, status string) {
	Emit(Event{
		Type:   "task_updated",
		ID:     id,
		Status: status,
	})
}

func Progress(message string) {
	Emit(Event{
		Type:    "progress",
		Message: message,
	})
}
```

### 2. Emit Events During Execution

```go
// cmd/boatmanmode/execute.go
package main

import (
	"github.com/philjestin/boatmanmode/pkg/events"
)

func executeTicket(ticketID string) error {
	// Start planning agent
	planAgentID := "plan-" + generateID()
	events.AgentStarted(planAgentID, "Planning Implementation", "Creating implementation plan for "+ticketID)

	// Run planning
	plan, err := runPlanningAgent(ticketID)
	if err != nil {
		events.AgentCompleted(planAgentID, "Planning Implementation", "failed")
		return err
	}
	events.AgentCompleted(planAgentID, "Planning Implementation", "success")

	// Start implementation agent
	implAgentID := "impl-" + generateID()
	events.AgentStarted(implAgentID, "Implementation", "Writing code based on plan")

	// Implementation creates tasks
	for _, step := range plan.Steps {
		taskID := "task-" + generateID()
		events.TaskCreated(taskID, step.Title, step.Description)

		events.TaskUpdated(taskID, "in_progress")
		// ... do work ...
		events.TaskUpdated(taskID, "completed")
	}

	events.AgentCompleted(implAgentID, "Implementation", "success")

	return nil
}
```

### 3. Hook into Existing Agents

For tmux-based agents, emit events before/after launching them:

```go
// When launching a tmux agent
agentID := "peer-review-" + generateID()
events.AgentStarted(agentID, "Peer Review", "Reviewing code quality and best practices")

// Launch tmux session
session := tmux.NewSession("peer-review")
session.Start()

// Wait for completion
session.Wait()

events.AgentCompleted(agentID, "Peer Review", "success")
```

### 4. Forward Claude CLI Task Events

When Claude CLI emits task events (via `--output-format stream-json`), forward them:

```go
// Parse Claude output
scanner := bufio.NewScanner(stdout)
for scanner.Scan() {
	line := scanner.Text()

	var claudeEvent map[string]interface{}
	if err := json.Unmarshal([]byte(line), &claudeEvent); err == nil {
		// Check if it's a task event
		if eventType, ok := claudeEvent["event_type"].(string); ok {
			switch eventType {
			case "task_create":
				task := claudeEvent["task"].(map[string]interface{})
				events.TaskCreated(
					task["id"].(string),
					task["subject"].(string),
					task["description"].(string),
				)
			case "task_update":
				task := claudeEvent["task"].(map[string]interface{})
				events.TaskUpdated(
					task["id"].(string),
					task["status"].(string),
				)
			}
		}
	}
}
```

## Testing Event Emission

### Manual Test

Run boatmanmode with streaming:

```bash
cd ~/workspace/handshake/boatmanmode
go run ./cmd/boatmanmode execute --ticket TICKET-123 --repo ~/workspace/myproject --stream
```

Expected output:
```json
{"type":"agent_started","id":"plan-abc123","name":"Planning Implementation","description":"Creating implementation plan"}
{"type":"progress","message":"Analyzing codebase..."}
{"type":"agent_completed","id":"plan-abc123","status":"success"}
{"type":"agent_started","id":"impl-def456","name":"Implementation","description":"Writing code"}
{"type":"task_created","id":"task-1","name":"Add API endpoint","description":"Implement /api/users"}
{"type":"task_updated","id":"task-1","status":"in_progress"}
{"type":"task_updated","id":"task-1","status":"completed"}
{"type":"agent_completed","id":"impl-def456","status":"success"}
```

### Integration Test with BoatmanApp

1. Build boatmanmode:
   ```bash
   cd ~/workspace/handshake/boatmanmode
   go build -o boatman ./cmd/boatmanmode
   ```

2. Run boatmanapp:
   ```bash
   cd /Users/pmiddleton/workspace/personal/boatmanapp
   wails dev
   ```

3. Create a boatmanmode session:
   - Click "Boatman Mode" button
   - Enter ticket ID: `TICKET-123`
   - Click "Start"

4. Verify in UI:
   - Switch to Tasks tab
   - Should see agents appearing as tasks:
     - 🤖 Planning Implementation (in_progress)
     - ✅ Planning Implementation (completed)
     - 🤖 Implementation (in_progress)
     - 📋 Add API endpoint (in_progress)
   - Output stream should show formatted messages

## Frontend Integration

Listen to boatmanmode events and update UI:

```typescript
// frontend/src/hooks/useAgent.ts

useEffect(() => {
  const unsubscribe = EventsOn("boatmanmode:event", (data: any) => {
    const { sessionId, event } = data;

    // Call backend to update session tasks
    HandleBoatmanModeEvent(sessionId, event.type, event);
  });

  return unsubscribe;
}, []);
```

## Troubleshooting

### Events not appearing in UI

1. Check boatmanmode output contains JSON events:
   ```bash
   boatman execute --ticket TICKET-123 --stream | grep '{"type"'
   ```

2. Check browser console for `boatmanmode:event` logs:
   ```javascript
   EventsOn("boatmanmode:event", (data) => console.log("Event:", data))
   ```

3. Verify HandleBoatmanModeEvent is being called:
   ```typescript
   console.log("Handling event:", sessionId, eventType, eventData)
   ```

### Duplicate tasks

Ensure agent IDs are unique and consistent:
- ✅ `plan-` + UUID or timestamp
- ❌ Hardcoded `plan-agent` (will cause duplicates across tickets)

### Tasks stuck in "in_progress"

Always emit `agent_completed` or `task_updated` with final status:
```go
defer func() {
  if err != nil {
    events.AgentCompleted(agentID, name, "failed")
  } else {
    events.AgentCompleted(agentID, name, "success")
  }
}()
```

## Summary

**BoatmanMode needs to:**
1. Output JSON events to stdout (one per line)
2. Emit `agent_started` when agents begin
3. Emit `agent_completed` when agents finish
4. Emit `task_created` / `task_updated` for task lifecycle
5. Use unique, consistent IDs for agents/tasks

**BoatmanApp will:**
1. Parse JSON events from stdout
2. Emit Wails events to frontend
3. Create/update tasks in session
4. Display agents/tasks in UI

This gives users full visibility into boatmanmode's multi-agent orchestration! 🚀
