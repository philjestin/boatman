package runner

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/philjestin/boatman-ecosystem/harness/review"
)

// --- mock implementations ---

type mockDeveloper struct {
	executeFn  func(ctx context.Context, req *Request, plan *Plan) (*ExecuteResult, error)
	refactorFn func(ctx context.Context, req *Request, issues []review.Issue,
		guidance string, prev *ExecuteResult) (*RefactorResult, error)
}

func (m *mockDeveloper) Execute(ctx context.Context, req *Request, plan *Plan) (*ExecuteResult, error) {
	if m.executeFn != nil {
		return m.executeFn(ctx, req, plan)
	}
	return &ExecuteResult{
		FilesChanged: []string{"main.go"},
		Diff:         "diff --git a/main.go",
		Summary:      "implemented changes",
	}, nil
}

func (m *mockDeveloper) Refactor(ctx context.Context, req *Request, issues []review.Issue,
	guidance string, prev *ExecuteResult) (*RefactorResult, error) {
	if m.refactorFn != nil {
		return m.refactorFn(ctx, req, issues, guidance, prev)
	}
	return &RefactorResult{
		FilesChanged: []string{"main.go"},
		Diff:         "diff --git a/main.go (refactored)",
		Summary:      "addressed review issues",
	}, nil
}

type mockReviewer struct {
	results []*review.ReviewResult
	call    int
}

func (m *mockReviewer) Review(_ context.Context, _ string, _ string) (*review.ReviewResult, error) {
	if m.call < len(m.results) {
		r := m.results[m.call]
		m.call++
		return r, nil
	}
	return &review.ReviewResult{Passed: true}, nil
}

type mockPlanner struct {
	plan *Plan
	err  error
}

func (m *mockPlanner) Plan(_ context.Context, _ *Request) (*Plan, error) {
	return m.plan, m.err
}

type mockTester struct {
	result *TestResult
	err    error
}

func (m *mockTester) Test(_ context.Context, _ *Request, _ []string) (*TestResult, error) {
	return m.result, m.err
}

// --- helpers ---

func simpleRequest() *Request {
	return &Request{
		ID:          "test-1",
		Title:       "Test task",
		Description: "A test task",
		WorkDir:     "/tmp/test",
	}
}

// --- tests ---

