package triage

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/philjestin/boatmanmode/internal/linear"
)

// FormatTriageComment formats a Classification as a Linear-compatible markdown comment.
func FormatTriageComment(c Classification) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("## AI Triage: %s\n\n", c.Category))

	sb.WriteString("| Dimension | Score | |\n")
	sb.WriteString("|-----------|-------|-|\n")
	sb.WriteString(fmt.Sprintf("| Clarity | %d/5 | %s |\n", c.Rubric.Clarity, positiveMarker(c.Rubric.Clarity)))
	sb.WriteString(fmt.Sprintf("| Code Locality | %d/5 | %s |\n", c.Rubric.CodeLocality, positiveMarker(c.Rubric.CodeLocality)))
	sb.WriteString(fmt.Sprintf("| Pattern Match | %d/5 | %s |\n", c.Rubric.PatternMatch, positiveMarker(c.Rubric.PatternMatch)))
	sb.WriteString(fmt.Sprintf("| Validation Strength | %d/5 | %s |\n", c.Rubric.ValidationStrength, positiveMarker(c.Rubric.ValidationStrength)))
	sb.WriteString(fmt.Sprintf("| Dependency Risk | %d/5 | %s |\n", c.Rubric.DependencyRisk, negativeMarker(c.Rubric.DependencyRisk)))
	sb.WriteString(fmt.Sprintf("| Product Ambiguity | %d/5 | %s |\n", c.Rubric.ProductAmbiguity, negativeMarker(c.Rubric.ProductAmbiguity)))
	sb.WriteString(fmt.Sprintf("| Blast Radius | %d/5 | %s |\n", c.Rubric.BlastRadius, negativeMarker(c.Rubric.BlastRadius)))

	sb.WriteString("\n### Gate Results\n")
	for _, g := range c.GateResults {
		if g.Passed {
			sb.WriteString(fmt.Sprintf("- %s: pass\n", g.Gate))
		} else {
			reason := g.Reason
			if reason == "" {
				reason = "failed"
			}
			sb.WriteString(fmt.Sprintf("- %s: FAIL — %s\n", g.Gate, reason))
		}
	}

	if len(c.Reasons) > 0 {
		sb.WriteString("\n### Reasons\n")
		for _, r := range c.Reasons {
			sb.WriteString(fmt.Sprintf("- %s\n", r))
		}
	}

	if len(c.UncertainAxes) > 0 {
		sb.WriteString("\n### Uncertain Axes\n")
		for _, axis := range c.UncertainAxes {
			sb.WriteString(fmt.Sprintf("- %s\n", axis))
		}
	}

	if len(c.HardStops) > 0 {
		sb.WriteString("\n### Hard Stops\n")
		for _, stop := range c.HardStops {
			sb.WriteString(fmt.Sprintf("- %s\n", stop))
		}
	}

	sb.WriteString(fmt.Sprintf("\n_Triaged by Boatman at %s_\n", time.Now().UTC().Format(time.RFC3339)))

	return sb.String()
}

// positiveMarker returns a text marker for positive dimensions (higher is better).
func positiveMarker(score int) string {
	if score >= 3 {
		return "(good)"
	}
	return ""
}

// negativeMarker returns a text marker for negative dimensions (lower is better).
func negativeMarker(score int) string {
	if score <= 2 {
		return "(low risk)"
	}
	return ""
}

// PostTriageComments posts triage comment to each classified ticket in Linear.
// ticketUUIDs maps human-readable identifiers (e.g., "ENG-123") to Linear UUIDs.
func PostTriageComments(ctx context.Context, client *linear.Client, classifications []Classification, ticketUUIDs map[string]string) error {
	var errs []string

	for _, c := range classifications {
		uuid, ok := ticketUUIDs[c.TicketID]
		if !ok {
			errs = append(errs, fmt.Sprintf("no UUID mapping for ticket %s", c.TicketID))
			continue
		}

		comment := FormatTriageComment(c)
		if err := client.AddComment(ctx, uuid, comment); err != nil {
			errs = append(errs, fmt.Sprintf("failed to post comment for %s: %v", c.TicketID, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors posting triage comments: %s", strings.Join(errs, "; "))
	}

	return nil
}
