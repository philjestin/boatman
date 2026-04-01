package plan

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
	"github.com/philjestin/boatmanmode/internal/events"
	"github.com/philjestin/boatmanmode/internal/triage"
)

// Generator uses Claude with tools to generate execution plans for tickets.
type Generator struct {
	client   *claude.Client
	cfg      *config.Config
	repoPath string

	// OnPlanGenerated is called after each ticket is planned in GenerateBatch.
	OnPlanGenerated func(result PlanResult, index, total int)
}

// NewGenerator creates a Generator that uses Claude with Read/Grep/Glob tools
// to explore the repo and produce execution plans.
func NewGenerator(cfg *config.Config, repoPath string) *Generator {
	client := claude.NewWithTools(repoPath, "triage-planner", []string{"Read", "Grep", "Glob"})

	if cfg.Claude.Models.Planner != "" {
		client.Model = cfg.Claude.Models.Planner
	}
	client.Effort = cfg.Claude.Effort
	client.EnablePromptCaching = cfg.Claude.EnablePromptCaching
	client.SkipPermissions = true

	// Set BOATMAN_NO_TMUX so desktop subprocess mode uses streaming instead of tmux
	client.Env["BOATMAN_NO_TMUX"] = "1"

	// Forward Claude stream events for desktop visibility
	client.EventForwarder = func(rawLine string) {
		events.ClaudeStream("triage-planner", rawLine)
	}

	return &Generator{
		client:   client,
		cfg:      cfg,
		repoPath: repoPath,
	}
}

const plannerSystemPrompt = `You are a senior software architect generating a concrete execution plan for an AI coding agent.

You have access to Read, Grep, and Glob tools to explore the codebase. Use them — do NOT guess file paths.

## Your Process

1. Use Glob to find files in the specified repo areas
2. Use Read to examine key files and understand existing patterns
3. Use Grep to find related code, imports, and test patterns
4. Generate a concrete, actionable execution plan

## Output Requirements

After exploring the codebase, output ONLY a JSON object in this exact format. No text before or after the JSON:

` + "```json" + `
{
  "approach": "Detailed description of the implementation approach — what to change, how, and why",
  "candidateFiles": [
    "path/relative/to/repo/root/file1.tsx",
    "path/relative/to/repo/root/file2.rb"
  ],
  "newFiles": [],
  "deletedFiles": [],
  "validation": [
    "yarn test path/to/tests",
    "yarn check-types",
    "bundle exec rspec path/to/spec"
  ],
  "rollback": "How to undo this change (e.g. 'Revert PR — no migration, no data changes')",
  "stopConditions": [
    "If [specific condition], stop because [specific reason]",
    "If the change requires modifying files not listed in candidateFiles, stop and re-plan"
  ],
  "uncertainties": [
    "Specific things that are unclear from the ticket or codebase"
  ]
}
` + "```" + `

## Rules

- candidateFiles must be real paths you verified exist with Glob or Read. Never guess.
- validation must use real test runners (yarn test, bundle exec rspec, yarn check-types, etc.)
- stopConditions must be non-empty — always include at least one meaningful stop condition.
- uncertainties should list things that could block or change the approach.
- approach should be specific enough that a competent engineer (or AI agent) could follow it.`

// GeneratePlan produces an execution plan for a single ticket by calling Claude
// with tool access to explore the repo.
func (g *Generator) GeneratePlan(
	ctx context.Context,
	ticket triage.NormalizedTicket,
	classification triage.Classification,
	contextDoc *triage.ContextDoc,
) (*TicketPlan, *cost.Usage, error) {
	userPrompt := buildPlannerPrompt(ticket, classification, contextDoc)

	response, usage, err := g.client.Message(ctx, plannerSystemPrompt, userPrompt)
	if err != nil {
		return nil, nil, fmt.Errorf("planner Claude call failed for %s: %w", ticket.TicketID, err)
	}

	plan, err := parsePlanResponse(response)
	if err != nil {
		preview := response
		if len(preview) > 500 {
			preview = preview[:500] + "...(truncated)"
		}
		fmt.Fprintf(os.Stderr, "[triage-planner] Parse failed for %s. Raw response (%d chars):\n---\n%s\n---\n",
			ticket.TicketID, len(response), preview)
		return nil, usage, fmt.Errorf("failed to parse planner response for %s: %w", ticket.TicketID, err)
	}

	plan.TicketID = ticket.TicketID
	return plan, usage, nil
}

// GenerateBatch generates plans for multiple tickets concurrently.
func (g *Generator) GenerateBatch(
	ctx context.Context,
	tickets []triage.NormalizedTicket,
	classifications []triage.Classification,
	contextDocs []triage.ContextDoc,
	concurrency int,
) []PlanResult {
	if concurrency <= 0 {
		concurrency = 1
	}

	EmitPlanStarted(len(tickets))

	// Build lookups
	classMap := make(map[string]triage.Classification, len(classifications))
	for _, c := range classifications {
		classMap[c.TicketID] = c
	}

	docMap := make(map[string]*triage.ContextDoc)
	for i := range contextDocs {
		for _, tid := range contextDocs[i].TicketIDs {
			docMap[tid] = &contextDocs[i]
		}
	}

	results := make([]PlanResult, len(tickets))
	var wg sync.WaitGroup
	sem := make(chan struct{}, concurrency)

	for i, ticket := range tickets {
		wg.Add(1)
		go func(idx int, t triage.NormalizedTicket) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			classification := classMap[t.TicketID]
			doc := docMap[t.TicketID]

			EmitTicketPlanning(t.TicketID, idx, len(tickets))

			plan, usage, err := g.GeneratePlan(ctx, t, classification, doc)

			results[idx] = PlanResult{
				TicketID: t.TicketID,
				Plan:     plan,
				Usage:    usage,
			}
			if err != nil {
				results[idx].Error = err.Error()
			}

			EmitTicketPlanned(results[idx], idx, len(tickets))

			if g.OnPlanGenerated != nil {
				g.OnPlanGenerated(results[idx], idx, len(tickets))
			}
		}(i, ticket)
	}

	wg.Wait()
	return results
}

