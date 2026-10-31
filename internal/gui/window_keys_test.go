package gui

import (
	"strconv"
	"strings"
	"testing"

	"gioui.org/io/key"

	"cairn/internal/engine"
	"cairn/internal/migration"
)

type behind struct {
	*fake
	waiting []string
	current *bool
}

func (b behind) Ask(op engine.Operation, in engine.Input) (engine.Answer, error) {
	if b.current != nil && op.Thing == engine.MigrationApply.Thing && op.Verb == engine.MigrationApply.Verb {
		_, _ = b.fake.Ask(op, in)
		*b.current = true
		return engine.Answer{}, nil
	}
	if b.current != nil && *b.current {
		return b.fake.Ask(op, in)
	}
	if op.Thing == engine.MigrationList.Thing && op.Verb == engine.MigrationList.Verb {
		var changes []migration.Migration
		for i, name := range b.waiting {
			changes = append(changes, migration.Made(migration.Carried{Number: 3 + i, Name: name}))
		}
		return engine.Answer{Changes: changes}, nil
	}
	return engine.Answer{}, engine.Told(engine.ErrStoreBehind)
}

func TestTheWindowAsksToApplyWhatIsWaitingWhenItsLookIsRefused(t *testing.T) {
	w := Window{work: behind{fake: newFake(), waiting: []string{"add-the-estimate"}}}.Look()
	if w.Screen() != OnAsking || w.question != aboutApplying {
		t.Fatalf("the window is on %v asking %v", w.Screen(), w.question)
	}
	if w.agreeing.Op.Verb != engine.MigrationApply.Verb || w.agreeing.Op.Thing != engine.MigrationApply.Thing {
		t.Errorf("the window holds %+v", w.agreeing)
	}
	if w.Wrong() != nil {
		t.Errorf("the window says it went wrong: %v", w.Wrong())
	}
	if head := Heads(w) + " " + sentences(w.Asks()); !strings.Contains(head, "add-the-estimate") {
		t.Errorf("the question reads %q, naming no change", head)
	}
	for range 2 * Lingers() {
		w = w.aged()
	}
	if w.Screen() != OnAsking || w.question != aboutApplying {
		t.Errorf("after the beats the window is on %v asking %v", w.Screen(), w.question)
	}
}

func TestTheWindowNamesEveryChangeWaitingInOneSentence(t *testing.T) {
	w := Window{work: behind{fake: newFake(), waiting: []string{"add-the-estimate", "split-into-steps"}}}.Look()
	if head := Heads(w); head != "Apply what this store is waiting for?" {
		t.Errorf("the question is headed %q", head)
	}
	if said := sentences(w.Asks()); !strings.HasPrefix(said, "Waiting: add-the-estimate, split-into-steps.") {
		t.Errorf("the question reads %q", said)
	}
}

func TestTheWindowLeavesWhenQuitIsChosen(t *testing.T) {
	var current bool
	asked := behind{newFake(), []string{"add-the-estimate"}, &current}
	w := Window{work: asked}.Look()
	if w.Leaving() {
		t.Fatal("the window is leaving before anything was chosen")
	}
	for name, left := range map[string]Window{"chosen": w.Choose("Quit"), "pressed": w.Key(Shortcut("Quit"))} {
		if !left.Leaving() {
			t.Errorf("quit %s and the window is not leaving", name)
		}
		if current {
			t.Errorf("quit %s and the change was applied", name)
		}
	}
}

func sentences(said []Said) string {
	var out []string
	for _, s := range said {
		out = append(out, s.Means)
	}
	return strings.Join(out, " ")
}

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
	if got := w.Key("right").Key("right").Key("right").Chosen(); got != 2 {
		t.Errorf("walking right across two empty columns chose %d", got)
	}
}

func TestAnArrowWithNothingChosenChoosesSomewhereToBegin(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "Only", State: "backlog"})
	if got := w.Key("down").Chosen(); got != 1 {
		t.Errorf("an arrow on an untouched board chose %d", got)
	}
}

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

func TestAnArrowBeginsInTheFirstColumnThatHasAnything(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 7, Title: "Only", State: "current"})
	if got := w.Key("down").Chosen(); got != 7 {
		t.Errorf("an arrow chose %d, and the only card is 7", got)
	}
}

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

func TestShiftAndAnArrowMovesTheChosenCard(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "One", State: "todo"},
		engine.Card{Number: 2, Title: "Two", State: "todo"})
	w = w.clicked(middleOf(w, 1, 0))

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
	f.asked = nil
	holding(f, engine.Card{Number: 9, Title: "Last", State: "done"}).
		clicked(middleOf(w, 3, 0)).Key("shift+right")
	if askedFor(f, "task", "edit") != nil {
		t.Fatalf("shift+right off the last column asked %v", f.asked)
	}

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

