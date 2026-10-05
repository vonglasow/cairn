package gui

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"testing"

	"cairn/internal/engine"
	"cairn/internal/reading"
)

// Pressing, dragging and dropping: a card onto the board, a tile under a task,
// and what a drop the store refuses says.

func TestDroppingACardInAnotherColumnChangesItsState(t *testing.T) {
	f := newFake()
	w := holding(f, card(1, "Write the spec", "backlog"))

	w = w.PressAt(middleOf(w, 0, 0))
	w = w.MoveTo(middleOf(w, 1, 0))
	w.LetGo(middleOf(w, 1, 0))

	if f.what("task", "edit") != 1 {
		t.Fatalf("the store was asked %d edits", f.what("task", "edit"))
	}
	in, _ := f.gave("task", "edit")
	if in.Given["number"] != "1" || in.Given["state"] != "todo" {
		t.Errorf("asked to edit %+v, want task 1 into todo", in.Given)
	}
}

// A drop says two things at once, and the engine is told them in that order:
// a column is an edit, a place is a move, and it refuses to do both in one act.
func TestDroppingACardAboveAnotherCarriesBothTheColumnAndThePlace(t *testing.T) {
	f := newFake()
	w := holding(f,
		card(1, "First", "todo"),
		card(2, "Second", "todo"),
		card(3, "Third", "backlog"))

	w = w.PressAt(middleOf(w, 0, 0))
	w = w.MoveTo(middleOf(w, 1, 1))
	w.LetGo(middleOf(w, 1, 1))

	if f.what("task", "edit") != 1 || f.what("task", "move") != 1 {
		t.Fatalf("asked %d edits and %d moves, want one of each",
			f.what("task", "edit"), f.what("task", "move"))
	}
	if f.asked[0].op.Verb != "edit" {
		t.Error("the move was asked for before the column had changed")
	}
	in, _ := f.gave("task", "move")
	if in.Given["number"] != "3" || in.Given["after"] != "1" {
		t.Errorf("asked to move %+v, want task 3 behind task 1", in.Given)
	}
}

func TestDroppingACardInItsOwnColumnOnlyMovesIt(t *testing.T) {
	f := newFake()
	w := holding(f, card(1, "First", "todo"), card(2, "Second", "todo"))

	w = w.PressAt(middleOf(w, 1, 1))
	w = w.MoveTo(middleOf(w, 1, 0))
	w.LetGo(middleOf(w, 1, 0))

	if f.what("task", "edit") != 0 {
		t.Error("a card dropped in its own column had its column changed")
	}
	if f.what("task", "move") != 1 {
		t.Fatalf("asked %d moves", f.what("task", "move"))
	}
	// Dropped on the first slot, which is the head of the column and not
	// behind anything.
	in, _ := f.gave("task", "move")
	if in.Given["first"] != "true" {
		t.Errorf("asked to move %+v, want it first", in.Given)
	}
}

// Which lane a card lands in is which kind of stopped it is. The terminal asks
// with a popup because a terminal band has no room for three named lanes.
func TestDroppingACardInALaneSaysWhichKindOfStopped(t *testing.T) {
	for lane, kind := range Lanes {
		t.Run(kind, func(t *testing.T) {
			f := newFake()
			w := holding(f, card(1, "Client's answer", "todo"))

			w = w.PressAt(middleOf(w, 1, 0))
			w = w.MoveTo(inLane(w, lane))
			w.LetGo(inLane(w, lane))

			in, asked := f.gave("task", "edit")
			if !asked {
				t.Fatal("nothing was asked of the store")
			}
			if in.Given["state"] != kind {
				t.Errorf("asked for %q, want %q", in.Given["state"], kind)
			}
		})
	}
}

// A refusal is shown and the board keeps saying what is really there, rather
// than drawing the move as though it had happened.
func TestADropTheStoreRefusesSaysSoAndChangesNothing(t *testing.T) {
	f := newFake()
	f.failOp = errRefuse
	w := holding(f, card(1, "Write the spec", "backlog"))

	w = w.PressAt(middleOf(w, 0, 0))
	w = w.MoveTo(middleOf(w, 1, 0))
	w = w.LetGo(middleOf(w, 1, 0))

	if w.Wrong() == nil {
		t.Fatal("a refused drop said nothing")
	}
	if w.Dragging() {
		t.Error("the card is still held after a refusal")
	}
}

