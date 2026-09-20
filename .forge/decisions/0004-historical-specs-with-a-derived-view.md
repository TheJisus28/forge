---
id: 0004
date: "2026-09-20"
status: accepted
spec: SPEC-010
---

# The compatibility surface is historical specs plus a derived view

## Context

Forge is approaching 1.0, so its public surface has to be frozen: the CLI
(command and flag names, exit codes), the `.forge/` frontmatter and folder
layout, the planted host paths and `{{forge-role:...}}` markers, and the
`vX.Y.Z` tag `forge upgrade` resolves. Freezing a surface and moving it
afterwards is worse than not freezing, because it breaks every installed
repository at the moment it promised stability.

The unresolved question is the **artifact model**. Today a spec is one
change: it is born `proposed`, grows a contract, ends `done`, and its folder
stays as the record. That answers "what changed and why". It does not answer
"what does the system do now": to reconstruct current behaviour you read the
union of every `done` contract, in order, and know which one supersedes
which. `forge brief` listing the five most recent `done` specs is a
mitigation, not an answer. At forty specs it is archaeology.

The alternative, used by OpenSpec, is a **living spec**: a consolidated,
hand-curated statement of current behaviour, kept current by applying each
change's deltas (ADDED/MODIFIED/REMOVED requirements) when the change is
archived. It answers the second question directly. It also reintroduces
exactly what Forge's layout was built to avoid — a shared file that two
parallel changes want to edit — plus a delta grammar and an archive-time
merge step whose semantics would themselves be frozen at 1.0.

This decision settles the model, because it dictates the layout, the
frontmatter and what `forge archive` means. Everything else in the
compatibility freeze depends on it.

## Decision

The source of truth stays **historical**: each spec is one change, and its
contract is never rewritten after approval. What the system currently does is
not a hand-maintained document but a **view derived** from the contracts of
specs in `done`, computed on demand.

Three parts:

1. **Historical specs are the input; the current state is a view.** A spec
   folder, its contract and its history remain the durable record. No
   consolidated spec is written or edited by hand.

2. **The input is frozen; the output is not.** What 1.x stabilizes is the
   frontmatter and the folder layout, not the shape of the generated view.
   The generation format — columns, ordering, prose — may change in 1.x
   without a breaking change.

   The frozen input includes two fields on the contract's frontmatter:

   - **`capability:`** groups a contract by what it touches (for example
     `guard`, `upgrade`, `workflow`). It is **mandatory in new specs**;
     a spec without it is a **warning** so the delivered specs can be
     backfilled. `forge new --capability <name>` sets it, and warns when the
     name is not one any existing spec already declares, so a typo is
     visible; introducing a genuinely new capability is allowed.
   - **`supersedes: [SPEC-NNN]`** (optional) names the contract this one
     replaces, in whole. Without it the derived view would list
     contradictory contracts as if both were current, so it is the field
     that makes this path viable. `forge validate` **fails** when a
     `supersedes` points at a spec that does not exist or is not `done`,
     when following `supersedes` reaches a cycle, or when two specs that are
     not terminal replace the same target.

3. **The view is generated at read time, deterministically, and never
   committed.** `forge capabilities` (and `forge capabilities <name>`),
   `forge status` and `forge brief` assemble it from the contracts on disk.
   No model is called. Nothing is written to `.forge/`, which avoids both a
   shared file that branches conflict on and a file that goes stale. The
   `superseded` state of an old spec is **computed**, never written into its
   file.

## Alternatives

- **A living spec with deltas (OpenSpec's model).** Rejected for now. It
  freezes the folder layout, a delta grammar (ADDED/MODIFIED/REMOVED) and
  the semantics of `archive` as a merge step, and it reintroduces a shared
  file that two parallel changes edit. Moving from this decision to a living
  spec later is **additive** — a derived view can gain a curated layer.
  Moving from a living spec back to a derived view would break repositories
  that already carry the consolidated files and deltas. The asymmetry is why
  the cheaper model is chosen first. It is reconsidered if large repositories
  with several teams editing the same capabilities appear.

- **Historical only, status quo.** Rejected. "What exists" stays
  archaeology, and the answer gets worse with every closed spec.

- **A hand-maintained index of contracts.** Rejected. An index is a shared
  file, so it conflicts on every parallel branch and drifts from the specs
  it summarizes. Generating it removes both failure modes.

- **Freeze nothing until the model is proven.** Rejected. 1.0 is a promise;
  shipping it without a decided surface makes the promise false.

## Consequences

- The four specs this decision opens (SPEC-010 … SPEC-013) implement it in
  order: `capability` (010) and `supersedes` (011) freeze the input first,
  the derived view (012) consumes them, and the backfill of `capability` on
  the delivered specs (013) lands once the view is in use.

- **Known limitation.** `supersedes` is at the granularity of a whole spec,
  so it cannot express "this contract replaces section three of that one".
  The initial convention is **one spec per behaviour change**, which keeps
  the granularity honest. `supersedes: [SPEC-NNN#section]` is evaluated only
  if this hurts in practice.

- **Known risk.** The view reflects the **intent of the contracts**, not the
  real state of the code. A contract can be `done` while the code drifts
  from it. The view is a description of what was agreed, and `forge
  validate` only checks that the contracts are internally consistent; it
  cannot read the product.

- The delivered specs (001–009) warn until their `capability` is backfilled;
  validate turns that warning off as each is filled.

- The CLI surface grows by `forge capabilities` and the `--capability` flag,
  and `forge validate` grows the `capability` and `supersedes` rules. These
  are part of the frozen surface.
