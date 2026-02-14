// Package types defines common types shared between CLI and desktop.
package types

import "time"

// ExecutionResult represents the result of executing a task.
type ExecutionResult struct {
	Success      bool          `json:"success"`
	Error        string        `json:"error,omitempty"`
	FilesChanged []string      `json:"files_changed"`
	Diff         string        `json:"diff"`
	Duration     time.Duration `json:"duration"`
}

// PlanResult represents the result of planning analysis.
type PlanResult struct {
	Summary      string   `json:"summary"`
	Files        []string `json:"files"`         // Files to be modified
	Dependencies []string `json:"dependencies"`  // Dependencies needed
	Risks        []string `json:"risks"`         // Potential risks
	Estimate     string   `json:"estimate"`      // Time estimate
}

// ReviewResult represents code review results.
type ReviewResult struct {
	Passed  bool          `json:"passed"`
	Summary string        `json:"summary"`
	Issues  []ReviewIssue `json:"issues"`
	Score   int           `json:"score"` // 0-100
}

// ReviewIssue represents a single code review issue.
type ReviewIssue struct {
	File     string `json:"file"`
	Line     int    `json:"line,omitempty"`
	Severity string `json:"severity"` // "error", "warning", "info"
	Message  string `json:"message"`
	Code     string `json:"code,omitempty"`
}

// TestResult represents test execution results.
type TestResult struct {
	Passed       bool     `json:"passed"`
	TestsRun     int      `json:"tests_run"`
	TestsPassed  int      `json:"tests_passed"`
	TestsFailed  int      `json:"tests_failed"`
	FailedTests  []string `json:"failed_tests,omitempty"`
	Output       string   `json:"output,omitempty"`
	Duration     float64  `json:"duration"` // seconds
	CoverageText string   `json:"coverage_text,omitempty"`
}

// Usage represents token usage and cost information.
type Usage struct {
	InputTokens      int     `json:"input_tokens"`
	OutputTokens     int     `json:"output_tokens"`
	CacheReadTokens  int     `json:"cache_read_tokens"`
	CacheWriteTokens int     `json:"cache_write_tokens"`
	TotalCostUSD     float64 `json:"total_cost_usd"`
}

// IsEmpty returns true if usage has no data.
func (u *Usage) IsEmpty() bool {
	return u.InputTokens == 0 && u.OutputTokens == 0
}

// TaskInfo represents a task to be executed.
type TaskInfo struct {
	ID          string            `json:"id"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Labels      []string          `json:"labels,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// GetID returns the task ID.
func (t *TaskInfo) GetID() string {
	return t.ID
}

// GetTitle returns the task title.
func (t *TaskInfo) GetTitle() string {
	return t.Title
}

// GetDescription returns the task description.
func (t *TaskInfo) GetDescription() string {
	return t.Description
}
