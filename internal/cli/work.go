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

Someone accepts it into the queue when the team is ready:

  forge accept %s
`, s.ID, s.Title, filepath.ToSlash(rel), project.Slug(s.Title),
		filepath.ToSlash(rel), strings.ToLower(s.ID), s.Title, s.ID)
	return nil
}

func cmdAccept(args []string, out io.Writer) error {
	fs := newFlagSet("accept", "usage: forge accept <id> [--by <you>] [--note ...]", out)
	by := fs.String("by", "", "who is accepting the work (defaults to git user.name)")
	note := fs.String("note", "", "why, in one line")
	rest, err := parseArgs(fs, args)
	if err != nil {
		return err
	}
	p, s, err := specArg(rest)
	if err != nil {
		return err
	}
	actor, err := resolveActor(p, *by)
	if err != nil {
		return err
	}
	if err := workflow.Check(s.Status, workflow.Accepted); err != nil {
		return err
	}
	s.AcceptedBy = actor
	s.SetStatus(workflow.Accepted, actor, *note)
	if err := s.Save(); err != nil {
		return err
	}
	fmt.Fprintf(out, "%s accepted by %s. Anyone can now run: forge start %s\n", s.ID, actor, s.ID)
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
until it is approved.
`, s.ID, project.BranchName(s.ID, s.Title), filepath.Base(s.Path))
	return nil
}

func cmdApprove(args []string, out io.Writer) error {
	fs := newFlagSet("approve", "usage: forge approve <id> [--by <you>] [--note ...]", out)
	by := fs.String("by", "", "who is approving the contract (defaults to git user.name)")
	note := fs.String("note", "", "why, in one line")
	rest, err := parseArgs(fs, args)
	if err != nil {
		return err
	}
	p, s, err := specArg(rest)
	if err != nil {
		return err
	}
	actor, err := resolveActor(p, *by)
	if err != nil {
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
	s.ApprovedBy = actor
	s.ContractHash = project.HashContract(contract)
	s.SetStatus(workflow.Planning, actor, *note)
	if err := s.Save(); err != nil {
		return err
	}
	fmt.Fprintf(out, "%s approved by %s, contract %s.\n", s.ID, actor, s.ContractHash)
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
	_, s, err := specArg(rest)
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
	// accept and approve exist as their own commands because they record
	// who did them and, for approve, fingerprint the contract. Route
	// through them so `forge advance` cannot bypass that bookkeeping.
	switch target {
	case workflow.Accepted:
		return cmdAccept([]string{s.ID, "--by", *by, "--note", *note}, out)
	case workflow.Planning:
		return cmdApprove([]string{s.ID, "--by", *by, "--note", *note}, out)
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
			"record them in .forge/conventions/ or remove the section", s.ID,
			strings.Join(pending, "\n  "))
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

// resolveActor names whoever is making a move. Anyone may: Forge has no
// authorization model, the same way git has no authorization model for who
// may commit. An empty --by falls back to git user.name, so nobody has to
// type their own name to move their own work forward; the record still
// says who did it, which is what a teammate reviewing the pull request
// reads instead of asking Forge to referee anything.
func resolveActor(p *project.Project, by string) (string, error) {
	by = strings.TrimSpace(by)
	if by == "" {
		by = project.UserName(p.Root)
	}
	if by == "" {
		return "", fmt.Errorf("--by is required: git has no user.name configured either")
	}
	return by, nil
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
