package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	runtimeproviders "boatman/agent/providers"
	"boatman/agent/providers/claudecli"
	"boatman/config"
	"boatman/mcp"

	agentruntime "github.com/philjestin/boatman-ecosystem/shared/agentruntime"
	"github.com/philjestin/boatman-ecosystem/shared/agentruntime/integrations"
	"github.com/philjestin/boatman-ecosystem/shared/agentruntime/routines"
	"github.com/philjestin/boatman-ecosystem/shared/agentruntime/runprep"
	"github.com/philjestin/boatman-ecosystem/shared/agentruntime/runstore"
)

// DesktopRoutine is the compact routine definition rendered by the frontend.
type DesktopRoutine struct {
	ID               string             `json:"id"`
	Name             string             `json:"name"`
	Description      string             `json:"description,omitempty"`
	Schedule         string             `json:"schedule,omitempty"`
	WorkflowTemplate string             `json:"workflowTemplate,omitempty"`
	Role             string             `json:"role"`
	Profile          string             `json:"profile"`
	Integrations     []string           `json:"integrations,omitempty"`
	Parameters       []RoutineParameter `json:"parameters,omitempty"`
	Output           RoutineOutput      `json:"output"`
	Metadata         map[string]string  `json:"metadata,omitempty"`
}

// RoutineParameter is one user-supplied routine input.
type RoutineParameter struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
	Default     string `json:"default,omitempty"`
	Required    bool   `json:"required,omitempty"`
}

// RoutineOutput describes routine output shape for the desktop UI.
type RoutineOutput struct {
	Format      string `json:"format"`
	DefaultPath string `json:"defaultPath,omitempty"`
}

// RoutineRunRequest is the desktop request for dry-running or executing a routine.
type RoutineRunRequest struct {
	RoutineID   string            `json:"routineId"`
	ProjectPath string            `json:"projectPath"`
	Values      map[string]string `json:"values,omitempty"`
	Provider    string            `json:"provider,omitempty"`
	Model       string            `json:"model,omitempty"`
	RunID       string            `json:"runId,omitempty"`
	ReportOut   string            `json:"reportOut,omitempty"`
}

// RoutineRequestPreview is a safe summary of the provider-neutral request.
type RoutineRequestPreview struct {
	RunID                 string            `json:"runId"`
	Provider              string            `json:"provider"`
	Model                 string            `json:"model,omitempty"`
	Role                  string            `json:"role"`
	Profile               string            `json:"profile,omitempty"`
	WorkDir               string            `json:"workDir,omitempty"`
	ApprovalPolicy        string            `json:"approvalPolicy,omitempty"`
	ReasoningEffort       string            `json:"reasoningEffort,omitempty"`
	MCPServerLabels       []string          `json:"mcpServerLabels,omitempty"`
	Metadata              map[string]string `json:"metadata,omitempty"`
	InstructionsPreview   string            `json:"instructionsPreview,omitempty"`
	FirstMessagePreview   string            `json:"firstMessagePreview,omitempty"`
	FirstMessageCharacter int               `json:"firstMessageCharacterCount,omitempty"`
}

// RoutineDryRunResult is returned without invoking a model.
type RoutineDryRunResult struct {
	Routine      DesktopRoutine          `json:"routine"`
	Values       map[string]string       `json:"values"`
	Request      RoutineRequestPreview   `json:"request"`
	Integrations []mcp.IntegrationStatus `json:"integrations,omitempty"`
	ReportPath   string                  `json:"reportPath,omitempty"`
	Command      string                  `json:"command,omitempty"`
}

// RoutineRunResult is returned after a routine run completes.
type RoutineRunResult struct {
	RoutineID    string                  `json:"routineId"`
	RunID        string                  `json:"runId"`
	Provider     string                  `json:"provider"`
	Model        string                  `json:"model,omitempty"`
	Values       map[string]string       `json:"values,omitempty"`
	ReportPath   string                  `json:"reportPath,omitempty"`
	Integrations []mcp.IntegrationStatus `json:"integrations,omitempty"`
	Usage        *agentruntime.Usage     `json:"usage,omitempty"`
	Report       string                  `json:"report,omitempty"`
}

