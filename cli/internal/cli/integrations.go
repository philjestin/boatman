package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	agentruntime "github.com/philjestin/boatman-ecosystem/shared/agentruntime"
	"github.com/philjestin/boatman-ecosystem/shared/agentruntime/integrations"
	"github.com/philjestin/boatman-ecosystem/shared/agentruntime/runstore"
	"github.com/spf13/cobra"
)

var integrationsCmd = &cobra.Command{
	Use:   "integrations",
	Short: "Inspect service integration descriptors",
	Long: `Inspect Boatman's shared service integration catalog.

These commands do not open network connections. They validate local descriptor
metadata and environment configuration before providers or desktop sessions try
to use the integration.`,
	RunE: runIntegrationsList,
}

var integrationsCheckCmd = &cobra.Command{
	Use:   "check [name...]",
	Short: "Check integration configuration from environment variables",
	RunE:  runIntegrationsCheck,
}

func init() {
	rootCmd.AddCommand(integrationsCmd)
	integrationsCmd.AddCommand(integrationsCheckCmd)
	integrationsCmd.Flags().Bool("json", false, "Print integration descriptors as JSON")
	integrationsCheckCmd.Flags().Bool("json", false, "Print integration statuses as JSON")
	integrationsCheckCmd.Flags().Bool("emit-events", false, "Print normalized runtime events instead of the status table")
	integrationsCheckCmd.Flags().String("run-id", "", "Runtime run ID for emitted or recorded integration check events")
}

func runIntegrationsList(cmd *cobra.Command, args []string) error {
	items := integrations.DefaultCatalog().List()
	jsonOut, _ := cmd.Flags().GetBool("json")
	return writeIntegrations(cmd.OutOrStdout(), items, jsonOut)
}

func runIntegrationsCheck(cmd *cobra.Command, args []string) error {
	catalog := integrations.DefaultCatalog()
	names := args
	if len(names) == 0 {
		for _, item := range catalog.List() {
			names = append(names, item.Name)
		}
	}

	broker := integrations.NewBroker(catalog)
	statuses := make([]integrations.Status, 0, len(names))
	for _, name := range names {
		item, ok := catalog.Lookup(name)
		if !ok {
			return fmt.Errorf("unknown integration %q", name)
		}
		status, err := broker.Status(context.Background(), item.Name, integrations.ResolveOptions{
			Enabled: true,
			Env:     envForIntegration(item),
		})
		if err != nil {
			return err
		}
		statuses = append(statuses, status)
	}

	emitEvents, _ := cmd.Flags().GetBool("emit-events")
	runID, _ := cmd.Flags().GetString("run-id")
	if emitEvents || runtimeStoreEnabled() {
		events, err := integrationCheckEvents(statuses, runID)
		if err != nil {
			return err
		}
		if err := recordIntegrationEvents(context.Background(), events); err != nil {
			return err
		}
		if emitEvents {
			return writeRuntimeEvents(cmd.OutOrStdout(), events)
		}
	}

	jsonOut, _ := cmd.Flags().GetBool("json")
	return writeIntegrationStatuses(cmd.OutOrStdout(), statuses, jsonOut)
}

func writeIntegrations(out io.Writer, items []integrations.Integration, jsonOut bool) error {
	if jsonOut {
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "  ")
		return encoder.Encode(items)
	}
	if len(items) == 0 {
		fmt.Fprintln(out, "No integrations")
		return nil
	}
	fmt.Fprintf(out, "%-12s %-24s %-32s %s\n", "NAME", "DISPLAY", "REQUIRED ENV", "DESCRIPTION")
	for _, item := range items {
		required := "-"
		if item.MCP != nil {
			if keys := item.MCP.RequiredEnv(); len(keys) > 0 {
				required = strings.Join(keys, ",")
			}
		}
		fmt.Fprintf(out, "%-12s %-24s %-32s %s\n",
			truncate(item.Name, 12),
			truncate(item.DisplayName, 24),
			truncate(required, 32),
			item.Description,
		)
	}
	return nil
}

