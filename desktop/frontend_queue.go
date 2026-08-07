package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"boatman/agent"
	bmintegration "boatman/boatmanmode"

	agentruntime "github.com/philjestin/boatman-ecosystem/shared/agentruntime"
	"github.com/philjestin/boatman-ecosystem/shared/agentruntime/frontendqueue"
)

const defaultLinearGraphQLURL = "https://api.linear.app/graphql"

// FrontendTicketQueueRequest asks Boatman to fetch and plan frontend ticket work.
type FrontendTicketQueueRequest struct {
	ProjectPath     string                 `json:"projectPath,omitempty"`
	LinearProject   string                 `json:"linearProject,omitempty"`
	Query           string                 `json:"query,omitempty"`
	Limit           int                    `json:"limit,omitempty"`
	States          []string               `json:"states,omitempty"`
	IncludeCanceled bool                   `json:"includeCanceled,omitempty"`
	MaxParallel     int                    `json:"maxParallel,omitempty"`
	Tickets         []frontendqueue.Ticket `json:"tickets,omitempty"`
}

// FrontendTicketQueueResult is the desktop response for frontend queue planning.
type FrontendTicketQueueResult struct {
	Source        string                 `json:"source"`
	LinearProject string                 `json:"linearProject,omitempty"`
	Tickets       []frontendqueue.Ticket `json:"tickets"`
	Plan          frontendqueue.Plan     `json:"plan"`
	Warning       string                 `json:"warning,omitempty"`
}

// FrontendTicketQueueRunRequest starts planned frontend tickets as BoatmanMode
// worker sessions. The frontend passes a previously generated plan so execution
// is stable even if Linear changes between planning and running.
type FrontendTicketQueueRunRequest struct {
	ProjectPath   string                    `json:"projectPath,omitempty"`
	LinearProject string                    `json:"linearProject,omitempty"`
	Plan          frontendqueue.Plan        `json:"plan"`
	BatchID       string                    `json:"batchId,omitempty"`
	TicketIDs     []string                  `json:"ticketIds,omitempty"`
	RunAll        bool                      `json:"runAll,omitempty"`
	BaseBranch    string                    `json:"baseBranch,omitempty"`
	Models        agentruntime.ModelProfile `json:"models,omitempty"`
}

// FrontendTicketQueueWorkerSession describes one spawned worker session.
type FrontendTicketQueueWorkerSession struct {
	TicketID   string                         `json:"ticketId"`
	SessionID  string                         `json:"sessionId,omitempty"`
	Policy     frontendqueue.AutomationPolicy `json:"policy"`
	BranchName string                         `json:"branchName,omitempty"`
	Status     string                         `json:"status"`
	Message    string                         `json:"message,omitempty"`
}

// FrontendTicketQueueRunResult is returned immediately after worker sessions
// are queued. Workers continue streaming in the background.
type FrontendTicketQueueRunResult struct {
	RunID        string                             `json:"runId"`
	BatchID      string                             `json:"batchId,omitempty"`
	ProjectPath  string                             `json:"projectPath"`
	StartedAt    string                             `json:"startedAt"`
	Sessions     []FrontendTicketQueueWorkerSession `json:"sessions"`
	StartedCount int                                `json:"startedCount"`
	SkippedCount int                                `json:"skippedCount"`
}

// PlanFrontendTicketQueue fetches Linear tickets when needed, classifies them,
// and creates a conflict-aware queue plan for frontend automation.
func (a *App) PlanFrontendTicketQueue(input FrontendTicketQueueRequest) (*FrontendTicketQueueResult, error) {
	input = normalizeFrontendQueueRequest(input)
	tickets := input.Tickets
	source := "request"
	if len(tickets) == 0 {
		project := normalizeLinearProjectInput(input.LinearProject)
		if project == "" {
			return nil, fmt.Errorf("linear project is required when no tickets are provided")
		}
		prefs := a.config.GetPreferences()
		apiKey := strings.TrimSpace(prefs.LinearAPIKey)
		if apiKey == "" {
			return nil, fmt.Errorf("configure Linear API key in Settings before fetching frontend tickets")
		}
		client := newDesktopLinearClient(apiKey)
		fetched, err := client.FetchProjectIssues(context.Background(), project, input.Limit)
		if err != nil {
			return nil, err
		}
		tickets = filterFrontendTickets(fetched, input)
		source = "linear"
		input.LinearProject = project
	}

	plan := frontendqueue.PlanTickets(tickets, frontendqueue.PlanOptions{
		ProjectPath:   input.ProjectPath,
		LinearProject: input.LinearProject,
		MaxParallel:   input.MaxParallel,
	})
	return &FrontendTicketQueueResult{
		Source:        source,
		LinearProject: input.LinearProject,
		Tickets:       tickets,
		Plan:          plan,
	}, nil
}

