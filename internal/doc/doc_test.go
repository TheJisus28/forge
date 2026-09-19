package doc_test

import (
	"strings"
	"testing"

	"github.com/TheJisus28/forge/internal/doc"
)

const sample = `---
id: SPEC-004
title: "Pay: with a saved card"
status: proposed
covers: [AC1, AC3]
depends_on:
  - SPEC-003
  - SPEC-011@contract
blocked_by_external:
  - what: production credentials
    who: ana
  - what: legal review
    who: jesus
empty: []
---

## Problem

Cards are retyped every time.

## Acceptance criteria

- AC1: a saved card can be reused
`

func TestParse_ReadsEveryShape(t *testing.T) {
	d, err := doc.Parse([]byte(sample))
	if err != nil {
		t.Fatal(err)
	}
	if got := d.Str("id"); got != "SPEC-004" {
		t.Errorf("id = %q", got)
	}
	if got := d.Str("title"); got != "Pay: with a saved card" {
		t.Errorf("quoted scalar = %q", got)
	}
	if got := d.List("covers"); len(got) != 2 || got[0] != "AC1" || got[1] != "AC3" {
		t.Errorf("inline list = %v", got)
	}
	if got := d.List("depends_on"); len(got) != 2 || got[1] != "SPEC-011@contract" {
		t.Errorf("block list = %v", got)
	}
	ext := d.MapList("blocked_by_external")
	if len(ext) != 2 || ext[0]["what"] != "production credentials" || ext[1]["who"] != "jesus" {
		t.Errorf("map list = %v", ext)
	}
	if !d.Has("empty") || len(d.List("empty")) != 0 {
		t.Errorf("an empty list should be present and empty: %v", d.List("empty"))
	}
	if !strings.Contains(d.Body, "Cards are retyped") {
		t.Errorf("body lost: %q", d.Body)
	}
}

func TestString_RoundTrips(t *testing.T) {
	d, err := doc.Parse([]byte(sample))
	if err != nil {
		t.Fatal(err)
	}
	again, err := doc.Parse([]byte(d.String()))
	if err != nil {
		t.Fatalf("re-parsing our own output: %v", err)
	}
	if again.Str("title") != d.Str("title") {
		t.Errorf("title changed: %q", again.Str("title"))
	}
	if len(again.MapList("blocked_by_external")) != 2 {
		t.Errorf("map list lost on write:\n%s", d.String())
	}
	if got, want := again.Keys(), d.Keys(); strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("key order changed: %v vs %v", got, want)
	}
}

func TestSection(t *testing.T) {
	d, _ := doc.Parse([]byte(sample))
	if got := d.Section("Problem"); got != "Cards are retyped every time." {
		t.Errorf("Problem = %q", got)
	}
	if got := d.Section("Acceptance criteria"); !strings.HasPrefix(got, "- AC1:") {
		t.Errorf("criteria = %q", got)
	}
	if got := d.Section("Nowhere"); got != "" {
		t.Errorf("missing section = %q", got)
	}
}

func TestAppendToSection(t *testing.T) {
	d, _ := doc.Parse([]byte(sample))
	d.AppendToSection("Problem", "- one more line")
	if got := d.Section("Problem"); !strings.HasSuffix(got, "- one more line") {
		t.Errorf("append inside a section put it elsewhere: %q", got)
	}
	if !strings.Contains(d.Section("Acceptance criteria"), "AC1") {
		t.Error("append damaged the next section")
	}
	d.AppendToSection("History", "- first transition")
	if got := d.Section("History"); got != "- first transition" {
		t.Errorf("missing section was not created: %q", got)
	}
}

func TestParse_CommentsAndNoFrontmatter(t *testing.T) {
	d, err := doc.Parse([]byte("status: not frontmatter\n"))
	if err != nil {
		t.Fatal(err)
	}
	if len(d.Keys()) != 0 || !strings.Contains(d.Body, "not frontmatter") {
		t.Errorf("a body without fences must stay a body: %v", d.Keys())
	}

	d, err = doc.Parse([]byte("---\nguard: on   # comment\nwhat: \"a # inside quotes\"\n---\n"))
	if err != nil {
		t.Fatal(err)
	}
	if got := d.Str("guard"); got != "on" {
		t.Errorf("comment not stripped: %q", got)
	}
	if got := d.Str("what"); got != "a # inside quotes" {
		t.Errorf("quoted hash lost: %q", got)
	}
}

func TestParse_UnclosedFrontmatter(t *testing.T) {
	if _, err := doc.Parse([]byte("---\nid: SPEC-001\n")); err == nil {
		t.Fatal("expected an error for unclosed frontmatter")
	}
}

func TestDelete(t *testing.T) {
	d, _ := doc.Parse([]byte(sample))
	d.Delete("covers")
	if d.Has("covers") {
		t.Fatal("key still present")
	}
	if strings.Contains(d.String(), "covers") {
		t.Fatal("key still written")
	}
}
