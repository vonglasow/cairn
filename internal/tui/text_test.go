package tui

import (
	"strings"
	"testing"
)

func TestWrap(t *testing.T) {
	for _, c := range []struct {
		name, in     string
		width, lines int
		want         []string
	}{
		{"one short line", "hello", 10, 2, []string{"hello", ""}},
		{"cut at a word", "hello there you", 11, 2, []string{"hello there", "you"}},
		{"a word longer than the line is cut", "abcdefghij", 4, 2, []string{"abc…", ""}},
		{"what will not fit is marked on the last line", "one two three four five", 7, 2, []string{"one two", "three …"}},
		{"no width answers empty lines", "hello", 0, 2, []string{"", ""}},
		{"no lines answers nothing", "hello", 10, 0, []string{}},
		{"nothing to say fills nothing", "", 10, 2, []string{"", ""}},
	} {
		t.Run(c.name, func(t *testing.T) {
			got := Wrap(c.in, c.width, c.lines)
			if len(got) != len(c.want) {
				t.Fatalf("Wrap(%q, %d, %d) gave %d lines, want %d", c.in, c.width, c.lines, len(got), len(c.want))
			}
			for i := range got {
				if got[i] != c.want[i] {
					t.Fatalf("Wrap(%q, %d, %d) line %d = %q, want %q", c.in, c.width, c.lines, i, got[i], c.want[i])
				}
			}
		})
	}
}

func TestSpaced(t *testing.T) {
	if got := spaced(nil); got != nil {
		t.Fatalf("spaced on nothing gave %v", got)
	}
	got := spaced([]string{"a", "b"})
	if strings.Join(got, "") != "a b" {
		t.Fatalf("spaced gave %v", got)
	}
}

func TestATextIsLaidOutParagraphByParagraph(t *testing.T) {
	got := Paragraphs("un deux trois\n\nquatre", 8)
	want := []string{"un deux", "trois", "", "quatre"}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Fatalf("laid out as %q, want %q", got, want)
	}
	if got := Paragraphs("", 8); len(got) != 0 {
		t.Fatalf("nothing laid out as %q", got)
	}
}
