# Review - SPEC-013 Backfill capability on the delivered specs

Verdict: **pass**

Reviewed on branch `spec/013-backfill-capability-on-the-delivered-specs`
against contract `4d8a82b3f4cf`. Product code not touched beyond one test;
no criterion failed.

## Criterion / evidence

| # | Criterion | Result | Evidence |
|---|---|---|---|
| AC1 | SPEC-001…009 each declare a capability matching the subsystem | pass | `forge capabilities` groups them exactly as decision 1: `init`→001, `delivery`→002/004, `specs`→003, `agents`→005/006, `packaging`→007, `upgrade`→008, `guard`→009. Each diff is one added line. |
| AC2 | `forge validate` reports no missing-capability warning | pass | `go run . validate` → `19 specs, no problems`, exit 0, no `warning SPEC-…: has no capability`. Decision 3's "silent tree" also required the open specs, so SPEC-014 (`guard`) and SPEC-015…018 (`workflow`) were backfilled too. |
| AC3 | `forge capabilities` lists the delivered specs under their capabilities | pass | The command printed all eight groups with SPEC-001…009 under them (see AC1). |
| AC4 | Only `capability` is added; no contract, status or history line changes | pass | `git show ee0d0e8 --stat` shows `1 +` for each of the nine specs; `git show` on SPEC-001 and SPEC-009 shows the single `+capability:` line. `contract_hash` unchanged (validate is clean). |

Suite: `go test ./...` all `ok`; `gofmt -l .` empty; `go vet ./...` clean.

## Contract integrity

The recorded `contract_hash` matches and `forge brief` reports no drift. The
new test `TestDeliveredSpecsDeclareCapability` (in
`internal/validate/validate_test.go`) loads the repository root and fails if
any `done` spec lacks a `ValidCapability`, which is the guard decision 3
asked for.

## Problems

None blocking the merge and none blocking the close.

## Proposed conventions

None.
