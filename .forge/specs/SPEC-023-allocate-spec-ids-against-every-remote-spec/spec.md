---
id: SPEC-023
title: Allocate spec ids against every remote spec branch, not just main
status: contracting
capability: specs
created: 2026-09-20
updated: 2026-09-20
accepted_by: TheJisus28
orchestrator: TheJisus28
---

## Problem

Two specs can be born at the same time on different branches and take the
same number, because nothing looks at the branches that have not merged yet.

`forge new` writes a provisional number from the local tree. `forge accept`
is the only place that confirms it, and it does so against the ids committed
on the shared branch — `origin/main`, then `main` (`internal/cli/work.go:139`
calls `sharedSpecIDs`, defined at `internal/cli/work.go:433`). A branch that
has not merged is invisible there. Two people who start specs in parallel
read the same `main`, both accept the same number, and the collision surfaces
only when the second branch merges, where `forge validate` fails with
`duplicate id, also in ...` (`internal/validate/validate.go:58`).

By then the fix is expensive. `forge renumber` moves the folder and rewrites
the id (`internal/cli/work.go:407`), but it refuses a spec that anything
already references (`referencedError`, `internal/cli/work.go:426`), so a
collision found late can block a branch whose children or dependents already
point at the id. `forge check` and CI catch the duplicate, but only after the
work is done. SPEC-022 made the session brief fetch, which makes the `main`
check trustworthy; it does not make in-flight branches visible, so the
parallel window stays open. A spec-driven workflow exists so several specs
can be in flight at once; the id allocation still assumes only one at a time
reaches `main`.

## Acceptance criteria

Observable outcomes. Someone else must be able to mark each one pass or
fail with evidence. A criterion is verifiable when it names the evidence that
settles it: a backticked command, a `test`/`TestName`, or an observable verb
such as `returns`/`refuses`. `forge approve` refuses while a criterion names
none of those, so write it before approval.

- AC1: `forge new` reads the folder names under `.forge/specs/` on every
  remote-tracking ref, not only `main`, so an id a remote branch already
  holds on a *different* folder is skipped when the provisional id is minted,
  and the `forge accept` confirmation reads the same wider set;
  `TestNew_SkipsIdsOnOtherBranches` mints a fresh number when a local bare
  `origin` carries that number on both a `spec/*` branch and an `intake/*`
  branch.
- AC2: The wider read uses only refs already present and never fetches, even
  when `.forge/project.md` sets `fetch: on`; with no remote-tracking refs it
  falls back to the local `main` ref and never fails, proven by a before/after
  snapshot test (`TestNew_NoRemoteBranchesStaysOffline`).
- AC3: The wider read covers every remote-tracking ref that carries
  `.forge/specs/` — `spec/*`, `intake/*` and any other branch name — and
  touches no network; `TestRemoteSpecIDs_ScansEveryBranch` settles it over a
  local bare remote.
- AC4: `forge renumber` can still resolve a duplicate that the wider read
  did not prevent, and `forge validate` keeps failing on a real duplicate, so
  the change reduces the window without removing the backstop; the existing
  `TestRenumber_ResolvesTheRace` and `TestAccept_RenumbersWhenTakenOnMain`
  still pass, and `forge renumber --to <n>` refuses `n` when a remote ref
  holds it under a different folder, settled by
  `TestRenumber_ToRefusesTakenOnAnotherFolder`.
- AC5: `docs/cli.md` and `docs/workflow.md` state the wider, best-effort
  source of truth; name both residual windows — two branches created before
  either pushes, and two specs with the same title minting the same folder
  (same id and same slug, so not a collision, and visible only as a git
  conflict at merge); say that `forge new`/`forge accept` never fetch; and
  tell a repository with more than one person to run `git fetch` before
  `forge new` so the read sees the latest branches.
- AC6: `go test ./...` passes, and `gofmt -l .` and `go vet ./...` report
  nothing.
- AC7: The same id on the same folder is the same spec, never a collision:
  `forge accept` keeps the id when the only reference to it is the spec's own
  published branch; `TestAccept_KeepsIdForItsOwnPublishedBranch` settles it.

## Open questions

None.

