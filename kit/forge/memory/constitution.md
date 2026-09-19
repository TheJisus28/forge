# Constitution

Project principles. Analogous to Spec Kit's constitution: stable rules
agents must not silently break. Edit this file; do not fork the kit.

```yaml
working_language: en   # en | es | …  specs and user-facing copy
code_comments: en      # language of comments, logs, API docs
```

## Principles

1. Follow `forge/memory/stack.md`. Do not switch runtime or language
   without a SPEC.
2. Do not invent backlog. The user asks for work.
3. Specs before product code (see `forge/LIFECYCLE.md`).
4. Kit and agent instructions stay in English even if
   `working_language` is not `en`.

## Style

- Commits: Conventional Commits, English summaries unless this file
  says otherwise (`type(scope): summary`).
- User-visible product copy may use `working_language`.
