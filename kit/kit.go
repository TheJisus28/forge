// Package kit holds the Markdown that forge init plants in a repository.
// Editing the files in this directory is how the process evolves; the Go
// code only copies them.
package kit

import "embed"

// FS is the kit, rooted at this directory. Destination in the target repo:
//
//	forge/    → .forge/
//	claude/   → .claude/
//	opencode/ → .opencode/
//	github/   → .github/   (only with --ci github)
//	the rest  → as-is (AGENTS.md, CLAUDE.md)
//
// machine/ is the exception: it is embedded but never planted. It is the
// machinery (the workflow, the roles and the templates) that the binary
// serves through `forge workflow`, `forge roles` and `forge template`.
//
//go:embed AGENTS.md CLAUDE.md all:machine all:forge all:claude all:opencode all:github
var FS embed.FS
