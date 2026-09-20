# specs

One folder per unit of work, in any state. This folder is the backlog, the
board and the archive at once: what state a spec is in is written in its
frontmatter, not in which folder it sits.

- One folder per spec, created by `forge new`:
  `.forge/specs/SPEC-004-short-slug/`, with four files: `spec.md` (problem,
  criteria, contract, history), `plan.md` (approach and the existing state
  it builds on), `tasks.md` (the phases, ticked as they land) and
  `review.md` (evidence per criterion).
- The number is confirmed by `forge accept` against the ids already on
  `main`, and the spec's pull request opens early even if the spec is just a
  paragraph.
- A spec with children is what other tools call an epic. It is not built
  directly: its children are.

Do not create files here by hand and do not edit `status`: `forge` owns
ids, state and history.
