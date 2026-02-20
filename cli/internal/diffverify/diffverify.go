// Package diffverify provides an adapter wrapping the harness diffverify
// with coordinator integration and scottbott.Issue conversion.
package diffverify

import (
	"context"
	"fmt"

	"github.com/philjestin/boatman-ecosystem/harness/diffverify"
	"github.com/philjestin/boatman-ecosystem/harness/review"
	"github.com/philjestin/boatmanmode/internal/coordinator"
	"github.com/philjestin/boatmanmode/internal/scottbott"
)

// Type aliases from harness
type VerificationResult = diffverify.VerificationResult
type AddressedIssue = diffverify.AddressedIssue
type UnaddressedIssue = diffverify.UnaddressedIssue
type DiffChange = diffverify.DiffChange
type VerificationHandoff = diffverify.VerificationHandoff
type VerifyHandoff = diffverify.VerifyHandoff

// Agent wraps the harness Verifier with coordinator integration.
type Agent struct {
	inner *diffverify.Verifier
	coord *coordinator.Coordinator
}

// New creates a new Agent.
func New(worktreePath string) *Agent {
	return &Agent{
		inner: diffverify.New(worktreePath),
	}
}

// SetCoordinator sets the coordinator for work claiming.
func (a *Agent) SetCoordinator(c *coordinator.Coordinator) {
	a.coord = c
}

// SetMinConfidence sets the minimum confidence override.
func (a *Agent) SetMinConfidence(minConfidence int) {
	a.inner.SetMinConfidence(minConfidence)
}

// Verify converts scottbott issues and delegates to the harness verifier.
func (a *Agent) Verify(ctx context.Context, issues []scottbott.Issue, oldDiff, newDiff string) (*VerificationResult, error) {
	// Claim work if coordinator is available
	if a.coord != nil {
		claim := &coordinator.WorkClaim{
			WorkID:      "verify-diff",
			WorkType:    "verification",
			Description: fmt.Sprintf("Verifying %d issues", len(issues)),
		}
		a.coord.ClaimWork("verifier", claim)
		defer a.coord.ReleaseWork("verify-diff", "verifier")
	}

	// Convert scottbott issues to review issues
	reviewIssues := make([]review.Issue, len(issues))
	for i, issue := range issues {
		reviewIssues[i] = review.Issue{
			Severity:    issue.Severity,
			File:        issue.File,
			Line:        issue.Line,
			Description: issue.Description,
			Suggestion:  issue.Suggestion,
		}
	}

	result, err := a.inner.Verify(ctx, reviewIssues, oldDiff, newDiff)

	// Set context if coordinator is available
	if a.coord != nil && result != nil {
		a.coord.SetContext("verification_result", result)
	}

	return result, err
}
