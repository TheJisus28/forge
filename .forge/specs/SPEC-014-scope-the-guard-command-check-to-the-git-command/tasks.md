# Tasks — SPEC-014

- [x] Phase 1 — Segment-scoped command check. Landed in
  `internal/cli/guard.go`: `commandDenial` now walks `splitSegments(command)`
  and, for each segment, asks `gitCommand` for the subcommand. `gitCommand`
  only accepts `git` as the segment's first word (skipping `VAR=value` and
  global options `-C`, `-c`, `--git-dir`, `--work-tree`, `--namespace`,
  `--exec-path`); `refspecsToDefault` + `targetsDefault` judge only that
  segment's positional args (either side of `:`, after stripping `+`,
  `refs/heads/`, `heads/`, `refs/`). `reGHMerge` and `onDefaultBranch` are
  unchanged; the whole-line `reGitPush`/`reGitMerge`/`pushesToDefault` are
  gone.
  Tests in `internal/cli/cli_test.go`: `TestGuardCommand` gains the compound
  allows (`&&`, `;`, `||`, `echo git push…`) and `git push origin
  main:feature`/`git -C . push origin master` denies, plus `--explain`
  checks; `TestGuardCommand_BarePushFollowsTheBranch` builds a real repo on
  `main` and shows `git push` denied there and allowed on `feature`.
  Verified: `go test ./...` all `ok`; `gofmt -l .` empty; `go vet ./...`
  clean. By hand, feeding the line from a file (the installed guard in this
  environment still blocks the literal string),
  `go run . guard --command "<compound>" --explain` prints `would allow` and
  `git push origin main` prints `would deny`.

## Proposed conventions

None.

## Notes for the reviewer

- The AC1 line cannot be typed directly in this workspace: the opencode
  guard hook runs the installed (pre-fix) `forge`, whose whole-line scan
  blocks the literal command before it executes. That is the bug, observed
  live; the fix was verified through the test suite and by passing the line
  from a file so the old hook never sees the token.
