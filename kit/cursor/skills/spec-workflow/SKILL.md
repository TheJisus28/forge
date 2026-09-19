---
name: spec-workflow
description: Create or advance Forge SDD work (backlog, spec, plan, changes, review). Use when the user asks for a feature, a change, a spec, a backlog item, or mentions BL-XXX / SPEC-XXX. Do not use for questions or for initializing Forge itself.
---

# Spec workflow

Do not open a backlog item because it would be "the next step". Only if the user asked for product work or said to open one.

Read `forge/memory/stack.md` before writing any spec: the contract must match **this** runtime.

## Create a backlog item

1. Next `BL-00N` from `forge/BOARD.md`.
2. Copy `forge/templates/backlog.md` → `forge/backlog/BL-00N.md`.
3. Add a BOARD row. Status `open`.

## Promote to a spec

1. Next `SPEC-00N`.
2. `forge/specs/SPEC-00N/record.md` (`specifying`) + `spec.md`.
3. Backlog: `promoted` + `spec_id`. Update BOARD.

## Design / plan / review

The architect writes the contract. Human gate to `specified` (skill `advance-spec`). Implementer and QA do not touch `status`. Durable decisions → `forge/memory/decisions.md` or an ADR.
