# Boatman Platform

The platform module is the organizational layer for the Boatman ecosystem. It provides shared memory, cost governance, policy enforcement, and analytics on top of the core harness, enabling teams to share learnings and enforce development standards across repositories.

## Features

- **Multi-tenant scoping** -- All data is scoped by org, team, and repo via `X-Boatman-Org`, `X-Boatman-Team`, and `X-Boatman-Repo` headers.
- **Hierarchical policy** -- Policies merge from org to team to repo using a most-restrictive-wins strategy (lowest iteration cap, intersection of allowed models, `OR` for boolean requirements).
- **Shared memory** -- Patterns, preferences, and common issues are stored per scope and merged hierarchically so that repo-level data overrides team-level, which overrides org-level.
- **Cost tracking** -- Usage records capture per-step token counts and USD cost. Budgets set daily, monthly, and per-run limits with configurable alert thresholds.
- **Real-time events** -- An embedded NATS server powers a pub/sub event bus. Events are persisted to SQLite and streamed to the web dashboard via SSE.
- **Web dashboard** -- A React + TypeScript + Vite + Tailwind + Recharts single-page app embedded into the Go binary at build time. Pages for Runs, Memory, Costs, Policies, and Live Events.
- **CLI integration** -- The CLI connects to the platform with a TryConnect pattern and graceful degradation. When the platform is unreachable the CLI continues with local-only operation.

## Quick Start

### 1. Build

```bash
make build-platform
```

This produces the `platform/boatman-platform` binary.

### 2. Run

```bash
./platform/boatman-platform --port 8080 --data-dir ~/.boatman/platform
```

The server creates its data directory automatically. SQLite stores relational data in `platform.db` and NATS state lives in a `nats/` subdirectory.

### 3. Verify

```bash
curl http://localhost:8080/api/v1/health
```

A `200 OK` response with `{"status":"ok"}` confirms the server is ready.

### 4. Connect the CLI

Add the platform section to your `.boatman.yaml`:

```yaml
platform:
  server: "http://localhost:8080"
  org_id: "my-org"
  team_id: "my-team"
```

### 5. Open the Dashboard

Navigate to `http://localhost:8080` in your browser. The embedded SPA provides pages for Runs, Memory, Costs, Policies, and Live Events.

## Architecture

```
┌──────────────────────────────────────────────────────────┐
│                   Platform Server                        │
│                                                          │
│  ┌────────────┐  ┌────────────┐  ┌────────────────────┐  │
│  │  HTTP API   │  │  Dashboard │  │    Event Bus       │  │
│  │  (net/http) │  │  (React    │  │    (embedded NATS) │  │
│  │             │  │   SPA)     │  │                    │  │
│  └─────┬──────┘  └────────────┘  └────────┬───────────┘  │
│        │                                  │              │
│  ┌─────┴──────────────────────────────────┴───────────┐  │
│  │                   Services                         │  │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────────────┐  │  │
│  │  │  Memory   │  │  Policy  │  │  Cost Governor   │  │  │
│  │  │  Service  │  │  Engine  │  │                  │  │  │
│  │  └──────────┘  └──────────┘  └──────────────────┘  │  │
│  └────────────────────────┬───────────────────────────┘  │
│                           │                              │
│  ┌────────────────────────┴───────────────────────────┐  │
│  │                   Storage                          │  │
│  │  ┌──────────┐  ┌───────────┐  ┌─────────────────┐  │  │
│  │  │  SQLite   │  │ Artifacts │  │  Compliance     │  │  │
│  │  │  Backend  │  │ (local    │  │  Test Suite     │  │  │
│  │  │          │  │  fs / S3) │  │                 │  │  │
│  │  └──────────┘  └───────────┘  └─────────────────┘  │  │
│  └────────────────────────────────────────────────────┘  │
│                                                          │
│  ┌────────────────────────────────────────────────────┐  │
│  │                   Bridge Adapters                  │  │
│  │  HooksAdapter  ObserverAdapter  LegacyBridge       │  │
│  └────────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────┘
         ▲                                    ▲
         │  HTTP + scope headers              │  NATS pub/sub
         │                                    │
    ┌────┴────┐                          ┌────┴────┐
    │  CLI    │                          │ Desktop │
    │ client  │                          │   app   │
    └─────────┘                          └─────────┘
```

## Directory Structure

