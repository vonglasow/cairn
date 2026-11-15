package geom

import (
	"cairn/internal/board"
	"fmt"
	"strings"
	"testing"
)

func TestTooSmallIsTheTwoSidesTogether(t *testing.T) {
	for _, c := range []struct {
		what          string
		width, height int
		small         bool
	}{
		{"room enough", Narrowest, Shortest, false},
		{"one short of the width", Narrowest - 1, Shortest, true},
		{"one short of the height", Narrowest, Shortest - 1, true},
		{"short of both", 320, 240, true},
		{"far more than enough", 1600, 1000, false},
		{"a small laptop's half screen", 720, 560, false},
	} {
		if got := TooSmall(c.width, c.height); got != c.small {
			t.Errorf("%s: TooSmall(%d, %d) = %v", c.what, c.width, c.height, got)
		}
	}
}

// Showing is how many whole cards a column has room for.
func (l Layout) Showing(column int) int { return l.Column[column].H / CardHeight }

// Written is the whole shape as a person can read it, so a test can pin it and a sum changed by accident shows up as a number that moved.
func (l Layout) Written() string {
	var out strings.Builder
	say := func(name string, b Box) {
		fmt.Fprintf(&out, "%-12s x=%-5d y=%-5d w=%-5d h=%d\n", name, b.X, b.Y, b.W, b.H)
	}
	say("tabs", l.Tabs)
	for i := range board.Columns {
		say(fmt.Sprintf("head %d", i), l.Head[i])
		say(fmt.Sprintf("column %d", i), l.Column[i])
	}
	say("band", l.Band)
	for i, name := range board.Stalled {
		say("lane "+name, l.Lane[i])
	}
	say("bar", l.Bar)
	return out.String()
}
