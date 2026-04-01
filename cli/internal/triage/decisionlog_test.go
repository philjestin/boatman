package triage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewDecisionLog(t *testing.T) {
	dir := t.TempDir()
	logDir := filepath.Join(dir, "logs")

	dl, err := NewDecisionLog(logDir)
	if err != nil {
		t.Fatalf("NewDecisionLog failed: %v", err)
	}
	if dl == nil {
		t.Fatal("expected non-nil DecisionLog")
	}

	// Directory should be created
	info, err := os.Stat(logDir)
	if err != nil {
		t.Fatalf("expected log directory to exist: %v", err)
	}
	if !info.IsDir() {
		t.Error("expected log path to be a directory")
	}
}

func TestDecisionLog_AppendAndReadAll(t *testing.T) {
	dl, err := NewDecisionLog(t.TempDir())
	if err != nil {
		t.Fatalf("NewDecisionLog failed: %v", err)
	}

	now := time.Now().UTC()
	entries := []DecisionLogEntry{
		{
			TicketID:  "ENG-1",
			Stage:     StageScore,
			Verdict:   "AI_DEFINITE",
			Agent:     "triage-pipeline",
			Rationale: "all gates passed",
			Timestamp: now,
		},
		{
			TicketID:   "ENG-2",
			Stage:      StageScore,
			Verdict:    "HUMAN_ONLY",
			Agent:      "triage-pipeline",
			Rationale:  "hard stop: payments",
			Timestamp:  now,
			TokensUsed: 5000,
			CostUSD:    0.01,
		},
	}

	for _, e := range entries {
		if err := dl.Append(e); err != nil {
			t.Fatalf("Append failed: %v", err)
		}
	}

	// Read all
	read, err := dl.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}

	if len(read) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(read))
	}
	if read[0].TicketID != "ENG-1" {
		t.Errorf("expected first entry ENG-1, got %s", read[0].TicketID)
	}
	if read[1].TicketID != "ENG-2" {
		t.Errorf("expected second entry ENG-2, got %s", read[1].TicketID)
	}
	if read[1].TokensUsed != 5000 {
		t.Errorf("expected 5000 tokens, got %d", read[1].TokensUsed)
	}
}

func TestDecisionLog_ReadForTicket(t *testing.T) {
	dl, err := NewDecisionLog(t.TempDir())
	if err != nil {
		t.Fatalf("NewDecisionLog failed: %v", err)
	}

	now := time.Now().UTC()
	dl.Append(DecisionLogEntry{TicketID: "ENG-1", Stage: StageIngest, Verdict: "ingested", Timestamp: now})
	dl.Append(DecisionLogEntry{TicketID: "ENG-2", Stage: StageScore, Verdict: "AI_LIKELY", Timestamp: now})
	dl.Append(DecisionLogEntry{TicketID: "ENG-1", Stage: StageScore, Verdict: "AI_DEFINITE", Timestamp: now})

	// Filter for ENG-1
	entries, err := dl.ReadForTicket("ENG-1")
	if err != nil {
		t.Fatalf("ReadForTicket failed: %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("expected 2 entries for ENG-1, got %d", len(entries))
	}
	for _, e := range entries {
		if e.TicketID != "ENG-1" {
			t.Errorf("expected all entries for ENG-1, got %s", e.TicketID)
		}
	}

	// Filter for non-existent ticket
	entries, err = dl.ReadForTicket("ENG-999")
	if err != nil {
		t.Fatalf("ReadForTicket failed: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries for ENG-999, got %d", len(entries))
	}
}

func TestDecisionLog_ReadAll_EmptyLog(t *testing.T) {
	dl, err := NewDecisionLog(t.TempDir())
	if err != nil {
		t.Fatalf("NewDecisionLog failed: %v", err)
	}

	entries, err := dl.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll on empty log failed: %v", err)
	}
	if entries != nil {
		t.Errorf("expected nil entries for empty log, got %v", entries)
	}
}

