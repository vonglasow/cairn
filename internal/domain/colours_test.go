package domain

import (
	"errors"
	"testing"
)

func TestALabelsColourIsOneOfTheFourteen(t *testing.T) {
	if len(LabelColours) != 14 || LabelColours[0] != "rosewater" || LabelColours[13] != "lavender" {
		t.Fatalf("the colours are %v", LabelColours)
	}
	if got, err := ParseColour(" Peach "); err != nil || got != "peach" {
		t.Errorf("Peach reads %q, %v", got, err)
	}
	if _, err := ParseColour("brown"); !errors.Is(err, ErrNoSuchColour) {
		t.Errorf("brown reads %v", err)
	}
}

func TestALabelWithNoColourTakesOneOfItsName(t *testing.T) {
	if got := ColourOf("release"); got != "green" {
		t.Fatalf("release wears %q, read off its name it is green", got)
	}
	seen := map[string]bool{}
	for _, label := range []string{"release", "doc", "bug", "gui", "sp26-20", "design", "v1", "team"} {
		seen[ColourOf(label)] = true
	}
	if len(seen) < 4 {
		t.Errorf("eight labels share %d colours", len(seen))
	}
}

func TestEveryColourIsInkedAsCatppuccinMochaInksIt(t *testing.T) {
	for _, name := range LabelColours {
		if _, ok := LabelInks[name]; !ok {
			t.Errorf("%s has no ink", name)
		}
	}
	if len(LabelInks) != len(LabelColours) || LabelInks["green"] != [3]uint8{0xa6, 0xe3, 0xa1} || LabelInks["rosewater"] != [3]uint8{0xf5, 0xe0, 0xdc} {
		t.Errorf("the inks are %v", LabelInks)
	}
}
