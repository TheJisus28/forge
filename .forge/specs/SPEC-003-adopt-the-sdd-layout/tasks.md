# Tasks — SPEC-003

- [x] Phase 1 — `project.Load` reads `specs/<id>/spec.md` and exposes `Dir`,
  `PlanPath`, `TasksPath`, `ReviewPath`, `SpecDir`, `SpecDirName`. Where:
  `internal/project/project.go`.
- [x] Phase 2 — `new` creates the folder, `archive` keeps it, `renumber`
  moves it. Where: `internal/cli/work.go`.
- [x] Phase 3 — `validate` requires plan and tasks while implementing,
  review while reviewing and when done. Where: `internal/validate/validate.go`.
- [x] Phase 4 — kit templates, roles, skill, wrappers and docs. Where:
  `kit/`, `docs/`, `AGENTS.md`, `README.md`.
- [x] Phase 5 — migrate `.forge/` and write this record. Where:
  `.forge/specs/SPEC-001-.../`, `.forge/specs/SPEC-002-.../`, this folder.

## Proposed conventions

None.
