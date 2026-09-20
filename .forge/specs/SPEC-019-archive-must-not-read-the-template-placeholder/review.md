# Review - SPEC-019 Archive must not read the template placeholder as a pending convention

Verdict: **pass**

Reviewed on branch `spec/019-archive-must-not-read-the-template-placeholder`
against contract `57f52814717d`. Product code not touched by the review.

## Criterion / evidence

| # | Criterion | Result | Evidence |
|---|---|---|---|
| AC1 | Archive succeeds when the section carries only the template's text and `None.` | pass | `TestArchive_IgnoresTemplateConventionsComment` writes the shipped `forge template tasks` output as `tasks.md` and `forge archive SPEC-001` succeeds (`status: done`). The section body is the HTML comment plus `None.`, and `stripComments` leaves `None.`. |
| AC2 | Archive still refuses a real proposal and names the file and first line | pass | Same test: with the comment kept and `- Errors use an envelope.` under it, archive exits non-zero and the message contains `Errors use an envelope.` and not the comment text. The existing `TestArchiveKeepsTheRecord` (`Errors use an envelope.`) still passes. |
| AC3 | Templates and detection agree; the chosen shape is in the contract | pass | Contract decision 1: the guidance ships as an HTML comment; decision 2: `isNone`/detection sees the body after `stripComments`. `kit/machine/templates/tasks.md` now carries `<!-- ... -->` above `None.`; `review.md`'s template was already a bare `None.`. |
| AC4 | A test covers the template default (archives) and a real proposal (refuses) | pass | `TestArchive_IgnoresTemplateConventionsComment` covers both branches in one test, driven by the shipped template rather than a hand-written string. |

Suite: `go test ./...` all `ok`; `gofmt -l .` empty; `go vet ./...` clean.

After the first pass the spec was sent back to `implementing`: its own
`spec.md` blocked `forge archive`, because the frozen contract quotes the
`## Proposed conventions` heading inside a fenced code block and
`doc.Section` read that example as the real section. The fix tracks fenced
code blocks in `doc.Section`
(`TestSection_IgnoresHeadingsInCodeFences`), leaves the contract untouched,
and makes `forge archive SPEC-019` succeed. All criteria were re-verified on
the fixed tree.

## Contract integrity

The recorded `contract_hash` matches and `forge brief` reports no drift. The
regex `(?s)<!--.*?-->` is non-greedy and DOTALL, so a multi-line comment is
removed whole and a later comment is not merged with an earlier one; a
malformed `<!--` without `-->` is not treated as a comment and remains a
proposal, which is the safe direction.

## Problems

None blocking the merge and none blocking the close.

## Proposed conventions

None.
