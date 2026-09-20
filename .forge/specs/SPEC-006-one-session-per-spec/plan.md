# Plan — SPEC-006

## Existing state

- `kit/forge/kit/agents/orchestrator.md` already says "One phase per
  implementer run. Short contexts beat long ones." but never mentions
  sessions.
- `kit/claude/skills/forge-work/SKILL.md` drives the loop and does not speak
  about sessions either.
- `docs/workflow.md` and `docs/teams.md` describe the flow, not the session.
- Delivered specs: SPEC-001..SPEC-005. The brief is injected at session start
  and after compaction by the Claude hook and the opencode plugin.

## Phase 1 — The note

- Scope: a "Sessions" note in `orchestrator.md` and one line in
  `forge-work`; `forge update` to plant them.
- Done when: AC1 and AC2 hold.
- Verify with: reading the files and `forge validate`.

## Risks

- Guidance can be ignored; it is intentionally not a gate.
