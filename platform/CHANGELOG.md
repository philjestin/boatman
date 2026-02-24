# Changelog

All notable changes to the Boatman Platform module will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2026-02-24

### Added

- Storage layer with SQLite backend (`modernc.org/sqlite`, no CGO required) implementing Run, Pattern, Preference, CommonIssue, UsageRecord, Budget, Policy, and Event stores.
- Event bus powered by embedded NATS server with pub/sub helpers and well-known subject constants for run lifecycle, step progress, cost recording, and budget alerts.
- HTTP API with scope-based multi-tenancy via `X-Boatman-Org`, `X-Boatman-Team`, and `X-Boatman-Repo` headers, including endpoints for runs, memory, costs, policies, events, and SSE streaming.
- Memory service with hierarchical pattern merging (org -> team -> repo) and automatic learning from successful runs.
- Policy engine with most-restrictive-wins merge strategy across organizational scopes, including runner config enforcement for iteration caps and test requirements.
- Cost governance with budget tracking (daily, monthly, per-run limits), alert thresholds, and real-time budget status checks via the Governor service.
- Web dashboard built with React, TypeScript, Vite, Tailwind CSS, and Recharts, embedded into the Go binary via `go:embed`. Includes pages for Runs, Memory, Costs, Policies, and Live Events.
- CLI integration with TryConnect pattern and graceful degradation -- the CLI operates in standalone mode when the platform is unreachable.
- Bridge adapters: `HooksAdapter` (maps runner hooks to bus events), `ObserverAdapter` (maps runner observer callbacks to bus events), and `LegacyBridge` (converts bus events to the `shared/events.Event` JSON format for backward compatibility with the desktop subprocess integration).
- Compliance test suite (`storage.RunStoreTests`) that any storage backend must pass, exercising all sub-stores (Run, Memory, Cost, Policy, Event) with standard CRUD operations and filter queries.
- Artifact storage with `ArtifactStore` interface and local filesystem backend for large binary blobs (diffs, logs, outputs) keyed by `{org}/{team}/{repo}/{run_id}/{artifact_type}`.
