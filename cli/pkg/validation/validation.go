// Package validation provides code validation utilities.
// This package can be imported directly by the desktop app.
package validation

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/philjestin/boatman-ecosystem/shared/types"
)

// Validator provides code validation functionality.
type Validator struct {
	repoPath string
}

// New creates a new validator for the given repository.
func New(repoPath string) *Validator {
	return &Validator{repoPath: repoPath}
}

// ValidateGo runs go vet and returns any issues found.
func (v *Validator) ValidateGo(ctx context.Context, files []string) ([]types.ReviewIssue, error) {
	// Filter for .go files
	goFiles := make([]string, 0)
	for _, f := range files {
		if strings.HasSuffix(f, ".go") {
			goFiles = append(goFiles, f)
		}
	}

	if len(goFiles) == 0 {
		return nil, nil
	}

	// Run go vet
	args := append([]string{"vet"}, goFiles...)
	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = v.repoPath

	output, err := cmd.CombinedOutput()
	if err == nil {
		// No issues found
		return nil, nil
	}

	// Parse vet output
	issues := parseGoVetOutput(string(output))
	return issues, nil
}

// CheckSyntax checks if files have valid syntax.
func (v *Validator) CheckSyntax(ctx context.Context, files []string) ([]types.ReviewIssue, error) {
	var issues []types.ReviewIssue

	for _, file := range files {
		fullPath := filepath.Join(v.repoPath, file)

		// Check if file exists
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			continue
		}

		// Language-specific syntax checks
		switch {
		case strings.HasSuffix(file, ".go"):
			if errs := v.checkGoSyntax(ctx, file); errs != nil {
				issues = append(issues, errs...)
			}
		case strings.HasSuffix(file, ".js") || strings.HasSuffix(file, ".jsx"):
			// Could add JS syntax check
		case strings.HasSuffix(file, ".ts") || strings.HasSuffix(file, ".tsx"):
			// Could add TS syntax check
		}
	}

	return issues, nil
}

// checkGoSyntax checks Go file syntax using go/parser.
func (v *Validator) checkGoSyntax(ctx context.Context, file string) []types.ReviewIssue {
	cmd := exec.CommandContext(ctx, "go", "build", "-o", "/dev/null", file)
	cmd.Dir = v.repoPath

	output, err := cmd.CombinedOutput()
	if err == nil {
		return nil
	}

	// Parse build errors
	return parseGoBuildOutput(string(output), file)
}

// ValidateAll runs all validators and returns combined results.
func (v *Validator) ValidateAll(ctx context.Context, files []string) (*types.ReviewResult, error) {
	var allIssues []types.ReviewIssue

	// Run syntax check
	syntaxIssues, err := v.CheckSyntax(ctx, files)
	if err != nil {
		return nil, fmt.Errorf("syntax check failed: %w", err)
	}
	allIssues = append(allIssues, syntaxIssues...)

	// Run go vet
	vetIssues, err := v.ValidateGo(ctx, files)
	if err != nil {
		return nil, fmt.Errorf("go vet failed: %w", err)
	}
	allIssues = append(allIssues, vetIssues...)

	// Calculate score (simple: 100 - (10 * error count) - (5 * warning count))
	score := 100
	for _, issue := range allIssues {
		if issue.Severity == "error" {
			score -= 10
		} else if issue.Severity == "warning" {
			score -= 5
		}
	}
	if score < 0 {
		score = 0
	}

	return &types.ReviewResult{
		Passed:  len(allIssues) == 0,
		Summary: fmt.Sprintf("Found %d issues", len(allIssues)),
		Issues:  allIssues,
		Score:   score,
	}, nil
}

// parseGoVetOutput parses go vet output into ReviewIssues.
func parseGoVetOutput(output string) []types.ReviewIssue {
	var issues []types.ReviewIssue

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Format: file.go:line:col: message
		parts := strings.SplitN(line, ":", 4)
		if len(parts) < 3 {
			continue
		}

		issue := types.ReviewIssue{
			File:     parts[0],
			Severity: "warning",
			Message:  strings.TrimSpace(strings.Join(parts[2:], ":")),
		}

		issues = append(issues, issue)
	}

	return issues
}

// parseGoBuildOutput parses go build output into ReviewIssues.
func parseGoBuildOutput(output, file string) []types.ReviewIssue {
	var issues []types.ReviewIssue

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || !strings.Contains(line, file) {
			continue
		}

		issue := types.ReviewIssue{
			File:     file,
			Severity: "error",
			Message:  line,
		}

		issues = append(issues, issue)
	}

	return issues
}
