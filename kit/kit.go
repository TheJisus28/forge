// Package kit holds the Markdown that forge init plants in a repository.
// Editing the files in this directory is how the process evolves; the Go
// code only copies them.
package kit

import "embed"

// FS is the kit, rooted at this directory. Destination in the target repo:
//
//	forge/   → .forge/
//	claude/  → .claude/
//	github/  → .github/   (only with --ci github)
//	the rest → as-is (AGENTS.md, CLAUDE.md)
//
//go:embed AGENTS.md CLAUDE.md all:forge all:claude all:github
var FS embed.FS
