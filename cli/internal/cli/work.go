package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/philjestin/boatmanmode/internal/agent"
	"github.com/philjestin/boatmanmode/internal/config"
	"github.com/philjestin/boatmanmode/internal/linear"
	"github.com/philjestin/boatmanmode/internal/plan"
	"github.com/philjestin/boatmanmode/internal/planner"
	"github.com/philjestin/boatmanmode/internal/task"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// workCmd represents the work command - the main workflow executor.
var workCmd = &cobra.Command{
	Use:   "work [ticket-id-or-prompt]",
	Short: "Execute a task from Linear, prompt, or file",
	Long: `Execute a development task from multiple input sources.

Input modes:
  1. Linear ticket (default):    boatman work ENG-123
  2. Inline prompt:              boatman work --prompt "Add authentication"
  3. File-based prompt:          boatman work --file ./task.txt

The agent will:
  1. Prepare the task
  2. Create a git worktree for isolated development
  3. Execute the task using Claude
  4. Review with ScottBott
  5. Refactor if needed until review passes
  6. Create a pull request

Flags like --title and --branch-name can override auto-generated values for prompt/file mode.`,
	Args: cobra.ExactArgs(1),
	RunE: runWork,
}

func init() {
	rootCmd.AddCommand(workCmd)

	// Existing flags
	workCmd.Flags().Int("max-iterations", 3, "Maximum review/refactor iterations")
	workCmd.Flags().String("base-branch", "main", "Base branch for worktree")
	workCmd.Flags().Bool("auto-pr", true, "Automatically create PR on success")
	workCmd.Flags().Bool("dry-run", false, "Run without making changes")
	workCmd.Flags().Int("timeout", 60, "Timeout in minutes for each Claude agent")
	workCmd.Flags().String("review-skill", "peer-review", "Claude skill/agent to use for code review")

	// New input mode flags
	workCmd.Flags().Bool("prompt", false, "Treat argument as inline prompt text")
	workCmd.Flags().Bool("file", false, "Read prompt from file")
	workCmd.Flags().String("title", "", "Override auto-generated task title (prompt/file mode only)")
	workCmd.Flags().String("branch-name", "", "Override auto-generated branch name (prompt/file mode only)")

	// Pre-generated plan from triage pipeline
	workCmd.Flags().String("plan-file", "", "Path to a pre-generated triage plan JSON file (skips planning step)")

	// Resume a failed execution from review/refactor stage
	workCmd.Flags().Bool("resume", false, "Resume a failed execution from the review/refactor stage using the existing worktree")

	viper.BindPFlag("max_iterations", workCmd.Flags().Lookup("max-iterations"))
	viper.BindPFlag("base_branch", workCmd.Flags().Lookup("base-branch"))
	viper.BindPFlag("auto_pr", workCmd.Flags().Lookup("auto-pr"))
	viper.BindPFlag("timeout", workCmd.Flags().Lookup("timeout"))
	viper.BindPFlag("review_skill", workCmd.Flags().Lookup("review-skill"))
}

