package gui

import (
	"fmt"
	"testing"

	"cairn/internal/reading"
)

// onTheList is task 1 opened with the parts named hanging under it, the
// keyboard on its list of parts, in a window of the height given.
func onTheList(t *testing.T, high int, titles ...string) Window {
	t.Helper()
	cards := []reading.Card{card(1, "Ship it", "todo")}
	for i, title := range titles {
		cards = append(cards, reading.Card{Number: i + 2, Title: title, State: "todo", Parent: 1})
	}
	w := holding(newFake(), cards...)
	w.height = high
	w = w.clicked(middleOf(w, 1, 0)).DoubleClick(middleOf(w, 1, 0)).On("parts")
	if !w.form.onParts() {
		t.Fatal("the keyboard is not on the parts")
	}
	return w
}

// The arrows stop at the line that adds, and up from the first part leaves
// for the field before; staying on the list, the keyboard keeps its line.
func TestTheArrowsStopAtBothEndsOfTheList(t *testing.T) {
	w := onTheList(t, 800, "A", "B")
	for range 5 {
		w = w.Key("down")
	}
	if w.form.partAt != 2 {
		t.Errorf("five down from the first of two parts is line %d, want the line that adds", w.form.partAt)
	}
	if again := w.form.onto(w.form.On()); again.partAt != 2 {
		t.Errorf("staying on the list put the keyboard on line %d", again.partAt)
	}
	w.form.partAt = 1
	if up := w.Key("up"); !up.form.onParts() || up.form.partAt != 0 {
		t.Errorf("up from the second part left the list, or went to line %d", up.form.partAt)
	}
	parts, _ := w.form.partsField()
	w.form.partAt = 0
	if up := w.Key("up"); up.form.On() != parts-1 {
		t.Errorf("up from the first part left the keyboard on field %d, want %d", up.form.On(), parts-1)
	}
	if _, on := w.Key("down").Key("down").PartOn(); on {
		t.Error("on the line that adds, the keyboard is said to be on a part")
	}
}

// Space and backspace do nothing on the line that adds; a part struck out is
// not ticked; a part never saved is taken away rather than struck out.
func TestTickingAndStrikingKnowWhatTheyAreOn(t *testing.T) {
	w := onTheList(t, 800, "A")
	adding := w.Key("down")
	if got := adding.Key("space").Key("backspace").form.fields; fmt.Sprint(got) != fmt.Sprint(adding.form.fields) || len(adding.Key("backspace").form.struck) != 0 {
		t.Error("space or backspace on the line that adds changed something")
	}
	struck := w.Key("backspace").Key("space")
	if parts := struck.form.parts(); parts[0].Done {
		t.Error("a part struck out was ticked")
	}
	added := w.Key("down").Key("enter").Types("New").Key("enter").Key("esc")
	if parts := added.form.parts(); len(parts) != 2 {
		t.Fatalf("adding left %+v", parts)
	}
	added.form.partAt = 1
	if parts := added.Key("backspace").form.parts(); len(parts) != 1 || len(added.Key("backspace").form.struck) != 0 {
		t.Errorf("striking a part never saved left %+v struck %v", parts, added.Key("backspace").form.struck)
	}
}

// Once a part is added, the next is waiting on the line that adds; esc puts
// down what is being written, and leaves the form open.
func TestTheNextPartIsWaiting(t *testing.T) {
	w := onTheList(t, 800, "A").Key("down").Key("enter").Types("B").Key("enter")
	if !w.form.writingPart || w.form.partAt != w.form.lines()-1 || w.form.named.Reads() != "" {
		t.Errorf("after adding, writing %v on line %d of %d, holding %q", w.form.writingPart, w.form.partAt, w.form.lines(), w.form.named.Reads())
	}
	if put := w.Key("esc"); put.Screen() != OnForm || put.form.writingPart {
		t.Errorf("esc while writing left the window on %v, writing %v", put.Screen(), put.form.writingPart)
	}
}

