package brain

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestBrainJSONRoundTrip(t *testing.T) {
	b := &Brain{
		ID:          "test-brain",
		Name:        "Test Brain",
		Version:     1,
		Description: "A test brain",
		Triggers: Triggers{
			Keywords:     []string{"test", "example"},
			Entities:     []string{"TestEntity"},
			FilePatterns: []string{"pkg/test/"},
		},
		Confidence:  0.9,
		LastUpdated: "2025-01-01",
		Sections: []Section{
			{Title: "Overview", Content: "This is the overview."},
			{Title: "Rules", Content: "Rule 1\nRule 2"},
		},
		References: []Reference{
			{Path: "pkg/test/main.go", Description: "Main file", Checksum: "abc123"},
		},
	}

	data, err := json.Marshal(b)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var b2 Brain
	if err := json.Unmarshal(data, &b2); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if b2.ID != b.ID {
		t.Errorf("Expected ID %q, got %q", b.ID, b2.ID)
	}
	if b2.Name != b.Name {
		t.Errorf("Expected Name %q, got %q", b.Name, b2.Name)
	}
	if len(b2.Sections) != 2 {
		t.Errorf("Expected 2 sections, got %d", len(b2.Sections))
	}
	if len(b2.Triggers.Keywords) != 2 {
		t.Errorf("Expected 2 keywords, got %d", len(b2.Triggers.Keywords))
	}
}

func TestJSONReader(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "brain-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	b := &Brain{
		ID:          "json-test",
		Name:        "JSON Test",
		Version:     1,
		Description: "Testing JSON reader",
		Triggers:    Triggers{Keywords: []string{"json"}},
		Confidence:  0.8,
		Sections:    []Section{{Title: "Test", Content: "Content"}},
	}

	data, _ := json.MarshalIndent(b, "", "  ")
	path := filepath.Join(tmpDir, "test.json")
	os.WriteFile(path, data, 0644)

	reader := &JSONReader{}

	// Test extensions
	exts := reader.Extensions()
	if len(exts) != 1 || exts[0] != ".json" {
		t.Errorf("Expected [.json], got %v", exts)
	}

	// Test read
	loaded, err := reader.Read(path)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if loaded.ID != "json-test" {
		t.Errorf("Expected ID 'json-test', got %q", loaded.ID)
	}
	if loaded.Confidence != 0.8 {
		t.Errorf("Expected confidence 0.8, got %f", loaded.Confidence)
	}
}

func TestJSONReaderNoID(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "brain-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	path := filepath.Join(tmpDir, "noid.json")
	os.WriteFile(path, []byte(`{"name": "No ID"}`), 0644)

	reader := &JSONReader{}
	_, err = reader.Read(path)
	if err == nil {
		t.Error("Expected error for brain without ID")
	}
}

func TestJSONReaderInvalidJSON(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "brain-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	path := filepath.Join(tmpDir, "bad.json")
	os.WriteFile(path, []byte(`{not json`), 0644)

	reader := &JSONReader{}
	_, err = reader.Read(path)
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestJSONReaderMissingFile(t *testing.T) {
	reader := &JSONReader{}
	_, err := reader.Read("/nonexistent/file.json")
	if err == nil {
		t.Error("Expected error for missing file")
	}
}
