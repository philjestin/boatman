# AGENTS.md

## Repository Overview

This is a **single-package repository**.
Primary language(s): go.

### Architecture


This is a minimal Go module named `harness` within the `boatman-ecosystem` organization. The repository uses Go 1.24.1 and follows standard Go project conventions with `gofmt` for formatting. Based on the structural scan, this appears to be a single-module Go project without any visible subdirectories, packages, or entrypoints in the provided data. The module has no external dependencies listed in `go.mod`, suggesting it's either a very early-stage project, a library with only standard library dependencies, or the scan captured limited information. The lack of CI configuration, entrypoints, and detailed package structure indicates this is likely a foundational or skeletal project. Without access to actual Go source files, the specific domain, purpose, and internal architecture remain unclear, though the name "harness" suggests it may be a testing or orchestration utility within the broader boatman ecosystem.


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
| root | `go build ./...` | `go test ./...` | - | `gofmt -w .` | - |

## Conventions & Constraints

### Do

- **Use gofmt** for formatting

### Don't

- Don't bypass linter rules without team discussion
- Don't add new dependencies without checking for existing alternatives

### Inferred Architectural Constraints

- **Use gofmt for all Go code formatting** — The structural scan identified gofmt as the conventional formatter, which is the standard Go formatting tool *(enforced via: Add a pre-commit hook or CI check that runs 'gofmt -l .' and fails if any files are unformatted)*
- **Maintain Go 1.24.1 as the minimum supported version** — The go.mod file specifies 'go 1.24.1', establishing this as the language version baseline *(enforced via: CI should run tests with Go 1.24.1 to ensure compatibility; consider using 'go mod tidy -go=1.24.1')*
- **Follow standard Go project layout conventions** — The build system uses standard Go commands (go build ./..., go test ./...), implying adherence to Go's conventional project structure
- **Keep the module path as github.com/philjestin/boatman-ecosystem/harness** — This is the declared module path in go.mod and changing it would break imports for any consumers *(enforced via: Add a test that parses go.mod and verifies the module path hasn't changed unexpectedly)*

## Known Risks & Ambiguities

- 🔴 **[architecture]** No source files were provided in the scan, making it impossible to determine the actual package structure, domains, or code organization — *Scan and provide representative .go files from the repository to understand the actual architecture*
- 🟡 **[build]** No entrypoints (main packages) were detected, so it's unclear if this is a library or if there are executable commands to build — *Verify if there are cmd/ directories or main packages, and document the intended build artifacts*
- 🟡 **[ci]** No CI/CD configuration detected (no .github/workflows, .gitlab-ci.yml, etc.), which may lead to inconsistent testing and deployment practices — *Add CI configuration to automate testing, linting, and builds*
- ℹ️ **[dependency]** The go.mod file shows no dependencies, which is unusual for most Go projects and may indicate incomplete module initialization — *Run 'go mod tidy' to ensure dependencies are properly tracked*
- ℹ️ **[test]** No test framework configuration beyond standard go test was detected, and no test files were provided for analysis — *Verify that test files exist and follow Go testing conventions (*_test.go)*
- 🟡 **[convention]** No linting tools (golangci-lint, staticcheck, etc.) are configured, which may lead to code quality inconsistencies — *Add golangci-lint or similar tooling with a .golangci.yml configuration*

## Sources of Truth

These files are the authoritative source for their respective concerns:

