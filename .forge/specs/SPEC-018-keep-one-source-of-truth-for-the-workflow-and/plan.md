# Plan — SPEC-018

## Existing state

- `internal/workflow/workflow.go` owns `All()`, `transitions`, `Next()`,
  `Check()` and `WaitingFor()`. It has no `Meaning`: the one-line meaning of
  each state lives only in `kit/machine/WORKFLOW.md`, which is the drift the
  spec names. `WaitingFor(Blocked)` returns `"conductor: clear the blocker"`
  (line 143), the one place the role shows up in code.
- `internal/cli/machine.go:cmdWorkflow` prints `kit.Workflow()` verbatim
  (lines 13–24); `cmdRoles` prints `kit.Roles()` (27–43). Neither renders.
- `kit/machine.go` serves `Workflow()`, `Roles()`, `Role()`, `Template()`
  from the embedded `kit/machine/` tree. `kit/machine_test.go` already pins
  `Roles()` to `architect,implementer,orchestrator,reviewer`.
- `kit/machine/WORKFLOW.md` carries the hand-written `## States` table whose
  `blocked` row says `conductor`; it is the second copy.
- `internal/project/project.go`: `Spec.Conductor` (line 64), `FromDoc` reads
  `d.Str("conductor")` (505), `Save` writes it with `setOrDelete` (460).
  `Criteria` falls back to `Criterios de aceptación` (360), `Contract` to
  `Contrato` (376), `OpenQuestions` to `Preguntas abiertas` (385).
- `internal/cli/work.go`: `cmdStart` sets `s.Conductor` (179–196) and the
  `--by` help; `pendingConventions` reads both `Proposed conventions` and
  `Convenciones propuestas` (490).
- `internal/view/view.go` prints `conductor` (271).
- `internal/cli/cli_test.go` runs commands end to end (`newRepo`, `run`,
  `mustRun`, `write`, `read`); `internal/cli/machine_test.go` asserts the
  workflow output contains `## States`. These are the homes for the new
  checks.
- Docs that restate the machine: `docs/workflow.md` (loop lines 5–10,
  States table 16–29), `docs/teams.md` (conductor role paragraph, line 88),
  `docs/cli.md:62` (`forge start` "creates `plan.md` and `tasks.md`"),
  `docs/customizing.md:55-56` (claims both heading languages are accepted);
  loop blocks in `AGENTS.md:22-23`, `kit/AGENTS.md:21-22`, `README.md:74-75`.
  `kit/claude/skills/forge-work/SKILL.md` (and the planted `.claude/` copy)
  already say `orchestrator` and carry no state list, so it needs no edit.
- Conventions that apply: `.forge/conventions/testing.md` (add a seam only
  where a test must substitute; these tests need none), `cli-output.md`
  (notices are `warning: ` lines; errors are returned), `frontmatter.md`
  (list fields normalise through `normalizeIDs`, empty removes the key).
  Decision 0002's release step still names the conductor and is updated by
  this spec; the other `.forge/decisions/` and delivered specs are history
  and are not rewritten.
- Genuinely new: `workflow.Meaning`, the `<!-- forge:states -->` renderer,
  the `orchestrator` frontmatter key with its read-only `conductor` fallback,
  and the tests that hold the Markdown to the binary.

## Phase 1 — Workflow states: one declaration, rendered

- Scope: add `Meaning(State) string` to `internal/workflow/workflow.go` and
  change `WaitingFor(Blocked)` to `"orchestrator: clear the blocker"`. In
  `kit/machine/WORKFLOW.md` replace the table body with a `## States` heading
  and the `<!-- forge:states -->` marker. In `internal/cli/machine.go` make
  `cmdWorkflow` render the marker into a table from `workflow.All()`,
  `Meaning()` and `WaitingFor()` (a small `renderWorkflow` the internal test
  can call); a missing marker is an error naming `kit/machine/WORKFLOW.md`.
- Done when: `forge workflow` prints the table generated from Go and the
  Markdown no longer holds a second table. Moves AC1, and the `blocked` row
  of AC2.
