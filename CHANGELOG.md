# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog 1.1.0](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Planning surveys what exists: `## Existing state` in `plan.md` records the
  delivered specs, modules and conventions a change builds on and what it
  reuses, so a new spec does not treat the repository as empty. The
  orchestrator, architect and implementer roles and the `forge-work` skill
  require the survey, and `forge validate` warns when a plan never did it.
- `forge brief` lists the specs already `done` so a fresh session sees what
  has been delivered before planning, capped at the five most recent so the
  context does not grow with every closed spec; `forge status` lists the rest.
- opencode integration: `forge init` plants `.opencode/agents/` (the
  architect, implementer and reviewer) and `.opencode/plugins/forge-guard.js`.
  opencode reads `AGENTS.md` and Forge's skills natively, so no config file is
  required. See [docs/opencode.md](docs/opencode.md).
- opencode now receives the brief automatically too:
  `.opencode/plugins/forge-brief.js` puts `forge brief` in the system prompt
  at session start and after every compaction, and feeds it into the
  compaction prompt. In a repository without Forge, or without `forge`, it
  injects nothing.
- `forge guard --file <path>`: the hook-free mode other agents call. It exits
  1 and prints the reason when an edit must be denied, reusing the rule the
  Claude Code hook already enforces.
- `forge submit [id] [--base <branch>]`: closes a spec as a pull request.
  It pushes the branch and opens the PR with `gh`, records it on the spec,
  and never merges; without `gh` it prints the `git push` and `gh pr create`
  commands. The actor recorded by `accept`, `start` and `approve` now
  defaults to the authenticated `gh` login, falling back to git user.name.

### Fixed

- `forge init` and `forge update` no longer overwrite an existing
  `AGENTS.md` or `CLAUDE.md`. They are the project's own instructions,
  kept like every other file outside `.forge/kit/`; only `--force` rewrites
  them. This is what let the Forge repository manage itself without losing
  the rules in its `AGENTS.md`.

### Changed

- Removed `forge board` and `.forge/BOARD.md`; `forge init` no longer touches
  `.gitignore`, and everything the board showed is in `forge status`.
- `forge guard` now guards shell commands too: it denies `gh pr merge` and any
  `git push` or `git merge` that lands on the default branch (`main` or
  `master`). It runs hook-free with `forge guard --command "<cmd>"`, and from
  the Claude `PreToolUse` hook (matcher `Write|Edit|Bash`) and the opencode
  plugin. A person merges the pull request; the agent never pushes or merges
  into the default branch.
- `forge approve` refuses while the spec's `## Open questions` section lists a
  question and `forge validate` warns; `forge brief` and `forge status <id>`
  show task progress as `tasks done/total` read from the spec's `tasks.md`.
- Specs now live in a folder per spec, `.forge/specs/SPEC-NNN-slug/`, with the
  standard `spec.md`, `plan.md`, `tasks.md` and `review.md`. `wip/` and
  `changes.md` are gone, and `forge archive` keeps the folder as the durable
  record instead of deleting it. Not backward compatible.

## [0.2.0] - 2026-09-19

### Removed

- The maintainer list, `gates`, `allow_self_approval` and the `forge gate`
  command. Forge has no authorization model: `forge accept` and `forge
  approve` work for anyone, `--by` defaults to `git config user.name`,
  and the move is recorded in the spec's history without checking who made
  it. Real teams review pull requests already; Forge does not add a second,
  Forge-specific approval on top of that.
- The `forge-gate.yml` workflow and `forge validate --approvers`, which
  existed only to enforce the maintainer list.

### Changed

- `forge validate` no longer checks who accepted or approved a spec, only
  what is objectively inconsistent: missing contracts, uncovered promises,
  dependency cycles, contract drift.
- Denial messages from `forge guard` point at the command to run instead of
  naming a maintainer role.

## [0.1.1] - 2026-09-19

### Fixed

- `forge version` reported `dev` when installed with
  `go install github.com/TheJisus28/forge@version`, because that path
  compiles locally and never runs the `-ldflags` GoReleaser bakes into the
  release binaries. It now falls back to the module version Go itself
  records in the build info.

## [0.1.0]

### Added

- Spec-driven workflow built on a single artifact, the spec, moving through the
  states `proposed`, `accepted`, `specifying`, `awaiting-approval`, `planning`,
  `implementing`, `blocked`, `reviewing`, `done` and `dropped`. Two gates belong
  to maintainers: accepting work into the queue and approving a contract.
- Hierarchy between specs through `parent` and `covers`, with a coverage matrix
  showing what a parent spec still has uncovered.
- Dependencies through `depends_on`, with an optional `@contract` level, plus
  `needs` and `blocked_by_external`. Unmet dependencies block work, and the
  `--force` escape is recorded in the spec.
- Contract fingerprinting, so changes to an approved contract are detected as
  drift instead of passing unnoticed.
- Claude Code integration: subagents and skills under `.claude/`, a SessionStart
  brief, and a PreToolUse guard, wired through `.claude/settings.json`.
- GitHub workflows for validating specs and for recording approval gates when a
  maintainer approves a pull request.
- `forge validate` as the check to run in CI.
- Commands: `init`, `update`, `new`, `accept`, `start`, `approve`, `advance`,
  `archive`, `gate`, `status`, `brief`, `board`, `validate`, `sync`, `renumber`,
  `guard` and `version`.

[Unreleased]: https://github.com/TheJisus28/forge/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/TheJisus28/forge/compare/v0.1.1...v0.2.0
[0.1.1]: https://github.com/TheJisus28/forge/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/TheJisus28/forge/releases/tag/v0.1.0
