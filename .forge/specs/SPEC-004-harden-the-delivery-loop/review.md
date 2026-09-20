# Review — SPEC-004

Verdict: pass

## Acceptance criteria

| Criterion | Result | Evidence |
|---|---|---|
| AC1: `forge guard` denies `gh pr merge` and pushes/merges to the default branch, through `--command` and the Bash hooks | pass | `commandDenial` in `internal/cli/guard.go`; `TestGuardCommand`; by hand `forge guard --command "git push origin main"` exits 1 and `--command "git push -u origin spec/004-x"` exits 0. The Claude matcher is `Write|Edit|Bash` and the opencode plugin handles `bash`. |
| AC2: `## Open questions` gates `approve` and warns in `validate` | pass | `Spec.HasOpenQuestions`, the `cmdApprove` refusal and the validate warning; `TestApprove_RefusesOpenQuestions`, `TestRun_WarnsOpenQuestions`; the template includes the section. |
| AC3: `brief` and `status <id>` show `tasks done/total` | pass | `Spec.TaskProgress`; `TestStatusShowsTaskProgress`; by hand `forge brief` printed `tasks 0/4`. |
| AC4: `forge board`/`BOARD.md` are gone and tests pass | pass | `view.Board`, `cmdBoard`, the command and `ensureGitignore` removed; `go test ./...`, `go vet`, `gofmt` and `forge validate` clean. |

## Blocking problems

None.

## Notes

- The guard only refuses merges and pushes to the default branch; commits and
  spec-branch pushes stay allowed. `guard: off` remains the escape hatch.
- The open-questions check counts list items, so the template's explanatory
  prose does not read as an open question.
- Removing the board also removed the `.gitignore` writing; `forge init` no
  longer touches `.gitignore`.

## Proposed conventions

None.
