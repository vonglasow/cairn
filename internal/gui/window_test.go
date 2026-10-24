package gui

import (
	"fmt"

	"strconv"
	"strings"
	"testing"

	"gioui.org/io/pointer"

	"cairn/internal/engine"
)

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
	if w.Screen() != OnForm {
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

func TestTheWheelScrollsOneColumnAndNoOther(t *testing.T) {
	f := newFake()
	var many []engine.Card
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

func TestAReleaseWhereNoCardWasHeldIsStillAClick(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "Write the spec", State: "todo"})
	tab := TabAt(w.Layout(), 1)
	at := Point{X: tab.X + tab.W/2, Y: tab.Y + tab.H/2}

	w = w.PressAt(at)
	w = w.LetGo(at)
	if w.Tab() != Objectives {
		t.Fatalf("pressing and releasing on a tab left the window on %v", w.Tab())
	}
}

func TestTheBarOffersEditingAndClickingItOpensTheForm(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "Write the spec", State: "todo"})
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
	box := w.BarOffers()[at]
	w = w.clicked(Point{X: box.X + box.W/2, Y: box.Y + box.H/2})
	if w.Screen() != OnForm {
		t.Fatalf("clicking Edit left the window on %v", w.Screen())
	}
}

func TestTheBarOffersNothingToEditWithNothingChosen(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "Write the spec", State: "todo"})
	for _, offer := range w.Offers() {
		if offer == "Edit" {
			t.Error("the bar offers to edit with nothing chosen")
		}
	}
}

