# Review - SPEC-014 Scope the guard command check to the git command

Verdict: **pass with notes**

Reviewed on branch `spec/014-scope-the-guard-command-check-to-the-git-command`
against contract `cd828464703f`. Product code not touched by this review;
every acceptance criterion passes. One contract ambiguity is noted but does
not fail an AC.

## Criterion / evidence

| # | Criterion | Result | Evidence |
|---|---|---|---|
| AC1 | `git push -u origin my-branch && gh pr create --base main` is allowed; the later `main` does not make the push target the default branch | pass | `internal/cli/cli_test.go:TestGuardCommand` lists the line in its allow set and passes. By hand in a scratch repo outside this one, built from this branch, current branch `feature`: the binary's `guard --command "<line>" --explain` printed `would allow`. `splitSegments`/`gitCommand` judge each segment on its own, so `gh pr create --base main` is not read as a push. |
| AC2 | `git push origin main` and `git push origin HEAD:main` are denied | pass | `TestGuardCommand` deny set (also `HEAD:master`, `main:feature`, `git -C . push origin master`) passes. By hand in the scratch repo, `would deny` for `git push origin main`, `HEAD:main`, `HEAD:master`, `+main`, `--force origin main`, `-d origin main`, `:main`, `refs/heads/main`, `refs/heads/main:refs/heads/feature`, `"main"` and `'master'`. `refspecsToDefault`/`targetsDefault` split each refspec on `:` and strip `+`, `refs/heads/`, `heads/` and `refs/`. |
| AC3 | `git push` on `main`/`master` is denied, the same command on a feature branch is allowed | pass | `TestGuardCommand_BarePushFollowsTheBranch` builds a real repo, checks out `main`, asserts denial, then checks out `feature` and asserts allow; it passes. By hand in the scratch repo: on `main` the binary printed `would deny`; on `feature`, `would allow`. `onDefaultBranch` feeds the unchanged `def` check. |
| AC4 | The decision looks only at the `git push`/`git merge` command, not at later `;`, `&&`, `||` or `|` segments | pass | `TestGuardCommand` allows `git push -u origin my-branch ; echo main`, `... || gh pr create --base main` and the AC1 `&&` line. By hand: `git status && git push -u origin feat` allowed, while `git push -u origin feat && git push origin main` denied, so the second segment is still judged. Newline, named in decision 1, also holds: `git status` + newline + `git push origin main` denied, `git push -u origin my-branch` + newline + `echo main` allowed. |
| AC5 | `forge guard --command "..." --explain` prints `would allow`/`would deny` consistent with AC1-AC4 | pass | `TestGuardCommand` asserts `would allow` for the compound line and `would deny` for `git push origin main`. `cmdGuard` prints both from the one `commandDenial` result (decision 4), so there is no second path. By hand, every AC1-AC4 case returned the matching `would allow` / `would deny: ...` line, exit code 0. |

Suite: `go test ./...` all `ok` (`TestGuardCommand`,
`TestGuardCommand_BarePushFollowsTheBranch` included); `gofmt -l .` empty;
`go vet ./...` clean.

## Contract integrity

`commandDenial(p, command)` keeps its signature and is the only entry point
used by the hook and by `--explain` (`internal/cli/guard.go:79`). No command
or flag was added. The whole-line `reGitPush`, `reGitMerge` and
`pushesToDefault` are gone; `reGHMerge` and `onDefaultBranch` remain, as
decision 3 promised. A repository-wide `grep` finds the removed names only
in the spec folder, never in code.

## Problems

Blocking the merge: none.

Non-blocking notes:

- **Contract ambiguity in decision 2.** `commandDenial` uses
  `def || refspecsToDefault(args)` for every push, so while the current
  branch is `main` or `master` even an explicit feature refspec is denied. In
  the scratch repo on `main`, `git push -u origin feat` and
  `git push origin feature/main` both printed `would deny`. Decision 2's
  example reads as if `git push -u origin my-branch` is never denied, while
  the next sentence ("a bare `git push` is denied only while the current
  branch is `main` or `master`") and the parenthetical "(unchanged, AC3)" read
  as if the branch check is preserved for all pushes. The implementation
  keeps the older behaviour, which is pre-existing, so no AC fails; the team
  should record which reading is the contract, or split it into a follow-up.
- The guard stays a heuristic by design (contract "Out of scope"). With
  `git` not the first word it does not see the push: `sudo git push origin
  main` printed `would allow`, as did `echo "git push origin main"` and
  `git commit -m "git push origin main"`. These are the boundaries the
  contract chose, not defects.
- `go run . validate` on this branch reports an unrelated
  `SPEC-019: the contract changed after TheJisus28 approved it` error. The
  SPEC-014 diff touches only `internal/cli/guard.go`, `internal/cli/cli_test.go`
  and its own spec folder, so this is not caused here and does not affect
  the criteria. `validate` also asks for this `review.md`, which this file
  supplies; the SPEC-019 error remains for its own owner.

## Proposed conventions

None.