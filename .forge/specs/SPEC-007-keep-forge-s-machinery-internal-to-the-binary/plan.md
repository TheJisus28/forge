# Plan — SPEC-007

Phases, in order. One phase is one run of the implementer and one commit:
small enough to verify, large enough to mean something.

## Existing state

The machinery is already embedded and copied by one walk; this spec changes
what is planted, not how planting works.

- `kit.FS` and the walk in `internal/cli/init.go:79` already copy every
  embedded file to a mapped destination. Reuse `plant`, `mapDest`
  (init.go:142) and `kitOwned` (init.go:162); the change is one new case,
  one removed case and one cleanup step.
- `loadTemplate` (internal/cli/work.go:367) already prefers an on-disk copy
  and falls back to `kit.FS`. Reuse the function; delete the on-disk branch.
- The command shape is already uniform: `newFlagSet` + `parseArgs` +
  a case in the `Main` switch (internal/cli/cli.go:48) and one line in
  `usage`. The three new commands are read-only wrappers over `kit.FS`.
- The role text already exists once as `kit/forge/kit/agents/*.md`, and the
  host adapters already carry the host frontmatter plus a "mandatory
  context" paragraph. Reuse both; merge them.
- Docs already describe the current layout: `docs/cli.md`,
  `docs/customizing.md`, `docs/opencode.md`, `kit/forge/README.md` and the
  skills under `kit/claude/skills/`. They are updated, not rewritten.
- `kit/kit_test.go` already fails when a file on disk is not embedded, so
  moving the machinery under `machine/` is caught if the embed line is
  forgotten.
- `.forge/conventions/` has no written rules yet, so none constrain this
  work.

What does not exist yet and has to be built: the `machine/` embed root, the
`kit.Workflow/Roles/Role/Template` accessors, the marker composition in
`plant`, the three commands, and the deletion of a stale `.forge/kit/`.

## Phase 1 — Move the machinery out of the planted tree

- Scope: `git mv kit/forge/kit kit/machine`, rename `agents/` to `roles/`;
  embed `all:machine`; `mapDest` returns `""` for `machine/…`; `kitOwned`
  drops the `.forge/kit/` arm; `plant` removes a stale `.forge/kit/` after
  the walk and says so.
- Done when: `forge init` in a scratch directory writes no `.forge/kit/`,
  and `forge update` deletes one that exists. Moves AC1 and AC5.
- Verify with: `go test ./...` plus a scratch `forge init`/`update`.

## Phase 2 — Internalise the templates and expose the machinery

- Scope: add `kit.Workflow/Roles/Role/Template`; make `loadTemplate` read
  `kit.Template` only; add `forge workflow`, `forge roles [name]` and
  `forge template <name>` with their `usage` lines.
- Done when: the three commands print from the binary in a directory with
  no `.forge/`; a `.forge/kit/templates/spec.md` left in the tree does not
  change what `forge new` writes. Moves AC2 and AC4.
- Verify with: `go test ./...` and the three commands in a scratch repo.

## Phase 3 — Inline the roles into the host adapters

- Scope: move each "mandatory context" paragraph into its
  `machine/roles/<name>.md`; replace the body of every
  `kit/claude/agents/forge-*.md` and `kit/opencode/agents/forge-*.md` with
  the `{{forge-role:<name>}}` marker; compose in `plant` and fail loudly on
  an unknown marker.
- Done when: a planted adapter contains the full role text and references
  nothing under `.forge/`. Moves AC3.
- Verify with: planted `.claude/agents/forge-architect.md` in a scratch
  repo and `go test ./...`.

## Phase 4 — Documentation and tests

- Scope: update `AGENTS.md`, `kit/AGENTS.md`, `kit/forge/README.md`,
  `kit/forge/decisions/README.md`, `kit/forge/conventions/README.md`, the
  three skills, `docs/*`, `README.md`, `CONTRIBUTING.md` and `CHANGELOG.md`;
  update `internal/cli/cli_test.go`; add tests for the new commands and the
  marker composition.
- Done when: `go test ./...`, `gofmt -l .` and `go vet ./...` pass and no
  file still documents `.forge/kit/` as a planted path. Moves AC6.
- Verify with: `go test ./...`, `gofmt -l .`, `go vet ./...`.

## Risks

- An embed pattern is easy to get subtly wrong; `kit/kit_test.go` is the
  guard and must pass in Phase 1.
- A stale `.forge/kit/` deletion is destructive; it only ever removes a
  path Forge used to own and prints what it did, which is acceptable
  because decision 0001 drops backward compatibility.
- Missing a documentation reference is likely; the Phase 4 grep for
  `.forge/kit` and `kit/templates` is the check.
