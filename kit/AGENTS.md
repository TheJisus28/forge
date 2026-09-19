# AGENTS

This repository uses **Forge**: work is agreed before it is built, and the
agreement lives in `.forge/`. Keep this file short; it is read every session.

## Before anything

1. `.forge/project.md` — stack, commands, maintainers. If it is still
   unanswered, run the Forge onboarding and fill it before writing code.
2. `.forge/conventions/` — how code is written here. Never assume a
   convention that is not written down; propose it instead.
3. `forge status` — what is open, who is waiting, what is blocked.

## The loop

```
proposed → accepted → specifying → awaiting-approval → planning →
implementing → reviewing → done
```

- Anyone proposes: `forge new "<title>"`.
- Only a maintainer accepts (`forge accept`) and approves a contract
  (`forge approve`). Ask; never assume approval.
- No product code until the spec is `implementing`.
- The CLI owns ids, state, history and the board. Never edit `status` by
  hand and never renumber a spec yourself.

## Roles

Read the role file before acting as one. They live in `.forge/kit/agents/`:
`orchestrator.md`, `architect.md`, `implementer.md`, `reviewer.md`.

You are the orchestrator unless you were launched as another role.

## Language

Process files are English. Specs, decisions and product copy may use the
`working_language` declared in `.forge/project.md`.
