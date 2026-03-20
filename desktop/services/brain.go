// Package services provides business logic for the desktop app.
package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	harnessbrain "github.com/philjestin/boatman-ecosystem/harness/brain"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// BrainEntry is the frontend-friendly brain index entry.
type BrainEntry struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Confidence   float64  `json:"confidence"`
	Version      int      `json:"version"`
	LastUpdated  string   `json:"lastUpdated"`
	Keywords     []string `json:"keywords"`
	Entities     []string `json:"entities"`
	FilePatterns []string `json:"filePatterns"`
}

// BrainDetail is the full brain content for display.
type BrainDetail struct {
	BrainEntry
	Sections   []BrainSection   `json:"sections"`
	References []BrainReference `json:"references"`
}

// BrainSection is a named content block.
type BrainSection struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

// BrainReference is a code file reference.
type BrainReference struct {
	Path        string `json:"path"`
	Description string `json:"description"`
	Checksum    string `json:"checksum"`
}

// BrainValidationResult holds validation output.
type BrainValidationResult struct {
	Valid  bool              `json:"valid"`
	Errors []string          `json:"errors"`
	Stale  []StaleRefResult  `json:"stale"`
}

// StaleRefResult is a frontend-friendly stale reference.
type StaleRefResult struct {
	Path   string `json:"path"`
	Reason string `json:"reason"`
}

// SignalEntry is a frontend-friendly signal.
type SignalEntry struct {
	Type      string   `json:"type"`
	Domain    string   `json:"domain"`
	Details   string   `json:"details"`
	FilePaths []string `json:"filePaths"`
	Count     int      `json:"count"`
	FirstSeen string   `json:"firstSeen"`
	LastSeen  string   `json:"lastSeen"`
}

// BrainService manages brain operations for the desktop app.
type BrainService struct {
	ctx         context.Context
	projectPath string
	mu          sync.RWMutex
}

// NewBrainService creates a new brain service.
func NewBrainService() *BrainService {
	return &BrainService{}
}

// SetContext sets the Wails context for event emission.
func (s *BrainService) SetContext(ctx context.Context) {
	s.ctx = ctx
}

// SetProjectPath sets the current project for brain resolution.
func (s *BrainService) SetProjectPath(path string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.projectPath = path
}

func (s *BrainService) getProjectPath() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.projectPath
}

// ListBrains returns all available brains for the current project.
func (s *BrainService) ListBrains() ([]BrainEntry, error) {
	loader := s.newLoader()
	idx, err := loader.LoadIndex()
	if err != nil {
		return nil, fmt.Errorf("failed to load brain index: %w", err)
	}

	entries := make([]BrainEntry, len(idx.Entries))
	for i, e := range idx.Entries {
		entries[i] = BrainEntry{
			ID:           e.ID,
			Name:         e.Name,
			Description:  e.Description,
			Confidence:   e.Confidence,
			Keywords:     orEmpty(e.Triggers.Keywords),
			Entities:     orEmpty(e.Triggers.Entities),
			FilePatterns: orEmpty(e.Triggers.FilePatterns),
		}
	}
	return entries, nil
}

// GetBrain loads a specific brain by ID with full content.
func (s *BrainService) GetBrain(id string) (*BrainDetail, error) {
	loader := s.newLoader()
	b, err := loader.LoadBrain(id)
	if err != nil {
		return nil, err
	}
	return brainToDetail(b), nil
}

// MatchBrains finds brains matching the given keywords and file paths.
func (s *BrainService) MatchBrains(keywords []string, filePaths []string, entities []string) ([]BrainEntry, error) {
	loader := s.newLoader()
	idx, err := loader.LoadIndex()
	if err != nil {
		return nil, err
	}

	matches := idx.Match(harnessbrain.MatchContext{
		Keywords:  keywords,
		FilePaths: filePaths,
		Entities:  entities,
	})

	entries := make([]BrainEntry, len(matches))
	for i, m := range matches {
		entries[i] = BrainEntry{
			ID:           m.ID,
			Name:         m.Name,
			Description:  m.Description,
			Confidence:   m.Confidence,
			Keywords:     orEmpty(m.Triggers.Keywords),
			Entities:     orEmpty(m.Triggers.Entities),
			FilePatterns: orEmpty(m.Triggers.FilePatterns),
		}
	}

	// Emit event for UI
	if s.ctx != nil && len(entries) > 0 {
		runtime.EventsEmit(s.ctx, "brain:matched", map[string]any{
			"matches": entries,
			"count":   len(entries),
		})
	}

	return entries, nil
}

// ValidateBrain validates a brain by ID.
func (s *BrainService) ValidateBrain(id string) (*BrainValidationResult, error) {
	loader := s.newLoader()
	b, err := loader.LoadBrain(id)
	if err != nil {
		return nil, err
	}

	errs := harnessbrain.Validate(b)
	stale := harnessbrain.CheckStaleness(b, s.getProjectPath())

	errMsgs := make([]string, len(errs))
	for i, e := range errs {
		errMsgs[i] = e.Error()
	}

	staleResults := make([]StaleRefResult, len(stale))
	for i, sr := range stale {
		staleResults[i] = StaleRefResult{
			Path:   sr.Reference.Path,
			Reason: sr.Reason,
		}
	}

	return &BrainValidationResult{
		Valid:  len(errs) == 0,
		Errors: errMsgs,
		Stale:  staleResults,
	}, nil
}

