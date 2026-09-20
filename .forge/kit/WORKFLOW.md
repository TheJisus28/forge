# Workflow

One artifact: the **spec**. It is born cheap, grows a contract, and ends as
the record of what was promised and delivered. There is no separate backlog
item and no separate epic: the backlog is the specs in `proposed` or
`accepted`, and an epic is a spec that has children.

## States

| State | Meaning | Who acts next |
|---|---|---|
| `proposed` | Written, not in the queue | anyone: `forge accept` or drop |
| `accepted` | In the queue | anyone: `forge start` |
| `specifying` | The contract is being written | architect |
| `awaiting-approval` | Contract ready | anyone: `forge approve` |
| `planning` | Splitting into phases | orchestrator |
| `implementing` | Product code, phase by phase | implementer |
| `blocked` | Cannot continue | conductor |
| `reviewing` | Verifying acceptance criteria | reviewer |
| `done` | Archived and closed | nobody |
| `dropped` | Will not be done | nobody |

Forge has no authorization model. Anyone can accept work into the queue
and anyone can approve a contract, the same way anyone with push access can
commit; `--by` defaults to `git config user.name`, and the move is recorded
in the spec's history either way. There is no restricted role to ask
permission from, because a real team already reviews pull requests, and
that review is where scrutiny belongs — not in a second, Forge-specific
gate. `forge validate` still catches what is objectively broken: missing
contracts, uncovered promises, dependency cycles, contract drift.

## Hierarchy and coverage

A spec can declare `parent`. The child also declares `covers: [AC1, AC3]`:
which of the parent's acceptance criteria it delivers.

Coverage is declared from below and computed. The parent file is never
edited when a child advances, so two branches closing different children
never conflict. A parent cannot be closed while a criterion is uncovered.

## Dependencies

```yaml
depends_on: [SPEC-003, SPEC-011@contract]
needs: ["a retry engine"]        # no spec exists for this yet
blocked_by_external:
  - what: "production credentials"
    who: ana
```

`@contract` unblocks as soon as the other spec's contract is approved,
without waiting for its code. That is how a front end and a back end are
built in parallel against the same agreed shape.

`forge start` refuses while any of these is open. `--force` starts anyway
and records the exception in the spec, where the reviewer will see it.

When you start against someone's contract, Forge stores its fingerprint.
If that contract changes later, `forge validate` fails and names your spec,
because you are building against a version that no longer exists.

## Files

```
.forge/specs/SPEC-004-slug.md       the spec: problem, criteria, contract, history
.forge/wip/SPEC-004/plan.md         phases
.forge/wip/SPEC-004/changes.md      what each phase did
.forge/wip/SPEC-004/review.md       evidence per criterion
```

`forge archive` deletes `wip/` and leaves the spec. Nothing is lost: the
scaffolding stays in git history.

## Rules

- No product code without a spec in `implementing`.
- A spec ends as a pull request: `forge submit <id>` pushes the branch and
  opens it. Nothing in Forge merges it; a person merges on the forge host.
- Subagents never change `status`; the orchestrator runs `forge advance`.
- A conventions file is written only after the team agrees on it, not
  invented on the spot by whoever is implementing.
- Decisions that outlive the spec go to `.forge/decisions/`.
- If the contract is wrong, push back and stop. Do not improvise scope.
