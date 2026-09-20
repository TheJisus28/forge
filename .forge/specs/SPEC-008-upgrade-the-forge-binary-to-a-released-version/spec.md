---
id: SPEC-008
title: Upgrade the forge binary to a released version
status: implementing
created: 2026-09-20
updated: 2026-09-20
accepted_by: TheJisus28
conductor: TheJisus28
approved_by: TheJisus28
contract_hash: dc4450490120
---

## Problem

Forge is installed once, with `go install github.com/TheJisus28/forge@latest`
or from a release archive, and there is no way to move to a newer release
from the tool itself. `forge update` refreshes the kit planted in a
repository but not the binary, so after an upgrade two things are stale and
only one has a command. Users have to remember the exact install invocation
and re-run it by hand, and nothing in `forge help` tells them how.

## Acceptance criteria

Observable outcomes. Someone else must be able to mark each one pass or
fail with evidence.

- AC1: `forge upgrade` runs the Go toolchain for the user
  (`go install github.com/TheJisus28/forge@latest`), streams its output, and
  exits 0 when the install succeeds.
- AC2: `forge upgrade <version>` installs that exact version, passing
  `@<version>` (for example `forge upgrade v0.2.0`).
- AC3: The binary still reaches no network on its own: no banned import is
  added and `go test ./...` (including `nonet_test.go`) passes. The only
  network is the `go` toolchain the user invoked.
- AC4: When `go` is not on `PATH`, `forge upgrade` fails with a sentence
  that tells the user to install Go 1.22+ or run
  `go install github.com/TheJisus28/forge@latest` themselves.
- AC5: `forge upgrade` does not require and does not modify a `.forge/`
  project; it runs outside one and leaves the current repository's kit
  untouched. Refreshing the kit stays `forge update`.
- AC6: The command prints on stdout the version it is replacing (from
  `forge version` / `cli.Version`) and the installed version: for an explicit
  version that exact version, and for `@latest` the module version resolved
  from the new binary, falling back to `latest` when it cannot be read.
- AC7: `forge help` lists `forge upgrade` and `docs/cli.md` documents it.

## Open questions

One per line, like `- OQ1: ...`, for anything that must be answered before
the contract is approved. Write `None.` when there are none.

None.

## Contract

One new command, `forge upgrade`, plus one maintenance side effect: a
stale sidecar beside the running binary is removed on the next invocation.
No new package and no new dependency.

### Decisions

1. **Command and dispatch.** Add `case "upgrade": err = cmdUpgrade(rest,
   stdout, stderr)` to the switch in `internal/cli/cli.go:51`, next to
   `init`/`update`. `cmdUpgrade` lives in a new file
   `internal/cli/upgrade.go` (package `cli`) with signature
   `func cmdUpgrade(args []string, stdout, stderr io.Writer) error`. Unlike
   most commands it takes both writers because it streams the Go toolchain,
   which writes progress and errors to stderr. Discards a stderr-less
   signature that would silently drop the toolchain output (`cmdValidate`
   already takes both writers, so the shape is precedent).
2. **Argument.** `forge upgrade [version]`: zero positionals means ref
   `latest`; exactly one is used verbatim, with no `@` and no `v`
   normalisation (`forge upgrade v0.2.0` yields `@v0.2.0`); an empty string
   counts as absent; two or more is an error. Uses
   `newFlagSet("upgrade", "usage: forge upgrade [version]", stdout)` and
   `parseArgs`, so `--help` works. Discards parsing or rewriting versions.
3. **Module path.** `const forgeModule = "github.com/TheJisus28/forge"` in
   `internal/cli/upgrade.go`. Discards repeating the literal and discards
   reading it from `go.mod`, which the binary does not ship.
4. **Toolchain call.** `go install <forgeModule>@<ref>` through the seam
   `var goInstall = func(goBin, destDir, ref string, stdout, stderr
   io.Writer) error`; it sets `cmd.Env = append(os.Environ(), "GOBIN="+
   destDir)` and `cmd.Dir = destDir`, then streams `cmd.Stdout`/`cmd.Stderr`
   to the writers. Running in the temp dir keeps the user's workspace and
   module files out of the call. Discards installing into the real
   `GOBIN`/`GOPATH/bin` first, which would overwrite the running binary on
   Windows.
