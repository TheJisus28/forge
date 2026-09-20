# Review — SPEC-018

Verdict: pass with notes

## Acceptance criteria

| Criterion | Result | Evidence |
|---|---|---|
| AC1 | pass | `internal/workflow.All()` is the single ordered declaration, and `renderWorkflow` (`internal/cli/machine.go:46-55`) builds the table by iterating `workflow.All()` and calling `Meaning()`/`WaitingFor()`. `kit/machine/WORKFLOW.md:10` holds only `<!-- forge:states -->`, no hand-written table. `go run . workflow` prints the ten states in lifecycle order. In an isolated copy, adding `State("probing")` to `All()` made `go run . workflow` print a `probing` row; appending a `proposed -> accepted` arrow chain to `AGENTS.md` made `TestDocs_DoNotRestateTheStateMachine` fail with "restates the state machine as an arrow chain". `kit.Roles()` derives from `kit/machine/roles/*.md`; `TestWorkflow_StatesAreRenderedFromTheWorkflowPackage` fails if any `WaitingFor` token is not `anyone`, `nobody` or a `kit.Roles()` member, and pins `architect,implementer,orchestrator,reviewer`. Tests: `TestWorkflowCommand_RendersTheStatesFromGo` (two runs byte-identical, rows in `All()` order), `TestRenderWorkflow_MissingMarkerIsAnError`, `TestMeaning_EveryStateHasALine`. |
| AC2 | pass | `orchestratorFrom` (`internal/project/project.go:522-529`) reads `orchestrator` then the read-only legacy `conductor`; `Save` sets `orchestrator` and `doc.Delete("conductor")` (455-456). `go run . status SPEC-001` prints `orchestrator   carra` while `.forge/specs/SPEC-001-.../spec.md` still carries `conductor: carra`; `go run . status SPEC-018` prints `orchestrator   TheJisus28`, and commit `060ec2a` shows the advance rewrote the key. `go run . roles` lists `architect,implementer,orchestrator,reviewer`; `go run . start --help` says `who orchestrates this spec`. `rg -i conductor` over `AGENTS.md`, `kit/AGENTS.md`, `README.md`, `kit/machine/WORKFLOW.md` and `docs/*.md` finds nothing (only the fallback and test scaffolding remain). Tests: `TestSave_MigratesLegacyConductorKey`, `TestStart_RecordsOneOrchestratorName`, `TestDetail_ShowsOrchestrator`, `TestWaitingFor_BlockedNamesTheOrchestrator`, `TestDocs_DoNotSayConductor`. |
| AC3 | pass | `docs/cli.md:57-64` says `forge start` creates the spec folder and records fingerprints, and "does not create `plan.md` or `tasks.md`". End-to-end in a scratch repo: after `init`, `new`, `accept --by jesus`, `start --by ana`, the spec folder holds only `spec.md`, frontmatter is `orchestrator: ana`, and `status` prints `orchestrator   ana`. `cmdStart` (`internal/cli/work.go:165-202`) does only `os.MkdirAll` + `Save`. Tests: `TestDocs_ForgeStartDoesNotCreatePlanningFiles`. |
| AC4 | pass | `project.Criteria`/`Contract`/`OpenQuestions` read only `Acceptance criteria`/`Contract`/`Open questions` (`internal/project/project.go:360,373,380`); `pendingConventions` reads only `Proposed conventions` (`internal/cli/work.go:492`). `docs/customizing.md:51-57` names the five fixed English headings and says a translated heading is not known. Tests: `TestSpec_ReadsEnglishSectionHeadingsOnly`, `TestArchive_ReadsOnlyEnglishConventionHeading` (English blocks, Spanish invisible), `TestDocs_CustomizingDoesNotAcceptTranslatedHeadings`. |
| AC5 | pass | `docs/workflow.md:5-8` points at `forge workflow` for the states and transitions and keeps only the explanation (open questions, criteria, hierarchy, coverage, dependencies, drift). No state arrow chain and no table row whose first cell is a backticked state name in `docs/*.md`, `AGENTS.md`, `kit/AGENTS.md`, `README.md`; the only arrow left is the Conventional Commits `fix -> patch` line at `AGENTS.md:53`. Tests: `TestDocs_DoNotRestateTheStateMachine`, `TestDocs_WorkflowPointsAtForgeWorkflow`. |

Full gate: `go test ./...` all packages `ok`; `gofmt -l .` and `go vet ./...` no output; `go run . validate` passed after this file was written.

## Blocking problems

None.

## Notes

- The approved contract's `Tests` line for `internal/cli/cli_test.go` also promised "the spec folder holds only `spec.md` (`plan.md` and `tasks.md` are absent)" after `new`/`accept`/`start`. `TestStart_RecordsOneOrchestratorName` checks the frontmatter key and `forge status` but not the folder contents, and no other test asserts the absence. The behaviour is correct (verified end to end above), so this is a test-coverage gap against the contract, not an AC failure. Reproduce: `go test ./internal/cli/ -run TestStart_RecordsOneOrchestratorName -v` passes while holding no folder-contents assertion.
- `TestDocs_DoNotRestateTheStateMachine` builds its regex from the current `workflow.All()` names, so it catches a page that lists the current states but would miss a stale list of names that a future rename (SPEC-015) removes. Out of scope here, worth revisiting when states are renamed.
- The `conductor` to `orchestrator` save migration is visible in the spec's own transition commit `060ec2a`, which rewrote `conductor: TheJisus28` to `orchestrator: TheJisus28` without touching `contract_hash` (`ec06b4bfa420`), so there is no contract drift.

## Proposed conventions

None.