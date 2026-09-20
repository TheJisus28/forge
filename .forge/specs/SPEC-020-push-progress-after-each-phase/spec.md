---
id: SPEC-020
title: Push progress after each phase
status: implementing
capability: delivery
created: 2026-09-20
updated: 2026-09-20
accepted_by: TheJisus28
orchestrator: TheJisus28
approved_by: TheJisus28
contract_hash: 4b53e1926bf1
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

Builds on SPEC-002 (`forge submit` opens the pull request) and SPEC-021
(verifiable criteria, the coverage rules). Every name below is this
repository's.

1. **A new `forge push [id]` command owns the checkpoint.** It lives in
   `internal/cli/push.go:cmdPush(args []string, out io.Writer) error`,
   dispatched from `internal/cli/cli.go` beside `submit` and listed in the
   usage text. It loads through `specArg`, then runs one checkpoint and
   prints one line. It discards a `forge advance --push` flag: the
   checkpoint is its own verb, so an implementer can run it after each phase
   without changing the state.

2. **The checkpoint commits the pending work, then pushes the branch.**
   `internal/project/git.go` gains three helpers the command, `forge
   advance` and `forge submit` share: `CommitAll(root, message) (bool,
   error)` (`git add -A` then `git commit -m message`; false when the tree
   is clean), `HasUnpushed(root) (bool, error)` (false when the branch has
   no upstream; otherwise `git rev-list --count @{upstream}..HEAD` is more
   than zero) and `OnDefaultBranch(root) bool` (the branch is `main` or
   `master`). Pushing reuses the existing `project.Push`. Staging is `git
   add -A`, the policy the archive step already tells the user to run,
   because a spec branch is single-purpose.

3. **The commit names the spec and the phase.** The subject is
   `chore(<ID>): checkpoint <state>`, where `<state>` is the spec's current
   workflow status, for example `chore(SPEC-020): checkpoint implementing`.
   The workflow state is the phase the CLI can name without judgement; the
   plan's numbered phases are not machine-readable and a checkbox count
   would be a brittle substitute. A future spec that wants the phase number
   must first give `tasks.md` a readable marker.

4. **`forge push` refuses the default branch before it commits.** When
   `OnDefaultBranch(root)` is true, `cmdPush` returns an error — a spec ends
   as a pull request, so run `forge submit` — and exits non-zero without
   committing or pushing (AC2). This repeats the rule
   `internal/cli/guard.go:commandDenial` already enforces for a raw
   `git push`, because the guard reads shell commands and never sees what
   `forge` runs internally.

5. **Nothing to do succeeds.** When `CommitAll` committed nothing and
   `HasUnpushed` is false, `forge push` prints
   `nothing to push: <branch> is up to date` and exits 0 (AC3). It discards
   matching git's "Everything up-to-date" text: the helpers answer the
   question directly.

6. **The auto-checkpoint is opt-in through `push:`.** `internal/project/
   project.go` gains `func (p *Project) PushEnabled() bool`, reading the
   frontmatter scalar `push` exactly as `GuardEnabled` reads `guard`
   (`on`/`true`/`yes`; anything else, including absent, is off). In
   `cmdAdvance`, after `s.Save()` on a transition it performs, a
   `PushEnabled` project runs the same checkpoint and prints its result; a
   project without the scalar runs no `git` and no `gh`, so `forge advance`
   is offline by default (AC4). The `--to accepted` and `--to planning`
   routes delegate to `forge accept`/`forge approve`, which keep their own
   bookkeeping and do not auto-checkpoint; run `forge push` after them.

7. **A failed checkpoint after a state move is a warning, not a failure.**
   `forge advance` saves the new state first; if the commit or push then
   fails it prints `warning: ...` on `out` and exits 0, per
   `.forge/conventions/cli-output.md`, because the move stands and the work
   is still local, and the next `forge push` retries. `forge push` itself
   returns the git error, because it has nothing else to report.

8. **`forge submit` reuses the checkpoint.** Before its existing
   `project.Push` and `gh pr create`, `cmdSubmit` commits pending work with
   `CommitAll` (skipping the commit when the tree is clean), so the final
   step never leaves uncommitted work out of the pull request. It still
   opens the PR; nothing here merges or changes the guard.

