# Migration Guide: Standalone → Monorepo

This guide helps you transition from the old standalone repositories to the new monorepo structure.

## What Changed

### Before (Standalone)
```
~/workspace/handshake/boatmanmode/     # CLI
~/workspace/personal/boatmanapp/       # Desktop
```

### After (Monorepo)
```
~/workspace/personal/boatman-ecosystem/
├── cli/                               # CLI (was boatmanmode)
└── desktop/                           # Desktop (was boatmanapp)
```

## For Users

### If you were using the CLI

**Before:**
```bash
cd ~/workspace/handshake/boatmanmode
go build -o boatman ./cmd/boatman
./boatman work --prompt "task"
```

**After:**
```bash
cd ~/workspace/personal/boatman-ecosystem
make build-cli
./cli/boatman work --prompt "task"

# Or install to ~/bin
make install-cli
boatman work --prompt "task"
```

### If you were using the desktop app

**Before:**
```bash
cd ~/workspace/personal/boatmanapp
wails dev
```

**After:**
```bash
cd ~/workspace/personal/boatman-ecosystem
make dev

# Or manually
cd desktop
wails dev
```

## For Developers

### Setting up the new structure

```bash
# Clone the monorepo
git clone <repo-url> boatman-ecosystem
cd boatman-ecosystem

# Run setup
./scripts/setup.sh

# Or manually
go work sync
make build-cli
cd desktop/frontend && npm install
```

### Import path changes

**No changes needed!** Each component still has its own `go.mod`:
- CLI: `github.com/philjestin/boatmanmode`
- Desktop: (desktop doesn't publish as a library)

### Environment variables

**No changes needed.** Everything works the same:
- `ANTHROPIC_API_KEY` for Claude access
- `LINEAR_API_KEY` for Linear integration
- `BOATMAN_DEBUG=1` for verbose output

### Binary locations

The desktop app now looks for the CLI binary in this order:

1. `boatman` in PATH
2. `~/workspace/personal/boatman-ecosystem/cli/boatman` (monorepo)
3. `~/workspace/handshake/boatmanmode/boatman` (old location, fallback)

This means your existing setup will continue to work during migration.

### Configuration files

**CLI config** (`.boatman.yaml`):
- Old location: `~/workspace/handshake/boatmanmode/.boatman.yaml`
- New location: `~/workspace/personal/boatman-ecosystem/cli/.boatman.yaml`

You can copy your existing config:
```bash
cp ~/workspace/handshake/boatmanmode/.boatman.yaml \
   ~/workspace/personal/boatman-ecosystem/cli/
```

**Desktop config** (stored in home directory):
- No changes needed - config lives in `~/.boatman/` regardless

## Migration Checklist

### For CLI users

- [ ] Clone the monorepo
- [ ] Run `make build-cli`
- [ ] Copy your `.boatman.yaml` config
- [ ] Test with a simple prompt
- [ ] Update any scripts that reference the old path
- [ ] Optional: `make install-cli` to add to PATH

### For Desktop users

- [ ] Clone the monorepo
- [ ] Run `./scripts/setup.sh`
- [ ] Test `make dev`
- [ ] Verify CLI binary is found (check logs)
- [ ] Optional: Remove old directories once verified

### For Contributors

- [ ] Clone the monorepo
- [ ] Review CONTRIBUTING.md
- [ ] Understand the event protocol
- [ ] Set up pre-commit hooks (if any)
- [ ] Join the discussions

## Rollback Plan

If you need to temporarily go back to the old structure:

```bash
# The old repositories still exist
cd ~/workspace/handshake/boatmanmode      # Old CLI
cd ~/workspace/personal/boatmanapp        # Old desktop

# They'll continue to work (though won't get updates)
```

## Benefits of the Monorepo

1. **Atomic commits** - Features that span CLI and desktop are one commit
2. **Single source of truth** - Event protocol is clearly visible
3. **Easier development** - No context switching between repos
4. **Better testing** - Integration tests see both sides
5. **Simpler setup** - One clone, one build command

## Common Issues

### "boatman binary not found"

**Solution:** Build the CLI first
```bash
cd ~/workspace/personal/boatman-ecosystem
make build-cli
```

### "go.work not found"

**Solution:** Sync the workspace
```bash
cd ~/workspace/personal/boatman-ecosystem
go work sync
```

### "Module not found" errors

**Solution:** Download dependencies
```bash
make deps
```

### Desktop can't find CLI

**Solution:** Check search paths in order:
1. Is `boatman` in PATH? (`which boatman`)
2. Does `cli/boatman` exist? (`ls cli/boatman`)
3. Check desktop logs for actual path being used

## Questions?

Open an issue in the monorepo with the `migration` label.
