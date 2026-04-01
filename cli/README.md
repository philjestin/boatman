# BoatmanMode 🚣

An AI-powered development agent that automates ticket execution with peer review. BoatmanMode fetches tickets from Linear, generates code using Claude, reviews changes with a configurable Claude skill (default: `peer-review`), iterates until quality passes, and creates pull requests.

## Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                            BoatmanMode Orchestrator                          │
│                                                                               │
│  ┌─────────────┐    ┌─────────────────────────────────────────────────────┐ │
│  │   Linear    │───▶│                   Workflow Engine                    │ │
│  │  (tickets)  │    │                                                       │ │
│  └─────────────┘    │  1. Fetch ticket         5. Review (peer-review)     │ │
│                     │  2. Create worktree      6. Refactor loop            │ │
│  ┌─────────────┐    │  3. Plan (parallel)      7. Verify diff              │ │
│  │ Coordinator │◀──▶│  4. Validate & Execute   8. Create PR (gh)           │ │
│  │  (agents)   │    └─────────────────────────────────────────────────────┘ │
│  └─────────────┘                      │                                      │
│            ┌──────────────────────────┼──────────────────────────┐          │
│            ▼                          ▼                          ▼          │
│  ┌─────────────────┐       ┌─────────────────┐       ┌─────────────────┐   │
│  │  Preflight +    │       │ Test Runner +   │       │  Diff Verify +  │   │
│  │  Planner Agent  │       │ Review Agent    │       │  Refactor Agent │   │
│  │ ┌─────────────┐ │       │ ┌─────────────┐ │       │ ┌─────────────┐ │   │
│  │ │   Claude    │ │       │ │ peer-review │ │       │ │   Claude    │ │   │
│  │ │  (planning) │ │       │ │   + tests   │ │       │ │ (refactor)  │ │   │
│  │ └─────────────┘ │       │ └─────────────┘ │       │ └─────────────┘ │   │
│  └─────────────────┘       └─────────────────┘       └─────────────────┘   │
│                                                                              │
│  ┌──────────────────────────────────────────────────────────────────────┐  │
│  │                        Support Systems                                 │  │
│  │  📌 Context Pin  │  💾 Checkpoint  │  🧠 Memory  │  📝 Issue Tracker  │  │
│  └──────────────────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Key Features

### 🤖 AI-Powered Development
- Generates complete implementations from ticket descriptions
- Understands project conventions via Claude's context
- Creates appropriate tests alongside code

### 👀 Peer Review with Claude Skill
- Uses a configurable Claude skill for code review (default: `peer-review`)
- Specify a custom skill via `--review-skill` or config
- Automated pass/fail verdict with detailed feedback
- Falls back to built-in review if skill not found

### 🔄 Iterative Refinement
- Automatically refactors based on review feedback
- Fresh agent per iteration (clean context, no token bloat)
- Structured handoffs between agents (concise context)

### 📺 Live Activity Streaming
- Watch Claude work in real-time via tmux
- See every tool call: file reads, edits, bash commands
- Full visibility into AI decision-making
- `claude_stream` events forward raw Claude output for desktop app integration

### 🌲 Git Worktree Isolation
- Each ticket works in an isolated worktree
- No interference with your main working directory
- Commit and push changes at any time

---

## 🆕 New Features

### 📝 Multiple Input Modes (NEW)
Work from Linear tickets, inline prompts, or files - same 9-step workflow:

```bash
# Linear mode (existing)
boatman work ENG-123

# Prompt mode (new)
boatman work --prompt "Add authentication with JWT tokens"

# File mode (new)
boatman work --file ./tasks/authentication.md
```

**Features:**
- Auto-generates unique task IDs for prompt/file modes
- Extracts titles from markdown headers or first line
- Auto-generates safe git branch names
- Same quality workflow regardless of input source
- Override auto-generation with `--title` and `--branch-name` flags

See [TASK_MODES.md](TASK_MODES.md) for complete documentation.