func buildPlannerPrompt(
	ticket triage.NormalizedTicket,
	classification triage.Classification,
	contextDoc *triage.ContextDoc,
) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# Ticket: %s\n\n", ticket.TicketID))
	sb.WriteString(fmt.Sprintf("## Title\n%s\n\n", ticket.Title))

	if ticket.Description != "" {
		desc := ticket.Description
		if len(desc) > 3000 {
			desc = desc[:3000] + "\n...(truncated)"
		}
		sb.WriteString(fmt.Sprintf("## Description\n%s\n\n", desc))
	}

	// Classification info
	sb.WriteString(fmt.Sprintf("## Classification: %s\n", classification.Category))
	sb.WriteString(fmt.Sprintf("- Clarity: %d/5, Code Locality: %d/5, Pattern Match: %d/5\n",
		classification.Rubric.Clarity, classification.Rubric.CodeLocality, classification.Rubric.PatternMatch))
	sb.WriteString(fmt.Sprintf("- Dependency Risk: %d/5, Product Ambiguity: %d/5, Blast Radius: %d/5\n\n",
		classification.Rubric.DependencyRisk, classification.Rubric.ProductAmbiguity, classification.Rubric.BlastRadius))

	if len(classification.Reasons) > 0 {
		sb.WriteString("## Scoring Reasons\n")
		for _, r := range classification.Reasons {
			sb.WriteString(fmt.Sprintf("- %s\n", r))
		}
		sb.WriteString("\n")
	}

	// Signals
	if len(ticket.Signals.MentionsFiles) > 0 {
		sb.WriteString(fmt.Sprintf("## Mentioned Files\n%s\n\n", strings.Join(ticket.Signals.MentionsFiles, ", ")))
	}
	if len(ticket.Signals.Domains) > 0 {
		sb.WriteString(fmt.Sprintf("## Domains\n%s\n\n", strings.Join(ticket.Signals.Domains, ", ")))
	}

	// Cluster context
	if contextDoc != nil {
		sb.WriteString("## Cluster Context\n\n")
		sb.WriteString(fmt.Sprintf("**Cluster:** %s\n", contextDoc.ClusterID))
		sb.WriteString(fmt.Sprintf("**Rationale:** %s\n\n", contextDoc.Rationale))

		if len(contextDoc.RepoAreas) > 0 {
			sb.WriteString("**Repo Areas to explore:**\n")
			for _, area := range contextDoc.RepoAreas {
				sb.WriteString(fmt.Sprintf("- %s\n", area))
			}
			sb.WriteString("\n")
		}

		if len(contextDoc.KnownPatterns) > 0 {
			sb.WriteString("**Known Patterns (follow these):**\n")
			for _, p := range contextDoc.KnownPatterns {
				sb.WriteString(fmt.Sprintf("- %s\n", p))
			}
			sb.WriteString("\n")
		}

		if len(contextDoc.ValidationPlan) > 0 {
			sb.WriteString("**Validation Commands:**\n")
			for _, v := range contextDoc.ValidationPlan {
				sb.WriteString(fmt.Sprintf("- %s\n", v))
			}
			sb.WriteString("\n")
		}

		if len(contextDoc.Risks) > 0 {
			sb.WriteString("**Risks:**\n")
			for _, r := range contextDoc.Risks {
				sb.WriteString(fmt.Sprintf("- %s\n", r))
			}
			sb.WriteString("\n")
		}

		sb.WriteString(fmt.Sprintf("**Cost Ceiling:** %dK tokens, %d min per ticket\n\n",
			contextDoc.CostCeiling.MaxTokensPerTicket/1000,
			contextDoc.CostCeiling.MaxAgentMinutesPerTicket))
	}

	sb.WriteString("Explore the codebase using the repo areas above, then generate the execution plan JSON.")
	return sb.String()
}

// parsePlanResponse extracts a TicketPlan JSON from Claude's response.
func parsePlanResponse(response string) (*TicketPlan, error) {
	jsonRe := regexp.MustCompile("(?s)```(?:json)?\\s*\\n?(.*?)\\n?```")
	matches := jsonRe.FindStringSubmatch(response)

	var jsonStr string
	if len(matches) > 1 {
		jsonStr = matches[1]
	} else {
		start := strings.Index(response, "{")
		end := strings.LastIndex(response, "}")
		if start >= 0 && end > start {
			jsonStr = response[start : end+1]
		} else {
			return nil, fmt.Errorf("no JSON found in planner response")
		}
	}

	var plan TicketPlan
	if err := json.Unmarshal([]byte(jsonStr), &plan); err != nil {
		return nil, fmt.Errorf("invalid plan JSON: %w", err)
	}

	return &plan, nil
}
