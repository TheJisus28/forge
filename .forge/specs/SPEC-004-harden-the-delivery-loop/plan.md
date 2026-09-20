# Plan — SPEC-004

## Existing state

- `internal/cli/guard.go`: `cmdGuard` reads the hook payload (`tool_name`,
  `tool_input.file_path`, `cwd`), or `--file` in plain mode, and answers
  through `denial`/`isProcessFile`. It only knows about file edits.
- `internal/cli/init.go`: `writeClaudeSettings` + `mergeHook` install the
  `PreToolUse` hook with matcher `Write|Edit`; `ensureGitignore` adds
  `.forge/BOARD.md`.
- `kit/opencode/plugins/forge-guard.js`: handles `edit`, `write`,
  `apply_patch` via `forge guard --file`.
- `internal/cli/work.go`: `cmdApprove` and the `isNone` helper that already
  recognises "None.".
- `internal/validate/validate.go`: `checkArtifacts` and `planSurveysExisting`.
- `internal/project/project.go`: `Spec.Dir/PlanPath/TasksPath/ReviewPath`.
- `internal/cli/report.go`: `cmdBoard`; `internal/view/view.go`: `Board`.
- Delivered specs: SPEC-001 init/AGENTS.md, SPEC-002 submit/gh actor,
  SPEC-003 the SDD folder layout.

## Phase 1 — Guard commands, not only files

- Scope: `forge guard --command <cmd>` and a Bash hook input; deny
  `gh pr merge`, `git push` to `main`/`master`, and `git merge` on the
  default branch. Wire the Claude matcher to `Write|Edit|Bash` and teach the
  OpenCode plugin the `bash` tool.
- Verify with: `go test ./...`.

## Phase 2 — Open questions gate

- Scope: `## Open questions` in the spec template; `cmdApprove` refuses while
  it is unresolved; `validate` warns.
- Verify with: `go test ./...`.

## Phase 3 — Task progress

- Scope: `Spec.TaskProgress` from `tasks.md`; show `tasks done/total` in
  `Brief` and `Detail`.
- Verify with: `go test ./...`.

## Phase 4 — Drop the board

- Scope: remove `forge board`, `view.Board`, `cmdBoard` and
  `ensureGitignore`; update docs.
- Verify with: `go test ./...` and `forge validate`.

## Risks

- The guard could refuse a legitimate command. It only targets merges and
  pushes to the default branch; `guard: off` remains the escape hatch.
- `TaskProgress` reads a file per spec; only called for the current spec.
