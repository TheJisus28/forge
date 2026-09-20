# Tasks — SPEC-007

The phases from the plan, as checkboxes. One phase is one implementer run
and one commit. Tick a phase when it lands and say where the work is, so a
later spec knows what exists without reading the diff.

- [x] Phase 1 — Move the machinery out of the planted tree. Where: `kit/forge/kit/` → `kit/machine/`, `kit/kit.go`, `internal/cli/init.go` (`mapDest`, `kitOwned`, `plant`, `removeStaleKit`), `internal/cli/work.go` (embed path), `internal/cli/cli_test.go`.
- [x] Phase 2 — Internalise the templates and expose the machinery. Where: `kit/machine.go` (new), `internal/cli/machine.go` (new), `internal/cli/work.go` (`loadTemplate`), `internal/cli/cli.go` (switch + usage).
- [x] Phase 3 — Inline the roles into the host adapters. Where: `kit/machine/roles/{architect,implementer,reviewer}.md`, `kit/claude/agents/forge-*.md`, `kit/opencode/agents/forge-*.md`, `compose` in `internal/cli/init.go`.
- [x] Phase 4 — Documentation and tests. Where: `AGENTS.md`, `kit/AGENTS.md`, `kit/forge/README.md`, `kit/forge/{decisions,conventions}/README.md`, `kit/claude/skills/forge-*/SKILL.md`, `docs/*`, `README.md`, `CONTRIBUTING.md`, `CHANGELOG.md`, `internal/cli/{cli,machine}_test.go`, `kit/machine_test.go`, plus a `forge update` in this repository (removed `.forge/kit/`, regenerated the host adapters).

## Proposed conventions

None.
