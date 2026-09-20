// Package workflow holds the spec lifecycle: the states and the legal
// transitions between them. It has no notion of who may make a move;
// Forge records who did, the same way git records an author, without
// deciding whether they were allowed to.
package workflow

import (
	"fmt"
	"sort"
	"strings"
)

// State is a point in the life of a spec.
type State string

const (
	Proposed     State = "proposed"
	Accepted     State = "accepted"
	Contracting  State = "contracting"
	Planning     State = "planning"
	Implementing State = "implementing"
	Blocked      State = "blocked"
	Reviewing    State = "reviewing"
	Done         State = "done"
	Dropped      State = "dropped"
)

// All lists every state in lifecycle order.
func All() []State {
	return []State{Proposed, Accepted, Contracting, Planning,
		Implementing, Blocked, Reviewing, Done, Dropped}
}

// Canonical maps a retired state name to its current one, so a repository
// written before the rename keeps loading and every command reports the new
// name. Every other state is returned unchanged; `forge migrate` is what
// converges the written tree.
func Canonical(s State) State {
	switch s {
	case "specifying", "awaiting-approval":
		return Contracting
	}
	return s
}

// Meaning returns the one-line explanation of a state, the normative copy
// `forge workflow` renders into its table.
func Meaning(s State) string {
	switch s {
	case Proposed:
		return "Written, not in the queue"
	case Accepted:
		return "In the queue"
	case Contracting:
		return "The contract is being written"
	case Planning:
		return "Splitting into phases"
	case Implementing:
		return "Product code, phase by phase"
	case Blocked:
		return "Cannot continue"
	case Reviewing:
		return "Verifying acceptance criteria"
	case Done:
		return "Archived and closed"
	case Dropped:
		return "Will not be done"
	}
	return ""
}

// Valid reports whether s is a known state.
func Valid(s State) bool {
	for _, k := range All() {
		if k == s {
			return true
		}
	}
	return false
}

// Terminal reports whether nothing follows this state.
func Terminal(s State) bool { return s == Done || s == Dropped }

// Transition describes one legal move.
type Transition struct {
	From State
	To   State
	// Why explains the move in the history line and in help output.
	Why string
}

var transitions = []Transition{
	{Proposed, Accepted, "accepted into the queue"},
	{Proposed, Dropped, "the work will not be done"},
	{Accepted, Contracting, "someone starts the work"},
	{Accepted, Dropped, "the work will not be done"},
	{Contracting, Planning, "the contract is approved"},
	{Contracting, Dropped, "the work will not be done"},
	{Planning, Implementing, "the plan is split into phases"},
	{Implementing, Reviewing, "every phase is finished"},
	{Implementing, Blocked, "the work cannot continue"},
	{Blocked, Implementing, "the blocker is gone"},
	{Blocked, Dropped, "the work will not be done"},
	{Reviewing, Implementing, "the review found failures"},
	{Reviewing, Done, "the review passed and the spec is archived"},
}

// Find returns the transition between two states.
func Find(from, to State) (Transition, bool) {
	for _, t := range transitions {
		if t.From == from && t.To == to {
			return t, true
		}
	}
	return Transition{}, false
}

// Next lists the states reachable from s.
func Next(s State) []State {
	var out []State
	for _, t := range transitions {
		if t.From == s {
			out = append(out, t.To)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

// Check validates a move and returns a readable error when it is not legal.
func Check(from, to State) error {
	if !Valid(from) {
		return fmt.Errorf("unknown state %q", from)
	}
	if !Valid(to) {
		return fmt.Errorf("unknown state %q, expected one of %s", to, join(All()))
	}
	if from == to {
		return fmt.Errorf("already in %s", from)
	}
	if _, ok := Find(from, to); !ok {
		if n := Next(from); len(n) == 0 {
			return fmt.Errorf("%s is final; nothing follows it", from)
		} else {
			return fmt.Errorf("cannot go from %s to %s; from %s you can go to %s",
				from, to, from, join(n))
		}
	}
	return nil
}

// InFlight reports whether work on the spec has started and not finished.
func InFlight(s State) bool {
	switch s {
	case Contracting, Planning, Implementing, Blocked, Reviewing:
		return true
	}
	return false
}

// WaitingFor names who has to act next, for the board and the session brief.
func WaitingFor(s State) string {
	switch s {
	case Proposed:
		return "anyone: forge accept or drop"
	case Accepted:
		return "anyone: forge start"
	case Contracting:
		return "architect: write the contract"
	case Planning:
		return "orchestrator: split into phases"
	case Implementing:
		return "implementer: next phase"
	case Blocked:
		return "orchestrator: clear the blocker"
	case Reviewing:
		return "reviewer: verify acceptance criteria"
	case Done:
		return "nobody: closed"
	case Dropped:
		return "nobody: dropped"
	}
	return "unknown"
}

func join(states []State) string {
	parts := make([]string, len(states))
	for i, s := range states {
		parts[i] = string(s)
	}
	return strings.Join(parts, ", ")
}
