# BoatmanMode Event System

This document explains how boatmanapp tracks boatmanmode's internal agents and tasks in the UI.

## Overview

When boatmanmode executes a ticket, it orchestrates multiple agents (planning, implementation, testing, peer review, etc.) in tmux sessions. To surface this activity in boatmanapp's UI, boatmanmode emits structured JSON events to stdout. Additionally, Claude's raw stream-json lines are forwarded via `claude_stream` events for full visibility into Claude's work within each phase.

## Event Flow

There are two event routing paths:

### Standard Mode (Claude Agent Sessions)

When Claude spawns subagents via the Task tool, the session tracks them automatically:

```
Claude API (stream-json)
    │
    ▼
┌─────────────────────────────────────┐
│  Session (session.go)               │
│                                     │
│  parseStreamLine() handles:         │
│  - tool_use(Task) → registers agent │
│    + maps tool_use_id → agent_id    │
│  - user(parent_tool_use_id) →       │
│    pushes agentStack, switches      │
│    currentAgentID to subagent       │
│  - content/tool events → attributed │
│    to currentAgentID                │
│  - tool_result(matching id) →       │
│    pops stack, marks completed      │
└────────┬────────────────────────────┘
         │ onMessage callback
         │
         ▼
┌─────────────────────────────────────┐
│  Wails EventsEmit("agent:message") │
└────────┬────────────────────────────┘
         │
         ▼
┌─────────────────────────────────────┐
│  Frontend (useAgent.ts)             │
│  AgentLogsPanel groups by agent     │
│  MessageBubble shows agent badges   │
└─────────────────────────────────────┘
```

### BoatmanMode (CLI Subprocess)

Boatmanmode events take two paths into the session:

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
         │ {"type": "agent_started", ...}
         │ {"type": "claude_stream", ...}
         │ Regular text output
         │
         ▼
┌─────────────────────────────────────┐
│  Integration (integration.go)       │
│                                     │
│  Parses JSON events:                │
│  - Structured events → Wails emit   │
│  - claude_stream → onMessage()      │
│  - Non-JSON text → onMessage()      │
└────────┬──────────┬─────────────────┘
         │          │
    Wails emit   MessageCallback
         │          │
         ▼          ▼
┌────────────┐  ┌────────────────────────────┐
│  Frontend  │  │  app.go                    │
│  handles   │  │  HandleBoatmanModeEvent()  │
│  boatman   │  │                            │
│  mode:     │  │  agent_started →           │
│  event     │  │    RegisterBoatmanAgent()  │
│            │  │    SetCurrentAgent()       │
│            │  │    AddBoatmanMessage()     │
│            │  │                            │
│            │  │  agent_completed →         │
│            │  │    AddBoatmanMessage()     │
│            │  │    CompleteCurrentAgent()  │
│            │  │                            │
│            │  │  progress →               │
│            │  │    AddBoatmanMessage()     │
│            │  │                            │
│            │  │  claude_stream →           │
│            │  │    ProcessExternalStream   │
│            │  │    Line() → parseStream   │
│            │  │    Line()                  │
└────────────┘  └────────────┬───────────────┘
                             │
                             │ Session message system
                             │ (onMessage callback)
                             ▼
                ┌────────────────────────────┐
                │  Wails "agent:message"     │
                │  → Frontend chat UI        │
                │  → AgentLogsPanel          │
                └────────────────────────────┘
```

**Key change:** Messages now primarily flow through the session message system (`agent:message` Wails channel) with proper agent attribution, rather than through `boatmanmode:output`. The `boatmanmode:output` channel remains as a fallback for raw text.

## Event Format

Boatmanmode should output JSON events to stdout, one per line:

```json
{"type": "agent_started", "id": "plan-123", "name": "Planning Implementation", "description": "Creating implementation plan"}
{"type": "progress", "message": "Analyzing codebase structure..."}
{"type": "agent_completed", "id": "plan-123", "name": "Planning Implementation", "status": "success"}
{"type": "claude_stream", "id": "executor", "message": "{\"type\":\"content_block_delta\",\"content\":\"...\"}"}
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
- Creates a new task in the session's task list (status: `in_progress`)
- Registers a new agent via `session.RegisterBoatmanAgent(id, name, description)`
- Switches context via `session.SetCurrentAgent(id)` - subsequent messages are attributed to this agent
- Adds a system message: "Started: {name}"
- AgentLogsPanel shows a new tab with the agent's type and a pulsing cyan status dot

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
- Adds a system message: "Completed: {name} ({status})"
- Calls `session.CompleteCurrentAgent()` which marks the agent as completed and restores the parent agent
- AgentLogsPanel tab shows a solid green dot for completed agents

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
- Routed through `session.AddBoatmanMessage("system", message)` to appear as a system message in the chat, attributed to the current agent
- Displayed in the AgentLogsPanel under the active agent's tab

### 6. `claude_stream`

