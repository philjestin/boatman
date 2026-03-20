package brain

import (
	"fmt"
	"sort"
	"strings"
)

// BrainHandoff wraps loaded brains as a Handoff for context injection.
type BrainHandoff struct {
	Brains []*Brain
}

// NewBrainHandoff creates a handoff from loaded brains.
func NewBrainHandoff(brains []*Brain) *BrainHandoff {
	return &BrainHandoff{Brains: brains}
}

// Full returns the complete content of all brains.
func (h *BrainHandoff) Full() string {
	var sb strings.Builder
	sb.WriteString("# Domain Knowledge\n\n")

	for _, b := range h.Brains {
		sb.WriteString(fmt.Sprintf("## %s\n", b.Name))
		sb.WriteString(fmt.Sprintf("*%s*\n\n", b.Description))

		for _, s := range b.Sections {
			sb.WriteString(fmt.Sprintf("### %s\n", s.Title))
			sb.WriteString(s.Content)
			sb.WriteString("\n\n")
		}

		if len(b.References) > 0 {
			sb.WriteString("### Key Files\n")
			for _, ref := range b.References {
				sb.WriteString(fmt.Sprintf("- `%s` — %s\n", ref.Path, ref.Description))
			}
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// Concise returns a brief summary of the brains.
func (h *BrainHandoff) Concise() string {
	var sb strings.Builder
	sb.WriteString("Domain Knowledge: ")

	names := make([]string, len(h.Brains))
	for i, b := range h.Brains {
		names[i] = b.Name
	}
	sb.WriteString(strings.Join(names, ", "))

	return sb.String()
}

// ForTokenBudget returns brain content sized for a token budget.
// Progressive truncation: drop low-confidence brains first, then truncate sections.
func (h *BrainHandoff) ForTokenBudget(maxTokens int) string {
	if len(h.Brains) == 0 {
		return ""
	}

	full := h.Full()
	if len(full)/4 <= maxTokens {
		return full
	}

	// Sort by confidence descending (work on a copy)
	sorted := make([]*Brain, len(h.Brains))
	copy(sorted, h.Brains)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Confidence > sorted[j].Confidence
	})

	// Progressively drop lowest-confidence brains until we fit
	for count := len(sorted); count > 0; count-- {
		subset := &BrainHandoff{Brains: sorted[:count]}
		content := subset.Full()
		if len(content)/4 <= maxTokens {
			return content
		}
	}

	// Still too large — truncate the highest-confidence brain's sections
	top := sorted[0]
	var sb strings.Builder
	sb.WriteString("# Domain Knowledge\n\n")
	sb.WriteString(fmt.Sprintf("## %s\n", top.Name))
	sb.WriteString(fmt.Sprintf("*%s*\n\n", top.Description))

	maxChars := maxTokens * 4
	for _, s := range top.Sections {
		entry := fmt.Sprintf("### %s\n%s\n\n", s.Title, s.Content)
		if sb.Len()+len(entry) > maxChars {
			remaining := maxChars - sb.Len()
			if remaining > 50 {
				sb.WriteString(entry[:remaining])
				sb.WriteString("\n... (truncated)")
			}
			break
		}
		sb.WriteString(entry)
	}

	return sb.String()
}

// Type returns the handoff type identifier.
func (h *BrainHandoff) Type() string {
	return "brain"
}
