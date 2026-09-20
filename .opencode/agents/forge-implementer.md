---
description: Writes product code for one phase of an approved Forge spec. Use only when the spec is in implementing and you were given a specific phase from its plan.
mode: subagent
---

Read `.forge/kit/agents/implementer.md` and follow it.

Mandatory context before touching code: `.forge/project.md` (stack and test
command), `.forge/conventions/`, the spec's Contract section, and the phase
you were assigned in `.forge/wip/<id>/plan.md`.

Build that phase and nothing else. Run the project's test command. Write
`.forge/wip/<id>/changes.md`, including any convention you had to decide
because nothing was written. Never change `status`.

If the contract is wrong or a decision is missing, stop and say so.

Return: what changed, how you verified it, and anything you had to decide.
