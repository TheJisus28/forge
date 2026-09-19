package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/TheJisus28/forge/internal/doc"
	"github.com/TheJisus28/forge/internal/project"
	"github.com/TheJisus28/forge/internal/workflow"
	"github.com/TheJisus28/forge/kit"
)

func cmdNew(args []string, out io.Writer) error {
	fs := newFlagSet("new", `usage: forge new "<title>" [--parent SPEC-002] [--covers AC1,AC3]`, out)
	parent := fs.String("parent", "", "the spec this one is part of")
	covers := fs.String("covers", "", "criteria of the parent this spec covers")
	rest, err := parseArgs(fs, args)
	if err != nil {
		return err
	}
	title := strings.TrimSpace(strings.Join(rest, " "))
	if title == "" {
		return fmt.Errorf(`a title is required: forge new "Pay with a saved card"`)
	}
	p, err := project.Load(cwd())
	if err != nil {
		return err
	}

	s := &project.Spec{
		ID:     project.FormatID(p.NextNum()),
		Title:  title,
		Status: workflow.Proposed,
		Parent: project.NormalizeID(*parent),
	}
	if *covers != "" {
		if s.Parent == "" {
			return fmt.Errorf("--covers needs --parent: criteria belong to the parent spec")
		}
		for _, ac := range strings.Split(*covers, ",") {
			if ac = strings.ToUpper(strings.TrimSpace(ac)); ac != "" {
				s.Covers = append(s.Covers, ac)
			}
		}
	}
	if s.Parent != "" {
		parentSpec, ok := p.Spec(s.Parent)
		if !ok {
			return fmt.Errorf("parent %s does not exist", s.Parent)
		}
		known := map[string]bool{}
		for _, c := range parentSpec.Criteria() {
			known[c.ID] = true
		}
		for _, ac := range s.Covers {
			if !known[ac] {
				return fmt.Errorf("%s does not declare %s; its criteria are written in %s",
					parentSpec.ID, ac, filepath.Base(parentSpec.Path))
			}
		}
	}

	d, err := loadTemplate(p, "spec.md")
	if err != nil {
		return err
	}
	d.SetStr("id", s.ID)
	d.SetStr("title", s.Title)
	d.SetStr("status", string(s.Status))
	d.SetStr("created", time.Now().Format("2006-01-02"))
	d.SetStr("updated", time.Now().Format("2006-01-02"))

	path := filepath.Join(p.SpecsDir(), project.FileName(s.ID, s.Title))
	built, err := project.FromDoc(path, d)
	if err != nil {
		return err
	}
	built.Parent, built.Covers = s.Parent, s.Covers
	s = built
	if err := s.Save(); err != nil {
		return err
	}

	rel, _ := filepath.Rel(p.Root, s.Path)
	fmt.Fprintf(out, `%s  %s
  %s

Write the problem and the acceptance criteria, then open the intake pull
request so the number is taken:

  git checkout -b intake/%s
  git add %s && git commit -m "spec(%s): %s"

A maintainer accepts it by approving that pull request.
`, s.ID, s.Title, filepath.ToSlash(rel), project.Slug(s.Title),
		filepath.ToSlash(rel), strings.ToLower(s.ID), s.Title)
	return nil
}

func cmdAccept(args []string, out io.Writer) error {
	fs := newFlagSet("accept", "usage: forge accept <id> --by <maintainer> [--note ...]", out)
	by := fs.String("by", "", "maintainer accepting the work")
	note := fs.String("note", "", "why, in one line")
	rest, err := parseArgs(fs, args)
	if err != nil {
		return err
	}
	p, s, err := specArg(rest)
	if err != nil {
		return err
	}
	if err := requireMaintainer(p, *by); err != nil {
		return err
	}
	if err := workflow.Check(s.Status, workflow.Accepted); err != nil {
		return err
	}
	s.AcceptedBy = *by
	s.SetStatus(workflow.Accepted, *by, *note)
	if err := s.Save(); err != nil {
		return err
	}
	fmt.Fprintf(out, "%s accepted by %s. Anyone can now run: forge start %s\n", s.ID, *by, s.ID)
	return nil
}

