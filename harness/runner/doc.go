// Package runner provides a composable pipeline for AI agent harnesses.
//
// The Runner orchestrates a role-based execute-test-review-refactor loop.
// Two roles are required (Developer and review.Reviewer); two are optional
// (Tester and Planner). The runner never calls an LLM directly — all model
// interaction happens inside role implementations, making it model-agnostic.
//
// # Minimal usage
//
//	r := runner.New(myDev, myReviewer)
//	result, err := r.Run(ctx, &runner.Request{
//	    ID:          "task-1",
//	    Title:       "Add auth",
//	    Description: "Implement JWT auth for the API",
//	    WorkDir:     "/path/to/repo",
//	})
//
// # Optional features
//
// Use functional options to add a Tester, Planner, cost tracker, checkpoint
// manager, lifecycle hooks, Observer, or Guard:
//
//	r := runner.New(dev, rev,
//	    runner.WithPlanner(planner),
//	    runner.WithTester(runner.NewTestRunnerTester("/repo")),
//	    runner.WithMaxIterations(5),
//	    runner.WithObserver(observer),
//	    runner.WithGuard(guard),
//	    runner.WithHooks(runner.Hooks{
//	        OnStepStart: func(name string) { log.Println("starting", name) },
//	    }),
//	)
//
// The Observer interface receives lifecycle events (OnRunStart, OnRunComplete,
// OnStepStart, OnStepComplete) for logging, metrics, or event publishing.
//
// The Guard interface enables mid-run policy enforcement. Before each step,
// AllowStep is called with the current GuardState (iterations, elapsed time,
// cost, files changed). Return an error to halt the run.
//
// # Pipeline flow
//
//  1. Plan (if Planner set) — produces a Plan for the Developer.
//  2. Execute — Developer.Execute produces initial code changes.
//  3. Review loop (1..MaxIterations):
//     a. Test (if Tester set and TestBeforeReview) — runs tests.
//     b. Review — Reviewer.Review evaluates the diff.
//     c. If passed, break.
//     d. Refactor — Developer.Refactor addresses review issues.
//  4. Finalize — return Result with status, metrics, and history.
//
// # Primitive integration
//
// Flow-affecting primitives (checkpoint, cost, issuetracker) are managed by
// the Runner. Step-affecting primitives (contextpin, memory, filesummary,
// compression) belong inside role implementations.
package runner
