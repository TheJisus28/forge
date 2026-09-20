# Plan — SPEC-020

The survey lives once, in `spec.md`'s `## Existing state` (the architect
wrote it). Phases below reuse what it names instead of rebuilding it.

## Phase 1 — The git helpers and the opt-in flag

- Scope: `internal/project/git.go` gains `CommitAll(root, message) (bool,
  error)` (`git add -A` then `git commit`; false on a clean tree),
  `HasUnpushed(root) (bool, error)` (false with no upstream; otherwise
  `git rev-list --count @{upstream}..HEAD` more than zero) and
  `OnDefaultBranch(root) bool`. `internal/project/project.go` gains
  `func (p *Project) PushEnabled() bool`, reading the `push` scalar beside
  `GuardEnabled`. Tests in `internal/project/project_test.go` (or a new
  `git_test.go`): a clean tree commits nothing, a dirty one commits, an
  unpushed branch reports true and an up-to-date one false, `main`/`master`
  are default and a `spec/...` branch is not, and `push: on`/`true`/`yes`
  enable with absent/off disabled.
- Done when: the helpers and `PushEnabled` behave as decision 2 and 6 say,
  against a temporary repository with a local bare remote. Moves AC1, AC4.
- Verify with: `go test ./...`, `gofmt -l .`, `go vet ./...`.

## Phase 2 — `forge push` and the checkpoint

- Scope: new `internal/cli/push.go` with `cmdPush(args []string, out
  io.Writer) error` and `checkpoint(p *project.Project, s *project.Spec,
  out io.Writer) error` (commit then push, one report line). `cmdPush`
  loads through `specArg`, refuses the default branch first (decision 4),
  builds the subject `chore(<ID>): checkpoint <state>` (decision 3), runs
  the checkpoint, and prints `nothing to push: <branch> is up to date` when
  there is nothing to do (decision 5). Dispatched from `internal/cli/cli.go`
  beside `submit` and added to the usage text. Tests in
  `internal/cli/cli_test.go`: `TestPush_CommitsAndPushes`,
  `TestPush_RefusesDefaultBranch`, `TestPush_NothingToPush`,
  `TestCommitMessage_NamesSpecAndState`.
- Done when: on a spec branch with a changed file, `forge push SPEC-001`
  commits the message, pushes and sets the upstream; on `main` it exits
  non-zero without committing; clean and up to date it exits 0 with the
  line. Moves AC1, AC2, AC3.
- Verify with: `go test ./...`, `gofmt -l .`, `go vet ./...`.

## Phase 3 — `forge advance` checkpoints when opted in

- Scope: `internal/cli/work.go:cmdAdvance` calls `checkpoint` after
  `s.Save()` when `p.PushEnabled()`, and prints its result; without the
  scalar it runs no `git` and no `gh` (decision 6). A failed checkpoint is a
  `warning: ` line on `out` and exit 0, the state move having already been
  saved (decision 7). The `--to accepted`/`--to planning` routes keep their
  own bookkeeping and do not auto-checkpoint. Tests:
  `TestAdvance_CheckpointsWhenOptedIn`, `TestAdvance_PushFailureIsAWarning`.
- Done when: with `push: on` and a bare remote, `forge advance SPEC-001
  --to implementing` commits and pushes; without the scalar, the tree and
  the remote are untouched; a failing push still leaves the new state and
  exits 0. Moves AC4.
- Verify with: `go test ./...`, `gofmt -l .`, `go vet ./...`.

## Phase 4 — `forge submit` commits before it pushes

- Scope: `internal/cli/report.go:cmdSubmit` calls `CommitAll` (skipping a
  clean tree) before its existing `project.Push` and `gh pr create`
  (decision 8). Test: `TestSubmit_CommitsBeforePush`, with the `gh` seam
  faked as the suite already does.
- Done when: a dirty tree is committed before the PR step and a clean one is
  left alone. Moves AC1.
- Verify with: `go test ./...`, `gofmt -l .`, `go vet ./...`.

## Phase 5 — Docs, roles and changelog

- Scope: `docs/cli.md` documents `forge push` and the `push:` opt-in;
  `docs/workflow.md` names the checkpoint and when to run it;
  `kit/machine/roles/implementer.md` runs `forge push` after each phase,
  and `kit/machine/roles/orchestrator.md` names the opt-in on `forge
  advance`; `CHANGELOG.md` gains the entry under `[Unreleased]` (AGENTS.md
  asks for it in the same pull request). Tests in
  `internal/cli/cli_test.go` and `internal/cli/machine_test.go` scope to the
  command's section and assert stable anchors (extending the SPEC-021 docs
  test). Moves AC5.
- Done when: the docs name `forge push` and the opt-in, the roles tell the
  implementer to checkpoint after a phase, and the changelog describes the
  change. Moves AC5.
- Verify with: `go test ./...`, `gofmt -l .`, `go vet ./...`,
  `go run . template tasks`.

## Risks

- Tests must not touch the network: the push helpers are exercised against a
  local bare repository (`git init --bare` under `t.TempDir()`), never
  `origin` on a host.
- `git add -A` stages the whole tree (decision 2). A spec branch is
  single-purpose, so this matches the archive instruction; a stray file is
  the user's to avoid. Do not add per-file staging here.
- Auto-push only after a state move succeeds; the state move must never be
  rolled back by a failed push (decision 7).
- The guard blocks raw `git push` on `main`; `cmdPush` repeats the refusal
  so the guard never has to see inside Forge. Do not weaken either.
- Scope creep: no new state, no merge, no background pusher, no change to
  `forge sync`.
