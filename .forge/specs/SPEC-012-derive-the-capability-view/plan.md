# Plan — SPEC-012

## Existing state

- Builds on SPEC-010 (`capability`) and SPEC-011 (`supersedes`): the two
  frontmatter fields, `project.Spec.Capability`, `project.Spec.Supersedes`,
  and the deterministic-read pattern are all here. The view is a projection
  over `p.Specs`, not new parsing.
- `internal/project/project.go` already exposes `p.Specs`, `p.Spec`, and the
  `Capability`/`Supersedes` fields; grouping belongs there so `brief` and the
  command share it.
- `internal/view/view.go` already builds `Brief`; the summary is one more
  block. `internal/cli` already has the command-registration pattern in
  `cli.go` and a `report.go`.
- Conventions: `.forge/conventions/frontmatter.md` (field shape and list
  normalisation), decision 0004.
- Duplication avoided: no second grouping pass, no `git` call, no cached
  file.
- Genuinely new: `project.Project.Capabilities()`, the `forge capabilities`
  command, the brief summary, and the docs pointers.

## Phase 1 — `Capabilities()` in the project package

- Scope: group `done` specs by capability, order, mark superseded.
- Verify with: `go test ./internal/project/...`.

## Phase 2 — `forge capabilities [name]`

- Scope: the command, its usage, `cli.go` registration, end-to-end tests.
- Verify with: `go test ./internal/cli/...`.

## Phase 3 — Brief summary and docs

- Scope: `view.Brief`, `docs/cli.md`, the orchestrator role and the
  `forge-work` skill survey step.
- Verify with: `go test ./...` and `go run . capabilities`.

## Risks

- "Superseded" must be computed from non-dropped specs only; a dropped
  replacement must not bury the current contract. The contract fixes this.
