---
id: SPEC-010
title: Add capability to every contract
status: specifying
created: 2026-09-20
updated: 2026-09-20
accepted_by: TheJisus28
conductor: TheJisus28
---

## Problem

A contract says what a change does, but not which part of the system it
touches. Nothing groups the contracts about the guard, or the upgrade path,
or the workflow; a reader asking "which contracts concern this capability?"
has to read every `done` contract, and a change that later replaces another
looks unrelated to it. Without a shared axis to group by, the derived view
that decision 0004 chooses has nothing to organise on.

## Acceptance criteria

- AC1: `forge new "<title>"` without `--capability` fails with a message
  that says to pass `--capability <name>`; a spec is no longer creatable
  without one.
- AC2: `forge new "<title>" --capability guard` writes `capability: guard`
  to the new spec's frontmatter.
- AC3: `forge validate` reports a warning naming each spec that has no
  `capability`, so the delivered specs are visible without failing the build.
- AC4: `forge new --capability <name>` prints a warning when no existing
  spec declares `<name>`, and creates the spec anyway, so a typo is visible
  but a genuinely new capability is allowed.
- AC5: `forge validate` reports an error when `capability` is present but
  empty or not a lowercase slug (`[a-z0-9-]+`).
- AC6: `forge status <id>` and `forge brief` show the capability, and
  `forge template spec` and `docs/cli.md` document the field.

## Open questions

One per line, like `- OQ1: ...`, for anything that must be answered before
the contract is approved. Write `None.` when there are none.

None.

## Contract

Written by the architect once the work is accepted, and frozen once
approved. Real names from this repository: modules, endpoints,
tables, screens. Numbered decisions with what they discard. Anything other
specs will build against goes here.

## Out of scope

A closed list.

## History

Written by `forge`. Do not edit by hand.
- 2026-09-20  accepted  by TheJisus28
- 2026-09-20  specifying  by TheJisus28
