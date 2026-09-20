---
id: SPEC-021
title: Check the criteria before approving and across artifacts
status: implementing
capability: workflow
created: 2026-09-20
updated: 2026-09-20
accepted_by: TheJisus28
orchestrator: TheJisus28
approved_by: TheJisus28
contract_hash: 0037636337b0
---

## Problem

`forge approve` refuses only while `## Open questions` is non-empty. Nothing
checks that an acceptance criterion is verifiable, and `forge validate`
checks structure — ids, references, cycles, coverage between parent and
child — but not whether every criterion has a task that delivers it and, at
review time, an evidence line that settles it. A vague criterion, or a
criterion no task mentions, passes the gates untouched.

## Acceptance criteria

- AC1: `forge approve` refuses, or warns with a clear message, when a
  criterion cannot be verified by a command, a test, or a request and its
  response; the rule is stated in the contract of this spec.
- AC2: A read-only `forge check` reports each criterion with no matching
  task in `tasks.md` and each criterion with no evidence line in
  `review.md`, naming the criterion and the file it is missing from.
- AC3: `forge check` exits non-zero when a criterion is uncovered while the
  spec is `reviewing` or `done`.
- AC4: `forge validate` includes the criterion-to-task and
  criterion-to-evidence coverage as a warning, and as an error once the spec
  is `done`.
- AC5: `docs/cli.md` documents the check and the criterion rule.

## Open questions

None.

## Contract

### Decisions

**1. A criterion is verifiable when its text names the evidence that settles
it; `forge approve` refuses while one does not.**
`internal/project/project.go` gains `func (c Criterion) Verifiable() bool`,
the one place the rule lives. A criterion is verifiable when its trimmed
`Text` contains at least one *anchor*:

- an inline code span: a pair of backticks with a non-space character
  between them (`\`forge check\``, `\`tasks.md\``, a test name, a field);
- a test reference: the whole word `test`/`tests` (case-insensitive) or an
  identifier matching `Test[A-Za-z0-9_]*`;
- an observable outcome: one of the whole words (case-insensitive)
  `returns`, `prints`, `outputs`, `exits`, `succeeds`, `fails`, `refuses`,
  `rejects`, `reports`, `lists`, `names`, `matches`, `emits`, `responds`.

Those are the three evidence kinds `docs/workflow.md` already names (a
command, a test, a request and its response), and they are the anchors every
criterion already in `.forge/specs/` writes. `cmdApprove` in
`internal/cli/work.go` gains a check beside the empty-contract and
open-questions refusals: it walks `s.Criteria()`, collects the ids where
`!c.Verifiable()`, and when the list is non-empty returns an error naming
them and saying to name the command, the test or the response that settles
each one. Discards: a warning instead of a refusal (the vague criterion
would still be approved, which is the problem the spec names); an explicit
`verify:`/`[manual]` tag (new syntax for every criterion and a parser
change); checking criterion shape in `forge validate` too (approval is the
gate AC1 names, and `forge check` stays about coverage). The fast lane
(SPEC-016), which skips approval, is out of scope.

**2. `forge check [id]` is a new read-only command that reports the criteria
no task and no evidence settles.**
New `internal/cli/check.go` with `func cmdCheck(args []string, out io.Writer) int`,
dispatched from `internal/cli/cli.go` the way `cmdValidate` is
(`case "check": return cmdCheck(rest, stdout)`) and added to the usage text
and `docs/cli.md`. It loads through `project.Load(cwd())`, reads `spec.md`,
`tasks.md` and `review.md`, and writes nothing: no `Save`, no `git`, no
network, so a before/after tree hash and `git status --porcelain` are
unchanged (the `forge capabilities` determinism test, extended).

- Selection: with an id (`project.NormalizeID`), that spec; without one,
  every spec whose status is `implementing`, `blocked`, `reviewing` or
  `done`, in id order (`p.Specs` is already sorted). An unknown id is a
  returned error naming `forge status`.
