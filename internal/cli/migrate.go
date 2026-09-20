package cli

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/TheJisus28/forge/internal/project"
	"github.com/TheJisus28/forge/internal/workflow"
)

// cmdMigrate converges the written tree onto the current state names. A spec
// whose raw frontmatter `status` is a retired name (`specifying` or
// `awaiting-approval`) has that scalar rewritten to `contracting`; the body
// is left byte-identical, including `## History`, because a rename is not a
// state move and history is never rewritten (decision 0004). A tree with
// nothing retired prints `nothing to migrate` and succeeds; `--dry-run`
// reports the same list and writes nothing (SPEC-015, decision 6).
func cmdMigrate(args []string, out io.Writer) error {
	fs := newFlagSet("migrate", "usage: forge migrate [--dry-run]", out)
	dryRun := fs.Bool("dry-run", false, "list the retired names to rewrite and write nothing")
	if _, err := parseArgs(fs, args); err != nil {
		return err
	}
	p, err := project.Load(cwd())
	if err != nil {
		return err
	}

	migrated := 0
	for _, s := range p.Specs {
		raw := workflow.State(s.Doc().Str("status"))
		canonical := workflow.Canonical(raw)
		if canonical == raw {
			continue
		}
		rel, err := filepath.Rel(p.Root, s.Path)
		if err != nil {
			rel = s.Path
		}
		fmt.Fprintf(out, "%s: %s -> %s\n", filepath.ToSlash(rel), raw, canonical)
		migrated++
		if *dryRun {
			continue
		}
		s.Doc().SetStr("status", string(canonical))
		if err := s.Doc().Save(s.Path); err != nil {
			return err
		}
	}
	if migrated == 0 {
		fmt.Fprintln(out, "nothing to migrate")
	}
	return nil
}
