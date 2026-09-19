# AGENTS

Forge is a Go CLI that plants a spec-driven agent kit into other
repositories. This file describes **this** repo, not the kit it ships.

## Stack

Go 1.22+, standard library only. No external dependencies.

## Commands

```bash
go test ./...
go run . init /tmp/scratch-repo
gofmt -l .
```

## Layout

- `kit/` — the Markdown planted into other repos, embedded with `go:embed`
- `main.go` — command parsing (`init`, `detect`, `version`, `help`)
- `internal/detect` — stack detection from repo manifests
- `internal/stackmd` — renders `forge/memory/stack.md`
- `internal/initcmd` — copies the kit, preserves project memory

## Conventions

- Everything shipped in `kit/` is English.
- Keep the scaffolded `AGENTS.md` short; agents read it every session.
- Never write outside the target directory passed to `init`.
- Project memory (`stack.md`, `decisions.md`, `constitution.md`) is only
  overwritten with `--reset-memory`.
- Conventional Commits, English summaries.
