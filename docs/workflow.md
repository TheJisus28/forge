# The workflow

Forge has one artifact: the **spec**. It is born cheap and grows.

```
proposed → accepted → specifying → awaiting-approval → planning →
implementing → reviewing → done

any of them → dropped          implementing ↔ blocked
```

The backlog is not a folder: it is the specs in `proposed` and `accepted`.
An epic is not a type: it is a spec that has children. Collapsing those into
one artifact removes the copying, and the drift that copying causes.

## States

| State | What it means | Who moves it |
|---|---|---|
| `proposed` | Written, not in the queue | maintainer accepts or drops |
| `accepted` | In the queue | anyone, with `forge start` |
| `specifying` | The contract is being written | architect |
| `awaiting-approval` | Contract ready for a human | maintainer |
| `planning` | Being split into phases | orchestrator |
| `implementing` | Product code, phase by phase | implementer |
| `blocked` | Cannot continue | conductor |
| `reviewing` | Criteria being verified | reviewer |
| `done` | Archived | nobody |
| `dropped` | Will not be done | nobody |

Only `forge` writes `status`, and every move appends a line to the spec's
History section saying who did it and why.

## The two human gates

```bash
forge accept SPEC-004 --by jesus      # into the queue
forge approve SPEC-004 --by jesus     # the contract is right
```

Both require a handle listed in `maintainers` in `.forge/project.md`. With
more than one maintainer, nobody approves the contract they conducted
themselves; set `allow_self_approval: true` if your team wants otherwise.

In a repository with CI, approving the pull request is the gate: the
workflow runs `forge gate --by <approver>` and records it for you.

## Acceptance criteria

Criteria live in the spec as a list:

```markdown
## Acceptance criteria

- AC1: a saved card can be reused without retyping it
- AC2: charging twice with the same request id charges once
```

They must be verifiable by someone who did not write them: a command, a
test, a request and its response. The reviewer marks each one with evidence.

## Hierarchy and coverage

A spec that spans several deliverables becomes a parent:

```bash
forge new "Card payments"                                  # SPEC-002
forge new "Saved cards" --parent SPEC-002 --covers AC1,AC3 # SPEC-004
forge new "Reconciliation" --parent SPEC-002 --covers AC4  # SPEC-005
```

Coverage is declared by the child and computed by the tool. The parent file
is never edited when a child advances, which is why two branches closing
different children never conflict.

```
$ forge status SPEC-002
coverage
  AC1   SPEC-004 (implementing)
  AC3   SPEC-004 (implementing)
  AC4   SPEC-005 (accepted)
  AC2   NOT COVERED
```

`NOT COVERED` is a warning while everything is open and an error once any
child is closed: that is the moment where a promise silently disappears.

## Dependencies

```yaml
depends_on: [SPEC-003, SPEC-011@contract]
needs:
  - "a retry engine"
blocked_by_external:
  - what: "production credentials"
    who: ana
```

- **`depends_on: [SPEC-003]`** waits for that spec to be `done`.
- **`depends_on: [SPEC-011@contract]`** waits only for its contract to be
  approved. This is how a front end and a back end are built in parallel
  against the same agreed shape.
- **`needs`** is a dependency on something that has no spec yet. The only
  way out is to create it and replace the text with its id, which is what
  stops "we will look at it later" from meaning "never".
- **`blocked_by_external`** is out of Forge's hands, so it makes it visible
  with an owner.

`forge start` refuses while any of these is open, lists what is ready
instead, and accepts `--force` when you decide otherwise. The exception is
written into the spec, where the reviewer will see it.

## Contract drift

When `forge approve` runs, it stores a fingerprint of the contract. Every
spec that starts against it with `@contract` records which fingerprint it
agreed to.

If the contract changes afterwards, `forge validate` fails and names the
specs building against a version that no longer exists. It also fails if a
contract is edited after approval without being approved again. This is the
classic back-and-front integration failure, caught in the pull request
instead of in staging.

## Files

```
.forge/specs/SPEC-004-saved-cards.md    the spec, in every state
.forge/wip/SPEC-004/plan.md             phases
.forge/wip/SPEC-004/changes.md          what each phase did
.forge/wip/SPEC-004/review.md           evidence per criterion
```

`forge archive` deletes `wip/` in the last commit of the pull request. The
scaffolding stays in the branch history; the main branch keeps the contract,
the decisions and the conventions.
