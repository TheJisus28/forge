package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/TheJisus28/forge/internal/project"
	"github.com/TheJisus28/forge/kit"
)

type plantOptions struct {
	force bool
	guard bool
	ci    string
	// update refuses to run where the kit was never planted.
	update bool
}

func cmdInit(args []string, out io.Writer) error {
	fs := newFlagSet("init", "usage: forge init [dir] [--ci github] [--no-guard] [--force]", out)
	ci := fs.String("ci", "", "plant CI workflows for a provider (github)")
	noGuard := fs.Bool("no-guard", false, "do not deny product code edits without an active spec")
	force := fs.Bool("force", false, "rewrite files that already exist")
	if err := fs.Parse(args); err != nil {
		return err
	}
	dir := "."
	if fs.NArg() > 0 {
		dir = fs.Arg(0)
	}
	return plant(dir, plantOptions{force: *force, guard: !*noGuard, ci: *ci}, out)
}

func cmdUpdate(args []string, out io.Writer) error {
	fs := newFlagSet("update", "usage: forge update [--force]", out)
	force := fs.Bool("force", false, "also rewrite the explanatory READMEs of your folders")
	if err := fs.Parse(args); err != nil {
		return err
	}
	root, err := project.Find(cwd())
	if err != nil {
		return err
	}
	ci := ""
	if _, err := os.Stat(filepath.Join(root, ".github", "workflows", "forge-validate.yml")); err == nil {
		ci = "github"
	}
	guard := true
	if p, err := project.Load(root); err == nil {
		guard = p.GuardEnabled()
	}
	return plant(root, plantOptions{force: *force, guard: guard, ci: ci, update: true}, out)
}

func plant(dir string, opt plantOptions, out io.Writer) error {
	root, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	st, err := os.Stat(root)
	if err != nil {
		return fmt.Errorf("%s: %w", dir, err)
	}
	if !st.IsDir() {
		return fmt.Errorf("%s is not a directory", dir)
	}
	marker := filepath.Join(root, project.Dir, "README.md")
	_, planted := os.Stat(marker)
	if opt.update && planted != nil {
		return project.ErrNotInitialized
	}

	written, kept := 0, 0
	err = fs.WalkDir(kit.FS, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		dest := mapDest(p)
		if dest == "" {
			return nil
		}
		if strings.HasPrefix(dest, ".github/") && opt.ci != "github" {
			return nil
		}
		data, err := kit.FS.ReadFile(p)
		if err != nil {
			return err
		}
		full := filepath.Join(root, filepath.FromSlash(dest))
		_, exists := os.Stat(full)
		switch {
		case exists != nil: // missing, always write
		case kitOwned(dest), opt.force && !protected(dest):
		default:
			kept++
			return nil
		}
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			return err
		}
		written++
		return os.WriteFile(full, data, 0o644)
	})
	if err != nil {
		return err
	}

	if err := writeClaudeSettings(root, opt.guard); err != nil {
		return err
	}
	if err := ensureGitignore(root); err != nil {
		return err
	}
	for _, dir := range []string{"specs", "wip", "decisions", "conventions"} {
		if err := os.MkdirAll(filepath.Join(root, project.Dir, dir), 0o755); err != nil {
			return err
		}
	}

	if opt.update {
		fmt.Fprintf(out, "kit updated: %d files written, %d of yours untouched\n", written, kept)
		return nil
	}
	fmt.Fprintf(out, `forge ready in %s

Next: open Claude Code here and say "run the Forge onboarding".
The agent will inspect the repo, ask what it cannot infer, and fill
.forge/project.md. Nothing else is configured until you answer.
`, root)
	return nil
}

// mapDest turns a path inside the embedded kit into a path in the repository.
func mapDest(p string) string {
	switch {
	case p == "forge" || strings.HasPrefix(p, "forge/"):
		return project.Dir + "/" + strings.TrimPrefix(p, "forge/")
	case strings.HasPrefix(p, "claude/"):
		return ".claude/" + strings.TrimPrefix(p, "claude/")
	case strings.HasPrefix(p, "github/"):
		return ".github/" + strings.TrimPrefix(p, "github/")
	default:
		return p
	}
}

// kitOwned files belong to Forge and are rewritten on every update.
func kitOwned(dest string) bool {
	switch dest {
	case "AGENTS.md", "CLAUDE.md", project.Dir + "/README.md":
		return true
	}
	return strings.HasPrefix(dest, project.Dir+"/kit/") ||
		strings.HasPrefix(dest, ".claude/") ||
		strings.HasPrefix(dest, ".github/workflows/forge-")
}

// protected files are never rewritten: they are the project's own memory.
func protected(dest string) bool {
	return dest == project.Dir+"/project.md"
}

func ensureGitignore(root string) error {
	path := filepath.Join(root, ".gitignore")
	current, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if strings.Contains(string(current), project.Dir+"/BOARD.md") {
		return nil
	}
	body := string(current)
	if body != "" && !strings.HasSuffix(body, "\n") {
		body += "\n"
	}
	body += "\n# Forge: generated views, regenerate with `forge board`\n" +
		project.Dir + "/BOARD.md\n"
	return os.WriteFile(path, []byte(body), 0o644)
}

// writeClaudeSettings adds the session hooks to .claude/settings.json without
// discarding anything the team already configured there.
func writeClaudeSettings(root string, guard bool) error {
	path := filepath.Join(root, ".claude", "settings.json")
	settings := map[string]any{}
	if data, err := os.ReadFile(path); err == nil {
		if err := json.Unmarshal(data, &settings); err != nil {
			return fmt.Errorf(".claude/settings.json is not valid JSON: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return err
	}

	hooks, _ := settings["hooks"].(map[string]any)
	if hooks == nil {
		hooks = map[string]any{}
	}
	hooks["SessionStart"] = mergeHook(hooks["SessionStart"], "startup|resume|clear|compact",
		"forge brief --json")
	if guard {
		hooks["PreToolUse"] = mergeHook(hooks["PreToolUse"], "Write|Edit", "forge guard")
	}
	settings["hooks"] = hooks

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}

// mergeHook appends our command to a hook list, leaving other entries alone
// and never adding the same command twice.
func mergeHook(existing any, matcher, command string) []any {
	list, _ := existing.([]any)
	for _, entry := range list {
		if strings.Contains(fmt.Sprint(entry), command) {
			return list
		}
	}
	return append(list, map[string]any{
		"matcher": matcher,
		"hooks": []any{map[string]any{
			"type":    "command",
			"command": command,
		}},
	})
}
