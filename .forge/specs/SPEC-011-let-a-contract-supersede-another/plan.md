# Plan — SPEC-011

## Existing state

- Builds on SPEC-010 (`capability`): the frontmatter field pattern, the
  `project.Spec` typed field, `FromDoc`/`Save` through `setOrDelete` and the
  list helpers, the intrinsic-check-before-state-gate placement, and the
  `view.Detail` row pattern are all already here and are reused as-is.
- `internal/project/project.go` already parses and writes the `covers` list
  (`setStrList`); `supersedes` is the same shape, not new parser machinery.
- `internal/validate/validate.go` already has `depCycle`, the `add(sev,
  spec, ...)` helper and the `%q`/`%s` message style; `checkSupersedes`
  reuses them.
- `internal/view/view.go` already renders `status` and `capability` rows and
  the `(none)` helper.
- Conventions: `.forge/conventions/frontmatter.md` (field checks before the
  state gate), `.forge/conventions/cli-output.md` (`%q` quoting), and
  decision 0004.
- Duplication avoided: no second cycle walker (model on `depCycle`), no
  second list parser.
- Genuinely new: the `supersedes` field, its four validate rules, and the
  two `Detail` rows.

## Phase 1 — Field and project API

- Scope: `Spec.Supersedes`, `FromDoc`, `Save`, unit tests.
- Verify with: `go test ./internal/project/...`.

## Phase 2 — `forge validate` rules

- Scope: `checkSupersedes` with the four findings.
- Verify with: `go test ./internal/validate/...`.

## Phase 3 — Status view and end-to-end

- Scope: `view.Detail` `supersedes`/`superseded by` rows; end-to-end test.
- Verify with: `go test ./...` and `go run . status SPEC-011`.

## Risks

- Uniqueness must count live specs only; counting terminal ones would flag
  legitimate history. The contract fixes the rule.
