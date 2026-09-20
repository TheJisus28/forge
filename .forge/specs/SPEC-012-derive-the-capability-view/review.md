# Review - SPEC-012 Derive the capability view

Verdict: **pass with notes**

Reviewed on branch `spec/012-derive-the-capability-view` against contract
`d16e1e49dec2`. Product code not touched; no criterion failed.

## Criterion / evidence

| # | Criterion | Result | Evidence |
|---|---|---|---|
| AC1 | `forge capabilities` lists every capability any `done` contract declares, with the current contracts under each | pass | Scratch repo (outside this one) with `done` SPEC-001 (payments), SPEC-002 (workflow), SPEC-003 (payments) and `proposed` SPEC-004..006. `forge capabilities` printed `payments` then `workflow` (alphabetical), with `SPEC-001`, `SPEC-003` under payments and `SPEC-002` under workflow; no `SPEC-004..006`. Unit `TestCapabilities_GroupsOrdersAndMarksSuperseded`, CLI `TestCapabilities_ListsByCapability`. |
| AC2 | `forge capabilities <name>` lists one capability | pass | `forge capabilities payments` printed only the payments group. Unknown name exits 1: `forge: no done spec declares the capability "nope"; forge capabilities lists them`. CLI `TestCapabilities_OneName`. |
| AC3 | A superseded contract is marked and its own file is not modified; `git status` stays clean | pass | SPEC-001 printed `SPEC-001  Old payments (superseded by SPEC-003)`. SHA256 of every file under `.forge` identical before/after (`forge tree unchanged: True`) and `git status --porcelain` empty after running. CLI `TestCapabilities_MarksSuperseded`, `TestCapabilities_IsDeterministicAndWritesNothing`; project test also proves a `dropped` superseder does not bury its target. |
| AC4 | Deterministic, no network, no model | pass | Two runs redirected to raw files had identical SHA256 (`D06E0E3E...C55`). `cmdCapabilities` only calls `project.Load` and formats; `TestBinaryMakesNoNetworkCalls` passes. |
| AC5 | `forge brief` includes the capability summary | pass | After onboarding the scratch project, `forge brief` printed `capabilities:` with `payments  1 current contract` and `workflow  1 current contract`, before the delivered list. CLI `TestBrief_ShowsCapabilitySummary`, view `TestBrief_ShowsCapabilitySummary`. |
| AC6 | `forge status` / plan survey can use the view; `docs/cli.md` documents `forge capabilities` | pass | `docs/cli.md` documents `forge capabilities [name]` and the brief summary (CLI `TestDocs_DescribeCapabilities`); `docs/workflow.md` "Planning from what exists" and `kit/machine/roles/orchestrator.md` "Before planning" start at `forge capabilities`; both `forge-work` skills mention it. The shared grouping is `project.Project.Capabilities()`, the interface `status`/plan callers build on. |

Suite: `go test ./...` all `ok`; `gofmt -l .` empty; `go vet ./...` clean.

## Contract integrity

The `Contract` section hashes to the recorded `contract_hash: d16e1e49dec2`;
`forge brief` reports no contract drift, so the interface frozen for other
specs (`project.Project.Capabilities()` returning
`[]CapabilityGroup{Name, Contracts{ID, Title, SupersededBy}}` in name/contract
order, and CLI `forge capabilities [name]`) is the one that was approved.
Inputs are the `capability` and `supersedes` frontmatter fields only.

## Problems

Blocking the merge: none. Every acceptance criterion passes.

Blocking closing the spec (not the code):

- `tasks.md` still carries a `## Proposed conventions` section with three
  items. `forge archive` scans every `.md` in the spec folder and refuses
  while proposals are undecided (`internal/cli/work.go:pendingConventions`),
  so the team must record them in `.forge/conventions/` or remove the
  section before `forge archive SPEC-012`.

Non-blocking notes:

- Where `project.md` is not onboarded, `view.Brief` returns the onboarding
  message before `capabilitiesSection`, so the summary is absent until
  onboarding. This matches the existing early return and AC5's premise (a
  `done` contract implies a configured project), but is worth knowing.
- AC6's "`forge status` ... can use the view" is met by sharing
  `Capabilities()`, not by `forge status` printing the grouping. The
  contract deliberately leaves the status tree per-state ("The `forge
  status` tree ... keeps listing every spec by state"), so this is as
  specified.
- A done spec with an empty `capability` is invisible (AC1 asks for the
  capabilities contracts declare). The delivered pre-SPEC-010 specs are
  therefore absent until SPEC-013 backfills them; this is the first
  proposal in `tasks.md`.

## Proposed conventions

None.