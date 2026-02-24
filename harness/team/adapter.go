package team

import (
	"context"
	"fmt"
	"strings"

	"github.com/philjestin/boatman-ecosystem/harness/review"
	"github.com/philjestin/boatman-ecosystem/harness/runner"
)

// Compile-time interface checks.
var (
	_ runner.Developer = (*DeveloperTeam)(nil)
	_ runner.Planner   = (*PlannerTeam)(nil)
	_ runner.Tester    = (*TesterTeam)(nil)
	_ review.Reviewer  = (*ReviewerTeam)(nil)
)

// DeveloperTeam wraps a Team as a runner.Developer.
type DeveloperTeam struct {
	team *Team
}

// NewDeveloperTeam creates a DeveloperTeam from the given Team.
func NewDeveloperTeam(t *Team) *DeveloperTeam {
	return &DeveloperTeam{team: t}
}

// Execute delegates execution to the underlying team.
func (d *DeveloperTeam) Execute(ctx context.Context, req *runner.Request, plan *runner.Plan) (*runner.ExecuteResult, error) {
	task := requestToTask(req, "execute")
	task.Input["plan_summary"] = plan.Summary
	task.Input["plan_steps"] = plan.Steps
	task.Input["plan_files"] = plan.RelevantFiles

	result, err := d.team.Handle(ctx, task)
	if err != nil {
		return nil, err
	}

	return &runner.ExecuteResult{
		FilesChanged: result.FilesChanged,
		Diff:         result.Diff,
		Summary:      result.Output,
	}, nil
}

// Refactor delegates refactoring to the underlying team.
func (d *DeveloperTeam) Refactor(ctx context.Context, req *runner.Request, issues []review.Issue, guidance string, prev *runner.ExecuteResult) (*runner.RefactorResult, error) {
	task := requestToTask(req, "refactor")
	task.Input["issues"] = issues
	task.Input["guidance"] = guidance
	if prev != nil {
		task.Input["prev_files_changed"] = prev.FilesChanged
		task.Input["prev_diff"] = prev.Diff
	}

	result, err := d.team.Handle(ctx, task)
	if err != nil {
		return nil, err
	}

	return &runner.RefactorResult{
		FilesChanged: result.FilesChanged,
		Diff:         result.Diff,
		Summary:      result.Output,
	}, nil
}

// PlannerTeam wraps a Team as a runner.Planner.
type PlannerTeam struct {
	team *Team
}

// NewPlannerTeam creates a PlannerTeam from the given Team.
func NewPlannerTeam(t *Team) *PlannerTeam {
	return &PlannerTeam{team: t}
}

// Plan delegates planning to the underlying team.
func (p *PlannerTeam) Plan(ctx context.Context, req *runner.Request) (*runner.Plan, error) {
	task := requestToTask(req, "plan")

	result, err := p.team.Handle(ctx, task)
	if err != nil {
		return nil, err
	}

	plan := &runner.Plan{
		Summary: result.Output,
	}

	if steps, ok := result.Data["steps"].([]string); ok {
		plan.Steps = steps
	}
	if files, ok := result.Data["relevant_files"].([]string); ok {
		plan.RelevantFiles = files
	}

	return plan, nil
}

// TesterTeam wraps a Team as a runner.Tester.
type TesterTeam struct {
	team *Team
}

// NewTesterTeam creates a TesterTeam from the given Team.
func NewTesterTeam(t *Team) *TesterTeam {
	return &TesterTeam{team: t}
}

// Test delegates testing to the underlying team.
func (tt *TesterTeam) Test(ctx context.Context, req *runner.Request, changedFiles []string) (*runner.TestResult, error) {
	task := requestToTask(req, "test")
	task.Input["changed_files"] = changedFiles

	result, err := tt.team.Handle(ctx, task)
	if err != nil {
		return nil, err
	}

	tr := &runner.TestResult{
		Output: result.Output,
	}

	if passed, ok := result.Data["passed"].(bool); ok {
		tr.Passed = passed
	}
	if failed, ok := result.Data["failed_tests"].([]string); ok {
		tr.FailedTests = failed
	}
	if coverage, ok := result.Data["coverage"].(float64); ok {
		tr.Coverage = coverage
	}

	return tr, nil
}

// ReviewerTeam wraps a Team as a review.Reviewer.
type ReviewerTeam struct {
	team *Team
}

// NewReviewerTeam creates a ReviewerTeam from the given Team.
func NewReviewerTeam(t *Team) *ReviewerTeam {
	return &ReviewerTeam{team: t}
}

// Review delegates review to the underlying team.
func (r *ReviewerTeam) Review(ctx context.Context, diff string, reviewContext string) (*review.ReviewResult, error) {
	task := &Task{
		ID:          "review",
		Description: "review code changes",
		Context:     reviewContext,
		Input: map[string]any{
			"diff": diff,
		},
	}

	result, err := r.team.Handle(ctx, task)
	if err != nil {
		return nil, err
	}

	rr := &review.ReviewResult{
		Summary: result.Output,
	}

	if passed, ok := result.Data["passed"].(bool); ok {
		rr.Passed = passed
	}
	if score, ok := result.Data["score"].(int); ok {
		rr.Score = score
	}
	if issues, ok := result.Data["issues"].([]review.Issue); ok {
		rr.Issues = issues
	}
	if praise, ok := result.Data["praise"].([]string); ok {
		rr.Praise = praise
	}
	if guidance, ok := result.Data["guidance"].(string); ok {
		rr.Guidance = guidance
	}

	return rr, nil
}

// requestToTask converts a runner.Request into a team.Task.
func requestToTask(req *runner.Request, action string) *Task {
	var constraints []string
	for _, l := range req.Labels {
		constraints = append(constraints, l)
	}

	desc := req.Description
	if desc == "" {
		desc = req.Title
	}

	return &Task{
		ID:          fmt.Sprintf("%s-%s", req.ID, action),
		Description: desc,
		Context:     strings.Join([]string{req.Title, req.Description}, "\n"),
		Input: map[string]any{
			"work_dir": req.WorkDir,
			"metadata": req.Metadata,
		},
		Constraints: constraints,
	}
}
