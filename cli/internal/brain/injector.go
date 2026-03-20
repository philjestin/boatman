package brain

import (
	"fmt"
	"strings"

	harnessbrain "github.com/philjestin/boatman-ecosystem/harness/brain"
)

// Injector selects and loads brains for a given task context.
type Injector struct {
	loader    *harnessbrain.Loader
	maxBrains int
}

// NewInjector creates an injector that scans project and global brain directories.
func NewInjector(projectPath string, maxBrains int) *Injector {
	dirs := harnessbrain.DefaultDirs(projectPath)
	reader := NewCompositeReader()
	loader := harnessbrain.NewLoader(dirs, reader)

	if maxBrains <= 0 {
		maxBrains = 3
	}

	return &Injector{
		loader:    loader,
		maxBrains: maxBrains,
	}
}

// ForContext matches brains against keywords, file paths, and entities
// extracted from a task description and plan.
func (inj *Injector) ForContext(keywords, filePaths, entities []string) (*harnessbrain.BrainHandoff, error) {
	idx, err := inj.loader.LoadIndex()
	if err != nil {
		return nil, fmt.Errorf("failed to load brain index: %w", err)
	}

	if len(idx.Entries) == 0 {
		return nil, nil
	}

	ctx := harnessbrain.MatchContext{
		Keywords:  keywords,
		FilePaths: filePaths,
		Entities:  entities,
	}

	matches := idx.Match(ctx)
	if len(matches) == 0 {
		return nil, nil
	}

	// Limit to maxBrains
	if len(matches) > inj.maxBrains {
		matches = matches[:inj.maxBrains]
	}

	// Load the full brains
	ids := make([]string, len(matches))
	for i, m := range matches {
		ids[i] = m.ID
	}

	brains, err := inj.loader.LoadBrains(ids)
	if err != nil {
		return nil, fmt.Errorf("failed to load brains: %w", err)
	}

	return harnessbrain.NewBrainHandoff(brains), nil
}

// ExtractKeywords extracts potential matching keywords from a text description.
func ExtractKeywords(text string) []string {
	// Simple word extraction — split on whitespace and filter short words
	words := strings.Fields(strings.ToLower(text))
	seen := make(map[string]bool)
	var keywords []string

	for _, w := range words {
		// Strip common punctuation
		w = strings.Trim(w, ".,;:!?()[]{}\"'`")
		if len(w) < 3 || seen[w] {
			continue
		}
		seen[w] = true
		keywords = append(keywords, w)
	}

	return keywords
}
