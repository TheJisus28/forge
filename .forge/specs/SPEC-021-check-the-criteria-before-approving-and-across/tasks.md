# Tasks — SPEC-021

The phases from the plan, as checkboxes. One phase is one implementer run
and one commit. Tick a phase when it lands and say where the work is, so a
later spec knows what exists without reading the diff.

- [x] Phase 1 — A criterion is verifiable, and `forge approve` refuses the
  ones that are not. Moves: AC1. Where: `internal/project/project.go`
  (`Criterion.Verifiable`), `internal/cli/work.go` (`cmdApprove`); tests in
  `internal/project/project_test.go`, `internal/cli/cli_test.go`.
  Landed: `Criterion.Verifiable` tests the trimmed `Text` against three
  anchor patterns — an inline code span with a non-space character
  (`codeSpanRe`), the whole word `test`/`tests` case-insensitively
  (`testWordRe`) or a `Test[A-Za-z0-9_]*` identifier (`testNameRe`), and the
  case-insensitive observable verbs (`outcomeRe`). `cmdApprove` walks
  `s.Criteria()` after the open-questions refusal and before
  `approved_by`/`contract_hash` are set, and returns an error naming the
  `!Verifiable()` ids and telling the author to name the command, the test or
  the response that settles each one.
  Tests added: `TestCriterion_Verifiable` (a backticked command, `test`,
  `TestName`, `returns` and `refuses` are verifiable; `Works well`,
  `The system is fast`, a no-anchor text, an all-space code span and
  `supersedes AC10` are not). `TestApprove_RefusesUnverifiableCriterion` (a
  non-empty contract and `None.` questions with `- AC1: The UI is fast`
  refuses, names `AC1` and leaves `contract_hash` unset; rewriting it with a
  backticked command approves to `planning` and fingerprints the contract).
  Four existing approval tests (`TestLifecycle`,
  `TestHierarchyAndDependencies`, `TestApprove_StraightFromContracting`,
  `TestStatusShowsSupersedes`) now give their criteria an anchor, because the
  new gate rejects the template's `- AC1: ...` placeholders.
  Verified: `go test ./...` all `ok`; `gofmt -l .` empty; `go vet ./...`
  clean; the two new tests pass with `-v`.
- [x] Phase 2 — The coverage derivation, in one place. Moves: AC2, AC4.
  Where: `internal/project/project.go` (`CriterionGap`,
  `Spec.CriterionGaps`); tests in `internal/project/project_test.go`.
  Landed: `CriterionGap{Criterion, File, Kind}` and
  `Spec.CriterionGaps()` — declaration order, a task gap (from
  `implementing` on) before an evidence gap (from `reviewing` on) for the
  same criterion, and nothing at all when `## Existing state` is empty
  (forward-only, decision 5). `tasks.md` is matched as a whole file;
  `review.md` is scoped through `doc.Load(...).Section("Acceptance
  criteria")`. A missing or unreadable artifact leaves its record empty, so
  every applicable criterion is a gap. The bounded token is built by
  `acTokenRe(id)`; `taskGapApplies`/`evidenceGapApplies` hold the state sets.
  Tests added: `TestCriterionGaps_MatchesBoundedTokens` (reviewing spec,
  `AC1..AC3`; tasks name `AC1,AC3`, the review row covers `AC1`; gaps are
  `AC2`/task, `AC2`/evidence, `AC3`/evidence; `AC10` and `AC1x` match
  nothing; an `AC2` mention under `## Notes` is not evidence) and
  `TestCriterionGaps_AppliesByState` (task gap from `implementing`, evidence
  gap only from `reviewing`, none for `proposed`/`accepted`/`contracting`/
  `planning`/`dropped`, and none at any state without `## Existing state`).
  Verified: `go test ./...` all `ok`; `gofmt -l .` empty; `go vet ./...`
  clean; `go test ./internal/project/ -run TestCriterionGaps -v` both PASS.
- [x] Phase 3 — `forge check`. Moves: AC2, AC3. Where: `internal/cli/check.go`
  (new), `internal/cli/cli.go`; tests in `internal/cli/cli_test.go`.
  Landed: `cmdCheck(args []string, out, errOut io.Writer) int`, dispatched
  from `Main` beside `cmdValidate` (`case "check": return cmdCheck(rest,
  stdout, stderr)`) and listed in the usage text. It loads through
  `project.Load(cwd())`, reads `spec.md`, `tasks.md` and `review.md`, and
  writes nothing. With an id it checks that spec whatever its state; without
  one, every spec in `implementing`, `blocked`, `reviewing` or `done`, in id
  order (`p.Specs` is sorted). It prints one line per gap from
  `Spec.CriterionGaps()` — `<ID>: <AC> has no task in <rel>` /
  `<ID>: <AC> has no evidence in <rel>`, the path relative to the repository
  root with forward slashes. When nothing is reported and at least one
  selected spec declares criteria it prints `<N> specs checked, every
  criterion is covered` (N is the selected-spec count); when nothing is
  selected, `no spec is building or reviewing yet; nothing to check`. It
  exits 1 only for an evidence gap on a `reviewing` or `done` spec
  (decision 5's failure), 0 otherwise. A load or unknown-id failure prints
  `forge: ...` to stderr and returns 1.
  Contract correction: decision 2 names the signature
  `func cmdCheck(args []string, out io.Writer) int` but also requires a load
  failure on stderr "like `cmdValidate`". `cmdValidate` takes a separate
  `errOut io.Writer`, and `cmdCheck` does too, with `Main` passing `stderr`,
  so the error does not bypass the `Main` test seam. The rest of the
  signature and the dispatch match the contract.
  Tests added in `internal/cli/cli_test.go`, with a `checkSpec` helper that
  writes a criterion plus the `## Existing state` marker:
  `TestCheck_ReportsUncoveredCriteria` (a `done` spec whose `tasks.md` and
  review's `## Acceptance criteria` omit the criterion prints one `no task`
  line and one `no evidence` line naming the relative file, and exits 1),
  `TestCheck_ImplementingGapExitsZero` (an `implementing` task gap prints and
  exits 0) and `TestCheck_IsDeterministicAndWritesNothing` (two runs are
  byte-identical; the `.forge` tree and `git status --porcelain` are
  unchanged).
  Verified: `go test ./...` all `ok`; `gofmt -l .` empty; `go vet ./...`
  clean; `go test ./internal/cli/ -run TestCheck_ -v` — all three PASS;
  `go run . check` reports SPEC-015's five task gaps (warnings, evidence
  present) and exits 0.
