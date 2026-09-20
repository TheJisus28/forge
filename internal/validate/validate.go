// Package validate holds the consistency rules Forge enforces in CI: the
// ones that keep the process honest when nobody is watching.
package validate

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/TheJisus28/forge/internal/doc"
	"github.com/TheJisus28/forge/internal/project"
	"github.com/TheJisus28/forge/internal/workflow"
)

// Severity separates what breaks the build from what is only worth saying.
type Severity int

const (
	Warning Severity = iota
	Error
)

func (s Severity) String() string {
	if s == Error {
		return "error"
	}
	return "warning"
}

// Finding is one problem found in the project.
type Finding struct {
	Severity Severity
	Spec     string
	Message  string
}

func (f Finding) String() string {
	if f.Spec == "" {
		return fmt.Sprintf("%-7s %s", f.Severity, f.Message)
	}
	return fmt.Sprintf("%-7s %s: %s", f.Severity, f.Spec, f.Message)
}

// Run checks the project and returns every finding, errors first.
func Run(p *project.Project) []Finding {
	var out []Finding
	add := func(sev Severity, spec, format string, args ...any) {
		out = append(out, Finding{sev, spec, fmt.Sprintf(format, args...)})
	}

	if !p.Configured() {
		add(Warning, "", "project.md has no test command; run the onboarding")
	}

	seen := map[string]string{}
	for _, s := range p.Specs {
		if prev, dup := seen[s.ID]; dup {
			add(Error, s.ID, "duplicate id, also in %s; run forge renumber",
				filepath.Base(prev))
		}
		seen[s.ID] = s.Path

		if strings.TrimSpace(s.Title) == "" {
			add(Error, s.ID, "missing title")
		}
		checkCapability(s, add)
		checkSupersedes(p, s, add)
		if !workflow.Valid(s.Status) {
			add(Error, s.ID, "unknown status %q", s.Status)
			continue
		}
		checkRelations(p, s, add)
		checkArtifacts(p, s, add)
		if s.HasOpenQuestions() {
			q := s.OpenQuestions()
			if i := strings.IndexByte(q, '\n'); i > 0 {
				q = q[:i]
			}
			add(Warning, s.ID, "has open questions: %s", strings.TrimSpace(q))
		}
	}

	for _, s := range p.Specs {
		checkCoverage(p, s, add)
	}
	for _, d := range drift(p) {
		add(Error, "", "%s", d)
	}

	sortFindings(out)
	return out
}

// checkCapability reports a spec with no `capability` key as a warning, so
// the specs delivered before the field existed stay visible without failing
// the build, and a malformed one as an error. The typed field cannot tell
// "absent" from "present but empty", so the raw document does.
func checkCapability(s *project.Spec, add func(Severity, string, string, ...any)) {
	if !s.Doc().Has("capability") {
		add(Warning, s.ID, "has no capability; a new spec sets one with forge new --capability <name>")
		return
	}
	if !project.ValidCapability(s.Capability) {
		add(Error, s.ID, "capability %q is not a lowercase slug ([a-z0-9-]+)", s.Capability)
	}
}

// checkSupersedes enforces the supersede link: it must point at a delivered
// contract, it must not loop, and only one live spec may claim the same
// replacement. It is an intrinsic check, so it runs before the state gate:
// the shape of the list does not depend on the state the spec claims.
func checkSupersedes(p *project.Project, s *project.Spec, add func(Severity, string, string, ...any)) {
	if len(s.Supersedes) == 0 {
		return
	}
	for _, target := range s.Supersedes {
		other, ok := p.Spec(target)
		if !ok {
			add(Error, s.ID, "supersedes %s, which does not exist", target)
			continue
		}
		if other.Status != workflow.Done {
			add(Error, s.ID, "supersedes %s, which is not done", target)
		}
	}
	if cycle := supersedeCycle(p, s); cycle != "" {
		add(Error, s.ID, "supersedes cycle: %s", cycle)
	}
	if workflow.Terminal(s.Status) {
		return
	}
	// Only a live spec can claim a replacement: history may hold several
	// successors over time, so terminal specs are not counted here.
	for _, target := range s.Supersedes {
		for _, other := range p.Specs {
			if other.ID == s.ID || workflow.Terminal(other.Status) {
				continue
			}
			if hasID(other.Supersedes, target) {
				add(Error, s.ID, "supersedes %s, which %s also supersedes", target, other.ID)
			}
		}
	}
}

