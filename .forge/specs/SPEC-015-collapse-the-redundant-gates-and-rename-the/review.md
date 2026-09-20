# Review — SPEC-015

Verdict: pass with notes

Reviewed on branch `spec/015-collapse-the-redundant-gates-and-rename-the`
at `6e863d0`, against contract `652e4ef43814`. Product code was not touched
by this review; `status` was not changed.

## Acceptance criteria

| Criterion | Result | Evidence |
|---|---|---|
| AC1 | pass | `internal/workflow/workflow.go`: `Contracting State = "contracting"` (line 19), `All()` returns `proposed, accepted, contracting, planning, implementing, blocked, reviewing, done, dropped` (30-31), `InFlight` includes `Contracting` (156), `WaitingFor(Contracting)` is `architect: write the contract` (170), and `Canonical` maps the retired names (38-44). `go run . workflow` prints ``| `contracting` | The contract is being written | architect: write the contract |`` and no retired name. `go run . roles architect` says "The spec is `contracting` while you write this" and `go run . roles orchestrator` says "Launch the architect while the spec is `contracting`". A scratch repo carrying `status: specifying` reports `SPEC-001  contracting` in `forge status`. `rg -n "specifying\|awaiting-approval" docs AGENTS.md README.md kit` returns no matches. Tests: `TestAll_UsesContracting`, `TestCheck_LegalAndIllegalMoves`, `TestCheck_CannotSkipApproval`, `TestInFlightAndTerminal`, `TestWorkflowCommand_RendersTheStatesFromGo`, `TestDocPages_UseContracting`. |
| AC2 | pass | End to end in scratch `ac2`: `forge new` -> `forge accept` -> `forge start` (status `contracting`) -> write the `## Contract` body -> `forge approve SPEC-001`, with **no** `forge advance`. The spec ends `status: planning` with `approved_by: TheJisus28` and `contract_hash: 8ee24165c4f2`, exactly the hash `forge approve` printed, and its `## History` is `accepted`, `contracting`, `planning` — no `awaiting-approval` line. Test: `TestApprove_StraightFromContracting`; also `TestRun_MissingArtifacts` (empty-contract case now gated at `planning`) and `TestBrief_ContractingIsInFlight`. |
| AC3 | pass | One gate: `forge new "..." --capability workflow` in scratch `ac3b` prints "then accept it into the queue" and `forge accept SPEC-001`, with no `intake`; `rg -i intake docs/teams.md docs/cli.md kit/forge/specs/README.md kit/claude/skills/forge-work/SKILL.md .forge/specs/README.md AGENTS.md kit/AGENTS.md README.md` returns no matches; `docs/teams.md`'s section is `## Entering the queue` with "one entry into the queue". Id authority (contract decision 7): a repo whose `main` carries `.forge/specs/SPEC-001-from-main/spec.md` while the branch provisionally holds `SPEC-001`; `forge accept SPEC-001` printed `SPEC-001 was taken on main; renumbered to SPEC-002`, moved the folder to `SPEC-002-collides-with-main`, wrote `id: SPEC-002` and `accepted_by: TheJisus28`, and the history line `- 2026-09-20  accepted  by TheJisus28: renumbered from SPEC-001: taken on main`; `forge status` then shows `SPEC-002`. Residual race: scratch `ac6` with two `SPEC-020` specs makes `forge validate` print `error   SPEC-020: duplicate id, also in spec.md; run forge renumber` and exit 1; `forge renumber SPEC-020` prints `SPEC-020 is now SPEC-021` and a second `forge validate` is clean. Tests: `TestNew_DescribesOneGate`, `TestAccept_RenumbersWhenTakenOnMain`, `TestAccept_RefusesWhenReferenced`, `TestRenumber_ResolvesTheRace`, `TestGuard_DeniesForgeAcceptForAgents`. |
| AC4 | pass | `forge template spec` contains `## Existing state`; `forge template plan` does not. `Spec.ExistingState()` reads `spec.md`'s section (`internal/project/project.go:386`), `checkSurvey` warns from it (`internal/validate/validate.go:242`), and `planSurveysExisting` is gone (`rg` finds no such symbol or a `plan.md` read). `kit/machine/roles/architect.md` records the survey in the spec's `## Existing state`; `orchestrator.md`'s `## Before planning` reads it there; `docs/workflow.md`, `AGENTS.md`, `kit/AGENTS.md` and the `forge-work` skill say the same. Tests: `TestTemplate_SurveySection`, `TestExistingState_ComesFromSpec`, `TestRun_SurveyWarnsFromSpec`. |
| AC5 | pass | Scratch `ac5` carrying `status: specifying` and `status: awaiting-approval`: `forge status` reports both as `contracting`; `forge migrate --dry-run` printed `.forge/specs/SPEC-001-old/spec.md: specifying -> contracting` and `.forge/specs/SPEC-002-old/spec.md: awaiting-approval -> contracting`, and the SHA256 of every file under `.forge` was identical before and after (writes nothing); `forge migrate` rewrote only the `status` scalar while the `## History` lines `... specifying ...` / `... awaiting-approval ...` stayed byte-identical; a second `forge migrate` printed `nothing to migrate` and exited 0 with no legacy `forge validate` warning. Before migrating, `forge validate` warned `status "specifying" is the old name for "contracting"; run forge migrate` and the same for `awaiting-approval`. `forge guard --file src/x` on the `specifying` spec exited 1 with `forge: SPEC-001 is contracting; write the Contract section first, then: forge approve SPEC-001` — the new name and no retired name. `forge --help` lists `forge migrate [--dry-run]`, documented in `docs/cli.md`. Tests: `TestMigrate_RewritesRetiredStatus`, `TestGuard_NamesContracting`, `TestRun_WarnsLegacyStatusName`. |

