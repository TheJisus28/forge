# .forge

Everything this project agreed on, in files you can read and review.

| Path | What it is |
|---|---|
| `project.md` | What this project is: stack, commands, language |
| `specs/` | One file per unit of work, in any state from proposed to done |
| `wip/` | Scaffolding of work in flight: plan, changes, review. Deleted when the spec is archived |
| `decisions/` | Why the system is the way it is. One file per decision |
| `conventions/` | How code is written here. One file per domain |
| `BOARD.md` | A generated view. Gitignored; run `forge board` |
| `kit/` | The machinery: the workflow and the agent roles |

## The one rule

**Everything outside `kit/` belongs to the team and Forge never overwrites
it. `kit/` belongs to Forge and `forge update` rewrites it whole.**

## Where state lives

The state of a spec is the `status` field in its frontmatter, and only
`forge` writes it. The board is a projection; if they disagree, the spec
wins and the board should be regenerated.

## Day to day

```bash
forge status                  # what is open and who is waiting
forge new "<title>"           # propose work
forge accept <id>             # into the queue
forge start <id>              # begin: checks dependencies first
forge approve <id>            # the contract is right; code can start
forge archive <id>            # close it, last commit of the pull request
```

The full state machine is in [kit/WORKFLOW.md](kit/WORKFLOW.md).