// A part carried past the last lands last, and only a part is carried.
func TestAPartIsCarriedNoFurtherThanTheList(t *testing.T) {
	f := onTheList(t, 800, "A", "B", "C").form
	if got := f.carrying(0, 99).parts(); got[2].Title != "A" {
		t.Errorf("carried past the last, the parts read %+v", got)
	}
	if got := f.carrying(3, 0).parts(); got[0].Title != "A" {
		t.Errorf("carrying the line that adds moved %+v", got)
	}
}

// A click on a part's box ticks it, on the rest of its tile chooses it, and
// on the line that adds begins a new part — wherever the list is scrolled to.
func TestAClickOnTheListDoesWhatItLandsOn(t *testing.T) {
	titles := make([]string, 12)
	for i := range titles {
		titles[i] = fmt.Sprintf("Part %d", i+1)
	}
	w := onTheList(t, 720, titles...)
	w.form.partAt = 11
	w = w.keepingPartInSight()
	from := w.partsFrom()
	tiles := w.partTiles()
	if from == 0 || len(tiles) < 3 {
		t.Fatalf("this proves nothing: the list shows %d lines from %d", len(tiles), from)
	}
	middle := func(b Box) Point { return Point{X: b.X + b.W/2, Y: b.Y + b.H/2} }
	box := SubCheck(tiles[0])
	ticked := w.clicked(Point{X: box.X + 2, Y: box.Y + 2})
	if parts := ticked.form.parts(); !parts[from].Done || ticked.form.partAt != from {
		t.Errorf("a click on the box of line %d ticked %+v, the keyboard on %d", from, parts[from], ticked.form.partAt)
	}
	chosen := w.clicked(middle(tiles[1]))
	if parts := chosen.form.parts(); parts[from+1].Done || chosen.form.partAt != from+1 {
		t.Errorf("a click beside the box ticked %+v, the keyboard on %d", parts[from+1], chosen.form.partAt)
	}
	adding := w.clicked(middle(tiles[len(tiles)-1]))
	if !adding.form.writingPart || adding.form.partAt != 12 {
		t.Errorf("a click on the line that adds wrote %v on line %d", adding.form.writingPart, adding.form.partAt)
	}
	top := w
	top.form.partAt, top.form.partsTop = 0, 0
	if adds := top.clicked(middle(top.partTiles()[len(top.partTiles())-1])); adds.form.partAt != 12 {
		t.Errorf("a click on the line that adds, the list at its top, put the keyboard on line %d", adds.form.partAt)
	}
	renaming := w.DoubleClick(middle(tiles[1]))
	if !renaming.form.writingPart || renaming.form.named.Reads() != titles[from+1] {
		t.Errorf("a double click on a part writes %v on %q", renaming.form.writingPart, renaming.form.named.Reads())
	}
}

// The list shows the line the keyboard is on, and starts as far up as it can
// while it does.
func TestTheListFollowsTheKeyboard(t *testing.T) {
	titles := make([]string, 12)
	for i := range titles {
		titles[i] = fmt.Sprintf("Part %d", i+1)
	}
	w := onTheList(t, Shortest, titles...)
	at, _ := w.form.partsField()
	_, value, rows := FormPlace(w.formBox(), w.form.Fields(), at)
	frame := Inside(Framed(value, rows))
	for line := range w.form.lines() {
		w.form.partAt = line
		tiles, from := w.partTiles(), w.partsFrom()
		shown := tiles[len(tiles)-1]
		if line < 12 {
			shown = tiles[line-from]
		}
		if shown.Y < frame.Y || shown.Y+shown.H > frame.Y+frame.H {
			t.Errorf("line %d is drawn at %v, out of %v", line, shown, frame)
		}
		if from > 0 {
			w.form.partAt = line
			earlier, _ := TickedPlaces(Framed(value, rows), 12, from-1)
			if line < 12 && line-from+1 < len(earlier) {
				t.Errorf("line %d would show from %d as well, and the list starts at %d", line, from-1, from)
			}
		}
	}
}

// The tiles of a list scrolled down start with the first in sight, at the
// top of the frame.
func TestTicksStartFromTheFirstInSight(t *testing.T) {
	frame := Box{X: 100, Y: 200, W: 400, H: 8 + 12*lineHigh}
	tiles, _ := TickedPlaces(frame, 5, 2)
	if len(tiles) != 3 || tiles[0].Y != Inside(frame).Y {
		t.Errorf("five parts from the third gave %v", tiles)
	}
}

