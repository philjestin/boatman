package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// SessionData represents the persistable data of a session
type SessionData struct {
	ID             string                `json:"id"`
	ProjectPath    string                `json:"projectPath"`
	Status         SessionStatus         `json:"status"`
	Messages       []Message             `json:"messages"`
	Tasks          []Task                `json:"tasks"`
	CreatedAt      string                `json:"createdAt"`
	UpdatedAt      string                `json:"updatedAt"`
	Model          string                `json:"model"`
	ConversationID string                `json:"conversationId"`
	CurrentAgentID string                `json:"currentAgentId"`
	Agents         map[string]*AgentInfo `json:"agents"`
	Tags           []string              `json:"tags,omitempty"`
	IsFavorite     bool                  `json:"isFavorite,omitempty"`
}

// SessionsDirGetter is a function type for getting sessions directory (for testing)
type SessionsDirGetter func() (string, error)

// defaultSessionsDirGetter is the default implementation
var defaultSessionsDirGetter SessionsDirGetter = getSessionsDir

// GetSessionsDir returns the directory where sessions are stored
func GetSessionsDir() (string, error) {
	return defaultSessionsDirGetter()
}

// getSessionsDir is the actual implementation
func getSessionsDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	sessionsDir := filepath.Join(homeDir, ".boatman", "sessions")

	// Create directory if it doesn't exist
	if err := os.MkdirAll(sessionsDir, 0755); err != nil {
		return "", err
	}

	return sessionsDir, nil
}

// SaveSession persists a session to disk
func SaveSession(session *Session) error {
	sessionsDir, err := GetSessionsDir()
	if err != nil {
		return fmt.Errorf("failed to get sessions directory: %w", err)
	}

	session.mu.RLock()
	defer session.mu.RUnlock()

	// Convert to persistable format
	data := SessionData{
		ID:             session.ID,
		ProjectPath:    session.ProjectPath,
		Status:         session.Status,
		Messages:       session.Messages,
		Tasks:          session.Tasks,
		CreatedAt:      session.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:      session.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		Model:          session.Model,
		ConversationID: session.conversationID,
		CurrentAgentID: session.currentAgentID,
		Agents:         session.agents,
		Tags:           session.Tags,
		IsFavorite:     session.IsFavorite,
	}

	// Marshal to JSON
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	// Write to file
	filename := filepath.Join(sessionsDir, session.ID+".json")
	if err := os.WriteFile(filename, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write session file: %w", err)
	}

	return nil
}

// LoadSession loads a session from disk
func LoadSession(sessionID string) (*Session, error) {
	sessionsDir, err := GetSessionsDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get sessions directory: %w", err)
	}

	filename := filepath.Join(sessionsDir, sessionID+".json")
	jsonData, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read session file: %w", err)
	}

	var data SessionData
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	// Create session from persisted data
	session := &Session{
		ID:             data.ID,
		ProjectPath:    data.ProjectPath,
		Status:         data.Status,
		Messages:       data.Messages,
		Tasks:          data.Tasks,
		Model:          data.Model,
		conversationID: data.ConversationID,
		currentAgentID: data.CurrentAgentID,
		agents:         data.Agents,
		Tags:           data.Tags,
		IsFavorite:     data.IsFavorite,
	}

	// Initialize tags if nil
	if session.Tags == nil {
		session.Tags = []string{}
	}

	// Initialize agents map if nil
	if session.agents == nil {
		session.agents = make(map[string]*AgentInfo)
		mainAgent := &AgentInfo{
			AgentID:   "main",
			AgentType: "main",
		}
		session.agents["main"] = mainAgent
		if session.currentAgentID == "" {
			session.currentAgentID = "main"
		}
	}

	// Parse timestamps
	createdAt, err := parseTimestamp(data.CreatedAt)
	if err == nil {
		session.CreatedAt = createdAt
	}

	updatedAt, err := parseTimestamp(data.UpdatedAt)
	if err == nil {
		session.UpdatedAt = updatedAt
	}

	// Initialize context for the session (required for sending messages)
	session.ctx, session.cancel = context.WithCancel(context.Background())

	// Set status to idle if it was stopped/error (make session usable again)
	if session.Status == SessionStatusStopped || session.Status == SessionStatusError {
		session.Status = SessionStatusIdle
	}

	return session, nil
}

// LoadAllSessions loads all persisted sessions
func LoadAllSessions() ([]*Session, error) {
	sessionsDir, err := GetSessionsDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get sessions directory: %w", err)
	}

	entries, err := os.ReadDir(sessionsDir)
	if err != nil {
		// Directory might not exist yet
		if os.IsNotExist(err) {
			return []*Session{}, nil
		}
		return nil, fmt.Errorf("failed to read sessions directory: %w", err)
	}

	var sessions []*Session
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		sessionID := entry.Name()[:len(entry.Name())-5] // Remove .json extension
		session, err := LoadSession(sessionID)
		if err != nil {
			fmt.Printf("Warning: failed to load session %s: %v\n", sessionID, err)
			continue
		}

		sessions = append(sessions, session)
	}

	return sessions, nil
}

// DeleteSessionFile removes a session file from disk
func DeleteSessionFile(sessionID string) error {
	sessionsDir, err := GetSessionsDir()
	if err != nil {
		return fmt.Errorf("failed to get sessions directory: %w", err)
	}

	filename := filepath.Join(sessionsDir, sessionID+".json")
	if err := os.Remove(filename); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete session file: %w", err)
	}

	return nil
}

