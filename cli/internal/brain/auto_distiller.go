package brain

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	harnessbrain "github.com/philjestin/boatman-ecosystem/harness/brain"
	"github.com/philjestin/boatmanmode/internal/claude"
	"github.com/philjestin/boatmanmode/internal/config"
	"gopkg.in/yaml.v3"
)

// AutoDistiller combines signal thresholds, LLM synthesis, and file context
// to automatically generate brain YAML files from accumulated knowledge gaps.
type AutoDistiller struct {
	projectPath string
	outputDir   string
	cfg         *config.Config
}

// NewAutoDistiller creates a distiller that outputs to the project's .boatman/brains/ directory.
func NewAutoDistiller(projectPath string, cfg *config.Config) *AutoDistiller {
	return &AutoDistiller{
		projectPath: projectPath,
		outputDir:   filepath.Join(projectPath, ".boatman", "brains"),
		cfg:         cfg,
	}
}

// DistillResult describes what was generated.
type DistillResult struct {
	Domain    string
	Path      string
	BrainID   string
	Signals   int
	UsedLLM   bool
}

// SignalThreshold defines when a domain has enough signals to warrant a brain.
const (
	MinSignalsForDraft   = 3 // minimum distinct signals to trigger distillation
	MinCountForRecurring = 5 // a single signal recurring this many times also triggers
)

// ShouldDistill checks if any domain has enough signals to warrant brain generation.
// Returns domains that are ready.
func (d *AutoDistiller) ShouldDistill() ([]string, error) {
	store, err := harnessbrain.NewSignalStore("")
	if err != nil {
		return nil, err
	}

	signals := store.GetAll()
	if len(signals) == 0 {
		return nil, nil
	}

	// Cluster by domain
	clusters := make(map[string][]harnessbrain.Signal)
	for _, sig := range signals {
		clusters[sig.Domain] = append(clusters[sig.Domain], sig)
	}

	// Load existing brains to avoid re-generating
	existing := d.existingBrainIDs()

	var ready []string
	for domain, domainSignals := range clusters {
		brainID := "auto-" + sanitize(domain)
		if existing[brainID] {
			continue // already have a brain for this domain
		}

		if d.meetsThreshold(domainSignals) {
			ready = append(ready, domain)
		}
	}

	return ready, nil
}

// meetsThreshold checks if signals warrant brain generation.
func (d *AutoDistiller) meetsThreshold(signals []harnessbrain.Signal) bool {
	if len(signals) >= MinSignalsForDraft {
		return true
	}

	for _, sig := range signals {
		if sig.Count >= MinCountForRecurring {
			return true
		}
	}

	return false
}

