# Tasks — SPEC-008

The phases from the plan, as checkboxes. One phase is one implementer run
and one commit. Tick a phase when it lands and say where the work is, so a
later spec knows what exists without reading the diff.

- [x] Phase 1 — install and report. Where: `internal/cli/upgrade.go`
  (`cmdUpgrade`, `forgeModule`, `exeSuffix`, `installModule`, `moduleVersion`,
  seams `lookupGo`, `goInstall`, `goVersion`, `executable`),
  `internal/cli/cli.go` (dispatch case and usage line),
  `internal/cli/upgrade_internal_test.go` and
  `internal/cli/upgrade_test.go`. Verified: `go test ./...`, `gofmt -l .`,
  `go vet ./...` all clean; `go run . help` lists `forge upgrade`.
- [ ] Phase 2 — safe replacement, cleanup and docs. Where:
  `internal/cli/upgrade.go` (`replaceExecutable`, `copyFile`,
  `cleanupStaleBinary`), `internal/cli/cli.go` (`cleanupStaleBinary()` at the
  top of `Main`), `docs/cli.md`.

## Proposed conventions

Patterns decided because nothing was written. The team decides whether they
become rules in `.forge/conventions/`.

None.
