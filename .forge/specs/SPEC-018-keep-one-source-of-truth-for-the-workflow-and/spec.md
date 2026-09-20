---
id: SPEC-018
title: Keep one source of truth for the workflow and roles
status: reviewing
capability: workflow
created: 2026-09-20
updated: 2026-09-20
accepted_by: TheJisus28
approved_by: TheJisus28
contract_hash: ec06b4bfa420
orchestrator: TheJisus28
---

## Problem

The loop is written down six times — `AGENTS.md`, `kit/AGENTS.md`,
`docs/workflow.md`, `kit/machine/WORKFLOW.md`, the `forge-work` skill and
the roles — and they have already drifted. `conductor` is a frontmatter
field and a role in `docs/teams.md` and in the `blocked` message, but
`forge roles` lists four roles and none is named conductor. `docs/cli.md`
says `forge start` creates `plan.md` and `tasks.md`; the code only creates
the folder. `forge archive` matches Spanish headings (`Convenciones
propuestas`) by literal string in a machine contract.

## Acceptance criteria

- AC1: The states, the transitions and the roles have one normative source
  in the binary (`forge workflow`, `forge roles`); the Markdown pages link
  to it instead of restating it, and a test fails if a page's state list
  diverges.
- AC2: There is a single name for whoever drives a spec, it is one of the
  roles `forge roles` lists, and it is used consistently in frontmatter,
  help text and docs.
- AC3: `docs/cli.md` matches behaviour: either `forge start` creates
  `plan.md` and `tasks.md`, or the page no longer claims it does.
- AC4: The headings the CLI reads are English-only and covered by a test
  that fails if a second-language alias is relied on.
- AC5: `docs/` does not restate the state machine; it links to `forge
  workflow`.

## Open questions

One per line, like `- OQ1: ...`, for anything that must be answered before
the contract is approved. Write `None.` when there are none.

None.

## Contract

### Decisions

**1. `internal/workflow` is the only declaration of the states and their
order; `forge workflow` renders the table from it.**
`internal/workflow/workflow.go` keeps `All()` and the `transitions` slice and
gains `Meaning(State) string`, the one-line meaning that today lives only in
the Markdown. `kit/machine/WORKFLOW.md` loses its hand-written `## States`
table and carries the marker `<!-- forge:states -->` where the table was;
`cmdWorkflow` in `internal/cli/machine.go` replaces that marker with a table
built from `workflow.All()`, `workflow.Meaning()` and
`workflow.WaitingFor()`. A missing marker is an error naming
`kit/machine/WORKFLOW.md`, so the source cannot silently lose the table.
Discards: the second copy in Markdown (it already drifted: `blocked` said
`conductor`) and any second state list in Go.

**2. Every Markdown page stops restating the state list and points at `forge
workflow`.** The loop diagram and the `## States` table are deleted from
`docs/workflow.md`, and the loop line is deleted from `AGENTS.md`,
`kit/AGENTS.md` and `README.md`; each says the states and transitions live in
the binary. `docs/workflow.md` keeps what is explanation — open questions,
criteria, hierarchy, coverage, dependencies, drift — but no ordered state
list. The `kit/` copy is the source; `forge update` refreshes the planted
`.claude/` and `.opencode/` copies. Discards: keeping a copy in a page and
testing it for equality (a copy is a second source even when a test guards
it), and deleting `docs/workflow.md` entirely (the rationale is worth
keeping).

**3. `kit/machine/roles/` is the only list of roles; `workflow.WaitingFor`
may only name a role from it.** `forge roles` already lists the role files,
so no role list is added to Go. `WaitingFor(Blocked)` changes from
`"conductor: clear the blocker"` to `"orchestrator: clear the blocker"`. A
test walks every state and fails when the word before `:` is not `anyone`,
`nobody`, or a member of `kit.Roles()`. Discards: a fifth `conductor` role
(the driver already exists as `orchestrator`) and a Go role-name constant
that could drift from the files.

**4. The driver is `orchestrator` everywhere; the `conductor` frontmatter key
is retired.** `project.Spec.Conductor` becomes `Spec.Orchestrator`;
`project.Save` writes `orchestrator` and deletes a legacy `conductor`;
`FromDoc` reads `orchestrator` and falls back to `conductor`, so specs
written before this change keep their recorded actor. `cmdStart`'s `--by`
help in `internal/cli/work.go` says "who orchestrates this spec";
`internal/view/view.go` prints `orchestrator` in `Brief` and `Detail`;
`docs/teams.md`, `docs/cli.md` and the release step in `AGENTS.md` say
orchestrator. Delivered `.forge/specs/` history and `.forge/decisions/`
records are left as written; decision 0002's rule is unchanged by the role's
name. Discards: keeping `conductor` as a second name (AC2 asks for one) and
renaming the existing `orchestrator` role instead (that would touch every
host adapter and leave old history lines contradicting the new name).
**Expensive, hard to reverse:** it renames a frontmatter key, so any tool
that reads `conductor` must move to `orchestrator`; the fallback keeps old
records readable until their next save, and no migration command is added.

**5. `docs/cli.md` matches `forge start`; the command is not changed.**
`forge start` records the orchestrator and `@contract` fingerprints and does
not create `plan.md` or `tasks.md`; those are written during planning, after
`forge approve`. The `forge start` paragraph is corrected to say so.
Discards: making `forge start` create empty `plan.md`/`tasks.md` (artifacts
before approval, and `forge validate` expects them only at `implementing`).

