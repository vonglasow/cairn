package gui

import (
	"strconv"
	"testing"

	"gioui.org/io/key"

	"cairn/internal/engine"
)

// The keyboard on the window: the arrows walking the board, the band and the
// objectives, shift moving what is chosen, and the few keys the bar does not
// offer.

// e edits whatever is chosen, because a hand that has just chosen something is
// not on the other side of the window.
func TestEEditsWhateverIsChosen(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "Write the spec", State: "todo"})

	if got := w.Key("e"); got.Screen() != OnTheBoard {
		t.Error("e with nothing chosen opened something")
	}
	w = w.clicked(middleOf(w, 1, 0))
	if got := w.Key("e"); got.Screen() != OnForm {
		t.Fatalf("e left the window on %v", got.Screen())
	}
}

// Deleting, searching and the help: what the terminal board offers and the
// window did not.

// The arrows walk the board, so a hand on the keyboard does not have to reach
// for the mouse to look at the next task.
func TestTheArrowsWalkTheBoard(t *testing.T) {
	f := newFake()
	w := holding(f,
		engine.Card{Number: 1, Title: "In the heap", State: "backlog"},
		engine.Card{Number: 2, Title: "First", State: "todo"},
		engine.Card{Number: 3, Title: "Second", State: "todo"})

	w = w.clicked(middleOf(w, 0, 0))
	if got := w.Key("right").Chosen(); got != 2 {
		t.Errorf("right of the backlog chose %d", got)
	}
	across := w.Key("right")
	if got := across.Key("down").Chosen(); got != 3 {
		t.Errorf("down from the first of todo chose %d", got)
	}
	if got := across.Key("down").Key("up").Chosen(); got != 2 {
		t.Errorf("up again chose %d", got)
	}
	if got := across.Key("left").Chosen(); got != 1 {
		t.Errorf("left of todo chose %d", got)
	}
}

// The arrows stop at the edges rather than wrapping round, and an empty column
// is walked past rather than stopped in.
func TestTheArrowsStopAtTheEdges(t *testing.T) {
	f := newFake()
	w := holding(f,
		engine.Card{Number: 1, Title: "Only", State: "backlog"},
		engine.Card{Number: 2, Title: "Last", State: "done"})

	w = w.clicked(middleOf(w, 0, 0))
	if got := w.Key("left").Chosen(); got != 1 {
		t.Errorf("left of the first column chose %d", got)
	}
	if got := w.Key("up").Chosen(); got != 1 {
		t.Errorf("up from the first card chose %d", got)
	}
	if got := w.Key("down").Chosen(); got != 1 {
		t.Errorf("down from the only card of a column chose %d", got)
	}
	// Two columns stand empty between them; walking right reaches the last.
	if got := w.Key("right").Key("right").Key("right").Chosen(); got != 2 {
		t.Errorf("walking right across two empty columns chose %d", got)
	}
}

// With nothing chosen, an arrow chooses somewhere to begin rather than nothing.
func TestAnArrowWithNothingChosenChoosesSomewhereToBegin(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "Only", State: "backlog"})
	if got := w.Key("down").Chosen(); got != 1 {
		t.Errorf("an arrow on an untouched board chose %d", got)
	}
}

// The arrows walk the objectives too, which are a list and so have no columns.
func TestTheArrowsWalkTheObjectives(t *testing.T) {
	f := newFake()
	w := New(f, 1200, 800)
	w.snap = engine.Snapshot{Goals: []engine.Goal{
		{Number: 3, Title: "Ship cairn", State: "active"},
		{Number: 4, Title: "Keep it honest", State: "active"},
	}}
	w.tab = Objectives

	first := GoalAt(w.Layout(), 0, 0)
	w = w.clicked(Point{X: first.X + first.W/2, Y: first.Y + first.H/2})
	if got := w.Key("down").Chosen(); got != 4 {
		t.Errorf("down the list chose %d", got)
	}
	if got := w.Key("down").Key("up").Chosen(); got != 3 {
		t.Errorf("up the list chose %d", got)
	}
	if got := w.Key("down").Key("down").Chosen(); got != 4 {
		t.Errorf("past the end of the list chose %d", got)
	}
}

