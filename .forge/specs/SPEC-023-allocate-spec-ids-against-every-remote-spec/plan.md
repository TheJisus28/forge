# Plan — SPEC-023

The survey lives once, in `spec.md`'s `## Existing state` (the architect
wrote it). Phases below reuse what it names instead of rebuilding it.

## Phase 1 — The read layer over every remote ref

- Scope: `internal/project/git.go` gains `type SpecRef struct { ID, Dir
  string }`, `func RemoteRefs(root string) []string` (one `git for-each-ref
  --format=%(refname) refs/remotes/`) and `func RemoteSpecDirs(root, ref
  string) []string` (the `SPEC-NNN-slug` folders, deduped and sorted).
  `func RemoteSpecIDs(root, ref string) []string` is rebuilt as the id view
  over `RemoteSpecDirs`, so `TestRemoteSpecIDs_ReadsTheRef` passes unchanged.
  `func SharedSpecRefs(root string) []SpecRef` unions the dirs over
  `RemoteRefs`, dedupes by `Dir`, and falls back to the local `main` ref when
  no remote-tracking ref carries a spec (decision 2, 8). Tests:
  `internal/project/project_test.go:TestRemoteSpecIDs_ScansEveryBranch` over
  a local bare `origin` (no network) plus the existing
  `TestRemoteSpecIDs_ReadsTheRef`. Moves AC3.
- Done when: `SharedSpecRefs` returns the union of the folders on a `spec/*`
  and an `intake/*` branch, deduped, and a missing ref yields nothing without
  error or fetch.
- Verify with: `go test ./...`, `gofmt -l .`, `go vet ./...`.

## Phase 2 — The id commands use the wider, folder-aware set

- Scope: `internal/cli/work.go` removes `sharedSpecIDs` and reads
  `project.SharedSpecRefs` in `cmdNew`, `cmdAccept` and `cmdRenumber`.
  `cmdNew` seeds the number from `nextFreeNum(p, shared)` and confirms the
  candidate `project.SpecDirName(id, title)` with a folder-aware check.
  `idTaken(p, s, shared []project.SpecRef)` treats a ref as a collision only
  when `ref.ID == s.ID && ref.Dir != filepath.Base(s.Dir())`, so the spec's
  own published branch never renumbers it (decision 1, 9). `nextFreeNum`
  takes `[]project.SpecRef`. `cmdRenumber --to <n>` refuses `n` when a shared
  ref holds it under a different folder (decision 6). Tests in
  `internal/cli/cli_test.go`: `TestNew_SkipsIdsOnOtherBranches` (AC1),
  `TestNew_NoRemoteBranchesStaysOffline` (AC2),
  `TestAccept_KeepsIdForItsOwnPublishedBranch` (AC7),
  `TestRenumber_ToRefusesTakenOnAnotherFolder` (AC4); the existing
  `TestAccept_RenumbersWhenTakenOnMain` and `TestRenumber_ResolvesTheRace`
  still pass (AC4). Moves AC1, AC2, AC4, AC7.
- Done when: `forge new` skips an id a `spec/*` or `intake/*` ref holds on a
  different folder; `forge accept` keeps the id of the spec's own published
  branch; `forge renumber --to` refuses a target another folder holds; no
  command fetches.
- Verify with: `go test ./...`, `gofmt -l .`, `go vet ./...`.

## Phase 3 — Docs

- Scope: `docs/cli.md` (`forge new`, `forge accept`, `forge renumber`)
  documents the wider source of truth, that it is best-effort, the two
  residual windows, that no command fetches, and the `git fetch` before
  `forge new` step for a shared repository. `docs/workflow.md` gains the same
  in the id/planning prose. Test
  `internal/cli/cli_test.go:TestDocs_DescribeTheWiderIdRead` reads both pages
  and asserts the statements, the way `TestDocs_DescribeCapability` reads
  `docs/cli.md`. Moves AC5, AC6.
- Done when: the docs agree with the binary, and `go test ./...`,
  `gofmt -l .` and `go vet ./...` are clean.
- Verify with: `go test ./...`, `gofmt -l .`, `go vet ./...`.

## Risks

- **No network in any test.** Every case runs against a local bare `origin`
  under `t.TempDir()`, as `gitRepo`/`checkpointRepo` already do;
  `nonet_test.go` stays intact. Never point a test at `github.com`.
- **`forge new`/`forge accept` must not fetch.** `SharedSpecRefs` only reads
  refs already present; a `fetch: on` project is freshened by the session
  start `forge brief`, not here (decision 3). AC2's snapshot is the guard.
- **The same folder is not a collision.** Compare `filepath.Base(s.Dir())`
  and `SpecDirName(id, title)`, never recompute from a possibly hand-edited
  `title`, or `forge accept` renumbers a spec against its own branch
  (decisions 1 and 9).
- **Do not remove the backstop.** `renumberSpec`, `referencedError` and
  `forge validate`'s duplicate check stay; the wider read only narrows the
  window (decision 4).
- Scope creep: no central allocator or reservation, no change to `fetch: on`
  or `forge brief`, no age/merge filter, no cap, no remote read in
  `forge validate`/`forge check`, no local-branch read beyond the `main`
  fallback.