func cmdStart(args []string, out io.Writer) error {
	fs := newFlagSet("start", "usage: forge start <id> [--by <you>] [--force]", out)
	by := fs.String("by", "", "who conducts this spec (defaults to git user.name)")
	force := fs.Bool("force", false, "start despite open dependencies, recording the exception")
	rest, err := parseArgs(fs, args)
	if err != nil {
		return err
	}
	p, s, err := specArg(rest)
	if err != nil {
		return err
	}
	if err := workflow.Check(s.Status, workflow.Specifying); err != nil {
		return err
	}
	blockers := p.Blockers(s)
	if len(blockers) > 0 && !*force {
		msg := fmt.Sprintf("%s cannot start yet:\n", s.ID)
		for _, b := range blockers {
			msg += "  - " + b + "\n"
		}
		if ready := readyToStart(p); len(ready) > 0 {
			msg += "ready instead: " + strings.Join(ready, ", ") + "\n"
		}
		msg += "use --force to start anyway; the exception is recorded in the spec"
		return fmt.Errorf("%s", msg)
	}

	conductor := *by
	if conductor == "" {
		conductor = project.UserName(p.Root)
	}
	s.Conductor = conductor
	for _, d := range s.Deps {
		if d.Level != "contract" {
			continue
		}
		if other, ok := p.Spec(d.ID); ok && other.ContractHash != "" {
			s.Agreed[d.ID] = other.ContractHash
		}
	}
	note := ""
	if len(blockers) > 0 {
		note = "forced despite: " + strings.Join(blockers, "; ")
	}
	s.SetStatus(workflow.Specifying, conductor, note)
	if err := os.MkdirAll(p.WipDirFor(s.ID), 0o755); err != nil {
		return err
	}
	if err := s.Save(); err != nil {
		return err
	}

	fmt.Fprintf(out, `%s is yours. Branch:

  git checkout -b %s

Next: the architect writes the Contract section of %s. No product code
until a maintainer approves it.
`, s.ID, project.BranchName(s.ID, s.Title), filepath.Base(s.Path))
	return nil
}

func cmdApprove(args []string, out io.Writer) error {
	fs := newFlagSet("approve", "usage: forge approve <id> --by <maintainer> [--note ...]", out)
	by := fs.String("by", "", "maintainer approving the contract")
	note := fs.String("note", "", "why, in one line")
	rest, err := parseArgs(fs, args)
	if err != nil {
		return err
	}
	p, s, err := specArg(rest)
	if err != nil {
		return err
	}
	if err := requireMaintainer(p, *by); err != nil {
		return err
	}
	if err := workflow.Check(s.Status, workflow.Planning); err != nil {
		if s.Status == workflow.Specifying && strings.TrimSpace(s.Contract()) != "" {
			return fmt.Errorf("%s still says it is being specified; when the contract is "+
				"ready run: forge advance %s --to awaiting-approval", s.ID, s.ID)
		}
		return err
	}
	contract := strings.TrimSpace(s.Contract())
	if contract == "" {
		return fmt.Errorf("%s has an empty Contract section; there is nothing to approve", s.ID)
	}
	if s.Conductor != "" && strings.EqualFold(s.Conductor, *by) && !p.AllowSelfApproval() {
		return fmt.Errorf("%s conducted %s; with more than one maintainer somebody else approves "+
			"(set allow_self_approval: true in project.md to change this)", *by, s.ID)
	}
	s.ApprovedBy = *by
	s.ContractHash = project.HashContract(contract)
	s.SetStatus(workflow.Planning, *by, *note)
	if err := s.Save(); err != nil {
		return err
	}
	fmt.Fprintf(out, "%s approved by %s, contract %s.\n", s.ID, *by, s.ContractHash)
	if waiting := waitingOnContract(p, s.ID); len(waiting) > 0 {
		fmt.Fprintf(out, "unblocked: %s\n", strings.Join(waiting, ", "))
	}
	fmt.Fprintf(out, "Next: split it into phases in .forge/wip/%s/plan.md, then\n"+
		"  forge advance %s --to implementing\n", s.ID, s.ID)
	return nil
}

func cmdAdvance(args []string, out io.Writer) error {
	fs := newFlagSet("advance", "usage: forge advance <id> --to <state> [--by ...] [--note ...]", out)
	to := fs.String("to", "", "target state")
	by := fs.String("by", "", "who is making the move")
	note := fs.String("note", "", "why, in one line")
	rest, err := parseArgs(fs, args)
	if err != nil {
		return err
	}
	p, s, err := specArg(rest)
	if err != nil {
		return err
	}
	target := workflow.State(strings.TrimSpace(*to))
	if target == "" {
		return fmt.Errorf("--to is required; from %s you can go to %s",
			s.Status, joinStates(workflow.Next(s.Status)))
	}
	if err := workflow.Check(s.Status, target); err != nil {
		return err
	}
	if target == workflow.Done {
		return fmt.Errorf("use forge archive %s: done is reached by distilling the spec", s.ID)
	}
	if workflow.NeedsMaintainer(s.Status, target) {
		if err := requireMaintainer(p, *by); err != nil {
			return err
		}
		switch target {
		case workflow.Accepted:
			s.AcceptedBy = *by
		case workflow.Planning:
			return cmdApprove([]string{s.ID, "--by", *by, "--note", *note}, out)
		}
	}
	who := *by
	if who == "" {
		who = "orchestrator"
	}
	s.SetStatus(target, who, *note)
	if err := s.Save(); err != nil {
		return err
	}
	fmt.Fprintf(out, "%s is now %s. Waiting on %s.\n", s.ID, target, workflow.WaitingFor(target))
	return nil
}

