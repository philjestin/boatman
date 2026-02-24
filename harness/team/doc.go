// Package team provides composable agent teams for the Boatman pipeline.
//
// A Team groups multiple Agents together with a Router (to select which agents
// handle a given task), a Strategy (sequential or parallel execution), and an
// Aggregator (to combine results). Teams implement the Handler interface and
// can be nested via AsAgent, enabling hierarchical delegation patterns.
//
// The top-level pipeline orchestration (1→9 fixed steps) remains deterministic.
// Agent teams operate within stages — e.g., the planning stage could route to
// a codebase analyzer and risk assessor in parallel, or the execution stage
// could delegate to frontend/backend specialists.
//
// Adapters (DeveloperTeam, PlannerTeam, TesterTeam, ReviewerTeam) allow any
// Team to satisfy the runner role interfaces so it can be plugged into existing
// pipeline stages without modification.
package team
