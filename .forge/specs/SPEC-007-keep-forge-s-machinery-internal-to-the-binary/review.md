# Review — SPEC-007

Verdict: pass

## Acceptance criteria

| Criterion | Result | Evidence |
|---|---|---|
| AC1 | pass | `TestInit_PlantsTheKitAndKeepsYourContent` fails if `.forge/kit` is planted; a scratch `forge init` leaves `.forge/` with only `project.md`, `README.md`, `specs/`, `decisions/`, `conventions/`. |
| AC2 | pass | `TestNew_IgnoresAPlantedTemplate`: with a `JUNK` `.forge/kit/templates/spec.md` present, `forge new` writes the embedded spec (`## Contract`, no `JUNK`). |
| AC3 | pass | `TestInit_InlinesTheRolesIntoHostAdapters`: planted `.claude/agents/forge-*.md` and `.opencode/agents/forge-*.md` carry the role text (`# Architect`), reference nothing under `.forge/kit`, and keep no `{{forge-role:…}}` marker. |
| AC4 | pass | `TestWorkflowCommand_PrintsTheWorkflow`, `TestRolesCommand_ListsAndPrints`, `TestTemplateCommand_PrintsAndRejects`; in a directory with no `.forge/`, `forge workflow` prints `# Workflow` and `forge roles` lists the four roles. |
| AC5 | pass | `TestUpdate_RemovesStaleKit` and `TestInit_PlantsTheKitAndKeepsYourContent` (project.md survives `init --force` + `update`); dogfooding here printed `removed the old .forge/kit` and kept a hand edit in `.forge/decisions/README.md`. |
| AC6 | pass | `go test ./...` is ok, `gofmt -l .` is empty, `go vet ./...` is clean, and `forge validate` reports `7 specs, no problems`. |

Commands run:

```
go test ./...            # all packages ok
gofmt -l .               # (empty)
go vet ./...             # (empty)
forge validate           # 7 specs, no problems
forge update             # removed the old .forge/kit
```

## Blocking problems

None.

## Notes

- The kit adapters now hold only their host frontmatter and a
  `{{forge-role:<name>}}` marker; `compose` in `internal/cli/init.go` fills
  them from `kit/machine/roles/<name>.md`. A marker naming no role fails the
  plant instead of shipping an empty agent.
- This repository was dogfooded: `forge update` deleted `.forge/kit/` and
  regenerated `.claude/` and `.opencode/`, which is why those files appear in
  the diff. The two sub-READMEs under `.forge/` are not Forge-owned, so their
  pointer to the new `forge template` was updated by hand.
- The working copy had CRLF line endings in several Go files (Windows
  `core.autocrlf`), so `go fmt ./...` normalised them to LF. Git normalises
  on commit, so this adds no content diff.
- `CHANGELOG.md` keeps the previous release's historical mention of
  `.forge/kit/`; released history is not rewritten.

## Proposed conventions

None.
