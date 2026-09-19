---
name: forge-convention
description: Propose or record a project convention in .forge/conventions/. Use when you notice a repeated pattern that is not written down, when someone asks how something should be written here, or when a maintainer approves a proposed convention.
---

# Conventions

`.forge/conventions/` is how this project writes code. It is empty until
the team fills it, and only a maintainer decides what goes in.

## Proposing

When you had to decide something that was not written, or you see the same
pattern three times, propose it:

1. Name the domain: `api`, `database`, `naming`, `testing`.
2. State the rule so it can be followed without judgement.
3. Show an example taken from **this** repository.
4. Explain in one sentence why, and where it does not apply.

If you are running as implementer or reviewer, put this in the
"Proposed conventions" section of your report and stop there.

## Recording

Only after a maintainer says yes, and only then:

- Copy `.forge/kit/templates/convention.md` to
  `.forge/conventions/<domain>.md`, or add the rule to the existing file.
- Fill `approved_by` with the maintainer's handle and the date.

If the answer is no, the code changes to match what the project already
does. Do not leave the rule half agreed.

## Rules

- Never apply a convention that is not written in `.forge/conventions/`.
- Never write a file there without an explicit approval.
- Never import rules from another project, another language or a style
  guide the team did not choose.