## Contract

Builds on SPEC-015 decision 7 (the id confirmation reads the shared branch;
`RemoteSpecIDs`, `renumberSpec`, `referencedError`) and SPEC-022 (`fetch: on`
freshens the remote refs at session start). Every name below is this
repository's.

**_1. A collision is the same id on a different folder; the same folder is the
same spec._** Forge compares folder names (`SPEC-NNN-slug`, the
`project.SpecDirName` shape), not numbers alone. A remote ref that carries the
spec's own folder is that spec — its own published branch — and never makes
`forge accept` renumber. Only the same id under a different folder is a
collision. `idTaken` takes the refs as `project.SpecRef` and treats a ref as a
collision only when `ref.ID == s.ID && ref.Dir != filepath.Base(s.Dir())`;
`cmdNew` compares the candidate `project.SpecDirName(id, title)` the same way.
Discards: comparing ids alone, which would renumber a spec against its own
pushed branch.

**_2. The wider read covers every remote-tracking ref that carries
`.forge/specs/`, across all configured remotes, not just `spec/*` on
`origin`._** `RemoteRefs(root)` runs one `git for-each-ref
--format=%(refname) refs/remotes/`; each ref that has `.forge/specs/`
contributes its folders through `RemoteSpecDirs(root, ref)`, so `spec/*`,
`intake/*` and any other branch name count, on any remote. A ref with no
`.forge/specs/` contributes nothing. Discards: globbing
`refs/remotes/origin/spec/*`, which misses `intake/*` — the branch this spec
itself travels on — and the current `sharedSpecIDs`, which reads `origin/main`
then `main` and no other ref.

**_3. `forge new` and `forge accept` never fetch; `fetch: on` is honoured at
session start, not in the id commands._** The read uses only refs already
present. A project that set `fetch: on` (SPEC-022) has fresh remote refs from
the `forge brief` fetch the session starts with, which runs before the id
commands; keeping the id commands offline keeps them fast and free of the
network failure modes SPEC-015 decision 7 left to the user. Discards: fetching
inside `forge new`/`forge accept` when `fetch: on`, a second fetch in the same
session, added latency, and a new network path in commands the no-network rule
kept clean. The docs point a repository with more than one person at `git
fetch` before `forge new`, as the manual replacement for the fetch these
commands deliberately do not make.

**_4. Best-effort, not a reservation._** Forge has no central allocator, so
"never collide" is not promised. The read keeps two windows open: branches
whose remote refs are absent locally (created after the last fetch) and two
branches created before either pushes; and, because a shared folder reads as
the same spec (decision 1), two specs with the same title that mint the same
number share a folder, are not flagged as a collision, and surface only as a
git conflict on that folder plus a `duplicate id` at merge. The docs say
exactly that, and no more.

**_5. Every matching ref is scanned, regardless of age, and merged branches
are not filtered out._** A merged branch's ids are already on the default
branch, so scanning it is redundant but harmless; an age or merged filter
would reopen the window, since an old branch can still merge. Gaps in the
sequence are accepted because numbering is max+1; an abandoned unmerged branch
pins its number only until its ref is deleted.

**_6. `forge new`, `forge accept` and `forge renumber` read the wider set;
`forge validate` and `forge check` stay local._** The wider read is where an id
is minted and confirmed; validate and check remain the local duplicate
backstop and must not depend on refs being fresh. `forge renumber --to <n>`
checks that explicit target against the same set and refuses `n` when a remote
ref holds it under a different folder, not only when it is taken locally.

**_7. No cap on the number of refs scanned._** The read is local plumbing (one
`for-each-ref` plus one `ls-tree` per ref), touches no network, and is
O(refs) like the other git reads here; a cap that silently dropped refs would
reopen the window. If a project ever has enough refs to matter, the answer is
an explicit warning and opt-out, deferred out of this spec.

**_8. With no remote-tracking refs the read falls back to the local `main`
ref, then to local numbering, and never fails._** `SharedSpecRefs` unions the
remote refs; when that union is empty it reads `main`; a missing ref yields no
ids rather than an error, so a repository with no remote keeps the SPEC-015
behaviour.

