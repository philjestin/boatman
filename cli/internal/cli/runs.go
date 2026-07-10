package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/philjestin/boatman-ecosystem/shared/agentruntime"
	"github.com/philjestin/boatman-ecosystem/shared/agentruntime/runstore"
	"github.com/spf13/cobra"
)

var runsCmd = &cobra.Command{
	Use:   "runs",
	Short: "Inspect recorded runtime runs",
	Long: `Inspect normalized provider runtime runs recorded by BOATMAN_RUNTIME_STORE
or BOATMAN_RUNTIME_STORE_DIR.`,
}

var runsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List recorded runtime runs",
	RunE: func(cmd *cobra.Command, args []string) error {
		store := runStoreFromCommand(cmd)
		runs, err := store.ListRuns(context.Background())
		if err != nil {
			return err
		}
		out := cmd.OutOrStdout()
		if len(runs) == 0 {
			fmt.Fprintln(out, "No recorded runtime runs")
			return nil
		}
		fmt.Fprintf(out, "%-32s %-10s %-18s %-12s %-20s %s\n", "RUN ID", "STATUS", "PROVIDER", "ROLE", "UPDATED", "EVENTS")
		for _, run := range runs {
			fmt.Fprintf(out, "%-32s %-10s %-18s %-12s %-20s %d\n",
				truncate(run.RunID, 32),
				run.Status,
				truncate(run.Provider, 18),
				run.Role,
				formatRunTime(run.UpdatedAt),
				run.EventCount,
			)
		}
		return nil
	},
}

var runsShowCmd = &cobra.Command{
	Use:   "show [run-id]",
	Short: "Show a recorded runtime run",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		store := runStoreFromCommand(cmd)
		metadata, events, err := store.LoadRun(context.Background(), args[0])
		if err != nil {
			return err
		}
		jsonOut, _ := cmd.Flags().GetBool("json")
		return writeRunShow(cmd.OutOrStdout(), metadata, events, jsonOut)
	},
}

var runsArtifactsCmd = &cobra.Command{
	Use:   "artifacts [run-id]",
	Short: "List artifacts recorded for a runtime run",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		store := runStoreFromCommand(cmd)
		artifacts, err := store.ListArtifacts(context.Background(), args[0])
		if err != nil {
			return err
		}
		jsonOut, _ := cmd.Flags().GetBool("json")
		return writeRunArtifacts(cmd.OutOrStdout(), artifacts, jsonOut)
	},
}

var runsRequestCmd = &cobra.Command{
	Use:   "request [run-id]",
	Short: "Show the original runtime run request",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		store := runStoreFromCommand(cmd)
		req, err := store.LoadRequest(context.Background(), args[0])
		if err != nil {
			return err
		}
		jsonOut, _ := cmd.Flags().GetBool("json")
		return writeRunRequest(cmd.OutOrStdout(), req, jsonOut)
	},
}

func init() {
	rootCmd.AddCommand(runsCmd)
	runsCmd.AddCommand(runsListCmd)
	runsCmd.AddCommand(runsShowCmd)
	runsCmd.AddCommand(runsArtifactsCmd)
	runsCmd.AddCommand(runsRequestCmd)
	runsCmd.PersistentFlags().String("store-dir", "", "Runtime run store directory (default: BOATMAN_RUNTIME_STORE_DIR or .boatman/runs)")
	runsShowCmd.Flags().Bool("json", false, "Print metadata and events as JSON")
	runsArtifactsCmd.Flags().Bool("json", false, "Print artifacts as JSON")
	runsRequestCmd.Flags().Bool("json", false, "Print request as JSON")
}

func runStoreFromCommand(cmd *cobra.Command) *runstore.FileStore {
	dir, _ := cmd.Flags().GetString("store-dir")
	if strings.TrimSpace(dir) == "" {
		dir, _ = cmd.InheritedFlags().GetString("store-dir")
	}
	if strings.TrimSpace(dir) == "" {
		dir = os.Getenv("BOATMAN_RUNTIME_STORE_DIR")
	}
	if strings.TrimSpace(dir) == "" {
		cwd, err := os.Getwd()
		if err != nil {
			cwd = "."
		}
		dir = filepath.Join(cwd, ".boatman", "runs")
	}
	return runstore.NewFileStore(dir)
}

