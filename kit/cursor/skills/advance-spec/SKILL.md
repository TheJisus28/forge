---
name: advance-spec
description: Advance a Forge SPEC status in record.md and BOARD.md. Use when the owner approves a spec, a phase finishes, QA reports, or the user says approved / lgtm / dale / to review / close the spec.
---

# Advance spec

Orchestrator only. Never a subagent.

## Steps

1. Read `forge/specs/SPEC-XXX/record.md`.
2. Check the destination artifact (`forge/LIFECYCLE.md`).
3. `specifying → specified` requires owner `approved` / `lgtm` / `dale`.
4. Frontmatter: `status` + `updated_at`.
5. Append history; never delete previous lines:

```markdown
<!-- history at="ISO-8601" from="FROM" to="TO" by="orchestrator" -->
One or two sentences of why.
```

6. Update `forge/BOARD.md`.
7. If needed, backlog `promoted` / `spec_id`.

## Illegal

- Skipping `specified`.
- `reviewing → done` with P0 findings.
- Status changes by architect / implementer / qa.
