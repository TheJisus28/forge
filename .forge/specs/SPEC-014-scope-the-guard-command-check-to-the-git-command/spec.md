---
id: SPEC-014
title: Scope the guard command check to the git command
status: reviewing
capability: guard
created: 2026-09-20
updated: 2026-09-20
accepted_by: TheJisus28
conductor: TheJisus28
approved_by: TheJisus28
contract_hash: cd828464703f
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

### Decisions

**1. The command is read segment by segment.** `commandDenial` splits the
line into command segments at `;`, `&&`, `||`, `|` and newline, ignoring
those characters inside single or double quotes. A `git` command is
recognised only when it is the first word of its segment (leading
`VAR=value` assignments are skipped), so `echo git push origin main` is not
a push and the segment after `&&` is judged on its own. `git`'s global
options before the subcommand (`-C`, `-c`, `--git-dir`, `--work-tree`,
`--namespace`, `--exec-path`) are skipped. Discards: the current
whole-line `reGitPush`/`reGitMerge` regexes and `strings.FieldsFunc` scan,
which cross segment boundaries.

**2. A push is judged on its own refspecs.** Within a `git push` segment,
every positional argument is read as a refspec and denied when either side
of a `:` — after stripping a leading `+`, then `refs/heads/`, `heads/` or
`refs/` — is `main` or `master`. So `git push origin main`,
`git push origin HEAD:main` and `git push origin main:feature` are denied;
`git push -u origin my-branch` and `git push origin feature/main` are not.
A bare `git push` is denied only while the current branch is `main` or
`master` (unchanged, AC3). Discards: tokenising the whole line for `main`
(the bug), and treating the remote name (`origin`) as a refspec.

**3. `git merge` and `gh pr merge` keep their rule, scoped.** A `git merge`
segment is denied only when the current branch is the default (unchanged):
merging the default *into* a feature is fine. `gh pr merge` is still denied
wherever it appears, via the existing regex over the line, because merging
the pull request is the act being prevented, not a git invocation.
Discards: denying `git merge main` on a feature branch (that is a normal
update from the default branch).

**4. `--explain` uses the same path.** `forge guard --command "..." --explain`
prints `would allow`/`would deny` from `commandDenial`, so AC5 needs no
separate logic.

### Interfaces other specs build against

`commandDenial(p, command)` keeps its signature; only its internals change.
No command or flag is added. `pushesToDefault` and the whole-line push/merge
regexes are deleted.

### Tests

- `internal/cli/cli_test.go` extends `TestGuardCommand`: the compound line
  `git push -u origin my-branch && gh pr create --base main` is allowed; the
  existing denies (`git push origin main`, `HEAD:master`, `gh pr merge`) and
  allows (`git push -u origin spec/004-x`, `git push origin feature/main`)
  still hold.
- A git repository on `main` is created in the test; `git push` is denied
  there and allowed after `git checkout -b feature` (AC3).
- A `;`/`||` compound that pushes a feature then names `main` is allowed
  (AC4), and `echo git push origin main` is allowed (decision 1).
- `forge guard --command ... --explain` prints the matching verdict (AC5).

### Out of scope of this contract

- Any change to the file-edit guard (`denial`) or to `isProcessFile`.
- Shell parsing beyond the listed separators (command substitution,
  subshells, `eval`); the guard stays a heuristic, not a shell.
- Detecting `git push --all`/`--mirror` reaching the default branch.

## History

Written by `forge`. Do not edit by hand.
- 2026-09-20  accepted  by TheJisus28
- 2026-09-20  specifying  by TheJisus28
- 2026-09-20  awaiting-approval  by orchestrator
- 2026-09-20  planning  by TheJisus28
- 2026-09-20  implementing  by orchestrator
- 2026-09-20  reviewing  by orchestrator
