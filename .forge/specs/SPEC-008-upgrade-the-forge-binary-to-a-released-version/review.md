# Review — SPEC-008

Verdict: **pass with notes**

Reviewed branch `spec/008-upgrade-the-forge-binary-to-a-released-version`
at `3984c11` (re-review of the first pass at `ff49aef`, which raised three
non-blocking notes). Evidence commands were run on Windows.

## Criteria

- **AC1 — pass.** `go install github.com/TheJisus28/forge@latest` is what
  runs: `TestUpgrade_InstallsLatestIntoATempDir` records ref `latest` and
  asserts stdout contains `installing github.com/TheJisus28/forge@latest`.
  `installModule` (upgrade.go:141) sets `cmd.Stdout`/`cmd.Stderr` to the
  writers, and `Main` returns 0 when `cmdUpgrade` returns nil
  (cli.go:101-108). Note: streaming the child's output is verified by
  inspection only — every test stubs `goInstall` (still true, see
  Problems).
- **AC2 — pass.** `TestUpgrade_UsesTheGivenVersionVerbatim` records ref
  `v0.2.0` for args `["v0.2.0"]`, used verbatim; `installModule` builds
  `github.com/TheJisus28/forge@v0.2.0`. `forge upgrade v0.2.0` ⇒
  `@v0.2.0`.
- **AC3 — pass.** `go test ./... -count=1` → all 8 packages `ok`; the root
  package includes `nonet_test.go`, which enforces the banned-import list
  (`TestBinaryMakesNoNetworkCalls` → PASS). `internal/cli/upgrade.go`
  imports only `fmt`, `io`, `os`, `os/exec`, `path/filepath`, `runtime`,
  `strings`; no banned import anywhere under review.
- **AC4 — pass.** `TestUpgrade_WithoutGoExplainsHowToInstallIt` stubs
  `lookupGo` to `exec.ErrNotFound`, asserts the error contains `Go 1.22+`
  and `go install github.com/TheJisus28/forge@latest`, and asserts
  `goInstall` was never called. `go` is looked up at upgrade.go:50, before
  any install.
- **AC5 — pass.** `cmdUpgrade` contains no reference to `internal/project`
  and no `.forge/` path (`Select-String` over upgrade.go is empty). The
  contract's promised end-to-end test now exists:
  `TestUpgrade_OutsideAProjectCreatesNoForgeDirectory`
  (upgrade_internal_test.go:239) chdirs into a `.forge`-less
  `t.TempDir()`, calls `Main([]string{"upgrade"}, …)`, asserts exit 0 and
  that no `.forge/` was created → PASS.
- **AC6 — pass.** `TestUpgrade_PrintsOldAndNewVersions` (Version `dev`,
  `goVersion` `v0.3.0`) asserts output contains `dev` and
  `upgraded to v0.3.0`; `TestUpgrade_FallsBackToTheRefWhenVersionIsUnreadable`
  asserts `upgraded to latest`. `moduleVersion` reads `go version -m` and
  falls back to the requested ref (upgrade.go:75-78, 155-167).
- **AC7 — pass.** `go run . help` prints `forge upgrade [version]       upgrade the forge binary itself`
  (cli.go:20); `TestHelp_ListsUpgrade` and `TestDocs_DocumentUpgrade` pass;
  `docs/cli.md:31` has the `### forge upgrade [version]` section. The page
  intro is now true for the whole CLI (lines 3-7).

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

The re-review diff `git diff ff49aef..3984c11` contains only the intended
changes: the `docs/cli.md` intro rewrite, the new `Main` test, the two
`spec.md` history lines, and this `review.md`. No product code moved.

## Problems

Blocking the merge: **none**.

Non-blocking:

1. **AC5's end-to-end test — resolved at `3984c11`.** The contract's Test
   plan promised a `cli.Main([]string{"upgrade"}, …)` run in a
   `.forge`-less `t.TempDir()` asserting exit 0 and no `.forge/` created.
   It is now `TestUpgrade_OutsideAProjectCreatesNoForgeDirectory`
   (upgrade_internal_test.go:239), and `go test ./internal/cli -run
   TestUpgrade_OutsideAProjectCreatesNoForgeDirectory -v` → PASS.
2. **`docs/cli.md` intro — resolved at `3984c11`.** Lines 3-7 now say most
   commands use `.forge/`, `forge upgrade` works outside a project and does
   not touch the kit, and the Go toolchain is one of the user's own network
   tools. The page is true for the whole CLI.
3. **AC1 streaming is still verified by inspection only.** Every upgrade
   test still replaces `goInstall` (`stubToolchain`,
   upgrade_internal_test.go:48), so `cmd.Stdout`/`cmd.Stderr` forwarding
   in `installModule` (upgrade.go:144-145) has no test that runs a child
   process. This is not a contract failure — the contract's Test plan
   requires every upgrade test to stub the toolchain and forbids a real
   `go install` — but it remains a note.
4. **Team decision on the proposed convention — recorded.** The test-seam
   pattern (`lookupGo`, `goInstall`, `executable`, `renameFile`,
   `removeFile` as package-level variables, swapped in tests) was raised as
   a convention. The team adopted a more general version, now written to
   `.forge/conventions/testing.md`: a substitution point is allowed when a
   test needs it to stay hermetic and deterministic, but a mutable
   package-level variable is not the mandatory dependency-injection
   mechanism; prefer the simplest shape the case allows.

Checks: `go test ./... -count=1` ok (all 8 packages); `gofmt -l .` empty;
`go vet ./...` clean; `go run . help` lists the command; `go run .
validate` → `8 specs, no problems` (exit 0). No test invokes a real
self-install and none reaches the network.

## Proposed conventions

None.
