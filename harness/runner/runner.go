package runner

import (
	"context"
	"fmt"
	"time"

	"github.com/philjestin/boatman-ecosystem/harness/checkpoint"
	"github.com/philjestin/boatman-ecosystem/harness/cost"
	"github.com/philjestin/boatman-ecosystem/harness/issuetracker"
	"github.com/philjestin/boatman-ecosystem/harness/review"
)

// Runner orchestrates the execute-test-review-refactor loop.
type Runner struct {
	developer    Developer
	reviewer     review.Reviewer
	tester       Tester
	planner      Planner
	config       Config
	costTracker  *cost.Tracker
	issueHistory *issuetracker.IssueHistory
	checkpoint   *checkpoint.Manager
	hooks        Hooks
}

// Option configures a Runner.
type Option func(*Runner)

// WithTester adds a Tester to the pipeline.
func WithTester(t Tester) Option {
	return func(r *Runner) { r.tester = t }
}

// WithPlanner adds a Planner to the pipeline.
func WithPlanner(p Planner) Option {
	return func(r *Runner) { r.planner = p }
}

// WithConfig overrides the default configuration.
func WithConfig(c Config) Option {
	return func(r *Runner) { r.config = c }
}

// WithMaxIterations sets the maximum review/refactor cycles.
func WithMaxIterations(n int) Option {
	return func(r *Runner) { r.config.MaxIterations = n }
}

// WithCostTracker attaches a cost tracker to the runner.
func WithCostTracker(t *cost.Tracker) Option {
	return func(r *Runner) { r.costTracker = t }
}

// WithCheckpointManager attaches a checkpoint manager.
func WithCheckpointManager(m *checkpoint.Manager) Option {
	return func(r *Runner) { r.checkpoint = m }
}

// WithHooks attaches lifecycle hooks.
func WithHooks(h Hooks) Option {
	return func(r *Runner) { r.hooks = h }
}