type builtRoutineRun struct {
	routine      routines.Routine
	values       map[string]string
	request      agentruntime.RunRequest
	statuses     []mcp.IntegrationStatus
	secretEnv    map[string]string
	reportPath   string
	providerName string
}

// ListRoutines returns Boatman's built-in repeatable routines.
func (a *App) ListRoutines() ([]DesktopRoutine, error) {
	items := routines.DefaultLibrary().List()
	return desktopRoutines(items)
}

// ListProjectRoutines returns built-in routines plus definitions from the
// selected project's .boatman routine files.
func (a *App) ListProjectRoutines(projectPath string) ([]DesktopRoutine, error) {
	workDir, err := routineWorkDir(projectPath)
	if err != nil {
		return nil, err
	}
	library, err := routines.ProjectLibrary(workDir)
	if err != nil {
		return nil, err
	}
	return desktopRoutines(library.List())
}

func desktopRoutines(items []routines.Routine) ([]DesktopRoutine, error) {
	out := make([]DesktopRoutine, 0, len(items))
	for _, item := range items {
		if err := routines.Validate(item); err != nil {
			return nil, err
		}
		out = append(out, desktopRoutine(item))
	}
	return out, nil
}

// DryRunRoutine validates configuration and returns the provider-neutral request
// that would be sent without invoking a model.
func (a *App) DryRunRoutine(input RoutineRunRequest) (*RoutineDryRunResult, error) {
	built, err := a.buildRoutineRun(context.Background(), input, true)
	if err != nil {
		return nil, err
	}
	return &RoutineDryRunResult{
		Routine:      desktopRoutine(built.routine),
		Values:       cloneStringMap(built.values),
		Request:      routineRequestPreview(built.request),
		Integrations: built.statuses,
		ReportPath:   built.reportPath,
		Command:      routineCommandPreview(built.routine.ID, built.values),
	}, nil
}

// RunRoutine executes a repeatable routine and writes its Markdown report.
func (a *App) RunRoutine(input RoutineRunRequest) (*RoutineRunResult, error) {
	built, err := a.buildRoutineRun(context.Background(), input, false)
	if err != nil {
		return nil, err
	}
	provider, err := routineProvider(built.request.Provider, built.secretEnv)
	if err != nil {
		return nil, err
	}
	report, usage, err := runRoutineText(context.Background(), provider, built.request)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(built.reportPath) != "" && built.reportPath != "-" {
		if err := writeRoutineReport(built.reportPath, report); err != nil {
			return nil, err
		}
	}
	return &RoutineRunResult{
		RoutineID:    built.routine.ID,
		RunID:        built.request.RunID,
		Provider:     built.providerName,
		Model:        built.request.Model,
		Values:       cloneStringMap(built.values),
		ReportPath:   built.reportPath,
		Integrations: built.statuses,
		Usage:        usage,
		Report:       strings.TrimSpace(report),
	}, nil
}

func (a *App) buildRoutineRun(ctx context.Context, input RoutineRunRequest, dryRun bool) (*builtRoutineRun, error) {
	workDir, err := routineWorkDir(input.ProjectPath)
	if err != nil {
		return nil, err
	}
	library, err := routines.ProjectLibrary(workDir)
	if err != nil {
		return nil, err
	}
	routine, ok := library.Get(input.RoutineID)
	if !ok {
		return nil, fmt.Errorf("unknown routine %q", input.RoutineID)
	}
	values, err := routines.Values(routine, input.Values)
	if err != nil {
		return nil, err
	}
	runID := strings.TrimSpace(input.RunID)
	if runID == "" {
		runID = routines.NewRunID(routine.ID, time.Now())
	}
	prefs := a.config.GetPreferences()
	statuses, mcpRefs, secretEnv, err := desktopRoutineIntegrationRefs(ctx, routine, prefs, dryRun)
	if err != nil {
		return nil, err
	}
	providerName := strings.TrimSpace(input.Provider)
	if providerName == "" {
		providerName = runtimeproviders.DefaultProvider
	}
	model := strings.TrimSpace(input.Model)
	if model == "" {
		model = strings.TrimSpace(prefs.DefaultModel)
	}
	req, err := routines.BuildRequest(routine, values, routines.BuildOptions{
		RunID:      runID,
		WorkDir:    workDir,
		Provider:   providerName,
		Model:      model,
		MCPServers: mcpRefs,
		Metadata:   routineAuthMetadata(prefs),
	})
	if err != nil {
		return nil, err
	}
	reportPath, err := desktopRoutineReportPath(input.ReportOut, routine, req.RunID, workDir)
	if err != nil {
		return nil, err
	}
	return &builtRoutineRun{
		routine:      routine,
		values:       values,
		request:      req,
		statuses:     statuses,
		secretEnv:    secretEnv,
		reportPath:   reportPath,
		providerName: providerName,
	}, nil
}

