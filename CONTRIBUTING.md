# Contributing

Thanks for looking. Forge is small on purpose, and the best contributions
usually make it smaller.

## Ground rules

- **Standard library only.** A new dependency needs a very good reason.
- **The kit stays technology-agnostic.** No rules about frameworks,
  naming or style in `kit/`: those belong to each project's
  `.forge/conventions/`, decided by its own team.
- **No authorization model.** Forge records who accepted or approved
  something; it never checks whether they were allowed to. Do not add a
  maintainer list, a role, or a permission check back in.
- **The binary never touches the network.** `nonet_test.go` enforces it.
- **The CLI owns ids and state.** If a change would let an agent edit
  `status` by hand, it is the wrong change.
- Conventional Commits, English summaries: `type(scope): summary`.

## Local development

```bash
go test ./...
gofmt -l .

# try it against a scratch repository
go build -o /tmp/forge . && cd /tmp/scratch && /tmp/forge init && /tmp/forge status
```

## Where things live

| Path | What |
|---|---|
| `kit/` | The Markdown planted in other repositories, embedded via `go:embed` |
| `internal/doc` | Frontmatter parsing and writing |
| `internal/workflow` | States and the legal transitions between them |
| `internal/project` | `.forge` loading, specs, coverage, dependencies |
| `internal/view` | Brief, status and board |
| `internal/validate` | The CI rules |
| `internal/cli` | Commands |

Most contributions are Markdown under `kit/`. Anything you add there is
planted by the next `forge init`; `kit/forge/` lands as `.forge/`,
`kit/claude/` as `.claude/` and `kit/github/` as `.github/`. A test fails if
you add a file that the `go:embed` line does not cover.

## Adding a command

1. Implement it in `internal/cli`, keeping the rules in the packages that
   own them rather than in the command.
2. Add it to the usage text in `internal/cli/cli.go` and to
   `docs/cli.md`.
3. Cover it in `internal/cli/cli_test.go`, which runs commands end to end
   against a temporary repository.

## Supporting another agent

`AGENTS.md` is the single source of truth and the roles live in
`kit/forge/kit/agents/`. Support for a new tool is a thin wrapper that
points at them, plus the mapping in `internal/cli/init.go`. Keep the
wrapper short: duplicating a role is how the two copies start disagreeing.

## Releasing

Tag `vX.Y.Z` on `main`. The release workflow runs GoReleaser and publishes
binaries; `go install github.com/TheJisus28/forge@latest` picks up the tag.
Update `CHANGELOG.md` in the same pull request as the change, not at
release time.
