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
	ID         string
	Num        int
	Title      string
	Capability string
	Status     workflow.State
	Parent     string
	Covers     []string
	Supersedes []string
	Deps       []Dep
	Needs      []string
	External   []map[string]string
	Path       string

	Orchestrator string
	AcceptedBy   string
	ApprovedBy   string
	ContractHash string
	Agreed       map[string]string // spec ID -> contract hash this spec started against
	PR           string

	doc *doc.Doc
}

// Doc exposes the underlying document for commands that edit the body.
func (s *Spec) Doc() *doc.Doc { return s.doc }

// Dir is the folder that holds the spec and its plan, tasks and review.
func (s *Spec) Dir() string { return filepath.Dir(s.Path) }

// PlanPath, TasksPath and ReviewPath are the spec's working artifacts.
func (s *Spec) PlanPath() string   { return filepath.Join(s.Dir(), "plan.md") }
func (s *Spec) TasksPath() string  { return filepath.Join(s.Dir(), "tasks.md") }
func (s *Spec) ReviewPath() string { return filepath.Join(s.Dir(), "review.md") }

// TaskProgress counts the checked and total checkbox tasks in tasks.md. Both
// are zero when the file has no tasks yet.
func (s *Spec) TaskProgress() (done, total int) {
	data, err := os.ReadFile(s.TasksPath())
	if err != nil {
		return 0, 0
	}
	for _, line := range strings.Split(string(data), "\n") {
		l := strings.TrimSpace(line)
		if !strings.HasPrefix(l, "- [") || len(l) < 5 || l[4] != ']' {
			continue
		}
		total++
		if l[3] == 'x' || l[3] == 'X' {
			done++
		}
	}
	return done, total
}

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
		if !e.IsDir() {
			continue
		}
		path := filepath.Join(p.SpecsDir(), e.Name(), "spec.md")
		if _, err := os.Stat(path); err != nil {
			continue
		}
		s, err := loadSpec(path)
		if err != nil {
			return nil, err
		}
		p.Specs = append(p.Specs, s)
	}
	sort.Slice(p.Specs, func(i, j int) bool { return p.Specs[i].Num < p.Specs[j].Num })
	return p, nil
}

// SpecsDir is where one folder per spec lives.
func (p *Project) SpecsDir() string { return filepath.Join(p.Root, Dir, "specs") }

// SpecDir is the folder of one spec: specs/<id>-<slug>/.
func (p *Project) SpecDir(id, title string) string {
	return filepath.Join(p.SpecsDir(), SpecDirName(id, title))
}

// SpecDirName is the folder name of a spec.
func SpecDirName(id, title string) string { return id + "-" + Slug(title) }

// GuardEnabled reports whether the PreToolUse guard should deny edits.
func (p *Project) GuardEnabled() bool {
	v := strings.ToLower(strings.TrimSpace(p.Config.Str("guard")))
	return v != "off" && v != "false" && v != "no"
}

// PushEnabled reports whether `forge advance` checkpoints automatically. It is
// opt-in through the `push` scalar: only on, true and yes enable it; anything
// else, including an absent scalar, leaves it off (SPEC-020, decision 6).
func (p *Project) PushEnabled() bool {
	switch strings.ToLower(strings.TrimSpace(p.Config.Str("push"))) {
	case "on", "true", "yes":
		return true
	}
	return false
}

// FetchEnabled reports whether `forge brief` refreshes the remote refs at
// session start. It is opt-in through the `fetch` scalar: only on, true and
// yes enable it; anything else, including an absent scalar, leaves it off
// (SPEC-022, decision 1).
func (p *Project) FetchEnabled() bool {
	switch strings.ToLower(strings.TrimSpace(p.Config.Str("fetch"))) {
	case "on", "true", "yes":
		return true
	}
	return false
}

