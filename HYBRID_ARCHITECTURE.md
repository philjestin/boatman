# Hybrid Architecture: Subprocess + Direct Imports

The Boatman ecosystem uses a **hybrid architecture** that combines the benefits of subprocess execution with direct package imports.

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                    Desktop Application                       │
│                                                              │
│  ┌────────────────────┐         ┌────────────────────┐     │
│  │  Subprocess Layer  │         │  Direct Imports    │     │
│  │                    │         │                    │     │
│  │  • Full execution  │         │  • Diff analysis   │     │
│  │  • Streaming       │         │  • Validation      │     │
│  │  • Process control │         │  • Utilities       │     │
│  └─────────┬──────────┘         └─────────┬──────────┘     │
│            │                              │                 │
│            │ Spawns CLI                   │ Direct call     │
│            ▼                              ▼                 │
└────────────┼──────────────────────────────┼─────────────────┘
             │                              │
             │                              │
┌────────────▼──────────────────────────────▼─────────────────┐
│                      CLI Module                              │
│                                                              │
│  ┌──────────────────┐  ┌──────────────────────────────────┐ │
│  │  cmd/boatman     │  │  pkg/ (public utilities)         │ │
│  │  (executable)    │  │                                  │ │
│  │                  │  │  • pkg/diff                      │ │
│  │  • main()        │  │  • pkg/validation                │ │
│  │  • CLI commands  │  │  • pkg/git                       │ │
│  └──────────────────┘  └──────────────────────────────────┘ │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐   │
│  │  internal/ (private implementation)                  │   │
│  │                                                       │   │
│  │  • internal/agent  • internal/executor               │   │
│  │  • internal/events  • internal/planner               │   │
│  └──────────────────────────────────────────────────────┘   │
│                                                              │
│  Uses ▼                                                      │
└─────────┼────────────────────────────────────────────────────┘
          │
┌─────────▼────────────────────────────────────────────────────┐
│                  Shared Package                              │
│                                                              │
│  • shared/events  - Event types (JSON protocol)             │
│  • shared/types   - Common structs (Results, Usage, etc.)   │
└──────────────────────────────────────────────────────────────┘
```

## When to Use Each Approach

### Subprocess Execution ⚙️

**Use for:**
- User-facing execution (streaming output to UI)
- Long-running operations (need kill/restart capability)
- Full boatmanmode workflows (multi-phase pipelines)
- Process isolation (memory cleanup between runs)
- Multiple parallel agents

**Example:**
```go
// Execute user prompt with streaming
bmIntegration, _ := boatmanmode.NewIntegration("", "", projectPath)
outputChan := make(chan string, 100)

go func() {
    for msg := range outputChan {
        ui.ShowMessage(msg) // Stream to UI
    }
}()

bmIntegration.StreamExecution(ctx, sessionID, prompt, "prompt", outputChan)
```

**Pros:**
- ✅ Streaming output
- ✅ Can kill/restart
- ✅ Process isolation
- ✅ CLI remains standalone
- ✅ User sees progress

**Cons:**
- ❌ Subprocess overhead
- ❌ JSON serialization
- ❌ No compile-time type checking
- ❌ Harder debugging

### Direct Import 📦

**Use for:**
- UI queries (task details, diff stats)
- Quick validation (syntax checks, linting)
- Utility functions (git operations, parsing)
- Performance-critical paths
- Type-safe operations

**Example:**
```go
// Quick diff analysis for task details modal
import "github.com/philjestin/boatmanmode/pkg/diff"

hybrid := services.NewHybrid(projectPath)
diff, _ := hybrid.GetDiff()
stats := hybrid.GetDiffStats(diff)

fmt.Printf("Files: %d, Lines: +%d/-%d\n",
    stats.FilesChanged, stats.LinesAdded, stats.LinesDeleted)
```

**Pros:**
- ✅ No subprocess overhead
- ✅ Type-safe
- ✅ Easier debugging
- ✅ Faster execution
- ✅ Shared types

**Cons:**
- ❌ Tight coupling
- ❌ Can't kill operation
- ❌ Single process memory
- ❌ No streaming UI

## Package Organization

### CLI Packages

```
cli/
├── cmd/boatman/           # CLI entry point (subprocess)
│   └── main.go
│
├── pkg/                   # Public packages (direct import)
│   ├── diff/              # Diff analysis utilities
│   ├── validation/        # Code validation
│   └── git/               # Git operations (future)
│
└── internal/              # Private implementation
    ├── agent/             # Main execution logic
    ├── events/            # Event emission
    ├── executor/          # Code execution
    └── planner/           # Planning phase
```

### Desktop Packages

```
desktop/
├── boatmanmode/           # Subprocess integration (existing)
│   └── integration.go
│
└── services/              # Hybrid services (new)
    ├── boatman_hybrid.go  # Hybrid wrapper
    └── examples.go        # Usage examples
```

### Shared Packages

```
shared/
├── events/                # Event types (JSON protocol)
│   └── events.go
│
└── types/                 # Common types
    └── types.go           # Result structs, Usage, etc.
