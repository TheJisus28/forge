# Customizing

Forge is Markdown plus a small binary. Almost everything you might want to
change is a file in your own repository.

## Templates

`forge new` reads `.forge/kit/templates/spec.md` from **your** project
before falling back to the embedded copy. Add sections your team always
wants — threat model, rollout plan, metrics — and every new spec will have
them.

The same folder holds the templates for plans, tasks, reviews, decisions
and conventions. Those are copied by hand or by the agent, so editing them
is enough.

Note that `forge update` rewrites `.forge/kit/`. If you customize templates,
either re-apply your changes after upgrading or keep them in a file the kit
does not own.

## Agent roles

`.forge/kit/agents/*.md` are the instructions for the orchestrator,
architect, implementer and reviewer. They are deliberately short and free of
technology opinions. Anything you add there applies to every session.

The Claude Code wrappers in `.claude/agents/` are five lines each: a
`description` that decides when Claude delegates, a `model`, and a pointer
to the role. Change the model there if you want the architect on a different
one, or add `tools` to narrow what a role can do.

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
contributions are welcome.

For a tool that runs a command before file edits, wire it to
`forge guard --file <path>`: it exits 1 and prints the reason when the edit
must be denied, so the rule stays in the binary instead of being copied.

## Making the CLI part of your own workflow

`forge validate` is the only command you need in CI. Everything else is
optional. If your team does not use GitHub, skip `--ci github` entirely:
Forge does not depend on it. `forge accept` and `forge approve` work the
same locally as anywhere; there is no server-side gate to miss.
