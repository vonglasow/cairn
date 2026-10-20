package gui

import (
	"cairn/internal/engine"
	"fmt"
	"strings"
	"testing"
)

func onTheList(t *testing.T, high int, titles ...string) Window {
	t.Helper()
	return onTheListHigh(t, newFake(), high, titles...)
}

func onTheListOf(t *testing.T, f *fake, titles ...string) Window {
	t.Helper()
	return onTheListHigh(t, f, 800, titles...)
}

func onTheListHigh(t *testing.T, f *fake, high int, titles ...string) Window {
	t.Helper()
	shipped := card(1, "Ship it", "todo")
	for i, title := range titles {
		shipped.Steps = append(shipped.Steps, engine.Part{Number: i + 1, Title: title})
	}
	w := holding(f, shipped)
	w.height = high
	w = w.clicked(middleOf(w, 1, 0)).DoubleClick(middleOf(w, 1, 0)).On("steps")
	if !w.form.onParts() {
		t.Fatal("the keyboard is not on the steps")
	}
	return w
}

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

func TestADeleteKeyAsksOnlyOnAPart(t *testing.T) {
	w := onTheList(t, 800, "A")
	adding := w.Key("down")
	for _, key := range []string{"space", "backspace", "delete-forward"} {
		if got := adding.Key(key); got.Screen() != OnForm || fmt.Sprint(got.form.fields) != fmt.Sprint(adding.form.fields) {
			t.Errorf("%s on the line that adds left the window on %v with %v", key, got.Screen(), got.form.fields)
		}
	}
	for _, key := range []string{"backspace", "delete-forward"} {
		asked := w.Key(key)
		if asked.Screen() != OnAsking || Heads(asked) != "Delete the step “A”?" {
			t.Errorf("%s on a part left the window on %v asking %q", key, asked.Screen(), Heads(asked))
		}
		if back := asked.Key("esc"); back.Screen() != OnForm || len(back.form.parts()) != 1 {
			t.Errorf("cancelled, the window is on %v with %+v", back.Screen(), back.form.parts())
		}
	}
}

func TestAgreedAPartIsDeletedAtOnceAndTheOnesBehindMoveUp(t *testing.T) {
	f := newFake()
	shipped := card(1, "Ship it", "todo")
	shipped.Steps = []engine.Part{{Number: 1, Title: "A"}, {Number: 2, Title: "B"}, {Number: 3, Title: "C"}}
	w := holding(f, shipped)
	w.height = 800
	w = w.clicked(middleOf(w, 1, 0)).DoubleClick(middleOf(w, 1, 0)).On("steps")
	w = w.Key("down").Key("down").Key("down").Key("enter").Types("D").Key("enter").Key("esc")
	w.form.partAt = 1
	w = w.Key("backspace").Key("enter")
	in, asked := f.gave("step", "delete")
	if !asked || in.Given["number"] != "1" || in.Given["step"] != "2" {
		t.Fatalf("the engine was asked %v %+v", asked, in)
	}
	if w.Screen() != OnForm {
		t.Errorf("agreed, the window is on %v", w.Screen())
	}
	want := []engine.Part{{Number: 1, Title: "A"}, {Number: 2, Title: "C"}, {Title: "D"}}
	if got := w.form.parts(); fmt.Sprint(got) != fmt.Sprint(want) {
		t.Errorf("the form holds %+v, want %+v", got, want)
	}
	at, _ := w.form.partsField()
	if got := engine.PartsIn(w.form.opened[at]); fmt.Sprint(got) != fmt.Sprint(want[:2]) {
		t.Errorf("the form was opened on %+v, want %+v", got, want[:2])
	}
	if w.form.partAt != 1 {
		t.Errorf("the keyboard is on line %d, want the step that moved up", w.form.partAt)
	}
}

func TestAPartTheEngineWillNotDeleteStays(t *testing.T) {
	f := newFake()
	w := onTheListOf(t, f, "A", "B")
	f.failOp = errRefuse
	w = w.Key("backspace").Key("enter")
	if got := w.form.parts(); len(got) != 2 {
		t.Errorf("a step the engine refused to delete left %+v", got)
	}
	if w.Screen() != OnForm {
		t.Errorf("refused, the window is on %v", w.Screen())
	}
}

func TestAPartNeverSavedIsTakenOutWithoutTheEngine(t *testing.T) {
	f := newFake()
	w := onTheListOf(t, f, "A")
	added := w.Key("down").Key("enter").Types("New").Key("enter").Key("esc")
	added.form.partAt = 1
	gone := added.Key("backspace").Key("enter")
	if parts := gone.form.parts(); len(parts) != 1 || parts[0].Title != "A" {
		t.Errorf("deleting a part never saved left %+v", parts)
	}
	if _, asked := f.gave("step", "delete"); asked {
		t.Error("the engine was asked to delete a step it never had")
	}
	if gone.form.partAt != 1 {
		t.Errorf("the keyboard is on line %d, want the line that adds", gone.form.partAt)
	}
}

