package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"boatman/agent"
	"boatman/auth"
	bmintegration "boatman/boatmanmode"
	triageintegration "boatman/triage"
	"boatman/config"
	"boatman/diff"
	gitpkg "boatman/git"
	"boatman/harnessui"
	"boatman/mcp"
	"boatman/project"
	"boatman/services"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct holds application state and dependencies
type App struct {
	ctx            context.Context
	config         *config.Config
	agentManager   *agent.Manager
	projectManager *project.ProjectManager
	mcpManager     *mcp.Manager
	brainService   *services.BrainService
	harnessRuns    map[string]context.CancelFunc
	harnessMu      sync.Mutex
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
		brainService:   services.NewBrainService(),
		harnessRuns:    make(map[string]context.CancelFunc),
	}
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.agentManager.SetContext(ctx)
	a.agentManager.SetWailsReady()
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

	// Initialize brain service
	a.brainService.SetContext(ctx)

	// Load persisted sessions from disk
	if err := a.agentManager.LoadPersistedSessions(); err != nil {
		fmt.Printf("Warning: failed to load persisted sessions: %v\n", err)
	}

	// Run session cleanup asynchronously on startup
	go func() {
		if count, err := a.agentManager.CleanupSessions(); err == nil && count > 0 {
			runtime.LogInfof(ctx, "Cleaned up %d old sessions", count)
		}
	}()

	// Run periodic brain auto-distillation on startup
	go func() {
		if !services.ShouldAutoDistill() {
			return
		}
		results, err := a.brainService.AutoDistillBrains()
		if err != nil {
			runtime.LogWarningf(ctx, "Auto-distillation failed: %v", err)
			return
		}
		if len(results) > 0 {
			runtime.LogInfof(ctx, "Auto-generated %d brain(s) from signals", len(results))
		}
	}()
}

// shutdown is called when the app is closing
func (a *App) shutdown(ctx context.Context) {
	// Save all sessions BEFORE stopping them, so we persist current status (not "stopped")
	a.agentManager.SaveAllSessions()
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
	ID              string              `json:"id"`
	ProjectPath     string              `json:"projectPath"`
	Status          agent.SessionStatus `json:"status"`
	CreatedAt       string              `json:"createdAt"`
	Tags            []string            `json:"tags,omitempty"`
	IsFavorite      bool                `json:"isFavorite,omitempty"`
	Model           string              `json:"model,omitempty"`
	ReasoningEffort string              `json:"reasoningEffort,omitempty"`
	Mode            string              `json:"mode,omitempty"`
}

// CreateAgentSession creates a new agent session
func (a *App) CreateAgentSession(projectPath string) (*AgentSessionInfo, error) {
	session, err := a.agentManager.CreateSession(projectPath)
	if err != nil {
		return nil, err
	}

	return &AgentSessionInfo{
		ID:              session.ID,
		ProjectPath:     session.ProjectPath,
		Status:          session.Status,
		CreatedAt:       session.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		Model:           session.Model,
		ReasoningEffort: session.ReasoningEffort,
		Mode:            session.Mode,
	}, nil
}

