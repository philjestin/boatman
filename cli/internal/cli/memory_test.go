package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	agentruntime "github.com/philjestin/boatman-ecosystem/shared/agentruntime"
	"github.com/philjestin/boatman-ecosystem/shared/agentruntime/memorydocs"
	"github.com/spf13/cobra"
)

func TestWriteMemoryListText(t *testing.T) {
	var buf bytes.Buffer
	err := writeMemoryList(&buf, []memorydocs.Document{
		{ID: "project", Scope: memorydocs.ScopeProject, Title: "Project Memory"},
	}, false)
	if err != nil {
		t.Fatalf("writeMemoryList error: %v", err)
	}
	output := buf.String()
	for _, want := range []string{"ID", "project", "Project Memory"} {
		if !strings.Contains(output, want) {
			t.Fatalf("output %q should contain %q", output, want)
		}
	}
}

func TestWriteMemoryShowJSON(t *testing.T) {
	var buf bytes.Buffer
	err := writeMemoryShow(&buf, memorydocs.Document{
		ID:    "domains/payments",
		Scope: memorydocs.ScopeDomain,
		Title: "Payments",
		Body:  "Idempotency matters.",
	}, true)
	if err != nil {
		t.Fatalf("writeMemoryShow error: %v", err)
	}
	var decoded memorydocs.Document
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("json output did not decode: %v", err)
	}
	if decoded.ID != "domains/payments" || decoded.Body != "Idempotency matters." {
		t.Fatalf("decoded = %#v, want payments memory doc", decoded)
	}
}

func TestRunMemoryContextEmitsLoadedEvent(t *testing.T) {
	dir := t.TempDir()
	store := memorydocs.NewFileStore(dir)
	if _, err := store.Write(context.Background(), memorydocs.Document{
		ID:    "project",
		Title: "Project",
		Body:  "Use existing provider contracts.",
	}); err != nil {
		t.Fatalf("Write error: %v", err)
	}

	cmd := &cobra.Command{Use: "context"}
	cmd.Flags().String("memory-dir", dir, "")
	cmd.Flags().Int("max-bytes", 12000, "")
	cmd.Flags().Bool("emit-event", true, "")
	cmd.Flags().String("run-id", "run-memory", "")
	var out bytes.Buffer
	cmd.SetOut(&out)
	if err := runMemoryContext(cmd, []string{"project"}); err != nil {
		t.Fatalf("memory context error: %v", err)
	}
	output := out.String()
	if !strings.Contains(output, "Use existing provider contracts.") {
		t.Fatalf("output = %q, want rendered memory body", output)
	}
	var event agentruntime.Event
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &event); err != nil {
		t.Fatalf("last line did not decode as event: %v\n%s", err, output)
	}
	if event.Type != agentruntime.EventMemoryLoaded || event.RunID != "run-memory" {
		t.Fatalf("event = %#v, want memory.loaded run-memory", event)
	}
}
