package brain

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIndexMatch(t *testing.T) {
	idx := NewIndex([]IndexEntry{
		{
			ID:   "billing",
			Name: "Billing Domain",
			Triggers: Triggers{
				Keywords:     []string{"invoice", "payment", "billing"},
				Entities:     []string{"Invoice", "Subscription"},
				FilePatterns: []string{"packs/billing/"},
			},
			Confidence: 0.9,
		},
		{
			ID:   "auth",
			Name: "Auth Domain",
			Triggers: Triggers{
				Keywords:     []string{"login", "auth", "session"},
				Entities:     []string{"User", "Session"},
				FilePatterns: []string{"packs/auth/"},
			},
			Confidence: 0.85,
		},
		{
			ID:   "unrelated",
			Name: "Unrelated",
			Triggers: Triggers{
				Keywords: []string{"unrelated"},
			},
			Confidence: 0.5,
		},
	})

	// Match by keyword
	matches := idx.Match(MatchContext{Keywords: []string{"billing", "payment"}})
	if len(matches) != 1 {
		t.Fatalf("Expected 1 match for billing keywords, got %d", len(matches))
	}
	if matches[0].ID != "billing" {
		t.Errorf("Expected billing match, got %s", matches[0].ID)
	}

	// Match by entity
	matches = idx.Match(MatchContext{Entities: []string{"User"}})
	if len(matches) != 1 {
		t.Fatalf("Expected 1 match for User entity, got %d", len(matches))
	}
	if matches[0].ID != "auth" {
		t.Errorf("Expected auth match, got %s", matches[0].ID)
	}

	// Match by file pattern
	matches = idx.Match(MatchContext{FilePaths: []string{"packs/billing/models/invoice.rb"}})
	if len(matches) != 1 {
		t.Fatalf("Expected 1 match for billing file, got %d", len(matches))
	}
	if matches[0].ID != "billing" {
		t.Errorf("Expected billing match, got %s", matches[0].ID)
	}

	// No match
	matches = idx.Match(MatchContext{Keywords: []string{"nothing"}})
	if len(matches) != 0 {
		t.Errorf("Expected 0 matches, got %d", len(matches))
	}

	// Multiple matches sorted by score
	matches = idx.Match(MatchContext{
		Keywords: []string{"billing", "auth"},
	})
	if len(matches) != 2 {
		t.Fatalf("Expected 2 matches, got %d", len(matches))
	}
}

func TestIndexMatchCaseInsensitive(t *testing.T) {
	idx := NewIndex([]IndexEntry{
		{
			ID:   "test",
			Name: "Test",
			Triggers: Triggers{
				Keywords: []string{"Billing"},
				Entities: []string{"Invoice"},
			},
		},
	})

	matches := idx.Match(MatchContext{Keywords: []string{"billing"}})
	if len(matches) != 1 {
		t.Errorf("Expected case-insensitive keyword match, got %d matches", len(matches))
	}

	matches = idx.Match(MatchContext{Entities: []string{"invoice"}})
	if len(matches) != 1 {
		t.Errorf("Expected case-insensitive entity match, got %d matches", len(matches))
	}
}

func TestIndexMatchScoring(t *testing.T) {
	idx := NewIndex([]IndexEntry{
		{
			ID:   "weak",
			Name: "Weak Match",
			Triggers: Triggers{
				Keywords: []string{"test"},
			},
		},
		{
			ID:   "strong",
			Name: "Strong Match",
			Triggers: Triggers{
				Keywords: []string{"test"},
				Entities: []string{"TestEntity"},
			},
		},
	})

	// Entity match should score higher, so "strong" should be first
	matches := idx.Match(MatchContext{
		Keywords: []string{"test"},
		Entities: []string{"TestEntity"},
	})
	if len(matches) != 2 {
		t.Fatalf("Expected 2 matches, got %d", len(matches))
	}
	if matches[0].ID != "strong" {
		t.Errorf("Expected strong match first, got %s", matches[0].ID)
	}
}

func TestIndexFromBrains(t *testing.T) {
	brains := []*Brain{
		{
			ID:          "b1",
			Name:        "Brain One",
			Description: "First",
			Triggers:    Triggers{Keywords: []string{"one"}},
			Confidence:  0.9,
		},
		{
			ID:          "b2",
			Name:        "Brain Two",
			Description: "Second",
			Triggers:    Triggers{Keywords: []string{"two"}},
			Confidence:  0.8,
		},
	}

	idx := IndexFromBrains(brains)
	if len(idx.Entries) != 2 {
		t.Fatalf("Expected 2 entries, got %d", len(idx.Entries))
	}
	if idx.Entries[0].ID != "b1" {
		t.Errorf("Expected b1, got %s", idx.Entries[0].ID)
	}
}