func TestTheArrowsWalkTheBandAndClimbOutOfIt(t *testing.T) {
	f := newFake()
	w := holding(f,
		engine.Card{Number: 1, Title: "In todo", State: "todo"},
		engine.Card{Number: 2, Title: "Blocked one", State: "blocked"},
		engine.Card{Number: 3, Title: "Blocked two", State: "blocked"},
		engine.Card{Number: 4, Title: "Watched", State: "watching"})

	w = w.clicked(middleOf(w, 1, 0)).Key("down")
	if w.Chosen() != 2 {
		t.Fatalf("down out of TODO chose %d, want the first blocked task", w.Chosen())
	}
	if got := w.Key("down").Chosen(); got != 3 {
		t.Fatalf("down in a lane chose %d, want 3", got)
	}
	if got := w.Key("down").Key("down").Chosen(); got != 3 {
		t.Fatalf("down past the end of a lane chose %d, want 3", got)
	}
	if got := w.Key("right").Chosen(); got != 4 {
		t.Fatalf("right out of the blocked lane chose %d, want 4", got)
	}
	if got := w.Key("left").Chosen(); got != 2 {
		t.Fatalf("left out of the first lane chose %d, want to stay", got)
	}
	if got := w.Key("right").Key("right").Chosen(); got != 4 {
		t.Fatalf("right off the last lane chose %d, want to stay on 4", got)
	}
	if got := w.Key("down").Key("right").Chosen(); got != 4 {
		t.Fatalf("right from the second blocked card chose %d, want 4", got)
	}
	if got := w.clicked(middleOf(w, 1, 0)).Key("right").Chosen(); got != 1 {
		t.Fatalf("right out of TODO chose %d, and should not have fallen into the band", got)
	}
	if got := w.Key("up").Chosen(); got != 1 {
		t.Fatalf("up out of a lane chose %d, want the task above it", got)
	}
}

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

func TestWalkingOffTheBottomOfTheBacklogGoesNowhere(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "One", State: "backlog"},
		engine.Card{Number: 2, Title: "Stuck", State: "blocked"})
	w = w.clicked(middleOf(w, 0, 0))
	if got := w.Key("down").Chosen(); got != 1 {
		t.Fatalf("down out of the backlog chose %d, want to stay on 1", got)
	}
}

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

func TestTheWindowKeepsTheEnginesAgreeingAfterARefusal(t *testing.T) {
	f := newFake()
	f.failOp = engine.ErrStepStillOpen
	w := Window{work: f}
	w = w.asking(engine.TaskMove, engine.Input{Given: map[string]string{"number": "4"}})
	if w.Screen() != OnAsking || w.question != aboutFinishingParts {
		t.Fatalf("the window is on %v asking %v", w.Screen(), w.question)
	}
	if w.agreeing.Reason != "step-still-open" || !w.agreeing.In.Confirmed || w.agreeing.In.Given["number"] != "4" {
		t.Errorf("the window holds %+v", w.agreeing)
	}
	f.failOp = nil
	w.agreed()
	if in, ok := f.gave("task", "move"); !ok || !in.Confirmed || in.Given["number"] != "4" || f.what("task", "move") != 2 {
		t.Errorf("agreeing asked %+v after %d moves", in, f.what("task", "move"))
	}
}

func TestTheWindowAppliesWhatIsWaitingOnceAgreed(t *testing.T) {
	var current bool
	f := newFake()
	w := Window{work: behind{f, []string{"add-the-estimate"}, &current}}.Look()
	if w.question != aboutApplying {
		t.Fatalf("the window asks %v", w.question)
	}
	w = w.Choose("agree")
	if _, ok := f.gave(engine.MigrationApply.Thing, engine.MigrationApply.Verb); !ok {
		t.Fatalf("agreeing did not apply the change")
	}
	if w.Screen() != OnTheBoard || w.agreeing.Op.Verb != "" || w.Wrong() != nil {
		t.Errorf("the window is on %v holding %+v, wrong: %v", w.Screen(), w.agreeing, w.Wrong())
	}
	if w.Showing().Version != f.version {
		t.Errorf("the board was not drawn again")
	}
	var late bool
	w = Window{work: behind{newFake(), []string{"add-the-estimate"}, &late}}.asking(engine.TaskMove, engine.Input{})
	if w.Screen() != OnAsking || w.question != aboutApplying {
		t.Errorf("an act refused for a store behind lands on %v asking %v", w.Screen(), w.question)
	}
}
