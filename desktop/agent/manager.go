package agent

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// AuthConfig holds authentication configuration
type AuthConfig struct {
	Method       string // "anthropic-api" or "google-cloud"
	APIKey       string
	GCPProjectID string
	GCPRegion    string
	ApprovalMode string // "suggest", "auto-edit", "full-auto"
}

// ConfigGetter retrieves memory management configuration
type ConfigGetter interface {
	GetMaxMessagesPerSession() int
	GetArchiveOldMessages() bool
	GetMaxSessionAgeDays() int
	GetMaxTotalSessions() int
	GetAutoCleanupSessions() bool
	GetMaxAgentsPerSession() int
	GetKeepCompletedAgents() bool
}

// Manager handles multiple agent sessions
type Manager struct {
	ctx              context.Context
	sessions         map[string]*Session
	mu               sync.RWMutex
	defaultModel     string
	authConfigGetter func() AuthConfig
	configGetter     ConfigGetter
}

// NewManager creates a new agent manager
func NewManager() *Manager {
	return &Manager{
		sessions:     make(map[string]*Session),
		defaultModel: "sonnet",
	}
}

// SetContext sets the Wails runtime context
func (m *Manager) SetContext(ctx context.Context) {
	m.ctx = ctx
}

// SetDefaultModel sets the default model for new sessions
func (m *Manager) SetDefaultModel(model string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.defaultModel = model
}

// SetAuthConfigGetter sets the function to retrieve auth configuration
func (m *Manager) SetAuthConfigGetter(getter func() AuthConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.authConfigGetter = getter
}

// SetAPIKeyGetter sets the function to retrieve the API key (deprecated, use SetAuthConfigGetter)
func (m *Manager) SetAPIKeyGetter(getter func() string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.authConfigGetter = func() AuthConfig {
		return AuthConfig{
			Method: "anthropic-api",
			APIKey: getter(),
		}
	}
}

// SetConfigGetter sets the config getter for memory management settings
func (m *Manager) SetConfigGetter(getter ConfigGetter) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.configGetter = getter
}

// GetConfigGetter returns the config getter
func (m *Manager) GetConfigGetter() ConfigGetter {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.configGetter
}

// CreateSession creates a new agent session for a project
func (m *Manager) CreateSession(projectPath string) (*Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	sessionID := uuid.New().String()
	session := NewSession(sessionID, projectPath)

	// Set up event handlers
	m.setupSessionHandlers(session, sessionID)

	// Set trim settings from config
	if m.configGetter != nil {
		maxMessages := m.configGetter.GetMaxMessagesPerSession()
		archive := m.configGetter.GetArchiveOldMessages()
		session.SetTrimSettings(maxMessages, archive)

		// Set agent cleanup settings
		maxAgents := m.configGetter.GetMaxAgentsPerSession()
		keepCompleted := m.configGetter.GetKeepCompletedAgents()
		session.SetAgentCleanupSettings(maxAgents, keepCompleted)
	}

	m.sessions[sessionID] = session
	return session, nil
}

// CreateFirefighterSession creates a new firefighter agent session
func (m *Manager) CreateFirefighterSession(projectPath string, scope string) (*Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	sessionID := uuid.New().String()
	session := NewSession(sessionID, projectPath)

	session.Mode = "firefighter"
	session.ModeConfig = map[string]interface{}{
		"scope": scope,
	}
	session.Tags = append(session.Tags, "firefighter")

	// Set up event handlers
	m.setupSessionHandlers(session, sessionID)

	// Set trim settings from config
	if m.configGetter != nil {
		maxMessages := m.configGetter.GetMaxMessagesPerSession()
		archive := m.configGetter.GetArchiveOldMessages()
		session.SetTrimSettings(maxMessages, archive)

		// Set agent cleanup settings
		maxAgents := m.configGetter.GetMaxAgentsPerSession()
		keepCompleted := m.configGetter.GetKeepCompletedAgents()
		session.SetAgentCleanupSettings(maxAgents, keepCompleted)
	}

	m.sessions[sessionID] = session
	return session, nil
}

// CreateBoatmanModeSession creates a new boatmanmode agent session
// mode can be "ticket" or "prompt"
func (m *Manager) CreateBoatmanModeSession(projectPath string, input string, mode string) (*Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	sessionID := uuid.New().String()
	session := NewSession(sessionID, projectPath)

	session.Mode = "boatmanmode"
	session.ModeConfig = map[string]interface{}{
		"input": input,
		"mode":  mode,
	}
	session.Tags = append(session.Tags, "boatmanmode")

	// Set up event handlers
	m.setupSessionHandlers(session, sessionID)

	// Set trim settings from config
	if m.configGetter != nil {
		maxMessages := m.configGetter.GetMaxMessagesPerSession()
		archive := m.configGetter.GetArchiveOldMessages()
		session.SetTrimSettings(maxMessages, archive)

		// Set agent cleanup settings
		maxAgents := m.configGetter.GetMaxAgentsPerSession()
		keepCompleted := m.configGetter.GetKeepCompletedAgents()
		session.SetAgentCleanupSettings(maxAgents, keepCompleted)
	}

	m.sessions[sessionID] = session
	return session, nil
}

