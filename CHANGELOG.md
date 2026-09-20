# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog 1.1.0](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.9.0] - 2026-09-20

### Added

- `forge brief` refreshes the remote refs at session start when
  `.forge/project.md` sets `fetch: on`, so a session does not start from a
  stale clone. It runs `git fetch` only — never `pull`, `merge` or `rebase` —
  and only for a GitHub remote with an authenticated `gh`; when the fetch
  cannot run or fails it prints a `warning: ` line, renders from local refs
  and exits 0. The default stays offline, and the docs, the roles and
  `project.md` say so.
- `forge new`, `forge accept` and `forge renumber` read the spec folders
  committed under `.forge/specs/` on every remote-tracking ref, not just
  `main`, so a number a parallel branch already pushed is skipped. The read is
  best-effort and never fetches, and `forge accept` keeps the id of a spec's
  own published branch; in a repository with more than one person, run
  `git fetch` before `forge new`.

## [0.8.0] - 2026-09-20

### Added

- `forge push [id]` checkpoints the work between phases: it commits the
  pending tree as `chore(<id>): checkpoint <state>` and pushes the spec
  branch with its upstream set, refusing the default branch and succeeding
  when there is nothing to push. With `push: on` in `.forge/project.md`,
  `forge advance` runs the same checkpoint at every state boundary; without
  it, advance never touches the network. `forge submit` commits pending work
  before it pushes.

## [0.7.0] - 2026-09-20

### Added

- `forge approve` refuses a contract whose criteria name no command, test or
  observable response, and `forge check [id]` reports every criterion that no
  task or evidence settles, read-only, exiting 1 only for a missing evidence
  line on a `reviewing` or `done` spec. `forge validate` reports the same
  coverage as a warning, or an error at `done`. Coverage is read from
  comment-stripped `tasks.md` and the review's `## Acceptance criteria`,
  matching ids as whole tokens so `AC1` never counts `AC10`, `AC1x` or `AC1-`;
  the templates and the reviewer role teach the rule.

## [0.6.0] - 2026-09-20

### Changed

- The state after `accepted` is `contracting` and `awaiting-approval` is gone:
  `forge approve` moves a spec straight from `contracting` to `planning`,
  still freezing the approver and the contract fingerprint. `forge accept` is
  the single gate into the queue and confirms the spec's id against
  `origin/main`, renumbering it when the number is taken; the agent guard
  denies `forge accept`, so only a person accepts a spec. `forge migrate`
  rewrites retired state names, and the existing-state survey lives once, in
  `spec.md` `## Existing state`, written by the architect.

## [0.5.0] - 2026-09-20

### Changed

- The workflow states, transitions and roles have one source of truth: the
  states live in `internal/workflow` and `forge workflow` renders them into
  `kit/machine/WORKFLOW.md`, so the docs and the pages no longer restate the
  loop and point at the command instead.
- Whoever drives a spec is named `orchestrator` everywhere: the frontmatter
  key, `forge status`, `forge roles`, the help text and the docs. Specs
  written before the rename keep their recorded actor through a read-only
  `conductor` fallback, and `Save` writes only the new key.
- The CLI reads English section headings only (`Acceptance criteria`,
  `Contract`, `Open questions`, `Proposed conventions`, `Existing state`);
  `working_language` governs the prose, not the headings. `docs/cli.md` no
  longer claims `forge start` writes `plan.md` or `tasks.md`.

## [0.4.0] - 2026-09-20

### Changed

- The guard treats repository-root paperwork — Markdown and licence or notice
  files — as process files, so editing the changelog or the readme no longer
  needs a spec. The unused community boilerplate at the repository root was
  removed, and the contributor guide now lives in `AGENTS.md`.

## [0.3.0] - 2026-09-20

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
- The README now leads with what Forge is for, how to install it and how to
  use it, and links every document.
- `forge guard --file <path>`: the hook-free mode other agents call. It exits
  1 and prints the reason when an edit must be denied, reusing the rule the
  Claude Code hook already enforces.
- `forge submit [id] [--base <branch>]`: closes a spec as a pull request.
  It pushes the branch and opens the PR with `gh`, records it on the spec,
  and never merges; without `gh` it prints the `git push` and `gh pr create`
  commands. The actor recorded by `accept`, `start` and `approve` now
  defaults to the authenticated `gh` login, falling back to git user.name.
- `forge workflow`, `forge roles [name]` and `forge template <name>`: the
  workflow, the four roles and the file templates are read from the binary,
  so any agent follows the process without a planted file. See
  [docs/cli.md](docs/cli.md).

### Fixed

- `forge init` and `forge update` no longer overwrite an existing
  `AGENTS.md` or `CLAUDE.md`. They are the project's own instructions,
  kept like every other file outside `.forge/kit/`; only `--force` rewrites
  them. This is what let the Forge repository manage itself without losing
  the rules in its `AGENTS.md`.

### Changed

- The machinery no longer lives in `.forge/`. The workflow, the roles and
  the file templates ship in the binary, so `.forge/` holds only project
  content. File templates are a standard and can no longer be overridden per
  project, and the host adapters (`.claude/agents/`, `.opencode/agents/`)
  carry the role text inlined instead of pointing at a file. `forge update`
  deletes a stale `.forge/kit/` from an older version.
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
- The kit states the session rule: one spec is one session, each phase goes to
  a subagent to keep the context short, and compaction is the fallback inside
  a session. The brief is built again on start and after a compaction, so a
  fresh session loses nothing.

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

[Unreleased]: https://github.com/TheJisus28/forge/compare/v0.8.0...HEAD
[0.8.0]: https://github.com/TheJisus28/forge/compare/v0.7.0...v0.8.0
[0.7.0]: https://github.com/TheJisus28/forge/compare/v0.6.0...v0.7.0
[0.6.0]: https://github.com/TheJisus28/forge/compare/v0.5.0...v0.6.0
[0.5.0]: https://github.com/TheJisus28/forge/compare/v0.4.0...v0.5.0
[0.4.0]: https://github.com/TheJisus28/forge/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/TheJisus28/forge/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/TheJisus28/forge/compare/v0.1.1...v0.2.0
[0.1.1]: https://github.com/TheJisus28/forge/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/TheJisus28/forge/releases/tag/v0.1.0
