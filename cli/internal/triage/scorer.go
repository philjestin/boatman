package triage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/philjestin/boatmanmode/internal/claude"
	"github.com/philjestin/boatmanmode/internal/config"
	"github.com/philjestin/boatmanmode/internal/cost"
)

const scorerSystemPrompt = `You are a ticket triage evaluator for a software engineering team. You score development tickets on seven rubric dimensions to determine if they are suitable for autonomous AI execution.

Score each dimension from 0 (worst) to 5 (best):

## Positive Dimensions (higher = better for AI)

1. **clarity** (0-5): Are requirements explicit and testable? 5 = clear acceptance criteria with testable behavior. 0 = vague wish with no definition of done.
2. **codeLocality** (0-5): Is the change confined to one module/package? 5 = single file or directory. 0 = spans entire codebase across many services.
3. **patternMatch** (0-5): Has this kind of problem been solved before in this repo? 5 = exact template exists to follow. 0 = completely novel architecture required.
4. **validationStrength** (0-5): Can success be proven with automated tests? 5 = existing test suite covers the area thoroughly. 0 = no automated validation possible.

## Negative Dimensions (higher = worse for AI)

5. **dependencyRisk** (0-5): Does it depend on unknown infra, external APIs, or cross-team sequencing? 5 = multiple external dependencies and coordination required. 0 = entirely self-contained.
6. **productAmbiguity** (0-5): Does it require UX judgment, stakeholder interpretation, or hidden business context? 5 = highly subjective decisions required. 0 = purely mechanical transformation.
7. **blastRadius** (0-5): If the AI implementation is wrong, how bad is the failure? 5 = catastrophic and irreversible (data loss, security breach). 0 = trivially revertible with no user impact.

## Output Format

Respond with ONLY a JSON object. No markdown fencing, no explanation before or after:

{"clarity": <int>, "codeLocality": <int>, "patternMatch": <int>, "validationStrength": <int>, "dependencyRisk": <int>, "productAmbiguity": <int>, "blastRadius": <int>, "uncertainAxes": ["<dimension names where you were least certain>"], "reasons": ["<1 sentence per dimension explaining the score>"]}`

// Scorer uses Claude to score tickets on 7 rubric dimensions.
type Scorer struct {
	claudeClient *claude.Client
	model        string
	cfg          *config.Config

	// OnTicketScored is called after each ticket is scored in ScoreBatch.
	// The int parameters are the zero-based index and total count.
	OnTicketScored func(result ScoredTicket, index, total int)
}

// NewScorer creates a new Scorer that uses Claude for rubric evaluation.
func NewScorer(cfg *config.Config) *Scorer {
	client := claude.New()
	client.Model = cfg.Claude.Models.Scorer
	client.EnablePromptCaching = cfg.Claude.EnablePromptCaching
	client.SkipPermissions = true
	client.EnableTools = false

	return &Scorer{
		claudeClient: client,
		model:        cfg.Claude.Models.Scorer,
		cfg:          cfg,
	}
}

// Score evaluates a single ticket against the 7-dimension rubric using Claude.
func (s *Scorer) Score(ctx context.Context, ticket NormalizedTicket) (*ScorerResponse, *cost.Usage, error) {
	userPrompt := buildUserPrompt(ticket)

	response, usage, err := s.claudeClient.Message(ctx, scorerSystemPrompt, userPrompt)
	if err != nil {
		return nil, nil, fmt.Errorf("scorer Claude call failed for %s: %w", ticket.TicketID, err)
	}

	scored, err := parseScoreResponse(response)
	if err != nil {
		// Log the raw response to stderr so desktop integration captures it
		preview := response
		if len(preview) > 500 {
			preview = preview[:500] + "...(truncated)"
		}
		fmt.Fprintf(os.Stderr, "[triage-scorer] Parse failed for %s. Raw response (%d chars):\n---\n%s\n---\n", ticket.TicketID, len(response), preview)
		return nil, usage, fmt.Errorf("failed to parse scorer response for %s: %w", ticket.TicketID, err)
	}

	return scored, usage, nil
}

// ScoredTicket pairs a ticket with its scoring result.
type ScoredTicket struct {
	Ticket   NormalizedTicket
	Response *ScorerResponse
	Usage    *cost.Usage
	Err      error
}