// A card being dragged follows the pointer, and the board says where it would
// land — without that, a drop is a guess.
func TestACardBeingDraggedFollowsThePointerAndShowsWhereItLands(t *testing.T) {
	f := newFake()
	w := holding(f,
		reading.Card{Number: 1, Title: "First", State: "todo"},
		reading.Card{Number: 2, Title: "Second", State: "backlog"})

	if _, dragging := w.Ghost(); dragging {
		t.Fatal("a board nobody is dragging on shows a card in mid-air")
	}

	from := middleOf(w, 0, 0)
	over := middleOf(w, 1, 0)
	w = w.PressAt(from)
	w = w.MoveTo(over)

	ghost, dragging := w.Ghost()
	if !dragging {
		t.Fatal("a card being dragged is not shown")
	}
	if ghost.X > over.X || over.X > ghost.X+ghost.W || ghost.Y > over.Y || over.Y > ghost.Y+ghost.H {
		t.Errorf("the card at %+v does not hold the pointer at %+v", ghost, over)
	}

	landing, somewhere := w.Landing()
	if !somewhere {
		t.Fatal("the board does not say where the card would land")
	}
	if landing.X != w.Layout().Column[1].X {
		t.Errorf("it lands at %d, and todo is at %d", landing.X, w.Layout().Column[1].X)
	}
}

// Over nothing at all, there is nowhere to land and the board says so rather
// than pointing somewhere.
func TestOverNothingThereIsNowhereToLand(t *testing.T) {
	f := newFake()
	w := holding(f, reading.Card{Number: 1, Title: "First", State: "todo"})
	w = w.PressAt(middleOf(w, 1, 0))
	w = w.MoveTo(Point{X: 0, Y: 0})
	if _, somewhere := w.Landing(); somewhere {
		t.Error("a card over the margin was told where it would land")
	}
}

// The line saying where a card would land is snapped to the grid of cards,
// not left wherever the pointer happens to be: a line between two rows says
// neither.
func TestWhereACardWouldLandIsSnappedToTheGrid(t *testing.T) {
	f := newFake()
	w := holding(f,
		reading.Card{Number: 1, Title: "First", State: "todo"},
		reading.Card{Number: 2, Title: "Second", State: "backlog"})

	l := w.Layout()
	for _, off := range []int{2, CardHeight / 2, CardHeight - 2} {
		over := Point{X: l.Column[1].X + 20, Y: l.Column[1].Y + CardHeight + off}
		held := w.PressAt(middleOf(w, 0, 0)).MoveTo(over)
		landing, somewhere := held.Landing()
		if !somewhere {
			t.Fatalf("%d into the second slot, the board says nowhere", off)
		}
		if want := l.Column[1].Y + CardHeight; landing.Y != want {
			t.Errorf("%d into the second slot, the line is at %d, want %d", off, landing.Y, want)
		}
		if landing.X != l.Column[1].X || landing.W != l.Column[1].W {
			t.Errorf("the line runs from %d for %d, and the column from %d for %d",
				landing.X, landing.W, l.Column[1].X, l.Column[1].W)
		}
	}
}

// Where a point falls in the box under the fields, and where it does not.
func TestTileUnder_saysWhichPartAPointIsOn(t *testing.T) {
	w := openOnAParent(t, 3)
	box := ServingBox(Over(w.width, w.height), w.Shown())
	for _, tc := range []struct {
		what string
		at   Point
		want int
		on   bool
	}{
		{"the first tile", Point{X: box.X + 5, Y: box.Y + 2}, 0, true},
		{"the second", Point{X: box.X + 5, Y: box.Y + SubHeight + 2}, 1, true},
		{"above the box", Point{X: box.X + 5, Y: box.Y - 1}, 0, false},
		{"below the box", Point{X: box.X + 5, Y: box.Y + box.H + 1}, 0, false},
		{"left of it", Point{X: box.X - 1, Y: box.Y + 2}, 0, false},
		{"in the box, past the last tile", Point{X: box.X + 5, Y: box.Y + 3*SubHeight + 2}, 0, false},
	} {
		at, on := w.tileUnder(tc.at)
		if on != tc.on || (on && at != tc.want) {
			t.Errorf("%s: tileUnder = %d, %v, want %d, %v", tc.what, at, on, tc.want, tc.on)
		}
	}
}

