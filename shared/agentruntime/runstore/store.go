// Package runstore persists normalized agent runtime events.
package runstore

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	agentruntime "github.com/philjestin/boatman-ecosystem/shared/agentruntime"
)

const (
	metadataFile  = "metadata.json"
	requestFile   = "request.json"
	eventsFile    = "events.ndjson"
	artifactsFile = "artifacts.json"
)

// RunMetadata is the durable index record for a runtime run.
type RunMetadata struct {
	RunID         string              `json:"runId"`
	Provider      string              `json:"provider,omitempty"`
	Model         string              `json:"model,omitempty"`
	Role          agentruntime.Role   `json:"role,omitempty"`
	Profile       string              `json:"profile,omitempty"`
	WorkDir       string              `json:"workDir,omitempty"`
	Status        agentruntime.Status `json:"status,omitempty"`
	StartedAt     time.Time           `json:"startedAt"`
	UpdatedAt     time.Time           `json:"updatedAt"`
	EndedAt       *time.Time          `json:"endedAt,omitempty"`
	EventCount    int                 `json:"eventCount"`
	ArtifactCount int                 `json:"artifactCount,omitempty"`
}

// ArtifactRecord is a compact index entry for durable outputs produced by a run.
type ArtifactRecord struct {
	Kind        string                 `json:"kind"`
	Path        string                 `json:"path,omitempty"`
	URL         string                 `json:"url,omitempty"`
	Diff        string                 `json:"diff,omitempty"`
	Message     string                 `json:"message,omitempty"`
	PhaseID     string                 `json:"phaseId,omitempty"`
	TaskID      string                 `json:"taskId,omitempty"`
	EventType   agentruntime.EventType `json:"eventType,omitempty"`
	FirstSeenAt time.Time              `json:"firstSeenAt"`
	LastSeenAt  time.Time              `json:"lastSeenAt"`
	EventCount  int                    `json:"eventCount"`
}

// Store persists run metadata and events.
type Store interface {
	StartRun(ctx context.Context, req agentruntime.RunRequest) error
	Append(ctx context.Context, event agentruntime.Event) error
	LoadRun(ctx context.Context, runID string) (RunMetadata, []agentruntime.Event, error)
	ListRuns(ctx context.Context) ([]RunMetadata, error)
}

// ForRequest returns the configured file store for a run request. Recording is
// enabled when request metadata contains runStoreDir, BOATMAN_RUNTIME_STORE_DIR
// is set, or BOATMAN_RUNTIME_STORE=1. With BOATMAN_RUNTIME_STORE=1 and no
// explicit directory, the store lives at <workdir>/.boatman/runs.
func ForRequest(req agentruntime.RunRequest) (*FileStore, bool, error) {
	if isFalsey(req.Metadata["runStore"]) {
		return nil, false, nil
	}
	dir := strings.TrimSpace(req.Metadata["runStoreDir"])
	if dir == "" {
		dir = strings.TrimSpace(os.Getenv("BOATMAN_RUNTIME_STORE_DIR"))
	}
	if dir == "" {
		if os.Getenv("BOATMAN_RUNTIME_STORE") != "1" {
			return nil, false, nil
		}
		defaultDir, err := DefaultDir(req.WorkDir)
		if err != nil {
			return nil, false, err
		}
		dir = defaultDir
	}
	return NewFileStore(dir), true, nil
}

func isFalsey(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "0", "false", "no", "off", "disabled":
		return true
	default:
		return false
	}
}

// DefaultDir returns the default runtime store directory for a workdir.
func DefaultDir(workDir string) (string, error) {
	workDir = strings.TrimSpace(workDir)
	if workDir == "" {
		var err error
		workDir, err = os.Getwd()
		if err != nil {
			return "", err
		}
	}
	return filepath.Join(workDir, ".boatman", "runs"), nil
}

// FileStore is a directory-backed run store.
type FileStore struct {
	root string
}

// NewFileStore creates a file-backed store rooted at root.
func NewFileStore(root string) *FileStore {
	return &FileStore{root: root}
}

// Root returns the configured store root.
func (s *FileStore) Root() string {
	return s.root
}

// StartRun initializes metadata for a run.
func (s *FileStore) StartRun(ctx context.Context, req agentruntime.RunRequest) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	runID := strings.TrimSpace(req.RunID)
	if runID == "" {
		runID = "run"
	}
	req.RunID = runID
	now := time.Now().UTC()
	metadata := RunMetadata{
		RunID:     runID,
		Provider:  req.Provider,
		Model:     req.Model,
		Role:      req.Role,
		Profile:   req.Profile,
		WorkDir:   req.WorkDir,
		Status:    agentruntime.StatusStarted,
		StartedAt: now,
		UpdatedAt: now,
	}
	if err := os.MkdirAll(s.runDir(runID), 0755); err != nil {
		return err
	}
	if err := s.writeRequest(runID, req); err != nil {
		return err
	}
	return s.writeMetadata(runID, metadata)
}