- Verify with: `go test ./...`, `gofmt -l .`, `go vet ./...`, and
  `go run . workflow` by eye.

## Phase 2 — Pages stop restating; `docs/cli` matches behaviour

- Scope: delete the loop diagram and `## States` table from
  `docs/workflow.md` and have it point at `forge workflow`; delete the loop
  block from `AGENTS.md`, `kit/AGENTS.md` and `README.md`. Correct the
  `forge start` paragraph in `docs/cli.md` (no `plan.md`/`tasks.md` at
  start; planning writes them after approval). Add the `cli_test.go`
  scanning checks: no `state → state` chain and no `| \`state\` |` row in
  `AGENTS.md`, `kit/AGENTS.md`, `README.md` or `docs/*.md`; `docs/workflow.md`
  names `forge workflow`; `docs/cli.md` drops the old claim.
- Done when: no page carries an ordered state list, the docs link to the
  binary, and `docs/cli.md` describes `forge start` as it behaves. Moves AC1,
  AC3, AC5.
- Verify with: `go test ./...`, `gofmt -l .`, `go vet ./...`.

## Phase 3 — One name for the driver: `orchestrator`

- Scope: rename `project.Spec.Conductor` to `Spec.Orchestrator`; `FromDoc`
  reads `orchestrator` and falls back to `conductor`; `Save` writes
  `orchestrator` and deletes the legacy `conductor` key. Update
  `cmdStart` in `internal/cli/work.go` (assignment and `--by` help) and
  `internal/view/view.go` (`Brief`, `Detail`). Update `docs/teams.md`,
  `docs/cli.md` and the release step in `AGENTS.md` to say orchestrator;
  leave `.forge/` records as written. Tests: legacy `conductor` loads and the
  next save rewrites it (`internal/project/project_test.go`), `Detail` prints
  `orchestrator` (`internal/view/view_test.go`), and after
  `start --by ana` the spec has `orchestrator:` and no `conductor:`
  (`internal/cli/cli_test.go`), plus no `conductor` in the scanned pages.
- Done when: frontmatter, help and docs use `orchestrator` only, and old
  specs still read. Moves AC2.
- Verify with: `go test ./...`, `gofmt -l .`, `go vet ./...`, and
  `go run . status SPEC-018`.

## Phase 4 — English section headings only

- Scope: drop the Spanish fallbacks in `project.Criteria`, `project.Contract`
  and `project.OpenQuestions` (`internal/project/project.go`) and the
  `Convenciones propuestas` heading in `internal/cli/work.go`
  (`pendingConventions`). Correct the working-language paragraph in
  `docs/customizing.md`. Tests: a Spanish-only body yields empty criteria,
  contract and questions (`internal/project/project_test.go`); a proposal
  under `Convenciones propuestas` no longer blocks archive while one under
  `Proposed conventions` does, and `docs/customizing.md` no longer claims
  both (`internal/cli/cli_test.go`).
- Done when: the CLI reads English headings only and the docs say so. Moves
  AC4.
- Verify with: `go test ./...`, `gofmt -l .`, `go vet ./...`.

## Risks

- Decision 4 renames a frontmatter key. The fallback must stay read-only and
  `Save` must delete `conductor`, or a legacy spec that is saved again would
  carry both keys and fail the single-name check. If a host tool reads
  `conductor` directly, it must move to `orchestrator`; this is the one call
  a reviewer could reasonably flip.
- `internal/cli/machine_test.go` asserts the output contains `## States`.
  Keeping the heading while rendering its body keeps that test meaningful;
  deleting the heading would make it silently weaker.
- `docs/workflow.md` is hand-written and long. Removing the table must not
  remove the prose that explains open questions, coverage, dependencies or
  drift.
- Scope creep: SPEC-015 and SPEC-016 change states and add a fast lane inside
  `internal/workflow`; this spec must not. Do not rename `specifying` and do
  not add a flag here.
