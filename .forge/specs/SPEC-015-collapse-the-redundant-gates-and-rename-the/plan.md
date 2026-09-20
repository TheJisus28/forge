# Plan — SPEC-015

## Existing state

- `internal/workflow/workflow.go` owns `All()`, the `transitions` map,
  `Next()`, `Check()`, `WaitingFor()`, `Meaning()`, `InFlight()` and
  `Terminal()`. SPEC-018 added `Meaning` and made `forge workflow` render the
  states from here. The names to change are `Specifying State = "specifying"`
  and `AwaitingApproval State = "awaiting-approval"` plus their rows in
  `All()`, `transitions`, `InFlight` and `WaitingFor`.
- `internal/cli/machine.go:cmdWorkflow` renders the
  `<!-- forge:states -->` marker from `All()`/`Meaning()`/`WaitingFor()`; it
  needs no change once the machine changes.
- `internal/project/project.go`: `FromDoc` sets the status from
  `d.Str("status")`; `Save` writes `string(s.Status)`; `Contract`,
  `ContractHash` and `Criteria` already exist. This is where `Canonical` and
  `Spec.ExistingState` hook in.
- `internal/cli/work.go`: `cmdNew` prints the `intake/<slug>` branch and PR;
  `cmdAccept` records `accepted_by` and nothing else; `cmdStart` checks the
  status is `specifying`; `cmdApprove` special-cases the move to
  `awaiting-approval`; `cmdRenumber` owns the folder move and id rewrite that
  decision 7 factors into `renumberSpec`.
- `internal/cli/guard.go`: `commandDenial` + `splitSegments` + `shellFields`
  already decode a compound shell command to deny `git push` to the default
  branch and `gh pr merge`; a `case workflow.Specifying` names the contract
  phase. Decision 8 adds one arm here, not a new mechanism.
- `internal/validate/validate.go`: `checkArtifacts` requires a non-empty
  `## Contract` while `AwaitingApproval`; `planSurveysExisting` reads
  `plan.md`; a `duplicate id, also in <file>; run forge renumber` check
  already exists and is the race backstop for decision 7.
- `internal/view/view.go:Brief` lists `awaiting-approval` under
  `open decisions`; `Detail` prints `waiting on` from `WaitingFor`.
- Templates and roles: `kit/machine/templates/spec.md` and `plan.md`,
  `kit/machine/roles/architect.md`, `orchestrator.md`, `kit/AGENTS.md`,
  `kit/forge/specs/README.md`, `kit/claude/skills/forge-work/SKILL.md` and
  `kit/opencode/plugins/forge-guard.js`. `forge update` refreshes the planted
  `.claude/` and `.opencode/` copies; `.forge/specs/README.md` is user
  content edited deliberately.
- `internal/project/git.go` already has the `run` helper decision 7 reuses for
  `git ls-tree`; no network and no fetch.
- Docs that restate the states or intake: `docs/workflow.md`,
  `docs/teams.md` (`## Intake`), `docs/cli.md`, `docs/customizing.md`,
  `AGENTS.md`, `kit/AGENTS.md`, `README.md`.
- Conventions that apply: `testing.md` (a seam only where a test must
  substitute; before/after snapshot proves "writes nothing"),
  `cli-output.md` (warnings are `warning: ` lines, errors are returned),
  `frontmatter.md` (empty values delete the key). Decision 0004: delivered
  `.forge/` records and `## History` are never rewritten.
- Genuinely new: `workflow.Canonical`, the retired-name read alias and
  `forge migrate`, `project.RemoteSpecIDs` + shared `renumberSpec` for id
  confirmation, the `forge accept` agent-denial arm, and the survey moving to
  `spec.md` `## Existing state`.

## Phase 1 — The state machine and the approval path

Both of the plan's first concerns land here: the rename and collapse do not
compile without `cmdApprove` moving straight to `planning`, so one implementer
run covers decisions 1, 2 and 4.

- Scope: in `internal/workflow/workflow.go` rename `Specifying` to
  `Contracting`, delete `AwaitingApproval`, update `All()`, `transitions`,
  `InFlight`, `WaitingFor` and `Meaning`, and add `Canonical(State) State`
  mapping `specifying`/`awaiting-approval` to `Contracting`. Update
  `project.FromDoc` to load through `Canonical`, the `case workflow.Specifying`
  in `internal/cli/guard.go`, `cmdStart`, and the tests that name the old
  states. Add `workflow_test.go` and `project_test.go` cases, and the
  `validate` warning naming `contracting` and `forge migrate`.
- Done when: the machine has the nine states in order, retired names load as
  `contracting`, and `forge workflow`/`forge guard` say `contracting`. Moves
  AC1, AC5.
- Verify with: `go test ./...`, `gofmt -l .`, `go vet ./...`,
  `go run . workflow`.

- Also in this phase: `cmdApprove` moves `contracting → planning`, dropping the
  `awaiting-approval` bounce and its error while still freezing `approved_by`
  and `contract_hash`; `validate.checkArtifacts` moves its non-empty-contract
  case to `planning`; `view.Brief` puts `contracting` in flight;
  `kit/machine/roles/architect.md` and `orchestrator.md` name `contracting`.
  Update the four `advance --to awaiting-approval` test call sites to a
  written contract plus `approve`.