- [x] Phase 4 — `forge validate` includes the coverage. Moves: AC4. Where:
  `internal/validate/validate.go` (`checkCriteriaCoverage`); tests in
  `internal/validate/validate_test.go`.
  Landed: `Run` calls `checkCriteriaCoverage(p, s, add)` right after
  `checkArtifacts`, beside the parent-to-child `checkCoverage` it leaves
  untouched. The new rule walks `s.CriterionGaps()` and turns each gap into a
  finding with the `forge check` messages — `<AC> has no task in <rel>` /
  `<AC> has no evidence in <rel>`, the path relative to `p.Root` with forward
  slashes. Severity per decision 5: a task gap is always a `Warning` (in
  flight and at `done`), an evidence gap is a `Warning` in flight and an
  `Error` at `done`. The contract's signature
  `checkCriteriaCoverage(p *project.Project, s *project.Spec, add func(...))`
  is what shipped; the implementer prompt's `(p *project.Project) []string`
  and its "evidence gap on `reviewing` is an error" both contradict contract
  decision 5 / AC4 (`review.md` is a warning until `done`), so the contract
  and plan won.
  Test added: `TestRun_CriterionCoverageWarnsAndErrors` (an `implementing`
  spec missing a task warns with `AC1 has no task in` and produces no error;
  a `done` spec whose review's `## Acceptance criteria` omits `AC1` errors
  with `AC1 has no evidence in`; a `done` spec missing only the task warns
  and produces no error). Each spec carries `## Existing state` so
  `CriterionGaps` judges it.
  Regression check: `go run . validate` prints five warnings and exits 0 —
  `warning SPEC-015: AC1..AC5 has no task in
  .forge/specs/SPEC-015-.../tasks.md`. SPEC-015 is the only delivered spec
  with `## Existing state`, its review's `## Acceptance criteria` covers all
  five criteria, and its `tasks.md` predates criterion naming, so the gaps
  are task warnings at `done`, exactly decision 5's forward-only outcome. No
  delivered spec fails, the rule was not weakened, and `Spec.CriterionGaps`
  needed no change: the `## Existing state` gate Phase 2 already added is the
  forward-only exception. SPEC-021 itself is clean (its tasks name
  AC1..AC5 and evidence does not apply while `implementing`).
  Verified: `go test ./...` all `ok`; `gofmt -l .` empty; `go vet ./...`
  clean; `go test ./internal/validate/ -run
  TestRun_CriterionCoverageWarnsAndErrors -v` PASS; `go run . validate` as
  above (exit 0).
- [ ] Phase 5 — Templates, roles and docs. Moves: AC5. Where:
  `kit/machine/templates/review.md`, `tasks.md`, `spec.md`, `docs/cli.md`,
  `docs/workflow.md`, `kit/machine/roles/reviewer.md`; tests in
  `internal/cli/machine_test.go`, `internal/cli/cli_test.go`.

## Proposed conventions

- **A bounded id token must exclude `-` as well as word characters.**
  Go's `\b` matches `AC1` inside `AC1-` (a hyphen is not a word character),
  so `acTokenRe` uses `(?:^|[^0-9A-Za-z_-])` on each side instead of `\b`.
  A later matcher of ids in artifacts should reuse that boundary rather than
  write `\b<id>\b`, which would silently accept `AC1-`.
- **A command that owns an exit code takes `errOut` and prints its own
  failure, instead of returning the error to `Main`.** `cmdValidate` is the
  precedent and `cmdCheck` follows it: a command that returns `int` has no
  way to reach `Main`'s `forge: ...` handler, so it writes `forge: ...` to a
  separate `errOut io.Writer` and returns 1 itself. That keeps the test seam
  working — `Main`'s `stderr` argument is what a test replaces — rather than
  writing to `os.Stderr` directly. A later `int`-returning command should
  take `(args []string, out, errOut io.Writer) int`.
- **A `validate` rule that judges a spec's artifacts delegates the
  derivation and only assigns severity.** `checkCriteriaCoverage` does not
  re-read `tasks.md` or `review.md`: it calls `project.Spec.CriterionGaps()`
  and maps each gap to a `Finding`, so the token matcher and the state table
  live in one place (the same single-source rule SPEC-018 enforced for the
  workflow). A later rule over the same artifacts should reuse the
  `internal/project` derivation rather than grow a second reader.
