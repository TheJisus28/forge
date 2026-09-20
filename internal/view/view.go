// Package view renders what a human or an agent needs to read: the session
// brief, the status tree and one spec's detail.
package view

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/TheJisus28/forge/internal/project"
	"github.com/TheJisus28/forge/internal/workflow"
)

// deliveredLimit bounds the "already delivered" list in the brief. The list
// must not grow with the number of closed specs: every session and
// compaction pays for it. forge status stays the way to see all of them.
const deliveredLimit = 5

// Brief is the short project state injected at the start of an agent session.
// It is deliberately small: it is paid for on every session and every compaction.
func Brief(p *project.Project) string {
	var b strings.Builder
	b.WriteString("Forge")
	branch := project.Branch(p.Root)
	if branch != "" {
		b.WriteString(" — branch " + branch)
	}
	b.WriteString("\n\n")

	if !p.Configured() {
		b.WriteString("This project is not onboarded yet: .forge/project.md has no stack\n")
		b.WriteString("and no test command.\n\n")
		b.WriteString("Next action: run the forge-onboard skill. Inspect the repository,\n")
		b.WriteString("propose what you can infer, ask about the rest, and write project.md\n")
		b.WriteString("only with confirmed answers.\n")
		return b.String()
	}

	if cur, ok := p.Current(); ok {
		fmt.Fprintf(&b, "%s  %s  %s\n", cur.ID, cur.Status, cur.Title)
		fmt.Fprintf(&b, "  capability  %s\n", capabilityOrNone(cur.Capability))
		fmt.Fprintf(&b, "  waiting on %s\n", workflow.WaitingFor(cur.Status))
		if done, total := cur.TaskProgress(); total > 0 {
			fmt.Fprintf(&b, "  tasks %d/%d\n", done, total)
		}
		if blockers := p.Blockers(cur); len(blockers) > 0 {
			fmt.Fprintf(&b, "  blocked: %s\n", strings.Join(blockers, "; "))
		}
		b.WriteString("\n")
	}

	capabilitiesSection(&b, p)

	var waiting, inFlight, ready, delivered []*project.Spec
	for _, s := range p.Specs {
		switch {
		case s.Status == workflow.Proposed:
			waiting = append(waiting, s)
		case workflow.InFlight(s.Status):
			inFlight = append(inFlight, s)
		case s.Status == workflow.Accepted && len(p.Blockers(s)) == 0:
			ready = append(ready, s)
		case s.Status == workflow.Done:
			delivered = append(delivered, s)
		}
	}
	section(&b, "open decisions", waiting, func(*project.Spec) string {
		return "forge accept or drop"
	})
	section(&b, "in flight", inFlight, func(s *project.Spec) string {
		if s.Orchestrator != "" {
			return string(s.Status) + ", " + s.Orchestrator
		}
		return string(s.Status)
	})
	section(&b, "ready to start", ready, func(*project.Spec) string { return "forge start" })
	deliveredSection(&b, delivered)

	if drift := Drift(p); len(drift) > 0 {
		b.WriteString("contract drift:\n")
		for _, d := range drift {
			b.WriteString("  " + d + "\n")
		}
		b.WriteString("\n")
	}

	if age, ok := project.FetchAge(p.Root); ok && age > 48*time.Hour {
		fmt.Fprintf(&b, "note: this clone last fetched %s ago; run forge status --fetch\n\n",
			humanAge(age))
	}

	b.WriteString("Rules that matter here: no product code without a spec in implementing;\n")
	b.WriteString("conventions live in .forge/conventions/ (propose new ones, never\n")
	b.WriteString("assume them); and before planning a change, survey what already exists\n")
	b.WriteString("in the specs and the code and reuse it.\n")
	return b.String()
}

func section(b *strings.Builder, title string, specs []*project.Spec, note func(*project.Spec) string) {
	if len(specs) == 0 {
		return
	}
	fmt.Fprintf(b, "%s:\n", title)
	for _, s := range specs {
		fmt.Fprintf(b, "  %s  %-40s %s\n", s.ID, truncate(s.Title, 40), note(s))
	}
	b.WriteString("\n")
}

