---
id: SPEC-009
title: Treat repository paperwork as process files and drop unused community docs
status: done
capability: guard
created: 2026-09-20
updated: 2026-09-20
accepted_by: TheJisus28
conductor: TheJisus28
approved_by: TheJisus28
contract_hash: 6dea588b01b7
---

## Problem

The guard classifies everything outside a short allowlist as product code,
so a repository's own paperwork at the root is denied while no spec is
implementing. Correcting `CHANGELOG.md` or deleting a boilerplate file
forces a spec that delivers no product change. This repository also carries
community boilerplate — a code of conduct, a contributing guide and a
security policy — that duplicates `AGENTS.md` or serves a single maintainer,
and `CONTRIBUTING.md` is the only contributor guide while `AGENTS.md` is
already the single source of truth. The two problems are the same boundary
seen from both sides: the guard does not know what paperwork is.

## Acceptance criteria

Observable outcomes. Someone else must be able to mark each one pass or
fail with evidence.

- AC1: With `guard: on` and no spec implementing, editing a Markdown file at
  the repository root is allowed: `forge guard --file CHANGELOG.md` exits 0.
- AC2: The same guard still denies product code: with no spec implementing,
  `forge guard --file internal/cli/guard.go` and `forge guard --file main.go`
  exit 1.
- AC3: `CODE_OF_CONDUCT.md`, `CONTRIBUTING.md` and `SECURITY.md` are gone
  from the tree, and no tracked file references them by name.
- AC4: The contributor instructions that still matter — adding a command and
  supporting another agent — live in `AGENTS.md`; `README.md` no longer links
  to `CONTRIBUTING.md`.
- AC5: The guard's definition of process files is documented where the guard
  is documented, and `go test ./...`, `gofmt -l .` and `go vet ./...` pass.

## Open questions

One per line, like `- OQ1: ...`, for anything that must be answered before
the contract is approved. Write `None.` when there are none.

None.

## Contract

One change in `internal/cli/guard.go` (the `isProcessFile` rule), the
documentation that states that rule, and the removal of three unused
repository-root files. No new command, no new package, no new dependency, no
change under `kit/`.

### Decisions

1. **The rule: repository-root paperwork is a process file.** `isProcessFile`
   keeps its signatures and its existing allowlist, and gains one more
   condition: a path is a process file when it is *at the repository root* and
   is either Markdown (extension `.md` or `.markdown`, case-insensitive) or a
   licence/notice file (`LICENSE` or `NOTICE`, or any name starting
   `LICENSE.`/`NOTICE.`, case-insensitive). Concretely, `isProcessFile` ends
   with `return isRootPaperwork(rel)` after the prefix loop and the exact-name
   switch. This is a category, which is what the Problem asks for, and it
   absorbs `CHANGELOG.md`, `CONTRIBUTING.md`, `SECURITY.md` and
   `CODE_OF_CONDUCT.md` without naming any of them. Discards adding four more
   names to the exact-name switch: that list would deny the next paperwork
   file (`GOVERNANCE.md`, `ARCHITECTURE.md`) and force another spec like this
   one for every new boilerplate file.
2. **Root-only, and what stays product code.** "Root" means the
   slash-relative path contains no `/`. Markdown inside a directory (`docs/*`,
   `internal/*`, `kit/*`) stays product code, because Forge's own documentation
   and a user's hosted docs live there and must be gated. The existing
   prefixes `.forge/`, `.claude/`, `.github/`, `.opencode/` and the exact-root
   names `.gitignore`, `opencode.json`, `opencode.jsonc` are unchanged; the
   root `.md` names already in the switch (`AGENTS.md`, `CLAUDE.md`,
   `README.md`) stay listed even though rule 1 now also covers them, so the
   existing behaviour is not re-derived from the new rule. Everything else is
   product code, including `main.go`, `go.mod`, `Makefile`, `package.json`
   and every document beneath a directory. Discards a blanket "all Markdown is
   process" rule, which would let an agent edit a product's `docs/` without a
   spec.
3. **Accepted risk, stated in the docs.** In a repository whose product *is*
   root-level Markdown (a book, a template collection), rule 1 lets an agent
   edit those files with no spec implementing. The escape stays the same as
   today: work on a spec branch, or set `guard: off`. Rule 4 records this in
   the guard documentation rather than pretending the category is exact.
