# Plan — SPEC-001

## Existing state

- `internal/cli/init.go`: `kitOwned` listed `AGENTS.md`, `CLAUDE.md` and
  `.forge/README.md`, so `plant` rewrote them on init and update; `protected`
  only covered `.forge/project.md`.
- Tests: `internal/cli/cli_test.go` ran init end to end but never against a
  pre-existing `AGENTS.md`. `docs/cli.md` already promised "existing files are
  kept".
- Delivered specs: none before this one.

## Phase 1 — Stop claiming the root pointers

- Scope: remove `AGENTS.md` and `CLAUDE.md` from `kitOwned`; add
  `TestInit_KeepsExistingRootPointers`; fix the `docs/cli.md` update wording.
- Done when: init and update keep an existing `AGENTS.md`/`CLAUDE.md`, and
  `--force` rewrites them.
- Verify with: `go test ./...`.

## Risks

- `forge update` no longer refreshes the root pointers; `--force` remains for a
  deliberate refresh.
