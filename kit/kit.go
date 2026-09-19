// Package kit holds the Markdown that forge init plants in a repository.
// Editing the files in this directory is how the kit evolves; the Go code
// only copies them.
package kit

import "embed"

// FS is the kit, rooted at this directory. Destination in the target repo:
//
//	cursor/  → .cursor/
//	claude/  → .claude/
//	github/  → .github/
//	the rest → as-is (AGENTS.md, CLAUDE.md, GEMINI.md, forge/)
//
//go:embed AGENTS.md CLAUDE.md GEMINI.md all:forge all:cursor all:claude all:github
var FS embed.FS