// Where a tile lands when it is let go, which is the slot the pointer is over.
func TestDroppedUnder_takesThePlaceItWasLetGoOn(t *testing.T) {
	box := func(w Window) Box { return ServingBox(Over(w.width, w.height), w.Shown()) }
	// Every slot is inside the box. Past its bottom is not a place a part can
	// land — it has no column to fall into — so that is the other test.
	for _, tc := range []struct {
		what       string
		held, slot int
		want       map[string]string
	}{
		{"the third onto the first", 2, 0, map[string]string{"number": "4", "first": "true"}},
		{"the third onto the second", 2, 1, map[string]string{"number": "4", "after": "2"}},
		// Let go on the slot the last tile occupies, which names that tile: the
		// box is three deep for a task, so a slot past all three is outside it,
		// and outside is not a place a part can land.
		{"the first onto the third", 0, 3, map[string]string{"number": "2", "after": "4"}},
	} {
		w := openOnAParent(t, 3)
		from := Point{X: box(w).X + 5, Y: box(w).Y + tc.held*SubHeight + 2}
		to := Point{X: box(w).X + 5, Y: box(w).Y + tc.slot*SubHeight + 2}
		w = w.PressAt(from).MoveTo(to).LetGo(to)
		// The move among what was asked, and not the last of them: a move that
		// went through is followed by the reading that redraws the board.
		var moved *asked
		for i, one := range w.work.(*fake).asked {
			if one.op.Thing == engine.TaskMove.Thing && one.op.Verb == engine.TaskMove.Verb {
				moved = &w.work.(*fake).asked[i]
			}
		}
		if moved == nil {
			t.Fatalf("%s: no move was asked", tc.what)
		}
		for name, value := range tc.want {
			if moved.in.Given[name] != value {
				t.Errorf("%s: asked %v, want %s=%s", tc.what, moved.in.Given, name, value)
			}
		}
	}
}

// A tile let go on its own place asks nothing: it is already there.
func TestDroppedUnder_asksNothingForAPartAlreadyThere(t *testing.T) {
	w := openOnAParent(t, 3)
	box := ServingBox(Over(w.width, w.height), w.Shown())
	at := Point{X: box.X + 5, Y: box.Y + SubHeight + 2}
	past := Point{X: at.X + Grip + 2, Y: at.Y}
	w = w.PressAt(at).MoveTo(past).LetGo(at)
	for _, one := range w.work.(*fake).asked {
		if one.op.Verb == engine.TaskMove.Verb && one.op.Thing == engine.TaskMove.Thing {
			t.Errorf("a part let go where it already was asked %v", one.in.Given)
		}
	}
}

// openOnAParent is a window with a task open on the parts hanging under it,
// numbered 2 upward in the order they are given.
func openOnAParent(t *testing.T, parts int) Window {
	t.Helper()
	f := newFake()
	cards := []reading.Card{{Number: 1, Title: "Ship it", State: "todo"}}
	for i := range parts {
		cards = append(cards, reading.Card{
			Number: i + 2, Title: fmt.Sprintf("Part %d", i+1), State: "todo", Parent: 1,
		})
	}
	w := New(f, 1200, 800)
	w.snap = reading.Snapshot{Cards: cards}
	return w.clicked(middleOf(w, 1, 0)).Key("enter")
}

// What a tile in hand looks like, and where it would land. Pinned because the
// arithmetic that places them is the arithmetic a drop is aimed by: a ghost
// half a tile out or a line on the wrong slot makes the gesture lie.
func TestGhostAndLanding_areTheTilesOwnWhileOneIsHeld(t *testing.T) {
	w := openOnAParent(t, 3)
	box := ServingBox(Over(w.width, w.height), w.Shown())
	from := Point{X: box.X + 5, Y: box.Y + 2*SubHeight + 2}
	to := Point{X: box.X + 5, Y: box.Y + SubHeight + 4}
	w = w.PressAt(from).MoveTo(to)

	ghost, dragging := w.Ghost()
	if !dragging {
		t.Fatal("nothing is in hand")
	}
	// A whole tile: its body is derived the way a tile's is, so that what is
	// carried is laid out exactly as it will be once it is put down.
	want := Box{X: box.X, Y: to.Y - SubHeight/2, W: box.W, H: SubHeight}
	if ghost != want {
		t.Errorf("Ghost = %+v, want %+v", ghost, want)
	}
	if !ghost.In(to) {
		t.Errorf("Ghost = %+v does not hold the pointer at %+v", ghost, to)
	}

	landing, somewhere := w.Landing()
	if !somewhere {
		t.Fatal("nothing says where it would land")
	}
	// Snapped to the slot the hand is over, which is the second.
	wantLanding := Box{X: box.X, Y: box.Y + SubHeight, W: box.W, H: landingHigh}
	if landing != wantLanding {
		t.Errorf("Landing = %+v, want %+v", landing, wantLanding)
	}
}

