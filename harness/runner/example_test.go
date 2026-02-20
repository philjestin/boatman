package runner_test

import (
	"context"
	"fmt"

	"github.com/philjestin/boatman-ecosystem/harness/review"
	"github.com/philjestin/boatman-ecosystem/harness/runner"
)

// simpleDev is a minimal Developer for examples.
type simpleDev struct{}

func (d *simpleDev) Execute(_ context.Context, req *runner.Request, _ *runner.Plan) (*runner.ExecuteResult, error) {
	return &runner.ExecuteResult{
		FilesChanged: []string{"main.go"},
		Diff:         fmt.Sprintf("diff for %s", req.Title),
		Summary:      "implemented " + req.Title,
	}, nil
}

func (d *simpleDev) Refactor(_ context.Context, _ *runner.Request, _ []review.Issue,
	_ string, _ *runner.ExecuteResult) (*runner.RefactorResult, error) {
	return &runner.RefactorResult{
		FilesChanged: []string{"main.go"},
		Diff:         "refactored diff",
		Summary:      "addressed issues",
	}, nil
}

// simpleReviewer is a Reviewer that always passes.
type simpleReviewer struct{}

func (r *simpleReviewer) Review(_ context.Context, _ string, _ string) (*review.ReviewResult, error) {
	return &review.ReviewResult{
		Passed:  true,
		Score:   9,
		Summary: "looks good",
	}, nil
}

func Example_minimal() {
	r := runner.New(&simpleDev{}, &simpleReviewer{})

	result, err := r.Run(context.Background(), &runner.Request{
		ID:          "task-1",
		Title:       "Add auth",
		Description: "Implement JWT auth for the API",
		WorkDir:     "/tmp/repo",
	})
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	fmt.Printf("Status: %v, Iterations: %d\n", result.Status, result.Iterations)
	// Output: Status: passed, Iterations: 1
}

func Example_withHooks() {
	r := runner.New(&simpleDev{}, &simpleReviewer{},
		runner.WithMaxIterations(5),
		runner.WithHooks(runner.Hooks{
			OnStepStart: func(name string) {
				fmt.Printf("step: %s\n", name)
			},
		}),
	)

	result, err := r.Run(context.Background(), &runner.Request{
		ID:          "task-2",
		Title:       "Fix bug",
		Description: "Fix the login bug",
		WorkDir:     "/tmp/repo",
	})
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	fmt.Printf("Status: %v\n", result.Status)
	// Output:
	// step: execute
	// step: review_1
	// Status: passed
}
