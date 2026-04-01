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
- **Triages backlogs** — scores and classifies Linear tickets by AI-readiness, clusters related work, and generates validated execution plans
- **Executes tasks end-to-end** — takes tickets, prompts, or files and implements them autonomously with planning, execution, code review, and refactoring
- **Creates isolated git worktrees** for safe parallel work
- **Reviews its own code** via ScottBott peer review with iterative refactoring
- **Creates draft PRs as safety checkpoints** so work is preserved even if later stages fail
- **Supports resume** — pick up a failed execution from the review/refactor stage without re-doing the work
- Integrates with Linear, GitHub, and provides both CLI and desktop GUI interfaces

## Components

### CLI (`./cli`)

The command-line interface and core autonomous agent.

**Key features:**
- **9-step work pipeline**: prepare → worktree → plan → validate → execute → draft PR → test/review → refactor → finalize PR
- **Backlog triage pipeline**: fetch → score (7-dimension rubric) → classify (deterministic gates) → cluster → plan (optional)
- Git worktree isolation for safe parallel work
- Claude AI integration for code generation and review
- Draft PR safety checkpoints — work is preserved even if review/refactor fails
- Resume failed executions from the review/refactor stage
- Generated file filtering (protobuf, GraphQL codegen, Wails bindings)
- Structured JSON event emission for desktop integration
- Linear ticket integration
- Multiple input modes: Linear tickets, inline prompts, or file-based tasks

**Quick start:**
```bash
cd cli
go build -o boatman ./cmd/boatman

# Execute a task
./boatman work --prompt "Add user authentication"

# Triage a backlog
./boatman triage --teams EMP --states backlog --limit 20

# Resume a failed execution
./boatman work EMP-1234 --resume
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
- **BoatmanMode integration** for autonomous ticket execution with draft PR checkpoints and resume
- **Triage mode** for scoring, classifying, and planning entire backlogs
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

**BoatmanMode event types:**
- `agent_started` / `agent_completed` — Track each workflow phase
- `progress` — General status updates
- `claude_stream` — Raw Claude stream-json lines for full visibility into Claude's work
- `task_created` / `task_updated` — Task lifecycle events

**Triage event types:**
- `triage_started` / `triage_complete` — Pipeline lifecycle
- `triage_ticket_scoring` / `triage_ticket_scored` — Per-ticket scoring progress
- `triage_classifying` / `triage_clustering` — Stage transitions
- `plan_started` / `plan_ticket_planned` / `plan_ticket_validated` / `plan_complete` — Plan generation

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
- **Seamless ticket execution**: Click button, enter ticket ID, watch autonomous workflow
- **9-step pipeline**: plan → validate → execute → draft PR → test → review → refactor → finalize PR
- **Draft PR safety checkpoint**: Work is preserved as a draft PR before review/refactor starts
- **Resume failed runs**: Pick up from review/refactor without re-executing
- **Generated file filtering**: Protobuf, GraphQL codegen, and Wails bindings excluded from review prompts
- **Real-time event streaming**: See each phase with per-agent attribution
- **Purple session badges**: Visual distinction for boatmanmode sessions

### Triage Mode
- **Backlog analysis**: Score and classify entire backlogs from Linear
- **7-dimension AI rubric**: clarity, codeLocality, patternMatch, validationStrength, dependencyRisk, productAmbiguity, blastRadius
- **Deterministic classification**: Hard stops (payments, auth, migrations) and threshold gates, no LLM guesswork
- **4 categories**: AI_DEFINITE, AI_LIKELY, HUMAN_REVIEW_REQUIRED, HUMAN_ONLY
- **Ticket clustering**: Groups related tickets by shared domains, files, and dependencies
- **Plan generation**: Claude explores the repo and generates validated execution plans with candidate files and test commands
- **4-gate plan validation**: files exist, within scope, stop conditions present, valid test runners
- **Execute from triage**: One click to create a BoatmanMode session with the pre-generated plan

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
- **[Triage Mode](./desktop/TRIAGE.md)** - Backlog triage pipeline, scoring rubric, classification, plan generation
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

### BoatmanMode Work Pipeline

```
boatman work EMP-1234
  │
  ├─ Step 1: Prepare Task        (display info)
  ├─ Step 2: Setup Worktree      (git worktree add)
  ├─ Step 3: Planning            (Claude analyzes codebase, or use --plan-file)
  ├─ Step 4: Preflight           (validate plan, pin context files)
  ├─ Step 5: Execute             (Claude implements with tools)
  ├─ Step 5b: Draft PR           (safety checkpoint: commit + push + draft PR)
  ├─ Step 6: Test & Review       (parallel: test runner + ScottBott review)
  ├─ Step 7: Refactor Loop       (iterate: review → refactor → verify)
  ├─ Step 8: Commit & Push       (final reviewed changes)
  └─ Step 9: Finalize PR         (update body, mark ready)