func writeRunShow(out io.Writer, metadata runstore.RunMetadata, events []agentruntime.Event, jsonOut bool) error {
	if jsonOut {
		return json.NewEncoder(out).Encode(struct {
			Metadata runstore.RunMetadata `json:"metadata"`
			Events   []agentruntime.Event `json:"events"`
		}{Metadata: metadata, Events: events})
	}

	fmt.Fprintf(out, "Run:      %s\n", metadata.RunID)
	fmt.Fprintf(out, "Status:   %s\n", metadata.Status)
	fmt.Fprintf(out, "Provider: %s\n", metadata.Provider)
	fmt.Fprintf(out, "Model:    %s\n", metadata.Model)
	fmt.Fprintf(out, "Role:     %s\n", metadata.Role)
	fmt.Fprintf(out, "Profile:  %s\n", metadata.Profile)
	fmt.Fprintf(out, "Workdir:  %s\n", metadata.WorkDir)
	fmt.Fprintf(out, "Started:  %s\n", formatRunTime(metadata.StartedAt))
	fmt.Fprintf(out, "Updated:  %s\n", formatRunTime(metadata.UpdatedAt))
	if metadata.EndedAt != nil {
		fmt.Fprintf(out, "Ended:    %s\n", formatRunTime(*metadata.EndedAt))
	}
	fmt.Fprintf(out, "Events:   %d\n\n", len(events))
	fmt.Fprintf(out, "Artifacts: %d\n\n", metadata.ArtifactCount)
	for _, event := range events {
		fmt.Fprintf(out, "%s %-18s %-12s %s\n",
			formatRunTime(event.Timestamp),
			event.Type,
			event.Status,
			eventSummary(event),
		)
	}
	return nil
}

func writeRunArtifacts(out io.Writer, artifacts []runstore.ArtifactRecord, jsonOut bool) error {
	if jsonOut {
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "  ")
		return encoder.Encode(artifacts)
	}
	if len(artifacts) == 0 {
		fmt.Fprintln(out, "No artifacts")
		return nil
	}
	fmt.Fprintf(out, "%-12s %-42s %-24s %-20s %s\n", "KIND", "PATH/URL", "EVENT", "UPDATED", "MESSAGE")
	for _, artifact := range artifacts {
		location := artifact.Path
		if location == "" {
			location = artifact.URL
		}
		fmt.Fprintf(out, "%-12s %-42s %-24s %-20s %s\n",
			truncate(artifact.Kind, 12),
			truncate(location, 42),
			artifact.EventType,
			formatRunTime(artifact.LastSeenAt),
			truncate(strings.ReplaceAll(artifact.Message, "\n", " "), 80),
		)
	}
	return nil
}

func writeRunRequest(out io.Writer, req agentruntime.RunRequest, jsonOut bool) error {
	if jsonOut {
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "  ")
		return encoder.Encode(req)
	}
	fmt.Fprintf(out, "Run:       %s\n", req.RunID)
	fmt.Fprintf(out, "Provider:  %s\n", req.Provider)
	fmt.Fprintf(out, "Model:     %s\n", req.Model)
	fmt.Fprintf(out, "Role:      %s\n", req.Role)
	fmt.Fprintf(out, "Profile:   %s\n", req.Profile)
	fmt.Fprintf(out, "Workdir:   %s\n", req.WorkDir)
	fmt.Fprintf(out, "Messages:  %d\n", len(req.Messages))
	fmt.Fprintf(out, "Tools:     %d\n", len(req.Tools))
	fmt.Fprintf(out, "MCP:       %d\n", len(req.MCPServers))
	if req.OutputSchema != nil {
		fmt.Fprintf(out, "Schema:    %s\n", req.OutputSchema.Name)
	}
	if req.ApprovalPolicy != "" {
		fmt.Fprintf(out, "Approval:  %s\n", req.ApprovalPolicy)
	}
	if req.Background {
		fmt.Fprintln(out, "Background: true")
	}
	if len(req.Metadata) > 0 {
		keys := make([]string, 0, len(req.Metadata))
		for key := range req.Metadata {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		fmt.Fprintf(out, "Metadata:  %s\n", strings.Join(keys, ","))
	}
	return nil
}

