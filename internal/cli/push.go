package cli

import (
	"fmt"
	"io"

	"github.com/TheJisus28/forge/internal/project"
)

// cmdPush checkpoints a spec: it commits the pending work and publishes the
// branch, so the work survives leaving the machine and a handoff is possible
// (SPEC-020, decision 1).
func cmdPush(args []string, out io.Writer) error {
	fs := newFlagSet("push", "usage: forge push [id]", out)
	rest, err := parseArgs(fs, args)
	if err != nil {
		return err
	}
	p, s, err := specArg(rest)
	if err != nil {
		return err
	}
	if project.OnDefaultBranch(p.Root) {
		return fmt.Errorf("refusing to push the default branch; a spec ends as a pull request, run forge submit")
	}
	if branch := project.Branch(p.Root); branch == "" {
		return fmt.Errorf("not on a git branch; forge push publishes a spec branch")
	}
	return checkpoint(p, s, out)
}

// checkpoint commits the pending work and pushes the spec's branch. It reports
// one line: what it pushed, or that there was nothing to publish (SPEC-020,
// decisions 2, 3 and 5).
func checkpoint(p *project.Project, s *project.Spec, out io.Writer) error {
	branch := project.Branch(p.Root)
	committed, err := project.CommitAll(p.Root, checkpointMessage(s))
	if err != nil {
		return err
	}
	ahead, err := project.HasUnpushed(p.Root)
	if err != nil {
		return err
	}
	if !committed && !ahead {
		fmt.Fprintf(out, "nothing to push: %s is up to date\n", branch)
		return nil
	}
	if err := project.Push(p.Root, branch); err != nil {
		return err
	}
	fmt.Fprintf(out, "pushed %s to origin\n", branch)
	return nil
}

// checkpointMessage is the Conventional Commit subject a checkpoint writes: the
// spec and the phase the CLI can name without judgement (SPEC-020, decision 3).
func checkpointMessage(s *project.Spec) string {
	return fmt.Sprintf("chore(%s): checkpoint %s", s.ID, s.Status)
}
