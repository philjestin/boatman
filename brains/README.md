# Domain Brains

Domain brains are curated context files that make Boatman's agents competent in specific areas of your codebase.

## How It Works

1. Drop a `.md` file in this directory
2. Add YAML frontmatter with matching criteria (keywords, labels, file paths)
3. Boatman automatically loads matching brains for each task
4. Phase-specific sections (`## Planning`, `## Execution`, `## Review`) target the right agent

## Brain Format

```markdown
---
domain: Human-readable domain name
keywords:
  - terms that trigger this brain
  - matched against task title + description
labels:
  - matched against task labels (e.g. Linear labels)
paths:
  - directory/paths that trigger matching
  - when planned files touch these paths
---

## Domain Model
Entities, relationships, invariants. The stuff that takes 6 months to learn.

## Planning
Architecture context the planner needs. Key decisions, boundaries, dependencies.

## Execution
Implementation patterns. Code paths. Conventions. "Don't do X" rules.

## Review
What to check for. Common mistakes. Domain-specific review criteria.

## Common Mistakes
Things AI agents always get wrong in this domain.
```

## Matching

Brains are matched in two passes:

1. **Before planning** - matched on task title, description, and labels
2. **After planning** - re-matched with planned file paths for broader coverage

Multiple brains can match a single task. Phase-specific sections ensure each agent gets only the context it needs.

## Writing Good Brains

- **Be specific.** "Always use `PaymentService`" beats "follow best practices."
- **Include what goes wrong.** The `## Common Mistakes` section prevents the errors you've seen before.
- **Keep it focused.** One brain per domain. A 200-line brain is better than a 2000-line one.
- **Use the phase sections.** The planner doesn't need implementation details. The executor doesn't need architecture rationale.
- **Update when things change.** A brain that references deprecated code is worse than no brain.

## Files

- `_example.md` - template to copy when creating a new brain
- Files starting with `_` are ignored by the loader
