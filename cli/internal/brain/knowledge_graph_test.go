package brain

import (
	"os"
	"path/filepath"
	"testing"

	harnessbrain "github.com/philjestin/boatman-ecosystem/harness/brain"
)

func TestKnowledgeGraphRecordsTaskSignalsAndBrains(t *testing.T) {
	projectPath := t.TempDir()
	graph, err := LoadKnowledgeGraph(projectPath)
	if err != nil {
		t.Fatalf("LoadKnowledgeGraph: %v", err)
	}

	graph.RecordTaskContext("ENG-123", "Fix checkout retries", []string{
		"packs/payments/app/services/retry.rb",
		"packs/payments/app/models/payment.rb",
	})
	graph.RecordTaskExecution("ENG-123", "Fix checkout retries", []string{
		"packs/payments/app/services/retry.rb",
	})
	signal := harnessbrain.Signal{
		Type:      harnessbrain.SignalReviewFailure,
		Domain:    "payments",
		Details:   "Agent missed idempotency",
		FilePaths: []string{"packs/payments/app/services/retry.rb"},
	}
	graph.RecordSignal(signal)
	graph.RecordTaskSignal("ENG-123", "Fix checkout retries", signal)
	graph.RecordBrain(DistillResult{
		Domain:     "payments",
		BrainID:    "auto-payments",
		Path:       filepath.Join(projectPath, ".boatman", "brains", "payments.yaml"),
		MemoryPath: filepath.Join(projectPath, ".boatman", "memory", "domains", "payments.md"),
		Signals:    3,
		UsedLLM:    true,
	})

	if err := graph.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	graphPath := filepath.Join(projectPath, ".boatman", "knowledge", "graph.json")
	if _, err := os.Stat(graphPath); err != nil {
		t.Fatalf("expected graph file: %v", err)
	}

	loaded, err := LoadKnowledgeGraph(projectPath)
	if err != nil {
		t.Fatalf("reload graph: %v", err)
	}
	assertNode(t, loaded, "task:eng-123", "task")
	assertNode(t, loaded, "domain:payments", "domain")
	assertNode(t, loaded, "file:packs-payments-app-services-retry-rb", "file")
	assertNode(t, loaded, "signal:review-failure-payments", "signal")
	assertNode(t, loaded, "brain:auto-payments", "brain")
	assertEdge(t, loaded, "task:eng-123", "file:packs-payments-app-services-retry-rb", "planned_file")
	assertEdge(t, loaded, "task:eng-123", "file:packs-payments-app-services-retry-rb", "changed_file")
	assertEdge(t, loaded, "task:eng-123", "signal:review-failure-payments", "emitted_signal")
	assertEdge(t, loaded, "domain:payments", "brain:auto-payments", "distilled_into")
}

func assertNode(t *testing.T, graph *KnowledgeGraph, id, kind string) {
	t.Helper()
	node, ok := graph.Nodes[id]
	if !ok {
		t.Fatalf("missing node %q", id)
	}
	if node.Kind != kind {
		t.Fatalf("node %q kind = %q, want %q", id, node.Kind, kind)
	}
	if node.Count == 0 {
		t.Fatalf("node %q Count = 0, want positive", id)
	}
}

func assertEdge(t *testing.T, graph *KnowledgeGraph, from, to, kind string) {
	t.Helper()
	for _, edge := range graph.Edges {
		if edge.From == from && edge.To == to && edge.Kind == kind {
			if edge.Count == 0 {
				t.Fatalf("edge %s -> %s %s Count = 0, want positive", from, to, kind)
			}
			return
		}
	}
	t.Fatalf("missing edge %s -> %s %s", from, to, kind)
}
