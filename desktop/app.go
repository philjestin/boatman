package main

import (
	"context"
	"fmt"
	"time"

	"boatman/agent"
	"boatman/auth"
	bmintegration "boatman/boatmanmode"
	"boatman/config"
	"boatman/diff"
	gitpkg "boatman/git"
	"boatman/mcp"
	"boatman/project"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct holds application state and dependencies
type App struct {
	ctx            context.Context
	config         *config.Config
	agentManager   *agent.Manager
	projectManager *project.ProjectManager
	mcpManager     *mcp.Manager
}

// NewApp creates a new App application struct
func NewApp() *App {
	cfg, err := config.NewConfig()
	if err != nil {
		panic(err)
	}

	pm, err := project.NewProjectManager()
	if err != nil {
		panic(err)
	}

	mcpMgr, err := mcp.NewManager()
	if err != nil {
		panic(err)
	}

	return &App{
		config:         cfg,
		agentManager:   agent.NewManager(),
		projectManager: pm,
		mcpManager:     mcpMgr,
	}
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.agentManager.SetContext(ctx)
	a.agentManager.SetAuthConfigGetter(func() agent.AuthConfig {
		prefs := a.config.GetPreferences()
		gcpProjectID, gcpRegion := a.config.GetGCPConfig()
		return agent.AuthConfig{
			Method:       string(prefs.AuthMethod),
			APIKey:       prefs.APIKey,
			GCPProjectID: gcpProjectID,
			GCPRegion:    gcpRegion,
			ApprovalMode: string(prefs.ApprovalMode),
		}
	})

	// Set config getter for memory management
	a.agentManager.SetConfigGetter(a)

	// Run session cleanup asynchronously on startup
	go func() {
		if count, err := a.agentManager.CleanupSessions(); err == nil && count > 0 {
			runtime.LogInfof(ctx, "Cleaned up %d old sessions", count)
		}
	}()
}

// shutdown is called when the app is closing
func (a *App) shutdown(ctx context.Context) {
	a.agentManager.StopAllSessions()
}

// =============================================================================
// Configuration Methods
// =============================================================================

// GetPreferences returns user preferences
func (a *App) GetPreferences() config.UserPreferences {
	return a.config.GetPreferences()
}

// SetPreferences updates user preferences
func (a *App) SetPreferences(prefs config.UserPreferences) error {
	return a.config.SetPreferences(prefs)
}

// IsOnboardingCompleted checks if onboarding is done
func (a *App) IsOnboardingCompleted() bool {
	return a.config.IsOnboardingCompleted()
}

// CompleteOnboarding marks onboarding as done
func (a *App) CompleteOnboarding() error {
	return a.config.CompleteOnboarding()
}

// =============================================================================
// Agent Session Methods
// =============================================================================

// AgentSessionInfo represents session info for the frontend
type AgentSessionInfo struct {
	ID          string              `json:"id"`
	ProjectPath string              `json:"projectPath"`
	Status      agent.SessionStatus `json:"status"`
	CreatedAt   string              `json:"createdAt"`
	Tags        []string            `json:"tags,omitempty"`
	IsFavorite  bool                `json:"isFavorite,omitempty"`
}

// CreateAgentSession creates a new agent session
func (a *App) CreateAgentSession(projectPath string) (*AgentSessionInfo, error) {
	session, err := a.agentManager.CreateSession(projectPath)
	if err != nil {
		return nil, err
	}

	return &AgentSessionInfo{
		ID:          session.ID,
		ProjectPath: session.ProjectPath,
		Status:      session.Status,
		CreatedAt:   session.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

// CreateFirefighterSession creates a new firefighter agent session
func (a *App) CreateFirefighterSession(projectPath string, scope string) (*AgentSessionInfo, error) {
	session, err := a.agentManager.CreateFirefighterSession(projectPath, scope)
	if err != nil {
		return nil, err
	}

	return &AgentSessionInfo{
		ID:          session.ID,
		ProjectPath: session.ProjectPath,
		Status:      session.Status,
		CreatedAt:   session.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		Tags:        session.Tags,
	}, nil
}

// CreateBoatmanModeSession creates a new boatmanmode agent session
// mode can be "ticket" or "prompt"
func (a *App) CreateBoatmanModeSession(projectPath string, input string, mode string) (*AgentSessionInfo, error) {
	session, err := a.agentManager.CreateBoatmanModeSession(projectPath, input, mode)
	if err != nil {
		return nil, err
	}

	return &AgentSessionInfo{
		ID:          session.ID,
		ProjectPath: session.ProjectPath,
		Status:      session.Status,
		CreatedAt:   session.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		Tags:        session.Tags,
	}, nil
}

// StartAgentSession starts an agent session
func (a *App) StartAgentSession(sessionID string) error {
	return a.agentManager.StartSession(sessionID)
}

// StopAgentSession stops an agent session
func (a *App) StopAgentSession(sessionID string) error {
	return a.agentManager.StopSession(sessionID)
}

// DeleteAgentSession deletes an agent session
func (a *App) DeleteAgentSession(sessionID string) error {
	return a.agentManager.DeleteSession(sessionID)
}

// SendAgentMessage sends a message to an agent session
func (a *App) SendAgentMessage(sessionID, content string) error {
	return a.agentManager.SendMessage(sessionID, content)
}

// ApproveAgentAction approves a pending action
func (a *App) ApproveAgentAction(sessionID, actionID string) error {
	return a.agentManager.ApproveAction(sessionID, actionID)
}

// RejectAgentAction rejects a pending action
func (a *App) RejectAgentAction(sessionID, actionID string) error {
	return a.agentManager.RejectAction(sessionID, actionID)
}

// GetAgentMessages returns messages for a session
func (a *App) GetAgentMessages(sessionID string) ([]agent.Message, error) {
	return a.agentManager.GetSessionMessages(sessionID)
}

// MessagePage represents a page of messages
type MessagePage struct {
	Messages []agent.Message `json:"messages"`
	Total    int             `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"pageSize"`
	HasMore  bool            `json:"hasMore"`
}

// GetAgentMessagesPaginated returns a paginated list of messages for a session
func (a *App) GetAgentMessagesPaginated(sessionID string, page, pageSize int) (*MessagePage, error) {
	allMessages, err := a.agentManager.GetSessionMessages(sessionID)
	if err != nil {
		return nil, err
	}

	total := len(allMessages)

	// Default page size
	if pageSize <= 0 {
		pageSize = 50
	}

	// Default to page 0
	if page < 0 {
		page = 0
	}

	// Calculate start and end indices
	start := page * pageSize
	if start >= total {
		// Page is beyond available messages
		return &MessagePage{
			Messages: []agent.Message{},
			Total:    total,
			Page:     page,
			PageSize: pageSize,
			HasMore:  false,
		}, nil
	}

	end := start + pageSize
	if end > total {
		end = total
	}

	messages := allMessages[start:end]
	hasMore := end < total

	return &MessagePage{
		Messages: messages,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		HasMore:  hasMore,
	}, nil
}

// GetAgentTasks returns tasks for a session
func (a *App) GetAgentTasks(sessionID string) ([]agent.Task, error) {
	return a.agentManager.GetSessionTasks(sessionID)
}

// ListAgentSessions returns all agent sessions
func (a *App) ListAgentSessions() []AgentSessionInfo {
	sessions := a.agentManager.ListSessions()
	infos := make([]AgentSessionInfo, len(sessions))
	for i, s := range sessions {
		infos[i] = AgentSessionInfo{
			ID:          s.ID,
			ProjectPath: s.ProjectPath,
			Status:      s.Status,
			CreatedAt:   s.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			Tags:        s.Tags,
			IsFavorite:  s.IsFavorite,
		}
	}
	return infos
}

// =============================================================================
// Project Methods
// =============================================================================

// OpenProject opens or creates a project
func (a *App) OpenProject(path string) (*project.Project, error) {
	return a.projectManager.AddProject(path)
}

// RemoveProject removes a project from recents
func (a *App) RemoveProject(id string) error {
	return a.projectManager.RemoveProject(id)
}

// GetProject returns a project by ID
func (a *App) GetProject(id string) (*project.Project, error) {
	return a.projectManager.GetProject(id)
}

// ListProjects returns all projects
func (a *App) ListProjects() []project.Project {
	return a.projectManager.ListProjects()
}

// GetRecentProjects returns recent projects
func (a *App) GetRecentProjects(limit int) []project.Project {
	return a.projectManager.GetRecentProjects(limit)
}

// SelectFolder opens a folder selection dialog
func (a *App) SelectFolder() (string, error) {
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Project Folder",
	})
}

// GetWorkspaceInfo returns information about a workspace
func (a *App) GetWorkspaceInfo(path string) (*project.WorkspaceInfo, error) {
	ws := project.NewWorkspace(path)
	return ws.GetInfo()
}

// =============================================================================
// Git Methods
// =============================================================================

// GitStatus represents git status for a project
type GitStatus struct {
	IsRepo    bool            `json:"isRepo"`
	Branch    string          `json:"branch"`
	Modified  []string        `json:"modified"`
	Added     []string        `json:"added"`
	Deleted   []string        `json:"deleted"`
	Untracked []string        `json:"untracked"`
}

// GetGitStatus returns git status for a project
func (a *App) GetGitStatus(projectPath string) (*GitStatus, error) {
	repo := gitpkg.NewRepository(projectPath)

	if !repo.IsGitRepo() {
		return &GitStatus{IsRepo: false}, nil
	}

	branch, err := repo.GetCurrentBranch()
	if err != nil {
		branch = "unknown"
	}

	status, err := repo.GetStatus()
	if err != nil {
		return nil, err
	}

	return &GitStatus{
		IsRepo:    true,
		Branch:    branch,
		Modified:  status.Modified,
		Added:     status.Added,
		Deleted:   status.Deleted,
		Untracked: status.Untracked,
	}, nil
}

// GetGitDiff returns diff for a file
func (a *App) GetGitDiff(projectPath, filePath string) (string, error) {
	repo := gitpkg.NewRepository(projectPath)
	return repo.GetDiff(filePath)
}

// =============================================================================
// Diff Methods
// =============================================================================

// ParseDiff parses a unified diff string
func (a *App) ParseDiff(diffText string) ([]diff.FileDiff, error) {
	return diff.ParseUnifiedDiff(diffText)
}

// GetSideBySideDiff generates side-by-side diff
func (a *App) GetSideBySideDiff(fileDiff diff.FileDiff) []diff.SideBySideLine {
	return diff.GenerateSideBySide(fileDiff)
}

// =============================================================================
// MCP Methods
// =============================================================================

// GetMCPServers returns configured MCP servers
func (a *App) GetMCPServers() ([]mcp.Server, error) {
	return a.mcpManager.GetServers()
}

// AddMCPServer adds a new MCP server
func (a *App) AddMCPServer(server mcp.Server) error {
	return a.mcpManager.AddServer(server)
}

// RemoveMCPServer removes an MCP server
func (a *App) RemoveMCPServer(name string) error {
	return a.mcpManager.RemoveServer(name)
}

// UpdateMCPServer updates an MCP server
func (a *App) UpdateMCPServer(server mcp.Server) error {
	return a.mcpManager.UpdateServer(server)
}

// GetMCPPresets returns preset MCP servers
func (a *App) GetMCPPresets() []mcp.Server {
	return mcp.GetPresetServers()
}

// =============================================================================
// Config Getter Implementation (for agent.ConfigGetter interface)
// =============================================================================

// GetMaxMessagesPerSession returns the max messages per session setting
func (a *App) GetMaxMessagesPerSession() int {
	prefs := a.config.GetPreferences()
	if prefs.MaxMessagesPerSession <= 0 {
		return 1000 // Default
	}
	return prefs.MaxMessagesPerSession
}

// GetArchiveOldMessages returns the archive old messages setting
func (a *App) GetArchiveOldMessages() bool {
	return a.config.GetPreferences().ArchiveOldMessages
}

// GetMaxSessionAgeDays returns the max session age in days
func (a *App) GetMaxSessionAgeDays() int {
	prefs := a.config.GetPreferences()
	if prefs.MaxSessionAgeDays <= 0 {
		return 30 // Default
	}
	return prefs.MaxSessionAgeDays
}

// GetMaxTotalSessions returns the max total sessions setting
func (a *App) GetMaxTotalSessions() int {
	prefs := a.config.GetPreferences()
	if prefs.MaxTotalSessions <= 0 {
		return 100 // Default
	}
	return prefs.MaxTotalSessions
}

// GetAutoCleanupSessions returns the auto cleanup sessions setting
func (a *App) GetAutoCleanupSessions() bool {
	return a.config.GetPreferences().AutoCleanupSessions
}

// GetMaxAgentsPerSession returns the max agents per session setting
func (a *App) GetMaxAgentsPerSession() int {
	prefs := a.config.GetPreferences()
	if prefs.MaxAgentsPerSession <= 0 {
		return 20 // Default
	}
	return prefs.MaxAgentsPerSession
}

// GetKeepCompletedAgents returns the keep completed agents setting
func (a *App) GetKeepCompletedAgents() bool {
	return a.config.GetPreferences().KeepCompletedAgents
}

// =============================================================================
// Session Cleanup Methods
// =============================================================================

// CleanupOldSessions manually triggers session cleanup
func (a *App) CleanupOldSessions() (int, error) {
	maxAgeDays := a.GetMaxSessionAgeDays()
	maxTotal := a.GetMaxTotalSessions()
	return agent.CleanupOldSessions(maxAgeDays, maxTotal)
}

// GetSessionStats returns statistics about all sessions
func (a *App) GetSessionStats() (map[string]interface{}, error) {
	stats, err := agent.GetSessionStats()
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"total":      stats.Total,
		"oldestDate": stats.OldestDate.Format("2006-01-02T15:04:05Z07:00"),
		"newestDate": stats.NewestDate.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

// =============================================================================
// Search and Organization Methods
// =============================================================================

// SearchSessionsRequest represents a search request
type SearchSessionsRequest struct {
	Query       string   `json:"query"`
	Tags        []string `json:"tags"`
	ProjectPath string   `json:"projectPath"`
	IsFavorite  *bool    `json:"isFavorite"`
	FromDate    string   `json:"fromDate"`
	ToDate      string   `json:"toDate"`
}

// SearchSessionsResponse represents a search response
type SearchSessionsResponse struct {
	SessionID    string              `json:"sessionId"`
	ProjectPath  string              `json:"projectPath"`
	CreatedAt    string              `json:"createdAt"`
	UpdatedAt    string              `json:"updatedAt"`
	Tags         []string            `json:"tags"`
	IsFavorite   bool                `json:"isFavorite"`
	MessageCount int                 `json:"messageCount"`
	Score        int                 `json:"score"`
	MatchReasons []string            `json:"matchReasons"`
	Status       agent.SessionStatus `json:"status"`
}

// SearchSessions searches sessions based on criteria
func (a *App) SearchSessions(req SearchSessionsRequest) ([]SearchSessionsResponse, error) {
	// Parse dates
	var fromDate, toDate time.Time
	var err error

	if req.FromDate != "" {
		fromDate, err = time.Parse("2006-01-02", req.FromDate)
		if err != nil {
			return nil, fmt.Errorf("invalid from date: %w", err)
		}
	}

	if req.ToDate != "" {
		toDate, err = time.Parse("2006-01-02", req.ToDate)
		if err != nil {
			return nil, fmt.Errorf("invalid to date: %w", err)
		}
	}

	// Create filter
	filter := agent.SearchFilter{
		Query:       req.Query,
		Tags:        req.Tags,
		ProjectPath: req.ProjectPath,
		IsFavorite:  req.IsFavorite,
		FromDate:    fromDate,
		ToDate:      toDate,
	}

	// Perform search
	results, err := agent.SearchSessions(filter)
	if err != nil {
		return nil, err
	}

	// Convert to response format
	response := make([]SearchSessionsResponse, len(results))
	for i, result := range results {
		response[i] = SearchSessionsResponse{
			SessionID:    result.Session.ID,
			ProjectPath:  result.Session.ProjectPath,
			CreatedAt:    result.Session.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:    result.Session.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
			Tags:         result.Session.GetTags(),
			IsFavorite:   result.Session.IsFavorite,
			MessageCount: len(result.Session.Messages),
			Score:        result.Score,
			MatchReasons: result.MatchReason,
			Status:       result.Session.Status,
		}
	}

	return response, nil
}

// AddSessionTag adds a tag to a session
func (a *App) AddSessionTag(sessionID, tag string) error {
	return a.agentManager.AddTag(sessionID, tag)
}

// RemoveSessionTag removes a tag from a session
func (a *App) RemoveSessionTag(sessionID, tag string) error {
	return a.agentManager.RemoveTag(sessionID, tag)
}

// SetSessionFavorite sets the favorite status of a session
func (a *App) SetSessionFavorite(sessionID string, favorite bool) error {
	return a.agentManager.SetFavorite(sessionID, favorite)
}

// GetAllTags returns all unique tags across all sessions
func (a *App) GetAllTags() ([]string, error) {
	return agent.GetAllTags()
}

// =============================================================================
// Firefighter Monitoring Methods
// =============================================================================

// StartFirefighterMonitoring enables active monitoring for a firefighter session
func (a *App) StartFirefighterMonitoring(sessionID string) error {
	session, err := a.agentManager.GetSession(sessionID)
	if err != nil {
		return err
	}
	return session.StartFirefighterMonitoring()
}

// StopFirefighterMonitoring disables active monitoring
func (a *App) StopFirefighterMonitoring(sessionID string) error {
	session, err := a.agentManager.GetSession(sessionID)
	if err != nil {
		return err
	}
	session.StopFirefighterMonitoring()
	return nil
}

// IsFirefighterMonitoringActive checks if monitoring is active
func (a *App) IsFirefighterMonitoringActive(sessionID string) (bool, error) {
	session, err := a.agentManager.GetSession(sessionID)
	if err != nil {
		return false, err
	}
	return session.IsFirefighterMonitoringActive(), nil
}

// InvestigateLinearTicket triggers investigation of a specific Linear ticket
func (a *App) InvestigateLinearTicket(sessionID, linearIssueID string) error {
	session, err := a.agentManager.GetSession(sessionID)
	if err != nil {
		return err
	}
	return session.InvestigateLinearTicket(linearIssueID)
}

// InvestigateSlackAlert triggers investigation of a Slack alert
func (a *App) InvestigateSlackAlert(sessionID, slackThreadID, alertMessage string) error {
	session, err := a.agentManager.GetSession(sessionID)
	if err != nil {
		return err
	}
	return session.InvestigateSlackAlert(slackThreadID, alertMessage)
}

// GetFirefighterMonitorStatus returns monitoring status
func (a *App) GetFirefighterMonitorStatus(sessionID string) (map[string]interface{}, error) {
	session, err := a.agentManager.GetSession(sessionID)
	if err != nil {
		return nil, err
	}
	return session.GetFirefighterMonitorStatus(), nil
}

// =============================================================================
// Google Cloud OAuth Authentication Methods
// =============================================================================

// IsGCloudInstalled checks if gcloud CLI is installed
func (a *App) IsGCloudInstalled() bool {
	gcloud := auth.NewGCloudAuth()
	return gcloud.IsInstalled()
}

// IsGCloudAuthenticated checks if user is authenticated with gcloud
func (a *App) IsGCloudAuthenticated() (bool, error) {
	gcloud := auth.NewGCloudAuth()
	return gcloud.IsAuthenticated()
}

// GetGCloudAuthInfo returns current authentication info
func (a *App) GetGCloudAuthInfo() (map[string]interface{}, error) {
	gcloud := auth.NewGCloudAuth()
	return gcloud.GetAuthInfo()
}

// GCloudLogin triggers OAuth login flow
func (a *App) GCloudLogin() error {
	gcloud := auth.NewGCloudAuth()
	return gcloud.Login()
}

// GCloudLoginApplicationDefault triggers application default OAuth login
func (a *App) GCloudLoginApplicationDefault() error {
	gcloud := auth.NewGCloudAuth()
	return gcloud.LoginApplicationDefault()
}

// GCloudSetProject sets the active GCP project
func (a *App) GCloudSetProject(projectID string) error {
	gcloud := auth.NewGCloudAuth()
	return gcloud.SetProject(projectID)
}

// GCloudGetAvailableProjects returns list of available GCP projects
func (a *App) GCloudGetAvailableProjects() ([]string, error) {
	gcloud := auth.NewGCloudAuth()
	return gcloud.GetAvailableProjects()
}

// GCloudVerifyVertexAIAccess verifies access to Vertex AI
func (a *App) GCloudVerifyVertexAIAccess(projectID, region string) error {
	gcloud := auth.NewGCloudAuth()
	return gcloud.VerifyVertexAIAccess(projectID, region)
}

// GCloudRevoke revokes authentication
func (a *App) GCloudRevoke() error {
	gcloud := auth.NewGCloudAuth()
	return gcloud.Revoke()
}

// =============================================================================
// Okta OAuth Methods
// =============================================================================

// OktaLogin initiates Okta OAuth flow
func (a *App) OktaLogin(domain, clientID, clientSecret string) error {
	okta := auth.NewOktaAuth(domain, clientID, clientSecret)
	// Request scopes for Datadog and Bugsnag access
	scopes := []string{"openid", "profile", "email", "offline_access"}
	return okta.Login(scopes)
}

// IsOktaAuthenticated checks if Okta OAuth is valid
func (a *App) IsOktaAuthenticated(domain, clientID, clientSecret string) bool {
	okta := auth.NewOktaAuth(domain, clientID, clientSecret)
	return okta.IsAuthenticated()
}

// GetOktaAccessToken returns current Okta access token
func (a *App) GetOktaAccessToken(domain, clientID, clientSecret string) (string, error) {
	okta := auth.NewOktaAuth(domain, clientID, clientSecret)
	return okta.GetAccessToken()
}

// OktaRefreshToken refreshes the Okta access token
func (a *App) OktaRefreshToken(domain, clientID, clientSecret string) error {
	okta := auth.NewOktaAuth(domain, clientID, clientSecret)
	return okta.RefreshToken()
}

// OktaRevoke revokes Okta authentication
func (a *App) OktaRevoke(domain, clientID, clientSecret string) error {
	okta := auth.NewOktaAuth(domain, clientID, clientSecret)
	return okta.Revoke()
}

// =============================================================================
// BoatmanMode Integration Methods
// =============================================================================

// ExecuteLinearTicketWithBoatmanMode runs the full boatmanmode workflow for a ticket
func (a *App) ExecuteLinearTicketWithBoatmanMode(linearAPIKey, ticketID, projectPath string) error {
	// Get Claude API key from config
	prefs := a.config.GetPreferences()
	claudeAPIKey := prefs.APIKey
	if claudeAPIKey == "" {
		return fmt.Errorf("Claude API key not configured")
	}

	// Create boatmanmode integration
	bmIntegration, err := bmintegration.NewIntegration(linearAPIKey, claudeAPIKey, projectPath)
	if err != nil {
		return fmt.Errorf("failed to create boatmanmode integration: %w", err)
	}

	// Execute the ticket (use app context for Wails events)
	_, err = bmIntegration.ExecuteTicket(a.ctx, ticketID)
	if err != nil {
		return fmt.Errorf("failed to execute ticket: %w", err)
	}

	return nil
}

// StreamBoatmanModeExecution runs boatmanmode workflow with streaming output
// mode can be "ticket" or "prompt"
// This function returns immediately and runs the execution in the background
func (a *App) StreamBoatmanModeExecution(sessionID, input, mode, linearAPIKey, projectPath string) error {
	// Get auth config using the same mechanism as regular sessions
	prefs := a.config.GetPreferences()
	claudeAPIKey := prefs.APIKey
	// Note: We don't error if Claude API key is missing - the boatman CLI will handle auth
	// It can use ~/.claude/config.json or other configured auth methods

	bmIntegration, err := bmintegration.NewIntegration(linearAPIKey, claudeAPIKey, projectPath)
	if err != nil {
		return fmt.Errorf("failed to create boatmanmode integration: %w", err)
	}

	// Emit started event
	runtime.EventsEmit(a.ctx, "boatmanmode:started", map[string]interface{}{
		"sessionId": sessionID,
	})

	// Run execution in background to avoid blocking the frontend
	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("[boatmanmode] Recovered from panic in execution: %v\n", r)
				// Try to emit error event to frontend (with panic recovery)
				func() {
					defer func() {
						if r2 := recover(); r2 != nil {
							fmt.Printf("[boatmanmode] Failed to emit panic error: %v\n", r2)
						}
					}()
					runtime.EventsEmit(a.ctx, "boatmanmode:error", map[string]interface{}{
						"sessionId": sessionID,
						"error":     fmt.Sprintf("Execution panic: %v", r),
					})
				}()
			}
		}()

		// Create output channel
		outputChan := make(chan string, 100)

		// Stream output to frontend via events
		go func() {
			defer func() {
				if r := recover(); r != nil {
					fmt.Printf("[boatmanmode] Recovered from panic in output emitter: %v\n", r)
				}
			}()

			for msg := range outputChan {
				func() {
					defer func() {
						if r := recover(); r != nil {
							fmt.Printf("[boatmanmode] Failed to emit output: %v\n", r)
						}
					}()
					runtime.EventsEmit(a.ctx, "boatmanmode:output", map[string]interface{}{
						"sessionId": sessionID,
						"input":     input,
						"mode":      mode,
						"message":   msg,
					})
				}()
			}
		}()

		// Execute with streaming (use app context for Wails events)
		_, err := bmIntegration.StreamExecution(a.ctx, sessionID, input, mode, outputChan)
		close(outputChan)

		if err != nil {
			fmt.Printf("[boatmanmode] Execution error: %v\n", err)
			// Emit error event to frontend
			runtime.EventsEmit(a.ctx, "boatmanmode:error", map[string]interface{}{
				"sessionId": sessionID,
				"error":     err.Error(),
			})
		} else {
			// Emit completion event to frontend
			runtime.EventsEmit(a.ctx, "boatmanmode:complete", map[string]interface{}{
				"sessionId": sessionID,
			})
		}
	}()

	return nil
}

