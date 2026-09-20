# Plan — SPEC-021

The survey lives once, in `spec.md`'s `## Existing state` (the architect
wrote it). Phases below reuse what it names instead of rebuilding it.

## Phase 1 — A criterion is verifiable, and `forge approve` refuses the ones that are not

- Scope: `internal/project/project.go` gains
  `func (c Criterion) Verifiable() bool`, the one place the anchor rule lives
  (an inline code span; the whole word `test`/`tests` or a `Test[A-Za-z0-9_]*`
  identifier; one of the observable-outcome verbs in decision 1).
  `internal/cli/work.go:cmdApprove` walks `s.Criteria()`, collects the ids
  where `!Verifiable()`, and refuses beside the empty-contract and
  open-questions refusals, before `approved_by`/`contract_hash` are set.
  Tests: `internal/project/project_test.go:TestCriterion_Verifiable`,
  `internal/cli/cli_test.go:TestApprove_RefusesUnverifiableCriterion`.
- Done when: a spec with a criterion `The UI is fast` makes `forge approve`
  exit 1 and name the id; rewriting it with a backticked command approves to
  `planning` with the contract hash set. Moves AC1.
- Verify with: `go test ./...`, `gofmt -l .`, `go vet ./...`.

## Phase 2 — The coverage derivation, in one place

- Scope: `internal/project/project.go` gains the `CriterionGap` struct and
  `func (s *Spec) CriterionGaps() []CriterionGap`: a task gap when `tasks.md`
  has no bounded `\bAC<n>\b` for the criterion (from `implementing` on), an
  evidence gap when the `## Acceptance criteria` section of `review.md` has
  none (from `reviewing` on), declaration order, a task gap before an evidence
  gap, and nothing at all for a spec whose `## Existing state` is empty
  (decision 5). `internal/validate` and `forge check` will both consume it.
  Tests: `TestCriterionGaps_MatchesBoundedTokens`,
  `TestCriterionGaps_AppliesByState`.
- Done when: the table in the contract's `### Tests` section holds. Moves
  AC2, AC4.
- Verify with: `go test ./...`, `gofmt -l .`, `go vet ./...`.

## Phase 3 — `forge check`

- Scope: new `internal/cli/check.go:cmdCheck(args []string, out io.Writer) int`,
  dispatched from `internal/cli/cli.go` the way `cmdValidate` is and listed in
  the usage text. Selection, two-line output and exit code follow decision 2;
  it writes nothing, runs no git and no network. Tests:
  `internal/cli/cli_test.go:TestCheck_ReportsUncoveredCriteria`,
  `TestCheck_ImplementingGapExitsZero`,
  `TestCheck_IsDeterministicAndWritesNothing` (extend the
  `forge capabilities` before/after tree snapshot).
- Done when: a `done` spec with an uncovered criterion prints one `no task`
  line and one `no evidence` line naming the relative file and exits 1; an
  `implementing` task gap prints and exits 0; two runs are byte-identical and
  leave the tree unchanged. Moves AC2, AC3.
- Verify with: `go test ./...`, `gofmt -l .`, `go vet ./...`, `go run . check`.

## Phase 4 — `forge validate` includes the coverage

- Scope: `internal/validate/validate.go` gains
  `checkCriteriaCoverage(p *project.Project, s *project.Spec, add func(Severity, string, string, ...any))`,
  called from `Run` right after `checkArtifacts`; each gap from
  `CriterionGaps()` becomes a `Finding` with the same two messages as
  `forge check`. Severity per decision 5: both are `Warning` in flight; at
  `done` a missing evidence line is an `Error` and a missing task stays a
  `Warning`. `checkCoverage` (parent to child) is untouched. Test:
  `TestRun_CriterionCoverageWarnsAndErrors`.
- Done when: an `implementing` spec missing a task warns; a `done` spec
  missing an evidence line errors; a `done` spec missing only a task warns.
  Moves AC4.
- Verify with: `go test ./...`, `gofmt -l .`, `go vet ./...`, `go run . validate`.

## Phase 5 — Templates, roles and docs

- Scope: `kit/machine/templates/review.md` moves its example rows into an HTML
  comment and writes the placeholder `ACn`, leaving only the
  `| Criterion | Result | Evidence |` header live; `kit/machine/templates/tasks.md`
  prompts `moves: <criterion ids>`; `kit/machine/templates/spec.md`'s
  Acceptance criteria guidance names the three evidence kinds; `docs/cli.md`
  gains a `forge check` section and updates the `forge approve` and
  `forge validate` paragraphs; `docs/workflow.md` names `forge check`;
  `kit/machine/roles/reviewer.md` runs `forge check <id>` before writing the
  review. Tests: `internal/cli/machine_test.go:TestTemplates_CarryNoRealCriterionId`,
  `internal/cli/cli_test.go:TestDocPages_DocumentTheCriterionRule`.
- Done when: no shipped template carries a literal `\bAC\d+\b`, `docs/cli.md`
  documents the check and the approve rule, and no page restates the state
  machine. Moves AC5.
- Verify with: `go test ./...`, `gofmt -l .`, `go vet ./...`,
  `go run . template tasks`, `go run . template review`.

## Risks

- Phase 1 starts enforcing on this repository at once: every criterion of a
  spec approved from then on must carry an anchor. SPEC-021's own five already
  do; SPEC-016 must be written that way.
- `CriterionGaps` matching is a heuristic token (decision 3); a criterion
  named only inside `AC1x` or `AC10` does not match, by design.
- Forward-only (decision 5, OQ1): a delivered spec without `## Existing state`
  is never judged. SPEC-021 must keep its survey or `forge check` goes silent
  on it.
- Phase 4 runs on every spec; the first pass warns on in-flight specs whose
  `tasks.md` predates the rule. That is intended: name the criteria in
  `tasks.md`, do not loosen the rule.
- Scope creep: SPEC-016 (fast lane) skips approval and is not built here. Do
  not add a state, a transition, a frontmatter key or a flag.
