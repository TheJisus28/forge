---
id: SPEC-010
title: Add capability to every contract
status: reviewing
capability: workflow
created: 2026-09-20
updated: 2026-09-20
accepted_by: TheJisus28
conductor: TheJisus28
approved_by: TheJisus28
contract_hash: 88df1f38fca2
---

## Problem

A contract says what a change does, but not which part of the system it
touches. Nothing groups the contracts about the guard, or the upgrade path,
or the workflow; a reader asking "which contracts concern this capability?"
has to read every `done` contract, and a change that later replaces another
looks unrelated to it. Without a shared axis to group by, the derived view
that decision 0004 chooses has nothing to organise on.

## Acceptance criteria

- AC1: `forge new "<title>"` without `--capability` fails with a message
  that says to pass `--capability <name>`; a spec is no longer creatable
  without one.
- AC2: `forge new "<title>" --capability guard` writes `capability: guard`
  to the new spec's frontmatter.
- AC3: `forge validate` reports a warning naming each spec that has no
  `capability`, so the delivered specs are visible without failing the build.
- AC4: `forge new --capability <name>` prints a warning when no existing
  spec declares `<name>`, and creates the spec anyway, so a typo is visible
  but a genuinely new capability is allowed.
- AC5: `forge validate` reports an error when `capability` is present but
  empty or not a lowercase slug (`[a-z0-9-]+`).
- AC6: `forge status <id>` and `forge brief` show the capability, and
  `forge template spec` and `docs/cli.md` document the field.

## Open questions

One per line, like `- OQ1: ...`, for anything that must be answered before
the contract is approved. Write `None.` when there are none.

None.

## Contract

### Decisions

**1. The frontmatter key is `capability`, a scalar lowercase slug.** A spec
carries one scalar in its frontmatter: `capability: guard`. Its value shape
is `[a-z0-9-]+`. It is mandatory on specs created from now on; on the specs
delivered before this change it warns while absent (SPEC-013 backfills
them). It groups a contract by the part of the system it touches. Discards:
a list of tags, free text, an enum closed in the binary (which needs a code
change per capability), and a body field (nothing could parse it). This is
the public surface frozen by decision 0004, so it is recorded there and not
as a new decision.

**2. `project.Spec.Capability string` is the typed field; the existing
frontmatter helpers carry it.**
- Add `Capability string` to the `Spec` struct in
  `internal/project/project.go`, beside `Title`.
- `FromDoc` sets it with `strings.TrimSpace(d.Str("capability"))`.
- `Save` writes it with `setOrDelete(s.doc, "capability", s.Capability)`, so
  a non-empty value is written (in place when the template already carries
  the key) and an empty one deletes the key, the same as every other empty
  field. `forge new` therefore never leaves `capability: ""` behind.
- Export `func ValidCapability(name string) bool` from `internal/project`,
  matching `^[a-z0-9-]+$`, so `validate` and SPEC-012 share one definition.
- `Doc().Has("capability")` stays the way to tell "absent" from "present but
  empty"; the typed field alone cannot.

Discards: keeping capability only in `doc` (SPEC-012 could not group on it),
a distinct string type, and a private regex per caller.

**3. `forge new` requires `--capability <name>`, and warns only about an
unknown name.**
- Usage becomes
  `usage: forge new "<title>" --capability <name> [--parent SPEC-002] [--covers AC1,AC3]`.
- Missing or blank: return before any file is created, with exactly
  `a capability is required: pass --capability <name>, as in forge new "Pay with a saved card" --capability payments`.
- Present: `strings.TrimSpace` it, then `d.SetStr("capability", cap)` on the
  template before `project.FromDoc`, so the new `Spec` gets it through the
  same read path as a spec on disk and `Save` writes it.
- Unknown name: build the set of non-empty `s.Capability` over every spec in
  `p.Specs` (any status; exact, case-sensitive match, because capabilities
  are lowercase). When the name is not in it, print one warning to `out` and
  create the spec anyway, exit 0:
  `warning: no existing spec declares the capability "<name>"; creating it as a new one`.
