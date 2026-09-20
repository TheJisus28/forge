# Tasks — SPEC-009

The phases from the plan, as checkboxes. One phase is one implementer run
and one commit. Tick a phase when it lands and say where the work is, so a
later spec knows what exists without reading the diff.

- [x] Phase 1 — the guard rule. Where: `internal/cli/guard.go`
  (`isRootPaperwork`, `isProcessFile`), `internal/cli/cli_test.go`
  (`TestGuardFileMode`). Commit `dfc8e39`.
- [ ] Phase 2 — drop the community docs and migrate what matters. Where:
  `CONTRIBUTING.md`, `SECURITY.md`, `CODE_OF_CONDUCT.md` (removed),
  `AGENTS.md`, `README.md`, `CHANGELOG.md`, `internal/cli/cli_test.go`.
- [ ] Phase 3 — document the rule. Where: `docs/customizing.md`,
  `docs/cli.md`, `docs/opencode.md`, `internal/cli/cli_test.go`.

## Proposed conventions

Patterns decided because nothing was written. The team decides whether they
become rules in `.forge/conventions/`.

None.
