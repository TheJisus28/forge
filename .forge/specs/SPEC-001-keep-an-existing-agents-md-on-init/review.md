# Review — SPEC-001

Verdict: pass

## Acceptance criteria

| Criterion | Result | Evidence |
|---|---|---|
| AC1: `forge init` keeps an existing `AGENTS.md`/`CLAUDE.md` | pass | `TestInit_KeepsExistingRootPointers` writes both files, runs `init`, asserts they are byte-identical. |
| AC2: `forge init --force` overwrites them | pass | The same test asserts the file changes after `init --force`. |
| AC3: `forge update` keeps an existing `AGENTS.md` | pass | The same test runs `update` and asserts the custom content survives. |
| AC4: `go test ./...` passes | pass | All packages `ok`; `go vet` and `gofmt` clean. |

## Blocking problems

None.

## Notes

- The fix removed two entries from the `kitOwned` switch; `plant`'s existing
  "keep unless --force" branch does the rest.
- Discovered while verifying, out of scope: `cmdInit` parses with `fs.Parse`,
  so a flag after the positional `[dir]` is ignored.

## Proposed conventions

None.
