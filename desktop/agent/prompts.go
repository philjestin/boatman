package agent

import (
	"fmt"
	"strings"
)

// firefighterMCPTools maps MCP server names (lowercase) to their prompt descriptions
var firefighterMCPTools = map[string]struct {
	toolLine     string
	workflowStep string
}{
	"bugsnag": {
		toolLine:     "- Bugsnag MCP server: Use the available Bugsnag MCP tools to query errors, projects, and events",
		workflowStep: "- Error Discovery: Use Bugsnag MCP tools to query recent errors (NOT bash commands)",
	},
	"datadog": {
		toolLine:     "- Datadog MCP server: Use the available Datadog MCP tools to query logs, metrics, and monitors",
		workflowStep: "- Context Gathering: Use Datadog MCP tools to check logs and metrics around error timestamps",
	},
	"slack": {
		toolLine:     "- Slack MCP server: Use the available Slack MCP tools to read messages from alert channels and reply in threads",
		workflowStep: "- Slack Monitoring: Use Slack MCP tools to read alert channels for Datadog notifications",
	},
	"linear": {
		toolLine:     "- Linear MCP server: Use the available Linear MCP tools to query and update tickets",
		workflowStep: "",
	},
}

// hasMCPServer checks if a server name contains the given keyword (case-insensitive)
func hasMCPServer(serverNames []string, keyword string) bool {
	keyword = strings.ToLower(keyword)
	for _, name := range serverNames {
		if strings.Contains(strings.ToLower(name), keyword) {
			return true
		}
	}
	return false
}

// buildMCPToolsSection generates the "Available MCP Tools" and workflow steps
// based on which MCP servers are actually configured.
func buildMCPToolsSection(serverNames []string) (string, string) {
	// Order matters for display
	keys := []string{"bugsnag", "datadog", "slack", "linear"}

	var toolLines []string
	var workflowSteps []string

	for _, key := range keys {
		if hasMCPServer(serverNames, key) {
			info := firefighterMCPTools[key]
			toolLines = append(toolLines, info.toolLine)
			if info.workflowStep != "" {
				workflowSteps = append(workflowSteps, info.workflowStep)
			}
		}
	}

	if len(toolLines) == 0 {
		return "No MCP tools are currently configured. Use bash commands or available tools to investigate.", ""
	}

	return strings.Join(toolLines, "\n"), strings.Join(workflowSteps, "\n")
}

const firefighterSystemTemplate = `You are a Firefighter Agent specialized in investigating and fixing production incidents.

IMPORTANT: You have access to monitoring tools through MCP (Model Context Protocol). These tools are ALREADY AVAILABLE to you - do NOT try to use bash commands for services that have MCP tools configured.

Your mission:
1. Monitor production systems for errors and anomalies using MCP tools
2. Investigate issues proactively as they occur
3. Attempt automatic fixes in isolated git worktrees
4. Generate comprehensive incident reports

Available MCP Tools:
%TOOLS%

Investigation Workflow:
%WORKFLOW_STEPS%
- Code Analysis: Use git log/blame to find recent changes and owners
- Root Cause Hypothesis: Correlate timeline of errors with deployments
- Fix Attempt: For high-severity issues, attempt automatic fix
- Report Generation: Provide structured incident report

Sub-Agent Spawning:
- For each alert requiring investigation, use the Agent tool to spawn a sub-agent
- Sub-agents investigate and attempt fixes in isolated git worktrees
- Use isolation: "worktree" when spawning sub-agents for fix attempts
- Sub-agents should: analyze root cause, implement fix, run tests, create draft PR

Auto-Fix Workflow (for HIGH severity issues):
1. Create isolated worktree: git worktree add ../worktrees/fix-[issue-id] -b fix/[descriptive-name]
2. Navigate to worktree: cd ../worktrees/fix-[issue-id]
3. Implement the fix
4. Run tests to verify: npm test / bundle exec rspec / etc.
5. If tests pass: Create draft PR with detailed description
6. If tests fail: Document the attempt and findings
7. Clean up: git worktree remove ../worktrees/fix-[issue-id]

Report Format:
### Incident Summary
- Error: [description]
- First Seen: [timestamp]
- Frequency: [count/rate]
- Severity: [level]

### Affected Systems
- Services: [list]
- Code Owners: [teams/individuals from git blame]

### Timeline
[Chronological events]

### Error Details
[Stacktraces, messages, context]

### Code Analysis
[Recent changes, relevant commits]

### Root Cause Analysis
[Primary hypothesis with evidence]

### Recommended Actions
1. [Immediate action]
2. [Short-term fix]
3. [Long-term prevention]

### References
- Monitoring links
- Git commits: [SHAs]

%SCOPE%`

// GetFirefighterPrompt returns the firefighter prompt with optional scope,
// dynamically including only the MCP tools that are actually configured.
func GetFirefighterPrompt(scope string, mcpServerNames ...string) string {
	toolsSection, workflowSteps := buildMCPToolsSection(mcpServerNames)

	prompt := strings.Replace(firefighterSystemTemplate, "%TOOLS%", toolsSection, 1)
	prompt = strings.Replace(prompt, "%WORKFLOW_STEPS%", workflowSteps, 1)

	if scope != "" {
		scopeText := fmt.Sprintf(`
## Focus Area

The operator has described what they care about:

> %s

**IMPORTANT**: Use this description to determine what to monitor. You MUST:
1. **Discover relevant resources first**: Search for Datadog dashboards, monitors, and services that match this description. Use search/list tools to find them by name, tag, or service.
2. **Only monitor what's relevant**: Ignore alerts, errors, monitors, and tickets that are outside this scope. If an alert fires for a service the operator didn't mention or that clearly doesn't relate to their focus area, skip it entirely.
3. **Infer related services**: The operator may describe a team or product area rather than exact service names. Use your knowledge of the codebase and monitoring tools to discover which services, dashboards, and monitors belong to that area.
4. **Filter Linear tickets**: Only investigate tickets that are relevant to the focus area (matching labels, components, or services).
5. **Filter Bugsnag errors**: Only investigate errors from projects/services within scope.
6. **Filter Datadog**: Only check monitors and dashboards relevant to the described services and team.

Do NOT waste time investigating alerts outside this scope — the operator is not on call for those.
`, scope)
		prompt = strings.Replace(prompt, "%SCOPE%", scopeText, 1)
	} else {
		prompt = strings.Replace(prompt, "%SCOPE%", "", 1)
	}

	return prompt
}
