package gui

import (
	"cairn/internal/board"
	"errors"
	"fmt"
	"strings"
	"testing"

	"gioui.org/io/key"

	"cairn/internal/engine"
	"cairn/internal/engine/migration"
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
	return engine.Answer{}, engine.Told(migration.ErrStoreBehind)
}

func TestTheWindowAsksToApplyWhatIsWaitingWhenItsLookIsRefused(t *testing.T) {
	w := Window{work: behind{fake: newFake(), waiting: []string{"add-the-estimate"}}}.Look()
	if w.Screen != board.OnAsking || w.Question != board.AboutApplying {
		t.Fatalf("the window is on %v asking %v", w.Screen, w.Question)
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
	if w.Screen != board.OnAsking || w.Question != board.AboutApplying {
		t.Errorf("after the beats the window is on %v asking %v", w.Screen, w.Question)
	}
}

type unappliable struct{ behind }

func (u unappliable) Ask(op engine.Operation, in engine.Input) (engine.Answer, error) {
	if op.Thing == engine.MigrationApply.Thing && op.Verb == engine.MigrationApply.Verb {
		return engine.Answer{}, errors.New("disk full")
	}
	return u.behind.Ask(op, in)
}

func TestTheWindowAsksAgainWhenApplyingFails(t *testing.T) {
	var current bool
	w := Window{work: unappliable{behind{newFake(), []string{"add-the-estimate"}, &current}}}.Look()
	if w.Question != board.AboutApplying {
		t.Fatalf("the window asks %v", w.Question)
	}
	w = w.Choose("Apply")
	if w.Screen != board.OnAsking || w.Question != board.AboutApplying {
		t.Fatalf("after a failed apply the window is on %v asking %v", w.Screen, w.Question)
	}
	if w.agreeing.Op.Verb != engine.MigrationApply.Verb || w.agreeing.Op.Thing != engine.MigrationApply.Thing {
		t.Errorf("the window holds %+v", w.agreeing)
	}
	if w.Wrong() == nil || !strings.Contains(w.Wrong().Error(), "disk full") {
		t.Errorf("the failure is not said: %v", w.Wrong())
	}
}

type newer struct{ *fake }

func (n newer) Ask(op engine.Operation, in engine.Input) (engine.Answer, error) {
	return engine.Answer{}, engine.Told(fmt.Errorf("%w: a change it does not know", migration.ErrNewerStore))
}

func TestTheWindowSaysAStoreFromANewerCairnAndOffersOnlyToQuit(t *testing.T) {
	f := newFake()
	w := Window{work: newer{f}}.Look()
	if w.Screen != board.OnAsking || w.Question != board.AboutNewer {
		t.Fatalf("the window is on %v asking %v", w.Screen, w.Question)
	}
	if offers := w.Offers(); len(offers) != 1 || offers[0] != "Quit" {
		t.Errorf("the window offers %q, want only Quit", offers)
	}
	if w.Wrong() != nil {
		t.Errorf("the window says it went wrong: %v", w.Wrong())
	}
	for range 2 * Lingers() {
		w = w.aged()
	}
	if w.Screen != board.OnAsking || w.Question != board.AboutNewer {
		t.Errorf("after the beats the window is on %v asking %v", w.Screen, w.Question)
	}
	w = w.Choose("agree")
	if w.Screen != board.OnAsking || w.Question != board.AboutNewer || len(f.wrote()) != 0 {
		t.Errorf("agreeing left the window on %v asking %v after %d writes", w.Screen, w.Question, len(f.wrote()))
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

func TestADeletionIsAgreedToWithEnter(t *testing.T) {
	f := newFake()
	w := holding(f, card(1, "Write the spec", "todo"))
	w = w.clicked(middleOf(w, 1, 0)).Key("cmd+backspace")
	if w.Screen != board.OnAsking {
		t.Fatalf("d left the window on %v", w.Screen)
	}
	if again := w.Key("cmd+backspace"); len(f.asked) != 0 {
		t.Errorf("a second d deleted without being agreed to: %v, screen %v", f.asked, again.Screen)
	}
	if agreed := w.Key("enter"); agreed.Screen != board.OnTheBoard {
		t.Errorf("enter left the question on %v", agreed.Screen)
	}
}

func TestAWholeColumnGoesOffTheBoardAtOnce(t *testing.T) {
	f := newFake()
	w := holding(f, board.Card{Number: 1, Title: "One", State: "done"},
		board.Card{Number: 2, Title: "Two", State: "done"})
	w = w.clicked(middleOf(w, 3, 0))
	w = Typed(w, "X", key.ModShift)
	if w.Screen != board.OnAsking {
		t.Fatalf("a whole column went without a question: %v", w.Screen)
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
	w := holding(f, board.Card{Number: 1, Title: "One", State: "todo"},
		board.Card{Number: 2, Title: "Two", State: "todo"})
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
	holding(f, board.Card{Number: 9, Title: "Last", State: "done"}).
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

func TestShiftSidewaysOutOfTheBand(t *testing.T) {
	f := newFake()
	w := holding(f,
		board.Card{Number: 1, Title: "In todo", State: "todo"},
		board.Card{Number: 2, Title: "Blocked", State: "blocked"})
	w = w.clicked(middleOf(w, 1, 0)).Key("down")
	if w.Chosen != 2 {
		t.Fatalf("down out of TODO chose %d, want the blocked task", w.Chosen)
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
		board.Card{Number: 1, Title: "In todo", State: "todo"},
		board.Card{Number: 2, Title: "Blocked one", State: "blocked"},
		board.Card{Number: 3, Title: "Blocked two", State: "blocked"},
		board.Card{Number: 4, Title: "Watched", State: "watching"})

	w = w.clicked(middleOf(w, 1, 0)).Key("down")
	if w.Chosen != 2 {
		t.Fatalf("down out of TODO chose %d, want the first blocked task", w.Chosen)
	}
	if got := w.Key("down").Chosen; got != 3 {
		t.Fatalf("down in a lane chose %d, want 3", got)
	}
	if got := w.Key("down").Key("down").Chosen; got != 3 {
		t.Fatalf("down past the end of a lane chose %d, want 3", got)
	}
	if got := w.Key("right").Chosen; got != 4 {
		t.Fatalf("right out of the blocked lane chose %d, want 4", got)
	}
	if got := w.Key("left").Chosen; got != 2 {
		t.Fatalf("left out of the first lane chose %d, want to stay", got)
	}
	if got := w.Key("right").Key("right").Chosen; got != 4 {
		t.Fatalf("right off the last lane chose %d, want to stay on 4", got)
	}
	if got := w.Key("down").Key("right").Chosen; got != 4 {
		t.Fatalf("right from the second blocked card chose %d, want 4", got)
	}
	if got := w.clicked(middleOf(w, 1, 0)).Key("right").Chosen; got != 1 {
		t.Fatalf("right out of TODO chose %d, and should not have fallen into the band", got)
	}
	if got := w.Key("up").Chosen; got != 1 {
		t.Fatalf("up out of a lane chose %d, want the task above it", got)
	}
}

func TestTheWindowAppliesWhatIsWaitingOnceAgreed(t *testing.T) {
	var current bool
	f := newFake()
	w := Window{work: behind{f, []string{"add-the-estimate"}, &current}}.Look()
	if w.Question != board.AboutApplying {
		t.Fatalf("the window asks %v", w.Question)
	}
	w = w.Choose("agree")
	if _, ok := f.gave(engine.MigrationApply.Thing, engine.MigrationApply.Verb); !ok {
		t.Fatalf("agreeing did not apply the change")
	}
	if w.Screen != board.OnTheBoard || w.agreeing.Op.Verb != "" || w.Wrong() != nil {
		t.Errorf("the window is on %v holding %+v, wrong: %v", w.Screen, w.agreeing, w.Wrong())
	}
	if w.Showing().Version != f.version {
		t.Errorf("the board was not drawn again")
	}
	var late bool
	w = Window{work: behind{newFake(), []string{"add-the-estimate"}, &late}}.asking(engine.TaskMove, engine.Input{})
	if w.Screen != board.OnAsking || w.Question != board.AboutApplying {
		t.Errorf("an act refused for a store behind lands on %v asking %v", w.Screen, w.Question)
	}
}

func TestDeletingATaskSaysItsStepsGoWithIt(t *testing.T) {
	w := holding(newFake(), board.Card{Number: 1, Title: "Write the spec", State: "todo", SubTotal: 3})
	w = w.clicked(middleOf(w, 1, 0)).Key("cmd+backspace")
	said := ""
	for _, line := range w.Asks() {
		said += line.Means + " "
	}
	if !strings.Contains(said, "Its 3 steps and its labels go with it") {
		t.Errorf("deleting a task with three steps says %q", said)
	}
}