func TestIndexHandoff(t *testing.T) {
	idx := NewIndex([]IndexEntry{
		{ID: "a", Name: "Alpha", Description: "First brain", Confidence: 0.9},
		{ID: "b", Name: "Beta", Description: "Second brain", Confidence: 0.8},
	})

	// Full
	full := idx.Full()
	if !strings.Contains(full, "Alpha") || !strings.Contains(full, "Beta") {
		t.Error("Full should contain both brain names")
	}
	if !strings.Contains(full, "90%") {
		t.Error("Full should contain confidence percentage")
	}

	// Concise
	concise := idx.Concise()
	if !strings.Contains(concise, "Alpha") || !strings.Contains(concise, "Beta") {
		t.Error("Concise should contain both brain names")
	}

	// Type
	if idx.Type() != "brain-index" {
		t.Errorf("Expected type 'brain-index', got %q", idx.Type())
	}

	// ForTokenBudget with large budget returns full
	result := idx.ForTokenBudget(10000)
	if result != full {
		t.Error("Large budget should return full content")
	}

	// ForTokenBudget with tiny budget returns concise
	result = idx.ForTokenBudget(5)
	if result != concise {
		t.Error("Tiny budget should return concise content")
	}
}

func TestMatchFilePattern(t *testing.T) {
	tests := []struct {
		pattern  string
		filePath string
		expected bool
	}{
		{"packs/billing/", "packs/billing/models/invoice.rb", true},
		{"packs/billing/", "packs/auth/models/user.rb", false},
		{"*.rb", "user.rb", true},
		{"*.go", "user.rb", false},
	}

	for _, tc := range tests {
		result := matchFilePattern(tc.pattern, tc.filePath)
		if result != tc.expected {
			t.Errorf("matchFilePattern(%q, %q) = %v, want %v",
				tc.pattern, tc.filePath, result, tc.expected)
		}
	}
}

func TestLoaderLoadIndex(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "brain-loader-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create two brain files
	writeBrainJSON(t, tmpDir, "billing.json", &Brain{
		ID:       "billing",
		Name:     "Billing",
		Version:  1,
		Triggers: Triggers{Keywords: []string{"billing"}},
		Sections: []Section{{Title: "T", Content: "C"}},
	})
	writeBrainJSON(t, tmpDir, "auth.json", &Brain{
		ID:       "auth",
		Name:     "Auth",
		Version:  1,
		Triggers: Triggers{Keywords: []string{"auth"}},
		Sections: []Section{{Title: "T", Content: "C"}},
	})

	loader := NewLoader([]string{tmpDir}, &JSONReader{})
	idx, err := loader.LoadIndex()
	if err != nil {
		t.Fatalf("LoadIndex failed: %v", err)
	}

	if len(idx.Entries) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(idx.Entries))
	}
}

func TestLoaderLoadBrain(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "brain-loader-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	writeBrainJSON(t, tmpDir, "test.json", &Brain{
		ID:         "test",
		Name:       "Test Brain",
		Version:    1,
		Confidence: 0.9,
		Triggers:   Triggers{Keywords: []string{"test"}},
		Sections:   []Section{{Title: "Overview", Content: "Details"}},
	})

	loader := NewLoader([]string{tmpDir}, &JSONReader{})
	brain, err := loader.LoadBrain("test")
	if err != nil {
		t.Fatalf("LoadBrain failed: %v", err)
	}

	if brain.Name != "Test Brain" {
		t.Errorf("Expected 'Test Brain', got %q", brain.Name)
	}
}

func TestLoaderLoadBrainNotFound(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "brain-loader-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	loader := NewLoader([]string{tmpDir}, &JSONReader{})
	_, err = loader.LoadBrain("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent brain")
	}
}

