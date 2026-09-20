# AGENTS

Forge is a Go CLI that plants a spec-driven workflow into other
repositories. This file describes **this** repo, not the kit it ships.

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
