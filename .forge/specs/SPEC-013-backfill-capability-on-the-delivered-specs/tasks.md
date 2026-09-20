# Tasks — SPEC-013

- [x] Phase 1 — Backfill and guard. Landed the `capability:` line under
  `status:` in SPEC-001 (`init`), SPEC-002 (`delivery`), SPEC-003 (`specs`),
  SPEC-004 (`delivery`), SPEC-005 (`agents`), SPEC-006 (`agents`), SPEC-007
  (`packaging`), SPEC-008 (`upgrade`) and SPEC-009 (`guard`). Everything else
  in those files is untouched (`git diff` shows one added line each).
  Decision 3 also required a silent tree, so the open specs that predate the
  field got theirs too: SPEC-014 `guard`, SPEC-015…018 `workflow`.
  New regression test `TestDeliveredSpecsDeclareCapability` in
  `internal/validate/validate_test.go` loads the repository root and fails if
  any `done` spec lacks a `ValidCapability`.
  Verified: `go run . validate` → `19 specs, no problems` (exit 0, no
  missing-capability warning); `go run . capabilities` lists SPEC-001…009
  under their capabilities; `go test ./...` all `ok`; `gofmt -l .` empty;
  `go vet ./...` clean.

## Proposed conventions

None.

## Notes for the reviewer

- `capability` is added by hand rather than through a `forge` writer: there
  is no command to set a single frontmatter field on an archived spec, and
  the point of the backfill is to add exactly one line. `contract_hash` is
  unaffected (it covers the Contract section), confirmed by `forge validate`.
