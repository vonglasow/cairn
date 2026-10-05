package gui

import (
	"errors"
	"fmt"
	"slices"

	"strconv"
	"strings"
	"testing"

	"gioui.org/io/key"
	"gioui.org/io/pointer"

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
			found = found || strings.Contains(line.Key, about) || strings.Contains(line.Means, about)
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

// The wheel stops at a panel. It reached the board behind a form, the help and
// the question, so turning it over a modal scrolled something nobody could see.
func TestTheWheelStopsAtAPanel(t *testing.T) {
	f := newFake()
	cards := []reading.Card{}
	for i := 1; i <= 30; i++ {
		cards = append(cards, card(i, "Une de plus", "backlog"))
	}
	w := holding(f, cards...)
	middle := Point{X: w.width / 2, Y: w.height / 2}
	for _, what := range []struct {
		name string
		on   Window
	}{
		{"a form", w.clicked(middleOf(w, 0, 0)).Choose("Edit")},
		{"the help", w.Choose("Help")},
		{"the question", w.clicked(middleOf(w, 0, 0)).Choose("Delete")},
	} {
		was := what.on.scroll[0]
		if got := what.on.Scroll(middle, 5*lineHigh).scroll[0]; got != was {
			t.Errorf("over %s, the wheel moved the column behind it from %d to %d", what.name, was, got)
		}
	}
}

// Only the primary button acts. A right or a middle click arrived as a press
// and a release like any other, so it chose cards and pressed offers while
// nothing on screen suggested it would.
func TestOnlyTheFirstButtonActs(t *testing.T) {
	if Acts(pointer.ButtonSecondary) || Acts(pointer.ButtonTertiary) {
		t.Error("a button other than the first was allowed to act")
	}
	if !Acts(pointer.ButtonPrimary) {
		t.Error("the first button was not allowed to act")
	}
	// A scroll carries no button at all, and has to go through.
	if !Acts(0) {
		t.Error("an event with no button was refused")
	}
}

// An offer drawn on a panel answers where it is drawn. The list is drawn in a
// panel of its own, inset from the one the others use, so its three words were
// painted in one place and listened for in another — words that look like
// buttons and answer only somewhere else.
func TestAnOfferAnswersWhereItIsDrawn(t *testing.T) {
	f := newFake()
	w := New(f, 1200, 800)
	w.snap = reading.Snapshot{
		Cards: []reading.Card{card(1, "Write the spec", "todo"), card(2, "Ship it", "todo")},
	}
	w = w.clicked(middleOf(w, 1, 0)).Choose("Edit")
	for w.Form().Fields()[w.Form().On()].Arg != "parent" {
		w = w.Key("tab")
	}
	w = w.Key("enter")
	if w.Screen() != OnChoosing {
		t.Fatalf("the list did not open: %v", w.Screen())
	}

	box := ChoosingOver(w.width, w.height)
	for i, offer := range w.Offers() {
		at := PanelOfferAt(box, i)
		got, on := w.offerAt(Point{X: at.X + 10, Y: at.Y + at.H/2})
		if !on || got != offer {
			t.Errorf("the word %q drawn at %+v answers %q, %v", offer, at, got, on)
		}
	}
}

// A search can be left with the pointer. The box took the whole bar, so the
// one way out was a key — and the card the search had just found could not be
// acted on at all.
func TestASearchCanBeLeftWithThePointer(t *testing.T) {
	f := newFake()
	w := holding(f, card(1, "Write the spec", "todo")).Choose("Search")
	if w.Screen() != OnSearch {
		t.Fatalf("the window is on %v", w.Screen())
	}
	if len(w.Offers()) == 0 {
		t.Fatal("a search offers nothing, so nothing can be pointed at")
	}
	at := OfferAt(w.Layout(), len(w.Offers())-1)
	got := w.clicked(Point{X: at.X + 10, Y: at.Y + at.H/2})
	if got.Screen() == OnSearch {
		t.Error("clicking the last offer left the window in the search")
	}
}

// A field holding one of a closed set is answered by the pointer, with the
// list a parent is chosen from: what it may hold is read before it is picked.
func TestAClosedSetIsAnsweredByThePointer(t *testing.T) {
	f := newFake()
	w := holding(f, card(1, "Write the spec", "todo"))
	w = w.clicked(middleOf(w, 1, 0)).Choose("Edit")
	box := Over(w.width, w.height)
	fields := w.Form().Fields()
	for i, field := range fields {
		if field.Name != "column" {
			continue
		}
		_, value := FieldRow(box, FieldTop(box, fields, i))
		got := w.clicked(Point{X: value.X + 10, Y: value.Y + value.H/2})
		if got.Screen() != OnChoosing || len(got.Choosing()) != len(field.Choices) {
			t.Errorf("clicking the column field left the window on %v, offering %d",
				got.Screen(), len(got.Choosing()))
		}
		return
	}
	t.Fatal("the form has no column field")
}

// Every bar beside a run of text can be pulled. Two of the four were drawn
// with a grip sized "never smaller than can be taken hold of" and answered
// nothing — a handle that invites the hand and does not move.
func TestEveryBarCanBePulled(t *testing.T) {
	f := newFake()
	for range 40 {
		f.acts = append(f.acts, anAct("moved", "column", "todo", "current"))
	}
	cards := []reading.Card{{Number: 1, Title: "Ship it", State: "todo",
		Description: strings.Repeat("une ligne de description de plus ", 40)}}
	for i := range 6 {
		cards = append(cards, reading.Card{
			Number: i + 2, Title: fmt.Sprintf("Part %d", i+1), State: "todo", Parent: 1})
	}
	w := New(f, 1200, 800)
	w.snap = reading.Snapshot{Cards: cards}
	w = w.clicked(middleOf(w, 1, 0)).Key("enter")

	box := Over(w.width, w.height)
	at, _ := LongField(w.Shown())
	for _, c := range []struct {
		what  string
		track Box
		read  func(Window) int
	}{
		{"the history", HistoryBar(box), func(w Window) int { return w.History() }},
		{"the description", SayingBar(box, w.Shown(), at), func(w Window) int { return w.Saying() }},
		{"what hangs under it", ServingBar(box, w.Shown()), func(w Window) int { return w.ServingAt() }},
	} {
		middle := Point{X: c.track.X + c.track.W/2, Y: c.track.Y + 4}
		far := Point{X: middle.X, Y: c.track.Y + c.track.H + 999}
		pulled := w.PressAt(middle).MoveTo(far).LetGo(far)
		if c.read(pulled) == c.read(w) {
			t.Errorf("pulling the bar beside %s left it at %d", c.what, c.read(w))
		}
	}
}

// A click on a text of several lines lands on the line it was over. It read
// only how far across the point was, so clicking the third line of a
// description put the caret on the first.
func TestAClickOnAWrappedTextLandsOnItsOwnLine(t *testing.T) {
	f := newFake()
	w := holding(f, card(1, "Write the spec", "todo"))
	w = w.clicked(middleOf(w, 1, 0)).Choose("Edit").On("description")
	w = w.Types(strings.Repeat("une ligne de description assez longue pour envelopper ", 5))

	box := Over(w.width, w.height)
	fields := w.Form().Fields()
	var value Box
	for i, field := range fields {
		if field.Arg == "desc" {
			_, value = FieldRow(box, FieldTop(box, fields, i))
		}
	}
	inside := Inside(Framed(value, longRows))
	first := w.clicked(Point{X: inside.X + 20, Y: inside.Y + lineHigh/2}).Form().Writing().At
	third := w.clicked(Point{X: inside.X + 20, Y: inside.Y + 2*lineHigh + lineHigh/2}).Form().Writing().At
	if third <= first {
		t.Errorf("a click on the third line gave %d and on the first %d", third, first)
	}
}

// The band is walked like the columns. A task that is blocked, watching or
// paused could not be reached by the arrows at all — and those are the tasks
// that want a decision.
func TestTheArrowsReachTheBand(t *testing.T) {
	f := newFake()
	w := holding(f,
		card(1, "Write the spec", "todo"),
		reading.Card{Number: 2, Title: "Waiting on them", State: "blocked"})
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

// Nothing in the window answered the pointer before it was pressed: "Delete"
// looked exactly like a label, and the one affordance that says a thing can be
// clicked was missing from the whole surface.
func TestTheWindowKnowsWhatThePointerIsOver(t *testing.T) {
	w := New(newFake(), 1200, 800).Wonder()
	if _, over := w.Hovered(); over {
		t.Fatal("something is hovered before the pointer has moved")
	}
	l := w.Layout()
	first := OfferAt(l, 0)
	w = Pointed(w, pointer.Move, Point{X: first.X + first.W/2, Y: first.Y + first.H/2}, 0)
	said, over := w.Hovered()
	if !over || said != w.Offers()[0] {
		t.Fatalf("the pointer is over %q %v, want the first offer", said, over)
	}
	w = Pointed(w, pointer.Leave, Point{}, 0)
	if _, over := w.Hovered(); over {
		t.Fatal("the pointer left the window and something is still hovered")
	}
}

// Shrinking the window below what the board needs used to throw away whatever
// was open: a form filled in for ten minutes came back as an empty board.
func TestShrinkingAndGrowingBackKeepsWhatWasOpen(t *testing.T) {
	w := holding(newFake(), reading.Card{Number: 1, Title: "Write the spec", State: "todo"})
	w = w.clicked(middleOf(w, 1, 0)).Choose("Edit").Types("x")
	if w.Screen() != OnForm {
		t.Fatalf("the form did not open: %v", w.Screen())
	}
	small := w.Resized(200, 100)
	if small.Screen() != TooSmallToDraw {
		t.Fatalf("a window of 200 by 100 drew the board: %v", small.Screen())
	}
	back := small.Resized(1200, 800)
	if back.Screen() != OnForm {
		t.Fatalf("growing back left the window on %v", back.Screen())
	}
	if !back.Form().Changed() {
		t.Error("what was typed did not come back with the form")
	}
}

// The list's scrollbar has a thumb sized never to be too small to grab, and
// grabbing it did nothing: only two of the four bars in the window answered a
// drag at all.
func TestTheListsBarCanBePulled(t *testing.T) {
	f := newFake()
	var goals []reading.Goal
	for i := 1; i <= 40; i++ {
		goals = append(goals, reading.Goal{Number: i, Title: "Objective " + strconv.Itoa(i)})
	}
	w := New(f, 1200, 800)
	w.snap = reading.Snapshot{Cards: []reading.Card{{Number: 1, Title: "One", State: "todo"}}, Goals: goals}
	w = w.clicked(middleOf(w, 1, 0)).Choose("Edit")
	for i, field := range w.Form().Fields() {
		if field.Arg == "objective" {
			w.form = w.form.onto(i)
		}
	}
	w = w.Choose("Choose")
	if w.Screen() != OnChoosing {
		t.Fatalf("the list did not open: %v", w.Screen())
	}
	box := ChoosingOver(w.width, w.height)
	track := ChoosingBar(box)
	w = w.PressAt(Point{X: track.X + track.W/2, Y: track.Y + 2}).
		MoveTo(Point{X: track.X + track.W/2, Y: track.Y + track.H - 2})
	if w.ChoosingFromRow() == 0 {
		t.Fatal("pulling the list's bar to the bottom left it at the top")
	}
}

// The same gesture as the terminal's: the column the board is on goes off it in
// one, rather than a card at a time.
func TestAWholeColumnGoesOffTheBoardAtOnce(t *testing.T) {
	f := newFake()
	w := holding(f, reading.Card{Number: 1, Title: "One", State: "done"},
		reading.Card{Number: 2, Title: "Two", State: "done"})
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
	w := holding(f, reading.Card{Number: 1, Title: "One", State: "todo"},
		reading.Card{Number: 2, Title: "Two", State: "todo"})
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
	holding(f, reading.Card{Number: 9, Title: "Last", State: "done"}).
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
		reading.Card{Number: 1, Title: "First", State: "todo"},
		reading.Card{Number: 2, Title: "Second", State: "todo"},
		reading.Card{Number: 3, Title: "Third", State: "todo"})
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
		reading.Card{Number: 1, Title: "In todo", State: "todo"},
		reading.Card{Number: 2, Title: "Blocked", State: "blocked"})
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

// askedFor finds the first thing asked of the store of a given kind.
func askedFor(f *fake, thing, verb string) *engine.Input {
	for _, one := range f.asked {
		if one.op.Thing == thing && one.op.Verb == verb {
			in := one.in
			return &in
		}
	}
	return nil
}

// The band used to be beyond the arrows altogether, and the tasks in it are
// exactly the ones waiting for a decision.
func TestTheArrowsWalkTheBandAndClimbOutOfIt(t *testing.T) {
	f := newFake()
	w := holding(f,
		reading.Card{Number: 1, Title: "In todo", State: "todo"},
		reading.Card{Number: 2, Title: "Blocked one", State: "blocked"},
		reading.Card{Number: 3, Title: "Blocked two", State: "blocked"},
		reading.Card{Number: 4, Title: "Watched", State: "watching"})

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
	var cards []reading.Card
	for i := 1; i <= 20; i++ {
		cards = append(cards, reading.Card{Number: i, Title: "Task " + strconv.Itoa(i), State: "todo"})
	}
	cards = append(cards, reading.Card{Number: 99, Title: "Stuck", State: "blocked"})
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

// Escape closes the search box and widens the board again, by the key as well
// as by the offer: the two are the same act and were written apart.
func TestEscapeOutOfTheSearchByTheKey(t *testing.T) {
	w := holding(newFake(), reading.Card{Number: 1, Title: "Write the spec", State: "todo"})
	w = w.Key("/").Types("spec")
	if w.Screen() != OnSearch || w.Looking() == "" {
		t.Fatalf("the search did not open: %v %q", w.Screen(), w.Looking())
	}
	w = w.Key("esc")
	if w.Screen() != OnTheBoard {
		t.Fatalf("escape left the window on %v", w.Screen())
	}
	if w.Looking() != "" {
		t.Fatalf("escape left the board narrowed to %q", w.Looking())
	}
}

// Lighting is what the drawing asks before it lights an offer, so what the
// pointer is resting on is a question a scenario can put rather than a colour
// it has to count pixels for.
func TestOnlyTheOfferUnderThePointerIsLit(t *testing.T) {
	w := New(newFake(), 1200, 800).Wonder()
	offers := w.Offers()
	for _, offer := range offers {
		if w.Lighting(offer) {
			t.Fatalf("%q is lit before the pointer has moved", offer)
		}
	}
	first := OfferAt(w.Layout(), 0)
	w = w.Hovering(Point{X: first.X + first.W/2, Y: first.Y + first.H/2})
	if !w.Lighting(offers[0]) {
		t.Fatalf("%q is not lit with the pointer on it", offers[0])
	}
	if w.Lighting(offers[1]) {
		t.Fatalf("%q is lit with the pointer elsewhere", offers[1])
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
			w := holding(f, reading.Card{Number: 1, Title: "One", State: "todo"},
				reading.Card{Number: 2, Title: "Two", State: "todo"})
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
	w := holding(f, reading.Card{Number: 1, Title: "One", State: "backlog"},
		reading.Card{Number: 2, Title: "Stuck", State: "blocked"})
	w = w.clicked(middleOf(w, 0, 0))
	if got := w.Key("down").Chosen(); got != 1 {
		t.Fatalf("down out of the backlog chose %d, want to stay on 1", got)
	}
}

// Climbing out of a lane whose column above is empty stays where it is, rather
// than reaching for a card that is not there.
func TestClimbingOutOfALaneWithNothingAboveItStays(t *testing.T) {
	f := newFake()
	w := holding(f, reading.Card{Number: 1, Title: "Stuck", State: "blocked"})
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
	w := holding(f, reading.Card{Number: 1, Title: "Stuck", State: "blocked"},
		reading.Card{Number: 2, Title: "Watched", State: "watching"})
	w = w.clicked(onLaneCard(w, 1, 0))
	if w.Chosen() != 2 {
		t.Fatalf("the watched card was not chosen: %d", w.Chosen())
	}
	if got := w.Key("left").Chosen(); got != 1 {
		t.Fatalf("left out of the watching lane chose %d, want the blocked one", got)
	}
}

// onLaneCard is the point a person would click to hit the n-th card of a lane.
func onLaneCard(w Window, lane, at int) Point {
	box := w.Layout().LaneCardAt(lane, at, 0)
	return Point{X: box.X + box.W/2, Y: box.Y + box.H/2}
}

// The bar beside the work under a task can be pulled, and pulling it to the
// bottom brings the last tile into the box rather than somewhere near it.
func TestPullingTheBarUnderATaskReachesTheLastTile(t *testing.T) {
	f := newFake()
	cards := []reading.Card{{Number: 1, Title: "Ship it", State: "todo"}}
	for i := 2; i <= 30; i++ {
		cards = append(cards, reading.Card{
			Number: i, Title: "Part " + strconv.Itoa(i), State: "todo", Parent: 1})
	}
	w := holding(f, cards...)
	w = w.clicked(middleOf(w, 1, 0)).Choose("Edit").Cancel()
	w = w.DoubleClick(middleOf(w, 1, 0))
	if w.Screen() != OnDetail {
		t.Fatalf("the task did not open: %v", w.Screen())
	}
	last := len(w.Under()) - 1
	track := ServingBar(Over(w.width, w.height), w.Shown())
	w = w.PressAt(Point{X: track.X + track.W/2, Y: track.Y + 1}).
		MoveTo(Point{X: track.X + track.W/2, Y: track.Y + track.H - 1})
	box := ServingBox(Over(w.width, w.height), w.Shown())
	tile := w.ServingCardAt(last)
	if tile.Y+tile.H > box.Y+box.H || tile.Y < box.Y {
		t.Fatalf("the last tile sits at %d..%d, outside the box %d..%d",
			tile.Y, tile.Y+tile.H, box.Y, box.Y+box.H)
	}
}

// Scrolling a description to the end shows its last line. It did not: the end
// of the scroll was worked out from the text laid out on one width and the text
// was drawn on another, a few points narrower — so what was drawn ran longer
// than what was counted, and the last lines of a long description could never
// be brought into the box, bar or no bar.
func TestScrollingADescriptionToTheEndShowsItsLastLine(t *testing.T) {
	text := strings.Repeat("une phrase de description assez longue pour replier ", 60) + "LA FIN"
	w := holding(newFake(), reading.Card{Number: 1, Title: "Write the spec", State: "todo",
		Description: text})
	w = w.clicked(middleOf(w, 1, 0))
	w = w.DoubleClick(middleOf(w, 1, 0))
	if w.Screen() != OnDetail {
		t.Fatalf("the task did not open: %v", w.Screen())
	}
	frame, _, lines, rows, ok := w.sayingOn()
	if !ok {
		t.Fatal("an open task with a description has none on screen")
	}
	drawn := Wrapped(text, room(Inside(frame))).Said
	if len(drawn) != len(lines.Said) {
		t.Fatalf("the description is counted as %d lines and drawn as %d", len(lines.Said), len(drawn))
	}
	w.saying = w.settledSaying(1 << 20)
	if last := drawn[min(len(drawn)-1, w.saying+rows-1)]; !strings.HasSuffix(last, "LA FIN") {
		t.Fatalf("scrolled to the end, the last line in sight is %q", last)
	}
}

// ^D and ^U walk the description on an open task, half its box at a time —
// the terminal board's keys for it, taken here too so the same fingers work.
func TestHalfABoxAtATimeOnTheDescription(t *testing.T) {
	text := strings.Repeat("une phrase de description assez longue pour replier ", 60) + "LA FIN"
	w := holding(newFake(), reading.Card{Number: 1, Title: "Write the spec", State: "todo",
		Description: text})
	w = w.clicked(middleOf(w, 1, 0)).DoubleClick(middleOf(w, 1, 0))
	_, _, lines, rows, _ := w.sayingOn()
	if w = w.Key("cmd+d"); w.saying != rows/2 {
		t.Fatalf("^D moved the description to %d, want %d", w.saying, rows/2)
	}
	for range len(lines.Said) {
		w = w.Key("cmd+d")
	}
	if w.saying != len(lines.Said)-rows {
		t.Fatalf("^D to the end left the description at %d, want %d", w.saying, len(lines.Said)-rows)
	}
	for range len(lines.Said) {
		w = w.Key("cmd+u")
	}
	if w.saying != 0 {
		t.Fatalf("^U to the top left the description at %d", w.saying)
	}
}

// A task opens on the top of its description, whatever was scrolled before it.
// The scroll was left where the last task or the form had put it, and drawn as
// it stood: a task opened after a long description had been read to its end
// showed an empty box until the wheel brought the scroll back inside the text.
func TestATaskOpensOnTheTopOfItsDescription(t *testing.T) {
	long := strings.Repeat("une phrase de description assez longue pour replier ", 60)
	w := holding(newFake(),
		reading.Card{Number: 1, Title: "Long", State: "todo", Description: long},
		reading.Card{Number: 2, Title: "Short", State: "todo", Description: "PREMIERE LIGNE"})
	// Read the long one to its end, and put it away.
	w = w.clicked(middleOf(w, 1, 0)).DoubleClick(middleOf(w, 1, 0))
	_, _, lines, _, _ := w.sayingOn()
	for range len(lines.Said) {
		w = w.Key("cmd+d")
	}
	if w.Saying() == 0 {
		t.Fatal("this proves nothing: the long description was not scrolled")
	}
	w = w.Cancel()
	// The short one opens on its first line.
	w = w.clicked(middleOf(w, 1, 1)).DoubleClick(middleOf(w, 1, 1))
	if w.Screen() != OnDetail || w.Chosen() != 2 {
		t.Fatalf("the short task did not open: %v on %d", w.Screen(), w.Chosen())
	}
	if got := w.Saying(); got != 0 {
		t.Fatalf("the short task opened with its description scrolled to %d", got)
	}
	// And it is the same coming from the form.
	w = w.Cancel().clicked(middleOf(w, 1, 0)).Choose("Edit")
	for i, field := range w.Form().Fields() {
		if field.Name == "description" {
			w.form = w.form.onto(i)
		}
	}
	w = w.inSightOfTheCaret().Cancel()
	w = w.clicked(middleOf(w, 1, 1)).DoubleClick(middleOf(w, 1, 1))
	if got := w.Saying(); got != 0 {
		t.Fatalf("a task opened after the form showed its description from %d", got)
	}
}

// Whatever the scroll holds, what is drawn is inside the text: a scroll past
// the end is read as the end.
func TestAScrollPastTheEndIsReadAsTheEnd(t *testing.T) {
	w := holding(newFake(), reading.Card{Number: 1, Title: "Short", State: "todo",
		Description: "PREMIERE LIGNE"})
	w = w.clicked(middleOf(w, 1, 0)).DoubleClick(middleOf(w, 1, 0))
	w.saying = 40
	if got := w.Saying(); got != 0 {
		t.Fatalf("a scroll of 40 on a description of one line reads as %d", got)
	}
}

// The box under an open task says how many of its parts are left and how many
// are done, beside what it is.
func TestTheBoxUnderATaskCountsWhatIsLeft(t *testing.T) {
	w := holding(newFake(),
		reading.Card{Number: 1, Title: "Ship it", State: "todo"},
		reading.Card{Number: 2, Title: "A", State: "done", Parent: 1},
		reading.Card{Number: 3, Title: "B", State: "todo", Parent: 1},
		reading.Card{Number: 4, Title: "C", State: "todo", Parent: 1})
	w = w.clicked(middleOf(w, 1, 0)).DoubleClick(middleOf(w, 1, 0))
	if got := w.UnderHead(); !strings.Contains(got, "2 to do · 1 done") {
		t.Fatalf("the box under the task is headed %q", got)
	}
	if got := w.Under()[0].Number; got != 3 {
		t.Fatalf("the first part shown is %d, want the first one left, 3", got)
	}
}

// Every key the bar teaches is in the help too. The two are written apart —
// Shortcut says which key an offer has, Helping says what each key does — so
// nothing but this keeps the second from forgetting one the first added.
func TestTheHelpNamesEveryKeyTheBarTeaches(t *testing.T) {
	var said strings.Builder
	for _, f := range holding(newFake()).Helping() {
		said.WriteString(f.Key + " ")
	}
	for _, offer := range []string{"new", "search", "help", "archived", "archive", "take", "keep",
		"edit", "delete", "agree", "save", "done", "close"} {
		key := Pretty(Shortcut(offer))
		if !strings.Contains(said.String(), key) {
			t.Errorf("%s is %q on the bar and the help never names it", offer, key)
		}
	}
}
