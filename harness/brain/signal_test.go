package brain

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSignalStoreRecordAndGet(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "signal-test")
	defer os.RemoveAll(tmpDir)

	path := filepath.Join(tmpDir, "signals.jsonl")
	store, err := NewSignalStore(path)
	if err != nil {
		t.Fatalf("NewSignalStore failed: %v", err)
	}

	store.Record(Signal{
		Type:      SignalReviewFailure,
		Domain:    "billing",
		Details:   "Missing error handling",
		FilePaths: []string{"packs/billing/model.rb"},
	})

	all := store.GetAll()
	if len(all) != 1 {
		t.Fatalf("Expected 1 signal, got %d", len(all))
	}

	if all[0].Type != SignalReviewFailure {
		t.Errorf("Expected review_failure, got %s", all[0].Type)
	}
	if all[0].Count != 1 {
		t.Errorf("Expected count 1, got %d", all[0].Count)
	}
}

func TestSignalStoreDeduplication(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "signal-test")
	defer os.RemoveAll(tmpDir)

	path := filepath.Join(tmpDir, "signals.jsonl")
	store, _ := NewSignalStore(path)

	// Record same type+domain twice
	store.Record(Signal{
		Type:      SignalReviewFailure,
		Domain:    "billing",
		Details:   "First issue",
		FilePaths: []string{"file1.rb"},
	})
	store.Record(Signal{
		Type:      SignalReviewFailure,
		Domain:    "billing",
		Details:   "Second issue",
		FilePaths: []string{"file2.rb"},
	})

	all := store.GetAll()
	if len(all) != 1 {
		t.Fatalf("Expected 1 deduplicated signal, got %d", len(all))
	}
	if all[0].Count != 2 {
		t.Errorf("Expected count 2, got %d", all[0].Count)
	}
	if all[0].Details != "Second issue" {
		t.Errorf("Expected updated details, got %q", all[0].Details)
	}
	// File paths should be merged
	if len(all[0].FilePaths) != 2 {
		t.Errorf("Expected 2 merged file paths, got %d", len(all[0].FilePaths))
	}
}

func TestSignalStoreGetByDomain(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "signal-test")
	defer os.RemoveAll(tmpDir)

	path := filepath.Join(tmpDir, "signals.jsonl")
	store, _ := NewSignalStore(path)

	store.Record(Signal{Type: SignalReviewFailure, Domain: "billing"})
	store.Record(Signal{Type: SignalReviewFailure, Domain: "auth"})
	store.Record(Signal{Type: SignalRefactorLoop, Domain: "billing"})

	billing := store.GetByDomain("billing")
	if len(billing) != 2 {
		t.Errorf("Expected 2 billing signals, got %d", len(billing))
	}

	auth := store.GetByDomain("auth")
	if len(auth) != 1 {
		t.Errorf("Expected 1 auth signal, got %d", len(auth))
	}
}

func TestSignalStoreSaveAndReload(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "signal-test")
	defer os.RemoveAll(tmpDir)

	path := filepath.Join(tmpDir, "signals.jsonl")
	store, _ := NewSignalStore(path)

	store.Record(Signal{
		Type:      SignalReviewFailure,
		Domain:    "billing",
		Details:   "Test signal",
		FilePaths: []string{"file1.rb"},
	})
	store.Record(Signal{
		Type:   SignalRepeatedFileRead,
		Domain: "auth",
	})

	if err := store.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Reload from disk
	store2, err := NewSignalStore(path)
	if err != nil {
		t.Fatalf("Reload failed: %v", err)
	}

	all := store2.GetAll()
	if len(all) != 2 {
		t.Fatalf("Expected 2 signals after reload, got %d", len(all))
	}
	if all[0].Domain != "billing" {
		t.Errorf("Expected billing domain, got %q", all[0].Domain)
	}
}

func TestMergeStrings(t *testing.T) {
	a := []string{"a", "b"}
	b := []string{"b", "c"}
	result := mergeStrings(a, b)

	if len(result) != 3 {
		t.Errorf("Expected 3 unique strings, got %d", len(result))
	}
}
