package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/TheJisus28/forge/internal/doc"
	"github.com/TheJisus28/forge/internal/project"
	"github.com/TheJisus28/forge/internal/workflow"
	"github.com/TheJisus28/forge/kit"
)

func cmdNew(args []string, out io.Writer) error {
	fs := newFlagSet("new", `usage: forge new "<title>" --capability <name> [--parent SPEC-002] [--covers AC1,AC3]`, out)
	parent := fs.String("parent", "", "the spec this one is part of")
	capability := fs.String("capability", "", "the part of the system this spec touches")
	covers := fs.String("covers", "", "criteria of the parent this spec covers")
	rest, err := parseArgs(fs, args)
	if err != nil {
		return err
	}
	title := strings.TrimSpace(strings.Join(rest, " "))
	if title == "" {
		return fmt.Errorf(`a title is required: forge new "Pay with a saved card"`)
	}
	cap := strings.TrimSpace(*capability)
	if cap == "" {
		return fmt.Errorf(`a capability is required: pass --capability <name>, as in forge new "Pay with a saved card" --capability payments`)
	}
	p, err := project.Load(cwd())
	if err != nil {
		return err
	}

	known := map[string]bool{}
	for _, existing := range p.Specs {
		if existing.Capability != "" {
			known[existing.Capability] = true
		}
	}
	if !known[cap] {
		fmt.Fprintf(out, "warning: no existing spec declares the capability %q; creating it as a new one\n", cap)
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

	d, err := loadTemplate("spec.md")
	if err != nil {
		return err
	}
	d.SetStr("id", s.ID)
	d.SetStr("title", s.Title)
	d.SetStr("status", string(s.Status))
	d.SetStr("created", time.Now().Format("2006-01-02"))
	d.SetStr("updated", time.Now().Format("2006-01-02"))
	d.SetStr("capability", cap)

	path := filepath.Join(p.SpecDir(s.ID, s.Title), "spec.md")
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

Write the problem and the acceptance criteria, then accept it into the
queue:

  forge accept %s
`, s.ID, s.Title, filepath.ToSlash(rel), s.ID)
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
	// The number forge new wrote is provisional: confirm it against the ids
	// already committed on the shared branch before recording the acceptance
	// (SPEC-015, decision 7).
	remote := sharedSpecIDs(p.Root)
	was := s.ID
	historyNote := *note
	if idTaken(p, s, remote) {
		if len(referencesTo(p, s.ID)) > 0 {
			return referencedError(p, s.ID)
		}
		if err := renumberSpec(p, s, nextFreeNum(p, remote)); err != nil {
			return err
		}
		historyNote = fmt.Sprintf("renumbered from %s: taken on main", was)
		if *note != "" {
			historyNote += "; " + *note
		}
	}
	s.AcceptedBy = actor
	s.SetStatus(workflow.Accepted, actor, historyNote)
	if err := s.Save(); err != nil {
		return err
	}
	if was != s.ID {
		fmt.Fprintf(out, "%s was taken on main; renumbered to %s\n", was, s.ID)
	}
	fmt.Fprintf(out, "%s accepted by %s. Anyone can now run: forge start %s\n", s.ID, actor, s.ID)
	return nil
}

func cmdStart(args []string, out io.Writer) error {
	fs := newFlagSet("start", "usage: forge start <id> [--by <you>] [--force]", out)
	by := fs.String("by", "", "who orchestrates this spec (defaults to git user.name)")
	force := fs.Bool("force", false, "start despite open dependencies, recording the exception")
	rest, err := parseArgs(fs, args)
	if err != nil {
		return err
	}
	p, s, err := specArg(rest)
	if err != nil {
		return err
	}
	if err := workflow.Check(s.Status, workflow.Contracting); err != nil {
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

	orchestrator, err := resolveActor(p, *by)
	if err != nil {
		return err
	}
	s.Orchestrator = orchestrator
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
	s.SetStatus(workflow.Contracting, orchestrator, note)
	if err := os.MkdirAll(s.Dir(), 0o755); err != nil {
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
		return err
	}
	contract := strings.TrimSpace(s.Contract())
	if contract == "" {
		return fmt.Errorf("%s has an empty Contract section; there is nothing to approve", s.ID)
	}
	if s.HasOpenQuestions() {
		return fmt.Errorf("%s still has open questions:\n  %s\n"+
			"answer them in the spec, or write None.", s.ID, firstLine(s.OpenQuestions()))
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
	plan, _ := filepath.Rel(p.Root, s.PlanPath())
	fmt.Fprintf(out, "Next: write %s and tick the phases in %s, then\n"+
		"  forge advance %s --to implementing\n",
		filepath.ToSlash(plan), filepath.Base(s.TasksPath()), s.ID)
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
	review, _ := filepath.Rel(p.Root, s.ReviewPath())
	if _, err := os.Stat(s.ReviewPath()); err != nil {
		return fmt.Errorf("%s has no review: %s is missing", s.ID, filepath.ToSlash(review))
	}
	if pending := pendingConventions(s.Dir()); len(pending) > 0 {
		return fmt.Errorf("%s still proposes conventions that nobody decided:\n  %s\n"+
			"record them in .forge/conventions/ or remove the section", s.ID,
			strings.Join(pending, "\n  "))
	}
	s.SetStatus(workflow.Done, "orchestrator", "archived")
	if err := s.Save(); err != nil {
		return err
	}
	fmt.Fprintf(out, `%s archived. The record stays in the tree: spec, plan, tasks and review.

  git add -A && git commit -m "spec(%s): archive"

A future spec reads that folder to know what exists and how it was verified.
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
	old := s.ID
	if len(referencesTo(p, old)) > 0 {
		return referencedError(p, old)
	}
	if err := renumberSpec(p, s, num); err != nil {
		return err
	}
	if err := s.Save(); err != nil {
		return err
	}
	fmt.Fprintf(out, "%s is now %s\n", old, s.ID)
	return nil
}

// renumberSpec moves a spec's folder and rewrites its `id` field and number to
// num. It is the one implementation behind `forge renumber` and the id
// confirmation `forge accept` does (SPEC-015, decision 7); the caller saves.
func renumberSpec(p *project.Project, s *project.Spec, num int) error {
	newID := project.FormatID(num)
	if _, taken := p.Spec(newID); taken {
		return fmt.Errorf("%s is already taken", newID)
	}
	oldDir := s.Dir()
	newDir := p.SpecDir(newID, s.Title)
	if err := os.Rename(oldDir, newDir); err != nil {
		return fmt.Errorf("move %s to %s: %w", filepath.Base(oldDir), filepath.Base(newDir), err)
	}
	s.ID = newID
	s.Num = num
	s.Doc().SetStr("id", newID)
	s.Path = filepath.Join(newDir, "spec.md")
	return nil
}

// referencedError is the refusal both `forge renumber` and `forge accept` use:
// a number can only be changed while nothing points at the spec.
func referencedError(p *project.Project, id string) error {
	return fmt.Errorf("%s is referenced by %s; renumber is only safe before anything points "+
		"at a spec", id, strings.Join(referencesTo(p, id), ", "))
}

// sharedSpecIDs reads the ids committed on the shared branch, trying
// origin/main and then main. A repository with neither keeps local numbering.
func sharedSpecIDs(root string) []string {
	if ids := project.RemoteSpecIDs(root, "origin/main"); len(ids) > 0 {
		return ids
	}
	return project.RemoteSpecIDs(root, "main")
}

// idTaken reports whether the spec's number is already used, either on the
// shared branch or by another spec in the local tree.
func idTaken(p *project.Project, s *project.Spec, remote []string) bool {
	for _, id := range remote {
		if id == s.ID {
			return true
		}
	}
	for _, other := range p.Specs {
		if other != s && other.ID == s.ID {
			return true
		}
	}
	return false
}

// nextFreeNum is the next number over the union of the local tree and the ids
// on the shared branch, so a renumbered spec cannot collide with either.
func nextFreeNum(p *project.Project, remote []string) int {
	max := p.NextNum() - 1
	for _, id := range remote {
		if n, err := idNum(id); err == nil && n > max {
			max = n
		}
	}
	return max + 1
}

// idNum reads the number out of a normalised SPEC-NNN id.
func idNum(id string) (int, error) {
	return strconv.Atoi(strings.TrimPrefix(project.NormalizeID(id), "SPEC-"))
}

// loadTemplate reads a file template from the binary. It has no project
// override on purpose: the templates are a standard, not a project file
// (decision 0001, SPEC-007).
func loadTemplate(name string) (*doc.Doc, error) {
	data, err := kit.Template(name)
	if err != nil {
		return nil, err
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
// may commit. An empty --by falls back to the authenticated gh user and
// then to git user.name, so nobody has to type their own name to move their
// own work forward; the record still says who did it, which is what a
// teammate reviewing the pull request reads instead of asking Forge to
// referee anything.
func resolveActor(p *project.Project, by string) (string, error) {
	by = strings.TrimSpace(by)
	if by == "" {
		by = project.GHUser(p.Root)
	}
	if by == "" {
		by = project.UserName(p.Root)
	}
	if by == "" {
		return "", fmt.Errorf("--by is required: no authenticated gh user and git has no user.name")
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

// pendingConventions finds convention proposals left unresolved in the spec
// folder, so archiving does not quietly drop them.
func pendingConventions(dir string) []string {
	var out []string
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		d, err := doc.Load(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		// Headings are fixed English: `working_language` governs the prose,
		// never the heading the CLI reads.
		if body := stripComments(d.Section("Proposed conventions")); strings.TrimSpace(body) != "" && !isNone(body) {
			out = append(out, e.Name()+": "+firstLine(body))
		}
	}
	return out
}

// commentRe matches an HTML comment, the shape the templates use to ship
// guidance inside a section without proposing anything. The dash form is
// excluded so a `<--` typo is not eaten as a comment.
var commentRe = regexp.MustCompile(`(?s)<!--.*?-->`)

// stripComments removes HTML comments, so the guidance a template ships in a
// `Proposed conventions` section is not read as a proposal.
func stripComments(s string) string {
	return commentRe.ReplaceAllString(s, "")
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
