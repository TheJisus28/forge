# Review — SPEC-008

Verdict: **pass with notes**

Reviewed branch `spec/008-upgrade-the-forge-binary-to-a-released-version`
at `ff49aef`. Evidence commands were run on Windows.

## Criteria

- **AC1 — pass.** `go install github.com/TheJisus28/forge@latest` is what
  runs: `TestUpgrade_InstallsLatestIntoATempDir` records ref `latest` and
  asserts stdout contains `installing github.com/TheJisus28/forge@latest`.
  `installModule` (upgrade.go:141) sets `cmd.Stdout`/`cmd.Stderr` to the
  writers, and `Main` returns 0 when `cmdUpgrade` returns nil
  (cli.go:101-108). Note: streaming the child's output is verified by
  inspection only — every test stubs `goInstall`.
- **AC2 — pass.** `TestUpgrade_UsesTheGivenVersionVerbatim` records ref
  `v0.2.0` for args `["v0.2.0"]`, used verbatim; `installModule` builds
  `github.com/TheJisus28/forge@v0.2.0`. `forge upgrade v0.2.0` ⇒
  `@v0.2.0`.
- **AC3 — pass.** `go test ./... -count=1` → all 8 packages `ok`; the root
  package includes `nonet_test.go`, which enforces the banned-import list.
  `internal/cli/upgrade.go` imports only `fmt`, `io`, `os`, `os/exec`,
  `path/filepath`, `runtime`, `strings`; no banned import anywhere under
  review.
- **AC4 — pass.** `TestUpgrade_WithoutGoExplainsHowToInstallIt` stubs
  `lookupGo` to `exec.ErrNotFound`, asserts the error contains `Go 1.22+`
  and `go install github.com/TheJisus28/forge@latest`, and asserts
  `goInstall` was never called. `go` is looked up at upgrade.go:50, before
  any install.
- **AC5 — pass (structural).** `cmdUpgrade` contains no reference to
  `internal/project` and no `.forge/` path; every upgrade test calls it with
  no project present and passes. `grep internal/project internal/cli/upgrade.go`
  is empty. Note: the contract's Test plan also promised a
  `Main([]string{"upgrade"}, …)` run in a project-less temp dir asserting
  exit 0 and no `.forge/`; that test was not written (see Problems).
- **AC6 — pass.** `TestUpgrade_PrintsOldAndNewVersions` (Version `dev`,
  `goVersion` `v0.3.0`) asserts output contains `dev` and
  `upgraded to v0.3.0`; `TestUpgrade_FallsBackToTheRefWhenVersionIsUnreadable`
  asserts `upgraded to latest`. `moduleVersion` reads `go version -m` and
  falls back to the requested ref (upgrade.go:75-78, 155-167).
- **AC7 — pass.** `go run . help` prints `forge upgrade [version]       upgrade the forge binary itself`
  (cli.go:20); `TestHelp_ListsUpgrade` and `TestDocs_DocumentUpgrade` pass;
  `docs/cli.md:29` has the `### forge upgrade [version]` section.

## Contract interface check

All promised symbols exist with the exact shape in
`internal/cli/upgrade.go`: `const forgeModule` (line 15),
`cmdUpgrade(args []string, stdout, stderr io.Writer) error` (31),
`cleanupStaleBinary()` (120), `replaceExecutable(exe, fresh, ref string) error`
(88), `copyFile(src, dst string) error` (109), `exeSuffix() string` (130),
and the seams `lookupGo = exec.LookPath`, `goInstall`, `goVersion`,
`executable = os.Executable`, `renameFile = os.Rename`,
`removeFile = os.Remove` (19-26). `cli.go` has the dispatch case (64) and
`cleanupStaleBinary()` as the first line of `Main` (46). `git diff 79fa4f4 HEAD`
shows `main.go`, `internal/project/*` and `go.mod` unchanged; no new
dependency. The sidecar order (`old` remove → rename → copy → best-effort
remove, with rollback) matches decisions 7 and 11.

## Problems

Blocking the merge: **none**.

Non-blocking:

1. **Promised AC5 test is missing.** The contract's Test plan requires a
   `cli.Main([]string{"upgrade"}, …)` test in a `.forge`-less `t.TempDir()`
   asserting exit 0 and no `.forge/` created. Neither `upgrade_test.go` nor
   `upgrade_internal_test.go` calls `Main` with `upgrade`, so the
   project-independence is proven structurally and by direct `cmdUpgrade`
   calls rather than end to end.
2. **AC1 streaming is not exercised by a test.** `goInstall` is stubbed in
   every case, so `cmd.Stdout`/`cmd.Stderr` forwarding is only verified by
   reading `installModule`.
3. **docs/cli.md intro is now inaccurate.** Lines 3-5 still say every
   command reads/writes `.forge/` and "None of them reach the network",
   which `forge upgrade` contradicts: it touches no `.forge/` and runs the
   Go toolchain. Pre-existing wording, not updated by this change.

Checks: `go test ./... -count=1` ok; `gofmt -l .` empty; `go vet ./...`
clean; `go run . help` lists the command. `go run . validate` exits 1 only
because this `review.md` did not exist before now, which is expected.
No test invokes a real self-install and none reaches the network.

## Proposed conventions

- Commands that shell out or touch the OS take their calls through
  package-level function variables that default to the real function
  (`lookupGo`, `goInstall`, `executable`, `renameFile`, `removeFile`), and
  the internal test file replaces them with a generic `swap` helper
  registered on `t.Cleanup`. It keeps `go test ./...` hermetic: no test
  runs the real toolchain.
