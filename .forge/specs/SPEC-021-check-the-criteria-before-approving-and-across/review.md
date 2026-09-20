# Review — SPEC-021

Verdict: pass with notes

Reviewed by the `reviewer` role on
`spec/021-check-the-criteria-before-approving-and-across`. Commits under
review: `f8e895e` (verifiable criterion + approve refusal), `db32c71`
(validate coverage), `5f1d810` (templates, roles, docs), on top of the
contract `3fd16fa` (contract hash `0037636337b0`). This is a read-only
review: no `status` change and no archive.

## Acceptance criteria

- AC1: pass — `Criterion.Verifiable()` (`internal/project/project.go:384`)
  is the single anchor rule and `cmdApprove` (`internal/cli/work.go:257`)
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
- AC2: pass — `Spec.CriterionGaps()` derives the gaps once
  (`internal/project/project.go:435`) and `cmdCheck`
  (`internal/cli/check.go:20`) prints
  `<ID>: <AC> has no task in <rel>` / `<ID>: <AC> has no evidence in <rel>`.
  `go test ./internal/cli/ -run TestCheck_ -v` PASS (reports, in-flight
  exit 0, deterministic and writes nothing). Independent scratch at `done`
  printed `SPEC-001: AC1 has no task in
  .forge/specs/SPEC-001-boundary/tasks.md` and `SPEC-001: AC1 has no
  evidence in .forge/specs/SPEC-001-boundary/review.md`. Boundaries hold:
  `AC10` and `AC1-` did not satisfy `AC1` (still reported as gaps), and
  `AC1`/`AC2` under `## Notes` were not evidence (both still reported).
- AC3: pass — `cmdCheck` returns 1 for an evidence gap on a `reviewing` or
  `done` spec (`internal/cli/check.go:70`). `TestCheck_ReportsUncoveredCriteria`
  PASS. Scratch `forge check` at `done` and at `reviewing` with no evidence
  printed the gap lines and exited 1; an `implementing` task gap printed and
  exited 0 (`TestCheck_ImplementingGapExitsZero` PASS). Note: per contract
  decision 5 a `done` spec whose task is missing but whose evidence is
  present exits 0 — the criterion is settled, so it is not "uncovered".
- AC4: pass — `checkCriteriaCoverage` (`internal/validate/validate.go:287`)
  is called from `Run` beside `checkArtifacts` and maps each gap to a
  `Finding`. `go test ./internal/validate/ -run
  TestRun_CriterionCoverageWarnsAndErrors -v` PASS. Independent scratch
  `forge validate` at `reviewing` printed `warning SPEC-001: AC1 has no
  evidence in ...` and exited 0; at `done` with no evidence it printed
  `error   SPEC-001: AC1 has no evidence in ...` and exited 1; at `done`
  with evidence present but the task missing it printed
  `warning SPEC-001: AC1 has no task in ...` and no error.
- AC5: pass — `docs/cli.md` has a `### forge check [id]` section (line 142)
  and the `forge approve` paragraph states the verifiable rule (lines
  70-78); `docs/workflow.md`'s `## Acceptance criteria` names `forge check`
  and `forge validate`; `kit/machine/roles/reviewer.md` runs
  `forge check <id>` before writing the review. `TestDocPages_DocumentTheCriterionRule`
  and `TestTemplates_CarryNoRealCriterionId` PASS; `forge template tasks`
  shows `Moves: <criterion ids>` and `forge template review` keeps only the
  live table header plus a commented `ACn` example (no `\bAC\d+\b`).

## Problems

Blocking the merge: none. `go test ./...`, `gofmt -l .` and `go vet ./...`
are clean on the branch (`exit 0`, empty output).

Not blocking:

1. **An HTML comment inside `## Acceptance criteria` counts as evidence.**
   `doc.Section` does not strip comments, and `CriterionGaps` reads it
   directly, so a hand-written `<!-- AC1 and AC2 are covered -->` in the
   evidence section cleared both evidence gaps in a scratch `done` spec and
   made `forge check` exit 0, even with an empty evidence table. The shipped
   template is safe because its commented example uses `ACn` (verified: the
   exact `forge template review` output fed as `review.md` still reported
   `AC1 has no evidence`), so this is not an AC failure but it is the
   SPEC-019 trap one level down. `cmdArchive` already strips comments with
   `stripComments` (`internal/cli/work.go:586`); the derivation could do the
   same.
2. **AC4 / decision 5 wording.** Contract decision 5 keeps a `done` task
   gap a `Warning` while AC4's literal "as an error once the spec is `done`"
   would make both task and evidence coverage errors. The implementation
   follows the contract. Decision 5 also cites "OQ1" while
   `## Open questions` is `None.`, so the question is not actually recorded
   anywhere.
3. **`CHANGELOG.md` is not touched on this branch** although `AGENTS.md`
   says to update it in the same pull request as the change. The precedent
   (`f8bad28 spec(spec-015): archive`) adds it in the archive commit, so
   this may be intentional, but the branch as merged from `main` carries no
   entry for `forge check`.

No delivered spec regresses: after this review is written the only
`forge validate` output on `main` is the five `warning SPEC-015: ACn has no
task in .../tasks.md` lines (evidence present, forward-only per decision 5);
every other delivered spec has no `## Existing state` and is not judged.

## Proposed conventions

See `tasks.md`'s `## Proposed conventions`. One addition proposed by this
review:

- **A derivation that reads a section strips HTML comments before matching
  a token.** `project.Spec.CriterionGaps` trusts `doc.Section`, which keeps
  comments, so a comment naming a real criterion reads as evidence;
  `cmdArchive` already strips them with `stripComments`. Any later matcher
  over a section populated by a template should strip comments, or the
  template's example must stay inert.
