package view_test

import (
	"strings"
	"testing"

	"github.com/TheJisus28/forge/internal/doc"
	"github.com/TheJisus28/forge/internal/project"
	"github.com/TheJisus28/forge/internal/view"
)

// doneProject builds an onboarded project with n specs, all done.
func doneProject(t *testing.T, n int) *project.Project {
	t.Helper()
	cfg := doc.New()
	cfg.SetStr("test", "go test ./...")
	p := &project.Project{Root: t.TempDir(), Config: cfg}
	for i := 1; i <= n; i++ {
		d := doc.New()
		d.SetStr("id", project.FormatID(i))
		d.SetStr("title", "Spec "+project.FormatID(i))
		d.SetStr("status", "done")
		s, err := project.FromDoc(project.FormatID(i)+".md", d)
		if err != nil {
			t.Fatal(err)
		}
		p.Specs = append(p.Specs, s)
	}
	return p
}

func TestBrief_ListsEveryDeliveredSpecWhenFew(t *testing.T) {
	brief := view.Brief(doneProject(t, 3))
	for i := 1; i <= 3; i++ {
		if !strings.Contains(brief, project.FormatID(i)) {
			t.Errorf("brief should list %s:\n%s", project.FormatID(i), brief)
		}
	}
	if strings.Contains(brief, "not listed") {
		t.Errorf("nothing is hidden with three specs:\n%s", brief)
	}
}

// The brief must not grow with the number of closed specs: it shows the
// most recent ones and points at the command that lists the rest.
func TestBrief_CapsDeliveredAndPointsAtStatus(t *testing.T) {
	brief := view.Brief(doneProject(t, 8))

	for i := 4; i <= 8; i++ {
		if !strings.Contains(brief, project.FormatID(i)) {
			t.Errorf("brief should list recent %s:\n%s", project.FormatID(i), brief)
		}
	}
	if strings.Contains(brief, project.FormatID(1)) {
		t.Errorf("the oldest spec should be hidden:\n%s", brief)
	}
	if !strings.Contains(brief, "3 earlier") || !strings.Contains(brief, "forge status") {
		t.Errorf("brief should count the hidden specs and point at status:\n%s", brief)
	}
}

// The brief summarises the current shape, one line per capability, or
// nothing at all when no done spec declares one.
func TestBrief_ShowsCapabilitySummary(t *testing.T) {
	p := doneProject(t, 2)
	if brief := view.Brief(p); strings.Contains(brief, "capabilities:") {
		t.Errorf("with no capabilities the brief should stay silent:\n%s", brief)
	}

	p.Specs[0].Capability = "workflow"
	p.Specs[1].Capability = "workflow"
	p.Specs[1].Supersedes = []string{"SPEC-001"}
	brief := view.Brief(p)
	if !strings.Contains(brief, "capabilities:") || !strings.Contains(brief, "workflow") {
		t.Fatalf("the brief should summarise the capabilities:\n%s", brief)
	}
	if !strings.Contains(brief, "1 current contract") {
		t.Errorf("one of the two workflow contracts is superseded:\n%s", brief)
	}
}

// The detail surfaces the capability right below the status, so an agent
// reads which part of the system it touches before anything else.
func TestDetail_ShowsCapability(t *testing.T) {
	p := doneProject(t, 1)
	p.Specs[0].Capability = "guard"

	out := view.Detail(p, p.Specs[0])
	if !strings.Contains(out, "capability  guard") {
		t.Errorf("the detail should show the capability:\n%s", out)
	}
}

// A spec from before the field existed still reads as a person would expect:
// an explicit "(none)" rather than an empty label.
func TestDetail_ShowsNone(t *testing.T) {
	p := doneProject(t, 1)

	out := view.Detail(p, p.Specs[0])
	if !strings.Contains(out, "capability  (none)") {
		t.Errorf("an empty capability should render as (none):\n%s", out)
	}
}
