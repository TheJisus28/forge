# Project memory

The only part of the kit that **does not** copy as-is to another repo.

| File | What it stores |
|---|---|
| [stack.md](stack.md) | Language, runtime, commands. Agents read this first. |
| [constitution.md](constitution.md) | Principles and working language (Spec Kit analogue). |
| [decisions.md](decisions.md) | Short log of technical decisions. |
| [adrs/](adrs/) | Formal ADRs when the log is not enough. |

When copying Forge: rewrite `stack.md` and `constitution.md`; leave
`decisions.md` and `adrs/` empty.

## When to write a decision

Runtime, contract, layout, or a dependency another chat would have to
re-ask → `decisions.md`. Shape of the system changes → ADR.

This folder is not a backlog. Work items live in `../backlog/`.
