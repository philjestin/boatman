package cost_test

import (
	"context"
	"testing"

	harnesscost "github.com/philjestin/boatman-ecosystem/harness/cost"
	"github.com/philjestin/boatman-ecosystem/platform/eventbus"
	costservice "github.com/philjestin/boatman-ecosystem/platform/services/cost"
	"github.com/philjestin/boatman-ecosystem/platform/storage"
	"github.com/philjestin/boatman-ecosystem/platform/storage/sqlite"
)

func setupGovernor(t *testing.T) (*costservice.Governor, storage.Store) {
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
	return costservice.NewGovernor(store.Costs(), bus), store
}

func TestRecordStep(t *testing.T) {
	gov, store := setupGovernor(t)
	ctx := context.Background()
	scope := storage.Scope{OrgID: "org1", TeamID: "team1", RepoID: "repo1"}

	usage := harnesscost.Usage{
		InputTokens:  1000,
		OutputTokens: 500,
		TotalCostUSD: 0.01,
	}

	if err := gov.RecordStep(ctx, "run-1", "execute", usage, scope); err != nil {
		t.Fatalf("RecordStep: %v", err)
	}

	// Verify persisted
	records, err := store.Costs().GetUsage(ctx, storage.UsageFilter{RunID: "run-1"})
	if err != nil {
		t.Fatalf("GetUsage: %v", err)
	}
	if len(records) != 1 {
		t.Errorf("expected 1 usage record, got %d", len(records))
	}
}

func TestCheckBudgetNoBudget(t *testing.T) {
	gov, _ := setupGovernor(t)
	ctx := context.Background()
	scope := storage.Scope{OrgID: "org1", TeamID: "team1"}

	status, err := gov.CheckBudget(ctx, scope)
	if err != nil {
		t.Fatalf("CheckBudget: %v", err)
	}
	if status.AtLimit {
		t.Error("should not be at limit with no budget set")
	}
}

func TestCheckBudgetWithLimit(t *testing.T) {
	gov, store := setupGovernor(t)
	ctx := context.Background()
	scope := storage.Scope{OrgID: "org1", TeamID: "team1", RepoID: "repo1"}

	// Set budget
	store.Costs().SetBudget(ctx, &storage.Budget{
		ID:           "budget-1",
		Scope:        scope,
		DailyLimit:   1.00,
		MonthlyLimit: 10.00,
		AlertAt:      0.8,
	})

	// Record usage that exceeds daily alert threshold
	usage := harnesscost.Usage{
		InputTokens:  50000,
		OutputTokens: 25000,
		TotalCostUSD: 0.90,
	}
	gov.RecordStep(ctx, "run-1", "execute", usage, scope)

	status, err := gov.CheckBudget(ctx, scope)
	if err != nil {
		t.Fatalf("CheckBudget: %v", err)
	}

	if !status.AlertTriggered {
		t.Error("expected alert to be triggered at 90% of daily limit")
	}
}

func TestCheckBudgetAtLimit(t *testing.T) {
	gov, store := setupGovernor(t)
	ctx := context.Background()
	scope := storage.Scope{OrgID: "org1", TeamID: "team1", RepoID: "repo1"}

	store.Costs().SetBudget(ctx, &storage.Budget{
		ID:         "budget-2",
		Scope:      scope,
		DailyLimit: 0.50,
		AlertAt:    0.8,
	})

	// Record usage that exceeds daily limit
	usage := harnesscost.Usage{TotalCostUSD: 0.60}
	gov.RecordStep(ctx, "run-2", "execute", usage, scope)

	status, err := gov.CheckBudget(ctx, scope)
	if err != nil {
		t.Fatal(err)
	}
	if !status.AtLimit {
		t.Error("expected to be at limit")
	}
}
