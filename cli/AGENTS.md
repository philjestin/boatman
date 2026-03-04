# AGENTS.md

## Repository Overview

This is a **single-package repository**.
Primary language(s): go.
CI: github_actions.

### Architecture


## BoatmanMode CLI Architecture

**BoatmanMode** is a Go-based CLI tool that orchestrates AI-powered development workflows. It automates ticket execution by fetching tasks from Linear, generating code using Claude (via Vertex AI), performing peer review with configurable Claude skills, and creating pull requests via GitHub CLI.

**Core workflow engine** (`internal/cli`) coordinates a multi-agent pipeline with 9 distinct steps: ticket fetch, worktree creation, planning (with preflight validation), code execution, test running, peer review, iterative refactoring (with diff verification), and PR creation. Each agent runs in isolated tmux sessions with real-time streaming visibility.

**Multi-mode input system** supports Linear tickets, inline prompts (`--prompt`), or file-based tasks (`--file`), all flowing through the same quality workflow with auto-generated task IDs and branch names.

**Event-driven integration layer** emits structured JSON events (`agent_started`, `progress`, `claude_stream`, etc.) to stdout for external desktop app integration and real-time monitoring.

**Resilience and observability** are built-in: exponential backoff retries for Linear/Claude APIs, health checks for dependencies (`git`, `gh`, `claude`, `tmux`), structured logging via `log/slog`, and git-integrated checkpoints for crash recovery.

**Modular architecture** separates concerns across packages: `platform` (shared utilities, Linear/GitHub clients), `harness` (test mocks and fixtures), and `shared` (common types). The CLI uses Cobra for command routing and Viper for configuration management, with support for multi-model Claude selection and prompt caching for cost optimization.


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

### Inferred Command Matrix

The following commands were identified by analyzing project configuration:

| Scope | Build | Test | Lint | Format | Dev |
|-------|-------|------|------|--------|-----|
| root | `go build ./...` | `go test ./...` | `go vet ./...` | `gofmt -w .` | - |

## Conventions & Constraints

### Do

- **Use gofmt** for formatting

### Don't

- Don't bypass linter rules without team discussion
- Don't add new dependencies without checking for existing alternatives

### Inferred Architectural Constraints

- **All CLI commands must be implemented under cmd/boatman/ as the single entrypoint** — The structural model identifies exactly one entrypoint: 'boatman' CLI at cmd/boatman. This follows Go's standard project layout for single-binary CLI tools. *(enforced via: Enforce via code review or a custom linter that checks for additional main packages outside cmd/boatman/)*
- **Use Cobra for command routing and Viper for configuration management** — go.mod explicitly imports github.com/spf13/cobra and github.com/spf13/viper, and main.go delegates to internal/cli.Execute(). This is the established pattern for CLI structure. *(enforced via: Dependency analysis: ensure no alternative CLI frameworks (urfave/cli, etc.) are introduced)*
- **Version information (version, commit, date, builtBy) must be injected at build time via ldflags** — main.go declares these as package-level vars with default values ('dev', 'none', etc.) and passes them to cli.SetVersionInfo(). GoReleaser typically injects these via -ldflags. *(enforced via: Verify .goreleaser.yml contains ldflags for -X main.version, -X main.commit, etc.)*
- **External integrations must emit structured JSON events to stdout for observability** — README documents an event system with specific event types (agent_started, progress, claude_stream, etc.) for desktop app integration. This is a core architectural contract. *(enforced via: Integration tests that parse stdout and validate event schema conformance)*
- **All agents must support structured logging via log/slog with configurable levels** — README explicitly mentions 'Structured logging via log/slog with levels (DEBUG, INFO, WARN, ERROR)' and BOATMAN_DEBUG=1 for verbose output. This is the observability standard. *(enforced via: Static analysis to ensure all logging uses slog package, not fmt.Println or log.Printf)*
- **Runtime dependencies (git, gh, claude, tmux) must be validated via health checks at startup** — README states 'Health checks verify git, gh, claude, tmux at startup' as part of resilience features. This prevents cryptic failures mid-workflow. *(enforced via: Unit test that mocks missing dependencies and verifies startup fails with clear error messages)*
- **Configuration must support both environment variables and YAML files with clear precedence** — README shows LINEAR_API_KEY env var and ~/.boatman.yaml file. Viper supports layered config, so precedence must be documented and consistent. *(enforced via: Integration test that sets config via multiple sources and validates precedence order)*
- **All API clients (Linear, Claude) must implement retry logic with exponential backoff** — README explicitly lists 'Retry logic with exponential backoff for Linear API and Claude CLI' under resilience features. This is critical for production reliability. *(enforced via: Unit tests for API clients that simulate transient failures and verify retry behavior)*
- **Git operations must use worktrees for isolation; never modify the main working directory** — README emphasizes 'Each ticket works in an isolated worktree' and 'No interference with your main working directory' as a core feature. This prevents data loss. *(enforced via: Integration tests that verify git worktree creation and cleanup; static analysis to flag direct operations on main worktree)*
- **Commit messages must follow conventional format for auto-versioning (breaking:, feat:, fix:)** — The auto-version.yml workflow parses commit messages for version bump hints using regex patterns for 'breaking|major', 'feat|feature|minor', defaulting to patch. *(enforced via: Git commit-msg hook or CI check using commitlint to validate conventional commit format)*
- **Each agent type must support configurable Claude model selection via claude.models config** — README documents per-agent model configuration (planner, executor, reviewer, refactor, preflight, test_runner) for cost optimization. This is a user-facing contract. *(enforced via: Schema validation for .boatman.yaml that ensures model IDs are valid (claude-opus-4-6, claude-sonnet-4-5, claude-haiku-4))*
- **All file operations during agent execution must use context pinning with checksums** — README describes 'Context Pinning' feature that 'Pins file contents with checksums' and 'Detects stale files during long operations' to ensure consistency. *(enforced via: Unit tests for file operations that verify checksum validation and stale file detection)*

