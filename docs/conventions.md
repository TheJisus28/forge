# Conventions

`.forge/conventions/` is how your project writes code. Forge ships it
**empty**, and that is deliberate.

A tool that arrives with opinions about your framework is wrong about half
the projects it lands in, and it teaches agents to follow rules the team
never chose. What belongs here is what your team decided, in your words.

## How it fills up

Agents propose, maintainers decide.

When an implementer has to make a choice that nothing covers — the shape of
an error response, where a new module goes, how a migration is named — it
does the reasonable thing, and then says so in the
"Proposed conventions" section of `.forge/wip/<id>/changes.md`. The reviewer
does the same when it notices a pattern repeating.

The orchestrator brings it to a maintainer:

> The implementer noticed the three new services return errors with the same
> envelope, but that is not written anywhere. Should it become a convention?
> If you approve it, other agents will follow it without being told.

On approval it becomes `.forge/conventions/api.md`, with the maintainer's
handle and the date. On rejection, the code changes to match what the
project already does. What does not happen is leaving it half agreed.

## What a convention looks like

```markdown
---
domain: api
approved_by: jesus
date: 2026-03-14
---

# Convention — api

## Rule

Every error response is `{ "error": { "code", "message" } }`. `code` is a
stable string; `message` is for humans and may change.

## Example

Code from this repository, not from a tutorial.

## Why

Clients switch on `code`. Changing it breaks them silently.

## Exceptions

Health endpoints return plain text.
```

Write the rule so it can be followed without judgement calls. A convention
an agent has to interpret is a convention it will interpret differently
tomorrow.

## The rules agents follow

- Never apply a convention that is not written here.
- Never write a file here without an explicit approval.
- Never import rules from another project, another language, or a style
  guide the team did not choose.

## Why this pays for itself

Without it, every session re-derives your style by reading code, which costs
tokens and produces inconsistent results. With it, twenty lines settle it,
and the same decision is not re-litigated in every pull request.
