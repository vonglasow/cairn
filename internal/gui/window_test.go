package gui

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"testing"

	"cairn/internal/engine"
	"cairn/internal/reading"
)

// What a gesture decides, read without drawing anything. The window keeps its
// state in one struct of ours, a gesture turns one into another, and a test
// reads what the store was asked. Gio is immediate mode and would rather each
// widget kept its own state; letting that happen is how this surface would
// stop being testable at all.

func TestAClickChoosesTheCardUnderIt(t *testing.T) {
	f := newFake()
	w := holding(f, card(1, "Write the spec", "todo"), card(2, "Ship it", "todo"))

	w = w.clicked(middleOf(w, 1, 0))
	if w.Chosen() != 1 {
		t.Fatalf("chose %d, want task 1", w.Chosen())
	}
	w = w.clicked(middleOf(w, 1, 1))
	if w.Chosen() != 2 {
		t.Fatalf("chose %d, want task 2", w.Chosen())
	}
}

func TestAClickOnNothingChoosesNothing(t *testing.T) {
	f := newFake()
	w := holding(f, card(1, "Write the spec", "todo"))
	w = w.clicked(middleOf(w, 1, 0))
	w = w.clicked(Point{X: 0, Y: 0})
	if w.Chosen() != 0 {
		t.Fatalf("clicking the margin left %d chosen", w.Chosen())
	}
}

func TestADoubleClickOpensTheDetail(t *testing.T) {
	f := newFake()
	w := holding(f, card(1, "Write the spec", "todo"))

	w = w.doubleClicked(middleOf(w, 1, 0))
	if w.Screen() != OnDetail {
		t.Fatalf("the window is on %v, want the detail", w.Screen())
	}
	if w.Chosen() != 1 {
		t.Fatalf("the detail is open on %d", w.Chosen())
	}
	if f.what("trace", "show") != 1 {
		t.Error("the detail did not ask for the task's history")
	}
	if w = w.Key("esc"); w.Screen() != OnTheBoard {
		t.Fatalf("escape left the window on %v", w.Screen())
	}
}

// A click that barely moves is a click. Without the threshold, a shaky hand
// would move a task every time it chose one.
func TestAClickThatBarelyMovesIsStillAClick(t *testing.T) {
	f := newFake()
	w := holding(f, card(1, "Write the spec", "backlog"))

	at := middleOf(w, 0, 0)
	w = w.PressAt(at)
	w = w.MoveTo(Point{X: at.X + 2, Y: at.Y + 1})
	w = w.LetGo(Point{X: at.X + 2, Y: at.Y + 1})

	if w.Chosen() != 1 {
		t.Errorf("chose %d, want task 1", w.Chosen())
	}
	if len(f.asked) != 0 {
		t.Errorf("the store was asked %d things for a click", len(f.asked))
	}
}

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

func TestTheWheelScrollsOneColumnAndNoOther(t *testing.T) {
	f := newFake()
	var many []reading.Card
	for i := 1; i <= 12; i++ {
		many = append(many, card(i, "Task", "backlog"))
	}
	w := holding(f, many...)

	w = w.Scroll(middleOf(w, 0, 0), CardHeight)
	if w.scroll[0] == 0 {
		t.Error("the backlog did not scroll")
	}
	if w.scroll[1] != 0 {
		t.Error("a column nobody pointed at scrolled")
	}
}

func TestScrollingStopsAtBothEnds(t *testing.T) {
	f := newFake()
	w := holding(f, card(1, "Write the spec", "backlog"))

	w = w.Scroll(middleOf(w, 0, 0), -CardHeight)
	if w.scroll[0] != 0 {
		t.Errorf("scrolled to %d above the first card", w.scroll[0])
	}
	w = w.Scroll(middleOf(w, 0, 0), CardHeight*20)
	if w.scroll[0] != 0 {
		t.Errorf("one card scrolled to %d, and it all fits", w.scroll[0])
	}
}

func TestAWindowTooSmallSaysSo(t *testing.T) {
	f := newFake()
	w := New(f, 320, 240)
	if w.Screen() != TooSmallToDraw {
		t.Fatalf("the window is on %v", w.Screen())
	}
	w = w.Resized(1200, 800)
	if w.Screen() != OnTheBoard {
		t.Fatalf("growing back left the window on %v", w.Screen())
	}
}

