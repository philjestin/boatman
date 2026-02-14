// Package services provides hybrid integration with boatmanmode.
// Uses subprocess for main execution but direct imports for utilities.
package services

import (
	"context"

	// Direct imports from CLI packages
	clidiff "github.com/philjestin/boatmanmode/pkg/diff"
	"github.com/philjestin/boatmanmode/pkg/validation"

	// Shared types
	"github.com/philjestin/boatman-ecosystem/shared/types"
)

// BoatmanHybrid provides hybrid access to boatmanmode functionality.
// - Uses subprocess for full execution (streaming, isolation, killable)
// - Uses direct imports for utilities (fast, type-safe, no overhead)
type BoatmanHybrid struct {
	repoPath string

	// Direct utility instances
	diffAnalyzer *clidiff.Analyzer
	validator    *validation.Validator
}

// NewHybrid creates a new hybrid boatman service.
func NewHybrid(repoPath string) *BoatmanHybrid {
	return &BoatmanHybrid{
		repoPath:     repoPath,
		diffAnalyzer: clidiff.New(repoPath),
		validator:    validation.New(repoPath),
	}
}

// ============================================================================
// Direct Import Methods (Fast, No Subprocess)
// ============================================================================

// GetDiff gets the current git diff using direct import (no subprocess).
func (h *BoatmanHybrid) GetDiff() (string, error) {
	return h.diffAnalyzer.GetDiff()
}

// GetDiffStats analyzes a diff and returns statistics.
func (h *BoatmanHybrid) GetDiffStats(diffContent string) *clidiff.DiffStats {
	return clidiff.ParseDiff(diffContent)
}

// ValidateFiles validates files using direct import (no subprocess).
func (h *BoatmanHybrid) ValidateFiles(ctx context.Context, files []string) (*types.ReviewResult, error) {
	return h.validator.ValidateAll(ctx, files)
}

// QuickSyntaxCheck performs a quick syntax check (no subprocess).
func (h *BoatmanHybrid) QuickSyntaxCheck(ctx context.Context, files []string) ([]types.ReviewIssue, error) {
	return h.validator.CheckSyntax(ctx, files)
}

// ============================================================================
// Usage Examples
// ============================================================================

/*
Example 1: Get diff stats for task details modal

	hybrid := NewHybrid(projectPath)
	diff, _ := hybrid.GetDiff()
	stats := hybrid.GetDiffStats(diff)

	fmt.Printf("Files changed: %d, Lines: +%d/-%d\n",
		stats.FilesChanged, stats.LinesAdded, stats.LinesDeleted)

Example 2: Validate code before committing

	issues, err := hybrid.ValidateFiles(ctx, changedFiles)
	if err != nil {
		return err
	}
	if !issues.Passed {
		fmt.Printf("Found %d issues\n", len(issues.Issues))
	}

Example 3: Quick syntax check on file save

	issues, _ := hybrid.QuickSyntaxCheck(ctx, []string{"main.go"})
	for _, issue := range issues {
		fmt.Printf("%s:%d: %s\n", issue.File, issue.Line, issue.Message)
	}

*/
