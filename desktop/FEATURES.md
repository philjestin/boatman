# Desktop App Features Guide

Complete guide to all features available in the Boatman desktop application.

## Table of Contents

- [Session Management](#session-management)
- [Search & Organization](#search--organization)
- [Diff Viewing & Review](#diff-viewing--review)
- [Task Tracking](#task-tracking)
- [BoatmanMode Integration](#boatmanmode-integration)
- [Firefighter Mode](#firefighter-mode)
- [Agent Logs](#agent-logs)
- [Settings & Configuration](#settings--configuration)
  - [General Settings](#general-settings)
  - [Firefighter Settings](#firefighter-settings)
  - [MCP Servers](#mcp-servers)
  - [Memory & Storage](#memory--storage)
  - [Appearance & Theming](#appearance--theming)
- [Keyboard Shortcuts](#keyboard-shortcuts)
- [Tips & Best Practices](#tips--best-practices)
- [Troubleshooting](#troubleshooting)

---

## Session Management

### Creating Sessions

**Standard Mode** (Interactive Chat):
1. Click "New Chat" or press `Cmd+N`
2. Select model (Opus 4.6, Sonnet 4.5, Haiku 4, Claude 3.5 Sonnet)
3. Start chatting with Claude about your codebase

**BoatmanMode** (Autonomous Execution):
1. Click "Boatman Mode" button (purple, in header)
2. Enter Linear ticket ID
3. Watch autonomous workflow execute

**Firefighter Mode** (Incident Investigation):
1. Click "Firefighter" button (flame icon)
2. Select ticket from triage queue
3. Click "Investigate" to start analysis

### Session Modes

Each session has a visual indicator:
- **Standard**: No badge (default)
- **Firefighter**: 🔥 Flame badge (orange)
- **BoatmanMode**: ▶️ Play badge (purple)

### Session Organization

**Favorites**:
- Click star icon on any session
- Favorited sessions appear at the top of the list
- Filter to show only favorites in search

**Tags**:
- Add custom tags to sessions for organization
- Examples: `bug-fix`, `feature`, `urgent`, `backend`
- Filter by tags in search modal

**Project-based**:
- Sessions automatically linked to project path
- Filter by project to see all related work
- Switch between projects easily

---

## Search & Organization

### Quick Search

**Keyboard shortcut**: `Cmd+K` (macOS) or `Ctrl+K` (Windows/Linux)

**Search capabilities**:
- Full-text search across session content
- Search in messages, prompts, and responses
- Real-time results as you type

### Advanced Filters

Click "Filters" in search modal to access:

**By Tags**:
- Select one or more tags
- Shows sessions matching ANY selected tag

**By Project**:
- Filter sessions from specific project path
- Useful for multi-project workflows

**By Favorite Status**:
- Show only favorited sessions
- Quick access to important work

**By Date Range**:
- From/To date pickers
- Find sessions from specific time periods
- Examples: "last week", "last month", "Q1 2026"

### Search Results

Results display:
- Session title
- Timestamp
- Matching excerpt (highlighted)
- Mode badge (standard/firefighter/boatmanmode)
- Tags and favorite status

Click any result to open that session.

---

## Diff Viewing & Review

### Diff View

**Access diff viewer**:
1. Click on a task with code changes
2. View diff in task detail modal
3. Or use standalone diff view for large changes

**Diff features**:
- Side-by-side or unified view
- Syntax highlighting
- Line numbers
- Addition/deletion highlighting
- Collapsible hunks

### Diff Summary Cards

Quick overview at the top of diff view:
- **Files changed**: Total count
- **Additions**: Green +X lines
- **Deletions**: Red -X lines
- **Modifications**: Yellow ~X lines
- **File tree**: Navigate between files

### Inline Comments

**Add comments to specific lines**:
1. Click comment icon (💬) on any diff line
2. Type your comment
3. Click "Send" or press `Cmd+Enter`

**Comment threads**:
- View all comments on a line
- Nested replies
- Delete your own comments
- Timestamp and author tracking

**Use cases**:
- Code review feedback
- Questions about changes
- Suggestions for improvements
- Documentation notes

### Batch Approval

**Select multiple files**:
1. Check boxes next to files in diff view
2. Batch approval bar appears at bottom
3. Shows "X of Y files selected"

**Batch actions**:
- **Approve Selected**: Accept all selected changes
- **Reject Selected**: Reject all selected changes
- **Clear**: Deselect all

**Benefits**:
- Faster code review process
- Approve obvious changes quickly
- Reject entire categories (e.g., all test files)

---

## Task Tracking

### Task List

**View tasks**:
- Click "Tasks" tab in session view
- Shows all tasks created during execution
- Color-coded status indicators

**Task statuses**:
- 🟡 **Pending**: Not started
- 🔵 **In Progress**: Currently executing
- 🟢 **Completed**: Successfully finished
- 🔴 **Failed**: Error or issue occurred

### Task Detail Modal

**View detailed information**:
1. Click any task card
2. Modal shows full metadata

**Metadata included**:
- **Diff**: Code changes made by agent
- **Plan**: Execution plan (for planning tasks)
- **Feedback**: Review feedback (for review tasks)
- **Issues**: Problems found (for review/validation)
- **Refactor Diff**: Changes from refactoring (for refactor tasks)

**Use cases**:
- Understand what agent did
- Review code quality
- Debug failures
- Learn from successful patterns

---

## BoatmanMode Integration

### Starting BoatmanMode Execution

1. Ensure Linear API key is configured (Settings → Firefighter)
2. Click "Boatman Mode" button (purple, top right)
3. Enter Linear ticket ID (e.g., `ENG-123`)
4. Click "Start Execution"

### Execution Phases

Watch real-time progress through:

1. **Planning**: Analyzes ticket and creates implementation plan
2. **Validation**: Pre-flight checks on plan
3. **Execution**: Generates code changes
4. **Testing**: Runs test suite
5. **Review**: Peer review with configurable Claude skill
6. **Refactoring**: Fixes issues from review (if needed)
7. **PR Creation**: Creates pull request and updates ticket

### Event Streaming

**Real-time updates**:
- Events appear as they happen
- No polling, true streaming
- Task list updates live

**Event types**:
- `agent_started`: New agent begins work
- `agent_completed`: Agent finishes (success/failure)
- `task_created`: New task added
- `task_updated`: Task status changed
- `progress`: General progress messages

**Event metadata**:
All events include rich context (diffs, feedback, plans, issues) that populates task detail modals.

### Benefits Over Manual Execution

| Feature | Manual | BoatmanMode |
|---------|--------|-------------|
| Planning | Manual | Automated |
| Implementation | You code | AI codes |
| Testing | Remember to run | Automatic |
| Review | Request peer review | Built-in |
| Refactoring | Manual fixes | Iterative AI |
| PR Creation | Manual | Automatic |
| Ticket Updates | Manual | Automatic |

---

## Firefighter Mode

### Incident Investigation

**Workflow**:
1. Click "Firefighter" button (flame icon)
2. View Linear triage queue
3. Select incident ticket
4. Click "Investigate"

**What happens**:
- Fetches error details from Bugsnag
- Queries relevant logs from Datadog
- Analyzes git history for recent changes
- Identifies code owners via git blame
- Correlates deployments with error spikes
- Generates comprehensive investigation report

### Investigation Report

**Includes**:
- Error summary and stack trace
- Affected users/requests
- Recent deployments
- Related code changes
- Suspected root cause
- Recommended fixes
- Links to Bugsnag/Datadog

### Auto-Fix (High/Urgent Priority)

For high-priority incidents:
1. Creates isolated git worktree
2. Implements potential fix
3. Runs test suite
4. Creates draft PR if tests pass
5. Updates Linear ticket with PR link

### Configuration

**Required**:
- Linear API key (Settings → Firefighter)
- Okta SSO for Bugsnag/Datadog access

**Optional**:
- Slack integration for alert mentions
- Custom triage queue labels

---

## Agent Logs

### Real-Time Visibility

**Access logs panel**:
- Click "Logs" tab in session view
- Auto-scrolls as new logs appear
- Shows all agent tool usage

**Log types**:
- 📖 **File reads**: Which files Claude accessed
- ✏️ **File edits**: What Claude modified
- 📝 **File writes**: New files Claude created
- 🔧 **Bash commands**: Commands Claude ran
- 🔍 **Searches**: Codebase searches performed
- 🧠 **Thinking**: Agent reasoning (if available)

### Use Cases

**Debugging**:
- See exactly what agent did
- Identify where it went wrong
- Understand decision-making process

**Learning**:
- Watch how Claude approaches problems
- Learn effective patterns
- Improve your prompts

**Trust**:
- Full transparency of AI actions
- No hidden operations
- Approve before destructive commands

---

## Settings & Configuration

### General Settings

**Authentication**:
- **Google Cloud OAuth**: For Vertex AI Claude access
- **API Key**: Direct Anthropic API key
- **Okta SSO**: For enterprise MCP integrations

**Model Selection**:
- Default model for new sessions
- Override per session if needed
- Available: Opus 4.6, Sonnet 4.5, Haiku 4, Claude 3.5 Sonnet

**Approval Mode**:
- **Auto-approve**: Agent can execute without confirmation
- **Manual approve**: Review each action before execution

### Firefighter Settings

**Linear API Key**:
- Personal API key from Linear settings
- Used for both Firefighter and BoatmanMode
- Required for ticket integration

**Triage Queue**:
- Configure queue labels
- Default: `firefighter`, `triage`

### MCP Servers

**Built-in servers**:
- GitHub (for repository operations)
- Datadog (for log queries)
- Bugsnag (for error tracking)
- Linear (for ticket management)
- Slack (for notifications)

**Add custom server**:
1. Click "Add Server" in settings
2. Enter server name and command
3. Configure authentication if needed
4. Save and restart

**Server status**:
- ✅ Connected and ready
- ⚠️ Authentication needed
- ❌ Failed to start

### Memory & Storage

**Session Storage**:
- Sessions automatically saved to `~/.boatman/sessions/`
- Persistent across app restarts
- Includes full conversation history and metadata

**Storage Management**:
- View total storage usage in Settings
- Clear old sessions to free space
- Export sessions for backup
- Import sessions from backups

**Session Cleanup**:
- Manually delete individual sessions (right-click → Delete)
- Bulk delete by date range (Settings → Storage)
- Auto-cleanup after configurable retention period (optional)

**Export/Import**:
- **Export**: Right-click session → Export as JSON
- **Import**: Settings → Import Sessions → Select JSON file
- Useful for sharing sessions with team or moving between machines

**Data Privacy**:
- All data stored locally on your machine
- No cloud sync (unless you configure it)
- Sessions contain prompts and responses but not API keys

### Appearance & Theming

**Display Preferences**:
- **Font Size**: Adjust code and chat font sizes independently
- **Line Height**: Configure spacing for better readability
- **Code Theme**: Choose syntax highlighting theme (light/dark variants)

**UI Customization**:
- **Compact Mode**: Reduce padding for more content on screen
- **Show/Hide Panels**: Toggle visibility of logs, tasks, and diff panels
- **Panel Layout**: Adjust panel sizes and positions

**Accessibility**:
- High contrast mode for better visibility
- Keyboard navigation throughout the app
- Screen reader support (experimental)

**Coming Soon**:
- Full dark mode support
- Custom color themes
- Panel layout presets

---

## Keyboard Shortcuts

### Global

| Shortcut | Action |
|----------|--------|
| `Cmd+K` / `Ctrl+K` | Open search modal |
| `Cmd+N` / `Ctrl+N` | New chat session |
| `Cmd+,` / `Ctrl+,` | Open settings |
| `Cmd+F` / `Ctrl+F` | Toggle favorites filter |
| `Esc` | Close modal/dialog |

### Chat View

| Shortcut | Action |
|----------|--------|
| `Cmd+Enter` / `Ctrl+Enter` | Send message |
| `Cmd+L` / `Ctrl+L` | Toggle logs panel |
| `Cmd+T` / `Ctrl+T` | Switch to tasks tab |

### Diff View

| Shortcut | Action |
|----------|--------|
| `Cmd+A` / `Ctrl+A` | Select all files |
| `Cmd+D` / `Ctrl+D` | Deselect all |
| `Cmd+Enter` / `Ctrl+Enter` | Approve selected |
| `Cmd+Backspace` / `Ctrl+Backspace` | Reject selected |

### Search Modal

| Shortcut | Action |
|----------|--------|
| `Enter` | Select first result |
| `↑` / `↓` | Navigate results |
| `Cmd+F` / `Ctrl+F` | Toggle filters |
| `Esc` | Close search |

---

## Tips & Best Practices

### Session Organization

✅ **DO**:
- Tag sessions immediately after creation
- Favorite important sessions
- Use descriptive session titles
- Clean up old sessions periodically

❌ **DON'T**:
- Let sessions pile up without tags
- Forget to favorite critical work
- Use generic titles like "Chat 1"

### Code Review

✅ **DO**:
- Review diffs in task detail modals
- Add comments for questions
- Use batch approval for obvious changes
- Check agent logs for context

❌ **DON'T**:
- Blindly approve all changes
- Skip reviewing test changes
- Ignore inline comments from team

### BoatmanMode

✅ **DO**:
- Use for well-defined tickets
- Check task metadata after execution
- Review PR before merging
- Learn from successful patterns

❌ **DON'T**:
- Use for vague requirements
- Skip reviewing generated code
- Merge without testing locally
- Ignore failed iterations

### Firefighter Mode

✅ **DO**:
- Investigate before attempting fixes
- Read full investigation report
- Correlate with recent deployments
- Update Linear ticket with findings

❌ **DON'T**:
- Skip investigation step
- Assume first guess is correct
- Forget to document findings
- Ignore related errors

---

## Troubleshooting

### Search Not Working

**Symptom**: Search returns no results

**Solutions**:
1. Clear filters (some may be too restrictive)
2. Check date range isn't too narrow
3. Try broader search terms
4. Verify sessions exist for that project

### BoatmanMode Execution Stuck

**Symptom**: No events appearing

**Solutions**:
1. Check Linear API key is configured
2. Verify `boatman` CLI is in PATH
3. Look at agent logs for errors
4. Check browser console for errors

### Diff View Not Loading

**Symptom**: Blank diff view

**Solutions**:
1. Refresh the session
2. Check task has diff metadata
3. Try reopening task detail modal
4. Check browser console for errors

### Agent Logs Empty

**Symptom**: No logs appearing

**Solutions**:
1. Ensure session is active
2. Check logs tab is selected
3. Scroll to bottom (auto-scroll enabled)
4. Verify agent is actually running

---

## Feature Request & Feedback

Have ideas for new features or improvements?

1. Open an issue on GitHub
2. Tag with `enhancement` label
3. Describe use case and expected behavior
4. Include screenshots if applicable

We're actively developing and love hearing from users!

---

**Last Updated**: February 2026
**Version**: Desktop v1.0.0+
