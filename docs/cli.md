# CLI reference

Every command reads and writes files under `.forge/`. None of them reach
the network; `forge status --fetch` and `forge sync` run `git` and `gh`,
which are the user's tools with the user's credentials.

## Setting up

### `forge init [dir] [--ci github] [--no-guard] [--force]`

Plants the kit: `.forge/`, `AGENTS.md`, `CLAUDE.md`, the Claude Code and
opencode subagents, skills and hooks, and the workflows when `--ci github`
is given.

Existing files are kept. `--force` rewrites them, except `.forge/project.md`,
which is never overwritten. `AGENTS.md` and `CLAUDE.md` are the project's
own instructions, so they are never overwritten without `--force` either.
`--no-guard` skips the `PreToolUse` hook and the opencode guard plugin, so
nothing denies product code edits without an active spec.

### `forge update [--force]`

Refreshes the kit Forge owns: `.forge/kit/`, `.forge/README.md`, the
`.claude/` and `.opencode/` integrations, and the workflows. Never touches
your `AGENTS.md`, `CLAUDE.md`, `project.md`, `specs/`, `decisions/` or
`conventions/`; `--force` rewrites the files it does not own. Run it after
upgrading the binary.

## Moving work

### `forge new "<title>" [--parent SPEC-002] [--covers AC1,AC3]`

Creates a spec in `proposed` with the next free number, as
`.forge/specs/<id-slug>/spec.md`. `--covers` requires `--parent`, and fails
if the parent does not declare those criteria.

### `forge accept <id> [--by <you>] [--note ...]`

Into the queue. `--by` defaults to `git config user.name`. Anyone can run
this; Forge has no list to check the handle against.

### `forge start <id> [--by <you>] [--force]`

Begins the work. Refuses if the spec is not `accepted`, has children, or has
open dependencies, and prints what is ready instead. `--force` records the
exception in the spec. Records the contract fingerprints of any
`@contract` dependencies, creates the spec folder's `plan.md` and `tasks.md`,
and prints the branch to create.

### `forge approve <id> [--by <you>] [--note ...]`

The contract is right; code can start. Requires a non-empty `## Contract`,
stores its fingerprint, and reports which specs it unblocks.

### `forge advance <id> --to <state> [--by ...] [--note ...]`

Any other move. Validates it against the state machine and appends the
history line. `--to done` is refused: that is what `archive` is for.

### `forge archive <id>`

Closes a spec that passed review. Marks the spec `done` and keeps its folder
as the durable record, and reports whether the parent can now be closed.
Refuses if the spec folder has no `review.md` or
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

It lists the most recent five `done` specs, so a session knows what already
exists without the context growing with every closed spec; `forge status`
shows the whole list and is the place to drill down.

### `forge board [--print]`

Regenerates `.forge/BOARD.md`, which is gitignored on purpose.

## CI and integration

### `forge validate [--quiet]`

Exits 1 when the project is inconsistent: unknown or duplicate ids, illegal
history, missing artifacts for a state, broken references, dependency
cycles, uncovered promises once a child closes, and contract drift. It
does not check who accepted or approved anything: Forge has no
authorization model to enforce.

### `forge guard [--explain] [--file path]`

The `PreToolUse` hook. Reads the payload on stdin and denies edits to
product code while no spec is `implementing`, explaining how to unblock.
Silence means no decision, so the normal permission flow continues.

With `--file` and no `--explain` it is hook-free: it decides for that one
path, prints the reason and exits 1 when the edit must be denied, so any
agent can call it. This is what the opencode plugin uses. `--explain` never
exits 1: it prints `would deny` or `would allow` for a human.

### `forge submit [id] [--base <branch>] [--dry-run]`

Closes a spec as a pull request. Pushes the current branch with `git` and
opens the PR with `gh pr create`, then records its number and URL on the
spec. It never merges: a person reviews and merges.

Without `gh`, or with `--dry-run`, it prints the exact `git push` and
`gh pr create` commands instead of running them and succeeds. `--base`
chooses the target branch (default `main`).

### `forge sync [id]`

Reads the pull request for the current branch through `gh` and records its
number and state in the spec. Optional; everything else works without `gh`.

### `forge version`
