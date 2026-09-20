# Review — SPEC-010

Verdict: pass with notes

Reviewed `spec/010-add-capability-to-every-contract` at `8083d66` (phases
`a5a6e36`, `4be8a29`, `1ed2700`, `c9f26f1`; `8083d66` ticks tasks and moves
to reviewing). The binary was built from this tree into `%TEMP%` and every
CLI criterion was run from a scratch repository outside this one.

## Acceptance criteria

| Criterion | Result | Evidence |
|---|---|---|
| AC1 `forge new` without `--capability` fails, creates nothing | pass | Scratch repo: `forge.exe new "X"` -> `forge: a capability is required: pass --capability <name>, as in forge new "Pay with a saved card" --capability payments`, exit 1; `.forge/specs` gained no folder. `TestNew_RequiresCapability` PASS. |
| AC2 `--capability guard` writes `capability: guard` | pass | Scratch repo: `forge.exe new "Guest access" --capability guard` -> `.forge/specs/SPEC-001-guest-access/spec.md` frontmatter contains `capability: guard`. `TestNew_WritesCapabilityAndWarnsOnANewName` PASS. |
| AC3 `forge validate` warns per spec lacking `capability`, no build failure | pass | `go run . validate` here prints `warning SPEC-001:` … `SPEC-009: has no capability; a new spec sets one with forge new --capability <name>` (and SPEC-011…018, the untracked proposals). A scratch repo with one missing-capability spec prints the warning and exits 0. `TestRun_WarnsMissingCapability` PASS. See notes on the repo's own exit code. |
| AC4 undeclared name warns and still creates, exit 0 | pass | Scratch repo: first `new "Guest access" --capability guard` prints `warning: no existing spec declares the capability "guard"; creating it as a new one`, creates the spec, exit 0; a second `new … --capability guard` prints no warning. `TestNew_WritesCapabilityAndWarnsOnANewName` PASS. |
| AC5 present but empty or non-slug is an error | pass | `new "Bad cap" --capability Guard` then `validate` -> `error SPEC-003: capability "Guard" is not a lowercase slug ([a-z0-9-]+)`, exit 1; hand-edited `capability: ""` -> `error SPEC-002: capability "" is not a lowercase slug ([a-z0-9-]+)`, exit 1. `TestRun_RejectsInvalidCapability` PASS. |
| AC6 status/brief show it; template and docs document it | pass | `go run . status SPEC-010` -> `capability  workflow`; `forge brief` on the spec branch -> `  capability  workflow`; empty value renders `(none)` in both. `forge template spec` -> `capability: ""  # lowercase slug ([a-z0-9-]+), …`. `docs/cli.md:43` heading has `--capability <name>`, body lines 46-48 explain it. `TestStatusAndBriefShowCapability`, `TestDocs_DescribeCapability`, `TestDetail_ShowsCapability`, `TestDetail_ShowsNone` PASS. |

Toolchain: `go test ./...` all `ok`, exit 0; `gofmt -l .` no output; `go vet ./...` no output.

Contract integrity: no drift. `forge validate` reports no "contract changed
after approval" for SPEC-010, the `contract_hash` is intact, and the promised
interfaces exist as written — `project.Spec.Capability`,
`project.ValidCapability`, `Save` preserving a non-empty value and dropping
an empty one. Decision 7 is honored: commit `c9f26f1` adds only
`capability: workflow` to SPEC-010's frontmatter; the contract, status and
history are untouched by that phase (the later `status: reviewing` is the
normal Forge transition).

## Blocking problems

None block the merge. All six acceptance criteria pass.

## Notes

- **`go run . validate` on this repository exits 1 while the review is being
  written.** The only error is `SPEC-010: is reviewing without
  …/review.md`; every capability finding is a warning. With this file in
  place that error clears and the command exits 0 (warnings alone do not set
  the exit code, `internal/validate/validate.go:304` `HasErrors`, and
  `--quiet` hides them, `internal/cli/report.go:96`). This matches AC3: the
  warnings are visible, the build is not failed by them.
- **`TestValidate_Capability` was not written.** The contract's `### Tests`
  and the AC3/AC5 Verification lines name an end-to-end
  `internal/cli/cli_test.go` test `TestValidate_Capability`. It does not
  exist; validate behaviour is covered by `TestRun_WarnsMissingCapability`
  and `TestRun_RejectsInvalidCapability` (unit), plus `TestLifecycle`'s
  final `forge validate` (cli). The criteria themselves are verified above,
  so this is a deviation from the contract's test list, not a behaviour
  defect.
- **Stale `forge new "<title>"` uses outside the three documents decision 6
  names.** `kit/AGENTS.md:25`, `.forge/README.md:28`,
  `kit/forge/README.md:28`, `kit/claude/skills/forge-work/SKILL.md:22`,
  `kit/machine/roles/orchestrator.md:26`, `internal/view/view.go:162` and
  `internal/cli/guard.go:122` still invite a bare `forge new`, which now
  errors. Decision 6 scoped the documentation change to the template,
  `cli.go` and `docs/cli.md`, so this is out of contract and non-blocking,
  but following those hints leads straight into the required-flag error.
- **AC3 on this working tree warns for SPEC-011…018 too**, not only
  SPEC-001…009. Those folders are untracked proposals created outside this
  spec; they legitimately have no capability yet. The criterion says "each
  spec that has no capability", so the extra warnings are correct.
- **The first spec in a fresh repository always warns** (no existing spec
  declares the name). The contract accepted this (`plan.md`, risks). Exit is
  0, so AC4 holds.

## Proposed conventions

None.