Full gate: `go test ./...` all packages `ok`; `gofmt -l .` and `go vet ./...`
produce no output; `go run . validate` is clean after this file exists (before
it, the only finding was this spec `reviewing` without its `review.md`).

## Blocking problems

None.

## Notes

- `CHANGELOG.md` was not updated. `AGENTS.md` says "Update `CHANGELOG.md` in
  the same pull request as the change". This is a behaviour change (rename,
  deleted state, `forge migrate`, one-gate accept, the survey move) and the
  `[Unreleased]` section's newest entry is still SPEC-018's. The immediate
  precedent is inconsistent — the delivered SPEC-019 also merged without an
  entry — so I did not treat it as blocking, but one `[Unreleased]` line
  should be added in this pull request if the team enforces AGENTS.md.
- The survey check is softer than it looks: `kit/machine/templates/spec.md`
  ships `## Existing state` with an HTML-comment guidance block, and
  `checkSurvey` only asks `strings.TrimSpace(s.ExistingState()) != ""`, so the
  comment counts as content. Reproduce: a spec created by `forge new` and
  moved to an in-flight state (scratch `ac2`, `status: planning`) makes
  `forge validate` print only the missing-test-command warning, never
  `spec.md has no Existing state section`. The contract's test only pins the
  section-absent case, so this is not an AC failure, but the warning cannot
  fire on the normal creation path until the check strips comments (the same
  `stripComments` shape SPEC-019 introduced). The implementer flagged this in
  `tasks.md` as a convention to decide.
- SPEC-015's own `plan.md` still carries a full `## Existing state` section
  and its `spec.md` says the full survey is there "during this spec". The
  change is template/role level and `forge template plan` no longer ships the
  section, so this is a legacy artifact of the spec that made the change, not
  a regression; the validate warning is satisfied because the spec's own
  `## Existing state` is non-empty.
- The hand-off brief reported "SPEC-019's contract changed after approval".
  At the branch tip neither `forge validate` nor `forge brief` reports any
  drift, and SPEC-019's recorded `contract_hash` (`c32a9d7895fe`) matches its
  contract, so the note was stale.

## Proposed conventions

None.