Raw stream-json line from Claude, forwarded by the CLI for full visibility into Claude's work.

**Fields:**
- `type`: `"claude_stream"` (required)
- `id`: Phase identifier (required) - e.g., `"executor"`, `"planner"`, `"refactor-1"`
- `message`: Raw stream-json line from Claude (required)

**Example:**
```json
{
  "type": "claude_stream",
  "id": "executor",
  "message": "{\"type\":\"content_block_delta\",\"content\":\"Implementing the feature...\"}"
}
```

**How it works:**
1. The CLI's `claude.Client` has an `EventForwarder` callback that fires for each raw stream-json line
2. The executor/planner sets this callback to call `events.ClaudeStream(phaseID, rawLine)`
3. The event is emitted as JSON to stdout
4. `integration.go` parses it and calls `onMessage("claude_stream", rawLine)`
5. `app.go` routes it to `session.ProcessExternalStreamLine()` which wraps `parseStreamLine()`
6. The session parses the raw line exactly like standard mode events (tool use, content deltas, etc.)

**UI Behavior:**
- Claude's streaming text, tool usage, and tool results appear in the chat attributed to the current boatmanmode agent
- Provides full visibility into what Claude is doing during each phase (previously invisible)

## Standard Mode Subagent Tracking

When using the desktop app in standard mode (direct Claude sessions), subagent tracking happens automatically:

### Subagent Lifecycle

```
1. tool_use (name=Task, id=toolu_abc)
   → Session registers agent-123
   → Maps toolu_abc → agent-123 in toolIDToAgentID

2. user (parent_tool_use_id=toolu_abc)
   → Looks up agent-123 from toolIDToAgentID
   → Pushes "main" onto agentStack
   → Sets currentAgentID = agent-123
   → Subsequent messages attributed to agent-123

3. content_block_delta, tool_use, tool_result
   → All attributed to agent-123

4. tool_result (tool_use_id=toolu_abc)
   → Detects toolu_abc in toolIDToAgentID
   → Marks agent-123 as completed
   → Pops agentStack → currentAgentID = "main"
   → Cleans up toolIDToAgentID entry
```

### Session Fields

```go
type Session struct {
    // ...existing fields...
    toolIDToAgentID map[string]string // Maps Task tool_use ID → spawned agent ID
    agentStack      []string          // Stack of active agent IDs for nested subagents
}
```

### Nested Subagents

The stack-based approach supports nested subagents (e.g., a Task agent spawning its own Task agents). Each level pushes onto the stack and pops on completion.

## Implementation in BoatmanMode CLI

### Event Emitter

Events are emitted using the `internal/events` package:

```go
import "github.com/philjestin/boatmanmode/internal/events"

// Start an agent
events.AgentStarted("execute-ENG-123", "Execution", "Implementing code changes")

// Complete an agent
events.AgentCompleted("execute-ENG-123", "Execution", "success")

// Progress update
events.Progress("Running tests...")

// Forward Claude stream line (called via EventForwarder callback)
events.ClaudeStream("executor", rawLine)
```

### Claude Event Forwarding

The `claude.Client` supports an `EventForwarder` callback:

```go
client := claude.NewWithTools(worktreePath, "executor", nil)
client.EventForwarder = func(rawLine string) {
    events.ClaudeStream("executor", rawLine)
}
```

This is set automatically in `executor.go` and `planner.go`.

## Boatmanmode Agent Management API

The desktop session provides these exported methods for managing boatmanmode agents:

```go
// Register a new agent for a boatmanmode phase
session.RegisterBoatmanAgent(agentID, agentType, description string)

// Push current agent and switch to the given agent
session.SetCurrentAgent(agentID string)

// Mark current agent completed and pop the stack
session.CompleteCurrentAgent()

// Add a message attributed to the current agent
session.AddBoatmanMessage(role, content string)

// Process a raw stream-json line from an external source
session.ProcessExternalStreamLine(line string, responseBuilder *strings.Builder, currentMessageID *string)
```

## Testing Event Emission

### Manual Test

Run boatmanmode with streaming:

```bash
cd ~/workspace/personal/boatman-ecosystem/cli
go run ./cmd/boatman work --prompt "Add health check" --repo ~/workspace/myproject
```

Expected output includes both high-level events and claude_stream events:
```json
{"type":"agent_started","id":"plan-abc123","name":"Planning Implementation","description":"Creating implementation plan"}
{"type":"claude_stream","id":"planner","message":"{\"type\":\"content_block_delta\",\"content\":\"Analyzing...\"}"}
{"type":"progress","message":"Analyzing codebase..."}
{"type":"agent_completed","id":"plan-abc123","status":"success"}
{"type":"agent_started","id":"impl-def456","name":"Execution","description":"Writing code"}
{"type":"claude_stream","id":"executor","message":"{\"type\":\"content_block_delta\",\"content\":\"Creating file...\"}"}
{"type":"agent_completed","id":"impl-def456","status":"success"}
```

