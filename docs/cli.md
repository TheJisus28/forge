# CLI reference

Most commands read and write files under `.forge/`; `forge upgrade` works
outside a project and does not touch the kit. The binary itself never
reaches the network: the only network comes from the user's own tools —
`git`/`gh` for `forge status --fetch`, `forge sync`, `forge push` and
`forge submit`, and the Go toolchain for `forge upgrade`.

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

### `forge new "<title>" --capability <name> [--parent SPEC-002] [--covers AC1,AC3]`

Creates a spec in `proposed` with the next free number, as
`.forge/specs/<id-slug>/spec.md`. The number is provisional: `forge accept`
confirms it against the ids already committed on `main`. `--capability` is
required and must be a lowercase slug naming the part of the system the spec
touches, such as `guard`; an undeclared name prints a warning and the spec
is still created. `--covers` requires `--parent`, and fails if the parent
does not declare those criteria.

### `forge accept <id> [--by <you>] [--note ...]`

Into the queue, and the single gate. Confirms the provisional number against
the ids committed on `origin/main` (or `main`), renumbers the spec when the
number is taken, and records who accepted it. `--by` defaults to `git config
user.name`. Anyone can run this; Forge has no list to check the handle
against.

### `forge start <id> [--by <you>] [--force]`

Begins the work. Refuses if the spec is not `accepted`, has children, or has
open dependencies, and prints what is ready instead. `--force` records the
exception in the spec. Creates the spec folder, records the orchestrator and
the contract fingerprints of any `@contract` dependencies, and prints the
branch to create. It does not create `plan.md` or `tasks.md`; those are
written during planning, after `forge approve`.

### `forge approve <id> [--by <you>] [--note ...]`

The contract is right; code can start. Requires a non-empty `## Contract`
and refuses while `## Open questions` lists a question, so questions are
settled before approval. It also refuses while an acceptance criterion is not
verifiable: each one must name the command, the test or the request and
response that settles it, so a vague criterion cannot be approved. Stores the
contract fingerprint and reports which specs it unblocks.

### `forge advance <id> --to <state> [--by ...] [--note ...]`

Any other move. Validates it against the state machine and appends the
history line. `--to done` is refused: that is what `archive` is for. When
`.forge/project.md` sets `push: on`, it also checkpoints the move — commit
and push, the same as `forge push`; without that scalar `forge advance`
never touches `git` or the network. A checkpoint that fails is a `warning:`
line and the move still stands.

### `forge push [id]`

Checkpoints the work so it survives leaving the machine: commits everything
pending with the message `chore(<ID>): checkpoint <state>` — the spec and
the phase the CLI can name — then pushes the current branch and sets its
upstream. Run it after each phase, so another machine can fetch the branch
and resume.

It refuses the default branch (`main` or `master`) before it commits, and
when there is nothing to commit and the branch is already up to date it
prints `nothing to push: <branch> is up to date` and succeeds.

### `forge archive <id>`

Closes a spec that passed review. Marks the spec `done` and keeps its folder
as the durable record, and reports whether the parent can now be closed.
Refuses if the spec folder has no `review.md` or if `tasks.md` or `review.md`
still propose conventions nobody decided.

### `forge renumber <id> [--to N]`

Resolves a duplicate id. Refuses once anything points at the spec.

### `forge migrate [--dry-run]`

Rewrites the `status` of every spec whose frontmatter still carries a retired
state name to `contracting`, so a tree written before the rename converges.
Only the `status` field changes: the body, including `## History`, is left
byte-identical, because a rename is not a state move. `--dry-run` prints the
same list and writes nothing. When the tree is already current it prints
`nothing to migrate` and succeeds.

## Seeing the state

### `forge status [id] [--fetch]`

Without an id: the tree of specs with coverage, blockers and who is waiting.
With an id: the full detail of one spec, including task progress as
`tasks done/total` read from its `tasks.md`, and its `capability` (or
`(none)`). `--fetch` runs `git fetch` first.

### `forge capabilities [name]`

The current shape of the system, derived from the contracts on disk. Every
`done` spec is a contract, grouped by the `capability` it declares;
capabilities print in name order and the contracts inside one in number
order. A contract that a non-dropped spec supersedes stays visible with a
`(superseded by SPEC-MMM)` suffix, so the history is not lost while the
current shape stays legible, and a superseded contract is never written to.
Without a name it prints every capability; with a name it prints only that
one, and fails when no done spec declares it.

It reads `.forge/specs/` and writes nothing: no `git`, no network, no model
and no file, so the same tree prints byte-identical output and `git status`
is unchanged afterwards.

### `forge brief [--json]`

The short state an agent reads at the start of a session. It shows the
current spec's task progress as `tasks done/total`, read from its
`tasks.md`, and names that spec's `capability`. It also summarises the
current shape with one line per capability and the count of current
contracts, using the same view as `forge capabilities`. `--json` emits the
Claude Code `SessionStart` payload. In a repository without Forge it prints
nothing and succeeds, so the hook is harmless everywhere.

It lists the most recent five `done` specs, so a session knows what already
exists without the context growing with every closed spec; `forge status`
shows the whole list and is the place to drill down.

### `forge check [id]`

`forge check` reports every acceptance criterion no task in `tasks.md`
delivers and no evidence line under `## Acceptance criteria` in `review.md`
settles. Each gap prints as one `no task` or `no evidence` line naming the
criterion and the file it is missing from, relative to the repository root.

Without an id it checks every spec that is `implementing`, `blocked`,
`reviewing` or `done`, in id order; with one it checks that spec whatever its
state. It exits 1 when an evidence gap leaves a `reviewing` or `done` spec
unsettled, because `review.md` is the durable proof; a task gap in flight is
advice and exits 0. It reads and writes nothing: no `git`, no network and no
file.

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
cycles, uncovered promises once a child closes, and contract drift. A spec
with no `capability` is a warning; a present value that is not a lowercase
slug is an error. It also reports a criterion with no task in `tasks.md` or
no evidence line in `review.md`: both are warnings in flight, and a missing
evidence line is an error once the spec is `done`. It does not check who
accepted or approved anything: Forge has no authorization model to enforce.

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

Closes a spec as a pull request. Commits anything still pending, pushes the
current branch with `git` and opens the PR with `gh pr create`, then records
its number and URL on the spec. It never merges: a person reviews and merges.

Without `gh`, or with `--dry-run`, it prints the exact `git push` and
`gh pr create` commands instead of running them and succeeds. `--base`
chooses the target branch (default `main`).

### `forge sync [id]`

Reads the pull request for the current branch through `gh` and records its
number and state in the spec. Optional; everything else works without `gh`.

### `forge version`
