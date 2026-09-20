# Tasks — SPEC-010

The phases from the plan, as checkboxes. One phase is one implementer run
and one commit. Tick a phase when it lands and say where the work is, so a
later spec knows what exists without reading the diff.

- [ ] Phase 1 — The field and the project API. Where: `internal/project/project.go` (`Spec.Capability`, `FromDoc`, `Save`, `ValidCapability`), `internal/project/project_test.go`.
- [ ] Phase 2 — `forge new --capability`. Where: `internal/cli/work.go:cmdNew`, `internal/cli/cli.go`, `internal/cli/cli_test.go`.
- [ ] Phase 3 — `forge validate`. Where: `internal/validate/validate.go:checkCapability`, `internal/validate/validate_test.go`.
- [ ] Phase 4 — Surface, template and docs. Where: `internal/view/view.go`, `kit/machine/templates/spec.md`, `docs/cli.md`, this spec's frontmatter.

## Proposed conventions

Patterns decided because nothing was written. The team decides whether they
become rules in `.forge/conventions/`.

None.