func cmdArchive(args []string, out io.Writer) error {
	fs := newFlagSet("archive", "usage: forge archive <id>", out)
	rest, err := parseArgs(fs, args)
	if err != nil {
		return err
	}
	p, s, err := specArg(rest)
	if err != nil {
		return err
	}
	if s.Status != workflow.Reviewing {
		return fmt.Errorf("%s is %s; archive closes a spec that passed review", s.ID, s.Status)
	}
	wip := p.WipDirFor(s.ID)
	if _, err := os.Stat(filepath.Join(wip, "review.md")); err != nil {
		return fmt.Errorf("%s has no review: .forge/wip/%s/review.md is missing", s.ID, s.ID)
	}
	if pending := pendingConventions(wip); len(pending) > 0 {
		return fmt.Errorf("%s still proposes conventions that nobody decided:\n  %s\n"+
			"record them in .forge/conventions/ (a maintainer decides) or remove the section",
			s.ID, strings.Join(pending, "\n  "))
	}
	if err := os.RemoveAll(wip); err != nil {
		return err
	}
	s.SetStatus(workflow.Done, "orchestrator", "archived")
	if err := s.Save(); err != nil {
		return err
	}
	fmt.Fprintf(out, `%s archived. The scaffolding is gone from the tree and stays in git history.

  git add -A && git commit -m "spec(%s): archive"

What remains in main: the contract, the decisions and the conventions.
`, s.ID, strings.ToLower(s.ID))
	if parent, ok := p.Spec(s.Parent); ok {
		reportParent(out, p, parent)
	}
	return nil
}