- Output: one line per gap from `Spec.CriterionGaps()` (decision 3):
  `<ID>: <AC> has no task in <relative path>` and
  `<ID>: <AC> has no evidence in <relative path>`, the path relative to the
  repository root with forward slashes. When the selected specs have
  criteria and no gaps, `<N> specs checked, every criterion is covered`;
  when nothing is selected,
  `no spec is building or reviewing yet; nothing to check`.
- Exit code: `1` when a reported gap is a failure (decision 5) on a spec
  that is `reviewing` or `done`; `0` otherwise, including a gap on an
  `implementing` or `blocked` spec, where the review does not exist yet. A
  load failure prints `forge: ...` to stderr and returns `1`, like
  `cmdValidate`.

Discards: reporting evidence gaps while a spec is still `implementing`
(every criterion of every in-flight spec would read as uncovered before the
reviewer writes anything); a `--json` or `--quiet` flag (no caller needs one
yet); leaving the report to `forge validate` alone (the reviewer needs one
focused, read-only command before archiving).

**3. Criterion-to-task and criterion-to-evidence coverage is derived once, in
`internal/project`, and both `forge check` and `forge validate` consume it.**
`internal/project/project.go` gains:

```go
// CriterionGap is one criterion the spec's artifacts do not settle.
type CriterionGap struct {
    Criterion Criterion
    File      string // absolute path of the artifact it is missing from
    Kind      string // "task" or "evidence"
}

// CriterionGaps lists every criterion no task in tasks.md delivers and no
// evidence line in review.md settles, for the spec's current state.
func (s *Spec) CriterionGaps() []CriterionGap
```

Matching is a bounded, case-insensitive literal token: a criterion has a
task when `tasks.md` contains `\bAC<n>\b`, so `AC1` never matches `AC10`,
`AC1x` or `AC1-`. Evidence is read from the `## Acceptance criteria` section
of `review.md` (`doc.Section`, the SPEC-018 fixed English heading) and not
from the whole file, so a mention under `## Notes` or
`## Blocking problems` is not evidence. Applicability follows when each file
is the record: a task gap is reported from `implementing` on; an evidence
gap from `reviewing` on (the states `validate.checkArtifacts` requires the
files for). Criteria keep declaration order and a task gap precedes an
evidence gap. A missing or unreadable file reports the applicable kind for
every criterion: the file names nothing.

`internal/validate/validate.go` gains
`func checkCriteriaCoverage(p *project.Project, s *project.Spec, add func(Severity, string, string, ...any))`,
called from `Run` right after `checkArtifacts`. Each gap becomes a `Finding`
whose severity is decision 5's, with the same two messages as `forge check`
(`<AC> has no task in <rel>` / `<AC> has no evidence in <rel>`). The existing
`checkCoverage` (parent to child) is untouched; the new name keeps them
apart. Discards: a second token matcher in `internal/cli` and a second state
table in `internal/validate` (the drift SPEC-018 removed elsewhere); reading
the whole `review.md` (the template's example rows and any prose would count
as evidence); folding `checkArtifacts` into this.

**4. The shipped templates stop carrying tokens the check reads as real, and
the docs state the rule.**
Following SPEC-019: a template must not ship a literal `AC<digit>`, or a
fresh file copied from it would read as covered.
`kit/machine/templates/review.md` moves its example rows into an HTML
comment and writes the placeholder as `ACn`, leaving the live
`| Criterion | Result | Evidence |` table header only.
`kit/machine/templates/tasks.md` tells the author to name the criteria each
phase moves, with the placeholder `moves: <criterion ids>` and no digit.
`kit/machine/templates/spec.md`'s Acceptance criteria guidance names the
three evidence kinds. `docs/cli.md` gains a `forge check [id]` section and
changes the `forge approve` and `forge validate` paragraphs to the new gates
(AC5); `docs/workflow.md`'s `## Acceptance criteria` names `forge check`
beside the existing human rule; `kit/machine/roles/reviewer.md` runs
`forge check <id>` before writing the review; `internal/cli/cli.go`'s usage
lists `forge check`. Discards: leaving the review template's `AC1`/`AC2`
rows (a reviewer who copies it and forgets to edit would look covered — the
SPEC-019 bug one artifact over); deleting the guidance instead of commenting
it (the section becomes unexplained).

