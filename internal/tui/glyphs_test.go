package tui

import (
	"errors"
	"strings"
	"testing"

	"cairn/internal/engine"
)

func TestGlyphsFor(t *testing.T) {
	if GlyphsFor(NerdFont) != nerd {
		t.Fatal("the full set was not answered for nerdfont")
	}
	if GlyphsFor(Unicode) != plain {
		t.Fatal("the plain set was not answered for unicode")
	}
	if GlyphsFor("something else") != plain {
		t.Fatal("a word nobody planned for should still draw a board")
	}
}

func TestTheTwoSetsSayTheSameThings(t *testing.T) {
	// Every mark exists in both, or a board would draw a hole where the other
	// draws something.
	for name, pair := range map[string][2]string{
		"blocked":    {nerd.Blocked, plain.Blocked},
		"watching":   {nerd.Watching, plain.Watching},
		"paused":     {nerd.Paused, plain.Paused},
		"active":     {nerd.Active, plain.Active},
		"on-hold":    {nerd.OnHold, plain.OnHold},
		"reached":    {nerd.Reached, plain.Reached},
		"dropped":    {nerd.Dropped, plain.Dropped},
		"described":  {nerd.Described, plain.Described},
		"sub-tasks":  {nerd.SubTasks, plain.SubTasks},
		"label":      {nerd.Label, plain.Label},
		"in the way": {nerd.InTheWay, plain.InTheWay},
	} {
		if pair[0] == "" || pair[1] == "" {
			t.Fatalf("%s is missing from one of the two sets", name)
		}
	}
}

func TestSetsAreTheTwoThatCanBeAsked(t *testing.T) {
	if len(Sets) != 2 {
		t.Fatalf("there are %d sets, and the setting says there are two", len(Sets))
	}
}

func TestParseSet(t *testing.T) {
	for _, want := range Sets {
		got, err := ParseSet(string(want))
		if err != nil || got != want {
			t.Fatalf("ParseSet(%q) = %v, %v", want, got, err)
		}
	}
	_, err := ParseSet("nerdfonts")
	if !errors.Is(err, ErrNotASet) {
		t.Fatalf("a word that is nearly right answered %v", err)
	}
	if !strings.Contains(err.Error(), "nerdfonts") {
		t.Fatalf("the refusal does not name what was typed: %v", err)
	}
	if said := engine.RefusalOf(err); said.Kind != engine.KindRefused || said.Reason != "not-a-glyph-set" {
		t.Fatalf("RefusalOf = %s/%s, want it answered as a refusal like any other", said.Kind, said.Reason)
	}
	if !strings.Contains(err.Error(), "unicode") {
		t.Fatalf("the refusal does not say what it takes: %v", err)
	}
}
