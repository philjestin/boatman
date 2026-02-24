# AGENTS.md

## Repository Overview

This is a **single-package repository**.
Primary language(s): go.

### Architecture


## Architecture Summary

**Boatman** is a desktop application built with **Wails v2** (Go backend + React TypeScript frontend) that provides a Claude AI interface for codebase interaction and production incident investigation.

**Core Systems:**
1. **Wails Application Shell** (`main.go`, `app.go`): Embeds the React frontend and exposes Go methods to JavaScript via bindings. Manages application lifecycle, event routing, and session state.
2. **Agent Session Management** (`agent/`): Handles Claude AI interactions, sub-agent tracking, stream parsing, and message attribution. Supports multiple concurrent agents with lifecycle tracking.
3. **BoatmanMode Integration** (`boatmanmode/`): Integrates with an external CLI (`github.com/philjestin/boatmanmode`) for autonomous ticket execution workflows (plan → execute → review → refactor → PR). Manages subprocess execution and event routing.
4. **Firefighter Mode**: Specialized workflow for production incident investigation, integrating Linear (tickets), Bugsnag (errors), Datadog (logs), and Slack. Automates triage, root cause analysis, and fix generation.
5. **MCP (Model Context Protocol) Integration** (`mcp/`): Manages external tool servers (GitHub, Datadog, Bugsnag, Linear, Slack) with OAuth/SSO support.
6. **Project & Git Management** (`project/`, `git/`): Handles workspace management, git operations, worktree isolation for parallel investigations, and diff parsing (`diff/`).
7. **React Frontend** (`frontend/`): TypeScript UI with components for chat, agent logs panel, task tracking, diff approval, session search, and onboarding wizard.

**Data Flow:** User interacts with React UI → Wails bindings invoke Go methods → Go backend orchestrates Claude API calls, MCP servers, git operations, or BoatmanMode CLI → Events stream back to frontend via Wails runtime → UI updates with agent messages, tasks, diffs, and logs.

**Key Patterns:** Event-driven architecture with structured streaming (Claude events, BoatmanMode events), git worktree isolation for safe parallel work, and modular MCP server integration for extensibility.


## Repository Map

Single package at repository root.

## Commands

### Quick Reference

| Action | Command |
|--------|---------|
| build | `go build ./...` |
| test | `go test ./...` |
| format | `gofmt -w .` |

### Test Frameworks

- **go_test**: `go test ./...`

### Agent Scripts

Use these wrapper scripts for deterministic, machine-readable output:

- `scripts/agent-test` — run tests with structured output
- `scripts/agent-lint` — run linter with structured output
- `scripts/agent-build` — run build and capture results

### Inferred Command Matrix

The following commands were identified by analyzing project configuration:

| Scope | Build | Test | Lint | Format | Dev |
|-------|-------|------|------|--------|-----|
| root | `wails build` | `go test ./...` | - | `gofmt -w .` | `wails dev` |
| frontend | `npm install` | - | - | - | - |

## Conventions & Constraints

### Do

- **Use gofmt** for formatting

### Don't

- Don't bypass linter rules without team discussion
- Don't add new dependencies without checking for existing alternatives

### Inferred Architectural Constraints

