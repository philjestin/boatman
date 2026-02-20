// Package issuetracker provides an adapter wrapping the harness issuetracker
// with scottbott.Issue <-> review.Issue conversion.
package issuetracker

import (
	"github.com/philjestin/boatman-ecosystem/harness/issuetracker"
	"github.com/philjestin/boatman-ecosystem/harness/review"
	"github.com/philjestin/boatmanmode/internal/scottbott"
)

// Type aliases from harness
type TrackedIssue = issuetracker.TrackedIssue
type IssueStats = issuetracker.IssueStats
type IssueHistory = issuetracker.IssueHistory
type IterationRecord = issuetracker.IterationRecord

// IssueTracker wraps the harness issue tracker with scottbott.Issue conversion.
type IssueTracker struct {
	inner *issuetracker.IssueTracker
}

// New creates a new IssueTracker.
func New() *IssueTracker {
	return &IssueTracker{inner: issuetracker.New()}
}

// NextIteration advances to the next iteration.
func (t *IssueTracker) NextIteration() {
	t.inner.NextIteration()
}

// Track converts scottbott issues to review issues and tracks them.
func (t *IssueTracker) Track(issues []scottbott.Issue) []TrackedIssue {
	reviewIssues := make([]review.Issue, len(issues))
	for i, issue := range issues {
		reviewIssues[i] = ScottbottToReview(issue)
	}
	return t.inner.Track(reviewIssues)
}

// GetNewIssues returns issues first seen in the current iteration.
func (t *IssueTracker) GetNewIssues() []TrackedIssue {
	return t.inner.GetNewIssues()
}

// GetPersistentIssues returns issues seen in multiple iterations.
func (t *IssueTracker) GetPersistentIssues() []TrackedIssue {
	return t.inner.GetPersistentIssues()
}

// GetAddressedIssues returns issues that have been fixed.
func (t *IssueTracker) GetAddressedIssues() []TrackedIssue {
	return t.inner.GetAddressedIssues()
}

// GetUnaddressedIssues returns issues not yet fixed.
func (t *IssueTracker) GetUnaddressedIssues() []TrackedIssue {
	return t.inner.GetUnaddressedIssues()
}

// GetCriticalIssues returns critical severity issues.
func (t *IssueTracker) GetCriticalIssues() []TrackedIssue {
	return t.inner.GetCriticalIssues()
}

// Stats returns issue tracking statistics.
func (t *IssueTracker) Stats() IssueStats {
	return t.inner.Stats()
}

// FormatIssues delegates to the harness package.
var FormatIssues = issuetracker.FormatIssues

// NewIssueHistory creates a new issue history tracker.
func NewIssueHistory() *IssueHistoryAdapter {
	return &IssueHistoryAdapter{inner: issuetracker.NewIssueHistory()}
}

// IssueHistoryAdapter wraps harness IssueHistory with scottbott conversion.
type IssueHistoryAdapter struct {
	inner *issuetracker.IssueHistory
}

// RecordIteration converts scottbott issues and records them.
func (h *IssueHistoryAdapter) RecordIteration(issues []scottbott.Issue) []TrackedIssue {
	reviewIssues := make([]review.Issue, len(issues))
	for i, issue := range issues {
		reviewIssues[i] = ScottbottToReview(issue)
	}
	return h.inner.RecordIteration(reviewIssues)
}

// FormatHistory returns formatted history.
func (h *IssueHistoryAdapter) FormatHistory() string {
	return h.inner.FormatHistory()
}

// ScottbottToReview converts a scottbott.Issue to a review.Issue.
func ScottbottToReview(issue scottbott.Issue) review.Issue {
	return review.Issue{
		Severity:    issue.Severity,
		File:        issue.File,
		Line:        issue.Line,
		Description: issue.Description,
		Suggestion:  issue.Suggestion,
	}
}

// ReviewToScottbott converts a review.Issue to a scottbott.Issue.
func ReviewToScottbott(issue review.Issue) scottbott.Issue {
	return scottbott.Issue{
		Severity:    issue.Severity,
		File:        issue.File,
		Line:        issue.Line,
		Description: issue.Description,
		Suggestion:  issue.Suggestion,
	}
}