// ListSignals returns all recorded signals.
func (s *BrainService) ListSignals() ([]SignalEntry, error) {
	store, err := harnessbrain.NewSignalStore("")
	if err != nil {
		return nil, err
	}

	signals := store.GetAll()
	entries := make([]SignalEntry, len(signals))
	for i, sig := range signals {
		entries[i] = SignalEntry{
			Type:      string(sig.Type),
			Domain:    sig.Domain,
			Details:   sig.Details,
			FilePaths: orEmpty(sig.FilePaths),
			Count:     sig.Count,
			FirstSeen: sig.FirstSeen.Format("2006-01-02T15:04:05Z"),
			LastSeen:  sig.LastSeen.Format("2006-01-02T15:04:05Z"),
		}
	}
	return entries, nil
}

// GetSignalsByDomain returns signals for a specific domain.
func (s *BrainService) GetSignalsByDomain(domain string) ([]SignalEntry, error) {
	store, err := harnessbrain.NewSignalStore("")
	if err != nil {
		return nil, err
	}

	signals := store.GetByDomain(domain)
	entries := make([]SignalEntry, len(signals))
	for i, sig := range signals {
		entries[i] = SignalEntry{
			Type:      string(sig.Type),
			Domain:    sig.Domain,
			Details:   sig.Details,
			FilePaths: orEmpty(sig.FilePaths),
			Count:     sig.Count,
			FirstSeen: sig.FirstSeen.Format("2006-01-02T15:04:05Z"),
			LastSeen:  sig.LastSeen.Format("2006-01-02T15:04:05Z"),
		}
	}
	return entries, nil
}

// GetBrainDirs returns the brain directories being scanned.
func (s *BrainService) GetBrainDirs() []string {
	pp := s.getProjectPath()
	if pp == "" {
		pp = "."
	}
	return harnessbrain.DefaultDirs(pp)
}

// newLoader creates a loader with YAML+JSON support for the current project.
func (s *BrainService) newLoader() *harnessbrain.Loader {
	pp := s.getProjectPath()
	if pp == "" {
		pp = "."
	}
	reader := &compositeReader{}
	return harnessbrain.NewLoader(harnessbrain.DefaultDirs(pp), reader)
}

// compositeReader supports both JSON and YAML brain files.
// YAML parsing uses a simple approach since we can't import gopkg.in/yaml.v3
// in the desktop module without adding the dependency. For the desktop app,
// we parse the YAML files using encoding/json after a simple YAML-to-JSON conversion,
// or we just support JSON files directly and rely on the harness JSONReader.
type compositeReader struct{}

func (r *compositeReader) Read(path string) (*harnessbrain.Brain, error) {
	ext := filepath.Ext(path)
	switch ext {
	case ".json":
		jr := &harnessbrain.JSONReader{}
		return jr.Read(path)
	case ".yaml", ".yml":
		return readYAMLBrain(path)
	default:
		return nil, fmt.Errorf("unsupported extension: %s", ext)
	}
}

func (r *compositeReader) Extensions() []string {
	return []string{".json", ".yaml", ".yml"}
}

// readYAMLBrain reads a YAML brain file by shelling out to a simple parser
// or using a lightweight approach. For now, we use Go's JSON parser on
// a pre-converted format, or we parse YAML manually for the simple schema.
func readYAMLBrain(path string) (*harnessbrain.Brain, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Try JSON first (some files might actually be JSON with yaml extension)
	var b harnessbrain.Brain
	if err := json.Unmarshal(data, &b); err == nil && b.ID != "" {
		return &b, nil
	}

	// For YAML, we need a real parser. Use the simple YAML format that maps
	// directly to the Brain struct. Since the desktop module already has access
	// to the harness, we'll add gopkg.in/yaml.v3 as a dependency.
	return parseYAML(data, path)
}

func brainToDetail(b *harnessbrain.Brain) *BrainDetail {
	sections := make([]BrainSection, len(b.Sections))
	for i, s := range b.Sections {
		sections[i] = BrainSection{Title: s.Title, Content: s.Content}
	}

	refs := make([]BrainReference, len(b.References))
	for i, r := range b.References {
		refs[i] = BrainReference{Path: r.Path, Description: r.Description, Checksum: r.Checksum}
	}

	return &BrainDetail{
		BrainEntry: BrainEntry{
			ID:           b.ID,
			Name:         b.Name,
			Description:  b.Description,
			Confidence:   b.Confidence,
			Version:      b.Version,
			LastUpdated:  b.LastUpdated,
			Keywords:     orEmpty(b.Triggers.Keywords),
			Entities:     orEmpty(b.Triggers.Entities),
			FilePatterns: orEmpty(b.Triggers.FilePatterns),
		},
		Sections:   sections,
		References: refs,
	}
}

