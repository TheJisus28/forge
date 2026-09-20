---
id: SPEC-004
title: Harden the delivery loop
status: done
capability: delivery
created: 2026-09-20
updated: 2026-09-20
accepted_by: TheJisus28
conductor: TheJisus28
approved_by: TheJisus28
contract_hash: c24dd3ad858f
pr_url: "https://github.com/TheJisus28/forge/pull/5"
pr_state: open
pr: 5
---

## Problem

The loop has four rough edges. An agent can still merge or push to the
default branch even though the work is meant to end as a pull request.
Nothing stops a contract being approved while the spec still has open
questions. The brief does not say how far a spec's tasks have gone. And
`forge board` renders information `forge status` already shows, at the cost
of one more generated file.

## Acceptance criteria

Observable outcomes. Someone else must be able to mark each one pass or
fail with evidence.

- AC1: `forge guard` denies an agent's `gh pr merge` and any `git push` or
  `git merge` that lands on the default branch, through `--command` and
  through the Claude and OpenCode Bash hooks.
- AC2: `spec.md` has an `## Open questions` section; `forge approve` refuses
  while it is unresolved and `forge validate` warns.
- AC3: `forge brief` and `forge status <id>` show the spec's task progress as
  `done/total`, read from `tasks.md`.
- AC4: `forge board` and `.forge/BOARD.md` are gone, and `go test ./...`
  passes.

## Contract

- `internal/cli/guard.go`: `forge guard` learns `--command <cmd>` and a Bash
  input shape; `commandDenial` refuses `gh pr merge`, a `git push` to
  `main`/`master`, and a `git merge` while on the default branch.
- `internal/cli/init.go`: the Claude `PreToolUse` hook matches
  `Write|Edit|Bash`; `forge update` updates the matcher of the existing hook.
- `kit/opencode/plugins/forge-guard.js`: also handles `input.tool === "bash"`.
- `internal/cli/work.go`: `cmdApprove` refuses while `## Open questions` is
  unresolved, reusing `isNone`.
- `internal/validate`: warns when `## Open questions` is unresolved.
- `internal/project/project.go`: `Spec.TaskProgress` counts `- [x]` against
  `- [ ]` in `tasks.md`; `internal/view` shows it in `Brief` and `Detail`.
- Removed: `view.Board`, `cmdBoard`, the `board` command and `ensureGitignore`.

- Decision 1: the guard denies commands, not only file edits, so "a person
  merges" is enforced rather than merely asked. It covers `gh pr merge` and
  pushes or merges to the default branch, not commits.
- Decision 2: open questions gate `approve`, because a contract built on
  unresolved questions is the expensive mistake Forge exists to prevent.
- Decision 3: task progress is derived from `tasks.md`, never stored.
- Decision 4: `board` is removed; `forge status` already renders its table.

## Open questions

Questions that must be answered before the contract is approved. Write
`None.` when there are none.

None.

## Out of scope

- Blocking `git commit`; only pushes and merges to the default branch.
- Judging whether the answers to open questions are any good.

## History

Written by `forge`. Do not edit by hand.
- 2026-09-20  accepted  by TheJisus28
- 2026-09-20  specifying  by TheJisus28
- 2026-09-20  awaiting-approval  by TheJisus28
- 2026-09-20  planning  by TheJisus28
- 2026-09-20  implementing  by TheJisus28: 4 phases: guard command, clarify gate, task progress, drop board
- 2026-09-20  reviewing  by TheJisus28: 4 phases done, tests pass
- 2026-09-20  done  by orchestrator: archived
