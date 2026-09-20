# Tasks — SPEC-012

- [x] Phase 1 — `Capabilities()` in the project package. Where:
  `internal/project/project.go` (`CapabilityGroup{Name, Contracts}`,
  `Contract{ID, Title, SupersededBy}`, `Contract.Current()`,
  `(*Project).Capabilities()`, placed before `criterionRe`),
  `internal/project/project_test.go`
  (`TestCapabilities_GroupsOrdersAndMarksSuperseded`). Landed: the supersede
  index is built once over every non-`dropped` spec (a dropped superseder
  does not bury its target), then every `done` spec with a non-empty
  `capability` is bucketed; groups are sorted by name and contracts keep the
  number order of `p.Specs`. Verified: `go test ./internal/project/...`
  (ok), `go test ./...` (all ok), `gofmt -l .` (empty), `go vet ./...`
  (clean).
- [x] Phase 2 — `forge capabilities [name]`. Where:
  `internal/cli/report.go` (`cmdCapabilities`, after `cmdBrief`),
  `internal/cli/cli.go` (usage line under "Seeing the state" and the
  `case "capabilities"` registration), `internal/cli/cli_test.go` (helpers
  `doneSpec` and `tree`; `TestCapabilities_ListsByCapability`,
  `TestCapabilities_MarksSuperseded`, `TestCapabilities_OneName`,
  `TestCapabilities_IsDeterministicAndWritesNothing`). Landed: the command
  calls `p.Capabilities()`, filters to one name when given, prints each
  capability then its contracts as `  SPEC-NNN  <title>` with a
  ` (superseded by SPEC-MMM)` suffix; an unknown name is a returned error.
  It only `Load`s and prints, so no `Save`, no file. Verified:
  `go test ./internal/cli/...` (ok) and `go test ./...` (all ok); a scratch
  repository in `%TEMP%` printed the expected groups, named capability and
  `(superseded by SPEC-002)`, then `git status --porcelain` was empty — AC3
  holds. An empty tree prints `No done specs declare a capability yet.` and
  exits 0; an unknown name exits 1.
- [x] Phase 3 — Brief summary and docs. Where: `internal/view/view.go`
  (`capabilitiesSection` and its call in `Brief`, after the current spec and
  before the open-decisions sections), `internal/view/view_test.go`
  (`TestBrief_ShowsCapabilitySummary`), `internal/cli/cli_test.go`
  (`TestBrief_ShowsCapabilitySummary`, `TestDocs_DescribeCapabilities`),
  `docs/cli.md` (new `forge capabilities [name]` section; `forge brief`
  mentions the summary), `docs/workflow.md` ("Planning from what exists"
  starts at `forge capabilities`), `kit/machine/roles/orchestrator.md`
  ("Before planning" starts at `forge capabilities`),
  `.claude/skills/forge-work/SKILL.md` and
  `kit/claude/skills/forge-work/SKILL.md` ("Approved work" starts at `forge
  capabilities`). Landed: the brief prints `capabilities:` then
  `  <name>  N current contract(s)` for the non-superseded contracts, and
  prints nothing when no done spec declares a capability. Verified:
  `go test ./...` (all ok), `gofmt -l .` (empty), `go vet ./...` (clean);
  a scratch `forge capabilities`/`brief` run showed the summary counts.

## Proposed conventions

- **An empty `capability` is not a group.** `Capabilities()` skips a `done`
  spec with no `capability`, because AC1 asks for "every capability that any
  `done` contract declares" and `""` is not one. The delivered pre-010 specs
  therefore stay invisible until SPEC-013 backfills them, and the command
  says `No done specs declare a capability yet.` rather than inventing a
  `(none)` bucket that would sort before real names. If the team prefers the
  view to surface backfill debt, the alternative is a synthetic group, but
  that is a product decision this contract did not make.
- **An unknown `forge capabilities <name>` is an error** (exit 1,
  `no done spec declares the capability %q; forge capabilities lists them`),
  matching `forge status <id>` for a missing spec. Printing nothing and
  succeeding was the other option; an error makes a typo visible.
- **The brief count is "current" contracts**, `Contract.Current()` being
  "no non-dropped spec supersedes it", and the line lands right after the
  current spec. The contract fixed the count and the position in prose; this
  records the exact rendering (`N current contract(s)`).

## Notes for the reviewer

- The task brief mentioned `forge capabilities [name] [--quiet?]`; the
  contract's decision 1 and its interface line define only `[name]`, so no
  `--quiet` flag was added. The CLI surface is part of the freeze in decision
  0004 and a flag that was not decided should not be minted here.
- `forge status` was left alone deliberately: the contract puts the status
  tree out of scope, and the brief derives from `Capabilities()` rather than
  a second grouping pass.
- The `superseded by` list is a slice because validate can allow several
  terminal specs to point at the same delivered target; it is rendered
  comma-joined, and only one is expected in normal use.
- `AGENTS.md` and `kit/AGENTS.md` also describe the pre-planning survey but
  are not named by contract decision 4, so they still say `forge status`; a
  later spec can fold `forge capabilities` into them if the team wants one
  survey sentence everywhere.
