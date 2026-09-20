---
id: SPEC-020
title: Push progress after each phase
status: accepted
capability: delivery
created: 2026-09-20
updated: 2026-09-20
accepted_by: TheJisus28
---

## Problem

A spec's work reaches `origin` only at the end: `forge submit` pushes the
branch when the pull request opens. Between phases the work is local — a
commit that was never published or, worse, an uncommitted tree — so leaving
the machine, continuing from another one, or handing the spec to another
person loses the in-progress state. There is nothing between "shared with
nobody" and "the spec is finished".

## Acceptance criteria

Observable outcomes. Someone else must be able to mark each one pass or
fail with evidence.

- AC1: `forge push <id>` on a spec branch commits the pending work with a
  Conventional Commit message that names the spec and the phase, and pushes
  the branch to `origin` with its upstream set, so another machine can
  fetch the branch and resume.
- AC2: `forge push` targets only the spec branch; it refuses on `main` or
  `master` and never pushes or merges into the default branch.
- AC3: With nothing to commit and the branch already up to date,
  `forge push` succeeds and reports that there is nothing to push, instead
  of failing.
- AC4: When `.forge/project.md` opts in (a `push: on` frontmatter scalar),
  `forge advance` commits and pushes at each phase or state boundary;
  without the opt-in, `forge advance` never touches the network.
- AC5: `docs/cli.md`, `forge workflow`/`docs/workflow.md` and the roles
  describe the checkpoint and when to run it; the implementer runs
  `forge push` after each phase so a handoff is possible.
- AC6: `go test ./...` passes.

## Open questions

One per line, like `- OQ1: ...`, for anything that must be answered before
the contract is approved. Write `None.` when there are none.

None.

## Contract

Written by the architect once the work is accepted, and frozen once
approved. Real names from this repository: modules, endpoints,
tables, screens. Numbered decisions with what they discard. Anything other
specs will build against goes here.

## Existing state

<!-- Written by the architect, read while planning: name what already
     exists that this builds on, in this repository's real names.
     - Delivered specs and contracts this one builds on (`forge status`).
     - Modules, files or tools that already do part of the job: reuse them.
     - Conventions in `.forge/conventions/` that apply.
     - Duplication this change deliberately avoids.
     - What does not exist yet and genuinely has to be built. -->

## Out of scope

A closed list.

## History

Written by `forge`. Do not edit by hand.
- 2026-09-20  accepted  by TheJisus28
