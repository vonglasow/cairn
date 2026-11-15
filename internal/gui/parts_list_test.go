package gui

import (
	"cairn/internal/board"
	"cairn/internal/engine"
	"cairn/internal/gui/design"
	"cairn/internal/gui/geom"
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
	if !w.Form.OnParts() {
		t.Fatal("the keyboard is not on the steps")
	}
	return w
}

func TestTheArrowsStopAtBothEndsOfTheList(t *testing.T) {
	w := onTheList(t, 800, "A", "B")
	for range 5 {
		w = w.Key("down")
	}
	if w.Form.PartAt != 2 {
		t.Errorf("five down from the first of two parts is line %d, want the line that adds", w.Form.PartAt)
	}
	if again := w.Form.Onto(w.Form.On()); again.PartAt != 2 {
		t.Errorf("staying on the list put the keyboard on line %d", again.PartAt)
	}
	w.Form.PartAt = 1
	if up := w.Key("up"); !up.Form.OnParts() || up.Form.PartAt != 0 {
		t.Errorf("up from the second part left the list, or went to line %d", up.Form.PartAt)
	}
	parts, _ := w.Form.PartsField()
	w.Form.PartAt = 0
	if up := w.Key("up"); up.Form.On() != parts-1 {
		t.Errorf("up from the first part left the keyboard on field %d, want %d", up.Form.On(), parts-1)
	}
	if _, on := w.Key("down").Key("down").PartOn(); on {
		t.Error("on the line that adds, the keyboard is said to be on a part")
	}
}

