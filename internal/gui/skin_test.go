package gui

import (
	"image/color"
	"math"
	"reflect"
	"testing"

	"cairn/internal/theme"
)

func TestTheDefaultSchemeIsDarkTokenByToken(t *testing.T) {
	if !reflect.DeepEqual(Dark().Colour, theme.Default()) {
		t.Error("the default scheme moves a colour of the window as cairn has always drawn it")
	}
}

func TestSkinnedWearsTheColoursGivenAndKeepsEveryOtherToken(t *testing.T) {
	s, err := theme.Find("", "gruvbox-dark")
	if err != nil {
		t.Fatal(err)
	}
	skinned, dark := Skinned(s.Colours()), Dark()
	if !reflect.DeepEqual(skinned.Colour, s.Colours()) {
		t.Error("the skin does not wear the scheme's colours")
	}
	skinned.Colour = dark.Colour
	if !reflect.DeepEqual(skinned, dark) {
		t.Error("wearing a scheme moved a token that is not a colour")
	}
}

func TestEverySkinCanBeRead(t *testing.T) {
	for _, slug := range theme.Shipped() {
		s, err := theme.Find("", slug)
		if err != nil {
			t.Fatal(err)
		}
		skin := Skinned(s.Colours())
		for _, short := range theme.Shortfalls(skin.Colour) {
			t.Errorf("%s falls short: %s", slug, short)
		}
		for name, said := range skin.Colour.Priority {
			if near(said, skin.Colour.Accent.Base) {
				t.Errorf("%s: priority %q is the colour of what is chosen", slug, name)
			}
			if alike(said, skin.Colour.Line.Edge) {
				t.Errorf("%s: priority %q is the colour of no priority at all", slug, name)
			}
			for other, also := range skin.Colour.Priority {
				if other != name && alike(said, also) {
					t.Errorf("%s: priority %q and %q are the same colour", slug, name, other)
				}
			}
		}
		if near(skin.Colour.State.Wrong, skin.Colour.Priority["critical"]) {
			t.Errorf("%s: an error is the colour of a critical task", slug)
		}
	}
}

func near(a, b color.NRGBA) bool  { return howFar(a, b) < 60 }
func alike(a, b color.NRGBA) bool { return howFar(a, b) < 45 }

func howFar(a, b color.NRGBA) float64 {
	dr, dg, db := float64(a.R)-float64(b.R), float64(a.G)-float64(b.G), float64(a.B)-float64(b.B)
	return math.Sqrt(dr*dr + dg*dg + db*db)
}

func TestTheWindowIsHandedTheSkinItWears(t *testing.T) {
	if opens := reflect.TypeOf(Run); opens.NumIn() < 3 || opens.In(2) != reflect.TypeFor[Skin]() {
		t.Errorf("the window is opened by %v, handed no skin to wear", opens)
	}
}
