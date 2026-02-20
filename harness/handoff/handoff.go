// Package handoff provides structured context passing between pipeline stages.
// It defines the Handoff interface and utility types for token-efficient
// context management in AI agent pipelines.
package handoff

import (
	"fmt"
	"strings"
)

// Handoff is the interface for passing context between agents.
// It supports multiple sizing strategies for token efficiency.
type Handoff interface {
	// Full returns the complete context
	Full() string
	// Concise returns a summary suitable for quick handoffs
	Concise() string
	// ForTokenBudget returns context sized to fit within token budget
	ForTokenBudget(maxTokens int) string
	// Type returns the handoff type for routing
	Type() string
}

// TokenBudget represents token limits for different contexts.
type TokenBudget struct {
	System  int // Max tokens for system prompt
	User    int // Max tokens for user prompt
	Context int // Max tokens for additional context
	Total   int // Total token budget
}

// DefaultBudget provides reasonable defaults for most models.
var DefaultBudget = TokenBudget{
	System:  8000,
	User:    50000,
	Context: 30000,
	Total:   100000,
}

// EstimateTokens provides a rough token count for a string.
// Assumes ~4 chars per token for English text.
func EstimateTokens(s string) int {
	return len(s) / 4
}

// TruncateToTokens truncates a string to fit within a token budget.
func TruncateToTokens(s string, maxTokens int) string {
	maxChars := maxTokens * 4
	if len(s) <= maxChars {
		return s
	}
	return s[:maxChars] + "\n... (truncated)"
}

// CompoundHandoff combines multiple handoffs into one.
type CompoundHandoff struct {
	Handoffs []Handoff
}

// NewCompoundHandoff creates a compound handoff from multiple sources.
func NewCompoundHandoff(handoffs ...Handoff) *CompoundHandoff {
	return &CompoundHandoff{Handoffs: handoffs}
}

// Full returns all handoffs combined.
func (h *CompoundHandoff) Full() string {
	var parts []string
	for _, ho := range h.Handoffs {
		parts = append(parts, ho.Full())
	}
	return strings.Join(parts, "\n\n---\n\n")
}

// Concise returns concise versions of all handoffs.
func (h *CompoundHandoff) Concise() string {
	var parts []string
	for _, ho := range h.Handoffs {
		parts = append(parts, ho.Concise())
	}
	return strings.Join(parts, "\n\n")
}

// ForTokenBudget distributes budget across handoffs.
func (h *CompoundHandoff) ForTokenBudget(maxTokens int) string {
	if len(h.Handoffs) == 0 {
		return ""
	}

	// First pass: try concise for all
	conciseTotal := EstimateTokens(h.Concise())
	if conciseTotal <= maxTokens {
		// We have room for more detail
		budgetPerHandoff := maxTokens / len(h.Handoffs)
		var parts []string
		for _, ho := range h.Handoffs {
			parts = append(parts, ho.ForTokenBudget(budgetPerHandoff))
		}
		return strings.Join(parts, "\n\n---\n\n")
	}

	// Tight on tokens, use pure concise
	return h.Concise()
}

// Type returns the handoff type.
func (h *CompoundHandoff) Type() string {
	return "compound"
}

// PipelineHandoff tracks context through a pipeline of agents.
type PipelineHandoff struct {
	// Original is the original handoff that started the pipeline
	Original Handoff
	// History is the sequence of handoffs from each agent
	History []Handoff
	// Current is the current handoff
	Current Handoff
}

// NewPipelineHandoff creates a new pipeline handoff.
func NewPipelineHandoff(original Handoff) *PipelineHandoff {
	return &PipelineHandoff{
		Original: original,
		History:  []Handoff{},
		Current:  original,
	}
}

// Advance moves to the next stage with a new handoff.
func (h *PipelineHandoff) Advance(next Handoff) {
	h.History = append(h.History, h.Current)
	h.Current = next
}

// Full returns the current handoff's full content.
func (h *PipelineHandoff) Full() string {
	return h.Current.Full()
}

// Concise returns the current handoff's concise content.
func (h *PipelineHandoff) Concise() string {
	return h.Current.Concise()
}

// ForTokenBudget returns current handoff sized for budget.
func (h *PipelineHandoff) ForTokenBudget(maxTokens int) string {
	return h.Current.ForTokenBudget(maxTokens)
}

// Type returns the current handoff type.
func (h *PipelineHandoff) Type() string {
	return h.Current.Type()
}

// WithHistory returns context with history for debugging.
func (h *PipelineHandoff) WithHistory(maxHistoryItems int) string {
	var sb strings.Builder

	sb.WriteString("# Pipeline Context\n\n")
	sb.WriteString("## Original\n")
	sb.WriteString(h.Original.Concise())
	sb.WriteString("\n\n")

	// Include recent history
	start := 0
	if len(h.History) > maxHistoryItems {
		start = len(h.History) - maxHistoryItems
	}

	if len(h.History) > 0 {
		sb.WriteString("## History (recent)\n")
		for i := start; i < len(h.History); i++ {
			sb.WriteString(fmt.Sprintf("\n### Step %d: %s\n", i+1, h.History[i].Type()))
			sb.WriteString(h.History[i].Concise())
			sb.WriteString("\n")
		}
	}

	sb.WriteString("\n## Current\n")
	sb.WriteString(h.Current.Full())

	return sb.String()
}
