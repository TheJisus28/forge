# Plan — SPEC-013

## Existing state

- `capability` is a first-class frontmatter field (SPEC-010): parsed into
  `project.Spec.Capability`, validated by `validate.checkCapability`, and
  shown by the view (SPEC-012). This spec only fills the field in.
- `internal/project/project_test.go` and `internal/validate/validate_test.go`
  already build scratch `.forge` trees and know the `doc` frontmatter shape.
- `forge validate` under this repo currently prints one missing-capability
  warning per delivered spec; the fix is data, not code.
- Convention: `.forge/conventions/frontmatter.md` (field under `status:`,
  slug shape). No product code changes are needed; only the nine spec files
  and one regression test.

## Phase 1 — Backfill and guard

- Scope: add the `capability:` line to SPEC-001…009, add a test that every
  `done` spec under this repo declares a valid capability, run `forge
  validate`.
- Verify with: `go test ./...`, `go run . validate`, `go run . capabilities`.
- The test loads the repository root (`filepath.Join("..", "..")` from the
  package) and asserts `project.ValidCapability` for every `done` spec, so a
  future spec created without the field fails the suite here.

## Risks

- Editing a `done` spec must not touch `contract_hash`; the hash covers the
  Contract section only, and the new line is outside it. The test pins this
  indirectly (a changed contract would fail `forge validate`).
