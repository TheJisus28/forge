package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/TheJisus28/forge/internal/project"
	"github.com/TheJisus28/forge/internal/validate"
	"github.com/TheJisus28/forge/internal/view"
)

// Seams a test replaces: the network belongs to git and gh, so a unit test
// stubs them rather than reaching either (SPEC-020, decision 8).
var (
	hasGH      = project.HasGH
	pushBranch = project.Push
	ghRun      = project.GH
)

func cmdStatus(args []string, out io.Writer) error {
	fs := newFlagSet("status", "usage: forge status [id] [--fetch]", out)
	fetch := fs.Bool("fetch", false, "run git fetch first, so the picture includes the remote")
	rest, err := parseArgs(fs, args)
	if err != nil {
		return err
	}
	p, err := project.Load(cwd())
	if err != nil {
		return err
	}
	if *fetch {
		if err := project.Fetch(p.Root); err != nil {
			return fmt.Errorf("git fetch: %w", err)
		}
		if p, err = project.Load(p.Root); err != nil {
			return err
		}
	}
	if len(rest) > 0 {
		id := project.NormalizeID(rest[0])
		s, ok := p.Spec(id)
		if !ok {
			return fmt.Errorf("%s does not exist", id)
		}
		fmt.Fprint(out, view.Detail(p, s))
		return nil
	}
	fmt.Fprint(out, view.Status(p))
	return nil
}