// RunFrontendTicketQueue starts a batch or explicit set of planned frontend
// tickets as parallel BoatmanMode worker sessions.
func (a *App) RunFrontendTicketQueue(input FrontendTicketQueueRunRequest) (*FrontendTicketQueueRunResult, error) {
	input = normalizeFrontendQueueRunRequest(input)
	if input.ProjectPath == "" {
		input.ProjectPath = strings.TrimSpace(input.Plan.Options.ProjectPath)
	}
	if input.ProjectPath == "" {
		return nil, fmt.Errorf("project path is required to run frontend queue workers")
	}
	if len(input.Plan.Tickets) == 0 {
		return nil, fmt.Errorf("frontend queue plan has no tickets")
	}

	selected, batchID := selectFrontendQueueRunPlans(input)
	if len(selected) == 0 {
		return nil, fmt.Errorf("no matching frontend queue tickets were selected")
	}

	prefs := a.config.GetPreferences()
	models := input.Models.WithDefault(prefs.DefaultModel)
	integration, err := bmintegration.NewIntegration(prefs.LinearAPIKey, prefs.APIKey, input.ProjectPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create BoatmanMode integration: %w", err)
	}

	runID := "frontend-queue-" + time.Now().UTC().Format("20060102-150405")
	result := &FrontendTicketQueueRunResult{
		RunID:       runID,
		BatchID:     batchID,
		ProjectPath: input.ProjectPath,
		StartedAt:   time.Now().UTC().Format(time.RFC3339),
		Sessions:    make([]FrontendTicketQueueWorkerSession, 0, len(selected)),
	}

	for _, plan := range selected {
		if !frontendQueuePolicyRunnable(plan.Policy) {
			result.SkippedCount++
			result.Sessions = append(result.Sessions, FrontendTicketQueueWorkerSession{
				TicketID: plan.Ticket.ID,
				Policy:   plan.Policy,
				Status:   "skipped",
				Message:  fmt.Sprintf("Policy %q is not runnable without clarification.", plan.Policy),
			})
			continue
		}

		branchName := frontendQueueBranchName(runID, plan)
		prompt := frontendQueueExecutionPrompt(plan, input, runID, branchName, models)
		session, err := a.agentManager.CreateBoatmanModeSession(input.ProjectPath, prompt, "prompt")
		if err != nil {
			result.SkippedCount++
			result.Sessions = append(result.Sessions, FrontendTicketQueueWorkerSession{
				TicketID: plan.Ticket.ID,
				Policy:   plan.Policy,
				Status:   "failed",
				Message:  err.Error(),
			})
			continue
		}
		_ = session.Start(models.Implementation)
		session.AddTag("frontend-queue")
		session.AddTag(plan.Ticket.ID)
		session.SetModeConfigValue("frontendQueueRunId", runID)
		session.SetModeConfigValue("frontendQueueBatchId", batchID)
		session.SetModeConfigValue("frontendQueueTicketId", plan.Ticket.ID)
		session.SetModeConfigValue("frontendQueuePolicy", string(plan.Policy))
		session.SetModeConfigValue("frontendQueueTargetKey", plan.TargetKey)
		session.SetModeConfigValue("baseBranch", input.BaseBranch)
		session.SetModeConfigValue("branchName", branchName)
		session.SetModeConfigValue("models", models)
		a.emitAgentSession(session)

		session.AddOrUpdateTask("frontend-queue-"+safeSlug(plan.Ticket.ID, 64), "Frontend ticket: "+plan.Ticket.Title, plan.TargetKey, "pending")
		session.AddBoatmanMessage("system", fmt.Sprintf("Queued frontend worker for %s. Branch: %s", plan.Ticket.ID, branchName))
		if err := agent.SaveSession(session); err != nil {
			session.AddBoatmanMessage("system", "Warning: failed to persist queued worker session: "+err.Error())
		}

		go a.runFrontendQueueWorker(routineContext(a.ctx), integration, session, plan, prompt, branchName, models)

		result.StartedCount++
		result.Sessions = append(result.Sessions, FrontendTicketQueueWorkerSession{
			TicketID:   plan.Ticket.ID,
			SessionID:  session.ID,
			Policy:     plan.Policy,
			BranchName: branchName,
			Status:     "started",
			Message:    "Worker session started.",
		})
	}
	if result.StartedCount == 0 && result.SkippedCount > 0 {
		return result, nil
	}
	return result, nil
}

