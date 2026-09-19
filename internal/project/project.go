// Package project loads a repository's .forge directory: the project
// configuration and every spec, with the relationships between them.
package project

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/TheJisus28/forge/internal/doc"
	"github.com/TheJisus28/forge/internal/workflow"
)

// Dir is the directory Forge plants in a repository.
const Dir = ".forge"

// ErrNotInitialized is returned when no .forge directory is found.
var ErrNotInitialized = errors.New("no .forge directory found; run forge init")

// Dep is a dependency on another spec. Level "contract" is satisfied as soon
// as the other spec's contract is approved; the default waits for done.
type Dep struct {
	ID    string
	Level string // "" or "contract"
}

// String renders the dependency the way it is written in frontmatter.
func (d Dep) String() string {
	if d.Level == "" {
		return d.ID
	}
	return d.ID + "@" + d.Level
}

// Criterion is one acceptance criterion of a spec.
type Criterion struct {
	ID   string // AC1
	Text string
}

// Spec is one unit of work, in any state from proposed to done.
type Spec struct {
	ID       string
	Num      int
	Title    string
	Status   workflow.State
	Parent   string
	Covers   []string
	Deps     []Dep
	Needs    []string
	External []map[string]string
	Path     string

	Conductor    string
	AcceptedBy   string
	ApprovedBy   string
	ContractHash string
	Agreed       map[string]string // spec ID -> contract hash this spec started against
	PR           string

	doc *doc.Doc
}

// Doc exposes the underlying document for commands that edit the body.
func (s *Spec) Doc() *doc.Doc { return s.doc }

// Project is a loaded .forge directory.
type Project struct {
	Root   string // repository root
	Config *doc.Doc
	Specs  []*Spec
}

// Find walks up from dir looking for a .forge directory.
func Find(dir string) (string, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	for {
		if st, err := os.Stat(filepath.Join(abs, Dir)); err == nil && st.IsDir() {
			return abs, nil
		}
		parent := filepath.Dir(abs)
		if parent == abs {
			return "", ErrNotInitialized
		}
		abs = parent
	}
}

// Load reads the project rooted at the repository containing dir.
func Load(dir string) (*Project, error) {
	root, err := Find(dir)
	if err != nil {
		return nil, err
	}
	p := &Project{Root: root}
	cfg, err := doc.Load(filepath.Join(root, Dir, "project.md"))
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		cfg = doc.New()
	}
	p.Config = cfg

	entries, err := os.ReadDir(p.SpecsDir())
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".md") || strings.EqualFold(name, "README.md") {
			continue
		}
		s, err := loadSpec(filepath.Join(p.SpecsDir(), name))
		if err != nil {
			return nil, err
		}
		p.Specs = append(p.Specs, s)
	}
	sort.Slice(p.Specs, func(i, j int) bool { return p.Specs[i].Num < p.Specs[j].Num })
	return p, nil
}

// SpecsDir is where spec files live.
func (p *Project) SpecsDir() string { return filepath.Join(p.Root, Dir, "specs") }

// WipDir is where the scaffolding of in-flight specs lives.
func (p *Project) WipDir() string { return filepath.Join(p.Root, Dir, "wip") }

// WipDirFor is the scaffolding directory of one spec.
func (p *Project) WipDirFor(id string) string { return filepath.Join(p.WipDir(), id) }

// Maintainers are the people allowed to accept and approve.
func (p *Project) Maintainers() []string { return p.Config.List("maintainers") }

// IsMaintainer reports whether handle may accept or approve.
func (p *Project) IsMaintainer(handle string) bool {
	for _, m := range p.Maintainers() {
		if strings.EqualFold(strings.TrimSpace(m), strings.TrimSpace(handle)) {
			return true
		}
	}
	return false
}

// AllowSelfApproval reports whether a conductor may approve their own spec.
// With a single maintainer it is always allowed: there is nobody else.
func (p *Project) AllowSelfApproval() bool {
	if len(p.Maintainers()) < 2 {
		return true
	}
	return truthy(p.Config.Str("allow_self_approval"))
}

// GuardEnabled reports whether the PreToolUse guard should deny edits.
func (p *Project) GuardEnabled() bool {
	v := strings.ToLower(strings.TrimSpace(p.Config.Str("guard")))
	return v != "off" && v != "false" && v != "no"
}

// Configured reports whether onboarding has been done.
func (p *Project) Configured() bool {
	return len(p.Maintainers()) > 0 && p.Config.Str("test") != ""
}

// Spec returns a spec by ID, case insensitive.
func (p *Project) Spec(id string) (*Spec, bool) {
	id = NormalizeID(id)
	for _, s := range p.Specs {
		if s.ID == id {
			return s, true
		}
	}
	return nil, false
}

// Children returns the specs whose parent is id, in ID order.
func (p *Project) Children(id string) []*Spec {
	var out []*Spec
	for _, s := range p.Specs {
		if s.Parent == id {
			out = append(out, s)
		}
	}
	return out
}

