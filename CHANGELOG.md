# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog 1.1.0](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

[Unreleased]: https://github.com/TheJisus28/forge/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/TheJisus28/forge/releases/tag/v0.1.0
