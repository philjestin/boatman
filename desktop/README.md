# Boatman

> A desktop application that brings Claude AI to your codebase with specialized **Firefighter Mode** for production incident investigation.

## About

Boatman is a native desktop application built with Wails (Go + React) that provides a powerful interface for Claude AI. It includes a specialized **Firefighter Mode** that integrates Linear, Bugsnag, and Datadog for automated production incident investigation and response.

## Features

### 🤖 Claude AI Agent
- **Interactive Chat**: Ask questions, request changes, get explanations about your codebase
- **Code Analysis**: Deep understanding of your project structure and patterns
- **Autonomous Actions**: Edit files, run commands, create commits with your approval
- **Sub-Agent Tracking**: Spawns specialized agents for complex multi-step tasks, with automatic context switching and per-agent message attribution
- **Agent Logs Panel**: View messages grouped by agent in separate tabs with status indicators (active/completed)
- **Agent Badges**: Messages show which agent produced them via inline badges
- **Project Management**: Open and manage multiple coding projects
- **Task Tracking**: View and monitor tasks created by the agent during sessions
- **Agent Logs**: Real-time streaming logs panel to see agent activity and tool usage
- **Task Detail Modal**: Clickable task cards showing diffs, feedback, plans, and issues

### 🔥 Firefighter Mode
- **Linear Integration**: Automatically monitors triage queue for production incidents
- **Error Investigation**: Connects Bugsnag errors with Datadog logs and git history
- **Root Cause Analysis**: Correlates deployments, code changes, and error patterns
- **Auto-Fix**: Creates isolated worktrees, attempts fixes, runs tests, and opens PRs
- **Dual Workflow**: Handles both ticket-based investigations and proactive monitoring
- **Slack Integration**: Responds to @mentions for urgent production issues

### 🔌 MCP Integration
- **Extensible**: Connect to any MCP-compatible tool or service
- **Built-in Servers**: GitHub, Datadog, Bugsnag, Linear, Slack, and more
- **OAuth Support**: Authenticate via Okta SSO for enterprise integrations
- **Custom Servers**: Build your own MCP servers for specialized workflows
- **MCP Server Dialog**: Easy configuration and management of MCP servers via UI

### 🔍 Advanced UI Features
- **Smart Search**: Full-text search across sessions with filters (tags, dates, favorites, projects)
- **Favorites & Tags**: Organize sessions with favorites and custom tags for easy retrieval
- **Batch Diff Approval**: Select and approve/reject multiple file changes at once
- **Inline Diff Comments**: Add threaded comments directly on diff lines for review discussions
- **Diff Summary Cards**: Quick overview of file changes with stats (additions, deletions, modifications)
- **Onboarding Wizard**: Guided first-run experience for setting up authentication and preferences
- **Model Selection**: Choose between Claude Opus 4.6, Sonnet 4.5, Haiku 4, and Claude 3.5 Sonnet
- **Session Modes**: Visual badges to distinguish standard, firefighter, and boatmanmode sessions

## Quick Start

```bash
# Clone and build
git clone <repository-url>
cd boatmanapp
wails build

# Run
open build/bin/boatman.app
```

**📚 First time?** See the complete [Getting Started Guide](./GETTING_STARTED.md)

## Prerequisites

### Required
- **Go 1.18+**: Build the application
- **Node.js 16+**: Frontend development
- **Wails CLI**: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
- **Claude API**: Get key from https://console.anthropic.com

### For Firefighter Mode
- **Okta Account**: For SSO authentication
- **Linear API Token**: From https://linear.app/settings/api
- **Datadog/Bugsnag**: Access via Okta SSO
- **Slack Bot** (optional): For alert integration

## Installation

1. Clone the repository:
   ```bash
   git clone <repository-url>
   cd boatmanapp
   ```

2. Install dependencies:
   ```bash
   wails doctor  # Check if all dependencies are installed
   cd frontend && npm install && cd ..
   ```

3. Build the application:
   ```bash
   wails build
   ```

The built application will be in the `build/bin` directory.

## Firefighter Mode

Firefighter Mode is designed for **on-call engineers** handling production incidents. It combines Linear ticket management with Bugsnag error tracking and Datadog monitoring for automated investigation.

### How It Works

```
Linear Ticket → Extract Context → Query Bugsnag/Datadog → Analyze Git History → Generate Report → Auto-Fix → Update Ticket
```

