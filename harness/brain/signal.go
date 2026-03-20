package brain

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// SignalType identifies the kind of knowledge gap signal.
type SignalType string

const (
	SignalUserCorrection   SignalType = "user_correction"
	SignalRepeatedFileRead SignalType = "repeated_file_read"
	SignalErrorRetryLoop   SignalType = "error_retry_loop"
	SignalReviewFailure    SignalType = "review_failure"
	SignalRefactorLoop     SignalType = "refactor_loop"
)

// Signal represents a detected knowledge gap.
type Signal struct {
	Type      SignalType `json:"type"`
	Domain    string     `json:"domain"`
	Details   string     `json:"details"`
	FilePaths []string   `json:"file_paths,omitempty"`
	Count     int        `json:"count"`
	FirstSeen time.Time  `json:"first_seen"`
	LastSeen  time.Time  `json:"last_seen"`
}

// SignalStore is an append-only store for signals, persisted as JSONL.
type SignalStore struct {
	mu      sync.Mutex
	signals []Signal
	path    string
}

// NewSignalStore creates a new signal store at the given path.
func NewSignalStore(path string) (*SignalStore, error) {
	if path == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		path = filepath.Join(homeDir, ".boatman", "signals", "signals.jsonl")
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, fmt.Errorf("failed to create signals directory: %w", err)
	}

	store := &SignalStore{path: path}

	// Load existing signals
	if err := store.load(); err != nil {
		// Not fatal — start fresh
		store.signals = nil
	}

	return store, nil
}

// Record adds or updates a signal. If a matching signal exists (same type, domain),
// it increments the count and updates LastSeen.
func (s *SignalStore) Record(sig Signal) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()

	for i, existing := range s.signals {
		if existing.Type == sig.Type && existing.Domain == sig.Domain {
			s.signals[i].Count++
			s.signals[i].LastSeen = now
			if sig.Details != "" {
				s.signals[i].Details = sig.Details
			}
			// Merge file paths
			s.signals[i].FilePaths = mergeStrings(s.signals[i].FilePaths, sig.FilePaths)
			return
		}
	}

	// New signal
	if sig.Count == 0 {
		sig.Count = 1
	}
	if sig.FirstSeen.IsZero() {
		sig.FirstSeen = now
	}
	sig.LastSeen = now
	s.signals = append(s.signals, sig)
}

// GetByDomain returns all signals for a given domain.
func (s *SignalStore) GetByDomain(domain string) []Signal {
	s.mu.Lock()
	defer s.mu.Unlock()

	var result []Signal
	for _, sig := range s.signals {
		if sig.Domain == domain {
			result = append(result, sig)
		}
	}
	return result
}

// GetAll returns all stored signals.
func (s *SignalStore) GetAll() []Signal {
	s.mu.Lock()
	defer s.mu.Unlock()

	result := make([]Signal, len(s.signals))
	copy(result, s.signals)
	return result
}

// Save persists all signals to disk as JSONL.
func (s *SignalStore) Save() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	f, err := os.Create(s.path)
	if err != nil {
		return fmt.Errorf("failed to create signals file: %w", err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	for _, sig := range s.signals {
		if err := enc.Encode(sig); err != nil {
			return fmt.Errorf("failed to write signal: %w", err)
		}
	}

	return nil
}

// load reads existing signals from the JSONL file.
func (s *SignalStore) load() error {
	f, err := os.Open(s.path)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var sig Signal
		if err := json.Unmarshal(scanner.Bytes(), &sig); err != nil {
			continue // skip malformed lines
		}
		s.signals = append(s.signals, sig)
	}

	return scanner.Err()
}

// mergeStrings appends unique strings from b into a.
func mergeStrings(a, b []string) []string {
	seen := make(map[string]bool, len(a))
	for _, s := range a {
		seen[s] = true
	}
	for _, s := range b {
		if !seen[s] {
			a = append(a, s)
			seen[s] = true
		}
	}
	return a
}
