---
id: SPEC-023
title: Allocate spec ids against every remote spec branch, not just main
status: accepted
capability: specs
created: 2026-09-20
updated: 2026-09-20
accepted_by: TheJisus28
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

- AC1: `forge new` reads the ids committed on every remote spec branch, not
  only `main`, so a number a remote branch already holds is skipped when the
  provisional id is minted, and the `forge accept` confirmation reads the
  same wider set; `TestNew_SkipsIdsOnOtherBranches` returns a fresh number
  when a local bare `origin` carries that number on a `spec/*` branch.
- AC2: The wider read uses only refs already present and never fetches by
  itself; with no remote branches it falls back to local plus `main` and
  never fails, proven by a before/after snapshot test
  (`TestNew_NoRemoteBranchesStaysOffline`).
- AC3: The extra cost is bounded to one pass over remote refs (a single
  `git for-each-ref` plus one `git ls-tree` per ref, or equivalent) and adds
  no per-branch network call; `TestRemoteSpecIDs_ScansEveryBranch` settles it
  over a local bare remote.
- AC4: `forge renumber` can still resolve a duplicate that the wider read
  did not prevent, and `forge validate` keeps failing on a real duplicate, so
  the change reduces the window without removing the backstop; the existing
  `TestRenumber_ResolvesTheRace` and `TestAccept_RenumbersWhenTakenOnMain`
  still pass.
- AC5: `docs/cli.md` and `docs/workflow.md` state the wider source of truth
  and that it is best-effort, naming the residual collision window (two
  branches created before either pushes).
- AC6: `go test ./...` passes, and `gofmt -l .` and `go vet ./...` report
  nothing.

## Open questions

Settled before acceptance; the contract records them as numbered decisions.

OQ1 — Best-effort, not a reservation. Forge has no central allocator, so
"never collide" cannot be promised: the read only narrows the window to
branches whose remote refs are not present locally (created after the last
fetch) and to two branches created before either pushes. The contract states
that guarantee and no more.

OQ2 — Every remote spec branch is scanned, regardless of age, and none is
skipped for being merged or old. A merged branch's ids are already on `main`,
so scanning it is redundant but harmless, and an age filter would reopen the
window (an old branch can still merge). Gaps in the sequence are accepted:
numbering is max+1 and already tolerates them. An abandoned, unmerged branch
pins its number only until its ref is deleted.

OQ3 — `forge new` (mint), `forge accept` (confirm) and `forge renumber`
(next free) read the wider set. `forge validate` and `forge check` stay
local: they are the duplicate backstop and must not depend on refs being
fresh.

OQ4 — No hard cap is introduced. The read is local plumbing (`git
for-each-ref` plus one `git ls-tree` per ref), touches no network, and is
O(branches) like the other git operations here; a cap that silently dropped
refs would reopen the window. If a project ever has enough branches to
matter, the answer is an explicit warning and opt-out, deferred out of this
spec.

## Contract

Written by the architect once the work is accepted, and frozen once
approved. Real names from this repository: modules, endpoints,
tables, screens. Numbered decisions with what they discard. Anything other
specs will build against goes here.

## Existing state

<!-- Written by the architect, read while planning: name what already
     exists that this builds on, in this repository's real names.
     - Delivered specs and contracts this one builds on (`forge status`).
     - Modules, files or tools that already do part of the job: reuse them.
     - Conventions in `.forge/conventions/` that apply.
     - Duplication this change deliberately avoids.
     - What does not exist yet and genuinely has to be built. -->

## Out of scope

A closed list.

## History

Written by `forge`. Do not edit by hand.
- 2026-09-20  accepted  by TheJisus28