func writeIntegrationStatuses(out io.Writer, statuses []integrations.Status, jsonOut bool) error {
	if jsonOut {
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "  ")
		return encoder.Encode(statuses)
	}
	if len(statuses) == 0 {
		fmt.Fprintln(out, "No integration statuses")
		return nil
	}
	fmt.Fprintf(out, "%-12s %-22s %-32s %s\n", "NAME", "STATE", "MISSING ENV", "MESSAGE")
	for _, status := range statuses {
		missing := "-"
		if len(status.MissingEnv) > 0 {
			missing = strings.Join(status.MissingEnv, ",")
		}
		fmt.Fprintf(out, "%-12s %-22s %-32s %s\n",
			truncate(status.Name, 12),
			status.State,
			truncate(missing, 32),
			status.Message,
		)
	}
	return nil
}

func envForIntegration(item integrations.Integration) map[string]string {
	if item.MCP == nil {
		return nil
	}
	env := make(map[string]string, len(item.MCP.Env))
	for key, fallback := range item.MCP.Env {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			env[key] = value
		} else {
			env[key] = fallback
		}
	}
	return env
}

func integrationCheckEvents(statuses []integrations.Status, runID string) ([]agentruntime.Event, error) {
	if strings.TrimSpace(runID) == "" {
		runID = fmt.Sprintf("integration-check-%d", time.Now().UnixNano())
	}
	broker := integrations.NewBroker(integrations.DefaultCatalog())
	events := make([]agentruntime.Event, 0, len(statuses)+2)
	started := agentruntime.NewEvent(agentruntime.EventRunStarted)
	started.RunID = runID
	started.Provider = "boatman-cli"
	started.Role = agentruntime.RoleIntegration
	started.Status = agentruntime.StatusStarted
	started.Name = "integration-check"
	events = append(events, started)

	terminal := agentruntime.StatusSucceeded
	for _, status := range statuses {
		event := broker.Event(status)
		event.RunID = runID
		event.Provider = "boatman-cli"
		event.Role = agentruntime.RoleIntegration
		events = append(events, event)
		if event.Status == agentruntime.StatusWaiting && terminal != agentruntime.StatusFailed {
			terminal = agentruntime.StatusWaiting
		}
		if event.Status == agentruntime.StatusFailed {
			terminal = agentruntime.StatusFailed
		}
	}

	completed := agentruntime.NewEvent(agentruntime.EventRunCompleted)
	completed.RunID = runID
	completed.Provider = "boatman-cli"
	completed.Role = agentruntime.RoleIntegration
	completed.Status = terminal
	completed.Name = "integration-check"
	events = append(events, completed)
	return events, nil
}

func recordIntegrationEvents(ctx context.Context, events []agentruntime.Event) error {
	if len(events) == 0 {
		return nil
	}
	req := agentruntime.RunRequest{
		RunID:    events[0].RunID,
		Provider: "boatman-cli",
		Role:     agentruntime.RoleIntegration,
		Profile:  "integration-check",
	}
	store, enabled, err := runstore.ForRequest(req)
	if err != nil {
		return err
	}
	if !enabled {
		return nil
	}
	if err := store.StartRun(ctx, req); err != nil {
		return fmt.Errorf("start integration run store: %w", err)
	}
	for _, event := range events {
		if err := store.Append(ctx, event); err != nil {
			return fmt.Errorf("record integration event: %w", err)
		}
	}
	return nil
}

func runtimeStoreEnabled() bool {
	return os.Getenv("BOATMAN_RUNTIME_STORE") == "1" || strings.TrimSpace(os.Getenv("BOATMAN_RUNTIME_STORE_DIR")) != ""
}

func writeRuntimeEvents(out io.Writer, events []agentruntime.Event) error {
	encoder := json.NewEncoder(out)
	for _, event := range events {
		if err := encoder.Encode(event); err != nil {
			return err
		}
	}
	return nil
}
