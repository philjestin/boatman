// Package scottbott provides the ScottBott peer-review integration.
// It invokes the existing peer-review Claude skill in the target repository.
package scottbott

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/philjestin/boatman-ecosystem/harness/review"
	agentruntime "github.com/philjestin/boatman-ecosystem/shared/agentruntime"
	"github.com/philjestin/boatmanmode/internal/config"
	"github.com/philjestin/boatmanmode/internal/cost"
	"github.com/philjestin/boatmanmode/internal/events"
	runtimeproviders "github.com/philjestin/boatmanmode/internal/providers"
)

// ReviewResult represents the outcome of a code review.
type ReviewResult struct {
	Passed   bool     `json:"passed"`
	Score    int      `json:"score"`
	Summary  string   `json:"summary"`
	Issues   []Issue  `json:"issues"`
	Praise   []string `json:"praise"`
	Guidance string   `json:"guidance"`
}

// Issue represents a specific problem found during review.
type Issue struct {
	Severity    string `json:"severity"`
	File        string `json:"file"`
	Line        int    `json:"line"`
	Description string `json:"description"`
	Suggestion  string `json:"suggestion"`
}

// IssueToReviewIssue converts a scottbott Issue to a review.Issue.
func IssueToReviewIssue(issue Issue) review.Issue {
	return review.Issue{
		Severity:    issue.Severity,
		File:        issue.File,
		Line:        issue.Line,
		Description: issue.Description,
		Suggestion:  issue.Suggestion,
	}
}

// ReviewIssueToIssue converts a review.Issue to a scottbott Issue.
func ReviewIssueToIssue(issue review.Issue) Issue {
	return Issue{
		Severity:    issue.Severity,
		File:        issue.File,
		Line:        issue.Line,
		Description: issue.Description,
		Suggestion:  issue.Suggestion,
	}
}

// ScottBott invokes the review skill.
type ScottBott struct {
	workDir             string
	sessionName         string
	outputDir           string
	skill               string
	model               string
	effort              string
	enablePromptCaching bool
	provider            agentruntime.Provider
	runText             func(context.Context, agentruntime.Provider, agentruntime.RunRequest, func(agentruntime.Event)) (string, *cost.Usage, error)
	cfg                 *config.Config
}

// New creates a new ScottBott instance.
func New(cfg *config.Config) *ScottBott {
	return newScottBottWithProvider("", "reviewer", cfg.ReviewSkill, cfg, defaultProvider(cfg))
}

// NewForIteration creates a ScottBott for a specific review iteration.
func NewForIteration(iteration int, cfg *config.Config) *ScottBott {
	return newScottBottWithProvider("", fmt.Sprintf("reviewer-%d", iteration), cfg.ReviewSkill, cfg, defaultProvider(cfg))
}

// NewWithWorkDir creates a ScottBott that runs in a specific directory.
func NewWithWorkDir(workDir string, iteration int, cfg *config.Config) *ScottBott {
	return newScottBottWithProvider(workDir, fmt.Sprintf("reviewer-%d", iteration), cfg.ReviewSkill, cfg, defaultProvider(cfg))
}

// NewWithSkill creates a ScottBott with a specific skill/agent for review.
func NewWithSkill(workDir string, iteration int, skill string, cfg *config.Config) *ScottBott {
	if skill == "" {
		skill = "peer-review"
	}
	return newScottBottWithProvider(workDir, fmt.Sprintf("reviewer-%d", iteration), skill, cfg, defaultProvider(cfg))
}

func defaultProvider(cfg *config.Config) agentruntime.Provider {
	return runtimeproviders.MustFromConfig(cfg, agentruntime.RoleReviewer, "reviewer")
}

func newScottBottWithProvider(workDir, sessionName, skill string, cfg *config.Config, provider agentruntime.Provider) *ScottBott {
	return &ScottBott{
		workDir:             workDir,
		sessionName:         sessionName,
		outputDir:           filepath.Join(os.TempDir(), "boatman-sessions"),
		skill:               skill,
		model:               cfg.Claude.Models.Reviewer,
		effort:              cfg.Claude.Effort,
		enablePromptCaching: cfg.Claude.EnablePromptCaching,
		provider:            provider,
		runText:             runtimeproviders.RunTextWithEvents,
		cfg:                 cfg,
	}
}

