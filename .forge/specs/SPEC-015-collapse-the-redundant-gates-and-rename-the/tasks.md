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
- [ ] Phase 2 — `forge accept` is the single gate and the id authority.
  Where: `internal/cli/work.go` (`cmdNew`, `cmdAccept`, `renumberSpec`,
  `cmdRenumber`), `internal/project/git.go` (`RemoteSpecIDs`),
  `internal/cli/guard.go` (`commandDenial`), `docs/teams.md`, `docs/cli.md`,
  `kit/machine/roles/orchestrator.md`,
  `kit/claude/skills/forge-work/SKILL.md`, `kit/forge/specs/README.md`,
  `.forge/specs/README.md`; tests in `internal/cli/cli_test.go`,
  `internal/project/project_test.go`.
- [ ] Phase 3 — `forge migrate` and the reported names. Where:
  `internal/cli/migrate.go` (new), `internal/cli/cli.go`, `docs/cli.md`;
  tests in `internal/cli/cli_test.go`.
- [ ] Phase 4 — The survey lives once, in `spec.md`. Where:
  `kit/machine/templates/spec.md`, `kit/machine/templates/plan.md`,
  `internal/project/project.go` (`ExistingState`),
  `internal/validate/validate.go` (`planSurveysExisting`),
  `kit/machine/roles/architect.md`, `kit/machine/roles/orchestrator.md`,
  `kit/AGENTS.md`, `AGENTS.md`, `docs/workflow.md`; the survey is already in
  this spec's `spec.md`; tests in `internal/project/project_test.go`,
  `internal/validate/validate_test.go`, `internal/cli/machine_test.go`.
- [ ] Phase 5 — Docs hold the new name only. Where: the scanned pages and
  `internal/cli/cli_test.go` (`TestDocPages_UseContracting`),
  `internal/cli/machine_test.go`.

## Proposed conventions

None.
