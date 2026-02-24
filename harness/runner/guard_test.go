package runner_test

import (
	"context"
	"errors"
	"testing"

	"github.com/philjestin/boatman-ecosystem/harness/runner"
)

func TestNopGuardSatisfiesInterface(t *testing.T) {
	var _ runner.Guard = runner.NopGuard{}
}

func TestNopGuardAlwaysAllows(t *testing.T) {
	g := runner.NopGuard{}
	err := g.AllowStep(context.Background(), "execute", &runner.GuardState{})
	if err != nil {
		t.Errorf("NopGuard should always allow: %v", err)
	}
}

// denyGuard denies all steps.
type denyGuard struct {
	deniedSteps []string
}

func (g *denyGuard) AllowStep(_ context.Context, step string, _ *runner.GuardState) error {
	g.deniedSteps = append(g.deniedSteps, step)
	return errors.New("policy denied: " + step)
}

func TestGuardDeniesExecution(t *testing.T) {
	guard := &denyGuard{}
	dev := &passingDeveloper{}
	rev := &passingReviewer{}

	r := runner.New(dev, rev,
		runner.WithGuard(guard),
		runner.WithMaxIterations(1),
	)

	result, err := r.Run(context.Background(), &runner.Request{Description: "test"})
	if err != nil {
		t.Fatalf("Run should not return a Go error: %v", err)
	}

	// The guard should have denied the execute step, causing the run to fail
	if result.Error == nil {
		t.Fatal("expected result.Error to be non-nil when guard denies")
	}

	if len(guard.deniedSteps) == 0 {
		t.Error("expected guard to have been consulted")
	}
}

// costGuard denies when cost exceeds threshold.
type costGuard struct {
	maxCost float64
}

func (g *costGuard) AllowStep(_ context.Context, _ string, state *runner.GuardState) error {
	if state.TotalCostUSD > g.maxCost {
		return errors.New("cost budget exceeded")
	}
	return nil
}

func TestGuardReceivesState(t *testing.T) {
	guard := &costGuard{maxCost: 100.0} // high limit, should allow
	dev := &passingDeveloper{}
	rev := &passingReviewer{}

	r := runner.New(dev, rev,
		runner.WithGuard(guard),
		runner.WithMaxIterations(1),
	)

	result, err := r.Run(context.Background(), &runner.Request{Description: "test"})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.Error != nil {
		t.Fatalf("expected no error with high cost limit: %v", result.Error)
	}
}

func TestRunnerWithoutGuard(t *testing.T) {
	dev := &passingDeveloper{}
	rev := &passingReviewer{}

	r := runner.New(dev, rev, runner.WithMaxIterations(1))

	_, err := r.Run(context.Background(), &runner.Request{Description: "test"})
	if err != nil {
		t.Fatalf("Run without guard should succeed: %v", err)
	}
}