func desktopRoutine(item routines.Routine) DesktopRoutine {
	item = routines.Normalize(item)
	params := make([]RoutineParameter, 0, len(item.Parameters))
	for _, param := range item.Parameters {
		defaultValue := param.Default
		if item.Defaults != nil {
			if value, ok := item.Defaults[param.Name]; ok {
				defaultValue = value
			}
		}
		params = append(params, RoutineParameter{
			Name:        param.Name,
			Type:        string(param.Type),
			Description: param.Description,
			Default:     defaultValue,
			Required:    param.Required,
		})
	}
	return DesktopRoutine{
		ID:               item.ID,
		Name:             item.Name,
		Description:      item.Description,
		Schedule:         item.Schedule,
		WorkflowTemplate: item.WorkflowTemplate,
		Role:             string(item.Role),
		Profile:          item.Profile,
		Integrations:     append([]string(nil), item.Integrations...),
		Parameters:       params,
		Output: RoutineOutput{
			Format:      item.Output.Format,
			DefaultPath: item.Output.DefaultPath,
		},
		Metadata: cloneStringMap(item.Metadata),
	}
}

func desktopRoutineIntegrationRefs(ctx context.Context, routine routines.Routine, prefs config.UserPreferences, dryRun bool) ([]mcp.IntegrationStatus, []agentruntime.MCPServerRef, map[string]string, error) {
	catalog := integrations.DefaultCatalog()
	manager := integrations.NewManager(catalog)
	statuses := make([]mcp.IntegrationStatus, 0, len(routine.Integrations))
	refs := make([]agentruntime.MCPServerRef, 0, len(routine.Integrations))
	secretEnv := map[string]string{}
	for _, name := range routine.Integrations {
		item, ok := catalog.Lookup(name)
		if !ok {
			return nil, nil, nil, fmt.Errorf("routine %q references unknown integration %q", routine.ID, name)
		}
		env := desktopEnvForIntegration(item, prefs)
		conn, err := manager.Connect(ctx, item.Name, integrations.ResolveOptions{
			Enabled: true,
			Env:     env,
		})
		statuses = append(statuses, conn.Status)
		if err != nil {
			return statuses, nil, nil, err
		}
		if conn.State != integrations.StateConnected {
			if dryRun {
				continue
			}
			missing := ""
			if len(conn.Status.MissingEnv) > 0 {
				missing = fmt.Sprintf(" missing env: %s", strings.Join(conn.Status.MissingEnv, ","))
			}
			return statuses, nil, nil, fmt.Errorf("routine %q integration %q is %s:%s %s", routine.ID, item.Name, conn.State, missing, conn.Status.Message)
		}
		if conn.MCP != nil {
			for key, value := range conn.MCP.Env {
				if strings.TrimSpace(key) != "" && strings.TrimSpace(value) != "" {
					secretEnv[key] = value
				}
			}
			ref := *conn.MCP
			ref.Env = sanitizeRoutineMCPEnv(ref.Env, item)
			refs = append(refs, ref)
		}
	}
	if len(secretEnv) == 0 {
		secretEnv = nil
	}
	return statuses, refs, secretEnv, nil
}

