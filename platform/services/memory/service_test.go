package memory_test

import (
	"context"
	"testing"

	"github.com/philjestin/boatman-ecosystem/platform/services/memory"
	"github.com/philjestin/boatman-ecosystem/platform/storage"
	"github.com/philjestin/boatman-ecosystem/platform/storage/sqlite"
)

func setupService(t *testing.T) (*memory.Service, storage.Store) {
	t.Helper()
	store, err := sqlite.New(sqlite.WithInMemory())
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Migrate(context.Background()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { store.Close() })
	return memory.NewService(store.Memory()), store
}

func TestGetMergedPatterns(t *testing.T) {
	svc, store := setupService(t)
	ctx := context.Background()

	orgScope := storage.Scope{OrgID: "org1"}
	repoScope := storage.Scope{OrgID: "org1", TeamID: "team1", RepoID: "repo1"}

	// Add org-level pattern
	store.Memory().CreatePattern(ctx, &storage.Pattern{
		ID:    "org-pat",
		Scope: orgScope,
		Type:  "naming",
		Description: "Org naming convention",
		Weight: 0.8,
	})

	// Add repo-level pattern
	store.Memory().CreatePattern(ctx, &storage.Pattern{
		ID:    "repo-pat",
		Scope: repoScope,
		Type:  "testing",
		Description: "Repo test pattern",
		Weight: 0.9,
	})

	patterns, err := svc.GetMergedPatterns(ctx, repoScope)
	if err != nil {
		t.Fatalf("GetMergedPatterns: %v", err)
	}

	if len(patterns) != 2 {
		t.Errorf("expected 2 merged patterns, got %d", len(patterns))
	}

	// Should be sorted by weight (repo pattern first)
	if len(patterns) >= 2 && patterns[0].Weight < patterns[1].Weight {
		t.Error("patterns should be sorted by weight descending")
	}
}

func TestLearnFromRun(t *testing.T) {
	svc, store := setupService(t)
	ctx := context.Background()

	scope := storage.Scope{OrgID: "org1", TeamID: "team1", RepoID: "repo1"}

	run := &storage.Run{
		ID:           "run-1",
		Scope:        scope,
		FilesChanged: []string{"main.go", "handler.go"},
	}

	// Low score should not learn
	if err := svc.LearnFromRun(ctx, run, 50); err != nil {
		t.Fatal(err)
	}
	patterns, _ := store.Memory().ListPatterns(ctx, scope)
	if len(patterns) != 0 {
		t.Errorf("expected 0 patterns for low score, got %d", len(patterns))
	}

	// High score should learn
	if err := svc.LearnFromRun(ctx, run, 90); err != nil {
		t.Fatal(err)
	}
	patterns, _ = store.Memory().ListPatterns(ctx, scope)
	if len(patterns) != 2 {
		t.Errorf("expected 2 patterns for high score, got %d", len(patterns))
	}
}

func TestToHarnessMemory(t *testing.T) {
	svc, store := setupService(t)
	ctx := context.Background()

	scope := storage.Scope{OrgID: "org1", TeamID: "team1", RepoID: "repo1"}

	store.Memory().CreatePattern(ctx, &storage.Pattern{
		ID: "pat-1", Scope: scope, Type: "naming",
		Description: "Use camelCase", Weight: 0.9,
	})
	store.Memory().SetPreferences(ctx, &storage.Preferences{
		ID: "pref-1", Scope: scope,
		PreferredTestFramework: "go test",
		NamingConventions:      map[string]string{"vars": "camelCase"},
		FileOrganization:       map[string]string{},
		CodeStyle:              map[string]string{},
		ReviewerThresholds:     map[string]int{},
	})
	store.Memory().CreateIssue(ctx, &storage.CommonIssue{
		ID: "issue-1", Scope: scope, Type: "style",
		Description: "Missing error check", Frequency: 3,
	})

	mem, err := svc.ToHarnessMemory(ctx, scope)
	if err != nil {
		t.Fatalf("ToHarnessMemory: %v", err)
	}

	if len(mem.Patterns) != 1 {
		t.Errorf("expected 1 pattern, got %d", len(mem.Patterns))
	}
	if len(mem.CommonIssues) != 1 {
		t.Errorf("expected 1 issue, got %d", len(mem.CommonIssues))
	}
	if mem.Preferences.PreferredTestFramework != "go test" {
		t.Errorf("expected 'go test', got %q", mem.Preferences.PreferredTestFramework)
	}
}

func TestPlatformMemoryStoreAdapter(t *testing.T) {
	svc, store := setupService(t)
	ctx := context.Background()
	scope := storage.Scope{OrgID: "org1", TeamID: "team1", RepoID: "repo1"}

	store.Memory().CreatePattern(ctx, &storage.Pattern{
		ID: "pat-adapter", Scope: scope, Type: "naming",
		Description: "Test pattern", Weight: 0.8,
	})

	adapter := memory.NewPlatformMemoryStore(svc, scope)

	// Get via adapter
	mem, err := adapter.Get("/any/path")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if len(mem.Patterns) != 1 {
		t.Errorf("expected 1 pattern via adapter, got %d", len(mem.Patterns))
	}

	// Save via adapter
	mem.Preferences.PreferredTestFramework = "pytest"
	if err := adapter.Save(mem); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// Verify it persisted
	prefs, _ := store.Memory().GetPreferences(ctx, scope)
	if prefs.PreferredTestFramework != "pytest" {
		t.Errorf("expected 'pytest' after save, got %q", prefs.PreferredTestFramework)
	}
}
