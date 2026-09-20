---
id: 0001
date: "2026-09-20"
status: accepted
spec: SPEC-007
---

# Forge's machinery is internal and the templates are a protected standard

## Context

Forge plants its machinery — the workflow, the agent roles and the file
templates — inside `.forge/kit/`. That folder mixes two different things:
the project's decisions (specs, decisions, conventions) and the artifacts
that define how Forge itself behaves.

`docs/customizing.md` went further and promoted editing
`.forge/kit/templates/` per project, with `forge new` reading the local
copy before the embedded one. That turned the shape of every spec, plan,
decision and convention into something any contributor could change in a
pull request.

We judged that dangerous. These documents are a standard: the whole value
of Forge is that the same flow means the same thing in every repository.
A standard that anyone can edit "a diestra y siniestra" drifts, and a
reviewer cannot tell a deliberate change from an accident.

## Decision

Forge's machinery lives in the binary and is not planted into `.forge/`.
`.forge/` holds only project decisions (`project.md`, `specs/`,
`decisions/`, `conventions/`). Templates are read only from the embedded
copy; per-project override is not supported.

## Alternatives

- **Keep templates overridable (status quo).** Rejected. A mutable
  standard invites drift, and the templates are precisely the artifacts
  least safe to edit casually.
- **Move only the agent roles, keep the templates.** Rejected. It leaves
  the most sensitive artifact editable and makes the rule arbitrary.
- **Move the machinery to a separate, Forge-owned top-level folder.**
  Rejected. It adds a second planted root and breaks the single rule that
  everything outside `.forge/kit/` belongs to the team while `kit/` is
  refreshed by `forge update`.

## Consequences

- We lose per-project templates. If a real need appears, it is revisited
  with a new decision that supersedes this one; it is not worked around
  by editing an internal file.
- `.forge/` becomes fully team-owned; `forge update` stops owning
  `.forge/kit/` entirely.
- The host adapters (`.claude/agents/`, `.opencode/agents/`) carry the
  role text inlined. Hosts Forge does not wrap lose the neutral copy that
  `.forge/kit/agents/` used to provide.
- The workflow and roles stay readable as documentation, but they are
  produced by Forge, never edited in place by a project.
