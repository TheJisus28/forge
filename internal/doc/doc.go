// Package doc reads and writes the Markdown files Forge owns: a YAML
// frontmatter block followed by a Markdown body.
//
// The subset of YAML supported here is the one the templates use: scalars,
// inline and block lists of scalars, and lists of flat maps. Anything richer
// is rejected with an error rather than silently misread.
package doc

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const fence = "---"

// Kind distinguishes the three shapes a frontmatter value can take.
type Kind int

const (
	Scalar Kind = iota
	List
	MapList
)

// Value is a frontmatter value. Only the field matching Kind is meaningful.
type Value struct {
	Kind    Kind
	Str     string
	Items   []string
	Maps    []map[string]string
	mapKeys [][]string // key order per map, so writing round-trips
}

// Doc is a parsed Forge file. Key order is preserved so rewriting a file
// produces a minimal diff.
type Doc struct {
	keys   []string
	values map[string]Value
	Body   string
}

// New returns an empty document.
func New() *Doc {
	return &Doc{values: map[string]Value{}}
}

// Keys returns the frontmatter keys in file order.
func (d *Doc) Keys() []string { return append([]string(nil), d.keys...) }

// Has reports whether the key is present.
func (d *Doc) Has(key string) bool { _, ok := d.values[key]; return ok }

// Str returns a scalar value, or "" when absent or of another kind.
func (d *Doc) Str(key string) string {
	v, ok := d.values[key]
	if !ok || v.Kind != Scalar {
		return ""
	}
	return v.Str
}

// List returns a list value, or nil when absent or of another kind.
func (d *Doc) List(key string) []string {
	v, ok := d.values[key]
	if !ok || v.Kind != List {
		return nil
	}
	return append([]string(nil), v.Items...)
}

// MapList returns a list-of-maps value, or nil when absent or of another kind.
func (d *Doc) MapList(key string) []map[string]string {
	v, ok := d.values[key]
	if !ok || v.Kind != MapList {
		return nil
	}
	out := make([]map[string]string, 0, len(v.Maps))
	for _, m := range v.Maps {
		c := make(map[string]string, len(m))
		for k, val := range m {
			c[k] = val
		}
		out = append(out, c)
	}
	return out
}

// SetStr sets a scalar, appending the key when it is new.
func (d *Doc) SetStr(key, val string) { d.set(key, Value{Kind: Scalar, Str: val}) }

// SetList sets a list, appending the key when it is new.
func (d *Doc) SetList(key string, items []string) {
	d.set(key, Value{Kind: List, Items: append([]string(nil), items...)})
}

// SetMapList sets a list of flat maps, appending the key when it is new.
func (d *Doc) SetMapList(key string, maps []map[string]string) {
	v := Value{Kind: MapList}
	for _, m := range maps {
		keys := make([]string, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		c := make(map[string]string, len(m))
		for k, val := range m {
			c[k] = val
		}
		v.Maps = append(v.Maps, c)
		v.mapKeys = append(v.mapKeys, keys)
	}
	d.set(key, v)
}

// Delete removes a key.
func (d *Doc) Delete(key string) {
	if _, ok := d.values[key]; !ok {
		return
	}
	delete(d.values, key)
	for i, k := range d.keys {
		if k == key {
			d.keys = append(d.keys[:i], d.keys[i+1:]...)
			break
		}
	}
}

func (d *Doc) set(key string, v Value) {
	if _, ok := d.values[key]; !ok {
		d.keys = append(d.keys, key)
	}
	d.values[key] = v
}

// Section returns the body of a "## Heading" section, without the heading.
func (d *Doc) Section(heading string) string {
	want := strings.ToLower(strings.TrimSpace(heading))
	var out []string
	in := false
	fenced := false
	for _, line := range strings.Split(d.Body, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			fenced = !fenced
		}
		if !fenced && strings.HasPrefix(line, "## ") {
			if in {
				break
			}
			in = strings.ToLower(strings.TrimSpace(strings.TrimPrefix(line, "## "))) == want
			continue
		}
		if in {
			out = append(out, line)
		}
	}
	return strings.TrimSpace(strings.Join(out, "\n"))
}

