package team_test

import (
	"context"
	"fmt"

	"github.com/philjestin/boatman-ecosystem/harness/cost"
	"github.com/philjestin/boatman-ecosystem/harness/team"
)

func ExampleTeam_sequential() {
	analyzer := team.NewAgent("analyzer", "analyzes codebase structure",
		team.HandlerFunc(func(_ context.Context, _ *team.Task) (*team.Result, error) {
			return &team.Result{Output: "found 3 packages"}, nil
		}),
	)
	riskAssessor := team.NewAgent("risk-assessor", "assesses risk of changes",
		team.HandlerFunc(func(_ context.Context, _ *team.Task) (*team.Result, error) {
			return &team.Result{Output: "low risk"}, nil
		}),
	)

	t := team.New("planning-team",
		team.WithAgents(analyzer, riskAssessor),
		team.WithStrategy(team.Sequential),
	)

	result, err := t.Handle(context.Background(), &team.Task{
		ID:          "plan-1",
		Description: "analyze and assess",
	})
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
	fmt.Println(result.Output)
	// Output: found 3 packages
	// low risk
}

func ExampleTeam_parallel() {
	frontend := team.NewAgent("frontend", "handles UI changes",
		team.HandlerFunc(func(_ context.Context, _ *team.Task) (*team.Result, error) {
			return &team.Result{
				Output:       "updated components",
				FilesChanged: []string{"app.tsx"},
			}, nil
		}),
	)
	backend := team.NewAgent("backend", "handles API changes",
		team.HandlerFunc(func(_ context.Context, _ *team.Task) (*team.Result, error) {
			return &team.Result{
				Output:       "updated handlers",
				FilesChanged: []string{"handler.go"},
			}, nil
		}),
	)

	t := team.New("dev-team",
		team.WithAgents(frontend, backend),
		team.WithStrategy(team.Parallel),
	)

	result, err := t.Handle(context.Background(), &team.Task{
		ID:          "dev-1",
		Description: "implement feature",
	})
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
	fmt.Printf("files changed: %v\n", result.FilesChanged)
	fmt.Printf("children: %d\n", len(result.Children))
	// Output: files changed: [app.tsx handler.go]
	// children: 2
}

func ExampleTeam_nested() {
	devTeam := team.New("dev-team",
		team.WithAgents(
			team.NewAgent("backend", "backend dev",
				team.HandlerFunc(func(_ context.Context, _ *team.Task) (*team.Result, error) {
					return &team.Result{Output: "backend done"}, nil
				}),
			),
		),
		team.WithDescription("development team"),
	)

	reviewTeam := team.New("review-team",
		team.WithAgents(
			team.NewAgent("security", "security review",
				team.HandlerFunc(func(_ context.Context, _ *team.Task) (*team.Result, error) {
					return &team.Result{Output: "no vulnerabilities"}, nil
				}),
			),
		),
		team.WithDescription("review team"),
	)

	superTeam := team.New("delivery-team",
		team.WithAgents(devTeam.AsAgent(), reviewTeam.AsAgent()),
		team.WithStrategy(team.Sequential),
	)

	result, err := superTeam.Handle(context.Background(), &team.Task{
		ID:          "deliver-1",
		Description: "ship feature",
	})
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
	fmt.Println(result.Output)
	// Output: backend done
	// no vulnerabilities
}

func ExampleTeam_withCostTracking() {
	tracker := cost.NewTracker()

	agent := team.NewAgent("worker", "does work",
		team.HandlerFunc(func(_ context.Context, _ *team.Task) (*team.Result, error) {
			return &team.Result{
				Output: "done",
				Usage:  cost.Usage{InputTokens: 500, OutputTokens: 100, TotalCostUSD: 0.005},
			}, nil
		}),
	)

	t := team.New("tracked-team",
		team.WithAgents(agent),
		team.WithCostTracker(tracker),
	)

	_, err := t.Handle(context.Background(), &team.Task{ID: "t1"})
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}

	total := tracker.Total()
	fmt.Printf("tokens: %d in, %d out\n", total.InputTokens, total.OutputTokens)
	// Output: tokens: 500 in, 100 out
}
