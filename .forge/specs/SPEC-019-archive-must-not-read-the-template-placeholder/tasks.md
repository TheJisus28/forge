# Tasks — SPEC-019

- [x] Phase 1 — Template comment and comment-aware detection. Landed:
  `kit/machine/templates/tasks.md` ships the `Proposed conventions` guidance
  as an HTML comment above `None.`; `internal/cli/work.go` gains
  `stripComments` (a `<!-- ... -->` non-greedy, DOTALL regex) and calls it in
  `pendingConventions` before `isNone`/`firstLine`, so the comment is neither
  a proposal nor the reported line.
  `internal/cli/cli_test.go` adds
  `TestArchive_IgnoresTemplateConventionsComment`, which reads the shipped
  template with `forge template tasks`: with a real proposal under the
  comment archiving is refused and the proposal line is named (not the
  comment); with the template default the spec archives.
  Verified: `go test ./...` all `ok`; `gofmt -l .` empty; `go vet ./...`
  clean.

## Proposed conventions

None.

## Notes for the reviewer

- The first attempt failed because the test replaced the wrong `None.`: the
  comment text itself ends with "replace this comment with None.", so a naive
  `strings.Replace` rewrote inside the comment. The test now replaces the
  standalone `\nNone.\n` line, which is the real section body.