// CreateFirefighterSession creates a new firefighter agent session
func (a *App) CreateFirefighterSession(projectPath string, scope string, slackChannels string) (*AgentSessionInfo, error) {
	session, err := a.agentManager.CreateFirefighterSession(projectPath, scope, slackChannels)
	if err != nil {
		return nil, err
	}

	return &AgentSessionInfo{
		ID:              session.ID,
		ProjectPath:     session.ProjectPath,
		Status:          session.Status,
		CreatedAt:       session.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		Tags:            session.Tags,
		Model:           session.Model,
		ReasoningEffort: session.ReasoningEffort,
		Mode:            session.Mode,
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
		ID:              session.ID,
		ProjectPath:     session.ProjectPath,
		Status:          session.Status,
		CreatedAt:       session.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		Tags:            session.Tags,
		Model:           session.Model,
		ReasoningEffort: session.ReasoningEffort,
		Mode:            session.Mode,
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
			ID:              s.ID,
			ProjectPath:     s.ProjectPath,
			Status:          s.Status,
			CreatedAt:       s.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			Tags:            s.Tags,
			IsFavorite:      s.IsFavorite,
			Model:           s.Model,
			ReasoningEffort: s.ReasoningEffort,
			Mode:            s.Mode,
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

// GetWorktreeDiff returns the diff of all changes on the worktree branch relative to the base branch.
// This is used for boatman mode sessions where changes are committed in a worktree.
func (a *App) GetWorktreeDiff(worktreePath, baseBranch string) (string, error) {
	repo := gitpkg.NewRepository(worktreePath)
	return repo.GetDiffAgainstBase(baseBranch)
}

// GetBoatmanModeSessionConfig returns the worktree path and base branch for a boatman mode session.
func (a *App) GetBoatmanModeSessionConfig(sessionID string) (map[string]interface{}, error) {
	session, err := a.agentManager.GetSession(sessionID)
	if err != nil {
		return nil, err
	}
	return session.ModeConfig, nil
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

// GetMCPServerNames returns the names of configured MCP servers
func (a *App) GetMCPServerNames() []string {
	servers := a.config.GetMCPServers()
	names := make([]string, len(servers))
	for i, s := range servers {
		names[i] = s.Name
	}
	return names
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

// SetSessionModel updates the model for a session
func (a *App) SetSessionModel(sessionID, model string) error {
	return a.agentManager.SetSessionModel(sessionID, model)
}

// SetSessionReasoningEffort updates the reasoning effort for a session
func (a *App) SetSessionReasoningEffort(sessionID, effort string) error {
	return a.agentManager.SetSessionReasoningEffort(sessionID, effort)
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

	// Get the session so we can route messages through it
	session, sessionErr := a.agentManager.GetSession(sessionID)

	// Write triage plan to temp file if one is attached to the session.
	var planFile string
	if sessionErr == nil && session != nil {
		if triagePlan, ok := session.ModeConfig["triagePlan"]; ok && triagePlan != nil {
			planJSON, err := json.Marshal(triagePlan)
			if err == nil {
				tmpFile, err := os.CreateTemp("", "boatman-plan-*.json")
				if err == nil {
					tmpFile.Write(planJSON)
					tmpFile.Close()
					planFile = tmpFile.Name()
					fmt.Printf("[boatmanmode] Using triage plan: %s\n", planFile)
				}
			}
		}
	}

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

		// Create message callback to route messages through the session
		var onMessage bmintegration.MessageCallback
		if sessionErr == nil && session != nil {
			// Persistent stream state for processing Claude stream lines
			var streamBuilder strings.Builder
			var streamMsgID string

			onMessage = func(role, content string) {
				if role == "claude_stream" {
					// Route raw Claude stream-json line through the session parser
					session.ProcessExternalStreamLine(content, &streamBuilder, &streamMsgID)
				} else {
					session.AddBoatmanMessage(role, content)
				}
			}
		}

		// Execute with streaming (use app context for Wails events)
		// Check if this session is marked for resume
		isResume := false
		if session != nil {
			if r, ok := session.ModeConfig["resume"]; ok {
				isResume, _ = r.(bool)
			}
		}

		_, err := bmIntegration.StreamExecution(a.ctx, sessionID, input, mode, planFile, isResume, outputChan, onMessage)
		close(outputChan)

		// Clean up temp plan file.
		if planFile != "" {
			os.Remove(planFile)
		}

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

		// Persist session after boatmanmode execution completes
		if sessionErr == nil && session != nil {
			if saveErr := agent.SaveSession(session); saveErr != nil {
				fmt.Printf("[boatmanmode] Warning: failed to save session %s: %v\n", sessionID, saveErr)
			}
		}
	}()

	return nil
}

// ResumeBoatmanModeExecution resumes a failed boatmanmode session from the review/refactor stage.
// It sets the "resume" flag on the session's ModeConfig and re-runs StreamBoatmanModeExecution.
func (a *App) ResumeBoatmanModeExecution(sessionID string) error {
	session, err := a.agentManager.GetSession(sessionID)
	if err != nil {
		return fmt.Errorf("session not found: %w", err)
	}

	// Mark session for resume
	session.SetModeConfigValue("resume", true)

	// Extract the original input and mode from the session
	input, _ := session.ModeConfig["input"].(string)
	mode, _ := session.ModeConfig["mode"].(string)
	if input == "" {
		return fmt.Errorf("session has no input to resume")
	}
	if mode == "" {
		mode = "ticket"
	}

	// Get preferences for API keys
	prefs := a.config.GetPreferences()
	linearAPIKey := prefs.LinearAPIKey

	// Re-run execution with resume flag (StreamBoatmanModeExecution reads ModeConfig["resume"])
	return a.StreamBoatmanModeExecution(sessionID, input, mode, linearAPIKey, session.ProjectPath)
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

		// Register and switch to the new boatmanmode agent
		session.RegisterBoatmanAgent(id, name, description)
		session.SetCurrentAgent(id)
		session.AddBoatmanMessage("system", fmt.Sprintf("Started: %s", name))

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

		// Add completion message and restore parent agent
		session.AddBoatmanMessage("system", fmt.Sprintf("Completed: %s (%s)", name, status))
		session.CompleteCurrentAgent()

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

			// Store worktree path in session ModeConfig for the Changes tab
			if worktreePath, ok := data["worktree_path"].(string); ok {
				session.SetModeConfigValue("worktreePath", worktreePath)
			}
			if baseBranch, ok := data["base_branch"].(string); ok {
				session.SetModeConfigValue("baseBranch", baseBranch)
			}
		}
		if len(metadata) > 0 {
			session.UpdateTaskMetadata(id, metadata)
		}

	case "progress":
		// Route progress messages through the session
		message, _ := eventData["message"].(string)
		if message != "" {
			session.AddBoatmanMessage("system", message)
			// Also append to active task metadata for the task detail modal
			session.AppendCurrentTaskLog(message)
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
// Triage Integration Methods
// =============================================================================

// CreateTriageSession creates a new triage agent session.
func (a *App) CreateTriageSession(projectPath string) (*AgentSessionInfo, error) {
	session, err := a.agentManager.CreateTriageSession(projectPath)
	if err != nil {
		return nil, err
	}

	return &AgentSessionInfo{
		ID:          session.ID,
		ProjectPath: session.ProjectPath,
		Status:      session.Status,
		CreatedAt:   session.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		Tags:        session.Tags,
		Mode:        session.Mode,
	}, nil
}

// StreamTriageExecution runs the triage pipeline with streaming output.
// Returns immediately and runs the execution in the background.
func (a *App) StreamTriageExecution(sessionID string, opts triageintegration.TriageOptions, linearAPIKey, projectPath string) error {
	prefs := a.config.GetPreferences()
	claudeAPIKey := prefs.APIKey

	integration, err := triageintegration.NewIntegration(linearAPIKey, claudeAPIKey, projectPath)
	if err != nil {
		return fmt.Errorf("failed to create triage integration: %w", err)
	}

	session, sessionErr := a.agentManager.GetSession(sessionID)
	if sessionErr == nil && session != nil {
		session.Status = "running"
	}

	// Emit status change so frontend updates
	runtime.EventsEmit(a.ctx, "agent:status", map[string]interface{}{
		"sessionId": sessionID,
		"status":    "running",
	})

	runtime.EventsEmit(a.ctx, "triage:started", map[string]interface{}{
		"sessionId": sessionID,
	})

	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("[triage] Recovered from panic: %v\n", r)
				func() {
					defer func() { recover() }()
					runtime.EventsEmit(a.ctx, "triage:error", map[string]interface{}{
						"sessionId": sessionID,
						"error":     fmt.Sprintf("Execution panic: %v", r),
					})
				}()
			}
		}()

		outputChan := make(chan string, 100)

		go func() {
			defer func() { recover() }()
			for msg := range outputChan {
				func() {
					defer func() { recover() }()
					runtime.EventsEmit(a.ctx, "triage:output", map[string]interface{}{
						"sessionId": sessionID,
						"message":   msg,
					})
				}()
			}
		}()

		// Handle events directly in the Go backend to avoid race conditions
		// with the frontend round-trip. Store triage_complete result before
		// emitting the completion event to the frontend.
		onEvent := func(event triageintegration.TriageEvent) {
			if sessionErr != nil || session == nil {
				return
			}
			switch event.Type {
			case "triage_complete":
				if event.Data != nil {
					if result, ok := event.Data["result"]; ok {
						session.SetModeConfigValue("triageResult", result)
					}
					if stats, ok := event.Data["stats"]; ok {
						session.SetModeConfigValue("triageStats", stats)
					}
				}
			case "plan_complete":
				if event.Data != nil {
					if results, ok := event.Data["results"]; ok {
						session.SetModeConfigValue("plans", results)
					}
					if stats, ok := event.Data["stats"]; ok {
						session.SetModeConfigValue("planStats", stats)
					}
				}
			}
		}

		err := integration.StreamTriageExecution(a.ctx, sessionID, opts, outputChan, onEvent)
		close(outputChan)

		if err != nil {
			fmt.Printf("[triage] Execution error: %v\n", err)
			runtime.EventsEmit(a.ctx, "triage:error", map[string]interface{}{
				"sessionId": sessionID,
				"error":     err.Error(),
			})
		} else {
			runtime.EventsEmit(a.ctx, "triage:complete", map[string]interface{}{
				"sessionId": sessionID,
			})
		}

		if sessionErr == nil && session != nil {
			if saveErr := agent.SaveSession(session); saveErr != nil {
				fmt.Printf("[triage] Warning: failed to save session %s: %v\n", sessionID, saveErr)
			}
		}
	}()

	return nil
}

// HandleTriageEvent processes triage events and updates session state.
func (a *App) HandleTriageEvent(sessionID string, eventType string, eventData map[string]interface{}) error {
	session, err := a.agentManager.GetSession(sessionID)
	if err != nil {
		return err
	}

	switch eventType {
	case "triage_started":
		session.AddBoatmanMessage("system", "Triage pipeline started")
	case "triage_fetch_complete":
		message, _ := eventData["message"].(string)
		session.AddBoatmanMessage("system", message)
	case "triage_scoring_started":
		message, _ := eventData["message"].(string)
		session.AddBoatmanMessage("system", message)
	case "triage_ticket_scored":
		message, _ := eventData["message"].(string)
		session.AddBoatmanMessage("system", message)
	case "triage_scoring_complete":
		message, _ := eventData["message"].(string)
		session.AddBoatmanMessage("system", message)
	case "triage_complete":
		// Store the full result in ModeConfig for the frontend to retrieve
		if data, ok := eventData["data"].(map[string]interface{}); ok {
			if result, ok := data["result"]; ok {
				session.SetModeConfigValue("triageResult", result)
			}
			if stats, ok := data["stats"]; ok {
				session.SetModeConfigValue("triageStats", stats)
			}
		}
		session.AddBoatmanMessage("system", "Triage pipeline complete")
	case "plan_started":
		message, _ := eventData["message"].(string)
		session.AddBoatmanMessage("system", message)
	case "plan_ticket_planned":
		message, _ := eventData["message"].(string)
		session.AddBoatmanMessage("system", message)
	case "plan_complete":
		if data, ok := eventData["data"].(map[string]interface{}); ok {
			if results, ok := data["results"]; ok {
				session.SetModeConfigValue("plans", results)
			}
			if stats, ok := data["stats"]; ok {
				session.SetModeConfigValue("planStats", stats)
			}
		}
		session.AddBoatmanMessage("system", "Plan generation complete")
	case "triage_error":
		message, _ := eventData["message"].(string)
		session.AddBoatmanMessage("system", fmt.Sprintf("Triage error: %s", message))
	}

	return nil
}

// GetTriageResult returns the stored triage result for a session.
func (a *App) GetTriageResult(sessionID string) (map[string]interface{}, error) {
	session, err := a.agentManager.GetSession(sessionID)
	if err != nil {
		return nil, err
	}
	return session.ModeConfig, nil
}

// ExecuteTriageTicket creates a BoatmanMode session to execute a triage-identified ticket.
// If the triage session has a pre-generated plan for this ticket, it is stored
// in the new session's ModeConfig so that StreamBoatmanModeExecution can pass
// it to the CLI via --plan-file, skipping the planning step.
func (a *App) ExecuteTriageTicket(sessionID, ticketID string) (*AgentSessionInfo, error) {
	triageSession, err := a.agentManager.GetSession(sessionID)
	if err != nil {
		return nil, fmt.Errorf("triage session not found: %w", err)
	}

	info, err := a.CreateBoatmanModeSession(triageSession.ProjectPath, ticketID, "ticket")
	if err != nil {
		return nil, err
	}

	// Look up the matching triage plan and attach it to the new session.
	if plans, ok := triageSession.ModeConfig["plans"]; ok {
		if planList, ok := plans.([]interface{}); ok {
			for _, p := range planList {
				pm, ok := p.(map[string]interface{})
				if !ok {
					continue
				}
				if pm["ticketId"] == ticketID {
					if planData, ok := pm["plan"]; ok && planData != nil {
						bmSession, _ := a.agentManager.GetSession(info.ID)
						if bmSession != nil {
							bmSession.SetModeConfigValue("triagePlan", planData)
						}
					}
					break
				}
			}
		}
	}

	return info, nil
}

// ExportTriagePDF exports the triage result for a session as a PDF file.
// Opens a native save dialog and writes the PDF to the chosen path.
func (a *App) ExportTriagePDF(sessionID string) (string, error) {
	session, err := a.agentManager.GetSession(sessionID)
	if err != nil {
		return "", fmt.Errorf("session not found: %w", err)
	}

	resultData, ok := session.ModeConfig["triageResult"]
	if !ok {
		return "", fmt.Errorf("no triage result available for this session")
	}

	resultMap, ok := resultData.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("triage result has unexpected format")
	}

	// Open native save dialog
	savePath, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Export Triage Report",
		DefaultFilename: fmt.Sprintf("triage-report-%s.pdf", time.Now().Format("2006-01-02")),
		Filters: []runtime.FileFilter{
			{DisplayName: "PDF Files", Pattern: "*.pdf"},
		},
	})
	if err != nil {
		return "", fmt.Errorf("save dialog error: %w", err)
	}
	if savePath == "" {
		return "", nil // User cancelled
	}

	// Ensure .pdf extension
	if !strings.HasSuffix(strings.ToLower(savePath), ".pdf") {
		savePath += ".pdf"
	}

	// Attach plan data if available.
	if plans, ok := session.ModeConfig["plans"]; ok {
		resultMap["plans"] = plans
	}
	if planStats, ok := session.ModeConfig["planStats"]; ok {
		resultMap["planStats"] = planStats
	}

	if err := triageintegration.GeneratePDF(resultMap, savePath); err != nil {
		return "", fmt.Errorf("failed to generate PDF: %w", err)
	}

	return savePath, nil
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

