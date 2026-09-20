# Tasks — SPEC-015

The phases from the plan, as checkboxes. Phases 1 and 2 of the earlier draft
are one run here: the collapse does not compile without the approval path.
Tick a phase when it lands and say where the work is, so a later spec knows
what exists without reading the diff.

- [x] Phase 1 — The state machine and the approval path. Where:
  `internal/workflow/workflow.go` (`Contracting`, `Canonical`, `All`,
  `transitions`, `InFlight`, `WaitingFor`, `Meaning`),
  `internal/project/project.go` (`FromDoc`), `internal/cli/guard.go`,
  `internal/cli/work.go` (`cmdStart`, `cmdApprove`),
  `internal/validate/validate.go` (`checkArtifacts`),
  `internal/view/view.go` (`Brief`), `kit/machine/roles/architect.md`,
  `kit/machine/roles/orchestrator.md`; tests in
  `internal/workflow/workflow_test.go`, `internal/project/project_test.go`,
  `internal/validate/validate_test.go`, `internal/view/view_test.go`,
  `internal/cli/cli_test.go`, `internal/cli/machine_test.go`.
  Landed 2026-09-20: the machine has the nine states in order and
  `workflow.Canonical` maps `specifying`/`awaiting-approval` to
  `contracting`; `project.FromDoc` reads through it; `cmdStart` sets
  `contracting` and `cmdApprove` moves straight to `planning` (still freezing
  `approved_by`/`contract_hash`); the guard message names `contracting`;
  `validate.checkArtifacts` moved its non-empty-contract case to `planning`
  and warns on a retired raw status naming `forge migrate`; `view.Brief`
  files `contracting` under `in flight`; both roles name `contracting`.
  Verified: `go test ./...` (all packages ok), `gofmt -l .` (clean),
  `go vet ./...` (clean), `go run . workflow` (nine states, `contracting`,
  no retired name), `go run . validate` (19 specs, no problems). The guard
  message was checked on a scratch repo: `contracting`, `specifying` and
  `awaiting-approval` frontmatter all deny as `contracting` naming
  `forge approve`.
  Known: `CHANGELOG.md:142` (the released 0.1.0 entry) still names the old
  states; it is release history and the phase 5 scan excludes it. The
  task's guard check "on this spec" cannot say `contracting` because
  SPEC-015 is itself `implementing`, where the guard allows product code.
- [x] Phase 2 — `forge accept` is the single gate and the id authority.
  Where: `internal/cli/work.go` (`cmdNew`, `cmdAccept`, `renumberSpec`,
  `cmdRenumber`), `internal/project/git.go` (`RemoteSpecIDs`),
  `internal/cli/guard.go` (`commandDenial`), `docs/teams.md`, `docs/cli.md`,
  `kit/machine/roles/orchestrator.md`,
  `kit/claude/skills/forge-work/SKILL.md`, `kit/forge/specs/README.md`,
  `.forge/specs/README.md`; tests in `internal/cli/cli_test.go`,
  `internal/project/project_test.go`.
  Landed 2026-09-20: `forge new` prints only `forge accept`; the loop pages
  name one entry into the queue; `project.RemoteSpecIDs(root, ref)` reads
  `git ls-tree -d --name-only <ref> .forge/specs/` without fetching;
  `cmdAccept` confirms the provisional id against `origin/main` then `main`,
  renumbers through the shared `renumberSpec` when taken (history line
  `renumbered from SPEC-NNN: taken on main`) and refuses when the id is
  referenced; `commandDenial` denies a `forge`/`forge.exe accept` segment
  inside compounds. Verified: `go test ./...` (all packages ok),
  `gofmt -l .` (clean), `go vet ./...` (clean),
  `go run . guard --explain --command "forge accept SPEC-020"` →
  `would deny: Forge: only a person accepts...`; `forge new "x" --capability
  workflow` in a scratch repo prints `forge accept SPEC-001` and no `intake`.
  Also fixed `.claude/skills/forge-work/SKILL.md`, the planted copy `forge
  update` refreshes, so the two do not disagree in this tree.
