# Changelog - Desktop (Boatman App)

All notable changes to the desktop component will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Hybrid architecture: subprocess + direct imports
- Direct import utilities via `services/boatman_hybrid.go`
- Shared types from `shared/` package

### Changed
- Moved from standalone repo to monorepo at `desktop/`
- Integration layer updated to use shared event types

## [1.0.0] - 2026-02-14

### Added
- Initial release in monorepo
- Cross-platform desktop application (macOS, Linux, Windows)
- Real-time boatmanmode execution with streaming output
- Session management and history
- Project-based organization
- Task tracking with status indicators
- Clickable task details showing:
  - Execution diffs
  - Review feedback
  - Refactoring changes
  - Planning details
  - Issues found
- Git integration (status, diffs)
- Linear ticket integration
- Firefighter mode for incident response
- Search and filtering
- Favorites and tags

### Bundled CLI Version
- Bundles CLI v1.0.0

### Minimum CLI Version
- Requires CLI >= v1.0.0

### Changed
- Event handling now uses subprocess execution
- Task metadata captured from CLI events

### Technical
- Built with Wails v2
- React + TypeScript frontend
- Go backend
- Event-driven architecture

---

## Release Notes Template

When releasing, copy this template:

```markdown
## [X.Y.Z] - YYYY-MM-DD

### Added
- New features

### Changed
- Changes to existing functionality

### Fixed
- Bug fixes

### Bundled CLI Version
- Bundles CLI vX.Y.Z

### Minimum CLI Version
- Requires CLI >= vX.Y.Z

### Breaking Changes
- Changes that require user action

### Migration Guide
- Steps to migrate from previous version
```