func desktopEnvForIntegration(item integrations.Integration, prefs config.UserPreferences) map[string]string {
	env := map[string]string{}
	switch item.Name {
	case "datadog":
		setFirstNonEmpty(env, "DD_API_KEY", prefs.DatadogAPIKey, os.Getenv("DD_API_KEY"))
		setFirstNonEmpty(env, "DD_APP_KEY", prefs.DatadogAppKey, os.Getenv("DD_APP_KEY"))
		setFirstNonEmpty(env, "DD_SITE", prefs.DatadogSite, os.Getenv("DD_SITE"), "datadoghq.com")
	case "bugsnag":
		setFirstNonEmpty(env, "BUGSNAG_API_KEY", prefs.BugsnagAPIKey, os.Getenv("BUGSNAG_API_KEY"))
	case "linear":
		setFirstNonEmpty(env, "LINEAR_API_KEY", prefs.LinearAPIKey, os.Getenv("LINEAR_API_KEY"))
	case "slack":
		setFirstNonEmpty(env, "SLACK_BOT_TOKEN", os.Getenv("SLACK_BOT_TOKEN"))
		setFirstNonEmpty(env, "SLACK_TEAM_ID", os.Getenv("SLACK_TEAM_ID"))
	}
	return env
}

func setFirstNonEmpty(env map[string]string, key string, values ...string) {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			env[key] = strings.TrimSpace(value)
			return
		}
	}
}