// setupSessionHandlers sets up event handlers for a session
func (m *Manager) setupSessionHandlers(session *Session, sessionID string) {
	session.SetMessageHandler(func(msg Message) {
		if m.ctx != nil {
			runtime.EventsEmit(m.ctx, "agent:message", map[string]interface{}{
				"sessionId": sessionID,
				"message":   msg,
			})
		}
	})

	session.SetTaskHandler(func(task Task) {
		if m.ctx != nil {
			runtime.EventsEmit(m.ctx, "agent:task", map[string]interface{}{
				"sessionId": sessionID,
				"task":      task,
			})
		}
	})

	session.SetStatusHandler(func(status SessionStatus) {
		if m.ctx != nil {
			runtime.EventsEmit(m.ctx, "agent:status", map[string]interface{}{
				"sessionId": sessionID,
				"status":    status,
			})
		}
	})
}

// GetSession returns a session by ID
func (m *Manager) GetSession(sessionID string) (*Session, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	session, ok := m.sessions[sessionID]
	if !ok {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}
	return session, nil
}

// ListSessions returns all active sessions
func (m *Manager) ListSessions() []*Session {
	m.mu.RLock()
	defer m.mu.RUnlock()

	sessions := make([]*Session, 0, len(m.sessions))
	for _, s := range m.sessions {
		sessions = append(sessions, s)
	}
	return sessions
}

// StartSession starts an agent session
func (m *Manager) StartSession(sessionID string) error {
	session, err := m.GetSession(sessionID)
	if err != nil {
		return err
	}

	m.mu.RLock()
	model := m.defaultModel
	configGetter := m.configGetter
	m.mu.RUnlock()

	// Update trim settings in case they changed
	if configGetter != nil {
		maxMessages := configGetter.GetMaxMessagesPerSession()
		archive := configGetter.GetArchiveOldMessages()
		session.SetTrimSettings(maxMessages, archive)

		// Update agent cleanup settings
		maxAgents := configGetter.GetMaxAgentsPerSession()
		keepCompleted := configGetter.GetKeepCompletedAgents()
		session.SetAgentCleanupSettings(maxAgents, keepCompleted)
	}

	return session.Start(model)
}

// StopSession stops an agent session
func (m *Manager) StopSession(sessionID string) error {
	session, err := m.GetSession(sessionID)
	if err != nil {
		return err
	}
	return session.Stop()
}

// DeleteSession removes a session
func (m *Manager) DeleteSession(sessionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, ok := m.sessions[sessionID]
	if !ok {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	session.Stop()
	delete(m.sessions, sessionID)
	return nil
}

// SendMessage sends a message to a session
func (m *Manager) SendMessage(sessionID, content string) error {
	session, err := m.GetSession(sessionID)
	if err != nil {
		return err
	}

	// Get auth config
	var authConfig AuthConfig
	m.mu.RLock()
	if m.authConfigGetter != nil {
		authConfig = m.authConfigGetter()
	}
	m.mu.RUnlock()

	return session.SendMessage(content, authConfig)
}

// ApproveAction approves a pending action
func (m *Manager) ApproveAction(sessionID, actionID string) error {
	session, err := m.GetSession(sessionID)
	if err != nil {
		return err
	}
	return session.Approve(actionID)
}

// RejectAction rejects a pending action
func (m *Manager) RejectAction(sessionID, actionID string) error {
	session, err := m.GetSession(sessionID)
	if err != nil {
		return err
	}
	return session.Reject(actionID)
}

// GetSessionMessages returns messages for a session
func (m *Manager) GetSessionMessages(sessionID string) ([]Message, error) {
	session, err := m.GetSession(sessionID)
	if err != nil {
		return nil, err
	}
	return session.GetMessages(), nil
}

// GetSessionTasks returns tasks for a session
func (m *Manager) GetSessionTasks(sessionID string) ([]Task, error) {
	session, err := m.GetSession(sessionID)
	if err != nil {
		return nil, err
	}
	return session.GetTasks(), nil
}

// StopAllSessions stops all running sessions
func (m *Manager) StopAllSessions() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, session := range m.sessions {
		session.Stop()
	}
}

// CleanupSessions removes old sessions based on config settings
func (m *Manager) CleanupSessions() (int, error) {
	m.mu.RLock()
	configGetter := m.configGetter
	m.mu.RUnlock()

	if configGetter == nil || !configGetter.GetAutoCleanupSessions() {
		return 0, nil
	}

	maxAgeDays := configGetter.GetMaxSessionAgeDays()
	maxTotal := configGetter.GetMaxTotalSessions()

	return CleanupOldSessions(maxAgeDays, maxTotal)
}

// MarkAgentCompleted marks an agent as completed
func (m *Manager) MarkAgentCompleted(sessionID, agentID string) error {
	session, err := m.GetSession(sessionID)
	if err != nil {
		return err
	}
	session.MarkAgentCompleted(agentID)
	return nil
}

// AddTag adds a tag to a session
func (m *Manager) AddTag(sessionID, tag string) error {
	session, err := m.GetSession(sessionID)
	if err != nil {
		return err
	}
	session.AddTag(tag)
	return SaveSession(session)
}

// RemoveTag removes a tag from a session
func (m *Manager) RemoveTag(sessionID, tag string) error {
	session, err := m.GetSession(sessionID)
	if err != nil {
		return err
	}
	session.RemoveTag(tag)
	return SaveSession(session)
}

// SetFavorite sets the favorite status of a session
func (m *Manager) SetFavorite(sessionID string, favorite bool) error {
	session, err := m.GetSession(sessionID)
	if err != nil {
		return err
	}
	session.SetFavorite(favorite)
	return SaveSession(session)
}
