package kit_test

import (
	"strings"
	"testing"

	"github.com/TheJisus28/forge/kit"
)

func TestMachine_ServesWorkflowRolesAndTemplates(t *testing.T) {
	workflow, err := kit.Workflow()
	if err != nil {
		t.Fatalf("Workflow: %v", err)
	}
	if !strings.Contains(string(workflow), "# Workflow") {
		t.Errorf("Workflow did not return the workflow:\n%s", workflow)
	}

	got := strings.Join(kit.Roles(), ",")
	want := "architect,implementer,orchestrator,reviewer"
	if got != want {
		t.Errorf("Roles() = %q, want %q", got, want)
	}

	if _, err := kit.Role("architect"); err != nil {
		t.Errorf("Role(architect): %v", err)
	}
	if _, err := kit.Role("architect.md"); err != nil {
		t.Errorf("Role should accept the extension: %v", err)
	}
	if _, err := kit.Role("nope"); err == nil {
		t.Error("Role(nope) should fail")
	}

	if _, err := kit.Template("spec"); err != nil {
		t.Errorf("Template(spec): %v", err)
	}
	if _, err := kit.Template("spec.md"); err != nil {
		t.Errorf("Template should accept the extension: %v", err)
	}
	if _, err := kit.Template("nope"); err == nil {
		t.Error("Template(nope) should fail")
	}
}