func checkRelations(p *project.Project, s *project.Spec, add func(Severity, string, string, ...any)) {
	if s.Parent != "" {
		if _, ok := p.Spec(s.Parent); !ok {
			add(Error, s.ID, "parent %s does not exist", s.Parent)
		} else if cycle := parentCycle(p, s); cycle != "" {
			add(Error, s.ID, "parent cycle: %s", cycle)
		}
	}
	if len(s.Covers) > 0 && s.Parent == "" {
		add(Error, s.ID, "declares covers but has no parent")
	}
	if parent, ok := p.Spec(s.Parent); ok {
		known := map[string]bool{}
		for _, c := range parent.Criteria() {
			known[c.ID] = true
		}
		for _, ac := range s.Covers {
			if !known[ac] {
				add(Error, s.ID, "covers %s, which %s does not declare", ac, parent.ID)
			}
		}
	}
	for _, d := range s.Deps {
		if d.Level != "" && d.Level != "contract" {
			add(Error, s.ID, "dependency %s has unknown level %q; use @contract or nothing",
				d.ID, d.Level)
		}
		if _, ok := p.Spec(d.ID); !ok {
			add(Error, s.ID, "depends on %s, which does not exist", d.ID)
		}
	}
	if cycle := depCycle(p, s); cycle != "" {
		add(Error, s.ID, "dependency cycle: %s", cycle)
	}
}

func checkArtifacts(p *project.Project, s *project.Spec, add func(Severity, string, string, ...any)) {
	dir := s.Dir()
	exists := func(name string) bool {
		_, err := os.Stat(filepath.Join(dir, name))
		return err == nil
	}
	rel := func(path string) string {
		if r, err := filepath.Rel(p.Root, path); err == nil {
			return filepath.ToSlash(r)
		}
		return filepath.ToSlash(path)
	}
	switch s.Status {
	case workflow.AwaitingApproval, workflow.Planning, workflow.Implementing, workflow.Reviewing:
		if strings.TrimSpace(s.Contract()) == "" {
			add(Error, s.ID, "is %s with an empty Contract section", s.Status)
		}
	}
	switch s.Status {
	case workflow.Implementing:
		if !exists("plan.md") {
			add(Error, s.ID, "is implementing without %s", rel(s.PlanPath()))
		} else if !planSurveysExisting(dir) {
			add(Warning, s.ID, "plan.md has no Existing state section; name what to "+
				"reuse before building")
		}
		if !exists("tasks.md") {
			add(Error, s.ID, "is implementing without %s", rel(s.TasksPath()))
		}
	case workflow.Reviewing:
		if !exists("review.md") {
			add(Error, s.ID, "is reviewing without %s", rel(s.ReviewPath()))
		}
	case workflow.Done:
		if !exists("review.md") {
			add(Error, s.ID, "is done without %s", rel(s.ReviewPath()))
		}
		if s.ApprovedBy == "" {
			add(Error, s.ID, "is done without an approved contract")
		}
	}
}

// planSurveysExisting reports whether the plan recorded what already exists
// and can be reused. Only a warning: guidance, not a gate.
func planSurveysExisting(dir string) bool {
	d, err := doc.Load(filepath.Join(dir, "plan.md"))
	if err != nil {
		return true
	}
	return strings.TrimSpace(d.Section("Existing state")) != ""
}

