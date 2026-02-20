package memory

import (
	"testing"

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

	// Verify Store alias
	var _ Store = (*harnessmem.Store)(nil)

	// Verify Analyzer alias
	var _ Analyzer = (*harnessmem.Analyzer)(nil)

	t.Log("All type aliases verified successfully")
}

// TestNewStore verifies NewStore function alias works
func TestNewStore(t *testing.T) {
	// Verify function exists and is callable
	store := NewStore("/test/path")
	if store == nil {
		t.Fatal("NewStore returned nil")
	}

	// Verify it creates a harness.Store
	var _ *harnessmem.Store = store

	t.Log("NewStore alias verified successfully")
}

// TestNewAnalyzer verifies NewAnalyzer function alias works
func TestNewAnalyzer(t *testing.T) {
	store := NewStore("/test/path")

	// Verify function exists and is callable
	analyzer := NewAnalyzer(store)
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
		Frequency:   5,
	}

	if pattern.Type != "test-pattern" {
		t.Errorf("Expected Type = 'test-pattern', got %s", pattern.Type)
	}

	if pattern.Frequency != 5 {
		t.Errorf("Expected Frequency = 5, got %d", pattern.Frequency)
	}

	t.Log("Pattern struct works through alias")
}

// TestCommonIssueStruct tests the CommonIssue struct through alias
func TestCommonIssueStruct(t *testing.T) {
	issue := &CommonIssue{
		Description: "Common test issue",
		Solution:    "Test solution",
		Count:       3,
	}

	if issue.Description != "Common test issue" {
		t.Errorf("Expected Description = 'Common test issue', got %s", issue.Description)
	}

	if issue.Count != 3 {
		t.Errorf("Expected Count = 3, got %d", issue.Count)
	}

	t.Log("CommonIssue struct works through alias")
}

// TestPromptRecordStruct tests the PromptRecord struct through alias
func TestPromptRecordStruct(t *testing.T) {
	record := &PromptRecord{
		Prompt:   "Test prompt",
		Response: "Test response",
		Success:  true,
	}

	if record.Prompt != "Test prompt" {
		t.Errorf("Expected Prompt = 'Test prompt', got %s", record.Prompt)
	}

	if !record.Success {
		t.Error("Expected Success = true")
	}

	t.Log("PromptRecord struct works through alias")
}

// TestPreferencesStruct tests the Preferences struct through alias
func TestPreferencesStruct(t *testing.T) {
	prefs := &Preferences{
		CodingStyle:    "golang-standard",
		TestFramework:  "testing",
		PreferredTools: []string{"go", "git"},
	}

	if prefs.CodingStyle != "golang-standard" {
		t.Errorf("Expected CodingStyle = 'golang-standard', got %s", prefs.CodingStyle)
	}

	if len(prefs.PreferredTools) != 2 {
		t.Errorf("Expected 2 preferred tools, got %d", len(prefs.PreferredTools))
	}

	t.Log("Preferences struct works through alias")
}

// TestSessionStatsStruct tests the SessionStats struct through alias
func TestSessionStatsStruct(t *testing.T) {
	stats := &SessionStats{
		TotalPrompts:     10,
		SuccessfulPrompts: 8,
		FailedPrompts:    2,
	}

	if stats.TotalPrompts != 10 {
		t.Errorf("Expected TotalPrompts = 10, got %d", stats.TotalPrompts)
	}

	if stats.SuccessfulPrompts != 8 {
		t.Errorf("Expected SuccessfulPrompts = 8, got %d", stats.SuccessfulPrompts)
	}

	t.Log("SessionStats struct works through alias")
}

// TestStoreBasicOperations tests basic store operations through aliases
func TestStoreBasicOperations(t *testing.T) {
	// Create store with temp directory
	store := NewStore(t.TempDir())

	// Test recording a prompt
	store.RecordPrompt("test prompt", "test response", true)

	// Test recording a pattern
	store.RecordPattern("test-pattern", "description")

	// Test recording a common issue
	store.RecordCommonIssue("issue description", "solution")

	// Test getting memory (will be empty in basic test)
	memory := store.GetMemory()
	if memory == nil {
		t.Error("Expected GetMemory to return non-nil memory")
	}

	t.Log("Store basic operations work through aliases")
}

// TestAnalyzerBasicOperations tests basic analyzer operations through aliases
func TestAnalyzerBasicOperations(t *testing.T) {
	store := NewStore(t.TempDir())
	analyzer := NewAnalyzer(store)

	// Record some test data
	store.RecordPrompt("test prompt 1", "response 1", true)
	store.RecordPrompt("test prompt 2", "response 2", false)
	store.RecordPattern("pattern1", "desc1")
	store.RecordCommonIssue("issue1", "solution1")

	// Test getting session stats
	stats := analyzer.GetSessionStats()
	if stats == nil {
		t.Fatal("Expected GetSessionStats to return non-nil stats")
	}

	t.Logf("Session stats: total=%d, successful=%d, failed=%d",
		stats.TotalPrompts, stats.SuccessfulPrompts, stats.FailedPrompts)

	// Test getting patterns
	patterns := analyzer.GetTopPatterns(10)
	if patterns == nil {
		t.Error("Expected GetTopPatterns to return non-nil slice")
	}

	// Test getting common issues
	issues := analyzer.GetCommonIssues(10)
	if issues == nil {
		t.Error("Expected GetCommonIssues to return non-nil slice")
	}

	t.Log("Analyzer basic operations work through aliases")
}

// TestPackageBackwardCompatibility verifies the package maintains backward compatibility
func TestPackageBackwardCompatibility(t *testing.T) {
	// This test ensures that code using the old import path still works
	// by verifying all exported types and functions are available

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
