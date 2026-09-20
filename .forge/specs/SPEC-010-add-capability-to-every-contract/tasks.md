# Tasks — SPEC-010

The phases from the plan, as checkboxes. One phase is one implementer run
and one commit. Tick a phase when it lands and say where the work is, so a
later spec knows what exists without reading the diff.

- [x] Phase 1 — The field and the project API. Where: `internal/project/project.go` (`Spec.Capability`, `FromDoc`, `Save`, `ValidCapability`), `internal/project/project_test.go`. Landed: `Spec.Capability` beside `Title`; `FromDoc` reads `strings.TrimSpace(d.Str("capability"))`; `Save` writes `setOrDelete(s.doc, "capability", s.Capability)`; `capabilityRe = ^[a-z0-9-]+$` and `ValidCapability` beside `setOrDelete`. Tests: the `api` fixture (SPEC-002) carries `capability: notifications` and `TestLoad_ReadsSpecsAndConfig` asserts it; `TestSaveRoundTrip` sets `s.Capability = "notifications"` and asserts it survives; new `TestValidCapability` accepts `guard` and rejects `Guard`, `""`, `with_underscore`, `guard!`. Verified: `go test ./...` (all ok), `gofmt -l .` (empty), `go vet ./...` (clean).
- [x] Phase 2 — `forge new --capability`. Where: `internal/cli/work.go:cmdNew` (flag `--capability`, blank check before `project.Load`, undeclared-name warning over non-empty `Spec.Capability`, `d.SetStr("capability", cap)` before `project.FromDoc`), `internal/cli/cli.go` (usage), `internal/cli/cli_test.go` (new `TestNew_RequiresCapability`, `TestNew_WritesCapabilityAndWarnsOnANewName`; every existing `new` call given `--capability`), `internal/cli/machine_test.go` (`TestNew_IgnoresAPlantedTemplate`'s `new` call, not named by the contract). Verified: `go test ./...` (all ok), `gofmt -l .` (empty), `go vet ./...` (clean); manual run confirmed the exact missing-capability error, the exact `warning: no existing spec declares the capability "guard"; creating it as a new one`, `capability: guard` in the frontmatter, and no warning once a spec declares it.
- [ ] Phase 3 — `forge validate`. Where: `internal/validate/validate.go:checkCapability`, `internal/validate/validate_test.go`.
- [ ] Phase 4 — Surface, template and docs. Where: `internal/view/view.go`, `kit/machine/templates/spec.md`, `docs/cli.md`, this spec's frontmatter.

## Proposed conventions

Patterns decided because nothing was written. The team decides whether they
become rules in `.forge/conventions/`.

- **A non-fatal CLI notice is a `warning: ` line on `out`, and the command
  still exits 0.** `forge new` has to report an undeclared capability and
  still create the spec, and nothing in the tree said how. I wrote it with
  `fmt.Fprintf(out, "warning: ...\n")` to the command's `out` writer, leaving
  errors (which `Main` prefixes `forge: ` and turns into exit 1) as returned
  `error`s. `internal/validate` already has a `Warning` severity for its own
  findings; this is its CLI counterpart. Proposed because it is the first
  plain-text warning on a command and the next one (SPEC-011/012) will want
  the same shape.
- **Quoting a user-supplied name in a message uses `%q`.** The contract wrote
  the warning with literal double quotes; I used `%q` so the quotes and any
  escaping come from one place, matching the existing `unknown command %q`
  style in `Main`.
