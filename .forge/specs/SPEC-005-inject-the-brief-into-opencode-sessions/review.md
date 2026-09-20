# Review — SPEC-005

Verdict: pass

## Acceptance criteria

| Criterion | Result | Evidence |
|---|---|---|
| AC1: an opencode plugin appends `forge brief` to the system prompt | pass | `experimental.chat.system.transform` in `kit/opencode/plugins/forge-brief.js` replaces a `Forge brief:` block. A real `opencode run` in this repository reported the branch, `SPEC-005` and `implementing` without running any command. |
| AC2: the brief is fed into the compaction prompt | pass | The plugin pushes the same text into `output.context` on `experimental.session.compacting`, per the plugin API. |
| AC3: nothing is injected without Forge or without `forge` | pass | `brief()` returns an empty string on a non-zero exit and the hooks then do nothing. |
| AC4: `go test ./...` passes and the docs cover it | pass | All packages `ok`; `docs/opencode.md` has a "The brief" section and `CHANGELOG.md` records it. |

## Blocking problems

None.

## Notes

- The block is replaced on each transform, so it refreshes and never grows,
  and it sits after the static instructions so a provider cache still hits
  the stable prefix.
- Compaction was verified against the plugin API, not by forcing a real
  compaction.
- `forge update` planted `.opencode/plugins/forge-brief.js`; `forge validate`
  reports `5 specs, no problems`.

## Proposed conventions

None.
