---
id: SPEC-001
title: Keep an existing AGENTS.md on init
status: done
capability: init
created: 2026-09-19
updated: 2026-09-19
accepted_by: carra
conductor: carra
approved_by: carra
contract_hash: 545ec4fadeb3
---

## Problem

`forge init` overwrites an existing `AGENTS.md` and `CLAUDE.md`, deleting
the project-specific instructions they hold. This repository had to merge
its `AGENTS.md` by hand after `forge init`. It contradicts the documented
"existing files are kept" and the rule that anything outside `.forge/kit/`
is written once and never overwritten.

## Acceptance criteria

Observable outcomes. Someone else must be able to mark each one pass or
fail with evidence.

- AC1: `forge init` in a repository whose `AGENTS.md` already exists leaves
  its content unchanged. The same holds for `CLAUDE.md`.
- AC2: `forge init --force` overwrites them with the kit copies.
- AC3: `forge update` leaves an existing `AGENTS.md` unchanged.
- AC4: `go test ./...` passes.

## Contract

In `internal/cli/init.go`, `kitOwned` no longer claims `AGENTS.md` or
`CLAUDE.md`. Those two root pointers are written when missing and
overwritten only with `--force`, exactly like the rest of the tree outside
`.forge/kit/`. `.forge/README.md`, `.forge/kit/`, `.claude/` and
`.opencode/` stay kit-owned and keep refreshing on `forge update`.

- Decision 1: keep an existing file instead of merging into it. Forge
  cannot guess how to combine someone else's instructions, and the planted
  hooks and skills already carry the workflow when the user's file stays.
- Decision 2: `CLAUDE.md` follows `AGENTS.md`, because it is the same
  one-line pointer to it.

`docs/cli.md` is updated: `forge update` refreshes the kit, not the root
pointers a user owns.

## Out of scope

- Merging Forge instructions into an existing `AGENTS.md`.
- Which files are kit-owned beyond the two root pointers.
- Planting the GitHub validation workflow.

## History

Written by `forge`. Do not edit by hand.
- 2026-09-19  accepted  by carra
- 2026-09-19  specifying  by carra
- 2026-09-19  awaiting-approval  by carra
- 2026-09-19  planning  by carra
- 2026-09-19  implementing  by carra: Plan written: single phase, stop claiming root pointers
- 2026-09-19  reviewing  by carra: single phase implemented, tests pass
- 2026-09-19  done  by orchestrator: archived
