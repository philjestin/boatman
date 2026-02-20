// Package handoff provides structured context passing between agents.
// Each handoff contains only the essential information needed for the next agent.
// All handoffs implement the Handoff interface for adaptive sizing.
//
// Phase 3: Interface, utility types, and compound/pipeline handoffs are now
// imported from the harness package. Concrete handoffs remain here.
package handoff

import (
	"fmt"
	"strings"

	harnesshandoff "github.com/philjestin/boatman-ecosystem/harness/handoff"
	"github.com/philjestin/boatmanmode/internal/linear"
	"github.com/philjestin/boatmanmode/internal/task"
)

// Handoff is the interface for passing context between agents.
// Re-exported from harness.
type Handoff = harnesshandoff.Handoff

// TokenBudget represents token limits for different contexts.
// Re-exported from harness.
type TokenBudget = harnesshandoff.TokenBudget

// DefaultBudget provides reasonable defaults for most models.
var DefaultBudget = harnesshandoff.DefaultBudget

// EstimateTokens provides a rough token count for a string.
var EstimateTokens = harnesshandoff.EstimateTokens

// TruncateToTokens truncates a string to fit within a token budget.
var TruncateToTokens = harnesshandoff.TruncateToTokens

// CompoundHandoff combines multiple handoffs into one.
// Re-exported from harness.
type CompoundHandoff = harnesshandoff.CompoundHandoff

// NewCompoundHandoff creates a compound handoff from multiple sources.
var NewCompoundHandoff = harnesshandoff.NewCompoundHandoff

// PipelineHandoff tracks context through a pipeline of agents.
// Re-exported from harness.
type PipelineHandoff = harnesshandoff.PipelineHandoff

// NewPipelineHandoff creates a new pipeline handoff.
var NewPipelineHandoff = harnesshandoff.NewPipelineHandoff

// ExecutionHandoff contains context for the initial code execution.
type ExecutionHandoff struct {
	TicketID    string
	Title       string
	Description string
	Labels      []string
	BranchName  string
}

// NewExecutionHandoff creates a handoff from a Task.
func NewExecutionHandoff(t task.Task) *ExecutionHandoff {
	return &ExecutionHandoff{
		TicketID:    t.GetID(),
		Title:       t.GetTitle(),
		Description: t.GetDescription(),
		Labels:      t.GetLabels(),
		BranchName:  t.GetBranchName(),
	}
}

// NewExecutionHandoffFromTicket creates a handoff from a Linear ticket (backward compatibility).
func NewExecutionHandoffFromTicket(ticket *linear.Ticket) *ExecutionHandoff {
	return NewExecutionHandoff(task.NewLinearTask(ticket))
}

// Full returns the complete execution context.
func (h *ExecutionHandoff) Full() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# %s\n\n", h.Title))
	sb.WriteString(fmt.Sprintf("**Ticket:** %s\n", h.TicketID))
	if len(h.Labels) > 0 {
		sb.WriteString(fmt.Sprintf("**Labels:** %s\n", strings.Join(h.Labels, ", ")))
	}
	sb.WriteString("\n## Requirements\n\n")
	sb.WriteString(h.Description)
	return sb.String()
}

// Concise returns a summary of the execution context.
func (h *ExecutionHandoff) Concise() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# %s (%s)\n\n", h.Title, h.TicketID))
	sb.WriteString(extractRequirements(h.Description))
	return sb.String()
}

// ForTokenBudget returns context sized to fit within token budget.
func (h *ExecutionHandoff) ForTokenBudget(maxTokens int) string {
	full := h.Full()
	if EstimateTokens(full) <= maxTokens {
		return full
	}

	concise := h.Concise()
	if EstimateTokens(concise) <= maxTokens {
		return concise
	}

	return TruncateToTokens(concise, maxTokens)
}

// Type returns the handoff type.
func (h *ExecutionHandoff) Type() string {
	return "execution"
}

// ToPrompt formats the handoff as a prompt for the executor.
// Kept for backward compatibility.
func (h *ExecutionHandoff) ToPrompt() string {
	return h.Full()
}

// ReviewHandoff contains context for ScottBott peer review.
type ReviewHandoff struct {
	TicketID     string
	Title        string
	Requirements string // Concise summary of what was requested
	Diff         string // The actual code changes
	FilesChanged []string
}

// NewReviewHandoff creates a handoff for code review.
func NewReviewHandoff(t task.Task, diff string, filesChanged []string) *ReviewHandoff {
	return &ReviewHandoff{
		TicketID:     t.GetID(),
		Title:        t.GetTitle(),
		Requirements: extractRequirements(t.GetDescription()),
		Diff:         diff,
		FilesChanged: filesChanged,
	}
}

// NewReviewHandoffFromTicket creates a review handoff from a Linear ticket (backward compatibility).
func NewReviewHandoffFromTicket(ticket *linear.Ticket, diff string, filesChanged []string) *ReviewHandoff {
	return NewReviewHandoff(task.NewLinearTask(ticket), diff, filesChanged)
}