func TestTheNextPartIsWaiting(t *testing.T) {
	w := onTheList(t, 800, "A").Key("down").Key("enter").Types("B").Key("enter")
	if !w.form.writingPart || w.form.partAt != w.form.lines()-1 || w.form.named.Reads() != "" {
		t.Errorf("after adding, writing %v on line %d of %d, holding %q", w.form.writingPart, w.form.partAt, w.form.lines(), w.form.named.Reads())
	}
	if put := w.Key("esc"); put.Screen() != OnForm || put.form.writingPart {
		t.Errorf("esc while writing left the window on %v, writing %v", put.Screen(), put.form.writingPart)
	}
}

func TestAPartIsCarriedNoFurtherThanTheList(t *testing.T) {
	f := onTheList(t, 800, "A", "B", "C").form
	if got := f.carrying(0, 99).parts(); got[2].Title != "A" {
		t.Errorf("carried past the last, the parts read %+v", got)
	}
	if got := f.carrying(3, 0).parts(); got[0].Title != "A" {
		t.Errorf("carrying the line that adds moved %+v", got)
	}
}

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
		if line > 0 {
			w = w.Key("down")
		}
		if w.form.partAt != line {
			t.Fatalf("down went to line %d, want %d", w.form.partAt, line)
		}
		tiles, from := w.partTiles(), w.partsFrom()
		shown := tiles[len(tiles)-1]
		if line < 12 {
			shown = tiles[line-from]
		}
		if shown.Y < frame.Y || shown.Y+shown.H > frame.Y+frame.H {
			t.Errorf("line %d is drawn at %v, out of %v", line, shown, frame)
		}
		if from > 0 {
			earlier, _ := TickedPlaces(Framed(value, rows), 12, from-1)
			if line < 12 && line-from+1 < len(earlier) {
				t.Errorf("line %d would show from %d as well, and the list starts at %d", line, from-1, from)
			}
		}
	}
}

func TestTicksStartFromTheFirstInSight(t *testing.T) {
	frame := Box{X: 100, Y: 200, W: 400, H: 8 + 12*lineHigh}
	tiles, _ := TickedPlaces(frame, 5, 2)
	if len(tiles) != 3 || tiles[0].Y != Inside(frame).Y {
		t.Errorf("five parts from the third gave %v", tiles)
	}
}

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
	elsewhere := w.form.onto(0)
	w.form = elsewhere
	if painted(drawn(t, w), Edges(SubPanel(tiles[0]))[0], lit) != 0 {
		t.Error("a part is outlined with the keyboard elsewhere")
	}
}

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
	over := pressed.MoveTo(Point{X: middle(tiles[2]).X, Y: tiles[2].Y + tiles[2].H - 2})
	if line, _ := over.PartLanding(); painted(drawn(t, over), line, Dark().Chosen) < line.W*line.H/2 {
		t.Error("where it would land is hidden under the part in hand")
	}
}