If review fails: draft PR preserved with work intact
If execution hangs: --resume picks up from Step 6
```

### Triage Pipeline

```
boatman triage --teams EMP --states backlog
  │
  ├─ Stage 0: Fetch              (Linear API → raw tickets)
  ├─ Stage 1: Ingest             (normalize, extract signals)
  ├─ Stage 2a: Score             (Claude: 7-dimension rubric, concurrent)
  ├─ Stage 2b: Classify          (deterministic decision tree)
  ├─ Stage 3: Cluster            (signal-overlap grouping + context docs)
  └─ Stage 4: Plan (optional)    (Claude explores repo → validated plans)
        │
        └─ Execute from triage → creates BoatmanMode session with plan
```

### Event Flow

```
┌──────────────────┐        ┌──────────────────┐
│   CLI Agent      │        │  CLI Triage      │
│  (boatman work)  │        │ (boatman triage) │
└────────┬─────────┘        └────────┬─────────┘
         │                           │
         │  JSON events to stdout    │
         ▼                           ▼
┌──────────────────┐        ┌──────────────────┐
│  BoatmanMode     │        │  Triage          │
│  Integration     │        │  Integration     │
│  (subprocess)    │        │  (subprocess)    │
└────────┬─────────┘        └────────┬─────────┘
         │                           │
         │  Wails events             │
         ▼                           ▼
┌──────────────────────────────────────────────┐
│   Desktop UI (React)                         │
│                                              │
│   ChatView     TriageView                    │
│   ├─ Messages  ├─ ResultsTable (sortable)    │
│   ├─ Resume    ├─ ClusterView (grouped)      │
│   └─ AgentLogs └─ PlanView (validated)       │
└──────────────────────────────────────────────┘
```

### Directory Structure

```
boatman/
├── cli/
│   ├── cmd/boatman/         # CLI entry point
│   └── internal/
│       ├── agent/           # 9-step work pipeline + resume
│       ├── triage/          # Triage pipeline (ingest, score, classify, cluster)
│       ├── plan/            # Plan generation + 4-gate validation
│       ├── executor/        # Claude execution + generated file filtering
│       ├── scottbott/       # Code review agent
│       ├── planner/         # Codebase analysis + plan generation
│       ├── testrunner/      # Test detection + execution
│       ├── worktree/        # Git worktree management
│       ├── github/          # PR creation (draft + ready + update)
│       ├── events/          # JSON event emission
│       ├── brain/           # Domain knowledge extraction
│       ├── contextpin/      # Multi-file context coordination
│       ├── coordinator/     # Agent message bus + work claiming
│       ├── cost/            # Token usage tracking
│       ├── diffverify/      # Refactor verification
│       ├── handoff/         # Structured inter-agent context
│       └── cli/             # Cobra command definitions
│
├── desktop/
│   ├── app.go               # Wails app (sessions, streaming, triage)
│   ├── agent/               # Session management + persistence
│   ├── boatmanmode/          # CLI subprocess integration
│   ├── triage/               # Triage subprocess integration
│   └── frontend/src/
│       ├── components/
│       │   ├── chat/         # ChatView, MessageBubble, Resume banner
│       │   └── triage/       # TriageView, ResultsTable, PlanView, etc.
│       ├── hooks/            # useAgent.ts (session state management)
│       └── types/            # TypeScript type definitions
│
└── go.work                   # Go workspace configuration
```