// Full returns the complete review context.
func (h *ReviewHandoff) Full() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Review: %s (%s)\n\n", h.Title, h.TicketID))
	sb.WriteString("## Requirements Summary\n\n")
	sb.WriteString(h.Requirements)
	sb.WriteString("\n\n## Files Changed\n\n")
	for _, f := range h.FilesChanged {
		sb.WriteString(fmt.Sprintf("- %s\n", f))
	}
	sb.WriteString("\n## Diff\n\n```diff\n")
	sb.WriteString(h.Diff)
	sb.WriteString("\n```\n")
	return sb.String()
}

// Concise returns a summary of the review context.
func (h *ReviewHandoff) Concise() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Review: %s (%s)\n\n", h.Title, h.TicketID))
	sb.WriteString("## Requirements\n\n")
	sb.WriteString(h.Requirements)
	sb.WriteString(fmt.Sprintf("\n\n## Changes: %d files, %d lines\n",
		len(h.FilesChanged), strings.Count(h.Diff, "\n")))
	for _, f := range h.FilesChanged {
		sb.WriteString(fmt.Sprintf("- %s\n", f))
	}
	return sb.String()
}

// ForTokenBudget returns context sized to fit within token budget.
func (h *ReviewHandoff) ForTokenBudget(maxTokens int) string {
	full := h.Full()
	if EstimateTokens(full) <= maxTokens {
		return full
	}

	// Try with truncated diff
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Review: %s (%s)\n\n", h.Title, h.TicketID))
	sb.WriteString("## Requirements Summary\n\n")
	sb.WriteString(h.Requirements)
	sb.WriteString("\n\n## Files Changed\n\n")
	for _, f := range h.FilesChanged {
		sb.WriteString(fmt.Sprintf("- %s\n", f))
	}

	// Calculate remaining budget for diff
	headerTokens := EstimateTokens(sb.String())
	diffBudget := maxTokens - headerTokens - 100 // Reserve 100 for formatting

	sb.WriteString("\n## Diff (truncated)\n\n```diff\n")
	sb.WriteString(TruncateToTokens(h.Diff, diffBudget))
	sb.WriteString("\n```\n")

	return sb.String()
}

// Type returns the handoff type.
func (h *ReviewHandoff) Type() string {
	return "review"
}

// ToPrompt formats the handoff for the reviewer.
// Kept for backward compatibility.
func (h *ReviewHandoff) ToPrompt() string {
	return h.Full()
}

// RefactorHandoff contains context for a refactor iteration.
type RefactorHandoff struct {
	TicketID      string
	Title         string
	Requirements  string   // Original requirements
	Issues        []string // Specific issues to fix
	Guidance      string   // Review guidance
	FilesToUpdate []string // Files that need changes
	CurrentCode   string   // Current implementation
	ProjectRules  string   // Project coding standards and rules (critical for proper fixes)
}

// NewRefactorHandoff creates a handoff for refactoring.
// projectRules should contain the project's coding standards (from .cursorrules, CLAUDE.md, etc.)
func NewRefactorHandoff(t task.Task, issues []string, guidance string, filesToUpdate []string, currentCode string, projectRules string) *RefactorHandoff {
	return &RefactorHandoff{
		TicketID:      t.GetID(),
		Title:         t.GetTitle(),
		Requirements:  extractRequirements(t.GetDescription()),
		Issues:        issues,
		Guidance:      guidance,
		FilesToUpdate: filesToUpdate,
		CurrentCode:   currentCode,
		ProjectRules:  projectRules,
	}
}

// NewRefactorHandoffFromTicket creates a refactor handoff from a Linear ticket (backward compatibility).
func NewRefactorHandoffFromTicket(ticket *linear.Ticket, issues []string, guidance string, filesToUpdate []string, currentCode string, projectRules string) *RefactorHandoff {
	return NewRefactorHandoff(task.NewLinearTask(ticket), issues, guidance, filesToUpdate, currentCode, projectRules)
}

// Full returns the complete refactor context.
func (h *RefactorHandoff) Full() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Refactor: %s (%s)\n\n", h.Title, h.TicketID))

	// Project rules FIRST - these are critical for proper fixes
	if h.ProjectRules != "" {
		sb.WriteString("## Project Rules & Standards (MUST FOLLOW)\n\n")
		sb.WriteString(h.ProjectRules)
		sb.WriteString("\n\n---\n\n")
	}

	sb.WriteString("## Original Requirements\n\n")
	sb.WriteString(h.Requirements)

	sb.WriteString("\n\n## Issues to Fix (MUST ADDRESS ALL)\n\n")
	sb.WriteString("Review each issue below. The issues reference specific requirements from the ticket and project rules.\n\n")
	for i, issue := range h.Issues {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, issue))
	}

	sb.WriteString("\n## Guidance\n\n")
	sb.WriteString(h.Guidance)

	sb.WriteString("\n\n## Files to Update\n\n")
	for _, f := range h.FilesToUpdate {
		sb.WriteString(fmt.Sprintf("- %s\n", f))
	}

	sb.WriteString("\n## Current Implementation\n\n")
	sb.WriteString(h.CurrentCode)

	sb.WriteString("\n\n## Instructions\n\n")
	sb.WriteString("1. Review the Project Rules & Standards above\n")
	sb.WriteString("2. Review the Original Requirements from the ticket\n")
	sb.WriteString("3. Fix ALL listed issues while following the project rules\n")
	sb.WriteString("4. Output complete updated files using this format:\n\n")
	sb.WriteString("### FILE: path/to/file.ext\n```\n// complete file contents\n```\n")

	return sb.String()
}

