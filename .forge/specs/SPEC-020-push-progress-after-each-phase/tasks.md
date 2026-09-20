# Tasks — SPEC-020

The phases from the plan, as checkboxes. One phase is one implementer run
and one commit. Tick a phase when it lands and say where the work is, so a
later spec knows what exists without reading the diff.

- [x] Phase 1 — The git helpers and the opt-in flag. Moves: AC1, AC4.
  Where: `internal/project/git.go` (`CommitAll`, `HasUnpushed`,
  `OnDefaultBranch`), `internal/project/project.go` (`PushEnabled`); tests in
  `internal/project/git_test.go` (new).
  Landed: `CommitAll(root, message) (bool, error)` checks
  `git status --porcelain` first, so a clean tree returns `false, nil`; a
  dirty one runs `git add -A` and `git commit -m`, returning `true`.
  `HasUnpushed(root) (bool, error)` reports `true` when `@{upstream}` cannot
  be resolved, else `git rev-list --count @{upstream}..HEAD > 0`.
  `OnDefaultBranch(root) bool` is `main`/`master`, case-insensitive.
  `(*Project).PushEnabled` reads the `push` scalar (`on`/`true`/`yes`) beside
  `GuardEnabled`. Tests: `TestCommitAll`, `TestHasUnpushed`,
  `TestOnDefaultBranch`, `TestPushEnabled`, over `gitRepo` (a local bare
  `origin` under `t.TempDir()`, no network).
  Contract correction: decision 2 wrote that `HasUnpushed` is `false` when
  the branch has no upstream. That contradicts AC1, which requires a brand
  new branch to be pushed with its upstream set: the "nothing to push" path
  would swallow the first push. Shipped `true` for no upstream, so a new
  branch publishes and an in-sync branch reports `false`; the test and this
  note carry the correction.
  Verified: `go test ./...` all `ok`; `gofmt -l .` empty; `go vet ./...`
  clean; `TestCommitAll|TestHasUnpushed|TestOnDefaultBranch|TestPushEnabled`
  pass with `-v`.
- [x] Phase 2 — `forge push` and the checkpoint. Moves: AC1, AC2, AC3.
  Where: `internal/cli/push.go` (new), `internal/cli/cli.go`; tests in
  `internal/cli/cli_test.go`.
  Landed: `cmdPush(args, out) error` loads through `specArg`, refuses
  `project.OnDefaultBranch` first (decision 4) and a detached HEAD, then runs
  `checkpoint`. `checkpoint` builds `chore(<ID>): checkpoint <state>` through
  `checkpointMessage`, commits with `project.CommitAll`, and pushes with
  `project.Push` unless `!committed && !HasUnpushed`, when it prints
  `nothing to push: <branch> is up to date` (decision 5). Dispatched as
  `case "push"` and added to the usage text.
  Tests: `TestPush_CommitsAndPushes` (subject, clean tree, upstream set,
  remote holds HEAD), `TestCommitMessage_NamesSpecAndState`,
  `TestPush_RefusesDefaultBranch` (no commit on `main`),
  `TestPush_NothingToPush`, over the `checkpointRepo` local-bare-remote
  fixture.
  Verified: `go test ./...` all `ok`; `gofmt -l .` empty; `go vet ./...`
  clean; the four tests pass with `-v`.
- [x] Phase 3 — `forge advance` checkpoints when opted in. Moves: AC4.
  Where: `internal/cli/work.go` (`cmdAdvance`); tests in
  `internal/cli/cli_test.go`.
  Landed: `cmdAdvance` now keeps `p` from `specArg` and, after `s.Save()`,
  calls `checkpoint` only when `p.PushEnabled()`; a returned error becomes a
  `warning: ...` line on `out` and the command still exits 0 (decision 7).
  `checkpoint` absorbed the default-branch and detached-HEAD refusals, so no
  caller — `forge push` or the opt-in advance — can publish `main`. The
  `--to accepted`/`--to planning` routes keep their early return and do not
  checkpoint.
  Tests: `TestAdvance_CheckpointsWhenOptedIn` (commit, clean tree, branch on
  the remote), `TestAdvance_OfflineWithoutOptIn` (HEAD and remote untouched),
  `TestAdvance_PushFailureIsAWarning` (exit 0, `warning:`, state stands).
  Verified: `go test ./...` all `ok`; `gofmt -l .` empty; `go vet ./...`
  clean; the three tests pass with `-v`.
- [x] Phase 4 — `forge submit` commits before it pushes. Moves: AC1.
  Where: `internal/cli/report.go` (`cmdSubmit`); tests in
  `internal/cli/submit_internal_test.go` (new).
  Landed: `cmdSubmit` commits the pending work with `project.CommitAll` and
  `checkpointMessage(s)` after the `--dry-run`/no-gh early return and before
  its existing push and `gh pr create` (decision 8). To keep the unit test
  off the network, `report.go` gains the seams `hasGH`, `pushBranch` and
  `ghRun` (the `upgrade.go` pattern); `cmdSync` still calls `project`
  directly.
  Test: `TestSubmit_CommitsBeforePush` stubs the three seams, runs
  `cmdSubmit` on a spec branch with a pending file, and asserts the commit
  subject, a clean tree at push time, and the pushed branch.
  Verified: `go test ./...` all `ok`; `gofmt -l .` empty; `go vet ./...`
  clean; `TestSubmit_CommitsBeforePush` passes with `-v`.
- [x] Phase 5 — Docs, roles and changelog. Moves: AC5. Where:
  `docs/cli.md`, `docs/workflow.md`, `kit/machine/roles/implementer.md`,
  `kit/machine/roles/orchestrator.md`, `CHANGELOG.md`; tests in
  `internal/cli/cli_test.go`, `internal/cli/machine_test.go`.
  Landed: `docs/cli.md` gains a `forge push` section (commit subject,
  default-branch refusal, the `nothing to push` line), names the `push: on`
  opt-in in `forge advance`, and says `forge submit` commits pending work
  first; the network list in the intro names `forge push`/`forge submit`.
  `docs/workflow.md` gains a `## Checkpoints` section. The implementer role
  runs `forge push` after a phase; the orchestrator role names the `push: on`
  opt-in. `CHANGELOG.md` adds the `[Unreleased]` entry.
  Tests: `TestDocPages_DocumentTheCheckpoint` (scopes the `forge push` and
  `forge advance` sections, checks `docs/workflow.md`),
  `TestRoles_DocumentTheCheckpoint` (the two role printouts).
  Verified: `go test ./...` all `ok`; `gofmt -l .` empty; `go vet ./...`
  clean; both tests pass with `-v`.

## Proposed conventions

None.
