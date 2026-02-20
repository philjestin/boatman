# Boatman Ecosystem

A monorepo containing the Boatman CLI tool and desktop application for AI-powered autonomous software development.

> 🆕 **[What's New in February 2026](./WHATS_NEW.md)** - See recent enhancements: monorepo architecture, advanced search, batch diff approval, BoatmanMode integration, agent logs, and more!

## Repository Structure

```
boatman-ecosystem/
├── cli/              # Boatmanmode CLI - Autonomous development agent
│   ├── cmd/          # CLI entry points
│   ├── internal/     # CLI implementation
│   └── README.md     # CLI-specific documentation
│
├── desktop/          # Boatman Desktop - GUI wrapper for CLI
│   ├── frontend/     # React/TypeScript UI
│   ├── agent/        # Go backend for desktop app
│   └── wailsjs/      # Wails bindings
│
├── go.work           # Go workspace configuration
└── README.md         # This file
```

## What is Boatman?

Boatman is an AI-powered autonomous development system that:
- Takes tasks or prompts and implements them end-to-end
- Creates isolated git worktrees for safe experimentation
- Plans, executes, reviews, and refactors code automatically
- Integrates with Linear for ticket-based workflows
- Provides both CLI and desktop GUI interfaces

## Components

### CLI (`./cli`)

The command-line interface and core autonomous agent.

**Key features:**
- Autonomous multi-phase workflow (plan → execute → review → refactor)
- Git worktree isolation for safe parallel work
- Claude AI integration for code generation and review
- Structured JSON event emission for integration
- Linear ticket integration
- **Event metadata** including diffs, feedback, plans, and issues
- **Public utilities** in `pkg/` for desktop integration (diff analysis, validation)
- **Automatic environment filtering** for secure nested Claude sessions
- **Multiple input modes**: Linear tickets, inline prompts, or file-based tasks

**Quick start:**
```bash
cd cli
go build -o boatman ./cmd/boatman
./boatman work --prompt "Add user authentication"
```

See [cli/README.md](./cli/README.md) for detailed documentation.

### Desktop (`./desktop`)

A cross-platform desktop application built with Wails that provides a GUI for the CLI.

**Key features:**
- Visual task tracking with clickable phase details
- Real-time streaming of agent execution with per-agent attribution
- Subagent tracking in standard mode (Task/Explore agents get separate tabs)
- Claude event streaming in boatmanmode (full visibility into each phase)
- Session management and history
- Project-based organization
- Event-driven architecture with live updates
- **Smart search** with filters (tags, dates, favorites, projects)
- **Batch diff approval** for reviewing multiple file changes at once
- **Inline diff comments** for code review discussions
- **BoatmanMode integration** for autonomous Linear ticket execution
- **Firefighter mode** for production incident investigation
- **Agent logs panel** for real-time visibility into AI actions
- **Onboarding wizard** for first-time setup
- **MCP server management** via UI dialog

**Quick start:**
```bash
cd desktop
wails dev
```

See [desktop/README.md](./desktop/README.md) for detailed documentation.

## Development

### Prerequisites

- Go 1.24.1 or later
- Node.js 18+ (for desktop frontend)
- Wails v2 (for desktop app)
- Claude CLI or Anthropic API key

### Setting up the workspace

This repository uses a Go workspace to manage multiple modules:

```bash
# Clone the repository
git clone <repo-url> boatman-ecosystem
cd boatman-ecosystem

# The go.work file automatically configures the workspace
go work sync
```

### Building

**CLI only:**
```bash
cd cli
go build -o boatman ./cmd/boatman
```

**Desktop only:**
```bash
cd desktop
wails build
```

**Both:**
```bash
# Build CLI first (desktop bundles it)
cd cli && go build -o boatman ./cmd/boatman

# Build desktop
cd ../desktop && wails build
```

### Testing

**CLI tests:**
```bash
cd cli
go test ./...
```

**Desktop tests:**
```bash
cd desktop
go test ./...
cd frontend && npm test
```

## Integration

The CLI and desktop communicate via structured JSON events:

```json
{"type": "agent_started", "id": "execute-ENG-123", "name": "Execution", "description": "Implementing code"}
{"type": "claude_stream", "id": "executor", "message": "{\"type\":\"content_block_delta\",...}"}
{"type": "agent_completed", "id": "execute-ENG-123", "name": "Execution", "status": "success"}
```

Events are emitted by the CLI to stdout and captured by the desktop app for real-time UI updates.

**Event types:**
- `agent_started` / `agent_completed` — Track each workflow phase
- `progress` — General status updates
- `claude_stream` — Raw Claude stream-json lines for full visibility into Claude's work
- `task_created` / `task_updated` — Task lifecycle events

The desktop app routes these events through the session message system with proper agent attribution, so each workflow phase gets its own tab in the Agent Logs panel.

## Why a Monorepo?

The CLI and desktop are tightly coupled:
- They share a JSON event protocol that evolves together
- Desktop is essentially a GUI wrapper for the CLI
- Features often span both components (e.g., adding metadata to events)
- Single source of truth for integration contracts
- Atomic commits for cross-cutting changes
- **Hybrid architecture**: Desktop can use subprocess OR direct imports

## Recent Enhancements

### Monorepo Architecture (February 2026)
The project has been restructured into a unified monorepo with shared types and utilities:
- **Shared event types**: CLI and desktop use common event definitions
- **Public utilities**: CLI exposes `pkg/` for desktop integration
- **Hybrid architecture**: Desktop can use subprocess OR direct imports
- **Independent versioning**: Components version independently (`cli/v1.x.x`, `desktop/v1.x.x`)

### Desktop UI Improvements
- ✅ **Advanced search** with filters for tags, favorites, dates, and projects
- ✅ **Batch diff approval** for efficient code review
- ✅ **Inline diff comments** for collaborative review discussions
- ✅ **Agent logs panel** for real-time visibility into AI tool usage
- ✅ **Task detail modals** showing diffs, plans, feedback, and issues
- ✅ **Onboarding wizard** for guided first-time setup
- ✅ **Session favorites and tags** for better organization

### CLI Event System
- ✅ **Rich event metadata**: Events now include diffs, feedback, plans, and issues
- ✅ **Structured task tracking**: External tools can track agent progress via events
- ✅ **Environment security**: Automatic filtering of sensitive env vars in nested sessions

### BoatmanMode Integration
- ✅ **Seamless ticket execution**: Click button, enter ticket ID, watch autonomous workflow
- ✅ **Real-time event streaming**: See planning, execution, review, and refactor phases
- ✅ **Task visibility**: All agent steps tracked as tasks with full metadata
- ✅ **Purple session badges**: Visual distinction for boatmanmode sessions

## Hybrid Architecture

The desktop app uses a **hybrid approach** for maximum flexibility:

### Subprocess (Traditional)
- User-facing execution with streaming output
- Process isolation and control (can kill/restart)
- Full boatmanmode workflows

### Direct Import (New)
- Fast UI queries (diff analysis, validation)
- Type-safe operations
- No subprocess overhead

**Example:**
```go
// Fast diff stats for task detail modal (direct import)
import "github.com/philjestin/boatmanmode/pkg/diff"
hybrid := services.NewHybrid(projectPath)
stats := hybrid.GetDiffStats(diff)

// Full execution with streaming + session routing (subprocess)
onMessage := func(role, content string) {
    session.AddBoatmanMessage(role, content)
}
bmIntegration.StreamExecution(ctx, sessionID, prompt, "prompt", outputChan, onMessage)
```

See [HYBRID_ARCHITECTURE.md](./HYBRID_ARCHITECTURE.md) for details.

## Release Strategy

Components use **independent versioning** within the monorepo:

- **CLI**: `cli/v1.2.3` - Standalone releases for terminal users
- **Desktop**: `desktop/v1.0.5` - Bundles CLI, can release independently

### Quick Release

```bash
# Interactive wizard (recommended)
make release

# Or manually
make bump-cli-minor        # Bump CLI version
vim cli/CHANGELOG.md       # Update changelog
git commit -am "cli: Release v1.1.0"
git tag cli/v1.1.0
git push origin main --tags
```

GitHub Actions automatically:
- Builds binaries for all platforms
- Creates GitHub releases
- Uploads artifacts

See [RELEASE_SUMMARY.md](./RELEASE_SUMMARY.md) for quick reference or [RELEASES.md](./RELEASES.md) for the complete guide.

## Documentation

### Getting Started
- **[What's New](./WHATS_NEW.md)** - Recent enhancements and features (February 2026)
- **[Quickstart Guide](./QUICKSTART.md)** - Get up and running in 5 minutes
- **[Main README](./README.md)** - This file, project overview

### CLI Documentation
- **[CLI README](./cli/README.md)** - Complete CLI documentation
- **[Task Modes](./cli/TASK_MODES.md)** - Linear tickets, prompts, and file-based tasks
- **[Events](./cli/EVENTS.md)** - Event system specification
- **[Library Usage](./cli/LIBRARY_USAGE.md)** - Using CLI as a Go library
- **[CLI Changelog](./cli/CHANGELOG.md)** - CLI version history

### Desktop Documentation
- **[Desktop README](./desktop/README.md)** - Desktop app overview
- **[Features Guide](./desktop/FEATURES.md)** - Comprehensive feature documentation
- **[Getting Started](./desktop/GETTING_STARTED.md)** - Desktop setup and usage
- **[BoatmanMode Integration](./desktop/BOATMANMODE_INTEGRATION.md)** - Autonomous execution
- **[BoatmanMode Events](./desktop/BOATMANMODE_EVENTS.md)** - Event specification
- **[Desktop Changelog](./desktop/CHANGELOG.md)** - Desktop version history

### Architecture & Development
- **[Hybrid Architecture](./HYBRID_ARCHITECTURE.md)** - Subprocess vs direct imports
- **[Contributing](./CONTRIBUTING.md)** - Development guidelines
- **[Releases](./RELEASES.md)** - Complete release guide
- **[Release Summary](./RELEASE_SUMMARY.md)** - Quick reference
- **[Versioning](./VERSIONING.md)** - Independent component versioning

## Contributing

1. Make changes in the appropriate directory (`cli/` or `desktop/`)
2. Run tests in both if you modify the event protocol
3. Update documentation in component READMEs and relevant guides
4. Create a PR with a clear description of changes

## License

[Your License Here]

## Architecture Diagrams

### Event Flow

```
┌─────────────────┐
│   CLI Agent     │
│  (boatmanmode)  │
└────────┬────────┘
         │ Emits JSON events to stdout
         │ agent_started, claude_stream, agent_completed, progress
         ▼
┌─────────────────┐
│  Integration    │ (desktop/boatmanmode/integration.go)
│    Layer        │ Captures stdout, parses JSON
│                 │ Calls MessageCallback for session routing
└────────┬────────┘
         │
    ┌────┴─────────────────────────┐
    │                              │
    ▼                              ▼
┌─────────────┐          ┌─────────────────────────┐
│ Wails emit  │          │ app.go → Session methods │
│ boatmanmode │          │ RegisterBoatmanAgent()   │
│ :event      │          │ SetCurrentAgent()        │
│             │          │ AddBoatmanMessage()      │
│             │          │ ProcessExternalStream    │
│             │          │ Line() (claude_stream)   │
└──────┬──────┘          └───────────┬──────────────┘
       │                             │
       │                             │ agent:message (Wails)
       ▼                             ▼
┌──────────────────────────────────────┐
│   Desktop UI (React)                 │
│   useAgent.ts handles events         │
│   AgentLogsPanel groups by agent     │
│   MessageBubble shows agent badges   │
│   TaskDetail shows phase info        │
└──────────────────────────────────────┘
```

### Directory Structure Philosophy

- **`cli/`**: Core autonomous agent logic, no GUI dependencies
- **`desktop/`**: Presentation layer that consumes CLI events
- **`go.work`**: Workspace ties them together for development
- Each component maintains its own `go.mod` for independent versioning