func TestLoaderMultiDirPriority(t *testing.T) {
	dir1, _ := os.MkdirTemp("", "brain-prio1")
	dir2, _ := os.MkdirTemp("", "brain-prio2")
	defer os.RemoveAll(dir1)
	defer os.RemoveAll(dir2)

	// Same ID in both dirs, different names
	writeBrainJSON(t, dir1, "test.json", &Brain{
		ID:       "shared",
		Name:     "Project Version",
		Version:  1,
		Triggers: Triggers{Keywords: []string{"test"}},
		Sections: []Section{{Title: "T", Content: "C"}},
	})
	writeBrainJSON(t, dir2, "test.json", &Brain{
		ID:       "shared",
		Name:     "Global Version",
		Version:  1,
		Triggers: Triggers{Keywords: []string{"test"}},
		Sections: []Section{{Title: "T", Content: "C"}},
	})

	// dir1 has priority
	loader := NewLoader([]string{dir1, dir2}, &JSONReader{})

	// Index should deduplicate by ID
	idx, _ := loader.LoadIndex()
	if len(idx.Entries) != 1 {
		t.Fatalf("Expected 1 entry (deduped), got %d", len(idx.Entries))
	}
	if idx.Entries[0].Name != "Project Version" {
		t.Errorf("Expected project version (first dir wins), got %q", idx.Entries[0].Name)
	}

	// LoadBrain should also return first dir
	brain, _ := loader.LoadBrain("shared")
	if brain.Name != "Project Version" {
		t.Errorf("Expected project version, got %q", brain.Name)
	}
}

func TestLoaderLoadBrains(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "brain-multi")
	defer os.RemoveAll(tmpDir)

	writeBrainJSON(t, tmpDir, "a.json", &Brain{
		ID: "a", Name: "A", Version: 1,
		Triggers: Triggers{Keywords: []string{"a"}},
		Sections: []Section{{Title: "T", Content: "C"}},
	})
	writeBrainJSON(t, tmpDir, "b.json", &Brain{
		ID: "b", Name: "B", Version: 1,
		Triggers: Triggers{Keywords: []string{"b"}},
		Sections: []Section{{Title: "T", Content: "C"}},
	})

	loader := NewLoader([]string{tmpDir}, &JSONReader{})
	brains, err := loader.LoadBrains([]string{"a", "b"})
	if err != nil {
		t.Fatalf("LoadBrains failed: %v", err)
	}
	if len(brains) != 2 {
		t.Errorf("Expected 2 brains, got %d", len(brains))
	}
}

func TestLoaderNonexistentDir(t *testing.T) {
	loader := NewLoader([]string{"/nonexistent/path"}, &JSONReader{})
	idx, err := loader.LoadIndex()
	if err != nil {
		t.Fatalf("LoadIndex should not error for missing dirs: %v", err)
	}
	if len(idx.Entries) != 0 {
		t.Errorf("Expected 0 entries, got %d", len(idx.Entries))
	}
}

func TestBrainHandoff(t *testing.T) {
	brains := []*Brain{
		{
			Name:        "Billing",
			Description: "Billing domain",
			Confidence:  0.9,
			Sections: []Section{
				{Title: "Rules", Content: "Rule 1"},
			},
			References: []Reference{
				{Path: "packs/billing/model.rb", Description: "Main model"},
			},
		},
		{
			Name:        "Auth",
			Description: "Auth domain",
			Confidence:  0.7,
			Sections: []Section{
				{Title: "Overview", Content: "Auth overview"},
			},
		},
	}

	h := NewBrainHandoff(brains)

	// Full
	full := h.Full()
	if !strings.Contains(full, "Billing") || !strings.Contains(full, "Auth") {
		t.Error("Full should contain both brains")
	}
	if !strings.Contains(full, "Rule 1") {
		t.Error("Full should contain section content")
	}
	if !strings.Contains(full, "packs/billing/model.rb") {
		t.Error("Full should contain references")
	}

	// Concise
	concise := h.Concise()
	if !strings.Contains(concise, "Billing") || !strings.Contains(concise, "Auth") {
		t.Error("Concise should contain both brain names")
	}

	// Type
	if h.Type() != "brain" {
		t.Errorf("Expected type 'brain', got %q", h.Type())
	}

	// ForTokenBudget with large budget
	result := h.ForTokenBudget(10000)
	if result != full {
		t.Error("Large budget should return full content")
	}
}

