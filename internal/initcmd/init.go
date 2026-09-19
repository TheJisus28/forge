package initcmd

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/TheJisus28/forge/internal/detect"
	"github.com/TheJisus28/forge/internal/stackmd"
	"github.com/TheJisus28/forge/kit"
)

// Options control how Init writes files into a repository.
type Options struct {
	Force       bool
	ResetMemory bool
	SkipCopilot bool
}

var preserveMemory = map[string]bool{
	"forge/memory/stack.md":        true,
	"forge/memory/decisions.md":    true,
	"forge/memory/constitution.md": true,
}

// Init copies the generic kit into root and writes stack.md from detection.
func Init(root string, opt Options) error {
	root, err := filepath.Abs(root)
	if err != nil {
		return err
	}
	if st, err := os.Stat(root); err != nil {
		return fmt.Errorf("init: %w", err)
	} else if !st.IsDir() {
		return fmt.Errorf("init: %s is not a directory", root)
	}

	marker := filepath.Join(root, "forge", "README.md")
	if _, err := os.Stat(marker); err == nil && !opt.Force {
		return fmt.Errorf("init: forge is already initialized in %s (use --force to rewrite the kit; memory is kept unless --reset-memory)", root)
	}

	err = fs.WalkDir(kit.FS, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		destRel := mapDest(p)
		if destRel == "" {
			return nil
		}
		if opt.SkipCopilot && strings.HasPrefix(destRel, ".github/") {
			return nil
		}
		if preserveMemory[toSlash(destRel)] && !opt.ResetMemory {
			if _, err := os.Stat(filepath.Join(root, destRel)); err == nil {
				return nil
			}
		}
		return writeEmbedded(root, p, destRel)
	})
	if err != nil {
		return err
	}

	stackPath := filepath.Join(root, "forge", "memory", "stack.md")
	if _, err := os.Stat(stackPath); err == nil && !opt.ResetMemory {
		return nil
	}
	st, err := detect.Dir(root)
	if err != nil {
		return err
	}
	return os.WriteFile(stackPath, []byte(stackmd.Render(st)), 0o644)
}

func mapDest(rel string) string {
	switch {
	case rel == "cursor" || strings.HasPrefix(rel, "cursor/"):
		return ".cursor/" + strings.TrimPrefix(rel, "cursor/")
	case rel == "claude" || strings.HasPrefix(rel, "claude/"):
		return ".claude/" + strings.TrimPrefix(rel, "claude/")
	case rel == "github" || strings.HasPrefix(rel, "github/"):
		return ".github/" + strings.TrimPrefix(rel, "github/")
	default:
		return rel
	}
}

func writeEmbedded(root, embedPath, destRel string) error {
	data, err := kit.FS.ReadFile(embedPath)
	if err != nil {
		return err
	}
	dest := filepath.Join(root, filepath.FromSlash(destRel))
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}
	return os.WriteFile(dest, data, 0o644)
}

func toSlash(p string) string {
	return path.Clean(strings.ReplaceAll(p, `\`, "/"))
}