// NextNum returns the next free spec number.
func (p *Project) NextNum() int {
	max := 0
	for _, s := range p.Specs {
		if s.Num > max {
			max = s.Num
		}
	}
	return max + 1
}

// DepSatisfied reports whether a dependency is met, and why not when it is not.
func (p *Project) DepSatisfied(d Dep) (bool, string) {
	other, ok := p.Spec(d.ID)
	if !ok {
		return false, fmt.Sprintf("%s does not exist", d.ID)
	}
	if d.Level == "contract" {
		if other.ContractHash == "" {
			return false, fmt.Sprintf("%s has no approved contract yet", d.ID)
		}
		return true, ""
	}
	if other.Status != workflow.Done {
		return false, fmt.Sprintf("%s is %s, not done", d.ID, other.Status)
	}
	return true, ""
}

// Blockers lists every reason a spec cannot be started right now.
func (p *Project) Blockers(s *Spec) []string {
	var out []string
	for _, d := range s.Deps {
		if ok, why := p.DepSatisfied(d); !ok {
			out = append(out, why)
		}
	}
	for _, n := range s.Needs {
		out = append(out, fmt.Sprintf("unresolved need: %s", n))
	}
	for _, e := range s.External {
		who := e["who"]
		if who == "" {
			who = "unassigned"
		}
		out = append(out, fmt.Sprintf("external: %s (%s)", e["what"], who))
	}
	if len(p.Children(s.ID)) > 0 {
		out = append(out, "this spec has children; work happens in them")
	}
	return out
}

// Coverage maps each criterion of a parent spec to the children covering it.
type Coverage struct {
	Criterion Criterion
	By        []*Spec
}

// Coverage returns the coverage matrix of a spec with children.
func (p *Project) Coverage(s *Spec) []Coverage {
	children := p.Children(s.ID)
	out := make([]Coverage, 0)
	for _, c := range s.Criteria() {
		row := Coverage{Criterion: c}
		for _, child := range children {
			for _, ac := range child.Covers {
				if strings.EqualFold(ac, c.ID) {
					row.By = append(row.By, child)
					break
				}
			}
		}
		out = append(out, row)
	}
	return out
}

// Criteria parses the acceptance criteria out of the body.
var criterionRe = regexp.MustCompile(`^-\s*(?:\[[ xX]\]\s*)?(AC\d+)\s*[:.-]\s*(.+)$`)

// Criteria returns the acceptance criteria declared by the spec.
func (s *Spec) Criteria() []Criterion {
	body := s.doc.Section("Acceptance criteria")
	if body == "" {
		body = s.doc.Section("Criterios de aceptación")
	}
	var out []Criterion
	for _, line := range strings.Split(body, "\n") {
		if m := criterionRe.FindStringSubmatch(strings.TrimSpace(line)); m != nil {
			out = append(out, Criterion{ID: strings.ToUpper(m[1]), Text: strings.TrimSpace(m[2])})
		}
	}
	return out
}

// Contract returns the contract section, empty while it is not written.
func (s *Spec) Contract() string {
	if c := s.doc.Section("Contract"); c != "" {
		return c
	}
	return s.doc.Section("Contrato")
}

// ContractChanged reports whether the contract was edited after approval.
// Everything anyone else built on top of it was agreed against the old one.
func (s *Spec) ContractChanged() bool {
	return s.ContractHash != "" && HashContract(s.Contract()) != s.ContractHash
}

// HashContract is the fingerprint stored on approval and compared later.
func HashContract(contract string) string {
	norm := strings.Join(strings.Fields(contract), " ")
	sum := sha256.Sum256([]byte(norm))
	return hex.EncodeToString(sum[:])[:12]
}

// AppendHistory records a transition in the body of the spec.
func (s *Spec) AppendHistory(line string) {
	heading := "History"
	for _, h := range s.doc.Headings() {
		if strings.EqualFold(h, "Historia") {
			heading = "Historia"
		}
	}
	s.doc.AppendToSection(heading, line)
}

// SetStatus writes the new state and a history line.
func (s *Spec) SetStatus(to workflow.State, by, note string) {
	s.Status = to
	s.doc.SetStr("status", string(to))
	s.doc.SetStr("updated", time.Now().Format("2006-01-02"))
	line := fmt.Sprintf("- %s  %s  by %s", time.Now().Format("2006-01-02"), to, by)
	if note != "" {
		line += ": " + note
	}
	s.AppendHistory(line)
}

