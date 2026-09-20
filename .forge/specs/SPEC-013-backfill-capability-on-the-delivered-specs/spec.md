---
id: SPEC-013
title: Backfill capability on the delivered specs
status: reviewing
capability: workflow
created: 2026-09-20
updated: 2026-09-20
accepted_by: TheJisus28
conductor: TheJisus28
approved_by: TheJisus28
contract_hash: 4d8a82b3f4cf
---

## Problem

The specs delivered before this change (SPEC-001 to SPEC-009) declare no
`capability`, so the derived view would omit the entire history and
`forge validate` would warn on nine specs forever. The field only pays off
once the existing contracts carry it.

## Acceptance criteria

- AC1: Each of SPEC-001 to SPEC-009 declares a `capability` that matches
  the part of Forge it changed (for example `workflow`, `delivery`,
  `guard`, `upgrade`).
- AC2: `forge validate` reports no missing-capability warning.
- AC3: `forge capabilities` lists the delivered specs under their
  capabilities.
- AC4: Only `capability` is added to those specs' frontmatter; no contract,
  status or history line changes.

## Open questions

One per line, like `- OQ1: ...`, for anything that must be answered before
the contract is approved. Write `None.` when there are none.

None.

## Contract

### Decisions

**1. The nine delivered specs get these capabilities**, chosen as the part
of Forge each one changed:

| Spec | Capability | Why |
|---|---|---|
| SPEC-001 | `init` | Changes what `forge init` plants and preserves. |
| SPEC-002 | `delivery` | A spec ends as a pull request. |
| SPEC-003 | `specs` | The on-disk spec/plan/review layout. |
| SPEC-004 | `delivery` | Guard, gates and the delivery loop. |
| SPEC-005 | `agents` | Reaching the agent inside its session. |
| SPEC-006 | `agents` | One session per spec. |
| SPEC-007 | `packaging` | Kit served from the binary. |
| SPEC-008 | `upgrade` | `forge upgrade`. |
| SPEC-009 | `guard` | What the guard classifies as product code. |

Discards: one bucket such as `core` for everything (the view would say
nothing), and reusing a spec's own title as its capability (the capability
names the subsystem, not the change).

**2. Only the `capability:` line is added to each spec.** Everything else —
`status`, `updated`, `contract_hash`, `pr_url`, the Contract, the History —
stays byte for byte. The line goes directly under `status:`, where the
template (SPEC-010) puts it. `updated` is not bumped: a backfill is not a
change to the spec's content, and bumping nine dates would dirty the history
for no reader. Discards: re-running a `forge` writer over the files (it could
normalise unrelated fields), and hand-adding the field to specs the view does
not read.

**3. `forge validate` must fall silent about capability for the whole tree.**
The only capability finding allowed after this change is the existing
malformed-slug error; no `missing capability` warning. A test asserts the
delivered specs carry a valid capability so a future copy/paste cannot drop
one.

### Tests

- `internal/project/project_test.go` or `internal/validate/validate_test.go`:
  every spec under the repo's `.forge/specs/` that is `done` has a
  `ValidCapability` (a guard against a future backfill hole).
- `internal/cli/cli_test.go`: `forge validate` over the repository's own
  `.forge` prints no missing-capability warning (or, if the suite stays
  hermetic with a scratch tree, a scratch tree with the nine specs passes).

### Out of scope of this contract

- Any change to `capability` parsing, validation or the view.
- Backfilling anything other than `capability` (no `supersedes` history).

## History

Written by `forge`. Do not edit by hand.
- 2026-09-20  accepted  by TheJisus28
- 2026-09-20  specifying  by TheJisus28
- 2026-09-20  awaiting-approval  by orchestrator
- 2026-09-20  planning  by TheJisus28
- 2026-09-20  implementing  by orchestrator
- 2026-09-20  reviewing  by orchestrator