// Nothing is in hand, and nothing says where it would land, once a tile is let
// go outside the box.
func TestLanding_saysNothingOutsideTheBox(t *testing.T) {
	w := openOnAParent(t, 3)
	box := ServingBox(Over(w.width, w.height), w.Shown())
	from := Point{X: box.X + 5, Y: box.Y + 2}
	out := Point{X: box.X + 5, Y: box.Y + box.H + CardHeight}
	w = w.PressAt(from).MoveTo(out)
	if landing, somewhere := w.Landing(); somewhere {
		t.Errorf("it says a tile would land at %+v, outside its box", landing)
	}
}

// The box follows the cursor out of sight, and comes back with it.
func TestWalkingUnder_bringsWhatItLandsOnIntoSight(t *testing.T) {
	w := openOnAParent(t, 20)
	rows := w.ServingRows()
	if w.ServingAt() != 0 {
		t.Fatalf("it starts at %d, not at the top", w.ServingAt())
	}
	// Walked to the last tile the box can show, nothing has scrolled yet.
	for range rows - 1 {
		w = w.walkingUnder(1)
	}
	if w.ServingAt() != 0 {
		t.Errorf("it scrolled to %d before the cursor left the box", w.ServingAt())
	}
	// One more, and it follows by exactly one tile.
	w = w.walkingUnder(1)
	if got := w.ServingAt(); got != SubHeight {
		t.Errorf("it scrolled to %d, want one tile (%d)", got, SubHeight)
	}
	// Back up to the top, and it comes back.
	for range rows + 2 {
		w = w.walkingUnder(-1)
	}
	if got := w.ServingAt(); got != 0 {
		t.Errorf("it is at %d after walking back to the first, want the top", got)
	}
}

// A move the store refuses leaves the window saying so, and the parts where
// they were: a board that reordered itself on a refusal would be a board
// showing something nobody's store holds.
func TestDroppedUnder_saysSoWhenTheStoreRefuses(t *testing.T) {
	w := openOnAParent(t, 3)
	w.work.(*fake).failOp = errors.New("the store said no")
	box := ServingBox(Over(w.width, w.height), w.Shown())
	from := Point{X: box.X + 5, Y: box.Y + 2*SubHeight + 2}
	to := Point{X: box.X + 5, Y: box.Y + 2}
	w = w.PressAt(from).MoveTo(to).LetGo(to)
	if w.Wrong() == nil {
		t.Error("a refused move said nothing")
	}
	if got := titles(w.Under()); !slices.Equal(got, []string{"Part 1", "Part 2", "Part 3"}) {
		t.Errorf("the parts read %v, want them where they were", got)
	}
}

// Where a point falls, and where a tile lands, once the box has been scrolled.
//
// Every other test of these leaves the box at its top, where the scroll adds
// nothing — so the reckoning could have dropped it entirely and none of them
// would have noticed. A part let go on the fourth tile of a scrolled box must
// name the fourth, not whichever one happens to sit there on screen.
func TestUnderAScrolledBox_reckonsWithWhatIsOutOfSight(t *testing.T) {
	w := openOnAParent(t, 20)
	box := ServingBox(Over(w.width, w.height), w.Shown())
	// Walked far enough that the box has followed the cursor down.
	for range w.ServingRows() + 2 {
		w = w.walkingUnder(1)
	}
	if w.ServingAt() == 0 {
		t.Fatal("the box did not scroll")
	}
	scrolled := w.ServingAt() / SubHeight

	// The tile under the top of the box is the one the scroll put there.
	at, on := w.tileUnder(Point{X: box.X + 5, Y: box.Y + 2})
	if !on || at != scrolled {
		t.Errorf("the top of a scrolled box is tile %d (%v), want %d", at, on, scrolled)
	}

	// And a part let go there takes that tile's place, not the first one's.
	// Read before the drop: the window reads the store again once a move goes
	// through, and this one is answered by a fake that holds no tasks.
	want := strconv.Itoa(w.Under()[scrolled].Number)
	from := Point{X: box.X + 5, Y: box.Y + 3*SubHeight + 2}
	to := Point{X: box.X + 5, Y: box.Y + SubHeight + 2}
	w = w.PressAt(from).MoveTo(to).LetGo(to)
	var moved *asked
	for i, one := range w.work.(*fake).asked {
		if one.op.Thing == engine.TaskMove.Thing && one.op.Verb == engine.TaskMove.Verb {
			moved = &w.work.(*fake).asked[i]
		}
	}
	if moved == nil {
		t.Fatal("nothing was moved")
	}
	// Behind the tile above where it was let go, counted from the top of the
	// listing and not from the top of what is showing.
	if got := moved.in.Given["after"]; got != want {
		t.Errorf("it was put after %s, want %s — the scroll was not counted", got, want)
	}
}

