# Forge kit

Generic spec-driven kit for coding agents. Language-agnostic. The only
project-specific files live under [`memory/`](memory/).

Aligned with [AGENTS.md](https://agents.md/) (portable agent instructions),
[GitHub Spec Kit](https://github.com/github/spec-kit) (specify → plan →
implement → review), and [OpenSpec](https://openspec.dev/) (proposal →
design → tasks). Forge is smaller: one Go binary, Markdown on disk, no
runtime besides git.

## Layout

| Piece | Where | Generic? |
|---|---|---|
| This guide | README.md | yes |
| Lifecycle | [LIFECYCLE.md](LIFECYCLE.md) | yes |
| Board | [BOARD.md](BOARD.md) | project (starts empty) |
| Agents | [agents/](agents/) | yes |
| Language playbooks | [playbooks/](playbooks/) | yes; selected by stack |
| Templates | [templates/](templates/) | yes |
| Backlog / specs | `backlog/`, `specs/` | project (start empty) |
| Memory | [memory/](memory/) | **project** |

## Session start

1. `memory/stack.md` — runtime of **this** repo.
2. Playbooks listed there (`playbooks/<lang>.md`).
3. `memory/constitution.md` and `memory/decisions.md`.
4. [BOARD.md](BOARD.md). If nothing is active, do not invent work.

Do not create a BL/SPEC unless the user asked for a feature, a change,
or explicitly "open a BL".

## When there is work

1. Feature or change → backlog item.
2. Architect turns it into a spec (contract).
3. Owner approves (`approved` / `lgtm` / `dale`). No implementation before that.
4. Orchestrator plans and delegates: implementer, qa.
5. Every status change is written to `specs/SPEC-XXX/record.md`.
6. Subagents **never** change `status`.

Details: [LIFECYCLE.md](LIFECYCLE.md).

## Plant this kit in another repo

```bash
forge init /path/to/other/repo
```

`--force` rewrites the kit and keeps `memory/` unless `--reset-memory`.
