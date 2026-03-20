package brain

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	harnessbrain "github.com/philjestin/boatman-ecosystem/harness/brain"
	"gopkg.in/yaml.v3"
)

// Distiller converts accumulated signals into draft brain YAML files.
type Distiller struct {
	store     *harnessbrain.SignalStore
	outputDir string
}

// NewDistiller creates a distiller reading from the default signal store.
func NewDistiller(outputDir string) (*Distiller, error) {
	store, err := harnessbrain.NewSignalStore("")
	if err != nil {
		return nil, err
	}

	return &Distiller{
		store:     store,
		outputDir: outputDir,
	}, nil
}

// Distill clusters signals by domain and generates draft brain YAML files.
// Returns paths to generated files.
func (d *Distiller) Distill() ([]string, error) {
	signals := d.store.GetAll()
	if len(signals) == 0 {
		return nil, nil
	}

	// Cluster signals by domain
	clusters := make(map[string][]harnessbrain.Signal)
	for _, sig := range signals {
		clusters[sig.Domain] = append(clusters[sig.Domain], sig)
	}

	if err := os.MkdirAll(d.outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	var paths []string

	for domain, domainSignals := range clusters {
		if len(domainSignals) < 2 {
			continue // Need at least 2 signals to suggest a brain
		}

		path, err := d.generateDraft(domain, domainSignals)
		if err != nil {
			fmt.Printf("Warning: failed to generate draft for %s: %v\n", domain, err)
			continue
		}
		paths = append(paths, path)
	}

	sort.Strings(paths)
	return paths, nil
}

// generateDraft creates a draft brain YAML from signals in a domain.
func (d *Distiller) generateDraft(domain string, signals []harnessbrain.Signal) (string, error) {
	// Collect unique file paths and details
	var filePaths []string
	var details []string
	seenFiles := make(map[string]bool)

	for _, sig := range signals {
		for _, fp := range sig.FilePaths {
			if !seenFiles[fp] {
				seenFiles[fp] = true
				filePaths = append(filePaths, fp)
			}
		}
		if sig.Details != "" {
			details = append(details, fmt.Sprintf("- [%s] %s (count: %d)", sig.Type, sig.Details, sig.Count))
		}
	}

	// Infer keywords from domain and file paths
	keywords := []string{domain}
	for _, fp := range filePaths {
		base := strings.TrimSuffix(filepath.Base(fp), filepath.Ext(fp))
		if base != domain && len(base) > 2 {
			keywords = append(keywords, strings.ToLower(base))
		}
	}

	// Infer file patterns
	var filePatterns []string
	patternSeen := make(map[string]bool)
	for _, fp := range filePaths {
		parts := strings.SplitN(fp, "/", 3)
		if len(parts) >= 2 {
			pattern := parts[0] + "/" + parts[1] + "/"
			if !patternSeen[pattern] {
				patternSeen[pattern] = true
				filePatterns = append(filePatterns, pattern)
			}
		}
	}

	draft := yamlBrain{
		ID:          "draft-" + sanitize(domain),
		Name:        titleCase(domain) + " Domain",
		Version:     1,
		Description: fmt.Sprintf("Auto-generated draft brain for the %s domain. Review and curate before use.", domain),
		Triggers: yamlTriggers{
			Keywords:     keywords,
			FilePatterns: filePatterns,
		},
		Confidence: 0.5,
		Sections: []yamlSection{
			{
				Title:   "Signals Detected",
				Content: strings.Join(details, "\n"),
			},
			{
				Title:   "Domain Model",
				Content: "TODO: Document the key models, relationships, and invariants for this domain.",
			},
			{
				Title:   "Business Rules",
				Content: "TODO: Document the business rules and constraints that agents should know.",
			},
			{
				Title:   "Failure Modes",
				Content: "TODO: Document common failure modes and how to avoid them.",
			},
		},
	}

	// Add references
	for _, fp := range filePaths {
		draft.References = append(draft.References, yamlReference{
			Path:        fp,
			Description: "Referenced in signals",
		})
	}

	data, err := yaml.Marshal(draft)
	if err != nil {
		return "", fmt.Errorf("failed to marshal draft: %w", err)
	}

	outPath := filepath.Join(d.outputDir, sanitize(domain)+".yaml")
	if err := os.WriteFile(outPath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write draft: %w", err)
	}

	return outPath, nil
}

// titleCase capitalizes the first letter of a string.
func titleCase(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// sanitize makes a string safe for use in filenames.
func sanitize(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "/", "-")
	s = strings.ReplaceAll(s, "_", "-")
	if len(s) > 50 {
		s = s[:50]
	}
	return s
}
