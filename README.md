# Forge

A tiny **Go** CLI that plants a spec-driven agent kit into any git repository.

One binary, Markdown on disk, no runtime added to your project. Works with
the agents you already use: Codex, Cursor, Claude Code, Gemini CLI, Copilot.

```bash
go install github.com/TheJisus28/forge@latest
cd my-project
forge init
```

## Why

Every chat with a coding agent rediscovers the same things: what the stack is,
what was already decided, what is allowed. Forge writes that down once, in
files the agents already read.

The workflow is the usual spec-driven loop, kept deliberately thin:

**backlog → spec → plan → implement → review**

Same idea as [GitHub Spec Kit](https://github.com/github/spec-kit) and
[OpenSpec](https://openspec.dev/), without a heavy toolchain: no Python, no
node_modules, no 30 integrations to configure.

## Install

With Go 1.22+:

```bash
go install github.com/TheJisus28/forge@latest
```

Without Go: download a binary from
[Releases](https://github.com/TheJisus28/forge/releases) and put it on your
`PATH`.

From source:

```bash
git clone https://github.com/TheJisus28/forge.git
cd forge
go install .
```

## Usage

```bash
forge init [dir]     # copy the kit and write forge/memory/stack.md
forge detect [dir]   # print the stack.md that would be generated
forge version
```

| Flag for `init` | Effect |
|---|---|
| `--force` | rewrite the kit when `forge/README.md` already exists |
| `--reset-memory` | also overwrite `stack.md`, `decisions.md`, `constitution.md` |
| `--skip-copilot` | skip `.github/copilot-instructions.md` |

Without `--force`, an initialized repository is left untouched. With
`--force`, the kit is refreshed and your project memory is preserved unless
you pass `--reset-memory`. Re-running `init` after an upgrade is the intended
way to update the kit.

## What `init` writes

```
your-repo/
  AGENTS.md                  # canonical agent instructions, kept short
  CLAUDE.md                  # @AGENTS.md
  GEMINI.md
  forge/
    README.md                # how the kit works
    LIFECYCLE.md             # spec state machine
    BOARD.md                 # backlog + specs table (starts empty)
    agents/                  # orchestrator, architect, implementer, qa
      playbooks/               # java, go, node, python
    templates/               # backlog, spec, plan, changes, qa-report, adr
    backlog/  specs/         # your work, versioned in git
    memory/
      stack.md               # detected languages, runtime, test/dev commands
      constitution.md        # principles + working_language
      decisions.md           # decision log
      adrs/                  # architecture decision records
  .cursor/rules/forge.mdc
  .cursor/skills/{spec-workflow,advance-spec}/SKILL.md
  .claude/skills/{spec-workflow,advance-spec}/SKILL.md
  .github/copilot-instructions.md
```

Nothing else is touched. Forge never edits your source code.

## How agents pick it up

| Agent | File it reads |
|---|---|
| Codex, Cursor, Amp, Jules, Factory | `AGENTS.md` — the [agents.md](https://agents.md/) convention |
| Claude Code | `CLAUDE.md` importing `@AGENTS.md`, plus `.claude/skills/` |
| Gemini CLI | `GEMINI.md` (or point it at `AGENTS.md`) |
| GitHub Copilot | `.github/copilot-instructions.md` |
| Cursor | `.cursor/rules/`, `.cursor/skills/` |

`AGENTS.md` is the single source of truth; the rest are thin pointers, so you
never maintain the same instructions twice.

## Stack detection

`forge init` reads the manifests in the repository root and writes
`forge/memory/stack.md`:

| Detected | From |
|---|---|
| `go`, runtime (echo / chi / gin) | `go.mod` |
| `node`, `typescript`, package manager, runtime (next / express / fastify) | `package.json`, lockfiles, `tsconfig.json` |
| `java`, runtime (spring-boot), gradle or maven | `build.gradle`, `build.gradle.kts`, `pom.xml` |
| `python`, runtime (fastapi / django / flask), uv or poetry or pip | `pyproject.toml`, `requirements.txt`, lockfiles |
| `rust`, `dart` (noted, no playbook yet) | `Cargo.toml`, `pubspec.yaml` |

Detection is a starting point, not an oracle. Check the YAML and fill in the
`Facts` section by hand; agents trust that file before they trust habit.

Preview it without writing anything:

```bash
forge detect
```

## Language

The kit ships in **English** so it works for any team. Each project sets
`working_language` in `forge/memory/constitution.md`, so specs, ADRs, and
user-facing copy can be written in Spanish (or any other language) while the
agent instructions stay English.

## This repository

```
kit/            the Markdown that init plants — agents, memory, playbooks,
                templates, lifecycle. Edit here; it is embedded in the binary.
internal/detect stack detection from manifests
internal/stackmd renders memory/stack.md
internal/initcmd copies the kit, preserves project memory
main.go         command parsing
```

Everything the agents ever read lives under `kit/`. The Go code only
copies it, so a change to the methodology is a Markdown change.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

MIT — see [LICENSE](LICENSE).