func TestADeleteKeyAsksOnlyOnAPart(t *testing.T) {
	w := onTheList(t, 800, "A")
	adding := w.Key("down")
	for _, key := range []string{"space", "backspace", "delete-forward"} {
		if got := adding.Key(key); got.Screen != board.OnForm || fmt.Sprint(got.Form.Fields) != fmt.Sprint(adding.Form.Fields) {
			t.Errorf("%s on the line that adds left the window on %v with %v", key, got.Screen, got.Form.Fields)
		}
	}
	for _, key := range []string{"backspace", "delete-forward"} {
		asked := w.Key(key)
		if asked.Screen != board.OnAsking || Heads(asked) != "Delete the step “A”?" {
			t.Errorf("%s on a part left the window on %v asking %q", key, asked.Screen, Heads(asked))
		}
		if back := asked.Key("esc"); back.Screen != board.OnForm || len(back.Form.Parts()) != 1 {
			t.Errorf("cancelled, the window is on %v with %+v", back.Screen, back.Form.Parts())
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
	w.Form = w.Form.TakingPart(1, false, false)
	w = w.Key("backspace").Key("enter")
	in, asked := f.gave("step", "delete")
	if !asked || in.Given["number"] != "1" || in.Given["step"] != "2" {
		t.Fatalf("the engine was asked %v %+v", asked, in)
	}
	if w.Screen != board.OnForm {
		t.Errorf("agreed, the window is on %v", w.Screen)
	}
	want := []engine.Part{{Number: 1, Title: "A"}, {Number: 2, Title: "C"}, {Title: "D"}}
	if got := w.Form.Parts(); fmt.Sprint(got) != fmt.Sprint(want) {
		t.Errorf("the form holds %+v, want %+v", got, want)
	}
	at, _ := w.Form.PartsField()
	if got := engine.PartsIn(w.Form.Opened[at]); fmt.Sprint(got) != fmt.Sprint(want[:2]) {
		t.Errorf("the form was opened on %+v, want %+v", got, want[:2])
	}
	if w.Form.PartAt != 1 {
		t.Errorf("the keyboard is on line %d, want the step that moved up", w.Form.PartAt)
	}
}

func TestAPartTheEngineWillNotDeleteStays(t *testing.T) {
	f := newFake()
	w := onTheListOf(t, f, "A", "B")
	f.failOp = errRefuse
	w = w.Key("backspace").Key("enter")
	if got := w.Form.Parts(); len(got) != 2 {
		t.Errorf("a step the engine refused to delete left %+v", got)
	}
	if w.Screen != board.OnForm {
		t.Errorf("refused, the window is on %v", w.Screen)
	}
}

func TestAPartIsCarriedNoFurtherThanTheList(t *testing.T) {
	f := onTheList(t, 800, "A", "B", "C").Form
	if got := f.Carrying(0, 99).Parts(); got[2].Title != "A" {
		t.Errorf("carried past the last, the parts read %+v", got)
	}
	if got := f.Carrying(3, 0).Parts(); got[0].Title != "A" {
		t.Errorf("carrying the line that adds moved %+v", got)
	}
}

func TestAClickOnTheListDoesWhatItLandsOn(t *testing.T) {
	titles := make([]string, 12)
	for i := range titles {
		titles[i] = fmt.Sprintf("Part %d", i+1)
	}
	w := onTheList(t, 720, titles...)
	w.Form.PartAt = 11
	w = w.keepingPartInSight()
	from := w.partsFrom()
	tiles := w.partTiles()
	if from == 0 || len(tiles) < 3 {
		t.Fatalf("this proves nothing: the list shows %d lines from %d", len(tiles), from)
	}
	middle := func(b geom.Box) geom.Point { return geom.Point{X: b.X + b.W/2, Y: b.Y + b.H/2} }
	box := SubCheck(tiles[0])
	ticked := w.clicked(geom.Point{X: box.X + 2, Y: box.Y + 2})
	if parts := ticked.Form.Parts(); !parts[from].Done || ticked.Form.PartAt != from {
		t.Errorf("a click on the box of line %d ticked %+v, the keyboard on %d", from, parts[from], ticked.Form.PartAt)
	}
	chosen := w.clicked(middle(tiles[1]))
	if parts := chosen.Form.Parts(); parts[from+1].Done || chosen.Form.PartAt != from+1 {
		t.Errorf("a click beside the box ticked %+v, the keyboard on %d", parts[from+1], chosen.Form.PartAt)
	}
	adding := w.clicked(middle(tiles[len(tiles)-1]))
	if !adding.Form.WritingPart || adding.Form.PartAt != 12 {
		t.Errorf("a click on the line that adds wrote %v on line %d", adding.Form.WritingPart, adding.Form.PartAt)
	}
	top := w
	top.Form.PartAt, top.Form.PartsTop = 0, 0
	if adds := top.clicked(middle(top.partTiles()[len(top.partTiles())-1])); adds.Form.PartAt != 12 {
		t.Errorf("a click on the line that adds, the list at its top, put the keyboard on line %d", adds.Form.PartAt)
	}
	renaming := w.DoubleClick(middle(tiles[1]))
	if !renaming.Form.WritingPart || renaming.Form.Named.Reads() != titles[from+1] {
		t.Errorf("a double click on a part writes %v on %q", renaming.Form.WritingPart, renaming.Form.Named.Reads())
	}
}

func TestAPartInHandSaysWhereItLands(t *testing.T) {
	w := onTheList(t, 800, "A", "B", "C")
	tiles := w.partTiles()
	middle := func(b geom.Box) geom.Point { return geom.Point{X: b.X + b.W/2, Y: b.Y + b.H/2} }
	pressed := w.PressAt(middle(tiles[1]))
	if _, ok := pressed.PartGhost(); ok {
		t.Error("a part is in hand before it has moved")
	}
	if _, ok := pressed.PartLanding(); ok {
		t.Error("a part says where it lands before it has moved")
	}
	up := pressed.MoveTo(middle(tiles[0]))
	if line, ok := up.PartLanding(); !ok || line.Y != tiles[0].Y-1-landingHigh/2 || line.X != tiles[0].X || line.W != tiles[0].W {
		t.Errorf("carried up onto the first, it lands at %v (%v), want the gap above %v", line, ok, tiles[0])
	}
	down := pressed.MoveTo(middle(tiles[2]))
	if line, ok := down.PartLanding(); !ok || line.Y != tiles[2].Y+tiles[2].H+1-landingHigh/2 {
		t.Errorf("carried down onto the last, it lands at %v (%v), want the gap under %v", line, ok, tiles[2])
	}
	if ghost, ok := down.PartGhost(); !ok || ghost.H != tiles[1].H || ghost.Y != tiles[1].Y+middle(tiles[2]).Y-middle(tiles[1]).Y || ghost.X != tiles[1].X {
		t.Errorf("the part in hand is drawn at %v", ghost)
	}
	off := pressed.MoveTo(geom.Point{X: 1, Y: 1})
	if _, ok := off.PartLanding(); ok {
		t.Error("held off the list, it says it would land somewhere")
	}
	shot := drawn(t, down)
	if ghost, _ := down.PartGhost(); lit(shot, SubWriting(ghost), 90) == 0 {
		t.Error("the part in hand is not drawn")
	}
	over := pressed.MoveTo(geom.Point{X: middle(tiles[2]).X, Y: tiles[2].Y + tiles[2].H - 2})
	if line, _ := over.PartLanding(); painted(drawn(t, over), line, design.Dark().Colour.Accent.Base) < line.W*line.H/2 {
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
	if ticked.Form.Changed() {
		t.Error("a tick left the form something to save")
	}
	w2 := onTheList(t, 800, "A")
	w2.work.(*fake).failOp = errNope
	refused := w2.Key("space")
	if refused.Form.Parts()[0].Done || refused.Wrong() == nil {
		t.Errorf("a refused tick left the part %+v, saying %v", refused.Form.Parts()[0], refused.Wrong())
	}
}

func TestWhereAPartLandsAtTheEdges(t *testing.T) {
	w := onTheList(t, 800, "A", "B")
	tiles := w.partTiles()
	middle := func(b geom.Box) geom.Point { return geom.Point{X: b.X + b.W/2, Y: b.Y + b.H/2} }
	pressed := w.PressAt(middle(tiles[0])).MoveTo(geom.Point{X: middle(tiles[0]).X, Y: middle(tiles[0]).Y + Wobble + 1})
	adding := tiles[len(tiles)-1]
	if line, ok := pressed.MoveTo(middle(adding)).PartLanding(); !ok || line.Y != adding.Y || line.X != tiles[0].X {
		t.Errorf("over the line that adds it lands at %v (%v), want the top of %v", line, ok, adding)
	}
	if line, ok := pressed.MoveTo(geom.Point{X: middle(tiles[0]).X, Y: tiles[0].Y + 2}).PartLanding(); ok {
		t.Errorf("over its own tile it lands at %v: a drop there moves nothing", line)
	}
	if _, ok := w.PressAt(middle(adding)).MoveTo(middle(tiles[0])).PartGhost(); ok {
		t.Error("the line that adds was taken in hand")
	}
	alone := onTheList(t, 800, "Only")
	one := alone.partTiles()[0]
	if _, ok := alone.PressAt(middle(one)).MoveTo(geom.Point{X: middle(one).X, Y: middle(one).Y + Wobble + 1}).PartGhost(); !ok {
		t.Error("a part alone in its list is not drawn in hand")
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
	over := geom.Point{X: tiles[0].X + 4, Y: tiles[0].Y + 4}
	half := w.Scroll(over, geom.SubHeight/2)
	if half.partsFrom() != 0 || half.Scroll(over, geom.SubHeight-geom.SubHeight/2).partsFrom() != 1 {
		t.Errorf("half a notch showed from %d, and the rest of it from %d", half.partsFrom(), half.Scroll(over, geom.SubHeight-geom.SubHeight/2).partsFrom())
	}
	for notches := 1; notches <= 12; notches++ {
		turned := w
		for range notches {
			turned = turned.Scroll(over, geom.SubHeight)
		}
		if want := min(notches, 12-shown); turned.partsFrom() != want {
			t.Errorf("%d notches show from %d, want %d", notches, turned.partsFrom(), want)
		}
	}
}

func TestStepsAreGatheredAndStretched(t *testing.T) {
	f := onTheList(t, 800, "A", "B", "C", "D").Form
	if got := f.TakingPart(2, true, false).TakenParts(); fmt.Sprint(got) != "[0 2]" {
		t.Errorf("⌘ on C from A took %v", got)
	}
	if got := f.TakingPart(2, true, false).TakingPart(0, true, false).TakenParts(); got != nil {
		t.Errorf("⌘ on A again left %v taken", got)
	}
	f.PartAt = 1
	stretched := f.TakingPart(3, false, true)
	if got := stretched.TakenParts(); fmt.Sprint(got) != "[1 2 3]" || stretched.PartAt != 3 {
		t.Errorf("⇧ on D from B took %v, on line %d", got, stretched.PartAt)
	}
	if got := stretched.TakingPart(0, false, false).TakenParts(); got != nil {
		t.Errorf("a plain click left %v taken", got)
	}
	if got := f.TakingPart(4, true, false).TakenParts(); got != nil {
		t.Errorf("the line that adds was taken: %v", got)
	}
}

func TestStepsTakenTogetherMoveAsOneBlock(t *testing.T) {
	titles := func(f board.Form) string {
		var out []string
		for _, p := range f.Parts() {
			out = append(out, p.Title)
		}
		return strings.Join(out, "")
	}
	f := onTheList(t, 800, "A", "B", "C", "D").Form
	down := f.TakingPart(1, true, false).CarryingSeveral(3)
	if titles(down) != "CDAB" || fmt.Sprint(down.TakenParts()) != "[2 3]" {
		t.Errorf("A and B carried onto D give %s, taken %v", titles(down), down.TakenParts())
	}
	up := f.TakingPart(2, false, false).TakingPart(3, true, false).CarryingSeveral(0)
	if titles(up) != "CDAB" || fmt.Sprint(up.TakenParts()) != "[0 1]" {
		t.Errorf("C and D carried onto A give %s, taken %v", titles(up), up.TakenParts())
	}
	if titles(f.TakingPart(2, false, false).TakingPart(3, true, false).CarryingSeveral(3)) != "ABCD" {
		t.Error("a block carried onto itself moved")
	}
}

func TestSeveralStepsAreDeletedInOneAct(t *testing.T) {
	f := newFake()
	w := onTheListOf(t, f, "A", "B", "C")
	w = w.Key("down").Key("down").Key("down").Key("enter").Types("D").Key("enter").Key("esc")
	w.Form = w.Form.TakingPart(0, false, false).TakingPart(2, true, false).TakingPart(3, true, false)
	asked := w.Key("backspace")
	if asked.Screen != board.OnAsking || Heads(asked) != "Delete 3 steps?" {
		t.Fatalf("the delete key asked %q on %v", Heads(asked), asked.Screen)
	}
	gone := asked.Key("enter")
	in, ok := f.gave("task", "edit")
	if !ok || in.Given["number"] != "1" || titlesOf(engine.PartsIn(in.Given["steps"])) != "B" {
		t.Fatalf("the engine was asked %v %+v", ok, in)
	}
	if got := gone.Form.Parts(); len(got) != 1 || got[0] != (engine.Part{Number: 1, Title: "B"}) {
		t.Errorf("the form holds %+v", got)
	}
	if gone.Form.TakenParts() != nil || gone.Screen != board.OnForm {
		t.Errorf("after the delete, %v taken, on %v", gone.Form.TakenParts(), gone.Screen)
	}
	f.failOp = errRefuse
	refused := w.Key("backspace").Key("enter")
	if len(refused.Form.Parts()) != 4 {
		t.Errorf("a delete the engine refused left %+v", refused.Form.Parts())
	}
}

func titlesOf(parts []engine.Part) string {
	var out []string
	for _, p := range parts {
		out = append(out, p.Title)
	}
	return strings.Join(out, "")
}

func TestEveryStepIsTakenAtOnce(t *testing.T) {
	w := onTheList(t, 800, "A", "B", "C").Key("cmd+a")
	if got := w.Form.TakenParts(); fmt.Sprint(got) != "[0 1 2]" {
		t.Errorf("⌘A took %v", got)
	}
	if one := onTheList(t, 800, "A").Key("cmd+a"); one.Form.TakenParts() != nil {
		t.Errorf("⌘A on one step took %v", one.Form.TakenParts())
	}
}
