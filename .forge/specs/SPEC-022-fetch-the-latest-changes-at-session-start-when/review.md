# Review — SPEC-022

Verdict: pass

Reviewed by the `reviewer` role on
`spec/022-fetch-the-latest-changes-at-session-start-when`, against the
approved contract `346680c3f13a`. This is a read-only review: no `status`
change, no archive and no product code touched. Merge base
`9d131f648128c948f40173f0e7702803411ca342` (`main`).

## Acceptance criteria

| Criterion | Result | Evidence |
|---|---|---|
| AC1 | pass | `go test ./internal/cli/ -run TestBrief_FetchesWhenOptedIn -v` -> `--- PASS: TestBrief_FetchesWhenOptedIn (2.85s)`. The test (`internal/cli/brief_internal_test.go:132`) builds a local bare `origin` with a commit only the remote has, ages `FETCH_HEAD` 3 days, swaps `githubRemote`->true and `ghUser`->"tester", runs `cmdBrief`, then asserts `git rev-parse origin/main` equals the remote-only SHA, the output has no `last fetched` note and no `warning:`, and the local `SPEC-001` still lists. `fetchBeforeBrief` (`internal/cli/report.go:107`) calls the real `project.Fetch` (`git fetch --quiet`) after the `FetchEnabled -> HasGitHubRemote -> GHUser` gate. |
| AC2 | pass | `go test ./internal/cli/ -run TestBrief_OfflineByDefault -v` -> `--- PASS`. `brief_internal_test.go:168` runs with no `fetch` scalar and swaps `githubRemote` and `ghUser` for spies that call `t.Error` if invoked; it asserts `cmdBrief` returns nil, `origin/main` equals the before value, `FETCH_HEAD` is byte-identical and the output has no `warning:`. The spies prove no `gh` call, and `fetchBeforeBrief` returns at the `FetchEnabled` check before any seam (`report.go:108`). |
| AC3 | pass | `go test ./internal/cli/ -run TestBrief_FetchDoesNotTouchWorkingTree -v` -> `--- PASS`. `brief_internal_test.go:202` snapshots `git status --porcelain`, `git rev-parse HEAD` and `git rev-parse --abbrev-ref HEAD` around an opted-in `cmdBrief` on a dirty tree; all three are identical and the output carries no pull/merge/rebase notice. `git diff <merge-base> HEAD -- internal/` grepped for `pull|merge|rebase|checkout|reset` returns only test comments/data; the only network call added is `project.Fetch`, unchanged (`internal/project/git.go:131`, `git fetch --quiet`). |
| AC4 | pass | `go test ./internal/cli/ -run TestBrief_SkipsWithoutGitHub -v` -> `--- PASS`. `brief_internal_test.go:239` case 1: `githubRemote` false with a `ghUser` spy that must not be called; output starts `warning: no GitHub remote`, `SPEC-001` renders, `origin/main`/`FETCH_HEAD` unchanged, err nil. Case 2: `githubRemote` true, `ghUser` empty; output starts `warning: gh is not authenticated`, err nil. |
| AC5 | pass | `go test ./internal/cli/ -run TestBrief_FetchFailureIsAWarning -v` -> `--- PASS`. `brief_internal_test.go:286` points `origin` at a missing path (no network) and runs `cmdBrief`; err is nil, the output starts `warning: `, names `git fetch failed` and still contains the brief (`Rules that matter here`). The `--json` run also returns nil, is one line (`strings.TrimRight(payload,"\n")` has no `\n`) and carries `warning: git fetch failed`. |
| AC6 | pass | `go test ./internal/cli/ -run 'TestDocPages_DocumentTheOptInFetch\|TestRoles_DocumentTheOptInFetch' -v` -> both `--- PASS`. The docs test checks the `docs/cli.md` preamble names `forge brief`, the `### forge brief` section names `fetch: on`, `git fetch`, `pull`/`merge`/`rebase` and the three warnings, and `README.md`, `AGENTS.md` and `.forge/project.md` carry the opt-in. `go run . roles orchestrator` and `go run . roles architect` print `fetch: on` and `a fetch only`. `go run . brief --json` exits 0 and prints one line; `go run . brief` exits 0; `CHANGELOG.md` has the `[Unreleased]` `### Added` entry; `kit/forge/project.md` is untouched. |
| AC7 | pass | `go test ./...` -> every package `ok` (root, `internal/cli`, `internal/doc`, `internal/project`, `internal/validate`, `internal/view`, `internal/workflow`, `kit`), exit 0. `gofmt -l .` -> no output, exit 0. `go vet ./...` -> no output, exit 0. |