## Known Risks & Ambiguities

- 🔴 **[dependency]** The project uses local replace directives for three sibling modules (harness, shared, platform) that are not present in the scanned repository. These are referenced as '../harness', '../shared', '../platform' in go.mod, suggesting this is part of a monorepo or multi-repo ecosystem. — *Clarify the repository structure: is this a monorepo with missing sibling directories, or should these be published modules? AI agents need access to these dependencies to build/test successfully.*
- 🟡 **[build]** No linting configuration files detected (.golangci.yml, etc.), but CI workflow references a 'lint' job. The exact linter and its configuration are unknown. — *Verify what linter is used in .github/workflows/test.yml and document the lint command. Consider adding a .golangci.yml for consistent linting.*
- 🟡 **[dependency]** Runtime dependencies on external CLI tools (claude, gh, git, tmux) are required but not version-pinned. Health checks verify presence but not version compatibility. — *Document minimum required versions for claude, gh, git, and tmux. Consider adding version checks to health validation.*
- 🟡 **[architecture]** The README describes extensive features (coordinator, checkpoints, memory, event system, etc.) but only cmd/boatman/main.go is provided. The actual implementation in internal/cli and other packages is not visible. — *Provide representative files from internal/cli, internal/coordinator, internal/agents, etc. to validate architectural claims and understand code organization.*
- 🟡 **[test]** Test frameworks are listed as null in the structural model, but README mentions E2E test environment with mocks. The test setup and how to run E2E tests is unclear. — *Document test organization: unit tests vs E2E tests, how to run E2E tests with mocks, and any required test fixtures or environment setup.*
- ℹ️ **[ci]** Auto-versioning workflow calculates next version based on commit message prefixes (breaking:, feat:, etc.) but this convention is not documented in contributing guidelines. — *Document commit message conventions for version bumping in CONTRIBUTING.md or README. Consider using conventional commits specification.*
- 🟡 **[build]** GoReleaser is used for releases (.github/workflows/release.yml) but no .goreleaser.yml configuration file is visible in the scan. — *Ensure .goreleaser.yml exists and is properly configured. This file defines build targets, archives, and release artifacts.*
- ℹ️ **[convention]** Configuration file location is documented as ~/.boatman.yaml but the precedence order (env vars, local .boatman.yaml, global config) and schema validation are not specified. — *Document complete configuration precedence and provide a schema or example with all available options.*

## Sources of Truth

These files are the authoritative source for their respective concerns:

- **CI configuration**: [`.github/workflows/auto-version.yml`](.github/workflows/auto-version.yml)
- **CI configuration**: [`.github/workflows/release.yml`](.github/workflows/release.yml)
- **CI configuration**: [`.github/workflows/test.yml`](.github/workflows/test.yml)
