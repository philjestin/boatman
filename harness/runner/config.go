package runner

import (
	"time"

	"github.com/philjestin/boatman-ecosystem/harness/cost"
	"github.com/philjestin/boatman-ecosystem/harness/issuetracker"
	"github.com/philjestin/boatman-ecosystem/harness/review"
)

// Config controls Runner behavior.
type Config struct {
	MaxIterations       int    // Max review/refactor cycles. Default: 3
	TestBeforeReview    bool   // Run tests before each review. Default: true
	FailOnTestFailure   bool   // Treat test failure as review issue. Default: true
	SkipPlanningOnError bool   // Continue without plan if Planner errors. Default: true
	CheckpointDir       string // Empty = no checkpointing
	ResumeFrom          string // Checkpoint ID to resume from. Empty = fresh start
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		MaxIterations:       3,
		TestBeforeReview:    true,
		FailOnTestFailure:   true,
		SkipPlanningOnError: true,
	}
}

// Status represents the outcome of a Run.
type Status int

const (
	StatusPassed        Status = iota // Review passed
	StatusMaxIterations               // Hit max iterations
	StatusExecuteFailed               // Execute step failed
	StatusCanceled                    // Context canceled
	StatusError                       // Unexpected error
)

// String returns a human-readable status name.
func (s Status) String() string {
	switch s {
	case StatusPassed:
		return "passed"
	case StatusMaxIterations:
		return "max_iterations"
	case StatusExecuteFailed:
		return "execute_failed"
	case StatusCanceled:
		return "canceled"
	case StatusError:
		return "error"
	default:
		return "unknown"
	}
}

// Result is the final output of a Run.
type Result struct {
	Status       Status
	Iterations   int
	Plan         *Plan
	FinalDiff    string
	FilesChanged []string
	ReviewResult *review.ReviewResult
	TestResult   *TestResult
	CostTracker  *cost.Tracker
	IssueStats   *issuetracker.IssueStats
	Duration     time.Duration
	Error        error
	Steps        []StepRecord
}

// StepRecord records timing and outcome for a single step.
type StepRecord struct {
	Name     string
	Duration time.Duration
	Error    error
}
