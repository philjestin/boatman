package checkpoint

import (
	"testing"

	harnesscp "github.com/philjestin/boatman-ecosystem/harness/checkpoint"
)

// TestTypeAliases verifies that type aliases are properly exported
func TestTypeAliases(t *testing.T) {
	// Verify Step alias
	var _ Step = harnesscp.Step("")

	// Verify Status alias
	var _ Status = harnesscp.Status("")

	// Verify Checkpoint alias
	var _ Checkpoint = harnesscp.Checkpoint{}

	// Verify StepRecord alias
	var _ StepRecord = harnesscp.StepRecord{}

	// Verify Manager alias (pointer type since Manager is a struct)
	var _ *Manager = (*harnesscp.Manager)(nil)

	t.Log("All type aliases verified successfully")
}

// TestStepConstants verifies step constants match harness values
func TestStepConstants(t *testing.T) {
	tests := []struct {
		name    string
		local   Step
		harness harnesscp.Step
	}{
		{"StepFetchTicket", StepFetchTicket, harnesscp.StepFetchTicket},
		{"StepCreateWorktree", StepCreateWorktree, harnesscp.StepCreateWorktree},
		{"StepPlanning", StepPlanning, harnesscp.StepPlanning},
		{"StepValidation", StepValidation, harnesscp.StepValidation},
		{"StepExecution", StepExecution, harnesscp.StepExecution},
		{"StepTesting", StepTesting, harnesscp.StepTesting},
		{"StepReview", StepReview, harnesscp.StepReview},
		{"StepRefactor", StepRefactor, harnesscp.StepRefactor},
		{"StepVerify", StepVerify, harnesscp.StepVerify},
		{"StepCommit", StepCommit, harnesscp.StepCommit},
		{"StepPush", StepPush, harnesscp.StepPush},
		{"StepCreatePR", StepCreatePR, harnesscp.StepCreatePR},
		{"StepComplete", StepComplete, harnesscp.StepComplete},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.local != tt.harness {
				t.Errorf("%s: local = %v, harness = %v", tt.name, tt.local, tt.harness)
			}
		})
	}
}

// TestStatusConstants verifies status constants match harness values
func TestStatusConstants(t *testing.T) {
	tests := []struct {
		name    string
		local   Status
		harness harnesscp.Status
	}{
		{"StatusPending", StatusPending, harnesscp.StatusPending},
		{"StatusInProgress", StatusInProgress, harnesscp.StatusInProgress},
		{"StatusComplete", StatusComplete, harnesscp.StatusComplete},
		{"StatusFailed", StatusFailed, harnesscp.StatusFailed},
		{"StatusSkipped", StatusSkipped, harnesscp.StatusSkipped},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.local != tt.harness {
				t.Errorf("%s: local = %v, harness = %v", tt.name, tt.local, tt.harness)
			}
		})
	}
}

// TestNewManager verifies NewManager function alias works
func TestNewManager(t *testing.T) {
	tempDir := t.TempDir()

	// Verify function exists and is callable
	manager, err := NewManager(tempDir)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	if manager == nil {
		t.Fatal("NewManager returned nil")
	}

	// Verify it creates a harness.Manager
	var _ *harnesscp.Manager = manager

	t.Log("NewManager alias verified successfully")
}

// TestPackageBackwardCompatibility verifies the package maintains backward compatibility
func TestPackageBackwardCompatibility(t *testing.T) {
	// Types
	var _ Step
	var _ Status
	var _ Checkpoint
	var _ StepRecord
	var _ *Manager

	// Constants - Steps
	_ = StepFetchTicket
	_ = StepCreateWorktree
	_ = StepPlanning
	_ = StepValidation
	_ = StepExecution
	_ = StepTesting
	_ = StepReview
	_ = StepRefactor
	_ = StepVerify
	_ = StepCommit
	_ = StepPush
	_ = StepCreatePR
	_ = StepComplete

	// Constants - Status
	_ = StatusPending
	_ = StatusInProgress
	_ = StatusComplete
	_ = StatusFailed
	_ = StatusSkipped

	// Functions
	_ = NewManager

	t.Log("All exports are available for backward compatibility")
}
