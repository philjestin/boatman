# BoatmanMode Integration

This document explains how `boatmanapp` integrates with `boatmanmode` for automated Linear ticket execution.

## Overview

**BoatmanApp** is the desktop UI that provides:
- Chat interface with Claude
- Firefighter mode for production incident investigation
- Manual investigation of Linear tickets with Bugsnag/Datadog integration

**BoatmanMode** is the CLI tool that provides:
- Automated ticket execution with full workflow (plan → implement → test → peer review → PR)
- Git worktree isolation
- Multi-agent coordination
- Iterative refinement based on peer review feedback

## Integration Approach

### Subprocess-Based Integration

BoatmanApp calls the `boatman` CLI binary as a subprocess rather than importing the Go modules directly.

**Why?**
- BoatmanMode's packages are in `internal/` which can't be imported by external modules (Go restriction)
- Clean separation of concerns - UI vs automation engine
- No version coupling - each can be updated independently
- Can use boatmanmode as a standalone tool or via the UI

**How?**
```go
// boatmanmode/integration.go
cmd := exec.CommandContext(ctx, boatmanmodePath,
    "execute",
    "--ticket", ticketID,
    "--repo", repoPath,
)
```

### Message Routing

`StreamExecution()` accepts a `MessageCallback` that routes messages through the desktop session:

```go
type MessageCallback func(role, content string)

func (i *Integration) StreamExecution(ctx context.Context, sessionID, input, mode string,
    outputChan chan<- string, onMessage MessageCallback) (map[string]interface{}, error)
```

The callback handles two roles:
- `"assistant"` — regular CLI output text, routed via `session.AddBoatmanMessage()`
- `"claude_stream"` — raw Claude stream-json lines, routed via `session.ProcessExternalStreamLine()`

This means all boatmanmode messages flow through the session's message system with proper agent attribution, appearing in the chat UI via the `agent:message` Wails channel.

## Usage from BoatmanApp

### 1. Execute a Ticket (Simple)

```go
result, err := ExecuteLinearTicketWithBoatmanMode(
    linearAPIKey,
    ticketID,
    projectPath,
)
```

**What happens:**
1. Calls `boatman execute --ticket TICKET_ID --repo REPO_PATH`
2. BoatmanMode:
   - Fetches ticket from Linear
   - Creates isolated git worktree
   - Plans implementation (parallel agent)
   - Validates plan (preflight agent)
   - Implements code changes
   - Runs tests
   - Peer reviews code (using `peer-review` skill)
   - Refactors based on feedback (loop until pass)
   - Creates PR with `gh`
   - Updates Linear ticket
3. Returns result (success/failure, PR URL, test results)

### 2. Execute with Streaming Output

```go
err := StreamLinearTicketExecution(
    linearAPIKey,
    ticketID,
    projectPath,
)

// Messages now flow through agent:message (primary channel):
EventsOn("agent:message", (data) => {
    // Messages include metadata.agent with agentId, agentType, status
    addMessage(data.sessionId, data.message);
})

// Structured events for task tracking:
EventsOn("boatmanmode:event", (data) => {
    HandleBoatmanModeEvent(data.sessionId, data.event.type, data.event);
})

// boatmanmode:output is now a fallback (no-op logger):
EventsOn("boatmanmode:output", (data) => {
    console.log("Fallback:", data.message);
})
```

**What happens:**
- Same as above, but streams real-time output to the UI
- User can watch the execution progress live
- Shows all agent activity: file reads, edits, test runs, peer review feedback
- Each boatmanmode phase creates a separate agent tab in the AgentLogsPanel
- Claude's raw stream events are forwarded via `claude_stream` for full visibility
- Messages are attributed to the correct agent and appear in the chat UI

### 3. Fetch Linear Tickets

```go
tickets, err := FetchLinearTicketsForBoatmanMode(
    linearAPIKey,
    projectPath,
)
```

**What happens:**
1. Calls `boatman list-tickets --repo REPO_PATH --labels firefighter,triage,boatmanmode`
2. Returns list of tickets suitable for automated execution
3. UI displays them in the ticket list

## Configuration

### BoatmanMode Binary Location

The integration looks for the `boatman` binary in two places:

1. **In PATH**: `exec.LookPath("boatman")`
2. **Default location**: `~/workspace/handshake/boatmanmode/boatman`

**For production**, ensure `boatman` is in the PATH:

```bash
# Option 1: Add to PATH
export PATH="$PATH:~/workspace/handshake/boatmanmode"

# Option 2: Create symlink
ln -s ~/workspace/handshake/boatmanmode/boatman /usr/local/bin/boatman

# Option 3: Install globally
cd ~/workspace/handshake/boatmanmode
go install ./cmd/boatmanmode
```

