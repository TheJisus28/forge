# Plan — SPEC-014

## Existing state

- `internal/cli/guard.go` owns `commandDenial`: `reGHMerge`, `reGitPush`,
  `reGitMerge`, `pushesToDefault`, `onDefaultBranch`. The push/merge regexes
  and `pushesToDefault` are the bug; `reGHMerge` and `onDefaultBranch` stay.
- `internal/cli/cli_test.go:TestGuardCommand` already lists the deny/allow
  commands and is the place to add the compound cases; `runGit`,
  `newRepo`, `run`, `mustRun`, `write` are available.
- `project.Branch(root)` reads the current branch for the AC3 check; a test
  needs a real git repo with a commit to exercise it.
- Convention: `.forge/conventions/testing.md` (seams only where a test must
  substitute; here no seam is needed, git runs for real in `t.TempDir`).
- Duplication avoided: one `commandDenial` path for both the hook and
  `--explain` (decision 4).
- Genuinely new: segment splitting, quote-aware fielding, git-subcommand
  recognition, refspec classification.

## Phase 1 — Segment-scoped command check

- Scope: rewrite `commandDenial` internals in `internal/cli/guard.go`;
  extend `TestGuardCommand` in `internal/cli/cli_test.go` for AC1–AC5.
- Verify with: `go test ./...`, `gofmt -l .`, `go vet ./...`, and
  `go run . guard --command "..." --explain` by hand for the compound line.

## Risks

- Over-blocking a legitimate refspec is worse than under-blocking here, but
  `feature/main` must stay allowed; the classifier strips only the known
  ref prefixes, not any `/main` suffix.
- Splitting on `2>&1` produces empty segments; the git recognition ignores
  them, so no behaviour change.