func TestBrainHandoffTokenBudgetDropsLowConfidence(t *testing.T) {
	brains := []*Brain{
		{
			Name:        "High",
			Description: "High confidence",
			Confidence:  0.9,
			Sections:    []Section{{Title: "T", Content: strings.Repeat("x", 400)}},
		},
		{
			Name:        "Low",
			Description: "Low confidence",
			Confidence:  0.3,
			Sections:    []Section{{Title: "T", Content: strings.Repeat("y", 400)}},
		},
	}

	h := NewBrainHandoff(brains)

	// Budget that fits one brain but not both (~200 tokens per brain with headers)
	result := h.ForTokenBudget(200)

	if !strings.Contains(result, "High") {
		t.Error("Should keep high confidence brain")
	}
	if strings.Contains(result, "Low") {
		t.Error("Should drop low confidence brain")
	}
}

func TestBrainHandoffEmpty(t *testing.T) {
	h := NewBrainHandoff(nil)
	if h.ForTokenBudget(1000) != "" {
		t.Error("Empty handoff should return empty string")
	}
}

func TestValidate(t *testing.T) {
	// Valid brain
	valid := &Brain{
		ID:         "test",
		Name:       "Test",
		Version:    1,
		Confidence: 0.8,
		Triggers:   Triggers{Keywords: []string{"test"}},
		Sections:   []Section{{Title: "Overview", Content: "Content"}},
	}

	errs := Validate(valid)
	if len(errs) != 0 {
		t.Errorf("Expected no errors for valid brain, got %v", errs)
	}
}

func TestValidateErrors(t *testing.T) {
	// Missing everything
	empty := &Brain{}
	errs := Validate(empty)
	if len(errs) < 4 {
		t.Errorf("Expected at least 4 errors for empty brain, got %d: %v", len(errs), errs)
	}

	// Bad confidence
	badConf := &Brain{
		ID:         "test",
		Name:       "Test",
		Version:    1,
		Confidence: 1.5,
		Triggers:   Triggers{Keywords: []string{"test"}},
		Sections:   []Section{{Title: "T", Content: "C"}},
	}
	errs = Validate(badConf)
	if len(errs) != 1 {
		t.Errorf("Expected 1 error for bad confidence, got %d: %v", len(errs), errs)
	}
}

func TestValidateTooManySections(t *testing.T) {
	sections := make([]Section, MaxSections+1)
	for i := range sections {
		sections[i] = Section{Title: "T", Content: "C"}
	}

	b := &Brain{
		ID:       "test",
		Name:     "Test",
		Version:  1,
		Triggers: Triggers{Keywords: []string{"test"}},
		Sections: sections,
	}

	errs := Validate(b)
	found := false
	for _, e := range errs {
		if e.Field == "sections" && strings.Contains(e.Message, "too many") {
			found = true
		}
	}
	if !found {
		t.Error("Expected 'too many sections' error")
	}
}

func TestCheckStaleness(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "brain-stale")
	defer os.RemoveAll(tmpDir)

	// Create a file
	testFile := filepath.Join(tmpDir, "test.go")
	os.WriteFile(testFile, []byte("package test"), 0644)

	b := &Brain{
		References: []Reference{
			{Path: "test.go", Description: "Test file", Checksum: "wrong-checksum"},
			{Path: "missing.go", Description: "Missing file", Checksum: "abc"},
			{Path: "no-checksum.go", Description: "No checksum"},
		},
	}

	stale := CheckStaleness(b, tmpDir)

	if len(stale) != 2 {
		t.Fatalf("Expected 2 stale references, got %d", len(stale))
	}

	// Check reasons
	reasons := map[string]string{}
	for _, s := range stale {
		reasons[s.Reference.Path] = s.Reason
	}

	if reasons["test.go"] != "changed" {
		t.Errorf("Expected 'changed' for test.go, got %q", reasons["test.go"])
	}
	if reasons["missing.go"] != "missing" {
		t.Errorf("Expected 'missing' for missing.go, got %q", reasons["missing.go"])
	}
}

func TestDefaultDirs(t *testing.T) {
	dirs := DefaultDirs("/path/to/project")
	if len(dirs) < 1 {
		t.Fatal("Expected at least 1 directory")
	}
	if dirs[0] != "/path/to/project/.boatman/brains" {
		t.Errorf("First dir should be project-level, got %s", dirs[0])
	}
	if len(dirs) >= 2 {
		if !strings.Contains(dirs[1], ".boatman/brains") {
			t.Errorf("Second dir should be global, got %s", dirs[1])
		}
	}
}

// Helper to write a brain as JSON in a directory.
func writeBrainJSON(t *testing.T, dir, filename string, b *Brain) {
	t.Helper()
	data, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal brain: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, filename), data, 0644); err != nil {
		t.Fatalf("Failed to write brain file: %v", err)
	}
}