func normalizeFrontendQueueRequest(input FrontendTicketQueueRequest) FrontendTicketQueueRequest {
	input.ProjectPath = strings.TrimSpace(input.ProjectPath)
	input.LinearProject = normalizeLinearProjectInput(input.LinearProject)
	input.Query = strings.TrimSpace(input.Query)
	if input.Limit <= 0 {
		input.Limit = 50
	}
	if input.Limit > 250 {
		input.Limit = 250
	}
	if input.MaxParallel <= 0 {
		input.MaxParallel = 3
	}
	return input
}

func normalizeFrontendQueueRunRequest(input FrontendTicketQueueRunRequest) FrontendTicketQueueRunRequest {
	input.ProjectPath = strings.TrimSpace(input.ProjectPath)
	input.LinearProject = normalizeLinearProjectInput(input.LinearProject)
	input.BatchID = strings.TrimSpace(input.BatchID)
	input.BaseBranch = strings.TrimSpace(input.BaseBranch)
	if input.BaseBranch == "" {
		input.BaseBranch = strings.TrimSpace(input.Plan.Options.DefaultBaseBranch)
	}
	if input.BaseBranch == "" {
		input.BaseBranch = "main"
	}
	input.TicketIDs = uniqueTrimmedStrings(input.TicketIDs)
	return input
}

func normalizeLinearProjectInput(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if strings.Contains(value, "linear.app/") {
		parsed, err := url.Parse(value)
		if err == nil {
			parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
			for i, part := range parts {
				if part == "project" && i+1 < len(parts) {
					return strings.TrimSpace(parts[i+1])
				}
			}
		}
	}
	value = strings.TrimPrefix(value, "project:")
	value = strings.TrimSuffix(value, "/issues")
	return strings.Trim(value, "/")
}

func selectFrontendQueueRunPlans(input FrontendTicketQueueRunRequest) ([]frontendqueue.TicketPlan, string) {
	selectedIDs := map[string]bool{}
	batchID := input.BatchID
	if input.RunAll {
		for _, plan := range input.Plan.Tickets {
			if frontendQueuePolicyRunnable(plan.Policy) {
				selectedIDs[strings.ToLower(plan.Ticket.ID)] = true
			}
		}
	} else if batchID != "" {
		for _, batch := range input.Plan.Batches {
			if strings.EqualFold(batch.ID, batchID) {
				for _, ticketID := range batch.TicketIDs {
					selectedIDs[strings.ToLower(ticketID)] = true
				}
				batchID = batch.ID
				break
			}
		}
	} else {
		for _, ticketID := range input.TicketIDs {
			selectedIDs[strings.ToLower(ticketID)] = true
		}
	}

	if len(selectedIDs) == 0 {
		return nil, batchID
	}
	out := make([]frontendqueue.TicketPlan, 0, len(selectedIDs))
	for _, plan := range input.Plan.Tickets {
		if selectedIDs[strings.ToLower(plan.Ticket.ID)] {
			out = append(out, plan)
		}
	}
	return out, batchID
}

func frontendQueuePolicyRunnable(policy frontendqueue.AutomationPolicy) bool {
	return policy == frontendqueue.PolicyAutoPR || policy == frontendqueue.PolicyDraftPR
}

