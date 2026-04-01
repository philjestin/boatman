package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/philjestin/boatmanmode/internal/config"
	"github.com/philjestin/boatmanmode/internal/linear"
	"github.com/philjestin/boatmanmode/internal/plan"
	"github.com/philjestin/boatmanmode/internal/triage"
	"github.com/spf13/cobra"
)

var triageCmd = &cobra.Command{
	Use:   "triage",
	Short: "Triage backlog tickets for AI execution suitability",
	Long: `Triage Linear tickets using the ADR-004 rubric pipeline.

Fetches tickets from Linear, scores them on 7 rubric dimensions using Claude,
classifies them via a deterministic decision tree, and clusters related tickets.

Examples:
  boatman triage --teams ENG,FE --limit 50
  boatman triage --ticket-ids ENG-123,ENG-456
  boatman triage --teams ENG --post-comments --states backlog,unstarted
  boatman triage --ticket-ids ENG-123 --dry-run`,
	RunE: runTriage,
}

func init() {
	rootCmd.AddCommand(triageCmd)

	triageCmd.Flags().StringSlice("teams", nil, "Team keys to fetch (e.g., ENG,FE)")
	triageCmd.Flags().StringSlice("states", []string{"backlog", "triage", "unstarted"}, "Workflow state types to include")
	triageCmd.Flags().Int("limit", 50, "Maximum tickets to fetch")
	triageCmd.Flags().StringSlice("ticket-ids", nil, "Specific ticket identifiers (e.g., ENG-123,ENG-456)")
	triageCmd.Flags().Bool("post-comments", false, "Post triage results as Linear comments")
	triageCmd.Flags().Bool("dry-run", false, "Run without writing logs or posting comments")
	triageCmd.Flags().String("output-dir", "", "Output directory for decision logs (default: .boatman-triage)")
	triageCmd.Flags().Int("concurrency", 0, "Max concurrent Claude scoring calls (default: from config)")
	triageCmd.Flags().Bool("emit-events", false, "Emit JSON events to stdout for desktop app integration")
	triageCmd.Flags().Bool("generate-plans", false, "Generate execution plans for AI_DEFINITE and AI_LIKELY tickets")
	triageCmd.Flags().String("repo-path", "", "Path to repo for plan generation (default: current directory)")
}

func runTriage(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	ticketIDs, _ := cmd.Flags().GetStringSlice("ticket-ids")
	teams, _ := cmd.Flags().GetStringSlice("teams")
	states, _ := cmd.Flags().GetStringSlice("states")
	limit, _ := cmd.Flags().GetInt("limit")
	postComments, _ := cmd.Flags().GetBool("post-comments")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	outputDir, _ := cmd.Flags().GetString("output-dir")
	concurrency, _ := cmd.Flags().GetInt("concurrency")
	emitEvents, _ := cmd.Flags().GetBool("emit-events")
	generatePlans, _ := cmd.Flags().GetBool("generate-plans")
	repoPath, _ := cmd.Flags().GetString("repo-path")

	// Require either --ticket-ids or --teams.
	if len(ticketIDs) == 0 && len(teams) == 0 {
		if len(cfg.Triage.DefaultTeams) > 0 {
			teams = cfg.Triage.DefaultTeams
		} else {
			return fmt.Errorf("either --ticket-ids or --teams is required (or set triage.default_teams in config)")
		}
	}

	linearClient := linear.New(cfg.LinearKey)
	pipeline := triage.NewPipeline(cfg, linearClient)

	opts := triage.PipelineOptions{
		TicketIDs:    ticketIDs,
		TeamKeys:     teams,
		States:       states,
		Limit:        limit,
		PostComments: postComments,
		DryRun:       dryRun,
		OutputDir:    outputDir,
		Concurrency:  concurrency,
		EmitEvents:    emitEvents,
		GeneratePlans: generatePlans,
		RepoPath:      repoPath,
	}

	if dryRun {
		fmt.Println("Dry run mode - no logs or comments will be written")
	}

	result, err := pipeline.Run(ctx, opts)
	if err != nil {
		return fmt.Errorf("triage pipeline failed: %w", err)
	}

	// --- Stage 4: Plan Generation (optional) ---
	if generatePlans && (result.Stats.AIDefiniteCount+result.Stats.AILikelyCount) > 0 {
		repoDir := repoPath
		if repoDir == "" {
			repoDir = "."
		}

		// Filter to AI_DEFINITE and AI_LIKELY tickets.
		var planTickets []triage.NormalizedTicket
		var planClassifications []triage.Classification
		for _, c := range result.Classifications {
			if c.Category != triage.CategoryAIDefinite && c.Category != triage.CategoryAILikely {
				continue
			}
			planClassifications = append(planClassifications, c)
			for _, t := range result.Tickets {
				if t.TicketID == c.TicketID {
					planTickets = append(planTickets, t)
					break
				}
			}
		}

		generator := plan.NewGenerator(cfg, repoDir)
		planResults := generator.GenerateBatch(ctx, planTickets, planClassifications, result.ContextDocs, opts.Concurrency)

		// Validate each plan.
		docMap := make(map[string]*triage.ContextDoc)
		for i := range result.ContextDocs {
			for _, tid := range result.ContextDocs[i].TicketIDs {
				docMap[tid] = &result.ContextDocs[i]
			}
		}
		for i := range planResults {
			if planResults[i].Plan != nil {
				validation := plan.ValidatePlan(planResults[i].Plan, repoDir, docMap[planResults[i].TicketID])
				planResults[i].Validation = validation
				plan.EmitTicketValidated(planResults[i].TicketID, validation)
			}
		}

		// Compute plan stats.
		planStats := plan.PlanStats{Total: len(planResults)}
		for _, pr := range planResults {
			if pr.Error != "" {
				planStats.Failed++
			} else if pr.Validation != nil && pr.Validation.Passed {
				planStats.Passed++
			} else {
				planStats.Failed++
			}
			if pr.Usage != nil {
				planStats.TotalTokensUsed += pr.Usage.InputTokens + pr.Usage.OutputTokens
				planStats.TotalCostUSD += pr.Usage.TotalCostUSD
			}
		}

		// Emit plan_complete event with results and stats for desktop app.
		plan.EmitPlanComplete(planResults, planStats)

		// Store on result as json.RawMessage.
		result.Plans, _ = json.Marshal(planResults)
		result.PlanStats, _ = json.Marshal(planStats)
	}

	if !emitEvents {
		printTriageResult(result)
	}
	return nil
}