// =============================================================================
// Harness Methods
// =============================================================================

// ScaffoldHarness generates a new harness project from the scaffold templates.
func (a *App) ScaffoldHarness(req harnessui.ScaffoldRequest) (*harnessui.ScaffoldResponse, error) {
	return harnessui.GenerateScaffold(req)
}

// ListHarnesses returns all discovered harness projects in ~/.boatman/harnesses/.
func (a *App) ListHarnesses() ([]harnessui.HarnessInfo, error) {
	return harnessui.ListHarnesses()
}

// RunHarness starts a harness subprocess and streams output via events.
func (a *App) RunHarness(runID string, req harnessui.RunRequest) error {
	ctx, cancel := context.WithCancel(a.ctx)

	a.harnessMu.Lock()
	a.harnessRuns[runID] = cancel
	a.harnessMu.Unlock()

	go func() {
		defer func() {
			a.harnessMu.Lock()
			delete(a.harnessRuns, runID)
			a.harnessMu.Unlock()
		}()

		err := harnessui.RunHarness(ctx, a.ctx, runID, req)
		if err != nil {
			runtime.EventsEmit(a.ctx, "harness:error", map[string]any{
				"runId": runID,
				"error": err.Error(),
			})
		} else {
			runtime.EventsEmit(a.ctx, "harness:complete", map[string]any{
				"runId": runID,
			})
		}
	}()

	return nil
}

