---
id: SPEC-017
title: Check the criteria before approving and across artifacts
status: proposed
capability: workflow
created: 2026-09-20
updated: 2026-09-20
---

## Problem

`forge approve` refuses only while `## Open questions` is non-empty. Nothing
checks that an acceptance criterion is verifiable, and `forge validate`
checks structure — ids, references, cycles, coverage between parent and
child — but not whether every criterion has a task that delivers it and, at
review time, an evidence line that settles it. A vague criterion, or a
criterion no task mentions, passes the gates untouched.

## Acceptance criteria

- AC1: `forge approve` refuses, or warns with a clear message, when a
  criterion cannot be verified by a command, a test, or a request and its
  response; the rule is stated in the contract of this spec.
- AC2: A read-only `forge check` reports each criterion with no matching
  task in `tasks.md` and each criterion with no evidence line in
  `review.md`, naming the criterion and the file it is missing from.
- AC3: `forge check` exits non-zero when a criterion is uncovered while the
  spec is `reviewing` or `done`.
- AC4: `forge validate` includes the criterion-to-task and
  criterion-to-evidence coverage as a warning, and as an error once the spec
  is `done`.
- AC5: `docs/cli.md` documents the check and the criterion rule.

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