- [x] Phase 3 — `forge migrate` and the reported names. Where:
  `internal/cli/migrate.go` (new), `internal/cli/cli.go`, `docs/cli.md`;
  tests in `internal/cli/cli_test.go`.
  Landed 2026-09-20: `cmdMigrate [--dry-run]` loads through `project.Load`,
  rewrites the raw `status` scalar of every retired-name spec to
  `contracting` with `doc.SetStr` + `doc.Save` (the body, `## History`
  included, is byte-identical; no history line and no `updated` change),
  prints `rel/path: old -> new` for each, prints `nothing to migrate` and
  exits 0 on a current tree, and writes nothing under `--dry-run`; dispatched
  from `internal/cli/cli.go` and listed in the usage text and `docs/cli.md`.
  Confirmed unchanged and kept: `validate.checkLegacyStatus` already warns
  `status "specifying" is the old name for "contracting"; run forge migrate`
  (`TestRun_WarnsLegacyStatusName` still green) and `guard.denial`'s single
  `case workflow.Contracting` reports the new name for a retired frontmatter.
  Added `TestMigrate_RewritesRetiredStatus` (before/after `tree` snapshot
  proves `--dry-run` writes nothing; `## History` compared byte-for-byte; a
  second run prints `nothing to migrate`) and `TestGuard_NamesContracting`
  (`contracting` plus both retired names deny with `contracting` and
  `forge approve`, never a retired name). Verified: `go test ./...` (all
  packages ok), `gofmt -l .` (clean), `go vet ./...` (clean), and on a
  scratch copy `forge migrate --dry-run` printed `nothing to migrate` exit 0,
  then listed both retired specs and left their status lines unchanged; the
  real run rewrote only `status` and left `## History` intact; a third run
  printed `nothing to migrate`; `forge guard --file src/x` on a `specifying`
  scratch spec reported `SPEC-003 is contracting ... forge approve SPEC-003`.
  Decision recorded: `docs/cli.md`'s new section does not spell the retired
  names, because Phase 5's `TestDocPages_UseContracting` scans `docs/*.md`
  for them and the Phase 2 proposed convention chose to avoid the word rather
  than carve out an exception; the names live in the binary, the validate
  warning and this spec. `CHANGELOG.md` was not touched (Phases 1–2 did not
  either); the maintainer's release step owns it.
- [x] Phase 4 — The survey lives once, in `spec.md`. Where:
  `kit/machine/templates/spec.md`, `kit/machine/templates/plan.md`,
  `internal/project/project.go` (`ExistingState`),
  `internal/validate/validate.go` (`planSurveysExisting` → `checkSurvey`),
  `kit/machine/roles/architect.md`, `kit/machine/roles/orchestrator.md`,
  `kit/machine/roles/implementer.md`, `kit/AGENTS.md`, `AGENTS.md`,
  `docs/workflow.md`, `kit/claude/skills/forge-work/SKILL.md` and the
  planted `.claude/skills/forge-work/SKILL.md`; the survey is already in
  this spec's `spec.md`; tests in `internal/project/project_test.go`,
  `internal/validate/validate_test.go`, `internal/cli/machine_test.go`.
  Landed 2026-09-20: `kit/machine/templates/spec.md` gains `## Existing
  state` (HTML-comment guidance, empty body) between `## Contract` and
  `## Out of scope`, and the section is removed from
  `kit/machine/templates/plan.md`; `project.Spec.ExistingState()` returns
  `s.doc.Section("Existing state")` from `spec.md`, mirroring `Contract`
  and `OpenQuestions`; `validate.checkSurvey` warns a non-terminal spec
  past `accepted` (any `workflow.InFlight` state) whose section is empty,
  and `planSurveysExisting` is deleted, so the warning no longer reads
  `plan.md`; the architect role records the survey in `spec.md` and the
  orchestrator, implementer, `kit/AGENTS.md`, `AGENTS.md`,
  `docs/workflow.md` and both `forge-work` skill copies read it there.
  Added `TestExistingState_ComesFromSpec`,
  `TestRun_SurveyWarnsFromSpec` (replacing
  `TestRun_PlanWithoutExistingStateWarns`) and
  `TestTemplate_SurveySection`. Verified: `go test ./...` (all packages
  ok), `gofmt -l .` (clean), `go vet ./...` (clean), `go run . template
  spec` (carries `## Existing state`), `go run . template plan` (does not),
  `go run . validate` (`19 specs, no problems`).
  Beyond the contract's file list: `kit/machine/roles/implementer.md` was
  changed too, because it pointed at `plan.md`'s `## Existing state`, a
  section the phase removes; the planted `.claude/agents/` role copies were
  left as Phase 1 left them (the two disagreed already; `forge update`
  refreshes them).