// The line the keyboard is on is outlined, and no other; the line that adds
// is lit while the keyboard is on it; a line being written shows what is
// written; a part struck out is drawn faint.
func TestTheListDrawsWhereTheKeyboardIs(t *testing.T) {
	w := onTheList(t, 800, "A", "B")
	lit := Dark().Chosen
	tiles := w.partTiles()
	shot := drawn(t, w)
	if painted(shot, Edges(SubPanel(tiles[0]))[0], lit) == 0 || painted(shot, Edges(SubPanel(tiles[1]))[0], lit) != 0 {
		t.Error("the part the keyboard is on is not the one outlined")
	}
	adding := w.Key("down").Key("down")
	if painted(drawn(t, adding), tiles[2], lit) == 0 || painted(shot, tiles[2], lit) != 0 {
		t.Error("the line that adds is not lit while the keyboard is on it, and only then")
	}
	writing := adding.Key("enter").Types("New part")
	if painted(drawn(t, writing), tiles[2], Dark().Ink) == 0 {
		t.Error("what is being written on the line that adds is not drawn")
	}
	renaming := w.Key("down").Key("enter")
	if shot := drawn(t, renaming); painted(shot, Edges(SubPanel(tiles[1]))[0], lit) == 0 || painted(shot, Edges(SubPanel(tiles[0]))[0], lit) != 0 {
		t.Error("the part being renamed is not the one outlined")
	}
	struck := w.Key("backspace")
	if got := painted(drawn(t, struck), SubWriting(tiles[0]), Dark().Ink); got != 0 {
		t.Errorf("a part struck out is written in the reading ink: %d points", got)
	}
	if painted(drawn(t, struck), Edges(SubPanel(tiles[0]))[0], lit) == 0 {
		t.Error("a part struck out with the keyboard on it is not outlined")
	}
	elsewhere := w.form.onto(0)
	w.form = elsewhere
	if painted(drawn(t, w), Edges(SubPanel(tiles[0]))[0], lit) != 0 {
		t.Error("a part is outlined with the keyboard elsewhere")
	}
}

// A part in hand is drawn under the pointer only once it has moved past the
// grip, and lands above a part when carried up, below one when carried down.
func TestAPartInHandSaysWhereItLands(t *testing.T) {
	w := onTheList(t, 800, "A", "B", "C")
	tiles := w.partTiles()
	middle := func(b Box) Point { return Point{X: b.X + b.W/2, Y: b.Y + b.H/2} }
	pressed := w.PressAt(middle(tiles[1]))
	if _, ok := pressed.PartGhost(); ok {
		t.Error("a part is in hand before it has moved")
	}
	if _, ok := pressed.PartLanding(); ok {
		t.Error("a part says where it lands before it has moved")
	}
	up := pressed.MoveTo(middle(tiles[0]))
	if line, ok := up.PartLanding(); !ok || line.Y != tiles[0].Y || line.X != tiles[0].X || line.W != tiles[0].W {
		t.Errorf("carried up onto the first, it lands at %v (%v), want the top of %v", line, ok, tiles[0])
	}
	down := pressed.MoveTo(middle(tiles[2]))
	if line, ok := down.PartLanding(); !ok || line.Y != tiles[2].Y+tiles[2].H-landingHigh {
		t.Errorf("carried down onto the last, it lands at %v (%v), want the bottom of %v", line, ok, tiles[2])
	}
	if ghost, ok := down.PartGhost(); !ok || ghost.H != SubHeight || ghost.Y != middle(tiles[2]).Y-SubHeight/2 || ghost.X != tiles[2].X {
		t.Errorf("the part in hand is drawn at %v", ghost)
	}
	off := pressed.MoveTo(Point{X: 1, Y: 1})
	if _, ok := off.PartLanding(); ok {
		t.Error("held off the list, it says it would land somewhere")
	}
	shot := drawn(t, down)
	if ghost, _ := down.PartGhost(); painted(shot, ghost, Dark().Ghost) == 0 {
		t.Error("the part in hand is not drawn")
	}
	// Held right over where it would land, the line is still drawn: over the
	// part in hand, not under it.
	over := pressed.MoveTo(Point{X: middle(tiles[2]).X, Y: tiles[2].Y + tiles[2].H - 2})
	if line, _ := over.PartLanding(); painted(drawn(t, over), line, Dark().Chosen) < line.W*line.H/2 {
		t.Error("where it would land is hidden under the part in hand")
	}
}

