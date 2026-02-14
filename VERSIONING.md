# Versioning and Release Strategy

The Boatman ecosystem uses **independent versioning** with coordinated releases.

## Version Numbers

Each component has its own semantic version:

- **CLI**: `cli/v1.2.3`
- **Desktop**: `desktop/v1.0.5`
- **Shared**: Internal library, no independent releases

## Semantic Versioning

Both follow [SemVer](https://semver.org/):

```
MAJOR.MINOR.PATCH

MAJOR: Breaking changes
MINOR: New features (backward compatible)
PATCH: Bug fixes
```

### Examples

**CLI Version Bump:**
- Add metadata to events → `cli/v1.3.0` (new feature)
- Fix event parsing → `cli/v1.2.4` (bug fix)
- Change event structure → `cli/v2.0.0` (breaking)

**Desktop Version Bump:**
- Add task detail modal → `desktop/v1.1.0` (new feature)
- Fix UI bug → `desktop/v1.0.6` (bug fix)
- Require new CLI version → `desktop/v2.0.0` (breaking)

## Git Tags

Tags include the component prefix:

```bash
# CLI releases
git tag cli/v1.2.3
git tag cli/v1.3.0

# Desktop releases
git tag desktop/v1.0.5
git tag desktop/v1.1.0

# Can also tag both together for coordinated releases
git tag cli/v1.3.0 desktop/v1.1.0
```

## Release Workflow

### CLI Release

```bash
# 1. Update version in code
echo "v1.3.0" > cli/VERSION

# 2. Update changelog
vim cli/CHANGELOG.md

# 3. Commit
git add cli/VERSION cli/CHANGELOG.md
git commit -m "cli: Release v1.3.0"

# 4. Tag
git tag cli/v1.3.0

# 5. Push
git push origin main --tags

# GitHub Actions will:
# - Build binaries for all platforms
# - Create GitHub release
# - Upload artifacts
```

### Desktop Release

```bash
# 1. Update version in code
vim desktop/wails.json  # Update version field

# 2. Update changelog
vim desktop/CHANGELOG.md

# 3. Commit
git add desktop/wails.json desktop/CHANGELOG.md
git commit -m "desktop: Release v1.1.0"

# 4. Tag
git tag desktop/v1.1.0

# 5. Push
git push origin main --tags

# GitHub Actions will:
# - Build desktop app for all platforms
# - Bundle latest CLI
# - Create GitHub release
# - Upload .dmg, .exe, .AppImage
```

### Coordinated Release (Both)

When releasing features that span both components:

```bash
# 1. Update both versions
echo "v1.3.0" > cli/VERSION
vim desktop/wails.json

# 2. Update both changelogs
vim cli/CHANGELOG.md
vim desktop/CHANGELOG.md

# 3. Commit
git add cli/VERSION cli/CHANGELOG.md desktop/wails.json desktop/CHANGELOG.md
git commit -m "Release: CLI v1.3.0, Desktop v1.1.0

- Added metadata to event protocol
- Desktop now shows task details with diffs"

# 4. Tag both
git tag cli/v1.3.0
git tag desktop/v1.1.0

# 5. Push
git push origin main --tags

# Both workflows run in parallel
```

## Version Compatibility

Desktop releases specify which CLI version they bundle:

```markdown
# desktop/CHANGELOG.md

## [1.1.0] - 2026-02-14

### Changed
- Bundles CLI v1.3.0 (includes metadata support)
- Task details now show diffs and feedback

### Requires
- CLI >= v1.3.0 (for metadata support)
```

## Automated Versioning

### Using Scripts

```bash
# Bump CLI version
make bump-cli-patch   # 1.2.3 -> 1.2.4
make bump-cli-minor   # 1.2.3 -> 1.3.0
make bump-cli-major   # 1.2.3 -> 2.0.0

# Bump desktop version
make bump-desktop-patch
make bump-desktop-minor
make bump-desktop-major
```

### Using Conventional Commits

Optionally use commit prefixes to auto-determine version bumps:

```bash
# Patch bump
git commit -m "fix(cli): Fix event parsing"
git commit -m "fix(desktop): Fix modal rendering"

# Minor bump
git commit -m "feat(cli): Add metadata to events"
git commit -m "feat(desktop): Add task detail modal"

# Major bump (breaking change)
git commit -m "feat(cli)!: Change event structure"
git commit -m "feat(desktop)!: Require CLI v2.0.0"
```

## Release Branches (Optional)

For long-term support:

```
main                 (development)
├── release/cli/v1.x     (CLI v1 LTS)
└── release/desktop/v1.x (Desktop v1 LTS)
```

Backport critical fixes:

```bash
# Fix in main
git commit -m "fix(cli): Critical security fix"

# Cherry-pick to release branch
git checkout release/cli/v1.x
git cherry-pick <commit-sha>
git tag cli/v1.2.4
git push origin release/cli/v1.x --tags
```

## Pre-releases

For testing before official release:

```bash
# CLI alpha/beta/rc
git tag cli/v1.3.0-alpha.1
git tag cli/v1.3.0-beta.1
git tag cli/v1.3.0-rc.1
git tag cli/v1.3.0  # Final release

# Desktop
git tag desktop/v1.1.0-beta.1
git tag desktop/v1.1.0
```

## Version Files

### CLI Version

```bash
# cli/VERSION
v1.3.0
```

```go
// cli/version.go
package main

var Version = "1.3.0"
```

### Desktop Version

```json
// desktop/wails.json
{
  "name": "boatman",
  "version": "1.1.0",
  "outputfilename": "boatman"
}
```

## Release Checklist

### Before Release

- [ ] All tests pass
- [ ] Update VERSION/version.json
- [ ] Update CHANGELOG.md
- [ ] Update README if needed
- [ ] Test locally
- [ ] Review diff since last release

### Release

- [ ] Commit version bump
- [ ] Tag with correct version
- [ ] Push tags
- [ ] Wait for CI to complete
- [ ] Verify GitHub release created
- [ ] Test released artifacts

### After Release

- [ ] Announce in discussions/slack
- [ ] Update documentation site
- [ ] Close related issues/PRs
- [ ] Plan next release

## FAQ

### Can I release CLI without Desktop?

Yes! They're independent. Just tag `cli/v1.3.0` and only the CLI workflow runs.

### Does Desktop always need a new CLI?

No. Desktop can release bug fixes independently. Only update bundled CLI when needed.

### How do I know which CLI is bundled?

Check the desktop release notes or run:
```bash
./boatman --version  # Inside desktop.app/Contents/MacOS/
```

### What if event protocol changes?

1. Release CLI with new protocol: `cli/v2.0.0` (breaking)
2. Update desktop to use new protocol: `desktop/v2.0.0` (breaking)
3. Tag both, document migration

### Can users mix versions?

Technically yes (CLI is separate), but:
- Desktop bundles a specific CLI version
- Event protocol must match
- We document compatibility in changelogs

## Version Matrix

| Desktop | Bundled CLI | Event Protocol | Compatible CLI Range |
|---------|-------------|----------------|---------------------|
| v1.0.0  | v1.2.0      | v1             | v1.2.0 - v1.x.x     |
| v1.1.0  | v1.3.0      | v1             | v1.3.0 - v1.x.x     |
| v2.0.0  | v2.0.0      | v2             | v2.0.0+             |

## Deprecation Policy

### CLI
- Major versions supported for 1 year after new major
- Security fixes backported to N-1 major

### Desktop
- Latest major only
- Encourages updates for new features

## Tools

- **GoReleaser**: Automates CLI binary builds
- **Wails**: Builds desktop apps
- **GitHub Actions**: Runs release workflows
- **Conventional Commits**: Optional versioning hints
- **Semantic Release**: Optional automated versioning
