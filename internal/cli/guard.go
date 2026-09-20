package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/TheJisus28/forge/internal/project"
	"github.com/TheJisus28/forge/internal/workflow"
)

// hookInput is the subset of the Claude Code PreToolUse payload we read.
type hookInput struct {
	ToolName  string `json:"tool_name"`
	ToolInput struct {
		FilePath string `json:"file_path"`
	} `json:"tool_input"`
	CWD string `json:"cwd"`
}

// cmdGuard decides whether an edit to a file is allowed. It answers Claude
// Code's PreToolUse hook on stdin, or, with --file, exits 1 so any other
// agent can use the same rule. Staying silent means "no decision", which
// lets the normal permission flow continue; only a denial is loud.
func cmdGuard(args []string, out io.Writer) error {
	fs := newFlagSet("guard", "usage: forge guard [--explain] [--file <path>]  (reads the hook payload on stdin)", out)
	explain := fs.Bool("explain", false, "say what it would do without reading a hook payload")
	file := fs.String("file", "", "the file an agent would edit; without --explain, denies with exit 1")
	_, err := parseArgs(fs, args)
	if err != nil {
		return err
	}

	// A path with no --explain is the hook-free mode other agents call:
	// exit 1 and print the reason instead of answering Claude Code's JSON.
	plain := *file != "" && !*explain

	var in hookInput
	switch {
	case plain:
		in.ToolInput.FilePath = *file
	case *explain:
		in.ToolInput.FilePath = *file
		if in.ToolInput.FilePath == "" {
			// Ask about product code, which is what the guard is for.
			in.ToolInput.FilePath = "src/example"
		}
	default:
		data, err := io.ReadAll(os.Stdin)
		if err != nil || len(strings.TrimSpace(string(data))) == 0 {
			return nil
		}
		if err := json.Unmarshal(data, &in); err != nil {
			return nil
		}
	}

	dir := in.CWD
	if dir == "" {
		dir = cwd()
	}
	p, err := project.Load(dir)
	if err != nil || !p.GuardEnabled() {
		return nil
	}
	if reason := denial(p, in.ToolInput.FilePath); reason != "" {
		switch {
		case *explain:
			fmt.Fprintln(out, "would deny: "+reason)
			return nil
		case plain:
			// Drop the leading "Forge: " so the caller can prefix its own
			// label without the message reading "forge: Forge: ...".
			return errors.New(strings.TrimPrefix(reason, "Forge: "))
		}
		payload := map[string]any{
			"hookSpecificOutput": map[string]any{
				"hookEventName":            "PreToolUse",
				"permissionDecision":       "deny",
				"permissionDecisionReason": reason,
			},
		}
		data, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		fmt.Fprintln(out, string(data))
		return nil
	}
	if *explain {
		fmt.Fprintln(out, "would allow")
	}
	return nil
}

// denial returns the reason to block an edit, or "" to stay out of the way.
func denial(p *project.Project, file string) string {
	if isProcessFile(p.Root, file) {
		return ""
	}
	cur, ok := p.Current()
	if !ok {
		ready := readyToStart(p)
		msg := "Forge: this branch has no spec, so product code cannot change here.\n" +
			"Open one with forge new \"<title>\", or switch to a spec branch."
		if len(ready) > 0 {
			msg += "\nAccepted and ready to start: " + strings.Join(ready, ", ")
		}
		return msg
	}
	if cur.Status == workflow.Implementing {
		return ""
	}
	switch cur.Status {
	case workflow.Specifying:
		return fmt.Sprintf("Forge: %s is still being specified. Write the Contract section "+
			"first, then: forge advance %s --to awaiting-approval && forge approve %s",
			cur.ID, cur.ID, cur.ID)
	case workflow.AwaitingApproval:
		return fmt.Sprintf("Forge: %s has a contract but it is not approved yet. "+
			"Nothing is built until: forge approve %s", cur.ID, cur.ID)
	case workflow.Planning:
		return fmt.Sprintf("Forge: %s is approved but has no plan yet. Write the phases in "+
			".forge/wip/%s/plan.md, then: forge advance %s --to implementing",
			cur.ID, cur.ID, cur.ID)
	case workflow.Reviewing:
		return fmt.Sprintf("Forge: %s is under review. If the review found failures, run "+
			"forge advance %s --to implementing and say why.", cur.ID, cur.ID)
	default:
		return fmt.Sprintf("Forge: %s is %s; product code only changes while a spec is "+
			"implementing.", cur.ID, cur.Status)
	}
}

// isProcessFile reports whether the path is Forge's own paperwork, which is
// always editable: that is where the agent writes contracts and plans.
func isProcessFile(root, file string) bool {
	if file == "" {
		return true
	}
	rel := file
	if filepath.IsAbs(file) {
		if r, err := filepath.Rel(root, file); err == nil {
			rel = r
		}
	}
	rel = filepath.ToSlash(rel)
	if strings.HasPrefix(rel, "../") {
		return true // outside the repository; not our business
	}
	for _, prefix := range []string{project.Dir + "/", ".claude/", ".github/", ".opencode/"} {
		if strings.HasPrefix(rel, prefix) {
			return true
		}
	}
	switch rel {
	case "AGENTS.md", "CLAUDE.md", "README.md", ".gitignore",
		"opencode.json", "opencode.jsonc":
		return true
	}
	return false
}