// Append writes a runtime event to the run's event stream and updates metadata.
func (s *FileStore) Append(ctx context.Context, event agentruntime.Event) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	runID := strings.TrimSpace(event.RunID)
	if runID == "" {
		runID = "run"
		event.RunID = runID
	}
	if event.Version == 0 {
		event.Version = agentruntime.ProtocolVersion
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}
	if err := os.MkdirAll(s.runDir(runID), 0755); err != nil {
		return err
	}
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(s.eventsPath(runID), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	if _, err := file.Write(append(data, '\n')); err != nil {
		file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	if err := s.recordArtifact(runID, event); err != nil {
		return err
	}
	return s.updateMetadata(runID, event)
}

// LoadRun reads metadata and events for a run.
func (s *FileStore) LoadRun(ctx context.Context, runID string) (RunMetadata, []agentruntime.Event, error) {
	if err := ctx.Err(); err != nil {
		return RunMetadata{}, nil, err
	}
	runID = strings.TrimSpace(runID)
	if runID == "" {
		return RunMetadata{}, nil, fmt.Errorf("run ID is required")
	}
	metadata, err := s.readMetadata(runID)
	if err != nil {
		return RunMetadata{}, nil, err
	}
	events, err := s.readEvents(runID)
	if err != nil {
		return RunMetadata{}, nil, err
	}
	return metadata, events, nil
}

// ListRuns returns known runs newest first.
func (s *FileStore) ListRuns(ctx context.Context) ([]RunMetadata, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(s.root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var runs []RunMetadata
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		metadata, err := s.readMetadata(entry.Name())
		if err == nil {
			runs = append(runs, metadata)
		}
	}
	sort.Slice(runs, func(i, j int) bool {
		return runs[i].UpdatedAt.After(runs[j].UpdatedAt)
	})
	return runs, nil
}

// LoadRequest reads the original provider-neutral run request for resume and
// replay tooling.
func (s *FileStore) LoadRequest(ctx context.Context, runID string) (agentruntime.RunRequest, error) {
	if err := ctx.Err(); err != nil {
		return agentruntime.RunRequest{}, err
	}
	runID = strings.TrimSpace(runID)
	if runID == "" {
		return agentruntime.RunRequest{}, fmt.Errorf("run ID is required")
	}
	data, err := os.ReadFile(s.requestPath(runID))
	if err != nil {
		return agentruntime.RunRequest{}, err
	}
	var req agentruntime.RunRequest
	if err := json.Unmarshal(data, &req); err != nil {
		return agentruntime.RunRequest{}, err
	}
	return req, nil
}

// ListArtifacts reads the compact artifact index for a run.
func (s *FileStore) ListArtifacts(ctx context.Context, runID string) ([]ArtifactRecord, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	runID = strings.TrimSpace(runID)
	if runID == "" {
		return nil, fmt.Errorf("run ID is required")
	}
	return s.readArtifacts(runID)
}

func (s *FileStore) updateMetadata(runID string, event agentruntime.Event) error {
	metadata, err := s.readMetadata(runID)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		metadata = RunMetadata{
			RunID:     runID,
			Provider:  event.Provider,
			Model:     event.Model,
			Role:      event.Role,
			Status:    agentruntime.StatusStarted,
			StartedAt: event.Timestamp,
		}
	}
	if metadata.Provider == "" {
		metadata.Provider = event.Provider
	}
	if metadata.Model == "" {
		metadata.Model = event.Model
	}
	if metadata.Role == "" {
		metadata.Role = event.Role
	}
	metadata.EventCount++
	metadata.UpdatedAt = event.Timestamp
	if terminalStatus(event) {
		status := event.Status
		if status == "" {
			status = statusForTerminalEvent(event.Type)
		}
		metadata.Status = status
		endedAt := event.Timestamp
		metadata.EndedAt = &endedAt
	} else if event.Status != "" {
		metadata.Status = event.Status
	}
	if metadata.StartedAt.IsZero() {
		metadata.StartedAt = event.Timestamp
	}
	if artifacts, err := s.readArtifacts(runID); err == nil {
		metadata.ArtifactCount = len(artifacts)
	}
	return s.writeMetadata(runID, metadata)
}

