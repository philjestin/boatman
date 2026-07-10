package memorydocs

import (
	"context"
	"strings"
	"testing"
	"time"

	agentruntime "github.com/philjestin/boatman-ecosystem/shared/agentruntime"
)

func TestWriteReadAndList(t *testing.T) {
	store := NewFileStore(t.TempDir())
	expires := time.Date(2026, 7, 7, 0, 0, 0, 0, time.UTC)
	written, err := store.Write(context.Background(), Document{
		ID:          "integrations/linear",
		Title:       "Linear Memory",
		Body:        "Use team defaults when planning tickets.",
		Provenance:  "seeded by test",
		SourceRunID: "run-123",
		ExpiresAt:   &expires,
		Metadata:    map[string]string{"owner": "agents"},
	})
	if err != nil {
		t.Fatalf("Write error: %v", err)
	}
	if written.Scope != ScopeIntegration {
		t.Fatalf("Scope = %q, want integration", written.Scope)
	}

	read, err := store.Read(context.Background(), "integrations/linear")
	if err != nil {
		t.Fatalf("Read error: %v", err)
	}
	if read.ID != "integrations/linear" || read.Title != "Linear Memory" || read.SourceRunID != "run-123" {
		t.Fatalf("read = %#v, want round-tripped metadata", read)
	}
	if read.Metadata["owner"] != "agents" {
		t.Fatalf("Metadata = %#v, want owner", read.Metadata)
	}

	docs, err := store.List(context.Background())
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	if len(docs) != 1 || docs[0].ID != "integrations/linear" {
		t.Fatalf("docs = %#v, want linear memory doc", docs)
	}
}

func TestNormalizeIDRejectsUnsafePaths(t *testing.T) {
	for _, id := range []string{"../secret", "/abs", "domains//payments", ".hidden", "bad id"} {
		if _, err := NormalizeID(id); err == nil {
			t.Fatalf("NormalizeID(%q) succeeded, want error", id)
		}
	}
	if got, err := NormalizeID("domains/payments.md"); err != nil || got != "domains/payments" {
		t.Fatalf("NormalizeID returned %q, %v; want domains/payments", got, err)
	}
}

func TestLoadContextSkipsExpiredAndTruncates(t *testing.T) {
	store := NewFileStore(t.TempDir())
	past := time.Now().Add(-time.Hour)
	if _, err := store.Write(context.Background(), Document{
		ID:        "project",
		Body:      "Project memory body that is long enough to truncate.",
		ExpiresAt: nil,
	}); err != nil {
		t.Fatalf("write project: %v", err)
	}
	if _, err := store.Write(context.Background(), Document{
		ID:        "team",
		Body:      "Expired team note",
		ExpiresAt: &past,
	}); err != nil {
		t.Fatalf("write team: %v", err)
	}

	rendered, docs, err := store.LoadContext(context.Background(), nil, 60)
	if err != nil {
		t.Fatalf("LoadContext error: %v", err)
	}
	if len(docs) != 1 || docs[0].ID != "project" {
		t.Fatalf("docs = %#v, want only active project doc", docs)
	}
	if strings.Contains(rendered, "Expired") {
		t.Fatalf("rendered = %q, should skip expired docs", rendered)
	}
	if len(rendered) > 60 {
		t.Fatalf("rendered length = %d, want <= 60", len(rendered))
	}
}

func TestLoadedEvent(t *testing.T) {
	event := LoadedEvent("run-1", []Document{
		{ID: "project", Scope: ScopeProject, Title: "Project"},
	})
	if event.Type != agentruntime.EventMemoryLoaded || event.RunID != "run-1" || event.Status != agentruntime.StatusCompleted {
		t.Fatalf("event = %#v, want memory.loaded completed event", event)
	}
	items, ok := event.Data["documents"].([]map[string]any)
	if !ok || len(items) != 1 || items[0]["id"] != "project" {
		t.Fatalf("event documents = %#v, want project item", event.Data["documents"])
	}
}
