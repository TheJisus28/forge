---
name: forge-onboard
description: Fill .forge/project.md by inspecting the repository and interviewing the owner. Use when project.md is empty or unanswered, when Forge was just installed, or when the user says the project is not onboarded.
---

# Onboarding

Forge ships no assumptions about this project. Your job is to replace that
emptiness with confirmed facts, not with guesses.

## 1. Look

Read the repository: manifests and lockfiles, the test and CI
configuration, the directory layout, the entry points, recent commits.
Build a proposal of what the stack is and how it is run.

## 2. Ask

Show the proposal and ask about everything you could not settle. At
minimum:

- The exact test command, and the dev command.
- Anything you saw two ways of doing, and which one is the current one.
- Which parts are legacy and must not be touched.
- The working language for specs and product copy.

Ask in one batch, plainly, and wait. Never fill a field with a guess.
Forge has no maintainer or approval list to configure: anyone on the team
can accept work and approve a contract, so there is nothing to ask about
who is allowed to do what.

## 3. Write

Write `.forge/project.md`: the frontmatter (`test`, `dev`,
`working_language`) and the prose sections, using only confirmed answers.
Leave a question in place if it was not answered.

## 4. Offer conventions

If, while looking, you saw patterns the project clearly follows, list them
and ask whether any should become a convention. Write
`.forge/conventions/<domain>.md` only for the ones the team agrees on.

Finish by running `forge status` and telling the user what to do next.
