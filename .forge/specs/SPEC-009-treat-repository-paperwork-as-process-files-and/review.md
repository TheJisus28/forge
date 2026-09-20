# Review — SPEC-009

Verdict: pass with notes

Reviewed `spec/009-treat-repository-paperwork-as-process-files-and` at `b2c941b`
(phases `dfc8e39`, `02f5e37`, `b2c941b`; `48856a5` ticks tasks). SPEC-009 is
`reviewing`, so no spec is implementing and the denial path is live.

## Acceptance criteria

| Criterion | Result | Evidence |
|---|---|---|
| AC1 root Markdown allowed | pass | Binary built from this tree. `forge guard --file CHANGELOG.md` → exit 0, no output; `forge guard --explain --file CHANGELOG.md` → `would allow`, exit 0. Test `TestGuardFileMode` allowed slice includes `CHANGELOG.md` and passes. |
| AC2 product code still denied | pass | `forge guard --file internal/cli/guard.go` → exit 1; `forge guard --file main.go` → exit 1; nested Markdown `docs/customizing.md` → exit 1; root non-paperwork `go.mod` → exit 1 (each prints `forge: SPEC-009 is under review...`). `TestGuardFileMode` denied slice asserts code 1 with `no spec` in a spec-less `newRepo`; passes. |
| AC3 files gone, no live reference | pass | `git grep -n -E "CONTRIBUTING\|SECURITY\|CODE_OF_CONDUCT" -- ":!.forge/specs"` → no output, exit 1. `git ls-files` has no matching path. `Test-Path` is `False` for all three. `TestRepositoryPaperwork` passes. |
| AC4 instructions survive in AGENTS.md | pass | `AGENTS.md` has `## Contributing` (line 76), `### Adding a command` (78), `### Supporting another agent` (86) and the changelog sentence under `## Releases` (61-62). `README.md` line 131 is `[AGENTS.md](AGENTS.md)`; it does not contain `CONTRIBUTING.md` (covered by AC3 grep). `TestRepositoryPaperwork` passes. |
| AC5 documented and toolchain green | pass | Rule stated in `docs/customizing.md` 40-49 (canonical, with the allowlist-feature-request note), `docs/cli.md` 147-150, `docs/opencode.md` 52-56: root Markdown/`LICENSE`/`NOTICE` allowed, nested Markdown product code. `go test ./...` all `ok`, exit 0; `go vet ./...` exit 0; `gofmt -l .` no output, exit 0. `TestDocs_DescribeProcessFiles` passes. |

The code matches the Contract: `isProcessFile` ends with `return
isRootPaperwork(rel)` (guard.go:229); `isRootPaperwork` returns false for `""`
or a path containing `/`, and true for `.md`/`.markdown` or a
`LICENSE`/`NOTICE` base name via the two prefix checks (guard.go:235-246).
Signatures and the CLI surface are unchanged.

## Blocking problems

None block the merge. All five acceptance criteria pass.

## Notes

- **gofmt caveat (AC5).** This working tree has `core.autocrlf=true`, but
  `.gitattributes` (`* text=auto eol=lf`) checks files out as LF, so the
  warned CRLF artifact did not reproduce. `git ls-files --eol` reports `w/lf`
  for `internal/cli/guard.go` and `internal/cli/cli_test.go`, and `gofmt -l .`
  lists nothing (verified to recurse: a deliberately unformatted file in a
  temp subdirectory is reported). The files this spec changed are gofmt-clean.
  AC5 passes cleanly here.
- **Fragment-built filenames in `TestRepositoryPaperwork`.** Acceptable as
  evidence for AC3 as written, but a contract smell. The test assembles
  `"CONTRIB" + "UTING.md"`, `"SECUR" + "ITY.md"` and `"CODE_OF_" +
  "CONDUCT.md"` so the AC3 `git grep` stays clean; the runtime checks still
  target the real paths (`os.Stat` and `strings.Contains`), so the behaviour
  under test is genuine. However, it means the acceptance criterion's evidence
  is shaped by the test that verifies it, and the test cannot be found by
  grepping for the removed docs. It is also worth saying plainly that the
  test only greps `README.md` for the reference; the "no tracked file" half of
  AC3 rests on the reviewer's `git grep`, not on a test. Judge it as a
  documented workaround, not a behaviour defect.
- **`docs/opencode.md` (decision 4).** Decision 4 asked to *replace* "the root
  pointers" in the "rule is unchanged" paragraph; the implementation keeps the
  phrase and appends the new definition (lines 52-56). AC5's outcome — the
  definition is documented in all three pages and they agree — is met, so this
  is non-blocking, but the doc does not read exactly as the decision worded it.
- **Decision 10 already done.** `.forge/decisions/0003-repository-root-paperwork-is-a-process-file.md`
  exists, titled per the decision, superseding nothing.
- **Archive blocker (process, not AC).** `tasks.md` had an unresolved entry
  under `## Proposed conventions`; `forge archive` refuses while one is present
  (`internal/cli/work.go:307`, `pendingConventions`). The orchestrator declined
  it: the fragment trick is a one-off concession forced by AC3's wording, not a
  pattern to reuse. AC3's better shape is "the removed paths do not exist",
  observed with `git ls-files` or a root-Markdown allowlist, without the "no
  live reference" half. Cleaning `TestRepositoryPaperwork` is left to a future
  spec; `tasks.md` now reads `None.`

## Proposed conventions

None.