# What and why

Describe what this pull request changes and the reason for the change. Link any
related issue.

# How it was verified

```bash
go test ./...
gofmt -l .
```

Add anything else you ran or checked by hand.

# Checklist

- [ ] No new external dependencies; the CLI still builds on the standard library alone.
- [ ] Documentation is updated if behaviour changed.
- [ ] Kit Markdown stays free of technology-specific opinions; Forge remains language-agnostic.