// cmdGate turns a pull request approval into the state change it means.
// It is what the CI workflow runs, so that approving in GitHub and the
// spec file can never tell different stories.
func cmdGate(args []string, out io.Writer) error {
	fs := newFlagSet("gate", "usage: forge gate --by <maintainer> [--base origin/main]", out)
	by := fs.String("by", "", "the maintainer who approved the pull request")
	base := fs.String("base", "origin/main", "branch this pull request targets")
	dry := fs.Bool("dry-run", false, "say what would happen and change nothing")
	_, err := parseArgs(fs, args)
	if err != nil {
		return err
	}
	p, err := project.Load(cwd())
	if err != nil {
		return err
	}
	if err := requireMaintainer(p, *by); err != nil {
		return err
	}

	targets := specsInBranch(p, *base)
	if len(targets) == 0 {
		fmt.Fprintln(out, "no spec in this pull request; nothing to record")
		return nil
	}
	for _, s := range targets {
		var action string
		switch s.Status {
		case workflow.Proposed:
			action = "accept"
		case workflow.AwaitingApproval:
			action = "approve"
		default:
			fmt.Fprintf(out, "%s is %s; an approval does not move it\n", s.ID, s.Status)
			continue
		}
		if *dry {
			fmt.Fprintf(out, "would %s %s as %s\n", action, s.ID, *by)
			continue
		}
		args := []string{s.ID, "--by", *by, "--note", "approved the pull request"}
		var err error
		if action == "accept" {
			err = cmdAccept(args, out)
		} else {
			err = cmdApprove(args, out)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// specsInBranch finds the specs this branch is about: the ones it changed,
// or the one named by the branch.
func specsInBranch(p *project.Project, base string) []*project.Spec {
	var out []*project.Spec
	seen := map[string]bool{}
	for _, file := range project.ChangedSpecFiles(p.Root, base) {
		name := filepath.Base(strings.TrimSpace(file))
		for _, s := range p.Specs {
			if filepath.Base(s.Path) == name && !seen[s.ID] {
				seen[s.ID] = true
				out = append(out, s)
			}
		}
	}
	if len(out) == 0 {
		if cur, ok := p.Current(); ok {
			out = append(out, cur)
		}
	}
	return out
}

func cmdRenumber(args []string, out io.Writer) error {
	fs := newFlagSet("renumber", "usage: forge renumber <id> [--to <number>]", out)
	to := fs.Int("to", 0, "the new number (defaults to the next free one)")
	rest, err := parseArgs(fs, args)
	if err != nil {
		return err
	}
	p, s, err := specArg(rest)
	if err != nil {
		return err
	}
	num := *to
	if num == 0 {
		num = p.NextNum()
	}
	newID := project.FormatID(num)
	if _, taken := p.Spec(newID); taken {
		return fmt.Errorf("%s is already taken", newID)
	}
	old := s.ID
	if references := referencesTo(p, old); len(references) > 0 {
		return fmt.Errorf("%s is referenced by %s; renumber is only safe before anything points "+
			"at a spec", old, strings.Join(references, ", "))
	}
	newPath := filepath.Join(p.SpecsDir(), project.FileName(newID, s.Title))
	s.ID = newID
	s.Doc().SetStr("id", newID)
	s.Path = newPath
	if err := s.Save(); err != nil {
		return err
	}
	if err := os.Remove(filepath.Join(p.SpecsDir(), project.FileName(old, s.Title))); err != nil &&
		!os.IsNotExist(err) {
		return err
	}
	fmt.Fprintf(out, "%s is now %s\n", old, newID)
	return nil
}

func loadTemplate(p *project.Project, name string) (*doc.Doc, error) {
	path := filepath.Join(p.Root, project.Dir, "kit", "templates", name)
	if data, err := os.ReadFile(path); err == nil {
		return doc.Parse(data)
	}
	data, err := kit.FS.ReadFile("forge/kit/templates/" + name)
	if err != nil {
		return nil, fmt.Errorf("template %s is missing", name)
	}
	return doc.Parse(data)
}

func specArg(args []string) (*project.Project, *project.Spec, error) {
	p, err := project.Load(cwd())
	if err != nil {
		return nil, nil, err
	}
	if len(args) == 0 {
		if cur, ok := p.Current(); ok {
			return p, cur, nil
		}
		return nil, nil, fmt.Errorf("which spec? pass an id, or work on a spec branch")
	}
	id := project.NormalizeID(args[0])
	s, ok := p.Spec(id)
	if !ok {
		return nil, nil, fmt.Errorf("%s does not exist; forge status lists what does", id)
	}
	return p, s, nil
}

func requireMaintainer(p *project.Project, who string) error {
	if strings.TrimSpace(who) == "" {
		return fmt.Errorf("--by is required: this gate belongs to a maintainer (%s)",
			strings.Join(p.Maintainers(), ", "))
	}
	if len(p.Maintainers()) == 0 {
		return fmt.Errorf("project.md declares no maintainers; run the onboarding first")
	}
	if !p.IsMaintainer(who) {
		return fmt.Errorf("%q is not a maintainer; they are: %s", who,
			strings.Join(p.Maintainers(), ", "))
	}
	return nil
}

func readyToStart(p *project.Project) []string {
	var out []string
	for _, s := range p.Specs {
		if s.Status == workflow.Accepted && len(p.Blockers(s)) == 0 {
			out = append(out, s.ID)
		}
	}
	return out
}

func waitingOnContract(p *project.Project, id string) []string {
	var out []string
	for _, s := range p.Specs {
		for _, d := range s.Deps {
			if d.ID == id && d.Level == "contract" && s.Status == workflow.Accepted {
				out = append(out, s.ID)
			}
		}
	}
	return out
}

func referencesTo(p *project.Project, id string) []string {
	var out []string
	for _, s := range p.Specs {
		if s.Parent == id {
			out = append(out, s.ID)
			continue
		}
		for _, d := range s.Deps {
			if d.ID == id {
				out = append(out, s.ID)
				break
			}
		}
	}
	return out
}

// pendingConventions finds convention proposals left unresolved in the
// scaffolding, so archiving does not quietly drop them.
func pendingConventions(wip string) []string {
	var out []string
	entries, err := os.ReadDir(wip)
	if err != nil {
		return nil
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		d, err := doc.Load(filepath.Join(wip, e.Name()))
		if err != nil {
			continue
		}
		for _, heading := range []string{"Proposed conventions", "Convenciones propuestas"} {
			if body := d.Section(heading); body != "" && !isNone(body) {
				out = append(out, e.Name()+": "+firstLine(body))
			}
		}
	}
	return out
}

func isNone(body string) bool {
	s := strings.ToLower(strings.TrimSpace(body))
	return s == "none" || s == "none." || s == "ninguna" || s == "ninguna." || s == "-"
}

func firstLine(s string) string {
	if i := strings.Index(s, "\n"); i > 0 {
		return strings.TrimSpace(s[:i])
	}
	return strings.TrimSpace(s)
}

func reportParent(out io.Writer, p *project.Project, parent *project.Spec) {
	missing := 0
	for _, row := range p.Coverage(parent) {
		covered := false
		for _, child := range row.By {
			if child.Status == workflow.Done {
				covered = true
			}
		}
		if !covered {
			missing++
		}
	}
	if missing == 0 {
		fmt.Fprintf(out, "%s: every criterion is covered by a closed spec; it can be closed.\n",
			parent.ID)
		return
	}
	fmt.Fprintf(out, "%s: %d criteria still not delivered. Run forge status %s.\n",
		parent.ID, missing, parent.ID)
}

func joinStates(states []workflow.State) string {
	parts := make([]string, len(states))
	for i, s := range states {
		parts[i] = string(s)
	}
	return strings.Join(parts, ", ")
}