// Review performs a code review using the configured peer-review skill.
func (s *ScottBott) Review(ctx context.Context, ticketContext, diff string) (*ReviewResult, *cost.Usage, error) {
	os.MkdirAll(s.outputDir, 0755)

	// Write the review prompt to a file for debugging.
	promptFile := filepath.Join(s.outputDir, fmt.Sprintf("%s-prompt.txt", s.sessionName))
	prompt := formatReviewPrompt(ticketContext, diff)
	if err := os.WriteFile(promptFile, []byte(prompt), 0644); err != nil {
		return nil, nil, fmt.Errorf("failed to write prompt: %w", err)
	}
	defer os.Remove(promptFile)

	// Output file for capturing response
	outputFile := filepath.Join(s.outputDir, fmt.Sprintf("%s.out", s.sessionName))

	fmt.Printf("   📏 Review: %d chars context, %d chars diff\n", len(ticketContext), len(diff))
	fmt.Printf("   🔍 Invoking %s skill...\n", s.skill)

	start := time.Now()
	response, usage, err := s.runReviewProvider(ctx, "", prompt, true)
	elapsed := time.Since(start)

	if err != nil {
		// If skill doesn't exist, fall back to system prompt
		fmt.Printf("   ⚠️  %s skill not found, using fallback...\n", s.skill)
		return s.reviewWithFallback(ctx, ticketContext, diff)
	}

	fmt.Printf("   ⏱️  Review completed in %s\n", elapsed.Round(time.Second))

	// Save output for debugging
	os.WriteFile(outputFile, []byte(response), 0644)

	// Parse the response
	result, err := s.parseReviewResponse(strings.TrimSpace(response))
	return result, usage, err
}

// reviewWithFallback uses a system prompt if peer-review skill isn't available.
func (s *ScottBott) reviewWithFallback(ctx context.Context, ticketContext, diff string) (*ReviewResult, *cost.Usage, error) {
	systemPrompt := `You are a senior staff engineer conducting a peer code review.
Be thorough, constructive, and focused on correctness, security, and maintainability.

Respond with ONLY a JSON object:
{
  "passed": boolean,
  "score": number (0-100),
  "summary": "2-3 sentence summary",
  "issues": [{"severity": "critical|major|minor", "file": "path", "line": 0, "description": "what's wrong", "suggestion": "how to fix"}],
  "praise": ["good things"],
  "guidance": "guidance for fixing if failed"
}

Pass if: no critical issues, ≤2 major issues, code meets requirements.`

	prompt := formatReviewPrompt(ticketContext, diff)

	start := time.Now()
	response, usage, err := s.runReviewProvider(ctx, systemPrompt, prompt, false)
	elapsed := time.Since(start)

	if err != nil {
		return nil, nil, fmt.Errorf("review failed: %w", err)
	}

	fmt.Printf("   ⏱️  Review completed in %s\n", elapsed.Round(time.Second))

	result, err := s.parseReviewResponse(strings.TrimSpace(response))
	return result, usage, err
}

func (s *ScottBott) runReviewProvider(ctx context.Context, systemPrompt, prompt string, useSkill bool) (string, *cost.Usage, error) {
	metadata := map[string]string{
		"phaseId": s.sessionName,
	}
	profile := "reviewer"
	if s.enablePromptCaching {
		metadata["enablePromptCaching"] = "true"
	}
	if useSkill && s.skill != "" {
		metadata["claudeAgent"] = s.skill
		metadata["outputFormat"] = "text"
		profile = s.skill
	}

	return s.runText(ctx, s.provider, agentruntime.RunRequest{
		RunID:        s.sessionName,
		Role:         agentruntime.RoleReviewer,
		Profile:      profile,
		Provider:     s.provider.Name(),
		Model:        s.model,
		WorkDir:      s.workDir,
		Instructions: systemPrompt,
		Messages: []agentruntime.Message{
			{Role: "user", Content: prompt},
		},
		OutputSchema:   reviewOutputSchema(),
		ApprovalPolicy: agentruntime.ApprovalSuggest,
		Reasoning: &agentruntime.ReasoningOptions{
			Effort: s.effort,
		},
		Metadata: metadata,
	}, func(event agentruntime.Event) {
		if event.Type != agentruntime.EventProviderRaw {
			return
		}
		if len(event.Raw) > 0 {
			events.ProviderRaw(s.sessionName, event.Provider, string(event.Raw))
		} else if event.Message != "" {
			events.ProviderRaw(s.sessionName, event.Provider, event.Message)
		}
	})
}

func reviewOutputSchema() *agentruntime.OutputSchema {
	return &agentruntime.OutputSchema{
		Name:        "review_result",
		Description: "Peer review result for a Boatman work pipeline change.",
		Strict:      true,
		Schema: json.RawMessage(`{
  "type": "object",
  "additionalProperties": false,
  "required": ["passed", "score", "summary", "issues", "praise", "guidance"],
  "properties": {
    "passed": {"type": "boolean"},
    "score": {"type": "integer"},
    "summary": {"type": "string"},
    "issues": {
      "type": "array",
      "items": {
        "type": "object",
        "additionalProperties": false,
        "required": ["severity", "file", "line", "description", "suggestion"],
        "properties": {
          "severity": {"type": "string"},
          "file": {"type": "string"},
          "line": {"type": "integer"},
          "description": {"type": "string"},
          "suggestion": {"type": "string"}
        }
      }
    },
    "praise": {
      "type": "array",
      "items": {"type": "string"}
    },
    "guidance": {"type": "string"}
  }
}`),
	}
}

