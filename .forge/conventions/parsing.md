---
domain: parsing
approved_by: TheJisus28
date: 2026-09-20
---

# Convention — parsing

## Rule

An id read out of artifact text is matched as a whole token, never with
`\b`. A token is a run of `[0-9A-Za-z_-]+`, compared case-insensitively with
`strings.EqualFold`. So `AC1` never equals `AC10`, `AC1x` or `AC1-`, while
two ids side by side (`AC1 AC2`, `AC1,AC2`) are both found.

## Example

`internal/project/project.go:hasToken` tokenizes with
`regexp.MustCompile("[0-9A-Za-z_-]+")` and compares each token, and
`CriterionGaps` calls it for both `tasks.md` and the review section.
`TestCriterionGaps_MatchesBoundedTokens` pins the adjacency cases and the
`AC1-`/`AC10` negatives.

## Why

Go's `\b` treats `-` as a non-word character, so `\bAC1\b` matches inside
`AC1-` and grants false coverage; a consuming delimiter pattern can eat the
separator between two ids. Comparing whole tokens avoids both.

## Exceptions

An id that is already isolated by structure — a map key, a frontmatter value —
needs no tokenizer.
