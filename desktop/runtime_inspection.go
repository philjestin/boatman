package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	agentruntime "github.com/philjestin/boatman-ecosystem/shared/agentruntime"
	"github.com/philjestin/boatman-ecosystem/shared/agentruntime/memorydocs"
	"github.com/philjestin/boatman-ecosystem/shared/agentruntime/runstore"
)

// RuntimeRunSummary is the compact desktop-facing index record for a runtime run.
type RuntimeRunSummary struct {
	RunID         string `json:"runId"`
	Provider      string `json:"provider,omitempty"`
	Model         string `json:"model,omitempty"`
	Role          string `json:"role,omitempty"`
	Profile       string `json:"profile,omitempty"`
	WorkDir       string `json:"workDir,omitempty"`
	Status        string `json:"status,omitempty"`
	StartedAt     string `json:"startedAt,omitempty"`
	UpdatedAt     string `json:"updatedAt,omitempty"`
	EndedAt       string `json:"endedAt,omitempty"`
	EventCount    int    `json:"eventCount"`
	ArtifactCount int    `json:"artifactCount,omitempty"`
}

// RuntimeArtifactSummary is the desktop-facing artifact index record for one run.
type RuntimeArtifactSummary struct {
	Kind        string `json:"kind"`
	Path        string `json:"path,omitempty"`
	URL         string `json:"url,omitempty"`
	Diff        string `json:"diff,omitempty"`
	Message     string `json:"message,omitempty"`
	PhaseID     string `json:"phaseId,omitempty"`
	TaskID      string `json:"taskId,omitempty"`
	EventType   string `json:"eventType,omitempty"`
	FirstSeenAt string `json:"firstSeenAt,omitempty"`
	LastSeenAt  string `json:"lastSeenAt,omitempty"`
	EventCount  int    `json:"eventCount"`
}

// RuntimeEventSummary is a readable projection of one normalized runtime event.
type RuntimeEventSummary struct {
	Type         string `json:"type"`
	Status       string `json:"status,omitempty"`
	Role         string `json:"role,omitempty"`
	Provider     string `json:"provider,omitempty"`
	Model        string `json:"model,omitempty"`
	PhaseID      string `json:"phaseId,omitempty"`
	TaskID       string `json:"taskId,omitempty"`
	Name         string `json:"name,omitempty"`
	Message      string `json:"message,omitempty"`
	Timestamp    string `json:"timestamp,omitempty"`
	ToolName     string `json:"toolName,omitempty"`
	ToolError    bool   `json:"toolError,omitempty"`
	ArtifactKind string `json:"artifactKind,omitempty"`
	ArtifactPath string `json:"artifactPath,omitempty"`
	ArtifactURL  string `json:"artifactUrl,omitempty"`
	RawPreview   string `json:"rawPreview,omitempty"`
}

// RuntimeRunDetail contains the run index, event stream summary, and artifacts.
type RuntimeRunDetail struct {
	Metadata  RuntimeRunSummary        `json:"metadata"`
	Events    []RuntimeEventSummary    `json:"events"`
	Artifacts []RuntimeArtifactSummary `json:"artifacts"`
}