func TestExecuteOnly(t *testing.T) {
	dev := &mockDeveloper{}
	rev := &mockReviewer{
		results: []*review.ReviewResult{
			{Passed: true, Score: 9, Summary: "looks good"},
		},
	}

	r := New(dev, rev)
	result, err := r.Run(context.Background(), simpleRequest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Status != StatusPassed {
		t.Errorf("expected StatusPassed, got %v", result.Status)
	}
	if result.Iterations != 1 {
		t.Errorf("expected 1 iteration, got %d", result.Iterations)
	}
	if result.FinalDiff == "" {
		t.Error("expected non-empty FinalDiff")
	}
}

func TestReviewPassFirstIteration(t *testing.T) {
	dev := &mockDeveloper{}
	rev := &mockReviewer{
		results: []*review.ReviewResult{
			{Passed: true, Score: 8, Summary: "approved"},
		},
	}

	r := New(dev, rev)
	result, err := r.Run(context.Background(), simpleRequest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Status != StatusPassed {
		t.Errorf("expected StatusPassed, got %v", result.Status)
	}
	if result.Iterations != 1 {
		t.Errorf("expected 1 iteration, got %d", result.Iterations)
	}
}

func TestReviewFailThenRefactorThenPass(t *testing.T) {
	dev := &mockDeveloper{}
	rev := &mockReviewer{
		results: []*review.ReviewResult{
			{
				Passed:  false,
				Score:   4,
				Summary: "needs work",
				Issues: []review.Issue{
					{Severity: "major", Description: "missing error handling"},
				},
				Guidance: "add error handling to all public functions",
			},
			{Passed: true, Score: 8, Summary: "approved after refactor"},
		},
	}

	r := New(dev, rev)
	result, err := r.Run(context.Background(), simpleRequest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Status != StatusPassed {
		t.Errorf("expected StatusPassed, got %v", result.Status)
	}
	if result.Iterations != 2 {
		t.Errorf("expected 2 iterations, got %d", result.Iterations)
	}
}

func TestMaxIterationsExhausted(t *testing.T) {
	dev := &mockDeveloper{}
	rev := &mockReviewer{
		results: []*review.ReviewResult{
			{Passed: false, Score: 3, Issues: []review.Issue{{Severity: "major", Description: "issue 1"}}},
			{Passed: false, Score: 4, Issues: []review.Issue{{Severity: "minor", Description: "issue 2"}}},
			{Passed: false, Score: 5, Issues: []review.Issue{{Severity: "minor", Description: "issue 3"}}},
		},
	}

	r := New(dev, rev, WithMaxIterations(3))
	result, err := r.Run(context.Background(), simpleRequest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Status != StatusMaxIterations {
		t.Errorf("expected StatusMaxIterations, got %v", result.Status)
	}
	if result.Iterations != 3 {
		t.Errorf("expected 3 iterations, got %d", result.Iterations)
	}
}

func TestPlannerErrorSkipped(t *testing.T) {
	dev := &mockDeveloper{}
	rev := &mockReviewer{
		results: []*review.ReviewResult{
			{Passed: true, Score: 8},
		},
	}
	planner := &mockPlanner{
		err: errors.New("planner unavailable"),
	}

	r := New(dev, rev,
		WithPlanner(planner),
		WithConfig(Config{
			MaxIterations:       3,
			TestBeforeReview:    true,
			FailOnTestFailure:   true,
			SkipPlanningOnError: true,
		}),
	)
	result, err := r.Run(context.Background(), simpleRequest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Status != StatusPassed {
		t.Errorf("expected StatusPassed (planning skipped), got %v", result.Status)
	}
	if result.Plan != nil {
		t.Error("expected nil plan when planner errors")
	}
}

func TestPlannerErrorFatal(t *testing.T) {
	dev := &mockDeveloper{}
	rev := &mockReviewer{}
	planner := &mockPlanner{
		err: errors.New("planner unavailable"),
	}

	r := New(dev, rev,
		WithPlanner(planner),
		WithConfig(Config{
			MaxIterations:       3,
			SkipPlanningOnError: false,
		}),
	)
	result, err := r.Run(context.Background(), simpleRequest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Status != StatusError {
		t.Errorf("expected StatusError, got %v", result.Status)
	}
}

func TestContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	rev := &mockReviewer{
		results: []*review.ReviewResult{
			{Passed: false, Score: 3, Issues: []review.Issue{{Severity: "major", Description: "issue"}}},
		},
	}

	// Cancel context after execute completes but before review loop check
	dev := &mockDeveloper{
		executeFn: func(_ context.Context, _ *Request, _ *Plan) (*ExecuteResult, error) {
			cancel() // cancel before review loop
			return &ExecuteResult{
				FilesChanged: []string{"main.go"},
				Diff:         "diff",
				Summary:      "done",
			}, nil
		},
	}

	r := New(dev, rev)
	result, err := r.Run(ctx, simpleRequest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Status != StatusCanceled {
		t.Errorf("expected StatusCanceled, got %v", result.Status)
	}
}

func TestExecuteFailure(t *testing.T) {
	dev := &mockDeveloper{
		executeFn: func(_ context.Context, _ *Request, _ *Plan) (*ExecuteResult, error) {
			return nil, errors.New("compilation error")
		},
	}
	rev := &mockReviewer{}

	r := New(dev, rev)
	result, err := r.Run(context.Background(), simpleRequest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Status != StatusExecuteFailed {
		t.Errorf("expected StatusExecuteFailed, got %v", result.Status)
	}
}

func TestWithTester(t *testing.T) {
	dev := &mockDeveloper{}
	rev := &mockReviewer{
		results: []*review.ReviewResult{
			{Passed: true, Score: 8},
		},
	}
	tester := &mockTester{
		result: &TestResult{Passed: true, Coverage: 85.0},
	}

	r := New(dev, rev, WithTester(tester))
	result, err := r.Run(context.Background(), simpleRequest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Status != StatusPassed {
		t.Errorf("expected StatusPassed, got %v", result.Status)
	}
	if result.TestResult == nil {
		t.Fatal("expected non-nil TestResult")
	}
	if result.TestResult.Coverage != 85.0 {
		t.Errorf("expected coverage 85.0, got %f", result.TestResult.Coverage)
	}
}

func TestWithPlanner(t *testing.T) {
	dev := &mockDeveloper{
		executeFn: func(_ context.Context, _ *Request, plan *Plan) (*ExecuteResult, error) {
			if plan == nil {
				return nil, errors.New("expected plan")
			}
			return &ExecuteResult{
				FilesChanged: plan.RelevantFiles,
				Diff:         "diff",
				Summary:      plan.Summary,
			}, nil
		},
	}
	rev := &mockReviewer{
		results: []*review.ReviewResult{
			{Passed: true, Score: 9},
		},
	}
	planner := &mockPlanner{
		plan: &Plan{
			Summary:       "implement auth",
			Steps:         []string{"add middleware", "add JWT validation"},
			RelevantFiles: []string{"auth.go", "middleware.go"},
		},
	}

	r := New(dev, rev, WithPlanner(planner))
	result, err := r.Run(context.Background(), simpleRequest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Status != StatusPassed {
		t.Errorf("expected StatusPassed, got %v", result.Status)
	}
	if result.Plan == nil {
		t.Fatal("expected non-nil Plan")
	}
	if result.Plan.Summary != "implement auth" {
		t.Errorf("expected plan summary 'implement auth', got %q", result.Plan.Summary)
	}
}

func TestHooksAreCalled(t *testing.T) {
	dev := &mockDeveloper{}
	rev := &mockReviewer{
		results: []*review.ReviewResult{
			{Passed: true, Score: 8},
		},
	}

	var stepsStarted []string
	var stepsEnded []string

	r := New(dev, rev, WithHooks(Hooks{
		OnStepStart: func(name string) {
			stepsStarted = append(stepsStarted, name)
		},
		OnStepEnd: func(name string, _ time.Duration, _ error) {
			stepsEnded = append(stepsEnded, name)
		},
	}))

	_, err := r.Run(context.Background(), simpleRequest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(stepsStarted) == 0 {
		t.Error("expected OnStepStart to be called")
	}
	if len(stepsEnded) == 0 {
		t.Error("expected OnStepEnd to be called")
	}
}

func TestStatusString(t *testing.T) {
	tests := []struct {
		status Status
		want   string
	}{
		{StatusPassed, "passed"},
		{StatusMaxIterations, "max_iterations"},
		{StatusExecuteFailed, "execute_failed"},
		{StatusCanceled, "canceled"},
		{StatusError, "error"},
		{Status(99), "unknown"},
	}

	for _, tt := range tests {
		if got := tt.status.String(); got != tt.want {
			t.Errorf("Status(%d).String() = %q, want %q", tt.status, got, tt.want)
		}
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.MaxIterations != 3 {
		t.Errorf("expected MaxIterations=3, got %d", cfg.MaxIterations)
	}
	if !cfg.TestBeforeReview {
		t.Error("expected TestBeforeReview=true")
	}
	if !cfg.FailOnTestFailure {
		t.Error("expected FailOnTestFailure=true")
	}
	if !cfg.SkipPlanningOnError {
		t.Error("expected SkipPlanningOnError=true")
	}
}
