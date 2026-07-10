package runprep

import (
	"context"
	"strings"
	"testing"

	agentruntime "github.com/philjestin/boatman-ecosystem/shared/agentruntime"
	"github.com/philjestin/boatman-ecosystem/shared/agentruntime/memorydocs"
)

func TestPrepareDefaultsRunStoreForProjectWorkDir(t *testing.T) {
	workDir := t.TempDir()

	req, events, err := Prepare(context.Background(), agentruntime.RunRequest{
		WorkDir: workDir,
		Role:    agentruntime.RolePlanner,
	}, Options{DefaultRunStore: true})
	if err != nil {
		t.Fatalf("Prepare error: %v", err)
	}
	if got := req.Metadata[MetadataRunStoreDir]; !strings.HasPrefix(got, workDir) || !strings.HasSuffix(got, ".boatman/runs") {
		t.Fatalf("runStoreDir = %q, want project .boatman/runs", got)
	}
	if len(events) != 0 {
		t.Fatalf("events = %#v, want none without memory", events)
	}
}

func TestPrepareHonorsRunStoreOptOut(t *testing.T) {
	workDir := t.TempDir()

	req, _, err := Prepare(context.Background(), agentruntime.RunRequest{
		WorkDir: workDir,
		Metadata: map[string]string{
			MetadataRunStore: "false",
		},
	}, Options{DefaultRunStore: true})
	if err != nil {
		t.Fatalf("Prepare error: %v", err)
	}
	if req.Metadata[MetadataRunStoreDir] != "" {
		t.Fatalf("runStoreDir = %q, want opt out", req.Metadata[MetadataRunStoreDir])
	}
}

func TestPrepareLoadsMemoryIntoInstructions(t *testing.T) {
	workDir := t.TempDir()
	store := memorydocs.NewFileStore(memorydocs.DefaultDir(workDir))
	if _, err := store.Write(context.Background(), memorydocs.Document{
		ID:    "project",
		Scope: memorydocs.ScopeProject,
		Title: "Project Memory",
		Body:  "Use the shared runtime event contract.",
	}); err != nil {
		t.Fatalf("write memory: %v", err)
	}

	req, events, err := Prepare(context.Background(), agentruntime.RunRequest{
		RunID:        "run-1",
		Provider:     "claude-cli",
		Model:        "sonnet",
		Role:         agentruntime.RoleExecutor,
		Profile:      "executor",
		WorkDir:      workDir,
		Instructions: "Original system prompt.",
	}, Options{LoadMemory: true, MemoryMaxBytes: 4096})
	if err != nil {
		t.Fatalf("Prepare error: %v", err)
	}
	if !strings.Contains(req.Instructions, "# Boatman Memory") || !strings.Contains(req.Instructions, "Use the shared runtime event contract.") || !strings.Contains(req.Instructions, "Original system prompt.") {
		t.Fatalf("Instructions = %q, want memory prepended to system prompt", req.Instructions)
	}
	if len(events) != 1 || events[0].Type != agentruntime.EventMemoryLoaded || events[0].RunID != "run-1" {
		t.Fatalf("events = %#v, want memory.loaded for run-1", events)
	}
	if events[0].Data["targetRole"] != string(agentruntime.RoleExecutor) {
		t.Fatalf("event data = %#v, want target role", events[0].Data)
	}
}

func TestPrepareSkipsMemoryRoles(t *testing.T) {
	workDir := t.TempDir()
	store := memorydocs.NewFileStore(memorydocs.DefaultDir(workDir))
	if _, err := store.Write(context.Background(), memorydocs.Document{
		ID:    "project",
		Title: "Project Memory",
		Body:  "Do not inject into memory distillation.",
	}); err != nil {
		t.Fatalf("write memory: %v", err)
	}

	req, events, err := Prepare(context.Background(), agentruntime.RunRequest{
		Role:    agentruntime.RoleMemory,
		WorkDir: workDir,
	}, Options{LoadMemory: true})
	if err != nil {
		t.Fatalf("Prepare error: %v", err)
	}
	if strings.Contains(req.Instructions, "Do not inject") {
		t.Fatalf("Instructions = %q, should not include memory for memory role", req.Instructions)
	}
	if len(events) != 0 {
		t.Fatalf("events = %#v, want no memory events", events)
	}
}
