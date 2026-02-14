# Quick Start Guide

Get up and running with the Boatman ecosystem in 5 minutes.

## Prerequisites

- Go 1.24.1+
- Node.js 18+
- Wails v2 (for desktop): `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
- Claude CLI or Anthropic API key

## Setup

### 1. Clone and Setup

```bash
# Clone the repository
git clone <repo-url> boatman-ecosystem
cd boatman-ecosystem

# Run setup script (handles everything)
./scripts/setup.sh
```

Or manually:

```bash
# Sync Go workspace
go work sync

# Build CLI
make build-cli

# Install frontend dependencies
cd desktop/frontend && npm install
```

### 2. Configuration

**For CLI:**
```bash
# Copy example config
cp cli/.boatman.example.yaml cli/.boatman.yaml

# Edit with your settings
vim cli/.boatman.yaml
```

**For Desktop:**
The desktop app will prompt for configuration on first run.

## Usage

### CLI

```bash
# Build
make build-cli

# Execute a prompt
./cli/boatman work --prompt "Add user authentication"

# Execute a Linear ticket
./cli/boatman execute --ticket ENG-123

# Install to PATH
make install-cli
boatman work --prompt "your task"
```

### Desktop

```bash
# Dev mode (hot reload)
make dev

# Or manually
cd desktop
wails dev

# Build production app
make build-desktop
```

## Your First Task

### CLI

```bash
cd cli
./boatman work --prompt "Add a hello world function to main.go"
```

This will:
1. Create a git worktree
2. Plan the implementation
3. Execute the changes
4. Review the code
5. Run tests
6. Create a commit

### Desktop

1. Start the app: `make dev`
2. Open a project
3. Click "Boatman Mode"
4. Enter a prompt
5. Watch the execution in real-time
6. Click on tasks to see details (diffs, feedback, etc.)

## Project Structure

```
boatman-ecosystem/
├── cli/              # Autonomous CLI tool
│   ├── cmd/          # CLI commands
│   ├── pkg/          # Public utilities (for desktop)
│   └── internal/     # Private implementation
│
├── desktop/          # GUI application
│   ├── frontend/     # React UI
│   ├── services/     # Hybrid integration
│   └── boatmanmode/  # Subprocess integration
│
├── shared/           # Shared types
│   ├── events/       # Event protocol
│   └── types/        # Common structs
│
└── scripts/          # Setup and utilities
```

## Common Tasks

```bash
# Build everything
make build-all

# Run tests
make test-all

# Clean build artifacts
make clean

# Format code
make fmt

# Show all commands
make help
```

## Development Workflow

### Working on CLI

```bash
cd cli

# Make changes
vim internal/agent/agent.go

# Test
go test ./...

# Build
go build -o boatman ./cmd/boatman

# Test locally
./boatman work --prompt "test"
```

### Working on Desktop

```bash
# Build CLI first (desktop depends on it)
make build-cli

# Start dev mode
cd desktop
wails dev

# Frontend changes hot reload automatically
# Go changes require restart
```

### Working on Both (Event Protocol)

When modifying the event protocol:

```bash
# 1. Update shared types
vim shared/events/events.go

# 2. Update CLI emission
vim cli/internal/events/emitter.go

# 3. Update desktop handling
vim desktop/app.go

# 4. Update frontend
vim desktop/frontend/src/hooks/useAgent.ts

# 5. Test both
make build-cli && make dev
```

## Hybrid Architecture

The desktop can use two approaches:

### Subprocess (for execution)

```go
// Full execution with streaming
bmIntegration.StreamExecution(ctx, sessionID, prompt, "prompt", outputChan)
```

### Direct Import (for utilities)

```go
// Fast diff analysis
import "github.com/philjestin/boatmanmode/pkg/diff"
hybrid := services.NewHybrid(projectPath)
stats := hybrid.GetDiffStats(diff)
```

See [HYBRID_ARCHITECTURE.md](./HYBRID_ARCHITECTURE.md) for when to use each.

## Troubleshooting

### "boatman binary not found"

```bash
# Build it
make build-cli

# Or install to PATH
make install-cli
```

### "Module not found" errors

```bash
# Sync workspace
go work sync

# Tidy modules
cd cli && go mod tidy
cd ../desktop && go mod tidy
```

### Desktop can't find CLI

The desktop looks for the CLI in this order:
1. `boatman` in PATH
2. `~/workspace/personal/boatman-ecosystem/cli/boatman`
3. `~/workspace/handshake/boatmanmode/boatman` (old location)

Build the CLI to fix: `make build-cli`

### Frontend build errors

```bash
cd desktop/frontend
rm -rf node_modules package-lock.json
npm install
```

## Next Steps

- Read [CONTRIBUTING.md](./CONTRIBUTING.md) for development guidelines
- Read [HYBRID_ARCHITECTURE.md](./HYBRID_ARCHITECTURE.md) for architecture details
- Check [desktop/services/examples.go](./desktop/services/examples.go) for usage examples
- Look at [cli/README.md](./cli/README.md) for CLI documentation

## Getting Help

- Check existing issues
- Read the documentation
- Look at example code
- Ask in discussions

## Quick Commands Reference

```bash
# Setup
./scripts/setup.sh              # Initial setup
go work sync                    # Sync workspace

# Building
make build-cli                  # Build CLI only
make build-desktop              # Build desktop only
make build-all                  # Build everything

# Development
make dev                        # Start desktop in dev mode
make install-cli                # Install CLI to ~/bin

# Testing
make test-cli                   # Test CLI
make test-desktop               # Test desktop
make test-all                   # Test everything

# Utilities
make clean                      # Clean build artifacts
make fmt                        # Format code
make help                       # Show all commands
```

Happy shipping! 🚢