// A tick is said to the engine at once; refused, the box stays as it was and
// the refusal is said.
func TestATickIsDoneAtOnceOrNotAtAll(t *testing.T) {
	w := onTheList(t, 800, "A")
	f := w.work.(*fake)
	ticked := w.Key("space")
	if in, ok := f.gave("task", "edit"); !ok || in.Given["number"] != "2" || in.Given["state"] != "done" {
		t.Errorf("ticking asked %+v", in)
	}
	if ticked.form.Changed() {
		t.Error("a tick left the form something to save")
	}
	w2 := onTheList(t, 800, "A")
	w2.work.(*fake).failOp = errNope
	refused := w2.Key("space")
	if refused.form.parts()[0].Done || refused.Wrong() == nil {
		t.Errorf("a refused tick left the part %+v, saying %v", refused.form.parts()[0], refused.Wrong())
	}
}

// Held over the line that adds, a part would land at the end of the list; held
// over its own tile, where it was; and the line that adds is never in hand.
func TestWhereAPartLandsAtTheEdges(t *testing.T) {
	w := onTheList(t, 800, "A", "B")
	tiles := w.partTiles()
	middle := func(b Box) Point { return Point{X: b.X + b.W/2, Y: b.Y + b.H/2} }
	pressed := w.PressAt(middle(tiles[0])).MoveTo(Point{X: middle(tiles[0]).X, Y: middle(tiles[0]).Y + Grip + 1})
	adding := tiles[len(tiles)-1]
	if line, ok := pressed.MoveTo(middle(adding)).PartLanding(); !ok || line.Y != adding.Y || line.X != tiles[0].X {
		t.Errorf("over the line that adds it lands at %v (%v), want the top of %v", line, ok, adding)
	}
	if line, ok := pressed.MoveTo(Point{X: middle(tiles[0]).X, Y: tiles[0].Y + 2}).PartLanding(); !ok || line.Y != tiles[0].Y {
		t.Errorf("over its own tile it lands at %v (%v), want the top of %v", line, ok, tiles[0])
	}
	if _, ok := w.PressAt(middle(adding)).MoveTo(middle(tiles[0])).PartGhost(); ok {
		t.Error("the line that adds was taken in hand")
	}
	alone := onTheList(t, 800, "Only")
	one := alone.partTiles()[0]
	if _, ok := alone.PressAt(middle(one)).MoveTo(Point{X: middle(one).X, Y: middle(one).Y + Grip + 1}).PartGhost(); !ok {
		t.Error("a part alone in its list is not drawn in hand")
	}
}

// In a list scrolled down, a part lands by the tile drawn under the pointer,
// not by the one that would be there unscrolled.
func TestAPartLandsInAScrolledList(t *testing.T) {
	titles := make([]string, 12)
	for i := range titles {
		titles[i] = fmt.Sprintf("Part %d", i+1)
	}
	w := onTheList(t, 720, titles...)
	w.form.partAt = 11
	w = w.keepingPartInSight()
	tiles := w.partTiles()
	if w.partsFrom() == 0 || len(tiles) < 3 {
		t.Fatalf("this proves nothing: the list shows %d lines from %d", len(tiles), w.partsFrom())
	}
	middle := func(b Box) Point { return Point{X: b.X + b.W/2, Y: b.Y + b.H/2} }
	held := w.PressAt(middle(tiles[0])).MoveTo(Point{X: middle(tiles[0]).X, Y: middle(tiles[0]).Y + Grip + 1}).MoveTo(middle(tiles[1]))
	if line, ok := held.PartLanding(); !ok || line.Y != tiles[1].Y+tiles[1].H-landingHigh {
		t.Errorf("carried down onto the second tile in sight, it lands at %v (%v), want the bottom of %v", line, ok, tiles[1])
	}
}
