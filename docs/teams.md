# Working as a team

Forge assumes a protected main branch, pull requests, and people who are
not always online. Its layout exists to make merge conflicts structurally
unlikely rather than politely avoided.

## One file, one owner

- One folder per spec, so two people working on different things never edit
  the same file.
- One file per decision, because a shared log would collide on every merge.
- Coverage declared by children, so a parent is never rewritten.

## Intake

```bash
forge new "Export invoices to CSV"
git checkout -b intake/export-invoices
git add .forge/specs/SPEC-005-export-invoices/spec.md
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

The branch holds the spec folder `.forge/specs/SPEC-005-export-invoices/`,
with the contract, the plan, the tasks and the review. A teammate who checks
it out sees what each phase did and what the review found. That is the point
of keeping it in the branch instead of on someone's laptop.

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
forge submit SPEC-005
```

`forge submit` pushes the branch and opens the pull request through `gh`,
recording it on the spec. It never merges: a person reviews and merges on
GitHub. Without `gh` it prints the `git push` and `gh pr create` commands
instead of failing. `forge guard` denies `gh pr merge` and any `git push`
or `git merge` that lands on the default branch, so the pull request is the
only way in.

Archive refuses if the spec still proposes conventions nobody decided.
Decide them first: that is how the project accumulates criteria instead of
losing it.

The pull request that lands in main contains the code, the spec folder with
its contract, plan, tasks and review, any new decisions and conventions, and
nothing else. Archive keeps the folder as the durable record; it is not
deleted, so a later spec can read what was built and why.

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