### API Keys

BoatmanApp passes API keys to boatmanmode via environment variables (set by the CLI flags).

**Required:**
- `LINEAR_API_KEY` - from user preferences or FirefighterSettings
- `CLAUDE_API_KEY` - from user preferences

## User Experience

### Firefighter Mode (Manual Investigation)

**Current behavior:**
1. User clicks "Firefighter" button
2. Sees Linear triage queue in sidebar
3. Clicks "Investigate" on a ticket
4. Firefighter agent (in chat):
   - Fetches Bugsnag errors
   - Queries Datadog logs
   - Analyzes git history
   - Generates investigation report
   - Updates Linear ticket
5. User reviews report and decides next steps

**This is useful for:**
- Understanding what went wrong
- Root cause analysis
- Creating incident reports
- Manual fixes

### BoatmanMode Execution (Automated Fix)

**New behavior:**
1. User clicks "Auto-Execute with BoatmanMode" button on ticket
2. BoatmanMode (subprocess):
   - Creates worktree
   - Plans fix
   - Implements code
   - Runs tests
   - Peer reviews
   - Refactors until passing
   - Creates PR
3. User reviews the PR

**This is useful for:**
- Straightforward bug fixes
- Well-defined feature requests
- Tickets with clear acceptance criteria
- Reducing manual coding work

### When to Use Which?

| Use Case | Tool | Why |
|----------|------|-----|
| "What caused this production error?" | Firefighter Mode | Investigates across systems |
| "Fix this straightforward bug" | BoatmanMode | Automates the fix |
| "Unclear what's broken" | Firefighter Mode | Gathers context first |
| "Implement feature XYZ" | BoatmanMode | Writes code automatically |
| "Need postmortem report" | Firefighter Mode | Generates investigation docs |
| "Simple refactor task" | BoatmanMode | Automates implementation |

## UI Integration

### FirefighterTicketList Component

Add an "Auto-Execute" button next to the "Investigate" button:

```tsx
<button onClick={() => handleAutoExecute(ticket)}>
  <PlayCircle className="w-3 h-3" />
  Auto-Execute with BoatmanMode
</button>
```

### Message Flow

Boatmanmode messages now flow through the session message system with proper agent attribution:

1. **`agent:message`** (primary) - All session messages including boatmanmode output. Messages include `metadata.agent` with `agentId`, `agentType`, `description`, and `status`.
2. **`boatmanmode:event`** - Structured events for task tracking (agent_started, agent_completed, etc.)
3. **`boatmanmode:output`** (fallback) - Raw text output, now a no-op logger since messages route through `agent:message`

### AgentLogsPanel

The `AgentLogsPanel` component groups messages by agent. Boatmanmode phases (Execution, Planning, Review, Refactor, Testing) each get their own tab with color-coded labels:
- **Execution**: Orange
- **Planning**: Yellow
- **Review**: Pink
- **Refactor**: Indigo
- **Testing**: Teal

Each tab shows a status indicator:
- Pulsing cyan dot = active
- Solid green dot = completed

### MessageBubble Agent Badges

When a message is attributed to a non-main agent, a purple badge appears next to "Claude" in the message header showing the agent type and description (truncated to 30 characters).

## Future Enhancements

### 1. Interactive Mode

Allow user to:
- Pause execution
- Provide feedback during peer review
- Adjust plan before implementation
- Cancel if going wrong direction

### 2. Batch Execution

Execute multiple tickets in parallel:
```
Queue: TICKET-123, TICKET-124, TICKET-125
Execute all in separate worktrees
Show progress for each
```

### 3. Custom Workflows

Let users configure:
- Which peer review skill to use
- Test frameworks
- PR template
- Commit message format

### 4. Cost Tracking

Show estimated token usage before execution:
```
This ticket will use approximately:
- Planning: 5K tokens (~$0.10)
- Implementation: 20K tokens (~$0.40)
- Review: 10K tokens (~$0.20)
Total: ~$0.70
```

### 5. Execution History

Track all automated executions:
- Success rate
- Average time
- Common failure reasons
- Cost per ticket

## Architecture Diagram