**Workflow:**
1. **Monitor Linear Queue**: Checks triage queue for tickets labeled "firefighter"
2. **Investigate on Demand**: Click "Investigate" button on any ticket
3. **Gather Context**: Fetches error details from Bugsnag, logs from Datadog
4. **Code Analysis**: Uses git blame to find recent changes and code owners
5. **Root Cause**: Correlates deployments, errors, and code changes
6. **Generate Report**: Creates comprehensive investigation summary
7. **Attempt Fix**: For High/Urgent tickets, creates worktree and implements fix
8. **Test & PR**: Runs tests, opens draft PR if tests pass
9. **Update Ticket**: Adds findings and PR link to Linear ticket

### Key Capabilities

- ✅ **Automated Triage**: Monitors Linear queue every 5 minutes
- ✅ **Context Correlation**: Links Bugsnag errors with Datadog logs and git commits
- ✅ **Code Ownership**: Identifies responsible teams via git blame
- ✅ **Isolated Fixes**: Uses git worktrees for safe, parallel investigations
- ✅ **Test Validation**: Runs test suite before creating PRs
- ✅ **Documentation**: Generates structured reports for postmortems

**See the [Firefighter Mode Guide](./GETTING_STARTED.md#firefighter-mode) for setup instructions.**

## Getting Started

### First Launch

1. Launch Boatman
2. Complete onboarding: Choose auth method (API key or Google Cloud OAuth)
3. Select default model and approval mode
4. Create or open a project
5. Start chatting!

**For detailed walkthrough**, see [Getting Started → First Run](./GETTING_STARTED.md#first-run)

## Usage

### Standard Mode

**Ask Claude anything about your codebase:**
```
"Review this pull request and suggest improvements"
"Find why the authentication is failing"
"Refactor the payment service to use async/await"
"Explain how the database migration system works"
```

**Claude can:**
- Read and analyze code
- Edit files with your approval
- Run bash commands (git, npm, tests, etc.)
- Create commits and pull requests
- Spawn sub-agents for complex tasks (tracked in the Agent Logs panel)

### BoatmanMode Integration

**Full Claude event streaming:**
- Each workflow phase (Planning, Execution, Review, Refactor) creates a separate agent tab
- Claude's raw streaming output is forwarded via `claude_stream` events for full visibility
- Progress messages and agent lifecycle events appear as system messages in the chat
- See [BOATMANMODE_INTEGRATION.md](./BOATMANMODE_INTEGRATION.md) for details

### Firefighter Mode

**Investigate production incidents:**

1. Click "Firefighter" button (flame icon)
2. View Linear triage queue in left sidebar
3. Click "Investigate" on any ticket
4. Agent automatically:
   - Fetches Bugsnag error details
   - Queries Datadog logs
   - Analyzes git history
   - Generates investigation report
   - Attempts fix (if High/Urgent)
   - Updates Linear ticket

### BoatmanMode Integration

**Autonomous ticket execution with full workflow:**

1. Click "Boatman Mode" button (purple button in header)
2. Enter Linear ticket ID
3. Watch real-time execution with structured events:
   - Planning phase with codebase analysis
   - Implementation with code generation
   - Test execution and validation
   - Peer review with feedback
   - Automated refactoring
   - PR creation and ticket updates
4. Track progress in Tasks tab with clickable task details
5. Review diffs, feedback, and issues in task modals

**Key Benefits:**
- ✅ Full autonomous workflow (plan → execute → review → refactor → PR)
- ✅ Git worktree isolation for safe parallel work
- ✅ Real-time event streaming with structured task tracking
- ✅ Task metadata includes diffs, plans, feedback, and issues
- ✅ Integrated with CLI for subprocess execution

**See documentation:**
- [BoatmanMode Integration Guide](./BOATMANMODE_INTEGRATION.md)
- [BoatmanMode Events Specification](./BOATMANMODE_EVENTS.md)
- [BoatmanMode Implementation Details](./BOATMANMODE_IMPLEMENTATION.md)

**For detailed usage**, see:
- [Basic Usage Guide](./GETTING_STARTED.md#basic-usage)
- [Firefighter Mode Guide](./GETTING_STARTED.md#firefighter-mode)

## Development

### Live Development Mode

Run the app in development mode with hot reload:

```bash
wails dev
```

This starts:
- A Vite development server for fast frontend hot reload
- A dev server at http://localhost:34115 for browser-based development

### Project Structure

```
boatmanapp/
├── frontend/          # React TypeScript frontend
│   ├── src/
│   │   ├── components/  # UI components (MessageBubble, AgentLogsPanel, etc.)
│   │   ├── hooks/       # React hooks (useAgent for event handling)
│   │   ├── types/       # TypeScript types (AgentInfo, Message, etc.)
│   │   └── store/       # State management
├── agent/             # Agent session management (subagent tracking, stream parsing)
├── boatmanmode/       # BoatmanMode CLI integration (subprocess, event routing)
├── project/           # Project and workspace management
├── config/            # Configuration and preferences
├── git/               # Git integration
├── diff/              # Diff parsing and rendering
├── mcp/               # MCP server management
├── app.go             # Main application logic (event routing, session management)
└── main.go            # Application entry point
```

### Configuration

**User Configuration:**
- Settings: `~/.boatman/config.json`
- MCP Servers: `~/.claude/claude_mcp_config.json`
- Sessions: `~/.boatman/sessions/`

**For detailed configuration**, see [Configuration Guide](./GETTING_STARTED.md#configuration)

**Build Configuration:**
Edit `wails.json` for build settings. See https://wails.io/docs/reference/project-config

## Building

### Development Build

```bash
wails build
```

### Production Build

```bash
wails build -clean -production
```

### Platform-Specific Builds

```bash
# macOS
wails build -platform darwin/universal

# Windows
wails build -platform windows/amd64

# Linux
wails build -platform linux/amd64
```

Built applications will be in `build/bin/`.

## Troubleshooting

**Common Issues:**

```bash
# App won't launch
xattr -d com.apple.quarantine build/bin/boatman.app

# Check dependencies
wails doctor

# Test MCP server
npx -y @package/mcp-server
```

**Full troubleshooting guide:** [GETTING_STARTED.md → Troubleshooting](./GETTING_STARTED.md#troubleshooting)

## Documentation

- **[Getting Started Guide](./GETTING_STARTED.md)** - Complete setup and usage
- **[Features Guide](./FEATURES.md)** - Comprehensive feature documentation (NEW)
- **[Firefighter Mode](./GETTING_STARTED.md#firefighter-mode)** - Production incident investigation
- **[BoatmanMode Integration](./BOATMANMODE_INTEGRATION.md)** - Autonomous ticket execution
- **[BoatmanMode Events](./BOATMANMODE_EVENTS.md)** - Event specification and integration
- **[MCP Servers](./GETTING_STARTED.md#mcp-servers)** - Extending capabilities
- **[Configuration](./GETTING_STARTED.md#configuration)** - Settings and customization
- **[BoatmanMode Integration](./BOATMANMODE_INTEGRATION.md)** - CLI integration and event routing
- **[BoatmanMode Events](./BOATMANMODE_EVENTS.md)** - Event system and subagent tracking
- **[Changelog](./CHANGELOG.md)** - Version history and release notes

## FAQ

**Q: What models are supported?**
A: Claude Opus 4.6, Sonnet 4.5, Haiku 4, and Claude 3.5 Sonnet.

**Q: How much does it cost?**
A: Boatman is free. You pay for Claude API usage (via Anthropic or Google Cloud).

**Q: Is my code secure?**
A: Code is sent to Claude API following Anthropic's privacy policy. Use Google Cloud for private VPC.

**Q: Can I use it offline?**
A: No, requires internet for Claude API and MCP servers.

**Q: What OS is supported?**
A: Currently macOS (M1/M2/M3). Windows/Linux support planned.

**More FAQs:** [Getting Started → FAQ](./GETTING_STARTED.md#faq)

## Contributing

Contributions welcome! Please:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

See [CONTRIBUTING.md](./CONTRIBUTING.md) for detailed guidelines.

## License

[Add your license here - e.g., MIT, Apache 2.0, etc.]

## Support

- **Documentation**: [Getting Started Guide](./GETTING_STARTED.md)
- **Issues**: [GitHub Issues](https://github.com/your-org/boatmanapp/issues)
- **Discussions**: [GitHub Discussions](https://github.com/your-org/boatmanapp/discussions)

## Acknowledgments

- **Anthropic** - Claude AI and Model Context Protocol
- **Wails** - Native app framework
- **Linear, Datadog, Bugsnag** - Monitoring and issue tracking integrations

---

**Built with ❤️ for on-call engineers**