**5. Coverage is forward-only, and the `done`-state error keys on evidence.**
`Spec.CriterionGaps()` returns nothing for a spec whose `## Existing state`
section is empty. That section is the marker SPEC-015 introduced for a spec
written under the current workflow, and decision 0004 keeps a delivered spec
without it as history that is not re-judged. On this repository that leaves
SPEC-015 alone: its `review.md` covers every criterion, so it never errors.
In flight (`implementing`, `blocked`, `reviewing`) a missing task and a
missing evidence line are both `Warning`s, and `forge check` prints both.
At `done`, a missing evidence line is an `Error` in `forge validate` and the
non-zero exit in `forge check`, because `review.md` is the durable proof; a
missing task with evidence present stays a `Warning`, because the criterion
is settled and the delivered `tasks.md` that never named criteria are
history. `forge check` still prints that task gap (AC2) but counts only an
evidence gap as a failure.
Discards: erroring on a missing task at `done` (every delivered spec whose
`tasks.md` predates this rule would fail `forge validate`, and the only fix
is rewriting history, which decision 0004 rejects); gating on a new
frontmatter key (a second thing to keep in sync, no more precise than the
section the workflow already requires); skipping `forge check` for
delivered specs (it reports them, it just does not fail them on a task gap).
OQ1 asks whether the stricter reading of AC4 is intended.

### Interfaces other specs build against

- `project.Criterion.Verifiable() bool` is the criterion rule; a later
  reader of criteria uses it instead of restating the anchors.
- `project.CriterionGap{Criterion, File, Kind}` and
  `project.Spec.CriterionGaps() []CriterionGap` are the coverage
  derivation, including the forward-only gate. `forge check` and
  `internal/validate` are the callers; a later spec adds a consumer, not a
  second matcher.
- `forge check [id]` is read-only and exits non-zero only for a failure gap
  (decision 5) on a `reviewing` or `done` spec. `forge approve` still sets
  `approved_by` and `contract_hash`; the new refusal runs before either,
  beside the open-questions refusal.
- `forge validate` gains criterion coverage; `--quiet` (errors only) hides
  the task warnings with no change.
- The states, transitions, frontmatter keys and file layout are unchanged:
  no new state and no new key.

### Tests

- `internal/project/project_test.go:TestCriterion_Verifiable` — a table:
  a backticked command, a `test`/`TestName`, and a `returns`/`refuses`
  criterion are verifiable; `Works well`, `The system is fast` and a text
  with no anchor are not; `AC10` in the text does not change `AC1`.
- `internal/project/project_test.go:TestCriterionGaps_MatchesBoundedTokens`
  — a `reviewing` spec with `AC1..AC3`: `tasks.md` naming `AC1, AC3` and a
  `review.md` Acceptance criteria row for `AC1` yield gaps `AC2`/task,
  `AC2`/evidence and `AC3`/evidence; `AC10` and `AC1x` in either file match
  nothing.
- `internal/project/project_test.go:TestCriterionGaps_AppliesByState` — a
  task gap from `implementing` on, an evidence gap only from `reviewing`; a
  `planning`/`proposed` spec yields none, and a spec without
  `## Existing state` yields none at every state.
- `internal/validate/validate_test.go:TestRun_CriterionCoverageWarnsAndErrors`
  — an `implementing` spec missing a task is a `Warning`; a `done` spec
  missing an evidence line is an `Error`; a `done` spec missing only a task
  is a `Warning`.
- `internal/cli/cli_test.go:TestApprove_RefusesUnverifiableCriterion` — with
  a non-empty contract and `None.` questions, a criterion `The UI is fast`
  makes `forge approve` exit 1 and name the id; rewriting it with a
  backticked command approves the spec to `planning` with `contract_hash`
  set.
- `internal/cli/cli_test.go:TestCheck_ReportsUncoveredCriteria` — a `done`
  spec whose `tasks.md` and `review.md` omit a criterion: `forge check`
  prints the criterion and the relative file in one `no task` line and one
  `no evidence` line, and exits 1.
