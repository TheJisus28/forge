---
id: SPEC-016
title: Add a fast lane for bug fixes and chores
status: proposed
capability: workflow
created: 2026-09-20
updated: 2026-09-20
---

## Problem

Every change goes through the same road: architect, contract, approval,
planning, then code. For a one-line bug fix or a chore with no design
decision, that road costs more than the change, so people either avoid
Forge for small work or write a spec whose contract restates the obvious.
The tool becomes something to route around instead of the default.

## Acceptance criteria

- AC1: `forge new "<title>" --fix` (or an equivalent flag) creates a
  fast-lane spec that skips the contract and approval states.
- AC2: A fast-lane spec still requires a Problem, acceptance criteria, a
  `review.md` with evidence, and a pull request; nothing is merged without a
  human.
- AC3: `forge guard` allows product code edits once a fast-lane spec is
  started, and still blocks them otherwise.
- AC4: `forge validate` requires `review.md` for a fast-lane spec before
  `done`, exactly as for a full spec.
- AC5: `forge workflow` and `docs/workflow.md` describe the fast lane and
  when a change does not qualify for it.

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
