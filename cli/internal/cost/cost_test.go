package cost

import (
	"testing"

	harnesscost "github.com/philjestin/boatman-ecosystem/harness/cost"
)

// TestTypeAliases verifies that type aliases are properly exported
func TestTypeAliases(t *testing.T) {
	// Verify Usage alias
	var _ Usage = harnesscost.Usage{}

	// Verify StepUsage alias
	var _ StepUsage = harnesscost.StepUsage{}

	// Verify Tracker alias
	var _ Tracker = (*harnesscost.Tracker)(nil)

	t.Log("All type aliases verified successfully")
}

// TestNewTracker verifies NewTracker function alias works
func TestNewTracker(t *testing.T) {
	// Verify function exists and is callable
	tracker := NewTracker()
	if tracker == nil {
		t.Fatal("NewTracker returned nil")
	}

	// Verify it creates a harness.Tracker
	var _ *harnesscost.Tracker = tracker

	t.Log("NewTracker alias verified successfully")
}

// TestTrackerBasicOperations tests basic tracker operations through aliases
func TestTrackerBasicOperations(t *testing.T) {
	tracker := NewTracker()

	// Test adding usage
	usage := Usage{
		InputTokens:  100,
		OutputTokens: 50,
		TotalCostUSD: 0.001,
	}
	tracker.Add("test-step", usage)

	// Test getting total
	total := tracker.Total()
	if total.InputTokens != 100 {
		t.Errorf("Expected InputTokens = 100, got %d", total.InputTokens)
	}

	if total.OutputTokens != 50 {
		t.Errorf("Expected OutputTokens = 50, got %d", total.OutputTokens)
	}

	// Test HasUsage
	if !tracker.HasUsage() {
		t.Error("Expected HasUsage() = true")
	}

	t.Log("Tracker basic operations work through aliases")
}

// TestUsageStruct tests the Usage struct through alias
func TestUsageStruct(t *testing.T) {
	usage := &Usage{
		InputTokens:  1000,
		OutputTokens: 500,
		TotalCostUSD: 0.01,
	}

	if usage.InputTokens != 1000 {
		t.Errorf("Expected InputTokens = 1000, got %d", usage.InputTokens)
	}

	if usage.OutputTokens != 500 {
		t.Errorf("Expected OutputTokens = 500, got %d", usage.OutputTokens)
	}

	// Test Add method
	other := Usage{
		InputTokens:  100,
		OutputTokens: 50,
		TotalCostUSD: 0.001,
	}

	combined := usage.Add(other)
	if combined.InputTokens != 1100 {
		t.Errorf("Expected combined InputTokens = 1100, got %d", combined.InputTokens)
	}

	// Test IsEmpty method
	empty := Usage{}
	if !empty.IsEmpty() {
		t.Error("Expected IsEmpty() = true for empty usage")
	}

	t.Log("Usage struct methods work through alias")
}

// TestStepUsageStruct tests the StepUsage struct through alias
func TestStepUsageStruct(t *testing.T) {
	stepUsage := &StepUsage{
		Step: "test-step",
		Usage: Usage{
			InputTokens:  100,
			OutputTokens: 50,
			TotalCostUSD: 0.001,
		},
	}

	if stepUsage.Step != "test-step" {
		t.Errorf("Expected Step = 'test-step', got %s", stepUsage.Step)
	}

	if stepUsage.Usage.InputTokens != 100 {
		t.Errorf("Expected InputTokens = 100, got %d", stepUsage.Usage.InputTokens)
	}

	t.Log("StepUsage struct works through alias")
}

// TestMultipleStepTracking tests tracking multiple steps
func TestMultipleStepTracking(t *testing.T) {
	tracker := NewTracker()

	// Add usage for multiple steps
	tracker.Add("planning", Usage{InputTokens: 200, OutputTokens: 100})
	tracker.Add("execution", Usage{InputTokens: 300, OutputTokens: 150})
	tracker.Add("review", Usage{InputTokens: 100, OutputTokens: 50})

	total := tracker.Total()

	expectedInput := 200 + 300 + 100
	expectedOutput := 100 + 150 + 50

	if total.InputTokens != expectedInput {
		t.Errorf("Expected InputTokens = %d, got %d", expectedInput, total.InputTokens)
	}

	if total.OutputTokens != expectedOutput {
		t.Errorf("Expected OutputTokens = %d, got %d", expectedOutput, total.OutputTokens)
	}

	// Test Steps method
	steps := tracker.Steps()
	if len(steps) != 3 {
		t.Errorf("Expected 3 steps, got %d", len(steps))
	}

	t.Log("Multiple step tracking works correctly")
}

// TestPackageBackwardCompatibility verifies the package maintains backward compatibility
func TestPackageBackwardCompatibility(t *testing.T) {
	// This test ensures that code using the old import path still works
	// by verifying all exported types and functions are available

	// Types
	var _ Usage
	var _ StepUsage
	var _ *Tracker

	// Functions
	_ = NewTracker

	t.Log("All exports are available for backward compatibility")
}
