# Tasks — SPEC-022

The phases from the plan, as checkboxes. One phase is one implementer run
and one commit. Tick a phase when it lands and say where the work is, so a
later spec knows what exists without reading the diff. Name the criteria the
phase delivers (`Moves: <criterion ids>`), so every criterion is traceable to
the task that moves it; `forge check` reads these ids from this file.

- [ ] Phase 1 — The opt-in scalar and GitHub detection. Moves: AC2, AC4.
  Where: `internal/project/project.go` (`FetchEnabled`),
  `internal/project/git.go` (`HasGitHubRemote`); tests in
  `internal/project/project_test.go`, `internal/project/git_test.go`.
- [ ] Phase 2 — The brief fetches when opted in. Moves: AC1, AC2, AC3, AC4,
  AC5. Where: `internal/cli/report.go` (`fetchBeforeBrief`, the `githubRemote`
  and `ghUser` seams, `cmdBrief`); tests in
  `internal/cli/brief_internal_test.go` (new).
- [ ] Phase 3 — Docs, roles, changelog and dogfooding. Moves: AC6. Where:
  `docs/cli.md`, `README.md`, `AGENTS.md`, `.forge/project.md`,
  `kit/machine/roles/orchestrator.md`, `kit/machine/roles/architect.md`,
  `CHANGELOG.md`; tests in `internal/cli/cli_test.go`,
  `internal/cli/machine_test.go`.

## Proposed conventions

None.
