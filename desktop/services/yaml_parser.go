package services

import (
	"fmt"

	harnessbrain "github.com/philjestin/boatman-ecosystem/harness/brain"
	"gopkg.in/yaml.v3"
)

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

// parseYAML parses YAML data into a Brain struct.
func parseYAML(data []byte, path string) (*harnessbrain.Brain, error) {
	var yb yamlBrain
	if err := yaml.Unmarshal(data, &yb); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", path, err)
	}

	if yb.ID == "" {
		return nil, fmt.Errorf("brain in %s has no ID", path)
	}

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
