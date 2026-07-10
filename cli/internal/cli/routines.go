package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	agentruntime "github.com/philjestin/boatman-ecosystem/shared/agentruntime"
	"github.com/philjestin/boatman-ecosystem/shared/agentruntime/integrations"
	"github.com/philjestin/boatman-ecosystem/shared/agentruntime/routines"
	"github.com/philjestin/boatmanmode/internal/config"
	"github.com/philjestin/boatmanmode/internal/cost"
	runtimeproviders "github.com/philjestin/boatmanmode/internal/providers"
	"github.com/spf13/cobra"
)

var routinesCmd = &cobra.Command{
	Use:   "routines",
	Short: "Inspect and run repeatable Boatman routines",
	Long: `Inspect and run saved, repeatable routines.

Routines are provider-neutral run definitions with parameters, integrations,
runtime recording, and durable reports. They are designed to be invoked by
humans, cron, CI, or a future Boatman scheduler.`,
	RunE: runRoutinesList,
}

var routinesShowCmd = &cobra.Command{
	Use:   "show [routine-id]",
	Short: "Show a routine definition",
	Args:  cobra.ExactArgs(1),
	RunE:  runRoutinesShow,
}

var routinesRunCmd = &cobra.Command{
	Use:   "run [routine-id]",
	Short: "Run a repeatable routine",
	Args:  cobra.ExactArgs(1),
	RunE:  runRoutinesRun,
}

func init() {
	rootCmd.AddCommand(routinesCmd)
	routinesCmd.AddCommand(routinesShowCmd)
	routinesCmd.AddCommand(routinesRunCmd)
	routinesCmd.PersistentFlags().String("workdir", "", "Workspace directory used for project routines (default: current directory)")
	routinesCmd.Flags().Bool("json", false, "Print routines as JSON")
	routinesShowCmd.Flags().Bool("json", false, "Print routine as JSON")

	routinesRunCmd.Flags().String("graph-area", "", "Graph or product area for the Datadog GraphQL routine")
	routinesRunCmd.Flags().Int("top-n", 20, "Number of slow GraphQL operations to inspect")
	routinesRunCmd.Flags().String("lookback", "24h", "Datadog lookback window, such as 24h or 7d")
	routinesRunCmd.Flags().String("environment", "prod", "Datadog environment tag")
	routinesRunCmd.Flags().String("service", "", "Optional Datadog service tag")
	routinesRunCmd.Flags().Int("max-remediations", 3, "Maximum high-confidence findings to route into remediation automation")
	routinesRunCmd.Flags().StringArray("param", nil, "Additional routine parameter as key=value")
	routinesRunCmd.Flags().String("provider", "", "Runtime provider override")
	routinesRunCmd.Flags().String("model", "", "Runtime model override")
	routinesRunCmd.Flags().String("run-id", "", "Runtime run ID")
	routinesRunCmd.Flags().String("report-out", "", "Report path (default: .boatman/routines/<id>/<run-id>.md, '-' disables file output)")
	routinesRunCmd.Flags().String("mcp-url", "", "Remote MCP URL override for the routine integration")
	routinesRunCmd.Flags().Bool("dry-run", false, "Build and print the routine request without calling a model")
	routinesRunCmd.Flags().Bool("json", false, "Print run or dry-run result as JSON")
}

type routineRunResult struct {
	RoutineID    string                `json:"routineId"`
	RunID        string                `json:"runId"`
	Provider     string                `json:"provider"`
	Model        string                `json:"model,omitempty"`
	ReportPath   string                `json:"reportPath,omitempty"`
	Integrations []integrations.Status `json:"integrations,omitempty"`
	Usage        *cost.Usage           `json:"usage,omitempty"`
	Report       string                `json:"report,omitempty"`
}

type routineDryRun struct {
	Routine      routines.Routine        `json:"routine"`
	Values       map[string]string       `json:"values"`
	Request      agentruntime.RunRequest `json:"request"`
	Integrations []integrations.Status   `json:"integrations,omitempty"`
	ReportPath   string                  `json:"reportPath,omitempty"`
}

func runRoutinesList(cmd *cobra.Command, args []string) error {
	workDir, err := routineWorkDir(cmd)
	if err != nil {
		return err
	}
	library, err := routines.ProjectLibrary(workDir)
	if err != nil {
		return err
	}
	items := library.List()
	for _, item := range items {
		if err := routines.Validate(item); err != nil {
			return err
		}
	}
	jsonOut, _ := cmd.Flags().GetBool("json")
	return writeRoutineList(cmd.OutOrStdout(), items, jsonOut)
}