func TestTheObjectivesTabListsChoosesAndEdits(t *testing.T) {
	f := newFake()
	w := New(f, 1200, 800)
	w.snap = engine.Snapshot{Goals: []engine.Goal{
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

func TestClickingBelowTheObjectivesChoosesNothing(t *testing.T) {
	f := newFake()
	w := New(f, 1200, 800)
	w.snap = engine.Snapshot{Goals: []engine.Goal{{Number: 3, Title: "Ship cairn", State: "active"}}}
	w.tab = Objectives

	below := GoalAt(w.Layout(), 4, 0)
	if got := w.clicked(Point{X: below.X + below.W/2, Y: below.Y + below.H/2}); got.Chosen() != 0 {
		t.Errorf("clicking below the list chose %d", got.Chosen())
	}
}

func TestTheDetailShowsEveryFieldOfTheTask(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{
		Number: 1, Title: "Write the spec", State: "todo", Priority: "high",
		Description: "The client wants two columns",
	})

	w = w.doubleClicked(middleOf(w, 1, 0))
	shown := w.Form().Fields()
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

func TestTheDetailShowsNothingWithNoTaskOpen(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "Write the spec", State: "todo"})
	if got := w.Form().Fields(); len(got) != 0 {
		t.Errorf("a board with no detail open shows %d fields", len(got))
	}
}

func TestDoubleClickingAnObjectiveOpensThatObjective(t *testing.T) {
	f := newFake()
	w := New(f, 1200, 800)
	w.snap = engine.Snapshot{
		Cards: []engine.Card{{Number: 1, Title: "Write the spec", State: "backlog"}},
		Goals: []engine.Goal{
			{Number: 3, Title: "Ship cairn", State: "active"},
			{Number: 4, Title: "Keep it honest", State: "active"},
		},
	}
	w.tab = Objectives

	for at, want := range map[int]int{0: 3, 1: 4} {
		box := GoalAt(w.Layout(), at, 0)
		got := w.doubleClicked(Point{X: box.X + box.W/2, Y: box.Y + box.H/2})
		if got.Screen() != OnForm {
			t.Fatalf("double-clicking objective %d left the window on %v", at, got.Screen())
		}
		if got.Chosen() != want {
			t.Errorf("double-clicking objective %d opened %d, want %d", at, got.Chosen(), want)
		}
	}
}

func TestAPanelOverTheBoardHasTheWindow(t *testing.T) {
	f := newFake()
	w := New(f, 1200, 800)
	w.snap = engine.Snapshot{Goals: []engine.Goal{
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

	for i, offer := range w.Offers() {
		if offer != "Cancel" {
			continue
		}
		box := PanelOfferAt(w.panel(), i)
		w = w.clicked(Point{X: box.X + box.W/2, Y: box.Y + box.H/2})
	}
	if w.Screen() != OnTheBoard {
		t.Fatalf("Cancel left the window on %v", w.Screen())
	}
}

func TestNoCardIsPickedUpFromBehindAPanel(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "Write the spec", State: "todo"})
	w = w.doubleClicked(middleOf(w, 1, 0))

	w = w.PressAt(middleOf(w, 1, 0))
	w = w.MoveTo(Point{X: 20, Y: 400})
	if w.Dragging() {
		t.Error("a card was dragged out from behind an open detail")
	}
}

func TestTheDetailOfAnObjectiveShowsWhatItCarries(t *testing.T) {
	f := newFake()
	w := New(f, 1200, 800)
	w.snap = engine.Snapshot{Goals: []engine.Goal{{Number: 3, Title: "Ship cairn", State: "active"}}}
	w.tab = Objectives

	box := GoalAt(w.Layout(), 0, 0)
	w = w.doubleClicked(Point{X: box.X + box.W/2, Y: box.Y + box.H/2})

	shown := w.Form().Fields()
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

func TestATaskAndAnObjectiveOfTheSameNumberAreToldApart(t *testing.T) {
	f := newFake()
	w := New(f, 1200, 800)
	w.snap = engine.Snapshot{
		Cards: []engine.Card{{Number: 1, Title: "Write the spec", State: "backlog"}},
		Goals: []engine.Goal{{Number: 1, Title: "Ship cairn", State: "active"}},
	}

	w.tab = Objectives
	box := GoalAt(w.Layout(), 0, 0)
	goal := w.doubleClicked(Point{X: box.X + box.W/2, Y: box.Y + box.H/2})
	if got := goal.Form().Fields(); len(got) == 0 || got[0].Value != "Ship cairn" {
		t.Errorf("the detail of objective 1 shows %v", got)
	}
	if goal.Edit().Form().ForObjective != true {
		t.Error("editing objective 1 opened a task's form")
	}

	w.tab = Tasks
	task := w.doubleClicked(middleOf(w, 0, 0))
	if got := task.Form().Fields(); len(got) == 0 || got[0].Value != "Write the spec" {
		t.Errorf("the detail of task 1 shows %v", got)
	}
	if task.Edit().Form().ForObjective {
		t.Error("editing task 1 opened an objective's form")
	}
}

func TestDeletingAsksBeforeItLosesAnything(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "Write the spec", State: "todo"})
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

func TestDeletingAnObjective(t *testing.T) {
	f := newFake()
	w := New(f, 1200, 800)
	w.snap = engine.Snapshot{Goals: []engine.Goal{{Number: 3, Title: "Ship cairn", State: "active"}}}
	w.tab, w.chosen, w.chosenGoal = Objectives, 3, true

	w = w.Choose("Delete").Choose("Delete")
	if f.what("objective", "delete") != 1 {
		t.Errorf("it asked for %d objective deletes", f.what("objective", "delete"))
	}
}

func TestDeletingIsNotOfferedWithNothingChosen(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "Write the spec", State: "todo"})
	for _, offer := range w.Offers() {
		if offer == "Delete" {
			t.Error("the bar offers to delete with nothing chosen")
		}
	}
}

func TestTheHelpSaysWhatEverythingDoes(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "Write the spec", State: "todo"})

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

