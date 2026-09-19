# Stack (template)

Copied to `forge/memory/stack.md` by `forge init` (detected) or filled
by hand. Agents use `languages` / `playbooks` to load
`forge/playbooks/<lang>.md`.

```yaml
languages:
  - java          # java | go | node | typescript | …
playbooks:
  - java
runtime:          # spring-boot | echo | express | next | …
language_version: "21"
build:            # gradle | maven | go | pnpm | …
db:               # postgresql | mysql | sqlite | none
migrations:       # flyway | dbmate | prisma | none
test: "…"         # exact command
dev: "…"
```

## Facts

What a new agent must know that is not obvious from the tree (HTTP
envelope, auth, profiles).

## Not decided here

Leave empty. Migrations and features are not announced in the stack.
