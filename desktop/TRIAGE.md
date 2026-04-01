# Triage Mode

Triage Mode analyzes a Linear backlog and classifies tickets by AI-readiness using a multi-stage pipeline. It scores each ticket on a 7-dimension rubric, applies deterministic classification gates, clusters related tickets, and optionally generates validated execution plans.

## Pipeline Stages

```
Linear Backlog
  |
  v
[Stage 0: Fetch] ── Pull tickets from Linear (teams, states, limits)
  |
  v
[Stage 1: Ingest] ── Normalize tickets, extract signals (domains, files, deps)
  |
  v
[Stage 2a: Score] ── Claude scores each ticket on 7-dimension rubric (0-5)
  |
  v
[Stage 2b: Classify] ── Deterministic decision tree assigns category
  |
  v
[Stage 3: Cluster] ── Group related tickets by signal overlap
  |
  v
[Stage 4: Plan] ── (Optional) Claude explores repo, generates execution plans
  |
  v
[Output] ── Classifications, clusters, context docs, plans
```

## Scoring Rubric (Stage 2a)

Each ticket is scored by Claude on 7 dimensions (0-5 scale):

**Positive dimensions** (higher = better for AI):
| Dimension | What it measures |
|-----------|-----------------|
| `clarity` | Are requirements explicit and testable? |
| `codeLocality` | Is the change confined to one module? |
| `patternMatch` | Has this type of problem been solved before in the codebase? |
| `validationStrength` | Can success be proven with automated tests? |

**Negative dimensions** (higher = worse for AI):
| Dimension | What it measures |
|-----------|-----------------|
| `dependencyRisk` | Unknown infra, external APIs, cross-team sequencing |
| `productAmbiguity` | UX judgment calls, stakeholder interpretation needed |
| `blastRadius` | How bad if the implementation is wrong? |

## Classification (Stage 2b)

Classification is fully deterministic (no LLM). It runs a decision tree with hard stops, soft stops, and threshold gates.

### Categories

| Category | Meaning |
|----------|---------|
| `AI_DEFINITE` | Safe for fully autonomous execution |
| `AI_LIKELY` | Probably safe, but review the plan first |
| `HUMAN_REVIEW_REQUIRED` | Needs human judgment before AI can proceed |
| `HUMAN_ONLY` | Must be done by a human |

### Hard Stops -> `HUMAN_ONLY`

Tickets mentioning: payments, billing, authentication, authorization, database migrations, public API changes, incidents (sev1/sev2), legal/compliance, multi-repo coordination.

### Soft Stops -> `HUMAN_REVIEW_REQUIRED`

Tickets mentioning: feature flags, staged rollouts, deploy coordination, release trains.

### Gate Thresholds -> `AI_DEFINITE` requires all:

- `clarity` >= 2
- `blastRadius` < 3
- `productAmbiguity` < 3
- `dependencyRisk` < 3

If any gate fails, the ticket falls to `AI_LIKELY` or `HUMAN_REVIEW_REQUIRED` depending on severity.

## Clustering (Stage 3)

Tickets are grouped by signal overlap using greedy single-linkage clustering:

- Shared domains: 0.5 points each
- Shared file mentions: 0.5 points each
- Shared dependencies: 1.0 points each
- Merge threshold: 2.0 minimum overlap

Each cluster produces a **Context Document** containing:
- `repoAreas` — directories the cluster touches
- `knownPatterns` — existing patterns in those areas
- `validationPlan` — how to test changes
- `risks` — what could go wrong
- `costCeiling` — token/time budget per ticket

## Plan Generation (Stage 4, Optional)

When `--generate-plans` is passed, Claude explores the repo with Read/Grep/Glob tools for each `AI_DEFINITE` and `AI_LIKELY` ticket and produces:

```json
{
  "ticketId": "EMP-1234",
  "approach": "Add new GraphQL mutation with...",
  "candidateFiles": ["packs/employer/app/graphql/mutations/..."],
  "newFiles": [],
  "deletedFiles": [],
  "validation": ["bundle exec rspec packs/employer/spec/..."],
  "rollback": "Revert the migration and remove the mutation",
  "stopConditions": ["If the table doesn't exist, abort"],
  "uncertainties": ["Unclear if the index is needed for this query pattern"]
}
```

### Plan Validation (4 Gates)

Each plan is validated before it can be executed:

| Gate | Check | Pass condition |
|------|-------|----------------|
| `files_exist` | Do candidateFiles exist on disk? | >50% exist |
| `within_repo_areas` | Are files within the cluster's repoAreas? | <50% out of scope |
| `stop_conditions` | Are stop conditions present and meaningful? | Non-empty |
| `validation_commands` | Do test commands use known runners? | Known runners (rspec, jest, etc.) |

## CLI Usage

