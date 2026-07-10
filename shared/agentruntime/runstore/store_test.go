package runstore

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	agentruntime "github.com/philjestin/boatman-ecosystem/shared/agentruntime"
)

func TestFileStoreStartAppendLoadRun(t *testing.T) {
	store := NewFileStore(t.TempDir())
	ctx := context.Background()

	req := agentruntime.RunRequest{
		RunID:    "run-1",
		Provider: "claude-cli",
		Model:    "sonnet",
		Role:     agentruntime.RolePlanner,
		Profile:  "planner",
		WorkDir:  "/repo",
	}
	if err := store.StartRun(ctx, req); err != nil {
		t.Fatalf("StartRun error: %v", err)
	}
	storedReq, err := store.LoadRequest(ctx, "run-1")
	if err != nil {
		t.Fatalf("LoadRequest error: %v", err)
	}
	if storedReq.RunID != "run-1" || storedReq.Provider != "claude-cli" || storedReq.Profile != "planner" {
		t.Fatalf("stored request = %#v, want original request", storedReq)
	}

	started := agentruntime.NewEvent(agentruntime.EventRunStarted)
	started.RunID = "run-1"
	started.Provider = "claude-cli"
	started.Role = agentruntime.RolePlanner
	started.Status = agentruntime.StatusStarted
	if err := store.Append(ctx, started); err != nil {
		t.Fatalf("Append started error: %v", err)
	}

	message := agentruntime.NewEvent(agentruntime.EventMessageCompleted)
	message.RunID = "run-1"
	message.Message = "done"
	if err := store.Append(ctx, message); err != nil {
		t.Fatalf("Append message error: %v", err)
	}

	completed := agentruntime.NewEvent(agentruntime.EventRunCompleted)
	completed.RunID = "run-1"
	completed.Status = agentruntime.StatusSucceeded
	if err := store.Append(ctx, completed); err != nil {
		t.Fatalf("Append completed error: %v", err)
	}

	metadata, events, err := store.LoadRun(ctx, "run-1")
	if err != nil {
		t.Fatalf("LoadRun error: %v", err)
	}
	if metadata.RunID != "run-1" || metadata.Provider != "claude-cli" || metadata.Role != agentruntime.RolePlanner {
		t.Fatalf("metadata = %#v, want request metadata", metadata)
	}
	if metadata.Status != agentruntime.StatusSucceeded {
		t.Fatalf("Status = %q, want succeeded", metadata.Status)
	}
	if metadata.EndedAt == nil {
		t.Fatal("EndedAt should be set for completed run")
	}
	if metadata.EventCount != 3 {
		t.Fatalf("EventCount = %d, want 3", metadata.EventCount)
	}
	if len(events) != 3 || events[1].Message != "done" {
		t.Fatalf("events = %#v, want persisted message", events)
	}
}

func TestFileStoreIndexesArtifacts(t *testing.T) {
	store := NewFileStore(t.TempDir())
	ctx := context.Background()

	req := agentruntime.RunRequest{RunID: "artifact-run", Provider: "test", Role: agentruntime.RoleExecutor}
	if err := store.StartRun(ctx, req); err != nil {
		t.Fatalf("StartRun error: %v", err)
	}

	artifact := agentruntime.NewEvent(agentruntime.EventArtifactChanged)
	artifact.RunID = "artifact-run"
	artifact.PhaseID = "execute"
	artifact.TaskID = "ENG-1"
	artifact.Message = "updated implementation"
	artifact.Artifact = &agentruntime.ArtifactEvent{
		Kind: "file",
		Path: "internal/foo.go",
		Diff: "@@ old",
	}
	if err := store.Append(ctx, artifact); err != nil {
		t.Fatalf("Append artifact error: %v", err)
	}

	artifact.Message = "updated implementation again"
	artifact.Artifact.Diff = "@@ new"
	if err := store.Append(ctx, artifact); err != nil {
		t.Fatalf("Append artifact update error: %v", err)
	}

	artifacts, err := store.ListArtifacts(ctx, "artifact-run")
	if err != nil {
		t.Fatalf("ListArtifacts error: %v", err)
	}
	if len(artifacts) != 1 {
		t.Fatalf("artifacts = %#v, want one deduplicated artifact", artifacts)
	}
	if artifacts[0].Path != "internal/foo.go" || artifacts[0].EventCount != 2 || artifacts[0].Diff != "@@ new" {
		t.Fatalf("artifact = %#v, want updated artifact record", artifacts[0])
	}

	metadata, _, err := store.LoadRun(ctx, "artifact-run")
	if err != nil {
		t.Fatalf("LoadRun error: %v", err)
	}
	if metadata.ArtifactCount != 1 {
		t.Fatalf("ArtifactCount = %d, want 1", metadata.ArtifactCount)
	}
}

