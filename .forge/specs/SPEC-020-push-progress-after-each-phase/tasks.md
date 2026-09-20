# Tasks — SPEC-020

The phases from the plan, as checkboxes. One phase is one implementer run
and one commit. Tick a phase when it lands and say where the work is, so a
later spec knows what exists without reading the diff.

- [ ] Phase 1 — The git helpers and the opt-in flag. Moves: AC1, AC4.
  Where: `internal/project/git.go` (`CommitAll`, `HasUnpushed`,
  `OnDefaultBranch`), `internal/project/project.go` (`PushEnabled`); tests in
  `internal/project/`.
- [ ] Phase 2 — `forge push` and the checkpoint. Moves: AC1, AC2, AC3.
  Where: `internal/cli/push.go` (new), `internal/cli/cli.go`; tests in
  `internal/cli/cli_test.go`.
- [ ] Phase 3 — `forge advance` checkpoints when opted in. Moves: AC4.
  Where: `internal/cli/work.go` (`cmdAdvance`); tests in
  `internal/cli/cli_test.go`.
- [ ] Phase 4 — `forge submit` commits before it pushes. Moves: AC1.
  Where: `internal/cli/report.go` (`cmdSubmit`); tests in
  `internal/cli/cli_test.go`.
- [ ] Phase 5 — Docs, roles and changelog. Moves: AC5. Where:
  `docs/cli.md`, `docs/workflow.md`, `kit/machine/roles/implementer.md`,
  `kit/machine/roles/orchestrator.md`, `CHANGELOG.md`; tests in
  `internal/cli/cli_test.go`, `internal/cli/machine_test.go`.

## Proposed conventions

None.
