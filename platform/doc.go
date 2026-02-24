// Package platform provides organizational capabilities for the Boatman ecosystem.
//
// It adds shared memory, cost governance, policy enforcement, and analytics
// on top of the core harness, enabling teams to share learnings and enforce
// development standards across repositories.
//
// # Sub-packages
//
//   - storage: Interface-based persistence with SQLite backend
//   - eventbus: Embedded NATS server for pub/sub event distribution
//   - server: HTTP API with scope-based multi-tenancy and SSE streaming
//   - services/memory: Hierarchical pattern merging across org/team/repo scopes
//   - services/policy: Policy engine with most-restrictive-wins merge strategy
//   - services/cost: Budget tracking with alerts and governance hooks
//   - client: HTTP client for CLI integration
//   - dashboard: Embedded React web dashboard
//
// # Scope model
//
// All data is scoped by Org, Team, and Repo identifiers. Policies and memory
// merge hierarchically (org → team → repo), with more specific scopes
// overriding or restricting less specific ones.
//
// # Graceful degradation
//
// The CLI connects to the platform via a TryConnect pattern with a 3-second
// timeout. If the platform is unreachable or not configured, the CLI continues
// in standalone mode with full functionality. Platform features are additive.
package platform
