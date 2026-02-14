// Package diff provides utilities for working with git diffs.
// This package can be imported directly by the desktop app.
package diff

import (
	"fmt"
	"os/exec"
	"strings"
)

// Analyzer provides diff analysis utilities.
type Analyzer struct {
	repoPath string
}

// New creates a new diff analyzer for the given repository.
func New(repoPath string) *Analyzer {
	return &Analyzer{repoPath: repoPath}
}

// GetDiff returns the current git diff.
func (a *Analyzer) GetDiff() (string, error) {
	cmd := exec.Command("git", "diff", "HEAD")
	cmd.Dir = a.repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get diff: %w", err)
	}
	return string(output), nil
}

// GetStagedDiff returns the diff of staged changes.
func (a *Analyzer) GetStagedDiff() (string, error) {
	cmd := exec.Command("git", "diff", "--cached")
	cmd.Dir = a.repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get staged diff: %w", err)
	}
	return string(output), nil
}

// ParseDiff parses a diff string and returns statistics.
func ParseDiff(diff string) *DiffStats {
	stats := &DiffStats{}

	lines := strings.Split(diff, "\n")
	var currentFile string

	for _, line := range lines {
		if strings.HasPrefix(line, "diff --git") {
			stats.FilesChanged++
			// Extract filename
			parts := strings.Fields(line)
			if len(parts) >= 4 {
				currentFile = strings.TrimPrefix(parts[3], "b/")
			}
		} else if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
			stats.LinesAdded++
		} else if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
			stats.LinesDeleted++
		} else if strings.HasPrefix(line, "new file") {
			stats.FilesAdded++
			if currentFile != "" {
				stats.Files = append(stats.Files, currentFile)
			}
		} else if strings.HasPrefix(line, "deleted file") {
			stats.FilesDeleted++
		}
	}

	return stats
}

// DiffStats contains statistics about a diff.
type DiffStats struct {
	FilesChanged int      `json:"files_changed"`
	FilesAdded   int      `json:"files_added"`
	FilesDeleted int      `json:"files_deleted"`
	LinesAdded   int      `json:"lines_added"`
	LinesDeleted int      `json:"lines_deleted"`
	Files        []string `json:"files"`
}

// Total returns the total number of lines changed.
func (d *DiffStats) Total() int {
	return d.LinesAdded + d.LinesDeleted
}

// Summary returns a human-readable summary.
func (d *DiffStats) Summary() string {
	return fmt.Sprintf("%d files changed, +%d/-%d lines",
		d.FilesChanged, d.LinesAdded, d.LinesDeleted)
}