func frontendQueueBranchName(runID string, plan frontendqueue.TicketPlan) string {
	ticketID := safeSlug(plan.Ticket.ID, 24)
	title := safeSlug(plan.Ticket.Title, 42)
	run := strings.TrimPrefix(safeSlug(runID, 40), "frontend-queue-")
	parts := []string{"frontend-queue"}
	if ticketID != "" {
		parts = append(parts, ticketID)
	}
	if title != "" {
		parts = append(parts, title)
	}
	if run != "" {
		parts = append(parts, run)
	}
	return strings.Join(parts, "/")
}

func frontendQueueExecutionPrompt(plan frontendqueue.TicketPlan, input FrontendTicketQueueRunRequest, runID, branchName string, models agentruntime.ModelProfile) string {
	validation := make([]string, 0, len(plan.Validation))
	for _, step := range plan.Validation {
		line := step.Description
		if step.Command != "" {
			line += " (" + step.Command + ")"
		}
		validation = append(validation, line)
	}
	if len(validation) == 0 {
		validation = append(validation, "Run the most relevant frontend build, test, and visual checks for the changed surface.")
	}

	passK := "disabled"
	if plan.PassKCandidates > 0 {
		passK = fmt.Sprintf("%d candidates; keep the smallest passing implementation and discard the rest", plan.PassKCandidates)
	}

	return strings.TrimSpace(fmt.Sprintf(`# Frontend Queue Worker

Run ID: %s
Linear project: %s
Ticket: %s
Policy: %s
Target key: %s
Branch name: %s
Base branch: %s

## Model routing
- Planning model: %s
- Implementation model: %s
- Skill/review model: %s

## Deterministic queue plan
%s

## Automation contract
1. Start by using /plan for the ticket investigation and implementation plan.
2. Start from latest %s in a fresh git worktree, using branch %s.
3. Implement the smallest code change that satisfies the ticket. Prefer existing design-system components and tokens.
4. Use Chrome MCP or Playwright for visual validation when the ticket touches UI pixels, layout, copy, or interaction state.
5. Capture before/after evidence and run the validation gates below until they pass or a real blocker is proven.
6. Open a draft PR with Linear/Figma links, screenshots, validation output, and residual risk.
7. Run peer-review and lydia-code-review, address actionable feedback, then rerun the relevant checks.
8. Do not stop for a human unless the ticket needs product/design clarification or credentials/environment are unavailable.

Pass@K: %s

Validation gates:
- %s
`, runID, input.LinearProject, plan.Ticket.ID, plan.Policy, plan.TargetKey, branchName, input.BaseBranch, models.Plan, models.Implementation, models.Skills, plan.WorkerPrompt, input.BaseBranch, branchName, passK, strings.Join(validation, "\n- ")))
}

func (a *App) runFrontendQueueWorker(ctx context.Context, integration *bmintegration.Integration, session *agent.Session, plan frontendqueue.TicketPlan, prompt, branchName string, models agentruntime.ModelProfile) {
	taskID := "frontend-queue-" + safeSlug(plan.Ticket.ID, 64)
	session.SetStatus(agent.SessionStatusRunning)
	session.AddOrUpdateTask(taskID, "Frontend ticket: "+plan.Ticket.Title, plan.TargetKey, "in_progress")

	outputChan := make(chan string, 100)
	done := make(chan struct{})
	go func() {
		defer close(done)
		for line := range outputChan {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			if strings.Contains(line, "PR") || strings.Contains(line, "Worktree") || strings.Contains(line, "Review") || strings.Contains(line, "Validation") {
				session.AddBoatmanMessage("system", line)
			}
		}
	}()

	var streamBuilder strings.Builder
	var streamMsgID string
	onMessage := func(role, content string) {
		if role == "provider.raw" || role == "claude_stream" {
			session.ProcessExternalStreamLine(content, &streamBuilder, &streamMsgID)
			return
		}
		session.AddBoatmanMessage(role, content)
	}
	onEvent := func(event bmintegration.BoatmanEvent) {
		_ = a.HandleBoatmanModeEvent(session.ID, event.Type, boatmanEventMap(event))
	}

	_, err := integration.StreamExecutionWithOptions(ctx, session.ID, prompt, "prompt", bmintegration.StreamExecutionOptions{
		KeepDraft:         true,
		ReviewSkill:       "peer-review",
		ExtraReviewSkills: []string{"lydia-code-review"},
		Models:            models,
		Title:             fmt.Sprintf("%s: %s", plan.Ticket.ID, plan.Ticket.Title),
		BranchName:        branchName,
		SuppressWailsEmit: true,
		OnEvent:           onEvent,
	}, outputChan, onMessage)
	close(outputChan)
	<-done

	if err != nil {
		session.SetStatus(agent.SessionStatusError)
		session.AddOrUpdateTask(taskID, "Frontend ticket: "+plan.Ticket.Title, plan.TargetKey, "failed")
		session.AddBoatmanMessage("system", "Frontend queue worker failed: "+err.Error())
		_ = agent.SaveSession(session)
		return
	}

	session.AddOrUpdateTask(taskID, "Frontend ticket: "+plan.Ticket.Title, plan.TargetKey, "completed")
	session.AddBoatmanMessage("system", "Frontend queue worker completed. Review the draft PR and validation evidence.")
	session.SetStatus(agent.SessionStatusStopped)
	_ = agent.SaveSession(session)
}