### Integration Test with Desktop App

1. Build CLI:
   ```bash
   cd ~/workspace/personal/boatman-ecosystem/cli
   go build -o boatman ./cmd/boatman
   ```

2. Run desktop app:
   ```bash
   cd ~/workspace/personal/boatman-ecosystem/desktop
   wails dev
   ```

3. Create a boatmanmode session and verify:
   - AgentLogsPanel shows separate tabs per phase (Planning, Execution, Review, etc.)
   - Each tab has a status indicator (pulsing cyan = active, green = completed)
   - Messages within each tab show Claude's streaming activity
   - MessageBubble shows agent badges next to "Claude" for non-main agents

### Standard Mode Test

1. Start a standard Claude session
2. Send a message that triggers Claude to spawn a Task/Explore subagent
3. Verify:
   - AgentLogsPanel shows a new tab for the subagent
   - Messages in that tab are properly attributed
   - Tab shows active/completed status

## Frontend Integration

Messages from boatmanmode now primarily flow through the `agent:message` Wails channel, the same channel used by standard mode sessions. The `boatmanmode:output` handler in `useAgent.ts` is a no-op fallback that only logs:

```typescript
// frontend/src/hooks/useAgent.ts

// Primary message flow - handles all session messages including boatmanmode
EventsOn('agent:message', (data) => {
  addMessage(data.sessionId, data.message);
  // Messages include metadata.agent with agentId, agentType, status, etc.
});

// Boatmanmode structured events still flow through their own channel
// for task tracking (creating/updating tasks in the session)
EventsOn('boatmanmode:event', async (data) => {
  await HandleBoatmanModeEvent(data.sessionId, data.event.type, data.event);
});

// Fallback for raw text - now a no-op since messages route through agent:message
EventsOn('boatmanmode:output', (data) => {
  console.log('[FRONTEND] Boatmanmode raw output (fallback):', data.message?.substring(0, 100));
});
```

### AgentLogsPanel

The `AgentLogsPanel` component groups messages by agent and shows:
- Separate tabs per agent with color-coded types
- Status indicators per tab (pulsing cyan dot = active, solid green = completed)
- Agent hierarchy view showing parent-child relationships

**Agent type colors:**
| Type | Color |
|------|-------|
| main | Blue |
| task | Purple |
| Explore | Green |
| Plan | Amber |
| general-purpose | Cyan |
| Execution | Orange |
| Planning | Yellow |
| Review | Pink |
| Refactor | Indigo |
| Testing | Teal |

### MessageBubble

When a message's agent is not "main", a purple badge appears next to "Claude" showing the agent type and description (truncated to 30 chars).

## Troubleshooting

### Events not appearing in UI

1. Check boatmanmode output contains JSON events:
   ```bash
   boatman work --prompt "test" | grep '{"type"'
   ```

2. Check browser console for `agent:message` events (primary channel):
   ```javascript
   // Messages now flow through agent:message, not boatmanmode:output
   ```

3. Verify HandleBoatmanModeEvent is being called in the console logs

### Subagent tabs not appearing (Standard Mode)

1. Verify the Claude response includes `parent_tool_use_id` in user messages
2. Check console for `[user event] Switched to subagent` log lines
3. Verify `[handleToolResult] Subagent completed` appears when the subagent finishes

### Claude stream events not visible (Boatmanmode)

1. Verify the CLI was built with `EventForwarder` support:
   ```bash
   cd ~/workspace/personal/boatman-ecosystem/cli
   go build -o boatman ./cmd/boatman
   ```

2. Check that `claude_stream` events appear in CLI stdout:
   ```bash
   boatman work --prompt "test" 2>/dev/null | grep claude_stream
   ```

3. Verify `integration.go` is routing `claude_stream` events via `onMessage`

### Duplicate tasks

Ensure agent IDs are unique and consistent:
- Use `{phase}-{taskID}` format (e.g., `planning-ENG-123`)
- Avoid hardcoded IDs that cause conflicts across tickets

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

**BoatmanMode CLI emits:**
1. `agent_started` / `agent_completed` for workflow phases
2. `progress` for general status updates
3. `claude_stream` for raw Claude stream-json lines (full visibility)
4. `task_created` / `task_updated` for task lifecycle

**Desktop App routes events through the session:**
1. Structured events → `HandleBoatmanModeEvent()` → session agent methods
2. Claude stream lines → `ProcessExternalStreamLine()` → session message system
3. Raw CLI output → `AddBoatmanMessage()` → session message system
4. All messages flow via `agent:message` Wails channel with proper agent attribution

**Standard Mode tracks subagents automatically:**
1. `tool_use(Task)` → registers agent + maps tool ID
2. `user(parent_tool_use_id)` → switches context to subagent
3. Messages attributed to current agent via `agentStack`
4. `tool_result` → completes subagent + restores parent
