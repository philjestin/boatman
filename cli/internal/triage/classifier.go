package triage

import "strings"

// Decision tree gate thresholds (ADR-004). Adjust these to tune classification sensitivity.
const (
	clarityGate          = 2
	blastRadiusGate      = 3
	productAmbiguityGate = 3
	dependencyGate       = 3
)

// Hard-stop keyword groups. If any keyword in a group matches the ticket's
// title, description, or labels, the ticket is classified HUMAN_ONLY.
var hardStopKeywords = map[string][]string{
	"payments/billing": {
		"payments", "billing", "pricing", "stripe", "invoice", "subscription",
	},
	"authentication/authorization": {
		"authentication", "authorization", "authn", "authz", "permissions", "rbac", "oauth", "sso",
	},
	"migration/data repair": {
		"migration", "migrate", "backfill", "data repair",
	},
	"public API contracts": {
		"public api", "api contract", "breaking change", "api version",
	},
	"incident/hotfix": {
		"incident", "sev1", "sev2", "hotfix", "production fix",
	},
	"legal/compliance": {
		"legal", "compliance", "gdpr", "pii", "security vulnerability",
	},
	"multi-repo": {
		"multi-repo", "cross-repo", "monorepo boundary",
	},
}

// Soft-stop keyword groups. These trigger HUMAN_REVIEW_REQUIRED rather than HUMAN_ONLY.
var softStopKeywords = map[string][]string{
	"feature flag/staged rollout": {
		"feature flag", "staged rollout", "canary", "gradual rollout",
	},
	"deploy coordination": {
		"deploy coordination", "release train",
	},
}

// Classify applies the ADR-004 decision tree to a scored ticket and returns
// the final classification. This is purely deterministic -- no LLM calls.
func Classify(ticket *NormalizedTicket, scores RubricScores, uncertainAxes []string, reasons []string) Classification {
	c := Classification{
		TicketID:      ticket.TicketID,
		Rubric:        scores,
		UncertainAxes: uncertainAxes,
		Reasons:       reasons,
	}

	// Step 1: Check hard stops -- these override everything.
	c.HardStops = checkHardStops(ticket)
	if len(c.HardStops) > 0 {
		c.Category = CategoryHumanOnly
		return c
	}

	// Step 2: Check soft stops (feature flags, deploy coordination).
	softStops := checkSoftStops(ticket)
	if len(softStops) > 0 {
		c.Category = CategoryHumanReviewRequired
		c.HardStops = softStops // Record soft-stop reasons in the audit trail.
		return c
	}

	// Step 3: Check acceptance criteria / design spec signals.
	if !ticket.Signals.AcceptanceCriteriaPresent && !ticket.Signals.HasDesignSpec {
		c.Category = CategoryHumanReviewRequired
		c.HardStops = []string{"no acceptance criteria AND no linked spec"}
		return c
	}

	// Step 4: Evaluate rubric-based gates.
	c.GateResults = evaluateGates(scores)
	for _, gr := range c.GateResults {
		if !gr.Passed {
			c.Category = CategoryHumanReviewRequired
			return c
		}
	}

	// Step 5: Decision tree leaf nodes.
	if scores.Clarity >= 4 &&
		scores.CodeLocality >= 4 &&
		scores.PatternMatch >= 3 &&
		scores.ValidationStrength >= 3 &&
		scores.BlastRadius <= 1 &&
		scores.ProductAmbiguity <= 1 &&
		scores.DependencyRisk <= 1 {
		c.Category = CategoryAIDefinite
		return c
	}

	if scores.Clarity >= 3 &&
		scores.CodeLocality >= 3 &&
		scores.PatternMatch >= 2 &&
		scores.ValidationStrength >= 2 &&
		scores.BlastRadius <= 2 &&
		scores.ProductAmbiguity <= 2 &&
		scores.DependencyRisk <= 2 {
		c.Category = CategoryAILikely
		return c
	}

	// Default: not confident enough for AI.
	c.Category = CategoryHumanReviewRequired
	return c
}

// checkHardStops returns a list of triggered hard-stop reasons by matching
// the ticket's title, description, and labels against keyword groups.
func checkHardStops(ticket *NormalizedTicket) []string {
	text := buildSearchText(ticket)
	var triggered []string

	for group, keywords := range hardStopKeywords {
		if containsAny(text, keywords) {
			triggered = append(triggered, group)
		}
	}

	return triggered
}

// checkSoftStops returns a list of triggered soft-stop reasons.
func checkSoftStops(ticket *NormalizedTicket) []string {
	text := buildSearchText(ticket)
	var triggered []string

	for group, keywords := range softStopKeywords {
		if containsAny(text, keywords) {
			triggered = append(triggered, group)
		}
	}

	return triggered
}

// evaluateGates checks each rubric-based gate and returns the results.
func evaluateGates(scores RubricScores) []GateResult {
	return []GateResult{
		{
			Gate:   "CLARITY_GATE",
			Passed: scores.Clarity >= clarityGate,
			Reason: "clarity < 2 triggers human review",
		},
		{
			Gate:   "BLAST_RADIUS_GATE",
			Passed: scores.BlastRadius <= blastRadiusGate,
			Reason: "blastRadius > 3 triggers human review",
		},
		{
			Gate:   "PRODUCT_AMBIGUITY_GATE",
			Passed: scores.ProductAmbiguity <= productAmbiguityGate,
			Reason: "productAmbiguity > 3 triggers human review",
		},
		{
			Gate:   "DEPENDENCY_GATE",
			Passed: scores.DependencyRisk <= dependencyGate,
			Reason: "dependencyRisk > 3 triggers human review",
		},
	}
}

// buildSearchText concatenates the ticket's title, description, and labels
// into a single lowercase string for keyword matching.
func buildSearchText(ticket *NormalizedTicket) string {
	parts := []string{
		strings.ToLower(ticket.Title),
		strings.ToLower(ticket.Description),
	}
	for _, label := range ticket.Signals.Labels {
		parts = append(parts, strings.ToLower(label))
	}
	return strings.Join(parts, " ")
}

// containsAny returns true if text contains any of the given keywords.
func containsAny(text string, keywords []string) bool {
	for _, kw := range keywords {
		if strings.Contains(text, kw) {
			return true
		}
	}
	return false
}