func TestATickIsDoneAtOnceOrNotAtAll(t *testing.T) {
	w := onTheList(t, 800, "A")
	f := w.work.(*fake)
	ticked := w.Key("space")
	if in, ok := f.gave("step", "edit"); !ok || in.Given["number"] != "1" || in.Given["step"] != "1" || in.Given["done"] != "true" {
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

func TestTheWheelScrollsTheParts(t *testing.T) {
	titles := make([]string, 12)
	for i := range titles {
		titles[i] = fmt.Sprintf("Part %d", i+1)
	}
	w := onTheList(t, Shortest, titles...)
	w.form = w.form.onto(0)
	tiles := w.partTiles()
	shown := len(tiles) - 1
	over := Point{X: tiles[0].X + 4, Y: tiles[0].Y + 4}
	once := w.Scroll(over, SubHeight)
	if once.partsFrom() != 1 {
		t.Errorf("a notch over the parts shows them from %d", once.partsFrom())
	}
	far := once
	for range 99 {
		far = far.Scroll(over, SubHeight)
	}
	if got := far.partsFrom(); got != 12-shown {
		t.Errorf("turned to the end, the parts show from %d, want %d", got, 12-shown)
	}
	for range 99 {
		far = far.Scroll(over, -SubHeight)
	}
	if got := far.partsFrom(); got != 0 {
		t.Errorf("turned back, the parts show from %d", got)
	}
	walked := once.On("steps")
	if walked.partsFrom() != once.partsFrom() {
		t.Errorf("putting the keyboard on the list moved it from %d to %d", once.partsFrom(), walked.partsFrom())
	}
	if walked = walked.Key("down"); walked.form.partAt < walked.partsFrom() || walked.form.partAt >= walked.partsFrom()+shown {
		t.Errorf("down put the keyboard on line %d, out of sight from %d", walked.form.partAt, walked.partsFrom())
	}
}

func TestTheListMovesExactlyAsFarAsAsked(t *testing.T) {
	titles := make([]string, 12)
	for i := range titles {
		titles[i] = fmt.Sprintf("Part %d", i+1)
	}
	w := onTheList(t, 720, titles...)
	shown := w.partsShown()
	if shown < 3 || shown >= 12 {
		t.Fatalf("this proves nothing: %d parts are shown", shown)
	}
	for line := 1; line <= 12; line++ {
		w = w.Key("down")
		if want := max(0, min(line, 11)-shown+1); w.partsFrom() != want {
			t.Errorf("down to line %d shows from %d, want %d", line, w.partsFrom(), want)
		}
	}
	for line := 11; line >= 0; line-- {
		w = w.Key("up")
		if want := min(12-shown, line); w.partsFrom() != want {
			t.Errorf("up to line %d shows from %d, want %d", line, w.partsFrom(), want)
		}
	}
	tiles := w.partTiles()
	over := Point{X: tiles[0].X + 4, Y: tiles[0].Y + 4}
	half := w.Scroll(over, SubHeight/2)
	if half.partsFrom() != 0 || half.Scroll(over, SubHeight-SubHeight/2).partsFrom() != 1 {
		t.Errorf("half a notch showed from %d, and the rest of it from %d", half.partsFrom(), half.Scroll(over, SubHeight-SubHeight/2).partsFrom())
	}
	for notches := 1; notches <= 12; notches++ {
		turned := w
		for range notches {
			turned = turned.Scroll(over, SubHeight)
		}
		if want := min(notches, 12-shown); turned.partsFrom() != want {
			t.Errorf("%d notches show from %d, want %d", notches, turned.partsFrom(), want)
		}
	}
}

func TestStepsAreGatheredAndStretched(t *testing.T) {
	f := onTheList(t, 800, "A", "B", "C", "D").form
	if got := f.takingPart(2, true, false).takenParts(); fmt.Sprint(got) != "[0 2]" {
		t.Errorf("⌘ on C from A took %v", got)
	}
	if got := f.takingPart(2, true, false).takingPart(0, true, false).takenParts(); got != nil {
		t.Errorf("⌘ on A again left %v taken", got)
	}
	f.partAt = 1
	stretched := f.takingPart(3, false, true)
	if got := stretched.takenParts(); fmt.Sprint(got) != "[1 2 3]" || stretched.partAt != 3 {
		t.Errorf("⇧ on D from B took %v, on line %d", got, stretched.partAt)
	}
	if got := stretched.takingPart(0, false, false).takenParts(); got != nil {
		t.Errorf("a plain click left %v taken", got)
	}
	if got := f.takingPart(4, true, false).takenParts(); got != nil {
		t.Errorf("the line that adds was taken: %v", got)
	}
}

func TestStepsTakenTogetherMoveAsOneBlock(t *testing.T) {
	titles := func(f Form) string {
		var out []string
		for _, p := range f.parts() {
			out = append(out, p.Title)
		}
		return strings.Join(out, "")
	}
	f := onTheList(t, 800, "A", "B", "C", "D").form
	down := f.takingPart(1, true, false).carryingSeveral(3)
	if titles(down) != "CDAB" || fmt.Sprint(down.takenParts()) != "[2 3]" {
		t.Errorf("A and B carried onto D give %s, taken %v", titles(down), down.takenParts())
	}
	f.partAt = 2
	up := f.takingPart(3, true, false).carryingSeveral(0)
	if titles(up) != "CDAB" || fmt.Sprint(up.takenParts()) != "[0 1]" {
		t.Errorf("C and D carried onto A give %s, taken %v", titles(up), up.takenParts())
	}
	if titles(f.takingPart(3, true, false).carryingSeveral(3)) != "ABCD" {
		t.Error("a block carried onto itself moved")
	}
}

func TestSeveralStepsAreDeletedInOneAct(t *testing.T) {
	f := newFake()
	w := onTheListOf(t, f, "A", "B", "C")
	w = w.Key("down").Key("down").Key("down").Key("enter").Types("D").Key("enter").Key("esc")
	w.form.partAt = 0
	w.form = w.form.takingPart(2, true, false).takingPart(3, true, false)
	asked := w.Key("backspace")
	if asked.Screen() != OnAsking || Heads(asked) != "Delete 3 steps?" {
		t.Fatalf("the delete key asked %q on %v", Heads(asked), asked.Screen())
	}
	gone := asked.Key("enter")
	in, ok := f.gave("task", "edit")
	if !ok || in.Given["number"] != "1" || titlesOf(engine.PartsIn(in.Given["steps"])) != "B" {
		t.Fatalf("the engine was asked %v %+v", ok, in)
	}
	if got := gone.form.parts(); len(got) != 1 || got[0] != (engine.Part{Number: 1, Title: "B"}) {
		t.Errorf("the form holds %+v", got)
	}
	if gone.form.takenParts() != nil || gone.Screen() != OnForm {
		t.Errorf("after the delete, %v taken, on %v", gone.form.takenParts(), gone.Screen())
	}
	f.failOp = errRefuse
	refused := w.Key("backspace").Key("enter")
	if len(refused.form.parts()) != 4 {
		t.Errorf("a delete the engine refused left %+v", refused.form.parts())
	}
}

func titlesOf(parts []engine.Part) string {
	var out []string
	for _, p := range parts {
		out = append(out, p.Title)
	}
	return strings.Join(out, "")
}
