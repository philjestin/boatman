// Package checkpoint re-exports the harness checkpoint package for backward compatibility.
package checkpoint

import harnesscp "github.com/philjestin/boatman-ecosystem/harness/checkpoint"

// Type aliases for backward compatibility
type Step = harnesscp.Step
type Status = harnesscp.Status
type Checkpoint = harnesscp.Checkpoint
type StepRecord = harnesscp.StepRecord
type Manager = harnesscp.Manager
type Progress = harnesscp.Progress

// Step constants
const (
	StepFetchTicket    = harnesscp.StepFetchTicket
	StepCreateWorktree = harnesscp.StepCreateWorktree
	StepPlanning       = harnesscp.StepPlanning
	StepValidation     = harnesscp.StepValidation
	StepExecution      = harnesscp.StepExecution
	StepTesting        = harnesscp.StepTesting
	StepReview         = harnesscp.StepReview
	StepRefactor       = harnesscp.StepRefactor
	StepVerify         = harnesscp.StepVerify
	StepCommit         = harnesscp.StepCommit
	StepPush           = harnesscp.StepPush
	StepCreatePR       = harnesscp.StepCreatePR
	StepComplete       = harnesscp.StepComplete
)

// Status constants
const (
	StatusPending    = harnesscp.StatusPending
	StatusInProgress = harnesscp.StatusInProgress
	StatusComplete   = harnesscp.StatusComplete
	StatusFailed     = harnesscp.StatusFailed
	StatusSkipped    = harnesscp.StatusSkipped
)

// NewManager creates a new checkpoint manager.
var NewManager = harnesscp.NewManager
