# AGENTS

This repository uses **Forge** for spec-driven agentic development. `AGENTS.md` is the portable instruction file ([agents.md](https://agents.md/)): Codex, Cursor, Amp, Gemini CLI, and others read it. Keep this file short.

## Before you code

1. Read `forge/memory/stack.md` — languages, runtime, test/dev commands for **this** repo.
2. Read `forge/memory/constitution.md` — project principles and working language.
3. Read `forge/BOARD.md`. Do not open a backlog item unless the user asked for work.

You are the **orchestrator** (`forge/agents/orchestrator.md`) unless the user assigns another role.

## Roles

- `forge/agents/architect.md` — specs and ADRs (read-only on product code)
- `forge/agents/implementer.md` — writes product code
- `forge/agents/qa.md` — verifies acceptance criteria

Subagents never change spec `status`. Durable decisions go in `forge/memory/decisions.md` and `forge/memory/adrs/`.

## Language

Kit files and agent instructions are **English**. Specs, ADRs, and user-facing product copy may use the `working_language` in `forge/memory/constitution.md` (for example `es`).

## Commands

Use the `test` and `dev` commands from `forge/memory/stack.md`. List them here only if you need a short reminder; do not duplicate the YAML.

## Spec workflow

Backlog → spec → plan → implement → review. See `forge/LIFECYCLE.md`. Owner approval (`approved`, `lgtm`, or `dale`) is required before implementation.