func TestTheWindowFollowsTheStoreOnlyWhenItMoved(t *testing.T) {
	f := newFake()
	f.version = 7
	w := holding(f, card(1, "Write the spec", "backlog"))
	w.version = 7

	w = w.Wonder()
	if f.what("task", "list") != 0 {
		t.Error("a store nobody touched was read again")
	}
	f.version = 8
	w = w.Wonder()
	if f.what("task", "list") != 1 {
		t.Error("a store somebody else wrote to was not read again")
	}
}

func TestAStoreOutOfReachIsSaidRatherThanDrawnEmpty(t *testing.T) {
	f := newFake()
	f.fail = errNope
	w := New(f, 1200, 800)
	w = w.Wonder()
	if w.Wrong() == nil {
		t.Fatal("a store out of reach drew as though it were empty")
	}
}

// A click in a real window is a press and a release, so everything a click does
// has to happen on the release — including the things that are not cards. This
// is what stopped the tabs being reachable at all.
func TestAReleaseWhereNoCardWasHeldIsStillAClick(t *testing.T) {
	f := newFake()
	w := holding(f, reading.Card{Number: 1, Title: "Write the spec", State: "todo"})
	tab := TabAt(w.Layout(), 1)
	at := Point{X: tab.X + tab.W/2, Y: tab.Y + tab.H/2}

	w = w.PressAt(at)
	w = w.LetGo(at)
	if w.Tab() != Objectives {
		t.Fatalf("pressing and releasing on a tab left the window on %v", w.Tab())
	}
}

// The bar offers what can be done here, and clicking an offer does it. Without
// this there is no way to edit a task at all unless somebody knows a shortcut
// nobody wrote down.
func TestTheBarOffersEditingAndClickingItOpensTheForm(t *testing.T) {
	f := newFake()
	w := holding(f, reading.Card{Number: 1, Title: "Write the spec", State: "todo"})
	w = w.clicked(middleOf(w, 1, 0))

	at := -1
	for i, offer := range w.Offers() {
		if offer == "Edit" {
			at = i
		}
	}
	if at < 0 {
		t.Fatalf("the bar offers %v, and none of them edits", w.Offers())
	}
	box := OfferAt(w.Layout(), at)
	w = w.clicked(Point{X: box.X + box.W/2, Y: box.Y + box.H/2})
	if w.Screen() != OnForm {
		t.Fatalf("clicking Edit left the window on %v", w.Screen())
	}
}

// With nothing chosen there is nothing to edit, and the bar does not offer it.
func TestTheBarOffersNothingToEditWithNothingChosen(t *testing.T) {
	f := newFake()
	w := holding(f, reading.Card{Number: 1, Title: "Write the spec", State: "todo"})
	for _, offer := range w.Offers() {
		if offer == "Edit" {
			t.Error("the bar offers to edit with nothing chosen")
		}
	}
}

// The objectives tab lists what the work is for, and one of them can be chosen
// and edited the same way a task is.
func TestTheObjectivesTabListsChoosesAndEdits(t *testing.T) {
	f := newFake()
	w := New(f, 1200, 800)
	w.snap = reading.Snapshot{Goals: []reading.Goal{
		{Number: 3, Title: "Ship cairn", State: "active"},
		{Number: 4, Title: "Keep it honest", State: "active"},
	}}
	tab := TabAt(w.Layout(), 1)
	w = w.clicked(Point{X: tab.X + tab.W/2, Y: tab.Y + tab.H/2})

	second := GoalAt(w.Layout(), 1, 0)
	w = w.clicked(Point{X: second.X + second.W/2, Y: second.Y + second.H/2})
	if w.Chosen() != 4 {
		t.Fatalf("clicking the second objective chose %d", w.Chosen())
	}
	w = w.Choose("Edit")
	if !w.Form().ForObjective {
		t.Fatal("editing an objective opened a task's form")
	}
	w = w.Choose("Save")
	if f.what("objective", "edit") != 1 {
		t.Errorf("saving asked for %d objective edits", f.what("objective", "edit"))
	}
}

// A click below the last objective chooses nothing.
func TestClickingBelowTheObjectivesChoosesNothing(t *testing.T) {
	f := newFake()
	w := New(f, 1200, 800)
	w.snap = reading.Snapshot{Goals: []reading.Goal{{Number: 3, Title: "Ship cairn", State: "active"}}}
	w.tab = Objectives

	below := GoalAt(w.Layout(), 4, 0)
	if got := w.clicked(Point{X: below.X + below.W/2, Y: below.Y + below.H/2}); got.Chosen() != 0 {
		t.Errorf("clicking below the list chose %d", got.Chosen())
	}
}

