# Review — SPEC-023

Verdict: pass

Reviewed by the `reviewer` role on
`intake/spec-023-allocate-spec-ids`, against the approved contract
`bc039edb2034`. This is a read-only review: no `status` change, no archive and
no product code touched. Merge base
`ab516ea` (`main`).

## Acceptance criteria

| Criterion | Result | Evidence |
|---|---|---|
| AC1 | pass | `go test ./internal/cli/ -run TestNew_SkipsIdsOnOtherBranches -v` -> `--- PASS`. `internal/cli/cli_test.go:500` publishes `SPEC-030-elsewhere` on a `spec/*` ref and `SPEC-031-elsewhere` on an `intake/*` ref to a local bare `origin`, then `forge new "Local"` mints `SPEC-032-local`, so both refs are read and the taken number is skipped. `cmdNew` (`internal/cli/work.go:50`) numbers above `project.RemoteSpecRefs`; `freeNewNum`/`dirTaken` (`work.go:442,454`) skip a candidate held on a different folder. |
| AC2 | pass | `go test ./internal/cli/ -run TestNew_NoRemoteBranchesStaysOffline -v` -> `--- PASS`. `cli_test.go:516` sets `fetch: on`, leaves the caller with no remote-tracking refs (a second repository pushes to `origin`), snapshots `git for-each-ref refs/remotes/` before and after and checks `FETCH_HEAD` absent; `forge new` numbers from the local tree (`SPEC-001`) and nothing is fetched. The same test then runs `forge renumber SPEC-001` -> `SPEC-006`, proving `SharedSpecRefs` (`internal/project/git.go:133`) falls back to the local `main` ref (`SPEC-005`). `RemoteRefs`/`RemoteSpecRefs` (`git.go:44,124`) read only refs already present; `work.go` has no fetch path. |
| AC3 | pass | `go test ./internal/project/ -run TestRemoteSpecIDs_ScansEveryBranch -v` -> `--- PASS`. `internal/project/project_test.go:826` pushes `SPEC-010` on `main`, `SPEC-011` on `spec/*` and `SPEC-012` on `intake/*` to a local bare `origin` and asserts `SharedSpecRefs` returns all three folders. `RemoteRefs` is one `git for-each-ref refs/remotes/` (`git.go:44`) and `RemoteSpecDirs` one `git ls-tree` per ref (`git.go:63`); no network API is imported (`kit/nonet_test.go` passes). |
| AC4 | pass | `go test ./internal/cli/ -run 'TestRenumber_ToRefusesTakenOnAnotherFolder\|TestAccept_RenumbersWhenTakenOnMain\|TestRenumber_ResolvesTheRace' -v` -> all `--- PASS`. `cli_test.go:576` refuses `forge renumber SPEC-031 --to 30` because `SPEC-030-elsewhere` is on a `spec/*` ref, and `SPEC-031-local` stays; the guard is `dirTaken` at `work.go:391`. The existing `TestAccept_RenumbersWhenTakenOnMain` (`cli_test.go:327`) and `TestRenumber_ResolvesTheRace` (`cli_test.go:413`) pass unchanged, and `forge validate` still reports a real duplicate (`internal/validate/validate.go:58`). |
| AC5 | pass | `go test ./internal/cli/ -run 'TestDocs_DescribeTheWiderIdRead\|TestNew_DescribesOneGate' -v` -> both `--- PASS`. `cli_test.go:1516` reads `docs/cli.md` and `docs/workflow.md` and asserts `every remote-tracking ref`, `best-effort`, `` `git fetch` `` and `same title` in both. `docs/cli.md` (`forge new`, `forge accept`, `forge renumber`) and `docs/workflow.md` (`## Ids across branches`) carry the wider, best-effort read, both residual windows, that no command fetches, and the `git fetch` before `forge new` step. |
| AC6 | pass | `go test ./...` -> every package `ok` (root, `internal/cli`, `internal/doc`, `internal/project`, `internal/validate`, `internal/view`, `internal/workflow`, `kit`), exit 0. `gofmt -l .` -> no output. `go vet ./...` -> no output. |
| AC7 | pass | `go test ./internal/cli/ -run TestAccept_KeepsIdForItsOwnPublishedBranch -v` -> `--- PASS`. `cli_test.go:556` pushes the spec's own folder on `origin/spec/001-own-branch`, then `forge accept SPEC-001` keeps the id: the output has no `renumbered` and no `SPEC-002`, and the folder stays `SPEC-001-own-branch`. `idTaken` (`work.go:472`) compares `filepath.Base(s.Dir())`, so a ref with the same folder is this spec, not a collision. |

## Contract decisions

- Decision 1: `SpecRef{ID, Dir}` (`git.go:36`); `idTaken` (`work.go:472`) and
  `dirTaken` (`work.go:454`) collide only on the same id with a different
  `Dir`; `cmdNew` (`work.go:50`) checks the candidate `SpecDirName`.
- Decision 2: `RemoteRefs` lists every `refs/remotes/...` with one
  `for-each-ref` (`git.go:44`); `collectSpecRefs` (`git.go:143`) unions the
  folders across all of them, so `spec/*` and `intake/*` both count.
- Decision 3: no caller of `SharedSpecRefs`/`RemoteSpecRefs` runs `git fetch`
  or `gh`; the only reads are `git for-each-ref` and `git ls-tree`
  (`git.go:44,63`). `TestNew_NoRemoteBranchesStaysOffline` proves it with the
  ref/FETCH_HEAD snapshot.
- Decision 4/5: no age or merged filter; every remote-tracking ref is scanned.
  The two residual windows are documented, not prevented.
- Decision 6: `forge new` and `forge accept` use the wider set;
  `forge renumber` uses it for the default and for `--to` (`work.go:388,391`);
  `forge validate`/`forge check` are unchanged and stay local.
- Decision 7: no cap and no network call per ref.
- Decision 8: `SharedSpecRefs` (`git.go:133`) falls back to `main` when no
  remote ref carries a spec; `RemoteSpecRefs` (`git.go:124`) deliberately does
  not (see note 1).
- Decision 9: `forge new` fixes the folder once with `SpecDirName`; the
  collision compares `filepath.Base(s.Dir())`, never a title recomputed from
  frontmatter.

## Blocking problems

None. `go test ./...`, `gofmt -l .` and `go vet ./...` are clean, and every
criterion has a passing named test or command.

## Notes

1. **`forge new` reads remote refs only, not the `main` fallback.** The
   approved Interfaces say `cmdNew` calls `SharedSpecRefs`, but
   `SharedSpecRefs` includes the local `main` fallback (decision 8), and that
   would make `forge new` skip the provisional number in
   `TestAccept_RenumbersWhenTakenOnMain` (`cli_test.go:327`), which AC4
   requires to pass unchanged. `cmdNew` therefore calls `RemoteSpecRefs`
   (remote refs only) while `cmdAccept`/`cmdRenumber` use `SharedSpecRefs`.
   The implementation follows the criteria and decisions 1/8; the Interfaces
   line is the one place it reads narrower.
2. **`docs/cli.md` says "whatever the branch is named", not `intake/*`.**
   `TestNew_DescribesOneGate` (`cli_test.go` around line 303, SPEC-015
   decision 3) forbids the lowercase word `intake` on `docs/cli.md`, so the
   branch glob is described generically. The code still reads `intake/*`
   refs, and `TestNew_SkipsIdsOnOtherBranches` uses one.
3. The residual same-title window (decision 4) is real and documented: two
   parallel specs with the same title and number share a folder, are not a
   collision, and surface only as a git conflict at merge. Nothing in this
   spec claims otherwise.

## Proposed conventions

None.