// htmlCommentRe matches an HTML comment, the shape the templates use to ship
// guidance inside a section without it being read as content. The dash form
// is excluded so a `<--` typo is not eaten as a comment.
var htmlCommentRe = regexp.MustCompile(`(?s)<!--.*?-->`)

// StripComments removes HTML comments from s, so the guidance a template
// ships inside a section is not read as content. It is the one rule the
// archive gate and the criterion coverage derivation share (SPEC-021, the
// SPEC-019 principle one reader over). An unterminated comment is left as it
// is rather than swallowing the rest of the text.
func StripComments(s string) string {
	return htmlCommentRe.ReplaceAllString(s, "")
}

// AppendToSection adds a line at the end of a section, creating the section
// at the end of the document when it does not exist yet.
func (d *Doc) AppendToSection(heading, line string) {
	lines := strings.Split(strings.TrimRight(d.Body, "\n"), "\n")
	want := strings.ToLower(strings.TrimSpace(heading))
	start := -1
	for i, l := range lines {
		if strings.HasPrefix(l, "## ") &&
			strings.ToLower(strings.TrimSpace(strings.TrimPrefix(l, "## "))) == want {
			start = i
			break
		}
	}
	if start == -1 {
		d.Body = strings.TrimRight(d.Body, "\n") + "\n\n## " + heading + "\n\n" + line + "\n"
		return
	}
	end := len(lines)
	for i := start + 1; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "## ") {
			end = i
			break
		}
	}
	for end > start+1 && strings.TrimSpace(lines[end-1]) == "" {
		end--
	}
	rest := append([]string{line}, lines[end:]...)
	d.Body = strings.Join(append(lines[:end:end], rest...), "\n") + "\n"
}

// Parse reads a document from raw bytes.
func Parse(data []byte) (*Doc, error) {
	d := New()
	text := strings.ReplaceAll(string(data), "\r\n", "\n")
	if !strings.HasPrefix(text, fence+"\n") {
		d.Body = text
		return d, nil
	}
	rest := text[len(fence)+1:]
	end := strings.Index(rest, "\n"+fence)
	if end == -1 {
		return nil, fmt.Errorf("frontmatter is not closed with %q", fence)
	}
	front := rest[:end]
	body := rest[end+len(fence)+1:]
	d.Body = strings.TrimPrefix(body, "\n")
	if err := d.parseFront(front); err != nil {
		return nil, err
	}
	return d, nil
}

func (d *Doc) parseFront(front string) error {
	lines := strings.Split(front, "\n")
	for i := 0; i < len(lines); i++ {
		raw := lines[i]
		if strings.TrimSpace(raw) == "" || strings.HasPrefix(strings.TrimSpace(raw), "#") {
			continue
		}
		if raw != strings.TrimLeft(raw, " ") {
			return fmt.Errorf("unexpected indentation at %q", raw)
		}
		key, rawVal, ok := strings.Cut(raw, ":")
		if !ok {
			return fmt.Errorf("line %q is not a key: value pair", raw)
		}
		key = strings.TrimSpace(key)
		val := strings.TrimSpace(stripComment(rawVal))
		switch {
		case val == "":
			block, next, err := parseBlock(lines, i+1)
			if err != nil {
				return fmt.Errorf("key %q: %w", key, err)
			}
			i = next - 1
			d.set(key, block)
		case strings.HasPrefix(val, "["):
			d.set(key, Value{Kind: List, Items: parseInlineList(val)})
		default:
			d.set(key, Value{Kind: Scalar, Str: unquote(val)})
		}
	}
	return nil
}