func filterFrontendTickets(tickets []frontendqueue.Ticket, input FrontendTicketQueueRequest) []frontendqueue.Ticket {
	states := map[string]bool{}
	for _, state := range input.States {
		state = strings.ToLower(strings.TrimSpace(state))
		if state != "" {
			states[state] = true
		}
	}
	query := strings.ToLower(strings.TrimSpace(input.Query))
	out := make([]frontendqueue.Ticket, 0, len(tickets))
	for _, ticket := range tickets {
		statusType := strings.ToLower(ticket.StatusType)
		status := strings.ToLower(ticket.Status)
		if !input.IncludeCanceled && (statusType == "canceled" || statusType == "cancelled" || statusType == "completed") {
			continue
		}
		if len(states) > 0 && !states[statusType] && !states[status] {
			continue
		}
		if query != "" {
			body := strings.ToLower(ticket.ID + " " + ticket.Title + " " + ticket.Description + " " + strings.Join(ticket.Labels, " "))
			if !strings.Contains(body, query) {
				continue
			}
		}
		out = append(out, ticket)
	}
	return out
}

func uniqueTrimmedStrings(values []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		key := strings.ToLower(value)
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, value)
	}
	return out
}

type desktopLinearClient struct {
	apiKey     string
	apiURL     string
	httpClient *http.Client
}

func newDesktopLinearClient(apiKey string) *desktopLinearClient {
	apiURL := strings.TrimSpace(os.Getenv("LINEAR_API_URL"))
	if apiURL == "" {
		apiURL = defaultLinearGraphQLURL
	}
	return &desktopLinearClient{
		apiKey: apiKey,
		apiURL: apiURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *desktopLinearClient) FetchProjectIssues(ctx context.Context, projectInput string, limit int) ([]frontendqueue.Ticket, error) {
	project, err := c.resolveProject(ctx, projectInput)
	if err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = 50
	}
	if limit > 250 {
		limit = 250
	}
	query := `
query FrontendQueueIssues($projectID: String!, $first: Int!) {
  issues(filter: { project: { id: { eq: $projectID } } }, first: $first) {
    nodes {
      id
      identifier
      title
      description
      url
      branchName
      priority
      estimate
      createdAt
      updatedAt
      parent { identifier }
      state { name type }
      team { name key }
      project { id name slugId }
      labels { nodes { name } }
    }
  }
}`
	var result struct {
		Data struct {
			Issues struct {
				Nodes []linearIssueNode `json:"nodes"`
			} `json:"issues"`
		} `json:"data"`
		Errors []linearGraphQLError `json:"errors"`
	}
	if err := c.execute(ctx, query, map[string]any{
		"projectID": project.ID,
		"first":     limit,
	}, &result); err != nil {
		return nil, err
	}
	if len(result.Errors) > 0 {
		return nil, fmt.Errorf("linear API error: %s", result.Errors[0].Message)
	}
	tickets := make([]frontendqueue.Ticket, 0, len(result.Data.Issues.Nodes))
	for _, node := range result.Data.Issues.Nodes {
		tickets = append(tickets, linearIssueToFrontendTicket(node))
	}
	return tickets, nil
}

