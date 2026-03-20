// Package brain provides domain knowledge loading for Boatman's agent phases.
//
// A "brain" is a markdown file in the brains/ directory that contains
// domain-specific context for the planner, executor, and reviewer.
// Brains are matched to tasks by keywords in the task title, description,
// labels, or file paths.
package brain

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// Brain represents a loaded domain knowledge file.
type Brain struct {
	// Name is the brain identifier (filename without extension).
	Name string

	// Meta is the frontmatter metadata.
	Meta BrainMeta

	// Sections contains phase-specific content keyed by lowercase header.
	// e.g. "planning", "execution", "review", "domain model", "common mistakes"
	Sections map[string]string

	// Raw is the full markdown content (minus frontmatter).
	Raw string
}

// BrainMeta is the YAML frontmatter in a brain file.
type BrainMeta struct {
	// Domain is a human-readable domain name.
	Domain string `yaml:"domain"`

	// Keywords trigger brain matching against task content.
	Keywords []string `yaml:"keywords"`

	// Labels trigger brain matching against task labels.
	Labels []string `yaml:"labels"`

	// Paths trigger brain matching when planned files touch these directories.
	Paths []string `yaml:"paths"`
}

// ForPhase returns the brain content relevant to a specific agent phase.
// Falls back to the full raw content if no phase-specific section exists.
func (b *Brain) ForPhase(phase string) string {
	phase = strings.ToLower(phase)

	// Direct match
	if content, ok := b.Sections[phase]; ok {
		return b.header() + content
	}

	// Phase aliases
	aliases := map[string][]string{
		"planning":  {"planning", "plan", "architecture"},
		"execution": {"execution", "execute", "implementation"},
		"review":    {"review", "reviewing", "conventions"},
		"refactor":  {"review", "execution"},
	}

	if keys, ok := aliases[phase]; ok {
		for _, key := range keys {
			if content, exists := b.Sections[key]; exists {
				return b.header() + content
			}
		}
	}

	// Fallback: return everything
	return b.Raw
}

// header returns a formatted header identifying this brain.
func (b *Brain) header() string {
	domain := b.Meta.Domain
	if domain == "" {
		domain = b.Name
	}
	return fmt.Sprintf("# Domain Brain: %s\n\n", domain)
}

// Loader discovers and loads brain files from a project directory.
type Loader struct {
	projectPath string
}

// NewLoader creates a brain loader rooted at the given project path.
func NewLoader(projectPath string) *Loader {
	return &Loader{projectPath: projectPath}
}

// brainsDir returns the path to the brains directory.
func (l *Loader) brainsDir() string {
	return filepath.Join(l.projectPath, "brains")
}

// LoadAll loads every brain file in the brains/ directory.
func (l *Loader) LoadAll() ([]*Brain, error) {
	dir := l.brainsDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading brains directory: %w", err)
	}

	var brains []*Brain
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		if strings.HasPrefix(entry.Name(), "_") {
			continue // skip _example.md etc.
		}

		b, err := l.loadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			fmt.Printf("   [brain] skipping %s: %v\n", entry.Name(), err)
			continue
		}
		brains = append(brains, b)
	}

	return brains, nil
}

// Match finds brains relevant to the given task context.
func (l *Loader) Match(title, description string, labels, filePaths []string) ([]*Brain, error) {
	all, err := l.LoadAll()
	if err != nil {
		return nil, err
	}

	searchText := strings.ToLower(title + " " + description)
	var matched []*Brain

	for _, b := range all {
		if b.matches(searchText, labels, filePaths) {
			matched = append(matched, b)
		}
	}

	return matched, nil
}

// matches checks if a brain is relevant to the given context.
func (b *Brain) matches(searchText string, labels, filePaths []string) bool {
	searchText = strings.ToLower(searchText)

	// Check keywords against task title + description
	for _, kw := range b.Meta.Keywords {
		if strings.Contains(searchText, strings.ToLower(kw)) {
			return true
		}
	}

	// Check labels
	labelSet := make(map[string]bool, len(labels))
	for _, l := range labels {
		labelSet[strings.ToLower(l)] = true
	}
	for _, l := range b.Meta.Labels {
		if labelSet[strings.ToLower(l)] {
			return true
		}
	}

	// Check file paths
	for _, fp := range filePaths {
		for _, pattern := range b.Meta.Paths {
			if strings.Contains(fp, pattern) {
				return true
			}
		}
	}

	return false
}

// loadFile parses a single brain markdown file.
func (l *Loader) loadFile(path string) (*Brain, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	name := strings.TrimSuffix(filepath.Base(path), ".md")
	meta, body, err := parseFrontmatter(string(content))
	if err != nil {
		return nil, fmt.Errorf("parsing frontmatter: %w", err)
	}

	sections := parseSections(body)

	return &Brain{
		Name:     name,
		Meta:     meta,
		Sections: sections,
		Raw:      body,
	}, nil
}

// parseFrontmatter extracts YAML frontmatter from markdown content.
func parseFrontmatter(content string) (BrainMeta, string, error) {
	var meta BrainMeta

	if !strings.HasPrefix(content, "---\n") {
		return meta, content, nil
	}

	end := strings.Index(content[4:], "\n---")
	if end == -1 {
		return meta, content, nil
	}

	frontmatter := content[4 : 4+end]
	body := strings.TrimSpace(content[4+end+4:])

	if err := yaml.Unmarshal([]byte(frontmatter), &meta); err != nil {
		return meta, body, fmt.Errorf("invalid YAML frontmatter: %w", err)
	}

	return meta, body, nil
}

// parseSections splits markdown into sections by ## headers.
var sectionRe = regexp.MustCompile(`(?m)^##\s+(.+)$`)

func parseSections(body string) map[string]string {
	sections := make(map[string]string)
	matches := sectionRe.FindAllStringSubmatchIndex(body, -1)

	for i, match := range matches {
		header := strings.ToLower(strings.TrimSpace(body[match[2]:match[3]]))
		start := match[1]

		var end int
		if i+1 < len(matches) {
			end = matches[i+1][0]
		} else {
			end = len(body)
		}

		content := strings.TrimSpace(body[start:end])
		sections[header] = content
	}

	return sections
}

// MergeForPhase combines multiple brains' phase-specific content.
func MergeForPhase(brains []*Brain, phase string) string {
	if len(brains) == 0 {
		return ""
	}

	var sb strings.Builder
	for _, b := range brains {
		content := b.ForPhase(phase)
		if content != "" {
			sb.WriteString(content)
			sb.WriteString("\n\n---\n\n")
		}
	}

	return strings.TrimSuffix(sb.String(), "\n\n---\n\n")
}
