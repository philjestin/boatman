package team

import (
	"context"
	"testing"

	"github.com/philjestin/boatman-ecosystem/harness/review"
	"github.com/philjestin/boatman-ecosystem/harness/runner"
)

func TestDeveloperTeam_Execute(t *testing.T) {
	handler := HandlerFunc(func(_ context.Context, task *Task) (*Result, error) {
		if _, ok := task.Input["plan_summary"]; !ok {
			t.Error("expected plan_summary in input")
		}
		return &Result{
			Output:       "implemented feature",
			FilesChanged: []string{"main.go"},
			Diff:         "+// new code",
		}, nil
	})

	tm := New("dev", WithAgents(NewAgent("dev1", "developer", handler)))
	dt := NewDeveloperTeam(tm)

	result, err := dt.Execute(context.Background(),
		&runner.Request{ID: "req-1", Title: "Add feature", Description: "Add a new feature"},
		&runner.Plan{Summary: "plan", Steps: []string{"step1"}},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Summary != "implemented feature" {
		t.Fatalf("expected summary, got %q", result.Summary)
	}
	if len(result.FilesChanged) != 1 {
		t.Fatalf("expected 1 file changed, got %d", len(result.FilesChanged))
	}
}

func TestDeveloperTeam_Refactor(t *testing.T) {
	handler := HandlerFunc(func(_ context.Context, task *Task) (*Result, error) {
		if _, ok := task.Input["issues"]; !ok {
			t.Error("expected issues in input")
		}
		return &Result{
			Output:       "refactored",
			FilesChanged: []string{"main.go"},
			Diff:         "+// refactored",
		}, nil
	})

	tm := New("dev", WithAgents(NewAgent("dev1", "developer", handler)))
	dt := NewDeveloperTeam(tm)

	result, err := dt.Refactor(context.Background(),
		&runner.Request{ID: "req-1", Title: "Fix issues"},
		[]review.Issue{{Severity: "major", Description: "bug"}},
		"fix this",
		&runner.ExecuteResult{FilesChanged: []string{"main.go"}, Diff: "old diff"},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Summary != "refactored" {
		t.Fatalf("expected summary, got %q", result.Summary)
	}
}

func TestPlannerTeam_Plan(t *testing.T) {
	handler := HandlerFunc(func(_ context.Context, _ *Task) (*Result, error) {
		return &Result{
			Output: "analysis complete",
			Data: map[string]any{
				"steps":          []string{"step1", "step2"},
				"relevant_files": []string{"main.go"},
			},
		}, nil
	})

	tm := New("planner", WithAgents(NewAgent("p1", "planner", handler)))
	pt := NewPlannerTeam(tm)

	plan, err := pt.Plan(context.Background(),
		&runner.Request{ID: "req-1", Title: "Plan feature"},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if plan.Summary != "analysis complete" {
		t.Fatalf("expected summary, got %q", plan.Summary)
	}
	if len(plan.Steps) != 2 {
		t.Fatalf("expected 2 steps, got %d", len(plan.Steps))
	}
	if len(plan.RelevantFiles) != 1 {
		t.Fatalf("expected 1 relevant file, got %d", len(plan.RelevantFiles))
	}
}

func TestTesterTeam_Test(t *testing.T) {
	handler := HandlerFunc(func(_ context.Context, task *Task) (*Result, error) {
		if _, ok := task.Input["changed_files"]; !ok {
			t.Error("expected changed_files in input")
		}
		return &Result{
			Output: "all tests pass",
			Data: map[string]any{
				"passed":   true,
				"coverage": 85.5,
			},
		}, nil
	})

	tm := New("tester", WithAgents(NewAgent("t1", "tester", handler)))
	tt := NewTesterTeam(tm)

	result, err := tt.Test(context.Background(),
		&runner.Request{ID: "req-1", Title: "Test changes"},
		[]string{"main.go"},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Passed {
		t.Fatal("expected tests to pass")
	}
	if result.Coverage != 85.5 {
		t.Fatalf("expected 85.5 coverage, got %f", result.Coverage)
	}
}

func TestReviewerTeam_Review(t *testing.T) {
	handler := HandlerFunc(func(_ context.Context, task *Task) (*Result, error) {
		if _, ok := task.Input["diff"]; !ok {
			t.Error("expected diff in input")
		}
		return &Result{
			Output: "looks good",
			Data: map[string]any{
				"passed":   true,
				"score":    8,
				"issues":   []review.Issue{{Severity: "minor", Description: "nit"}},
				"guidance": "overall solid",
			},
		}, nil
	})

	tm := New("reviewer", WithAgents(NewAgent("r1", "reviewer", handler)))
	rt := NewReviewerTeam(tm)

	result, err := rt.Review(context.Background(), "+// new code", "feature context")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Passed {
		t.Fatal("expected review to pass")
	}
	if result.Score != 8 {
		t.Fatalf("expected score 8, got %d", result.Score)
	}
	if len(result.Issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(result.Issues))
	}
}