// Enter opens what the arrows walked to, the way a double-click opens what the
// pointer landed on.
func TestEnterOpensWhatIsChosen(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "Write the spec", State: "todo"})

	if got := w.Key("enter"); got.Screen() != OnTheBoard {
		t.Error("enter with nothing chosen opened something")
	}
	w = w.clicked(middleOf(w, 1, 0)).Key("enter")
	if w.Screen() != OnForm {
		t.Fatalf("enter left the window on %v", w.Screen())
	}
	if w.Chosen() != 1 {
		t.Errorf("it opened %d", w.Chosen())
	}
}

// Walking across into a shorter column lands on its last card rather than
// nowhere: the level is kept as far as there is one.
func TestWalkingAcrossIntoAShorterColumnLandsOnItsLast(t *testing.T) {
	f := newFake()
	w := holding(f,
		engine.Card{Number: 1, Title: "One", State: "backlog"},
		engine.Card{Number: 2, Title: "Two", State: "backlog"},
		engine.Card{Number: 3, Title: "Three", State: "backlog"},
		engine.Card{Number: 4, Title: "Only", State: "todo"})

	w = w.clicked(middleOf(w, 0, 2))
	if got := w.Chosen(); got != 3 {
		t.Fatalf("the third card of the backlog is %d", got)
	}
	if got := w.Key("right").Chosen(); got != 4 {
		t.Errorf("right into a column of one chose %d", got)
	}
}

// Walking the objectives from somewhere other than the first: the list has to
// find where it is before it can move.
func TestWalkingTheObjectivesFromTheMiddle(t *testing.T) {
	f := newFake()
	w := New(f, 1200, 800)
	w.snap = engine.Snapshot{Goals: []engine.Goal{
		{Number: 3, Title: "First", State: "active"},
		{Number: 4, Title: "Second", State: "active"},
		{Number: 5, Title: "Third", State: "active"},
	}}
	w.tab, w.chosen, w.chosenGoal = Objectives, 4, true

	if got := w.Key("down").Chosen(); got != 5 {
		t.Errorf("down from the second chose %d", got)
	}
	if got := w.Key("up").Chosen(); got != 3 {
		t.Errorf("up from the second chose %d", got)
	}
}

// An arrow on a board whose first columns are empty begins in the first column
// that has anything.
func TestAnArrowBeginsInTheFirstColumnThatHasAnything(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 7, Title: "Only", State: "current"})
	if got := w.Key("down").Chosen(); got != 7 {
		t.Errorf("an arrow chose %d, and the only card is 7", got)
	}
}

// The band is walked like the columns. A task that is blocked, watching or
// paused could not be reached by the arrows at all — and those are the tasks
// that want a decision.
func TestTheArrowsReachTheBand(t *testing.T) {
	f := newFake()
	w := holding(f,
		card(1, "Write the spec", "todo"),
		engine.Card{Number: 2, Title: "Waiting on them", State: "blocked"})
	w = w.clicked(middleOf(w, 1, 0))
	if w.Chosen() != 1 {
		t.Fatalf("the board chose %d", w.Chosen())
	}
	down := w
	for range 6 {
		down = down.Key("down")
		if down.Chosen() == 2 {
			return
		}
	}
	t.Errorf("walking down never reached the blocked task; it stopped on %d", down.Chosen())
}

// A card is moved with the keyboard. The terminal has done it from the start;
// the window had only the drag, so the central gesture of the product needed a
// mouse.
func TestACardIsMovedWithTheKeyboard(t *testing.T) {
	f := newFake()
	w := holding(f, card(1, "Write the spec", "backlog"), card(2, "Ship it", "backlog"))
	w = w.clicked(middleOf(w, 0, 0))

	w.Key("shift+right")
	var edited *asked
	for i, one := range f.asked {
		if one.op.Verb == engine.TaskEdit.Verb {
			edited = &f.asked[i]
		}
	}
	if edited == nil {
		t.Fatal("shift+right asked for no change of column")
	}
	if edited.in.Given["state"] != "todo" {
		t.Errorf("shift+right asked %v, want the next column", edited.in.Given)
	}
}

