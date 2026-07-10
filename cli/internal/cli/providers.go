package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

	agentruntime "github.com/philjestin/boatman-ecosystem/shared/agentruntime"
	"github.com/philjestin/boatmanmode/internal/config"
	runtimeproviders "github.com/philjestin/boatmanmode/internal/providers"
	"github.com/spf13/cobra"
)

var runtimeProvidersCmd = &cobra.Command{
	Use:   "providers",
	Short: "Inspect registered model providers",
	Long: `Inspect the model provider adapters compiled into this Boatman build.

The output is local capability metadata, so it is safe to run without provider
API credentials. Use --json when feeding the result into docs, checks, or other
automation.`,
	RunE: runProviders,
}

var runtimeProvidersCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Validate configured provider routing",
	RunE:  runProvidersCheck,
}

func init() {
	rootCmd.AddCommand(runtimeProvidersCmd)
	runtimeProvidersCmd.AddCommand(runtimeProvidersCheckCmd)
	runtimeProvidersCmd.Flags().Bool("json", false, "Print provider capabilities as JSON")
}

type providerInfo struct {
	Name         string                    `json:"name"`
	Capabilities agentruntime.Capabilities `json:"capabilities"`
}

func runProviders(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadRuntime()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	registry := runtimeproviders.NewRegistryForConfig(cfg)
	infos, err := collectProviderInfos(cmd.Context(), registry)
	if err != nil {
		return err
	}
	jsonOut, _ := cmd.Flags().GetBool("json")
	return writeProviders(cmd.OutOrStdout(), infos, jsonOut)
}

func runProvidersCheck(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadRuntime()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	registry := runtimeproviders.NewRegistryForConfig(cfg)
	missing := missingConfiguredProviders(cfg, registry)
	if len(missing) > 0 {
		return fmt.Errorf("unknown runtime provider route(s): %s", strings.Join(missing, ", "))
	}
	fmt.Fprintln(cmd.OutOrStdout(), "Provider routing ok")
	return nil
}

func collectProviderInfos(ctx context.Context, registry *runtimeproviders.Registry) ([]providerInfo, error) {
	names := registry.Names()
	infos := make([]providerInfo, 0, len(names))
	for _, name := range names {
		provider, err := registry.MustGet(name)
		if err != nil {
			return nil, err
		}
		caps, err := provider.Capabilities(ctx)
		if err != nil {
			return nil, fmt.Errorf("provider %q capabilities: %w", name, err)
		}
		if caps.Provider == "" {
			caps.Provider = name
		}
		infos = append(infos, providerInfo{Name: name, Capabilities: caps})
	}
	return infos, nil
}

func missingConfiguredProviders(cfg *config.Config, registry *runtimeproviders.Registry) []string {
	if cfg == nil {
		return nil
	}
	var missing []string
	check := func(label, providerName string) {
		providerName = strings.TrimSpace(providerName)
		if providerName == "" {
			return
		}
		if _, ok := registry.Get(providerName); !ok {
			missing = append(missing, fmt.Sprintf("%s=%s", label, providerName))
		}
	}

	check("default", cfg.Runtime.ProviderFor("", ""))
	for _, role := range sortedKeys(cfg.Runtime.RoleProviders) {
		check("role."+role, cfg.Runtime.RoleProviders[role])
	}
	for _, profile := range sortedKeys(cfg.Runtime.ProfileProviders) {
		check("profile."+profile, cfg.Runtime.ProfileProviders[profile])
	}
	return missing
}

func sortedKeys(values map[string]string) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func writeProviders(out io.Writer, infos []providerInfo, jsonOut bool) error {
	if jsonOut {
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "  ")
		return encoder.Encode(infos)
	}
	if len(infos) == 0 {
		fmt.Fprintln(out, "No registered model providers")
		return nil
	}

	fmt.Fprintf(out, "%-18s %-9s %-6s %-4s %-9s %-10s %-6s %s\n",
		"PROVIDER", "STREAMING", "TOOLS", "MCP", "APPROVAL", "STRUCTURED", "USAGE", "EXPERIMENTAL")
	for _, info := range infos {
		caps := info.Capabilities
		fmt.Fprintf(out, "%-18s %-9s %-6s %-4s %-9s %-10s %-6s %s\n",
			truncate(info.Name, 18),
			boolText(caps.SupportsStreaming),
			boolText(caps.SupportsToolCalls),
			boolText(caps.SupportsMCP),
			boolText(caps.SupportsApprovals),
			boolText(caps.SupportsStructuredOut),
			boolText(caps.SupportsUsage),
			strings.Join(caps.Experimental, ","),
		)
	}
	return nil
}

func boolText(value bool) string {
	if value {
		return "yes"
	}
	return "no"
}
