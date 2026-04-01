# AGENTS.md

## Repository Overview

This is a **go monorepo** with **4 packages**.
Primary language(s): go.
CI: github_actions.

### Architecture


## Boatman Ecosystem Architecture

This is a **Go workspace monorepo** containing an AI-powered autonomous development system with two primary components: a CLI agent and a desktop GUI wrapper.

**Core Components:**
- **CLI (`./cli`)**: Autonomous development agent with two main pipelines:
  - **Work pipeline** (9 steps): prepare → worktree → plan → validate → execute → draft PR → test/review → refactor → finalize PR. Creates draft PRs as safety checkpoints. Supports `--resume` to pick up from review/refactor.
  - **Triage pipeline** (4-6 stages): fetch → ingest → score (7-dimension rubric) → classify (deterministic gates) → cluster → plan (optional). Scores tickets by AI-readiness, generates validated execution plans.
  - Uses Claude AI for code generation, creates isolated git worktrees, emits structured JSON events to stdout.
- **Desktop (`./desktop`)**: Cross-platform Wails application (Go backend + React/TypeScript frontend) that provides a GUI for both work and triage pipelines. Includes BoatmanMode (autonomous execution with resume), Triage Mode (backlog analysis with results/clusters/plans tabs), and Firefighter Mode (incident response).
- **Shared (`./shared`)**: Common event types and utilities shared between CLI and desktop.
- **Harness (`./harness`)**: Test harness, brain knowledge system, diff verification, context pinning, and cost tracking libraries.

**Data Flow:**
The CLI emits JSON events to stdout during execution. Two event families:
- **Work events**: `agent_started`, `claude_stream`, `agent_completed`, `progress`, `task_created`, `task_updated`
- **Triage events**: `triage_started`, `triage_ticket_scored`, `triage_complete`, `plan_started`, `plan_ticket_planned`, `plan_complete`

The desktop app captures these events via subprocess execution, routes them through a session message system, and updates the UI in real-time.

**Integration Pattern:**
Desktop uses a **hybrid approach**: subprocess execution for user-facing workflows with streaming output, and direct Go imports (`pkg/` utilities) for fast UI queries like diff analysis and validation. This provides both process isolation and type-safe operations.

**Key Patterns:**
- Event-driven architecture with structured JSON protocol
- Git worktree isolation for safe parallel development
- Linear ticket integration for task management
- Independent component versioning (`cli/v*`, `desktop/v*`) within monorepo
- Go workspace for unified development with separate modules


## Repository Map

| Package | Path | Type | Language |
|---------|------|------|----------|
| cli | `./cli` | - | go |
| desktop | `./desktop` | - | go |
| harness | `./harness` | - | go |
| shared | `./shared` | - | go |

## Commands

### Quick Reference

| Action | Command |
|--------|---------|
| lint | `make lint` |
| dev | `make dev` |
| clean | `make clean` |

### Agent Scripts

Use these wrapper scripts for deterministic, machine-readable output:

- `scripts/agent-test` — run tests with structured output
- `scripts/agent-lint` — run linter with structured output
- `scripts/agent-build` — run build and capture results

## Conventions & Constraints

### Do

- **Use gofmt** for formatting

### Don't

- Don't bypass linter rules without team discussion
- Don't add new dependencies without checking for existing alternatives
- no circular deps

## Sources of Truth

These files are the authoritative source for their respective concerns:

- **CI configuration**: [`.github/workflows/release-cli.yml`](.github/workflows/release-cli.yml)
- **CI configuration**: [`.github/workflows/release-desktop.yml`](.github/workflows/release-desktop.yml)

## Scoped Documentation

Sub-packages may have their own `AGENTS.md` with local context:


### Adding/Updating Scoped AGENTS.md

When creating a new package or modifying an existing one, add or update its
`AGENTS.md` with:
- Package-specific build/test/lint commands
- Local architectural constraints or gotchas
- Subsystem-specific pointers