// Enter agrees to a deletion and d does not. The question was opened by d and
// agreed to by d, so a quick dd — which is how one deletes in the grammar this
// product borrows — removed a task before a word of it had been read.
func TestADeletionIsAgreedToWithEnter(t *testing.T) {
	f := newFake()
	w := holding(f, card(1, "Write the spec", "todo"))
	w = w.clicked(middleOf(w, 1, 0)).Key("d")
	if w.Screen() != OnAsking {
		t.Fatalf("d left the window on %v", w.Screen())
	}
	if again := w.Key("d"); len(f.asked) != 0 {
		t.Errorf("a second d deleted without being agreed to: %v, screen %v", f.asked, again.Screen())
	}
	if agreed := w.Key("enter"); agreed.Screen() != OnTheBoard {
		t.Errorf("enter left the question on %v", agreed.Screen())
	}
}

// The same gesture as the terminal's: the column the board is on goes off it in
// one, rather than a card at a time.
func TestAWholeColumnGoesOffTheBoardAtOnce(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "One", State: "done"},
		engine.Card{Number: 2, Title: "Two", State: "done"})
	w = w.clicked(middleOf(w, 3, 0))
	w = Typed(w, "X", key.ModShift)
	if w.Screen() != OnAsking {
		t.Fatalf("a whole column went without a question: %v", w.Screen())
	}
	w.Choose("Agree")
	for _, asked := range f.asked {
		if asked.op.Thing == "task" && asked.op.Verb == "archive" {
			if asked.in.Given["state"] != "done" {
				t.Fatalf("the sweep narrowed to %q", asked.in.Given["state"])
			}
			return
		}
	}
	t.Fatalf("nothing was swept: %v", f.asked)
}

// The central gesture of the product — move this card — wanted a mouse in the
// window: the keyboard walked the board and could not move anything on it.
func TestShiftAndAnArrowMovesTheChosenCard(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "One", State: "todo"},
		engine.Card{Number: 2, Title: "Two", State: "todo"})
	w = w.clicked(middleOf(w, 1, 0))

	// Sideways: the column it lands in is its new state, either way.
	w.Key("shift+right")
	asked := askedFor(f, "task", "edit")
	if asked == nil || asked.Given["state"] != "current" {
		t.Fatalf("shift+right asked %v", f.asked)
	}
	f.asked = nil
	w.Key("shift+left")
	back := askedFor(f, "task", "edit")
	if back == nil || back.Given["state"] != "backlog" {
		t.Fatalf("shift+left asked %v", f.asked)
	}
	// And no further than the two ends.
	f.asked = nil
	holding(f, engine.Card{Number: 9, Title: "Last", State: "done"}).
		clicked(middleOf(w, 3, 0)).Key("shift+right")
	if askedFor(f, "task", "edit") != nil {
		t.Fatalf("shift+right off the last column asked %v", f.asked)
	}

	// Down: its place among the ones it sits with, and no further than the end.
	f.asked = nil
	w.Key("shift+down")
	moved := askedFor(f, "task", "move")
	if moved == nil || moved.Given["after"] != "2" {
		t.Fatalf("shift+down asked %v", f.asked)
	}
	f.asked = nil
	w = w.clicked(middleOf(w, 1, 1))
	w.Key("shift+up")
	up := askedFor(f, "task", "move")
	if up == nil || up.Given["first"] != "true" {
		t.Fatalf("shift+up on the second card asked %v", f.asked)
	}
	f.asked = nil
	w.Key("shift+down")
	if askedFor(f, "task", "move") != nil {
		t.Fatalf("shift+down past the end of a column asked %v", f.asked)
	}
}

