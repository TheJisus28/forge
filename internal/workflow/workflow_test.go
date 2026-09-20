package workflow_test

import (
	"strings"
	"testing"

	"github.com/TheJisus28/forge/internal/workflow"
)

func TestCheck_LegalAndIllegalMoves(t *testing.T) {
	legal := [][2]workflow.State{
		{workflow.Proposed, workflow.Accepted},
		{workflow.Accepted, workflow.Specifying},
		{workflow.Specifying, workflow.AwaitingApproval},
		{workflow.AwaitingApproval, workflow.Planning},
		{workflow.AwaitingApproval, workflow.Specifying},
		{workflow.Planning, workflow.Implementing},
		{workflow.Implementing, workflow.Reviewing},
		{workflow.Reviewing, workflow.Implementing},
		{workflow.Reviewing, workflow.Done},
	}
	for _, m := range legal {
		if err := workflow.Check(m[0], m[1]); err != nil {
			t.Errorf("%s -> %s should be legal: %v", m[0], m[1], err)
		}
	}

	illegal := [][2]workflow.State{
		{workflow.Proposed, workflow.Implementing},
		{workflow.Accepted, workflow.Done},
		{workflow.Specifying, workflow.Implementing},
		{workflow.Done, workflow.Implementing},
		{workflow.Dropped, workflow.Accepted},
	}
	for _, m := range illegal {
		if err := workflow.Check(m[0], m[1]); err == nil {
			t.Errorf("%s -> %s should be rejected", m[0], m[1])
		}
	}
}

// Skipping the contract approval is the move the whole workflow exists to
// prevent, so it gets its own test.
func TestCheck_CannotSkipApproval(t *testing.T) {
	err := workflow.Check(workflow.AwaitingApproval, workflow.Implementing)
	if err == nil {
		t.Fatal("a spec must not reach implementing without passing through planning")
	}
	if !strings.Contains(err.Error(), "planning") {
		t.Errorf("the error should say where you can go: %v", err)
	}
}

func TestCheck_SameState(t *testing.T) {
	if err := workflow.Check(workflow.Implementing, workflow.Implementing); err == nil {
		t.Fatal("moving to the same state should be rejected")
	}
	if err := workflow.Check(workflow.Implementing, "shipping"); err == nil {
		t.Fatal("unknown states should be rejected")
	}
}

func TestInFlightAndTerminal(t *testing.T) {
	if workflow.InFlight(workflow.Proposed) || workflow.InFlight(workflow.Done) {
		t.Error("proposed and done are not in flight")
	}
	if !workflow.InFlight(workflow.Implementing) {
		t.Error("implementing is in flight")
	}
	if !workflow.Terminal(workflow.Done) || !workflow.Terminal(workflow.Dropped) {
		t.Error("done and dropped are terminal")
	}
}