4. **Documentation (AC5).** The rule is stated in the three pages that
   describe the guard today, so they cannot disagree:
   - `docs/customizing.md`, `## The guard` (the canonical paragraph): replace
     the sentence "Product code is anything outside `.forge/`, `.claude/`,
     `.github/` and the root pointers." with the rule above and the
     root-only caveat, and keep a closing note that a configurable allowlist
     is a feature request, not part of this change.
   - `docs/cli.md`, the `### forge guard` section: add one sentence naming the
     process files and that nested Markdown is product code.
   - `docs/opencode.md`, `## The guard`: replace "the root pointers" in the
     "rule is unchanged" paragraph with the same definition.
   Discards documenting only the code change: AC5 asks for prose, and
   `docs/opencode.md` currently states the old rule, so leaving it would be a
   false statement.
5. **Remove the unused community docs (AC3).** Delete `CONTRIBUTING.md`,
   `SECURITY.md` and `CODE_OF_CONDUCT.md`. `CODE_OF_CONDUCT.md` is already
   deleted in the working tree; the implementation commits all three
   deletions. Keep `AGENTS.md`, `README.md`, `CLAUDE.md`, `LICENSE`,
   `CHANGELOG.md`, `.gitignore`, `opencode.json` and `opencode.jsonc`.
   Discards keeping `SECURITY.md`'s vulnerability-reporting channel; that
   content is dropped, not moved, because it serves a single maintainer.
6. **Migrate what still matters into `AGENTS.md` (AC4).** Add a
   `## Contributing` section (placed after `## Layout`, before `## Rules that
   are not negotiable`) with the two sections from `CONTRIBUTING.md` that are
   not already in `AGENTS.md`:
   - `### Adding a command` — the three steps (implement in
     `internal/cli` keeping rules in their own packages; add to the usage
     text in `internal/cli/cli.go` and `docs/cli.md`; cover in
     `internal/cli/cli_test.go`).
   - `### Supporting another agent` — `AGENTS.md` is the single source, roles
     live in `kit/machine/roles/`, a new host is a thin wrapper with the
     frontmatter and the `{{forge-role:<name>}}` marker filled by
     `forge init`, mapped in `internal/cli/init.go`, and role text is never
     duplicated.
   `AGENTS.md` also gains one sentence under `## Releases`: update
   `CHANGELOG.md` in the same pull request as the change, not at release
   time. The old "Ground rules", "Local development", "Where things live" and
   "Releasing" sections are already covered by `AGENTS.md` and
   `.forge/project.md`, so they are not copied. Discards moving the content
   into a new file (the point is one source) and discards dropping the
   changelog sentence along with `CONTRIBUTING.md`.
7. **`README.md` stops linking `CONTRIBUTING.md` (AC4).** In `##
   Contributing`, replace the `[CONTRIBUTING.md](CONTRIBUTING.md)` link with
   `[AGENTS.md](AGENTS.md)` and keep the "kit is Markdown under `kit/`"
   sentence. Discards leaving a dangling link.
8. **How AC3 is read.** "No tracked file references them by name" means no
   *live* file: the durable record under `.forge/specs/` is excluded, because
   it necessarily quotes the names — the archived SPEC-007 folder mentions
   `CONTRIBUTING.md`, and SPEC-009's own Problem and acceptance criteria name
   all three. Archived specs are not rewritten; that would edit the record the
   project keeps on purpose. `CHANGELOG.md` will announce the removal without
   naming the files, so the live tree stays clean. Reviewer evidence:
   `git grep -n -E "CONTRIBUTING|SECURITY|CODE_OF_CONDUCT" -- ":!.forge/specs"`
   prints nothing, and the three paths do not exist.
9. **`CHANGELOG.md`.** Add one `### Changed` bullet under `[Unreleased]`
   describing the guard's process-file rule and the removal of the unused
   community boilerplate (without naming the files, per rule 8). This follows
   the rule preserved in rule 6 and is how every previous spec recorded its
   change. Discards skipping the entry.
10. **This decision outlives the spec.** The classification of repository-root
    paperwork is a durable product rule that later guard work (a configurable
    allowlist, more roots) will build on or supersede. The orchestrator should
    record it in `.forge/decisions/` after approval, as the next id, titled
    "Repository-root paperwork is a process file", superseding nothing.

### Interfaces (exact)

Changed: `internal/cli/guard.go`, package `cli`:

- `func isProcessFile(root, file string) bool` — signature unchanged; it
  normalises to a slash-relative path, allows `../` (outside the repository),
  the prefixes `.forge/`, `.claude/`, `.github/`, `.opencode/`, the exact names
  `.gitignore`, `opencode.json`, `opencode.jsonc`, `AGENTS.md`, `CLAUDE.md`,
  `README.md`, and finally calls the new helper.