func checkCoverage(p *project.Project, s *project.Spec, add func(Severity, string, string, ...any)) {
	children := p.Children(s.ID)
	if len(children) == 0 || s.Status == workflow.Dropped {
		return
	}
	anyDone := false
	for _, c := range children {
		if c.Status == workflow.Done {
			anyDone = true
		}
	}
	for _, row := range p.Coverage(s) {
		if len(row.By) > 0 {
			continue
		}
		sev := Warning
		if anyDone {
			sev = Error
		}
		add(sev, s.ID, "%s is not covered by any child: %s", row.Criterion.ID, row.Criterion.Text)
	}
	if s.Status == workflow.Done {
		for _, row := range p.Coverage(s) {
			for _, child := range row.By {
				if child.Status != workflow.Done {
					add(Error, s.ID, "is done but %s covering %s is %s",
						child.ID, row.Criterion.ID, child.Status)
				}
			}
		}
	}
}

func drift(p *project.Project) []string {
	var out []string
	for _, s := range p.Specs {
		if s.ContractChanged() {
			out = append(out, fmt.Sprintf(
				"%s: the contract changed after %s approved it; re-approve it so whoever "+
					"depends on it is told", s.ID, s.ApprovedBy))
		}
		if workflow.Terminal(s.Status) {
			continue
		}
		for dep, hash := range s.Agreed {
			other, ok := p.Spec(dep)
			if !ok || other.ContractHash == "" || other.ContractHash == hash {
				continue
			}
			out = append(out, fmt.Sprintf(
				"%s builds against %s's contract %s, but it is now %s; realign and re-approve",
				s.ID, dep, hash, other.ContractHash))
		}
	}
	return out
}

func parentCycle(p *project.Project, s *project.Spec) string {
	seen := map[string]bool{s.ID: true}
	path := []string{s.ID}
	cur := s
	for cur.Parent != "" {
		next, ok := p.Spec(cur.Parent)
		if !ok {
			return ""
		}
		path = append(path, next.ID)
		if seen[next.ID] {
			return strings.Join(path, " -> ")
		}
		seen[next.ID] = true
		cur = next
	}
	return ""
}

func depCycle(p *project.Project, start *project.Spec) string {
	var path []string
	visiting := map[string]bool{}
	var walk func(*project.Spec) bool
	walk = func(s *project.Spec) bool {
		if visiting[s.ID] {
			path = append(path, s.ID)
			return true
		}
		visiting[s.ID] = true
		path = append(path, s.ID)
		for _, d := range s.Deps {
			next, ok := p.Spec(d.ID)
			if !ok {
				continue
			}
			if walk(next) {
				return true
			}
		}
		visiting[s.ID] = false
		path = path[:len(path)-1]
		return false
	}
	if walk(start) {
		return strings.Join(path, " -> ")
	}
	return ""
}

// supersedeCycle walks the supersede links from start and returns the chain
// when it reaches a spec already on the path, mirroring depCycle.
func supersedeCycle(p *project.Project, start *project.Spec) string {
	var path []string
	visiting := map[string]bool{}
	var walk func(*project.Spec) bool
	walk = func(s *project.Spec) bool {
		if visiting[s.ID] {
			path = append(path, s.ID)
			return true
		}
		visiting[s.ID] = true
		path = append(path, s.ID)
		for _, target := range s.Supersedes {
			next, ok := p.Spec(target)
			if !ok {
				continue
			}
			if walk(next) {
				return true
			}
		}
		visiting[s.ID] = false
		path = path[:len(path)-1]
		return false
	}
	if walk(start) {
		return strings.Join(path, " -> ")
	}
	return ""
}

func hasID(ids []string, id string) bool {
	for _, v := range ids {
		if v == id {
			return true
		}
	}
	return false
}

// HasErrors reports whether any finding blocks the build.
func HasErrors(findings []Finding) bool {
	for _, f := range findings {
		if f.Severity == Error {
			return true
		}
	}
	return false
}

func sortFindings(f []Finding) {
	for i := 1; i < len(f); i++ {
		for j := i; j > 0 && less(f[j], f[j-1]); j-- {
			f[j], f[j-1] = f[j-1], f[j]
		}
	}
}

func less(a, b Finding) bool {
	if a.Severity != b.Severity {
		return a.Severity > b.Severity
	}
	if a.Spec != b.Spec {
		return a.Spec < b.Spec
	}
	return a.Message < b.Message
}
