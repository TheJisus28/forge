# Review — SPEC-021

Verdict: pass with notes

Reviewed by the `reviewer` role on
`spec/021-check-the-criteria-before-approving-and-across`. Commits under
review: `f8e895e` (verifiable criterion + approve refusal), `db32c71`
(validate coverage), `5f1d810` (templates, roles, docs), `697b70c`
(comment-aware, whole-token coverage) and `0f52a92` (agreed conventions and
this review), on top of the contract `3fd16fa` (contract hash
`0037636337b0`). This is a read-only review: no `status` change and no
archive. Re-verified after the post-review fix; everything below reflects
current HEAD.

## Acceptance criteria

- AC1: pass — `Criterion.Verifiable()` (`internal/project/project.go:384`)
  is the single anchor rule and `cmdApprove` (`internal/cli/work.go:263`)
  refuses beside the open-questions gate, before `ApprovedBy`/`ContractHash`
  are set; the rule is stated in contract decision 1.
  `go test ./... -run TestCriterion_Verifiable` and
  `-run TestApprove_RefusesUnverifiableCriterion` PASS. Independent scratch:
  `forge approve SPEC-001` on `- AC1: The UI is fast` printed
  `forge: SPEC-001 has criteria that name no command, test or response:` /
  `AC1` and exited 1, leaving `status: contracting` and no `contract_hash`;
  rewriting it to `` `GET /ui` returns the page `` printed
  `SPEC-001 approved by reviewer, contract a5e8877eb305.` and exited 0 with
  `status: planning`, `approved_by: reviewer`, `contract_hash: a5e8877eb305`.
  The fix did not touch this path.
- AC2: pass — `Spec.CriterionGaps()` (`internal/project/project.go:445`)
  strips HTML comments with `doc.StripComments` from the whole `tasks.md`
  text (line 464) and from the review's `## Acceptance criteria` section
  (line 470) before matching, and `hasToken` (`internal/project/project.go:409`)
  compares whole `[0-9A-Za-z_-]+` tokens case-insensitively;
  `cmdCheck` (`internal/cli/check.go:20`) prints
  `<ID>: <AC> has no task in <rel>` / `<ID>: <AC> has no evidence in <rel>`.
  `go test ./internal/cli/ -run TestCheck_ -v` PASS (`TestCheck_ReportsUncoveredCriteria`,
  `TestCheck_CommentedEvidenceIsNotCoverage`, `TestCheck_TemplateReviewLeavesCriterionUncovered`,
  `TestCheck_ImplementingGapExitsZero`, `TestCheck_IsDeterministicAndWritesNothing`);
  `go test ./internal/project/ -run TestCriterionGaps -v` PASS (including the
  extended adjacency and `AC1-`/`AC10`/`AC1x` sub-cases). Independent scratch
  at `done`: a commented-only evidence line reported
  `SPEC-001: AC1 has no evidence in .../review.md` and exit 1; the real
  `forge template review` output likewise left AC1 uncovered and exit 1;
  `AC1 AC2` and `AC1,AC2` in tasks and in the evidence section both counted
  (`1 specs checked, every criterion is covered`, exit 0); `AC1-`, `AC10` and
  `AC1x` satisfied neither `AC1` nor `AC2` (all four gap lines, exit 1); ids
  under `## Notes` are still not evidence.
- AC3: pass — `cmdCheck` returns 1 for an evidence gap on a `reviewing` or
  `done` spec (`internal/cli/check.go:72`).
  `TestCheck_ReportsUncoveredCriteria`,
  `TestCheck_CommentedEvidenceIsNotCoverage` and
  `TestCheck_TemplateReviewLeavesCriterionUncovered` PASS. Scratch
  `forge check` at `done` and at `reviewing` with no evidence printed the gap
  lines and exited 1; an `implementing` task gap printed and exited 0
  (`TestCheck_ImplementingGapExitsZero` PASS). Note: per contract decision 5
  a `done` spec whose task is missing but whose evidence is present exits 0
  — the criterion is settled, so it is not "uncovered".