func sanitizeRoutineMCPEnv(env map[string]string, item integrations.Integration) map[string]string {
	if len(env) == 0 {
		return nil
	}
	required := map[string]bool{}
	if item.MCP != nil {
		for _, key := range item.MCP.RequiredEnv() {
			required[key] = true
		}
	}
	out := make(map[string]string, len(env))
	for key, value := range env {
		if required[key] {
			continue
		}
		if strings.TrimSpace(value) != "" {
			out[key] = value
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func routineAuthMetadata(prefs config.UserPreferences) map[string]string {
	metadata := map[string]string{
		"authMethod":            string(prefs.AuthMethod),
		"normalizeClaudeResult": "true",
		"phaseId":               "routine",
	}
	if prefs.APIKey != "" {
		metadata["anthropicApiKey"] = prefs.APIKey
	}
	if prefs.GCPProjectID != "" {
		metadata["gcpProjectId"] = prefs.GCPProjectID
	}
	if prefs.GCPRegion != "" {
		metadata["gcpRegion"] = prefs.GCPRegion
	}
	return metadata
}

func routineProvider(name string, extraEnv map[string]string) (agentruntime.Provider, error) {
	if strings.TrimSpace(name) == "" || name == runtimeproviders.DefaultProvider {
		return claudecli.New(claudecli.WithExtraEnv(extraEnv)), nil
	}
	req := agentruntime.RunRequest{Provider: name}
	return runtimeproviders.NewDefaultRegistry().ForRequest(req)
}

func runRoutineText(ctx context.Context, provider agentruntime.Provider, req agentruntime.RunRequest) (string, *agentruntime.Usage, error) {
	preparedReq, initialEvents, err := runprep.Prepare(ctx, req, runprep.DefaultOptions())
	if err != nil {
		return "", nil, err
	}
	req = preparedReq
	provider = runprep.NewInitialEventsProvider(provider, initialEvents)
	store, storeEnabled, err := runstore.ForRequest(req)
	if err != nil {
		return "", nil, err
	}
	if storeEnabled {
		provider = runstore.NewRecordingProvider(provider, store)
	}
	stream, err := provider.StartRun(ctx, req)
	if err != nil {
		return "", nil, err
	}

	collector := routineStreamCollector{}
	for {
		select {
		case <-ctx.Done():
			return collector.Report(), collector.usage, ctx.Err()
		case event, ok := <-stream:
			if !ok {
				return collector.Report(), collector.usage, nil
			}
			if err := collector.Observe(event); err != nil {
				return collector.Report(), collector.usage, err
			}
		}
	}
}

type routineStreamCollector struct {
	fallback strings.Builder
	usage    *agentruntime.Usage
}

func (c *routineStreamCollector) Observe(event agentruntime.Event) error {
	switch event.Type {
	case agentruntime.EventMessageDelta:
		c.fallback.WriteString(event.Message)
	case agentruntime.EventMessageCompleted:
		if strings.TrimSpace(event.Message) != "" {
			if c.fallback.Len() > 0 {
				c.fallback.WriteString("\n\n")
			}
			c.fallback.WriteString(event.Message)
		}
	case agentruntime.EventUsageUpdated:
		if event.Usage != nil {
			c.usage = event.Usage
		}
	case agentruntime.EventRunFailed:
		if event.Message != "" {
			return fmt.Errorf("%s", event.Message)
		}
		return fmt.Errorf("routine provider run failed")
	}
	return nil
}

func (c *routineStreamCollector) Report() string {
	return strings.TrimSpace(c.fallback.String())
}

func routineRequestPreview(req agentruntime.RunRequest) RoutineRequestPreview {
	labels := make([]string, 0, len(req.MCPServers))
	for _, server := range req.MCPServers {
		if server.Label != "" {
			labels = append(labels, server.Label)
		}
	}
	sort.Strings(labels)
	effort := ""
	if req.Reasoning != nil {
		effort = req.Reasoning.Effort
	}
	firstMessage := ""
	if len(req.Messages) > 0 {
		firstMessage = req.Messages[0].Content
	}
	return RoutineRequestPreview{
		RunID:                 req.RunID,
		Provider:              req.Provider,
		Model:                 req.Model,
		Role:                  string(req.Role),
		Profile:               req.Profile,
		WorkDir:               req.WorkDir,
		ApprovalPolicy:        string(req.ApprovalPolicy),
		ReasoningEffort:       effort,
		MCPServerLabels:       labels,
		Metadata:              cloneStringMap(req.Metadata),
		InstructionsPreview:   truncateRoutineString(req.Instructions, 1000),
		FirstMessagePreview:   truncateRoutineString(firstMessage, 2000),
		FirstMessageCharacter: len(firstMessage),
	}
}

func routineWorkDir(projectPath string) (string, error) {
	if strings.TrimSpace(projectPath) == "" {
		return os.Getwd()
	}
	return filepath.Abs(projectPath)
}

func desktopRoutineReportPath(reportOut string, routine routines.Routine, runID, workDir string) (string, error) {
	if strings.TrimSpace(reportOut) == "-" {
		return "-", nil
	}
	if strings.TrimSpace(reportOut) != "" {
		return filepath.Abs(reportOut)
	}
	base := routine.Output.DefaultPath
	if strings.TrimSpace(base) == "" {
		base = filepath.Join(".boatman", "routines", routine.ID)
	}
	if !filepath.IsAbs(base) {
		base = filepath.Join(workDir, base)
	}
	return filepath.Join(base, runID+".md"), nil
}

func writeRoutineReport(path, report string) error {
	if strings.TrimSpace(path) == "" || path == "-" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(strings.TrimSpace(report)+"\n"), 0644)
}

func routineCommandPreview(routineID string, values map[string]string) string {
	args := []string{"boatman", "routines", "run", routineID}
	if value := strings.TrimSpace(values["graph_area"]); value != "" {
		args = append(args, "--graph-area", value)
	}
	if value := strings.TrimSpace(values["top_n"]); value != "" {
		args = append(args, "--top-n", value)
	}
	if value := strings.TrimSpace(values["lookback"]); value != "" {
		args = append(args, "--lookback", value)
	}
	if value := strings.TrimSpace(values["environment"]); value != "" {
		args = append(args, "--environment", value)
	}
	if value := strings.TrimSpace(values["service"]); value != "" {
		args = append(args, "--service", value)
	}
	return strings.Join(args, " ")
}

func cloneStringMap(values map[string]string) map[string]string {
	if len(values) == 0 {
		return nil
	}
	out := make(map[string]string, len(values))
	for key, value := range values {
		out[key] = value
	}
	return out
}

func truncateRoutineString(value string, limit int) string {
	if limit <= 0 || len(value) <= limit {
		return value
	}
	return value[:limit] + "..."
}
