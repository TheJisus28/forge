---
domain: cli-output
approved_by: TheJisus28
date: 2026-09-20
---

# Convention — CLI output

## Rule

A command reports a non-fatal notice as a line beginning `warning: ` written
to the command's `out` writer, and still exits 0. A failure is a returned
`error`: `Main` prefixes it with `forge: ` and exits 1. A user-supplied value
interpolated into any message is quoted with `%q`, never with manual quote
characters, so quoting and escaping come from one place.

## Warning that shares one text with a machine payload

When a command emits one text that also feeds a machine payload (`--json`)
or is injected into an agent session, the warning is prepended to that text
as `warning: ...\n` and is never written separately, so the payload is not
broken and the warning reaches the agent. In every other case the warning
goes to the command's `out` writer, as in the rule above.

`internal/cli/report.go:cmdBrief` is the case that motivated it: one
`text := view.Brief(p)` serves the human and the Claude Code `SessionStart`
hook (`forge brief --json`), so a `fetch: on` failure has to travel inside
the text rather than beside it:

```go
text := view.Brief(p)
if warning != "" {
	text = warning + "\n" + text
}
```

The `--json` payload is one line a hook parses: a warning printed before it
breaks the parse, and a warning printed after it is never seen. A bare
`warning: ` line on `out` is fine for a command that only prints for a
person, not when the same bytes are also a machine interface.

## Example

`internal/cli/work.go:cmdNew`, when no existing spec declares the requested
capability:

```go
fmt.Fprintf(out, "warning: no existing spec declares the capability %q; creating it as a new one\n", cap)
```

and `internal/validate/validate.go` renders the malformed value with
`capability %q is not a lowercase slug ([a-z0-9-]+)`, matching the existing
`unknown command %q` in `Main`.

## Why

`internal/validate` already separates `Warning` from `Error`; this is its
counterpart on the command line. Without it, a command that must report
something and still succeed would either exit non-zero and break CI, or
print a bare line indistinguishable from normal output.

## Exceptions

Progress echo that is the point of the command (`forge status`, `forge
new`) is normal output, not a warning. Findings of `forge validate` keep
their own `warning`/`error` prefixes and severity rules.
