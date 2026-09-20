---
id: SPEC-015
title: Collapse the redundant gates and rename the contract state
status: proposed
capability: workflow
created: 2026-09-20
updated: 2026-09-20
---

## Problem

The loop carries steps that decide the same thing twice and a state whose
name lies. The intake pull request and `forge accept` are two decisions for
one question ("is this worth doing?"). The architect finishing a contract
needs a manual `forge advance <id> --to awaiting-approval` before `forge
approve`, two commands for one hand-off. The existing-state survey is
required by the architect role and then repeated when planning fills `##
Existing state`. And `specifying` names the state reached after the
problem, criteria and spec file already exist: what is written there is the
contract, not the spec, so the name misleads every reader and every role.

## Acceptance criteria

- AC1: The state that follows `accepted` is named for the contract
  (`contracting`) in the workflow package, `forge workflow`, `forge roles`,
  `forge status` and the docs.
- AC2: `forge approve` accepts a contract written straight from
  `contracting`, without a separate `awaiting-approval` move, and still
  freezes the contract fingerprint and the approver.
- AC3: Accepting work is one gate: the workflow docs describe entering the
  queue once, not an intake pull request plus a separate command.
- AC4: The existing-state survey is written once, where the architect names
  what is reused, and planning does not ask for it again.
- AC5: `forge validate` and `forge guard` report the new state name, and a
  repository still carrying `specifying` is either migrated or reported with
  the command to fix it.

## Open questions

One per line, like `- OQ1: ...`, for anything that must be answered before
the contract is approved. Write `None.` when there are none.

None.

## Contract

Written by the architect once the work is accepted, and frozen once
approved. Real names from this repository: modules, endpoints,
tables, screens. Numbered decisions with what they discard. Anything other
specs will build against goes here.

## Out of scope

A closed list.

## History

Written by `forge`. Do not edit by hand.
