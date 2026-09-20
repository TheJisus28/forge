---
id: SPEC-002
title: Deliver work as a pull request
status: proposed
created: 2026-09-19
updated: 2026-09-19
---

## Problem

The agent that works a spec also merges its branch into `main` and pushes,
so there is no point where a person approves the change, and the user
recorded in the spec is the machine's git identity. A spec should end as a
pull request that someone reviews and merges, opened as the authenticated
GitHub user.

## Acceptance criteria

Observable outcomes. Someone else must be able to mark each one pass or
fail with evidence.

- AC1: A spec's work stays on its own branch and ends as a pull request
  against `main`. Nothing in Forge merges it; a person reviews and merges
  on GitHub.
- AC2: Forge can open that pull request itself through `gh`, and when `gh`
  is unavailable it prints the exact `git push` and `gh pr create` instead
  of failing.
- AC3: The user recorded by `forge accept`, `forge start`, `forge approve`
  and in the spec history is the login of the authenticated `gh` user,
  falling back to `git config user.name` when `gh` is missing or not
  authenticated.
- AC4: `go test ./...` passes.

## Contract

Written by the architect once the work is accepted, and frozen once
approved. Real names from this repository: modules, endpoints,
tables, screens. Numbered decisions with what they discard. Anything other
specs will build against goes here.

## Out of scope

- Branch protection or CODEOWNERS on GitHub; the team configures those.
- Choosing a merge strategy (squash, rebase) or auto-merge.
- Identifying the user for anything other than the recorded actor.

## History

Written by `forge`. Do not edit by hand.
