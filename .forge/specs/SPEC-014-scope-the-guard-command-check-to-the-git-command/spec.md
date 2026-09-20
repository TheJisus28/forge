---
id: SPEC-014
title: Scope the guard command check to the git command
status: proposed
capability: guard
created: 2026-09-20
updated: 2026-09-20
---

## Problem

`commandDenial` runs its push and merge patterns over the whole shell line,
and `pushesToDefault` scans every token of that line for `main` or `master`.
A compound command that pushes a feature branch and then does something that
merely mentions the default branch — for example

    git push -u origin my-branch && gh pr create --base main

is denied as if the push targeted `main`. The guard blocks a correct
command, so people work around it instead of trusting it, which is worse
than not having it.

## Acceptance criteria

- AC1: `git push -u origin my-branch && gh pr create --base main` is
  allowed; the token `main` after the push does not make the push target the
  default branch.
- AC2: `git push origin main` and `git push origin HEAD:main` are denied.
- AC3: `git push` while the current branch is `main` or `master` is denied;
  the same command on a feature branch is allowed.
- AC4: The decision looks only at the arguments of the `git push` (and
  `git merge`) command, not at later segments separated by `;`, `&&`, `||`
  or `|`.
- AC5: `forge guard --command "..." --explain` prints `would allow` or
  `would deny` consistent with AC1–AC4.

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