**6. The CLI reads English section headings only.** `project.Criteria`,
`project.Contract` and `project.OpenQuestions` stop falling back to
`Criterios de aceptación`, `Contrato` and `Preguntas abiertas`;
`pendingConventions` in `internal/cli/work.go` stops reading `Convenciones
propuestas`. `working_language` still governs the prose inside a spec,
decision or convention; only the `## ` headings are fixed English
(`Acceptance criteria`, `Contract`, `Open questions`, `Proposed
conventions`, `Existing state`). `docs/customizing.md` is corrected.
Discards: the Spanish heading aliases and the claim that both are accepted.
The `- ninguna` value `isNone` accepts is content, not a heading, and is
unchanged (SPEC-019, decision 2).

### Interfaces other specs build against

- `internal/workflow`: `All()`, `Next()`, `WaitingFor()` keep their
  signatures; new `Meaning(State) string`. `WaitingFor` returns
  `<role>: <action>` (or `anyone`/`nobody`), and the role token is a name from
  `kit.Roles()`. SPEC-015 (rename `specifying` → `contracting`) and SPEC-016
  (fast lane) change `All()`/`transitions`/`Meaning` in this one package;
  `forge workflow` and the tests follow with no Markdown edit.
- `forge workflow`: prints `kit/machine/WORKFLOW.md` with `<!-- forge:states
  -->` replaced by the generated `## States` table. `kit.Workflow()` stays the
  raw embedded copy; callers that need the rendered text go through
  `cmdWorkflow` or the same renderer in `internal/cli/machine.go`.
- Roles: `kit.Roles()` — the `.md` files in `kit/machine/roles/` — is the
  normative list; `forge roles` and the "who acts next" column of `forge
  workflow` both derive from it.
- Frontmatter: the key is `orchestrator`; `conductor` is read-only legacy and
  is never written. New readers use `orchestrator`.
- Sections: the CLI reads `Acceptance criteria`, `Contract`, `Open
  questions`, `Proposed conventions` and `Existing state`; headings are never
  translated.
- Boundaries: SPEC-014 (guard command scope) and SPEC-017 (`forge check`,
  criterion coverage) do not touch these surfaces; SPEC-015 and SPEC-016
  change contents inside `internal/workflow`, not this mechanism.

### Tests

- `internal/workflow/workflow_test.go`: `Meaning` returns a non-empty line
  for every state in `All()`; `WaitingFor(Blocked)` starts with
  `orchestrator:` and never contains `conductor`.
- `kit/machine_test.go`: `kit.Workflow()` contains `<!-- forge:states -->`
  and no `| \`proposed\`` table row; every `WaitingFor` role token is
  `anyone`, `nobody`, or in `kit.Roles()`, and `kit.Roles()` is
  `architect,implementer,orchestrator,reviewer`.
- `internal/cli/machine_test.go`: `forge workflow` lists `workflow.All()` in
  order with each state's `Meaning` and `WaitingFor`, and two runs are
  byte-identical. `internal/cli/machine_internal_test.go` covers
  `renderWorkflow` returning an error when the marker is absent.
- `internal/cli/cli_test.go`: scanning `AGENTS.md`, `kit/AGENTS.md`,
  `README.md` and `docs/*.md` finds no `state → state` chain and no
  `| \`state\` |` row; `docs/workflow.md` points at `forge workflow`. No
  `conductor` in `docs/*.md`, `AGENTS.md`, `kit/AGENTS.md`,
  `kit/machine/WORKFLOW.md` or `forge roles` output (the `.forge/` records
  are excluded). After `new`/`accept`/`start --by ana`, the spec carries
  `orchestrator:` and no `conductor:`, and `forge status SPEC-001` shows
  `orchestrator   ana`; the spec folder holds only `spec.md` (`plan.md` and
  `tasks.md` are absent), and `docs/cli.md` no longer claims `forge start`
  creates them. A real proposal under `## Convenciones propuestas` does not
  block `forge archive`, while one under `## Proposed conventions` does.
- `internal/project/project_test.go`: a spec with legacy `conductor: ana`
  loads `Orchestrator == "ana"` and the next `Save` writes `orchestrator:`
  and deletes `conductor:`; `Criteria`, `Contract` and `OpenQuestions` are
  empty for a body with only `## Criterios de aceptación`, `## Contrato` and
  `## Preguntas abiertas`.
- `internal/view/view_test.go`: `Detail` prints `orchestrator   ana` for a
  spec with an orchestrator.
- `docs`: a test reads `docs/customizing.md` and fails if it still claims the
  parser accepts `## Criterios de aceptación`.

### Out of scope of this contract

- Renaming any state, or adding or removing a transition (SPEC-015,
  SPEC-016).
- The fast lane for bug fixes and chores (SPEC-016).
- `forge check` and criterion-to-task/evidence coverage (SPEC-017).
- The guard's command scoping (SPEC-014).
- Rewriting delivered `.forge/specs/*` history or `.forge/decisions/*`
  records to say `orchestrator`.
- A migration command for the `conductor` key; the read fallback plus the
  next-save rewrite is the whole mechanism.
- Translating section headings through `working_language`.

## Out of scope

- The contents of the workflow: no state, transition, role or gate is added
  or removed here; this spec only unifies where they are declared and how
  they are rendered.
- The fast lane (SPEC-016), the criterion checks (SPEC-017) and the guard
  scoping (SPEC-014).
- `.forge/` records — delivered specs, decisions and conventions — are
  history and are not rewritten.

## History

Written by `forge`. Do not edit by hand.
- 2026-09-20  accepted  by TheJisus28
- 2026-09-20  specifying  by TheJisus28
- 2026-09-20  awaiting-approval  by orchestrator
- 2026-09-20  planning  by TheJisus28
- 2026-09-20  implementing  by orchestrator
- 2026-09-20  reviewing  by orchestrator
