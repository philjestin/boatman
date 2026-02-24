# AGENTS.md

## Repository Overview

This is a **single-package repository**.
Primary language(s): go.

### Architecture


This is a **shared Go library module** within the larger `boatman-ecosystem` monorepo. The module is located at `github.com/philjestin/boatman-ecosystem/shared` and serves as a common dependency for other services or applications in the ecosystem. It uses Go 1.24.1 and has no external dependencies (based on the minimal `go.mod`). The repository follows standard Go project conventions with `gofmt` for formatting. As a shared library, it likely exports reusable types, utilities, or domain logic consumed by sibling packages in the boatman-ecosystem. The absence of entrypoints confirms this is not a standalone application but rather a library meant to be imported. Data flow is inbound (consumed by other modules) rather than outbound.


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

