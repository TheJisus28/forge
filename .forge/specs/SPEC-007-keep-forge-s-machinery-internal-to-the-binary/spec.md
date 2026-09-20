---
id: SPEC-007
title: "Keep Forge's machinery internal to the binary"
status: done
capability: packaging
created: 2026-09-20
updated: 2026-09-20
accepted_by: TheJisus28
conductor: TheJisus28
approved_by: TheJisus28
contract_hash: b80586d8703a
pr_url: "https://github.com/TheJisus28/forge/pull/9"
pr_state: open
pr: 9
---

## Problem

Forge plants its own machinery — `WORKFLOW.md`, the agent roles and the
file templates — inside `.forge/kit/`, mixed with the project's decisions.
Worse, `forge new` reads a per-project template copy before the embedded
one, so any contributor can change the shape of every spec, plan, decision
and convention in a pull request. The boundary between what belongs to the
team and what belongs to Forge is blurred, and a standard is mutable.

## Acceptance criteria

Observable outcomes. Someone else must be able to mark each one pass or
fail with evidence.

- AC1: After `forge init` in an empty directory, nothing is written under
  `.forge/kit/`; `.forge/` holds only `README.md`, `project.md`, `specs/`,
  `decisions/` and `conventions/`.
- AC2: `forge new` builds a spec from the embedded template only. Creating
  or editing `.forge/kit/templates/spec.md` in the target repository has no
  effect on the generated spec.
- AC3: The role instructions are delivered inside the host adapters:
  `.claude/agents/forge-*.md` and `.opencode/agents/forge-*.md` contain the
  text in full and reference no file under `.forge/kit/`.
- AC4: The workflow and the four roles are obtainable from the installed
  binary alone, with no file under `.forge/`, by the mechanism fixed in the
  contract.
- AC5: `forge update` rewrites nothing under `.forge/specs/`,
  `.forge/decisions/`, `.forge/conventions/` or `.forge/project.md`, and its
  behaviour on a repository that still has an old `.forge/kit/` is the one
  fixed in the contract.
- AC6: `go test ./...`, `gofmt -l .` and `go vet ./...` pass, and the
  standard-library-only rule still holds.

## Open questions

One per line, like `- OQ1: ...`, for anything that must be answered before
the contract is approved. Write `None.` when there are none.

None. Settled: the workflow and the roles are served by `forge workflow`
and `forge roles` from the embedded copy; `forge update` deletes an old
`.forge/kit/` outright, since backward compatibility is not required, and
the change is documented.

## Contract

### Numbered decisions

1. **The machinery lives in the binary, in a new embedded root
   `kit/machine/`.** `git mv kit/forge/kit/ kit/machine/` and rename
   `agents/` to `roles/`. The embed line in `kit/kit.go` becomes
   `//go:embed AGENTS.md CLAUDE.md all:machine all:forge all:claude all:opencode all:github`,
   and `mapDest` (internal/cli/init.go:142) gains a case that returns `""`
   for `machine/…`, so the walk never plants it. Discards: keeping the
   machinery planted at `.forge/kit/` (status quo); a second Forge-owned
   root in the target repository (rejected in decision 0001).

2. **Templates are internal and not overridable.** `loadTemplate`
   (internal/cli/work.go:367) drops the on-disk read and resolves
   `machine/templates/<name>` from `kit.FS` only. The override documented
   in `docs/customizing.md` is removed. Discards: the per-project template
   override, per decision 0001.

3. **Roles are canonical at `kit/machine/roles/<name>.md` and inlined into
   the host adapters when planted.** Each role file becomes the full
   instruction: today's role text plus the host-neutral "mandatory context"
   paragraph that now lives duplicated in the adapters.
   `kit/claude/agents/forge-<name>.md` and
   `kit/opencode/agents/forge-<name>.md` keep only their host frontmatter
   and one marker line, `{{forge-role:<name>}}`. `plant` replaces the
   marker with `machine/roles/<name>.md`; a marker with no matching role is
   an error, never a silent pass. Discards: adapters that point at
   `.forge/kit/agents/` (current), a path that stops existing; keeping two
   copies of the role text in the repository.

4. **Three read-only commands serve the machinery from the binary.**
   `forge workflow` prints `machine/WORKFLOW.md`. `forge roles [name]`
   prints the four role names with no argument and one role with an
   argument. `forge template <name>` prints one template, so decisions and
   conventions are created from the binary's copy rather than a planted
   file. All three read `kit.FS` only, never the project, and succeed in a
   directory with no `.forge/`. Discards: a docs-only copy (absent from
   the target repository); folding the workflow into `AGENTS.md` (read
   every session, so it must stay short).

5. **`forge update` deletes a stale `.forge/kit/` outright.** After the
   walk, `plant` removes `<root>/.forge/kit` when it exists and prints one
   line saying so. No backward compatibility is promised (decision 0001).
   Discards: leaving it (dead machinery inside `.forge/`, contradicting the
   decision); migrating it.

6. **`.forge/kit/` stops being Forge-owned.** `kitOwned`
   (internal/cli/init.go:162) drops the
   `strings.HasPrefix(dest, project.Dir+"/kit/")` arm. Forge keeps owning
   `.forge/README.md`, `.claude/`, `.opencode/` and
   `.github/workflows/forge-*`.

### Interfaces other code builds on

- `kit.Workflow() ([]byte, error)`, `kit.Role(name) ([]byte, error)`,
  `kit.Roles() []string` and `kit.Template(name) ([]byte, error)`, all
  backed by `machine/`. `forge new` and `forge start` keep reading through
  `loadTemplate`, now a thin wrapper over `kit.Template`.
- The command switch and the `usage` text in `internal/cli/cli.go` gain
  `workflow`, `roles` and `template`.
- A planted `.claude/agents/forge-<role>.md` and
  `.opencode/agents/forge-<role>.md` is self-contained: nothing under
  `.forge/` is referenced.

### Tests that pin the behaviour

`internal/cli/cli_test.go` stops expecting `.forge/kit/…` and instead
asserts that a planted tree has no `.forge/kit/`, that a planted adapter
contains the role text, and that a `.forge/kit/templates/spec.md` left in
the tree does not change what `forge new` writes. `kit/kit_test.go` keeps
passing with `machine/` embedded. New unit tests cover `forge workflow`,
`forge roles` and `forge template`.

### Documentation that moves with the code

`AGENTS.md` and `kit/AGENTS.md` (roles path becomes `forge roles`),
`kit/forge/README.md` (`kit/` row and the one rule),
`kit/forge/decisions/README.md` and `kit/forge/conventions/README.md`
(`forge template`), the three `kit/claude/skills/forge-*/SKILL.md`,
`docs/cli.md`, `docs/customizing.md`, `docs/opencode.md`,
`docs/workflow.md`, `README.md`, `CONTRIBUTING.md` and `CHANGELOG.md`.

## Out of scope

- Adding first-class support for hosts other than Claude Code and opencode.
- Any change to the workflow states or to the semantics of specs,
  decisions and conventions.
- Any replacement for per-project template override; it is removed, not
  moved elsewhere.

## History

Written by `forge`. Do not edit by hand.
- 2026-09-20  accepted  by TheJisus28
- 2026-09-20  specifying  by TheJisus28
- 2026-09-20  awaiting-approval  by orchestrator
- 2026-09-20  planning  by TheJisus28
- 2026-09-20  implementing  by orchestrator
- 2026-09-20  reviewing  by orchestrator
- 2026-09-20  done  by orchestrator: archived