func (c *desktopLinearClient) resolveProject(ctx context.Context, projectInput string) (linearProjectNode, error) {
	query := `
query FrontendQueueProject($project: String!) {
  projects(filter: {
    or: [
      { id: { eq: $project } }
      { slugId: { eq: $project } }
      { name: { containsIgnoreCase: $project } }
    ]
  }, first: 1) {
    nodes { id name slugId }
  }
}`
	var result struct {
		Data struct {
			Projects struct {
				Nodes []linearProjectNode `json:"nodes"`
			} `json:"projects"`
		} `json:"data"`
		Errors []linearGraphQLError `json:"errors"`
	}
	if err := c.execute(ctx, query, map[string]any{"project": projectInput}, &result); err != nil {
		return linearProjectNode{}, err
	}
	if len(result.Errors) > 0 {
		return linearProjectNode{}, fmt.Errorf("linear API error: %s", result.Errors[0].Message)
	}
	if len(result.Data.Projects.Nodes) == 0 {
		return linearProjectNode{}, fmt.Errorf("linear project not found: %s", projectInput)
	}
	return result.Data.Projects.Nodes[0], nil
}

func (c *desktopLinearClient) execute(ctx context.Context, query string, variables map[string]any, out any) error {
	payload, err := json.Marshal(map[string]any{
		"query":     query,
		"variables": variables,
	})
	if err != nil {
		return fmt.Errorf("failed to encode Linear request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.apiURL, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("failed to create Linear request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", c.apiKey)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to call Linear: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("linear API returned %s", resp.Status)
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("failed to decode Linear response: %w", err)
	}
	return nil
}

type linearGraphQLError struct {
	Message string `json:"message"`
}

type linearProjectNode struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	SlugID string `json:"slugId"`
}

type linearIssueNode struct {
	ID          string   `json:"id"`
	Identifier  string   `json:"identifier"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	URL         string   `json:"url"`
	BranchName  string   `json:"branchName"`
	Priority    int      `json:"priority"`
	Estimate    *float64 `json:"estimate"`
	CreatedAt   string   `json:"createdAt"`
	UpdatedAt   string   `json:"updatedAt"`
	Parent      struct {
		Identifier string `json:"identifier"`
	} `json:"parent"`
	State struct {
		Name string `json:"name"`
		Type string `json:"type"`
	} `json:"state"`
	Team struct {
		Name string `json:"name"`
		Key  string `json:"key"`
	} `json:"team"`
	Project struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		SlugID string `json:"slugId"`
	} `json:"project"`
	Labels struct {
		Nodes []struct {
			Name string `json:"name"`
		} `json:"nodes"`
	} `json:"labels"`
}

func linearIssueToFrontendTicket(node linearIssueNode) frontendqueue.Ticket {
	labels := make([]string, 0, len(node.Labels.Nodes))
	for _, label := range node.Labels.Nodes {
		if strings.TrimSpace(label.Name) != "" {
			labels = append(labels, label.Name)
		}
	}
	team := node.Team.Name
	if node.Team.Key != "" {
		team = fmt.Sprintf("%s (%s)", node.Team.Name, node.Team.Key)
	}
	return frontendqueue.Ticket{
		ID:          firstNonEmpty(node.Identifier, node.ID),
		Title:       node.Title,
		Description: node.Description,
		URL:         node.URL,
		BranchName:  node.BranchName,
		Status:      node.State.Name,
		StatusType:  node.State.Type,
		Team:        team,
		Project:     node.Project.Name,
		ParentID:    node.Parent.Identifier,
		Labels:      labels,
		Estimate:    node.Estimate,
		Priority:    node.Priority,
		CreatedAt:   parseLinearTime(node.CreatedAt),
		UpdatedAt:   parseLinearTime(node.UpdatedAt),
	}
}

func parseLinearTime(value string) time.Time {
	if value == "" {
		return time.Time{}
	}
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return time.Time{}
	}
	return parsed
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}
