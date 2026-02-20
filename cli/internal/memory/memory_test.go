package memory

import (
	"testing"
	"time"

	harnessmem "github.com/philjestin/boatman-ecosystem/harness/memory"
)

// TestTypeAliases verifies that type aliases are properly exported
func TestTypeAliases(t *testing.T) {
	// Verify Memory alias
	var _ Memory = harnessmem.Memory{}

	// Verify Pattern alias
	var _ Pattern = harnessmem.Pattern{}

	// Verify CommonIssue alias
	var _ CommonIssue = harnessmem.CommonIssue{}

	// Verify PromptRecord alias
	var _ PromptRecord = harnessmem.PromptRecord{}

	// Verify Preferences alias
	var _ Preferences = harnessmem.Preferences{}

	// Verify SessionStats alias
	var _ SessionStats = harnessmem.SessionStats{}

	// Verify Store alias (pointer type since Store is a struct)
	var _ *Store = (*harnessmem.Store)(nil)

	// Verify Analyzer alias (pointer type since Analyzer is a struct)
	var _ *Analyzer = (*harnessmem.Analyzer)(nil)

	t.Log("All type aliases verified successfully")
}

// TestNewStore verifies NewStore function alias works
func TestNewStore(t *testing.T) {
	// NewStore returns (*Store, error)
	store, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}
	if store == nil {
		t.Fatal("NewStore returned nil")
	}

	// Verify it creates a harness.Store
	var _ *harnessmem.Store = store

	t.Log("NewStore alias verified successfully")
}

// TestNewAnalyzer verifies NewAnalyzer function alias works
func TestNewAnalyzer(t *testing.T) {
	store, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}

	// Get a Memory instance to pass to NewAnalyzer
	mem, err := store.Get("/test/project")
	if err != nil {
		t.Fatalf("store.Get failed: %v", err)
	}

	// NewAnalyzer takes *Memory, not *Store
	analyzer := NewAnalyzer(mem)
	if analyzer == nil {
		t.Fatal("NewAnalyzer returned nil")
	}

	// Verify it creates a harness.Analyzer
	var _ *harnessmem.Analyzer = analyzer

	t.Log("NewAnalyzer alias verified successfully")
}

// TestMemoryStruct tests the Memory struct through alias
func TestMemoryStruct(t *testing.T) {
	memory := &Memory{
		Patterns:     []Pattern{},
		CommonIssues: []CommonIssue{},
		Preferences:  Preferences{},
	}

	if memory.Patterns == nil {
		t.Error("Expected Patterns to be initialized")
	}

	if memory.CommonIssues == nil {
		t.Error("Expected CommonIssues to be initialized")
	}

	t.Log("Memory struct works through alias")
}

// TestPatternStruct tests the Pattern struct through alias
func TestPatternStruct(t *testing.T) {
	pattern := &Pattern{
		Type:        "test-pattern",
		Description: "A test pattern",
		UsageCount:  5,
		Weight:      0.8,
	}

	if pattern.Type != "test-pattern" {
		t.Errorf("Expected Type = 'test-pattern', got %s", pattern.Type)
	}

	if pattern.UsageCount != 5 {
		t.Errorf("Expected UsageCount = 5, got %d", pattern.UsageCount)
	}

	t.Log("Pattern struct works through alias")
}

// TestCommonIssueStruct tests the CommonIssue struct through alias
func TestCommonIssueStruct(t *testing.T) {
	issue := &CommonIssue{
		Description: "Common test issue",
		Solution:    "Test solution",
		Frequency:   3,
	}

	if issue.Description != "Common test issue" {
		t.Errorf("Expected Description = 'Common test issue', got %s", issue.Description)
	}

	if issue.Frequency != 3 {
		t.Errorf("Expected Frequency = 3, got %d", issue.Frequency)
	}

	t.Log("CommonIssue struct works through alias")
}

