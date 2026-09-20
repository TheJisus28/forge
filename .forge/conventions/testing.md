---
domain: testing
approved_by: TheJisus28
date: 2026-09-20
---

# Convention — testing

## Rule

A substitution point is allowed, never required. When an external
dependency — the OS, the Go toolchain, the clock, the network — would make a
test non-hermetic or non-deterministic, expose it through an explicit seam
that the test can replace. But a mutable package-level variable is not the
mandatory way to inject dependencies here: choose the simplest shape the
case allows — a direct call, a parameter, a small unexported interface, or a
package-level variable — and add a seam only where a test must actually
substitute it. A command that promises to write nothing proves it with a
before/after snapshot: the test hashes or fully lists the affected tree
around the call, rather than asserting on individual files, so a write
anywhere under the root fails the test.

## Example

`internal/cli/upgrade.go` declares only the calls a test must replace —
`lookupGo`, `goInstall`, `goVersion`, `executable`, `renameFile`,
`removeFile` — as package-level variables, and
`internal/cli/upgrade_internal_test.go` swaps them through one `swap` helper
registered on `t.Cleanup`. Nothing else in the package is a seam.

`internal/cli/cli_test.go:TestCapabilities_IsDeterministicAndWritesNothing`
runs `forge capabilities` twice, compares the SHA256 of every file under the
repository before and after, and asserts `git status --porcelain` is empty,
so any stray write fails the test rather than one overlooked file.

## Why

Tests must not run the real toolchain or touch the network, but routing every
external call through a package-level variable would make the code harder to
read for no test benefit. A seam is a tool, chosen per case, not a blanket
rule.

## Exceptions

Pure logic with no external dependency needs no substitution point and is
called directly in tests.
