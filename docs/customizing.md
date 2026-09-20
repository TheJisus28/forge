# Customizing

Forge's machinery — the workflow, the roles and the templates — ships in
the binary. What you change lives in your own repository: the conventions,
the decisions and the project facts.

## Templates

The file templates are a standard: they ship in the binary and `forge new`
always uses that copy. Forge does not read a template from your repository,
so a project cannot override the shape of a spec, a plan, a decision or a
convention. To read one:

```bash
forge template spec
```

If your team needs a section every spec must carry — a threat model, a
rollout plan, metrics — that is a change to Forge itself, not a local edit.

## Agent roles

The orchestrator, architect, implementer and reviewer roles ship in the
binary. `forge roles <name>` prints one; `forge roles` lists them. They are
deliberately short and free of technology opinions, and a project cannot
edit them.

`forge init` writes each role in full into its host adapter
(`.claude/agents/forge-*.md`, `.opencode/agents/forge-*.md`), so a generated
agent is self-contained; `forge update` rewrites it from the binary. Change
the host frontmatter there — the `model`, the `tools`, the `permission` —
not the role text.

## The guard

`guard: off` in `.forge/project.md` disables the `PreToolUse` denial without
touching the rest. `forge guard --explain` tells you what it would do on the
current branch, which is the fastest way to debug a surprise.

Product code is anything outside `.forge/`, `.claude/`, `.github/` and the
root pointers. If your repository keeps documentation somewhere the guard
should ignore, the cleanest answer today is to work on a spec branch; a
configurable allowlist is a reasonable feature request.

## Working language

`working_language` in `.forge/project.md` applies to specs, decisions and
product copy. The process files stay English so the same kit works across
teams; the parser accepts both `## Acceptance criteria` and
`## Criterios de aceptación` for the sections it needs to read.

## Other agents

`AGENTS.md` is the single source of truth: anything that reads it needs no
wrapper at all.

- **opencode**: first-class support. `forge init` plants
  `.opencode/agents/` (the architect, implementer and reviewer) and
  `.opencode/plugins/forge-guard.js` (the guard). opencode reads
  `AGENTS.md` and Forge's `.claude/skills/` on its own. See
  [Using Forge with opencode](opencode.md).
- Codex and others reading [agents.md](https://agents.md/): nothing to do.
- Cursor: `.cursor/rules/forge.mdc` and `.cursor/skills/<name>/SKILL.md`
- GitHub Copilot: `.github/copilot-instructions.md`

You can write those by hand now; first-class support is next, and
contributions are welcome. For any host, `forge workflow` prints the process
and `forge roles <name>` prints the role to follow, so nothing has to be
copied out of the binary.

For a tool that runs a command before file edits, wire it to
`forge guard --file <path>`: it exits 1 and prints the reason when the edit
must be denied, so the rule stays in the binary instead of being copied.

## Making the CLI part of your own workflow

`forge validate` is the only command you need in CI. Everything else is
optional. If your team does not use GitHub, skip `--ci github` entirely:
Forge does not depend on it. `forge accept` and `forge approve` work the
same locally as anywhere; there is no server-side gate to miss.