// printTriageResult displays a summary table of triage results.
func printTriageResult(result *triage.TriageResult) {
	if result.Stats.TotalTickets == 0 {
		fmt.Println("No tickets found matching the given filters.")
		return
	}

	// Header
	fmt.Println()
	fmt.Printf("%-12s %-50s %-24s %s\n", "TICKET", "TITLE", "CATEGORY", "CLUSTER")
	fmt.Println(strings.Repeat("-", 100))

	// Build cluster lookup.
	ticketCluster := make(map[string]string)
	for _, cl := range result.Clusters {
		for _, tid := range cl.TicketIDs {
			ticketCluster[tid] = cl.ClusterID
		}
	}

	for _, c := range result.Classifications {
		title := ticketTitle(result.Tickets, c.TicketID)
		if len(title) > 48 {
			title = title[:45] + "..."
		}
		cluster := ticketCluster[c.TicketID]
		if cluster == "" {
			cluster = "-"
		}
		fmt.Printf("%-12s %-50s %-24s %s\n", c.TicketID, title, c.Category, cluster)
	}

	// Summary
	fmt.Println(strings.Repeat("-", 100))
	fmt.Printf("\nSummary: %d tickets triaged\n", result.Stats.TotalTickets)
	fmt.Printf("  AI_DEFINITE:          %d\n", result.Stats.AIDefiniteCount)
	fmt.Printf("  AI_LIKELY:            %d\n", result.Stats.AILikelyCount)
	fmt.Printf("  HUMAN_REVIEW_REQUIRED: %d\n", result.Stats.HumanReviewCount)
	fmt.Printf("  HUMAN_ONLY:           %d\n", result.Stats.HumanOnlyCount)
	fmt.Printf("  Clusters:             %d\n", result.Stats.ClusterCount)

	if result.Stats.TotalTokensUsed > 0 {
		fmt.Printf("  Tokens used:          %d\n", result.Stats.TotalTokensUsed)
		fmt.Printf("  Cost:                 $%.4f\n", result.Stats.TotalCostUSD)
	}
	fmt.Println()
}

// ticketTitle finds the title for a ticket ID from the normalized tickets list.
func ticketTitle(tickets []triage.NormalizedTicket, ticketID string) string {
	for _, t := range tickets {
		if t.TicketID == ticketID {
			return t.Title
		}
	}
	return ""
}
