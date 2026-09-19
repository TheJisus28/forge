# CLI reference

Every command reads and writes files under `.forge/`. None of them reach
the network; `forge status --fetch` and `forge sync` run `git` and `gh`,
which are the user's tools with the user's credentials.

## Setting up

### `forge init [dir] [--ci github] [--no-guard] [--force]`

Plants the kit: `.forge/`, `AGENTS.md`, `CLAUDE.md`, the Claude Code
subagents, skills and hooks, and the workflows when `--ci github` is given.

Existing files are kept. `--force` rewrites them, except `.forge/project.md`,
which is never overwritten. `--no-guard` skips the `PreToolUse` hook that
denies product code edits without an active spec.

### `forge update [--force]`

Refreshes `.forge/kit/`, the root pointers and the agent integrations.
Never touches `project.md`, `specs/`, `decisions/` or `conventions/`. Run it
after upgrading the binary.

## Moving work

### `forge new "<title>" [--parent SPEC-002] [--covers AC1,AC3]`

Creates a spec in `proposed` with the next free number. `--covers` requires
`--parent`, and fails if the parent does not declare those criteria.

### `forge accept <id> --by <maintainer> [--note ...]`

The first human gate: into the queue. Fails if the handle is not in
`maintainers`.

### `forge start <id> [--by <you>] [--force]`

Begins the work. Refuses if the spec is not `accepted`, has children, or has
open dependencies, and prints what is ready instead. `--force` records the
exception in the spec. Records the contract fingerprints of any
`@contract` dependencies, creates `.forge/wip/<id>/`, and prints the branch
to create.

### `forge approve <id> --by <maintainer> [--note ...]`

The second human gate: the contract is right. Requires a non-empty
`## Contract`, stores its fingerprint, and reports which specs it unblocks.

### `forge advance <id> --to <state> [--by ...] [--note ...]`

Any other move. Validates it against the state machine and appends the
history line. `--to done` is refused: that is what `archive` is for.

### `forge archive <id>`

Closes a spec that passed review. Deletes `.forge/wip/<id>/` and reports
whether the parent can now be closed. Refuses if there is no `review.md` or
if the scaffolding still proposes conventions nobody decided.

### `forge renumber <id> [--to N]`

Resolves a duplicate id. Refuses once anything points at the spec.

## Seeing the state

### `forge status [id] [--fetch]`

Without an id: the tree of specs with coverage, blockers and who is waiting.
With an id: the full detail of one spec. `--fetch` runs `git fetch` first.

### `forge brief [--json]`

The short state an agent reads at the start of a session. `--json` emits the
Claude Code `SessionStart` payload. In a repository without Forge it prints
nothing and succeeds, so the hook is harmless everywhere.

### `forge board [--print]`

Regenerates `.forge/BOARD.md`, which is gitignored on purpose.

## CI and integration

### `forge validate [--approvers "ana,jose"] [--quiet]`

Exits 1 when the project is inconsistent: unknown or duplicate ids, illegal
history, missing artifacts for a state, broken references, dependency
cycles, uncovered promises once a child closes, contract drift, approvals by
non-maintainers, self-approval where it is not allowed, and `--approvers`
mismatches.

### `forge gate --by <maintainer> [--base origin/main] [--dry-run]`

Turns a pull request approval into the state change it means: accepts the
spec or approves its contract, depending on where it is. Run by CI.

### `forge guard [--explain] [--file path]`

The `PreToolUse` hook. Reads the payload on stdin and denies edits to
product code while no spec is `implementing`, explaining how to unblock.
Silence means no decision, so the normal permission flow continues.

### `forge sync [id]`

Reads the pull request for the current branch through `gh` and records its
number and state in the spec. Optional; everything else works without `gh`.

### `forge version`
