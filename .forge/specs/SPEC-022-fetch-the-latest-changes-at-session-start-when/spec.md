---
id: SPEC-022
title: Fetch the latest changes at session start when the project opts in
status: reviewing
capability: agents
created: 2026-09-20
updated: 2026-09-20
accepted_by: TheJisus28
orchestrator: TheJisus28
approved_by: TheJisus28
contract_hash: 346680c3f13a
---

## Problem

A session starts from whatever a clone already had, and the brief never
freshens it: SPEC-015 decision 7 made fetching the user's call, so the
`forge brief` injected at session start can describe refs that are days
behind the remote. The brief notices the staleness (`internal/view/view.go`)
and prints `run forge status --fetch`, but the agent cannot act on a warning
it only pays for after the fact. The work then starts against a picture
that is wrong, and the mistake is discovered late, if at all. The cost is
concrete: `forge new` picks the next id from the refs on `origin/main` and
renumbers on a collision, so a stale clone can mint an id the remote has
already taken and only find out when it pushes. In a project
whose remote is GitHub and whose `gh` is already authenticated, that is a
question the session could answer for itself at startup, cheaply and
safely: a fetch moves no branch and touches no file.

## Acceptance criteria

Observable outcomes. Someone else must be able to mark each one pass or
fail with evidence. A criterion is verifiable when it names the evidence that
settles it: a backticked command, a `test`/`TestName`, or an observable verb
such as `returns`/`refuses`. `forge approve` refuses while a criterion names
none of those, so write it before approval.

- AC1: With the opt-in scalar `fetch: on` in `.forge/project.md`, a GitHub
  remote and an authenticated `gh`, `forge brief` runs `git fetch` before it
  renders, so the remote-tracking refs are current without a manual
  `forge status --fetch`; `TestBrief_FetchesWhenOptedIn` returns the brief
  with `origin/main` advanced to a commit that exists only on the remote and
  the stale-clone note gone.
- AC2: Without the scalar (or when it is not `on`), `forge brief` performs
  no `git fetch` and no `gh` call and renders from the refs already present;
  `TestBrief_OfflineByDefault` returns a before/after snapshot in which the
  remote is untouched and the missing remote-only spec stays missing.
- AC3: `forge brief` never integrates remote changes: it does not run
  `git pull`, `merge` or `rebase`, and `git status --porcelain` and the
  current branch are identical before and after; a snapshot test
  `TestBrief_FetchDoesNotTouchWorkingTree` proves the tree is unchanged.
- AC4: When `gh` is missing or not authenticated, or the repository has no
  GitHub remote, `forge brief` skips the fetch, emits the brief from local
  refs and exits 0; `TestBrief_SkipsWithoutGitHub` returns the brief with no
  fetch.
- AC5: When the fetch fails (no network or an unreachable remote),
  `forge brief` prints a `warning: ` line naming the failure, still emits
  the brief from local state and exits 0; `TestBrief_FetchFailureIsAWarning`
  settles it, per `.forge/conventions/cli-output.md`.
- AC6: `docs/cli.md`, the roles and `.forge/project.md` document the `fetch`
  scalar, state that it only fetches, and correct the claim that the binary
  never touches the network by itself; `go run . brief` output matches the
  docs.
- AC7: `go test ./...` passes, and `gofmt -l .` and `go vet ./...` report
  nothing.

## Open questions

One per line, like `- OQ1: ...`, for anything that must be answered before
the contract is approved. Write `None.` when there are none.

None.

## Contract

