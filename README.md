# Forge

[![ci](https://github.com/TheJisus28/forge/actions/workflows/ci.yml/badge.svg)](https://github.com/TheJisus28/forge/actions/workflows/ci.yml)
[![license](https://img.shields.io/badge/license-MIT-blue)](LICENSE)

**A spec-driven workflow for coding agents, in files your team owns.**

One Go binary, no dependencies, no account, no server. It plants a `.forge/`
directory in your repository and then keeps the process honest: ids, states,
dependencies, coverage and validation are the tool's job, so the agents can
spend their tokens on the work.

```bash
go install github.com/TheJisus28/forge@latest
cd your-project
forge init
```

Then open your agent and say: **"run the Forge onboarding"**.

## Why

Agents are good at writing code and bad at remembering what your team
decided. Every session rediscovers the stack, re-infers the conventions, and
occasionally builds the wrong thing very well.

Forge writes that down once, in Markdown the agents already read, and adds
the one thing a chat cannot give you: **a contract approved by a human
before any code exists**.

It ships with **no opinions about your technology**. There is no list of
frameworks to prefer and no rules about how to name things. What belongs in
`.forge/conventions/` is what your team decided, and agents propose new ones
rather than inventing them.

## The loop

```
proposed → accepted → specifying → awaiting-approval → planning →
implementing → reviewing → done
```

One artifact moves through it: the **spec**. It is born as fifteen lines
describing a problem and grows a contract as it advances. There is no
separate backlog item and no separate epic; the backlog is the specs nobody
started, and an epic is a spec that has children.

Two moves are worth a human's attention: accepting work into the queue,
and approving a contract before code exists. Forge has no maintainer list
to check, though — anyone can do either, the same way anyone with push
access can commit. The record says who did it; the scrutiny happens where
your team already reviews, in the pull request.

## What `forge init` plants

```
your-project/
├── AGENTS.md  CLAUDE.md          pointers, three lines each
├── .forge/
│   ├── project.md                stack and commands
│   ├── specs/                    one folder per spec: spec, plan, tasks, review
│   ├── decisions/                why the system is like this
│   ├── conventions/              how code is written here
│   └── kit/                      the workflow and the agent roles
├── .claude/                      subagents, skills and session hooks
├── .opencode/                    subagents and the guard plugin
└── .github/workflows/            with --ci github
```

One rule: everything outside `.forge/kit/` is yours and Forge never
overwrites it.

## Agents that start knowing where the project is

For Claude Code, `forge init` installs a `SessionStart` hook that runs
`forge brief`. Before you type anything, the agent already knows which spec
your branch is about, what it is waiting for, what is blocked and what
changed. It runs again after every compaction, so long sessions do not drift.

A `PreToolUse` hook runs `forge guard`, which **denies edits to product code
while no spec is in `implementing`**, and says how to unblock. That is the
difference between a workflow people respect and one they respect when they
are not in a hurry. Turn it off with `guard: off` in `project.md`.

For opencode, `forge init` plants `.opencode/agents/` and a guard plugin:
opencode reads `AGENTS.md` and Forge's skills natively, and the plugin calls
`forge guard` before every edit, so the same rule holds without a hook.
[Using Forge with opencode](docs/opencode.md) has the details. Other agents
read `AGENTS.md`, which is the single source the wrappers point at.

## Built for teams

- The number of a spec is reserved by a one-file intake pull request.
- Work in flight lives in its own branch, so two people never touch the
  same file; `forge archive` marks the spec done and keeps its folder, and
  the main branch accumulates contracts, plans, tasks, reviews, decisions
  and conventions.
- `depends_on: [SPEC-011@contract]` unblocks a front end as soon as the back
  end's contract is approved, without waiting for its code. If that contract
  later changes, `forge validate` fails and names who was building against
  the old one.
- No approval list to manage: `forge accept` and `forge approve` work for
  anyone, and CI still catches what is objectively broken.

## Local-first

No account, no telemetry, no network calls. Forge reads and writes files;
`git` and `gh` are invoked explicitly when you ask for them. A test in this
repository fails if the binary ever imports `net/http`.

## Commands

| | |
|---|---|
| `forge init` / `update` | plant or refresh the kit |
| `forge new "<title>"` | propose work |
| `forge accept` / `approve` | into the queue, contract is right |
| `forge start` / `advance` / `archive` | move the work |
| `forge status` / `brief` | what is happening |
| `forge validate` | the CI check |
| `forge guard` / `sync` | hooks and GitHub integration |

Full reference: [docs/cli.md](docs/cli.md).

## Documentation

- [Workflow](docs/workflow.md) — states, hierarchy, coverage, dependencies
- [Teams](docs/teams.md) — branches, pull requests, gates, CI
- [Conventions](docs/conventions.md) — how a project accumulates criteria
- [CLI reference](docs/cli.md)
- [opencode](docs/opencode.md) — installing and using Forge from opencode
- [Customizing](docs/customizing.md) — templates, roles, other agents

## Contributing

Issues and pull requests are welcome; see [CONTRIBUTING.md](CONTRIBUTING.md).
The kit is Markdown under `kit/`, and most improvements are changes there.

## License

MIT — see [LICENSE](LICENSE).