// MemoryDocumentSummary is the desktop-facing index record for inspectable memory.
type MemoryDocumentSummary struct {
	ID          string            `json:"id"`
	Scope       string            `json:"scope,omitempty"`
	Title       string            `json:"title,omitempty"`
	BodyPreview string            `json:"bodyPreview,omitempty"`
	Provenance  string            `json:"provenance,omitempty"`
	SourceRunID string            `json:"sourceRunId,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	GeneratedAt string            `json:"generatedAt,omitempty"`
	UpdatedAt   string            `json:"updatedAt,omitempty"`
	ExpiresAt   string            `json:"expiresAt,omitempty"`
	Expired     bool              `json:"expired,omitempty"`
	Path        string            `json:"path,omitempty"`
}

// MemoryDocumentDetail contains the full inspectable memory document.
type MemoryDocumentDetail struct {
	MemoryDocumentSummary
	Body string `json:"body,omitempty"`
}

// ListRuntimeRuns returns recorded runtime runs for a project.
func (a *App) ListRuntimeRuns(projectPath string) ([]RuntimeRunSummary, error) {
	store, err := a.runtimeStore(projectPath)
	if err != nil {
		return nil, err
	}
	runs, err := store.ListRuns(context.Background())
	if err != nil {
		return nil, err
	}
	summaries := make([]RuntimeRunSummary, 0, len(runs))
	for _, run := range runs {
		summaries = append(summaries, runtimeRunSummary(run))
	}
	return summaries, nil
}

// GetRuntimeRun returns a recorded runtime run with event and artifact details.
func (a *App) GetRuntimeRun(projectPath, runID string) (*RuntimeRunDetail, error) {
	store, err := a.runtimeStore(projectPath)
	if err != nil {
		return nil, err
	}
	metadata, events, err := store.LoadRun(context.Background(), runID)
	if err != nil {
		return nil, err
	}
	artifacts, err := store.ListArtifacts(context.Background(), runID)
	if err != nil {
		return nil, err
	}
	detail := &RuntimeRunDetail{
		Metadata:  runtimeRunSummary(metadata),
		Events:    make([]RuntimeEventSummary, 0, len(events)),
		Artifacts: make([]RuntimeArtifactSummary, 0, len(artifacts)),
	}
	for _, event := range events {
		detail.Events = append(detail.Events, runtimeEventSummary(event))
	}
	for _, artifact := range artifacts {
		detail.Artifacts = append(detail.Artifacts, runtimeArtifactSummary(artifact))
	}
	return detail, nil
}

// ListMemoryDocuments returns inspectable runtime memory documents for a project.
func (a *App) ListMemoryDocuments(projectPath string) ([]MemoryDocumentSummary, error) {
	store := a.memoryStore(projectPath)
	docs, err := store.List(context.Background())
	if err != nil {
		return nil, err
	}
	summaries := make([]MemoryDocumentSummary, 0, len(docs))
	now := time.Now()
	for _, doc := range docs {
		summaries = append(summaries, memoryDocumentSummary(doc, now))
	}
	return summaries, nil
}

// GetMemoryDocument returns one full memory document by ID.
func (a *App) GetMemoryDocument(projectPath, id string) (*MemoryDocumentDetail, error) {
	store := a.memoryStore(projectPath)
	doc, err := store.Read(context.Background(), id)
	if err != nil {
		return nil, err
	}
	summary := memoryDocumentSummary(doc, time.Now())
	return &MemoryDocumentDetail{
		MemoryDocumentSummary: summary,
		Body:                  doc.Body,
	}, nil
}

func (a *App) runtimeStore(projectPath string) (*runstore.FileStore, error) {
	dir := strings.TrimSpace(os.Getenv("BOATMAN_RUNTIME_STORE_DIR"))
	if dir == "" {
		var err error
		dir, err = runstore.DefaultDir(projectPath)
		if err != nil {
			return nil, err
		}
	}
	return runstore.NewFileStore(dir), nil
}

func (a *App) memoryStore(projectPath string) *memorydocs.FileStore {
	dir := strings.TrimSpace(os.Getenv("BOATMAN_MEMORY_DIR"))
	if dir == "" {
		dir = memorydocs.DefaultDir(projectPath)
	}
	return memorydocs.NewFileStore(dir)
}

func runtimeRunSummary(run runstore.RunMetadata) RuntimeRunSummary {
	endedAt := ""
	if run.EndedAt != nil {
		endedAt = formatRuntimeTime(*run.EndedAt)
	}
	return RuntimeRunSummary{
		RunID:         run.RunID,
		Provider:      run.Provider,
		Model:         run.Model,
		Role:          string(run.Role),
		Profile:       run.Profile,
		WorkDir:       run.WorkDir,
		Status:        string(run.Status),
		StartedAt:     formatRuntimeTime(run.StartedAt),
		UpdatedAt:     formatRuntimeTime(run.UpdatedAt),
		EndedAt:       endedAt,
		EventCount:    run.EventCount,
		ArtifactCount: run.ArtifactCount,
	}
}

func runtimeArtifactSummary(artifact runstore.ArtifactRecord) RuntimeArtifactSummary {
	return RuntimeArtifactSummary{
		Kind:        artifact.Kind,
		Path:        artifact.Path,
		URL:         artifact.URL,
		Diff:        artifact.Diff,
		Message:     artifact.Message,
		PhaseID:     artifact.PhaseID,
		TaskID:      artifact.TaskID,
		EventType:   string(artifact.EventType),
		FirstSeenAt: formatRuntimeTime(artifact.FirstSeenAt),
		LastSeenAt:  formatRuntimeTime(artifact.LastSeenAt),
		EventCount:  artifact.EventCount,
	}
}

func runtimeEventSummary(event agentruntime.Event) RuntimeEventSummary {
	summary := RuntimeEventSummary{
		Type:      string(event.Type),
		Status:    string(event.Status),
		Role:      string(event.Role),
		Provider:  event.Provider,
		Model:     event.Model,
		PhaseID:   event.PhaseID,
		TaskID:    event.TaskID,
		Name:      event.Name,
		Message:   event.Message,
		Timestamp: formatRuntimeTime(event.Timestamp),
	}
	if event.Tool != nil {
		summary.ToolName = event.Tool.Name
		summary.ToolError = event.Tool.IsError
	}
	if event.Artifact != nil {
		summary.ArtifactKind = event.Artifact.Kind
		summary.ArtifactPath = event.Artifact.Path
		summary.ArtifactURL = event.Artifact.URL
	}
	if len(event.Raw) > 0 {
		summary.RawPreview = truncateRuntimeString(string(event.Raw), 240)
	}
	return summary
}

func memoryDocumentSummary(doc memorydocs.Document, now time.Time) MemoryDocumentSummary {
	return MemoryDocumentSummary{
		ID:          doc.ID,
		Scope:       string(doc.Scope),
		Title:       doc.Title,
		BodyPreview: truncateRuntimeString(strings.ReplaceAll(doc.Body, "\n", " "), 180),
		Provenance:  doc.Provenance,
		SourceRunID: doc.SourceRunID,
		Metadata:    doc.Metadata,
		GeneratedAt: formatRuntimeTime(doc.GeneratedAt),
		UpdatedAt:   formatRuntimeTime(doc.UpdatedAt),
		ExpiresAt:   formatRuntimeOptionalTime(doc.ExpiresAt),
		Expired:     doc.IsExpired(now),
		Path:        filepath.Clean(doc.Path),
	}
}

func formatRuntimeOptionalTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return formatRuntimeTime(*t)
}

func formatRuntimeTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}

func truncateRuntimeString(value string, limit int) string {
	value = strings.TrimSpace(value)
	if limit <= 0 || len(value) <= limit {
		return value
	}
	if limit <= 3 {
		return value[:limit]
	}
	return strings.TrimSpace(value[:limit-3]) + "..."
}