func runRoutinesShow(cmd *cobra.Command, args []string) error {
	workDir, err := routineWorkDir(cmd)
	if err != nil {
		return err
	}
	library, err := routines.ProjectLibrary(workDir)
	if err != nil {
		return err
	}
	routine, ok := library.Get(args[0])
	if !ok {
		return fmt.Errorf("unknown routine %q", args[0])
	}
	if err := routines.Validate(routine); err != nil {
		return err
	}
	jsonOut, _ := cmd.Flags().GetBool("json")
	return writeRoutineShow(cmd.OutOrStdout(), routine, jsonOut)
}

func runRoutinesRun(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	workDir, err := routineWorkDir(cmd)
	if err != nil {
		return err
	}
	library, err := routines.ProjectLibrary(workDir)
	if err != nil {
		return err
	}
	routine, ok := library.Get(args[0])
	if !ok {
		return fmt.Errorf("unknown routine %q", args[0])
	}
	values, err := routineValuesFromCommand(cmd, routine)
	if err != nil {
		return err
	}
	values, err = routines.Values(routine, values)
	if err != nil {
		return err
	}
	runID, _ := cmd.Flags().GetString("run-id")
	if strings.TrimSpace(runID) == "" {
		runID = routines.NewRunID(routine.ID, time.Now())
	}
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	mcpURL, _ := cmd.Flags().GetString("mcp-url")
	statuses, mcpRefs, err := routineIntegrationRefs(ctx, routine, mcpURL, dryRun)
	if err != nil {
		return err
	}

	cfg, err := config.LoadRuntime()
	if err != nil {
		return fmt.Errorf("failed to load runtime config: %w", err)
	}
	providerName, _ := cmd.Flags().GetString("provider")
	if strings.TrimSpace(providerName) == "" {
		providerName = cfg.Runtime.ProviderFor(string(routine.Role), routine.Profile)
	}
	model, _ := cmd.Flags().GetString("model")
	req, err := routines.BuildRequest(routine, values, routines.BuildOptions{
		RunID:      runID,
		WorkDir:    workDir,
		Provider:   providerName,
		Model:      model,
		MCPServers: mcpRefs,
	})
	if err != nil {
		return err
	}
	reportPath, err := routineReportPath(cmd, routine, req.RunID, workDir)
	if err != nil {
		return err
	}

	jsonOut, _ := cmd.Flags().GetBool("json")
	if dryRun {
		return writeRoutineDryRun(cmd.OutOrStdout(), routineDryRun{
			Routine:      routine,
			Values:       values,
			Request:      req,
			Integrations: statuses,
			ReportPath:   reportPath,
		}, jsonOut)
	}

	provider, err := runtimeproviders.NewRegistryForConfig(cfg).MustGet(providerName)
	if err != nil {
		return err
	}
	report, usage, err := runtimeproviders.RunText(ctx, provider, req)
	if err != nil {
		return err
	}
	if strings.TrimSpace(reportPath) != "" && reportPath != "-" {
		if err := writeRoutineReport(reportPath, report); err != nil {
			return err
		}
	}
	return writeRoutineRunResult(cmd.OutOrStdout(), routineRunResult{
		RoutineID:    routine.ID,
		RunID:        req.RunID,
		Provider:     providerName,
		Model:        model,
		ReportPath:   reportPath,
		Integrations: statuses,
		Usage:        usage,
		Report:       report,
	}, jsonOut)
}

func writeRoutineList(out io.Writer, items []routines.Routine, jsonOut bool) error {
	if jsonOut {
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "  ")
		return encoder.Encode(items)
	}
	if len(items) == 0 {
		fmt.Fprintln(out, "No routines")
		return nil
	}
	fmt.Fprintf(out, "%-32s %-18s %-18s %-14s %s\n", "ID", "SCHEDULE", "INTEGRATIONS", "WORKFLOW", "DESCRIPTION")
	for _, item := range items {
		fmt.Fprintf(out, "%-32s %-18s %-18s %-14s %s\n",
			truncate(item.ID, 32),
			truncate(emptyDash(item.Schedule), 18),
			truncate(strings.Join(item.Integrations, ","), 18),
			truncate(item.WorkflowTemplate, 14),
			truncate(item.Description, 80),
		)
	}
	return nil
}