// DistillAll checks thresholds and generates brains for all ready domains.
// Uses LLM when available, falls back to template-based generation.
func (d *AutoDistiller) DistillAll(ctx context.Context) ([]DistillResult, error) {
	domains, err := d.ShouldDistill()
	if err != nil {
		return nil, err
	}

	if len(domains) == 0 {
		return nil, nil
	}

	store, err := harnessbrain.NewSignalStore("")
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(d.outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	var results []DistillResult

	for _, domain := range domains {
		signals := store.GetByDomain(domain)
		result, err := d.distillDomain(ctx, domain, signals)
		if err != nil {
			fmt.Printf("   ⚠️  Auto-distill failed for %s: %v\n", domain, err)
			continue
		}
		results = append(results, *result)
	}

	return results, nil
}

// distillDomain generates a brain for a single domain.
func (d *AutoDistiller) distillDomain(ctx context.Context, domain string, signals []harnessbrain.Signal) (*DistillResult, error) {
	// Gather file contents for context
	fileContents := d.readReferencedFiles(signals)

	// Try LLM-powered synthesis first
	result, err := d.llmDistill(ctx, domain, signals, fileContents)
	if err == nil {
		return result, nil
	}

	// Fall back to template-based
	fmt.Printf("   📝 LLM distill unavailable for %s, using template: %v\n", domain, err)
	return d.templateDistill(domain, signals)
}

// llmDistill uses Claude to synthesize a brain from signals and file context.
func (d *AutoDistiller) llmDistill(ctx context.Context, domain string, signals []harnessbrain.Signal, fileContents map[string]string) (*DistillResult, error) {
	client := claude.New()
	client.WorkDir = d.projectPath
	client.SkipPermissions = true

	if d.cfg != nil && d.cfg.Claude.Models.Planner != "" {
		client.Model = d.cfg.Claude.Models.Planner
	}
	if d.cfg != nil && d.cfg.Claude.Effort != "" {
		client.Effort = d.cfg.Claude.Effort
	}

	systemPrompt := `You are a domain knowledge curator for an AI agent system called Boatman.
Your job is to synthesize observed knowledge gaps (signals) and source code into a structured
brain document that helps future agents work correctly in this domain.

Output ONLY valid YAML matching this schema — no markdown fences, no explanation:

id: auto-<domain>
name: <Domain> Domain
version: 1
description: <1-2 sentence description>
confidence: <0.6-0.9 based on signal quality>
last_updated: <today's date>
triggers:
  keywords: [<relevant keywords>]
  entities: [<key model/class names>]
  file_patterns: [<directory patterns like packs/foo/>]
sections:
  - title: Domain Model
    content: |
      <key models, relationships, invariants>
  - title: Business Rules
    content: |
      <numbered rules agents must follow>
  - title: Failure Modes
    content: |
      <common mistakes and how to avoid them>
  - title: Key Code Paths
    content: |
      <important function calls and patterns>
references:
  - path: <file path>
    description: <what this file does>

Guidelines:
- Extract concrete patterns from the code, don't make things up
- Business rules should be specific and actionable
- Failure modes should directly address the signals (review failures, refactor loops)
- Keep sections concise — agents have limited context windows`

	// Build user prompt with signals and file context
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Domain: %s\n\n", domain))

	sb.WriteString("## Signals (knowledge gaps detected during agent work)\n\n")
	for _, sig := range signals {
		sb.WriteString(fmt.Sprintf("- **%s** (count: %d): %s\n", sig.Type, sig.Count, sig.Details))
		if len(sig.FilePaths) > 0 {
			sb.WriteString(fmt.Sprintf("  Files: %s\n", strings.Join(sig.FilePaths, ", ")))
		}
	}

	if len(fileContents) > 0 {
		sb.WriteString("\n## Source Code Context\n\n")
		totalChars := 0
		maxChars := 30000 // Keep within reasonable prompt size
		for path, content := range fileContents {
			entry := fmt.Sprintf("### %s\n```\n%s\n```\n\n", path, content)
			if totalChars+len(entry) > maxChars {
				sb.WriteString(fmt.Sprintf("### %s\n(truncated — file too large)\n\n", path))
				continue
			}
			sb.WriteString(entry)
			totalChars += len(entry)
		}
	}

	sb.WriteString(fmt.Sprintf("\nGenerate the brain YAML for the %s domain. Today's date is %s.", domain, time.Now().Format("2006-01-02")))

	response, _, err := client.Message(ctx, systemPrompt, sb.String())
	if err != nil {
		return nil, fmt.Errorf("claude call failed: %w", err)
	}

	// Clean up response — strip markdown fences if present
	response = cleanYAMLResponse(response)

	// Validate it parses as YAML
	var parsed yamlBrain
	if err := yaml.Unmarshal([]byte(response), &parsed); err != nil {
		return nil, fmt.Errorf("LLM output is not valid YAML: %w", err)
	}

	if parsed.ID == "" {
		parsed.ID = "auto-" + sanitize(domain)
	}

	// Re-marshal to get clean YAML
	data, err := yaml.Marshal(parsed)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal brain: %w", err)
	}

	outPath := filepath.Join(d.outputDir, sanitize(domain)+".yaml")
	if err := os.WriteFile(outPath, data, 0644); err != nil {
		return nil, fmt.Errorf("failed to write brain: %w", err)
	}

	return &DistillResult{
		Domain:  domain,
		Path:    outPath,
		BrainID: parsed.ID,
		Signals: len(signals),
		UsedLLM: true,
	}, nil
}

