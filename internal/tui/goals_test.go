package tui

import (
	"testing"
)

func TestStandingGlyph(t *testing.T) {
	g := GlyphsFor(Unicode)
	for state, want := range map[string]string{
		"active":  g.Active,
		"on-hold": g.OnHold,
		"reached": g.Reached,
		"dropped": g.Dropped,
		"":        "",
	} {
		if got := StandingGlyph(state, g); got != want {
			t.Fatalf("%q is marked %q, want %q", state, got, want)
		}
	}
}

func TestBar(t *testing.T) {
	for _, c := range []struct {
		done, total, width int
		want               string
	}{
		{0, 0, 4, "░░░░"},
		{0, 4, 4, "░░░░"},
		{2, 4, 4, "██░░"},
		{4, 4, 4, "████"},
		{1, 3, 6, "██░░░░"},
	} {
		if got := Bar(c.done, c.total, c.width); got != c.want {
			t.Fatalf("Bar(%d, %d, %d) = %q, want %q", c.done, c.total, c.width, got, c.want)
		}
	}
	if got := Bar(1, 2, 0); got != "" {
		t.Fatalf("a bar of no width is %q", got)
	}
}