**_9. A published spec's slug cannot change, and the collision compares the
folder on disk, not one recomputed from the title._** `forge new` fixes the
folder once with `project.SpecDirName(id, title)`; no command renames an
existing folder except `forge renumber`, which changes the id at the same
time, so a spec's own published branch always carries the same `id-slug`. The
comparison uses `filepath.Base(s.Dir())` — the real folder — never
`project.SpecDirName(s.ID, s.Title)` recomputed from the current frontmatter,
so a hand-edited `title` that no longer matches the folder cannot make the
spec collide with its own pushed branch. `cmdNew`, which has no folder yet,
compares the candidate `project.SpecDirName(id, title)` it is about to create.
Discards: recomputing the folder from the current title, which would renumber a
spec against itself. A hand-renamed folder is outside the CLI-owned state and
out of scope.

### Interfaces

- `internal/project/git.go`:
  - `type SpecRef struct { ID, Dir string }` — one spec folder on a ref,
    `Dir` being the `SPEC-NNN-slug` basename.
  - `func RemoteRefs(root string) []string` — every `refs/remotes/...` name,
    from one `git for-each-ref --format=%(refname) refs/remotes/`; empty when
    git fails or there is no remote.
  - `func RemoteSpecDirs(root, ref string) []string` — the `SPEC-NNN-slug`
    folders under `.forge/specs/` on `ref`, deduped and sorted; the parser
    `RemoteSpecIDs` already has, moved here.
  - `func RemoteSpecIDs(root, ref string) []string` — kept, and now the id
    view over `RemoteSpecDirs`, so `TestRemoteSpecIDs_ReadsTheRef` passes
    unchanged.
  - `func SharedSpecRefs(root string) []SpecRef` — the union of
    `RemoteSpecDirs` over `RemoteRefs`, deduped by `Dir`; falls back to `main`
    when there are no remote refs. `RemoteRefs` runs one `git for-each-ref`,
    and each ref that carries specs costs one `git ls-tree`: no per-branch
    network call, whatever the ref count.
- `internal/cli/work.go`: `cmdNew` and `cmdAccept` build a folder with
  `project.SpecDirName` and call `project.SharedSpecRefs`; `sharedSpecIDs` is
  removed; `idTaken(p, s, shared []project.SpecRef)` applies decision 1;
  `nextFreeNum(p, shared)` reads the ids from `SpecRef.ID`; `cmdRenumber`
  with no `--to` uses the same wider set, and with an explicit `--to` it
  refuses a target a remote ref holds under a different folder.
- Reused, unchanged: `project.NormalizeID`, `project.SpecDirName`,
  `project.FormatID`, `project.Slug`, `project.NextNum`,
  `internal/project/git.go:run`, `p.Specs`, `referencesTo`, `renumberSpec`,
  `referencedError`.

### Tests

- AC1 `internal/cli/cli_test.go:TestNew_SkipsIdsOnOtherBranches`: a local bare
  `origin` whose `spec/...` branch carries `SPEC-030-x` and whose `intake/...`
  branch carries `SPEC-031-y`; `forge new` mints `SPEC-032`. Neither ref is
  fetched.
- AC2 `TestNew_NoRemoteBranchesStaysOffline`: no remote refs and `fetch: on`
  in `.forge/project.md`; a before/after snapshot of the repository shows
  `origin` untouched, `SharedSpecRefs` returns the `main`-based set, and no
  fetch runs.
- AC3 `internal/project/project_test.go:TestRemoteSpecIDs_ScansEveryBranch`:
  a local bare remote with a `spec/*` and an `intake/*` branch; the returned
  folders are the union, and no network is touched.
- AC4 reuses `internal/cli/cli_test.go:TestRenumber_ResolvesTheRace` and
  `TestAccept_RenumbersWhenTakenOnMain` unchanged, and adds
  `TestRenumber_ToRefusesTakenOnAnotherFolder`: a local bare `origin` with
  `SPEC-030-x` on a `spec/*` branch; `forge renumber --to 30` of a local spec
  is refused and its folder is unchanged.
