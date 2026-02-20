// Package review provides canonical review types and the Reviewer interface
// for pluggable code review backends.
package review

import "context"

// Issue is the canonical review issue type, unifying scottbott.Issue
// and shared/types.ReviewIssue.
type Issue struct {
	Severity    string `json:"severity"`              // "critical", "major", "minor"
	File        string `json:"file"`
	Line        int    `json:"line,omitempty"`
	Description string `json:"description"`
	Suggestion  string `json:"suggestion,omitempty"`
	Code        string `json:"code,omitempty"`
}

// ReviewResult is the outcome of a code review.
type ReviewResult struct {
	Passed   bool     `json:"passed"`
	Score    int      `json:"score"`
	Summary  string   `json:"summary"`
	Issues   []Issue  `json:"issues"`
	Praise   []string `json:"praise,omitempty"`
	Guidance string   `json:"guidance,omitempty"`
}

// Reviewer is the interface for pluggable code review backends.
// Implementations can use Claude, OpenAI, Gemini, static analysis, etc.
type Reviewer interface {
	Review(ctx context.Context, diff string, context string) (*ReviewResult, error)
}
