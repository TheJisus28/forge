# AGENTS

This repository uses **Forge**: work is agreed before it is built, and the
agreement lives in `.forge/`. Keep this file short; it is read every session.

## Before anything

1. `.forge/project.md` — stack and commands. If it is still unanswered,
   run the Forge onboarding and fill it before writing code.
2. `.forge/conventions/` — how code is written here. Never assume a
   convention that is not written down; propose it instead.
3. `forge status` — what is open, who is waiting, what is blocked.
4. The existing-state survey lives once, in the spec's `## Existing state`
   (`.forge/specs/<id>/spec.md`): the architect writes it and planning reads
   it. Reuse what it names instead of rebuilding it.

## The loop

The states and the legal transitions between them live in the binary: run
`forge workflow` to print them.

- Anyone proposes: `forge new "<title>"`.
- Anyone accepts (`forge accept`) and approves a contract (`forge
  approve`). Forge has no authorization model; it records who did it, the
  way a git commit records an author, and leaves scrutiny to the normal
  pull request review a team already does.
- No product code until the spec is `implementing`.
- Never push or merge into the default branch. Open the pull request with
  `forge submit`; a person merges it.
- The CLI owns ids, state, history and each spec's folder. Never edit
  `status` by hand and never renumber a spec yourself.

## Roles

Read the role before acting as one: `forge roles <orchestrator|architect|
implementer|reviewer>`. You are the orchestrator unless you were launched
as another role.

## Language

Process files are English. Specs, decisions and product copy may use the
`working_language` declared in `.forge/project.md`.
