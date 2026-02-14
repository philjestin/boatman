# Contributing to Boatman Ecosystem

Thank you for your interest in contributing! This guide will help you get started.

## Repository Structure

This is a monorepo containing two main components:

- **`cli/`** - The autonomous development CLI (boatmanmode)
- **`desktop/`** - The desktop GUI wrapper (boatmanapp)

## Getting Started

### Prerequisites

- Go 1.24.1 or later
- Node.js 18+ (for desktop development)
- Wails v2 (for desktop development)
- Claude CLI or Anthropic API key

### Initial Setup

```bash
# Clone the repository
git clone <repo-url> boatman-ecosystem
cd boatman-ecosystem

# Run setup script
./scripts/setup.sh

# Or manually:
go work sync
make build-cli
cd desktop/frontend && npm install
```

## Development Workflow

### Working on the CLI

```bash
# Make changes in cli/
cd cli

# Run tests
go test ./...

# Build
go build -o boatman ./cmd/boatman

# Test locally
./boatman work --prompt "test prompt"
```

### Working on the Desktop App

```bash
# Make changes in desktop/

# Build CLI first (desktop depends on it)
make build-cli

# Start dev mode with hot reload
cd desktop
wails dev

# Or from root
make dev
```

### Working on Both (Event Protocol Changes)

If you're modifying the event protocol that connects CLI and desktop:

1. **Update CLI event emission** (`cli/internal/events/emitter.go`)
2. **Update desktop event handling** (`desktop/app.go`, `desktop/frontend/src/hooks/useAgent.ts`)
3. **Test both sides**:
   ```bash
   # Rebuild CLI
   make build-cli

   # Test in desktop dev mode
   make dev
   ```
4. **Update event documentation** in both component READMEs

## Testing

### CLI Tests

```bash
cd cli
go test ./...

# With coverage
go test -cover ./...

# Specific package
go test ./internal/agent
```

### Desktop Tests

```bash
# Go backend tests
cd desktop
go test ./...

# Frontend tests
cd desktop/frontend
npm test
```

### Integration Testing

To test the full CLI → Desktop integration:

1. Build the CLI: `make build-cli`
2. Start desktop in dev mode: `make dev`
3. Create a boatmanmode session and verify events flow through

## Code Style

### Go

- Follow standard Go conventions
- Run `go fmt` before committing
- Run `golangci-lint run` if available
- Keep functions focused and well-documented

### TypeScript/React

- Use TypeScript for all new code
- Follow React hooks best practices
- Keep components small and focused
- Use meaningful variable names

## Commit Messages

Use clear, descriptive commit messages:

```
[cli] Add metadata to agent_completed events

- Added AgentCompletedWithData() function to emitter
- Updated execution, review, and refactor phases to emit metadata
- Includes diffs, feedback, and plan data for desktop UI

Fixes #123
```

Prefixes:
- `[cli]` - Changes to the CLI component
- `[desktop]` - Changes to the desktop component
- `[both]` - Changes affecting both components
- `[docs]` - Documentation only changes
- `[build]` - Build system or tooling changes

## Pull Request Process

1. **Create a feature branch** from `main`
2. **Make your changes** with clear commits
3. **Test thoroughly**:
   - Run tests: `make test-all`
   - Test manually in dev mode
4. **Update documentation** if needed
5. **Create a PR** with:
   - Clear description of changes
   - Why the change is needed
   - How to test it
   - Screenshots for UI changes

## Event Protocol Guidelines

The CLI and desktop communicate via JSON events. When modifying this protocol:

### Adding a New Event Type

1. **Define in CLI** (`cli/internal/events/emitter.go`):
   ```go
   func MyNewEvent(id, name string) {
       Emit(Event{
           Type: "my_new_event",
           ID:   id,
           Name: name,
       })
   }
   ```

2. **Emit in CLI** (`cli/internal/agent/agent.go`):
   ```go
   events.MyNewEvent("agent-123", "My Agent")
   ```

3. **Handle in Desktop** (`desktop/app.go`):
   ```go
   case "my_new_event":
       id, _ := eventData["id"].(string)
       name, _ := eventData["name"].(string)
       // Handle the event
   ```

4. **Update UI** (`desktop/frontend/src/hooks/useAgent.ts`):
   ```typescript
   const eventHandler = (data: BoatmanModeEventPayload) => {
       if (data.event.type === 'my_new_event') {
           // Handle in React
       }
   }
   ```

5. **Document** in event protocol docs

### Event Schema Rules

- Events must have a `type` field (string)
- Use snake_case for event types: `agent_started`, `task_completed`
- Optional fields: `id`, `name`, `description`, `status`, `message`, `data`
- The `data` field is a free-form object for phase-specific metadata
- Always emit events to stdout as newline-delimited JSON

## Common Tasks

### Adding a New CLI Feature

1. Implement in `cli/internal/`
2. Wire up to command in `cli/cmd/boatman/`
3. Add tests
4. Update CLI README

### Adding a New Desktop Feature

1. Add Go backend logic in `desktop/`
2. Create/update React components in `desktop/frontend/src/`
3. Wire up Wails bindings if needed
4. Update desktop README

### Debugging

**CLI debugging:**
```bash
# Enable verbose output
export BOATMAN_DEBUG=1
./cli/boatman work --prompt "test"

# Check emitted events
./cli/boatman work --prompt "test" | grep '"type":'
```

**Desktop debugging:**
```bash
# Check browser console in Wails dev mode
make dev

# Check Go backend logs in terminal

# Enable verbose Wails logging
wails dev -v
```

## Release Process

(To be defined - coordinated releases vs independent releases)

## Questions?

- Open an issue for bugs or feature requests
- Start a discussion for questions or ideas
- Check existing issues and PRs first

## License

By contributing, you agree that your contributions will be licensed under the same license as the project.
