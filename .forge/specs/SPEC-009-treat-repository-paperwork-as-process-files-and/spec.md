---
id: SPEC-009
title: Treat repository paperwork as process files and drop unused community docs
status: specifying
created: 2026-09-20
updated: 2026-09-20
accepted_by: TheJisus28
conductor: TheJisus28
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

Written by the architect once the work is accepted, and frozen once
approved. Real names from this repository: modules, endpoints,
tables, screens. Numbered decisions with what they discard. Anything other
specs will build against goes here.

## Out of scope

A closed list.

## History

Written by `forge`. Do not edit by hand.
- 2026-09-20  accepted  by TheJisus28
- 2026-09-20  specifying  by TheJisus28
