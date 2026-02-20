# Changelog - CLI (Boatmanmode)

All notable changes to the CLI component will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Monorepo structure with shared types
- Public utilities in `pkg/` for desktop integration
  - `pkg/diff`: Diff parsing and analysis utilities
  - `pkg/validation`: Plan and code validation
  - Available for direct import by desktop app
- Event metadata support (diffs, feedback, plans, issues)
  - `agent_completed` events include full context
  - Diffs embedded in event payloads
  - Feedback and issues from review phases
  - Plans and refactor changes tracked
- Automatic environment variable filtering for nested Claude sessions
  - Prevents Claude-in-Claude confusion
  - Filters `ANTHROPIC_*`, `CLAUDE_*` environment variables
  - Ensures clean execution context

## [1.0.0] - 2026-02-14

### Added
- Initial release in monorepo
- Autonomous development workflow (plan → execute → review → refactor)
- Git worktree isolation
- Claude AI integration
- Linear ticket support
- Structured JSON event emission
- Event types: `agent_started`, `agent_completed`, `task_created`, `task_updated`, `progress`
- Event metadata: `diff`, `plan`, `feedback`, `issues`, `refactor_diff`

### Changed
- Moved from standalone repo to monorepo at `cli/`
- Event protocol now uses shared types from `shared/events`

### Deprecated
- None

### Removed
- None

### Fixed
- None

### Security
- Automatic environment variable filtering for nested Claude sessions

---

## Release Notes Template

When releasing, copy this template:

```markdown
## [X.Y.Z] - YYYY-MM-DD

### Added
- New features

### Changed
- Changes to existing functionality

### Deprecated
- Soon-to-be removed features

### Removed
- Removed features

### Fixed
- Bug fixes

### Security
- Security fixes

### Breaking Changes
- API changes that require user action
```