## Contract decisions

- Decision 1: `FetchEnabled()` (`internal/project/project.go:200`) copies the
  `PushEnabled` shape, `on`/`true`/`yes` case-insensitive; `TestFetchEnabled`
  covers on/true/yes/ON/off/false/absent and PASSes. `.forge/project.md` carries
  `fetch: on` and the corrected network bullet.
- Decision 2: the gate order in `fetchBeforeBrief` is `p.FetchEnabled()`
  (line 108), `githubRemote(p.Root)` (111), `ghUser(p.Root) != ""` (114). The
  `ghUser` spy in `TestBrief_SkipsWithoutGitHub` proves no `gh` call when there
  is no GitHub remote. `HasGitHubRemote` (`internal/project/git.go:69`) lists
  remotes with the existing `run` and matches `github.com`;
  `TestHasGitHubRemote` passes over a local bare path, `https://` and `git@`,
  and over a directory with no repository.
- Decision 3: only `internal/cli/report.go` gained `fetchBeforeBrief`, beside
  the `githubRemote`/`ghUser` seams (lines 21-22). `git diff` shows no change
  under `internal/view/`, `view.Brief` is untouched, and the view package has
  no `os/exec` (`rg` exits 1). `cmdBrief` reloads with `project.Load` on
  `fetched` and keeps the old project when the reload fails (line 74).
- Decision 4: `fetchBeforeBrief` calls the existing `project.Fetch` unchanged;
  no `pull`/`merge`/`rebase`/`checkout`/`reset` invocation was added.
- Decision 5: the three warning strings match the contract verbatim
  (`report.go:112`, `:115`, `:118`) and `cmdBrief` returns nil in every
  cannot-fetch path, so `Main` exits 0.
- Decision 6: one `text := view.Brief(p)`; the warning is prepended to it
  before both the plain write and the JSON marshal, so the human and the hook
  see identical content and the payload stays one line.

## Blocking problems

None. `go test ./...`, `gofmt -l .` and `go vet ./...` are clean, and every
criterion has a passing named test or command.

## Notes

1. AC2's wording says "the missing remote-only spec stays missing"; the
   fixture has a remote-only *commit* (`remote.txt`), not a remote-only spec,
   and the test pins the refs (`origin/main`, `FETCH_HEAD`) as unchanged
   instead of a spec list. Contract decision 4 removed listing remote-only
   specs from scope (a fetch moves refs, not the working tree), so this is a
   spec-wording artefact, not a behaviour gap.
2. AC6's "correct the claim that the binary never touches the network by
   itself" is honoured only as far as contract decision 7 asks: `docs/cli.md`
   keeps its "binary never reaches the network" sentence but now lists
   `forge brief (only with fetch: on)` among the commands whose network comes
   from `git`/`gh`. The AC phrasing is broader than the approved contract; the
   implementation follows the contract.
3. The dogfood `go run . brief --json` performed a real `git fetch` (exit 0,
   one line, no warning) because this repository has `fetch: on`, a GitHub
   remote and an authenticated `gh`; the warning-carried JSON path is covered
   hermetically by `TestBrief_FetchFailureIsAWarning`.
4. `go run . check` reported the seven SPEC-022 evidence gaps only because
   `review.md` did not exist yet; the five SPEC-015 task warnings are
   pre-existing and untouched by this spec. This review supplies the
   AC1..AC7 tokens so `forge check SPEC-022` is covered.

5. The join rule this review would have proposed was decided and recorded in
   `.forge/conventions/cli-output.md` ("Warning that shares one text with a
   machine payload"), so `## Proposed conventions` is `None.`

## Proposed conventions

None.