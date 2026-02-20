# BoatmanMode Event System

BoatmanMode emits structured JSON events to stdout during workflow execution, enabling integration with external tools like the [boatmanapp](https://github.com/philjestin/boatmanapp) desktop application.

## Overview

Events are emitted as newline-delimited JSON (NDJSON) to stdout, allowing real-time tracking of the multi-agent orchestration workflow. Each event represents a significant state change in the workflow.

## Event Format

All events follow this JSON structure:

```json
{
  "type": "string",           // Event type (required)
  "id": "string",             // Unique identifier (optional)
  "name": "string",           // Human-readable name (optional)
  "description": "string",    // Detailed description (optional)
  "status": "string",         // Status: "success" or "failed" (optional)
  "message": "string",        // Progress message (optional)
  "data": {}                  // Additional metadata (optional)
}
```

## Event Types

### 1. `agent_started`

Emitted when an agent begins execution.

**Fields:**
- `type`: `"agent_started"` (required)
- `id`: Unique agent identifier (required) - e.g., `"prepare-ENG-123"`
- `name`: Human-readable agent name (required) - e.g., `"Planning & Analysis"`
- `description`: What the agent is doing (optional) - e.g., `"Analyzing codebase and creating implementation plan"`

**Example:**
```json
{
  "type": "agent_started",
  "id": "planning-ENG-123",
  "name": "Planning & Analysis",
  "description": "Analyzing codebase and creating implementation plan"
}
```

**When Emitted:**
- Step 1: Preparing task
- Step 2: Setting up git worktree
- Step 3: Planning & analysis
- Step 4: Pre-flight validation
- Step 5: Code execution
- Step 6: Running tests (parallel)
- Step 6: Code review (parallel)
- Step 7: Refactoring (each iteration)
- Step 8: Commit & push
- Step 9: Creating PR

### 2. `agent_completed`

Emitted when an agent finishes execution.

**Fields:**
- `type`: `"agent_completed"` (required)
- `id`: Agent identifier (must match `agent_started` id) (required)
- `name`: Agent name (optional, for display)
- `status`: `"success"` or `"failed"` (required)

**Example:**
```json
{
  "type": "agent_completed",
  "id": "planning-ENG-123",
  "name": "Planning & Analysis",
  "status": "success"
}
```

**When Emitted:**
- After each agent completes (corresponding to every `agent_started` event)

### 3. `progress`

General progress message not tied to a specific agent.

**Fields:**
- `type`: `"progress"` (required)
- `message`: Progress message (required) - e.g., `"Running tests..."`

**Example:**
```json
{
  "type": "progress",
  "message": "Review & refactor iteration 2 of 3"
}
```

**When Emitted:**
- During refactor loop iterations
- Other non-agent-specific progress updates

### 4. `task_created` *(Reserved for future use)*

Not currently emitted by boatmanmode, but supported by the event system for external integrations.

### 5. `task_updated` *(Reserved for future use)*

Not currently emitted by boatmanmode, but supported by the event system for external integrations.

### 6. `claude_stream`

Emitted for each raw stream-json line from Claude during execution. This provides full visibility into Claude's work (tool usage, streaming text, etc.) within each workflow phase.

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

**When Emitted:**
- During every Claude CLI interaction within executor, planner, and refactor agents
- Each line of Claude's `--output-format stream-json` output generates one event

**How It Works:**
The `claude.Client` struct has an `EventForwarder` callback field. When set, it is called for each raw stream-json line before the line is parsed by the client itself:

```go
// claude.Client.EventForwarder is called in doStreamingRequest()
// for each line before parsing
if c.EventForwarder != nil {
    c.EventForwarder(line)
}
```

The executor and planner set this callback during initialization:

```go
client.EventForwarder = func(rawLine string) {
    events.ClaudeStream("executor", rawLine)
}
```

The desktop app's `integration.go` handles `claude_stream` events by forwarding the raw line to `session.ProcessExternalStreamLine()`, which parses it using the same logic as standard mode sessions. This means Claude's tool usage, streaming text, and other events appear in the UI with proper agent attribution.

## Example Event Flow

Here's a typical event sequence for a successful workflow (with `claude_stream` events interspersed):

```json
{"type":"agent_started","id":"prepare-ENG-123","name":"Preparing Task","description":"Preparing task ENG-123"}
{"type":"agent_completed","id":"prepare-ENG-123","name":"Preparing Task","status":"success"}
{"type":"agent_started","id":"worktree-ENG-123","name":"Setup Worktree","description":"Creating isolated git worktree"}
{"type":"agent_completed","id":"worktree-ENG-123","name":"Setup Worktree","status":"success"}
{"type":"agent_started","id":"planning-ENG-123","name":"Planning & Analysis","description":"Analyzing codebase and creating implementation plan"}
{"type":"claude_stream","id":"planner","message":"{\"type\":\"content_block_start\",\"content_block\":{\"type\":\"text\"}}"}
{"type":"claude_stream","id":"planner","message":"{\"type\":\"content_block_delta\",\"delta\":{\"text\":\"Analyzing...\"}}"}
{"type":"agent_completed","id":"planning-ENG-123","name":"Planning & Analysis","status":"success"}
{"type":"agent_started","id":"preflight-ENG-123","name":"Pre-flight Validation","description":"Validating implementation plan"}
{"type":"agent_completed","id":"preflight-ENG-123","name":"Pre-flight Validation","status":"success"}
{"type":"agent_started","id":"execute-ENG-123","name":"Execution","description":"Implementing code changes"}
{"type":"claude_stream","id":"executor","message":"{\"type\":\"content_block_delta\",\"delta\":{\"text\":\"Creating file...\"}}"}
{"type":"claude_stream","id":"executor","message":"{\"type\":\"tool_use\",\"name\":\"Write\",\"input\":{\"path\":\"pkg/feature.go\"}}"}
{"type":"agent_completed","id":"execute-ENG-123","name":"Execution","status":"success"}
{"type":"agent_started","id":"test-ENG-123","name":"Running Tests","description":"Running unit tests for changed files"}
{"type":"agent_started","id":"review-1-ENG-123","name":"Code Review #1","description":"Reviewing code quality and best practices"}
{"type":"agent_completed","id":"test-ENG-123","name":"Running Tests","status":"success"}
{"type":"agent_completed","id":"review-1-ENG-123","name":"Code Review #1","status":"failed"}
{"type":"progress","message":"Review & refactor iteration 1 of 3"}
{"type":"agent_started","id":"refactor-1-ENG-123","name":"Refactoring #1","description":"Applying code review feedback"}
{"type":"claude_stream","id":"refactor-1","message":"{\"type\":\"content_block_delta\",\"delta\":{\"text\":\"Fixing...\"}}"}
{"type":"agent_completed","id":"refactor-1-ENG-123","name":"Refactoring #1","status":"success"}
{"type":"agent_started","id":"commit-ENG-123","name":"Commit & Push","description":"Committing and pushing changes to remote"}
{"type":"agent_completed","id":"commit-ENG-123","name":"Commit & Push","status":"success"}
{"type":"agent_started","id":"pr-ENG-123","name":"Create PR","description":"Creating pull request"}
{"type":"agent_completed","id":"pr-ENG-123","name":"Create PR","status":"success"}
```

Note: In practice, many `claude_stream` events are emitted between each `agent_started`/`agent_completed` pair. They are shown sparingly here for brevity.

## Consuming Events

### Command Line

Pipe boatmanmode output and filter for JSON events:

```bash
boatman work ENG-123 | grep '^{' | jq
```

### Go Integration

```go
package main

import (
    "bufio"
    "encoding/json"
    "log"
    "os/exec"
)

type Event struct {
    Type        string `json:"type"`
    ID          string `json:"id,omitempty"`
    Name        string `json:"name,omitempty"`
    Description string `json:"description,omitempty"`
    Status      string `json:"status,omitempty"`
    Message     string `json:"message,omitempty"`
}

func main() {
    cmd := exec.Command("boatman", "work", "ENG-123")
    stdout, _ := cmd.StdoutPipe()
    cmd.Start()

    scanner := bufio.NewScanner(stdout)
    for scanner.Scan() {
        line := scanner.Text()

        // Try to parse as JSON event
        var event Event
        if err := json.Unmarshal([]byte(line), &event); err == nil {
            // Handle event
            log.Printf("Event: %s - %s [%s]", event.Type, event.Name, event.Status)
        }
    }

    cmd.Wait()
}
```

### BoatmanApp Integration

The [boatmanapp](https://github.com/philjestin/boatmanapp) desktop application automatically parses these events and displays them in the UI:

1. **Go Backend** (`boatmanmode/integration.go`):
   - Parses JSON events from boatmanmode stdout
   - Emits Wails events to frontend: `EventsEmit("boatmanmode:event", data)`

2. **Frontend** (`frontend/src/hooks/useAgent.ts`):
   - Listens to `boatmanmode:event`
   - Calls `HandleBoatmanModeEvent(sessionId, event.type, event)`

3. **Backend** (`app.go`):
   - Creates/updates tasks in session
   - Displays in Tasks tab with appropriate icons:
     - 🤖 for agents (in_progress)
     - ✅ for success
     - ❌ for failed
     - ⏳ for progress messages

## Agent ID Format

Agent IDs follow a consistent pattern for easy tracking:

| Step | Agent ID Pattern | Example |
|------|------------------|---------|
| Prepare | `prepare-{taskID}` | `prepare-ENG-123` |
| Worktree | `worktree-{taskID}` | `worktree-ENG-123` |
| Planning | `planning-{taskID}` | `planning-ENG-123` |
| Preflight | `preflight-{taskID}` | `preflight-ENG-123` |
| Execute | `execute-{taskID}` | `execute-ENG-123` |
| Test | `test-{taskID}` | `test-ENG-123` |
| Review | `review-{iteration}-{taskID}` | `review-1-ENG-123` |
| Refactor | `refactor-{iteration}-{taskID}` | `refactor-2-ENG-123` |
| Commit | `commit-{taskID}` | `commit-ENG-123` |
| PR | `pr-{taskID}` | `pr-ENG-123` |

## Implementation Details

Events are emitted using the `internal/events` package:

```go
import "github.com/philjestin/boatmanmode/internal/events"

// Start an agent
events.AgentStarted("execute-ENG-123", "Execution", "Implementing code changes")

// Complete an agent
events.AgentCompleted("execute-ENG-123", "Execution", "success")

// Progress update
events.Progress("Running tests...")

// Forward raw Claude stream line
events.ClaudeStream("executor", rawLine)
```

All events are automatically flushed to stdout immediately, ensuring real-time updates.

### Claude Event Forwarding

The `claude.Client` struct has an `EventForwarder` field:

```go
type Client struct {
    // ...
    EventForwarder func(rawLine string) // Called for each stream-json line
}
```

When set, `doStreamingRequest()` calls it for each raw line before parsing. The executor and planner set this during initialization:

```go
// In executor.go
client.EventForwarder = func(rawLine string) {
    events.ClaudeStream("executor", rawLine)
}

// In planner.go
client.EventForwarder = func(rawLine string) {
    events.ClaudeStream("planner", rawLine)
}
```

This allows the desktop app to receive and display Claude's streaming activity (tool calls, text generation, etc.) within each workflow phase.

## Disabling Events *(Future)*

Currently, events are always emitted. A future version may add a `--no-events` flag to disable event emission for cleaner output in non-integrated environments.

## Troubleshooting

### Events not appearing

1. Check that stdout is not being redirected or buffered
2. Verify JSON parsing with: `boatman work ENG-123 | grep '^{' | jq`
3. Ensure events are not being filtered by shell pipes

### Duplicate events

Agent IDs include the task ID to prevent conflicts when running multiple workflows concurrently. Each workflow should have unique task IDs.

### Event timing

Events are emitted synchronously - `agent_started` is emitted immediately before an agent runs, and `agent_completed` immediately after. No delay or buffering occurs.

## Future Enhancements

- [ ] Add `--no-events` flag to disable event emission
- [x] ~~Emit `task_created` / `task_updated` events for internal Claude task tool usage~~ (Partially addressed: `claude_stream` now forwards all Claude events including task tool events, giving the desktop app full visibility)
- [x] ~~Add more granular events for sub-steps (file reads, code edits, etc.)~~ (Addressed: `claude_stream` forwards every Claude stream-json line, providing tool-level granularity)
- [ ] Support event output to separate file descriptor (e.g., `--events-fd=3`)
- [ ] Add event timestamps
- [ ] Support structured logging levels (debug, info, warn, error)
