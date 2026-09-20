---
id: SPEC-015
title: Collapse the redundant gates and rename the contract state
status: done
capability: workflow
created: 2026-09-20
updated: 2026-09-20
accepted_by: TheJisus28
orchestrator: TheJisus28
approved_by: TheJisus28
contract_hash: 652e4ef43814
---

## Problem

The loop carries steps that decide the same thing twice and a state whose
name lies. The intake pull request and `forge accept` are two decisions for
one question ("is this worth doing?"). The architect finishing a contract
needs a manual `forge advance <id> --to awaiting-approval` before `forge
approve`, two commands for one hand-off. The existing-state survey is
required by the architect role and then repeated when planning fills `##
Existing state`. And `specifying` names the state reached after the
problem, criteria and spec file already exist: what is written there is the
contract, not the spec, so the name misleads every reader and every role.

## Acceptance criteria

- AC1: The state that follows `accepted` is named for the contract
  (`contracting`) in the workflow package, `forge workflow`, `forge roles`,
  `forge status` and the docs.
- AC2: `forge approve` accepts a contract written straight from
  `contracting`, without a separate `awaiting-approval` move, and still
  freezes the contract fingerprint and the approver.
- AC3: Accepting work is one gate: the workflow docs describe entering the
  queue once, not an intake pull request plus a separate command.
- AC4: The existing-state survey is written once, where the architect names
  what is reused, and planning does not ask for it again.
- AC5: `forge validate` and `forge guard` report the new state name, and a
  repository still carrying `specifying` is either migrated or reported with
  the command to fix it.

## Open questions

One per line, like `- OQ1: ...`, for anything that must be answered before
the contract is approved. Write `None.` when there are none.

None.

## Contract

### Decisions

**1. `specifying` is renamed `contracting`; the states and their order keep
one declaration in `internal/workflow` (SPEC-018, decision 1).**
In `internal/workflow/workflow.go` the constant `Specifying` becomes
`Contracting State = "contracting"` and every use follows: `All()` returns
`proposed, accepted, contracting, planning, implementing, blocked, reviewing,
done, dropped` in that order; `Meaning(Contracting)` stays `"The contract is
being written"`; `InFlight` includes `Contracting`; `WaitingFor(Contracting)`
stays `"architect: write the contract"`. `cmdStart` in
`internal/cli/work.go` checks and sets `workflow.Contracting`; the guard's
`case workflow.Specifying` in `internal/cli/guard.go` follows. No Markdown
page gains a state table:
`forge workflow` renders the new name from `All()`, `forge status`/`forge
brief` print `s.Status`, and `forge roles` prints the role files, where
`kit/machine/roles/architect.md` names the state it works in (the spec is
`contracting` while the contract is written) and
`kit/machine/roles/orchestrator.md` launches the architect in `contracting`.
Discards: keeping `specifying` as a second name or alias inside the machine,
and restating the list in a page (SPEC-018 deleted those copies).

**2. `awaiting-approval` is deleted; `forge approve` moves `contracting`
straight to `planning`.**
`workflow.AwaitingApproval` and its four transitions are removed, replaced by
`{Contracting, Planning, "the contract is approved"}` and
`{Contracting, Dropped, "the work will not be done"}`. `cmdApprove` in
`internal/cli/work.go` keeps `workflow.Check(s.Status, workflow.Planning)`
and loses the special-case error that told the user to run `forge advance
--to awaiting-approval`; `approved_by` and `contract_hash` are still set
before the move. `validate.checkArtifacts` drops `AwaitingApproval` from the
case that requires a non-empty `## Contract`, because while `contracting` the
section may still be empty. `view.Brief` in `internal/view/view.go` stops
putting the contract state under `open decisions`: `contracting` is in
flight, and `open decisions` lists only `proposed`. Discards: keeping
`awaiting-approval` as a state (the hand-off it names *is* the approval) and
keeping the `contracting → awaiting-approval → contracting` bounce (a
requested change is an edit before approval; `ContractChanged` and
`forge validate` still catch edits after). A spec currently sitting in
`awaiting-approval` loads as `contracting` (decision 4) and its next
`forge approve` moves it to `planning`; its `## History` is untouched.

