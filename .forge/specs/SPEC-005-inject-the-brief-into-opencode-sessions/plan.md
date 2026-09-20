# Plan — SPEC-005

## Existing state

- `kit/opencode/plugins/forge-guard.js` is the plugin pattern: it receives
  `{ $, directory }`, runs `forge guard` through the shell, and throws to
  deny. The brief plugin follows the same shape.
- The OpenCode plugin API has `experimental.chat.system.transform`
  (`output.system: string[]`) and `experimental.session.compacting`
  (`output.context: string[]`), confirmed against `@opencode-ai/plugin`.
- `forge brief` exists and prints the session state; the Forge repository is
  itself a Forge project, so it is a real place to test.
- Delivered specs: SPEC-001..SPEC-004.

## Phase 1 — The plugin

- Scope: `kit/opencode/plugins/forge-brief.js`. On the system transform it
  replaces a `Forge brief:` block with the current `forge brief`; on
  compaction it pushes the same text into `output.context`. Nothing when
  `forge brief` fails.
- Verify with: a real `opencode run` in this repository.

## Phase 2 — Docs and planted copy

- Scope: `docs/opencode.md` and `CHANGELOG.md`; run `forge update` so
  `.opencode/plugins/forge-brief.js` exists.
- Verify with: `forge validate` and `opencode agent list`.

## Risks

- The transform runs per request; the replace-not-append keeps it from
  growing. The dynamic block sits after the static instructions, so a
  provider cache still hits the stable prefix.
- If a future OpenCode drops the experimental hooks, the plugin simply does
  nothing; the guard plugin is independent.
