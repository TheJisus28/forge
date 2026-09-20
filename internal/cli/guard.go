package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/TheJisus28/forge/internal/project"
	"github.com/TheJisus28/forge/internal/workflow"
)

// hookInput is the subset of the Claude Code PreToolUse payload we read.
type hookInput struct {
	ToolName  string `json:"tool_name"`
	ToolInput struct {
		FilePath string `json:"file_path"`
		Command  string `json:"command"`
	} `json:"tool_input"`
	CWD string `json:"cwd"`
}

// cmdGuard decides whether an edit or a shell command is allowed. It answers
// Claude Code's PreToolUse hook on stdin, or, with --file or --command, exits
// 1 so any other agent can use the same rule. Staying silent means "no
// decision", which lets the normal permission flow continue; only a denial
// is loud.
func cmdGuard(args []string, out io.Writer) error {
	fs := newFlagSet("guard", "usage: forge guard [--explain] [--file <path>] [--command <cmd>]  (reads the hook payload on stdin)", out)
	explain := fs.Bool("explain", false, "say what it would do without reading a hook payload")
	file := fs.String("file", "", "the file an agent would edit; without --explain, denies with exit 1")
	command := fs.String("command", "", "a shell command an agent would run; without --explain, denies with exit 1")
	_, err := parseArgs(fs, args)
	if err != nil {
		return err
	}

	// A path or command with no --explain is the hook-free mode other agents
	// call: exit 1 and print the reason instead of Claude Code's JSON.
	plain := (*file != "" || *command != "") && !*explain

	var in hookInput
	switch {
	case plain:
		in.ToolInput.FilePath = *file
		in.ToolInput.Command = *command
	case *explain:
		in.ToolInput.FilePath = *file
		in.ToolInput.Command = *command
		if in.ToolInput.FilePath == "" && in.ToolInput.Command == "" {
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

	var reason string
	if in.ToolName == "Bash" || in.ToolInput.Command != "" {
		reason = commandDenial(p, in.ToolInput.Command)
	} else {
		reason = denial(p, in.ToolInput.FilePath)
	}
	if reason == "" {
		if *explain {
			fmt.Fprintln(out, "would allow")
		}
		return nil
	}
	switch {
	case *explain:
		fmt.Fprintln(out, "would deny: "+reason)
		return nil
	case plain:
		// Drop the leading "Forge: " so the caller can prefix its own label
		// without the message reading "forge: Forge: ...".
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
		plan, _ := filepath.Rel(p.Root, cur.PlanPath())
		return fmt.Sprintf("Forge: %s is approved but has no plan yet. Write %s, "+
			"then: forge advance %s --to implementing",
			cur.ID, filepath.ToSlash(plan), cur.ID)
	case workflow.Reviewing:
		return fmt.Sprintf("Forge: %s is under review. If the review found failures, run "+
			"forge advance %s --to implementing and say why.", cur.ID, cur.ID)
	default:
		return fmt.Sprintf("Forge: %s is %s; product code only changes while a spec is "+
			"implementing.", cur.ID, cur.Status)
	}
}

var reGHMerge = regexp.MustCompile(`\bgh\s+pr\s+merge\b`)

// commandDenial returns the reason to block a shell command, or "" to stay
// out of the way. It refuses what would land on the default branch: a spec
// ends as a pull request a person merges, not as a direct push or merge.
// The line is read segment by segment, so a token that appears after the
// push — a base branch named on a later command — does not change what the
// push itself does.
func commandDenial(p *project.Project, command string) string {
	if strings.TrimSpace(command) == "" {
		return ""
	}
	if reGHMerge.MatchString(strings.ToLower(command)) {
		return "Forge: do not merge the pull request yourself; a person reviews and merges it."
	}
	def := onDefaultBranch(p)
	for _, seg := range splitSegments(command) {
		sub, args := gitCommand(seg)
		switch sub {
		case "push":
			if def || refspecsToDefault(args) {
				return "Forge: do not push to the default branch; a spec ends as a pull request. " +
					"Open it with forge submit."
			}
		case "merge":
			if def {
				return "Forge: do not merge into the default branch; a person merges the pull request."
			}
		}
	}
	return ""
}

// splitSegments splits a shell line into the commands separated by `;`, `&&`,
// `||`, `|` or a newline, leaving those characters alone inside quotes.
func splitSegments(s string) []string {
	var out []string
	var cur strings.Builder
	var quote byte
	flush := func() {
		if strings.TrimSpace(cur.String()) != "" {
			out = append(out, cur.String())
		}
		cur.Reset()
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if quote != 0 {
			cur.WriteByte(c)
			if c == quote {
				quote = 0
			}
			continue
		}
		switch c {
		case '\'', '"':
			quote = c
			cur.WriteByte(c)
		case ';', '\n':
			flush()
		case '&', '|':
			if i+1 < len(s) && s[i+1] == c {
				i++
			}
			flush()
		default:
			cur.WriteByte(c)
		}
	}
	flush()
	return out
}

// shellFields splits one command segment into words, keeping quoted spans
// together. It is not a shell parser, only enough to find the command name
// and its arguments.
func shellFields(s string) []string {
	var out []string
	var cur strings.Builder
	var quote byte
	flush := func() {
		if cur.Len() > 0 {
			out = append(out, cur.String())
			cur.Reset()
		}
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if quote != 0 {
			if c == quote {
				quote = 0
			} else {
				cur.WriteByte(c)
			}
			continue
		}
		switch c {
		case '\'', '"':
			quote = c
		case ' ', '\t':
			flush()
		default:
			cur.WriteByte(c)
		}
	}
	flush()
	return out
}

// gitCommand returns the subcommand of a segment that is a git invocation,
// and the words after it. A segment is a git command only when git is its
// first word, so `echo git push origin main` is not read as a push.
func gitCommand(segment string) (string, []string) {
	fields := shellFields(segment)
	i := 0
	for i < len(fields) && isAssignment(fields[i]) {
		i++
	}
	if i >= len(fields) {
		return "", nil
	}
	switch filepath.Base(fields[i]) {
	case "git", "git.exe":
	default:
		return "", nil
	}
	j := i + 1
	for j < len(fields) {
		switch fields[j] {
		case "-C", "-c", "--git-dir", "--work-tree", "--namespace", "--exec-path":
			j += 2
			continue
		}
		if strings.HasPrefix(fields[j], "-") {
			j++
			continue
		}
		return strings.ToLower(fields[j]), fields[j+1:]
	}
	return "", nil
}

// isAssignment reports whether a word is a leading VAR=value assignment, so
// `FOO=bar git push` still finds the git command.
func isAssignment(word string) bool {
	eq := strings.IndexByte(word, '=')
	if eq <= 0 || strings.HasPrefix(word, "-") {
		return false
	}
	for _, r := range word[:eq] {
		if !(r == '_' || r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9') {
			return false
		}
	}
	return true
}

// refspecsToDefault reports whether any positional argument of a git push
// names main or master, including HEAD:main, main:feature and a bare main.
func refspecsToDefault(args []string) bool {
	for _, a := range args {
		if strings.HasPrefix(a, "-") {
			continue
		}
		if targetsDefault(a) {
			return true
		}
	}
	return false
}

// targetsDefault reports whether a ref names the default branch on either
// side of a push refspec.
func targetsDefault(ref string) bool {
	r := strings.ToLower(strings.TrimPrefix(strings.TrimSpace(ref), "+"))
	if r == "" {
		return false
	}
	for _, part := range strings.Split(r, ":") {
		for _, prefix := range []string{"refs/heads/", "heads/", "refs/"} {
			part = strings.TrimPrefix(part, prefix)
		}
		if part == "main" || part == "master" {
			return true
		}
	}
	return false
}

// onDefaultBranch reports whether the current branch is main or master. It is
// false outside a repository, so a missing git leaves only the explicit check.
func onDefaultBranch(p *project.Project) bool {
	b := strings.ToLower(project.Branch(p.Root))
	return b == "main" || b == "master"
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
	return isRootPaperwork(rel)
}

// isRootPaperwork reports whether a repository-root file is paperwork: root
// Markdown and licence/notice files are always editable, so correcting a
// changelog or a licence needs no product spec.
func isRootPaperwork(rel string) bool {
	if rel == "" || strings.Contains(rel, "/") {
		return false
	}
	switch strings.ToLower(filepath.Ext(rel)) {
	case ".md", ".markdown":
		return true
	}
	name := strings.ToUpper(filepath.Base(rel))
	return name == "LICENSE" || name == "NOTICE" ||
		strings.HasPrefix(name, "LICENSE.") || strings.HasPrefix(name, "NOTICE.")
}