// Save writes the spec back to disk, syncing the typed fields first.
func (s *Spec) Save() error {
	s.doc.SetStr("title", s.Title)
	s.doc.SetStr("status", string(s.Status))
	setOrDelete(s.doc, "parent", s.Parent)
	setListOrDelete(s.doc, "covers", s.Covers)
	deps := make([]string, 0, len(s.Deps))
	for _, d := range s.Deps {
		deps = append(deps, d.String())
	}
	setListOrDelete(s.doc, "depends_on", deps)
	setListOrDelete(s.doc, "needs", s.Needs)
	if len(s.External) == 0 {
		s.doc.Delete("blocked_by_external")
	} else {
		s.doc.SetMapList("blocked_by_external", s.External)
	}
	setOrDelete(s.doc, "conductor", s.Conductor)
	setOrDelete(s.doc, "accepted_by", s.AcceptedBy)
	setOrDelete(s.doc, "approved_by", s.ApprovedBy)
	setOrDelete(s.doc, "contract_hash", s.ContractHash)
	agreed := make([]string, 0, len(s.Agreed))
	for id, h := range s.Agreed {
		agreed = append(agreed, id+":"+h)
	}
	sort.Strings(agreed)
	setListOrDelete(s.doc, "agreed_contracts", agreed)
	setOrDelete(s.doc, "pr", s.PR)
	return s.doc.Save(s.Path)
}

func loadSpec(path string) (*Spec, error) {
	d, err := doc.Load(path)
	if err != nil {
		return nil, err
	}
	return FromDoc(path, d)
}

// FromDoc builds a spec from an already parsed document, so a spec created
// from a template is exactly the same shape as one read from disk.
func FromDoc(path string, d *doc.Doc) (*Spec, error) {
	id := NormalizeID(d.Str("id"))
	if id == "" {
		return nil, fmt.Errorf("%s: missing id in frontmatter", filepath.Base(path))
	}
	num, err := numOf(id)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", filepath.Base(path), err)
	}
	s := &Spec{
		ID:           id,
		Num:          num,
		Title:        d.Str("title"),
		Status:       workflow.State(d.Str("status")),
		Parent:       NormalizeID(d.Str("parent")),
		Covers:       upperAll(d.List("covers")),
		Needs:        d.List("needs"),
		External:     d.MapList("blocked_by_external"),
		Path:         path,
		Conductor:    d.Str("conductor"),
		AcceptedBy:   d.Str("accepted_by"),
		ApprovedBy:   d.Str("approved_by"),
		ContractHash: d.Str("contract_hash"),
		Agreed:       map[string]string{},
		PR:           d.Str("pr"),
		doc:          d,
	}
	for _, raw := range d.List("depends_on") {
		id, level, _ := strings.Cut(strings.TrimSpace(raw), "@")
		s.Deps = append(s.Deps, Dep{ID: NormalizeID(id), Level: strings.ToLower(level)})
	}
	for _, raw := range d.List("agreed_contracts") {
		id, hash, ok := strings.Cut(raw, ":")
		if ok {
			s.Agreed[NormalizeID(id)] = strings.TrimSpace(hash)
		}
	}
	return s, nil
}

var idRe = regexp.MustCompile(`^SPEC-(\d+)$`)

// NormalizeID accepts "4", "spec-4" or "SPEC-004" and returns "SPEC-004".
func NormalizeID(raw string) string {
	s := strings.ToUpper(strings.TrimSpace(raw))
	if s == "" {
		return ""
	}
	s = strings.TrimPrefix(s, "SPEC-")
	n, err := strconv.Atoi(s)
	if err != nil {
		return strings.ToUpper(strings.TrimSpace(raw))
	}
	return FormatID(n)
}

// FormatID renders a spec number as an ID.
func FormatID(n int) string { return fmt.Sprintf("SPEC-%03d", n) }

func numOf(id string) (int, error) {
	m := idRe.FindStringSubmatch(id)
	if m == nil {
		return 0, fmt.Errorf("id %q is not SPEC-NNN", id)
	}
	return strconv.Atoi(m[1])
}

var slugRe = regexp.MustCompile(`[^a-z0-9]+`)

// Slug turns a title into the file name part after the ID.
func Slug(title string) string {
	s := strings.ToLower(strings.TrimSpace(title))
	s = strings.NewReplacer("á", "a", "é", "e", "í", "i", "ó", "o", "ú", "u", "ñ", "n").Replace(s)
	s = slugRe.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if len(s) > 48 {
		if cut := strings.LastIndex(s[:48], "-"); cut > 16 {
			s = s[:cut]
		} else {
			s = s[:48]
		}
	}
	if s == "" {
		s = "untitled"
	}
	return s
}

// FileName is the spec file name for an ID and title.
func FileName(id, title string) string { return id + "-" + Slug(title) + ".md" }

func setOrDelete(d *doc.Doc, key, val string) {
	if val == "" {
		d.Delete(key)
		return
	}
	d.SetStr(key, val)
}

func setListOrDelete(d *doc.Doc, key string, items []string) {
	if len(items) == 0 {
		d.Delete(key)
		return
	}
	d.SetList(key, items)
}

func upperAll(in []string) []string {
	out := make([]string, 0, len(in))
	for _, s := range in {
		out = append(out, strings.ToUpper(strings.TrimSpace(s)))
	}
	return out
}

func truthy(s string) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "true", "yes", "on", "1":
		return true
	}
	return false
}