// ScoreBatch scores multiple tickets concurrently with a semaphore to limit parallelism.
func (s *Scorer) ScoreBatch(ctx context.Context, tickets []NormalizedTicket, concurrency int) []ScoredTicket {
	results := make([]ScoredTicket, len(tickets))
	var wg sync.WaitGroup
	sem := make(chan struct{}, concurrency)

	for i, ticket := range tickets {
		wg.Add(1)
		go func(idx int, t NormalizedTicket) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			resp, usage, err := s.Score(ctx, t)
			results[idx] = ScoredTicket{
				Ticket:   t,
				Response: resp,
				Usage:    usage,
				Err:      err,
			}
			if s.OnTicketScored != nil {
				s.OnTicketScored(results[idx], idx, len(tickets))
			}
		}(i, ticket)
	}

	wg.Wait()
	return results
}

// buildUserPrompt constructs the user prompt from a NormalizedTicket.
func buildUserPrompt(ticket NormalizedTicket) string {
	description := ticket.Description
	if len(description) > 3000 {
		description = description[:3000]
	}

	labels := "none"
	if len(ticket.Signals.Labels) > 0 {
		labels = strings.Join(ticket.Signals.Labels, ", ")
	}

	mentionsFiles := "none detected"
	if len(ticket.Signals.MentionsFiles) > 0 {
		mentionsFiles = strings.Join(ticket.Signals.MentionsFiles, ", ")
	}

	domains := "none detected"
	if len(ticket.Signals.Domains) > 0 {
		domains = strings.Join(ticket.Signals.Domains, ", ")
	}

	dependencies := "none"
	if len(ticket.Signals.Dependencies) > 0 {
		dependencies = strings.Join(ticket.Signals.Dependencies, ", ")
	}

	return fmt.Sprintf(`Score this ticket:

## Ticket: %s
## Title: %s

## Description
%s

## Labels
%s

## Signals
- Mentioned files: %s
- Domains: %s
- Dependencies: %s
- Acceptance criteria present: %t
- Acceptance criteria explicit: %t
- Has design spec: %t`,
		ticket.TicketID,
		ticket.Title,
		description,
		labels,
		mentionsFiles,
		domains,
		dependencies,
		ticket.Signals.AcceptanceCriteriaPresent,
		ticket.Signals.AcceptanceCriteriaExplicit,
		ticket.Signals.HasDesignSpec,
	)
}

// parseScoreResponse extracts and validates a ScorerResponse from Claude's output.
func parseScoreResponse(response string) (*ScorerResponse, error) {
	var jsonStr string

	// Try to find JSON in a markdown fenced code block.
	codeBlockRe := regexp.MustCompile("(?s)`{3}(?:json)?\\s*\\n?(.*?)\\n?`{3}")
	matches := codeBlockRe.FindStringSubmatch(response)
	if len(matches) > 1 {
		jsonStr = matches[1]
	} else {
		// Fallback: find first '{' to last '}'.
		start := strings.Index(response, "{")
		end := strings.LastIndex(response, "}")
		if start >= 0 && end > start {
			jsonStr = response[start : end+1]
		} else {
			return nil, fmt.Errorf("no JSON found in scorer response")
		}
	}

	var scored ScorerResponse
	if err := json.Unmarshal([]byte(jsonStr), &scored); err != nil {
		return nil, fmt.Errorf("invalid JSON in scorer response: %w", err)
	}

	// Clamp all scores to 0-5.
	scored.Clarity = clampScore(scored.Clarity)
	scored.CodeLocality = clampScore(scored.CodeLocality)
	scored.PatternMatch = clampScore(scored.PatternMatch)
	scored.ValidationStrength = clampScore(scored.ValidationStrength)
	scored.DependencyRisk = clampScore(scored.DependencyRisk)
	scored.ProductAmbiguity = clampScore(scored.ProductAmbiguity)
	scored.BlastRadius = clampScore(scored.BlastRadius)

	return &scored, nil
}

// clampScore constrains a score to the 0-5 range.
func clampScore(score int) int {
	if score < 0 {
		return 0
	}
	if score > 5 {
		return 5
	}
	return score
}
