---
name: forge-work
description: Drive a Forge spec from proposal to archive. Use when the user asks for a feature or a change, wants to continue work on a spec, mentions a SPEC id, or asks what to work on next.
---

# Driving a spec

You are the orchestrator. Read `.forge/kit/agents/orchestrator.md` if you
have not in this session.

## Where are we

Run `forge status`. The state of the spec decides what happens next, not
your memory of the conversation.

## New work

1. If it is one deliverable: `forge new "<title>"`. Write the problem and
   the acceptance criteria with the user. Criteria must be verifiable by
   someone else.
2. If it spans several deliverables: propose a parent spec with the outcome
   criteria, get them confirmed, then create children with `--parent` and
   `--covers` so nothing is promised without an owner.
3. Tell the user to open the intake pull request. A maintainer accepts it
   by approving that pull request; you never accept on their behalf.

## Accepted work

```bash
forge start SPEC-00X          # refuses while dependencies are open
git checkout -b spec/00X-slug
```

Launch `forge-architect` for the contract. When it comes back, summarise
the decisions for the user and ask a maintainer to approve the pull
request. Approval is theirs; asking for it is yours.

## Approved work

Write the phases in `.forge/wip/<id>/plan.md`, then
`forge advance <id> --to implementing`. Launch `forge-implementer` once per
phase, and report back between phases.

When every phase is done: `forge advance <id> --to reviewing` and launch
`forge-reviewer`. If the review fails, go back to implementing with a note
saying why.

## Closing

```bash
forge archive <id>
```

It refuses if there are unresolved convention proposals, which is the
point: decide them first. Then commit and mark the pull request ready.
