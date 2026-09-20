# Review — SPEC-006

Verdict: pass

## Acceptance criteria

| Criterion | Result | Evidence |
|---|---|---|
| AC1: the orchestrator role states the session rule | pass | `kit/forge/kit/agents/orchestrator.md` has a `## Sessions` section: one spec is one session, phases go to subagents, compaction is the fallback. |
| AC2: the `forge-work` skill says the same | pass | `kit/claude/skills/forge-work/SKILL.md` opens with "One spec is one session: ...". |
| AC3: `go test ./...` passes | pass | All packages `ok`; `forge validate` reports `6 specs, no problems`. |

## Blocking problems

None.

## Notes

- `forge update` planted the note into `.forge/kit/agents/orchestrator.md` and
  `.claude/skills/forge-work/SKILL.md`, so this repository's own sessions get
  it too.
- No code changed; the rule is guidance.

## Proposed conventions

None.
