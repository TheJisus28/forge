# Plan — SPEC-009

Phases, in order. One phase is one run of the implementer and one commit:
small enough to verify, large enough to mean something.

## Existing state

- Delivered specs this builds on: SPEC-004 added `forge guard --file` and
  the command denial, so the guard is a tested command with a hook-free
  mode other agents call. SPEC-007 keeps Forge's machinery in the binary,
  so the rule lives in `internal/cli`, not in `kit/`. SPEC-005/007 own the
  documentation of the guard.
- Reuse `internal/cli/guard.go`: `isProcessFile` already normalises to a
  slash-relative path, allows paths outside the repository, the four
  process prefixes and the exact root names. This adds one condition, not a
  new module.
- Reuse `internal/cli/cli_test.go`: `newRepo` is a `guard: on` repository
  with no spec implementing, and `TestGuardFileMode`/`TestBriefAndGuard`
  already drive `forge guard --file` and `--explain`; the allowed/denied
  slices extend there. `upgrade_test.go` already reads repository files
  from `../../`, the pattern the AC3/AC4/AC5 black-box tests reuse.
- Reuse `internal/cli/guard.go`'s existing style: unexported helpers with
  comments that say what the rule is for (`commandDenial`, `pushesToDefault`,
  `onDefaultBranch`).
- Conventions in `.forge/conventions/testing.md`: a substitution point is
  chosen per case. The new rule is pure string logic with no external
  dependency, so it is called directly and gets no seam.
- Duplication avoided: no configurable allowlist, no second place that
  derives the rule, no new package, no `kit/` change. The rule is stated
  once in code and mirrored in the three pages that already describe the
  guard, kept in step by tests.
- What genuinely has to be built: `isRootPaperwork` and its call from
  `isProcessFile`; the extended AC1/AC2 tests; the AC3/AC4/AC5 black-box
  tests; the three deletion commits and the `AGENTS.md`/`README.md`/
  `CHANGELOG.md` edits; the rule text in `docs/customizing.md`,
  `docs/cli.md` and `docs/opencode.md`.

## Phase 1 — the guard rule

- Scope: `internal/cli/guard.go` gains `isRootPaperwork(rel string) bool`
  and `isProcessFile` ends with `return isRootPaperwork(rel)` after its
  prefix loop and exact-name switch. `internal/cli/cli_test.go` extends the
  allowed slice of `TestGuardFileMode` with `CHANGELOG.md` and adds a
  denied slice (`internal/cli/guard.go`, `main.go`, `docs/customizing.md`,
  `go.mod`).
- Done when: with `guard: on` and no spec implementing, `forge guard --file
  CHANGELOG.md` exits 0 while `forge guard --file main.go` and
  `forge guard --file docs/customizing.md` exit 1. Moves AC1 and AC2.
- Verify with: `go test ./...`.

## Phase 2 — drop the community docs and migrate what matters

- Scope: delete `CONTRIBUTING.md`, `SECURITY.md` and `CODE_OF_CONDUCT.md`
  (the last is already deleted in the working tree). Add `## Contributing`
  to `AGENTS.md` with `### Adding a command` and `### Supporting another
  agent` from `CONTRIBUTING.md`, and the changelog sentence under
  `## Releases`. `README.md` points its Contributing section at `AGENTS.md`
  instead of `CONTRIBUTING.md`. `CHANGELOG.md` records the guard rule and
  the removal under `[Unreleased]` without naming the removed files. Add
  the AC3/AC4 black-box tests to `internal/cli/cli_test.go`.
- Done when: the three paths are absent, `AGENTS.md` carries both
  migrated sections, `README.md` does not mention `CONTRIBUTING.md`, and
  `git grep -n -E "CONTRIBUTING|SECURITY|CODE_OF_CONDUCT" -- ":!.forge/specs"`
  prints nothing. Moves AC3 and AC4.
- Verify with: `go test ./...` and the `git grep` above.

## Phase 3 — document the rule

- Scope: `docs/customizing.md` (`## The guard`, the canonical paragraph),
  `docs/cli.md` (`### forge guard`) and `docs/opencode.md` (`## The
  guard`) state the root-paperwork rule, the root-only caveat and that a
  configurable allowlist is out of scope. Add the AC5 black-box test that
  reads `docs/customizing.md`.
- Done when: all three pages agree with `isRootPaperwork`, and the AC5 test
  reads the definition. Moves AC5.
- Verify with: `go test ./...`, `gofmt -l .` and `go vet ./...`.

## Risks

- A repository whose product is root-level Markdown loses the guard for
  those files. Accepted in decision 0003 and stated in the docs; the escape
  is a spec branch or `guard: off`.
- The three guard pages drift apart again. Mitigated by the AC5 test and by
  keeping the canonical text in `docs/customizing.md`.
- `git grep` for AC3 also matches this spec's own history. The evidence
  command excludes `.forge/specs/` by design (decision 8); the record is
  never rewritten.
