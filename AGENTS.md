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
   `.forge/specs/<id>/plan.md`.

## The loop

The states and the legal transitions between them live in the binary: run
`forge workflow` to print them. What matters when working:

- Anyone proposes with `forge new`, accepts with `forge accept` and
  approves a contract with `forge approve`. Forge has no authorization
  model; it records who did it, the way a git commit records an author.
- No product code until the spec is `implementing`.
- The CLI owns ids, state, history and each spec's folder. Never edit
  `status` by hand and never renumber a spec yourself.
- Read the role before acting as one: `forge roles <orchestrator|architect|
  implementer|reviewer>`. You are the orchestrator unless you were launched
  as another role.

## Stack

Go 1.22+, standard library only. No external dependencies, and none should
be added: the whole value proposition is a single portable binary.

## Commands

```bash
go test ./...
gofmt -l .
go run . init /tmp/scratch && go run . status
```

## Releases

Each merge to `main` that closes a spec is packaged as a release, cut by
hand (decision 0002). After the spec's pull request is merged:

1. The orchestrator proposes the next `vX.Y.Z` from Conventional Commits
   (`fix` → patch, `feat` → minor, breaking → major); the maintainer
   approves the exact number.
2. Tag the approved commit and push the tag: `git tag vX.Y.Z` then
   `git push origin vX.Y.Z`. GoReleaser publishes from there.

The tag is what `forge upgrade` resolves, so a merge without a tag leaves
installed binaries behind. Update `CHANGELOG.md` in the same pull request as
the change, not at release time.

## Layout

- `kit/` — the Markdown embedded with `go:embed`; `kit/machine/` is served
  by the binary, never planted
- `internal/doc` — frontmatter parser and writer
- `internal/workflow` — states and the legal transitions between them
- `internal/project` — loading `.forge`, specs, ids, coverage, dependencies
- `internal/view` — brief, status and one spec's detail
- `internal/validate` — the consistency rules CI enforces
- `internal/cli` — command parsing and output
- `main.go` — thin entry point

## Contributing

### Adding a command

1. Implement it in `internal/cli`, keeping the rules in the packages that
   own them rather than in the command.
2. Add it to the usage text in `internal/cli/cli.go` and to `docs/cli.md`.
3. Cover it in `internal/cli/cli_test.go`, which runs commands end to end
   against a temporary repository.

### Supporting another agent

`AGENTS.md` is the single source of truth and the roles live in
`kit/machine/roles/`. Support for a new tool is a thin wrapper with the host
frontmatter and a `{{forge-role:<name>}}` marker that `forge init` fills in
from the role, plus the mapping in `internal/cli/init.go`. Keep the wrapper to
that: duplicating a role is how two copies start disagreeing.

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
- **The CLI owns state.** Ids, `status`, history and each spec's folder are
  written by commands, never by hand and never by an agent editing Markdown.
- `.forge/` belongs to the user: `project.md`, `specs/`, `decisions/` and
  `conventions/` are written once and never overwritten. Only the
  explanatory `.forge/README.md` and the host integrations under `.claude/`
  and `.opencode/` are Forge's to refresh.

## Conventions

- Conventional Commits, English summaries: `type(scope): summary`.
- Errors are sentences that tell the reader what to do next.
- Tests describe behaviour, not implementation.

## Language

Process files are English. Specs, decisions and product copy may use the
`working_language` declared in `.forge/project.md`.
