# CLI reference

Most commands read and write files under `.forge/`; `forge upgrade` works
outside a project and does not touch the kit. The binary itself never
reaches the network: the only network comes from the user's own tools —
`git`/`gh` for `forge status --fetch` and `forge sync`, and the Go
toolchain for `forge upgrade`.

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

Refreshes what Forge owns: `.forge/README.md`, the `.claude/` and
`.opencode/` integrations, and the workflows. Never touches your `AGENTS.md`,
`CLAUDE.md`, `project.md`, `specs/`, `decisions/` or `conventions/`; `--force`
rewrites the files it does not own. It also removes a stale `.forge/kit/`
left by an older Forge. Run it after upgrading the binary.

### `forge upgrade [version]`

Upgrades the forge binary itself, not the kit. It installs a released
forge through the Go toolchain — `go install
github.com/TheJisus28/forge@latest`, or an explicit version such as
`forge upgrade v0.2.0` — and then replaces the running binary safely,
renaming it to a sidecar on Windows and cleaning that sidecar up on the
next invocation. It works outside a `.forge/` project and never touches the
planted kit; refreshing the kit stays `forge update`.

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

The contract is right; code can start. Requires a non-empty `## Contract`
and refuses while `## Open questions` lists a question, so questions are
settled before approval. Stores the contract fingerprint and reports which
specs it unblocks.

### `forge advance <id> --to <state> [--by ...] [--note ...]`

Any other move. Validates it against the state machine and appends the
history line. `--to done` is refused: that is what `archive` is for.

### `forge archive <id>`

Closes a spec that passed review. Marks the spec `done` and keeps its folder
as the durable record, and reports whether the parent can now be closed.
Refuses if the spec folder has no `review.md` or if `tasks.md` or `review.md`
still propose conventions nobody decided.

### `forge renumber <id> [--to N]`

Resolves a duplicate id. Refuses once anything points at the spec.

## Seeing the state

### `forge status [id] [--fetch]`

Without an id: the tree of specs with coverage, blockers and who is waiting.
With an id: the full detail of one spec, including task progress as
`tasks done/total` read from its `tasks.md`. `--fetch` runs `git fetch`
first.

### `forge brief [--json]`

The short state an agent reads at the start of a session. It shows the
current spec's task progress as `tasks done/total`, read from its
`tasks.md`. `--json` emits the Claude Code `SessionStart` payload. In a
repository without Forge it prints nothing and succeeds, so the hook is
harmless everywhere.

It lists the most recent five `done` specs, so a session knows what already
exists without the context growing with every closed spec; `forge status`
shows the whole list and is the place to drill down.

## Reading the process

The workflow, the roles and the templates live in the binary, not in your
repository, so the same text works in any host and nothing under `.forge/`
is a tool file.

### `forge workflow`

Prints the workflow: the states, the hierarchy, the dependencies and the
rules, exactly as the agents read them.

### `forge roles [name]`

Lists the four roles — orchestrator, architect, implementer, reviewer — or
prints one role's instructions by name. A host without a wrapper calls this
instead of reading a planted file.

### `forge template <name>`

Prints one file template: `spec`, `plan`, `tasks`, `review`, `decision` or
`convention`. `forge new` uses the same copy, which a project cannot
override.

## CI and integration

### `forge validate [--quiet]`

Exits 1 when the project is inconsistent: unknown or duplicate ids, illegal
history, missing artifacts for a state, broken references, dependency
cycles, uncovered promises once a child closes, and contract drift. It
does not check who accepted or approved anything: Forge has no
authorization model to enforce.

### `forge guard [--explain] [--file path] [--command <cmd>]`

The `PreToolUse` hook, matched on `Write|Edit|Bash`. Reads the payload on
stdin and denies edits to product code while no spec is `implementing`, and
denies commands that would land on the default branch: `gh pr merge`, and
any `git push` or `git merge` targeting `main` or `master`. The rule is that
a person merges the pull request; the agent never pushes or merges into the
default branch. It explains how to unblock. Silence means no decision, so
the normal permission flow continues. Process files are always editable:
`.forge/`, `.claude/`, `.github/`, `.opencode/`, the root pointers, and
repository-root Markdown or `LICENSE`/`NOTICE` files; Markdown inside a
directory stays product code.

With `--file` or `--command` and no `--explain` it is hook-free: it decides
for that one path or command, prints the reason and exits 1 when it must be
denied, so any agent can call it. This is what the opencode plugin uses.
`--explain` never exits 1: it prints `would deny` or `would allow` for a
human.

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