- `func isRootPaperwork(rel string) bool` — new, unexported. `rel` is
  slash-separated and relative to the repository root. Returns `false` when
  `rel == ""` or it contains `/`. Otherwise returns `true` when
  `strings.ToLower(filepath.Ext(rel))` is `.md` or `.markdown`, or when the
  upper-cased base name is `LICENSE`/`NOTICE` or starts with `LICENSE.`/
  `NOTICE.`. No exported API changes; the guard's CLI surface
  (`forge guard [--explain] [--file path] [--command <cmd>]`) is unchanged.

Changed documentation: `docs/customizing.md`, `docs/cli.md`,
`docs/opencode.md` (per rule 4).

Changed repository files: `README.md` (rule 7), `AGENTS.md` (rule 6),
`CHANGELOG.md` (rule 9).

Removed: `CONTRIBUTING.md`, `SECURITY.md`, `CODE_OF_CONDUCT.md`.

Unchanged, deliberately: `kit/` in full. `kit/opencode/plugins/forge-guard.js`
only shells out to `forge guard --file`, `kit/AGENTS.md` is the generic
planted pointer set, and no `kit/machine/` file enumerates process files, so
the planted copy needs no edit and must not gain this repository's opinions.
`internal/cli/init.go`, `internal/cli/cli.go`, `main.go`, `.forge/project.md`,
`.forge/conventions/*` are untouched.

### Test plan

- AC1 (root Markdown allowed): extend the allowed slice of
  `TestGuardFileMode` in `internal/cli/cli_test.go` (it already runs in
  `newRepo`, which is `guard: on` with no spec) with `CHANGELOG.md`, and
  assert `run(t, dir, "guard", "--file", "CHANGELOG.md")` returns code 0. The
  guard only inspects the string, so the file need not exist.
- AC2 (product code still denied): in the same test add a denied slice —
  `internal/cli/guard.go`, `main.go`, `docs/customizing.md` (nested Markdown)
  and `go.mod` (root non-paperwork) — and assert code 1 with the output
  containing `no spec` for each. The nested and root non-Markdown cases prove
  the rule is root-paperwork, not "root".
- AC3 (files gone, no live reference): a new black-box test in
  `internal/cli/cli_test.go` reads `../../` (the pattern `upgrade_test.go`
  already uses) and asserts `os.Stat` fails for `CONTRIBUTING.md`,
  `SECURITY.md` and `CODE_OF_CONDUCT.md`, and that `README.md` does not
  contain `CONTRIBUTING.md`. Review evidence is the `git grep` command in
  decision 8 over the whole tree except `.forge/specs/`.
- AC4 (instructions survive): the same test reads `../../AGENTS.md` and
  asserts it contains `## Contributing`, `Adding a command` and `Supporting
  another agent`; it reads `../../README.md` and asserts it contains
  `AGENTS.md` and does not contain `CONTRIBUTING.md`.
- AC5 (documented and green): a test reads `../../docs/customizing.md` and
  asserts it contains `Markdown` and `LICENSE` (the definition per rule 4).
  `go test ./...`, `gofmt -l .` and `go vet ./...` all pass.

### Out of scope (confirmed)

As in `## Out of scope` below: no configurable allowlist or new
`.forge/project.md` key, no change to nested Markdown, `.cursor/` or other
configuration roots, no change to `commandDenial`, no change to `guard: off`
or `forge init --no-guard`, no change under `kit/`, no rewriting of archived
specs, and no replacement for the removed `SECURITY.md` reporting channel.

### Open questions

None.

## Out of scope

- A configurable process-file allowlist, or any new key in
  `.forge/project.md`; the rule is compiled into the binary.
- Treating Markdown or documentation inside a directory (for example
  `docs/`, `internal/`) as a process file.
- Other configuration roots such as `.cursor/`, `.devcontainer/` or
  `CODEOWNERS`; they keep today's behaviour.
- Changing the shell-command denial (`gh pr merge`, `git push`/`git merge`
  into the default branch) or the `guard: off` switch.
- Editing the planted kit (`kit/`), including
  `kit/opencode/plugins/forge-guard.js`.
- Rewriting the archived SPEC-007 folder or any other delivered spec to
  remove the names of the deleted files.
- Preserving or replacing the content of `SECURITY.md`, `CONTRIBUTING.md` or
  `CODE_OF_CONDUCT.md` beyond the migration in decision 6.
- Adding, removing or renaming any other repository-root file.

## History

Written by `forge`. Do not edit by hand.
- 2026-09-20  accepted  by TheJisus28
- 2026-09-20  specifying  by TheJisus28
- 2026-09-20  awaiting-approval  by orchestrator
- 2026-09-20  planning  by TheJisus28
- 2026-09-20  implementing  by orchestrator
- 2026-09-20  reviewing  by orchestrator
- 2026-09-20  done  by orchestrator: archived