// Up from the third card or further goes behind the one two above it: the
// window asked for "after" the card directly above, which is where it already
// was, so shift+up did nothing past the second card.
func TestShiftUpFromFurtherDownGoesBehindTheCardTwoAbove(t *testing.T) {
	f := newFake()
	w := holding(f,
		engine.Card{Number: 1, Title: "First", State: "todo"},
		engine.Card{Number: 2, Title: "Second", State: "todo"},
		engine.Card{Number: 3, Title: "Third", State: "todo"})
	w = w.clicked(middleOf(w, 1, 2))
	f.asked = nil
	w.Key("shift+up")
	up := askedFor(f, "task", "move")
	if up == nil || up.Given["after"] != "1" || up.Given["number"] != "3" {
		t.Fatalf("shift+up on the third card asked %v", f.asked)
	}
}

// A card in the band moved sideways goes to the backlog on the left and
// nowhere on the right — never into a column read off a card it is not.
func TestShiftSidewaysOutOfTheBand(t *testing.T) {
	f := newFake()
	w := holding(f,
		engine.Card{Number: 1, Title: "In todo", State: "todo"},
		engine.Card{Number: 2, Title: "Blocked", State: "blocked"})
	w = w.clicked(middleOf(w, 1, 0)).Key("down")
	if w.Chosen() != 2 {
		t.Fatalf("down out of TODO chose %d, want the blocked task", w.Chosen())
	}
	f.asked = nil
	w.Key("shift+right")
	if askedFor(f, "task", "edit") != nil || askedFor(f, "task", "move") != nil {
		t.Fatalf("shift+right out of the band asked %v", f.asked)
	}
	w.Key("shift+left")
	left := askedFor(f, "task", "edit")
	if left == nil || left.Given["state"] != "backlog" || left.Given["number"] != "2" {
		t.Fatalf("shift+left out of the band asked %v", f.asked)
	}
	f.asked = nil
	w.Key("shift+up")
	w.Key("shift+down")
	if askedFor(f, "task", "move") != nil {
		t.Fatalf("shift+up and down in the band asked %v", f.asked)
	}
}

// The band used to be beyond the arrows altogether, and the tasks in it are
// exactly the ones waiting for a decision.
func TestTheArrowsWalkTheBandAndClimbOutOfIt(t *testing.T) {
	f := newFake()
	w := holding(f,
		engine.Card{Number: 1, Title: "In todo", State: "todo"},
		engine.Card{Number: 2, Title: "Blocked one", State: "blocked"},
		engine.Card{Number: 3, Title: "Blocked two", State: "blocked"},
		engine.Card{Number: 4, Title: "Watched", State: "watching"})

	// Down out of the column lands in the lane beneath it.
	w = w.clicked(middleOf(w, 1, 0)).Key("down")
	if w.Chosen() != 2 {
		t.Fatalf("down out of TODO chose %d, want the first blocked task", w.Chosen())
	}
	// Down again walks along the lane, and stops at its end.
	if got := w.Key("down").Chosen(); got != 3 {
		t.Fatalf("down in a lane chose %d, want 3", got)
	}
	if got := w.Key("down").Key("down").Chosen(); got != 3 {
		t.Fatalf("down past the end of a lane chose %d, want 3", got)
	}
	// Sideways moves to the next lane that holds anything, and stops at the
	// last one rather than walking off the end of them.
	if got := w.Key("right").Chosen(); got != 4 {
		t.Fatalf("right out of the blocked lane chose %d, want 4", got)
	}
	if got := w.Key("left").Chosen(); got != 2 {
		t.Fatalf("left out of the first lane chose %d, want to stay", got)
	}
	if got := w.Key("right").Key("right").Chosen(); got != 4 {
		t.Fatalf("right off the last lane chose %d, want to stay on 4", got)
	}
	// And arriving at a shorter lane lands on its last card, not past it.
	if got := w.Key("down").Key("right").Chosen(); got != 4 {
		t.Fatalf("right from the second blocked card chose %d, want 4", got)
	}
	// Sideways out of a column is not a step down into the lane under it.
	if got := w.clicked(middleOf(w, 1, 0)).Key("right").Chosen(); got != 1 {
		t.Fatalf("right out of TODO chose %d, and should not have fallen into the band", got)
	}
	// Up from the head of a lane climbs back into the column above it.
	if got := w.Key("up").Chosen(); got != 1 {
		t.Fatalf("up out of a lane chose %d, want the task above it", got)
	}
}