// HandleBoatmanModeEvent processes boatmanmode events and updates session state
func (a *App) HandleBoatmanModeEvent(sessionID string, eventType string, eventData map[string]interface{}) error {
	session, err := a.agentManager.GetSession(sessionID)
	if err != nil {
		return err
	}

	switch eventType {
	case "agent_started":
		// Create task for the agent
		id, _ := eventData["id"].(string)
		name, _ := eventData["name"].(string)
		description, _ := eventData["description"].(string)
		session.AddOrUpdateTask(id, name, description, "in_progress")

	case "agent_completed":
		// Update task status and capture metadata
		id, _ := eventData["id"].(string)
		name, _ := eventData["name"].(string)
		status, _ := eventData["status"].(string)
		if status == "success" {
			status = "completed"
		} else if status == "failed" {
			status = "failed"
		}
		session.AddOrUpdateTask(id, name, "", status)

		// Store phase-specific metadata (diffs, feedback, etc.)
		metadata := make(map[string]interface{})
		if data, ok := eventData["data"].(map[string]interface{}); ok {
			// Capture diff if present
			if diff, ok := data["diff"].(string); ok {
				metadata["diff"] = diff
			}
			// Capture feedback if present
			if feedback, ok := data["feedback"].(string); ok {
				metadata["feedback"] = feedback
			}
			// Capture review issues if present
			if issues, ok := data["issues"].([]interface{}); ok {
				metadata["issues"] = issues
			}
			// Capture plan if present
			if plan, ok := data["plan"].(string); ok {
				metadata["plan"] = plan
			}
			// Capture refactor diff if present
			if refactorDiff, ok := data["refactor_diff"].(string); ok {
				metadata["refactor_diff"] = refactorDiff
			}
			// Store all other data fields
			for k, v := range data {
				if k != "diff" && k != "feedback" && k != "issues" && k != "plan" && k != "refactor_diff" {
					metadata[k] = v
				}
			}
		}
		if len(metadata) > 0 {
			session.UpdateTaskMetadata(id, metadata)
		}

	case "task_created":
		// Add task from boatmanmode's task system
		id, _ := eventData["id"].(string)
		subject, _ := eventData["name"].(string)
		description, _ := eventData["description"].(string)
		session.AddOrUpdateTask(id, subject, description, "pending")

	case "task_updated":
		// Update task status
		id, _ := eventData["id"].(string)
		name, _ := eventData["name"].(string)
		status, _ := eventData["status"].(string)
		session.AddOrUpdateTask(id, name, "", status)
	}

	return nil
}

