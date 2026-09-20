---
id: SPEC-019
title: Archive must not read the template placeholder as a pending convention
status: done
capability: workflow
created: 2026-09-20
updated: 2026-09-20
accepted_by: TheJisus28
conductor: TheJisus28
approved_by: TheJisus28
contract_hash: c32a9d7895fe
---

## Problem

`forge archive` refuses while a spec's `plan.md`, `tasks.md` or `review.md`
still carries a `## Proposed conventions` section whose body is not exactly
`None.`. The file templates ship that section with an explanatory sentence
("Patterns decided because nothing was written. The team decides whether they
become rules…") above the `None.`, so a spec that never proposed anything
blocks its own archive until someone deletes the template's own text. It was
hit while archiving SPEC-010.

## Acceptance criteria

- AC1: `forge archive` succeeds when a spec's `Proposed conventions` section
  carries only the template's explanatory text and `None.`.
- AC2: `forge archive` still refuses when the section carries a real
  proposal, and names the file and the first line.
- AC3: The templates and the detection agree: either the placeholder text is
  not shipped in the section, or the detector ignores it. The chosen shape is
  stated in the contract.
- AC4: A test covers both the template default (archives) and a real
  proposal (refuses).

## Open questions

One per line, like `- OQ1: ...`, for anything that must be answered before
the contract is approved. Write `None.` when there are none.

None.

## Contract

### Decisions

**1. The template keeps its guidance, as an HTML comment.**
`kit/machine/templates/tasks.md`'s `## Proposed conventions` section becomes:

```markdown
## Proposed conventions

<!-- Patterns you had to decide because nothing was written. Record them in
     .forge/conventions/ or replace this comment with None. -->

None.
```

Markdown renders the comment invisibly, so the human still sees the help in
`forge template tasks` and in the file, while the section body is a comment
plus `None.`. Discards: deleting the guidance (the section becomes
unexplained), and putting the guidance in the template's intro above the
heading (it would no longer travel with the section).

**2. `pendingConventions` ignores HTML comments.** Before deciding whether a
body is a proposal, `isNone` strips `<!-- ... -->` spans (non-greedy, across
lines). A section is pending on the same rule as today otherwise: after
stripping comments and trimming, the body must be `None.` (or `-` /
`ninguna`). This is what makes the shipped default archive, and it stays
strict: a real proposal, in prose or as a bullet, is not a comment and still
blocks. Discards: teaching the detector the exact placeholder sentence (the
binary would carry a copy of the template's prose), and treating any
"mostly empty" body as None.

**3. Archive keeps naming the file and the first meaningful line.** The
existing message (`<file>: <first line>`) is unchanged; when the body keeps a
comment, the first line reported is the first non-comment, non-blank line, so
the reader sees the actual proposal.

### Interfaces other specs build against

None. This is internal to `internal/cli` and one template; no public command
or field changes.

### Tests

- `internal/cli/cli_test.go`: a spec whose `tasks.md` is the shipped template
  (read through `kit.Template("tasks")` or `forge template tasks`) archives
  successfully.
- The existing test that a real proposal (`Errors use an envelope.`) blocks
  archiving still passes, and the reported line is the proposal, not a
  comment.

### Out of scope of this contract

- Moving any other section's template prose.
- Changing when `forge archive` refuses for any reason other than a pending
  convention section.

## History

Written by `forge`. Do not edit by hand.
- 2026-09-20  accepted  by TheJisus28
- 2026-09-20  specifying  by TheJisus28
- 2026-09-20  awaiting-approval  by orchestrator
- 2026-09-20  planning  by TheJisus28
- 2026-09-20  implementing  by orchestrator
- 2026-09-20  reviewing  by orchestrator
- 2026-09-20  implementing  by orchestrator: review found doc.Section reads a ## heading inside a fenced code block, so the contract's own example blocked archive
- 2026-09-20  reviewing  by orchestrator: fixed doc.Section reading headings inside code fences
- 2026-09-20  done  by orchestrator: archived
- 2026-09-20  contract re-recorded  by orchestrator: the fence-aware Section no
  longer truncates the Contract at the quoted `## Proposed conventions`, so the
  stored hash was recomputed; the contract text is unchanged
