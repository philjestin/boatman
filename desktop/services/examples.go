package services

import (
	"context"
	"fmt"

	// Subprocess integration (existing)
	"boatman/boatmanmode"
)

// HybridUsageExamples demonstrates when to use each approach.

// Example 1: Use SUBPROCESS for user-facing execution
// - User sees streaming output
// - Can kill/restart execution
// - Process isolation
func ExecuteUserPromptExample(sessionID, projectPath, prompt string) error {
	// This uses the existing subprocess integration
	bmIntegration, err := boatmanmode.NewIntegration("", "", projectPath)
	if err != nil {
		return err
	}

	outputChan := make(chan string, 100)
	go func() {
		for msg := range outputChan {
			fmt.Println(msg) // Show to user in UI
		}
	}()

	// Subprocess execution with streaming
	_, err = bmIntegration.StreamExecution(
		context.Background(),
		sessionID,
		prompt,
		"prompt",
		outputChan,
		nil,
	)
	return err
}

// Example 2: Use DIRECT IMPORT for task details
// - Fast, no subprocess overhead
// - Type-safe
// - Perfect for UI queries
func GetTaskDetailsExample(projectPath, taskID string) (map[string]interface{}, error) {
	hybrid := NewHybrid(projectPath)

	// Fast diff analysis (no subprocess)
	diff, err := hybrid.GetDiff()
	if err != nil {
		return nil, err
	}

	stats := hybrid.GetDiffStats(diff)

	return map[string]interface{}{
		"diff":          diff,
		"files_changed": stats.FilesChanged,
		"lines_added":   stats.LinesAdded,
		"lines_deleted": stats.LinesDeleted,
		"summary":       stats.Summary(),
	}, nil
}

// Example 3: Use DIRECT IMPORT for pre-flight validation
// - Quick syntax check before committing
// - No subprocess overhead
// - Immediate feedback
func ValidateBeforeCommitExample(projectPath string, files []string) error {
	hybrid := NewHybrid(projectPath)

	// Quick validation (no subprocess)
	result, err := hybrid.ValidateFiles(context.Background(), files)
	if err != nil {
		return err
	}

	if !result.Passed {
		fmt.Printf("Found %d issues:\n", len(result.Issues))
		for _, issue := range result.Issues {
			fmt.Printf("  %s:%d: [%s] %s\n",
				issue.File, issue.Line, issue.Severity, issue.Message)
		}
		return fmt.Errorf("validation failed")
	}

	return nil
}

// Example 4: Use SUBPROCESS for multi-step workflows
// - Full boatmanmode pipeline
// - Streaming updates
// - User can monitor progress
func ExecuteTicketWorkflowExample(sessionID, projectPath, ticketID string) error {
	bmIntegration, err := boatmanmode.NewIntegration("linear-api-key", "claude-api-key", projectPath)
	if err != nil {
		return err
	}

	// This runs the full ticket workflow (plan → execute → review → refactor)
	// Uses subprocess for streaming and isolation
	result, err := bmIntegration.ExecuteTicket(context.Background(), ticketID)
	if err != nil {
		return err
	}

	fmt.Printf("Ticket execution result: %+v\n", result)
	return nil
}

// Example 5: Hybrid approach for real-time UI updates
// - Use subprocess for execution
// - Use direct import for UI queries
func HybridExecutionExample(sessionID, projectPath, prompt string) error {
	// Start execution in subprocess (streaming, killable)
	bmIntegration, _ := boatmanmode.NewIntegration("", "", projectPath)
	outputChan := make(chan string, 100)

	go func() {
		bmIntegration.StreamExecution(
			context.Background(),
			sessionID,
			prompt,
			"prompt",
			outputChan,
			nil,
		)
	}()

	// Meanwhile, use direct import for UI queries
	hybrid := NewHybrid(projectPath)

	// Show real-time diff stats as execution progresses
	ticker := make(chan bool)
	go func() {
		for range ticker {
			diff, _ := hybrid.GetDiff()
			if diff != "" {
				stats := hybrid.GetDiffStats(diff)
				fmt.Printf("Progress: %s\n", stats.Summary())
			}
		}
	}()

	// Wait for completion
	for msg := range outputChan {
		fmt.Println(msg)
	}

	return nil
}

/*
DECISION GUIDE: When to use which approach?

┌─────────────────────────────────────────────────────────────┐
│                        SUBPROCESS                            │
├─────────────────────────────────────────────────────────────┤
│ ✓ User-facing execution (show streaming output)            │
│ ✓ Long-running operations (need to kill/restart)           │
│ ✓ Full boatmanmode workflow (plan → execute → review)      │
│ ✓ Process isolation (memory cleanup)                       │
│ ✓ Multiple parallel agents                                 │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                      DIRECT IMPORT                           │
├─────────────────────────────────────────────────────────────┤
│ ✓ UI queries (task details, diff stats)                    │
│ ✓ Quick validation (syntax check, linting)                 │
│ ✓ Utilities (git operations, parsing)                      │
│ ✓ Performance critical (no subprocess overhead)            │
│ ✓ Type safety (compile-time checking)                      │
└─────────────────────────────────────────────────────────────┘

EXAMPLES:
- Opening task detail modal → Direct import (fast, type-safe)
- Executing user prompt → Subprocess (streaming, killable)
- Pre-commit validation → Direct import (quick feedback)
- Full ticket workflow → Subprocess (multi-phase, streaming)
- Real-time diff stats → Direct import (polling for UI)
*/
