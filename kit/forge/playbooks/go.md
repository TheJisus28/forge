# Playbook — Go

Load only if `stack.md` lists `go`.

## Signals

`go.mod`, `*.go`, `internal/`.

## Habits

- One module per repo unless `stack.md` or an ADR says otherwise.
- Keep the existing layout. Do not "hexagonalize" a flat package on
  the side.
- Sentinel errors (`var ErrNotFound = errors.New(...)`) when the repo
  already uses them.
- Tests: `go test` with the flags in `stack.md`.
- Migrations: the tool already in the stack (dbmate, golang-migrate,
  goose). Do not switch tools in an unrelated SPEC.

## Do not

- Add a `go.mod` per folder "just in case".
- Introduce Echo/Chi/Gin if the repo already chose one.
- Do network I/O inside a SQL transaction.
