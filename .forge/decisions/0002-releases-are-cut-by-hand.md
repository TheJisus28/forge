---
id: 0002
date: "2026-09-20"
status: accepted
---

# Releases are cut by hand when a spec is merged to main

## Context

Forge ships binaries through GoReleaser, triggered by pushing a `v*` tag
(`.github/workflows/release.yml` and `.goreleaser.yaml`). Until SPEC-008
there was no way to move an installed binary forward; now `forge upgrade`
installs whatever `@latest` resolves to. That makes the tag the artifact
users actually receive: if `main` moves without a tag, `forge upgrade` keeps
handing them the old release and the repository drifts from the binary.

There was no rule saying when to cut a release, so SPEC-008 was merged with
nothing for `forge upgrade` to reach.

## Decision

Every merge to `main` that closes a spec is packaged as a release, and
releases are cut by hand. After the spec is archived and its pull request
merged, the conductor proposes the next version from Conventional Commits
(`fix` → patch, `feat` → minor, `!`/`BREAKING CHANGE` → major), the
maintainer approves it, and that exact `vX.Y.Z` tag is pushed. The tag push
is the only thing that publishes.

## Alternatives

- **Automatic release pull request (release-please).** Rejected. It adds an
  external GitHub Action and a second pull request per release for a
  project with one maintainer.
- **Direct auto-tag on merge.** Rejected. The version would be published
  without the maintainer seeing it; correcting it means deleting and
  re-pushing a tag.
- **No releases until someone asks.** Rejected. That is the status quo that
  left SPEC-008 unreachable by `forge upgrade`.

## Consequences

- `main` and the latest release can diverge between the merge and the tag;
  cutting the tag is part of closing a spec, not an afterthought.
- `forge upgrade` reaches a new version only after the tag exists, so the
  tag is pushed before the change is announced.
- The process relies on the conductor, so the rule is written here and
  repeated in `AGENTS.md`.
