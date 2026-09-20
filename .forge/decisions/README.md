# decisions

Why the system is the way it is. One file per decision, numbered:
`0007-idempotent-charges.md`, from `kit/templates/decision.md`.

One file per decision on purpose: a shared log would make every branch
append to the same lines and conflict on every merge.

Write one when a choice will outlive the spec that forced it: a contract
another team depends on, a storage model, a boundary, a dependency someone
would otherwise question again in six months.

An accepted decision is not rewritten. Changing your mind means a new file
that supersedes the old one, and saying so in both.
