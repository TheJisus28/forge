# Plan — SPEC-008

Phases, in order. One phase is one run of the implementer and one commit:
small enough to verify, large enough to mean something.

## Existing state

- Delivered specs this builds on: SPEC-007 keeps Forge's machinery in the
  binary, so `forge upgrade` ships as a command, not as a planted script.
  The delivery loop (SPEC-002, SPEC-004) already ends every spec as a pull
  request.
- Reuse `internal/cli/cli.go`: the dispatch switch, the `usage` const and
  `cli.Version`; `cmdUpgrade` slots in next to `init`/`update`.
- Reuse `internal/cli/init.go`: `newFlagSet` + `parseArgs` is how every
  command parses flags, and `cmdValidate` is the precedent for a command
  that takes both `stdout` and `stderr`.
- Reuse `internal/project/git.go`: the `exec.Command` pattern
  (`project.Fetch`, `project.GH`) is the precedent for the toolchain owning
  the network while Forge owns the flow.
- Reuse `main.go`: `cli.Version` already resolves from ldflags or
  `debug.ReadBuildInfo`; the command only reads it.
- Constraint: `nonet_test.go` bans `net/http` and friends. Nothing in this
  change imports them; the only network is the invoked `go` toolchain.
- Conventions in `.forge/conventions/`: none written yet, so nothing to
  apply beyond the rules in `AGENTS.md` (stdlib only, errors are sentences,
  tests describe behaviour).
- Duplication avoided: no new package, no third dependency, no second
  shell-out helper; the temp-`GOBIN` install means the Go toolchain is the
  sole downloader, the same way `git`/`gh` are for the rest.
- What genuinely has to be built: `internal/cli/upgrade.go` with
  `cmdUpgrade`, `cleanupStaleBinary`, `replaceExecutable`, `copyFile`,
  `exeSuffix` and the test seams; the dispatch case and usage line in
  `cli.go`; the `cleanupStaleBinary()` call at the top of `cli.Main`; the
  `docs/cli.md` section; the tests.

## Phase 1 — install and report

- Scope: `internal/cli/upgrade.go` with `cmdUpgrade`, `forgeModule`,
  `exeSuffix` and the seams `lookupGo`, `goInstall`, `goVersion`,
  `executable`; `case "upgrade"` and the usage line in
  `internal/cli/cli.go`. No replacement yet: this phase installs into the
  temp `GOBIN` and prints the two output lines.
- Done when: `forge upgrade` accepts no arg (`latest`), one arg verbatim, or
  two and errors; a missing `go` fails with the actionable sentence; the
  command never calls `project.Find`; the before/after versions are printed.
  Moves AC1, AC2, AC4, AC5, AC6 (report) and AC7 (help).
- Verify with: `go test ./...` and `go run . help`.

## Phase 2 — safe replacement, cleanup and docs

- Scope: `replaceExecutable`, `copyFile`, `cleanupStaleBinary` and the
  `cleanupStaleBinary()` call at the top of `cli.Main`; the
  `docs/cli.md` section. Wire the replacement into `cmdUpgrade` after a
  successful install.
- Done when: the running executable is replaced through the sidecar
  (`<exe>.old`) on Windows and Unix, a stale sidecar is removed on the next
  invocation, a failed rename leaves the binary unchanged and returns the
  actionable sentence, and `docs/cli.md` documents the command. Moves AC3
  (guarded by `nonet_test.go`), AC6 (replacement) and AC7 (docs).
- Verify with: `go test ./...`, `gofmt -l .`, `go vet ./...`.

## Risks

- Windows sidecar release: the old image stays locked until the process
  exits, so the `.old` removal in later steps is best effort by design;
  cleanup on the next invocation covers it.
- `os.Executable()` under `go run .` is a temp build, so a dev run replaces
  a throwaway file. Harmless, and the tests stub `executable`.
- Cross-volume temp: `os.Rename` can fail across volumes, which is why the
  new binary is copied, not moved; the risk is covered by the rollback path.