// formatReviewPrompt creates the prompt for code review.
func formatReviewPrompt(ticketContext, diff string) string {
	return fmt.Sprintf(`## Ticket Context
%s

## Code Changes
%s

Review these changes against the requirements. Provide your assessment.`, ticketContext, diff)
}

// parseReviewResponse extracts ReviewResult from Claude's response.
func (s *ScottBott) parseReviewResponse(response string) (*ReviewResult, error) {
	// First try JSON extraction
	jsonStr := extractJSON(response)

	var result ReviewResult
	if err := json.Unmarshal([]byte(jsonStr), &result); err == nil {
		return &result, nil
	}

	// If JSON parsing fails, parse natural language response
	return s.parseNaturalLanguageReview(response)
}

// parseNaturalLanguageReview extracts review info from a natural language response.
func (s *ScottBott) parseNaturalLanguageReview(response string) (*ReviewResult, error) {
	lower := strings.ToLower(response)

	result := &ReviewResult{
		Score:  70, // Default score
		Issues: []Issue{},
		Praise: []string{},
	}

	// Determine pass/fail
	// Look for explicit pass indicators
	if strings.Contains(lower, "lgtm") ||
		strings.Contains(lower, "looks good") ||
		strings.Contains(lower, "approved") ||
		strings.Contains(lower, "ready to merge") ||
		strings.Contains(lower, "no critical issues") && !strings.Contains(lower, "major issues") {
		result.Passed = true
		result.Score = 85
	}

	// Look for explicit fail indicators - only if strict parsing is enabled
	if s.cfg != nil && s.cfg.Review.StrictParsing {
		if strings.Contains(lower, "must be addressed") ||
			strings.Contains(lower, "blocking") ||
			strings.Contains(lower, "critical issue") ||
			strings.Contains(lower, "cannot be merged") ||
			strings.Contains(lower, "needs work") ||
			strings.Contains(lower, "issues that need to be addressed") {
			result.Passed = false
			result.Score = 50
		}
	} else {
		// In relaxed mode, only fail on very explicit blocking language
		if strings.Contains(lower, "cannot be merged") ||
			strings.Contains(lower, "blocking issue") {
			result.Passed = false
			result.Score = 50
		}
	}

	// Extract summary - first paragraph or first 300 chars
	lines := strings.Split(response, "\n")
	var summaryBuilder strings.Builder
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			if summaryBuilder.Len() > 0 {
				break
			}
			continue
		}
		summaryBuilder.WriteString(line)
		summaryBuilder.WriteString(" ")
		if summaryBuilder.Len() > 200 {
			break
		}
	}
	result.Summary = strings.TrimSpace(summaryBuilder.String())
	if len(result.Summary) > 300 {
		result.Summary = result.Summary[:300] + "..."
	}

	// Extract issues by looking for common patterns
	issuePatterns := []string{
		"issue", "problem", "bug", "error", "fix", "should", "must", "need to",
		"incorrect", "missing", "wrong", "critical", "major", "minor",
	}

	for _, line := range lines {
		lineLower := strings.ToLower(line)
		for _, pattern := range issuePatterns {
			if strings.Contains(lineLower, pattern) {
				// This line might describe an issue
				line = strings.TrimSpace(line)
				line = strings.TrimPrefix(line, "- ")
				line = strings.TrimPrefix(line, "* ")
				line = strings.TrimPrefix(line, "• ")

				if len(line) > 20 && len(line) < 500 && !strings.HasPrefix(line, "#") {
					severity := "minor"
					if strings.Contains(lineLower, "critical") {
						severity = "critical"
					} else if strings.Contains(lineLower, "major") || strings.Contains(lineLower, "must") {
						severity = "major"
					}

					result.Issues = append(result.Issues, Issue{
						Severity:    severity,
						Description: line,
					})
				}
				break
			}
		}
	}

	// Deduplicate and limit issues
	seen := make(map[string]bool)
	var uniqueIssues []Issue
	for _, issue := range result.Issues {
		key := strings.ToLower(issue.Description[:min(50, len(issue.Description))])
		if !seen[key] && len(uniqueIssues) < 10 {
			seen[key] = true
			uniqueIssues = append(uniqueIssues, issue)
		}
	}
	result.Issues = uniqueIssues

	// Count critical/major issues to determine pass/fail
	criticalCount := 0
	majorCount := 0
	for _, issue := range result.Issues {
		if issue.Severity == "critical" {
			criticalCount++
		} else if issue.Severity == "major" {
			majorCount++
		}
	}

	// Use configurable thresholds or defaults
	maxCritical := 1
	maxMajor := 3
	if s.cfg != nil {
		maxCritical = s.cfg.Review.MaxCriticalIssues
		maxMajor = s.cfg.Review.MaxMajorIssues
	}

	if criticalCount > maxCritical || majorCount > maxMajor {
		result.Passed = false
		result.Score = 40 + (10 - criticalCount*10 - majorCount*5)
		if result.Score < 20 {
			result.Score = 20
		}
	}

	// Extract guidance - look for "fix" or "recommendation" sections
	var guidanceBuilder strings.Builder
	inGuidance := false
	for _, line := range lines {
		lineLower := strings.ToLower(line)
		if strings.Contains(lineLower, "recommend") ||
			strings.Contains(lineLower, "suggestion") ||
			strings.Contains(lineLower, "to fix") ||
			strings.Contains(lineLower, "next step") {
			inGuidance = true
		}
		if inGuidance {
			guidanceBuilder.WriteString(line)
			guidanceBuilder.WriteString("\n")
			if guidanceBuilder.Len() > 500 {
				break
			}
		}
	}
	result.Guidance = strings.TrimSpace(guidanceBuilder.String())

	// If no guidance extracted, use the issues as guidance
	if result.Guidance == "" && len(result.Issues) > 0 {
		var issueGuidance strings.Builder
		issueGuidance.WriteString("Please address the following issues:\n")
		for i, issue := range result.Issues {
			if i >= 5 {
				break
			}
			issueGuidance.WriteString(fmt.Sprintf("%d. %s\n", i+1, issue.Description))
		}
		result.Guidance = issueGuidance.String()
	}

	return result, nil
}

