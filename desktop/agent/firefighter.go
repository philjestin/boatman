package agent

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

// FirefighterMonitor manages active monitoring for a firefighter session
type FirefighterMonitor struct {
	session          *Session
	isActive         bool
	checkInterval    time.Duration
	sessionStartTime time.Time       // When the session was created — used as the time floor
	lastCheckTime    time.Time
	isFirstCheck     bool            // True until the first check completes
	seenIssues       map[string]bool // Track issues we've already investigated
	onAlert          func(Alert)
	onInvestigation  func(Investigation)
	mu               sync.RWMutex
	ctx              context.Context
	cancel           context.CancelFunc
}

// Alert represents a new issue detected by monitoring
type Alert struct {
	ID          string    `json:"id"`
	Source      string    `json:"source"` // "bugsnag", "datadog", "linear", "slack"
	Severity    string    `json:"severity"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	FirstSeen   time.Time `json:"firstSeen"`
	Count       int       `json:"count"`
	URL         string    `json:"url"`
	LinearID    string    `json:"linearId,omitempty"`    // Linear issue ID if from ticket
	SlackThread string    `json:"slackThread,omitempty"` // Slack thread if from alert
}

// Investigation represents an ongoing investigation
type Investigation struct {
	ID             string    `json:"id"`
	AlertID        string    `json:"alertId"`
	LinearIssueID  string    `json:"linearIssueId,omitempty"`
	Status         string    `json:"status"` // "investigating", "fixing", "testing", "done", "failed"
	WorktreePath   string    `json:"worktreePath,omitempty"`
	BranchName     string    `json:"branchName,omitempty"`
	PRNumber       string    `json:"prNumber,omitempty"`
	StartedAt      time.Time `json:"startedAt"`
	CompletedAt    time.Time `json:"completedAt,omitempty"`
	Summary        string    `json:"summary,omitempty"`
	FixDescription string    `json:"fixDescription,omitempty"`
}

// NewFirefighterMonitor creates a new firefighter monitor
func NewFirefighterMonitor(session *Session) *FirefighterMonitor {
	ctx, cancel := context.WithCancel(context.Background())
	return &FirefighterMonitor{
		session:          session,
		isActive:         false,
		checkInterval:    5 * time.Minute, // Check every 5 minutes by default
		sessionStartTime: time.Now(),
		isFirstCheck:     true,
		seenIssues:       make(map[string]bool),
		ctx:              ctx,
		cancel:           cancel,
	}
}

// Start begins active monitoring
func (fm *FirefighterMonitor) Start() error {
	fm.mu.Lock()
	if fm.isActive {
		fm.mu.Unlock()
		return fmt.Errorf("monitor already active")
	}
	fm.isActive = true
	fm.mu.Unlock()

	// Start monitoring loop in background
	go fm.monitorLoop()

	return nil
}

// Stop halts active monitoring
func (fm *FirefighterMonitor) Stop() {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	if !fm.isActive {
		return
	}

	fm.isActive = false
	fm.cancel()
}

// SetCheckInterval updates the monitoring interval
func (fm *FirefighterMonitor) SetCheckInterval(interval time.Duration) {
	fm.mu.Lock()
	defer fm.mu.Unlock()
	fm.checkInterval = interval
}

// SetAlertHandler sets the callback for new alerts
func (fm *FirefighterMonitor) SetAlertHandler(handler func(Alert)) {
	fm.mu.Lock()
	defer fm.mu.Unlock()
	fm.onAlert = handler
}

// SetInvestigationHandler sets the callback for investigation updates
func (fm *FirefighterMonitor) SetInvestigationHandler(handler func(Investigation)) {
	fm.mu.Lock()
	defer fm.mu.Unlock()
	fm.onInvestigation = handler
}

// monitorLoop runs the monitoring loop
func (fm *FirefighterMonitor) monitorLoop() {
	ticker := time.NewTicker(fm.checkInterval)
	defer ticker.Stop()

	// Run initial check immediately
	fm.performCheck()

	for {
		select {
		case <-fm.ctx.Done():
			return
		case <-ticker.C:
			fm.performCheck()
		}
	}
}

// performCheck executes a monitoring check
func (fm *FirefighterMonitor) performCheck() {
	fm.mu.Lock()
	if !fm.isActive {
		fm.mu.Unlock()
		return
	}
	fm.lastCheckTime = time.Now()
	firstCheck := fm.isFirstCheck
	fm.mu.Unlock()

	// Build monitoring prompt
	prompt := fm.buildMonitoringPrompt()

	// After first check, subsequent checks only look at new data
	if firstCheck {
		fm.mu.Lock()
		fm.isFirstCheck = false
		fm.mu.Unlock()
	}

	// Send to Claude for analysis
	// This will trigger the normal message flow, but with a specialized prompt
	if err := fm.session.SendMessage(prompt, AuthConfig{}); err != nil {
		fmt.Printf("[firefighter] Failed to send monitoring check: %v\n", err)
	}
}

// lookbackStart returns the earliest time to search for issues.
// First check: max 24 hours before session start.
// Subsequent checks: since the last check time.
func (fm *FirefighterMonitor) lookbackStart() time.Time {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	if fm.isFirstCheck {
		earliest := fm.sessionStartTime.Add(-24 * time.Hour)
		return earliest
	}
	return fm.lastCheckTime.Add(-fm.checkInterval)
}

// buildMonitoringPrompt creates a prompt for proactive monitoring
func (fm *FirefighterMonitor) buildMonitoringPrompt() string {
	lookback := fm.lookbackStart()
	lookbackStr := lookback.Format(time.RFC3339)
	sessionStartStr := fm.sessionStartTime.Format(time.RFC3339)

	fm.mu.RLock()
	firstCheck := fm.isFirstCheck
	fm.mu.RUnlock()

	var timeContext string
	if firstCheck {
		timeContext = fmt.Sprintf(`**TIME WINDOW**: This is the **first check** for this session.
- Session started at: %s
- Look back at most **24 hours** (since %s) to catch any recent issues
- After this check, subsequent checks will only look at **new** data since the previous check
- Do NOT investigate old/resolved issues — focus on currently active or recently triggered alerts`, sessionStartStr, lookbackStr)
	} else {
		timeContext = fmt.Sprintf(`**TIME WINDOW**: This is a **follow-up check**.
- Session started at: %s
- Only look at data **since the last check**: %s
- Ignore anything older — it was already checked
- Focus exclusively on NEW alerts, errors, and tickets created since the last check`, sessionStartStr, lookbackStr)
	}

	return `🔥 FIREFIGHTER MONITORING CHECK 🔥

You are in active firefighter monitoring mode. You handle BOTH proactive monitoring AND ticket-based investigations.

**CRITICAL**: Use your available MCP tools (` + fm.availableMCPToolsList() + `). DO NOT try to use bash commands - these are MCP tools that provide direct API access.

` + timeContext + `

**WORKFLOW**:

## PART 1: CHECK LINEAR TRIAGE QUEUE (Priority)

1. **Query Linear using MCP tools**:
   - Use Linear MCP tool to list issues in the triage queue
   - Filter by labels: "firefighter", "triage", or team-specific tags
   - Only consider tickets created or updated since ` + lookbackStr + `
   - Sort by priority (Urgent > High > Medium > Low)
   - Look for issues with Bugsnag/Datadog links in description

2. **For each high-priority ticket**:
   - Extract Bugsnag error ID or Datadog monitor ID from ticket description
   - Use those IDs to query detailed context from Bugsnag/Datadog
   - Start investigation if not already in progress
   - Update Linear ticket with investigation status

## PART 2: PROACTIVE MONITORING (Secondary)

3. **Query Bugsnag using MCP tools** for errors since ` + lookbackStr + `:
   - Use Bugsnag MCP tool to list recent errors
   - Look for NEW error types (not previously seen)
   - Focus on errors with increasing frequency
   - Prioritize critical/high severity
   - Check if Linear ticket already exists for this error

4. **Query Datadog using MCP tools** for triggered monitors:
   - Use Datadog MCP tool to check for currently triggered monitors
   - Look for monitors that transitioned to alert state since ` + lookbackStr + `
   - Check for log volume spikes and metric anomalies (error rates, latency, etc.)
   - Check if Linear ticket already exists for this alert

5. **Compare against previous issues**:
   - Only report NEWLY discovered issues since ` + lookbackStr + `
   - Skip if you've already investigated this issue
   - Skip if Linear ticket already exists

## REPORTING

6. **Report findings**:
   - **High-priority tickets**: "🎫 [Priority] Linear ticket [ID]: [title]"
   - **New proactive issues**: "🚨 NEW [severity]: [title]"
   - **All clear**: "✅ Linear queue: N tickets, Monitoring: No new issues"

7. **Auto-investigate based on priority**:
   - **Urgent/High from Linear**: Immediately investigate
   - **HIGH severity from monitoring**: Immediately investigate
   - **Medium/Low**: Alert only, wait for user approval

8. **Investigation workflow**:
   - Create git worktree: git worktree add ../worktrees/fix-[issue-id] -b fix/[issue-name]
   - Investigate root cause using Bugsnag/Datadog context
   - Attempt fix if straightforward
   - Run tests
   - If tests pass, create draft PR and update Linear ticket

` + fm.buildSlackMonitoringSection() + `

` + fm.buildScopeReminder() + `
**IMPORTANT**:
- **Prioritize Linear tickets over proactive monitoring**
- Use MCP tools, NOT bash commands
- Update Linear tickets with investigation progress
- Track which issues you've seen to avoid duplicates
- Be concise in monitoring reports
- **Only look at data since ` + lookbackStr + `** — do not scan all historical data

Last check time: ` + fm.lastCheckTime.Format(time.RFC3339) + `

Begin monitoring check now. Start with Linear triage queue, then proactive monitoring` + fm.slackMonitoringReminder() + `.`
}

// buildScopeReminder returns a scope reminder for the monitoring prompt if scope is configured
func (fm *FirefighterMonitor) buildScopeReminder() string {
	scope, _ := fm.session.ModeConfig["scope"].(string)
	if scope == "" {
		return ""
	}
	return fmt.Sprintf(`
**SCOPE REMINDER**: The operator's focus area is:
> %s
Only investigate alerts, errors, monitors, and tickets relevant to this scope. Skip everything else.

`, scope)
}

// availableMCPToolsList returns a comma-separated list of available MCP tool names
func (fm *FirefighterMonitor) availableMCPToolsList() string {
	return availableMCPToolsListFromConfig(fm.session.ModeConfig)
}

// availableMCPToolsListFromConfig extracts MCP server names from ModeConfig
func availableMCPToolsListFromConfig(modeConfig map[string]interface{}) string {
	var names []string
	if raw, ok := modeConfig["mcpServers"]; ok {
		if strs, ok := raw.([]string); ok {
			names = strs
		} else if ifaces, ok := raw.([]interface{}); ok {
			for _, v := range ifaces {
				if name, ok := v.(string); ok {
					names = append(names, name)
				}
			}
		}
	}
	if len(names) == 0 {
		return "Linear, Bugsnag, Datadog"
	}
	return strings.Join(names, ", ")
}

// hasSlackMCP checks if a Slack MCP server is configured
func (fm *FirefighterMonitor) hasSlackMCP() bool {
	return hasMCPServer(fm.getMCPNames(), "slack")
}

// getMCPNames extracts MCP server names from ModeConfig
func (fm *FirefighterMonitor) getMCPNames() []string {
	if raw, ok := fm.session.ModeConfig["mcpServers"]; ok {
		if strs, ok := raw.([]string); ok {
			return strs
		}
		if ifaces, ok := raw.([]interface{}); ok {
			var names []string
			for _, v := range ifaces {
				if name, ok := v.(string); ok {
					names = append(names, name)
				}
			}
			return names
		}
	}
	return nil
}

// buildSlackMonitoringSection returns the PART 3 section for Slack channel monitoring if channels are configured
func (fm *FirefighterMonitor) buildSlackMonitoringSection() string {
	slackChannels, _ := fm.session.ModeConfig["slackChannels"].(string)
	if slackChannels == "" || !fm.hasSlackMCP() {
		return ""
	}

	lookback := fm.lookbackStart()

	return fmt.Sprintf(`
## PART 3: SLACK CHANNEL MONITORING

5b. **Monitor Slack channels for Datadog alerts**:
   - Use Slack MCP tools to read recent messages from these channels: %s
   - Only look at messages posted since %s
   - Look for Datadog alert messages (triggered monitors, error spikes, anomalies)
   - Look for bot messages from Datadog containing monitor alerts or warning/error notifications
   - Skip messages you have already investigated (track by timestamp or thread ID)

6b. **For each NEW alert found in Slack**:
   - Parse the alert: extract monitor name, service, severity, and any Datadog links
   - Use Datadog MCP tools to gather full context (query the monitor, check related logs/metrics)
   - Spawn a sub-agent using the Agent tool to investigate and attempt a fix in an isolated worktree:
     - Use the Agent tool with isolation: "worktree" to investigate the alert
     - The sub-agent should: analyze root cause, attempt fix, run tests, create draft PR if tests pass
   - Reply in the Slack thread acknowledging the alert and sharing findings
   - Track the alert to avoid duplicate investigations
`, slackChannels, lookback.Format(time.RFC3339))
}

// slackMonitoringReminder returns a reminder suffix if Slack channels are configured and Slack MCP is available
func (fm *FirefighterMonitor) slackMonitoringReminder() string {
	slackChannels, _ := fm.session.ModeConfig["slackChannels"].(string)
	if slackChannels == "" || !fm.hasSlackMCP() {
		return ""
	}
	return ", then check Slack channels"
}

// IsActive returns whether monitoring is active
func (fm *FirefighterMonitor) IsActive() bool {
	fm.mu.RLock()
	defer fm.mu.RUnlock()
	return fm.isActive
}

// GetStatus returns current monitoring status
func (fm *FirefighterMonitor) GetStatus() map[string]interface{} {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	return map[string]interface{}{
		"active":        fm.isActive,
		"checkInterval": fm.checkInterval.String(),
		"lastCheck":     fm.lastCheckTime,
		"seenIssues":    len(fm.seenIssues),
	}
}

// MarkIssueSeen marks an issue as already investigated
func (fm *FirefighterMonitor) MarkIssueSeen(issueID string) {
	fm.mu.Lock()
	defer fm.mu.Unlock()
	fm.seenIssues[issueID] = true
}

// ClearSeenIssues resets the seen issues cache
func (fm *FirefighterMonitor) ClearSeenIssues() {
	fm.mu.Lock()
	defer fm.mu.Unlock()
	fm.seenIssues = make(map[string]bool)
}

// InvestigateLinearTicket triggers investigation for a specific Linear ticket
func (fm *FirefighterMonitor) InvestigateLinearTicket(linearIssueID string) error {
	prompt := fm.buildTicketInvestigationPrompt(linearIssueID)

	// Send to Claude for investigation
	if err := fm.session.SendMessage(prompt, AuthConfig{}); err != nil {
		return fmt.Errorf("failed to send investigation request: %w", err)
	}

	return nil
}

// buildTicketInvestigationPrompt creates a prompt for investigating a specific Linear ticket
func (fm *FirefighterMonitor) buildTicketInvestigationPrompt(linearIssueID string) string {
	return fmt.Sprintf(`🔥 FIREFIGHTER TICKET INVESTIGATION 🔥

You have been assigned to investigate Linear ticket: %s

**CRITICAL**: Use your available MCP tools to query Linear, Bugsnag, and Datadog. DO NOT use bash commands.

**INVESTIGATION WORKFLOW**:

1. **Fetch Linear Ticket Details**:
   - Use Linear MCP tool to get full details of ticket %s
   - Extract: title, description, priority, labels, comments
   - Look for Bugsnag error IDs or Datadog monitor IDs in description/comments

2. **Gather Context from Monitoring Tools**:
   - If Bugsnag error ID found:
     - Use Bugsnag MCP tool to fetch error details, stacktrace, events
     - Get recent occurrences and frequency trend
   - If Datadog monitor/link found:
     - Use Datadog MCP tool to query related logs around the time of the error
     - Check metrics for anomalies
   - If no IDs found:
     - Search Bugsnag for errors matching ticket description
     - Search Datadog logs for related errors

3. **Code Analysis**:
   - Use git log/blame to find recent changes related to the error
   - Identify code owners from git history
   - Look for related commits in the timeframe

4. **Root Cause Analysis**:
   - Correlate timeline: deployment times, error start time, code changes
   - Identify primary hypothesis with evidence
   - Determine if this is a regression, new bug, or infrastructure issue

5. **Investigation Report**:
   Generate a structured report with:
   - **Summary**: One-line description of the issue
   - **Severity**: Based on impact and frequency
   - **Root Cause**: Primary hypothesis with evidence
   - **Affected Systems**: Services, endpoints, user impact
   - **Timeline**: When it started, deployment correlation
   - **Code Analysis**: Recent changes, potential culprits
   - **Recommended Actions**: Immediate fix, short-term mitigation, long-term prevention

6. **Update Linear Ticket**:
   - Use Linear MCP tool to add a comment with investigation findings
   - Update ticket status to "In Progress" or "Ready for Fix"
   - Add relevant labels (e.g., "root-cause-identified", "needs-deployment")

7. **Attempt Fix (if High/Urgent priority)**:
   - Create git worktree: git worktree add ../worktrees/fix-%s -b fix/[issue-name]
   - Implement the fix
   - Run tests
   - If tests pass: create draft PR, link in Linear ticket
   - If tests fail: document findings in Linear ticket

**IMPORTANT**:
- Be thorough - this is a specific ticket investigation, not a quick scan
- Update Linear ticket with progress
- Use MCP tools for all external queries
- Include links to Bugsnag errors, Datadog logs, git commits in your report
- If you cannot determine root cause, document what you tried and what's still unclear

Begin investigation now.`, linearIssueID, linearIssueID, linearIssueID)
}

// InvestigateSlackAlert triggers investigation for a Slack alert mention
func (fm *FirefighterMonitor) InvestigateSlackAlert(slackThreadID, alertMessage string) error {
	prompt := fm.buildSlackAlertPrompt(slackThreadID, alertMessage)

	// Send to Claude for investigation
	if err := fm.session.SendMessage(prompt, AuthConfig{}); err != nil {
		return fmt.Errorf("failed to send alert investigation: %w", err)
	}

	return nil
}

// buildSlackAlertPrompt creates a prompt for investigating a Slack alert
func (fm *FirefighterMonitor) buildSlackAlertPrompt(slackThreadID, alertMessage string) string {
	return fmt.Sprintf(`🔥 FIREFIGHTER SLACK ALERT RESPONSE 🔥

You have been tagged in Slack for an urgent issue.

**Slack Thread**: %s
**Alert Message**: %s

**CRITICAL**: Use your available MCP tools to query Linear, Bugsnag, Datadog, and Slack. DO NOT use bash commands.

**ALERT RESPONSE WORKFLOW**:

1. **Parse Alert Message**:
   - Extract key information: service name, error type, severity
   - Look for Bugsnag/Datadog links in the message
   - Identify if this relates to an existing Linear ticket

2. **Check for Existing Ticket**:
   - Use Linear MCP tool to search for related issues
   - If ticket exists: Add Slack thread link to ticket and proceed with investigation
   - If no ticket: Create new Linear ticket with alert details

3. **Gather Context**:
   - Use Bugsnag/Datadog MCP tools to fetch error details
   - Get recent occurrences, stacktraces, logs
   - Check deployment timeline

4. **Quick Assessment**:
   - Determine severity and impact
   - Identify if immediate action required
   - Check if rollback needed

5. **Respond in Slack**:
   - Use Slack MCP tool to reply in thread with:
     - Acknowledgment: "🔥 On it - investigating now"
     - Initial findings (within 2-3 minutes)
     - Linear ticket link
     - Next steps

6. **Full Investigation**:
   - Follow standard investigation workflow
   - Update both Slack thread and Linear ticket with progress
   - If urgent: Attempt immediate fix in worktree

**IMPORTANT**:
- **Speed matters** - provide quick initial response in Slack
- Create Linear ticket if one doesn't exist
- Keep Slack thread updated with progress
- Escalate if issue is beyond your capability to fix

Begin investigation now. Prioritize quick Slack acknowledgment.`, slackThreadID, alertMessage)
}
