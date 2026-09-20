# Plan — SPEC-022

The survey lives once, in `spec.md`'s `## Existing state` (the architect
wrote it). Phases below reuse what it names instead of rebuilding it.

## Phase 1 — The opt-in scalar and GitHub detection

- Scope: `internal/project/project.go` gains `func (p *Project)
  FetchEnabled() bool`, reading the `fetch` scalar beside `GuardEnabled`/
  `PushEnabled` (`on`/`true`/`yes`, case-insensitive; anything else, absent
  included, is off). `internal/project/git.go` gains `func
  HasGitHubRemote(root string) bool`, listing the configured remotes with the
  existing `run` helper and reporting true when any URL contains
  `github.com`. `Fetch`, `GHUser`, `HasGH` and `run` are reused unchanged.
  Tests: `internal/project/project_test.go:TestFetchEnabled` (the
  on/true/yes/off/absent table, mirroring `TestPushEnabled`) and
  `internal/project/git_test.go:TestHasGitHubRemote` (an `https://` and a
  `git@` GitHub URL true; a local bare path or no remote false, over the
  `gitRepo` fixture, no network). Moves AC2, AC4.
- Done when: `FetchEnabled` answers exactly as decision 1 says and
  `HasGitHubRemote` recognises a GitHub remote without touching the network.
- Verify with: `go test ./...`, `gofmt -l .`, `go vet ./...`.

## Phase 2 — The brief fetches when opted in

- Scope: `internal/cli/report.go` gains `func fetchBeforeBrief(p
  *project.Project) (warning string, fetched bool)` and two seams beside the
  existing `hasGH`/`pushBranch`/`ghRun`: `githubRemote =
  project.HasGitHubRemote` and `ghUser = project.GHUser`. The gate is
  `p.FetchEnabled()`, then `githubRemote(p.Root)`, then `ghUser(p.Root) !=
  ""`; only then `project.Fetch(p.Root)` (decision 2, 4). `cmdBrief` calls it
  before `view.Brief`, reloads with `project.Load(p.Root)` when `fetched` is
  true (keeping the loaded project if the reload fails), prepends the
  `warning: ` line to the one text so plain and `--json` match, and returns
  nil so `Main` exits 0 (decision 3, 5, 6). `view.Brief` is not changed.
  Tests in a new `internal/cli/brief_internal_test.go` (package `cli`, so the
  seams are reachable): `TestBrief_FetchesWhenOptedIn` (AC1),
  `TestBrief_OfflineByDefault` (AC2), `TestBrief_FetchDoesNotTouchWorkingTree`
  (AC3), `TestBrief_SkipsWithoutGitHub` (AC4),
  `TestBrief_FetchFailureIsAWarning` (AC5), over a real local bare `origin`
  as `gitRepo`/`checkpointRepo` already do. Moves AC1, AC2, AC3, AC4, AC5.
- Done when: with `fetch: on` and a faked GitHub seam the brief advances
  `origin/main` and drops the stale note; without the scalar it makes no call
  and changes nothing; the tree, index and branch are untouched in every
  case; every cannot-fetch path warns and still renders the brief.
- Verify with: `go test ./...`, `gofmt -l .`, `go vet ./...`.

## Phase 3 — Docs, roles, changelog and dogfooding

- Scope: `docs/cli.md` names `forge brief` in the network preamble and
  documents `fetch: on`, the fetch-only behaviour and the warnings;
  `README.md` `## Local-first` gains the opt-in; `AGENTS.md` "No network in
  the binary" names the opt-in fetch; `.forge/project.md` gains `fetch: on`
  and corrects the `## Facts an agent cannot guess` network bullet;
  `kit/machine/roles/orchestrator.md` and `architect.md` state that a
  `fetch: on` project refreshes its refs through `forge brief` and that this
  is a fetch only; `CHANGELOG.md` gains the `[Unreleased]` entry. Tests in
  `internal/cli/cli_test.go` (the docs assertions) and
  `internal/cli/machine_test.go` (the role printouts) scope to the sections
  and assert stable anchors. Moves AC6, AC7.
- Done when: the docs, roles and `project.md` agree, `go run . brief` matches
  the documented behaviour, and the changelog describes the change.
- Verify with: `go test ./...`, `gofmt -l .`, `go vet ./...`.

## Risks

- **No network in any test.** The fetch is exercised against a local bare
  `origin` under `t.TempDir()` and the GitHub/gh checks are seams swapped in
  the `package cli` test; `nonet_test.go` and the hermetic fixture stay
  intact. Never point a test at `github.com`.
- **Only `git fetch`.** No `pull`, `merge`, `rebase`, `checkout` or `reset`
  anywhere (decision 4). AC3's snapshot test is the guard against a silent
  integration.
- **The hook must not fail the session.** Every cannot-fetch path is a
  `warning: ` line and exit 0 (decision 5), never a returned error like
  `forge status --fetch`. Do not weaken `forge status --fetch`, which stays
  fail-loud and flag-driven (out of scope).
- **The `--json` payload stays one line.** The warning is prepended to the
  text before the marshal (decision 6); writing it separately would corrupt
  the Claude Code hook payload.
- **Default stays offline.** `fetch: on` absent means `fetchBeforeBrief`
  returns before any `gh` or `git fetch` call (decision 1). Do not make it
  default on.
- Scope creep: no `--fetch` flag on `forge brief`, no fetch in `forge
  status`/`advance`/`push`/`submit`, no timeout, no credential prompting,
  and no listing of specs that exist only on `origin/main`.