- **All Go backend code must be bound to the Wails app instance in main.go's Bind slice to be callable from the frontend.** — Wails requires explicit binding of Go structs/methods to expose them to the JavaScript frontend. The main.go file shows app is bound via Bind: []interface{}{app}. *(enforced via: Create a test that verifies all exported methods on bound structs are accessible via Wails runtime. Use reflection to check bindings.)*
- **Frontend assets must be embedded in the Go binary using go:embed directive pointing to frontend/dist.** — main.go uses //go:embed all:frontend/dist to embed the built React app. This is required for Wails to serve the frontend. *(enforced via: Build process should fail if frontend/dist does not exist or is empty. Add a pre-build check in wails build.)*
- **Local module dependencies (harness, shared, cli) must exist as sibling directories to the desktop repository.** — go.mod contains replace directives pointing to ../harness, ../shared, and ../cli, indicating a monorepo structure at boatman-ecosystem/. *(enforced via: Add a Makefile or build script that validates the presence of ../harness, ../shared, and ../cli before running go build.)*
- **Configuration files must be stored in ~/.boatman/ and ~/.claude/ directories.** — README documents config.json in ~/.boatman/ and claude_mcp_config.json in ~/.claude/. This is a user-facing convention for settings persistence.
- **Agent session state and sub-agent tracking must be managed through the agent/ package.** — README and architecture description indicate agent/ handles session management, sub-agent lifecycle, and stream parsing. Centralizing this logic prevents state inconsistencies. *(enforced via: Use Go package visibility (unexported types) to prevent direct session state manipulation outside agent/. Add integration tests for session lifecycle.)*
- **Git worktrees must be used for isolated Firefighter Mode and BoatmanMode investigations.** — README emphasizes 'Isolated Fixes: Uses git worktrees for safe, parallel investigations' and 'Git worktree isolation for safe parallel work.' This prevents conflicts during concurrent operations. *(enforced via: Add tests that verify Firefighter/BoatmanMode workflows create separate worktrees and do not modify the main working directory.)*
- **MCP server configurations must follow the Model Context Protocol specification and be stored in ~/.claude/claude_mcp_config.json.** — README mentions MCP integration with built-in servers (GitHub, Datadog, etc.) and custom server support. The config path is documented as ~/.claude/claude_mcp_config.json. *(enforced via: Validate MCP config JSON schema on load. Provide clear error messages if config is malformed.)*
- **BoatmanMode CLI integration must use structured event streaming with defined event types (planning, execution, review, refactor).** — README describes 'Real-time event streaming with structured task tracking' and references BOATMANMODE_EVENTS.md for event specification. This ensures consistent UI updates. *(enforced via: Define event schema (e.g., JSON schema or Go structs) and validate all events from the CLI subprocess against it. Add tests for event parsing.)*
- **Frontend components must use TypeScript and follow the established directory structure (components/, hooks/, types/, store/).** — README documents 'frontend/src/components/, hooks/, types/, store/' structure. Consistency aids navigation and maintainability. *(enforced via: Use ESLint rules to enforce file placement (e.g., components must be in components/, hooks in hooks/). Add pre-commit hooks.)*
- **All user-facing features must support the onboarding wizard flow for first-run setup.** — README emphasizes 'Onboarding Wizard: Guided first-run experience for setting up authentication and preferences' and 'First Launch' section. This ensures new users can configure the app.

## Known Risks & Ambiguities

- 🔴 **[dependency]** Three local module replacements in go.mod point to relative paths (../harness, ../shared, ../cli). These dependencies are not in the scanned repository and will break builds outside the monorepo context. — *Verify that the repository is part of a monorepo structure at /Users/pmiddleton/workspace/personal/boatman-ecosystem/ and document the required sibling directories (harness, shared, cli) for builds to succeed.*
- 🟡 **[dependency]** go.mod uses version 0.0.0 for github.com/philjestin/boatman-ecosystem/harness and github.com/philjestin/boatmanmode, which are then replaced with local paths. This makes it unclear what the canonical remote versions are. — *Clarify whether these modules are published to GitHub or are strictly local-only. If published, use proper semantic versions. If local-only, document the monorepo structure.*
- 🟡 **[build]** README mentions 'wails doctor' for dependency checking but this is not captured in the build_systems commands. Developers may miss this critical pre-build step. — *Add 'wails doctor' to onboarding documentation or create a setup script that runs it automatically.*
- ℹ️ **[architecture]** The application integrates with external services (Linear, Bugsnag, Datadog, Slack) via MCP servers and OAuth/Okta SSO, but configuration paths (~/.boatman/config.json, ~/.claude/claude_mcp_config.json) are hardcoded to user home directory. This may conflict in multi-user or containerized environments. — *Consider making config paths configurable via environment variables or command-line flags for flexibility in different deployment scenarios.*
- 🟡 **[test]** No test files or test framework configuration detected in the scan. The README mentions 'go test ./...' but it's unclear if any tests exist. — *Verify whether tests exist in the codebase. If not, flag this as a gap. If they do exist, ensure they are discoverable and documented.*
- ℹ️ **[build]** Platform-specific builds are documented (darwin/universal, windows/amd64, linux/amd64) but README states 'Currently macOS (M1/M2/M3). Windows/Linux support planned.' This creates ambiguity about actual platform support. — *Clarify which platforms are tested and supported vs. theoretically buildable. Update README to reflect current state.*
- 🟡 **[ci]** No CI configuration detected (ci_systems: null). For a desktop app with multiple platform targets and external integrations, lack of CI increases risk of regressions. — *Implement CI pipelines (GitHub Actions, etc.) for automated testing, linting, and multi-platform builds.*
- ℹ️ **[convention]** Frontend uses TypeScript and React but no linting/formatting commands are documented for the frontend. Only Go formatting (gofmt) is mentioned. — *Document frontend linting/formatting commands (e.g., eslint, prettier) if they exist, or add them to the project.*
- 🟡 **[architecture]** BoatmanMode integration relies on an external CLI subprocess (github.com/philjestin/boatmanmode) with event streaming. If the CLI crashes or hangs, error handling and recovery mechanisms are unclear. — *Document error handling strategy for subprocess failures. Consider adding timeout mechanisms and graceful degradation.*

## Sources of Truth

These files are the authoritative source for their respective concerns:

