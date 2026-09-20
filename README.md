# Forge

[![ci](https://github.com/TheJisus28/forge/actions/workflows/ci.yml/badge.svg)](https://github.com/TheJisus28/forge/actions/workflows/ci.yml)
[![license](https://img.shields.io/badge/license-MIT-blue)](LICENSE)

**A spec-driven workflow for coding agents, in files your team owns.**

Forge is one Go binary with no dependencies, no account and no server. It
plants a `.forge/` directory in your repository and keeps the process honest:
ids, states, dependencies, coverage and validation are the tool's job, so the
agents spend their tokens on the work.

## Install

```bash
go install github.com/TheJisus28/forge@latest
cd your-project
forge init
```

You need Go 1.22+ and `git`. `gh` is optional, and only for opening pull
requests.

Then open your coding agent (Claude Code, opencode, or anything that reads
`AGENTS.md`) and say: **"run the Forge onboarding"**. The agent inspects the
repository, asks what it cannot infer, and fills `.forge/project.md`.

## Use it

Everything moves through one artifact: the **spec**.

```bash
forge new "Export invoices to CSV"   # propose work
forge accept SPEC-005                # into the queue
forge start SPEC-005                 # a branch, then the contract
forge approve SPEC-005               # the contract is right; code can start
forge advance SPEC-005 --to reviewing
forge archive SPEC-005               # close it; a review is required
forge submit SPEC-005                # push the branch, open the pull request
```

A person reviews and merges the pull request. Forge never merges for you.

Check where things stand at any time:

```bash
forge status     # every spec, what is waiting, what blocks
forge brief      # the short state an agent reads at session start
forge validate   # exit 1 when the project is inconsistent
```

## Why teams use it

- **A contract approved before any code.** The agent writes down what it will
  build and a human agrees; then it builds exactly that.
- **No re-litigating every session.** The stack, the commands, the conventions
  and the decisions that outlive a spec live in `.forge/`, in Markdown the
  agent already reads.
- **The tool owns the bookkeeping.** Ids, states, hierarchy, coverage,
  dependency cycles and contract drift are checked by `forge`, not by a tired
  reviewer.
- **A record that stays.** Every spec ends as a folder, `spec.md`, `plan.md`,
  `tasks.md` and `review.md`, so a later change knows what exists and how it
  was verified.
- **It enforces itself.** While no spec is `implementing`, the guard denies
  edits to product code; it also refuses pushes and merges into the default
  branch, so work ends as a pull request a person approves.
- **No opinions about your stack.** Forge works in a Rust repository and a
  Rails one. What belongs in `.forge/conventions/` is what your team decided.

## How the loop works

```
proposed → accepted → specifying → awaiting-approval → planning →
implementing → reviewing → done
```

The spec is born as a problem and a few acceptance criteria, and grows a
contract as it advances. Two moves are worth a human's attention: accepting
work into the queue, and approving the contract. Forge has no maintainer list
to check; anyone can do either, the way anyone with push access can commit.
The record says who did it, and the scrutiny happens in the pull request
review your team already does. [The workflow](docs/workflow.md) has the
detail.

## What `forge init` plants

```
your-project/
├── AGENTS.md  CLAUDE.md          pointers, a few lines each
├── .forge/
│   ├── project.md                stack and commands
│   ├── specs/                    one folder per spec: spec, plan, tasks, review
│   ├── decisions/                why the system is like this
│   ├── conventions/              how code is written here
│   └── kit/                      the workflow and the agent roles
├── .claude/                      subagents, skills and session hooks
├── .opencode/                    subagents and plugins
└── .github/workflows/            with --ci github
```

Everything outside `.forge/kit/` is yours; Forge never overwrites it.

## Works with your agent

`AGENTS.md` is the single source. For **Claude Code**, `forge init` installs a
`SessionStart` hook that runs `forge brief` (and again after every compaction)
and a `PreToolUse` hook that runs `forge guard`. For **opencode**, it plants
`.opencode/agents/` and two plugins that do the same: the brief in the system
prompt, the guard before edits and shell commands. Other agents read
`AGENTS.md` on their own. See [Using Forge with opencode](docs/opencode.md).

## Local-first

No account, no telemetry, no network. Forge reads and writes files; `git` and
`gh` run only when you ask. A test in this repository fails if the binary ever
imports `net/http`.

## Documentation

- [Workflow](docs/workflow.md) — states, hierarchy, coverage, dependencies
- [Teams](docs/teams.md) — branches, pull requests, CI
- [Conventions](docs/conventions.md) — how a project accumulates criteria
- [CLI reference](docs/cli.md)
- [opencode](docs/opencode.md) — installing and using Forge from opencode
- [Customizing](docs/customizing.md) — templates, roles, other agents

## Contributing

Issues and pull requests are welcome; see [CONTRIBUTING.md](CONTRIBUTING.md).
The kit is Markdown under `kit/`, and most improvements are changes there.

## License

MIT — see [LICENSE](LICENSE).
