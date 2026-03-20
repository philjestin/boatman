---
domain: Example Domain
keywords:
  - example
  - sample
labels:
  - example-team
paths:
  - app/models/example
  - packs/example_domain
---

## Domain Model

Describe the core entities, their relationships, and invariants here.

- **Widget** - the primary entity, belongs to an Account
- **WidgetConfig** - 1:1 with Widget, stores runtime settings
- Invariant: a Widget cannot be activated without a valid WidgetConfig

## Planning

Context the planner needs to make good architectural decisions.

- Widget creation goes through `WidgetFactory`, never instantiate directly
- The `widget_events` table is append-only, never update rows
- WidgetConfig changes require a version bump (optimistic locking)

## Execution

Implementation patterns and conventions.

- All Widget mutations go through `WidgetService` (app/services/widget_service.rb)
- Use `Widget.active` scope, never `where(status: 'active')` directly
- Events are published via `WidgetEventPublisher`, not created manually
- Tests use `create(:widget, :with_config)` factory, not manual setup

## Review

What reviewers should check for in this domain.

- Direct database updates to widgets table (must go through service)
- Missing event publication after state changes
- WidgetConfig changes without version bump
- N+1 queries on Widget -> WidgetConfig (always `.includes(:config)`)

## Common Mistakes

Things AI agents consistently get wrong here.

- Creating Widgets without going through WidgetFactory
- Forgetting to publish events after status transitions
- Using `Widget.find` instead of `Widget.active.find` (includes soft-deleted)
- Putting business logic in the controller instead of WidgetService