// Climbing out of the band lands on the last card of the column above, and
// brings it into sight. A card chosen out of sight is a card whose state
// somebody is about to change without seeing which one it is.
func TestClimbingOutOfTheBandBringsTheCardIntoSight(t *testing.T) {
	f := newFake()
	var cards []engine.Card
	for i := 1; i <= 20; i++ {
		cards = append(cards, engine.Card{Number: i, Title: "Task " + strconv.Itoa(i), State: "todo"})
	}
	cards = append(cards, engine.Card{Number: 99, Title: "Stuck", State: "blocked"})
	w := holding(f, cards...)

	w = w.clicked(middleOf(w, 1, 0))
	for range 20 {
		w = w.Key("down")
	}
	if w.Chosen() != 99 {
		t.Fatalf("down out of TODO chose %d, want the blocked task", w.Chosen())
	}
	w = w.Key("up")
	if w.Chosen() != 20 {
		t.Fatalf("up out of the band chose %d, want the last card of TODO", w.Chosen())
	}
	run := w.Layout().Column[1]
	last := w.Layout().CardAt(1, 19, w.scroll[1])
	if last.Y < run.Y || last.Y+last.H > run.Y+run.H {
		t.Fatalf("the card chosen sits at %d..%d, outside the column %d..%d",
			last.Y, last.Y+last.H, run.Y, run.Y+run.H)
	}
}

// A keyboard move the store refuses leaves the window saying so, rather than
// redrawing as though it had worked.
func TestAKeyboardMoveTheStoreRefusesIsSaid(t *testing.T) {
	for _, c := range []struct {
		what string
		key  string
	}{
		{"sideways", "shift+right"},
		{"up or down", "shift+down"},
	} {
		t.Run(c.what, func(t *testing.T) {
			f := newFake()
			f.failOp = errNope
			w := holding(f, engine.Card{Number: 1, Title: "One", State: "todo"},
				engine.Card{Number: 2, Title: "Two", State: "todo"})
			w = w.clicked(middleOf(w, 1, 0)).Key(c.key)
			if w.Wrong() == nil {
				t.Fatalf("%s was refused and the window says nothing", c.what)
			}
		})
	}
}

// The backlog stands alone, full height, with no lane under it: walking off the
// bottom of it goes nowhere rather than into the band.
func TestWalkingOffTheBottomOfTheBacklogGoesNowhere(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "One", State: "backlog"},
		engine.Card{Number: 2, Title: "Stuck", State: "blocked"})
	w = w.clicked(middleOf(w, 0, 0))
	if got := w.Key("down").Chosen(); got != 1 {
		t.Fatalf("down out of the backlog chose %d, want to stay on 1", got)
	}
}

// Climbing out of a lane whose column above is empty stays where it is, rather
// than reaching for a card that is not there.
func TestClimbingOutOfALaneWithNothingAboveItStays(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "Stuck", State: "blocked"})
	w = w.clicked(onLaneCard(w, 0, 0))
	if w.Chosen() != 1 {
		t.Fatalf("the blocked card was not chosen: %d", w.Chosen())
	}
	if got := w.Key("up").Chosen(); got != 1 {
		t.Fatalf("up out of a lane with nothing above it chose %d", got)
	}
}

// Walking left between lanes reaches the one before, which is the lane sitting
// at nought — the end a guard written one off would never let it reach.
func TestWalkingLeftReachesTheFirstLane(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "Stuck", State: "blocked"},
		engine.Card{Number: 2, Title: "Watched", State: "watching"})
	w = w.clicked(onLaneCard(w, 1, 0))
	if w.Chosen() != 2 {
		t.Fatalf("the watched card was not chosen: %d", w.Chosen())
	}
	if got := w.Key("left").Chosen(); got != 1 {
		t.Fatalf("left out of the watching lane chose %d, want the blocked one", got)
	}
}
