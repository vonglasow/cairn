package gui

import (
	"cairn/internal/board"
	"cairn/internal/gui/design"
	"cairn/internal/gui/geom"
	"image"
	"testing"

	"cairn/internal/engine"
	"cairn/internal/engine/act"
)

func openTask(t *testing.T) (Window, *fake) {
	t.Helper()
	f := newFake()
	f.acts = []act.Act{anAct("changed", "title", "Write the spec", "")}
	w := holding(f,
		board.Card{Number: 1, Title: "Write the spec", State: "todo", Labels: []string{"backend"},
			Description: "Two columns and a band", Context: "/vault/1",
			Steps: []engine.Part{{Number: 1, Title: "Relire l'ADR"}}})
	w = w.clicked(middleOf(w, 1, 0)).DoubleClick(middleOf(w, 1, 0)).Key("cmd+t")
	if w.Screen != board.OnForm {
		t.Fatalf("the task did not open: %v", w.Screen)
	}
	return w, f
}

func readable(t *testing.T, w Window, key string) Readable {
	t.Helper()
	for _, r := range w.Readables() {
		if r.Key == key {
			return r
		}
	}
	t.Fatalf("the column beside the form writes nothing called %q", key)
	return Readable{}
}

func on(r Readable, at int) geom.Point {
	line := 0
	for i, from := range r.Lines.From {
		if from <= at {
			line = i
		}
	}
	return geom.Point{X: r.Box.X + (at-r.Lines.From[line])*glyphOf(r.Size), Y: r.Box.Y + (line-r.From)*lineHigh + lineHigh/2}
}

func TestARunOfAnOpenTasksTextIsHeldByDraggingAcrossIt(t *testing.T) {
	w, _ := openTask(t)
	r := readable(t, w, "field:context")
	w = w.PressAt(on(r, 0)).MoveTo(on(r, 6)).LetGo(on(r, 6))
	if got, _ := w.Key("cmd+c").Copying(); got != "/vault" {
		t.Fatalf("dragging across the first six letters copied %q", got)
	}
	w = w.PressAt(on(r, 2)).LetGo(on(r, 2))
	if got, wants := w.Key("cmd+c").Copying(); wants {
		t.Fatalf("a click held %q", got)
	}
}

func TestAllOfATextIsHeldAtOnce(t *testing.T) {
	w, _ := openTask(t)
	r := readable(t, w, "trace")
	w = w.PressAt(on(r, 0)).LetGo(on(r, 0))
	if got, _ := w.Key("cmd+a").Key("cmd+c").Copying(); got != r.Said {
		t.Fatalf("⌘A then ⌘C copied %q", got)
	}
}

func TestWhatIsHeldOnAnOpenTaskIsPainted(t *testing.T) {
	for _, key := range []string{"field:context", "trace"} {
		t.Run(key, func(t *testing.T) {
			w, _ := openTask(t)
			r := readable(t, w, key)
			before := held(drawn(t, w), r)
			w = w.DoubleClick(on(r, 0))
			if after := held(drawn(t, w), r); after <= before {
				t.Fatalf("holding a word of %q painted %d points of the held ground, against %d before", key, after, before)
			}
		})
	}
}

func held(shot *image.RGBA, r Readable) int {
	ghost, count := design.Dark().Colour.Surface.Ghost, 0
	for y := r.Box.Y; y < r.Box.Y+lineHigh; y++ {
		for x := r.Box.X; x < r.Box.X+r.Box.W; x++ {
			if got := shot.RGBAAt(x, y); got.R == ghost.R && got.G == ghost.G && got.B == ghost.B {
				count++
			}
		}
	}
	return count
}

func TestAPartIsInSightWhenAllOfItIs(t *testing.T) {
	frame := geom.Box{X: 0, Y: 100, W: 300, H: 100}
	for _, c := range []struct {
		y    int
		want bool
	}{
		{100, true}, {99, false}, {200 - lineHigh, true}, {201 - lineHigh, false},
	} {
		if got := Within(geom.Box{Y: c.y, H: lineHigh}, frame); got != c.want {
			t.Errorf("a line at %d: in sight %v, want %v", c.y, got, c.want)
		}
	}
}

func TestARunGrowsFromItsAnchorByWholeUnits(t *testing.T) {
	said := "Write the spec"
	word := Picked{From: 6, To: 9}
	for _, c := range []struct {
		what          string
		by            Unit
		anchor        Picked
		at            int
		fixed, moving int
	}{
		{"a character on", Characters, Picked{From: 2, To: 2}, 7, 2, 7},
		{"a character back", Characters, Picked{From: 7, To: 7}, 2, 7, 2},
		{"inside its own word", Words, word, 7, 6, 9},
		{"into the word after", Words, word, 11, 6, 14},
		{"into the word before", Words, word, 2, 9, 0},
		{"a paragraph", Paragraphs, Picked{From: 0, To: 14}, 3, 0, 14},
	} {
		fixed, moving := Stretched(said, c.by, c.anchor, c.at)
		if fixed != c.fixed || moving != c.moving {
			t.Errorf("%s: %d to %d, want %d to %d", c.what, fixed, moving, c.fixed, c.moving)
		}
	}
}
