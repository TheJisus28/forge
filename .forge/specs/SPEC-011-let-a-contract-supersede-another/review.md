# Review — SPEC-011

Verdict: **pass**

Reviewed at commit `dfa6dfa` (`spec(spec-011): every phase is finished`),
on top of the feature commit `58fd7da` (`feat(validate): let a contract
supersede another`).

Test command from `.forge/project.md` is `go test ./...`. Extra gates
required before landing: `gofmt -l .` and `go vet ./...`.

- `go test ./...` — all packages `ok` (forge, internal/cli, internal/doc,
  internal/project, internal/validate, internal/view, internal/workflow,
  kit).
- `gofmt -l .` — no output (nothing unformatted).
- `go vet ./...` — clean, exit 0.

## Criteria

| AC | Result | Evidence |
| --- | --- | --- |
| AC1 — frontmatter accepts `supersedes: [SPEC-NNN]`, a list | pass | `project.Spec.Supersedes` at `internal/project/project.go:58`; read via `normalizeIDs(d.List("supersedes"))` at line 438; written via `setListOrDelete(s.doc, "supersedes", s.Supersedes)` at line 385. `TestLoad_ReadsSupersedes` reads `supersedes: [SPEC-002, spec-3]` back as `[SPEC-002 SPEC-003]`; `TestSaveRoundTrip` persists it and confirms an empty list removes the key. Both PASS. |
| AC2 — fails when the target does not exist | pass | `checkSupersedes` at `internal/validate/validate.go:117-121` emits `Error`: `supersedes SPEC-NNN, which does not exist`. `TestRun_SupersedesAMissingSpec` PASS. Scratch repo outside this one: `error SPEC-001: supersedes SPEC-404, which does not exist`, exit 1. |
| AC3 — fails when the target is not `done` | pass | `internal/validate/validate.go:123-125` (`other.Status != workflow.Done`) emits `supersedes SPEC-NNN, which is not done`. `TestRun_SupersedesASpecThatIsNotDone` PASS. Scratch: target `accepted` → `error SPEC-001: supersedes SPEC-002, which is not done`, exit 1. |
| AC4 — fails on a cycle and names the chain | pass | `supersedeCycle` at `internal/validate/validate.go:344-372`, same walk shape as `depCycle`, reported as `supersedes cycle: <chain>`. `TestRun_SupersedesCycle` PASS. Scratch (both `accepted`, mutual): `error SPEC-001: supersedes cycle: SPEC-001 -> SPEC-002 -> SPEC-001`, exit 1. |
| AC5 — two non-terminal specs superseding the same target fails, naming both | pass | `internal/validate/validate.go:130-144`: only runs when the reporting spec is non-terminal, skips terminal claimants, names the other spec in the message. `TestRun_TwoLiveSpecsSupersedeTheSameTarget` asserts both directions and PASSES. Scratch: `SPEC-001: supersedes SPEC-003, which SPEC-002 also supersedes` and `SPEC-002: supersedes SPEC-003, which SPEC-001 also supersedes`, exit 1. Variant with one claimant `dropped` produced no error, exit 0, confirming the "two live" condition. |
| AC6 — `forge status <id>` shows both directions | pass | `view.Detail` rows at `internal/view/view.go:231-236`, `supersededBy` at lines 301-315. `TestStatusShowsSupersedes` PASS. Scratch: `forge status SPEC-002` prints `supersedes     SPEC-001`; `forge status SPEC-001` prints `superseded by  SPEC-002`; both exit 0. |

## Contract conformance

- **Interface stable.** The frozen frontmatter key `supersedes`, the Go field
  `project.Spec.Supersedes []string`, `Save`, the four validate findings and
  the two `Detail` rows all match the contract's "Interfaces other specs
  build against". No change after approval.
- **Spec order.** `checkSupersedes` is called inside the per-spec loop before
  `if !workflow.Valid(s.Status) { continue }` (`internal/validate/validate.go:68`),
  as decision 3 and the `frontmatter` convention require.
- **No drift.** Running `forge validate` against this repository reported no
  `contract changed after ... approved` finding for SPEC-011, so `contract_hash`
  still matches the approved text.
- SPEC-012 (the derived view) only mentions `supersedes` in its problem
  statement; the field it will read exists.

## Blocking problems

None. Every acceptance criterion passes with the evidence above.

## Notes (non-blocking)

- The contract called the list writer `setStrList`; the real helper is
  `setListOrDelete`. The implementer used the real name and the behaviour is
  the intended one (empty list removes the key). The `tasks.md` note already
  records this; no code change is needed, but the contract's wording is
  slightly off for anyone reading it later.
- The two new `Detail` rows start their values at a wider column (15) than
  the existing 12-column rows, so `supersedes`/`superseded by` line up with
  each other but not with the rows above. Cosmetic only; no test depends on
  it. If the team wants one column, that is a follow-up across all rows.
- A cyclic pair where neither target is `done` reports both `which is not done`
  and `supersedes cycle`. That is a faithful reading of AC3 and AC4, not a
  defect: both statements are true.

## Proposed conventions

None.