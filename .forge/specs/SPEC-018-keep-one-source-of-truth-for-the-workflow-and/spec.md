---
id: SPEC-018
title: Keep one source of truth for the workflow and roles
status: proposed
capability: workflow
created: 2026-09-20
updated: 2026-09-20
---

## Problem

The loop is written down six times — `AGENTS.md`, `kit/AGENTS.md`,
`docs/workflow.md`, `kit/machine/WORKFLOW.md`, the `forge-work` skill and
the roles — and they have already drifted. `conductor` is a frontmatter
field and a role in `docs/teams.md` and in the `blocked` message, but
`forge roles` lists four roles and none is named conductor. `docs/cli.md`
says `forge start` creates `plan.md` and `tasks.md`; the code only creates
the folder. `forge archive` matches Spanish headings (`Convenciones
propuestas`) by literal string in a machine contract.

## Acceptance criteria

- AC1: The states, the transitions and the roles have one normative source
  in the binary (`forge workflow`, `forge roles`); the Markdown pages link
  to it instead of restating it, and a test fails if a page's state list
  diverges.
- AC2: There is a single name for whoever drives a spec, it is one of the
  roles `forge roles` lists, and it is used consistently in frontmatter,
  help text and docs.
- AC3: `docs/cli.md` matches behaviour: either `forge start` creates
  `plan.md` and `tasks.md`, or the page no longer claims it does.
- AC4: The headings the CLI reads are English-only and covered by a test
  that fails if a second-language alias is relied on.
- AC5: `docs/` does not restate the state machine; it links to `forge
  workflow`.

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
