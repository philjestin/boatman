// Package testrunner provides an adapter wrapping the harness testrunner
// with coordinator integration.
package testrunner

import (
	"context"

	"github.com/philjestin/boatman-ecosystem/harness/testrunner"
	"github.com/philjestin/boatmanmode/internal/coordinator"
)

// Type aliases from harness
type TestResult = testrunner.TestResult
type Framework = testrunner.Framework
type TestResultHandoff = testrunner.TestResultHandoff
type FilesHandoff = testrunner.FilesHandoff

// Agent wraps the harness Runner with coordinator integration.
type Agent struct {
	inner *testrunner.Runner
	coord *coordinator.Coordinator
}

// New creates a new Agent.
func New(worktreePath string) *Agent {
	return &Agent{
		inner: testrunner.New(worktreePath),
	}
}

// SetCoordinator sets the coordinator for work claiming.
func (a *Agent) SetCoordinator(c *coordinator.Coordinator) {
	a.coord = c
}

// DetectFramework detects the test framework.
func (a *Agent) DetectFramework() (*Framework, error) {
	return a.inner.DetectFramework()
}

// RunAll runs all tests.
func (a *Agent) RunAll(ctx context.Context) (*TestResult, error) {
	return a.inner.RunAll(ctx)
}

// RunForFiles runs tests for specific files.
func (a *Agent) RunForFiles(ctx context.Context, changedFiles []string) (*TestResult, error) {
	return a.inner.RunForFiles(ctx, changedFiles)
}

// FindRelatedTests finds tests related to changed files.
func (a *Agent) FindRelatedTests(changedFiles []string, framework *Framework) []string {
	return a.inner.FindRelatedTests(changedFiles, framework)
}

// IsTestFile checks if a file is a test file.
func (a *Agent) IsTestFile(file string, framework *Framework) bool {
	return a.inner.IsTestFile(file, framework)
}

// FindTestFile finds the test file for a source file.
func (a *Agent) FindTestFile(file string, framework *Framework) string {
	return a.inner.FindTestFile(file, framework)
}
