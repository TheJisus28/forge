# Plan — SPEC-010

Phases, in order. One phase is one run of the implementer and one commit:
small enough to verify, large enough to mean something.

## Existing state

- Delivered specs and contracts this builds on: decision 0004
  (`.forge/decisions/0004-historical-specs-with-a-derived-view.md`) declares
  the field; SPEC-009 established that repository-root paperwork and
  `.forge/` edits need no product spec, which is why this spec folder is
  editable while the work is not yet `implementing`.
- Modules that already do part of the job:
  - `internal/project/project.go` already carries typed frontmatter fields
    (`Title`, `Conductor`, `Parent`, `Covers`) parsed in `FromDoc` and
    written in `Save` through `setOrDelete` and `setStrList`; `capability` is
    one more of those, not new machinery.
  - `internal/validate/validate.go` already has `checkRelations` and
    `checkArtifacts` with the `add(sev, spec, ...)` helper and the
    `Warning`/`Error` split; `checkCapability` joins them.
  - `internal/cli/work.go:cmdNew` already builds the `Spec` and loads
    `kit/machine/templates/spec.md` through `loadTemplate`; the flag is
    parsed there.
  - `internal/view/view.go` already renders the detail block and the brief's
    current-spec block; the line is added there.
- Conventions that apply: `.forge/conventions/testing.md` — a seam only where
  a test must substitute one; here none is needed, the commands run end to
  end against a temporary repository. No new dependency.
- Duplication this plan avoids: no second slug regex (one exported
  `project.ValidCapability`), no second frontmatter reader, no separate
  "tags" structure.
- What does not exist yet and has to be built: the field itself, the
  `--capability` flag, the two validate rules, the two view lines, and the
  template and docs entries.

## Phase 1 — The field and the project API

- Scope: `Spec.Capability`, parsed by `FromDoc` and written by `Save`, plus
  exported `project.ValidCapability`.
- Done when: a spec with `capability: guard` round-trips and `ValidCapability`
  accepts `guard` and rejects `Guard`, `""`, `with_underscore`, `guard!`.
- Verify with: `go test ./internal/project/...`.
- Moves: groundwork for AC2, AC5.

## Phase 2 — `forge new --capability`

- Scope: required flag, exact missing-flag error, unknown-name warning that
  still creates the spec; usage text.
- Done when: `forge new "X"` fails and creates nothing; `forge new "X"
  --capability guard` writes the field and warns only while no spec declares
  it.
- Verify with: `go test ./internal/cli/...`.
- Moves: AC1, AC2, AC4.

## Phase 3 — `forge validate`

- Scope: `checkCapability` — warning when absent, error when present but not
  a valid slug or empty.
- Done when: a missing capability exits 0 with a named warning; `Guard` and
  `""` exit 1.
- Verify with: `go test ./internal/validate/...`.
- Moves: AC3, AC5.

## Phase 4 — Surface, template and docs

- Scope: `view.Detail` and `view.Brief` lines; `capability: ""` in the spec
  template; `cli.go` usage; `docs/cli.md`; SPEC-010 declares
  `capability: workflow` on itself.
- Done when: `forge status`/`forge brief`/`forge template spec` show the
  field and `docs/cli.md` documents the flag.
- Verify with: `go test ./...` and `go run . status SPEC-010`.
- Moves: AC6, and closes the self-warning the contract's decision 7 names.

## Risks

- Every existing `forge new` call in `cli_test.go` must gain `--capability`,
  or the suite fails once the flag is mandatory; Phase 2 enumerates them in
  the contract.
- If the warning for an unknown name fires on the first spec of a fresh
  repository, that is accepted (contract, judgment call a); it must exit 0.