```
┌──────────────────────────────────────────────────────────────────────┐
│                           BoatmanApp (UI)                            │
│                                                                      │
│  ┌──────────────────┐              ┌──────────────────┐             │
│  │   Firefighter    │              │   BoatmanMode    │             │
│  │      Mode        │              │   Integration    │             │
│  │                  │              │                  │             │
│  │ • Investigate    │              │ • Auto-Execute   │             │
│  │ • Report         │              │ • Stream Output  │             │
│  │ • Manual Fix     │              │ • Fetch Tickets  │             │
│  └────────┬─────────┘              └────────┬─────────┘             │
│           │                                 │                       │
│           │ MCP Tools                       │ Subprocess            │
│           │ (Bugsnag,                       │ exec.Command          │
│           │  Datadog,                       ▼                       │
│           │  Linear)              ┌──────────────────┐              │
│           │                       │  boatman CLI     │              │
│           │                       │  (boatmanmode)   │              │
│           │                       └────────┬─────────┘              │
│           │                                │                        │
│           │                   JSON events to stdout                 │
│           │                   (agent_started, claude_stream, etc.)  │
│           │                                │                        │
│           │                                ▼                        │
│           │                       ┌──────────────────┐              │
│           │                       │ integration.go   │              │
│           │                       │ Parses events    │              │
│           │                       │ + MessageCallback│              │
│           │                       └────────┬─────────┘              │
│           │                                │                        │
│           │                ┌───────────────┼───────────────┐        │
│           │                │               │               │        │
│           │                ▼               ▼               ▼        │
│           │         boatmanmode:     app.go routes   Session routes  │
│           │         event (Wails)    to session      messages via    │
│           │                         agent methods    agent:message   │
│           │                                                         │
│           │                         RegisterBoatmanAgent()          │
│           │                         SetCurrentAgent()               │
│           │                         AddBoatmanMessage()             │
│           │                         CompleteCurrentAgent()           │
│           │                         ProcessExternalStreamLine()     │
│           │                                │                        │
│           │                                ▼                        │
│           │                       ┌──────────────────┐              │
│           │                       │  Frontend        │              │
│           │                       │  AgentLogsPanel  │              │
│           │                       │  MessageBubble   │              │
│           │                       │  (agent badges)  │              │
│           │                       └──────────────────┘              │
└───────────┼─────────────────────────────────────────────────────────┘
            │
            ▼
    ┌──────────────┐                 ┌──────────────┐
    │   Bugsnag    │                 │    Linear    │
    │   Datadog    │                 │     API      │
    │     APIs     │                 │              │
    └──────────────┘                 └──────┬───────┘
                                            │
                                            ▼
                                    ┌──────────────┐
                                    │  Git Worktree│
                                    │  Claude API  │
                                    │  GitHub API  │
                                    └──────────────┘
```

## Development

### Testing Integration Locally

1. **Build boatmanmode:**
   ```bash
   cd ~/workspace/handshake/boatmanmode
   go build -o boatman ./cmd/boatmanmode
   ```

2. **Build boatmanapp:**
   ```bash
   cd /Users/pmiddleton/workspace/personal/boatmanapp
   wails dev
   ```

3. **Test execution:**
   - Create a test Linear ticket
   - Add labels: "firefighter", "boatmanmode"
   - Click "Auto-Execute with BoatmanMode" in UI
   - Watch live output

### Adding New BoatmanMode Commands

When boatmanmode adds new CLI commands, update the integration:

```go
// boatmanmode/integration.go

// Add new method
func (i *Integration) YourNewCommand(ctx context.Context, args...) error {
    cmd := exec.CommandContext(ctx, i.boatmanmodePath,
        "your-command",
        "--arg", value,
    )
    // ...
}
```

Then expose it in `app.go`:

```go
func (a *App) YourNewCommand(args...) error {
    bmIntegration, err := bmintegration.NewIntegration(...)
    return bmIntegration.YourNewCommand(ctx, args...)
}
```

## Troubleshooting

### "boatman binary not found"

```bash
# Check if in PATH
which boatman

# If not, add to PATH or use full path
export PATH="$PATH:~/workspace/handshake/boatmanmode"
```

### "execution failed"

Check boatmanmode logs:
```bash
tail -f ~/.boatman/logs/execution.log
```

### "no output streaming"

Ensure boatmanmode supports `--stream` flag or implement polling:
```go
// Poll for status instead of streaming
ticker := time.NewTicker(2 * time.Second)
for range ticker.C {
    status := checkStatus()
    emit("boatmanmode:output", status)
}
```

## Summary

**BoatmanMode integration gives BoatmanApp users:**
- ✅ Automated ticket execution (not just investigation)
- ✅ Full implementation workflow (plan → code → test → review → PR)
- ✅ Live progress streaming
- ✅ Choice between manual investigation vs automated fix
- ✅ Best of both worlds: UI convenience + CLI automation power

The integration is **clean** (subprocess), **flexible** (no version coupling), and **user-friendly** (streaming output, real-time feedback).
