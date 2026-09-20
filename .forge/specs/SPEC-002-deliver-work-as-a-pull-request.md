---
id: SPEC-002
title: Deliver work as a pull request
status: done
created: 2026-09-19
updated: 2026-09-19
accepted_by: TheJisus28
conductor: TheJisus28
approved_by: TheJisus28
contract_hash: 78f3960f4f31
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

A spec ends as a pull request. `forge submit` pushes the spec branch and
opens the PR through `gh`, and records it on the spec; it never merges, and
a person merges on GitHub.

- `internal/project/git.go`: `GHUser(root)` returns the login of the
  authenticated `gh` user, or `""` when `gh` is missing or not
  authenticated; `Push(root, branch)` runs `git push -u origin <branch>`.
- `internal/cli/work.go`: `resolveActor` prefers `--by`, then `GHUser`, then
  `git config user.name`. `forge start` records its conductor the same way
  instead of reading git directly.
- `internal/cli/report.go`: `forge submit [id] [--base main]` pushes the
  branch, runs `gh pr create` with the spec id and title, prints the URL and
  records `pr`, `pr_url` and `pr_state`. Without `gh` it prints the exact
  `git push` and `gh pr create` commands and succeeds.
- `kit/` and `docs/`: the workflow ends as a pull request; the roles and the
  `forge-work` skill open it with `forge submit` and never merge.

- Decision 1: prefer the `gh` login and fall back to git. Forge must keep
  working without an account; the gh user is used when there is one.
- Decision 2: a command, not only guidance, because Forge already owns the
  link between a spec and its pull request (`pr`, `pr_url`, `pr_state`, the
  fields `forge sync` writes).
- Decision 3: `submit` never merges. Merging is a human action on GitHub.

## Out of scope

- Branch protection or CODEOWNERS on GitHub; the team configures those.
- Choosing a merge strategy (squash, rebase) or auto-merge.
- Identifying the user for anything other than the recorded actor.

## History

Written by `forge`. Do not edit by hand.
- 2026-09-19  accepted  by TheJisus28
- 2026-09-19  specifying  by TheJisus28
- 2026-09-19  awaiting-approval  by TheJisus28
- 2026-09-19  planning  by TheJisus28
- 2026-09-19  implementing  by TheJisus28: Plan written: identity, submit, docs
- 2026-09-19  reviewing  by TheJisus28: identity, submit and docs done
- 2026-09-19  done  by orchestrator: archived