**3. `forge accept` is the single gate into the queue; the intake pull
request is removed from the process.**
`cmdNew`'s closing text in `internal/cli/work.go` stops printing the
`intake/<slug>` branch and the intake pull request: it tells the author to
write the problem and criteria and then run `forge accept <id>`, one entry
into the queue. The same edit lands in
`docs/teams.md` (`## Intake`), `docs/cli.md` (`forge new`, `forge accept`),
`kit/machine/roles/orchestrator.md`, the `forge-work` skill
(`kit/claude/skills/forge-work/SKILL.md`; `forge update` refreshes the
planted `.claude/skills/forge-work/SKILL.md`), and the specs README
(`kit/forge/specs/README.md`, plus this repository's
`.forge/specs/README.md`, which is user content edited deliberately, not
refreshed by Forge). Number collisions stay caught by `forge validate` and
fixed by `forge renumber`, which `docs/teams.md` already documents. Scrutiny
does not disappear: it stays in the spec's pull request, which `docs/teams.md`
already says opens early with the contract and no code. Discards: the intake
pull request as a second "is this worth doing?" decision (Forge cannot read a
merge, so a merge could never set the state; the recorded actor is what
matters and `--by` records it), and dropping `forge accept` instead (that
would lose `accepted_by` and take the change past a rename-plus-collapse).
The removed pull request carried two guarantees — a number reserved on `main`
and a person at the gate — so this decision holds only together with decision
7 (`forge accept` confirms the id against `origin/main`) and decision 8 (an
agent session cannot run it).

**4. Retired state names are a read-only alias; `forge migrate` (decision 6)
is the command that converges the tree.**
`internal/workflow/workflow.go` gains `Canonical(s State) State`: it returns
`Contracting` for `"specifying"` and `"awaiting-approval"`, and `s` unchanged
otherwise. `project.FromDoc` sets `Status:
workflow.Canonical(workflow.State(d.Str("status")))`, so a repository that
upgrades keeps loading and every command, the guard included, reports
`contracting` before anything is rewritten; `Save` already writes
`string(s.Status)`, so the first state move rewrites the frontmatter.
`forge validate` compares the raw `status` (`s.Doc().Str("status")`) with the
canonical one and, when they differ, warns `SPEC-004: status "specifying" is
the old name for "contracting"; run forge migrate`. Nothing rewrites
`## History` (decision 0004). Discards: deleting the names outright (every
command would fail on an old repository until a migration ran), and adding
them to `workflow.All()`/`Valid` (a second name in the machine, which AC1
forbids).

**5. The existing-state survey lives once, in `spec.md` `## Existing state`,
written by the architect; planning reads it.**
`kit/machine/templates/spec.md` gains `## Existing state` between
`## Contract` and `## Out of scope`, with the guidance as an HTML comment
(the SPEC-019 pattern) and an empty body. `kit/machine/templates/plan.md`
loses its `## Existing state` section. `internal/project/project.go` gains
`Spec.ExistingState()` returning `s.doc.Section("Existing state")` from
`spec.md`; `internal/validate/validate.go` replaces `planSurveysExisting`
(which read `plan.md`) with a check of `s.ExistingState()` that warns when a
live spec past `accepted` has not surveyed. `kit/machine/roles/architect.md`
tells the architect to record the survey there; `kit/machine/roles/
orchestrator.md`'s `## Before planning` reads it instead of filling a second
one; `kit/AGENTS.md`, `AGENTS.md`, `docs/workflow.md` and
`kit/claude/skills/forge-work/SKILL.md` say the same. `## Existing state` is
not folded into `ContractHash`: it is a reuse list, and correcting it should
not read as contract drift. Discards: leaving the survey in `plan.md` and
asking planning for it again (the duplication AC4 names), and putting it only
in the Contract prose (a later spec could not find it by heading; SPEC-018
already made `Existing state` a fixed English heading).

