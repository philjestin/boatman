package runner

import (
	"context"

	"github.com/philjestin/boatman-ecosystem/harness/review"
)

// Request describes what the runner should accomplish.
type Request struct {
	ID          string
	Title       string
	Description string
	WorkDir     string
	Labels      []string
	Metadata    map[string]string
}

// Plan is what a Planner produces.
type Plan struct {
	Summary       string
	Steps         []string
	RelevantFiles []string
}

// ExecuteResult is what a Developer returns after making changes.
type ExecuteResult struct {
	FilesChanged []string
	Diff         string
	Summary      string
}

// RefactorResult is what a Developer returns after refactoring.
type RefactorResult struct {
	FilesChanged []string
	Diff         string
	Summary      string
}

// TestResult wraps the outcome of running tests.
type TestResult struct {
	Passed      bool
	Output      string
	FailedTests []string
	Coverage    float64
}

// Developer implements code changes. Execute is called once,
// then Refactor zero or more times based on review feedback.
type Developer interface {
	Execute(ctx context.Context, req *Request, plan *Plan) (*ExecuteResult, error)
	Refactor(ctx context.Context, req *Request, issues []review.Issue,
		guidance string, prevResult *ExecuteResult) (*RefactorResult, error)
}

// Tester runs tests and reports results. Optional.
type Tester interface {
	Test(ctx context.Context, req *Request, changedFiles []string) (*TestResult, error)
}

// Planner analyzes the request and produces a plan. Optional.
type Planner interface {
	Plan(ctx context.Context, req *Request) (*Plan, error)
}