// capabilitiesSection prints the current shape of the system, one line per
// capability: the count of contracts no non-dropped spec has superseded. It
// shares project.Capabilities() with `forge capabilities`, so the brief and
// the command cannot disagree, and it prints nothing when no done spec
// declares a capability yet.
func capabilitiesSection(b *strings.Builder, p *project.Project) {
	groups := p.Capabilities()
	if len(groups) == 0 {
		return
	}
	b.WriteString("capabilities:\n")
	for _, g := range groups {
		current := 0
		for _, c := range g.Contracts {
			if c.Current() {
				current++
			}
		}
		noun := "contracts"
		if current == 1 {
			noun = "contract"
		}
		fmt.Fprintf(b, "  %-20s %d current %s\n", g.Name, current, noun)
	}
	b.WriteString("\n")
}

// deliveredSection lists the most recent done specs, capped so the brief
// stays small, and points at forge status when it hides older ones.
func deliveredSection(b *strings.Builder, delivered []*project.Spec) {
	if len(delivered) == 0 {
		return
	}
	hidden := len(delivered) - deliveredLimit
	shown := delivered
	if hidden > 0 {
		shown = delivered[hidden:] // specs arrive oldest first
	}
	b.WriteString("already delivered (newest first, reuse before rebuilding):\n")
	for i := len(shown) - 1; i >= 0; i-- {
		fmt.Fprintf(b, "  %s  %s\n", shown[i].ID, truncate(shown[i].Title, 40))
	}
	if hidden > 0 {
		fmt.Fprintf(b, "  +%d earlier, not listed; forge status shows every spec\n", hidden)
	}
	b.WriteString("\n")
}

// Drift lists specs building against a contract that changed after they
// started, which is the classic back/front integration failure.
func Drift(p *project.Project) []string {
	var out []string
	for _, s := range p.Specs {
		if s.ContractChanged() {
			out = append(out, fmt.Sprintf("%s's contract changed after approval", s.ID))
		}
		if workflow.Terminal(s.Status) {
			continue
		}
		for dep, hash := range s.Agreed {
			other, ok := p.Spec(dep)
			if !ok || other.ContractHash == "" || other.ContractHash == hash {
				continue
			}
			out = append(out, fmt.Sprintf("%s started against %s's contract %s, now %s",
				s.ID, dep, hash, other.ContractHash))
		}
	}
	sort.Strings(out)
	return out
}

// Status is the full report: the tree of specs with coverage and blockers.
func Status(p *project.Project) string {
	var b strings.Builder
	if len(p.Specs) == 0 {
		return "No specs yet. Describe what you need and let the agent open one, " +
			"or run: forge new \"title\"\n"
	}
	roots := make([]*project.Spec, 0)
	for _, s := range p.Specs {
		if s.Parent == "" {
			roots = append(roots, s)
		}
	}
	for _, s := range roots {
		writeSpec(&b, p, s, "")
	}
	orphans := make([]*project.Spec, 0)
	for _, s := range p.Specs {
		if s.Parent != "" {
			if _, ok := p.Spec(s.Parent); !ok {
				orphans = append(orphans, s)
			}
		}
	}
	for _, s := range orphans {
		writeSpec(&b, p, s, "")
		fmt.Fprintf(&b, "    parent %s does not exist\n", s.Parent)
	}
	if drift := Drift(p); len(drift) > 0 {
		b.WriteString("\ncontract drift:\n")
		for _, d := range drift {
			b.WriteString("  " + d + "\n")
		}
	}
	return b.String()
}

func writeSpec(b *strings.Builder, p *project.Project, s *project.Spec, indent string) {
	fmt.Fprintf(b, "%s%s  %-18s %s\n", indent, s.ID, s.Status, s.Title)
	if !workflow.Terminal(s.Status) {
		fmt.Fprintf(b, "%s    waiting on %s\n", indent, workflow.WaitingFor(s.Status))
		for _, reason := range p.Blockers(s) {
			fmt.Fprintf(b, "%s    blocked: %s\n", indent, reason)
		}
	}
	children := p.Children(s.ID)
	if len(children) == 0 {
		return
	}
	for _, row := range p.Coverage(s) {
		if len(row.By) == 0 {
			fmt.Fprintf(b, "%s    %s %-40s NOT COVERED\n", indent, row.Criterion.ID,
				truncate(row.Criterion.Text, 40))
			continue
		}
		ids := make([]string, 0, len(row.By))
		for _, c := range row.By {
			ids = append(ids, fmt.Sprintf("%s (%s)", c.ID, c.Status))
		}
		fmt.Fprintf(b, "%s    %s %-40s %s\n", indent, row.Criterion.ID,
			truncate(row.Criterion.Text, 40), strings.Join(ids, ", "))
	}
	for _, c := range children {
		writeSpec(b, p, c, indent+"  ")
	}
}