```
platform/
├── cmd/
│   └── boatman-platform/
│       └── main.go              # Server entry point
├── client/
│   ├── client.go                # HTTP client for the platform API
│   └── client_test.go
├── dashboard/
│   ├── embed.go                 # go:embed for compiled frontend assets
│   └── frontend/
│       ├── src/
│       │   ├── App.tsx
│       │   ├── main.tsx
│       │   ├── components/      # CostChart, Layout, PatternList, PolicyEditor, Sidebar
│       │   ├── hooks/           # useApi, useSSE
│       │   ├── pages/           # Costs, LiveEvents, Memory, Policies, Runs
│       │   └── types/
│       ├── index.html
│       ├── vite.config.ts
│       ├── tailwind.config.js
│       ├── tsconfig.json
│       └── package.json
├── eventbus/
│   ├── bus.go                   # Embedded NATS server + pub/sub helpers
│   ├── bus_test.go
│   ├── subjects.go              # Well-known event subject constants
│   └── bridge/
│       ├── bridge.go            # HooksAdapter, ObserverAdapter
│       ├── bridge_test.go
│       └── legacy.go            # LegacyBridge (shared/events compat)
├── server/
│   ├── server.go                # HTTP server lifecycle
│   ├── server_test.go
│   └── api/
│       ├── routes.go            # Route registration + dashboard handler
│       ├── middleware.go        # Scope extraction + request logging
│       ├── health.go
│       ├── runs.go
│       ├── memory.go
│       ├── costs.go
│       ├── policies.go
│       ├── events.go            # SSE streaming endpoint
│       └── api_test.go
├── services/
│   ├── memory/
│   │   ├── service.go           # Hierarchical pattern merging (org->team->repo)
│   │   ├── service_test.go
│   │   └── adapter.go           # PlatformMemoryStore (implements harness MemoryProvider)
│   ├── policy/
│   │   ├── engine.go            # Most-restrictive-wins merge + config enforcement
│   │   ├── engine_test.go
│   │   └── guard.go             # Pre-run policy guard
│   └── cost/
│       ├── governor.go          # Budget tracking, limit enforcement, alerts
│       ├── governor_test.go
│       └── hooks.go             # Harness cost hook integration
├── storage/
│   ├── storage.go               # Store interface + domain types (Run, Pattern, etc.)
│   ├── testing.go               # Compliance test suite for any backend
│   ├── sqlite/
│   │   ├── sqlite.go            # SQLite backend implementation
│   │   ├── sqlite_test.go
│   │   ├── migrations.go        # Schema migrations
│   │   ├── runs.go
│   │   ├── memory.go
│   │   ├── costs.go
│   │   ├── policies.go
│   │   └── events.go
│   └── s3/
│       ├── artifacts.go         # ArtifactStore interface
│       ├── artifacts_test.go
│       └── local.go             # Local filesystem artifact backend
├── doc.go                       # Package-level documentation
├── go.mod
├── go.sum
└── VERSION                      # Current version (v0.1.0)
```

## API Summary

All API endpoints are prefixed with `/api/v1` and require organizational scope headers.

| Method | Endpoint                     | Description                                 |
|--------|------------------------------|---------------------------------------------|
| GET    | `/api/v1/health`             | Health check                                |
| GET    | `/api/v1/runs`               | List runs (filterable by scope, status)     |
| GET    | `/api/v1/runs/{id}`          | Get a single run by ID                      |
| POST   | `/api/v1/runs`               | Create a new run record                     |
| GET    | `/api/v1/memory/patterns`    | List patterns for the current scope         |
| POST   | `/api/v1/memory/patterns`    | Create a new pattern                        |
| GET    | `/api/v1/memory/preferences` | Get preferences for the current scope       |
| PUT    | `/api/v1/memory/preferences` | Set preferences for the current scope       |
| GET    | `/api/v1/costs/summary`      | Get usage summary (grouped by time period)  |
| GET    | `/api/v1/costs/budget`       | Get budget for the current scope            |
| PUT    | `/api/v1/costs/budget`       | Set budget for the current scope            |
| GET    | `/api/v1/policies`           | Get policy for the current scope            |
| PUT    | `/api/v1/policies`           | Set policy for the current scope            |
| GET    | `/api/v1/policies/effective` | Get merged effective policy for the scope   |
| GET    | `/api/v1/events`             | Query stored events                         |
| GET    | `/api/v1/events/stream`      | SSE stream of real-time events              |
| GET    | `/`                          | Web dashboard (embedded SPA)                |

### Scope Headers

Every request should include the following headers to identify the organizational context:

```
X-Boatman-Org:  <org_id>
X-Boatman-Team: <team_id>
X-Boatman-Repo: <repo_id>
```

## Development Commands

```bash
# Build the platform binary
make build-platform

# Run all platform tests
make test-platform

# Format Go code
make fmt

# Download dependencies (Go + npm)
make deps

# Build the dashboard frontend (required for embedded SPA)
cd platform/dashboard/frontend && npm install && npm run build
```

## Dependencies

| Dependency                        | Purpose                                           |
|-----------------------------------|---------------------------------------------------|
| `modernc.org/sqlite`              | Pure-Go SQLite driver (no CGO required)           |
| `github.com/nats-io/nats-server`  | Embedded NATS server for the event bus            |
| `github.com/nats-io/nats.go`      | NATS client library                               |
| `github.com/philjestin/boatman-ecosystem/harness` | Core harness module (runner, memory, cost types) |
| `github.com/philjestin/boatman-ecosystem/shared`  | Shared event types for legacy bridge compatibility |

The dashboard frontend uses React, TypeScript, Vite, Tailwind CSS, and Recharts.

## Configuration

### Server Flags

| Flag         | Default                   | Description                           |
|--------------|---------------------------|---------------------------------------|
| `--port`     | `8080`                    | HTTP server port                      |
| `--data-dir` | `~/.boatman/platform`     | Data directory for SQLite and NATS    |

### CLI Configuration (`.boatman.yaml`)

```yaml
platform:
  server: "http://localhost:8080"   # Platform server URL
  org_id: "my-org"                  # Organization identifier
  team_id: "my-team"               # Team identifier
```

When `platform.server` is empty or the server is unreachable, the CLI operates in standalone mode with local-only memory and no policy enforcement.
