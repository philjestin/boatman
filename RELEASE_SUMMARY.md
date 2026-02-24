# Release System Summary

Complete guide to the versioning and release system for the Boatman ecosystem.

## Overview

The Boatman ecosystem uses **independent versioning** for CLI and Desktop components within a single monorepo.

## Components

| Component | Version | Tag Format | Location |
|-----------|---------|------------|----------|
| CLI       | v1.0.0  | `cli/v1.0.0` | `cli/` |
| Desktop   | v1.0.0  | `desktop/v1.0.0` | `desktop/` |
| Platform  | v0.1.0  | `platform/v0.1.0` | `platform/` |
| Harness   | (internal) | No tags | `harness/` |
| Shared    | (internal) | No tags | `shared/` |

## Quick Commands

```bash
# Show current versions
make show-versions

# Bump versions
make bump-cli-patch        # CLI patch: 1.0.0 -> 1.0.1
make bump-cli-minor        # CLI minor: 1.0.0 -> 1.1.0
make bump-cli-major        # CLI major: 1.0.0 -> 2.0.0
make bump-desktop-patch    # Desktop patch
make bump-desktop-minor    # Desktop minor
make bump-desktop-major    # Desktop major
make bump-platform-patch   # Platform patch: 0.1.0 -> 0.1.1
make bump-platform-minor   # Platform minor: 0.1.0 -> 0.2.0
make bump-platform-major   # Platform major: 0.1.0 -> 1.0.0

# Interactive release (recommended)
./scripts/release.sh

# Manual release
./scripts/bump-version.sh cli minor
vim cli/CHANGELOG.md
git add cli && git commit -m "cli: Release v1.1.0"
git tag cli/v1.1.0
git push origin main --tags
```

## Release Workflow

### 1. Prepare

```bash
# Ensure clean working directory
git status

# Run tests
make test-all

# Show current versions
make show-versions
```

### 2. Bump Version

**Option A: Interactive (Recommended)**
```bash
./scripts/release.sh
# Follow prompts - it handles everything
```

**Option B: Manual**
```bash
# CLI
make bump-cli-minor
vim cli/CHANGELOG.md

# Desktop
make bump-desktop-minor
vim desktop/CHANGELOG.md
```

### 3. Commit and Tag

**Interactive script handles this, or manually:**
```bash
# Single component
git add cli/VERSION cli/CHANGELOG.md
git commit -m "cli: Release v1.1.0"
git tag cli/v1.1.0

# Both components (coordinated)
git add cli desktop
git commit -m "Release: CLI v1.1.0, Desktop v1.1.0"
git tag cli/v1.1.0 desktop/v1.1.0
```

### 4. Push

```bash
git push origin main --tags
```

This triggers GitHub Actions to:
- Build binaries (CLI)
- Build installers (Desktop)
- Create GitHub releases
- Upload artifacts

### 5. Verify

```bash
# Check workflow status
gh run list

# View releases
gh release list

# Or via web
open https://github.com/YOUR_ORG/boatman-ecosystem/releases
```

## Automation

### GitHub Actions

**CLI Release** (`.github/workflows/release-cli.yml`)
- Triggers on: `cli/v*` tags
- Builds: Linux, macOS, Windows binaries
- Uploads: tar.gz, zip archives

**Desktop Release** (`.github/workflows/release-desktop.yml`)
- Triggers on: `desktop/v*` tags
- Builds: macOS, Linux, Windows installers
- Bundles: Latest CLI binary
- Uploads: .dmg, .AppImage, .exe

### Version Files

**CLI:** `cli/VERSION`
```
v1.0.0
```

**Desktop:** `desktop/wails.json`
```json
{
  "version": "1.0.0"
}
```

### Scripts

| Script | Purpose |
|--------|---------|
| `scripts/bump-version.sh` | Bump version for component |
| `scripts/release.sh` | Interactive release wizard |
| `scripts/setup.sh` | Initial setup |

## Release Types

### Patch (1.0.0 → 1.0.1)
- **When:** Bug fixes only
- **Command:** `make bump-cli-patch`
- **Changelog:** `### Fixed`

### Minor (1.0.0 → 1.1.0)
- **When:** New features, backward compatible
- **Command:** `make bump-cli-minor`
- **Changelog:** `### Added`, `### Changed`

### Major (1.0.0 → 2.0.0)
- **When:** Breaking changes
- **Command:** `make bump-cli-major`
- **Changelog:** `### Breaking Changes`

## Changelog Format

```markdown
## [1.1.0] - 2026-02-14

### Added
- New feature X

### Changed
- Improved Y

### Fixed
- Bug Z

### Breaking Changes
- API change requiring user action
```

**Desktop-specific:**
```markdown
### Bundled CLI Version
- Bundles CLI v1.1.0

### Minimum CLI Version
- Requires CLI >= v1.1.0
```

