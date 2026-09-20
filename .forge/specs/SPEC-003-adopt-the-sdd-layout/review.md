# Review — SPEC-003

Verdict: pass

## Acceptance criteria

| Criterion | Result | Evidence |
|---|---|---|
| AC1: a spec is `specs/<id>/` with `spec.md`, `plan.md`, `tasks.md`, `review.md`, created by `forge new` | pass | `cmdNew` writes `specDir/spec.md`; `TestLifecycle` runs `forge new` then reads `specs/SPEC-001-saved-card-payments/spec.md`. |
| AC2: `forge archive` keeps the folder | pass | `cmdArchive` no longer calls `RemoveAll`; `TestLifecycle` asserts the folder's `review.md` survives archive. |
| AC3: `.forge/wip/` and the `changes.md` template are gone | pass | `kit/forge/wip/` and `kit/forge/kit/templates/changes.md` deleted; `.forge/wip/` removed; `forge update` planted `tasks.md`. |
| AC4: validate/status/brief/guard work on folders and tests pass | pass | `go test ./...` ok; `go vet` and `gofmt` clean; `forge validate`, `forge status`, `forge brief`, `forge guard` run against the migrated tree. |

## Blocking problems

None.

## Notes

- No backward compatibility: the old flat layout is not read. This repository
  was migrated by hand.
- The spec was implemented directly, not through the state machine, because
  the layout it changes is the one the workflow itself runs on.

## Proposed conventions

None.