func TestFileStoreListRunsNewestFirst(t *testing.T) {
	store := NewFileStore(t.TempDir())
	ctx := context.Background()

	oldEvent := agentruntime.NewEvent(agentruntime.EventRunStarted)
	oldEvent.RunID = "old"
	oldEvent.Timestamp = time.Now().Add(-time.Hour)
	if err := store.Append(ctx, oldEvent); err != nil {
		t.Fatalf("append old: %v", err)
	}

	newEvent := agentruntime.NewEvent(agentruntime.EventRunStarted)
	newEvent.RunID = "new"
	newEvent.Timestamp = time.Now()
	if err := store.Append(ctx, newEvent); err != nil {
		t.Fatalf("append new: %v", err)
	}

	runs, err := store.ListRuns(ctx)
	if err != nil {
		t.Fatalf("ListRuns error: %v", err)
	}
	if len(runs) != 2 || runs[0].RunID != "new" || runs[1].RunID != "old" {
		t.Fatalf("runs = %#v, want newest first", runs)
	}
}

func TestFileStoreSanitizesRunIDPath(t *testing.T) {
	root := t.TempDir()
	store := NewFileStore(root)
	ctx := context.Background()

	event := agentruntime.NewEvent(agentruntime.EventRunStarted)
	event.RunID = "../bad/run"
	if err := store.Append(ctx, event); err != nil {
		t.Fatalf("Append error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(root, ".._bad_run", eventsFile)); err != nil {
		t.Fatalf("expected sanitized events file: %v", err)
	}

	metadata, _, err := store.LoadRun(ctx, "../bad/run")
	if err != nil {
		t.Fatalf("LoadRun error: %v", err)
	}
	if metadata.RunID != "../bad/run" {
		t.Fatalf("metadata RunID = %q, want original ID", metadata.RunID)
	}
}

func TestFileStoreMissingRun(t *testing.T) {
	store := NewFileStore(t.TempDir())

	_, _, err := store.LoadRun(context.Background(), "missing")
	if err == nil {
		t.Fatal("LoadRun should fail for missing run")
	}
}

func TestForRequestUsesMetadataDir(t *testing.T) {
	dir := t.TempDir()

	store, enabled, err := ForRequest(agentruntime.RunRequest{
		Metadata: map[string]string{"runStoreDir": dir},
	})
	if err != nil {
		t.Fatalf("ForRequest error: %v", err)
	}
	if !enabled {
		t.Fatal("ForRequest should be enabled for metadata dir")
	}
	if store.Root() != dir {
		t.Fatalf("Root = %q, want %q", store.Root(), dir)
	}
}

func TestForRequestUsesEnvDir(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("BOATMAN_RUNTIME_STORE_DIR", dir)

	store, enabled, err := ForRequest(agentruntime.RunRequest{})
	if err != nil {
		t.Fatalf("ForRequest error: %v", err)
	}
	if !enabled || store.Root() != dir {
		t.Fatalf("enabled/root = %v/%q, want %q", enabled, store.Root(), dir)
	}
}

func TestForRequestDefaultOptIn(t *testing.T) {
	workDir := t.TempDir()
	t.Setenv("BOATMAN_RUNTIME_STORE", "1")

	store, enabled, err := ForRequest(agentruntime.RunRequest{WorkDir: workDir})
	if err != nil {
		t.Fatalf("ForRequest error: %v", err)
	}
	if !enabled {
		t.Fatal("ForRequest should be enabled")
	}
	if !strings.HasSuffix(store.Root(), filepath.Join(".boatman", "runs")) {
		t.Fatalf("Root = %q, want .boatman/runs suffix", store.Root())
	}
}

func TestForRequestDisabledByDefault(t *testing.T) {
	store, enabled, err := ForRequest(agentruntime.RunRequest{})
	if err != nil {
		t.Fatalf("ForRequest error: %v", err)
	}
	if enabled || store != nil {
		t.Fatalf("enabled/store = %v/%#v, want disabled nil", enabled, store)
	}
}
