---
domain: coverage
approved_by: TheJisus28
date: 2026-09-20
---

# Convention — coverage

## Rule

A derivation that reads an artifact section to match tokens strips HTML
comments first, so a commented token is inert. One comment rule is shared by
every reader: `doc.StripComments`. A commented `AC1` is neither a task nor
evidence, exactly as it is not a pending convention for `forge archive`.

## Example

`internal/project/project.go:Spec.CriterionGaps` runs `doc.StripComments`
over the whole `tasks.md` text and over the `review.md`
`## Acceptance criteria` section before matching a criterion id. The rule
lives in `internal/doc/doc.go:StripComments`, and
`internal/cli/work.go:stripComments` delegates to it, so the archive gate and
the coverage derivation cannot drift.

## Why

`doc.Section` keeps comments, so without stripping, the guidance a template
ships — or a stray `<!-- AC1 covered -->` — reads as coverage and a spec
archives with none. This is SPEC-019's principle applied to every reader, not
just `forge archive`.

## Exceptions

A reader that must surface raw file content keeps comments as they are. Only
derivations that match tokens over a template-populated section strip them.
