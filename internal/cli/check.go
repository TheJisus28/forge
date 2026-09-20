package cli

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/TheJisus28/forge/internal/project"
	"github.com/TheJisus28/forge/internal/workflow"
)

// cmdCheck reports the acceptance criteria no task in tasks.md delivers and no
// evidence line in review.md settles, for the specs being built or reviewed. It
// loads .forge, reads the spec artifacts and writes nothing: no Save, no git
// and no network, so the tree and `git status` are unchanged. Without an id it
// checks every spec in flight or delivered, in id order; with one it checks
// that spec whatever its state. A gap that leaves a criterion unsettled while
// the review is the record (reviewing or done) fails, so the exit code is what
// a reviewer reads.
func cmdCheck(args []string, out, errOut io.Writer) int {
	fs := newFlagSet("check", "usage: forge check [id]", out)
	rest, err := parseArgs(fs, args)
	if err != nil {
		return 0
	}
	p, err := project.Load(cwd())
	if err != nil {
		fmt.Fprintf(errOut, "forge: %v\n", err)
		return 1
	}

	var specs []*project.Spec
	if len(rest) > 0 {
		id := project.NormalizeID(rest[0])
		s, ok := p.Spec(id)
		if !ok {
			fmt.Fprintf(errOut, "forge: %s does not exist; forge status lists what does\n", id)
			return 1
		}
		specs = append(specs, s)
	} else {
		for _, s := range p.Specs {
			switch s.Status {
			case workflow.Implementing, workflow.Blocked, workflow.Reviewing, workflow.Done:
				specs = append(specs, s)
			}
		}
	}

	if len(specs) == 0 {
		fmt.Fprintln(out, "no spec is building or reviewing yet; nothing to check")
		return 0
	}

	reported := false
	failed := false
	withCriteria := 0
	for _, s := range specs {
		if len(s.Criteria()) > 0 {
			withCriteria++
		}
		for _, g := range s.CriterionGaps() {
			reported = true
			rel := g.File
			if r, err := filepath.Rel(p.Root, g.File); err == nil {
				rel = r
			}
			fmt.Fprintf(out, "%s: %s has no %s in %s\n",
				s.ID, g.Criterion.ID, g.Kind, filepath.ToSlash(rel))
			if g.Kind == "evidence" &&
				(s.Status == workflow.Reviewing || s.Status == workflow.Done) {
				failed = true
			}
		}
	}
	if failed {
		return 1
	}
	if !reported && withCriteria > 0 {
		fmt.Fprintf(out, "%d specs checked, every criterion is covered\n", len(specs))
	}
	return 0
}