// New creates a Runner with the two required roles and optional configuration.
func New(dev Developer, rev review.Reviewer, opts ...Option) *Runner {
	r := &Runner{
		developer:    dev,
		reviewer:     rev,
		config:       DefaultConfig(),
		issueHistory: issuetracker.NewIssueHistory(),
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// Run executes the full pipeline: plan → execute → (test → review → refactor)*.
func (r *Runner) Run(ctx context.Context, req *Request) (*Result, error) {
	start := time.Now()
	result := &Result{
		CostTracker: r.costTracker,
	}

	// --- 1. Plan (optional) ---
	if r.planner != nil {
		plan, stepRec, err := r.runStep(ctx, "plan", func() (any, error) {
			return r.planner.Plan(ctx, req)
		})
		result.Steps = append(result.Steps, stepRec)
		r.hooks.callOnPlanComplete(ctx, asPlan(plan), err)

		if err != nil {
			if !r.config.SkipPlanningOnError {
				result.Status = StatusError
				result.Error = fmt.Errorf("planning failed: %w", err)
				result.Duration = time.Since(start)
				return result, nil
			}
			// skip planning, continue without a plan
		} else {
			result.Plan = asPlan(plan)
		}

		r.checkpointStep(checkpoint.StepPlanning)
	}

	// --- 2. Execute ---
	execOut, stepRec, err := r.runStep(ctx, "execute", func() (any, error) {
		return r.developer.Execute(ctx, req, result.Plan)
	})
	result.Steps = append(result.Steps, stepRec)
	r.hooks.callOnExecuteComplete(ctx, asExecuteResult(execOut), err)

	if err != nil {
		result.Status = StatusExecuteFailed
		result.Error = fmt.Errorf("execute failed: %w", err)
		result.Duration = time.Since(start)
		return result, nil
	}

	currentExec := asExecuteResult(execOut)
	currentFiles := currentExec.FilesChanged
	currentDiff := currentExec.Diff

	r.checkpointStep(checkpoint.StepExecution)

	// --- 3. Review Loop ---
	passed := false
	for i := 1; i <= r.config.MaxIterations; i++ {
		result.Iterations = i

		if err := ctx.Err(); err != nil {
			result.Status = StatusCanceled
			result.Error = err
			result.Duration = time.Since(start)
			return result, nil
		}

		// 3a. Test (optional)
		if r.tester != nil && r.config.TestBeforeReview {
			testOut, tStepRec, tErr := r.runStep(ctx, fmt.Sprintf("test_%d", i), func() (any, error) {
				return r.tester.Test(ctx, req, currentFiles)
			})
			result.Steps = append(result.Steps, tStepRec)

			tr := asTestResult(testOut)
			r.hooks.callOnTestComplete(ctx, tr, i)

			if tErr == nil && tr != nil {
				result.TestResult = tr
				if r.config.FailOnTestFailure && !tr.Passed {
					// Synthesize a review issue for the test failure
					testIssue := review.Issue{
						Severity:    "critical",
						Description: fmt.Sprintf("Tests failed: %s", formatFailedTests(tr.FailedTests)),
						Suggestion:  "Fix the failing tests before proceeding.",
					}
					// Run review with test failure context
					currentDiff = augmentDiffWithTestFailure(currentDiff, tr)
					_ = testIssue // used below via review
				}
			}

			r.checkpointStep(checkpoint.StepTesting)
		}

		// 3b. Review
		revOut, rStepRec, rErr := r.runStep(ctx, fmt.Sprintf("review_%d", i), func() (any, error) {
			return r.reviewer.Review(ctx, currentDiff, req.Description)
		})
		result.Steps = append(result.Steps, rStepRec)

		rr := asReviewResult(revOut)
		r.hooks.callOnReviewComplete(ctx, rr, i)

		if rErr != nil {
			result.Status = StatusError
			result.Error = fmt.Errorf("review failed on iteration %d: %w", i, rErr)
			result.Duration = time.Since(start)
			return result, nil
		}

		result.ReviewResult = rr

		// Feed issues to issueHistory
		if rr != nil && len(rr.Issues) > 0 {
			r.issueHistory.RecordIteration(rr.Issues)
		}

		r.checkpointStep(checkpoint.StepReview)

		// 3c. Check pass
		if rr != nil && rr.Passed {
			// Also check test result if we have one
			if result.TestResult == nil || result.TestResult.Passed || !r.config.FailOnTestFailure {
				passed = true
				result.Status = StatusPassed
				r.hooks.callOnIterationComplete(ctx, i, true)
				break
			}
		}

		r.hooks.callOnIterationComplete(ctx, i, false)

		// 3d. Refactor (if not last iteration)
		if i < r.config.MaxIterations {
			var issues []review.Issue
			var guidance string
			if rr != nil {
				issues = rr.Issues
				guidance = rr.Guidance
			}

			refOut, refStepRec, refErr := r.runStep(ctx, fmt.Sprintf("refactor_%d", i), func() (any, error) {
				return r.developer.Refactor(ctx, req, issues, guidance, currentExec)
			})
			result.Steps = append(result.Steps, refStepRec)

			rref := asRefactorResult(refOut)
			r.hooks.callOnRefactorComplete(ctx, rref, i)

			if refErr != nil {
				result.Status = StatusError
				result.Error = fmt.Errorf("refactor failed on iteration %d: %w", i, refErr)
				result.Duration = time.Since(start)
				return result, nil
			}

			if rref != nil {
				currentFiles = rref.FilesChanged
				currentDiff = rref.Diff
				// Update currentExec to reflect latest state for next refactor
				currentExec = &ExecuteResult{
					FilesChanged: rref.FilesChanged,
					Diff:         rref.Diff,
					Summary:      rref.Summary,
				}
			}

			r.checkpointStep(checkpoint.StepRefactor)
		}
	}

	// --- 4. Finalize ---
	if !passed {
		result.Status = StatusMaxIterations
	}

	result.FinalDiff = currentDiff
	result.FilesChanged = currentFiles
	result.Duration = time.Since(start)

	stats := r.issueHistory.GetTracker().Stats()
	result.IssueStats = &stats

	return result, nil
}

// runStep executes fn, records timing, and fires step hooks.
func (r *Runner) runStep(_ context.Context, name string, fn func() (any, error)) (any, StepRecord, error) {
	r.hooks.callOnStepStart(name)
	start := time.Now()

	val, err := fn()

	dur := time.Since(start)
	r.hooks.callOnStepEnd(name, dur, err)

	return val, StepRecord{Name: name, Duration: dur, Error: err}, err
}

// checkpointStep records a completed step if a checkpoint manager is configured.
func (r *Runner) checkpointStep(step checkpoint.Step) {
	if r.checkpoint != nil {
		r.checkpoint.CompleteStep(step, nil)
	}
}

// --- type assertion helpers ---

func asPlan(v any) *Plan {
	if v == nil {
		return nil
	}
	p, _ := v.(*Plan)
	return p
}

func asExecuteResult(v any) *ExecuteResult {
	if v == nil {
		return nil
	}
	e, _ := v.(*ExecuteResult)
	return e
}

func asRefactorResult(v any) *RefactorResult {
	if v == nil {
		return nil
	}
	r, _ := v.(*RefactorResult)
	return r
}

func asTestResult(v any) *TestResult {
	if v == nil {
		return nil
	}
	t, _ := v.(*TestResult)
	return t
}

func asReviewResult(v any) *review.ReviewResult {
	if v == nil {
		return nil
	}
	r, _ := v.(*review.ReviewResult)
	return r
}

// formatFailedTests returns a summary string of failed test names.
func formatFailedTests(names []string) string {
	if len(names) == 0 {
		return "one or more tests failed"
	}
	if len(names) == 1 {
		return names[0]
	}
	return fmt.Sprintf("%s and %d more", names[0], len(names)-1)
}

// augmentDiffWithTestFailure appends test failure context to the diff
// so the reviewer can see what went wrong.
func augmentDiffWithTestFailure(diff string, tr *TestResult) string {
	if tr == nil || tr.Passed {
		return diff
	}
	return diff + "\n\n# Test Failures\n" + tr.Output
}