// TestPromptRecordStruct tests the PromptRecord struct through alias
func TestPromptRecordStruct(t *testing.T) {
	record := &PromptRecord{
		Prompt:       "Test prompt",
		Result:       "Test result",
		SuccessScore: 85,
	}

	if record.Prompt != "Test prompt" {
		t.Errorf("Expected Prompt = 'Test prompt', got %s", record.Prompt)
	}

	if record.SuccessScore != 85 {
		t.Errorf("Expected SuccessScore = 85, got %d", record.SuccessScore)
	}

	t.Log("PromptRecord struct works through alias")
}

// TestPreferencesStruct tests the Preferences struct through alias
func TestPreferencesStruct(t *testing.T) {
	prefs := &Preferences{
		PreferredTestFramework: "testing",
		NamingConventions:      map[string]string{"go": "camelCase"},
		CodeStyle:              map[string]string{"indent": "tabs"},
	}

	if prefs.PreferredTestFramework != "testing" {
		t.Errorf("Expected PreferredTestFramework = 'testing', got %s", prefs.PreferredTestFramework)
	}

	if len(prefs.NamingConventions) != 1 {
		t.Errorf("Expected 1 naming convention, got %d", len(prefs.NamingConventions))
	}

	t.Log("Preferences struct works through alias")
}

// TestSessionStatsStruct tests the SessionStats struct through alias
func TestSessionStatsStruct(t *testing.T) {
	stats := &SessionStats{
		TotalSessions:      10,
		SuccessfulSessions: 8,
		TotalIterations:    20,
	}

	if stats.TotalSessions != 10 {
		t.Errorf("Expected TotalSessions = 10, got %d", stats.TotalSessions)
	}

	if stats.SuccessfulSessions != 8 {
		t.Errorf("Expected SuccessfulSessions = 8, got %d", stats.SuccessfulSessions)
	}

	t.Log("SessionStats struct works through alias")
}

// TestStoreBasicOperations tests basic store operations through aliases
func TestStoreBasicOperations(t *testing.T) {
	store, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}

	// Get memory for a project
	mem, err := store.Get("/test/project")
	if err != nil {
		t.Fatalf("store.Get failed: %v", err)
	}
	if mem == nil {
		t.Fatal("Expected Get to return non-nil memory")
	}

	// Learn a pattern
	mem.LearnPattern(Pattern{
		ID:          "test-pattern",
		Type:        "naming",
		Description: "test description",
		Weight:      0.8,
	})

	// Learn an issue
	mem.LearnIssue(CommonIssue{
		ID:          "test-issue",
		Type:        "style",
		Description: "test issue description",
		Solution:    "test solution",
	})

	// Learn a prompt
	mem.LearnPrompt("feature", "test prompt", "test result", 80)

	// Update stats
	mem.UpdateStats(true, 2, time.Minute)

	// Save
	if err := store.Save(mem); err != nil {
		t.Fatalf("store.Save failed: %v", err)
	}

	t.Log("Store basic operations work through aliases")
}

// TestAnalyzerBasicOperations tests basic analyzer operations through aliases
func TestAnalyzerBasicOperations(t *testing.T) {
	store, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}

	mem, err := store.Get("/test/project")
	if err != nil {
		t.Fatalf("store.Get failed: %v", err)
	}

	analyzer := NewAnalyzer(mem)

	// Test analyzing a success
	analyzer.AnalyzeSuccess([]string{"main.go", "main_test.go"}, 90)

	// Test analyzing an issue
	analyzer.AnalyzeIssue("warning", "style issue description", "use gofmt", "main.go")

	// Verify patterns were learned
	patterns := mem.GetPatternsForFile("main.go")
	if patterns == nil {
		t.Error("Expected GetPatternsForFile to return non-nil slice")
	}

	t.Log("Analyzer basic operations work through aliases")
}

// TestPackageBackwardCompatibility verifies the package maintains backward compatibility
func TestPackageBackwardCompatibility(t *testing.T) {
	// Types
	var _ Memory
	var _ Pattern
	var _ CommonIssue
	var _ PromptRecord
	var _ Preferences
	var _ SessionStats
	var _ *Store
	var _ *Analyzer

	// Functions
	_ = NewStore
	_ = NewAnalyzer

	t.Log("All exports are available for backward compatibility")
}