// parseBlock reads the indented lines that follow a bare "key:".
func parseBlock(lines []string, start int) (Value, int, error) {
	v := Value{Kind: List}
	i := start
	var cur map[string]string
	var curKeys []string
	flush := func() {
		if cur != nil {
			v.Maps = append(v.Maps, cur)
			v.mapKeys = append(v.mapKeys, curKeys)
			cur, curKeys = nil, nil
		}
	}
	for ; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if line == strings.TrimLeft(line, " ") {
			break
		}
		switch {
		case strings.HasPrefix(trimmed, "- "):
			item := strings.TrimSpace(stripComment(strings.TrimPrefix(trimmed, "- ")))
			if k, val, ok := strings.Cut(item, ": "); ok {
				flush()
				v.Kind = MapList
				cur = map[string]string{strings.TrimSpace(k): unquote(strings.TrimSpace(val))}
				curKeys = []string{strings.TrimSpace(k)}
				continue
			}
			if v.Kind == MapList {
				return v, i, fmt.Errorf("mixes plain items and maps")
			}
			v.Items = append(v.Items, unquote(item))
		case cur != nil:
			k, val, ok := strings.Cut(trimmed, ":")
			if !ok {
				return v, i, fmt.Errorf("line %q is not a key: value pair", trimmed)
			}
			k = strings.TrimSpace(k)
			cur[k] = unquote(strings.TrimSpace(stripComment(val)))
			curKeys = append(curKeys, k)
		default:
			return v, i, fmt.Errorf("unexpected line %q", trimmed)
		}
	}
	flush()
	if v.Kind == List && v.Items == nil {
		v.Items = []string{}
	}
	return v, i, nil
}

func parseInlineList(val string) []string {
	inner := strings.TrimSuffix(strings.TrimPrefix(val, "["), "]")
	if strings.TrimSpace(inner) == "" {
		return []string{}
	}
	parts := strings.Split(inner, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if s := unquote(strings.TrimSpace(p)); s != "" {
			out = append(out, s)
		}
	}
	return out
}

// stripComment drops a trailing "# ..." that is not inside quotes.
func stripComment(s string) string {
	quote := byte(0)
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case quote != 0:
			if c == quote {
				quote = 0
			}
		case c == '"' || c == '\'':
			quote = c
		case c == '#' && i > 0 && (s[i-1] == ' ' || s[i-1] == '\t'):
			return s[:i]
		}
	}
	return s
}

func unquote(s string) string {
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}

func quote(s string) string {
	if s == "" {
		return `""`
	}
	if strings.ContainsAny(s, `:#"'[]{}`) || strings.HasPrefix(s, " ") || strings.HasSuffix(s, " ") {
		return `"` + strings.ReplaceAll(s, `"`, `\"`) + `"`
	}
	return s
}

// String renders the document back to text.
func (d *Doc) String() string {
	var b strings.Builder
	if len(d.keys) > 0 {
		b.WriteString(fence + "\n")
		for _, k := range d.keys {
			v := d.values[k]
			switch v.Kind {
			case Scalar:
				fmt.Fprintf(&b, "%s: %s\n", k, quote(v.Str))
			case List:
				if len(v.Items) == 0 {
					fmt.Fprintf(&b, "%s: []\n", k)
					continue
				}
				fmt.Fprintf(&b, "%s:\n", k)
				for _, it := range v.Items {
					fmt.Fprintf(&b, "  - %s\n", quote(it))
				}
			case MapList:
				fmt.Fprintf(&b, "%s:\n", k)
				for i, m := range v.Maps {
					keys := v.mapKeys[i]
					if len(keys) == 0 {
						keys = sortedKeys(m)
					}
					for j, mk := range keys {
						prefix := "    "
						if j == 0 {
							prefix = "  - "
						}
						fmt.Fprintf(&b, "%s%s: %s\n", prefix, mk, quote(m[mk]))
					}
				}
			}
		}
		b.WriteString(fence + "\n\n")
	}
	b.WriteString(strings.TrimLeft(d.Body, "\n"))
	out := b.String()
	if !strings.HasSuffix(out, "\n") {
		out += "\n"
	}
	return out
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// Load reads and parses a file.
func Load(path string) (*Doc, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	d, err := Parse(data)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", filepath.Base(path), err)
	}
	return d, nil
}

// Save writes the document, creating parent directories as needed.
func (d *Doc) Save(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(d.String()), 0o644)
}

// Headings lists the "## " headings of the body, in order.
func (d *Doc) Headings() []string {
	var out []string
	sc := bufio.NewScanner(strings.NewReader(d.Body))
	for sc.Scan() {
		if line := sc.Text(); strings.HasPrefix(line, "## ") {
			out = append(out, strings.TrimSpace(strings.TrimPrefix(line, "## ")))
		}
	}
	return out
}
