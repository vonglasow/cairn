package board

import (
	"strings"
	"testing"
)

func TestClip(t *testing.T) {
	for _, c := range []struct {
		name, in string
		width    int
		want     string
	}{
		{"nothing fits in nothing", "hello", 0, ""},
		{"a negative width is nothing", "hello", -3, ""},
		{"what fits is left alone", "hello", 5, "hello"},
		{"room to spare changes nothing", "hi", 9, "hi"},
		{"one too many is marked", "hello", 4, "hel…"},
		{"the space before the mark goes", "a bc", 3, "a…"},
		{"a wide rune is counted as it is drawn", "日本語", 4, "日…"},
	} {
		t.Run(c.name, func(t *testing.T) {
			if got := Clip(c.in, c.width); got != c.want {
				t.Fatalf("Clip(%q, %d) = %q, want %q", c.in, c.width, got, c.want)
			}
		})
	}
}

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

func TestSpread(t *testing.T) {
	for _, c := range []struct {
		name, left, right string
		width             int
		want              string
	}{
		{"pushed to both ends", "1", "x", 5, "1   x"},
		{"the right wins when there is no room", "1", "xxxxx", 3, "xx…"},
		{"the left is cut, and the space between is kept", "abcdef", "x", 5, "ab… x"},
		{"nothing on the right still fills", "ab", "", 4, "ab  "},
	} {
		t.Run(c.name, func(t *testing.T) {
			if got := Spread(c.left, c.right, c.width); got != c.want {
				t.Fatalf("Spread(%q, %q, %d) = %q, want %q", c.left, c.right, c.width, got, c.want)
			}
		})
	}
}

func TestJoin(t *testing.T) {
	for _, c := range []struct {
		name  string
		width int
		facts []string
		want  string
	}{
		{"all of them fit", 20, []string{"2d", "3/5"}, "2d  3/5"},
		{"what does not fit is dropped", 7, []string{"2d", "3/5", "spec"}, "2d  3/5"},
		{"and so is everything after it", 7, []string{"2d", "abcdef", "x"}, "2d"},
		{"an empty fact is passed over", 20, []string{"", "2d"}, "2d"},
		{"nothing at all is nothing", 20, nil, ""},
		{"the first one alone can be too wide", 1, []string{"2d"}, ""},
	} {
		t.Run(c.name, func(t *testing.T) {
			if got := Join("  ", c.width, c.facts...); got != c.want {
				t.Fatalf("Join(%d, %v) = %q, want %q", c.width, c.facts, got, c.want)
			}
		})
	}
}

func TestTrimLast(t *testing.T) {
	if got := trimLast(""); got != "" {
		t.Fatalf("trimLast on nothing gave %q", got)
	}
	if got := trimLast("héllo"); got != "héll" {
		t.Fatalf("trimLast gave %q", got)
	}
	if got := trimLast("é"); got != "" {
		t.Fatalf("trimLast on one accented letter gave %q", got)
	}
}

func TestPad(t *testing.T) {
	if got := pad("ab", 4); got != "ab  " {
		t.Fatalf("pad gave %q", got)
	}
	if got := pad("abcdef", 4); got != "abcdef" {
		t.Fatalf("pad shortened something: %q", got)
	}
}

func TestWritten(t *testing.T) {
	if got := written(""); got != "—" {
		t.Fatalf("written on nothing gave %q", got)
	}
	if got := written("x"); got != "x" {
		t.Fatalf("written changed something: %q", got)
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

func TestSpreadWithExactlyOneColumnBetweenThem(t *testing.T) {
	// One column of room is enough: the two ends still both fit, with the one
	// space between them that keeps them apart.
	if got := Spread("a", "xx", 3); got != " xx" {
		t.Fatalf("Spread with room for one gave %q", got)
	}
	if got := Spread("a", "xx", 2); got != "xx" {
		t.Fatalf("Spread with no room gave %q", got)
	}
}
