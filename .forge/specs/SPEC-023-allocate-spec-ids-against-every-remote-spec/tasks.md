# Tasks — SPEC-023

The phases from the plan, as checkboxes. One phase is one implementer run
and one commit. Tick a phase when it lands and say where the work is, so a
later spec knows what exists without reading the diff. Name the criteria the
phase delivers (`Moves: <criterion ids>`), so every criterion is traceable to
the task that moves it; `forge check` reads these ids from this file.

- [x] Phase 1 — The read layer over every remote ref. Moves: AC3.
  Where: `internal/project/git.go` (`SpecRef`, `RemoteRefs`,
  `RemoteSpecDirs`, `RemoteSpecIDs`, `RemoteSpecRefs`, `SharedSpecRefs`,
  `collectSpecRefs`, `dirID`); tests in `internal/project/project_test.go`
  (`TestRemoteSpecIDs_ScansEveryBranch`, and `TestRemoteSpecIDs_ReadsTheRef`
  unchanged).
  Landed: `RemoteRefs` lists `refs/remotes/` with one `git for-each-ref`;
  `RemoteSpecDirs` parses the `SPEC-NNN-slug` folders on a ref, deduped and
  sorted; `RemoteSpecIDs` is now the id view over it, so the old test passes
  unchanged; `RemoteSpecRefs` is the union over every remote ref and
  `SharedSpecRefs` adds the local `main` fallback when no remote ref carries
  a spec. Nothing is fetched.
  Verified: `go test ./internal/project/ -run TestRemoteSpecIDs -v` passes
  both tests; `go test ./...`, `gofmt -l .` and `go vet ./...` clean.
- [x] Phase 2 — The id commands use the wider, folder-aware set. Moves: AC1,
  AC2, AC4, AC7. Where: `internal/cli/work.go` (`cmdNew`, `cmdAccept`,
  `cmdRenumber`, `freeNewNum`, `dirTaken`, `idTaken`, `nextFreeNum`;
  `sharedSpecIDs` removed); tests in `internal/cli/cli_test.go`
  (`bareOriginRepo`, `publishSpec`, `pushSpecElsewhere` fixtures).
  Landed: `forge new` numbers above `project.RemoteSpecRefs` and skips a
  number a remote branch holds on a different folder; `forge accept` keeps
  the id when the only match is the spec's own folder (`idTaken` compares
  `filepath.Base(s.Dir())`) and renumbers otherwise; `forge renumber --to`
  refuses a target another folder holds on a shared ref. No command fetches.
  Tests: `TestNew_SkipsIdsOnOtherBranches`, `TestNew_NoRemoteBranchesStaysOffline`,
  `TestAccept_KeepsIdForItsOwnPublishedBranch`,
  `TestRenumber_ToRefusesTakenOnAnotherFolder`, plus the existing
  `TestAccept_RenumbersWhenTakenOnMain`, `TestAccept_RefusesWhenReferenced`
  and `TestRenumber_ResolvesTheRace` unchanged.
  Verified: the seven tests pass with `-v`; `go test ./...`, `gofmt -l .` and
  `go vet ./...` clean.
- [x] Phase 3 — Docs. Moves: AC5, AC6.
  Where: `docs/cli.md` (`forge new`, `forge accept`, `forge renumber`),
  `docs/workflow.md` (`## Ids across branches`); test in
  `internal/cli/cli_test.go` (`TestDocs_DescribeTheWiderIdRead`).
  Landed: both pages state the wider, best-effort source of truth over every
  remote-tracking ref, name both residual windows (unfetched refs / two
  branches before either pushes; two specs with the same title share a
  folder), say no command fetches, and point at `git fetch` before
  `forge new` in a shared repository.
  Verified: `TestDocs_DescribeTheWiderIdRead` and `TestNew_DescribesOneGate`
  pass; `go test ./...`, `gofmt -l .` and `go vet ./...` clean.

## Proposed conventions

None.