- AC4: pass — `checkCriteriaCoverage` (`internal/validate/validate.go:293`)
  is called from `Run` beside `checkArtifacts` and maps each gap from
  `CriterionGaps()` to a `Finding`. `go test ./internal/validate/ -run
  TestRun_CriterionCoverageWarnsAndErrors -v` PASS. Independent scratch
  `forge validate` at `reviewing` printed `warning SPEC-001: AC1 has no
  evidence in ...` and exited 0; at `done` with no evidence it printed
  `error   SPEC-001: AC1 has no evidence in ...` and exited 1; at `done`
  with evidence present but the task missing it printed
  `warning SPEC-001: AC1 has no task in ...` and no error.
- AC5: pass — `docs/cli.md` has a `### forge check [id]` section (line 142)
  and the `forge approve` paragraph (line 70) states the verifiable rule;
  `docs/workflow.md`'s `## Acceptance criteria` names `forge check` and
  `forge validate`; `kit/machine/roles/reviewer.md` runs `forge check <id>`
  before writing the review. `TestDocPages_DocumentTheCriterionRule` and
  `TestTemplates_CarryNoRealCriterionId` PASS; `forge template tasks` shows
  `Moves: <criterion ids>` and `forge template review` keeps only the live
  table header plus a commented `ACn` example (no `\bAC\d+\b`).

## Problems

Blocking the merge: none. `go test ./...` all `ok`, `gofmt -l .` empty,
`go vet ./...` clean. `go run . check` prints the five SPEC-015 task gaps and
exits 0; `go run . check SPEC-021` prints `1 specs checked, every criterion
is covered` and exits 0; `go run . validate` prints the same five SPEC-015
warnings and exits 0. `forge check <id>` for SPEC-001..SPEC-015, SPEC-018
and SPEC-019 all exit 0 (SPEC-015 with its five task lines), so no delivered
spec changed exit code.

Residual risk #1 is closed. A `done` spec whose only `AC1` evidence is an
HTML comment such as `<!-- AC1: go test ./... is the evidence -->` inside
`## Acceptance criteria`
now reports `SPEC-001: AC1 has no evidence in .../review.md` (and AC2) and
exits 1; feeding the real `forge template review` output likewise leaves the
criterion uncovered and exits 1. `doc.StripComments` is the one rule, shared
by `cmdArchive`'s `stripComments` (which now delegates), and the unterminated
comment / `<--` typo cases are left intact (`TestStripComments` PASS). The
tokenizer replacement (`acTokenRe` → `hasToken`) is behaviour-equivalent for
the boundaries exercised; `AC1 AC2` and `AC1,AC2` both match, `AC1-`, `AC10`
and `AC1x` do not.

Not blocking, still open:

1. **AC4 / decision 5 wording.** Contract decision 5 keeps a `done` task
   gap a `Warning` while AC4's literal "as an error once the spec is `done`"
   would make both task and evidence coverage errors. The implementation
   follows the contract. Decision 5 also cites "OQ1" while
   `## Open questions` is `None.`, so the question is not recorded anywhere.
2. **`CHANGELOG.md` is not touched on this branch** although `AGENTS.md`
   says to update it in the same pull request as the change. The precedent
   (`f8bad28 spec(spec-015): archive`) adds it in the archive commit, so this
   may be intentional, but the branch as merged from `main` carries no entry
   for `forge check`.

No delivered spec regresses: the only `forge validate` output is the five
`warning SPEC-015: ACn has no task in .../tasks.md` lines (evidence present,
forward-only per decision 5); every other delivered spec has no
`## Existing state` and is not judged.

## Proposed conventions

The comment-stripping convention this review proposed was accepted and
recorded by the maintainer (decided 2026-09-20):
`.forge/conventions/coverage.md` (strip comments before matching tokens,
one shared `doc.StripComments`), `.forge/conventions/parsing.md` (whole-token
id matching instead of `\b`) and `.forge/conventions/architecture.md`
(`internal/validate` delegates the derivation to `internal/project`).
`tasks.md`'s `## Proposed conventions` is now `None.` Nothing further to
propose.