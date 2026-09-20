package cli

import (
	"strings"
	"testing"
)

// A workflow that lost the marker must fail loudly instead of printing a
// page whose states table silently vanished.
func TestRenderWorkflow_MissingMarkerIsAnError(t *testing.T) {
	_, err := renderWorkflow([]byte("# Workflow\n\n## States\n\nno marker here\n"))
	if err == nil {
		t.Fatal("a workflow without the marker should fail")
	}
	if !strings.Contains(err.Error(), "kit/machine/WORKFLOW.md") {
		t.Errorf("the error should name the file, got: %v", err)
	}
}