// StopHarness cancels a running harness subprocess.
func (a *App) StopHarness(runID string) error {
	a.harnessMu.Lock()
	cancel, ok := a.harnessRuns[runID]
	a.harnessMu.Unlock()

	if !ok {
		return fmt.Errorf("no running harness with ID %q", runID)
	}

	cancel()
	return nil
}

// SelectHarnessFolder opens a directory selection dialog for harness output.
func (a *App) SelectHarnessFolder() (string, error) {
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Harness Output Folder",
	})
}

// =============================================================================
// Brain Methods
// =============================================================================

// ListBrains returns all available brains for the active project.
func (a *App) ListBrains() ([]services.BrainEntry, error) {
	return a.brainService.ListBrains()
}

// GetBrain loads a specific brain by ID with full content.
func (a *App) GetBrain(id string) (*services.BrainDetail, error) {
	return a.brainService.GetBrain(id)
}

// MatchBrains finds brains matching the given context.
func (a *App) MatchBrains(keywords []string, filePaths []string, entities []string) ([]services.BrainEntry, error) {
	return a.brainService.MatchBrains(keywords, filePaths, entities)
}

// ValidateBrain validates a brain by ID.
func (a *App) ValidateBrain(id string) (*services.BrainValidationResult, error) {
	return a.brainService.ValidateBrain(id)
}