Builds on SPEC-005 (`forge brief` and the `SessionStart` hook), SPEC-015
decision 7 (the binary does not fetch by default; `forge status --fetch`
stays the user's call and reads only refs already present), SPEC-020
decision 6 (the opt-in scalar shape) and decision 7 (a failed side effect is
a `warning: ` line and exit 0). Every name below is this repository's.

1. **The brief fetches only when the project opts in through `fetch: on`,
   read by `FetchEnabled()`, default off.** `internal/project/project.go`
   gains `func (p *Project) FetchEnabled() bool`, reading the `fetch`
   frontmatter scalar exactly as `PushEnabled` reads `push`: `on`, `true` or
   `yes`, case-insensitive, is on; anything else, including an absent key, is
   off. `.forge/project.md` gains `fetch: on` so Forge dogfoods it, and its
   `## Facts an agent cannot guess` network bullet is corrected. Discards: a
   `--fetch` flag on `forge brief` (the `SessionStart` hook runs `forge brief
   --json` with no arguments, so the opt-in has to live in the repository)
   and fetching by default (SPEC-015 decision 7 left it to the user; this
   keeps the default offline and adds an opt-in path beside
   `push:`/`guard:`).

2. **The opt-in fetches only a GitHub remote with an authenticated `gh`.**
   New `internal/project/git.go: func HasGitHubRemote(root string) bool`: it
   lists the configured remotes with the existing `run` helper (`git remote`,
   then `git remote get-url <name>` for each) and reports true when any URL
   contains `github.com`. The gate in `cmdBrief` is `p.FetchEnabled()`, then
   `HasGitHubRemote(p.Root)`, then `project.GHUser(p.Root) != ""`; `GHUser`
   already reports `""` when `gh` is missing or not authenticated (it checks
   `HasGH`, then runs `gh api user`). The remote check runs before `GHUser`,
   so a repository with no GitHub remote makes no `gh` call. Discards:
   fetching whenever any remote exists (the no-GitHub-remote case AC4 names)
   and skipping the `gh` check (the premise is a GitHub project the user
   works through `gh`; a bare `git fetch` can also block on a credential
   prompt inside the session hook).

3. **The fetch lives in `cmdBrief`, never in `view.Brief`.** `internal/cli/
   report.go` gains `func fetchBeforeBrief(p *project.Project) (warning
   string, fetched bool)`. `cmdBrief` calls it before `view.Brief(p)`; when
   `fetched` is true it reloads with `project.Load(p.Root)`, the way
   `cmdStatus --fetch` already does, so the brief reads the refs just
   fetched (if the reload fails, the already-loaded project is kept and the
   brief still renders). `view.Brief` stays a formatter over the loaded
   project and gains no fetch. Discards: the fetch inside `view.Brief` (it
   would put `os/exec` in the view package and force a network seam through
   every view test) and orchestrating in `internal/project` (the warning
   text and the `out` writer are CLI concerns).

4. **The only network operation is the existing `git fetch --quiet`; the
   working tree, the index and the branch are untouched.** `fetchBeforeBrief`
   calls the existing `project.Fetch(root)` unchanged. No `pull`, `merge`,
   `rebase`, `checkout` or `reset` is added anywhere. Because `FetchAge`
   measures `FETCH_HEAD`/`refs/remotes/origin/HEAD`, a successful fetch also
   clears the brief's `note: this clone last fetched N days ago` line, which
   is the criterion's observable: a fetch updates refs, not the working
   tree, so a spec that exists only on `origin/main` cannot appear in a brief
   built from `.forge/specs/` (AC1's original wording is corrected for this;
   listing remote-only specs would need reading `origin/main`, which is out
   of scope). Discards: `git pull --ff-only` or `fetch && merge` (they move
   the branch and can conflict with uncommitted work, which "moves no branch
   and touches no file" forbids) and fetching a named ref (`git fetch` with
   no refspec is what `forge status --fetch` already runs).

5. **An opted-in brief that cannot fetch warns and continues; only a fetch
   that ran can change the refs.** `fetchBeforeBrief` returns one line
   beginning `warning: ` when the project opted in and the fetch could not be
   done or failed:
   - no GitHub remote: `warning: no GitHub remote configured; using local refs`
   - `gh` missing or not authenticated: `warning: gh is not authenticated; using local refs`
   - `git fetch` failed: `warning: git fetch failed: <err>; using local refs`

   `cmdBrief` prepends the warning to the text and returns nil, so `Main`
   exits 0. When the fetch succeeds there is no warning; when `fetch: on` is
   absent `fetchBeforeBrief` returns immediately and calls neither `gh` nor
   `git fetch`. Discards: returning the git error the way `forge status
   --fetch` does (a session-start hook must not fail the session because the
   network is down) and staying silent when an opted-in project cannot fetch
   (the user asked for it and the reason is actionable).

6. **The same text feeds `forge brief` and `forge brief --json`.**
   `cmdBrief` keeps its single `text := view.Brief(...)` and prepends the
   warning to it before either the plain write or the JSON marshal, so the
   human and the Claude Code hook see identical content. Discards: writing
   the warning separately to `out` (the `--json` payload is one line of
   JSON; a preceding warning line would corrupt it) and warning only in the
   human path (the hook is where sessions start).

7. **The docs and roles state the opt-in and that a fetch only fetches.**
   - `docs/cli.md`: the "binary never reaches the network" preamble lists
     `forge brief` among the commands whose network comes from `git`/`gh`,
     and `### forge brief [--json]` documents `fetch: on`, that it runs `git
     fetch` and never `pull`/`merge`/`rebase`, and the warning behaviour.
   - `README.md` `## Local-first`: qualify "`git` and `gh` run only when you
     ask" with "or when the project opts in with `fetch: on`".
   - `.forge/project.md` `## Facts an agent cannot guess`: the network bullet
     says the brief fetches at session start when `fetch: on`, and only
     fetches.
   - `AGENTS.md` "No network in the binary": name the opt-in fetch beside
     `forge status --fetch`.
   - `kit/machine/roles/orchestrator.md`: the session/state section says a
     project with `fetch: on` refreshes its remote refs through `forge brief`
     at session start, and that this is a fetch only. `kit/machine/roles/
     architect.md` `## Before writing` item 1 names the scalar the same way.
     Discards: documenting it in one place (the claim lives in several files
     and they drift) and changing `kit/forge/project.md` (the template mirrors
     `guard`, not `push`, and the scalar is optional, so a target project
     learns it from the docs).

### Interfaces

- `internal/project/project.go: func (p *Project) FetchEnabled() bool` — the
  `fetch` scalar, `on`/`true`/`yes`, default off.
- `internal/project/git.go: func HasGitHubRemote(root string) bool` — any
  configured remote URL contains `github.com`; reuses `run`. Existing
  `Fetch(root) error`, `GHUser(root) string`, `HasGH() bool` and `run` are
  reused unchanged.
- `internal/cli/report.go: func fetchBeforeBrief(p *project.Project)
  (warning string, fetched bool)`; `cmdBrief` calls it before `view.Brief`
  and reloads after a successful fetch. Two seams go beside the existing
  `hasGH`/`pushBranch`/`ghRun`: `githubRemote = project.HasGitHubRemote` and
  `ghUser = project.GHUser`. `fetchRefs` is deliberately not a seam: the
  tests use a real local bare remote, as `gitRepo`/`checkpointRepo` already
  do, so no test touches the network.
- `.forge/project.md` frontmatter: `fetch: on`.
- Reused, unchanged: `view.Brief`, `project.Fetch`, `project.FetchAge`,
  `project.Load`, `project.GHUser`, `project.HasGH`, `internal/project/git.go:run`.

### Tests

- AC1 `internal/cli/brief_internal_test.go:TestBrief_FetchesWhenOptedIn`
  (package `cli`, so the seams are reachable): a local bare `origin` with a
  commit that exists only on the remote, `.forge/project.md` with `fetch:
  on`, a `FETCH_HEAD` older than 48h; swap `githubRemote` to true and
  `ghUser` to `"tester"`; call `cmdBrief`; the brief advances `origin/main`
  to the remote commit and no longer prints the `last fetched` note, while
  the local spec list is unchanged.
- AC2 `TestBrief_OfflineByDefault`: no `fetch` scalar; swap `githubRemote`
  and `ghUser` for spies that fail the test if called; call `cmdBrief`; exit
  0, the brief renders, `origin/main` and `FETCH_HEAD` are byte-identical
  before and after, and no `warning: ` appears.
- AC3 `TestBrief_FetchDoesNotTouchWorkingTree`: fetch on with a dirty
  worktree; snapshot `git status --porcelain`, `git rev-parse HEAD` and `git
  rev-parse --abbrev-ref HEAD`; call `cmdBrief`; the three are identical and
  the output carries no pull/merge/rebase notice.
- AC4 `TestBrief_SkipsWithoutGitHub`: fetch on, `githubRemote` swapped to
  false; exit 0, the brief renders, `origin/main`/`FETCH_HEAD` unchanged and
  a `warning: no GitHub remote` line; a second case with `githubRemote` true
  and `ghUser` empty covers not authenticated.
- AC5 `TestBrief_FetchFailureIsAWarning`: fetch on, `githubRemote` true,
  `ghUser` `"tester"`, the `origin` URL set to a missing path (the pattern
  `TestAdvance_PushFailureIsAWarning` uses); `cmdBrief` returns nil and the
  output starts `warning: ` and still contains the brief.
- `internal/project/project_test.go:TestFetchEnabled` — the
  on/true/yes/off/absent table, mirroring `TestPushEnabled`.
- `internal/project/git_test.go:TestHasGitHubRemote` — `origin` at
  `https://github.com/x/y.git` or `git@github.com:x/y.git` is true; a local
  bare path or no remote is false. No network.

## Existing state

- SPEC-005 introduced `forge brief` and the Claude Code `SessionStart` hook
  (`internal/cli/init.go:263`, `forge brief --json`); SPEC-015 decision 7
  kept fetching the user's call; SPEC-020 decision 6 is the opt-in scalar
  shape and decision 7 the warn-and-continue rule. This spec builds on all
  three and rewrites none of their state.
- `internal/view/view.go:88` already prints the staleness note from
  `project.FetchAge`, so a successful fetch is observable through the
  existing brief with no view change.
- `internal/project/git.go`: `Fetch` (107), `FetchAge` (80), `GHUser` (126),
  `HasGH` (198), `Branch`, `run` (206). All are reused; only
  `HasGitHubRemote` is new, because nothing reads a remote URL today
  (`RemoteSpecIDs` reads a git ref, not the remote list).
- `internal/project/project.go`: frontmatter is read with `Config.Str`, and
  `GuardEnabled`/`PushEnabled` (180-194) are the exact shape `FetchEnabled`
  copies.
- `internal/cli/report.go`: `cmdStatus --fetch` (34-41) is the existing
  fetch-then-reload pattern; the seam block (17-21) is where the new seams
  go; `cmdBrief` (55-87) is the insertion point.
- Seams and hermetic tests already exist to copy: `swap` in `internal/cli/
  upgrade_internal_test.go`, the `package cli` seam-faking in `internal/cli/
  submit_internal_test.go:TestSubmit_CommitsBeforePush`, and the local bare
  remotes in `internal/project/git_test.go:gitRepo` and `internal/cli/
  cli_test.go:checkpointRepo`, so no test reaches the network.
- Conventions that govern this change: `.forge/conventions/cli-output.md`
  (`warning: ` line, exit 0), `testing.md` (a seam only where a test must
  substitute it; prove "writes nothing" with a before/after snapshot),
  `architecture.md` (the rule and its derivation live in the owning
  package), `frontmatter.md` (a scalar is read through `Config.Str`).
- Deliberately not duplicated: the fetch reuses `project.Fetch` and the
  reload reuses `project.Load`; the warning reuses the `warning: `
  convention; the scalar reuses the `PushEnabled` shape.
- Not built yet: `FetchEnabled`, `HasGitHubRemote`, `fetchBeforeBrief`, the
  two seams, the `fetch: on` value and the docs/role edits.

## Out of scope

- `git pull`, `git merge`, `git rebase`, `git checkout`, `git reset`, or any
  change to the working tree, the index, HEAD or the current branch.
- Listing spec folders that exist only on `origin/main`: a fetch updates
  refs, not the working tree, and discovery stays with `RemoteSpecIDs` and
  `forge new`/`forge accept`.
- A `--fetch` flag on `forge brief`; a background, scheduled or retried
  fetch; a fetch timeout.
- Changing `forge status --fetch`: it keeps its explicit flag and its
  fail-loud error.
- Fetching in any other command (`forge advance`, `forge push`, `forge
  submit`).
- Authenticating `gh`, prompting for credentials, or handling SSH keys.
- Making `fetch` default on, or reading it anywhere but `.forge/project.md`.
- Rewriting `## History`, delivered `.forge/specs/*` or `.forge/decisions/*`
  records.

## History

Written by `forge`. Do not edit by hand.
- 2026-09-20  accepted  by TheJisus28
- 2026-09-20  contracting  by TheJisus28
- 2026-09-20  planning  by TheJisus28
- 2026-09-20  implementing  by orchestrator
- 2026-09-20  reviewing  by orchestrator
