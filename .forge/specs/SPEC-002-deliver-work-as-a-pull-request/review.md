# Review — SPEC-002

Verdict: pass

## Acceptance criteria

| Criterion | Result | Evidence |
|---|---|---|
| AC1: work ends as a pull request; nothing in Forge merges it | pass | `cmdSubmit` only pushes and creates a PR; the kit, roles and docs state that a person merges. |
| AC2: Forge opens the PR through `gh`; without `gh` it prints the commands | pass | `cmdSubmit` and `submitMessage`; `TestSubmit_DryRunPrintsCommands`; the command opened this spec's own PR. |
| AC3: the recorded actor defaults to the gh login, falling back to git | pass | `resolveActor` prefers `--by`, then `project.GHUser`, then git; `forge accept` without `--by` wrote `accepted_by: TheJisus28`. |
| AC4: `go test ./...` passes | pass | All packages `ok`; `go vet` and `gofmt` clean. |

## Blocking problems

None.

## Notes

- `GHUser` runs `gh api user` only when there is no `--by`; it returns empty on
  failure, so `accept`, `start` and `approve` keep working offline.

## Proposed conventions

None.
