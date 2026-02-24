package runner_test

import (
	"context"
	"testing"
	"time"

	"github.com/philjestin/boatman-ecosystem/harness/review"
	"github.com/philjestin/boatman-ecosystem/harness/runner"
)

// mockObserver records calls for verification.
type mockObserver struct {
	runStartCalls    int
	runCompleteCalls int
	stepStarts       []string
	stepCompletes    []string
}

func (m *mockObserver) OnRunStart(_ context.Context, _ *runner.Request)   { m.runStartCalls++ }
func (m *mockObserver) OnRunComplete(_ context.Context, _ *runner.Result) { m.runCompleteCalls++ }
func (m *mockObserver) OnStepStart(_ context.Context, step string)        { m.stepStarts = append(m.stepStarts, step) }
func (m *mockObserver) OnStepComplete(_ context.Context, step string, _ time.Duration, _ error) {
	m.stepCompletes = append(m.stepCompletes, step)
}

func TestNopObserverSatisfiesInterface(t *testing.T) {
	var _ runner.Observer = runner.NopObserver{}
}

func TestObserverReceivesCalls(t *testing.T) {
	obs := &mockObserver{}
	dev := &passingDeveloper{}
	rev := &passingReviewer{}

	r := runner.New(dev, rev,
		runner.WithObserver(obs),
		runner.WithMaxIterations(1),
	)

	result, err := r.Run(context.Background(), &runner.Request{Description: "test"})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if obs.runStartCalls != 1 {
		t.Errorf("expected 1 OnRunStart call, got %d", obs.runStartCalls)
	}
	if obs.runCompleteCalls != 1 {
		t.Errorf("expected 1 OnRunComplete call, got %d", obs.runCompleteCalls)
	}
	if len(obs.stepStarts) == 0 {
		t.Error("expected at least one OnStepStart call")
	}
	if len(obs.stepCompletes) != len(obs.stepStarts) {
		t.Errorf("stepStarts (%d) != stepCompletes (%d)", len(obs.stepStarts), len(obs.stepCompletes))
	}

	_ = result
}

func TestRunnerWithoutObserver(t *testing.T) {
	dev := &passingDeveloper{}
	rev := &passingReviewer{}

	r := runner.New(dev, rev, runner.WithMaxIterations(1))

	_, err := r.Run(context.Background(), &runner.Request{Description: "test"})
	if err != nil {
		t.Fatalf("Run without observer should succeed: %v", err)
	}
}

// --- test helpers ---

type passingDeveloper struct{}

func (d *passingDeveloper) Execute(_ context.Context, _ *runner.Request, _ *runner.Plan) (*runner.ExecuteResult, error) {
	return &runner.ExecuteResult{
		FilesChanged: []string{"main.go"},
		Diff:         "diff",
		Summary:      "done",
	}, nil
}

func (d *passingDeveloper) Refactor(_ context.Context, _ *runner.Request, _ []review.Issue, _ string, _ *runner.ExecuteResult) (*runner.RefactorResult, error) {
	return &runner.RefactorResult{
		FilesChanged: []string{"main.go"},
		Diff:         "refactored diff",
		Summary:      "refactored",
	}, nil
}

type passingReviewer struct{}

func (r *passingReviewer) Review(_ context.Context, _ string, _ string) (*review.ReviewResult, error) {
	return &review.ReviewResult{
		Passed:   true,
		Score:    90,
		Summary:  "looks good",
		Guidance: "",
	}, nil
}
