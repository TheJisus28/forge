package workflow_test

import (
	"strings"
	"testing"

	"github.com/TheJisus28/forge/internal/workflow"
)

func TestCheck_LegalAndIllegalMoves(t *testing.T) {
	legal := [][2]workflow.State{
		{workflow.Proposed, workflow.Accepted},
		{workflow.Accepted, workflow.Contracting},
		{workflow.Contracting, workflow.Planning},
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
		{workflow.Contracting, workflow.Implementing},
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
	err := workflow.Check(workflow.Contracting, workflow.Implementing)
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

// The state after accepted is named for the contract it produces, and the
// retired name is not a second state in the machine (SPEC-015, decision 1).
func TestAll_UsesContracting(t *testing.T) {
	want := []workflow.State{
		workflow.Proposed, workflow.Accepted, workflow.Contracting,
		workflow.Planning, workflow.Implementing, workflow.Blocked,
		workflow.Reviewing, workflow.Done, workflow.Dropped,
	}
	got := workflow.All()
	if len(got) != len(want) {
		t.Fatalf("All() = %v, want %d states", got, len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("All()[%d] = %s, want %s", i, got[i], want[i])
		}
	}
	for _, s := range got {
		if s == "specifying" || s == "awaiting-approval" {
			t.Errorf("All() still carries the retired state %q", s)
		}
	}
	if strings.TrimSpace(workflow.Meaning(workflow.Contracting)) == "" {
		t.Error("Meaning(Contracting) is empty")
	}
	if got := workflow.WaitingFor(workflow.Contracting); got != "architect: write the contract" {
		t.Errorf("WaitingFor(Contracting) = %q, want the architect action", got)
	}
	if !workflow.InFlight(workflow.Contracting) {
		t.Error("contracting is in flight")
	}
}

// The table in `forge workflow` has one source, so every state must carry a
// meaning and the blocked line must name a role `forge roles` lists.
func TestMeaning_EveryStateHasALine(t *testing.T) {
	for _, s := range workflow.All() {
		if strings.TrimSpace(workflow.Meaning(s)) == "" {
			t.Errorf("Meaning(%s) is empty", s)
		}
	}
}

func TestWaitingFor_BlockedNamesTheOrchestrator(t *testing.T) {
	got := workflow.WaitingFor(workflow.Blocked)
	if !strings.HasPrefix(got, "orchestrator:") {
		t.Errorf("WaitingFor(blocked) = %q, want an orchestrator action", got)
	}
	if strings.Contains(got, "conductor") {
		t.Errorf("WaitingFor(blocked) still says conductor: %q", got)
	}
}

func TestInFlightAndTerminal(t *testing.T) {
	if workflow.InFlight(workflow.Proposed) || workflow.InFlight(workflow.Done) {
		t.Error("proposed and done are not in flight")
	}
	if !workflow.InFlight(workflow.Contracting) {
		t.Error("contracting is in flight")
	}
	if !workflow.InFlight(workflow.Implementing) {
		t.Error("implementing is in flight")
	}
	if !workflow.Terminal(workflow.Done) || !workflow.Terminal(workflow.Dropped) {
		t.Error("done and dropped are terminal")
	}
}