- Done when: `new → accept → start → approve` reaches `planning` with no
  intermediate move, from a `contracting` spec. Moves AC2.
- Verify with: `go test ./...`, `gofmt -l .`, `go vet ./...`.

## Phase 2 — `forge accept` is the single gate and the id authority

- Scope: `cmdNew` and the docs/roles/skill (`docs/teams.md` `## Intake`,
  `docs/cli.md`, `kit/machine/roles/orchestrator.md`,
  `kit/claude/skills/forge-work/SKILL.md`, `kit/forge/specs/README.md`,
  `.forge/specs/README.md`) stop describing the intake pull request and name
  one entry through `forge accept`. Add `project.RemoteSpecIDs(root, ref)` in
  `internal/project/git.go` (`git ls-tree`, refs tried `origin/main`, `main`,
  none) and factor `renumberSpec(p, s, num)` out of `cmdRenumber`; `cmdAccept`
  confirms the provisional id against the union and renumbers when taken,
  noting `renumbered from SPEC-NNN: taken on main`. Add the `forge accept`
  denial arm to `commandDenial`.
- Done when: accepting a spec whose id is taken on `main` renames it and
  records the confirmed id, the docs describe one gate, and `forge guard`
  denies `forge accept` for an agent. Moves AC3.
- Verify with: `go test ./...`, `gofmt -l .`, `go vet ./...`,
  `go run . guard --explain --command "forge accept SPEC-020"`.

## Phase 3 — `forge migrate` and the reported names

- Scope: new `internal/cli/migrate.go:cmdMigrate [--dry-run]`, dispatched
  from `internal/cli/cli.go` and listed in the usage text and `docs/cli.md`;
  it rewrites retired frontmatter `status` values through `internal/doc`,
  appends no history line, prints `nothing to migrate` and exits 0 when there
  is nothing to do, and writes nothing under `--dry-run`. Confirm the
  `validate` warning names it.
- Done when: a tree carrying `specifying` or `awaiting-approval` converges to
  `contracting` with `## History` byte-identical, and the no-op path is
  clean. Moves AC5.
- Verify with: `go test ./...`, `gofmt -l .`, `go vet ./...`.

## Phase 4 — The survey lives once, in `spec.md`

- Scope: add `## Existing state` to `kit/machine/templates/spec.md`
  (HTML-comment guidance, empty body) between `## Contract` and
  `## Out of scope`, and remove the section from
  `kit/machine/templates/plan.md`; add `Spec.ExistingState()`; replace
  `planSurveysExisting` with a check of `s.ExistingState()`;
  `kit/machine/roles/architect.md` writes it, `orchestrator.md`,
  `kit/AGENTS.md`, `AGENTS.md` and `docs/workflow.md` read it. Add this
  spec's own survey to `spec.md` so the new check passes.
- Done when: the architect records the survey in `spec.md`, planning does not
  ask again, and `forge template plan` no longer carries the section. Moves
  AC4.
- Verify with: `go test ./...`, `gofmt -l .`, `go vet ./...`,
  `go run . template spec`, `go run . template plan`.

## Phase 5 — Docs hold the new name only

- Scope: the cross-cutting scan `TestDocPages_UseContracting` over
  `AGENTS.md`, `kit/AGENTS.md`, `README.md` and `docs/*.md` for no
  `specifying` and no `awaiting-approval`, extending the SPEC-018 scan; pin
  ``| `contracting` |`` in the `forge workflow` output test; check the final
  tree for `intake` references.
- Done when: no reader-facing page names a retired state and the command
  output test pins the new table row. Moves AC1, AC3.
- Verify with: `go test ./...`, `gofmt -l .`, `go vet ./...`,
  `go run . workflow`.

## Risks

- This spec changes the state machine while it is itself a spec in that
  machine. Phases 1–2 land the rename; the spec is already past
  `contracting`, so its own status stays valid. Run the phases that touch
  `internal/workflow` before anything relies on `forge advance` from a
  retired state.
- The id confirmation reads only local refs; a stale `origin/main` can let
  two branches accept the same number. That is the residual race decision 7
  accepts, caught by `forge validate` on the merge and fixed by
  `forge renumber`. If the maintainer wants it closed, `forge accept` would
  have to fetch, which the contract puts out of scope.
- The agent denial for `forge accept` is best-effort host policy; a
  guard-off team or `bash -c` bypasses it. The contract states this plainly.
- Decision 5 moves a section between templates. `forge update` must plant the
  new `spec.md`/`plan.md` without clobbering a user's existing specs; the
  template change is additive to `spec.md` and subtractive to `plan.md`, so a
  live spec that never re-copies the template keeps its old plan section
  until it is saved.
- Scope creep: SPEC-016 (fast lane) and SPEC-017 (`forge check`) build on
  this state list but are not implemented here. Do not add a state, a flag or
  a check command.