func TestDecisionLog_AppendWithDetails(t *testing.T) {
	dl, err := NewDecisionLog(t.TempDir())
	if err != nil {
		t.Fatalf("NewDecisionLog failed: %v", err)
	}

	details := map[string]interface{}{
		"clarity":    5,
		"hardStops":  []string{"payments"},
	}
	detailsJSON, _ := json.Marshal(details)

	entry := DecisionLogEntry{
		TicketID:  "ENG-10",
		Stage:     StageScore,
		Verdict:   "HUMAN_ONLY",
		Agent:     "triage-pipeline",
		Rationale: "payments hard stop",
		Timestamp: time.Now().UTC(),
		Details:   detailsJSON,
	}

	if err := dl.Append(entry); err != nil {
		t.Fatalf("Append failed: %v", err)
	}

	entries, err := dl.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Details == nil {
		t.Error("expected details to be preserved")
	}

	// Verify details can be unmarshaled
	var parsed map[string]interface{}
	if err := json.Unmarshal(entries[0].Details, &parsed); err != nil {
		t.Fatalf("failed to unmarshal details: %v", err)
	}
	if parsed["clarity"] != float64(5) {
		t.Errorf("expected clarity 5, got %v", parsed["clarity"])
	}
}

func TestDecisionLog_WriteContextDoc(t *testing.T) {
	dir := t.TempDir()
	dl, err := NewDecisionLog(dir)
	if err != nil {
		t.Fatalf("NewDecisionLog failed: %v", err)
	}

	doc := ContextDoc{
		ClusterID:      "cluster-frontend-1",
		Rationale:      "2 tickets grouped by frontend",
		TicketIDs:      []string{"ENG-1", "ENG-2"},
		RepoAreas:      []string{"next/packages/ui/"},
		KnownPatterns:  []string{"Follow existing React component patterns"},
		ValidationPlan: []string{"yarn test"},
		Risks:          []string{"blastRadius"},
		CostCeiling: CostCeiling{
			MaxTokensPerTicket:       500000,
			MaxAgentMinutesPerTicket: 30,
		},
	}

	if err := dl.WriteContextDoc(doc); err != nil {
		t.Fatalf("WriteContextDoc failed: %v", err)
	}

	// Verify file exists
	path := filepath.Join(dir, "contexts", "cluster-frontend-1.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("context doc file not found: %v", err)
	}

	// Verify content
	var parsed ContextDoc
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to parse context doc: %v", err)
	}
	if parsed.ClusterID != "cluster-frontend-1" {
		t.Errorf("expected cluster ID cluster-frontend-1, got %s", parsed.ClusterID)
	}
	if len(parsed.TicketIDs) != 2 {
		t.Errorf("expected 2 ticket IDs, got %d", len(parsed.TicketIDs))
	}
	if parsed.CostCeiling.MaxTokensPerTicket != 500000 {
		t.Errorf("expected 500K tokens, got %d", parsed.CostCeiling.MaxTokensPerTicket)
	}
}

func TestDecisionLog_NDJSONFormat(t *testing.T) {
	dir := t.TempDir()
	dl, err := NewDecisionLog(dir)
	if err != nil {
		t.Fatalf("NewDecisionLog failed: %v", err)
	}

	now := time.Now().UTC()
	dl.Append(DecisionLogEntry{TicketID: "ENG-1", Stage: StageScore, Verdict: "AI_DEFINITE", Timestamp: now})
	dl.Append(DecisionLogEntry{TicketID: "ENG-2", Stage: StageScore, Verdict: "AI_LIKELY", Timestamp: now})

	// Read raw file and verify NDJSON format (one JSON object per line)
	data, err := os.ReadFile(filepath.Join(dir, "decision-log.ndjson"))
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	lines := 0
	for _, b := range data {
		if b == '\n' {
			lines++
		}
	}
	if lines != 2 {
		t.Errorf("expected 2 newline-terminated lines, got %d", lines)
	}
}