func writeRoutineShow(out io.Writer, routine routines.Routine, jsonOut bool) error {
	if jsonOut {
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "  ")
		return encoder.Encode(routine)
	}
	fmt.Fprintf(out, "Routine:     %s\n", routine.ID)
	fmt.Fprintf(out, "Name:        %s\n", routine.Name)
	fmt.Fprintf(out, "Description: %s\n", routine.Description)
	fmt.Fprintf(out, "Schedule:    %s\n", emptyDash(routine.Schedule))
	fmt.Fprintf(out, "Workflow:    %s\n", emptyDash(routine.WorkflowTemplate))
	fmt.Fprintf(out, "Role:        %s\n", routine.Role)
	fmt.Fprintf(out, "Profile:     %s\n", routine.Profile)
	fmt.Fprintf(out, "Integration: %s\n", strings.Join(routine.Integrations, ","))
	fmt.Fprintf(out, "Output:      %s %s\n\n", routine.Output.Format, routine.Output.DefaultPath)
	if len(routine.Parameters) > 0 {
		fmt.Fprintf(out, "%-16s %-10s %-9s %-12s %s\n", "PARAMETER", "TYPE", "REQUIRED", "DEFAULT", "DESCRIPTION")
		for _, param := range routine.Parameters {
			fmt.Fprintf(out, "%-16s %-10s %-9s %-12s %s\n",
				param.Name,
				param.Type,
				boolText(param.Required),
				emptyDash(param.Default),
				param.Description,
			)
		}
	}
	return nil
}

func writeRoutineDryRun(out io.Writer, dry routineDryRun, jsonOut bool) error {
	if jsonOut {
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "  ")
		return encoder.Encode(dry)
	}
	fmt.Fprintf(out, "Routine:   %s\n", dry.Routine.ID)
	fmt.Fprintf(out, "Run ID:    %s\n", dry.Request.RunID)
	fmt.Fprintf(out, "Provider:  %s\n", dry.Request.Provider)
	fmt.Fprintf(out, "Profile:   %s\n", dry.Request.Profile)
	fmt.Fprintf(out, "Workdir:   %s\n", dry.Request.WorkDir)
	fmt.Fprintf(out, "Report:    %s\n", emptyDash(dry.ReportPath))
	fmt.Fprintf(out, "MCP refs:  %d\n\n", len(dry.Request.MCPServers))
	if len(dry.Integrations) > 0 {
		fmt.Fprintf(out, "%-12s %-22s %-32s %s\n", "INTEGRATION", "STATE", "MISSING ENV", "MESSAGE")
		for _, status := range dry.Integrations {
			missing := "-"
			if len(status.MissingEnv) > 0 {
				missing = strings.Join(status.MissingEnv, ",")
			}
			fmt.Fprintf(out, "%-12s %-22s %-32s %s\n", status.Name, status.State, truncate(missing, 32), status.Message)
		}
		fmt.Fprintln(out)
	}
	keys := make([]string, 0, len(dry.Values))
	for key := range dry.Values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		fmt.Fprintf(out, "%s=%s\n", key, dry.Values[key])
	}
	return nil
}

func writeRoutineRunResult(out io.Writer, result routineRunResult, jsonOut bool) error {
	if jsonOut {
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "  ")
		return encoder.Encode(result)
	}
	fmt.Fprintf(out, "Routine: %s\n", result.RoutineID)
	fmt.Fprintf(out, "Run:     %s\n", result.RunID)
	fmt.Fprintf(out, "Provider: %s\n", result.Provider)
	if result.ReportPath != "" && result.ReportPath != "-" {
		fmt.Fprintf(out, "Report:  %s\n", result.ReportPath)
	}
	if result.Usage != nil {
		fmt.Fprintf(out, "Usage:   input=%d output=%d cost=%.4f\n", result.Usage.InputTokens, result.Usage.OutputTokens, result.Usage.TotalCostUSD)
	}
	if strings.TrimSpace(result.Report) != "" {
		fmt.Fprintln(out)
		fmt.Fprintln(out, strings.TrimSpace(result.Report))
	}
	return nil
}

