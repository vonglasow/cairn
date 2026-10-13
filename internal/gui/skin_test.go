package gui

import (
	"image/color"
	"math"
	"testing"
)

func TestEverySkinCanBeRead(t *testing.T) {
	for _, skin := range []struct {
		what string
		Skin
	}{{"dark", Dark()}} {
		ink := []struct {
			what string
			on   color.NRGBA
			said color.NRGBA
		}{
			{"ink on a panel", skin.Panel, skin.Ink},
			{"ink on the board", skin.Back, skin.Ink},
			{"faint on a panel", skin.Panel, skin.Faint},
			{"faint on the board", skin.Back, skin.Faint},
			{"an offer on the board", skin.Back, skin.Chosen},
			{"an error on the board", skin.Back, skin.Wrong},
			{"a day gone by, on a panel", skin.Panel, skin.Late},
			{"a day gone by, on the board", skin.Back, skin.Late},
		}
		for _, c := range ink {
			if got := Contrast(c.said, c.on); got < 4.5 {
				t.Errorf("%s: %s reads at %.2f, and text needs 4.5", skin.what, c.what, got)
			}
		}

		shapes := []struct {
			what string
			on   color.NRGBA
			said color.NRGBA
			at   float64
		}{
			{"an edge against the board", skin.Back, skin.Edge, 3},
			{"an edge against a panel", skin.Panel, skin.Edge, 2},
			{"a panel against the board", skin.Back, skin.Panel, 1.3},
		}
		for _, c := range shapes {
			if got := Contrast(c.said, c.on); got < c.at {
				t.Errorf("%s: %s stands at %.2f, and it needs %.1f", skin.what, c.what, got, c.at)
			}
		}

		for name, said := range skin.Priority {
			if near(said, skin.Chosen) {
				t.Errorf("%s: priority %q is the colour of what is chosen", skin.what, name)
			}
			if alike(said, skin.Edge) {
				t.Errorf("%s: priority %q is the colour of no priority at all", skin.what, name)
			}
			for other, also := range skin.Priority {
				if other != name && alike(said, also) {
					t.Errorf("%s: priority %q and %q are the same colour", skin.what, name, other)
				}
			}
		}
		if near(skin.Wrong, skin.Priority["critical"]) {
			t.Errorf("%s: an error is the colour of a critical task", skin.what)
		}
	}
}

func near(a, b color.NRGBA) bool  { return howFar(a, b) < 60 }
func alike(a, b color.NRGBA) bool { return howFar(a, b) < 45 }

func howFar(a, b color.NRGBA) float64 {
	dr, dg, db := float64(a.R)-float64(b.R), float64(a.G)-float64(b.G), float64(a.B)-float64(b.B)
	return math.Sqrt(dr*dr + dg*dg + db*db)
}

func TestContrast_answersTheKnownRatios(t *testing.T) {
	black := color.NRGBA{A: 255}
	white := color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	if got := Contrast(white, black); math.Abs(got-21) > 0.01 {
		t.Errorf("white on black is %.2f, want 21", got)
	}
	if got := Contrast(black, black); math.Abs(got-1) > 0.01 {
		t.Errorf("black on black is %.2f, want 1", got)
	}
	if Contrast(white, black) != Contrast(black, white) {
		t.Error("the ratio changed when the two were swapped")
	}
}

func TestTheCurveIsStraightenedWhereItChangesShape(t *testing.T) {
	for _, c := range []struct {
		v    uint8
		want float64
	}{
		{0, 0},
		{9, 0.0027317429},
		{10, 0.0030352698},
		{11, 0.0033465358},
		{255, 1},
	} {
		if got := straight(c.v); math.Abs(got-c.want) > 0.000000001 {
			t.Errorf("straight(%d) = %.10f, want %.10f", c.v, got, c.want)
		}
	}
}

func TestContrastIsTwentyOneAtItsWidestAndOneAtItsNarrowest(t *testing.T) {
	white := color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	black := color.NRGBA{A: 255}
	if got := Contrast(black, white); math.Abs(got-21) > 0.01 {
		t.Errorf("black on white is %.2f, want 21", got)
	}
	if got := Contrast(white, white); math.Abs(got-1) > 0.01 {
		t.Errorf("white on white is %.2f, want 1", got)
	}
}
