# Tasks — SPEC-011

- [x] Phase 1 — Field and project API. `project.Spec.Supersedes []string`,
  read in `FromDoc` through the new `normalizeIDs` and written in `Save`
  through `setListOrDelete` (empty list removes the key); tests
  `TestLoad_ReadsSupersedes` and the extended `TestSaveRoundTrip`. Where:
  `internal/project/project.go`, `internal/project/project_test.go`.
- [x] Phase 2 — `forge validate` rules. `checkSupersedes` (called from `Run`
  with `checkCapability`, before the unknown-status `continue`) plus
  `supersedeCycle` (modeled on `depCycle`) and `hasID`; tests for missing
  target, not-done target, cycle, double-supersede and a quiet valid
  supersede. Where: `internal/validate/validate.go`,
  `internal/validate/validate_test.go`.
- [x] Phase 3 — Status view and end-to-end. `view.Detail` rows
  `supersedes` / `superseded by` and the `supersededBy` helper; test
  `TestStatusShowsSupersedes` drives a spec to `done` and shows the link
  both ways. Where: `internal/view/view.go`, `internal/cli/cli_test.go`.

## Proposed conventions

None.

## Notes for the reviewer

- The contract and the plan call the list writer `setStrList`; the actual
  helper in `internal/project/project.go` is `setListOrDelete`. The
  implementer used the real one, so behaviour matches the intent.