func TestWhatEachPanelSaysItIs(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 7, Title: "Write the spec", State: "todo"})
	if got := Heads(w); got != "" {
		t.Errorf("the board is headed %q, and it is its own heading", got)
	}
	on := w.clicked(middleOf(w, 1, 0))
	if got := Heads(on.Key("enter").Types("x")); got != "task 7 · Write the spec" {
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

func TestThePagesAreWalkedRound(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "Write the spec", State: "todo"})
	for _, c := range []struct {
		keys []string
		want Tab
	}{
		{[]string{"tab"}, Objectives},
		{[]string{"tab", "tab"}, Time},
		{[]string{"tab", "tab", "tab"}, Tasks},
		{[]string{"shift+tab"}, Time},
		{[]string{"shift+tab", "shift+tab"}, Objectives},
		{[]string{"tab", "shift+tab"}, Tasks},
	} {
		got := w
		for _, key := range c.keys {
			got = got.Key(key)
		}
		if got.Tab() != c.want {
			t.Errorf("%v from the first page lands on %v, want %v", c.keys, got.Tab(), c.want)
		}
	}
	if got := w.clicked(middleOf(w, 1, 0)).other(1); got.Chosen() != 0 {
		t.Errorf("after changing page, task %d is still chosen", got.Chosen())
	}
}

func TestWhereWhatIsChosenIsDrawn(t *testing.T) {
	f := newFake()
	w := holding(f,
		engine.Card{Number: 1, Title: "First", State: "todo"},
		engine.Card{Number: 2, Title: "Second", State: "todo"})
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
	if _, _, ok := w.WhereChosen(); ok {
		t.Error("with nothing chosen, something was said to be drawn")
	}
}

func TestTheWheelMovesAHistoryByWholeLines(t *testing.T) {
	f := newFake()
	for range 40 {
		f.acts = append(f.acts, anAct("moved", "column", "todo", "current"))
	}
	w := holding(f, engine.Card{Number: 1, Title: "Write the spec", State: "todo"})
	w = w.clicked(middleOf(w, 1, 0)).Key("enter")
	over := BesideBox(Over(w.width, w.height))
	at := Point{X: over.X + over.W/2, Y: over.Y + over.H/2}

	if got := w.Scroll(at, 3*lineHigh).History(); got != 3 {
		t.Errorf("turned by three lines, the history is at %d", got)
	}
	if got := w.Scroll(at, lineHigh-1).History(); got != 0 {
		t.Errorf("turned by less than a line, the history is at %d", got)
	}
	if got := w.Scroll(Point{X: 1, Y: 1}, 9*lineHigh).History(); got != 0 {
		t.Errorf("turned away from the history, it is at %d", got)
	}
}

func TestHowFarTheObjectivesScroll(t *testing.T) {
	f := newFake()
	w := New(f, 1200, 800)
	for i := 1; i <= 60; i++ {
		w.snap.Goals = append(w.snap.Goals, engine.Goal{Number: i, Title: "Ship it", State: "live"})
	}
	w.tab = Objectives
	run := GoalsRun(w.Layout())
	rows := GoalRows(len(w.snap.Goals))

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

	if got := w.followingGoal(1).goals; got != 0 {
		t.Errorf("following the second objective moved the grid to %d", got)
	}
	last := len(w.snap.Goals) - 1
	if got, want := w.followingGoal(last).goals, end; got != want {
		t.Errorf("following the last objective moved the grid to %d, want %d", got, want)
	}
	fits := run.H / CardHeight
	if got, want := w.followingGoal(fits*GoalsAcross()).goals, (fits+1)*CardHeight-run.H; got != want {
		t.Errorf("following the first objective past what fits moved to %d, want %d", got, want)
	}
}

func TestWhereAChosenObjectiveIsDrawn(t *testing.T) {
	f := newFake()
	w := New(f, 1200, 800)
	w.snap.Goals = []engine.Goal{
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
	for i, g := range engine.Sorted(w.snap.Goals) {
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
	w.chosen = 99
	if _, _, ok := w.WhereChosen(); ok {
		t.Error("an objective that is not there was said to be drawn somewhere")
	}
}

func TestAPanelUnderAnotherKeepsItsOwnHeading(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 7, Title: "Write the spec", State: "todo"})
	open := onTheObjectives(w)
	if open.Screen() != OnChoosing {
		t.Fatalf("the list did not open: %v", open.Screen())
	}
	if got := Heads(open); got != "objectives for task 7" {
		t.Errorf("the list is headed %q", got)
	}
	if got := HeadsOn(open, OnForm); got != "task 7 · Write the spec" {
		t.Errorf("the form under it is headed %q, not its own", got)
	}
}