- `internal/cli/cli_test.go:TestCheck_ImplementingGapExitsZero` — an
  `implementing` spec with a task gap prints it and exits 0.
- `internal/cli/cli_test.go:TestCheck_IsDeterministicAndWritesNothing` — two
  runs are byte-identical and the tree hash plus `git status --porcelain`
  are unchanged.
- `internal/cli/machine_test.go:TestTemplates_CarryNoRealCriterionId` —
  `forge template tasks` and `forge template review` contain no `\bAC\d+\b`
  (the templates are read through `kit.Template`).
- `internal/cli/cli_test.go:TestDocPages_DocumentTheCriterionRule` —
  `docs/cli.md` has a `forge check` section and the `forge approve` section
  states the rule; no page restates the state machine (the SPEC-018 scan
  extended).

### Out of scope of this contract

- Changing what counts as an acceptance criterion, how many a spec needs, or
  how they are parsed (`project.criterionRe`).
- Checking criterion shape anywhere other than `forge approve`: `forge
  check` and `forge validate` report coverage, not verifiability.
- Gating `forge archive` on coverage; the review is where `forge check` is
  run.
- The fast lane (SPEC-016), which skips approval, and a backfill of the
  delivered specs' `tasks.md` (OQ1).
- A machine-readable output (`--json`) or CI wiring beyond the existing
  `forge validate` exit code.
- Rewriting delivered `.forge/specs/` or `.forge/decisions/` records.

## Existing state

Recorded by the architect; planning reads this instead of copying it.

- `internal/project/project.go` already parses criteria (`Spec.Criteria`,
  `criterionRe`), owns the artifact paths (`TasksPath`, `ReviewPath`) and
  reads `tasks.md` for `TaskProgress`. `Criterion`, `CriterionGap` and both
  new functions sit here. `Spec.ExistingState()` already reads the marker
  decision 5 uses.
- `internal/validate/validate.go` already separates `Warning` from `Error`
  (`Severity`, `Finding`), gates artifacts by state in `checkArtifacts`, and
  owns the parent-to-child `checkCoverage`; the new `checkCriteriaCoverage`
  sits beside them on the same types.
- `internal/doc` already reads sections with the fence-aware `Section`, so
  `review.md`'s evidence is scoped to `## Acceptance criteria`; the SPEC-019
  fix and its comment-stripping are the precedent for not treating template
  text as content.
- `internal/cli` already has the thin command shape (`cmdValidate` returns
  an exit code from `Main`), the `warning: ` / returned-error convention
  (`cli-output`), and `specArg`/`project.NormalizeID` for an optional id.
- A command that writes nothing is already proven by a before/after tree
  snapshot (`TestCapabilities_IsDeterministicAndWritesNothing`, the testing
  convention); `forge check` follows it.
- SPEC-015 already foresaw this spec (`forge check` reads `Spec.Criteria`,
  `tasks.md`/`review.md`, and its `forge approve` check sits in
  `cmdApprove`); SPEC-018 fixed the English headings and the single-source
  rule this reuses; decision 0004 fixes the history this must not rewrite.

What genuinely does not exist: `Criterion.Verifiable`, `CriterionGaps`, the
`forge check` command, and the `forge approve` refusal.

## Out of scope

- The contents of the workflow: no state, transition, role or gate other
  than the two this spec adds (`forge approve`'s criterion refusal and
  `forge check`) is changed.
- The fast lane (SPEC-016) and the state-name history (SPEC-015, SPEC-018).
- Rewriting delivered `.forge/specs/` or `.forge/decisions/` records, and
  the `tasks.md` backfill OQ1 raises.
- Changing `forge guard`, `forge archive` or the pull-request gate.

## History

Written by `forge`. Do not edit by hand.
- 2026-09-20  accepted  by TheJisus28: renumbered from SPEC-017: taken on main
- 2026-09-20  contracting  by TheJisus28
- 2026-09-20  planning  by TheJisus28
- 2026-09-20  implementing  by orchestrator
