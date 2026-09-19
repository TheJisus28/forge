# Security Policy

## Supported versions

Only the latest tagged release is supported. Fixes are published as a new
release; older tags do not receive backports.

## Reporting a vulnerability

Please report privately, not in a public issue:

- Open a report through GitHub private vulnerability reporting on
  [TheJisus28/forge](https://github.com/TheJisus28/forge/security/advisories/new), or
- Email jesus.carrascalh@gmail.com.

Include a description of the issue, the steps needed to reproduce it, and what
an attacker could achieve with it.

Forge is maintained by volunteers, so there is no formal SLA. The aim is to
acknowledge a report within a few days and to agree on a disclosure timeline
with you once the issue is understood.

## Threat model

Forge is a single Go binary with no external dependencies, and it is
local-first:

- It only reads and writes files inside the repository it is pointed at.
- It makes no network calls of its own. There is no account, no server, and no
  telemetry.
- `forge sync` shells out to the GitHub CLI (`gh`), which uses the credentials
  already configured on the machine. Forge never handles those credentials
  itself, and the actions it triggers are limited by whatever permissions that
  account has.
- The GitHub workflows Forge plants run `forge validate`, and can commit a
  state change when a maintainer approves a pull request. Those workflows run
  with the permissions granted by the repository that hosts them.

Spec files are plain Markdown committed to the repository, and they are read by
agents and by CI. Do not put secrets, credentials, or private data in them.
