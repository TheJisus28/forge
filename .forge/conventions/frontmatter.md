---
domain: frontmatter
approved_by: TheJisus28
date: 2026-09-20
---

# Convention — frontmatter

## Rule

A finding about an intrinsic frontmatter field — an absent `capability`, a
malformed `title` — is produced before the per-spec loop's unknown-status
`continue`, beside the missing-title check, because it does not depend on the
state the spec claims. A field's shape is stated as a trailing `# ` comment
on its line in the template, which `internal/doc` strips before parsing, so
`forge template <name>` shows the guidance while the value stays empty. A
list of spec ids read from frontmatter is normalised through `normalizeIDs`
(the id counterpart of `upperAll` for criteria) and written with
`setListOrDelete`, so forgiving input and canonical output stay in one place
and an empty typed slice removes the key rather than writing `[]`.

## Example

`internal/validate/validate.go:Run` calls `checkCapability` beside the
missing-title check, above `if !workflow.Valid(s.Status) { continue }`.
`kit/machine/templates/spec.md` carries:

```yaml
capability: ""  # lowercase slug ([a-z0-9-]+), which part of the system this spec touches
```

## Why

A field the CLI reads is part of the contract, not of the state: a spec with
an unrecognised status still has a `capability` that must be checked. Putting
the guidance on the template line means the rule travels with the file a
human copies and with `forge template`, without a second document to drift.

## Exceptions

A field whose validity genuinely depends on the state (for example an
artifact that must exist only while `implementing`) stays in
`checkArtifacts`, after the state gate.
