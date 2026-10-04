package gui

import (
	"image"
	"strings"
	"testing"

	"gioui.org/io/pointer"

	"cairn/internal/reading"
	"cairn/internal/trace"
)

// openTask is a task opened on its detail, carrying something in every place a
// detail writes: a title, fields, a description, a part, and a history.
func openTask(t *testing.T) (Window, *fake) {
	t.Helper()
	f := newFake()
	f.acts = []trace.Act{anAct("changed", "title", "Write the spec", "")}
	w := holding(f,
		reading.Card{Number: 1, Title: "Write the spec", State: "todo", Labels: []string{"backend"},
			Description: "Deux colonnes et une bande"},
		reading.Card{Number: 2, Title: "Relire l'ADR", State: "todo", Parent: 1})
	w = w.clicked(middleOf(w, 1, 0)).DoubleClick(middleOf(w, 1, 0))
	if w.Screen() != OnDetail {
		t.Fatalf("the task did not open: %v", w.Screen())
	}
	return w, f
}

// readable finds a piece of text the detail writes, by what it is.
func readable(t *testing.T, w Window, key string) Readable {
	t.Helper()
	for _, r := range w.Readables() {
		if r.Key == key {
			return r
		}
	}
	t.Fatalf("the detail writes nothing called %q", key)
	return Readable{}
}

// on is the point over a character of a piece of text, on its first line in sight.
func on(r Readable, at int) Point {
	return Point{X: r.Box.X + at*glyphOf(r.Size) + glyphOf(r.Size)/2, Y: r.Box.Y + lineHigh/2}
}

// TOUT TEXTE ÉCRIT DOIT ÊTRE SÉLECTIONNABLE ET COPIABLE. Every piece of text an
// open task writes can be held, a word at a time by a double click, and copied.
// None of it could: a window draws text as shapes, and nothing had said where
// a shape's letters were.
func TestEveryTextAnOpenTaskWritesCanBeCopied(t *testing.T) {
	w, _ := openTask(t)
	for _, c := range []struct {
		key, word string
	}{
		{"heading", "task"},
		{"name:labels", "labels"},
		{"field:title", "Write"},
		{"field:labels", "backend"},
		{"field:description", "Deux"},
		{"under", "WHAT"},
		{"part:2", "#2"},
	} {
		t.Run(c.key, func(t *testing.T) {
			r := readable(t, w, c.key)
			got, _ := w.DoubleClick(on(r, 0)).Key("cmd+c").Copying()
			if got != c.word {
				t.Fatalf("a double click on the first word of %q copied %q, want %q", c.key, got, c.word)
			}
		})
	}
	// A line of the history, a word into it.
	history := readable(t, w, "history:0")
	at := strings.Index(history.Said, "changed")
	if at < 0 {
		t.Fatalf("the history reads %q", history.Said)
	}
	if got, _ := w.DoubleClick(on(history, at)).Key("cmd+c").Copying(); got != "changed" {
		t.Fatalf("a double click on a word of the history copied %q", got)
	}
}

// A run of text is held by dragging across it, the way the form holds one, and
// copied with ⌘C.
func TestARunOfAnOpenTasksTextIsHeldByDraggingAcrossIt(t *testing.T) {
	w, _ := openTask(t)
	r := readable(t, w, "field:title")
	w = w.PressAt(on(r, 0)).MoveTo(on(r, 5)).LetGo(on(r, 5))
	if got, _ := w.Key("cmd+c").Copying(); got != "Write" {
		t.Fatalf("dragging across the first five letters copied %q", got)
	}
	// A press that goes nowhere holds nothing, and copies nothing.
	w = w.PressAt(on(r, 2)).LetGo(on(r, 2))
	if got, wants := w.Key("cmd+c").Copying(); wants {
		t.Fatalf("a click held %q", got)
	}
}

// ⌘A holds the whole of the text last pressed in, so a description of eighty
// lines is copied without being dragged across.
func TestAllOfATextIsHeldAtOnce(t *testing.T) {
	w, _ := openTask(t)
	r := readable(t, w, "field:description")
	w = w.PressAt(on(r, 0)).LetGo(on(r, 0))
	if got, _ := w.Key("cmd+a").Key("cmd+c").Copying(); got != "Deux colonnes et une bande" {
		t.Fatalf("⌘A then ⌘C copied %q", got)
	}
}

// What is held is what is painted: the run the window will copy is the run it
// shows held, on the text it belongs to and no other.
func TestWhatIsHeldIsWhatIsPainted(t *testing.T) {
	w, _ := openTask(t)
	r := readable(t, w, "field:labels")
	w = w.DoubleClick(on(r, 0))
	if held := w.Holding("field:labels", r.Said); held.Holds() != "backend" {
		t.Fatalf("the labels are drawn holding %q", held.Holds())
	}
	if held := w.Holding("field:title", "Write the spec"); held.Holding() {
		t.Fatalf("the title is drawn holding %q, though nothing was picked in it", held.Holds())
	}
}