// Concise returns a summary of the refactor context.
func (h *RefactorHandoff) Concise() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Refactor: %s (%s)\n\n", h.Title, h.TicketID))

	sb.WriteString("## Issues to Fix\n\n")
	for i, issue := range h.Issues {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, issue))
	}

	sb.WriteString(fmt.Sprintf("\n## Files: %d\n", len(h.FilesToUpdate)))
	for _, f := range h.FilesToUpdate {
		sb.WriteString(fmt.Sprintf("- %s\n", f))
	}

	return sb.String()
}

// ForTokenBudget returns context sized to fit within token budget.
func (h *RefactorHandoff) ForTokenBudget(maxTokens int) string {
	full := h.Full()
	if EstimateTokens(full) <= maxTokens {
		return full
	}

	// Build incrementally, prioritizing rules and issues
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Refactor: %s (%s)\n\n", h.Title, h.TicketID))

	// Project rules are critical - include them first (truncated if needed)
	if h.ProjectRules != "" {
		sb.WriteString("## Project Rules & Standards (MUST FOLLOW)\n\n")
		// Reserve ~2000 tokens for rules - they're critical for correct fixes
		sb.WriteString(TruncateToTokens(h.ProjectRules, 2000))
		sb.WriteString("\n\n---\n\n")
	}

	// Original requirements (truncated if long)
	if h.Requirements != "" {
		sb.WriteString("## Original Requirements\n\n")
		sb.WriteString(TruncateToTokens(h.Requirements, 1000))
		sb.WriteString("\n\n")
	}

	// Issues are most important
	sb.WriteString("## Issues to Fix (MUST ADDRESS ALL)\n\n")
	for i, issue := range h.Issues {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, issue))
	}

	// Guidance is helpful
	if h.Guidance != "" {
		sb.WriteString("\n## Guidance\n\n")
		sb.WriteString(TruncateToTokens(h.Guidance, 500))
	}

	sb.WriteString("\n\n## Files to Update\n\n")
	for _, f := range h.FilesToUpdate {
		sb.WriteString(fmt.Sprintf("- %s\n", f))
	}

	// Calculate remaining budget for code
	headerTokens := EstimateTokens(sb.String())
	codeBudget := maxTokens - headerTokens - 200

	if codeBudget > 500 {
		sb.WriteString("\n## Current Implementation (truncated)\n\n")
		sb.WriteString(TruncateToTokens(h.CurrentCode, codeBudget))
	}

	sb.WriteString("\n\n## Instructions\n\n")
	sb.WriteString("Fix ALL listed issues following project rules. Output complete updated files:\n")
	sb.WriteString("### FILE: path/to/file.ext\n```\n// contents\n```\n")

	return sb.String()
}

// Type returns the handoff type.
func (h *RefactorHandoff) Type() string {
	return "refactor"
}

// ToPrompt formats the handoff for the refactor agent.
// Kept for backward compatibility.
func (h *RefactorHandoff) ToPrompt() string {
	return h.Full()
}

// extractRequirements pulls out the key requirements from a description.
func extractRequirements(description string) string {
	// If description is short, use it as-is
	if len(description) < 500 {
		return description
	}

	// Try to extract just the goal/requirements sections
	lines := strings.Split(description, "\n")
	var result strings.Builder
	inRelevantSection := false

	for _, line := range lines {
		lower := strings.ToLower(line)

		// Start capturing at these sections
		if strings.Contains(lower, "goal") ||
		   strings.Contains(lower, "requirement") ||
		   strings.Contains(lower, "must") ||
		   strings.Contains(lower, "should") {
			inRelevantSection = true
		}

		// Stop at implementation details
		if strings.Contains(lower, "implementation approach") ||
		   strings.Contains(lower, "technical context") ||
		   strings.Contains(lower, "constraints") {
			inRelevantSection = false
		}

		if inRelevantSection || result.Len() < 300 {
			result.WriteString(line)
			result.WriteString("\n")
		}
	}

	summary := strings.TrimSpace(result.String())
	if len(summary) < 100 {
		// Fallback: just truncate
		if len(description) > 800 {
			return description[:800] + "..."
		}
		return description
	}

	return summary
}
