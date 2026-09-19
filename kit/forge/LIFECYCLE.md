# Spec lifecycle

State machine. Independent of stack. Only the orchestrator writes
transitions in `record.md`.

Closest cousins: Spec Kit (Specify → Plan → Tasks → Implement →
Converge) and OpenSpec (proposal → design → tasks → apply → archive).
Forge keeps one folder per spec and a thin backlog.

```
draft → specifying → specified → planning → planned → implementing → reviewing → done
                       ↑                    │
                       └── rejected ────────┘
```

## States

| Status | Meaning | Who acts | Artifact |
|---|---|---|---|
| `draft` | Backlog item ready, no contract yet | architect | `backlog/BL-XXX.md` |
| `specifying` | Contract being written | architect | `specs/SPEC-XXX/spec.md` |
| `specified` | Contract ready, waiting on owner | owner | `approved` / `lgtm` / `dale` |
| `planning` | Split into verifiable phases | architect / orchestrator | `plan.md` |
| `planned` | Plan written | orchestrator | complete `plan.md` |
| `implementing` | Product code | implementer | `changes.md` |
| `reviewing` | Verification, no new features | qa | `qa-report.md` |
| `done` | Closed | orchestrator | history on `record.md` |
| `rejected` | Scope changed or gate failed | orchestrator | reason in history |

## Human gates

- `specifying → specified` needs explicit owner approval.
- `reviewing → done` needs a QA pass and zero P0s. If scope grows:
  `reviewing → implementing` with a reason, spec updated, new human gate.

## How to advance

In `specs/SPEC-XXX/record.md`:

1. Change `status:` in the frontmatter.
2. Append history; do not delete previous lines:

```markdown
<!-- history at="ISO-8601" from="specifying" to="specified" by="orchestrator" -->
Owner approved the spec in chat.
```

3. Update the row in [BOARD.md](BOARD.md).
4. When a BL is promoted: `status: promoted` and `spec_id` on the BL.

Skill: `advance-spec`.

## Files in each SPEC

```
forge/specs/SPEC-XXX/
  record.md
  spec.md          # from specifying
  plan.md          # from planning
  changes.md       # from implementing
  qa-report.md     # from reviewing
```

Templates: [templates/](templates/).

## Rules

- One BL, one SPEC. Exploding scope → new SPEC or an expansion with a
  new human gate.
- No product code without a SPEC in `planned` or `implementing`.
- Exceptions: the Forge kit itself, typos, or a small local change the
  user asked for explicitly.
- The implementer does not invent scope. If a decision is missing,
  pushback on `record.md` and stop.
- Decisions that outlive the SPEC go to `memory/decisions.md` and, if
  architectural, to `memory/adrs/`.
