---
id: SPEC-005
title: Inject the brief into OpenCode sessions
status: done
created: 2026-09-20
updated: 2026-09-20
accepted_by: TheJisus28
conductor: TheJisus28
approved_by: TheJisus28
contract_hash: 9e5be0c0ed6a
---

## Problem

Claude Code runs `forge brief` from a `SessionStart` hook, so a session
starts knowing the Forge state and is refreshed after every compaction.
OpenCode has no such hook, so there the state only reaches the agent when it
runs `forge status` itself; the brief is never injected.

## Acceptance criteria

Observable outcomes. Someone else must be able to mark each one pass or
fail with evidence.

- AC1: An OpenCode plugin appends `forge brief` to the system prompt, so
  every request carries the current Forge state, including after a
  compaction.
- AC2: The plugin also feeds the brief into the compaction prompt, so the
  summary keeps that state.
- AC3: In a repository without Forge, or with `forge` missing, nothing is
  injected and OpenCode behaves as before.
- AC4: `go test ./...` passes and `docs/opencode.md` documents it.

## Open questions

One per line, like `- OQ1: ...`, for anything that must be answered before
the contract is approved. Write `None.` when there are none.

None.

## Contract

`kit/opencode/plugins/forge-brief.js` is the OpenCode counterpart of the
Claude Code `SessionStart` hook.

- It runs `forge brief` through the plugin shell and, on
  `experimental.chat.system.transform`, replaces a `Forge brief:` block in
  `output.system`, so the system prompt carries the current state and the
  block refreshes instead of piling up.
- On `experimental.session.compacting` it pushes the same text into
  `output.context`, so the compaction summary keeps the state.
- A non-zero `forge brief` (no `.forge`, or no `forge`) injects nothing.
- `docs/opencode.md` and `CHANGELOG.md` record it.

- Decision 1: the system-prompt transform, not a synthetic chat message,
  because it is the API for context, and the block goes at the end so the
  static instructions stay a cacheable prefix.
- Decision 2: replace, not append blindly, so state changes and compactions
  refresh the block.
- Decision 3: no configuration and no fallback text; the plugin is inert
  outside a Forge repository.

## Out of scope

- Changing the content or size of `forge brief`.
- Injecting anything other than the brief.

## History

Written by `forge`. Do not edit by hand.
- 2026-09-20  accepted  by TheJisus28
- 2026-09-20  specifying  by TheJisus28
- 2026-09-20  awaiting-approval  by TheJisus28
- 2026-09-20  planning  by TheJisus28
- 2026-09-20  implementing  by TheJisus28: two phases: plugin, docs
- 2026-09-20  reviewing  by TheJisus28: plugin and docs done
- 2026-09-20  done  by orchestrator: archived
