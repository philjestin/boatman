# Getting Started with Boatman

Boatman is a desktop application that provides a Claude AI agent for your codebase, with specialized **Firefighter Mode** for production incident investigation and response.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [First Run](#first-run)
- [Basic Usage](#basic-usage)
- [Firefighter Mode](#firefighter-mode)
- [Configuration](#configuration)
- [MCP Servers](#mcp-servers)
- [Troubleshooting](#troubleshooting)

---

## Prerequisites

### Required

- **macOS** (currently only darwin/arm64 supported)
- **Claude API Access** via one of:
  - Anthropic API key (get from https://console.anthropic.com)
  - Google Cloud account with Vertex AI enabled
  - OR Okta SSO (for firefighter mode)

### Optional (for Firefighter Mode)

- **Okta account** with admin access
- **Linear API token** (get from https://linear.app/settings/api)
- **Slack workspace** with bot token
- **Datadog** and **Bugsnag** access via Okta SSO
- **Git** installed for worktree-based fixes

---

## Installation

### Option 1: Build from Source

```bash
# Clone the repository
git clone <repository-url>
cd boatmanapp

# Install dependencies
go mod download

# Build the application
wails build

# The app will be at: build/bin/boatman.app
```

### Option 2: Download Release

Download the latest `.app` from the releases page and drag to `/Applications`.

---

## First Run

1. **Launch Boatman**
   ```bash
   open build/bin/boatman.app
   # or double-click the app in Finder
   ```

2. **Complete Onboarding**
   - Choose authentication method:
     - **Anthropic API**: Enter your API key
     - **Google Cloud**: Sign in with OAuth
   - Select default model (recommended: Claude Sonnet 4)
   - Choose approval mode:
     - **Suggest Mode**: You approve all changes (safest)
     - **Auto-Edit Mode**: Claude edits files, asks for bash commands
     - **Full Auto**: Claude has full control (use with caution)

3. **Create or Open a Project**
   - Click "New Project" and select your codebase directory
   - Or open a recent project from the sidebar

4. **Start Chatting**
   - Click "New Session" to start conversing with Claude
   - Ask questions about your code, request changes, or get explanations

---

## Basic Usage

### Sessions

**Creating a Session:**
```
1. Click "New Session" button
2. Type your message in the input area
3. Claude will respond with analysis, code suggestions, or actions
```

**Session Features:**
- **Favorites**: Star important sessions for quick access
- **Tags**: Organize sessions with custom tags
- **Search**: Find sessions by content or metadata
- **History**: View all past conversations and tasks

### Agent Interactions

Claude can:
- Read and analyze your codebase
- Search for files and code patterns
- Edit files with your approval
- Run bash commands (with approval in suggest/auto-edit modes)
- Create git commits and PRs
- Spawn sub-agents for complex tasks

### Approval Modes

**Suggest Mode** (Safest):
- Claude proposes all changes
- You review and approve each edit/command
- Best for production codebases

**Auto-Edit Mode** (Balanced):
- Claude can edit files directly
- Still asks permission for bash commands
- Good for trusted environments

**Full Auto Mode** (Risky):
- Claude has complete autonomy
- Use only in sandboxed/test environments

---

## Firefighter Mode

Firefighter Mode is a specialized agent for **investigating production incidents** using Linear tickets, Bugsnag errors, and Datadog monitoring.

### What Firefighter Mode Does

**Dual Workflow:**
1. **Ticket-Based Investigation** (Priority)
   - Monitors Linear triage queue for "firefighter" labeled tickets
   - Investigates tickets on-demand when you click "Investigate"
   - Extracts Bugsnag/Datadog links from ticket descriptions
   - Generates comprehensive investigation reports

2. **Proactive Monitoring** (Secondary)
   - Polls Bugsnag/Datadog every 5 minutes for NEW errors
   - Auto-investigates HIGH/Urgent severity issues
   - Creates Linear tickets for new incidents

**Investigation Workflow:**
```
1. Fetch ticket details from Linear
2. Query Bugsnag for error stacktraces and events
3. Query Datadog for logs and metrics around error time
4. Analyze git history for recent changes
5. Identify root cause with evidence
6. Generate structured investigation report
7. Update Linear ticket with findings
8. Attempt fix in isolated git worktree (if High/Urgent)
9. Run tests and create draft PR if fix succeeds
```

### Setup Firefighter Mode

#### Step 1: Configure Okta OAuth

Firefighter mode uses Okta SSO to access Datadog and Bugsnag APIs through OAuth instead of API keys.

**Prerequisites:**
- Okta account with admin access
- Datadog and Bugsnag integrated with Okta SSO

**Configuration:**

1. **Create Okta Application**
   - Go to Okta Admin Dashboard → Applications
   - Click "Create App Integration"
   - Choose "OIDC - OpenID Connect" → "Web Application"
   - **Redirect URI**: `http://localhost:8484/callback`
   - Note the **Client ID** and **Client Secret**

2. **Configure in Boatman**
   ```
   Settings → Firefighter Tab
   - Okta Domain: your-org.okta.com
   - Client ID: <from step 1>
   - Client Secret: <from step 1> (optional for public clients)
   - Click "Save Changes"
   ```

3. **Sign In**
   ```
   Click "Sign In with Okta"
   → Browser opens
   → Sign in with your Okta credentials
   → Grant permissions
   → Return to Boatman (automatically authenticated)
   ```

#### Step 2: Install Custom MCP Servers

Firefighter uses custom MCP servers that authenticate via Okta OAuth:

```bash
cd boatmanapp/mcp-servers
make all
make install
```

This installs:
- `datadog-okta` - Datadog MCP server with OAuth support
- `bugsnag-okta` - Bugsnag MCP server with OAuth support

#### Step 3: Configure MCP Servers

**Linear MCP Server:**

1. Get Linear API Key
   ```
   - Go to https://linear.app/settings/api
   - Create new API key with read/write permissions
   - Copy the key
   ```

2. Add to MCP config (`~/.claude/claude_mcp_config.json`):
   ```json
   {
     "mcpServers": {
       "linear": {
         "command": "npx",
         "args": ["-y", "@modelcontextprotocol/server-linear"],
         "env": {
           "LINEAR_API_KEY": "your-linear-api-key-here"
         }
       }
     }
   }
   ```

**Datadog & Bugsnag MCP Servers (OAuth):**

Add to `~/.claude/claude_mcp_config.json`:

```json
{
  "mcpServers": {
    "datadog-okta": {
      "command": "/Users/YOUR_USERNAME/.claude/mcp-servers/datadog-okta",
      "args": [],
      "env": {
        "OKTA_ACCESS_TOKEN": "[automatically-injected-by-boatman]",
        "DD_SITE": "datadoghq.com"
      }
    },
    "bugsnag-okta": {
      "command": "/Users/YOUR_USERNAME/.claude/mcp-servers/bugsnag-okta",
      "args": [],
      "env": {
        "OKTA_ACCESS_TOKEN": "[automatically-injected-by-boatman]"
      }
    }
  }
}
```

**Slack MCP Server (Optional):**

1. Create Slack App
   ```
   - Go to https://api.slack.com/apps
   - Create New App → From Scratch
   - Add Bot Token Scopes:
     - channels:history
     - channels:read
     - chat:write
     - users:read
   - Install to workspace
   - Copy Bot User OAuth Token
   ```

2. Add to MCP config:
   ```json
   {
     "mcpServers": {
       "slack": {
         "command": "npx",
         "args": ["-y", "@modelcontextprotocol/server-slack"],
         "env": {
           "SLACK_BOT_TOKEN": "xoxb-your-token-here",
           "SLACK_TEAM_ID": "your-workspace-id"
         }
       }
     }
   }
   ```

#### Step 4: Enable MCP Servers in Boatman

```
Settings → MCP Servers Tab
- Toggle "Enabled" for: linear, datadog-okta, bugsnag-okta, slack
- Click "Save Changes"
```

### Using Firefighter Mode

#### Starting a Firefighter Session

1. **Click "Firefighter" Button** (flame icon in header)
2. **Configure Session:**
   - Scope (optional): Specify service/team to focus on (e.g., "payment-service")
   - Enable Monitoring: Toggle for proactive monitoring
3. **Click "Start Investigation"**

#### Firefighter Interface

**Left Sidebar - Linear Triage Queue:**
- Shows all tickets labeled "firefighter" or "triage"
- Priority color-coding:
  - 🔥 Red = Urgent
  - 🟠 Orange = High
  - 🟡 Yellow = Medium
  - 🔵 Blue = Low
- Click "Investigate" on any ticket to start investigation

**Right Panel - Chat Interface:**
- Investigation reports appear here
- You can ask follow-up questions
- Sub-agents may spawn for complex tasks

**Top Bar - Monitoring Status:**
- 🟢 Active Monitoring: Checking every 5 minutes
- ⏸️ Paused: Manual investigation only
- Shows last check time and issue count

#### Investigating a Ticket

**Option A: Manual (Click "Investigate" button)**
```
1. Click "Investigate" on ticket in sidebar
2. Agent fetches ticket from Linear
3. Agent extracts Bugsnag/Datadog IDs from description
4. Agent queries monitoring tools for context
5. Agent analyzes git history
6. Agent generates investigation report
7. Agent updates Linear ticket with findings
8. Agent attempts fix (if High/Urgent priority)
```

**Option B: Automatic (Proactive Monitoring)**
```
1. Enable monitoring toggle
2. Agent polls every 5 minutes:
   - Checks Linear triage queue FIRST
   - Checks Bugsnag for new errors
   - Checks Datadog for triggered monitors
3. Auto-investigates HIGH/Urgent issues
4. Creates Linear tickets for new incidents
5. Updates you with findings
```

**Option B: Slack Alert (Future)**
```
1. Someone tags @employer-activation-firefighter in Slack
2. Agent receives alert via Slack MCP
3. Agent acknowledges in thread
4. Agent creates/finds Linear ticket
5. Agent investigates and updates both Slack + Linear
```

#### Investigation Report Format

Each investigation generates:

```markdown
### Incident Summary
- **Error**: NullPointerException in PaymentService.processRefund()
- **First Seen**: 2026-02-13 14:23:45 UTC
- **Frequency**: 47 occurrences in last hour
- **Severity**: High
- **Bugsnag**: https://app.bugsnag.com/errors/abc123

### Affected Systems
- **Services**: payment-service, billing-service
- **Code Owners**: @payments-team (git blame: john@example.com)
- **Endpoints**: POST /api/v1/payments/refund

### Timeline
- 14:20 UTC - Deployment: v2.3.4 to production
- 14:23 UTC - First error occurrence
- 14:25 UTC - Error rate spike to 15% of requests
- 14:30 UTC - Datadog alert triggered

### Error Details
[Stacktrace from Bugsnag]
[Datadog logs showing context]

### Code Analysis
**Recent Changes:**
- Commit abc123: "Refactor refund logic" by John Doe (2 hours ago)
- Modified: src/services/PaymentService.ts:142

**Problematic Code:**
Line 142 introduced null check AFTER accessing `transaction.customer`

### Root Cause Analysis
**Primary Hypothesis:**
Refund logic assumes `transaction.customer` is always present, but can be null for guest checkouts introduced in v2.3.0.

**Evidence:**
- 100% of errors have `transaction.customer === null`
- All errors from guest checkout flow
- Issue started immediately after deployment of commit abc123

### Recommended Actions
1. **Immediate**: Add null check before accessing `transaction.customer`
2. **Short-term**: Rollback to v2.3.3 if fix cannot deploy quickly
3. **Long-term**: Add integration tests for guest checkout flows

### Fix Attempted
✅ Created worktree: ../worktrees/fix-emp-456
✅ Applied fix: Added null guard in PaymentService.ts:142
✅ Tests passed: All 234 tests passing
✅ Draft PR created: #1234
🔗 Linear ticket updated with PR link

### References
- Bugsnag: https://app.bugsnag.com/errors/abc123
- Datadog Logs: https://app.datadoghq.com/logs?query=...
- Git Commit: abc123def456
- Pull Request: #1234
```

### Firefighter Best Practices

**1. Label Your Tickets**
- Use "firefighter" or "triage" labels in Linear
- Add service tags: "payment", "auth", "billing"
- Include Bugsnag/Datadog links in ticket description

**2. Monitoring Scope**
- Set scope when starting session to filter noise
- Example scopes:
  - "payment-service" - Only payment-related issues
  - "employer-activation" - Your team's services
  - "production" - All production incidents

**3. Priority Triage**
- Agent prioritizes Urgent/High tickets from Linear
- Auto-investigates HIGH severity from monitoring
- Medium/Low tickets: alert only, wait for approval

**4. Git Worktrees**
- Agent creates isolated worktrees for fixes: `../worktrees/fix-<issue-id>`
- Allows multiple simultaneous investigations
- Clean up old worktrees periodically:
  ```bash
  git worktree list
  git worktree remove ../worktrees/fix-old-issue
  ```

**5. Update Linear Tickets**
- Agent automatically updates tickets with findings
- Add additional context in comments
- Move to "In Progress" when investigating
- Move to "Ready for Fix" after investigation
- Link PRs when fix is ready

---

## Configuration

### Settings Overview

Access settings via the gear icon in the top-right.

**General Tab:**
- Authentication method (Anthropic API vs Google Cloud)
- API key or OAuth configuration
- Default model selection
- Theme (dark/light)
- Notifications

**Approval Tab:**
- Suggest Mode / Auto-Edit Mode / Full Auto Mode
- Permission granularity

**Memory Tab:**
- Max messages per session (default: 1000)
- Archive old messages (recommended: enabled)
- Auto-cleanup sessions (default: 30 days)
- Max total sessions (default: 100)

**MCP Servers Tab:**
- Enable/disable MCP servers
- Add custom MCP servers
- Configure environment variables

**Firefighter Tab:**
- Okta OAuth configuration
- Sign in/out
- Access token status

**About Tab:**
- Version information
- Links to documentation

### Project-Specific Settings

Each project can override global settings:

```
Right-click project → Project Settings
- Approval mode override
- Model override
```

---

## MCP Servers

Model Context Protocol (MCP) servers extend Claude's capabilities.

### Available Preset Servers

**Built-in Presets:**
- `filesystem` - File system access
- `github` - GitHub API integration
- `postgres` - PostgreSQL database access
- `datadog` - Datadog monitoring (API key auth)
- `bugsnag` - Bugsnag error tracking (API key auth)
- `linear` - Linear project management
- `slack` - Slack workspace integration

**Firefighter-Specific (OAuth):**
- `datadog-okta` - Datadog with Okta OAuth
- `bugsnag-okta` - Bugsnag with Okta OAuth

### Adding Custom MCP Servers

```
Settings → MCP Servers → Add Server

Option 1: Preset
- Select from dropdown
- Configure environment variables
- Enable

Option 2: Custom
- Name: my-custom-server
- Command: npx
- Args: -y, @my-org/mcp-server
- Environment variables:
  - API_KEY: your-key-here
- Enable
```

### MCP Configuration File

Direct editing: `~/.claude/claude_mcp_config.json`

```json
{
  "mcpServers": {
    "server-name": {
      "command": "npx",
      "args": ["-y", "@package/name"],
      "env": {
        "API_KEY": "your-key",
        "OTHER_VAR": "value"
      }
    }
  }
}
```

Changes take effect immediately (no restart required).

---

## Troubleshooting

### Common Issues

**"Claude CLI not found"**
```bash
# Install Claude CLI globally
npm install -g @anthropic/claude-cli

# Or use via npx (Boatman will detect)
```

**"Permission denied" errors**
```
Settings → Approval → Switch to Auto-Edit or Full Auto mode
- OR -
Approve the bash command when prompted
```

**"MCP server failed to start"**
```bash
# Test MCP server manually
npx -y @package/name

# Check logs
tail -f ~/.claude/logs/mcp-server.log

# Verify environment variables
cat ~/.claude/claude_mcp_config.json
```

**Okta OAuth "Authentication failed"**
```
1. Verify Okta domain is correct (without https://)
2. Check redirect URI in Okta app: http://localhost:8484/callback
3. Ensure client ID is correct
4. Check if client secret is required (confidential clients)
5. Verify your Okta account has access to Datadog/Bugsnag
```

**Firefighter "No tickets found"**
```
1. Verify Linear API key has read permissions
2. Check tickets have "firefighter" or "triage" labels
3. Ensure tickets are in correct project/workspace
4. Test Linear MCP server:
   npx -y @modelcontextprotocol/server-linear
```

**"Too many tokens" or context errors**
```
Settings → Memory → Reduce "Max messages per session"
- OR -
Enable "Archive old messages" to trim history
- OR -
Start a new session for new topics
```

### Debug Mode

Enable debug logging:

```bash
# Set environment variable before launching
export BOATMAN_DEBUG=1
open build/bin/boatman.app

# View logs
tail -f ~/Library/Logs/boatman/debug.log
```

### Getting Help

**Support Channels:**
- GitHub Issues: [repository-url]/issues
- Documentation: [repository-url]/docs
- Community: [community-link]

**When Reporting Bugs:**
Include:
1. Boatman version (Settings → About)
2. Operating system version
3. Steps to reproduce
4. Error messages / screenshots
5. Relevant logs

---

## Advanced Usage

### Custom System Prompts

Edit session behavior with custom prompts:

```go
// In agent/prompts.go
const CustomAgentPrompt = `You are a specialized agent for [purpose]...`
```

### Keyboard Shortcuts

- `Cmd+N` - New session
- `Cmd+K` - Quick search
- `Cmd+,` - Settings
- `Cmd+\` - Toggle sidebar
- `Cmd+Enter` - Send message
- `Esc` - Cancel current operation

### Session Tags

Organize sessions with tags:
```
#bug-fix #refactor #investigation #deploy #review
```

Filter by tag in sidebar search.

### Git Integration

Boatman can:
- Create commits with proper messages
- Generate PRs with descriptions
- Create worktrees for parallel work (firefighter mode)
- Run git blame for code ownership

Configure git hooks in settings if needed.

---

## Security Best Practices

**1. API Keys**
- Never commit API keys to version control
- Rotate keys regularly
- Use OAuth when available (Google Cloud, Okta)

**2. Approval Modes**
- Use Suggest Mode in production codebases
- Only use Full Auto in sandboxed environments
- Review all bash commands before approval

**3. MCP Servers**
- Only install trusted MCP servers
- Review server source code before enabling
- Use environment variables for secrets (never hardcode)

**4. Firefighter Mode**
- Limit OAuth scopes to minimum required
- Review auto-fixes before merging PRs
- Keep worktrees isolated from main branch

**5. Data Privacy**
- Code sent to Claude API follows Anthropic's privacy policy
- Self-hosted option: Use Google Cloud with private VPC
- Disable telemetry in settings if required

---

## License

[Include your license information here]

## Contributing

[Include contribution guidelines here]

## Changelog

See [CHANGELOG.md](./CHANGELOG.md) for version history and updates.
