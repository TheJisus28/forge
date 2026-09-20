# AGENTS

Forge is a Go CLI that plants a spec-driven workflow into other
repositories. This file describes **this** repo, and this repo is itself
managed with Forge: work is agreed before it is built, and the agreement
lives in `.forge/`.

## Before anything

1. `.forge/project.md` — what this project is, its stack and commands.
2. `.forge/conventions/` — how code is written here. Never assume a
   convention that is not written down; propose it instead.
3. `forge status` — what is open, who is waiting, what is blocked.
4. Before planning, survey what already exists: `forge status`, the
   delivered specs and the code. Reuse it instead of rebuilding it, and
   record what you reuse under `## Existing state` in
   `.forge/wip/<id>/plan.md`.

## The loop

```
proposed → accepted → specifying → awaiting-approval → planning →
implementing → reviewing → done
```

- Anyone proposes with `forge new`, accepts with `forge accept` and
  approves a contract with `forge approve`. Forge has no authorization
  model; it records who did it, the way a git commit records an author.
- No product code until the spec is `implementing`.
- The CLI owns ids, state, history and the board. Never edit `status` by
  hand and never renumber a spec yourself.
- Read the role file before acting as one. They live in
  `.forge/kit/agents/`: `orchestrator.md`, `architect.md`,
  `implementer.md`, `reviewer.md`. You are the orchestrator unless you were
  launched as another role.

## Stack

Go 1.22+, standard library only. No external dependencies, and none should
be added: the whole value proposition is a single portable binary.

## Commands

```bash
go test ./...
gofmt -l .
go run . init /tmp/scratch && go run . status
```

## Layout

- `kit/` — the Markdown planted into other repos, embedded with `go:embed`
- `internal/doc` — frontmatter parser and writer
- `internal/workflow` — states and the legal transitions between them
- `internal/project` — loading `.forge`, specs, ids, coverage, dependencies
- `internal/view` — brief, status and board rendering
- `internal/validate` — the consistency rules CI enforces
- `internal/cli` — command parsing and output
- `main.go` — thin entry point

## Rules that are not negotiable

- **No network in the binary.** `nonet_test.go` fails the build if any
  package imports `net/http` and friends. Shelling out to `git` or `gh` is
  how network happens, and only when the user asks.
- **No technology opinions in `kit/`.** Forge must be useful in a Rust
  repository and a Rails one. Rules about frameworks, naming or style
  belong to the user's `.forge/conventions/`, written by their own team.
- **No authorization model.** Forge records who accepted or approved
  something; it never checks whether they were allowed to. That decision
  belongs to the team and, if they want it enforced, to their own
  branch protection or `CODEOWNERS`, not to Forge.
- **The CLI owns state.** Ids, `status`, history and the board are written
  by commands, never by hand and never by an agent editing Markdown.
- Anything planted outside `.forge/kit/` belongs to the user and is written
  once, never overwritten.

## Conventions

- Conventional Commits, English summaries: `type(scope): summary`.
- Errors are sentences that tell the reader what to do next.
- Tests describe behaviour, not implementation.

## Language

Process files are English. Specs, decisions and product copy may use the
`working_language` declared in `.forge/project.md`.
