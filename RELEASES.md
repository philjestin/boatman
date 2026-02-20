# Release Guide

Complete guide to releasing CLI and Desktop components.

## Quick Release

### Interactive Release (Recommended)

```bash
# Guided release process
./scripts/release.sh
```

This interactive script will:
1. Check working directory is clean
2. Ask which component(s) to release
3. Bump version(s)
4. Open changelogs for editing
5. Run tests
6. Commit and tag
7. Push to trigger CI/CD

### Manual Release

**CLI:**
```bash
# 1. Bump version
make bump-cli-minor

# 2. Update changelog
vim cli/CHANGELOG.md

# 3. Commit and tag
git add cli/VERSION cli/CHANGELOG.md
git commit -m "cli: Release v1.3.0"
git tag cli/v1.3.0

# 4. Push
git push origin main --tags
```

**Desktop:**
```bash
# 1. Bump version
make bump-desktop-minor

# 2. Update changelog
vim desktop/CHANGELOG.md

# 3. Commit and tag
git add desktop/wails.json desktop/CHANGELOG.md
git commit -m "desktop: Release v1.1.0"
git tag desktop/v1.1.0

# 4. Push
git push origin main --tags
```

## Release Checklist

### Pre-Release

- [ ] All PRs merged to main
- [ ] All tests passing
- [ ] No open blockers
- [ ] Documentation updated
- [ ] Breaking changes documented

### Version Bump

- [ ] Bump version using script or Makefile
- [ ] Update CHANGELOG.md with release notes
- [ ] Update README.md if needed
- [ ] For desktop: note bundled CLI version

### Testing

```bash
# CLI
cd cli
go test ./...
go build -o boatman ./cmd/boatman
./boatman work --prompt "test"

# Desktop
cd desktop
go test ./...
cd frontend && npm test
```

### Commit and Tag

- [ ] Commit version bump
- [ ] Tag with correct prefix (`cli/v1.2.3` or `desktop/v1.0.5`)
- [ ] Push tags

### Post-Release

- [ ] Verify GitHub release created
- [ ] Download and test release artifacts
- [ ] Deploy documentation site (see below)
- [ ] Announce release
- [ ] Close milestone (if using)

### Deploy Documentation Site

The documentation site (`docs/`) is hosted on Vercel. After any release that includes documentation changes, deploy the updated site:

```bash
# Install Vercel CLI if needed
npm i -g vercel

# Deploy from the docs directory
cd docs
npm install
npx vercel --prod
```

For preview deployments (before merging):

```bash
cd docs
npx vercel
```

The Vercel project is already linked via `docs/.vercel/project.json`. If prompted, use the existing project settings.

## Release Types

### Patch Release (Bug Fixes)

**When:** Bug fixes, minor improvements, no new features

```bash
# CLI: 1.2.3 -> 1.2.4
make bump-cli-patch

# Desktop: 1.0.5 -> 1.0.6
make bump-desktop-patch
```

**Changelog:**
```markdown
## [1.2.4] - 2026-02-14

### Fixed
- Fix event parsing for large diffs
- Correct git worktree cleanup
```

### Minor Release (New Features)

**When:** New features, backward compatible

```bash
# CLI: 1.2.3 -> 1.3.0
make bump-cli-minor

# Desktop: 1.0.5 -> 1.1.0
make bump-desktop-minor
```

**Changelog:**
```markdown
## [1.3.0] - 2026-02-14

### Added
- Task metadata in agent_completed events
- Support for refactor_diff field

### Changed
- Improved error messages
```

### Major Release (Breaking Changes)

**When:** Breaking API changes, incompatible changes

```bash
# CLI: 1.2.3 -> 2.0.0
make bump-cli-major

# Desktop: 1.0.5 -> 2.0.0
make bump-desktop-major
```

**Changelog:**
```markdown
## [2.0.0] - 2026-02-14

### Breaking Changes
- Changed event structure (incompatible with v1.x desktop)
- Requires desktop v2.0.0+

### Migration Guide
- Update desktop to v2.0.0
- Event handlers need updating
```

## Coordinated Releases

When releasing features that span both components:

```bash
# 1. Bump both versions
make bump-cli-minor
make bump-desktop-minor

# 2. Update both changelogs
vim cli/CHANGELOG.md
vim desktop/CHANGELOG.md

# 3. Commit both
git add cli desktop
git commit -m "Release: CLI v1.3.0, Desktop v1.1.0

- Added event metadata support
- Desktop shows task details with diffs"

# 4. Tag both
git tag cli/v1.3.0
git tag desktop/v1.1.0

# 5. Push
git push origin main --tags
```

Both workflows will run in parallel.

## GitHub Actions

Release workflows automatically trigger on tag push:

### CLI Release (`cli/v*` tags)

**Workflow:** `.github/workflows/release-cli.yml`

**Steps:**
1. Checkout code
2. Set up Go
3. Run tests
4. Run GoReleaser
5. Build binaries for:
   - Linux (amd64, arm64)
   - macOS (amd64, arm64)
   - Windows (amd64)
6. Create GitHub release
7. Upload binaries