// A word in a field of the form is held by a double click, as on an open task.
func TestAWordInTheFormIsHeldByADoubleClick(t *testing.T) {
	w := holding(newFake(), reading.Card{Number: 1, Title: "Write the spec", State: "todo"})
	w = w.clicked(middleOf(w, 1, 0)).Choose("Edit")
	box := Over(w.width, w.height)
	fields := w.Form().Fields()
	_, value := FieldRow(box, FieldTop(box, fields, 0))
	inside := Inside(Framed(value, RowsOf(box, fields, 0)))
	at := Point{X: inside.X + 11*perGlyph + perGlyph/2, Y: inside.Y + lineHigh/2}
	w, _ = Pointing(w, Gesture{}, pointer.Press, at, 0, moment)
	if got := w.DoubleClick(at).Form().Writing().Holds(); got != "spec" {
		t.Fatalf("a double click on the third word of the title held %q", got)
	}
}

// What is held is painted, on both ways a line is written: a field's value,
// laid out on lines, and a line of the history, written whole.
func TestWhatIsHeldOnAnOpenTaskIsPainted(t *testing.T) {
	for _, key := range []string{"field:labels", "history:0"} {
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

// held counts the points of a piece of text's first line painted in the ground
// a held run is painted in.
func held(shot *image.RGBA, r Readable) int {
	ghost, count := Dark().Ghost, 0
	for y := r.Box.Y; y < r.Box.Y+lineHigh; y++ {
		for x := r.Box.X; x < r.Box.X+r.Box.W; x++ {
			if got := shot.RGBAAt(x, y); got.R == ghost.R && got.G == ghost.G && got.B == ghost.B {
				count++
			}
		}
	}
	return count
}

// How wide a character is reckoned to be, at every size the window writes in
// and either side of where the reckoning changes.
func TestACharacterIsReckonedAtEverySize(t *testing.T) {
	for size, want := range map[int]int{
		small - 1: perSmallGlyph, small: perSmallGlyph,
		reads: perGlyph, plain: perGlyph, heading - 1: perGlyph,
		heading: perGlyph + 1, heading + 1: perGlyph + 1,
	} {
		if got := glyphOf(size); got != want {
			t.Errorf("at %d a character is reckoned %d wide, want %d", size, got, want)
		}
	}
}

// The character under a point, on any line in sight, scrolled or not, and held
// to the text at all four of its edges.
func TestTheCharacterUnderAPoint(t *testing.T) {
	r := Readable{Box: Box{X: 100, Y: 200, W: 300, H: 3 * lineHigh}, Size: reads,
		Lines: Wrapped("un deux\ntrois\nquatre", 40)}
	at := func(column, row int) Point {
		return Point{X: r.Box.X + column*perGlyph + 1, Y: r.Box.Y + row*lineHigh + 1}
	}
	for _, c := range []struct {
		what string
		p    Point
		from int
		want int
	}{
		{"the first character", at(0, 0), 0, 0},
		{"two in on the second line", at(2, 1), 0, 10},
		{"the first line in sight, scrolled one down", at(1, 0), 1, 9},
		{"above the text", Point{X: r.Box.X + 1, Y: r.Box.Y - 50}, 0, 0},
		{"below the text", at(1, 9), 0, 15},
		{"left of the text", Point{X: r.Box.X - 40, Y: r.Box.Y + 1}, 0, 0},
		{"past the end of a line", at(30, 1), 0, 13},
	} {
		r.From = c.from
		if got := r.At(c.p); got != c.want {
			t.Errorf("%s: character %d, want %d", c.what, got, c.want)
		}
	}
}

// The room a piece of text takes to be pressed in: the lines in sight, or its
// own box when that is taller.
func TestTheRoomAPieceOfTextTakes(t *testing.T) {
	r := Readable{Box: Box{X: 100, Y: 200, W: 300, H: lineHigh}, Size: reads,
		Lines: Wrapped("1\n2\n3\n4\n5", 40), From: 2}
	if got := r.inSight().H; got != 3*lineHigh {
		t.Errorf("five lines scrolled two down take %d, want three lines", got)
	}
	r.Box.H, r.From = 10*lineHigh, 0
	if got := r.inSight().H; got != 10*lineHigh {
		t.Errorf("a box taller than its lines takes %d, want the box", got)
	}
}

// A part counts as in sight when the whole of its line is inside the box, its
// edges included, and not when any of it is outside.
func TestAPartIsInSightWhenAllOfItIs(t *testing.T) {
	frame := Box{X: 0, Y: 100, W: 300, H: 100}
	for _, c := range []struct {
		y    int
		want bool
	}{
		{100, true}, {99, false}, {200 - lineHigh, true}, {201 - lineHigh, false},
	} {
		if got := Within(Box{Y: c.y, H: lineHigh}, frame); got != c.want {
			t.Errorf("a line at %d: in sight %v, want %v", c.y, got, c.want)
		}
	}
}

// Where a held run is painted on its line: as far in as what comes before it,
// as wide as it is.
func TestWhereAHeldRunIsPainted(t *testing.T) {
	box := Box{X: 100, Y: 200, W: 300, H: lineHigh}
	if got, want := HeldRun(box, 40, 25), (Box{X: 140, Y: 200, W: 25, H: lineHigh}); got != want {
		t.Fatalf("the run is painted at %+v, want %+v", got, want)
	}
}