- [ ] Phase 5 — Docs hold the new name only. Where: the scanned pages and
  `internal/cli/cli_test.go` (`TestDocPages_UseContracting`),
  `internal/cli/machine_test.go`.

## Proposed conventions

- **A negative docs scan means the forbidden word cannot appear at all.**
  Decision 3's test reads each page and rejects any `intake`; prose written
  to explain its absence ("there is no separate intake pull request") fails
  the test it is describing. Either the scan is a positive check on the
  replacement command, or the prose must avoid the word. I chose to avoid
  it and renamed `docs/teams.md`'s `## Intake` heading to `## Entering the
  queue`. Worth deciding which shape the next such scan takes.
- **`forge accept` refuses a referenced id only when it would renumber.**
  Decision 7's sentence does not say whether "already referenced" alone
  blocks acceptance or only blocks the renumber. I read it as the latter,
  because a proposed parent may already have children and refusing
  acceptance whenever anything points at the spec would make the parent
  impossible to accept. `TestAccept_RefusesWhenReferenced` covers the
  collision case.
- **No convention was needed for the git call.** `RemoteSpecIDs` had to pass
  `.forge/specs/` with a trailing slash: `git ls-tree -d` without it returns
  the `specs` tree entry, not its child folders. Recorded as an
  implementation note, not a rule.
- **A frontmatter-only rewrite edits the `doc.Doc` directly, not `Spec.Save`.**
  `Spec.Save` syncs every typed field, so a rename that must leave the body
  and the untouched keys byte-identical would rewrite `conductor`, lists and
  empty values as a side effect. `cmdMigrate` calls `s.Doc().SetStr` +
  `s.Doc().Save` for exactly the one scalar. Rule to decide: a command whose
  contract promises "only this field changes" edits the document, and
  `Spec.Save` stays for a state move that owns the whole record.
- **A maintenance command's output read both ways.** `forge migrate` prints
  `rel/path: old -> new` and the same lines under `--dry-run`, because the
  contract says the dry run "prints the same list"; a no-op prints a fixed
  `nothing to migrate`. Rule to decide: a `--dry-run` is not distinguished in
  the output, and the "nothing to do" line is the stable contract a test can
  pin. (The `docs/cli.md` section deliberately does not name the retired
  states, extending the first bullet above: Phase 5 scans `docs/*.md` for
  them, so the reference lives only in the binary and the validate warning.)
- **`CHANGELOG.md` stays untouched until the release step.** `AGENTS.md` says
  to update it in the pull request that changes behaviour, but Phases 1–2 did
  not add an `[Unreleased]` entry and Phase 3 matched them. Worth deciding
  whether a multi-phase spec writes one changelog entry at the end or each
  phase writes its own.
- **"A live spec past `accepted`" is `workflow.InFlight`.** Decision 5 did
  not enumerate the states the survey warning covers. I used
  `workflow.InFlight(s.Status)` — `contracting`, `planning`, `implementing`,
  `blocked`, `reviewing` — which skips `proposed`, `accepted` and the
  terminal states, so the rule reuses the state machine instead of a second
  list. That means a spec warns from `contracting` on, before the architect
  has necessarily written the survey; it is a warning, not a gate.
- **The template's HTML-comment guidance counts as a present section.**
  `checkSurvey` asks `strings.TrimSpace(s.ExistingState()) != ""`, so the
  guidance comment the new `spec.md` template ships makes `## Existing
  state` non-empty and the warning cannot fire on a spec created by `forge
  new` until someone replaces the comment. SPEC-019 solved this shape by
  stripping HTML comments before judging a placeholder (`stripComments` in
  `internal/cli`); decision 5 only said "a check of `s.ExistingState()`",
  and `internal/validate` cannot reach that unexported helper without a
  shared one. Worth deciding whether the survey check strips comments, and
  where the one comment-stripper lives.
