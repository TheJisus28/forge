---
id: SPEC-012
title: Derive the capability view
status: implementing
capability: workflow
created: 2026-09-20
updated: 2026-09-20
accepted_by: TheJisus28
conductor: TheJisus28
approved_by: TheJisus28
contract_hash: d16e1e49dec2
---

## Problem

"What does the system currently do?" can only be answered by reading the
union of every `done` contract in order and knowing which one replaced
which. `forge brief` listing the five most recent `done` specs is a
mitigation, not an answer: at forty specs, finding the current behaviour of
one area is archaeology. The answer should be computed from the contracts
on disk, grouped by capability and pruned by `supersedes`, without a
hand-maintained file that branches conflict on or that goes stale.

## Acceptance criteria

- AC1: `forge capabilities` lists every capability that any `done` contract
  declares, with the current contracts under each.
- AC2: `forge capabilities <name>` lists the current contracts of one
  capability.
- AC3: A superseded contract is marked as such in the view and its own file
  is not modified; `git status` stays clean after running the command.
- AC4: The output is deterministic: the same repository state produces
  byte-identical output, with no network call and no model.
- AC5: `forge brief` includes the capability summary, so a session starts
  knowing the current shape without reading every contract.
- AC6: `forge status` and the plan's `## Existing state` step can use the
  view; `docs/cli.md` documents `forge capabilities`.

## Open questions

One per line, like `- OQ1: ...`, for anything that must be answered before
the contract is approved. Write `None.` when there are none.

None.

## Contract

### Decisions

**1. `forge capabilities [name]` is a read-only command that derives the
view.** Without a name it lists every capability that at least one `done`
spec declares, alphabetically, each with the current contracts under it; with
a name it prints only that capability. A contract is a `done` spec, shown as
`SPEC-NNN  <title>`. A spec superseded by another spec that is not `dropped`
is shown with a `(superseded by SPEC-MMM)` suffix under its capability rather
than being omitted, so the history is visible but the current shape is
legible. A non-`done` spec is not part of the view (it is not current state
yet). Discards: a written `.forge/capabilities.md` (it conflicts and goes
stale), and listing in-flight specs (that is `forge status`).

**2. The output is deterministic and side-effect free.** Capabilities sort
alphabetically, contracts inside one capability sort by spec number. The
command reads `.forge/specs/` only: no `git`, no network, no model, no
`Save`, no file written. Running it twice on the same tree prints identical
bytes and leaves `git status` unchanged.
Discards: sorting by recency or by filesystem order (unstable), and caching
(a cache is a file that goes stale).

**3. `forge brief` carries the summary.** After the current spec and before
the delivered list, `Brief` prints one line per capability with the count of
current contracts, or nothing when there are none. The grouping logic lives
in a new `internal/project` method (`Capabilities()` returning ordered
groups) so `brief`, `capabilities` and any future caller share one
definition; the CLI formats, the package groups.
Discards: a second grouping implementation in `internal/cli`.

**4. The survey step points at the view.** `docs/cli.md` documents `forge
capabilities`; `kit/machine/roles/orchestrator.md`, the `forge-work` skill
and `docs/workflow.md`'s "Planning from what exists" mention it as the first
thing the survey reads before the contracts.
Discards: leaving the derived view undiscoverable in the process.

### Interfaces other specs build against

CLI: `forge capabilities [name]`. Go: `project.Project.Capabilities()`,
returning capabilities in name order with their contracts (id, title,
superseded-by) in number order, so `brief`, the command and tests share one
shape. The `capability` and `supersedes` frontmatter fields (SPEC-010,
SPEC-011) are the only inputs.

### Tests

- `internal/project/project_test.go`: `Capabilities()` groups by capability,
  orders names and contracts, and marks a superseded contract.
- `internal/cli/cli_test.go`: end to end, two `done` specs in different
  capabilities show under `forge capabilities`; `forge capabilities <name>`
  shows one; a spec superseded by another shows the `(superseded by …)`
  suffix; running the command twice is byte-identical; `forge brief` contains
  the capability summary.
- A test asserts the command writes nothing: after `forge capabilities`, the
  spec files are unchanged.

### Out of scope of this contract

- Any hand-maintained capabilities file.
- Grouping by anything other than `capability`.
- The `forge status` tree (it keeps listing every spec by state).

## History

Written by `forge`. Do not edit by hand.
- 2026-09-20  accepted  by TheJisus28
- 2026-09-20  specifying  by TheJisus28
- 2026-09-20  awaiting-approval  by orchestrator
- 2026-09-20  planning  by TheJisus28
- 2026-09-20  implementing  by orchestrator
