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
3. Open the intake pull request, then accept it into the queue:
   `forge accept SPEC-00X`. Forge has no approval list to check against;
   the pull request review is the real check.

## Accepted work

```bash
forge start SPEC-00X          # refuses while dependencies are open
git checkout -b spec/00X-slug
```

Launch `forge-architect` for the contract. When it comes back, summarise
the decisions for the user, make sure `## Open questions` in
`.forge/specs/<id>/spec.md` says `None`, and once the contract is right:
`forge approve SPEC-00X`.

## Approved work

First survey what already exists: `forge status`, the contracts of the
delivered specs, and the code that already does part of the job. Fill
`## Existing state` in `.forge/specs/<id>/plan.md` with what you will reuse
and what is genuinely new, write the phases in `.forge/specs/<id>/plan.md`
and the same phases as checkboxes in `.forge/specs/<id>/tasks.md`, then run
`forge advance <id> --to implementing`. Launch `forge-implementer` once per
phase, and report back between phases.

When every phase is done: `forge advance <id> --to reviewing` and launch
`forge-reviewer`. If the review fails, go back to implementing with a note
saying why.

## Closing

```bash
forge archive <id>
git add -A && git commit -m "spec(spec-00X): archive"
forge submit <id>
```

`forge archive` keeps the spec folder in the tree as the record and refuses
if there are unresolved convention proposals, which is the point: decide
them first. `forge submit` pushes the branch and opens the pull request.
**Never push or merge into the default branch and never merge the pull
request yourself**; the guard denies those commands, and a person reviews
and merges on the forge host.