- AC5 `TestDocs_DescribeTheWiderIdRead` in `internal/cli/cli_test.go`: reads
  `../../docs/cli.md` and `../../docs/workflow.md` and asserts the wider,
  best-effort source of truth, both residual windows, that `forge new`/
  `forge accept` never fetch, and the `git fetch` before `forge new` step,
  the way `TestDocs_DescribeCapability` reads `docs/cli.md`.
- AC6 is `go test ./...`, `gofmt -l .` and `go vet ./...` themselves; they
  need no test of their own.
- AC7 `TestAccept_KeepsIdForItsOwnPublishedBranch`: push the current spec's
  folder to `origin` on the branch under test, then `forge accept`; the id is
  unchanged, there is no `renumbered from` line and no `forge renumber`
  prompt.

## Existing state

- SPEC-015 (`Collapse the redundant gates and rename the contract state`),
  decision 7 is the id authority this grows from: `forge accept` confirms the
  provisional number against the shared branch before recording it, through
  `internal/project/git.go:RemoteSpecIDs`, and `renumberSpec`/
  `referencedError` (`internal/cli/work.go:407,426`) are the only renumber
  path. It reads `origin/main` then `main` (`sharedSpecIDs`,
  `internal/cli/work.go:433`) and never fetches.
- SPEC-022 (`Fetch the latest changes at session start when the project opts
  in`) adds the opt-in `fetch: on` scalar and the session-start `forge brief`
  fetch; the id commands build on it by reading refs that are already fresh
  instead of fetching themselves.
- SPEC-020 decision 6 is the opt-in scalar shape; nothing in this change adds
  a scalar.
- `internal/project/git.go`: `RemoteSpecIDs` (36) and `run` (230) are the
  reads to reuse. `internal/project/project.go`: `NextNum` (236),
  `SpecDirName` (177), `Slug` (707), `NormalizeID` (680), `FormatID` (694).
  `internal/cli/work.go`: `cmdNew` (18), `cmdAccept` (139), `cmdRenumber`
  (375), `sharedSpecIDs` (433), `idTaken` (442), `nextFreeNum` (458),
  `renumberSpec` (407), `referencedError` (426).
- Conventions that apply: `.forge/conventions/architecture.md` (the id rule
  stays in `internal/project`, the package that owns ids, and `internal/cli`
  only assigns behaviour) and `.forge/conventions/testing.md` (a real local
  bare remote, as `internal/project/git_test.go:gitRepo` and
  `internal/cli/cli_test.go:checkpointRepo` already do, so no test reaches the
  network; "writes nothing" proven with a before/after snapshot).
- Deliberately not duplicated: `RemoteSpecIDs` is rebuilt over the new
  `RemoteSpecDirs` rather than gaining a second folder parser, and the ref
  enumeration reuses `run`.
- Not built yet: `SpecRef`, `RemoteRefs`, `RemoteSpecDirs`, `SharedSpecRefs`,
  the folder-aware collision rule, and the wider refs in
  `cmdNew`/`cmdAccept`/`cmdRenumber`.
- Dependency to keep visible, stated as two separate claims.
  **Today the `intake/` pull request is what publishes the number before it is
  accepted:** pushing the branch is how a provisional id becomes visible to
  anyone else's `SharedSpecRefs` read before `forge accept` records it.
  **If SPEC-015 removes that step:** publishing the branch becomes the manual
  reservation, and this spec protects only numbers already published on some
  ref — an id that was never pushed stays invisible until it merges. Nothing
  else in the tree reads ids from other refs.

## Out of scope

- A central allocator, id reservation, locking, or any guarantee that two
  branches can never collide.
- Fetching inside `forge new`/`forge accept`, and any change to SPEC-022's
  `fetch: on` behaviour or `forge brief`.
- Age- or merge-based filtering of refs; bounding or capping the number of
  refs scanned.
- Changing `forge validate` or `forge check` to read remote refs.
- Reading local branches other than the `main` fallback.
- Renaming a spec's folder when its title changes.
- Rewriting `## History`, delivered `.forge/specs/*` or `.forge/decisions/*`.

## History

Written by `forge`. Do not edit by hand.
- 2026-09-20  accepted  by TheJisus28
- 2026-09-20  contracting  by TheJisus28
