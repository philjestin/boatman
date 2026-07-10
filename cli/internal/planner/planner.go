// Package planner runs a runtime provider to analyze tickets and plan execution.
// This gives the executor agent focused context and a clear approach.
package planner

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	agentruntime "github.com/philjestin/boatman-ecosystem/shared/agentruntime"
	"github.com/philjestin/boatman-ecosystem/shared/agentruntime/toolbroker"
	"github.com/philjestin/boatmanmode/internal/config"
	"github.com/philjestin/boatmanmode/internal/cost"
	"github.com/philjestin/boatmanmode/internal/events"
	"github.com/philjestin/boatmanmode/internal/linear"
	runtimeproviders "github.com/philjestin/boatmanmode/internal/providers"
	"github.com/philjestin/boatmanmode/internal/task"
)

// Plan contains the structured output from the planning agent.
type Plan struct {
	// Summary is a brief description of what needs to be done
	Summary string `json:"summary"`

	// Approach is the step-by-step plan
	Approach []string `json:"approach"`

	// RelevantFiles are files Claude identified as important
	RelevantFiles []string `json:"relevant_files"`

	// RelevantDirs are directories to focus on
	RelevantDirs []string `json:"relevant_dirs"`

	// ExistingPatterns are patterns to follow
	ExistingPatterns []string `json:"existing_patterns"`

	// TestStrategy is how to test the implementation
	TestStrategy string `json:"test_strategy"`

	// Warnings are things to watch out for
	Warnings []string `json:"warnings"`
}

// Planner analyzes tickets before execution.
type Planner struct {
	provider     agentruntime.Provider
	worktreePath string
	model        string
	effort       string
	enableTools  bool
	runText      func(context.Context, agentruntime.Provider, agentruntime.RunRequest, func(agentruntime.Event)) (string, *cost.Usage, error)
}

// New creates a new Planner agent.
func New(worktreePath string, cfg *config.Config) *Planner {
	provider := runtimeproviders.MustFromConfig(cfg, agentruntime.RolePlanner, "planner")
	return newPlannerWithProvider(worktreePath, cfg, provider)
}

func newPlannerWithProvider(worktreePath string, cfg *config.Config, provider agentruntime.Provider) *Planner {
	return &Planner{
		provider:     provider,
		worktreePath: worktreePath,
		model:        cfg.Claude.Models.Planner,
		effort:       cfg.Claude.Effort,
		enableTools:  cfg.EnableTools,
		runText:      runtimeproviders.RunTextWithEvents,
	}
}

// Analyze runs the planning agent to understand the task.
func (p *Planner) Analyze(ctx context.Context, t task.Task) (*Plan, *cost.Usage, error) {
	fmt.Println("   🧠 Running planning agent...")

	systemPrompt := `You are a senior software architect planning a development task.
Your job is to analyze the task and codebase to create a focused execution plan.

IMPORTANT: Use your tools to explore the codebase. Do NOT guess - actually look at the code.

Your process:
1. Read the task requirements carefully
2. Search for existing similar implementations (use Glob, Grep)
3. Read key files to understand patterns
4. Identify files that need to be created or modified
5. Note any patterns or conventions to follow

After exploration, output a JSON plan in this exact format:

` + "```json" + `
{
  "summary": "One sentence describing the task",
  "approach": [
    "Step 1: Do X",
    "Step 2: Do Y"
  ],
  "relevant_files": [
    "path/to/file1.rb",
    "path/to/file2.rb"
  ],
  "relevant_dirs": [
    "packs/some_pack/app/graphql/"
  ],
  "existing_patterns": [
    "Pattern: Use XyzResolver for queries",
    "Pattern: Commands go in app/commands/"
  ],
  "test_strategy": "How to test this implementation",
  "warnings": [
    "Don't modify X",
    "Watch out for Y"
  ]
}
` + "```" + `

Output ONLY the JSON block after your exploration. No other text after the JSON.`

	prompt := fmt.Sprintf(`# Task: %s

## Description
%s

Analyze this task and explore the codebase to create an execution plan.
Focus on understanding existing patterns before proposing new code.`,
		t.GetTitle(),
		t.GetDescription())

	fmt.Println("   📝 Analyzing task and exploring codebase...")

	start := time.Now()
	response, usage, err := p.runText(ctx, p.provider, agentruntime.RunRequest{
		RunID:        "planner-" + t.GetID(),
		Role:         agentruntime.RolePlanner,
		Profile:      "planner",
		Provider:     p.provider.Name(),
		Model:        p.model,
		WorkDir:      p.worktreePath,
		Instructions: systemPrompt,
		Messages: []agentruntime.Message{
			{Role: "user", Content: prompt},
		},
		Tools:          plannerTools(p.enableTools),
		OutputSchema:   plannerOutputSchema(),
		ApprovalPolicy: agentruntime.ApprovalFullAuto,
		Reasoning: &agentruntime.ReasoningOptions{
			Effort: p.effort,
		},
		Metadata: map[string]string{
			"phaseId": "planner",
		},
	}, func(event agentruntime.Event) {
		if event.Type == agentruntime.EventProviderRaw {
			if len(event.Raw) > 0 {
				events.ProviderRaw("planner", event.Provider, string(event.Raw))
			} else if event.Message != "" {
				events.ProviderRaw("planner", event.Provider, event.Message)
			}
		}
	})
	elapsed := time.Since(start)

	if err != nil {
		return nil, nil, fmt.Errorf("planning agent failed: %w", err)
	}

	fmt.Printf("   ⏱️  Planning completed in %s\n", elapsed.Round(time.Second))

	// Parse the JSON plan from response
	plan, err := p.parsePlan(response)
	if err != nil {
		fmt.Printf("   ⚠️  Could not parse plan JSON: %v\n", err)
		// Return a basic plan from the response
		return &Plan{
			Summary:  "Planning agent explored codebase",
			Approach: []string{"See planning agent output for details"},
		}, usage, nil
	}

	// Display plan summary
	fmt.Printf("   📋 Plan: %s\n", plan.Summary)
	fmt.Printf("   📁 Found %d relevant files\n", len(plan.RelevantFiles))
	if len(plan.Approach) > 0 {
		fmt.Printf("   📝 Approach: %d steps\n", len(plan.Approach))
	}

	return plan, usage, nil
}

