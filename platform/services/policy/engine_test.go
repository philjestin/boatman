package policy_test

import (
	"context"
	"testing"

	"github.com/philjestin/boatman-ecosystem/harness/runner"
	"github.com/philjestin/boatman-ecosystem/platform/eventbus"
	"github.com/philjestin/boatman-ecosystem/platform/services/policy"
	"github.com/philjestin/boatman-ecosystem/platform/storage"
	"github.com/philjestin/boatman-ecosystem/platform/storage/sqlite"
)

func setupEngine(t *testing.T) (*policy.Engine, storage.Store, *eventbus.Bus) {
	t.Helper()
	store, err := sqlite.New(sqlite.WithInMemory())
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Migrate(context.Background()); err != nil {
		t.Fatal(err)
	}
	bus, err := eventbus.New(eventbus.WithEventStore(store.Events()))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { bus.Close(); store.Close() })
	return policy.NewEngine(store.Policies()), store, bus
}

func TestEnforceConfig(t *testing.T) {
	engine, store, _ := setupEngine(t)
	ctx := context.Background()

	scope := storage.Scope{OrgID: "org1", TeamID: "team1", RepoID: "repo1"}

	// Set org policy requiring tests and max 3 iterations
	store.Policies().Set(ctx, &storage.Policy{
		ID:            "pol-org",
		Scope:         storage.Scope{OrgID: "org1"},
		MaxIterations: 3,
		RequireTests:  true,
	})

	cfg := runner.Config{
		MaxIterations:    10,
		TestBeforeReview: false,
	}

	enforced, err := engine.EnforceConfig(ctx, scope, cfg)
	if err != nil {
		t.Fatalf("EnforceConfig: %v", err)
	}

	if enforced.MaxIterations != 3 {
		t.Errorf("expected MaxIterations=3, got %d", enforced.MaxIterations)
	}
	if !enforced.TestBeforeReview {
		t.Error("expected TestBeforeReview=true (policy requires tests)")
	}
}

func TestEnforceConfigNoPolicy(t *testing.T) {
	engine, _, _ := setupEngine(t)
	ctx := context.Background()

	cfg := runner.Config{MaxIterations: 10}
	enforced, err := engine.EnforceConfig(ctx, storage.Scope{OrgID: "nopol"}, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if enforced.MaxIterations != 10 {
		t.Errorf("expected unchanged MaxIterations=10, got %d", enforced.MaxIterations)
	}
}

func TestPolicyGuardAllows(t *testing.T) {
	engine, _, bus := setupEngine(t)
	ctx := context.Background()
	scope := storage.Scope{OrgID: "org1", TeamID: "team1"}

	guard := policy.NewPolicyGuard(engine, nil, scope, bus)

	err := guard.AllowStep(ctx, "execute", &runner.GuardState{
		TotalCostUSD: 0.01,
		FilesChanged: 1,
	})
	if err != nil {
		t.Errorf("expected allow, got: %v", err)
	}
}

func TestPolicyGuardDeniesOnCost(t *testing.T) {
	engine, store, bus := setupEngine(t)
	ctx := context.Background()
	scope := storage.Scope{OrgID: "org1", TeamID: "team1"}

	// Set policy with cost limit
	store.Policies().Set(ctx, &storage.Policy{
		ID:             "pol-cost",
		Scope:          scope,
		MaxCostPerRun:  0.50,
	})

	guard := policy.NewPolicyGuard(engine, store.Costs(), scope, bus)

	err := guard.AllowStep(ctx, "refactor", &runner.GuardState{
		TotalCostUSD: 0.60, // exceeds limit
		FilesChanged: 1,
	})
	if err == nil {
		t.Error("expected deny when cost exceeds limit")
	}
}

func TestPolicyGuardDeniesOnFilesChanged(t *testing.T) {
	engine, store, bus := setupEngine(t)
	ctx := context.Background()
	scope := storage.Scope{OrgID: "org1", TeamID: "team1"}

	store.Policies().Set(ctx, &storage.Policy{
		ID:              "pol-files",
		Scope:           scope,
		MaxFilesChanged: 5,
	})

	guard := policy.NewPolicyGuard(engine, store.Costs(), scope, bus)

	err := guard.AllowStep(ctx, "execute", &runner.GuardState{
		FilesChanged: 10, // exceeds limit
	})
	if err == nil {
		t.Error("expected deny when files changed exceeds limit")
	}
}
