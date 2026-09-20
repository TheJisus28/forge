# Tasks — SPEC-018

The phases from the plan, as checkboxes. One phase is one implementer run
and one commit. Tick a phase when it lands and say where the work is, so a
later spec knows what exists without reading the diff.

- [x] Phase 1 — Workflow states: one declaration, rendered. Where:
  `internal/workflow/workflow.go` (`Meaning`, `WaitingFor`),
  `kit/machine/WORKFLOW.md` (marker),
  `internal/cli/machine.go` (`cmdWorkflow`, `renderWorkflow`), tests in
  `internal/workflow/workflow_test.go`, `internal/cli/machine_test.go`,
  `internal/cli/machine_internal_test.go`, `kit/machine_test.go`.
  Landed: `workflow.Meaning` covers every `All()` state; `WaitingFor(Blocked)`
  is `orchestrator: clear the blocker`; `WORKFLOW.md` holds
  `<!-- forge:states -->` after `## States`; `renderWorkflow` replaces the
  marker with a table built from `All()`/`Meaning()`/`WaitingFor()` and errors
  naming `kit/machine/WORKFLOW.md` when the marker is absent. Verified with
  `go test ./...` (all packages ok), `gofmt -l .` (no output), `go vet ./...`
  (no output) and `go run . workflow` (table generated in lifecycle order).
- [x] Phase 2 — Pages stop restating; `docs/cli` matches behaviour. Where:
  `docs/workflow.md`, `AGENTS.md`, `kit/AGENTS.md`, `README.md`,
  `docs/cli.md`, tests in `internal/cli/cli_test.go`.
  Landed: the loop diagram and `## States` table are gone from
  `docs/workflow.md`, replaced by a pointer at `forge workflow`; the loop block
  is gone from `AGENTS.md`, `kit/AGENTS.md` and `README.md`, each pointing at
  `forge workflow`; `docs/cli.md`'s `forge start` paragraph now says start
  creates the folder and records fingerprints, not `plan.md`/`tasks.md`, which
  planning writes after `forge approve`.
  `TestDocs_DoNotRestateTheStateMachine` scans `AGENTS.md`, `kit/AGENTS.md`,
  `README.md` and `docs/*.md` for state arrow chains and state table rows
  built from `workflow.All()`; `TestDocs_WorkflowPointsAtForgeWorkflow` and
  `TestDocs_ForgeStartDoesNotCreatePlanningFiles` pin the pointer and the
  corrected command. Verified with `go test ./...` (all packages ok),
  `gofmt -l .` (no output) and `go vet ./...` (no output); the three new tests
  pass with `-v`, and a standalone reproduction confirmed the patterns match
  the pre-change pages and ignore the Conventional Commits arrow in
  `AGENTS.md`.
- [x] Phase 3 — One name for the driver: `orchestrator`. Where:
  `internal/project/project.go` (`Conductor` → `Orchestrator`, `FromDoc`,
  `Save`), `internal/cli/work.go` (`cmdStart`), `internal/view/view.go`,
  `docs/teams.md`, `docs/cli.md`, `AGENTS.md`; tests in
  `internal/project/project_test.go`, `internal/view/view_test.go`,
  `internal/cli/cli_test.go`.
  Landed: `project.Spec.Orchestrator` replaces `Conductor`; `orchestratorFrom`
  reads `orchestrator` then falls back to the read-only legacy `conductor`,
  and `Save` writes `orchestrator` with `setOrDelete` and always deletes
  `conductor`. `cmdStart` assigns `s.Orchestrator` and its `--by` help says
  "who orchestrates this spec"; `view.Brief` and `view.Detail` use the field
  and print `orchestrator`. `docs/teams.md`, `docs/cli.md` (the `forge start`
  paragraph now records the orchestrator) and the release step in `AGENTS.md`
  say orchestrator; `.forge/` records are untouched and still read through the
  fallback (`go run . status SPEC-018` prints `orchestrator   TheJisus28` from
  the on-disk `conductor:`).
  Tests added: `TestSave_MigratesLegacyConductorKey` (legacy loads, next save
  rewrites and drops the key), `TestDetail_ShowsOrchestrator`,
  `TestStart_RecordsOneOrchestratorName` (after `new`/`accept`/`start --by
  ana` the spec has `orchestrator: ana`, no `conductor:`, and `forge status`
  shows it), `TestDocs_DoNotSayConductor` (scans `AGENTS.md`,
  `kit/AGENTS.md`, `kit/machine/WORKFLOW.md`, `docs/*.md` and `forge roles`
  output). Verified with `go test ./...` (all packages ok), `gofmt -l .` (no
  output), `go vet ./...` (no output) and `go run . status SPEC-018`.
- [x] Phase 4 — English section headings only. Where:
  `internal/project/project.go` (`Criteria`, `Contract`, `OpenQuestions`),
  `internal/cli/work.go` (`pendingConventions`), `docs/customizing.md`;
  tests in `internal/project/project_test.go`, `internal/cli/cli_test.go`.
  Landed: `Criteria`, `Contract` and `OpenQuestions` read only `Acceptance
  criteria`, `Contract` and `Open questions`; `pendingConventions` reads only
  `Proposed conventions`; the `working_language` paragraph in
  `docs/customizing.md` now says the five section headings are fixed English
  and `working_language` governs the prose, with no claim that a translated
  heading is read. `isNone`'s `ninguna` value is untouched (content, not a
  heading; SPEC-019 decision 2). Tests added:
  `TestSpec_ReadsEnglishSectionHeadingsOnly` (a body with only `Criterios de
  aceptación`, `Contrato` and `Preguntas abiertas` yields empty criteria,
  contract and questions), `TestArchive_ReadsOnlyEnglishConventionHeading`
  (`Proposed conventions` blocks archive and names the proposal; `Convenciones
  propuestas` is invisible and archive succeeds) and
  `TestDocs_CustomizingDoesNotAcceptTranslatedHeadings` (names the five fixed
  English headings and fails on an acceptance claim for the Spanish alias,
  tolerating a negated mention). Verified with `go test ./...` (all packages
  ok), `gofmt -l .` (no output) and `go vet ./...` (no output); the three new
  tests pass with `-v`.

## Proposed conventions

None.
