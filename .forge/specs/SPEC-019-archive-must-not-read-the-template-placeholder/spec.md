---
id: SPEC-019
title: Archive must not read the template placeholder as a pending convention
status: proposed
capability: workflow
created: 2026-09-20
updated: 2026-09-20
---

## Problem

`forge archive` refuses while a spec's `plan.md`, `tasks.md` or `review.md`
still carries a `## Proposed conventions` section whose body is not exactly
`None.`. The file templates ship that section with an explanatory sentence
("Patterns decided because nothing was written. The team decides whether they
become rules…") above the `None.`, so a spec that never proposed anything
blocks its own archive until someone deletes the template's own text. It was
hit while archiving SPEC-010.

## Acceptance criteria

- AC1: `forge archive` succeeds when a spec's `Proposed conventions` section
  carries only the template's explanatory text and `None.`.
- AC2: `forge archive` still refuses when the section carries a real
  proposal, and names the file and the first line.
- AC3: The templates and the detection agree: either the placeholder text is
  not shipped in the section, or the detector ignores it. The chosen shape is
  stated in the contract.
- AC4: A test covers both the template default (archives) and a real
  proposal (refuses).

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
