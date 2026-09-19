// Package workflow holds the spec lifecycle: the states, the transitions
// between them and which ones need a human.
package workflow

import (
	"fmt"
	"sort"
	"strings"
)

// State is a point in the life of a spec.
type State string

const (
	Proposed         State = "proposed"
	Accepted         State = "accepted"
	Specifying       State = "specifying"
	AwaitingApproval State = "awaiting-approval"
	Planning         State = "planning"
	Implementing     State = "implementing"
	Blocked          State = "blocked"
	Reviewing        State = "reviewing"
	Done             State = "done"
	Dropped          State = "dropped"
)

// All lists every state in lifecycle order.
func All() []State {
	return []State{Proposed, Accepted, Specifying, AwaitingApproval, Planning,
		Implementing, Blocked, Reviewing, Done, Dropped}
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
	// Gate marks a move only a maintainer may make.
	Gate bool
	// Why explains the move in the history line and in help output.
	Why string
}

var transitions = []Transition{
	{Proposed, Accepted, true, "a maintainer accepts the work into the queue"},
	{Proposed, Dropped, false, "the work will not be done"},
	{Accepted, Specifying, false, "someone starts the work"},
	{Accepted, Dropped, false, "the work will not be done"},
	{Specifying, AwaitingApproval, false, "the contract is ready for review"},
	{AwaitingApproval, Specifying, false, "changes were requested"},
	{AwaitingApproval, Planning, true, "a maintainer approves the contract"},
	{AwaitingApproval, Dropped, false, "the work will not be done"},
	{Planning, Implementing, false, "the plan is split into phases"},
	{Implementing, Reviewing, false, "every phase is finished"},
	{Implementing, Blocked, false, "the work cannot continue"},
	{Blocked, Implementing, false, "the blocker is gone"},
	{Blocked, Dropped, false, "the work will not be done"},
	{Reviewing, Implementing, false, "the review found failures"},
	{Reviewing, Done, false, "the review passed and the spec is archived"},
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

// NeedsMaintainer reports whether the move is a human gate.
func NeedsMaintainer(from, to State) bool {
	t, ok := Find(from, to)
	return ok && t.Gate
}

// InFlight reports whether work on the spec has started and not finished.
func InFlight(s State) bool {
	switch s {
	case Specifying, AwaitingApproval, Planning, Implementing, Blocked, Reviewing:
		return true
	}
	return false
}

// WaitingFor names who has to act next, for the board and the session brief.
func WaitingFor(s State) string {
	switch s {
	case Proposed:
		return "maintainer: accept or drop"
	case Accepted:
		return "anyone: forge start"
	case Specifying:
		return "architect: write the contract"
	case AwaitingApproval:
		return "maintainer: approve the contract"
	case Planning:
		return "orchestrator: split into phases"
	case Implementing:
		return "implementer: next phase"
	case Blocked:
		return "conductor: clear the blocker"
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
