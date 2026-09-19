# Contributing

Forge is a small Go module with no dependencies. Keep it that way.

## Ground rules

- Kit Markdown is English. `working_language` exists for *target* repos.
- Standard library only in the CLI unless there is a strong reason.
- The scaffolded `AGENTS.md` stays short: Codex concatenates project docs
  under a size budget, and long instruction files get truncated.
- Forge never edits a project's source code, only its own files.
- Conventional Commits, English summaries (`type(scope): summary`).

## Local development

```bash
go test ./...
go run . detect /path/to/some/repo
go run . init /tmp/scratch-repo
```

## Where things live

| Path | What |
|---|---|
| `kit/` | the kit itself: agents, memory, playbooks, templates. Embedded via `go:embed` |
| `main.go` | command parsing and help text |
| `internal/detect` | manifest-based stack detection |
| `internal/stackmd` | renders `forge/memory/stack.md` |
| `internal/initcmd` | copies the kit, preserves project memory |

Most contributions are Markdown under `kit/`. Anything added there is planted
by the next `forge init`; `kit/cursor/`, `kit/claude/`, and `kit/github/` land
as the dotted directories `.cursor/`, `.claude/`, `.github/`.

## Adding a language playbook

1. Add `kit/forge/playbooks/<lang>.md` following the shape of the existing
   ones: signals, habits, "do not".
2. Teach `internal/detect` to append `<lang>` to `Playbooks` when the matching
   manifest exists, and to fill `test` / `dev` / `runtime`.
3. Add a case to `internal/detect/detect_test.go`.
4. List it in `kit/forge/playbooks/README.md` and in the detection table of
   the README.

## Releasing

Tag `vX.Y.Z` on `main`. The release workflow runs GoReleaser and publishes
binaries; `go install github.com/TheJisus28/forge@latest` picks up the tag.