**Artifacts:**
- `boatman_Linux_x86_64.tar.gz`
- `boatman_Darwin_x86_64.tar.gz`
- `boatman_Darwin_arm64.tar.gz`
- `boatman_Windows_x86_64.zip`

### Desktop Release (`desktop/v*` tags)

**Workflow:** `.github/workflows/release-desktop.yml`

**Steps:**
1. Checkout code
2. Set up Go and Node
3. Install Wails
4. **Build CLI first** (bundled with desktop)
5. Build desktop for:
   - macOS (Intel & Apple Silicon)
   - Linux (AppImage/deb)
   - Windows (exe/msi)
6. Create GitHub release
7. Upload installers

**Artifacts:**
- `boatman-desktop-macos.dmg`
- `boatman-desktop-linux.AppImage`
- `boatman-desktop-windows.exe`

## Version Files

### CLI

**Location:** `cli/VERSION`
```
v1.3.0
```

**Usage in code:**
```go
// cli/version.go (optional)
package main

var Version = "1.3.0"  // Updated by release script
```

### Desktop

**Location:** `desktop/wails.json`
```json
{
  "name": "boatman",
  "version": "1.1.0",
  "outputfilename": "boatman"
}
```

**Usage in code:**
```go
// desktop/app.go
func (a *App) GetVersion() string {
    return "1.1.0"  // Could read from wails.json
}
```

## Changelogs

### Format

Follow [Keep a Changelog](https://keepachangelog.com/):

```markdown
## [1.3.0] - 2026-02-14

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
- API changes requiring user action
```

### Desktop-Specific Sections

```markdown
## [1.1.0] - 2026-02-14

### Bundled CLI Version
- Bundles CLI v1.3.0

### Minimum CLI Version
- Requires CLI >= v1.3.0 (for metadata support)

### Migration
- If using external CLI, upgrade to v1.3.0+
```

## Pre-releases

For testing before official release:

```bash
# Alpha
git tag cli/v1.3.0-alpha.1
git push origin --tags

# Beta
git tag cli/v1.3.0-beta.1
git push origin --tags

# Release Candidate
git tag cli/v1.3.0-rc.1
git push origin --tags

# Final
git tag cli/v1.3.0
git push origin --tags
```

Pre-releases are marked in GitHub as "Pre-release".

## Hotfix Releases

For critical bugs in production:

```bash
# 1. Create hotfix branch from tag
git checkout -b hotfix/cli-v1.2.4 cli/v1.2.3

# 2. Fix the bug
vim cli/internal/agent/agent.go
git commit -m "fix: Critical bug in agent"

# 3. Bump patch version
echo "v1.2.4" > cli/VERSION

# 4. Update changelog
vim cli/CHANGELOG.md

# 5. Commit, tag, merge
git commit -am "cli: Hotfix v1.2.4"
git tag cli/v1.2.4
git checkout main
git merge hotfix/cli-v1.2.4
git push origin main --tags
```

## Rollback

If a release has critical issues:

### Mark as Pre-release

In GitHub, edit the release and mark it as "Pre-release" with a warning.

### Release New Version

```bash
# Don't delete tags - create new version
make bump-cli-patch  # Fix and release v1.3.1
```

### Communication

- Post issue explaining the problem
- Update release notes
- Notify users
- Document in changelog

## Version Compatibility Matrix

| Desktop | Bundled CLI | Min CLI | Event Protocol |
|---------|-------------|---------|----------------|
| v1.0.x  | v1.0.0      | v1.0.0  | v1             |
| v1.1.x  | v1.3.0      | v1.3.0  | v1             |
| v2.0.x  | v2.0.0      | v2.0.0  | v2             |

## Monitoring Releases

**Check workflow status:**
```bash
# Via gh CLI
gh run list --workflow=release-cli.yml
gh run list --workflow=release-desktop.yml

# Via web
open https://github.com/YOUR_ORG/boatman-ecosystem/actions
```

**View releases:**
```bash
# Via gh CLI
gh release list

# Via web
open https://github.com/YOUR_ORG/boatman-ecosystem/releases
```

## Tips

### Test Before Release

```bash
# Build and test locally first
make build-all
make test-all

# Test CLI
./cli/boatman work --prompt "test"

# Test Desktop
cd desktop && wails dev
```

### Use Milestones

Group related issues/PRs:
1. Create milestone: "CLI v1.3.0"
2. Assign issues/PRs
3. Close milestone on release

### Automate More

Optional automation:
- Conventional commits → auto-version bumps
- PR labels → changelog generation
- Automated testing on tag push
- Slack/Discord notifications

## Troubleshooting

### Workflow Failed

```bash
# Check logs
gh run view <run-id>

# Re-run failed jobs
gh run rerun <run-id>
```

### Wrong Version Tagged

```bash
# Delete local tag
git tag -d cli/v1.3.0

# Delete remote tag
git push origin :refs/tags/cli/v1.3.0

# Create correct tag
git tag cli/v1.3.0
git push origin --tags
```

### Binary Not Working

- Check GoReleaser config
- Test locally: `goreleaser build --snapshot`
- Verify ldflags are set correctly

## Questions?

- Check existing releases for examples
- Review GitHub Actions logs
- See VERSIONING.md for strategy details
