---
id: 0003
date: "2026-09-20"
status: accepted
spec: SPEC-009
---

# Repository-root paperwork is a process file

## Context

The guard denies edits to "product code" while no spec is `implementing`,
and everything outside a short allowlist counts as product code. That list
was the process directories (`.forge/`, `.claude/`, `.github/`,
`.opencode/`) and a handful of exact root names (`AGENTS.md`, `CLAUDE.md`,
`README.md`, `.gitignore`, `opencode.json`, `opencode.jsonc`). A
repository's own paperwork at the root — `CHANGELOG.md`, a contributing
guide, a security policy, a code of conduct — fell on the wrong side, so
correcting a changelog entry or deleting boilerplate demanded a spec that
delivered no product change. The exact-name list was the wrong shape: it
would deny the next paperwork file (`GOVERNANCE.md`, `ARCHITECTURE.md`) and
force another spec for every new one.

## Decision

A file at the repository root is a process file, and therefore editable
without a spec, when it is Markdown (`.md` or `.markdown`,
case-insensitive) or a licence/notice file (`LICENSE`/`NOTICE`, or any name
starting `LICENSE.`/`NOTICE.`). Markdown inside a directory stays product
code.

## Alternatives

- **Add the missing names to the exact-name switch.** Rejected. It fixes
  today's files and repeats the problem for every future one.
- **Treat all Markdown as a process file.** Rejected. It would let an agent
  edit a product's `docs/` without a spec; Forge's own `docs/` is exactly
  that surface.
- **Make the allowlist configurable in `.forge/project.md`.** Rejected for
  now. It is a public interface for a problem the root rule already covers,
  and it belongs in a later decision if a real project needs it.
- **Leave `guard: off` as the answer.** Rejected. It turns off command
  denial too, so the escape is far larger than the need.

## Consequences

- Adding or renaming root documentation needs no spec; product code still
  does.
- In a repository whose product *is* root-level Markdown (a book, a
  template collection), the guard lets an agent edit those files with no
  spec. We accept that and document it; the escape stays a spec branch or
  `guard: off`.
- The rule is compiled into the binary, not configurable. A project that
  needs a different boundary raises a new decision that supersedes this one.
