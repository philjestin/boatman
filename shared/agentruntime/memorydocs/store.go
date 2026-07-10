// Package memorydocs provides inspectable, provider-neutral memory documents
// for Boatman runtime sessions.
package memorydocs

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode"

	agentruntime "github.com/philjestin/boatman-ecosystem/shared/agentruntime"
)

// Scope describes the audience for a memory document.
type Scope string

const (
	ScopeUser        Scope = "user"
	ScopeProject     Scope = "project"
	ScopeTeam        Scope = "team"
	ScopeIntegration Scope = "integration"
	ScopeDomain      Scope = "domain"
)

// Document is a durable Markdown memory artifact.
type Document struct {
	ID          string            `json:"id"`
	Scope       Scope             `json:"scope,omitempty"`
	Title       string            `json:"title,omitempty"`
	Body        string            `json:"body,omitempty"`
	Provenance  string            `json:"provenance,omitempty"`
	SourceRunID string            `json:"sourceRunId,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	GeneratedAt time.Time         `json:"generatedAt,omitempty"`
	UpdatedAt   time.Time         `json:"updatedAt,omitempty"`
	ExpiresAt   *time.Time        `json:"expiresAt,omitempty"`
	Path        string            `json:"path,omitempty"`
}

// FileStore stores memory documents as Markdown files below one directory.
type FileStore struct {
	Dir string
}

// NewFileStore creates a file-backed memory document store.
func NewFileStore(dir string) *FileStore {
	return &FileStore{Dir: dir}
}

// DefaultDir returns the conventional per-project memory directory.
func DefaultDir(workDir string) string {
	if strings.TrimSpace(workDir) == "" {
		cwd, err := os.Getwd()
		if err == nil {
			workDir = cwd
		} else {
			workDir = "."
		}
	}
	return filepath.Join(workDir, ".boatman", "memory")
}

// NormalizeID validates and canonicalizes a memory document ID.
func NormalizeID(id string) (string, error) {
	id = filepath.ToSlash(strings.TrimSpace(id))
	if filepath.IsAbs(id) || strings.HasPrefix(id, "/") {
		return "", fmt.Errorf("invalid memory document ID %q", id)
	}
	id = strings.TrimSuffix(id, ".md")
	id = strings.Trim(id, "/")
	if id == "" {
		return "", errors.New("memory document ID is required")
	}
	if strings.HasPrefix(id, ".") || strings.Contains(id, "//") {
		return "", fmt.Errorf("invalid memory document ID %q", id)
	}
	clean := filepath.ToSlash(filepath.Clean(id))
	if clean == "." || clean != id {
		return "", fmt.Errorf("invalid memory document ID %q", id)
	}
	for _, part := range strings.Split(clean, "/") {
		if part == "" || part == "." || part == ".." {
			return "", fmt.Errorf("invalid memory document ID %q", id)
		}
		for _, r := range part {
			if unicode.IsLetter(r) || unicode.IsDigit(r) {
				continue
			}
			switch r {
			case '-', '_', '.':
				continue
			default:
				return "", fmt.Errorf("invalid memory document ID %q", id)
			}
		}
	}
	return clean, nil
}

// List returns all memory documents in deterministic ID order.
func (s *FileStore) List(ctx context.Context) ([]Document, error) {
	if err := checkContext(ctx); err != nil {
		return nil, err
	}
	if strings.TrimSpace(s.Dir) == "" {
		return nil, errors.New("memory directory is required")
	}
	if _, err := os.Stat(s.Dir); errors.Is(err, os.ErrNotExist) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	var docs []Document
	err := filepath.WalkDir(s.Dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if err := checkContext(ctx); err != nil {
			return err
		}
		if d.IsDir() || !strings.EqualFold(filepath.Ext(d.Name()), ".md") {
			return nil
		}
		doc, err := s.readPath(path)
		if err != nil {
			return err
		}
		docs = append(docs, doc)
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(docs, func(i, j int) bool {
		return docs[i].ID < docs[j].ID
	})
	return docs, nil
}

// Read loads one memory document by ID.
func (s *FileStore) Read(ctx context.Context, id string) (Document, error) {
	if err := checkContext(ctx); err != nil {
		return Document{}, err
	}
	path, err := s.pathForID(id)
	if err != nil {
		return Document{}, err
	}
	return s.readPath(path)
}

// Write writes one memory document as Markdown with stable frontmatter.
func (s *FileStore) Write(ctx context.Context, doc Document) (Document, error) {
	if err := checkContext(ctx); err != nil {
		return Document{}, err
	}
	id, err := NormalizeID(doc.ID)
	if err != nil {
		return Document{}, err
	}
	path, err := s.pathForID(id)
	if err != nil {
		return Document{}, err
	}
	now := time.Now().UTC()
	doc.ID = id
	doc.Path = path
	if doc.GeneratedAt.IsZero() {
		doc.GeneratedAt = now
	}
	if doc.UpdatedAt.IsZero() {
		doc.UpdatedAt = now
	}
	if strings.TrimSpace(string(doc.Scope)) == "" {
		doc.Scope = inferScope(id)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return Document{}, err
	}
	data := []byte(serialize(doc))
	if err := os.WriteFile(path, data, 0644); err != nil {
		return Document{}, err
	}
	return doc, nil
}

// LoadContext renders selected non-expired memory documents into prompt-ready
// context. Passing no IDs loads every document.
func (s *FileStore) LoadContext(ctx context.Context, ids []string, maxBytes int) (string, []Document, error) {
	if maxBytes < 0 {
		return "", nil, errors.New("maxBytes must be non-negative")
	}
	var docs []Document
	if len(ids) == 0 {
		all, err := s.List(ctx)
		if err != nil {
			return "", nil, err
		}
		docs = all
	} else {
		for _, id := range ids {
			doc, err := s.Read(ctx, id)
			if err != nil {
				return "", nil, err
			}
			docs = append(docs, doc)
		}
	}

	now := time.Now()
	active := make([]Document, 0, len(docs))
	var b strings.Builder
	for _, doc := range docs {
		if doc.IsExpired(now) {
			continue
		}
		active = append(active, doc)
		chunk := renderContextDocument(doc)
		if maxBytes > 0 && b.Len()+len(chunk) > maxBytes {
			remaining := maxBytes - b.Len()
			if remaining <= 0 {
				break
			}
			b.WriteString(chunk[:remaining])
			break
		}
		b.WriteString(chunk)
	}
	return strings.TrimSpace(b.String()), active, nil
}

// IsExpired reports whether a document has passed its expiration time.
func (d Document) IsExpired(now time.Time) bool {
	return d.ExpiresAt != nil && !d.ExpiresAt.IsZero() && now.After(*d.ExpiresAt)
}

// LoadedEvent creates a normalized memory.loaded event for runtime streams.
func LoadedEvent(runID string, docs []Document) agentruntime.Event {
	event := agentruntime.NewEvent(agentruntime.EventMemoryLoaded)
	event.RunID = runID
	event.Role = agentruntime.RoleMemory
	event.Status = agentruntime.StatusCompleted
	event.Message = fmt.Sprintf("loaded %d memory document(s)", len(docs))
	items := make([]map[string]any, 0, len(docs))
	for _, doc := range docs {
		item := map[string]any{
			"id":    doc.ID,
			"scope": string(doc.Scope),
			"title": doc.Title,
		}
		if doc.Path != "" {
			item["path"] = doc.Path
		}
		if doc.Provenance != "" {
			item["provenance"] = doc.Provenance
		}
		if doc.SourceRunID != "" {
			item["sourceRunId"] = doc.SourceRunID
		}
		if doc.ExpiresAt != nil {
			item["expiresAt"] = doc.ExpiresAt.Format(time.RFC3339)
		}
		items = append(items, item)
	}
	event.Data = map[string]any{"documents": items}
	return event
}

func (s *FileStore) pathForID(id string) (string, error) {
	if strings.TrimSpace(s.Dir) == "" {
		return "", errors.New("memory directory is required")
	}
	normalized, err := NormalizeID(id)
	if err != nil {
		return "", err
	}
	path := filepath.Join(s.Dir, filepath.FromSlash(normalized)+".md")
	cleanDir, err := filepath.Abs(s.Dir)
	if err != nil {
		return "", err
	}
	cleanPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	if cleanPath != cleanDir && !strings.HasPrefix(cleanPath, cleanDir+string(os.PathSeparator)) {
		return "", fmt.Errorf("memory document path escapes store: %s", normalized)
	}
	return path, nil
}

func (s *FileStore) readPath(path string) (Document, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Document{}, err
	}
	rel, err := filepath.Rel(s.Dir, path)
	if err != nil {
		return Document{}, err
	}
	id := strings.TrimSuffix(filepath.ToSlash(rel), ".md")
	doc, err := parse(id, string(data))
	if err != nil {
		return Document{}, fmt.Errorf("parse %s: %w", path, err)
	}
	doc.Path = path
	return doc, nil
}

func parse(id, text string) (Document, error) {
	normalized, err := NormalizeID(id)
	if err != nil {
		return Document{}, err
	}
	doc := Document{ID: normalized, Scope: inferScope(normalized)}
	text = strings.ReplaceAll(text, "\r\n", "\n")
	if strings.HasPrefix(text, "---\n") {
		rest := strings.TrimPrefix(text, "---\n")
		end := strings.Index(rest, "\n---")
		if end < 0 {
			return Document{}, errors.New("unterminated frontmatter")
		}
		metadata := rest[:end]
		body := rest[end+len("\n---"):]
		body = strings.TrimPrefix(body, "\n")
		if err := applyMetadata(&doc, metadata); err != nil {
			return Document{}, err
		}
		doc.Body = strings.TrimSpace(body)
	} else {
		doc.Body = strings.TrimSpace(text)
	}
	if doc.ID == "" {
		doc.ID = normalized
	}
	if doc.Scope == "" {
		doc.Scope = inferScope(doc.ID)
	}
	return doc, nil
}

func applyMetadata(doc *Document, text string) error {
	metadata := make(map[string]string)
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			return fmt.Errorf("invalid frontmatter line %q", line)
		}
		key = normalizeMetadataKey(key)
		value = strings.TrimSpace(value)
		switch key {
		case "id":
			id, err := NormalizeID(value)
			if err != nil {
				return err
			}
			doc.ID = id
		case "scope":
			doc.Scope = Scope(value)
		case "title":
			doc.Title = value
		case "provenance":
			doc.Provenance = value
		case "source_run_id":
			doc.SourceRunID = value
		case "generated_at":
			t, err := parseTime(value)
			if err != nil {
				return fmt.Errorf("generated_at: %w", err)
			}
			doc.GeneratedAt = t
		case "updated_at":
			t, err := parseTime(value)
			if err != nil {
				return fmt.Errorf("updated_at: %w", err)
			}
			doc.UpdatedAt = t
		case "expires_at":
			t, err := parseTime(value)
			if err != nil {
				return fmt.Errorf("expires_at: %w", err)
			}
			doc.ExpiresAt = &t
		default:
			if doc.Metadata == nil {
				doc.Metadata = make(map[string]string)
			}
			metadata[key] = value
		}
	}
	for key, value := range metadata {
		doc.Metadata[key] = value
	}
	return nil
}

func serialize(doc Document) string {
	var b strings.Builder
	b.WriteString("---\n")
	writeMetadata(&b, "id", doc.ID)
	writeMetadata(&b, "scope", string(doc.Scope))
	writeMetadata(&b, "title", doc.Title)
	writeMetadata(&b, "provenance", doc.Provenance)
	writeMetadata(&b, "source_run_id", doc.SourceRunID)
	if !doc.GeneratedAt.IsZero() {
		writeMetadata(&b, "generated_at", doc.GeneratedAt.UTC().Format(time.RFC3339))
	}
	if !doc.UpdatedAt.IsZero() {
		writeMetadata(&b, "updated_at", doc.UpdatedAt.UTC().Format(time.RFC3339))
	}
	if doc.ExpiresAt != nil && !doc.ExpiresAt.IsZero() {
		writeMetadata(&b, "expires_at", doc.ExpiresAt.UTC().Format(time.RFC3339))
	}
	if len(doc.Metadata) > 0 {
		keys := make([]string, 0, len(doc.Metadata))
		for key := range doc.Metadata {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			writeMetadata(&b, normalizeMetadataKey(key), doc.Metadata[key])
		}
	}
	b.WriteString("---\n\n")
	b.WriteString(strings.TrimSpace(doc.Body))
	b.WriteString("\n")
	return b.String()
}

func writeMetadata(b *strings.Builder, key, value string) {
	value = strings.TrimSpace(strings.ReplaceAll(value, "\n", " "))
	if value == "" {
		return
	}
	fmt.Fprintf(b, "%s: %s\n", key, value)
}

func parseTime(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, nil
	}
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return t, nil
	}
	if t, err := time.Parse("2006-01-02", value); err == nil {
		return t, nil
	}
	return time.Time{}, fmt.Errorf("invalid time %q", value)
}

func inferScope(id string) Scope {
	switch {
	case id == "user":
		return ScopeUser
	case id == "project":
		return ScopeProject
	case id == "team":
		return ScopeTeam
	case strings.HasPrefix(id, "integrations/"):
		return ScopeIntegration
	case strings.HasPrefix(id, "domains/"):
		return ScopeDomain
	default:
		return ScopeProject
	}
}

func renderContextDocument(doc Document) string {
	title := strings.TrimSpace(doc.Title)
	if title == "" {
		title = doc.ID
	}
	var b strings.Builder
	fmt.Fprintf(&b, "## %s (%s)\n", title, doc.ID)
	if doc.Scope != "" {
		fmt.Fprintf(&b, "Scope: %s\n", doc.Scope)
	}
	if doc.Provenance != "" {
		fmt.Fprintf(&b, "Provenance: %s\n", doc.Provenance)
	}
	if doc.SourceRunID != "" {
		fmt.Fprintf(&b, "Source run: %s\n", doc.SourceRunID)
	}
	if doc.ExpiresAt != nil && !doc.ExpiresAt.IsZero() {
		fmt.Fprintf(&b, "Expires: %s\n", doc.ExpiresAt.Format(time.RFC3339))
	}
	b.WriteString("\n")
	b.WriteString(strings.TrimSpace(doc.Body))
	b.WriteString("\n\n")
	return b.String()
}

func normalizeMetadataKey(key string) string {
	key = strings.ToLower(strings.TrimSpace(key))
	key = strings.ReplaceAll(key, "-", "_")
	key = strings.ReplaceAll(key, " ", "_")
	return key
}

func checkContext(ctx context.Context) error {
	if ctx == nil {
		return nil
	}
	return ctx.Err()
}