// The detail shows everything a task carries. Only the history is what a card
// already marks; the rest is what a card had no room for.
func TestTheDetailShowsEveryFieldOfTheTask(t *testing.T) {
	f := newFake()
	w := holding(f, reading.Card{
		Number: 1, Title: "Write the spec", State: "todo", Priority: "high",
		Description: "The client wants two columns",
	})

	w = w.doubleClicked(middleOf(w, 1, 0))
	shown := w.Shown()
	if len(shown) == 0 {
		t.Fatal("the detail shows nothing of the task")
	}
	want := map[string]string{"title": "Write the spec", "column": "todo", "priority": "high"}
	for name, value := range want {
		found := false
		for _, field := range shown {
			if field.Name == name {
				found = true
				if field.Value != value {
					t.Errorf("the detail shows %q as %q, want %q", name, field.Value, value)
				}
			}
		}
		if !found {
			t.Errorf("the detail does not show %q", name)
		}
	}
}

// With nothing open there is nothing to show.
func TestTheDetailShowsNothingWithNoTaskOpen(t *testing.T) {
	f := newFake()
	w := holding(f, reading.Card{Number: 1, Title: "Write the spec", State: "todo"})
	if got := w.Shown(); len(got) != 0 {
		t.Errorf("a board with no detail open shows %d fields", len(got))
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

// Double-clicking an objective opens that objective, and not whatever task
// happens to lie under the same row of the board behind it.
func TestDoubleClickingAnObjectiveOpensThatObjective(t *testing.T) {
	f := newFake()
	w := New(f, 1200, 800)
	w.snap = reading.Snapshot{
		Cards: []reading.Card{{Number: 1, Title: "Write the spec", State: "backlog"}},
		Goals: []reading.Goal{
			{Number: 3, Title: "Ship cairn", State: "active"},
			{Number: 4, Title: "Keep it honest", State: "active"},
		},
	}
	w.tab = Objectives

	for at, want := range map[int]int{0: 3, 1: 4} {
		box := GoalAt(w.Layout(), at, 0)
		got := w.doubleClicked(Point{X: box.X + box.W/2, Y: box.Y + box.H/2})
		if got.Screen() != OnDetail {
			t.Fatalf("double-clicking objective %d left the window on %v", at, got.Screen())
		}
		if got.Chosen() != want {
			t.Errorf("double-clicking objective %d opened %d, want %d", at, got.Chosen(), want)
		}
	}
}

// A panel over the board has the window. A click that lands on the board
// behind it does nothing — changing what an open detail shows by clicking
// behind it is the sort of thing nobody can explain afterwards.
func TestAPanelOverTheBoardHasTheWindow(t *testing.T) {
	f := newFake()
	w := New(f, 1200, 800)
	w.snap = reading.Snapshot{Goals: []reading.Goal{
		{Number: 3, Title: "Ship cairn", State: "active"},
		{Number: 4, Title: "Keep it honest", State: "active"},
	}}
	w.tab = Objectives

	first := GoalAt(w.Layout(), 0, 0)
	w = w.doubleClicked(Point{X: first.X + first.W/2, Y: first.Y + first.H/2})
	if w.Chosen() != 3 {
		t.Fatalf("the detail opened on %d", w.Chosen())
	}

	second := GoalAt(w.Layout(), 1, 0)
	w = w.clicked(Point{X: second.X + second.W/2, Y: second.Y + second.H/2})
	if w.Chosen() != 3 {
		t.Errorf("clicking behind an open detail changed it to %d", w.Chosen())
	}

	// The bar still answers: it is what closes the panel.
	for i, offer := range w.Offers() {
		if offer != "Close" {
			continue
		}
		box := OfferAt(w.Layout(), i)
		w = w.clicked(Point{X: box.X + box.W/2, Y: box.Y + box.H/2})
	}
	if w.Screen() != OnTheBoard {
		t.Fatalf("Close left the window on %v", w.Screen())
	}
}

// Nor can a card be picked up from behind a panel.
func TestNoCardIsPickedUpFromBehindAPanel(t *testing.T) {
	f := newFake()
	w := holding(f, reading.Card{Number: 1, Title: "Write the spec", State: "todo"})
	w = w.doubleClicked(middleOf(w, 1, 0))

	w = w.PressAt(middleOf(w, 1, 0))
	w = w.MoveTo(Point{X: 20, Y: 400})
	if w.Dragging() {
		t.Error("a card was dragged out from behind an open detail")
	}
}

// An objective's detail shows what the objective carries, the same way a
// task's does.
func TestTheDetailOfAnObjectiveShowsWhatItCarries(t *testing.T) {
	f := newFake()
	w := New(f, 1200, 800)
	w.snap = reading.Snapshot{Goals: []reading.Goal{{Number: 3, Title: "Ship cairn", State: "active"}}}
	w.tab = Objectives

	box := GoalAt(w.Layout(), 0, 0)
	w = w.doubleClicked(Point{X: box.X + box.W/2, Y: box.Y + box.H/2})

	shown := w.Shown()
	if len(shown) == 0 {
		t.Fatal("an objective's detail shows nothing")
	}
	for _, field := range shown {
		if field.Name == "title" && field.Value != "Ship cairn" {
			t.Errorf("it shows the title as %q", field.Value)
		}
		if field.Name == "blockers" {
			t.Error("an objective's detail asks what is blocking it, and nothing blocks an objective")
		}
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

// A number does not say which kind of thing it is: task 1 and objective 1 both
// exist, and cairn numbers them apart. A board that remembered only the number
// opened the task when an objective was pointed at.
func TestATaskAndAnObjectiveOfTheSameNumberAreToldApart(t *testing.T) {
	f := newFake()
	w := New(f, 1200, 800)
	w.snap = reading.Snapshot{
		Cards: []reading.Card{{Number: 1, Title: "Write the spec", State: "backlog"}},
		Goals: []reading.Goal{{Number: 1, Title: "Ship cairn", State: "active"}},
	}

	// On the objectives tab, the objective.
	w.tab = Objectives
	box := GoalAt(w.Layout(), 0, 0)
	goal := w.doubleClicked(Point{X: box.X + box.W/2, Y: box.Y + box.H/2})
	if got := goal.Shown(); len(got) == 0 || got[0].Value != "Ship cairn" {
		t.Errorf("the detail of objective 1 shows %v", got)
	}
	if goal.Edit().Form().ForObjective != true {
		t.Error("editing objective 1 opened a task's form")
	}

	// On the tasks tab, the task.
	w.tab = Tasks
	task := w.doubleClicked(middleOf(w, 0, 0))
	if got := task.Shown(); len(got) == 0 || got[0].Value != "Write the spec" {
		t.Errorf("the detail of task 1 shows %v", got)
	}
	if task.Edit().Form().ForObjective {
		t.Error("editing task 1 opened an objective's form")
	}
}

// e edits whatever is chosen, because a hand that has just chosen something is
// not on the other side of the window.
func TestEEditsWhateverIsChosen(t *testing.T) {
	f := newFake()
	w := holding(f, reading.Card{Number: 1, Title: "Write the spec", State: "todo"})

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

// Deleting asks first, because it is the one thing here that loses work.
func TestDeletingAsksBeforeItLosesAnything(t *testing.T) {
	f := newFake()
	w := holding(f, reading.Card{Number: 1, Title: "Write the spec", State: "todo"})
	w = w.clicked(middleOf(w, 1, 0))

	asked := w.Choose("Delete")
	if asked.Screen() != OnAsking {
		t.Fatalf("Delete left the window on %v", asked.Screen())
	}
	if len(f.asked) != 0 {
		t.Fatal("it deleted before asking")
	}
	if got := asked.Choose("Cancel"); got.Screen() != OnTheBoard || len(f.asked) != 0 {
		t.Error("cancelling deleted it anyway")
	}
	if got := asked.Choose("Delete"); f.what("task", "delete") != 1 {
		t.Errorf("agreeing asked for %d deletes and left the window on %v",
			f.what("task", "delete"), got.Screen())
	}
}

// An objective is deleted the same way, through its own operation.
func TestDeletingAnObjective(t *testing.T) {
	f := newFake()
	w := New(f, 1200, 800)
	w.snap = reading.Snapshot{Goals: []reading.Goal{{Number: 3, Title: "Ship cairn", State: "active"}}}
	w.tab, w.chosen, w.chosenGoal = Objectives, 3, true

	w = w.Choose("Delete").Choose("Delete")
	if f.what("objective", "delete") != 1 {
		t.Errorf("it asked for %d objective deletes", f.what("objective", "delete"))
	}
}

// Nothing chosen is nothing to delete, and the bar does not offer it.
func TestDeletingIsNotOfferedWithNothingChosen(t *testing.T) {
	f := newFake()
	w := holding(f, reading.Card{Number: 1, Title: "Write the spec", State: "todo"})
	for _, offer := range w.Offers() {
		if offer == "Delete" {
			t.Error("the bar offers to delete with nothing chosen")
		}
	}
}

// A search narrows what is drawn to what it reaches, and reaches what a card
// does not show.
func TestSearchingNarrowsTheBoardToWhatItReaches(t *testing.T) {
	f := newFake()
	w := holding(f,
		reading.Card{Number: 1, Title: "Write the spec", State: "backlog", Description: "two columns"},
		reading.Card{Number: 2, Title: "Ship the board", State: "backlog"})

	w = w.Key("/")
	if w.Screen() != OnSearch {
		t.Fatalf("/ left the window on %v", w.Screen())
	}
	w = w.Types("spec")
	if got := len(w.Showing().Cards); got != 1 {
		t.Fatalf("a search for \"spec\" reaches %d tasks", got)
	}

	deep := w.Key("esc").Key("/").Types("columns")
	if got := len(deep.Showing().Cards); got != 1 {
		t.Errorf("a search of the description reaches %d tasks", got)
	}
	if got := deep.Showing().Cards[0].Number; got != 1 {
		t.Errorf("it reached task %d", got)
	}
}

// Leaving a search puts everything back, because a search is a way of looking
// and not a way of changing.
func TestLeavingASearchPutsEverythingBack(t *testing.T) {
	f := newFake()
	w := holding(f,
		reading.Card{Number: 1, Title: "Write the spec", State: "backlog"},
		reading.Card{Number: 2, Title: "Ship the board", State: "backlog"})

	w = w.Key("/").Types("spec")
	if got := len(w.Showing().Cards); got != 1 {
		t.Fatalf("the search reaches %d", got)
	}
	w = w.Key("esc")
	if got := len(w.Showing().Cards); got != 2 {
		t.Errorf("leaving the search left %d tasks showing", got)
	}
	if w.Screen() != OnTheBoard {
		t.Errorf("leaving the search left the window on %v", w.Screen())
	}
}

// A search that reaches nothing draws nothing rather than everything: an empty
// answer is an answer.
func TestASearchThatReachesNothingShowsNothing(t *testing.T) {
	f := newFake()
	w := holding(f, reading.Card{Number: 1, Title: "Write the spec", State: "backlog"})
	w = w.Key("/").Types("nothing like it")
	if got := len(w.Showing().Cards); got != 0 {
		t.Errorf("a search that reaches nothing shows %d tasks", got)
	}
}

// A search reaches objectives too, on the tab that shows them.
func TestASearchReachesObjectives(t *testing.T) {
	f := newFake()
	w := New(f, 1200, 800)
	w.snap = reading.Snapshot{Goals: []reading.Goal{
		{Number: 3, Title: "Ship cairn", State: "active"},
		{Number: 4, Title: "Keep it honest", State: "active"},
	}}
	w.tab = Objectives

	w = w.Key("/").Types("honest")
	if got := len(w.Showing().Goals); got != 1 {
		t.Fatalf("a search of the objectives reaches %d", got)
	}
	if got := w.Showing().Goals[0].Number; got != 4 {
		t.Errorf("it reached objective %d", got)
	}
}

// The help says what everything does, which is the only thing here that has to
// be read rather than pointed at.
func TestTheHelpSaysWhatEverythingDoes(t *testing.T) {
	f := newFake()
	w := holding(f, reading.Card{Number: 1, Title: "Write the spec", State: "todo"})

	w = w.Choose("Help")
	if w.Screen() != OnHelp {
		t.Fatalf("Help left the window on %v", w.Screen())
	}
	said := w.Helping()
	if len(said) == 0 {
		t.Fatal("the help says nothing")
	}
	for _, about := range []string{"drag", "double-click", "/", "e"} {
		found := false
		for _, line := range said {
			found = found || strings.Contains(line.Name, about) || strings.Contains(line.Value, about)
		}
		if !found {
			t.Errorf("the help says nothing about %q", about)
		}
	}
	if got := w.Key("esc"); got.Screen() != OnTheBoard {
		t.Fatalf("escape left the help on %v", got.Screen())
	}
}

// Backspace in a search takes back one character of what is being looked for,
// and one character is one rune.
func TestBackspaceInASearchTakesBackOneCharacter(t *testing.T) {
	f := newFake()
	w := holding(f, reading.Card{Number: 1, Title: "réunion", State: "backlog"})

	w = w.Key("/").Types("réu")
	if got := w.Key("backspace").Looking(); got != "ré" {
		t.Fatalf("taking one back left %q", got)
	}
	if got := w.Key("backspace").Key("backspace").Looking(); got != "r" {
		t.Errorf("taking two back left %q", got)
	}
	// Taking back what is not there leaves nothing rather than failing.
	empty := w.Key("esc").Key("/")
	for range 3 {
		empty = empty.Key("backspace")
	}
	if got := empty.Looking(); got != "" {
		t.Errorf("an empty search reads %q", got)
	}
	// And backspace on the board itself is not a search at all.
	if got := w.Key("esc").Key("backspace").Looking(); got != "" {
		t.Errorf("backspace on the board left a search of %q", got)
	}
}

// The arrows walk the board, so a hand on the keyboard does not have to reach
// for the mouse to look at the next task.
func TestTheArrowsWalkTheBoard(t *testing.T) {
	f := newFake()
	w := holding(f,
		reading.Card{Number: 1, Title: "In the heap", State: "backlog"},
		reading.Card{Number: 2, Title: "First", State: "todo"},
		reading.Card{Number: 3, Title: "Second", State: "todo"})

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
		reading.Card{Number: 1, Title: "Only", State: "backlog"},
		reading.Card{Number: 2, Title: "Last", State: "done"})

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
	w := holding(f, reading.Card{Number: 1, Title: "Only", State: "backlog"})
	if got := w.Key("down").Chosen(); got != 1 {
		t.Errorf("an arrow on an untouched board chose %d", got)
	}
}

// The arrows walk the objectives too, which are a list and so have no columns.
func TestTheArrowsWalkTheObjectives(t *testing.T) {
	f := newFake()
	w := New(f, 1200, 800)
	w.snap = reading.Snapshot{Goals: []reading.Goal{
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
	w := holding(f, reading.Card{Number: 1, Title: "Write the spec", State: "todo"})

	if got := w.Key("enter"); got.Screen() != OnTheBoard {
		t.Error("enter with nothing chosen opened something")
	}
	w = w.clicked(middleOf(w, 1, 0)).Key("enter")
	if w.Screen() != OnDetail {
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
		reading.Card{Number: 1, Title: "One", State: "backlog"},
		reading.Card{Number: 2, Title: "Two", State: "backlog"},
		reading.Card{Number: 3, Title: "Three", State: "backlog"},
		reading.Card{Number: 4, Title: "Only", State: "todo"})

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
	w.snap = reading.Snapshot{Goals: []reading.Goal{
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
	w := holding(f, reading.Card{Number: 7, Title: "Only", State: "current"})
	if got := w.Key("down").Chosen(); got != 7 {
		t.Errorf("an arrow chose %d, and the only card is 7", got)
	}
}

// Every panel says what it is, and says which thing it is about. Read only by
// the drawing, so it is asked for here: a heading nothing checks is a heading
// that goes out saying the wrong number.
func TestWhatEachPanelSaysItIs(t *testing.T) {
	f := newFake()
	w := holding(f, reading.Card{Number: 7, Title: "Write the spec", State: "todo"})
	if got := Heads(w); got != "" {
		t.Errorf("the board is headed %q, and it is its own heading", got)
	}
	on := w.clicked(middleOf(w, 1, 0))
	// The title goes with the number on an open task: the line was empty, it
	// holds ninety-nine characters at the smallest window, and the field below
	// it cuts a title at thirty.
	if got := Heads(on.Key("enter")); got != "task 7 · Write the spec" {
		t.Errorf("an open task is headed %q", got)
	}
	// Not while it is being edited: the title is in the field being typed
	// into, and a heading repeating it would go stale between two keystrokes.
	if got := Heads(on.Choose("Edit")); got != "task 7" {
		t.Errorf("a task being edited is headed %q", got)
	}
	if got := Heads(on.Choose("Delete")); got != "Delete task 7?" {
		t.Errorf("the question is headed %q", got)
	}
	if got := Heads(w.Key("?")); got != "What everything does" {
		t.Errorf("the help is headed %q", got)
	}
	if got := Heads(w.Key("cmd+n")); got != "New task" {
		t.Errorf("adding a task is headed %q", got)
	}
	second := TabAt(w.Layout(), 1)
	goals := w.clicked(Point{X: second.X + second.W/2, Y: second.Y + second.H/2})
	if got := Heads(goals.Key("cmd+n")); got != "New objective" {
		t.Errorf("adding an objective is headed %q", got)
	}
}

// The other page is the one somebody is not on, from either of them. Both tab
// and shift+tab ask for it: with two pages, which way round a walk goes is a
// difference nothing can see.
func TestTheOtherPageIsTheOneNobodyIsOn(t *testing.T) {
	f := newFake()
	w := holding(f, reading.Card{Number: 1, Title: "Write the spec", State: "todo"})
	if got := w.other().Tab(); got != Objectives {
		t.Errorf("the other page from the first is %v", got)
	}
	if got := w.other().other().Tab(); got != Tasks {
		t.Errorf("the other page from the second is %v", got)
	}
	if got := w.Key("shift+tab").Tab(); got != Objectives {
		t.Errorf("shift+tab from the first page is %v", got)
	}
	if got := w.Key("tab").Key("shift+tab").Tab(); got != Tasks {
		t.Errorf("tab then shift+tab is %v", got)
	}
	// And changing page lets go of what was chosen: a card chosen on the page
	// you left is a card the arrows would walk from on the page you are on.
	if got := w.clicked(middleOf(w, 1, 0)).other(); got.Chosen() != 0 {
		t.Errorf("after changing page, task %d is still chosen", got.Chosen())
	}
}

// Where what is chosen is drawn is the box of that very thing, and not of
// whatever else happened to be looked at first.
func TestWhereWhatIsChosenIsDrawn(t *testing.T) {
	f := newFake()
	w := holding(f,
		reading.Card{Number: 1, Title: "First", State: "todo"},
		reading.Card{Number: 2, Title: "Second", State: "todo"})
	l := w.Layout()

	first := w.clicked(middleOf(w, 1, 0))
	box, run, ok := first.WhereChosen()
	if !ok {
		t.Fatal("the first card is chosen and drawn nowhere")
	}
	if box != l.CardAt(1, 0, 0) {
		t.Errorf("the first card is drawn at %+v, want %+v", box, l.CardAt(1, 0, 0))
	}
	if run != l.Column[1] {
		t.Errorf("it is drawn in %+v, want the column %+v", run, l.Column[1])
	}
	second := first.Key("down")
	if got, _, _ := second.WhereChosen(); got != l.CardAt(1, 1, second.scroll[1]) {
		t.Errorf("the second card is drawn at %+v, want %+v", got, l.CardAt(1, 1, second.scroll[1]))
	}
	// Nothing chosen is drawn nowhere, and says so rather than answering a box.
	if _, _, ok := w.WhereChosen(); ok {
		t.Error("with nothing chosen, something was said to be drawn")
	}
}

// The wheel moves a history by whole lines, and by as many as it was turned.
func TestTheWheelMovesAHistoryByWholeLines(t *testing.T) {
	f := newFake()
	for range 40 {
		f.acts = append(f.acts, anAct("moved", "column", "todo", "current"))
	}
	w := holding(f, reading.Card{Number: 1, Title: "Write the spec", State: "todo"})
	w = w.clicked(middleOf(w, 1, 0)).Key("enter")
	over := HistoryBox(Over(w.width, w.height))
	at := Point{X: over.X + over.W/2, Y: over.Y + over.H/2}

	if got := w.Scroll(at, 3*lineHigh).History(); got != 3 {
		t.Errorf("turned by three lines, the history is at %d", got)
	}
	if got := w.Scroll(at, lineHigh-1).History(); got != 0 {
		t.Errorf("turned by less than a line, the history is at %d", got)
	}
	// And the wheel anywhere else over the panel moves nothing.
	if got := w.Scroll(Point{X: 1, Y: 1}, 9*lineHigh).History(); got != 0 {
		t.Errorf("turned away from the history, it is at %d", got)
	}
}

// The objectives are a grid, so how far they scroll is reckoned in rows and
// not in objectives — and how far they may scroll is reckoned from the rows
// there are against the room there is.
func TestHowFarTheObjectivesScroll(t *testing.T) {
	f := newFake()
	w := New(f, 1200, 800)
	for i := 1; i <= 60; i++ {
		w.snap.Goals = append(w.snap.Goals, reading.Goal{Number: i, Title: "Ship it", State: "live"})
	}
	w.tab = Objectives
	run := GoalsRun(w.Layout())
	rows := GoalRows(len(w.snap.Goals))

	// The far end is every row there is, less the room there is to show them.
	end := rows*CardHeight - run.H
	if end <= 0 {
		t.Fatalf("%d objectives in a run of %d all fit: this proves nothing", len(w.snap.Goals), run.H)
	}
	if got := w.settledGoals(9999); got != end {
		t.Errorf("scrolled past the end, the objectives stand at %d, want %d", got, end)
	}
	if got := w.settledGoals(-9); got != 0 {
		t.Errorf("scrolled before the beginning, the objectives stand at %d", got)
	}

	// Following the second objective moves nothing: it is on the first row.
	if got := w.followingGoal(1).goals; got != 0 {
		t.Errorf("following the second objective moved the grid to %d", got)
	}
	// Following the last one puts its row against the bottom of the run.
	last := len(w.snap.Goals) - 1
	if got, want := w.followingGoal(last).goals, end; got != want {
		t.Errorf("following the last objective moved the grid to %d, want %d", got, want)
	}
	// And the first one on the row past what fits moves just far enough to
	// show that row's bottom, which is less than a whole row when the run is
	// not a whole number of rows tall.
	fits := run.H / CardHeight
	if got, want := w.followingGoal(fits*GoalsAcross()).goals, (fits+1)*CardHeight-run.H; got != want {
		t.Errorf("following the first objective past what fits moved to %d, want %d", got, want)
	}
}

// Where a chosen objective is drawn is the box of that very objective, and not
// of whichever one happened to be looked at first.
func TestWhereAChosenObjectiveIsDrawn(t *testing.T) {
	f := newFake()
	w := New(f, 1200, 800)
	w.snap.Goals = []reading.Goal{
		{Number: 3, Title: "Ship cairn", State: "live"},
		{Number: 4, Title: "Keep it honest", State: "live"},
	}
	w.tab = Objectives
	l := w.Layout()

	w.chosen, w.chosenGoal = 4, true
	box, run, ok := w.WhereChosen()
	if !ok {
		t.Fatal("the chosen objective is drawn nowhere")
	}
	at := 0
	for i, g := range reading.Sorted(w.snap.Goals) {
		if g.Number == 4 {
			at = i
		}
	}
	if want := GoalAt(l, at, 0); box != want {
		t.Errorf("objective 4 is drawn at %+v, want %+v", box, want)
	}
	if run != GoalsRun(l) {
		t.Errorf("it is drawn in %+v, want %+v", run, GoalsRun(l))
	}
	// An objective nobody chose is drawn nowhere, and says so.
	w.chosen = 99
	if _, _, ok := w.WhereChosen(); ok {
		t.Error("an objective that is not there was said to be drawn somewhere")
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
		{"the first past the last", 0, 4, map[string]string{"number": "2", "last": "true"}},
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

// A panel drawn under another keeps its own heading. The form is still there
// while a list is open over it, and both were asking the window what it was
// on — so both read the list's heading and the form said it twice.
func TestAPanelUnderAnotherKeepsItsOwnHeading(t *testing.T) {
	f := newFake()
	w := holding(f, reading.Card{Number: 7, Title: "Write the spec", State: "todo"})
	w = w.clicked(middleOf(w, 1, 0)).Choose("Edit")
	for w.Form().Fields()[w.Form().On()].Arg != "parent" {
		w = w.Key("tab")
	}
	open := w.Key("enter")
	if open.Screen() != OnChoosing {
		t.Fatalf("the list did not open: %v", open.Screen())
	}
	if got := Heads(open); got != "a parent for task 7" {
		t.Errorf("the list is headed %q", got)
	}
	if got := HeadsOn(open, OnForm); got != "task 7" {
		t.Errorf("the form under it is headed %q, not its own", got)
	}
}
