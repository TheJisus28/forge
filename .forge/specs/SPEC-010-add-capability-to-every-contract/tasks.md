# Tasks — SPEC-010

The phases from the plan, as checkboxes. One phase is one implementer run
and one commit. Tick a phase when it lands and say where the work is, so a
later spec knows what exists without reading the diff.

- [x] Phase 1 — The field and the project API. Where: `internal/project/project.go` (`Spec.Capability`, `FromDoc`, `Save`, `ValidCapability`), `internal/project/project_test.go`. Landed: `Spec.Capability` beside `Title`; `FromDoc` reads `strings.TrimSpace(d.Str("capability"))`; `Save` writes `setOrDelete(s.doc, "capability", s.Capability)`; `capabilityRe = ^[a-z0-9-]+$` and `ValidCapability` beside `setOrDelete`. Tests: the `api` fixture (SPEC-002) carries `capability: notifications` and `TestLoad_ReadsSpecsAndConfig` asserts it; `TestSaveRoundTrip` sets `s.Capability = "notifications"` and asserts it survives; new `TestValidCapability` accepts `guard` and rejects `Guard`, `""`, `with_underscore`, `guard!`. Verified: `go test ./...` (all ok), `gofmt -l .` (empty), `go vet ./...` (clean).
- [x] Phase 2 — `forge new --capability`. Where: `internal/cli/work.go:cmdNew` (flag `--capability`, blank check before `project.Load`, undeclared-name warning over non-empty `Spec.Capability`, `d.SetStr("capability", cap)` before `project.FromDoc`), `internal/cli/cli.go` (usage), `internal/cli/cli_test.go` (new `TestNew_RequiresCapability`, `TestNew_WritesCapabilityAndWarnsOnANewName`; every existing `new` call given `--capability`), `internal/cli/machine_test.go` (`TestNew_IgnoresAPlantedTemplate`'s `new` call, not named by the contract). Verified: `go test ./...` (all ok), `gofmt -l .` (empty), `go vet ./...` (clean); manual run confirmed the exact missing-capability error, the exact `warning: no existing spec declares the capability "guard"; creating it as a new one`, `capability: guard` in the frontmatter, and no warning once a spec declares it.
- [x] Phase 3 — `forge validate`. Where: `internal/validate/validate.go` (`checkCapability`, called from `Run` beside the title check so it fires for every spec, including one with an unknown status), `internal/validate/validate_test.go`. Landed: absent key → `Warning` on `s.ID`, `has no capability; a new spec sets one with forge new --capability <name>`; present but `!project.ValidCapability(s.Capability)` → `Error`, `capability %q is not a lowercase slug ([a-z0-9-]+)`, so an empty value also errors. Tests: new `TestRun_WarnsMissingCapability` (via `warns`, plus no errors so exit stays 0) and `TestRun_RejectsInvalidCapability` (via `expectError`, `Guard` and `""`); `TestRun_CleanProject`'s fixture gained `capability: a`. Verified: `go test ./...` (all ok), `go test ./internal/validate/ -run 'TestRun_(WarnsMissingCapability|RejectsInvalidCapability|CleanProject)' -v` (3 pass), `gofmt -l .` (clean), `go vet ./...` (clean).
- [x] Phase 4 — Surface, template and docs. Where: `internal/view/view.go` (`Detail` row after `status`, `Brief` current-spec row between the header and `waiting on`, unexported `capabilityOrNone`), `kit/machine/templates/spec.md` (`capability: ""  # lowercase slug ([a-z0-9-]+), which part of the system this spec touches` after `status: proposed`), `docs/cli.md` (`forge new` heading and body gain `--capability`; one sentence each in `forge status`, `forge brief`, `forge validate`), `.forge/specs/SPEC-010-add-capability-to-every-contract/spec.md` (`capability: workflow`), `internal/view/view_test.go` (`TestDetail_ShowsCapability`, `TestDetail_ShowsNone`), `internal/cli/cli_test.go` (`TestStatusAndBriefShowCapability`, `TestDocs_DescribeCapability`). `internal/cli/cli.go`'s usage line was already updated in Phase 2; confirmed and left alone. Verified: `go test ./...` (all ok, including the new tests), `gofmt -l .` (empty), `go vet ./...` (clean), `go run . status SPEC-010` (prints `capability  workflow`); a scratch `forge` in `%TEMP%` confirmed `forge template spec` shows the commented line, `forge new "Guest access" --capability guard` rewrites it to `capability: guard`, and `forge status SPEC-001` prints `capability  guard`.

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
  style in `Main`. Phase 3 follows it: `capability %q is not a lowercase slug`
  over the raw `s.Capability`, so `Guard` renders as `"Guard"` and an empty
  value as `""`.
- **An intrinsic frontmatter check runs before the unknown-status `continue`
  in `Run`.** Decision 4 says `checkCapability` fires "for every spec", but
  the per-spec loop skips `checkRelations`/`checkArtifacts` when
  `!workflow.Valid(s.Status)`. I placed `checkCapability` beside the
  `missing title` check, before that `continue`, so a spec with an
  unrecognised status is still told its `capability` is absent or malformed.
  Proposed because the next field-level rule (SPEC-011's `supersedes`) faces
  the same choice: a finding about a field does not depend on the state the
  spec claims, so it belongs with the title check, not after the state gate.
- **Field-shape guidance lives as a trailing `#` comment on the template
  line.** `kit/machine/templates/spec.md` carries
  `capability: ""  # lowercase slug ([a-z0-9-]+)...`. `internal/doc`'s
  `stripComment` already drops a `# ...` comment before parsing, so the value
  stays empty and `forge new` replaces the line; `forge template spec` writes
  the embedded bytes, so a reader still sees the guidance. Proposed because it
  is the first field shape documented in a template and the next one
  (SPEC-011's `supersedes`) needs the same place to say it.