func routineValuesFromCommand(cmd *cobra.Command, routine routines.Routine) (map[string]string, error) {
	values := map[string]string{}
	params, _ := cmd.Flags().GetStringArray("param")
	for _, entry := range params {
		key, value, ok := strings.Cut(entry, "=")
		if !ok || strings.TrimSpace(key) == "" {
			return nil, fmt.Errorf("--param must be key=value, got %q", entry)
		}
		values[strings.TrimSpace(key)] = strings.TrimSpace(value)
	}
	if graphArea, _ := cmd.Flags().GetString("graph-area"); strings.TrimSpace(graphArea) != "" {
		values["graph_area"] = graphArea
	}
	if cmd.Flags().Changed("top-n") && routineHasParameter(routine, "top_n") {
		topN, _ := cmd.Flags().GetInt("top-n")
		values["top_n"] = strconv.Itoa(topN)
	}
	if cmd.Flags().Changed("lookback") && routineHasParameter(routine, "lookback") {
		lookback, _ := cmd.Flags().GetString("lookback")
		values["lookback"] = lookback
	}
	if cmd.Flags().Changed("environment") && routineHasParameter(routine, "environment") {
		environment, _ := cmd.Flags().GetString("environment")
		values["environment"] = environment
	}
	if service, _ := cmd.Flags().GetString("service"); strings.TrimSpace(service) != "" {
		values["service"] = service
	}
	if cmd.Flags().Changed("max-remediations") && routineHasParameter(routine, "max_remediations") {
		maxRemediations, _ := cmd.Flags().GetInt("max-remediations")
		values["max_remediations"] = strconv.Itoa(maxRemediations)
	}
	return values, nil
}

func routineHasParameter(routine routines.Routine, name string) bool {
	for _, param := range routine.Parameters {
		if param.Name == name {
			return true
		}
	}
	return false
}

func routineWorkDir(cmd *cobra.Command) (string, error) {
	workDir := ""
	if flag := cmd.Flags().Lookup("workdir"); flag != nil {
		workDir = flag.Value.String()
	} else if flag := cmd.InheritedFlags().Lookup("workdir"); flag != nil {
		workDir = flag.Value.String()
	}
	if strings.TrimSpace(workDir) == "" {
		return os.Getwd()
	}
	return filepath.Abs(workDir)
}

func routineIntegrationRefs(ctx context.Context, routine routines.Routine, mcpURL string, dryRun bool) ([]integrations.Status, []agentruntime.MCPServerRef, error) {
	catalog := integrations.DefaultCatalog()
	manager := integrations.NewManager(catalog)
	statuses := make([]integrations.Status, 0, len(routine.Integrations))
	refs := make([]agentruntime.MCPServerRef, 0, len(routine.Integrations))
	for _, name := range routine.Integrations {
		item, ok := catalog.Lookup(name)
		if !ok {
			return nil, nil, fmt.Errorf("routine %q references unknown integration %q", routine.ID, name)
		}
		opts := integrations.ResolveOptions{
			Enabled: true,
			Env:     envForIntegration(item),
		}
		if strings.TrimSpace(mcpURL) != "" {
			opts.URL = strings.TrimSpace(mcpURL)
		}
		conn, err := manager.Connect(ctx, item.Name, opts)
		statuses = append(statuses, conn.Status)
		if err != nil {
			return statuses, nil, err
		}
		if conn.State != integrations.StateConnected {
			if dryRun {
				continue
			}
			missing := ""
			if len(conn.Status.MissingEnv) > 0 {
				missing = fmt.Sprintf(" missing env: %s", strings.Join(conn.Status.MissingEnv, ","))
			}
			return statuses, nil, fmt.Errorf("routine %q integration %q is %s:%s %s", routine.ID, item.Name, conn.State, missing, conn.Status.Message)
		}
		if conn.MCP != nil {
			ref := *conn.MCP
			ref.Env = sanitizeMCPEnv(ref.Env, item)
			refs = append(refs, ref)
		}
	}
	return statuses, refs, nil
}

func sanitizeMCPEnv(env map[string]string, item integrations.Integration) map[string]string {
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

func routineReportPath(cmd *cobra.Command, routine routines.Routine, runID, workDir string) (string, error) {
	reportOut, _ := cmd.Flags().GetString("report-out")
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

func emptyDash(value string) string {
	if strings.TrimSpace(value) == "" {
		return "-"
	}
	return value
}