// FetchLinearTicketsForBoatmanMode retrieves tickets from Linear
func (a *App) FetchLinearTicketsForBoatmanMode(linearAPIKey, projectPath string) ([]map[string]interface{}, error) {
	prefs := a.config.GetPreferences()
	claudeAPIKey := prefs.APIKey
	if claudeAPIKey == "" {
		return nil, fmt.Errorf("Claude API key not configured")
	}

	bmIntegration, err := bmintegration.NewIntegration(linearAPIKey, claudeAPIKey, projectPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create boatmanmode integration: %w", err)
	}

	// Fetch tickets with firefighter labels (use app context for Wails events)
	tickets, err := bmIntegration.FetchTickets(a.ctx, map[string]string{
		"labels": "firefighter,triage,boatmanmode",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tickets: %w", err)
	}

	// Tickets are already in the right format from boatmanmode CLI
	return tickets, nil
}

// =============================================================================
// Utility Methods
// =============================================================================

// CheckClaudeCLI checks if Claude CLI is installed
func (a *App) CheckClaudeCLI() bool {
	cli := agent.NewClaudeCLI()
	return cli.IsInstalled()
}

// GetClaudeCLIVersion returns the Claude CLI version
func (a *App) GetClaudeCLIVersion() (string, error) {
	cli := agent.NewClaudeCLI()
	return cli.GetVersion()
}

// SendNotification sends a desktop notification
func (a *App) SendNotification(title, message string) {
	runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
		Type:    runtime.InfoDialog,
		Title:   title,
		Message: message,
	})
}
