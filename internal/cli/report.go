package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/TheJisus28/forge/internal/project"
	"github.com/TheJisus28/forge/internal/validate"
	"github.com/TheJisus28/forge/internal/view"
)

func cmdStatus(args []string, out io.Writer) error {
	fs := newFlagSet("status", "usage: forge status [id] [--fetch]", out)
	fetch := fs.Bool("fetch", false, "run git fetch first, so the picture includes the remote")
	rest, err := parseArgs(fs, args)
	if err != nil {
		return err
	}
	p, err := project.Load(cwd())
	if err != nil {
		return err
	}
	if *fetch {
		if err := project.Fetch(p.Root); err != nil {
			return fmt.Errorf("git fetch: %w", err)
		}
		if p, err = project.Load(p.Root); err != nil {
			return err
		}
	}
	if len(rest) > 0 {
		id := project.NormalizeID(rest[0])
		s, ok := p.Spec(id)
		if !ok {
			return fmt.Errorf("%s does not exist", id)
		}
		fmt.Fprint(out, view.Detail(p, s))
		return nil
	}
	fmt.Fprint(out, view.Status(p))
	return nil
}

func cmdBrief(args []string, out io.Writer) error {
	fs := newFlagSet("brief", "usage: forge brief [--json]", out)
	asJSON := fs.Bool("json", false, "emit the Claude Code SessionStart hook payload")
	_, err := parseArgs(fs, args)
	if err != nil {
		return err
	}
	p, err := project.Load(cwd())
	if err != nil {
		// A hook runs everywhere, including repositories without Forge.
		if *asJSON {
			return nil
		}
		return err
	}
	text := view.Brief(p)
	if !*asJSON {
		fmt.Fprint(out, text)
		return nil
	}
	payload := map[string]any{
		"hookSpecificOutput": map[string]any{
			"hookEventName":     "SessionStart",
			"additionalContext": text,
		},
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	fmt.Fprintln(out, string(data))
	return nil
}

func cmdBoard(args []string, out io.Writer) error {
	fs := newFlagSet("board", "usage: forge board [--print]", out)
	print := fs.Bool("print", false, "write to stdout instead of .forge/BOARD.md")
	_, err := parseArgs(fs, args)
	if err != nil {
		return err
	}
	p, err := project.Load(cwd())
	if err != nil {
		return err
	}
	body := view.Board(p)
	if *print {
		fmt.Fprint(out, body)
		return nil
	}
	path := filepath.Join(p.Root, project.Dir, "BOARD.md")
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		return err
	}
	fmt.Fprintf(out, "%s regenerated (it is gitignored on purpose)\n", filepath.ToSlash(
		project.Dir+"/BOARD.md"))
	return nil
}

func cmdValidate(args []string, out, errOut io.Writer) int {
	fs := newFlagSet("validate", "usage: forge validate [--approvers \"ana,jose\"]", out)
	approvers := fs.String("approvers", "",
		"who really approved the pull request; CI passes this in")
	quiet := fs.Bool("quiet", false, "print only errors")
	_, err := parseArgs(fs, args)
	if err != nil {
		return 0
	}
	p, err := project.Load(cwd())
	if err != nil {
		fmt.Fprintf(errOut, "forge: %v\n", err)
		return 1
	}
	opt := validate.Options{}
	for _, a := range strings.Split(*approvers, ",") {
		if a = strings.TrimSpace(a); a != "" {
			opt.Approvers = append(opt.Approvers, a)
		}
	}
	findings := validate.Run(p, opt)
	shown := 0
	for _, f := range findings {
		if *quiet && f.Severity != validate.Error {
			continue
		}
		fmt.Fprintln(out, f)
		shown++
	}
	if validate.HasErrors(findings) {
		return 1
	}
	if shown == 0 {
		noun := "specs"
		if len(p.Specs) == 1 {
			noun = "spec"
		}
		fmt.Fprintf(out, "%d %s, no problems\n", len(p.Specs), noun)
	}
	return 0
}

func cmdSync(args []string, out io.Writer) error {
	fs := newFlagSet("sync", "usage: forge sync [id]", out)
	rest, err := parseArgs(fs, args)
	if err != nil {
		return err
	}
	if !project.HasGH() {
		return fmt.Errorf("sync needs the GitHub CLI (gh). Everything else works without it")
	}
	p, s, err := specArg(rest)
	if err != nil {
		return err
	}
	raw, err := project.GH(p.Root, "pr", "view", "--json", "number,state,url")
	if err != nil {
		return fmt.Errorf("gh could not read a pull request for this branch: %w", err)
	}
	var pr struct {
		Number int    `json:"number"`
		State  string `json:"state"`
		URL    string `json:"url"`
	}
	if err := json.Unmarshal([]byte(raw), &pr); err != nil {
		return err
	}
	s.PR = fmt.Sprintf("%d", pr.Number)
	s.Doc().SetStr("pr_url", pr.URL)
	s.Doc().SetStr("pr_state", strings.ToLower(pr.State))
	if err := s.Save(); err != nil {
		return err
	}
	fmt.Fprintf(out, "%s linked to pull request #%d (%s)\n", s.ID, pr.Number, strings.ToLower(pr.State))
	return nil
}