// templateDistill generates a brain using templates (no LLM needed).
func (d *AutoDistiller) templateDistill(domain string, signals []harnessbrain.Signal) (*DistillResult, error) {
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

	keywords := []string{domain}
	for _, fp := range filePaths {
		base := strings.TrimSuffix(filepath.Base(fp), filepath.Ext(fp))
		if base != domain && len(base) > 2 {
			keywords = append(keywords, strings.ToLower(base))
		}
	}

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

	brainID := "auto-" + sanitize(domain)
	draft := yamlBrain{
		ID:          brainID,
		Name:        titleCase(domain) + " Domain",
		Version:     1,
		Description: fmt.Sprintf("Auto-generated brain for the %s domain based on %d signals.", domain, len(signals)),
		Triggers: yamlTriggers{
			Keywords:     keywords,
			FilePatterns: filePatterns,
		},
		Confidence:  0.5,
		LastUpdated: time.Now().Format("2006-01-02"),
		Sections: []yamlSection{
			{
				Title:   "Signals Detected",
				Content: strings.Join(details, "\n"),
			},
			{
				Title:   "Domain Model",
				Content: "Auto-generated — review and curate. Key files: " + strings.Join(filePaths, ", "),
			},
			{
				Title:   "Failure Modes",
				Content: "Based on signals, agents frequently struggle with:\n" + strings.Join(details, "\n"),
			},
		},
	}

	for _, fp := range filePaths {
		draft.References = append(draft.References, yamlReference{
			Path:        fp,
			Description: "Referenced in signals",
		})
	}

	data, err := yaml.Marshal(draft)
	if err != nil {
		return nil, err
	}

	outPath := filepath.Join(d.outputDir, sanitize(domain)+".yaml")
	if err := os.WriteFile(outPath, data, 0644); err != nil {
		return nil, err
	}

	return &DistillResult{
		Domain:  domain,
		Path:    outPath,
		BrainID: brainID,
		Signals: len(signals),
		UsedLLM: false,
	}, nil
}

// readReferencedFiles reads file contents from the project for LLM context.
func (d *AutoDistiller) readReferencedFiles(signals []harnessbrain.Signal) map[string]string {
	contents := make(map[string]string)
	seen := make(map[string]bool)

	for _, sig := range signals {
		for _, fp := range sig.FilePaths {
			if seen[fp] {
				continue
			}
			seen[fp] = true

			fullPath := filepath.Join(d.projectPath, fp)
			data, err := os.ReadFile(fullPath)
			if err != nil {
				continue
			}

			// Skip very large files
			if len(data) > 20000 {
				contents[fp] = string(data[:20000]) + "\n... (truncated)"
			} else {
				contents[fp] = string(data)
			}
		}
	}

	return contents
}

// existingBrainIDs returns the set of brain IDs already present.
func (d *AutoDistiller) existingBrainIDs() map[string]bool {
	reader := NewCompositeReader()
	loader := harnessbrain.NewLoader([]string{d.outputDir}, reader)
	idx, err := loader.LoadIndex()
	if err != nil {
		return nil
	}

	ids := make(map[string]bool)
	for _, e := range idx.Entries {
		ids[e.ID] = true
	}
	return ids
}

// cleanYAMLResponse strips markdown fences and leading text from LLM output.
func cleanYAMLResponse(s string) string {
	s = strings.TrimSpace(s)

	// Strip ```yaml ... ``` fences
	if strings.HasPrefix(s, "```") {
		lines := strings.SplitN(s, "\n", 2)
		if len(lines) == 2 {
			s = lines[1]
		}
		if idx := strings.LastIndex(s, "```"); idx > 0 {
			s = s[:idx]
		}
	}

	// If it starts with text before the first YAML key, strip it
	if !strings.HasPrefix(s, "id:") {
		if idx := strings.Index(s, "\nid:"); idx >= 0 {
			s = s[idx+1:]
		}
	}

	return strings.TrimSpace(s)
}

// lastDistilledPath returns the path to the last-distilled timestamp file.
func lastDistilledPath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".boatman", "signals", "last_distilled.json")
}

// LastDistilledAt returns when distillation last ran.
func LastDistilledAt() time.Time {
	data, err := os.ReadFile(lastDistilledPath())
	if err != nil {
		return time.Time{}
	}
	var ts struct {
		At time.Time `json:"at"`
	}
	json.Unmarshal(data, &ts)
	return ts.At
}

// RecordDistilled records that distillation ran now.
func RecordDistilled() {
	ts := struct {
		At time.Time `json:"at"`
	}{At: time.Now()}
	data, _ := json.Marshal(ts)
	os.MkdirAll(filepath.Dir(lastDistilledPath()), 0755)
	os.WriteFile(lastDistilledPath(), data, 0644)
}

// ShouldRunPeriodicDistill checks if enough time has passed since last distillation.
func ShouldRunPeriodicDistill(interval time.Duration) bool {
	last := LastDistilledAt()
	if last.IsZero() {
		return true
	}
	return time.Since(last) > interval
}