// Configured reports whether onboarding has been done.
func (p *Project) Configured() bool {
	return p.Config.Str("test") != ""
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

// CapabilityGroup is one capability in the derived view: its name and the
// delivered contracts that declare it.
type CapabilityGroup struct {
	Name      string
	Contracts []Contract
}

// Contract is one delivered spec as the capability view shows it. SupersededBy
// names the non-dropped specs that replace it, in number order; an empty list
// means the contract still describes current behaviour.
type Contract struct {
	ID           string
	Title        string
	SupersededBy []string
}

// Current reports whether the contract still describes current behaviour:
// no non-dropped spec supersedes it.
func (c Contract) Current() bool { return len(c.SupersededBy) == 0 }

// Capabilities derives the capability view from the specs on disk: the
// contracts of every done spec, grouped by the capability it declares, with
// capabilities in name order and contracts in number order. Each contract
// names the non-dropped specs that supersede it, so a caller can tell the
// current shape from the history without reading every contract. Specs that
// are not done, and done specs that declare no capability, are not part of
// the view. Nothing is cached or written.
func (p *Project) Capabilities() []CapabilityGroup {
	supersededBy := map[string][]string{}
	for _, s := range p.Specs {
		if s.Status == workflow.Dropped {
			continue
		}
		for _, target := range s.Supersedes {
			supersededBy[target] = append(supersededBy[target], s.ID)
		}
	}

	byName := map[string]*CapabilityGroup{}
	for _, s := range p.Specs {
		if s.Status != workflow.Done || s.Capability == "" {
			continue
		}
		g := byName[s.Capability]
		if g == nil {
			g = &CapabilityGroup{Name: s.Capability}
			byName[s.Capability] = g
		}
		g.Contracts = append(g.Contracts, Contract{
			ID:           s.ID,
			Title:        s.Title,
			SupersededBy: supersededBy[s.ID],
		})
	}

	groups := make([]CapabilityGroup, 0, len(byName))
	for _, g := range byName {
		groups = append(groups, *g)
	}
	sort.Slice(groups, func(i, j int) bool { return groups[i].Name < groups[j].Name })
	return groups
}

// Criteria parses the acceptance criteria out of the body.
var criterionRe = regexp.MustCompile(`^-\s*(?:\[[ xX]\]\s*)?(AC\d+)\s*[:.-]\s*(.+)$`)

// Criteria returns the acceptance criteria declared by the spec. Headings are
// fixed English: `working_language` governs the prose, never the heading the
// CLI reads.
func (s *Spec) Criteria() []Criterion {
	body := s.doc.Section("Acceptance criteria")
	var out []Criterion
	for _, line := range strings.Split(body, "\n") {
		if m := criterionRe.FindStringSubmatch(strings.TrimSpace(line)); m != nil {
			out = append(out, Criterion{ID: strings.ToUpper(m[1]), Text: strings.TrimSpace(m[2])})
		}
	}
	return out
}

// The anchor patterns are the three kinds of evidence a criterion can name: a
// command in an inline code span, a test reference, or an observable outcome.
// A criterion is verifiable when its text carries at least one.
var (
	codeSpanRe = regexp.MustCompile("`[^`]*\\S[^`]*`")
	testWordRe = regexp.MustCompile(`(?i)\btests?\b`)
	testNameRe = regexp.MustCompile(`\bTest[A-Za-z0-9_]*`)
	outcomeRe  = regexp.MustCompile(`(?i)\b(?:returns|prints|outputs|exits|succeeds|fails|refuses|rejects|reports|lists|names|matches|emits|responds)\b`)
)

// Verifiable reports whether the criterion names the evidence that settles
// it: a command in backticks, a test, or an observable outcome. It is the one
// place the anchor rule lives; `forge approve` refuses a criterion whose text
// names none of them.
func (c Criterion) Verifiable() bool {
	text := strings.TrimSpace(c.Text)
	return codeSpanRe.MatchString(text) ||
		testWordRe.MatchString(text) ||
		testNameRe.MatchString(text) ||
		outcomeRe.MatchString(text)
}

// CriterionGap is one criterion the spec's artifacts do not settle.
type CriterionGap struct {
	Criterion Criterion
	File      string // absolute path of the artifact it is missing from
	Kind      string // "task" or "evidence"
}

// tokenRe splits an artifact into the tokens a criterion id is compared
// against. A token is a run of letters, digits, underscores and hyphens, so
// AC1, AC10, AC1x and AC1- are four distinct tokens.
var tokenRe = regexp.MustCompile(`[0-9A-Za-z_-]+`)

// hasToken reports whether text contains id as a whole token, case
// insensitively. It tokenizes rather than matching a consuming pattern so two
// ids side by side are both read: in `AC1 AC2` and `AC1,AC2` the delimiter
// between them is not eaten, while AC1 still does not equal AC10, AC1x or
// AC1- (SPEC-021, decision 3).
func hasToken(text, id string) bool {
	for _, tok := range tokenRe.FindAllString(text, -1) {
		if strings.EqualFold(tok, id) {
			return true
		}
	}
	return false
}

// taskGapApplies reports whether tasks.md is the record for the state: a task
// gap is meaningful from implementing on.
func taskGapApplies(s workflow.State) bool {
	switch s {
	case workflow.Implementing, workflow.Blocked, workflow.Reviewing, workflow.Done:
		return true
	}
	return false
}

// evidenceGapApplies reports whether review.md is the record for the state: an
// evidence gap is meaningful from reviewing on.
func evidenceGapApplies(s workflow.State) bool {
	switch s {
	case workflow.Reviewing, workflow.Done:
		return true
	}
	return false
}

// CriterionGaps lists every criterion no task in tasks.md delivers and no
// evidence line in review.md settles, for the spec's current state. A task gap
// is reported from implementing on; an evidence gap from reviewing on.
// Criteria keep declaration order and a task gap precedes an evidence gap for
// the same criterion. A missing or unreadable artifact names nothing, so every
// applicable criterion is a gap. A spec with no Existing state section is
// history under the workflow before SPEC-015 and is never judged (decision 5).
func (s *Spec) CriterionGaps() []CriterionGap {
	if s.ExistingState() == "" {
		return nil
	}
	criteria := s.Criteria()
	if len(criteria) == 0 {
		return nil
	}

	taskRecord := taskGapApplies(s.Status)
	evidenceRecord := evidenceGapApplies(s.Status)

	// A missing or unreadable file leaves the record empty, which names
	// nothing and so reads as a gap for every criterion. Comments are
	// stripped first: a commented `AC1` is template guidance, not a task and
	// not evidence (SPEC-021, the SPEC-019 principle one reader over).
	var tasks string
	if taskRecord {
		if data, err := os.ReadFile(s.TasksPath()); err == nil {
			tasks = doc.StripComments(string(data))
		}
	}
	var evidence string
	if evidenceRecord {
		if d, err := doc.Load(s.ReviewPath()); err == nil {
			evidence = doc.StripComments(d.Section("Acceptance criteria"))
		}
	}

	var out []CriterionGap
	for _, c := range criteria {
		if taskRecord && !hasToken(tasks, c.ID) {
			out = append(out, CriterionGap{Criterion: c, File: s.TasksPath(), Kind: "task"})
		}
		if evidenceRecord && !hasToken(evidence, c.ID) {
			out = append(out, CriterionGap{Criterion: c, File: s.ReviewPath(), Kind: "evidence"})
		}
	}
	return out
}

// Contract returns the contract section, empty while it is not written. The
// heading is fixed English, never translated.
func (s *Spec) Contract() string {
	return s.doc.Section("Contract")
}

// OpenQuestions returns the section that must be settled before the contract
// is approved, empty when it is not written. The heading is fixed English,
// never translated.
func (s *Spec) OpenQuestions() string {
	return s.doc.Section("Open questions")
}

// ExistingState returns the architect's survey of what already exists to
// reuse, empty while it is not written. It lives in spec.md; plan.md no
// longer owns that section. The heading is fixed English, never translated.
func (s *Spec) ExistingState() string {
	return s.doc.Section("Existing state")
}

// HasOpenQuestions reports whether the spec still lists an unanswered
// question. Questions are list items; prose and `None.` are not.
func (s *Spec) HasOpenQuestions() bool {
	for _, line := range strings.Split(s.OpenQuestions(), "\n") {
		l := strings.TrimSpace(line)
		if !strings.HasPrefix(l, "-") {
			continue
		}
		switch strings.ToLower(strings.TrimSpace(strings.TrimPrefix(l, "-"))) {
		case "", "none", "none.", "ninguna", "ninguna.", "-":
			continue
		}
		return true
	}
	return false
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
	setOrDelete(s.doc, "capability", s.Capability)
	setOrDelete(s.doc, "parent", s.Parent)
	setListOrDelete(s.doc, "covers", s.Covers)
	setListOrDelete(s.doc, "supersedes", s.Supersedes)
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
	setOrDelete(s.doc, "orchestrator", s.Orchestrator)
	s.doc.Delete("conductor")
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
		Capability:   strings.TrimSpace(d.Str("capability")),
		Status:       workflow.Canonical(workflow.State(d.Str("status"))),
		Parent:       NormalizeID(d.Str("parent")),
		Covers:       upperAll(d.List("covers")),
		Supersedes:   normalizeIDs(d.List("supersedes")),
		Needs:        d.List("needs"),
		External:     d.MapList("blocked_by_external"),
		Path:         path,
		Orchestrator: orchestratorFrom(d),
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

// orchestratorFrom reads the driver of a spec. Specs written before the key was
// renamed still carry `conductor`; it is read here and never written again.
func orchestratorFrom(d *doc.Doc) string {
	if o := d.Str("orchestrator"); o != "" {
		return o
	}
	return d.Str("conductor")
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

var capabilityRe = regexp.MustCompile(`^[a-z0-9-]+$`)

// ValidCapability reports whether name is a capability slug: one or more
// lowercase letters, digits or hyphens. It is the single shape both
// `forge validate` and the derived capability view rely on.
func ValidCapability(name string) bool { return capabilityRe.MatchString(name) }

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

// normalizeIDs reads a list of spec ids the way `parent` is read: forgiving
// input, canonical `SPEC-NNN` stored.
func normalizeIDs(in []string) []string {
	out := make([]string, 0, len(in))
	for _, s := range in {
		out = append(out, NormalizeID(s))
	}
	return out
}