5. **Temp `GOBIN`.** `destDir, err := os.MkdirTemp("", "forge-upgrade-*")`
   with `defer os.RemoveAll(destDir)`. The fresh binary is
   `filepath.Join(destDir, "forge"+exeSuffix())`, where `exeSuffix()`
   returns `.exe` when `runtime.GOOS == "windows"` and `""` otherwise.
   Discards a fixed temp name shared between concurrent runs.
6. **No project.** `cmdUpgrade` never calls `project.Find` or any
   `internal/project` function. Discards reusing `cmdUpdate`'s project
   lookup, which fails outside `.forge/` and would couple the binary to the
   kit.
7. **Windows-safe replacement** in `replaceExecutable(exe, fresh, ref
   string) error`, in order:
   1. `old := exe + ".old"`; best-effort `os.Remove(old)` (a stale sidecar
      from a previous run).
   2. `os.Rename(exe, old)`. Windows permits renaming a running image; it
      forbids overwriting one.
   3. `copyFile(fresh, exe)`: byte copy with mode `0o755` (not
      `os.Rename`), so it also works when the temp dir is on another volume.
   4. Best-effort `os.Remove(old)`. On Windows the running process still
      holds the old file, so this delete may fail and is deliberately
      ignored.
   5. If step 2 fails, return an actionable error with nothing changed. If
      step 3 fails, attempt `os.Rename(old, exe)` rollback and report
      accordingly.
   The sidecar is always `<exe>.old`. Discards `os.Rename(fresh, exe)`
   (cross-volume) and any in-place overwrite (fails on Windows).
8. **Stale sidecar cleanup.** `func cleanupStaleBinary()` in `upgrade.go`
   calls `os.Executable()` and then best-effort `os.Remove(exe + ".old")`,
   ignoring errors. It is the first statement of `cli.Main`, before the
   `len(args) == 0` check, so any next `forge` invocation cleans up, not
   only the next `upgrade`. Discards invoking cleanup from `main.go`
   (keeps `main.go` thin and lets tests drive it through `cli.Main`) and
   discards cleaning only inside `cmdUpgrade`. Under `go run .` the path is
   a temp build with no sidecar, which is harmless.
9. **Version resolution.** After install,
   `var goVersion = func(goBin, path string) (string, error)` runs
   `go version -m <fresh>` and returns the token after `mod <forgeModule>`;
   on any failure it returns the requested ref instead. Discards a network
   version query, which the binary may not make.
10. **Output.** To `stdout`, exactly two lines: before installing
    `forge <cli.Version>: installing <forgeModule>@<ref>\n`, and after the
    replacement succeeds `forge upgraded to <installed>\n`. The child's
    stdout and stderr go straight to our stdout and stderr. Discards
    buffering the toolchain, and discards printing "upgraded" before the
    replacement has succeeded.
11. **Error sentences** (lowercase, `%w` where a cause exists; `<ref>` is
    the requested ref):
    - `go` absent: `go is not on PATH; install Go 1.22+ (https://go.dev/dl)
      and run: go install github.com/TheJisus28/forge@latest`
    - install failed: `go install github.com/TheJisus28/forge@<ref>
      failed: <cause>`
    - rename failed: `cannot replace the running forge binary: <cause>;
      close other forge processes and run: go install
      github.com/TheJisus28/forge@<ref>`
    - copy failed, rolled back: `cannot write the new forge binary:
      <cause>; forge is unchanged, run: go install
      github.com/TheJisus28/forge@<ref>`
    - copy failed, rollback failed: `cannot write the new forge binary:
      <cause>; the previous forge is at <exe>.old, move it back or run: go
      install github.com/TheJisus28/forge@<ref>`
    - too many args: `forge upgrade takes at most one version, for example:
      forge upgrade v0.2.0`
12. **Help and docs.** Add `forge upgrade [version]        upgrade the
    forge binary itself` to the `usage` const in `internal/cli/cli.go`, and
    a `### forge upgrade [version]` subsection to `docs/cli.md` under
    "Setting up" after `forge update`, saying it shells out to the Go
    toolchain, replaces the running binary safely (sidecar on Windows),
    works outside a `.forge/` project, and never touches the kit.

### Interfaces (exact)

New file `internal/cli/upgrade.go`, package `cli`:

