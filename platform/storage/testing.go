package storage

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// RunStoreTests is a compliance test suite that any Store backend must pass.
// It exercises all sub-stores with standard operations.
func RunStoreTests(t *testing.T, store Store) {
	ctx := context.Background()

	if err := store.Migrate(ctx); err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	t.Run("MigrateIdempotent", func(t *testing.T) {
		if err := store.Migrate(ctx); err != nil {
			t.Fatalf("second Migrate should be idempotent: %v", err)
		}
	})

	t.Run("RunStore", func(t *testing.T) {
		testRunStore(t, ctx, store.Runs())
	})

	t.Run("MemoryStore", func(t *testing.T) {
		testMemoryStore(t, ctx, store.Memory())
	})

	t.Run("CostStore", func(t *testing.T) {
		testCostStore(t, ctx, store.Costs())
	})

	t.Run("PolicyStore", func(t *testing.T) {
		testPolicyStore(t, ctx, store.Policies())
	})

	t.Run("EventStore", func(t *testing.T) {
		testEventStore(t, ctx, store.Events())
	})
}

func testRunStore(t *testing.T, ctx context.Context, rs RunStore) {
	scope := Scope{OrgID: "org1", TeamID: "team1", RepoID: "repo1"}

	// Create
	run := &Run{
		ID:           "run-1",
		Scope:        scope,
		UserID:       "user1",
		Status:       RunStatusPending,
		Prompt:       "add a hello endpoint",
		FilesChanged: []string{"main.go"},
		Duration:     5 * time.Second,
	}
	if err := rs.Create(ctx, run); err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Get
	got, err := rs.Get(ctx, "run-1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.ID != "run-1" || got.UserID != "user1" || got.Status != RunStatusPending {
		t.Errorf("Get returned unexpected values: %+v", got)
	}
	if len(got.FilesChanged) != 1 || got.FilesChanged[0] != "main.go" {
		t.Errorf("FilesChanged mismatch: %v", got.FilesChanged)
	}

	// Update
	run.Status = RunStatusRunning
	run.TotalCostUSD = 0.05
	if err := rs.Update(ctx, run); err != nil {
		t.Fatalf("Update: %v", err)
	}
	got, _ = rs.Get(ctx, "run-1")
	if got.Status != RunStatusRunning || got.TotalCostUSD != 0.05 {
		t.Errorf("Update not reflected: status=%s cost=%f", got.Status, got.TotalCostUSD)
	}

	// List with filter
	runs, err := rs.List(ctx, RunFilter{Scope: &scope})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(runs) != 1 {
		t.Errorf("expected 1 run, got %d", len(runs))
	}

	// List with status filter
	runs, err = rs.List(ctx, RunFilter{Status: RunStatusPassed})
	if err != nil {
		t.Fatalf("List with status: %v", err)
	}
	if len(runs) != 0 {
		t.Errorf("expected 0 passed runs, got %d", len(runs))
	}
}

func testMemoryStore(t *testing.T, ctx context.Context, ms MemoryStore) {
	scope := Scope{OrgID: "org1", TeamID: "team1", RepoID: "repo1"}

	// Patterns
	p := &Pattern{
		ID:          "pat-1",
		Scope:       scope,
		Type:        "naming",
		Description: "Use camelCase for functions",
		Weight:      0.9,
	}
	if err := ms.CreatePattern(ctx, p); err != nil {
		t.Fatalf("CreatePattern: %v", err)
	}

	patterns, err := ms.ListPatterns(ctx, scope)
	if err != nil {
		t.Fatalf("ListPatterns: %v", err)
	}
	if len(patterns) != 1 || patterns[0].Description != "Use camelCase for functions" {
		t.Errorf("ListPatterns unexpected: %+v", patterns)
	}

	p.Weight = 0.95
	if err := ms.UpdatePattern(ctx, p); err != nil {
		t.Fatalf("UpdatePattern: %v", err)
	}

	if err := ms.DeletePattern(ctx, "pat-1"); err != nil {
		t.Fatalf("DeletePattern: %v", err)
	}
	patterns, _ = ms.ListPatterns(ctx, scope)
	if len(patterns) != 0 {
		t.Errorf("expected 0 patterns after delete, got %d", len(patterns))
	}

	// Preferences
	prefs := &Preferences{
		ID:                     "pref-1",
		Scope:                  scope,
		PreferredTestFramework: "go test",
		NamingConventions:      map[string]string{"vars": "camelCase"},
		FileOrganization:       map[string]string{},
		CodeStyle:              map[string]string{},
		ReviewerThresholds:     map[string]int{},
	}
	if err := ms.SetPreferences(ctx, prefs); err != nil {
		t.Fatalf("SetPreferences: %v", err)
	}

	got, err := ms.GetPreferences(ctx, scope)
	if err != nil {
		t.Fatalf("GetPreferences: %v", err)
	}
	if got.PreferredTestFramework != "go test" {
		t.Errorf("expected 'go test', got %q", got.PreferredTestFramework)
	}

	// Common issues
	issue := &CommonIssue{
		ID:          "issue-1",
		Scope:       scope,
		Type:        "style",
		Description: "Missing error check",
		Solution:    "Always check returned errors",
		Frequency:   3,
	}
	if err := ms.CreateIssue(ctx, issue); err != nil {
		t.Fatalf("CreateIssue: %v", err)
	}

	issues, err := ms.ListIssues(ctx, scope)
	if err != nil {
		t.Fatalf("ListIssues: %v", err)
	}
	if len(issues) != 1 {
		t.Errorf("expected 1 issue, got %d", len(issues))
	}

	issue.Frequency = 5
	if err := ms.UpdateIssue(ctx, issue); err != nil {
		t.Fatalf("UpdateIssue: %v", err)
	}
}