// extractJSON extracts JSON from a response that might be wrapped in markdown.
func extractJSON(text string) string {
	text = strings.TrimSpace(text)

	if strings.HasPrefix(text, "```json") {
		text = strings.TrimPrefix(text, "```json")
		if idx := strings.LastIndex(text, "```"); idx != -1 {
			text = text[:idx]
		}
	} else if strings.HasPrefix(text, "```") {
		text = strings.TrimPrefix(text, "```")
		if idx := strings.LastIndex(text, "```"); idx != -1 {
			text = text[:idx]
		}
	}

	return strings.TrimSpace(text)
}

// FormatReview returns a human-readable format of the review.
func (r *ReviewResult) FormatReview() string {
	var sb strings.Builder

	sb.WriteString("   ┌─────────────────────────────────────────┐\n")
	if r.Passed {
		sb.WriteString("   │  ✅ REVIEW PASSED                       │\n")
	} else {
		sb.WriteString("   │  ❌ REVIEW FAILED                       │\n")
	}
	sb.WriteString("   └─────────────────────────────────────────┘\n")

	sb.WriteString(fmt.Sprintf("   📊 Score: %d/100\n\n", r.Score))
	sb.WriteString(fmt.Sprintf("   📝 Summary:\n      %s\n\n", r.Summary))

	if len(r.Praise) > 0 {
		sb.WriteString("   👍 What's good:\n")
		for _, p := range r.Praise {
			sb.WriteString(fmt.Sprintf("      • %s\n", p))
		}
		sb.WriteString("\n")
	}

	if len(r.Issues) > 0 {
		sb.WriteString("   🔍 Issues found:\n")
		for i, issue := range r.Issues {
			icon := "💡"
			switch issue.Severity {
			case "critical":
				icon = "🚨"
			case "major":
				icon = "⚠️"
			case "minor":
				icon = "📝"
			}

			sb.WriteString(fmt.Sprintf("      %d. %s [%s] %s\n", i+1, icon, strings.ToUpper(issue.Severity), issue.Description))
			if issue.File != "" {
				location := issue.File
				if issue.Line > 0 {
					location = fmt.Sprintf("%s:%d", issue.File, issue.Line)
				}
				sb.WriteString(fmt.Sprintf("         📍 %s\n", location))
			}
			if issue.Suggestion != "" {
				sb.WriteString(fmt.Sprintf("         💡 %s\n", issue.Suggestion))
			}
			sb.WriteString("\n")
		}
	}

	if r.Guidance != "" && !r.Passed {
		sb.WriteString("   📋 Guidance:\n")
		for _, line := range strings.Split(r.Guidance, "\n") {
			sb.WriteString(fmt.Sprintf("      %s\n", line))
		}
	}

	return sb.String()
}

// GetIssueDescriptions returns a list of issue descriptions for handoff.
func (r *ReviewResult) GetIssueDescriptions() []string {
	issues := make([]string, len(r.Issues))
	for i, issue := range r.Issues {
		issues[i] = fmt.Sprintf("[%s] %s", issue.Severity, issue.Description)
		if issue.Suggestion != "" {
			issues[i] += " → " + issue.Suggestion
		}
	}
	return issues
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