```bash
# Basic triage
boatman triage --teams EMP --states backlog --limit 20

# With plan generation
boatman triage --teams EMP --states backlog --generate-plans --repo-path .

# Specific tickets
boatman triage --ticket-ids EMP-1234,EMP-5678

# Full options
boatman triage \
  --teams EMP,EDU \
  --states backlog,unstarted \
  --limit 50 \
  --concurrency 5 \
  --generate-plans \
  --repo-path /path/to/repo \
  --emit-events \
  --output-dir .boatman-triage
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--teams` | | Team keys to fetch (comma-separated) |
| `--states` | | Workflow states to filter |
| `--limit` | 50 | Maximum tickets to process |
| `--ticket-ids` | | Specific ticket IDs (skips team/state filters) |
| `--concurrency` | 3 | Parallel Claude scoring calls |
| `--generate-plans` | false | Run Stage 4 plan generation |
| `--repo-path` | `.` | Repository path for plan generation |
| `--emit-events` | false | Emit JSON events to stdout (used by desktop) |
| `--output-dir` | `.boatman-triage` | Decision log output directory |
| `--post-comments` | false | Write classification comments to Linear |
| `--dry-run` | false | Run without side effects |

## Desktop UI

### Starting a Triage

1. Click the triage icon in the sidebar
2. Configure: teams, states, ticket limit
3. Optionally check "Generate Plans"
4. Click "Start Triage"

### Viewing Results

The triage view shows three tabs:

**Results Tab** — Sortable table of all tickets with:
- Ticket ID, title, category badge
- Rubric scores (clarity, blast radius, etc.)
- Click to expand full classification detail (gate results, hard stops, signals)

**Clusters Tab** — Grouped view showing:
- Cluster rationale and ticket membership
- Repo areas touched
- Context doc: validation plan, risks, cost ceiling
- "Execute All" button for cluster-wide execution

**Plans Tab** (when plans are generated) — Per-ticket plan details:
- Approach, candidate files (validated/missing/out-of-scope)
- Validation commands, stop conditions, uncertainties
- Gate pass/fail results
- "Execute" button (only if all gates pass)

### Executing from Triage

Click "Execute" on a ticket or plan to create a new BoatmanMode session with the triage plan pre-loaded. The plan skips the planning step and goes straight to execution with the validated candidate files.

## Event Flow

```
CLI (boatman triage --emit-events)
  |  Emits JSON events to stdout:
  |  triage_started, triage_ticket_scored, triage_complete, etc.
  v
Desktop Integration (triage/integration.go)
  |  Parses JSON, emits Wails events
  v
Frontend (TriageView.tsx)
  |  Subscribes to triage:event
  |  Shows live progress bar during scoring
  |  Renders results when complete
  v
User clicks Execute
  |
  v
App.ExecuteTriageTicket()
  |  Creates BoatmanMode session with plan pre-loaded
  v
StreamBoatmanModeExecution() with --plan-file flag
```

## Decision Log

Triage writes an audit log to `--output-dir` (default `.boatman-triage/`):

- `log.jsonl` — One entry per ticket with classification decision, scores, gate results
- `context_<clusterId>.json` — Full context document per cluster

Each log entry includes: ticketID, stage, verdict, agent, rationale, timestamp, tokens used, cost.

## Architecture

```
cli/internal/triage/
  ├── pipeline.go      # Pipeline orchestrator (fetch → score → classify → cluster)
  ├── ingest.go        # Stage 1: Normalize tickets, extract signals
  ├── scorer.go        # Stage 2a: Claude scoring with 7-dimension rubric
  ├── classifier.go    # Stage 2b: Deterministic decision tree
  ├── cluster.go       # Stage 3: Signal-overlap clustering + context docs
  ├── decisionlog.go   # Audit log persistence
  ├── events.go        # Event emission helpers
  └── types.go         # All triage types (NormalizedTicket, Classification, etc.)

cli/internal/plan/
  ├── generator.go     # Stage 4: Claude with tools generates plans
  ├── validator.go     # 4-gate plan validation
  ├── events.go        # Plan event emission helpers
  └── types.go         # TicketPlan, PlanValidation, PlanResult, etc.

desktop/triage/
  └── integration.go   # Subprocess wrapper, event forwarding

desktop/frontend/src/components/triage/
  ├── TriageView.tsx         # Main orchestrator (progress + results tabs)
  ├── TriageResultsTable.tsx # Sortable ticket table
  ├── TriageTicketDetail.tsx # Full classification detail sidebar
  ├── TriageClusterView.tsx  # Cluster cards with context docs
  ├── TriagePlanView.tsx     # Plan results list
  ├── TriagePlanDetail.tsx   # Plan detail with gate validation
  ├── TriageStatsBar.tsx     # Summary stats (counts, tokens, cost)
  └── TriageBadge.tsx        # Category badge component
```