- `forge new` does not apply the slug rule; `forge validate` owns shape
  (decision 4), so an invalid value is created and then reported, not
  silently normalised.

Discards: an optional flag (the field would drift, which is the Problem), a
hard failure on an undeclared name (a genuinely new capability must be
allowed), a `--capability` list, and lowercasing the input. The flag name
and its mandatory-ness are the surface frozen in decision 0004.

**4. `forge validate` warns when capability is absent and errors when it is
malformed.** A new `checkCapability` in `internal/validate/validate.go`,
called from `Run` for every spec:
- `!s.Doc().Has("capability")`: a `Warning` on `s.ID` with message
  `has no capability; a new spec sets one with forge new --capability <name>`.
  A warning never sets an exit code, so the historical specs are visible and
  CI does not fail (the repo's own `forge-validate.yml` gates on exit 1).
- present but `!project.ValidCapability(s.Capability)`: an `Error` on `s.ID`
  with message
  `capability "Guard" is not a lowercase slug ([a-z0-9-]+)`, rendering the
  actual value. This covers a bad value and a present-but-empty key such as
  `capability: ""` (or a bare `capability:`), because both give
  `Spec.Capability == ""`.

Discards: failing the build on the historical specs (SPEC-013 would be
blocked and every pre-1.0 clone would break), and folding a present-but-empty
value into the warning (it would hide a broken field).

**5. `forge status <id>` and `forge brief` surface the capability.**
- `view.Detail`: a row immediately after `status`,
  `capability  guard` (same padded label column as `status`/`waiting on`),
  and `capability  (none)` when `Spec.Capability` is empty.
- `view.Brief`: in the current-spec block, `  capability  guard` immediately
  after the `<ID>  <status>  <title>` header and before `waiting on`;
  `(none)` when empty. No current spec, no line.

Discards: annotating every section line (that is the grouped summary
SPEC-012 adds), and omitting it (an agent would start without the one field
the derived view is built on).

**6. The template, the usage text and `docs/cli.md` are the documentation.**
- `kit/machine/templates/spec.md`: add `capability: ""` after
  `status: proposed`, with a short trailing comment naming the shape, so
  `forge template spec` prints the field and `forge new` copies it.
- `internal/cli/cli.go`: the `forge new "<title>"` usage line gains
  `--capability <name>`.
- `docs/cli.md`: `forge new` gains `--capability <name>` in its heading and
  body (required, lowercase slug, warns on an undeclared name and creates);
  one sentence each in `forge status`, `forge brief` and `forge validate`.

Discards: documenting the field only in code or only in this spec.

**7. This spec declares `capability: workflow` on itself.** The implementing
PR adds that one frontmatter line to
`.forge/specs/SPEC-010-add-capability-to-every-contract/spec.md`, otherwise
the spec that introduces the field becomes a permanent missing-capability
warning beside SPEC-001…009. It does not touch the contract, status or
history, so it is not contract drift. Discards: extending SPEC-013's scope
(its AC1 fixes SPEC-001…009) or leaving this spec warning forever.

### Interfaces other specs build against

Frontmatter contract (frozen): key `capability`, scalar, `^[a-z0-9-]+$`,
required by `forge new`, warning while absent.

Go, in `internal/project`:
- `type Spec struct { … ; Capability string ; … }`.
- `func (s *Spec) Save() error` preserves a non-empty `capability` and drops
  an empty one; setting `s.Capability` and calling `Save` adds only that line.
- `func ValidCapability(name string) bool`.

CLI: `forge new --capability <name>`; the `forge validate` warning and error
text above; the `view.Detail` and `view.Brief` line formats above.

Consumers:
- SPEC-011 adds `supersedes` beside `capability` and does not alter it.
- SPEC-012 groups `p.Specs` by `s.Capability` for the `done` view and may
  call `project.ValidCapability`; it re-parses no frontmatter.
- SPEC-013 sets `s.Capability` on SPEC-001…009 and calls `s.Save()`, which
  adds only the `capability` line.

### Tests

End to end, `internal/cli/cli_test.go` (package `cli_test`), against
`newRepo(t)`:
- `TestNew_RequiresCapability`: `run(t, dir, "new", "Thing")` exits non-zero,
  output contains `--capability <name>`, and no new folder appears under
  `.forge/specs`.
- `TestNew_WritesCapabilityAndWarnsOnANewName`: the first
  `mustRun(t, dir, "new", "Guest access", "--capability", "guard")` prints
  `no existing spec declares` and
  `.forge/specs/SPEC-001-guest-access/spec.md` contains `capability: guard`;
  a second `new … --capability guard` prints no warning.
- `TestValidate_Capability`: a spec with no `capability` makes
  `forge validate` exit 0 and print `has no capability` with its id; a spec
  with `capability: Guard` exits 1 and prints `is not a lowercase slug`; the
  same for `capability: ""`.
- `TestStatusAndBriefShowCapability`: `forge status SPEC-001` contains
  `capability` and `guard`. For the brief, `git init` and a checkout of the
  spec branch (the `TestSubmit_DryRunPrintsCommands` shape), then
  `forge brief` contains `capability` and `guard`.
- `TestDocs_DescribeCapability`: `docs/cli.md` contains `--capability`.
- Update every existing `new` call — `TestActor_UsesTheFlag`,
  `TestSubmit_DryRunPrintsCommands`, `TestLifecycle`,
  `TestHierarchyAndDependencies` (three calls),
  `TestApprove_RefusesOpenQuestions`, `TestStatusShowsTaskProgress` — to pass
  a `--capability`, or the suite fails. `TestLifecycle`'s final
  `forge validate` stays clean because SPEC-001 then carries one.

Unit, where the testing convention asks:
- `internal/project/project_test.go`: put `capability: notifications` on a
  fixture and assert `s.Capability`; set it in `TestSaveRoundTrip` and assert
  it survives; `TestValidCapability` accepts `guard` and rejects `Guard`,
  `""`, `with_underscore` and `guard!`.
- `internal/validate/validate_test.go`: `TestRun_WarnsMissingCapability`
  through the existing `warns` helper, and
  `TestRun_RejectsInvalidCapability` through `expectError`, for a bad value
  and an empty value. Add `capability: a` to the `TestRun_CleanProject`
  fixture so "clean" means no finding of either severity.
- `internal/view/view_test.go`: `TestDetail_ShowsCapability` and
  `TestDetail_ShowsNone`. The brief's current-spec line needs a branch, so it
  is covered end to end.

### Verification

- AC1 — `forge new "X"` in an onboarded repo exits 1, prints
  `--capability <name>`, creates nothing. `TestNew_RequiresCapability`.
- AC2 — `forge new "X" --capability guard`; the new `spec.md` contains
  `capability: guard`. `TestNew_WritesCapabilityAndWarnsOnANewName`.
- AC3 — `forge validate` on this repository prints one `has no capability`
  warning per spec of SPEC-001…009 and exits 0 (warnings are hidden by
  `--quiet`). `TestValidate_Capability`; observed here before SPEC-013.
- AC4 — the first `new --capability guard` prints `no existing spec
  declares`, still creates the spec, exits 0.
  `TestNew_WritesCapabilityAndWarnsOnANewName`.
- AC5 — frontmatter `capability: Guard` or `capability: ""` makes
  `forge validate` exit 1 with `is not a lowercase slug`.
  `TestValidate_Capability`, `TestRun_RejectsInvalidCapability`.
- AC6 — `forge status <id>` prints `capability  guard`; `forge brief` prints
  `  capability  guard` on the spec branch; `forge template spec` prints the
  `capability` frontmatter line; `docs/cli.md` documents the flag.
  `TestStatusAndBriefShowCapability`, `TestDocs_DescribeCapability`.

### Explicitly out of scope of this contract

- `forge capabilities` and the derived view (SPEC-012).
- `supersedes`, the computed `superseded` state and its validate rules
  (SPEC-011).
- Backfilling `capability` on SPEC-001…009 (SPEC-013); this contract only
  makes them warn.
- Rejecting or normalising a malformed `--capability` at `forge new`.
- Exempting `dropped` or any other state from the missing-capability warning.
- Any change to the `.forge/` folder layout or to ids, status and history.

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
