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

- **A list of spec ids read from frontmatter is normalised through
  `normalizeIDs`, the id counterpart of `upperAll` for criteria.** `covers`
  goes through `upperAll` in `FromDoc`; `supersedes` now goes through
  `normalizeIDs` in `project.go`, and `parent`/`depends_on` already call
  `NormalizeID`. `Save` uses the same `setListOrDelete` as every other list,
  so an empty typed slice removes the key rather than writing `[]`.
- **A validate finding about a shared defect names the other party inside
  the message, not in a second finding field.** The contract left the cycle
  and duplicate texts open. I used `supersedes cycle: A -> B -> A` (the
  `dependency cycle:` shape) and `supersedes SPEC-NNN, which SPEC-MMM also
  supersedes`, with the reporting spec carried by `Finding.Spec`. Every
  offending live spec gets its own finding, matching `depCycle`; the pair
  therefore yields two findings, each naming both.
- **Detail row labels are hardcoded to a column, and a label longer than the
  existing 12-column field starts a wider one.** All current `Detail` rows
  put the value at column 13. `superseded by` is 13 characters, so the two
  new rows use a 15-column value start (`supersedes` + 5 spaces,
  `superseded by` + 2) so they align with each other without reformatting
  the rows above. If the team wants one column for `Detail`, the existing
  rows would need the same treatment.

## Notes for the reviewer

- The contract and the plan call the list writer `setStrList`; the actual
  helper in `internal/project/project.go` is `setListOrDelete`. I used the
  real one, so the behaviour matches the intent (an empty list removes the
  key). No other name in the contract diverged from the code.