// ListSignals returns all recorded knowledge gap signals.
func (a *App) ListSignals() ([]services.SignalEntry, error) {
	return a.brainService.ListSignals()
}

// GetSignalsByDomain returns signals for a specific domain.
func (a *App) GetSignalsByDomain(domain string) ([]services.SignalEntry, error) {
	return a.brainService.GetSignalsByDomain(domain)
}

// GetBrainDirs returns the directories being scanned for brains.
func (a *App) GetBrainDirs() []string {
	return a.brainService.GetBrainDirs()
}

// SetBrainProjectPath sets the project path for brain resolution.
func (a *App) SetBrainProjectPath(path string) {
	a.brainService.SetProjectPath(path)
}

// =============================================================================
// Claude Chat Methods
// =============================================================================

// SetSessionSystemPrompt sets the system prompt for a session.
func (a *App) SetSessionSystemPrompt(sessionID, prompt string) error {
	session, err := a.agentManager.GetSession(sessionID)
	if err != nil {
		return err
	}
	session.SetSystemPrompt(prompt)
	return nil
}

// GetSessionSystemPrompt returns the system prompt for a session.
func (a *App) GetSessionSystemPrompt(sessionID string) (string, error) {
	session, err := a.agentManager.GetSession(sessionID)
	if err != nil {
		return "", err
	}
	return session.GetSystemPrompt(), nil
}

// ClearSessionConversation clears the conversation for a session.
func (a *App) ClearSessionConversation(sessionID string) error {
	session, err := a.agentManager.GetSession(sessionID)
	if err != nil {
		return err
	}
	session.ClearConversation()
	return nil
}

// GetSessionInfo returns session metadata for display.
func (a *App) GetSessionInfo(sessionID string) (map[string]interface{}, error) {
	session, err := a.agentManager.GetSession(sessionID)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"projectPath":    session.ProjectPath,
		"model":          session.Model,
		"reasoningEffort": session.ReasoningEffort,
		"conversationId": session.GetConversationID(),
		"systemPrompt":   session.GetSystemPrompt(),
		"messageCount":   len(session.GetMessages()),
		"mode":           session.Mode,
	}, nil
}

// AutoDistillBrains triggers auto-distillation of signals into brain files.
func (a *App) AutoDistillBrains() ([]services.AutoDistillResult, error) {
	return a.brainService.AutoDistillBrains()
}