func TestTheWheelStopsAtAPanel(t *testing.T) {
	f := newFake()
	cards := []engine.Card{}
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

func TestOnlyTheFirstButtonActs(t *testing.T) {
	if Acts(pointer.ButtonSecondary) || Acts(pointer.ButtonTertiary) {
		t.Error("a button other than the first was allowed to act")
	}
	if !Acts(pointer.ButtonPrimary) {
		t.Error("the first button was not allowed to act")
	}
	if !Acts(0) {
		t.Error("an event with no button was refused")
	}
}

func TestAnOfferAnswersWhereItIsDrawn(t *testing.T) {
	f := newFake()
	w := New(f, 1200, 800)
	w.snap = engine.Snapshot{
		Cards: []engine.Card{card(1, "Write the spec", "todo")},
		Goals: []engine.Goal{{Number: 1, Title: "Ship it", State: "active"}},
	}
	w = onTheObjectives(w)
	if w.Screen() != OnChoosing {
		t.Fatalf("the list did not open: %v", w.Screen())
	}

	box := w.ChoosingBox()
	for i, offer := range w.Offers() {
		at := PanelOfferAt(box, i)
		got, on := w.offerAt(Point{X: at.X + 10, Y: at.Y + at.H/2})
		if !on || got != offer {
			t.Errorf("the word %q drawn at %+v answers %q, %v", offer, at, got, on)
		}
	}
}

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
		value := formValue(box, fields, i)
		got := w.clicked(Point{X: value.X + 10, Y: value.Y + value.H/2})
		if got.Screen() != OnChoosing || len(got.Choosing()) != len(field.Choices) {
			t.Errorf("clicking the column field left the window on %v, offering %d",
				got.Screen(), len(got.Choosing()))
		}
		return
	}
	t.Fatal("the form has no column field")
}

func TestEveryBarCanBePulled(t *testing.T) {
	f := newFake()
	for range 40 {
		f.acts = append(f.acts, anAct("moved", "column", "todo", "current"))
	}
	cards := []engine.Card{{Number: 1, Title: "Ship it", State: "todo",
		Description: strings.Repeat("une ligne de description de plus ", 40)}}
	for i := range 6 {
		cards[0].Steps = append(cards[0].Steps, engine.Part{Number: i + 1, Title: fmt.Sprintf("Step %d", i+1)})
	}
	w := New(f, 1200, 800)
	w.snap = engine.Snapshot{Cards: cards}
	w = w.clicked(middleOf(w, 1, 0)).Key("enter")

	box := Over(w.width, w.height)
	_, saying, _, _, _ := w.sayingOn()
	for _, c := range []struct {
		what  string
		track Box
		read  func(Window) int
	}{
		{"the history", BesideBar(box), func(w Window) int { return w.History() }},
		{"the description", saying, func(w Window) int { return w.Saying() }},
	} {
		middle := Point{X: c.track.X + c.track.W/2, Y: c.track.Y + 4}
		far := Point{X: middle.X, Y: c.track.Y + c.track.H + 999}
		pulled := w.PressAt(middle).MoveTo(far).LetGo(far)
		if c.read(pulled) == c.read(w) {
			t.Errorf("pulling the bar beside %s left it at %d", c.what, c.read(w))
		}
	}
}

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
			value = formValue(box, fields, i)
		}
	}
	inside := Inside(Framed(value, longRows))
	first := w.clicked(Point{X: inside.X + 20, Y: inside.Y + lineHigh/2}).Form().Writing().At
	third := w.clicked(Point{X: inside.X + 20, Y: inside.Y + 2*lineHigh + lineHigh/2}).Form().Writing().At
	if third <= first {
		t.Errorf("a click on the third line gave %d and on the first %d", third, first)
	}
}

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

