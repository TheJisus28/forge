---
test: "go test ./..."
dev: "go run ."
working_language: en
guard: on
push: on
fetch: on
---

# Project

## What this product does

Forge is a Go CLI that plants a spec-driven workflow for coding agents into
a repository. It owns the ids, states, dependencies and validation, and
ships no opinions about the target project's technology. This repository is
Forge itself, so the workflow is used to develop the tool.

## Stack

Go 1.22+, standard library only. No external dependencies: the value is a
single portable binary, and `nonet_test.go` fails the build if a package
imports `net/http` and friends.

## Commands

`test` and `dev` are in the frontmatter. Before a change lands, also run
`gofmt -l .` and `go vet ./...`.

## Facts an agent cannot guess

- The CLI owns state: ids, `status`, history and the board are written by
  commands, never by hand.
- No technology opinions belong in `kit/`; it must work in a Rust repo too.
- No authorization model: Forge records who acted, never whether they were
  allowed.
- The binary never touches the network by itself: only `git` and `gh` are
  invoked, and only when the user asks — or when the project opts in with
  `fetch: on`, which lets `forge brief` fetch the remote refs at session
  start. That is a fetch only: it never pulls, merges or rebases.