// Helper function to parse timestamps
func parseTimestamp(s string) (time.Time, error) {
	return time.Parse("2006-01-02T15:04:05Z07:00", s)
}

// GetArchivesDir returns the directory where archived messages are stored
func GetArchivesDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	archivesDir := filepath.Join(homeDir, ".boatman", "sessions", "archives")

	// Create directory if it doesn't exist
	if err := os.MkdirAll(archivesDir, 0755); err != nil {
		return "", err
	}

	return archivesDir, nil
}

// ArchiveMessages appends messages to an archive file for a session
func ArchiveMessages(sessionID string, messages []Message) error {
	if len(messages) == 0 {
		return nil
	}

	archivesDir, err := GetArchivesDir()
	if err != nil {
		return fmt.Errorf("failed to get archives directory: %w", err)
	}

	filename := filepath.Join(archivesDir, sessionID+".json")

	// Load existing archive if it exists
	var existingMessages []Message
	if data, err := os.ReadFile(filename); err == nil {
		if err := json.Unmarshal(data, &existingMessages); err != nil {
			return fmt.Errorf("failed to unmarshal existing archive: %w", err)
		}
	}

	// Append new messages
	allMessages := append(existingMessages, messages...)

	// Write back to file
	jsonData, err := json.MarshalIndent(allMessages, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal archive: %w", err)
	}

	if err := os.WriteFile(filename, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write archive file: %w", err)
	}

	return nil
}

// GetArchivedMessageCount returns the number of archived messages for a session
func GetArchivedMessageCount(sessionID string) (int, error) {
	archivesDir, err := GetArchivesDir()
	if err != nil {
		return 0, err
	}

	filename := filepath.Join(archivesDir, sessionID+".json")
	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to read archive file: %w", err)
	}

	var messages []Message
	if err := json.Unmarshal(data, &messages); err != nil {
		return 0, fmt.Errorf("failed to unmarshal archive: %w", err)
	}

	return len(messages), nil
}

// DeleteArchiveFile removes an archive file from disk
func DeleteArchiveFile(sessionID string) error {
	archivesDir, err := GetArchivesDir()
	if err != nil {
		return fmt.Errorf("failed to get archives directory: %w", err)
	}

	filename := filepath.Join(archivesDir, sessionID+".json")
	if err := os.Remove(filename); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete archive file: %w", err)
	}

	return nil
}

// SessionStats contains statistics about all sessions
type SessionStats struct {
	Total       int       `json:"total"`
	OldestDate  time.Time `json:"oldestDate"`
	NewestDate  time.Time `json:"newestDate"`
}

// GetSessionStats returns statistics about all sessions
func GetSessionStats() (*SessionStats, error) {
	sessions, err := LoadAllSessions()
	if err != nil {
		return nil, err
	}

	if len(sessions) == 0 {
		return &SessionStats{
			Total: 0,
		}, nil
	}

	stats := &SessionStats{
		Total:      len(sessions),
		OldestDate: sessions[0].UpdatedAt,
		NewestDate: sessions[0].UpdatedAt,
	}

	for _, session := range sessions {
		if session.UpdatedAt.Before(stats.OldestDate) {
			stats.OldestDate = session.UpdatedAt
		}
		if session.UpdatedAt.After(stats.NewestDate) {
			stats.NewestDate = session.UpdatedAt
		}
	}

	return stats, nil
}

// CleanupOldSessions deletes sessions based on age and count limits
func CleanupOldSessions(maxAgeDays, maxTotal int) (int, error) {
	sessions, err := LoadAllSessions()
	if err != nil {
		return 0, err
	}

	if len(sessions) == 0 {
		return 0, nil
	}

	var toDelete []string
	now := time.Now()
	ageCutoff := now.AddDate(0, 0, -maxAgeDays)

	// First, mark sessions that are too old
	for _, session := range sessions {
		if session.UpdatedAt.Before(ageCutoff) {
			toDelete = append(toDelete, session.ID)
		}
	}

	// Sort remaining sessions by UpdatedAt (newest first)
	type sessionWithTime struct {
		id        string
		updatedAt time.Time
	}
	var remaining []sessionWithTime
	for _, session := range sessions {
		// Skip sessions already marked for deletion
		shouldSkip := false
		for _, id := range toDelete {
			if id == session.ID {
				shouldSkip = true
				break
			}
		}
		if !shouldSkip {
			remaining = append(remaining, sessionWithTime{
				id:        session.ID,
				updatedAt: session.UpdatedAt,
			})
		}
	}

	// Sort by updatedAt descending (newest first)
	for i := 0; i < len(remaining); i++ {
		for j := i + 1; j < len(remaining); j++ {
			if remaining[j].updatedAt.After(remaining[i].updatedAt) {
				remaining[i], remaining[j] = remaining[j], remaining[i]
			}
		}
	}

	// Mark oldest sessions for deletion if we exceed maxTotal
	if len(remaining) > maxTotal {
		for i := maxTotal; i < len(remaining); i++ {
			toDelete = append(toDelete, remaining[i].id)
		}
	}

	// Delete all marked sessions and their archives
	deletedCount := 0
	for _, sessionID := range toDelete {
		if err := DeleteSessionFile(sessionID); err != nil {
			fmt.Printf("Warning: failed to delete session %s: %v\n", sessionID, err)
			continue
		}
		// Also delete archive if it exists
		_ = DeleteArchiveFile(sessionID)
		deletedCount++
	}

	return deletedCount, nil
}
