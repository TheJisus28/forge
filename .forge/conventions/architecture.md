---
domain: architecture
approved_by: TheJisus28
date: 2026-09-20
---

# Convention — architecture

## Rule

A rule in `internal/validate` that judges a spec's artifacts delegates the
derivation to the package that owns it and only assigns severity. It does not
re-read `tasks.md` or `review.md` itself.

## Example

`internal/validate/validate.go:checkCriteriaCoverage` calls
`project.Spec.CriterionGaps()` and maps each gap to a `Warning` or `Error`,
so the token matcher and the state table live only in `internal/project`.

## Why

A second reader of the same artifacts drifts from the first and gives a
different verdict for the same spec. This is the single-source rule SPEC-018
enforced for the workflow and roles.

## Exceptions

A rule that checks artifact structure rather than derived content — whether a
file exists, whether a heading is present (`checkArtifacts`) — reads what it
needs directly.