func plannerTools(enabled bool) []agentruntime.ToolRef {
	if !enabled {
		return nil
	}
	return toolbroker.PlannerRefs()
}

func plannerOutputSchema() *agentruntime.OutputSchema {
	return &agentruntime.OutputSchema{
		Name:        "work_planner_plan",
		Description: "Focused implementation plan for the work pipeline executor.",
		Strict:      true,
		Schema: json.RawMessage(`{
  "type": "object",
  "additionalProperties": false,
  "required": [
    "summary",
    "approach",
    "relevant_files",
    "relevant_dirs",
    "existing_patterns",
    "test_strategy",
    "warnings"
  ],
  "properties": {
    "summary": {"type": "string"},
    "approach": {
      "type": "array",
      "items": {"type": "string"}
    },
    "relevant_files": {
      "type": "array",
      "items": {"type": "string"}
    },
    "relevant_dirs": {
      "type": "array",
      "items": {"type": "string"}
    },
    "existing_patterns": {
      "type": "array",
      "items": {"type": "string"}
    },
    "test_strategy": {"type": "string"},
    "warnings": {
      "type": "array",
      "items": {"type": "string"}
    }
  }
}`),
	}
}

// AnalyzeTicket is a backward-compatibility wrapper for Linear tickets.
func (p *Planner) AnalyzeTicket(ctx context.Context, ticket *linear.Ticket) (*Plan, *cost.Usage, error) {
	return p.Analyze(ctx, task.NewLinearTask(ticket))
}

// parsePlan extracts the JSON plan from Claude's response.
func (p *Planner) parsePlan(response string) (*Plan, error) {
	// Find JSON block in response
	jsonRe := regexp.MustCompile("```(?:json)?\\s*\\n?([\\s\\S]*?)\\n?```")
	matches := jsonRe.FindStringSubmatch(response)

	var jsonStr string
	if len(matches) > 1 {
		jsonStr = matches[1]
	} else {
		// Try to find raw JSON
		start := strings.Index(response, "{")
		end := strings.LastIndex(response, "}")
		if start >= 0 && end > start {
			jsonStr = response[start : end+1]
		} else {
			return nil, fmt.Errorf("no JSON found in response")
		}
	}

	var plan Plan
	if err := json.Unmarshal([]byte(jsonStr), &plan); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	return &plan, nil
}

// ToHandoff creates a concise handoff for the executor agent.
func (plan *Plan) ToHandoff() string {
	var sb strings.Builder

	sb.WriteString("# Execution Plan\n\n")

	if plan.Summary != "" {
		sb.WriteString(fmt.Sprintf("## Summary\n%s\n\n", plan.Summary))
	}

	if len(plan.Approach) > 0 {
		sb.WriteString("## Approach\n")
		for i, step := range plan.Approach {
			sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, step))
		}
		sb.WriteString("\n")
	}

	if len(plan.RelevantFiles) > 0 {
		sb.WriteString("## Key Files (read these first)\n")
		for _, f := range plan.RelevantFiles {
			sb.WriteString(fmt.Sprintf("- `%s`\n", f))
		}
		sb.WriteString("\n")
	}

	if len(plan.ExistingPatterns) > 0 {
		sb.WriteString("## Patterns to Follow\n")
		for _, p := range plan.ExistingPatterns {
			sb.WriteString(fmt.Sprintf("- %s\n", p))
		}
		sb.WriteString("\n")
	}

	if plan.TestStrategy != "" {
		sb.WriteString(fmt.Sprintf("## Testing\n%s\n\n", plan.TestStrategy))
	}

	if len(plan.Warnings) > 0 {
		sb.WriteString("## ⚠️ Warnings\n")
		for _, w := range plan.Warnings {
			sb.WriteString(fmt.Sprintf("- %s\n", w))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}
