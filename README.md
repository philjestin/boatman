# Boatman Ecosystem

A monorepo containing the Boatman CLI tool and desktop application for AI-powered autonomous software development.

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
- Real-time streaming of agent execution
- Session management and history
- Project-based organization
- Event-driven architecture with live updates

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
{
  "type": "agent_completed",
  "id": "execution-task-123",
  "name": "Execution",
  "status": "success",
  "data": {
    "diff": "...",
    "feedback": "...",
    "issues": [...]
  }
}
```

Events are emitted by the CLI to stdout and captured by the desktop app for real-time UI updates.

## Why a Monorepo?

The CLI and desktop are tightly coupled:
- They share a JSON event protocol that evolves together
- Desktop is essentially a GUI wrapper for the CLI
- Features often span both components (e.g., adding metadata to events)
- Single source of truth for integration contracts
- Atomic commits for cross-cutting changes

## Release Strategy

While developed together, the components can be released independently:

- **CLI**: Standalone releases for terminal users
- **Desktop**: Bundles the CLI internally, synchronized releases

## Contributing

1. Make changes in the appropriate directory (`cli/` or `desktop/`)
2. Run tests in both if you modify the event protocol
3. Update documentation in component READMEs
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
         │ {"type": "agent_completed", "data": {...}}
         ▼
┌─────────────────┐
│  Integration    │ (desktop/boatmanmode/integration.go)
│    Layer        │ Captures stdout, parses JSON
└────────┬────────┘
         │ Wails EventsEmit
         │ runtime.EventsEmit(ctx, "boatmanmode:event", ...)
         ▼
┌─────────────────┐
│   Desktop UI    │ (desktop/frontend/src/hooks/useAgent.ts)
│   React App     │ EventsOn('boatmanmode:event', handler)
└─────────────────┘
         │ Updates task metadata
         │ task.metadata = { diff, feedback, plan, ... }
         ▼
┌─────────────────┐
│  TaskDetail     │ Displays clickable task details
│     Modal       │ with diffs, feedback, issues
└─────────────────┘
```

### Directory Structure Philosophy

- **`cli/`**: Core autonomous agent logic, no GUI dependencies
- **`desktop/`**: Presentation layer that consumes CLI events
- **`go.work`**: Workspace ties them together for development
- Each component maintains its own `go.mod` for independent versioning