func cmdBrief(args []string, out io.Writer) error {
	fs := newFlagSet("brief", "usage: forge brief [--json]", out)
	asJSON := fs.Bool("json", false, "emit the Claude Code SessionStart hook payload")
	_, err := parseArgs(fs, args)
	if err != nil {
		return err
	}
	p, err := project.Load(cwd())
	if err != nil {
		// A hook runs everywhere, including repositories without Forge.
		if *asJSON {
			return nil
		}
		return err
	}
	text := view.Brief(p)
	if !*asJSON {
		fmt.Fprint(out, text)
		return nil
	}
	payload := map[string]any{
		"hookSpecificOutput": map[string]any{
			"hookEventName":     "SessionStart",
			"additionalContext": text,
		},
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	fmt.Fprintln(out, string(data))
	return nil
}

// cmdCapabilities prints the current shape of the system: every done spec is
// a contract, grouped by the capability it declares. A superseded contract
// stays visible with a suffix rather than being omitted, so the history is
// legible without hiding the current shape. It reads .forge/specs/ and writes
// nothing: no Save, no file, no git and no network, so running it leaves
// `git status` unchanged.
func cmdCapabilities(args []string, out io.Writer) error {
	fs := newFlagSet("capabilities", "usage: forge capabilities [name]", out)
	rest, err := parseArgs(fs, args)
	if err != nil {
		return err
	}
	p, err := project.Load(cwd())
	if err != nil {
		return err
	}

	groups := p.Capabilities()
	if len(rest) > 0 {
		name := strings.TrimSpace(rest[0])
		found := false
		for _, g := range groups {
			if g.Name == name {
				groups = []project.CapabilityGroup{g}
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("no done spec declares the capability %q; forge capabilities lists them", name)
		}
	}
	if len(groups) == 0 {
		fmt.Fprintln(out, "No done specs declare a capability yet.")
		return nil
	}

	var b strings.Builder
	for i, g := range groups {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString(g.Name + "\n")
		for _, c := range g.Contracts {
			fmt.Fprintf(&b, "  %s  %s", c.ID, c.Title)
			if len(c.SupersededBy) > 0 {
				fmt.Fprintf(&b, " (superseded by %s)", strings.Join(c.SupersededBy, ", "))
			}
			b.WriteString("\n")
		}
	}
	fmt.Fprint(out, b.String())
	return nil
}

func cmdValidate(args []string, out, errOut io.Writer) int {
	fs := newFlagSet("validate", "usage: forge validate [--quiet]", out)
	quiet := fs.Bool("quiet", false, "print only errors")
	_, err := parseArgs(fs, args)
	if err != nil {
		return 0
	}
	p, err := project.Load(cwd())
	if err != nil {
		fmt.Fprintf(errOut, "forge: %v\n", err)
		return 1
	}
	findings := validate.Run(p)
	shown := 0
	for _, f := range findings {
		if *quiet && f.Severity != validate.Error {
			continue
		}
		fmt.Fprintln(out, f)
		shown++
	}
	if validate.HasErrors(findings) {
		return 1
	}
	if shown == 0 {
		noun := "specs"
		if len(p.Specs) == 1 {
			noun = "spec"
		}
		fmt.Fprintf(out, "%d %s, no problems\n", len(p.Specs), noun)
	}
	return 0
}

func cmdSync(args []string, out io.Writer) error {
	fs := newFlagSet("sync", "usage: forge sync [id]", out)
	rest, err := parseArgs(fs, args)
	if err != nil {
		return err
	}
	if !project.HasGH() {
		return fmt.Errorf("sync needs the GitHub CLI (gh). Everything else works without it")
	}
	p, s, err := specArg(rest)
	if err != nil {
		return err
	}
	raw, err := project.GH(p.Root, "pr", "view", "--json", "number,state,url")
	if err != nil {
		return fmt.Errorf("gh could not read a pull request for this branch: %w", err)
	}
	var pr struct {
		Number int    `json:"number"`
		State  string `json:"state"`
		URL    string `json:"url"`
	}
	if err := json.Unmarshal([]byte(raw), &pr); err != nil {
		return err
	}
	s.PR = fmt.Sprintf("%d", pr.Number)
	s.Doc().SetStr("pr_url", pr.URL)
	s.Doc().SetStr("pr_state", strings.ToLower(pr.State))
	if err := s.Save(); err != nil {
		return err
	}
	fmt.Fprintf(out, "%s linked to pull request #%d (%s)\n", s.ID, pr.Number, strings.ToLower(pr.State))
	return nil
}

// cmdSubmit closes a spec as a pull request. It pushes the branch and opens
// the PR through gh; it never merges. Without gh, or with --dry-run, it
// prints the commands instead of running them.
func cmdSubmit(args []string, out io.Writer) error {
	fs := newFlagSet("submit", "usage: forge submit [id] [--base <branch>] [--dry-run]", out)
	base := fs.String("base", "main", "the branch the pull request targets")
	dryRun := fs.Bool("dry-run", false, "print the git and gh commands instead of running them")
	rest, err := parseArgs(fs, args)
	if err != nil {
		return err
	}
	p, s, err := specArg(rest)
	if err != nil {
		return err
	}
	branch := project.Branch(p.Root)
	if branch == "" {
		branch = project.BranchName(s.ID, s.Title)
	}
	title, body := submitMessage(p, s)

	if *dryRun || !hasGH() {
		if !*dryRun {
			fmt.Fprintln(out, "gh is not installed; open the pull request yourself:")
		}
		fmt.Fprintf(out, "  git push -u origin %s\n", branch)
		fmt.Fprintf(out, "  gh pr create --base %s --head %s --title %q --fill\n",
			*base, branch, title)
		return nil
	}

	// The final step never leaves uncommitted work out of the pull request
	// (SPEC-020, decision 8).
	if _, err := project.CommitAll(p.Root, checkpointMessage(s)); err != nil {
		return fmt.Errorf("git commit: %w", err)
	}
	if err := pushBranch(p.Root, branch); err != nil {
		return fmt.Errorf("git push: %w", err)
	}
	url, err := ghRun(p.Root, "pr", "create",
		"--base", *base, "--head", branch, "--title", title, "--body", body)
	if err != nil {
		return fmt.Errorf("gh pr create: %w", err)
	}
	s.PR = prNumber(url)
	s.Doc().SetStr("pr_url", url)
	s.Doc().SetStr("pr_state", "open")
	if err := s.Save(); err != nil {
		return err
	}
	fmt.Fprintf(out, "pull request opened: %s\n", url)
	return nil
}

// submitMessage builds the pull request title and body from the spec, so a
// reviewer reads the agreement and not only the diff.
func submitMessage(p *project.Project, s *project.Spec) (title, body string) {
	title = fmt.Sprintf("%s: %s", s.ID, s.Title)
	var b strings.Builder
	rel, err := filepath.Rel(p.Root, s.Path)
	if err != nil {
		rel = s.Path
	}
	fmt.Fprintf(&b, "Spec: %s\n\n", filepath.ToSlash(rel))
	if problem := s.Doc().Section("Problem"); problem != "" {
		fmt.Fprintf(&b, "## Problem\n\n%s\n\n", problem)
	}
	if ac := s.Doc().Section("Acceptance criteria"); ac != "" {
		fmt.Fprintf(&b, "## Acceptance criteria\n\n%s\n\n", ac)
	}
	if contract := s.Contract(); contract != "" {
		fmt.Fprintf(&b, "## Contract\n\n%s\n\n", contract)
	}
	return title, strings.TrimSpace(b.String())
}

// prNumber reads the number out of the URL gh prints.
func prNumber(url string) string {
	url = strings.TrimRight(strings.TrimSpace(url), "/")
	if i := strings.LastIndex(url, "/"); i >= 0 {
		return url[i+1:]
	}
	return url
}