## Coordinated Releases

When features span both components:

```bash
# Bump both
make bump-cli-minor
make bump-desktop-minor

# Update both changelogs
vim cli/CHANGELOG.md
vim desktop/CHANGELOG.md

# Commit and tag both
git add cli desktop
git commit -m "Release: CLI v1.1.0, Desktop v1.1.0"
git tag cli/v1.1.0 desktop/v1.1.0

# Push
git push origin main --tags
```

Both workflows run in parallel.

## Version Compatibility

Desktop releases specify bundled and minimum CLI versions:

| Desktop | Bundled CLI | Min CLI Required |
|---------|-------------|------------------|
| v1.0.0  | v1.0.0      | v1.0.0           |
| v1.1.0  | v1.1.0      | v1.1.0           |
| v2.0.0  | v2.0.0      | v2.0.0           |

## Pre-releases

For testing:

```bash
# Create pre-release tag
git tag cli/v1.1.0-beta.1
git push origin --tags

# GitHub marks it as "Pre-release"
# Users can test before stable release
```

Progression:
1. `v1.1.0-alpha.1` (internal testing)
2. `v1.1.0-beta.1` (external testing)
3. `v1.1.0-rc.1` (release candidate)
4. `v1.1.0` (stable)

## Hotfixes

For critical production bugs:

```bash
# Branch from production tag
git checkout -b hotfix/cli-1.0.1 cli/v1.0.0

# Fix bug
git commit -m "fix: Critical security issue"

# Bump patch
echo "v1.0.1" > cli/VERSION
vim cli/CHANGELOG.md
git commit -m "cli: Hotfix v1.0.1"

# Tag and merge
git tag cli/v1.0.1
git checkout main
git merge hotfix/cli-1.0.1
git push origin main --tags
```

## Directory Structure

```
boatman-ecosystem/
├── .github/workflows/
│   ├── release-cli.yml          # CLI release automation
│   └── release-desktop.yml      # Desktop release automation
│
├── cli/
│   ├── VERSION                  # CLI version file
│   ├── CHANGELOG.md             # CLI changelog
│   └── .goreleaser.yml          # GoReleaser config
│
├── desktop/
│   ├── wails.json               # Desktop version (in JSON)
│   └── CHANGELOG.md             # Desktop changelog
│
└── scripts/
    ├── bump-version.sh          # Version bump script
    └── release.sh               # Interactive release wizard
```

## Documentation

- **VERSIONING.md** - Versioning strategy and policy
- **RELEASES.md** - Complete release guide
- **RELEASE_SUMMARY.md** - This file (quick reference)

## Checklist

### Pre-Release
- [ ] Tests passing
- [ ] Documentation updated
- [ ] Breaking changes documented
- [ ] Changelog updated

### Release
- [ ] Version bumped
- [ ] Changes committed
- [ ] Tag created
- [ ] Pushed to remote

### Post-Release
- [ ] GitHub release verified
- [ ] Artifacts downloaded and tested
- [ ] Announcement posted
- [ ] Documentation site updated

## Examples

### Release CLI v1.1.0

```bash
# Option 1: Interactive
./scripts/release.sh
# Select "1) CLI only"
# Select "2) Minor"
# Follow prompts

# Option 2: Manual
make bump-cli-minor
vim cli/CHANGELOG.md
git add cli
git commit -m "cli: Release v1.1.0"
git tag cli/v1.1.0
git push origin main --tags
```

### Release Desktop v1.0.1

```bash
make bump-desktop-patch
vim desktop/CHANGELOG.md
git add desktop
git commit -m "desktop: Release v1.0.1"
git tag desktop/v1.0.1
git push origin main --tags
```

### Release Both (Coordinated)

```bash
./scripts/release.sh
# Select "3) Both"
# Follow prompts
```

## Troubleshooting

### Workflow Failed
```bash
gh run list --workflow=release-cli.yml
gh run view <run-id>
gh run rerun <run-id>
```

### Wrong Tag
```bash
# Delete tag
git tag -d cli/v1.1.0
git push origin :refs/tags/cli/v1.1.0

# Recreate
git tag cli/v1.1.0
git push origin --tags
```

### Test Locally
```bash
# Test CLI build
cd cli
goreleaser build --snapshot --clean

# Test desktop build
cd desktop
wails build
```

## Summary

- **Independent versions** for CLI and Desktop
- **Automated releases** via GitHub Actions
- **Semantic versioning** (MAJOR.MINOR.PATCH)
- **Tag format:** `cli/vX.Y.Z`, `desktop/vX.Y.Z`
- **Scripts:** Interactive or manual workflows
- **Changelogs:** Keep a Changelog format

**Most common workflow:**
```bash
./scripts/release.sh  # Interactive wizard handles everything
```

For more details, see:
- **VERSIONING.md** - Strategy details
- **RELEASES.md** - Complete guide