**6. `forge migrate [--dry-run]` runs the retired-name rewrite, and both
`forge validate` and `forge guard` report `contracting`.**
New `cmdMigrate` in `internal/cli/migrate.go`, dispatched from
`internal/cli/cli.go` and listed in the usage text and `docs/cli.md`, loads
the project through `project.Load`, and for each spec whose raw `status`
differs from its canonical value sets the frontmatter scalar through
`internal/doc` and saves the file, reporting each path; it appends no history
line, because a rename is not a state move. `--dry-run` prints the same list
and writes nothing; a tree with nothing retired prints `nothing to migrate`
and exits 0. `forge validate`'s warning (decision 4) names `forge migrate`.
`forge guard`'s denial for the contract phase collapses to one
`case workflow.Contracting`, whose message says the spec is `contracting` and
that `forge approve <id>` is next; a frontmatter still saying `specifying` or
`awaiting-approval` loads as `contracting` and is denied with the new name.
Discards: a silent rewrite (the user cannot see what changed), and relying on
the next state move alone to report it (a spec can sit in `contracting` for a
long time, and `forge validate` is what CI runs).

**7. The id is only confirmed by `forge accept`, against the ids on
`origin/main`; `forge new`'s number is provisional.**
`forge new` still writes `project.FormatID(p.NextNum())` from the local tree,
but that number is not final: the spec is `proposed` and a spec that has not
been accepted is not a stable target. `cmdAccept` in `internal/cli/work.go`
becomes the id authority. Before it records the acceptance it reads the ids
committed on the shared branch through a new
`project.RemoteSpecIDs(root, ref) []string` in `internal/project/git.go`,
which runs the existing `run` helper as `git ls-tree -d --name-only <ref>
.forge/specs` and parses each folder's `SPEC-NNN` prefix with
`project.NormalizeID`; the refs are tried `origin/main`, then `main`, then
none (a repository with no remote keeps today's local behaviour). `cmdAccept`
computes the next free number over the union of `p.Specs` and that ref.
If the provisional id is free in both, it is confirmed unchanged. If it is
taken on `origin/main`/`main`, `cmdAccept` renames the spec to the next free
number through the same helper `cmdRenumber` uses — factor the folder move
and the `id`/`Num` rewrite out of `cmdRenumber` in `internal/cli/work.go`
into `renumberSpec(p, s, num) error` and call it from both — notes
`renumbered from SPEC-020: taken on main` on the acceptance history line, and
records `accepted_by` on the confirmed id; its output says the new id.
So a branch that created `SPEC-020` while `main` also advanced to `SPEC-020`
is accepted as `SPEC-021`, with the folder, the `id:` field and the history
line all saying `SPEC-021`, and nothing on `main` conflicts. If the spec is
already referenced (`referencesTo(p, s.ID)` is non-empty), `cmdAccept`
refuses with the existing `forge renumber` message so the reference is
resolved first, and `cmdRenumber` keeps refusing once anything points at the
spec. The residual race — two branches accept from the same base and neither
has the other on `origin/main` — is caught on the merge by `forge validate`'s
existing `duplicate id, also in <file>; run forge renumber`; `forge renumber
<id>` is the migration, run on the branch **before** rebasing onto `main`,
when its own spec is the only one with that id in the tree and the target is
unambiguous. The guarantee is therefore: a spec's id is confirmed at `forge
accept` against `origin/main`/`main`, and no duplicate id can reach `main`
unseen — `forge validate` fails the merge and `forge renumber` is the
migration. Discards: keeping the final number at `forge new` (the collision
the condition names); a shared counter file (decision 0004 rejected a shared
file that parallel branches conflict on); a remote reservation ref or a `gh`
lookup (network, plus a shared write the guard would then have to permit);
and fetching inside `forge accept` (the binary does not fetch; `forge status
--fetch` stays the user's call, and `forge validate` is the backstop when
`origin/main` was stale).

**8. An agent session cannot run `forge accept`; the block is the existing
agent guard, not `cmdAccept`.**
The rule is one new arm of `commandDenial` in `internal/cli/guard.go`: after
`splitSegments` and `shellFields`, a segment whose first word is `forge` or
`forge.exe` (skipping leading `VAR=value` assignments, the same walk that
finds `git push`) and whose subcommand is `accept` is denied with
`Forge: only a person accepts a spec into the queue; ask a human to run:
forge accept <id>`. `forge guard --explain --command "forge accept SPEC-020"`
prints `would deny`. The hosts that call the rule are unchanged:
`.claude/settings.json` gets the PreToolUse hook `Write|Edit|Bash` running
`forge guard` from `writeClaudeSettings` in `internal/cli/init.go`, and
`.opencode/plugins/forge-guard.js` (planted from
`kit/opencode/plugins/forge-guard.js`) calls `forge guard --command <cmd>`;
both already forward every shell command, so the new arm is the whole change
and the rule stays in the binary, as that plugin's own comment says. A person
bypasses it by running `forge accept` in their own shell, where no hook runs,
or by setting `guard: off` in `.forge/project.md` (or re-running `forge init
--no-guard`). This does not break the non-negotiable rule ("it never checks
whether they were allowed to"): the denial lives in the agent guard the host runs
before an agent's tool call, is not consulted by `cmdAccept`, records no
identity and can be switched off, so `forge accept` still accepts anyone and
records them; it has the same shape as the existing `gh pr merge` denial,
which exists so that a person merges the pull request. It is best-effort like
the git checks: a guarded host stops the plain `forge accept ...` invocation
(including inside a `;`/`&&` compound), not `bash -c "forge accept ..."` or a
team that turned the guard off. The guarantee is therefore scoped to a
guarded host: an agent on Claude Code or opencode with the guard on cannot
run `forge accept`, and a person runs it outside the hook. Discards:
refusing inside `cmdAccept` (that
is a Forge-internal authorization check and would break the rule); a
`--by <human>` flag (a record, not a gate); and adding test-time guards to
the host assets outside `commandDenial`, which would put a second copy of the
rule where it can drift.

### Interfaces other specs build against

- The state list is stable and ordered: `proposed, accepted, contracting,
  planning, implementing, blocked, reviewing, done, dropped`. This is a
  rename of `specifying` and the deletion of `awaiting-approval`; no other
  state, and no state's meaning, changes.
- `internal/workflow`: `All()`, `Next()`, `WaitingFor()`, `Meaning()`,
  `Check()`, `Valid()`, `InFlight()` and `Terminal()` keep their signatures;
  `Canonical(State) State` is added. The approval edge is
  `Contracting → Planning`; `WaitingFor(Contracting)` is
  `"architect: write the contract"`, and the role token is one of
  `kit.Roles()` (SPEC-018).
- SPEC-016 (fast lane) inserts its skip inside `internal/workflow` and the
  fast-lane flag in `forge new`; it does not add a state name to a Markdown
  page. SPEC-017 (`forge check`) reads `Spec.Criteria`, `plan.md` and
  `review.md`; its `forge approve` checks sit beside the existing empty
  contract and open-questions checks in `cmdApprove`, now reached from
  `contracting`.
- `project.Spec.ExistingState() string` reads `spec.md`'s
  `## Existing state`. `plan.md` no longer owns that section.
- Frontmatter `status` values are the canonical names. Retired names stay
  readable through `workflow.Canonical` and are rewritten by `forge migrate`;
  `## History` is never rewritten.
- `project.RemoteSpecIDs(root, ref) []string` reads the ids under
  `.forge/specs/` on a git ref; `cmdAccept` and (for the race) `cmdRenumber`
  are the only callers. `renumberSpec(p, s, num) error` in `internal/cli`
  owns the folder move and the `id`/`Num` rewrite for both commands.
- The agent guard denies a `forge accept` invocation through
  `commandDenial`; hosts pick it up with no asset change. `cmdAccept` itself
  is unchanged in who it admits.

### Tests

- Decision 1: `internal/workflow/workflow_test.go:TestAll_UsesContracting` —
  `All()` contains `Contracting`, not `Specifying`/`AwaitingApproval`, in the
  order above; `Meaning(Contracting)` is non-empty;
  `WaitingFor(Contracting) == "architect: write the contract"`;
  `InFlight(Contracting)` is true. The existing `TestCheck_LegalAndIllegalMoves`,
  `TestCheck_CannotSkipApproval` and `TestInFlightAndTerminal` are updated:
  `Accepted → Contracting` and `Contracting → Planning` are legal;
  `Contracting → Implementing` is rejected and the error names `planning`.
- Decision 2: `internal/cli/cli_test.go:TestApprove_StraightFromContracting` —
  `new`, `accept`, `start`, write a non-empty `## Contract` with `None.`
  questions, then `forge approve` with no `forge advance`; the spec ends
  `planning`, `approved_by` and `contract_hash` are set, and no `## History`
  line says `awaiting-approval`. The four
  `advance ... --to awaiting-approval` calls in `internal/cli/cli_test.go`
  (lines ~344, ~501, ~640, ~698) are replaced by a written contract plus
  `approve`. `internal/validate/validate_test.go:TestRun_MissingArtifacts`
  moves its empty-contract case to `planning`.
  `internal/view/view_test.go:TestBrief_ContractingIsInFlight` — a
  `contracting` spec appears under `in flight`, not `open decisions`, and
  `Detail` prints `waiting on architect: write the contract`.
- Decision 3: `internal/cli/cli_test.go:TestNew_DescribesOneGate` — `forge
  new`'s output names `forge accept` and contains no `intake`; `forge accept`
  output contains no `intake`; `docs/teams.md`, `docs/cli.md`,
  `kit/forge/specs/README.md` and `kit/claude/skills/forge-work/SKILL.md`
  contain no `intake`.
- Decision 4: `internal/project/project_test.go:TestFromDoc_CanonicalisesRetiredStatus`
  — a spec with `status: specifying` and one with `status: awaiting-approval`
  load with `Status == workflow.Contracting`, and `workflow.Canonical` maps
  both while leaving `done` alone.
  `internal/validate/validate_test.go:TestRun_WarnsLegacyStatusName` — a
  spec whose raw `status` is `specifying` warns with `contracting` and
  `forge migrate`, and does not error.
- Decision 5: `internal/project/project_test.go:TestExistingState_ComesFromSpec`
  — `Spec.ExistingState()` returns `spec.md`'s `## Existing state`, and a
  `plan.md` carrying one does not change it.
  `internal/validate/validate_test.go:TestRun_SurveyWarnsFromSpec` (replacing
  `TestRun_PlanWithoutExistingStateWarns`) — a live spec past `accepted` with
  no `## Existing state` in `spec.md` warns; adding it clears the warning.
  `internal/cli/machine_test.go:TestTemplate_SurveySection` — `forge template
  spec` contains `## Existing state`; `forge template plan` does not.
- Decision 6: `internal/cli/cli_test.go:TestMigrate_RewritesRetiredStatus` —
  a repository with `specifying` and `awaiting-approval` specs: `forge
  migrate` rewrites both frontmatters to `contracting`, leaves every
  `## History` line byte-identical, and a second run prints `nothing to
  migrate`; `forge migrate --dry-run` writes nothing (the before/after
  snapshot the testing convention asks for).
  `internal/cli/cli_test.go:TestGuard_NamesContracting` — `forge guard --file
  src/x` on a `contracting` spec, and on a legacy `specifying` or
  `awaiting-approval` spec, denies with a message containing `contracting`
  and `forge approve`, and none of `specifying`/`awaiting-approval`.
- Decision 7: `internal/project/project_test.go:TestRemoteSpecIDs_ReadsTheRef`
  — a temporary repository whose `main`/`origin/main` carries
  `.forge/specs/SPEC-020-x/spec.md` returns `SPEC-020`; a repository with no
  such ref returns none.
  `internal/cli/cli_test.go:TestAccept_RenumbersWhenTakenOnMain` — with
  `SPEC-001` already on `main` and a branch whose provisional spec is also
  `SPEC-001`, `forge accept` renames the folder and the `id` to `SPEC-002`,
  notes `renumbered from SPEC-001: taken on main` in the history line, and
  ends `accepted` with `accepted_by`; `forge status` shows `SPEC-002`.
  `internal/cli/cli_test.go:TestAccept_RefusesWhenReferenced` — a proposed
  spec named in another spec's `depends_on` is not renumbered and `forge
  accept` returns the `forge renumber` message.
  `internal/cli/cli_test.go:TestRenumber_ResolvesTheRace` — a tree with two
  `SPEC-020` specs makes `forge validate` report
  `duplicate id ... run forge renumber`, and `forge renumber SPEC-020` before
  rebasing moves the branch's spec to the next free number so `forge
  validate` passes.
- Decision 8: `internal/cli/cli_test.go:TestGuard_DeniesForgeAcceptForAgents`
  — `forge guard --command "forge accept SPEC-020"` exits non-zero and prints
  `only a person accepts`; a compound `cd .forge && forge accept SPEC-020` is
  denied; `forge guard --command "forge new \"x\" --capability workflow"` is
  allowed; `forge guard --explain --command "forge accept SPEC-020"` prints
  `would deny`; and with `guard: off` in `.forge/project.md` the same command
  is allowed.
- Cross-cutting: `internal/cli/machine_test.go:TestWorkflowCommand_RendersTheStatesFromGo`
  keeps iterating `workflow.All()` and additionally pins that the output
  contains ``| `contracting` |`` and neither retired name.
  `internal/cli/cli_test.go:TestDocPages_UseContracting` scans `AGENTS.md`,
  `kit/AGENTS.md`, `README.md` and `docs/*.md` for no `specifying` and no
  `awaiting-approval`, extending the SPEC-018 scan.

### Out of scope of this contract

- Adding, removing or renaming any state other than `specifying →
  contracting` and deleting `awaiting-approval`; the fast lane (SPEC-016) and
  `forge check` (SPEC-017).
- Changing what `forge accept` records or when `forge approve` fingerprints
  the contract.
- Rewriting `## History`, delivered `.forge/specs/*` or `.forge/decisions/*`
  records, or the `conductor` key's read-only fallback (SPEC-018, decision 4).
- A general migration framework: `forge migrate` rewrites retired state
  names only.
- A remote id reservation (a pushed ref or a `gh` lookup) and fetching inside
  `forge accept`; decision 7 reads only refs that are already local.
- A Forge-internal authorization check in `cmdAccept`, and any change to the
  host guard assets beyond the new `commandDenial` arm.
- Changing `forge new`'s local numbering: it keeps assigning a provisional
  number.

## Existing state

Recorded by the architect (decision 5); planning reads it instead of copying
it. The full survey is in `plan.md` during this spec and is not repeated
here; the reusable pieces are `internal/workflow` (states, transitions,
`Meaning`, `Canonical` to be added), the existing `commandDenial` walk in
`internal/cli/guard.go`, the `run` helper in `internal/project/git.go`, the
existing `forge renumber` folder/id rewrite and `duplicate id` validation,
and the SPEC-018 state renderer. Templates, roles, guard assets and the
listed docs are the files this change edits.

## Out of scope

A closed list.

## History

Written by `forge`. Do not edit by hand.
- 2026-09-20  accepted  by TheJisus28
- 2026-09-20  specifying  by TheJisus28
- 2026-09-20  awaiting-approval  by orchestrator
- 2026-09-20  planning  by TheJisus28
- 2026-09-20  implementing  by orchestrator
- 2026-09-20  reviewing  by orchestrator
- 2026-09-20  done  by orchestrator: archived