func terminalStatus(event agentruntime.Event) bool {
	switch event.Type {
	case agentruntime.EventRunCompleted, agentruntime.EventRunFailed:
		return true
	default:
		return false
	}
}

func statusForTerminalEvent(eventType agentruntime.EventType) agentruntime.Status {
	switch eventType {
	case agentruntime.EventRunCompleted:
		return agentruntime.StatusSucceeded
	case agentruntime.EventRunFailed:
		return agentruntime.StatusFailed
	default:
		return agentruntime.StatusCompleted
	}
}

func (s *FileStore) readMetadata(runID string) (RunMetadata, error) {
	data, err := os.ReadFile(s.metadataPath(runID))
	if err != nil {
		return RunMetadata{}, err
	}
	var metadata RunMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return RunMetadata{}, err
	}
	return metadata, nil
}

func (s *FileStore) writeMetadata(runID string, metadata RunMetadata) error {
	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.metadataPath(runID), append(data, '\n'), 0644)
}

func (s *FileStore) writeRequest(runID string, req agentruntime.RunRequest) error {
	data, err := json.MarshalIndent(req, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.requestPath(runID), append(data, '\n'), 0644)
}

func (s *FileStore) recordArtifact(runID string, event agentruntime.Event) error {
	if event.Artifact == nil {
		return nil
	}
	records, err := s.readArtifacts(runID)
	if err != nil {
		return err
	}
	key := artifactKey(event.Artifact)
	now := event.Timestamp
	if now.IsZero() {
		now = time.Now().UTC()
	}
	for i, record := range records {
		if artifactRecordKey(record) != key {
			continue
		}
		records[i].Diff = event.Artifact.Diff
		records[i].Message = event.Message
		records[i].PhaseID = event.PhaseID
		records[i].TaskID = event.TaskID
		records[i].EventType = event.Type
		records[i].LastSeenAt = now
		records[i].EventCount++
		return s.writeArtifacts(runID, records)
	}
	record := ArtifactRecord{
		Kind:        event.Artifact.Kind,
		Path:        event.Artifact.Path,
		URL:         event.Artifact.URL,
		Diff:        event.Artifact.Diff,
		Message:     event.Message,
		PhaseID:     event.PhaseID,
		TaskID:      event.TaskID,
		EventType:   event.Type,
		FirstSeenAt: now,
		LastSeenAt:  now,
		EventCount:  1,
	}
	records = append(records, record)
	sort.Slice(records, func(i, j int) bool {
		return artifactRecordKey(records[i]) < artifactRecordKey(records[j])
	})
	return s.writeArtifacts(runID, records)
}

func (s *FileStore) readArtifacts(runID string) ([]ArtifactRecord, error) {
	data, err := os.ReadFile(s.artifactsPath(runID))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var records []ArtifactRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return nil, err
	}
	return records, nil
}

func (s *FileStore) writeArtifacts(runID string, records []ArtifactRecord) error {
	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.artifactsPath(runID), append(data, '\n'), 0644)
}

func (s *FileStore) readEvents(runID string) ([]agentruntime.Event, error) {
	file, err := os.Open(s.eventsPath(runID))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer file.Close()

	var events []agentruntime.Event
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var event agentruntime.Event
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return events, nil
}

func (s *FileStore) runDir(runID string) string {
	return filepath.Join(s.root, safeRunID(runID))
}

func (s *FileStore) metadataPath(runID string) string {
	return filepath.Join(s.runDir(runID), metadataFile)
}

func (s *FileStore) requestPath(runID string) string {
	return filepath.Join(s.runDir(runID), requestFile)
}

func (s *FileStore) eventsPath(runID string) string {
	return filepath.Join(s.runDir(runID), eventsFile)
}

func (s *FileStore) artifactsPath(runID string) string {
	return filepath.Join(s.runDir(runID), artifactsFile)
}

func artifactKey(artifact *agentruntime.ArtifactEvent) string {
	if artifact == nil {
		return ""
	}
	return strings.Join([]string{artifact.Kind, artifact.Path, artifact.URL}, "\x00")
}

func artifactRecordKey(record ArtifactRecord) string {
	return strings.Join([]string{record.Kind, record.Path, record.URL}, "\x00")
}

func safeRunID(runID string) string {
	runID = strings.TrimSpace(runID)
	if runID == "" {
		return "run"
	}
	var b strings.Builder
	for _, r := range runID {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r)
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '-', r == '_', r == '.':
			b.WriteRune(r)
		default:
			b.WriteRune('_')
		}
	}
	if b.Len() == 0 {
		return "run"
	}
	return b.String()
}