### Interfaces

- `internal/cli/push.go: cmdPush(args []string, out io.Writer) error`, and
  `checkpoint(p *project.Project, s *project.Spec, out io.Writer) error`
  (commit then push, one report line), used by `cmdPush`, `cmdAdvance` and
  `cmdSubmit`.
- `internal/project/git.go: CommitAll`, `HasUnpushed`, `OnDefaultBranch`.
- `internal/project/project.go: (*Project).PushEnabled`.
- `internal/cli/work.go: cmdAdvance` calls `checkpoint` when
  `p.PushEnabled()`; `internal/cli/report.go: cmdSubmit` calls `CommitAll`.
- `.forge/project.md` (and a target project's) frontmatter: `push: on`.

### Tests

- `TestPush_CommitsAndPushes` — a temp repo with an `origin` bare remote, on
  `spec/001-...` with a changed file: `forge push SPEC-001` commits
  `chore(SPEC-001): checkpoint <state>`, pushes, and sets the upstream.
- `TestPush_RefusesDefaultBranch` — on `main`, `forge push` exits non-zero
  and neither the tree nor the remote moved.
- `TestPush_NothingToPush` — clean tree, branch up to date: exit 0 and the
  `nothing to push` line.
- `TestCommitMessage_NamesSpecAndState` — the subject builder.
- `TestAdvance_CheckpointsWhenOptedIn` — with `push: on` and a bare remote,
  `forge advance SPEC-001 --to implementing` commits and pushes; without the
  scalar, the same call leaves `git status --porcelain` and the remote
  unchanged (the offline guarantee).
- `TestAdvance_PushFailureIsAWarning` — a failing push leaves the spec in
  the new state and exits 0.
- `TestSubmit_CommitsBeforePush` — a dirty tree is committed before the PR
  step, with the `gh` seam faked as the suite already does.

## Existing state

- SPEC-002 delivered `forge submit` (`internal/cli/report.go:cmdSubmit`):
  it pushes the branch with `project.Push` and opens the PR with `gh`. This
  change gives submit a commit step and reuses that push.
- SPEC-004 hardened the delivery loop. The phase states live in
  `internal/workflow/workflow.go` (`contracting`, `planning`, `implementing`,
  `reviewing`, ...) and `forge advance` crosses them in
  `internal/cli/work.go:cmdAdvance`.
- `internal/project/git.go` already has `Branch`, `SpecIDFromBranch`,
  `BranchName`, `Push`, `HasGH` and `GH`; the three new helpers sit beside
  them and `Push` is reused unchanged.
- `internal/cli/guard.go:commandDenial` already refuses `git push` while on
  the default branch; `forge push` repeats the rule because the guard never
  sees the git commands Forge runs itself.
- Config scalars are read with `p.Config.Str`; `GuardEnabled`
  (`internal/project/project.go`) is the shape `PushEnabled` copies.
- `internal/doc` reads and writes frontmatter (`Str`, `SetStr`); the
  `push: on` scalar needs no parser.
- `.forge/conventions/cli-output.md` (a non-fatal notice is a `warning: `
  line and exit 0) and `testing.md` (a command that promises not to touch
  the network is proven by a before/after snapshot and a faked seam) govern
  the tests.
- Not built: the command, the three git helpers, `PushEnabled`, the
  `cmdAdvance`/`cmdSubmit` wiring, and the docs.

## Out of scope

- Per-file or interactive staging; the checkpoint stages the whole tree.
- Any push to the default branch, any merge, or a change to the guard's
  rules.
- A background or scheduled push, or a push outside a spec branch.
- Reworking `forge sync` or the pull-request metadata.
- Reading a phase number out of `tasks.md`.

## History

Written by `forge`. Do not edit by hand.
- 2026-09-20  accepted  by TheJisus28
- 2026-09-20  contracting  by TheJisus28
- 2026-09-20  planning  by TheJisus28
- 2026-09-20  implementing  by orchestrator