func testCostStore(t *testing.T, ctx context.Context, cs CostStore) {
	scope := Scope{OrgID: "org1", TeamID: "team1", RepoID: "repo1"}

	// Record usage
	record := &UsageRecord{
		ID:           "usage-1",
		RunID:        "run-1",
		Scope:        scope,
		Step:         "execute",
		InputTokens:  1000,
		OutputTokens: 500,
		TotalCostUSD: 0.01,
	}
	if err := cs.RecordUsage(ctx, record); err != nil {
		t.Fatalf("RecordUsage: %v", err)
	}

	// Get usage
	records, err := cs.GetUsage(ctx, UsageFilter{RunID: "run-1"})
	if err != nil {
		t.Fatalf("GetUsage: %v", err)
	}
	if len(records) != 1 || records[0].Step != "execute" {
		t.Errorf("GetUsage unexpected: %+v", records)
	}

	// Budget
	budget := &Budget{
		ID:           "budget-1",
		Scope:        scope,
		MonthlyLimit: 100.0,
		DailyLimit:   10.0,
		PerRunLimit:  1.0,
		AlertAt:      0.8,
	}
	if err := cs.SetBudget(ctx, budget); err != nil {
		t.Fatalf("SetBudget: %v", err)
	}

	got, err := cs.GetBudget(ctx, scope)
	if err != nil {
		t.Fatalf("GetBudget: %v", err)
	}
	if got == nil || got.MonthlyLimit != 100.0 {
		t.Errorf("GetBudget unexpected: %+v", got)
	}
}

func testPolicyStore(t *testing.T, ctx context.Context, ps PolicyStore) {
	orgScope := Scope{OrgID: "org1"}
	teamScope := Scope{OrgID: "org1", TeamID: "team1"}
	repoScope := Scope{OrgID: "org1", TeamID: "team1", RepoID: "repo1"}

	// Set org policy
	orgPolicy := &Policy{
		ID:            "pol-org",
		Scope:         orgScope,
		MaxIterations: 5,
		RequireTests:  true,
		AllowedModels: []string{"claude-3-opus", "claude-3-sonnet", "claude-3-haiku"},
	}
	if err := ps.Set(ctx, orgPolicy); err != nil {
		t.Fatalf("Set org policy: %v", err)
	}

	// Set team policy (more restrictive iterations)
	teamPolicy := &Policy{
		ID:            "pol-team",
		Scope:         teamScope,
		MaxIterations: 3,
		AllowedModels: []string{"claude-3-opus", "claude-3-sonnet"},
	}
	if err := ps.Set(ctx, teamPolicy); err != nil {
		t.Fatalf("Set team policy: %v", err)
	}

	// Get effective policy at repo level
	effective, err := ps.GetEffectivePolicy(ctx, repoScope)
	if err != nil {
		t.Fatalf("GetEffectivePolicy: %v", err)
	}
	if effective == nil {
		t.Fatal("expected non-nil effective policy")
	}
	if effective.MaxIterations != 3 {
		t.Errorf("expected MaxIterations=3 (team override), got %d", effective.MaxIterations)
	}
	if !effective.RequireTests {
		t.Error("expected RequireTests=true (inherited from org)")
	}
	// Allowed models should be intersection: claude-3-opus, claude-3-sonnet
	if len(effective.AllowedModels) != 2 {
		t.Errorf("expected 2 allowed models (intersection), got %d: %v", len(effective.AllowedModels), effective.AllowedModels)
	}

	// Delete
	if err := ps.Delete(ctx, teamScope); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	got, _ := ps.Get(ctx, teamScope)
	if got != nil {
		t.Error("expected nil after delete")
	}
}

func testEventStore(t *testing.T, ctx context.Context, es EventStore) {
	scope := Scope{OrgID: "org1", TeamID: "team1", RepoID: "repo1"}

	// Publish events
	for i := 0; i < 5; i++ {
		event := &Event{
			ID:    fmt.Sprintf("evt-%d", i),
			RunID: "run-1",
			Scope: scope,
			Type:  "step.complete",
			Name:  fmt.Sprintf("step_%d", i),
		}
		if err := es.Publish(ctx, event); err != nil {
			t.Fatalf("Publish event %d: %v", i, err)
		}
		// Small delay so events have distinct timestamps
		time.Sleep(time.Millisecond)
	}

	// Query all
	events, err := es.Query(ctx, EventFilter{RunID: "run-1"})
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if len(events) != 5 {
		t.Errorf("expected 5 events, got %d", len(events))
	}

	// Query with type filter
	events, err = es.Query(ctx, EventFilter{
		RunID: "run-1",
		Types: []string{"step.complete"},
		Limit: 2,
	})
	if err != nil {
		t.Fatalf("Query with filter: %v", err)
	}
	if len(events) != 2 {
		t.Errorf("expected 2 events with limit, got %d", len(events))
	}
}