func orEmpty(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}

// AutoDistillResult describes a brain that was auto-generated.
type AutoDistillResult struct {
	Domain  string `json:"domain"`
	BrainID string `json:"brainId"`
	Path    string `json:"path"`
	Signals int    `json:"signals"`
	UsedLLM bool   `json:"usedLlm"`
}

// AutoDistillBrains checks signal thresholds and generates brains for ready domains.
// Uses template-based generation (no LLM in desktop context).
// Returns results and emits brain:auto-generated event.
func (s *BrainService) AutoDistillBrains() ([]AutoDistillResult, error) {
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

	// Load existing brain IDs to avoid re-generating
	loader := s.newLoader()
	idx, _ := loader.LoadIndex()
	existing := make(map[string]bool)
	if idx != nil {
		for _, e := range idx.Entries {
			existing[e.ID] = true
		}
	}

	pp := s.getProjectPath()
	if pp == "" {
		pp = "."
	}
	outputDir := filepath.Join(pp, ".boatman", "brains")

	var results []AutoDistillResult

	for domain, domainSignals := range clusters {
		brainID := "auto-" + sanitizeDomain(domain)
		if existing[brainID] {
			continue
		}

		// Check thresholds: 3+ signals OR 1 signal with count >= 5
		ready := len(domainSignals) >= 3
		if !ready {
			for _, sig := range domainSignals {
				if sig.Count >= 5 {
					ready = true
					break
				}
			}
		}
		if !ready {
			continue
		}

		// Generate template-based brain
		path, err := s.generateTemplateBrain(outputDir, domain, brainID, domainSignals)
		if err != nil {
			continue
		}

		results = append(results, AutoDistillResult{
			Domain:  domain,
			BrainID: brainID,
			Path:    path,
			Signals: len(domainSignals),
			UsedLLM: false,
		})
	}

	// Emit event for UI notification
	if s.ctx != nil && len(results) > 0 {
		runtime.EventsEmit(s.ctx, "brain:auto-generated", map[string]any{
			"results": results,
			"count":   len(results),
		})
	}

	// Record distillation timestamp
	if len(results) > 0 {
		recordDistilled()
	}

	return results, nil
}

// ShouldAutoDistill checks if enough time has passed since last distillation.
func ShouldAutoDistill() bool {
	homeDir, _ := os.UserHomeDir()
	path := filepath.Join(homeDir, ".boatman", "signals", "last_distilled.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return true // never distilled
	}
	var ts struct {
		At time.Time `json:"at"`
	}
	if err := json.Unmarshal(data, &ts); err != nil || ts.At.IsZero() {
		return true
	}
	return time.Since(ts.At) > 24*time.Hour
}

func recordDistilled() {
	homeDir, _ := os.UserHomeDir()
	dir := filepath.Join(homeDir, ".boatman", "signals")
	os.MkdirAll(dir, 0755)
	ts := struct {
		At time.Time `json:"at"`
	}{At: time.Now()}
	data, _ := json.Marshal(ts)
	os.WriteFile(filepath.Join(dir, "last_distilled.json"), data, 0644)
}

func (s *BrainService) generateTemplateBrain(outputDir, domain, brainID string, signals []harnessbrain.Signal) (string, error) {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", err
	}

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

	// Build a simple JSON brain (no yaml dependency needed)
	brain := &harnessbrain.Brain{
		ID:          brainID,
		Name:        domain + " Domain",
		Version:     1,
		Description: fmt.Sprintf("Auto-generated brain for the %s domain based on %d signals.", domain, len(signals)),
		Triggers: harnessbrain.Triggers{
			Keywords: []string{domain},
		},
		Confidence:  0.5,
		LastUpdated: time.Now().Format("2006-01-02"),
		Sections: []harnessbrain.Section{
			{Title: "Signals Detected", Content: joinLines(details)},
			{Title: "Domain Model", Content: "Auto-generated — review and curate. Key files: " + joinLines(filePaths)},
			{Title: "Failure Modes", Content: "Based on signals, agents frequently struggle with:\n" + joinLines(details)},
		},
	}

	for _, fp := range filePaths {
		brain.References = append(brain.References, harnessbrain.Reference{
			Path:        fp,
			Description: "Referenced in signals",
		})
	}

	data, err := json.MarshalIndent(brain, "", "  ")
	if err != nil {
		return "", err
	}

	outPath := filepath.Join(outputDir, sanitizeDomain(domain)+".json")
	if err := os.WriteFile(outPath, data, 0644); err != nil {
		return "", err
	}

	return outPath, nil
}

func sanitizeDomain(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "/", "-")
	s = strings.ReplaceAll(s, "_", "-")
	if len(s) > 50 {
		s = s[:50]
	}
	return s
}

func joinLines(lines []string) string {
	return strings.Join(lines, "\n")
}
