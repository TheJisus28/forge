# Tasks — SPEC-022

The phases from the plan, as checkboxes. One phase is one implementer run
and one commit. Tick a phase when it lands and say where the work is, so a
later spec knows what exists without reading the diff. Name the criteria the
phase delivers (`Moves: <criterion ids>`), so every criterion is traceable to
the task that moves it; `forge check` reads these ids from this file.

- [x] Phase 1 — The opt-in scalar and GitHub detection. Moves: AC2, AC4.
  Where: `internal/project/project.go` (`FetchEnabled`),
  `internal/project/git.go` (`HasGitHubRemote`); tests in
  `internal/project/project_test.go`, `internal/project/git_test.go`.
  Landed: `(*Project).FetchEnabled` reads the `fetch` scalar (`on`/`true`/
  `yes`, case-insensitive; absent off) beside `PushEnabled`.
  `HasGitHubRemote(root) bool` lists remotes with `run` and reports true when
  any URL contains `github.com`. Tests: `TestFetchEnabled` (on/true/yes/off/
  absent over the `write` fixture), `TestHasGitHubRemote` (https and `git@`
  GitHub URLs true; a local bare path or no remote false, over `gitRepo`, no
  network).
  Verified: `go test ./...` all `ok`; `gofmt -l .` empty; `go vet ./...`
  clean; both tests pass with `-v`.
- [x] Phase 2 — The brief fetches when opted in. Moves: AC1, AC2, AC3, AC4,
  AC5. Where: `internal/cli/report.go` (`fetchBeforeBrief`, the `githubRemote`
  and `ghUser` seams, `cmdBrief`); tests in
  `internal/cli/brief_internal_test.go` (new).
  Landed: `fetchBeforeBrief(p)` gates on `p.FetchEnabled()`, then
  `githubRemote(p.Root)`, then `ghUser(p.Root) != ""`, then
  `project.Fetch(p.Root)`; it returns the three decision-5 `warning: ` lines
  and `fetched=true` only on success. `cmdBrief` calls it before
  `view.Brief`, reloads with `project.Load` on `fetched` (keeping the loaded
  project if the reload fails), prepends the warning to the single `text`
  before both the plain write and the `--json` marshal, and returns nil.
  Tests: `TestBrief_FetchesWhenOptedIn`, `TestBrief_OfflineByDefault`,
  `TestBrief_FetchDoesNotTouchWorkingTree`, `TestBrief_SkipsWithoutGitHub`,
  `TestBrief_FetchFailureIsAWarning`, over a real local bare `origin` and a
  remote-only commit (no network).
  Verified: `go test ./...` all `ok`; `gofmt -l .` empty; `go vet ./...`
  clean; the five tests pass with `-v`.
- [x] Phase 3 — Docs, roles, changelog and dogfooding. Moves: AC6, AC7. Where:
  `docs/cli.md`, `README.md`, `AGENTS.md`, `.forge/project.md`,
  `kit/machine/roles/orchestrator.md`, `kit/machine/roles/architect.md`,
  `CHANGELOG.md`; tests in `internal/cli/cli_test.go`,
  `internal/cli/machine_test.go`.
  Landed: `docs/cli.md` names `forge brief` in the network preamble and
  documents `fetch: on`, the fetch-only behaviour (`git fetch`, never
  `pull`/`merge`/`rebase`) and the three warnings; `README.md`
  `## Local-first` and `AGENTS.md` "No network in the binary" carry the
  opt-in; `.forge/project.md` gains `fetch: on` and the corrected network
  bullet; `orchestrator.md`/`architect.md` state that a `fetch: on` project
  refreshes its remote refs through `forge brief` at session start and that
  this is a fetch only; `CHANGELOG.md` gains the `[Unreleased]` entry.
  Tests: `TestDocPages_DocumentTheOptInFetch` (`cli_test.go`) and
  `TestRoles_DocumentTheOptInFetch` (`machine_test.go`).
  Verified: `go test ./...` all `ok`; `gofmt -l .` empty; `go vet ./...`
  clean; the two tests pass with `-v`; `go run . roles orchestrator` and
  `go run . roles architect` render the new line; `go run . brief --json`
  exits 0 on one line.

## Proposed conventions

None.
