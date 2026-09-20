# Using Forge with opencode

Forge works with [opencode](https://opencode.ai) out of the box. `forge init`
plants the same process it plants for Claude Code, plus the two things
opencode needs that Claude Code has no equivalent for: the subagents and the
guard plugin.

## Install

1. Install Forge and opencode:

   ```bash
   go install github.com/TheJisus28/forge@latest
   # opencode: https://opencode.ai/docs
   ```

2. Plant the kit in your repository:

   ```bash
   cd your-project
   forge init
   ```

3. Open opencode in that directory and say: **"run the Forge onboarding"**.

`forge` must be on your `PATH`: the guard plugin runs it before edits, the
same way the Claude Code hook does.

## What opencode reads

| Forge plants | opencode mechanism |
|---|---|
| `AGENTS.md` | project rules, read natively every session |
| `.claude/skills/forge-*/SKILL.md` | skills, through opencode's Claude Code compatibility |
| `.opencode/agents/forge-*.md` | the architect, implementer and reviewer subagents |
| `.opencode/plugins/forge-guard.js` | the guard, as a `tool.execute.before` hook |

`CLAUDE.md` is planted too, but opencode prefers `AGENTS.md`, so it is only
used by Claude Code.

## The guard

Claude Code runs `forge guard` from a `PreToolUse` hook. opencode has no
hooks, so `.opencode/plugins/forge-guard.js` calls the same binary before
opencode writes or patches a file:

```bash
forge guard --file <path>   # exits 1 and prints why when the edit is denied
```

The rule is unchanged: no product code while no spec is `implementing`.
Editing `.forge/`, `.opencode/`, `.claude/`, `.github/` and the root pointers
is always allowed, because that is where the process happens.

Turn it off for a project with `guard: off` in `.forge/project.md`, or from
the start with `forge init --no-guard`. To see what it would decide:

```bash
forge guard --explain --file src/whatever.ts
```

## Verify it works

```bash
forge status            # the agent should have run the onboarding first
forge guard --explain   # says what the guard would do on this branch
```

Then, in opencode, ask the agent to edit a product file while no spec is
`implementing`: it should refuse and point you at `forge new` or `forge
start`. Open a spec, accept and start it, and the same edit is allowed.

## Updating

`forge update` refreshes `.opencode/` together with the rest of the kit. It
never touches your `specs/`, `decisions/`, `conventions/` or `project.md`.

## Notes

- opencode loads its configuration at startup. After `forge init` or `forge
  update`, restart opencode so the new subagents and plugin are picked up.
- Forge does not touch your `opencode.json`. The integration relies on the
  auto-discovered `.opencode/` directories and on `AGENTS.md`, so no config
  file is required.
- If you set `OPENCODE_DISABLE_CLAUDE_CODE_SKILLS=1`, opencode stops reading
  `.claude/skills/` and the Forge skills will not appear. Remove the
  variable, or copy the skills under `.opencode/skills/`.
