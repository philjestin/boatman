package brain

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// BrainReader reads brain data from a file and returns a Brain.
// The harness provides a JSON implementation; CLI provides YAML.
type BrainReader interface {
	Read(path string) (*Brain, error)
	Extensions() []string
}

// Loader scans directories for brain files and loads them.
type Loader struct {
	dirs    []string // priority-ordered: first match wins
	reader  BrainReader
}

// NewLoader creates a loader with the given directories and reader.
// Directories are searched in order; the first brain matching an ID wins.
func NewLoader(dirs []string, reader BrainReader) *Loader {
	return &Loader{
		dirs:   dirs,
		reader: reader,
	}
}

// DefaultDirs returns the standard brain directories for a project.
// Project-level brains take priority over global brains.
func DefaultDirs(projectPath string) []string {
	dirs := []string{
		filepath.Join(projectPath, ".boatman", "brains"),
	}

	if homeDir, err := os.UserHomeDir(); err == nil {
		dirs = append(dirs, filepath.Join(homeDir, ".boatman", "brains"))
	}

	return dirs
}

// LoadIndex scans all directories and builds an index from brain metadata.
func (l *Loader) LoadIndex() (*Index, error) {
	seen := make(map[string]bool)
	var entries []IndexEntry

	for _, dir := range l.dirs {
		dirEntries, err := l.scanDir(dir)
		if err != nil {
			continue // skip dirs that don't exist or can't be read
		}

		for _, e := range dirEntries {
			if !seen[e.ID] {
				seen[e.ID] = true
				entries = append(entries, e)
			}
		}
	}

	return NewIndex(entries), nil
}

// LoadBrain loads a specific brain by ID from the first matching directory.
func (l *Loader) LoadBrain(id string) (*Brain, error) {
	for _, dir := range l.dirs {
		brain, err := l.findInDir(dir, id)
		if err == nil {
			return brain, nil
		}
	}
	return nil, fmt.Errorf("brain %q not found", id)
}

// LoadBrains loads multiple brains by ID.
func (l *Loader) LoadBrains(ids []string) ([]*Brain, error) {
	var brains []*Brain
	for _, id := range ids {
		b, err := l.LoadBrain(id)
		if err != nil {
			continue
		}
		brains = append(brains, b)
	}
	if len(brains) == 0 && len(ids) > 0 {
		return nil, fmt.Errorf("no brains found for IDs: %s", strings.Join(ids, ", "))
	}
	return brains, nil
}

// scanDir reads all brain files in a directory and returns index entries.
func (l *Loader) scanDir(dir string) ([]IndexEntry, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var result []IndexEntry

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if !l.hasValidExtension(entry.Name()) {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		brain, err := l.reader.Read(path)
		if err != nil {
			continue
		}

		result = append(result, IndexEntry{
			ID:          brain.ID,
			Name:        brain.Name,
			Description: brain.Description,
			Triggers:    brain.Triggers,
			Confidence:  brain.Confidence,
		})
	}

	return result, nil
}

// findInDir searches a directory for a brain with the given ID.
func (l *Loader) findInDir(dir string, id string) (*Brain, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() || !l.hasValidExtension(entry.Name()) {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		brain, err := l.reader.Read(path)
		if err != nil {
			continue
		}

		if brain.ID == id {
			return brain, nil
		}
	}

	return nil, fmt.Errorf("brain %q not found in %s", id, dir)
}

// hasValidExtension checks if the filename has a reader-supported extension.
func (l *Loader) hasValidExtension(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	for _, valid := range l.reader.Extensions() {
		if ext == valid {
			return true
		}
	}
	return false
}

// JSONReader reads brains from JSON files (stdlib-only for harness).
type JSONReader struct{}

// Read parses a JSON brain file.
func (r *JSONReader) Read(path string) (*Brain, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", path, err)
	}

	var b Brain
	if err := json.Unmarshal(data, &b); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", path, err)
	}

	if b.ID == "" {
		return nil, fmt.Errorf("brain in %s has no ID", path)
	}

	return &b, nil
}

// Extensions returns the file extensions this reader handles.
func (r *JSONReader) Extensions() []string {
	return []string{".json"}
}