// runWork executes the main workflow for a given task.
func runWork(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Validate and parse input mode
	t, err := parseTaskInput(cmd, args, cfg)
	if err != nil {
		return err
	}

	dryRun, _ := cmd.Flags().GetBool("dry-run")
	if dryRun {
		fmt.Println("🏃 Dry run mode - no changes will be made")
	}

	a, err := agent.New(cfg)
	if err != nil {
		return fmt.Errorf("failed to create agent: %w", err)
	}

	// Check for resume mode
	resume, _ := cmd.Flags().GetBool("resume")
	if resume {
		fmt.Println("♻️  Resume mode — skipping to review/refactor using existing worktree")
		result, err := a.ResumeWork(ctx, t)
		if err != nil {
			return fmt.Errorf("resume failed: %w", err)
		}
		if result.PRCreated {
			fmt.Printf("✅ PR created: %s\n", result.PRURL)
		} else {
			fmt.Printf("⚠️  Resume completed but PR not created: %s\n", result.Message)
		}
		return nil
	}

	// Load pre-generated triage plan if provided.
	planFile, _ := cmd.Flags().GetString("plan-file")
	if planFile != "" {
		preloaded, err := loadTriagePlan(planFile)
		if err != nil {
			fmt.Printf("⚠️  Could not load plan file: %v (will generate plan from scratch)\n", err)
		} else {
			a.PreloadedPlan = preloaded
			fmt.Printf("📋 Using pre-generated triage plan (%d candidate files)\n", len(preloaded.RelevantFiles))
		}
	}

	result, err := a.Work(ctx, t)
	if err != nil {
		return fmt.Errorf("work failed: %w", err)
	}

	if result.PRCreated {
		fmt.Printf("✅ PR created: %s\n", result.PRURL)
	} else {
		fmt.Printf("⚠️  Work completed but PR not created: %s\n", result.Message)
	}

	return nil
}

// parseTaskInput determines the input mode and creates the appropriate Task.
func parseTaskInput(cmd *cobra.Command, args []string, cfg *config.Config) (task.Task, error) {
	input := args[0]

	// Get mode flags
	isPrompt, _ := cmd.Flags().GetBool("prompt")
	isFile, _ := cmd.Flags().GetBool("file")
	overrideTitle, _ := cmd.Flags().GetString("title")
	overrideBranch, _ := cmd.Flags().GetString("branch-name")

	// Validate: only one mode can be set
	modesSet := 0
	if isPrompt {
		modesSet++
	}
	if isFile {
		modesSet++
	}
	if modesSet > 1 {
		return nil, fmt.Errorf("only one of --prompt or --file can be specified")
	}

	// Validate: title/branch overrides only work with prompt/file mode
	if !isPrompt && !isFile {
		if overrideTitle != "" {
			return nil, fmt.Errorf("--title can only be used with --prompt or --file")
		}
		if overrideBranch != "" {
			return nil, fmt.Errorf("--branch-name can only be used with --prompt or --file")
		}
	}

	// Create task based on mode
	if isPrompt {
		fmt.Println("📝 Prompt mode")
		return task.CreateFromPrompt(input, overrideTitle, overrideBranch)
	}

	if isFile {
		fmt.Println("📄 File mode")
		// Validate file exists
		if _, err := os.Stat(input); err != nil {
			return nil, fmt.Errorf("task file does not exist: %s", input)
		}
		return task.CreateFromFile(input, overrideTitle, overrideBranch)
	}

	// Default: Linear mode
	fmt.Println("🎫 Linear mode")
	linearClient := linear.New(cfg.LinearKey)
	return task.CreateFromLinear(ctx, linearClient, input)
}

// loadTriagePlan reads a triage TicketPlan JSON file and converts it to a planner.Plan.
func loadTriagePlan(path string) (*planner.Plan, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read plan file: %w", err)
	}

	var tp plan.TicketPlan
	if err := json.Unmarshal(data, &tp); err != nil {
		return nil, fmt.Errorf("parse plan JSON: %w", err)
	}

	// Convert triage TicketPlan → planner.Plan
	p := &planner.Plan{
		Summary:       tp.Approach,
		Approach:      []string{tp.Approach},
		RelevantFiles: tp.CandidateFiles,
		TestStrategy:  strings.Join(tp.Validation, "\n"),
	}

	// Merge stop conditions and uncertainties into warnings.
	for _, sc := range tp.StopConditions {
		p.Warnings = append(p.Warnings, "STOP: "+sc)
	}
	for _, u := range tp.Uncertainties {
		p.Warnings = append(p.Warnings, "UNCERTAIN: "+u)
	}
	if tp.Rollback != "" {
		p.Warnings = append(p.Warnings, "ROLLBACK: "+tp.Rollback)
	}

	return p, nil
}

// ctx is needed for CreateFromLinear
var ctx = context.Background()