### 📡 Event System for External Integration (NEW)
Real-time JSON event stream for desktop app integration:

```bash
# Events are automatically emitted to stdout
boatman work ENG-123 | grep '^{' | jq
```

**Event Types:**
- `agent_started` / `agent_completed` - Track each workflow step
- `progress` - General progress updates
- `claude_stream` - Raw Claude stream-json lines for full UI visibility
- `task_created` / `task_updated` - Task lifecycle events (reserved)

**Use Cases:**
- Desktop app integration ([boatmanapp](https://github.com/philjestin/boatmanapp))
- Real-time workflow monitoring
- Custom dashboards and reporting
- CI/CD pipeline integration

See [EVENTS.md](EVENTS.md) for complete event specification and integration examples.

---

### 📊 Backlog Triage Pipeline (NEW)

Analyze and classify entire backlogs for AI-readiness:

```bash
# Score and classify tickets
boatman triage --teams EMP --states backlog --limit 20

# With plan generation
boatman triage --teams EMP --states backlog --generate-plans --repo-path .

# Specific tickets
boatman triage --ticket-ids EMP-1234,EMP-5678

# Stream events for desktop integration
boatman triage --teams EMP --states backlog --emit-events
```

**Pipeline stages:**
1. **Fetch** — Pull tickets from Linear (by team, state, or ID)
2. **Ingest** — Normalize tickets, extract signals (domains, files, dependencies)
3. **Score** — Claude rates each ticket on 7 dimensions (clarity, codeLocality, patternMatch, validationStrength, dependencyRisk, productAmbiguity, blastRadius)
4. **Classify** — Deterministic decision tree: hard stops (payments, auth) → soft stops (feature flags) → threshold gates → category assignment
5. **Cluster** — Group related tickets by signal overlap, generate context documents
6. **Plan** (optional) — Claude explores the repo with Read/Grep/Glob tools and generates validated execution plans

**Categories:** `AI_DEFINITE` | `AI_LIKELY` | `HUMAN_REVIEW_REQUIRED` | `HUMAN_ONLY`

**Plan validation gates:** files exist, within repo areas, stop conditions non-empty, valid test runners

See [desktop/TRIAGE.md](../desktop/TRIAGE.md) for the full triage specification.

### 🔄 Draft PR Safety Checkpoint (NEW)

After execution completes (Step 5), a draft PR is created immediately as a safety checkpoint before test/review/refactor begins. If those later stages hang or fail, the work is preserved:

```bash
# Normal flow: draft PR created automatically at Step 5b
boatman work EMP-1234

# If execution fails at review/refactor, resume from that point
boatman work EMP-1234 --resume
```

The draft PR is updated with review results and marked ready when the pipeline completes successfully.

### 🚀 Pre-flight Validation Agent
Validates the execution plan before any code changes:
- Verifies all referenced files exist
- Checks for deprecated patterns
- Validates approach clarity
- Warns about potential issues early

### 🧪 Test Runner Agent
Automatically runs tests after code changes:
- Auto-detects test framework (Go, Jest, RSpec, pytest)
- Parses test output for pass/fail
- Extracts coverage metrics
- Reports failed test names

### 🔍 Diff Verification Agent
Ensures refactors actually address review issues:
- Analyzes old vs new diffs
- Matches changes to specific issues
- Calculates confidence scores
- Detects newly introduced problems

### 🤝 Parallel Agent Coordination
Multiple agents can work simultaneously without conflicts:
- Central coordinator manages agent communication
- Work claiming prevents duplicate effort
- File locking prevents race conditions
- Shared context for agent collaboration

### 📌 Context Pinning
Ensures consistency during multi-file changes:
- Pins file contents with checksums
- Tracks file dependencies
- Detects stale files during long operations
- Refreshes context when needed

### 📦 Dynamic Handoff Compression
Adapts context size to token budgets:
- 4 compression levels (light → extreme)
- Priority-based content preservation
- Smart extraction of signatures and bullet points
- Automatic truncation with markers

### 📄 Smart File Summarization
Handles large files intelligently:
- Extracts function/class signatures
- Preserves imports and exports
- Keeps key comments and TODOs
- Language-aware parsing (Go, Python, Ruby, JS/TS, Java, Rust)

### 📝 Issue Deduplication
Tracks issues across review iterations:
- Prevents re-reporting same issues
- Detects similar issues via text similarity
- Tracks persistent vs addressed issues
- Provides iteration statistics

### 💾 Git-Integrated Checkpoints
Saves progress using git commits for durability:
- **Git commits** at each checkpoint for durability
- **Rollback** using `git reset` to any previous state
- **Snapshot branches** for important milestones
- **History browsing** with full audit trail
- **Squash** checkpoint commits before PR creation
- Resume from last successful step after crashes

### 🧠 Agent Memory
Cross-session learning for improved performance:
- Learns successful patterns
- Remembers common issues and solutions
- Caches effective prompts
- Per-project memory storage

### 🛡️ Resilience & Reliability (NEW)
Production-ready error handling and recovery:
- **Retry logic** with exponential backoff for Linear API and Claude CLI
- **Health checks** verify `git`, `gh`, `claude`, `tmux` at startup
- **Graceful degradation** when optional dependencies unavailable
- **Context cancellation** properly propagates to long-running operations

### 📊 Observability (NEW)
Structured logging and metrics for debugging:
- **Structured logging** via `log/slog` with levels (DEBUG, INFO, WARN, ERROR)
- **Dropped message tracking** when coordinator channels overflow
- **Debug mode** with `BOATMAN_DEBUG=1` for verbose output

### ⚙️ Configuration (NEW)
Externalized settings for all components:
- Coordinator buffer sizes
- Retry attempts and delays
- Claude CLI settings
- Token budgets for handoffs

### 🧪 E2E Test Environment (NEW)
Complete test harness for integration testing:
- Mock Linear GraphQL server
- Mock Claude CLI with canned responses
- Mock GitHub CLI for PR creation
- Fixture-based test scenarios

### 💰 Cost Optimization (NEW)
Intelligent model selection and prompt caching to reduce API costs by 50-90%:

#### Prompt Caching
- **Automatic caching** of system prompts (project rules, agent instructions)
- **50-90% cost reduction** on refactor iterations (cached context reused)
- **Faster response times** from cache hits
- **Enabled by default** with `enable_prompt_caching: true`

#### Multi-Model Strategy
- **Smart model selection** per agent type for optimal cost/performance
- **Configurable per agent** via `claude.models` in config
- **Example savings**: Preflight + test parsing with Haiku saves ~$0.50 per ticket

### 🧠 Model Configuration

Each agent in the workflow can use a different Claude model. Set the model for each agent type in your `.boatman.yaml` under `claude.models`:

```yaml
claude:
  models:
    planner: claude-opus-4-6         # Model for planning & codebase analysis
    executor: claude-opus-4-6        # Model for code generation
    reviewer: claude-opus-4-6        # Model for quality review
    refactor: claude-opus-4-6        # Model for fixing review issues
    preflight: claude-haiku-4        # Model for fast validation
    test_runner: claude-haiku-4      # Model for test output parsing
    scorer: claude-sonnet-4-5        # Model for triage rubric scoring
    triage_planner: claude-opus-4-6  # Model for triage plan generation
```

#### Available Models

| Model | ID | Best For |
|-------|----|----------|
| **Claude Opus 4.6** | `claude-opus-4-6` | Highest quality — complex planning, nuanced code generation, thorough review |
| **Claude Sonnet 4.5** | `claude-sonnet-4-5` | Good balance of quality and cost |
| **Claude Haiku 4** | `claude-haiku-4` | Fast, cheap — simple validation and parsing tasks |

#### Example Configurations

**Maximum quality** (use Opus for all complex agents):
```yaml
claude:
  models:
    planner: claude-opus-4-6
    executor: claude-opus-4-6
    reviewer: claude-opus-4-6
    refactor: claude-opus-4-6
    preflight: claude-haiku-4
    test_runner: claude-haiku-4
```

**Balanced** (Sonnet for most tasks, Haiku for simple ones):
```yaml
claude:
  models:
    planner: claude-sonnet-4-5
    executor: claude-sonnet-4-5
    reviewer: claude-sonnet-4-5
    refactor: claude-sonnet-4-5
    preflight: claude-haiku-4
    test_runner: claude-haiku-4
```

**Cost-optimized** (Haiku everywhere possible):
```yaml
claude:
  models:
    planner: claude-sonnet-4-5
    executor: claude-sonnet-4-5
    reviewer: claude-haiku-4
    refactor: claude-haiku-4
    preflight: claude-haiku-4
    test_runner: claude-haiku-4
```

If a model field is left empty or omitted, the Claude CLI's default model is used.

---

## Prerequisites

| Tool | Purpose | How to Authenticate |
|------|---------|---------------------|
| `claude` | AI code generation & review | `gcloud auth login` (Vertex AI) |
| `gh` | Pull request creation | `gh auth login` |
| `git` | Version control | SSH keys or credential helper |
| `tmux` | Agent session management | (no auth needed) |

### Claude CLI Setup (Vertex AI)

```bash
# Authenticate with Google Cloud
gcloud auth login
gcloud auth application-default login

# Set environment (or use an alias)
export CLAUDE_CODE_USE_VERTEX=1
export CLOUD_ML_REGION=us-east5
export ANTHROPIC_VERTEX_PROJECT_ID=your-project-id
```

## Installation

### Option 1: Download Pre-built Binary (Recommended)

Download the latest release for your platform from the [releases page](https://github.com/philjestin/boatmanmode/releases), or use the install script:

> **Note:** Releases are created automatically when code is pushed to `main`. See [RELEASING.md](RELEASING.md) for details on automatic versioning.

```bash
# macOS/Linux one-liner
curl -fsSL https://raw.githubusercontent.com/philjestin/boatmanmode/main/install.sh | bash

# Or download specific version
curl -fsSL https://raw.githubusercontent.com/philjestin/boatmanmode/main/install.sh | bash -s -- --version v1.0.0

# Or install to custom directory
curl -fsSL https://raw.githubusercontent.com/philjestin/boatmanmode/main/install.sh | bash -s -- --dir ~/bin
```

Supported platforms:
- **macOS**: Intel (amd64) and Apple Silicon (arm64)
- **Linux**: x86_64 and ARM64
- **Windows**: x86_64

### Option 2: Install with Go

```bash
go install github.com/philjestin/boatmanmode/cmd/boatman@latest
```

### Option 3: Build from Source

```bash
git clone https://github.com/philjestin/boatmanmode
cd boatmanmode
go build -o boatman ./cmd/boatman

# Optional: Add to PATH
sudo mv boatman /usr/local/bin/
```

### Verify Installation

```bash
boatman version
```

## Configuration

### Required: Linear API Key

```bash
export LINEAR_API_KEY=lin_api_xxxxx
```

### Optional: Config File

Create `~/.boatman.yaml`:

```yaml
linear_key: lin_api_xxxxx
max_iterations: 3
base_branch: main
review_skill: peer-review  # Claude skill/agent for code review

# Feature toggles
enable_preflight: true
enable_tests: true
enable_diff_verify: true
enable_memory: true
checkpoint_dir: ~/.boatman/checkpoints
memory_dir: ~/.boatman/memory

# Coordinator settings (advanced)
coordinator:
  message_buffer_size: 1000      # Main message channel buffer
  subscriber_buffer_size: 100    # Per-agent channel buffer

# Retry settings
retry:
  max_attempts: 3
  initial_delay: 500ms
  max_delay: 30s

# Claude CLI settings
claude:
  command: claude                     # Claude CLI command
  use_tmux: false                    # Use tmux for large prompts
  large_prompt_threshold: 100000     # Character count for tmux
  timeout: 0                         # 0 = no timeout
  enable_prompt_caching: true        # Enable prompt caching (reduces costs 50-90%)

  # Multi-model strategy: Use different models per agent type
  # Available models: claude-opus-4-6, claude-sonnet-4-5, claude-haiku-4
  # See "Model Configuration" section below for details and examples
  models:
    planner: claude-sonnet-4-5       # Complex planning & codebase analysis
    executor: claude-sonnet-4-5      # Code generation
    reviewer: claude-sonnet-4-5      # Quality review
    refactor: claude-sonnet-4-5      # Fixing review issues
    preflight: claude-haiku-4        # Fast validation (90% cheaper)
    test_runner: claude-haiku-4      # Simple test output parsing (90% cheaper)

# Token budgets for handoffs
token_budget:
  context: 8000
  plan: 2000
  review: 4000
```

## Usage

### Execute a Task

BoatmanMode supports three input modes:

```bash
cd /path/to/your/project

# 1. Linear ticket (default)
boatman work ENG-123

# 2. Inline prompt
boatman work --prompt "Add a health check endpoint at /health"

# 3. File-based prompt
boatman work --file ./tasks/authentication.md

# With custom title and branch
boatman work --prompt "Add auth" --title "Authentication" --branch-name "feature/auth"
```

### Watch Claude Work (Live Streaming)

```bash
# In another terminal
boatman watch

# Or attach to specific session
tmux attach -t boatman-executor
tmux attach -t boatman-reviewer-1
```

**What you'll see:**
```
🤖 Claude is working (with file write permissions)...
📝 Activity will stream below:

🔧 Running: ls -la packs/
📖 Reading: packs/annotations/app/graphql/consumer/types/...
✏️  Editing: packs/annotations/app/graphql/consumer/mutations/...
📝 Writing: packs/annotations/spec/graphql/consumer/...
🔍 Searching files...

📊 Task completed!
```

**tmux controls:**
- `Ctrl+B` then `D` - Detach
- `Ctrl+B` then arrow keys - Switch panes

### Manage Sessions

```bash
boatman sessions list       # List active sessions
boatman sessions kill       # Kill all boatman sessions
boatman sessions kill -f    # Also kill orphaned claude processes
boatman sessions cleanup    # Clean up idle sessions
```

### Manage Worktrees

```bash
boatman worktree list                    # List all worktrees
boatman worktree commit                  # Commit changes (WIP)
boatman worktree commit wt-name "msg"    # Commit with message
boatman worktree push                    # Push branch to origin
boatman worktree clean                   # Remove all worktrees
```

### View Changes Manually

```bash
# Go to worktree
cd .worktrees/philmiddleton-eng-123-feature

# See changes
git status
git diff

# Commit and push
git add -A
git commit -m "feat: implement feature"
git push -u origin HEAD

# Or checkout in main repo
cd /path/to/project
git checkout philmiddleton/eng-123-feature
```

### Command Options

```bash
boatman work ENG-123 --max-iterations 5        # More refactor attempts
boatman work ENG-123 --base-branch develop     # Different base branch
boatman work ENG-123 --dry-run                 # Preview without changes
boatman work ENG-123 --review-skill my-review  # Use custom review skill
```

## Workflow Details

### Enhanced Agent Pipeline

The workflow now uses **coordinated parallel agents** with intelligent handoffs:

```
┌─────────────────────────────────────────────────────────────┐
│  Step 1: PLANNER AGENT (tmux: boatman-planner)              │
│  🧠 Analyzes ticket → Explores codebase → Creates plan      │
│     Output: Summary, approach, relevant files, patterns     │
├─────────────────────────────────────────────────────────────┤
│  Step 2: PREFLIGHT VALIDATION                               │
│  ✅ Validates plan → Checks files exist → Warns of issues   │
│     Output: Validation result, warnings, suggestions        │
├─────────────────────────────────────────────────────────────┤
│              ↓ Compressed Handoff (token-aware) ↓           │
├─────────────────────────────────────────────────────────────┤
│  Step 3: EXECUTOR AGENT (tmux: boatman-executor)            │
│  🤖 Receives plan → Reads key files → Implements solution   │
│     Output: Modified files in worktree                      │
├─────────────────────────────────────────────────────────────┤
│  Step 4: TEST RUNNER                                        │
│  🧪 Detects framework → Runs tests → Reports results        │
│     Output: Pass/fail, coverage, failed test names          │
├─────────────────────────────────────────────────────────────┤
│              ↓ Git Diff + Test Results ↓                    │
├─────────────────────────────────────────────────────────────┤
│  Step 5: REVIEWER AGENT (tmux: boatman-reviewer-N)          │
│  👀 Reviews diff → Checks patterns → Pass/Fail verdict      │
│     Output: Score, issues (deduplicated), guidance          │
├─────────────────────────────────────────────────────────────┤
│              ↓ If Failed (with issue deduplication) ↓       │
├─────────────────────────────────────────────────────────────┤
│  Step 6: REFACTOR AGENT (tmux: boatman-refactor-N)          │
│  🔧 Receives feedback → Fixes issues → Updates files        │
├─────────────────────────────────────────────────────────────┤
│  Step 7: DIFF VERIFICATION                                  │
│  🔍 Compares diffs → Verifies issues addressed              │
│     Output: Confidence score, addressed/unaddressed issues  │
└─────────────────────────────────────────────────────────────┘
         💾 Checkpoint saved at each step
         🧠 Patterns learned on success
```

### Agent Coordination

The coordinator manages parallel agent execution:

```go
// Agents can claim work to prevent conflicts
coord.ClaimWork("executor", &WorkClaim{
    WorkID: "implement-feature",
    Files:  []string{"pkg/feature.go"},
})

// File locking prevents race conditions
coord.LockFiles("executor", []string{"pkg/feature.go"})

// Shared context for collaboration
coord.SetContext("plan", planJSON)
result, _ := coord.GetContext("plan")
```

### Structured Handoffs

Agents receive concise, focused context with dynamic compression:

| Handoff Type | Content | Token Budget |
|--------------|---------|--------------|
| Plan → Executor | Summary, approach, files | ~4000 tokens |
| Executor → Reviewer | Requirements, diff, test results | ~3000 tokens |
| Reviewer → Refactor | Issues (deduplicated), guidance | ~2000 tokens |

### Git-Integrated Checkpoint System

Progress is saved as git commits for durability and rollback:

```bash
# Each step creates a checkpoint commit
# Format: [checkpoint] ENG-123: complete execution (step: execution, iter: 1)

# Resume an interrupted workflow
boatman work ENG-123 --resume

# View checkpoint history
git log --oneline --grep "\[checkpoint\]"

# Rollback to a previous checkpoint
git reset --hard HEAD~2  # Go back 2 checkpoints

# Create a snapshot branch before risky operation
boatman checkpoint snapshot "before-refactor"

# Squash checkpoint commits before PR
boatman checkpoint squash "feat: implement feature ENG-123"
```

**Checkpoint commits include:**
- Ticket ID and step name
- Iteration number
- Serialized agent state in `.boatman-state.json`
- All file changes up to that point

**Rollback scenarios:**
```bash
# Undo last refactor attempt
git reset --hard HEAD~1

# Go back to before review started
boatman checkpoint rollback --step execution

# Restore from snapshot branch
git checkout checkpoint/ENG-123/before-review -- .
```

### Agent Memory

Cross-session learning improves over time:

```bash
# Memory is stored in ~/.boatman/memory/
# Per-project memory for patterns and issues

# Memory includes:
# - Successful code patterns
# - Common review issues
# - Effective prompts
# - Project preferences
```

## Using as a Go Library

BoatmanMode can be used as a library in your own Go applications:

```bash
go get github.com/philjestin/boatmanmode@latest
```

**Quick Example:**

```go
package main

import (
    "context"
    "github.com/philjestin/boatmanmode"
)

func main() {
    cfg := &boatmanmode.Config{
        LinearKey:     "your-api-key",
        BaseBranch:    "main",
        MaxIterations: 3,
        EnableTools:   true,
    }

    a, _ := boatmanmode.NewAgent(cfg)
    t, _ := boatmanmode.NewPromptTask("Add health check endpoint", "", "")
    result, _ := a.Work(context.Background(), t)

    if result.PRCreated {
        println("PR created:", result.PRURL)
    }
}
```

See [LIBRARY_USAGE.md](LIBRARY_USAGE.md) for complete API documentation and examples.

## Project Structure

```
boatmanmode/
├── cmd/boatman/main.go       # Entry point
├── internal/
│   ├── agent/                # Workflow orchestration (refactored into step methods)
│   ├── checkpoint/           # Progress saving/resume
│   ├── claude/               # Claude CLI wrapper (with retry + context cancellation)
│   ├── cli/                  # Cobra commands
│   ├── config/               # Configuration (expanded with nested configs)
│   ├── contextpin/           # File dependency tracking
│   ├── coordinator/          # Parallel agent coordination (thread-safe, observable)
│   ├── diffverify/           # Diff verification agent
│   ├── events/               # Event emission (agent_started, claude_stream, etc.)
│   ├── executor/             # Code generation (with EventForwarder for claude_stream)
│   ├── filesummary/          # Smart file summarization
│   ├── github/               # PR creation (gh CLI)
│   ├── handoff/              # Agent context passing + compression
│   ├── healthcheck/          # External dependency verification (NEW)
│   ├── issuetracker/         # Issue deduplication
│   ├── linear/               # Linear API client (with retry logic)
│   ├── logger/               # Structured logging via log/slog (NEW)
│   ├── memory/               # Cross-session learning
│   ├── planner/              # Plan generation
│   ├── preflight/            # Pre-execution validation
│   ├── retry/                # Exponential backoff retry logic (NEW)
│   ├── scottbott/            # Peer review
│   ├── testenv/              # E2E test environment with mocks (NEW)
│   ├── testrunner/           # Test execution
│   ├── tmux/                 # Session management
│   └── worktree/             # Git worktree management
└── README.md
```

## Environment Variables

| Variable | Description | Required |
|----------|-------------|----------|
| `LINEAR_API_KEY` | Linear API key | Yes |
| `CLAUDE_CODE_USE_VERTEX` | Set to `1` for Vertex AI | If using Vertex |
| `CLOUD_ML_REGION` | Vertex AI region | If using Vertex |
| `ANTHROPIC_VERTEX_PROJECT_ID` | GCP project ID | If using Vertex |
| `BOATMAN_DEBUG` | Set to `1` for debug output (structured logs) | No |
| `BOATMAN_CHECKPOINT_DIR` | Custom checkpoint directory | No |
| `BOATMAN_MEMORY_DIR` | Custom memory directory | No |
| `LINEAR_API_URL` | Override Linear API URL (for testing) | No |

## Troubleshooting

### "No files were changed in the worktree"

Claude ran but didn't modify any files. Possible causes:
- Ticket too vague - add more specific requirements
- Implementation already exists - Claude may just be analyzing
- Run `boatman watch` to see what Claude was doing

### Claude seems stuck

Check if Claude is actually working:
```bash
boatman watch  # See live activity
```

If truly stuck, kill and restart:
```bash
boatman sessions kill --force
boatman work ENG-123
```

### Session not found

```bash
boatman sessions kill  # Kill stuck sessions
boatman sessions list  # Verify clean state
```

### Want to see changes before PR

```bash
boatman worktree list                    # Find the worktree
cd .worktrees/<name>                     # Go there
git diff                                 # See changes
boatman worktree commit                  # Commit them
```

### Resume interrupted workflow

```bash
boatman work ENG-123 --resume  # Resume from checkpoint
```

### Timeout waiting for Claude

Large codebases take longer. The default timeout is 30 minutes. If Claude is actively working (visible in `boatman watch`), just wait. If stuck, use `boatman sessions kill --force`.

### Retry exhausted for API calls

If you see "failed after N attempts", the Linear API or Claude CLI is having issues:
```bash
# Check if services are accessible
curl -I https://api.linear.app/graphql
claude --version

# Increase retry attempts in config
# ~/.boatman.yaml
retry:
  max_attempts: 5
  initial_delay: 2s
```

### Dropped messages warning

If you see "coordinator message channel full, message dropped":
- This indicates high message volume between agents
- Increase buffer sizes in config:
```yaml
coordinator:
  message_buffer_size: 2000
  subscriber_buffer_size: 200
```

### Health check failures

If startup fails with "missing required dependencies":
```bash
# Verify all tools are installed and in PATH
which git gh claude tmux

# Check specific tool versions
git --version
gh --version
claude --version
```

### Debug mode

For detailed logging, enable debug mode:
```bash
export BOATMAN_DEBUG=1
boatman work ENG-123
```

This outputs structured logs showing:
- Retry attempts and delays
- Dropped messages
- Context cancellation
- Coordinator state changes

## Running Tests

```bash
# Run all unit tests
go test ./...

# Run with verbose output
go test -v ./...

# Run specific package tests
go test -v ./internal/coordinator
go test -v ./internal/checkpoint
go test -v ./internal/retry

# Run with coverage
go test -cover ./...

# Run E2E tests (includes mock servers)
go test ./internal/testenv/... -tags=e2e

# Run all tests including E2E
go test ./... -tags=e2e -v
```

### Test Packages

| Package | Tests | Description |
|---------|-------|-------------|
| `coordinator` | 17 | Work claiming, file locking, atomic ops, cleanup |
| `retry` | 14 | Exponential backoff, jitter, permanent errors |
| `healthcheck` | 12 | Dependency checks, timeouts, formatting |
| `logger` | 12 | Level filtering, JSON output, context |
| `config` | 13 | Defaults, custom values, nested configs |
| `testenv` | 18 | Mock servers, fixtures, e2e workflows |
| `agent` | 13 | Integration tests, parallel agents |

### E2E Test Environment

The `testenv` package provides a complete mock environment:

```go
func TestMyWorkflow(t *testing.T) {
    env := testenv.New(t).Setup()
    defer env.Cleanup()

    // Set custom Linear ticket
    env.SetLinearTicket("ENG-123", testenv.DefaultTicket())

    // Set Claude response
    env.SetClaudeResponse("I'll implement this feature...")

    // Run commands with mock environment
    output, err := env.RunInRepo(ctx, "go", "test", "./...")
}
```

## Code Quality

### Recent Improvements

The codebase has been hardened with the following improvements:

| Category | Changes |
|----------|---------|
| **Thread Safety** | Coordinator `running` flag uses `atomic.Bool`; no data races |
| **Error Handling** | Removed silent error swallowing (e.g., `os.Chdir` errors) |
| **Memory Management** | Coordinator `Stop()` clears all maps to prevent leaks |
| **Observability** | Dropped messages logged with `slog.Warn`; metrics tracked |
| **Resilience** | Exponential backoff retry for Linear API and Claude CLI |
| **Cancellation** | Claude streaming respects context cancellation |
| **Configuration** | All hardcoded values moved to config structs |
| **Testability** | `agent.Work()` refactored into 11 focused step methods |

### Architecture Decisions

- **No `os.Chdir`**: Commands use `cmd.Dir` instead of changing process state
- **Structured logging**: `log/slog` for consistent, parseable output
- **Atomic operations**: Thread-safe coordinator without excessive locking
- **Graceful cleanup**: Resources released in reverse order on shutdown

## License

MIT

---

*Built with 🚣 by the philjestin team*
