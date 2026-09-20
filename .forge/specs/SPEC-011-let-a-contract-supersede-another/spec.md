---
id: SPEC-011
title: Let a contract supersede another
status: reviewing
capability: workflow
created: 2026-09-20
updated: 2026-09-20
accepted_by: TheJisus28
conductor: TheJisus28
approved_by: TheJisus28
contract_hash: 9bab05096b81
---

## Problem

When a contract replaces an earlier one, nothing records the link. A
derived view of current state would list both contracts as if they were
still in force, because there is no way to tell that one superseded the
other. Nothing can catch a supersede that points at a spec that does not
exist, or is not delivered, or that loops back on itself, or two live specs
rewriting the same predecessor. The link is the field that makes the
derived view tell the truth.

## Acceptance criteria

- AC1: A spec's frontmatter accepts `supersedes: [SPEC-NNN]`, a list.
- AC2: `forge validate` fails when a `supersedes` target does not exist.
- AC3: `forge validate` fails when a `supersedes` target is not `done`.
- AC4: `forge validate` fails when following `supersedes` reaches a cycle,
  and names the chain.
- AC5: `forge validate` fails when two specs that are not terminal supersede
  the same target, and names both.
- AC6: `forge status <id>` shows what a spec supersedes and what supersedes
  it.

## Open questions

One per line, like `- OQ1: ...`, for anything that must be answered before
the contract is approved. Write `None.` when there are none.

None.

## Contract

### Decisions

**1. The frontmatter key is `supersedes`, a list of ids.**
`supersedes: [SPEC-007]` names the contracts this one replaces, in whole.
It is parsed and written with the same list helpers `covers` already uses in
`internal/project/project.go` (`setStrList` on save, the list read in
`FromDoc`), and each id is normalised with `project.NormalizeID`. Discards: a
scalar (a contract may replace more than one), free text, and a body field.

**2. `project.Spec.Supersedes []string` carries it.**
Add the field beside `Covers`; `FromDoc` reads it, `Save` writes it. An empty
list removes the key. No new type and no new parser.

**3. `forge validate` owns the rules, in a new `checkSupersedes`.**
Called from `Run` with the intrinsic field checks, before the unknown-status
`continue` (convention `.forge/conventions/frontmatter.md`):
- a target that does not exist is an `Error`:
  `supersedes SPEC-NNN, which does not exist`.
- a target that is not `done` is an `Error`:
  `supersedes SPEC-NNN, which is not done`.
- a cycle (following `supersedes` returns to a spec already on the path) is
  an `Error` naming the chain, the same shape as the existing dependency
  cycle finding.
- two specs that are **not terminal** (`status != done` and `status !=
  dropped`) superseding the same target is an `Error` naming both, so a
  replacement cannot be claimed twice while it is still in flight.
Discards: a warning (a broken supersede silently resurrects a dead
contract), and enforcing uniqueness across terminal specs (history may have
several successors over time; SPEC-012 reads the chain).

**4. `forge status <id>` shows the link both ways.**
In `view.Detail`, after `capability`, a `supersedes` row listing the ids
(omitted when empty) and a `superseded by` row listing the ids of specs whose
`Supersedes` contains this spec (omitted when none).

### Interfaces other specs build against

Frontmatter contract (frozen): key `supersedes`, a list of `SPEC-NNN`.
Go: `project.Spec.Supersedes []string`; `Save` adds or removes only that key.
Validate: the four findings above. View: the two `Detail` rows.

### Tests

- `internal/project/project_test.go`: a fixture with `supersedes: [SPEC-002]`
  reads back; `TestSaveRoundTrip` keeps it.
- `internal/validate/validate_test.go`: missing target, not-done target,
  cycle, and double-supersede each produce the expected `Error`; a clean
  fixture with a valid supersede is quiet.
- `internal/cli/cli_test.go`: end-to-end, a spec superseding a `done` one
  exits 0 and shows both `supersedes` and `superseded by` in `forge status`.

### Out of scope of this contract

- The derived view and the computed `superseded` marking (SPEC-012).
- Section-level supersedes (`SPEC-NNN#section`), deferred by decision 0004.
- Any change to the `capability` field (SPEC-010).

## History

Written by `forge`. Do not edit by hand.
- 2026-09-20  accepted  by TheJisus28
- 2026-09-20  specifying  by TheJisus28
- 2026-09-20  awaiting-approval  by orchestrator
- 2026-09-20  planning  by TheJisus28
- 2026-09-20  implementing  by orchestrator
- 2026-09-20  reviewing  by orchestrator