```

## Migration Strategy

### Phase 1: Add Shared Types (✅ Done)

- Created `shared/` module with common types
- CLI and desktop both import shared types
- Event protocol now type-safe

### Phase 2: Extract Public Utilities (✅ Done)

- Created `cli/pkg/` for public utilities
- Desktop can import directly
- Examples: diff analysis, validation

### Phase 3: Gradual Adoption (🔄 In Progress)

**Start with low-risk features:**
1. Task detail modal (diff stats) → Direct import
2. Pre-commit validation → Direct import
3. Real-time UI updates → Direct import
4. Main execution → Keep subprocess

**Keep existing:**
- User prompt execution → Subprocess
- Ticket workflows → Subprocess
- Streaming output → Subprocess

### Phase 4: Best of Both

**Final state:**
- Subprocess for execution (streaming, control)
- Direct import for utilities (speed, type-safety)
- Shared types everywhere (consistency)

## Implementation Examples

### Example 1: Enhanced Task Details

**Before (No details):**
```typescript
// frontend/src/components/tasks/TaskDetailModal.tsx
// Shows "No details available"
```

**After (Direct import):**
```go
// desktop/app.go
func (a *App) GetTaskDetails(taskID string) (*TaskDetails, error) {
    session, _ := a.agentManager.GetSession(sessionID)

    // Use direct import for fast diff analysis
    hybrid := services.NewHybrid(session.ProjectPath)
    diff, _ := hybrid.GetDiff()
    stats := hybrid.GetDiffStats(diff)

    return &TaskDetails{
        Diff:         diff,
        FilesChanged: stats.FilesChanged,
        LinesAdded:   stats.LinesAdded,
        LinesDeleted: stats.LinesDeleted,
        Summary:      stats.Summary(),
    }, nil
}
```

### Example 2: Pre-Commit Validation

**New feature using direct import:**
```go
// desktop/app.go
func (a *App) ValidateBeforeCommit(files []string) (*ValidationResult, error) {
    hybrid := services.NewHybrid(a.activeProject.Path)

    // Fast validation (no subprocess)
    result, err := hybrid.ValidateFiles(context.Background(), files)
    if err != nil {
        return nil, err
    }

    return &ValidationResult{
        Passed: result.Passed,
        Issues: result.Issues,
        Score:  result.Score,
    }, nil
}
```

### Example 3: Real-Time Progress

**Hybrid approach:**
```go
// Start subprocess for execution
go func() {
    bmIntegration.StreamExecution(ctx, sessionID, prompt, mode, outputChan)
}()

// Poll using direct import for UI
ticker := time.NewTicker(2 * time.Second)
go func() {
    hybrid := services.NewHybrid(projectPath)
    for range ticker.C {
        diff, _ := hybrid.GetDiff()
        stats := hybrid.GetDiffStats(diff)
        updateProgressBar(stats.FilesChanged, stats.Total())
    }
}()
```

## Benefits of This Architecture

### For Users
- ✅ Fast UI (no subprocess overhead for queries)
- ✅ Streaming execution (see progress in real-time)
- ✅ Can kill long-running operations
- ✅ Rich task details (diffs, stats, validation)

### For Developers
- ✅ Type safety (compile-time checking)
- ✅ Easier debugging (single process for utilities)
- ✅ Shared types (consistency)
- ✅ Flexibility (choose right tool for each job)

### For the CLI
- ✅ Remains standalone (users can still use CLI directly)
- ✅ Public API (pkg/) is stable
- ✅ Internal implementation can evolve

## Testing Strategy

### Subprocess Tests
```go
// Test CLI as external process
func TestBoatmanExecution(t *testing.T) {
    cmd := exec.Command("boatman", "work", "--prompt", "test")
    output, _ := cmd.CombinedOutput()
    // Assert output contains expected events
}
```

### Direct Import Tests
```go
// Test utilities directly
func TestDiffAnalysis(t *testing.T) {
    analyzer := diff.New(testRepo)
    stats := analyzer.GetDiffStats(testDiff)
    assert.Equal(t, 5, stats.FilesChanged)
}
```

### Integration Tests
```go
// Test both together
func TestHybridUsage(t *testing.T) {
    // Start subprocess execution
    go executeInSubprocess()

    // Query using direct import
    hybrid := services.NewHybrid(projectPath)
    diff, _ := hybrid.GetDiff()
    // Assert diff is being generated
}
```

## Decision Tree

```
Need to interact with boatmanmode?
│
├─ User-facing execution?
│  ├─ YES → Use subprocess
│  │        - Full workflow
│  │        - Streaming output
│  │        - Can kill/restart
│  │
│  └─ NO → Is it a query/utility?
│           ├─ YES → Use direct import
│           │        - Diff analysis
│           │        - Validation
│           │        - Stats
│           │
│           └─ NO → Use subprocess
│                    - When in doubt, subprocess
│                    - Easier to migrate later
```

## Future Enhancements

### Potential Direct Imports
- `pkg/git/` - Git utilities (diff, status, commit)
- `pkg/cost/` - Cost tracking and analysis
- `pkg/test/` - Test running utilities
- `pkg/format/` - Code formatting

### Shared Types Expansion
- `shared/config/` - Configuration types
- `shared/errors/` - Standard errors
- `shared/proto/` - Protocol buffers (if needed)

## Summary

The hybrid architecture gives us:
1. **Speed** - Direct imports for UI queries
2. **Control** - Subprocess for execution
3. **Type Safety** - Shared types everywhere
4. **Flexibility** - Choose right approach per feature
5. **Independence** - CLI remains standalone tool

This is the best of both worlds! 🚢