func eventSummary(event agentruntime.Event) string {
	switch {
	case event.Type == agentruntime.EventIntegrationState:
		return integrationEventSummary(event)
	case event.Type == agentruntime.EventMemoryLoaded:
		return memoryEventSummary(event)
	case event.Type == agentruntime.EventArtifactChanged && event.Artifact != nil:
		return artifactEventSummary(event)
	case event.Message != "":
		return truncate(strings.ReplaceAll(event.Message, "\n", " "), 120)
	case event.Tool != nil:
		if event.Tool.IsError {
			return fmt.Sprintf("%s error", event.Tool.Name)
		}
		return event.Tool.Name
	case event.Usage != nil:
		return fmt.Sprintf("input=%d output=%d cost=%.4f", event.Usage.InputTokens, event.Usage.OutputTokens, event.Usage.TotalCostUSD)
	case event.Schema != nil:
		return fmt.Sprintf("%s valid=%t", event.Schema.Name, event.Schema.Valid)
	case len(event.Raw) > 0:
		return truncate(string(event.Raw), 120)
	default:
		return event.Name
	}
}

func artifactEventSummary(event agentruntime.Event) string {
	location := event.Artifact.Path
	if location == "" {
		location = event.Artifact.URL
	}
	if location == "" {
		location = event.Artifact.Kind
	}
	if event.Message != "" {
		return fmt.Sprintf("%s %s", location, strings.ReplaceAll(event.Message, "\n", " "))
	}
	return location
}

func memoryEventSummary(event agentruntime.Event) string {
	docs := documentsFromEvent(event.Data["documents"])
	if len(docs) == 0 {
		if event.Message != "" {
			return strings.ReplaceAll(event.Message, "\n", " ")
		}
		return "loaded 0 memory documents"
	}
	return fmt.Sprintf("loaded=%s", strings.Join(docs, ","))
}

func integrationEventSummary(event agentruntime.Event) string {
	state := fmt.Sprint(event.Data["state"])
	if state == "" || state == "<nil>" {
		state = string(event.Status)
	}
	missing := stringList(event.Data["missing_env"])
	if len(missing) > 0 {
		return fmt.Sprintf("%s missing=%s", state, strings.Join(missing, ","))
	}
	if event.Message != "" {
		return fmt.Sprintf("%s %s", state, strings.ReplaceAll(event.Message, "\n", " "))
	}
	return state
}

func documentsFromEvent(value any) []string {
	switch typed := value.(type) {
	case []map[string]any:
		out := make([]string, 0, len(typed))
		for _, item := range typed {
			if id := strings.TrimSpace(fmt.Sprint(item["id"])); id != "" {
				out = append(out, id)
			}
		}
		return out
	case []any:
		out := make([]string, 0, len(typed))
		for _, item := range typed {
			if fields, ok := item.(map[string]any); ok {
				if id := strings.TrimSpace(fmt.Sprint(fields["id"])); id != "" {
					out = append(out, id)
				}
			}
		}
		return out
	default:
		return nil
	}
}

func stringList(value any) []string {
	switch typed := value.(type) {
	case []string:
		return typed
	case []any:
		out := make([]string, 0, len(typed))
		for _, item := range typed {
			if text := strings.TrimSpace(fmt.Sprint(item)); text != "" {
				out = append(out, text)
			}
		}
		return out
	default:
		return nil
	}
}

func formatRunTime(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	return t.Local().Format("2006-01-02 15:04:05")
}

func truncate(value string, max int) string {
	if len(value) <= max {
		return value
	}
	if max <= 1 {
		return value[:max]
	}
	if max <= 3 {
		return value[:max]
	}
	return value[:max-3] + "..."
}