- `const forgeModule = "github.com/TheJisus28/forge"`
- `func cmdUpgrade(args []string, stdout, stderr io.Writer) error`
- `func cleanupStaleBinary()`
- `func replaceExecutable(exe, fresh, ref string) error`
- `func copyFile(src, dst string) error`
- `func exeSuffix() string`
- package-level seam variables, overridden only by tests:
  `lookupGo = exec.LookPath`, `goInstall`, `goVersion`,
  `executable = os.Executable`, `renameFile = os.Rename`,
  `removeFile = os.Remove`.

Changed: `internal/cli/cli.go` (dispatch case, `usage` line,
`cleanupStaleBinary()` as the first line of `Main`), `docs/cli.md`.
Unchanged: `main.go`, `internal/project/*`, `go.mod`.

### Test seam

The OS and toolchain calls are package-level function variables in
`internal/cli/upgrade.go`. Because tests must swap unexported variables,
this spec adds the repository's first internal test file,
`internal/cli/upgrade_internal_test.go` with `package cli`; a sibling
`internal/cli/upgrade_test.go` with `package cli_test` keeps the black-box
help and documentation checks and reuses `mustRun`. No test runs a real
`go install`, and no test binary touches the network.

### Test plan

- AC1 (`@latest`, stream): stub `lookupGo`, `executable`, `goInstall`
  (records `(destDir, ref)` and writes a dummy `forge` file into
  `destDir`), `goVersion`; call `cmdUpgrade(nil, &out, &errOut)`; assert
  `ref == "latest"`, the recorded `destDir` is not the executable's
  directory, and `out` contains
  `installing github.com/TheJisus28/forge@latest`.
- AC2 (explicit version): args `["v0.2.0"]`; assert the recorded ref is
  `v0.2.0`.
- AC3 (no network): the existing `nonet_test.go`, plus `go test ./...`
  with the toolchain stubbed in every upgrade test.
- AC4 (no `go`): `lookupGo` returns `exec.ErrNotFound`; assert the error
  contains `Go 1.22+` and the manual `go install` command, and that
  `goInstall` was never called.
- AC5 (outside a project): in a `t.TempDir()` with no `.forge/`, stub the
  seams and call `Main([]string{"upgrade"}, &out, &errOut)`; assert exit 0
  and that no `.forge/` appears. AC5 is also structural: `cmdUpgrade`
  contains no reference to `internal/project`.
- AC6 (visible versions): `Version = "dev"` and `goVersion` returning
  `v0.3.0`; assert the output contains `dev` and `upgraded to v0.3.0`; a
  second case with `goVersion` returning an error asserts the ref is
  printed instead.
- AC7 (help and docs): in `package cli_test`,
  `mustRun(t, dir, "help")` contains `forge upgrade`; and a test reads
  `../../docs/cli.md` and asserts it contains `forge upgrade`.
- Replacement: with real `os.Rename`/`os.Remove`, create a temp `forge`
  (or `forge.exe`) holding `old` and an install that writes `new`; after
  `cmdUpgrade` the executable holds `new`. A case where `renameFile` fails
  asserts the actionable sentence and that the executable still holds
  `old`.
- Cleanup: create `<exe>.old`, call `cleanupStaleBinary`, assert it is
  gone; a case where `removeFile` errors asserts no panic and no returned
  error.

### Out of scope (confirmed)

As in `## Out of scope` below: no `gh release download` or release
archives, no detecting how Forge was installed, no network API version
check (no argument always means `@latest`), no downgrade protection beyond
passing a version, and no kit refresh (still `forge update`). The command
touches only the running executable and its temp `GOBIN`.

### Open questions

None.

## Out of scope

- Downloading release archives with `gh release download`, or replacing the
  running executable in place.
- Detecting whether Forge was installed from a module or a binary archive.
- Querying a network API to decide whether an upgrade is needed; without a
  version argument the command always installs `@latest`.
- Downgrade protection beyond passing an explicit version.
- Refreshing the planted kit; that remains `forge update`.

## History

Written by `forge`. Do not edit by hand.
- 2026-09-20  accepted  by TheJisus28
- 2026-09-20  specifying  by TheJisus28
- 2026-09-20  awaiting-approval  by orchestrator
- 2026-09-20  planning  by TheJisus28
- 2026-09-20  implementing  by orchestrator
