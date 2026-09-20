# Plan — SPEC-003

## Existing state

- The flat layout: `project.Load` read `specs/*.md`; `WipDirFor` held
  `plan.md`, `changes.md` and `review.md` under `.forge/wip/<id>/`; `archive`
  deleted that folder. `FileName` built `SPEC-NNN-slug.md`.
- Kit templates `spec.md`, `plan.md`, `changes.md`, `review.md`; the roles,
  the `forge-work` skill and the OpenCode/Claude wrappers pointed at `wip/`.
- Delivered specs: SPEC-001 (`forge init` keeps an existing `AGENTS.md`) and
  SPEC-002 (`forge submit` and the `gh` actor).

## Approach

- Change `internal/project` to read `specs/<id>/spec.md` and to expose the
  spec folder paths; make `new` create the folder and `archive` keep it.
- Require `plan.md` and `tasks.md` while implementing, `review.md` while
  reviewing and when done; drop the wip-exists check.
- Replace `changes.md` with `tasks.md` in the kit and point the roles and
  wrappers at the folder; remove `kit/forge/wip/`.
- Migrate this repository by hand; no backward compatibility.

## Risks

- Old repositories are not read. Intended: there is no beta to support yet.
- `.forge/kit/` and the planted `.claude/`, `.opencode/` copies are refreshed
  by `forge update`; removed files (`changes.md`) are deleted by hand.
