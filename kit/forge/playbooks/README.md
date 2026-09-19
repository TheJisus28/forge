# Language playbooks

The orchestrator and implementer load **only** the playbooks named in
`forge/memory/stack.md` (`languages` or `playbooks`).

| Language | File |
|---|---|
| Java | [java.md](java.md) |
| Go | [go.md](go.md) |
| Node / TypeScript | [node.md](node.md) |
| Python | [python.md](python.md) |

If the repo uses another language, add `playbooks/<lang>.md` in the same
shape and list it in `stack.md`. Do not copy rules from a playbook that
is not in the stack.