// A part let go just below the last one is put after it, and the window says
// which one it is going after.
//
// This is the slot that tells the reckoning apart: one below the last is still
// inside the box, and naming the last part is not the same as asking for the
// end of the list — both put it in the same place, and only one of them says
// what it is measuring from.
func TestDroppedUnder_namesTheLastPartWhenItGoesBelowIt(t *testing.T) {
	w := openOnAParent(t, 3)
	box := ServingBox(Over(w.width, w.height), w.Shown())
	last := strconv.Itoa(w.Under()[2].Number)
	from := Point{X: box.X + 5, Y: box.Y + 2}
	to := Point{X: box.X + 5, Y: box.Y + 3*SubHeight + 2}
	w = w.PressAt(from).MoveTo(to).LetGo(to)
	var moved *asked
	for i, one := range w.work.(*fake).asked {
		if one.op.Thing == engine.TaskMove.Thing && one.op.Verb == engine.TaskMove.Verb {
			moved = &w.work.(*fake).asked[i]
		}
	}
	if moved == nil {
		t.Fatal("nothing was moved")
	}
	if got := moved.in.Given["after"]; got != last {
		t.Errorf("it was put after %q, want after the last part (%s)", got, last)
	}
	if _, asLast := moved.in.Given["last"]; asLast {
		t.Error("it asked for the end of the list rather than naming what it goes after")
	}
}

// Let go past every tile there is, the move is to the end. Told with two tiles
// rather than three: the box holds three slots, so with three tiles there is
// no slot left past them, and "after the last" is the same gesture.
func TestDroppedUnder_pastThemAllIsTheEnd(t *testing.T) {
	w := openOnAParent(t, 2)
	box := ServingBox(Over(w.width, w.height), w.Shown())
	from := Point{X: box.X + 5, Y: box.Y + 2}
	to := Point{X: box.X + 5, Y: box.Y + 3*SubHeight + 2}
	w = w.PressAt(from).MoveTo(to).LetGo(to)

	var moved *asked
	for i, one := range w.work.(*fake).asked {
		if one.op.Verb == engine.TaskMove.Verb {
			moved = &w.work.(*fake).asked[i]
		}
	}
	if moved == nil {
		t.Fatal("no move was asked")
	}
	if moved.in.Given["last"] != "true" {
		t.Errorf("asked %v, want last=true", moved.in.Given)
	}
}

// An objective is not a task, and dragging one must not move the other. They
// are numbered apart — task 1 and objective 1 both exist — so a press that
// took hold by number alone, on the page where that number names an objective,
// handed a task number to the move that followed.
func TestDraggingAnObjectiveMovesNoTask(t *testing.T) {
	f := newFake()
	w := New(f, 1200, 800)
	w.snap = reading.Snapshot{
		Cards: []reading.Card{card(1, "Write the spec", "todo")},
		Goals: []reading.Goal{{Number: 1, Title: "Ship cairn", State: "active"}},
	}
	w = w.Key("tab")
	if w.Tab() != Objectives {
		t.Fatalf("the window is on %v", w.Tab())
	}
	g := GoalAt(w.Layout(), 0, 0)
	at := Point{X: g.X + g.W/2, Y: g.Y + g.H/2}
	w = w.PressAt(at).MoveTo(Point{X: at.X + 400, Y: at.Y}).LetGo(Point{X: at.X + 400, Y: at.Y})

	for _, one := range f.asked {
		if one.op.Thing == "task" && (one.op.Verb == "edit" || one.op.Verb == "move") {
			t.Errorf("dragging an objective asked %s %s: %v", one.op.Thing, one.op.Verb, one.in.Given)
		}
	}
}
