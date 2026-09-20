---
id: SPEC-003
title: Adopt the SDD spec layout
status: done
created: 2026-09-20
updated: 2026-09-20
accepted_by: TheJisus28
conductor: TheJisus28
approved_by: TheJisus28
pr_url: "https://github.com/TheJisus28/forge/pull/3"
pr_state: open
pr: 3
---

## Problem

Specs lived in one flat file, `.forge/specs/SPEC-NNN-slug.md`, and the plan,
changes and review scaffolding lived in `.forge/wip/<id>/` and was deleted at
archive. Nothing beyond the contract survived, so a later spec could not see
how earlier work was verified or where it landed. The layout also did not
match the per-feature folders people know from spec-driven-development tools.

## Acceptance criteria

- AC1: A spec is a folder `.forge/specs/SPEC-NNN-slug/` holding `spec.md`,
  `plan.md`, `tasks.md` and `review.md`, and `forge new` creates it.
- AC2: `forge archive` keeps the folder; it no longer deletes scaffolding.
- AC3: `.forge/wip/` and the `changes.md` template are gone from the kit and
  from an initialized repository.
- AC4: `forge validate`, `forge status`, `forge brief` and `forge guard` work
  against the folder layout, and `go test ./...` passes.

## Contract

A spec is a folder with four files, each owned by one role:

    .forge/specs/SPEC-004-saved-cards/
      spec.md     state + Problem + Acceptance criteria + Contract + History
      plan.md     Existing state + approach
      tasks.md    phases as checkboxes, ticked as they land, with where
      review.md   verdict + acceptance criteria evidence

- `internal/project`: `Spec.Path` is the folder's `spec.md`; `Spec` exposes
  `Dir`, `PlanPath`, `TasksPath`, `ReviewPath`; `Project.SpecDir` and
  `SpecDirName`; `Load` reads `specs/*/spec.md`.
- `internal/cli`: `new` creates the folder; `archive` keeps it and requires
  `review.md`; `renumber` moves the folder.
- `internal/validate`: `plan.md` and `tasks.md` are required while
  implementing, `review.md` while reviewing and when done; no wip check.
- `kit/`: `templates/tasks.md` replaces `changes.md`, `wip/` is removed, and
  the roles, the `forge-work` skill and the wrappers point at the folder.

- Decision 1: a folder per spec, matching Spec Kit and Kiro, with the files
  `spec.md`, `plan.md`, `tasks.md` and `review.md`.
- Decision 2: the folder is the durable record; `archive` no longer deletes.
- Decision 3: `changes.md` is dropped; the implementer records what landed and
  where in `tasks.md`, the reviewer records the evidence in `review.md`.
- Decision 4: no backward compatibility; the old flat layout is not read.

## Out of scope

- Splitting a `design.md` out of `spec.md`; the contract stays in `spec.md`.
- Reading the old flat `specs/*.md` layout.

## History

Written by `forge`. Do not edit by hand.
- 2026-09-20  done  by TheJisus28: implemented directly, because the layout it changes is the one the workflow runs on
