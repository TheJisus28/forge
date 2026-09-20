# Review — SPEC-020

Verdict: pass with notes

## Acceptance criteria

- AC1: pass — `forge push SPEC-001` commits the pending tree and publishes
  the branch. `internal/cli/push.go:checkpoint` commits through
  `project.CommitAll` with `checkpointMessage(s)` and pushes with
  `project.Push`. `TestPush_CommitsAndPushes` (subject
  `chore(SPEC-001): checkpoint proposed`, clean tree, upstream
  `origin/spec/001-something`, remote holds HEAD) and
  `TestCommitMessage_NamesSpecAndState` (`chore(SPEC-001): checkpoint
  implementing`) pass; `TestSubmit_CommitsBeforePush` covers the submit
  step.
- AC2: pass — `checkpoint` refuses the default branch before it commits and
  `cmdPush` returns that error. `TestPush_RefusesDefaultBranch` exits
  non-zero on `main` with `HEAD` unchanged;
  `TestOnDefaultBranch` covers `main`/`master`/`MAIN` and a `spec/...`
  branch.
- AC3: pass — with nothing to commit and the branch in sync, `checkpoint`
  prints `nothing to push: <branch> is up to date` and exits 0.
  `TestPush_NothingToPush` passes.
- AC4: pass — `cmdAdvance` calls `checkpoint` only when `p.PushEnabled()`;
  `TestAdvance_CheckpointsWhenOptedIn` (commit, clean tree, branch on the
  remote) and `TestAdvance_OfflineWithoutOptIn` (HEAD and remote untouched)
  pass. A failing push is a `warning:` and the move stands:
  `TestAdvance_PushFailureIsAWarning`. `TestPushEnabled` covers the scalar.
- AC5: pass — `docs/cli.md` has a `forge push` section and the `push: on`
  opt-in under `forge advance`; `docs/workflow.md` has `## Checkpoints`; the
  implementer and orchestrator roles name it.
  `TestDocPages_DocumentTheCheckpoint` and
  `TestRoles_DocumentTheCheckpoint` pass.
- AC6: pass — `go test ./...` reports every package `ok`; `gofmt -l .` is
  empty and `go vet ./...` is clean.

## Evidence

- `go test ./...` — all `ok`.
- `go test ./internal/project/ -run "TestCommitAll|TestHasUnpushed|TestOnDefaultBranch|TestPushEnabled" -v` — four PASS.
- `go test ./internal/cli/ -run "TestPush_|TestCommitMessage_NamesSpecAndState|TestAdvance_|TestSubmit_CommitsBeforePush|TestDocPages_DocumentTheCheckpoint|TestRoles_DocumentTheCheckpoint" -v` — ten PASS.
- `go run . check SPEC-020` — reports the six criteria with no evidence
  before this file exists (exit 1); `forge validate` raises the same rows,
  which is the coverage rule working, not a defect.
- The push tests use a local bare `origin` under `t.TempDir()`, so no test
  touches the network (the plan's Risks note).

## Problems

Blocking: none.

Non-blocking:

1. The contract's decision 2 said `HasUnpushed` is false with no upstream;
   the shipped helper returns true, so a brand new branch is published
   rather than reported as up to date. The correction is recorded in
   `tasks.md`; the acceptance criteria, not the parenthetical, are what this
   review verifies, and AC1 needs the true behaviour.
2. The commit subject names the workflow state as "the phase" (decision 3).
   The plan's numbered phases are not machine-readable, so this is the
   closest honest name; a future spec that wants the phase number must give
   `tasks.md` a marker.

## Contract drift

The contract was approved at `4b53e1926bf1` and the delivered interfaces match
decision 8 and the `### Interfaces` list: `cmdPush`/`checkpoint` in
`internal/cli/push.go`, `CommitAll`/`HasUnpushed`/`OnDefaultBranch` in
`internal/project/git.go`, `(*Project).PushEnabled`, and the `cmdAdvance` and
`cmdSubmit` wiring. The only drift is the decision 2 wording recorded above.

## Proposed conventions

None.
