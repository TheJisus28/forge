---
id: SPEC-006
title: One session per spec
status: done
created: 2026-09-20
updated: 2026-09-20
accepted_by: TheJisus28
conductor: TheJisus28
approved_by: TheJisus28
contract_hash: 159a4dc92be5
---

## Problem

Nothing tells an agent how to use sessions. A session that runs across
several specs accumulates noise and drifts, but the kit never says that a
spec is worked in one session, that the phases go to implementer subagents
to keep the context short, and that compaction is the fallback inside one
session, not the way to move between specs.

## Acceptance criteria

Observable outcomes. Someone else must be able to mark each one pass or
fail with evidence.

- AC1: The orchestrator role says a spec is one session, phases go to
  implementer subagents, and compaction is the fallback inside a session.
- AC2: The `forge-work` skill says the same where the work starts.
- AC3: `go test ./...` passes.

## Open questions

One per line, like `- OQ1: ...`, for anything that must be answered before
the contract is approved. Write `None.` when there are none.

None.

## Contract

A short note in the kit, next to the rule it protects. No code.

- `kit/forge/kit/agents/orchestrator.md`: a "Sessions" note.
- `kit/claude/skills/forge-work/SKILL.md`: one line where work starts.

- Decision 1: the rule lives in the kit Markdown, not the CLI. A session is
  the agent's concern; the CLI cannot see one.
- Decision 2: one session per spec, because the state lives in `.forge/` and
  the brief is injected again on start and after a compaction.

## Out of scope

- Enforcing the rule; it is guidance, not a gate.
- Changing the brief or the hooks.

## History

Written by `forge`. Do not edit by hand.
- 2026-09-20  accepted  by TheJisus28
- 2026-09-20  specifying  by TheJisus28
- 2026-09-20  awaiting-approval  by TheJisus28
- 2026-09-20  planning  by TheJisus28
- 2026-09-20  implementing  by TheJisus28: kit note only
- 2026-09-20  reviewing  by TheJisus28: kit note added
- 2026-09-20  done  by orchestrator: archived
