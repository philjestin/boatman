package cli

import (
	"context"
	"fmt"

	"github.com/philjestin/boatman-ecosystem/harness/scaffold"
	"github.com/philjestin/boatman-ecosystem/harness/scaffold/agent"
	"github.com/spf13/cobra"
)

// scaffoldCmd generates a new harness project from templates.
var scaffoldCmd = &cobra.Command{
	Use:   "scaffold",
	Short: "Generate a new AI agent harness project",
	Long: `Generate a new Go project with stub implementations of the harness
runner role interfaces (Developer, Reviewer, and optionally Tester/Planner).

The generated project compiles and runs immediately. Stubs return placeholder
values with provider-specific guidance in comments.

Examples:
  boatman scaffold --name github.com/user/my-agent --output ./my-agent
  boatman scaffold --name github.com/user/my-agent --provider claude --with-planner
  boatman scaffold --name github.com/user/my-agent --provider openai --with-tester --lang python`,
	RunE: runScaffold,
}

// enhanceCmd uses Claude to fill in stub implementations.
var enhanceCmd = &cobra.Command{
	Use:   "enhance",
	Short: "Use Claude to fill in scaffold stubs with real implementations",
	Long: `Enhance a scaffolded project by using the Claude CLI to replace stub
implementations with real LLM-calling code.

Requires the claude CLI to be installed and available on PATH.

Example:
  boatman scaffold enhance --dir ./my-agent`,
	RunE: runEnhance,
}

func init() {
	rootCmd.AddCommand(scaffoldCmd)
	scaffoldCmd.AddCommand(enhanceCmd)

	// Scaffold flags.
	scaffoldCmd.Flags().String("name", "", "Go module path (e.g. github.com/user/my-agent)")
	scaffoldCmd.Flags().String("output", "", "Output directory (defaults to last segment of --name)")
	scaffoldCmd.Flags().String("provider", "generic", "LLM provider: claude, openai, ollama, generic")
	scaffoldCmd.Flags().String("lang", "generic", "Target project language: go, typescript, python, ruby, generic")
	scaffoldCmd.Flags().Bool("with-planner", false, "Generate Planner implementation")
	scaffoldCmd.Flags().Bool("with-tester", false, "Generate Tester implementation")
	scaffoldCmd.Flags().Bool("with-cost-tracking", false, "Wire up cost.Tracker")
	scaffoldCmd.Flags().Int("max-iterations", 3, "Default max review iterations")
	scaffoldCmd.Flags().String("review-criteria", "", "Optional description for reviewer focus")

	_ = scaffoldCmd.MarkFlagRequired("name")

	// Enhance flags.
	enhanceCmd.Flags().String("dir", ".", "Directory containing the scaffolded project")
	enhanceCmd.Flags().String("provider", "", "LLM provider (auto-detected from project if not set)")
	enhanceCmd.Flags().String("lang", "", "Target project language (auto-detected from project if not set)")
	enhanceCmd.Flags().String("model", "", "Claude model (default: claude-sonnet-4-20250514)")
}

func runScaffold(cmd *cobra.Command, _ []string) error {
	name, _ := cmd.Flags().GetString("name")
	output, _ := cmd.Flags().GetString("output")
	provider, _ := cmd.Flags().GetString("provider")
	lang, _ := cmd.Flags().GetString("lang")
	withPlanner, _ := cmd.Flags().GetBool("with-planner")
	withTester, _ := cmd.Flags().GetBool("with-tester")
	withCost, _ := cmd.Flags().GetBool("with-cost-tracking")
	maxIter, _ := cmd.Flags().GetInt("max-iterations")
	criteria, _ := cmd.Flags().GetString("review-criteria")

	cfg := scaffold.ScaffoldConfig{
		ProjectName:         name,
		OutputDir:           output,
		Provider:            scaffold.LLMProvider(provider),
		ProjectLang:         scaffold.ProjectLanguage(lang),
		IncludePlanner:      withPlanner,
		IncludeTester:       withTester,
		IncludeCostTracking: withCost,
		MaxIterations:       maxIter,
		ReviewCriteria:      criteria,
	}

	result, err := scaffold.Generate(cfg)
	if err != nil {
		return fmt.Errorf("scaffold failed: %w", err)
	}

	fmt.Printf("Project generated at %s\n", result.OutputDir)
	fmt.Printf("Files created:\n")
	for _, f := range result.FilesCreated {
		fmt.Printf("  %s\n", f)
	}
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Printf("  cd %s\n", result.OutputDir)
	fmt.Println("  go build ./...")
	fmt.Println("  # Edit the stub files to add your LLM integration")
	fmt.Println()
	fmt.Println("Or use Claude to fill in the stubs:")
	fmt.Printf("  boatman scaffold enhance --dir %s\n", result.OutputDir)

	return nil
}

func runEnhance(cmd *cobra.Command, _ []string) error {
	dir, _ := cmd.Flags().GetString("dir")
	provider, _ := cmd.Flags().GetString("provider")
	lang, _ := cmd.Flags().GetString("lang")
	model, _ := cmd.Flags().GetString("model")

	cfg := agent.EnhanceConfig{
		ProjectDir:  dir,
		Provider:    scaffold.LLMProvider(provider),
		ProjectLang: scaffold.ProjectLanguage(lang),
		Model:       model,
	}

	ctx := context.Background()
	if err := agent.Enhance(ctx, cfg); err != nil {
		return fmt.Errorf("enhance failed: %w", err)
	}

	return nil
}
