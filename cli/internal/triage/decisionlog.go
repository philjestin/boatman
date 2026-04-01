package triage

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

const decisionLogFile = "decision-log.ndjson"
const contextDocDir = "contexts"

// DecisionLog is an append-only NDJSON log for pipeline decisions.
type DecisionLog struct {
	dir string
	mu  sync.Mutex
}

// NewDecisionLog creates a DecisionLog rooted at dir, creating the directory
// if it does not already exist.
func NewDecisionLog(dir string) (*DecisionLog, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("creating decision log directory: %w", err)
	}
	return &DecisionLog{dir: dir}, nil
}

// Append marshals the entry as JSON and appends it as a single line to the
// decision-log.ndjson file.
func (dl *DecisionLog) Append(entry DecisionLogEntry) error {
	dl.mu.Lock()
	defer dl.mu.Unlock()

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("marshaling decision log entry: %w", err)
	}

	f, err := os.OpenFile(dl.logPath(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("opening decision log: %w", err)
	}
	defer f.Close()

	if _, err := f.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("writing decision log entry: %w", err)
	}

	return nil
}

// ReadAll reads and returns every entry from the decision log.
func (dl *DecisionLog) ReadAll() ([]DecisionLogEntry, error) {
	dl.mu.Lock()
	defer dl.mu.Unlock()

	return dl.readEntries(func(_ DecisionLogEntry) bool { return true })
}

// ReadForTicket reads all entries from the decision log that match the given ticketID.
func (dl *DecisionLog) ReadForTicket(ticketID string) ([]DecisionLogEntry, error) {
	dl.mu.Lock()
	defer dl.mu.Unlock()

	return dl.readEntries(func(e DecisionLogEntry) bool {
		return e.TicketID == ticketID
	})
}

// WriteContextDoc writes a ContextDoc to contexts/{clusterId}.json,
// creating the contexts subdirectory if needed.
func (dl *DecisionLog) WriteContextDoc(doc ContextDoc) error {
	dl.mu.Lock()
	defer dl.mu.Unlock()

	dir := filepath.Join(dl.dir, contextDocDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating context doc directory: %w", err)
	}

	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling context doc: %w", err)
	}

	path := filepath.Join(dir, doc.ClusterID+".json")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("writing context doc: %w", err)
	}

	return nil
}

// logPath returns the full path to the decision log file.
func (dl *DecisionLog) logPath() string {
	return filepath.Join(dl.dir, decisionLogFile)
}

// readEntries scans the log file and returns entries that pass the filter.
func (dl *DecisionLog) readEntries(filter func(DecisionLogEntry) bool) ([]DecisionLogEntry, error) {
	f, err := os.Open(dl.logPath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("opening decision log: %w", err)
	}
	defer f.Close()

	var entries []DecisionLogEntry
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var entry DecisionLogEntry
		if err := json.Unmarshal(line, &entry); err != nil {
			return entries, fmt.Errorf("parsing decision log line: %w", err)
		}

		if filter(entry) {
			entries = append(entries, entry)
		}
	}

	if err := scanner.Err(); err != nil {
		return entries, fmt.Errorf("scanning decision log: %w", err)
	}

	return entries, nil
}
