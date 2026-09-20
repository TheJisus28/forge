# Plan — SPEC-019

## Existing state

- `internal/cli/work.go:pendingConventions` scans the spec folder's `.md`
  files, reads a `Proposed conventions` (or `Convenciones propuestas`)
  section and calls `isNone`. `isNone` already trims and accepts `None.`,
  `none`, `ninguna`, `-`.
- `kit/machine/templates/tasks.md` is the one template that puts prose inside
  that section (lines 12-13); `review.md` already ships a bare `None.`.
- `kit.Template(name)` exists (used by `cmdTemplate` and `loadTemplate`), so a
  test can read the shipped template.
- Duplication avoided: the detector is the single place that knows a section
  is resolved; the template is the single place the guidance lives.
- Genuinely new: comment-stripping in detection, and a test driven by the
  shipped template.

## Phase 1 — Template comment and comment-aware detection

- Scope: wrap the `tasks.md` guidance in `<!-- ... -->`; strip HTML comments
  before `isNone`/`firstLine` in `pendingConventions`; test.
- Verify with: `go test ./...`, `gofmt -l .`, `go vet ./...`.

## Risks

- A body that is only a comment with no `None.` still blocks; that is the
  intended strictness (the template always ships `None.`).
