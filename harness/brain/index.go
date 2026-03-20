package brain

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

// IndexEntry is a lightweight summary of a brain for the topic index.
type IndexEntry struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Triggers    Triggers `json:"triggers"`
	Confidence  float64  `json:"confidence"`
}

// Index is the topic map loaded into agent context.
type Index struct {
	Entries []IndexEntry
}

// NewIndex creates an index from a slice of entries.
func NewIndex(entries []IndexEntry) *Index {
	return &Index{Entries: entries}
}

// IndexFromBrains builds an index from loaded brains.
func IndexFromBrains(brains []*Brain) *Index {
	entries := make([]IndexEntry, len(brains))
	for i, b := range brains {
		entries[i] = IndexEntry{
			ID:          b.ID,
			Name:        b.Name,
			Description: b.Description,
			Triggers:    b.Triggers,
			Confidence:  b.Confidence,
		}
	}
	return &Index{Entries: entries}
}

// Match returns index entries whose triggers overlap with the context.
// Results are sorted by match score descending.
func (idx *Index) Match(ctx MatchContext) []IndexEntry {
	type scored struct {
		entry IndexEntry
		score float64
	}

	var matches []scored

	for _, entry := range idx.Entries {
		score := matchScore(entry.Triggers, ctx)
		if score > 0 {
			matches = append(matches, scored{entry: entry, score: score})
		}
	}

	sort.Slice(matches, func(i, j int) bool {
		return matches[i].score > matches[j].score
	})

	result := make([]IndexEntry, len(matches))
	for i, m := range matches {
		result[i] = m.entry
	}
	return result
}

// matchScore computes how well triggers overlap with context.
func matchScore(triggers Triggers, ctx MatchContext) float64 {
	score := 0.0

	// Keyword matches
	for _, kw := range triggers.Keywords {
		kwLower := strings.ToLower(kw)
		for _, ctxKw := range ctx.Keywords {
			if strings.ToLower(ctxKw) == kwLower {
				score += 1.0
			}
		}
	}

	// Entity matches (higher weight)
	for _, ent := range triggers.Entities {
		entLower := strings.ToLower(ent)
		for _, ctxEnt := range ctx.Entities {
			if strings.ToLower(ctxEnt) == entLower {
				score += 2.0
			}
		}
	}

	// File pattern matches
	for _, pattern := range triggers.FilePatterns {
		for _, fp := range ctx.FilePaths {
			if matchFilePattern(pattern, fp) {
				score += 1.5
			}
		}
	}

	return score
}

// matchFilePattern checks if a file path matches a trigger pattern.
// Supports both glob patterns and prefix matching (e.g. "packs/subscriptions/").
func matchFilePattern(pattern, filePath string) bool {
	// Prefix match for directory patterns
	if strings.HasSuffix(pattern, "/") {
		return strings.HasPrefix(filePath, pattern)
	}

	// Glob match
	matched, _ := filepath.Match(pattern, filePath)
	if matched {
		return true
	}

	// Also try matching just the base name
	matched, _ = filepath.Match(pattern, filepath.Base(filePath))
	return matched
}

// Full returns the complete index as formatted text.
func (idx *Index) Full() string {
	var sb strings.Builder
	sb.WriteString("# Brain Index\n\n")
	sb.WriteString("Available domain knowledge brains:\n\n")

	for _, e := range idx.Entries {
		sb.WriteString(fmt.Sprintf("## %s\n", e.Name))
		sb.WriteString(fmt.Sprintf("- **ID:** %s\n", e.ID))
		sb.WriteString(fmt.Sprintf("- **Description:** %s\n", e.Description))
		sb.WriteString(fmt.Sprintf("- **Confidence:** %.0f%%\n", e.Confidence*100))

		if len(e.Triggers.Keywords) > 0 {
			sb.WriteString(fmt.Sprintf("- **Keywords:** %s\n", strings.Join(e.Triggers.Keywords, ", ")))
		}
		if len(e.Triggers.Entities) > 0 {
			sb.WriteString(fmt.Sprintf("- **Entities:** %s\n", strings.Join(e.Triggers.Entities, ", ")))
		}
		if len(e.Triggers.FilePatterns) > 0 {
			sb.WriteString(fmt.Sprintf("- **File patterns:** %s\n", strings.Join(e.Triggers.FilePatterns, ", ")))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// Concise returns a compact summary of the index.
func (idx *Index) Concise() string {
	var sb strings.Builder
	sb.WriteString("Brain Index: ")

	names := make([]string, len(idx.Entries))
	for i, e := range idx.Entries {
		names[i] = e.Name
	}
	sb.WriteString(strings.Join(names, ", "))

	return sb.String()
}

// ForTokenBudget returns the index sized for a token budget.
func (idx *Index) ForTokenBudget(maxTokens int) string {
	full := idx.Full()
	if len(full)/4 <= maxTokens {
		return full
	}
	return idx.Concise()
}

// Type returns the handoff type identifier.
func (idx *Index) Type() string {
	return "brain-index"
}