func TestShrinkingAndGrowingBackKeepsWhatWasOpen(t *testing.T) {
	w := holding(newFake(), engine.Card{Number: 1, Title: "Write the spec", State: "todo"})
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

func TestTheListsBarCanBePulled(t *testing.T) {
	f := newFake()
	var goals []engine.Goal
	for i := 1; i <= 40; i++ {
		goals = append(goals, engine.Goal{Number: i, Title: "Objective " + strconv.Itoa(i)})
	}
	w := New(f, 1200, 800)
	w.snap = engine.Snapshot{Cards: []engine.Card{{Number: 1, Title: "One", State: "todo"}}, Goals: goals}
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
	box := w.ChoosingBox()
	track := ChoosingBar(box)
	w = w.PressAt(Point{X: track.X + track.W/2, Y: track.Y + 2}).
		MoveTo(Point{X: track.X + track.W/2, Y: track.Y + track.H - 2})
	if w.ChoosingFromRow() == 0 {
		t.Fatal("pulling the list's bar to the bottom left it at the top")
	}
}

func askedFor(f *fake, thing, verb string) *engine.Input {
	for _, one := range f.asked {
		if one.op.Thing == thing && one.op.Verb == verb {
			in := one.in
			return &in
		}
	}
	return nil
}

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

func onLaneCard(w Window, lane, at int) Point {
	box := w.Layout().LaneCardAt(lane, at, 0)
	return Point{X: box.X + box.W/2, Y: box.Y + box.H/2}
}

func TestScrollingADescriptionToTheEndShowsItsLastLine(t *testing.T) {
	text := strings.Repeat("une phrase de description assez longue pour replier ", 60) + "LA FIN"
	w := holding(newFake(), engine.Card{Number: 1, Title: "Write the spec", State: "todo",
		Description: text})
	w = w.clicked(middleOf(w, 1, 0))
	w = w.DoubleClick(middleOf(w, 1, 0))
	if w.Screen() != OnForm {
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

func TestATaskOpensOnTheTopOfItsDescription(t *testing.T) {
	long := strings.Repeat("une phrase de description assez longue pour replier ", 60)
	w := holding(newFake(),
		engine.Card{Number: 1, Title: "Long", State: "todo", Description: long},
		engine.Card{Number: 2, Title: "Short", State: "todo", Description: "PREMIERE LIGNE"})
	w = w.clicked(middleOf(w, 1, 0)).DoubleClick(middleOf(w, 1, 0))
	frame, _, lines, _, _ := w.sayingOn()
	for range len(lines.Said) {
		w = w.Scroll(Point{X: frame.X + 4, Y: frame.Y + 4}, lineHigh)
	}
	if w.Saying() == 0 {
		t.Fatal("this proves nothing: the long description was not scrolled")
	}
	w = w.Cancel()
	w = w.clicked(middleOf(w, 1, 1)).DoubleClick(middleOf(w, 1, 1))
	if w.Screen() != OnForm || w.Chosen() != 2 {
		t.Fatalf("the short task did not open: %v on %d", w.Screen(), w.Chosen())
	}
	if got := w.Saying(); got != 0 {
		t.Fatalf("the short task opened with its description scrolled to %d", got)
	}
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

func TestAScrollPastTheEndIsReadAsTheEnd(t *testing.T) {
	w := holding(newFake(), engine.Card{Number: 1, Title: "Short", State: "todo",
		Description: "PREMIERE LIGNE"})
	w = w.clicked(middleOf(w, 1, 0)).DoubleClick(middleOf(w, 1, 0))
	w.saying = 40
	if got := w.Saying(); got != 0 {
		t.Fatalf("a scroll of 40 on a description of one line reads as %d", got)
	}
}

func TestTheHelpNamesEveryKeyTheBarTeaches(t *testing.T) {
	for _, system := range []string{"darwin", "linux", "windows"} {
		var said strings.Builder
		w := holding(newFake())
		w.keyboard = KeyboardOf(system)
		for _, f := range w.Helping() {
			said.WriteString(f.Key + " ")
		}
		for _, offer := range []string{"new", "search", "help", "archived", "archive", "take", "keep",
			"edit", "delete", "agree", "save", "done", "close"} {
			key := w.keyboard.Spelled(Pretty(Shortcut(offer)))
			if !strings.Contains(said.String(), key) {
				t.Errorf("on %s, %s is %q on the bar and the help never names it", system, offer, key)
			}
		}
	}
}
