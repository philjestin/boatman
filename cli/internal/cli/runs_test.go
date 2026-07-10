package cli

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/philjestin/boatman-ecosystem/shared/agentruntime"
	"github.com/philjestin/boatman-ecosystem/shared/agentruntime/runstore"
)

func TestRunsShowJSONOutput(t *testing.T) {
	storeDir := t.TempDir()
	store := runstore.NewFileStore(storeDir)
	req := agentruntime.RunRequest{RunID: "run-json", Provider: "test", Role: agentruntime.RoleChat}
	if err := store.StartRun(context.Background(), req); err != nil {
		t.Fatalf("StartRun error: %v", err)
	}
	event := agentruntime.NewEvent(agentruntime.EventMessageCompleted)
	event.RunID = "run-json"
	event.Message = "hello"
	if err := store.Append(context.Background(), event); err != nil {
		t.Fatalf("Append error: %v", err)
	}

	metadata, events, err := store.LoadRun(context.Background(), "run-json")
	if err != nil {
		t.Fatalf("LoadRun error: %v", err)
	}
	var out bytes.Buffer
	if err := writeRunShow(&out, metadata, events, true); err != nil {
		t.Fatalf("writeRunShow error: %v", err)
	}
	if !strings.Contains(out.String(), `"runId":"run-json"`) || !strings.Contains(out.String(), `"message":"hello"`) {
		t.Fatalf("output = %s, want run metadata and event message", out.String())
	}
}

func TestWriteRunArtifactsText(t *testing.T) {
	var out bytes.Buffer
	err := writeRunArtifacts(&out, []runstore.ArtifactRecord{
		{
			Kind:      "file",
			Path:      "internal/foo.go",
			EventType: agentruntime.EventArtifactChanged,
			Message:   "updated implementation",
		},
	}, false)
	if err != nil {
		t.Fatalf("writeRunArtifacts error: %v", err)
	}
	output := out.String()
	for _, want := range []string{"KIND", "internal/foo.go", "updated implementation"} {
		if !strings.Contains(output, want) {
			t.Fatalf("output %q should contain %q", output, want)
		}
	}
}

func TestWriteRunRequestJSON(t *testing.T) {
	var out bytes.Buffer
	err := writeRunRequest(&out, agentruntime.RunRequest{
		RunID:    "run-request",
		Provider: "claude-cli",
		Role:     agentruntime.RolePlanner,
		Messages: []agentruntime.Message{{Role: "user", Content: "plan"}},
	}, true)
	if err != nil {
		t.Fatalf("writeRunRequest error: %v", err)
	}
	if !strings.Contains(out.String(), `"runId": "run-request"`) || !strings.Contains(out.String(), `"provider": "claude-cli"`) {
		t.Fatalf("output = %s, want request JSON", out.String())
	}
}

func TestEventSummary(t *testing.T) {
	event := agentruntime.NewEvent(agentruntime.EventToolResult)
	event.Tool = &agentruntime.ToolEvent{Name: "Read"}
	if got := eventSummary(event); got != "Read" {
		t.Fatalf("eventSummary = %q, want Read", got)
	}

	event = agentruntime.NewEvent(agentruntime.EventUsageUpdated)
	event.Usage = &agentruntime.Usage{InputTokens: 10, OutputTokens: 5}
	if got := eventSummary(event); !strings.Contains(got, "input=10") {
		t.Fatalf("eventSummary = %q, want usage", got)
	}

	event = agentruntime.NewEvent(agentruntime.EventIntegrationState)
	event.Data = map[string]any{
		"state":       "needs_configuration",
		"missing_env": []any{"DD_API_KEY", "DD_APP_KEY"},
	}
	if got := eventSummary(event); got != "needs_configuration missing=DD_API_KEY,DD_APP_KEY" {
		t.Fatalf("eventSummary = %q, want integration state summary", got)
	}

	event = agentruntime.NewEvent(agentruntime.EventMemoryLoaded)
	event.Data = map[string]any{
		"documents": []map[string]any{{"id": "project"}, {"id": "domains/payments"}},
	}
	if got := eventSummary(event); got != "loaded=project,domains/payments" {
		t.Fatalf("eventSummary = %q, want memory document summary", got)
	}

	event = agentruntime.NewEvent(agentruntime.EventArtifactChanged)
	event.Message = "created patch"
	event.Artifact = &agentruntime.ArtifactEvent{Kind: "diff", Path: "internal/foo.go"}
	if got := eventSummary(event); got != "internal/foo.go created patch" {
		t.Fatalf("eventSummary = %q, want artifact summary", got)
	}
}
