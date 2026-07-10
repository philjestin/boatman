package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	runtimeworkflows "github.com/philjestin/boatman-ecosystem/shared/agentruntime/workflows"
	"github.com/spf13/cobra"
)

var workflowsCmd = &cobra.Command{
	Use:   "workflows",
	Short: "Inspect built-in workflow templates",
	Long: `Inspect the provider-neutral workflow templates Boatman uses to model
agent-plane stages, gates, previews, skips, and validation loops.`,
	RunE: runWorkflowsList,
}

var workflowsShowCmd = &cobra.Command{
	Use:   "show [template-id]",
	Short: "Show a workflow template",
	Args:  cobra.ExactArgs(1),
	RunE:  runWorkflowsShow,
}

func init() {
	rootCmd.AddCommand(workflowsCmd)
	workflowsCmd.AddCommand(workflowsShowCmd)
	workflowsCmd.Flags().Bool("json", false, "Print workflow templates as JSON")
	workflowsShowCmd.Flags().Bool("json", false, "Print workflow template as JSON")
}

func runWorkflowsList(cmd *cobra.Command, args []string) error {
	templates := runtimeworkflows.DefaultLibrary().List()
	for _, template := range templates {
		if err := runtimeworkflows.Validate(template); err != nil {
			return err
		}
	}
	jsonOut, _ := cmd.Flags().GetBool("json")
	return writeWorkflowList(cmd.OutOrStdout(), templates, jsonOut)
}

func runWorkflowsShow(cmd *cobra.Command, args []string) error {
	template, ok := runtimeworkflows.DefaultLibrary().Get(args[0])
	if !ok {
		return fmt.Errorf("unknown workflow template %q", args[0])
	}
	if err := runtimeworkflows.Validate(template); err != nil {
		return err
	}
	jsonOut, _ := cmd.Flags().GetBool("json")
	return writeWorkflowShow(cmd.OutOrStdout(), template, jsonOut)
}

func writeWorkflowList(out io.Writer, templates []runtimeworkflows.Template, jsonOut bool) error {
	if jsonOut {
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "  ")
		return encoder.Encode(templates)
	}
	if len(templates) == 0 {
		fmt.Fprintln(out, "No workflow templates")
		return nil
	}
	fmt.Fprintf(out, "%-16s %-18s %-6s %-8s %s\n", "ID", "NAME", "STAGES", "GATES", "DESCRIPTION")
	for _, template := range templates {
		fmt.Fprintf(out, "%-16s %-18s %-6d %-8s %s\n",
			truncate(template.ID, 16),
			truncate(template.Name, 18),
			len(template.Stages),
			workflowGateSummary(template),
			truncate(template.Description, 80),
		)
	}
	return nil
}

func writeWorkflowShow(out io.Writer, template runtimeworkflows.Template, jsonOut bool) error {
	if jsonOut {
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "  ")
		return encoder.Encode(template)
	}
	fmt.Fprintf(out, "Workflow:    %s\n", template.ID)
	fmt.Fprintf(out, "Name:        %s\n", template.Name)
	fmt.Fprintf(out, "Description: %s\n\n", template.Description)
	fmt.Fprintf(out, "%-16s %-16s %-8s %-7s %-22s %s\n", "STAGE", "KIND", "GATE", "PREVIEW", "NEXT", "ON FAILURE")
	for _, stage := range template.Stages {
		fmt.Fprintf(out, "%-16s %-16s %-8s %-7s %-22s %s\n",
			truncate(stage.ID, 16),
			truncate(string(stage.Kind), 16),
			stage.Gate,
			boolText(stage.Preview),
			truncate(strings.Join(stage.Next, ","), 22),
			strings.Join(stage.OnFailure, ","),
		)
	}
	return nil
}

func workflowGateSummary(template runtimeworkflows.Template) string {
	var human, dynamic int
	for _, stage := range template.Stages {
		switch stage.Gate {
		case runtimeworkflows.GateHuman:
			human++
		case runtimeworkflows.GateDynamic:
			dynamic++
		}
	}
	if human == 0 && dynamic == 0 {
		return "none"
	}
	var parts []string
	if human > 0 {
		parts = append(parts, fmt.Sprintf("human:%d", human))
	}
	if dynamic > 0 {
		parts = append(parts, fmt.Sprintf("dynamic:%d", dynamic))
	}
	return strings.Join(parts, ",")
}