// Detail is the report for a single spec: state, blockers, criteria,
// coverage when it has children, and the command that moves it forward.
func Detail(p *project.Project, s *project.Spec) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s  %s\n%s\n\n", s.ID, s.Title, strings.Repeat("-", 60))
	fmt.Fprintf(&b, "status      %s\n", s.Status)
	fmt.Fprintf(&b, "capability  %s\n", capabilityOrNone(s.Capability))
	if len(s.Supersedes) > 0 {
		fmt.Fprintf(&b, "supersedes     %s\n", strings.Join(s.Supersedes, ", "))
	}
	if by := supersededBy(p, s); len(by) > 0 {
		fmt.Fprintf(&b, "superseded by  %s\n", strings.Join(by, ", "))
	}
	fmt.Fprintf(&b, "waiting on  %s\n", workflow.WaitingFor(s.Status))
	if done, total := s.TaskProgress(); total > 0 {
		fmt.Fprintf(&b, "tasks       %d/%d\n", done, total)
	}
	if s.Orchestrator != "" {
		fmt.Fprintf(&b, "orchestrator   %s\n", s.Orchestrator)
	}
	if s.ApprovedBy != "" {
		fmt.Fprintf(&b, "approved    %s, contract %s\n", s.ApprovedBy, s.ContractHash)
	}
	if s.PR != "" {
		fmt.Fprintf(&b, "pull req    #%s\n", s.PR)
	}
	if s.Parent != "" {
		fmt.Fprintf(&b, "parent      %s, covers %s\n", s.Parent, strings.Join(s.Covers, ", "))
	}
	if len(s.Deps) > 0 {
		b.WriteString("depends on\n")
		for _, d := range s.Deps {
			mark := "ok"
			if ok, why := p.DepSatisfied(d); !ok {
				mark = why
			}
			fmt.Fprintf(&b, "  %-22s %s\n", d.String(), mark)
		}
	}
	if blockers := p.Blockers(s); len(blockers) > 0 {
		b.WriteString("blocked by\n")
		for _, r := range blockers {
			fmt.Fprintf(&b, "  %s\n", r)
		}
	}
	if criteria := s.Criteria(); len(criteria) > 0 {
		b.WriteString("\nacceptance criteria\n")
		for _, c := range criteria {
			fmt.Fprintf(&b, "  %-5s %s\n", c.ID, c.Text)
		}
	}
	if len(p.Children(s.ID)) > 0 {
		b.WriteString("\ncoverage\n")
		for _, row := range p.Coverage(s) {
			if len(row.By) == 0 {
				fmt.Fprintf(&b, "  %-5s NOT COVERED\n", row.Criterion.ID)
				continue
			}
			ids := make([]string, 0, len(row.By))
			for _, c := range row.By {
				ids = append(ids, fmt.Sprintf("%s (%s)", c.ID, c.Status))
			}
			fmt.Fprintf(&b, "  %-5s %s\n", row.Criterion.ID, strings.Join(ids, ", "))
		}
	}
	if next := workflow.Next(s.Status); len(next) > 0 {
		names := make([]string, len(next))
		for i, n := range next {
			names[i] = string(n)
		}
		fmt.Fprintf(&b, "\nnext states  %s\n", strings.Join(names, ", "))
	}
	return b.String()
}

// supersededBy lists, in ID order, the specs that declare they supersede s.
// Specs are loaded in ID order, so the walk needs no sort of its own.
func supersededBy(p *project.Project, s *project.Spec) []string {
	var out []string
	for _, other := range p.Specs {
		if other.ID == s.ID {
			continue
		}
		for _, target := range other.Supersedes {
			if target == s.ID {
				out = append(out, other.ID)
				break
			}
		}
	}
	return out
}

// capabilityOrNone renders a spec's capability, or "(none)" when it does not
// declare one, so a reader never sees an empty field.
func capabilityOrNone(name string) string {
	if name == "" {
		return "(none)"
	}
	return name
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	if n <= 1 {
		return s[:n]
	}
	return s[:n-1] + "…"
}

func humanAge(d time.Duration) string {
	days := int(d.Hours() / 24)
	if days >= 1 {
		if days == 1 {
			return "1 day"
		}
		return fmt.Sprintf("%d days", days)
	}
	return fmt.Sprintf("%d hours", int(d.Hours()))
}
