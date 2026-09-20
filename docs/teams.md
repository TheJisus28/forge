# Working as a team

Forge assumes a protected main branch, pull requests, and people who are
not always online. Its layout exists to make merge conflicts structurally
unlikely rather than politely avoided.

## One file, one owner

- One file per spec, so two people working on different things never edit
  the same file.
- One file per decision, because a shared log would collide on every merge.
- Coverage declared by children, so a parent is never rewritten.
- `BOARD.md` is generated and gitignored: a projection cannot go stale if
  it is never stored.

## Intake

```bash
forge new "Export invoices to CSV"
git checkout -b intake/export-invoices
git add .forge/specs/SPEC-005-export-invoices.md
git commit -m "spec(spec-005): export invoices to CSV"
```

A one-file pull request that merges in minutes. This is what reserves the
number, which is why it happens before anyone starts working.

If two people run `forge new` at the same time, both get the same number,
`forge validate` catches it on the second pull request, and
`forge renumber SPEC-005` fixes it while nothing points at that id yet.

## Doing the work

```bash
forge start SPEC-005
git checkout -b spec/005-export-invoices
```

The branch holds the contract and the scaffolding in `.forge/wip/SPEC-005/`.
A teammate who checks it out sees the plan, what each phase did and what the
review found. That is the point of keeping it in the branch instead of on
someone's laptop.

Open the pull request **early, with the contract and no code**. It is the
cheapest moment to disagree.

## Review happens where your team already reviews

Forge has no maintainer list and no approval gate of its own. Accepting a
spec and approving its contract are commands anyone can run; `forge
accept SPEC-005` and `forge approve SPEC-005` record who did it, and that
is the whole mechanism. With `forge init --ci github`, `forge-validate.yml`
runs `forge validate` on every pull request and on every push to main, and
opens an issue if main ever became inconsistent.

The scrutiny you actually want — is this the right thing to build, is this
contract sane — happens in the normal pull request review, the same review
a team already does for the code. Forge does not add a second approval on
top of it. If your team wants a stricter rule (say, nobody merges their own
spec's contract), that is a branch protection or `CODEOWNERS` setting on
`.forge/specs/`, enforced by GitHub the way you enforce everything else,
not a Forge-specific concept.

## Closing

```bash
forge archive SPEC-005
git add -A && git commit -m "spec(spec-005): archive"
```

Archive refuses if the scaffolding still proposes conventions nobody
decided. Decide them first: that is how the project accumulates criteria
instead of losing it.

The pull request that lands in main contains the code, the contract with its
audit header, any new decisions and conventions, and nothing else.

## Roles

There is one role that matters to Forge: the **conductor**, whoever takes
a spec and drives the agents through it. Anyone can propose a spec, anyone
can accept one, anyone can approve a contract. Real teams already know who
should weigh in on what; Forge records who did, and does not referee it.

## Parallel work across a stack

Split by deliverable and connect with `@contract`:

```bash
forge new "Notifications"                                     # SPEC-010
forge new "API" --parent SPEC-010 --covers AC1,AC2            # SPEC-011
forge new "UI"  --parent SPEC-010 --covers AC3                # SPEC-012
```

Add `depends_on: [SPEC-011@contract]` to the UI. The moment the API contract
is approved, both are unblocked and build in parallel. If the API contract
changes later, CI tells the UI before integration does.
