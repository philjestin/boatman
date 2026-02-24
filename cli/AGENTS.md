# AGENTS.md

## Repository Overview

This is a **single-package repository**.
Primary language(s): go.
CI: github_actions.

### Architecture


## BoatmanMode CLI Architecture

**BoatmanMode** is a Go-based CLI tool that orchestrates AI-powered development workflows. It automates ticket execution by fetching tasks from Linear, generating code using Claude (via Vertex AI), performing peer review with configurable Claude skills, and creating pull requests via GitHub CLI.

### Core Systems

1. **Workflow Engine**: Orchestrates a 9-step pipeline (fetch ticket → create worktree → plan → validate → execute → review → refactor loop → verify diff → create PR). Supports multiple input modes: Linear tickets, inline prompts, or file-based tasks.

2. **Agent Coordination**: A central coordinator manages parallel agent execution (preflight validator, planner, executor, reviewer, refactor, test runner, diff verifier). Agents communicate via channels with work claiming and file locking to prevent conflicts.

3. **Claude Integration**: Each agent type can use a different Claude model (Opus, Sonnet, Haiku) configured per-agent in `.boatman.yaml`. Supports prompt caching for 50-90% cost reduction on iterative refactors. Communicates via `claude` CLI with Vertex AI backend.

4. **External Integrations**: Linear GraphQL API for ticket fetching, GitHub CLI (`gh`) for PR creation, Git worktrees for isolation, tmux for live activity streaming.

5. **Event System**: Emits structured JSON events to stdout (`agent_started`, `agent_completed`, `progress`, `claude_stream`, `task_created`, `task_updated`) for desktop app integration and real-time monitoring.

6. **Support Systems**: Git-integrated checkpoints for rollback, context pinning with checksums, dynamic handoff compression (4 levels), smart file summarization (language-aware), issue deduplication across review iterations, and cross-session agent memory.

### Data Flow

Configuration (`.boatman.yaml`, env vars) → CLI entry point (`cmd/boatman`) → Workflow engine fetches task → Coordinator spawns agents → Agents invoke Claude CLI → Results flow back through coordinator → Git operations (worktree, commits, PR) → Events emitted to stdout.


## Repository Map

Single package at repository root.

## Commands

### Quick Reference

| Action | Command |
|--------|---------|
| build | `go build ./...` |
| test | `go test ./...` |
| format | `gofmt -w .` |

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

## Sources of Truth

These files are the authoritative source for their respective concerns:

- **CI configuration**: [`.github/workflows/auto-version.yml`](.github/workflows/auto-version.yml)
- **CI configuration**: [`.github/workflows/release.yml`](.github/workflows/release.yml)
- **CI configuration**: [`.github/workflows/test.yml`](.github/workflows/test.yml)
