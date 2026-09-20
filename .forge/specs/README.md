# specs

One file per unit of work, in any state. This folder is the backlog, the
board and the archive at once: what state a spec is in is written in its
frontmatter, not in which folder it sits.

- Name: `SPEC-004-short-slug.md`, created by `forge new`.
- The number is reserved when the spec is merged into the main branch, so
  open the intake pull request early even if the spec is just a paragraph.
- A spec with children is what other tools call an epic. It is not built
  directly: its children are.

Do not create files here by hand and do not edit `status`: `forge` owns
ids, state and history.
