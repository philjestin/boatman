// Package harness provides model-agnostic primitives for building AI agent
// harnesses. It includes review loops, issue deduplication, diff verification,
// checkpoint/resume, context compression, file dependency tracking, and more.
//
// The harness module has zero intra-monorepo dependencies (stdlib only) and
// can be used independently of the boatman CLI. Implementations can plug in
// any LLM backend (Claude, OpenAI, Gemini, etc.) via the provided interfaces.
//
// Core packages:
//
//   - runner: Composable pipeline orchestrator with execute-test-review-refactor loop
//   - runner (Observer): Run lifecycle event interface for logging and metrics
//   - runner (Guard): Mid-run policy enforcement interface with GuardState
//   - review: Canonical review types and the Reviewer interface
//   - checkpoint: Progress saving with git integration
//   - memory: Cross-session learning (patterns, preferences, issues)
//   - memory (MemoryProvider): Pluggable memory backend interface
//   - team: Composable agent teams with routing, parallel execution, and aggregation
//   - team (Adapters): DeveloperTeam, PlannerTeam, TesterTeam, ReviewerTeam
//   - cost: Token usage and cost tracking
//   - filesummary: Intelligent file summarization
//   - handoff: Structured context passing between pipeline stages
//   - issuetracker: Issue deduplication across review iterations
//   - diffverify: Diff verification against review issues
//   - contextpin: File dependency tracking and pinning
//   - testrunner: Test framework detection and execution
package harness
