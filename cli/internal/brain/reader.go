package brain

import (
	"fmt"
	"os"

	harnessbrain "github.com/philjestin/boatman-ecosystem/harness/brain"
	"gopkg.in/yaml.v3"
)

// YAMLReader reads brains from YAML files.
type YAMLReader struct{}

// yamlBrain is the YAML-specific representation for unmarshaling.
type yamlBrain struct {
	ID          string          `yaml:"id"`
	Name        string          `yaml:"name"`
	Version     int             `yaml:"version"`
	Description string          `yaml:"description"`
	Triggers    yamlTriggers    `yaml:"triggers"`
	Confidence  float64         `yaml:"confidence"`
	LastUpdated string          `yaml:"last_updated"`
	Sections    []yamlSection   `yaml:"sections"`
	References  []yamlReference `yaml:"references"`
}

type yamlTriggers struct {
	Keywords     []string `yaml:"keywords"`
	Entities     []string `yaml:"entities"`
	FilePatterns []string `yaml:"file_patterns"`
}

type yamlSection struct {
	Title   string `yaml:"title"`
	Content string `yaml:"content"`
}

type yamlReference struct {
	Path        string `yaml:"path"`
	Description string `yaml:"description"`
	Checksum    string `yaml:"checksum"`
}

// Read parses a YAML brain file.
func (r *YAMLReader) Read(path string) (*harnessbrain.Brain, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", path, err)
	}

	var yb yamlBrain
	if err := yaml.Unmarshal(data, &yb); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", path, err)
	}

	if yb.ID == "" {
		return nil, fmt.Errorf("brain in %s has no ID", path)
	}

	// Convert to harness types
	sections := make([]harnessbrain.Section, len(yb.Sections))
	for i, s := range yb.Sections {
		sections[i] = harnessbrain.Section{Title: s.Title, Content: s.Content}
	}

	refs := make([]harnessbrain.Reference, len(yb.References))
	for i, r := range yb.References {
		refs[i] = harnessbrain.Reference{Path: r.Path, Description: r.Description, Checksum: r.Checksum}
	}

	return &harnessbrain.Brain{
		ID:          yb.ID,
		Name:        yb.Name,
		Version:     yb.Version,
		Description: yb.Description,
		Triggers: harnessbrain.Triggers{
			Keywords:     yb.Triggers.Keywords,
			Entities:     yb.Triggers.Entities,
			FilePatterns: yb.Triggers.FilePatterns,
		},
		Confidence:  yb.Confidence,
		LastUpdated: yb.LastUpdated,
		Sections:    sections,
		References:  refs,
	}, nil
}

// Extensions returns the file extensions this reader handles.
func (r *YAMLReader) Extensions() []string {
	return []string{".yaml", ".yml"}
}

// CompositeReader combines multiple readers to support all formats.
type CompositeReader struct {
	readers []harnessbrain.BrainReader
}

// NewCompositeReader creates a reader supporting both JSON and YAML.
func NewCompositeReader() *CompositeReader {
	return &CompositeReader{
		readers: []harnessbrain.BrainReader{
			&YAMLReader{},
			&harnessbrain.JSONReader{},
		},
	}
}

// Read tries each reader based on file extension.
func (r *CompositeReader) Read(path string) (*harnessbrain.Brain, error) {
	for _, reader := range r.readers {
		for _, ext := range reader.Extensions() {
			if len(path) > len(ext) && path[len(path)-len(ext):] == ext {
				return reader.Read(path)
			}
		}
	}
	return nil, fmt.Errorf("unsupported file format: %s", path)
}

// Extensions returns all supported extensions.
func (r *CompositeReader) Extensions() []string {
	var exts []string
	for _, reader := range r.readers {
		exts = append(exts, reader.Extensions()...)
	}
	return exts
}
