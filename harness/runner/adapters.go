package runner

import (
	"context"

	"github.com/philjestin/boatman-ecosystem/harness/testrunner"
)

// TestRunnerTester adapts harness/testrunner.Runner to the Tester interface.
type TestRunnerTester struct {
	runner *testrunner.Runner
}

// NewTestRunnerTester creates a Tester backed by the testrunner package.
func NewTestRunnerTester(workDir string) *TestRunnerTester {
	return &TestRunnerTester{
		runner: testrunner.New(workDir),
	}
}

// Test implements the Tester interface by delegating to testrunner.Runner.
func (t *TestRunnerTester) Test(ctx context.Context, req *Request, changedFiles []string) (*TestResult, error) {
	var tr *testrunner.TestResult
	var err error

	if len(changedFiles) > 0 {
		tr, err = t.runner.RunForFiles(ctx, changedFiles)
	} else {
		tr, err = t.runner.RunAll(ctx)
	}
	if err != nil {
		return nil, err
	}

	return &TestResult{
		Passed:      tr.Passed,
		Output:      tr.Output,
		FailedTests: tr.FailedNames,
		Coverage:    tr.Coverage,
	}, nil
}
