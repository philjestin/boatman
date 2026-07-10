package main

import (
	"context"
	"path/filepath"
	"testing"

	agentruntime "github.com/philjestin/boatman-ecosystem/shared/agentruntime"
	"github.com/philjestin/boatman-ecosystem/shared/agentruntime/memorydocs"
	"github.com/philjestin/boatman-ecosystem/shared/agentruntime/runstore"
)

func TestRuntimeInspectionListsAndLoadsRuns(t *testing.T) {
	t.Setenv("BOATMAN_RUNTIME_STORE_DIR", "")
	projectDir := t.TempDir()
	store := runstore.NewFileStore(filepath.Join(projectDir, ".boatman", "runs"))

	req := agentruntime.RunRequest{
		RunID:    "run-123",
		Provider: "claude-cli",
		Model:    "sonnet",
		Role:     agentruntime.RolePlanner,
		Profile:  "work-planner",
		WorkDir:  projectDir,
	}
	if err := store.StartRun(context.Background(), req); err != nil {
		t.Fatalf("StartRun() error = %v", err)
	}

	artifactEvent := agentruntime.NewEvent(agentruntime.EventArtifactChanged)
	artifactEvent.RunID = req.RunID
	artifactEvent.Provider = req.Provider
	artifactEvent.Model = req.Model
	artifactEvent.Role = req.Role
	artifactEvent.Status = agentruntime.StatusCompleted
	artifactEvent.Message = "plan written"
	artifactEvent.Artifact = &agentruntime.ArtifactEvent{Kind: "file", Path: "plan.md"}
	if err := store.Append(context.Background(), artifactEvent); err != nil {
		t.Fatalf("Append(artifact) error = %v", err)
	}

	completedEvent := agentruntime.NewEvent(agentruntime.EventRunCompleted)
	completedEvent.RunID = req.RunID
	completedEvent.Status = agentruntime.StatusSucceeded
	if err := store.Append(context.Background(), completedEvent); err != nil {
		t.Fatalf("Append(completed) error = %v", err)
	}

	app := &App{}
	runs, err := app.ListRuntimeRuns(projectDir)
	if err != nil {
		t.Fatalf("ListRuntimeRuns() error = %v", err)
	}
	if len(runs) != 1 {
		t.Fatalf("ListRuntimeRuns() len = %d, want 1", len(runs))
	}
	if runs[0].RunID != "run-123" || runs[0].Provider != "claude-cli" || runs[0].ArtifactCount != 1 {
		t.Fatalf("ListRuntimeRuns() = %+v", runs[0])
	}

	detail, err := app.GetRuntimeRun(projectDir, "run-123")
	if err != nil {
		t.Fatalf("GetRuntimeRun() error = %v", err)
	}
	if detail.Metadata.Status != string(agentruntime.StatusSucceeded) {
		t.Fatalf("detail status = %q, want succeeded", detail.Metadata.Status)
	}
	if len(detail.Events) != 2 {
		t.Fatalf("detail events len = %d, want 2", len(detail.Events))
	}
	if len(detail.Artifacts) != 1 || detail.Artifacts[0].Path != "plan.md" {
		t.Fatalf("detail artifacts = %+v", detail.Artifacts)
	}
}

func TestRuntimeInspectionListsAndLoadsMemoryDocuments(t *testing.T) {
	t.Setenv("BOATMAN_MEMORY_DIR", "")
	projectDir := t.TempDir()
	store := memorydocs.NewFileStore(filepath.Join(projectDir, ".boatman", "memory"))
	doc, err := store.Write(context.Background(), memorydocs.Document{
		ID:          "domains/payments",
		Title:       "Payments",
		Provenance:  "brain distiller",
		SourceRunID: "run-123",
		Body:        "Prefer the payment gateway test helpers before hand-rolled mocks.",
	})
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	app := &App{}
	docs, err := app.ListMemoryDocuments(projectDir)
	if err != nil {
		t.Fatalf("ListMemoryDocuments() error = %v", err)
	}
	if len(docs) != 1 {
		t.Fatalf("ListMemoryDocuments() len = %d, want 1", len(docs))
	}
	if docs[0].ID != doc.ID || docs[0].Scope != string(memorydocs.ScopeDomain) || docs[0].SourceRunID != "run-123" {
		t.Fatalf("ListMemoryDocuments() = %+v", docs[0])
	}

	detail, err := app.GetMemoryDocument(projectDir, "domains/payments")
	if err != nil {
		t.Fatalf("GetMemoryDocument() error = %v", err)
	}
	if detail.Title != "Payments" {
		t.Fatalf("detail title = %q, want Payments", detail.Title)
	}
	if detail.Body != "Prefer the payment gateway test helpers before hand-rolled mocks." {
		t.Fatalf("detail body = %q", detail.Body)
	}
}
