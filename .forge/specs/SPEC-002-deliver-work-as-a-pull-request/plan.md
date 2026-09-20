# Plan — SPEC-002

## Existing state

- `internal/project/git.go` owned every git/gh shell-out: `Branch`, `Fetch`,
  `UserName`, `HasGH`, `GH` and the `run` helper.
- `internal/cli/work.go`: `resolveActor`, used by `accept` and `approve`;
  `start` read `project.UserName` directly instead.
- `internal/cli/report.go`: `cmdSync` wrote the `pr`, `pr_url` and `pr_state`
  fields on a spec; `submit` reuses them.
- Delivered specs: SPEC-001.

## Phase 1 — Resolve the actor from gh, add the git helpers

- Scope: `project.GHUser`, `project.Push`; `resolveActor` prefers `--by`, then
  the gh login, then git; `forge start` uses `resolveActor`.
- Verify with: `go test ./...`.

## Phase 2 — `forge submit`

- Scope: `forge submit [id] [--base main] [--dry-run]` pushes the branch and
  opens the PR through `gh`, records it, and never merges; without `gh` it
  prints the commands.
- Verify with: `go test ./...`.

## Phase 3 — Kit and docs

- Scope: state the pull request as the end of the loop and `forge submit` as
  the command that opens it.
- Verify with: `forge validate`.

## Risks

- `GHUser` runs `gh api user` when there is no `--by`; it degrades to git and
  to empty on failure.
