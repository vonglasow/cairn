package tui

import (
	"cairn/internal/engine"
	"errors"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestPushingIntoAColumnAsksTaskEdit(t *testing.T) {
	f := newFake()
	f.cards = []engine.Card{{Number: 1, Title: "One", State: "current"}}
	b := hit(standing(f), "l", "l", "i")
	if _, cmd := b.pressed("shift+h"); cmd != nil {
		cmd()
	}
	if len(f.asked) == 0 {
		t.Fatal("the board asked for nothing")
	}
	got := f.asked[0]
	if got.op.Thing != "task" || got.op.Verb != "edit" {
		t.Errorf("asked %q %q, want task edit", got.op.Thing, got.op.Verb)
	}
	if got.in.Given["number"] != "1" {
		t.Errorf("number = %q, want 1", got.in.Given["number"])
	}
	if got.in.Given["state"] != "todo" {
		t.Errorf("state = %q, want todo", got.in.Given["state"])
	}
}

func asks(t *testing.T, f *fake, keys []string, press string) asked {
	t.Helper()
	b := hit(standing(f), keys...)
	if _, cmd := b.pressed(press); cmd != nil {
		cmd()
	}
	if len(f.asked) == 0 {
		t.Fatal("the board asked for nothing")
	}
	return f.asked[0]
}

func TestEveryGestureAsksTheOperationItMeans(t *testing.T) {
	two := []engine.Card{{Number: 1, Title: "One", State: "todo"}, {Number: 2, Title: "Two", State: "todo"}}
	for _, c := range []struct {
		what   string
		cards  []engine.Card
		goals  []engine.Goal
		keys   []string
		press  string
		thing  string
		verb   string
		number string
	}{
		{what: "taking the focus", cards: two[:1], keys: []string{"l", "i"}, press: "f", thing: "task", verb: "focus", number: "1"},
		{what: "deleting a task", cards: two[:1], keys: []string{"l", "i", "d"}, press: "enter", thing: "task", verb: "delete", number: "1"},
		{what: "pushing a task down", cards: two, keys: []string{"l", "i"}, press: "shift+j", thing: "task", verb: "move", number: "1"},
		{
			what: "deleting an objective", goals: []engine.Goal{{Number: 3, Title: "Ship"}},
			keys: []string{"tab", "d"}, press: "enter", thing: "objective", verb: "delete", number: "3",
		},
	} {
		t.Run(c.what, func(t *testing.T) {
			f := newFake()
			f.cards, f.goals = c.cards, c.goals
			got := asks(t, f, c.keys, c.press)
			if got.op.Thing != c.thing || got.op.Verb != c.verb {
				t.Errorf("asked %q %q, want %q %q", got.op.Thing, got.op.Verb, c.thing, c.verb)
			}
			if got.in.Given["number"] != c.number {
				t.Errorf("number = %q, want %q", got.in.Given["number"], c.number)
			}
		})
	}
}

func TestLettingTheFocusGoAsksTaskUnfocusByNumber(t *testing.T) {
	f := newFake()
	f.cards = []engine.Card{{Number: 1, Title: "One", State: "todo", InFocus: true}}
	f.focus = []int{1}
	got := asks(t, f, []string{"l", "i"}, "f")
	if got.op.Thing != "task" || got.op.Verb != "unfocus" {
		t.Errorf("asked %q %q, want task unfocus", got.op.Thing, got.op.Verb)
	}
	if got.in.Given["number"] != "1" {
		t.Errorf("it let go of %q, want task 1", got.in.Given["number"])
	}
}

func TestTakingUpATaskAsksTaskFocusByNumber(t *testing.T) {
	f := newFake()
	f.cards = []engine.Card{
		{Number: 1, Title: "One", State: "todo"},
		{Number: 2, Title: "Two", State: "todo", InFocus: true},
	}
	f.focus = []int{2}
	got := asks(t, f, []string{"l", "i"}, "f")
	if got.op.Thing != "task" || got.op.Verb != "focus" {
		t.Errorf("asked %q %q, want task focus", got.op.Thing, got.op.Verb)
	}
	if got.in.Given["number"] != "1" {
		t.Errorf("it took up %q, want task 1", got.in.Given["number"])
	}
}

func TestReadingTheHistoryAsksTraceShow(t *testing.T) {
	f := newFake()
	b := standing(f)
	_ = heardOf(b, 14)
	if len(f.asked) != 1 {
		t.Fatalf("the history asked for %d operations, want 1", len(f.asked))
	}
	got := f.asked[0]
	if got.op.Thing != "trace" || got.op.Verb != "show" {
		t.Errorf("asked %q %q, want trace show", got.op.Thing, got.op.Verb)
	}
	if got.in.Given["number"] != "14" {
		t.Errorf("number = %q, want 14", got.in.Given["number"])
	}
}

func TestTheBoardAsksWhatTheEngineSaysIsAQuestion(t *testing.T) {
	f := newFake()
	f.cards = []engine.Card{{Number: 1, Title: "One", State: "current"}}
	op := engine.Operation{Thing: "task", Verb: "done", Finishes: true}
	in := engine.Agreed(engine.Input{Given: map[string]string{"number": "1"}})
	model, _ := standing(f).Update(engine.Agreeing{Reason: "step-still-open", Op: op, In: in})
	b := model.(Board)
	if !b.asking {
		t.Fatal("the board is not asking")
	}
	if _, cmd := b.pressed("enter"); cmd != nil {
		cmd()
	}
	if len(f.asked) == 0 {
		t.Fatal("the board asked for nothing")
	}
	got := f.asked[0]
	if got.op.Verb != "done" || got.in.Given["number"] != "1" || !got.in.Confirmed {
		t.Errorf("asked %q %v confirmed=%v, want done 1 confirmed", got.op.Verb, got.in.Given, got.in.Confirmed)
	}
}

func TestALookRefusedForAStoreBehindIsAskedAbout(t *testing.T) {
	f := newFake()
	f.fail = engine.ErrStoreBehind
	b := pump(standing(f), standing(f).look())
	if !b.asking {
		t.Fatalf("the board is not asking, and says %q", strip(b.Draw()))
	}
	if said := strip(b.foot()); !strings.Contains(said, "apply what this store is waiting for?") {
		t.Errorf("the foot says %q", said)
	}
}

func TestAFailedApplyIsSaidAndTheStoreIsAskedAboutAgain(t *testing.T) {
	f := newFake()
	f.fail = engine.ErrStoreBehind
	b := pump(standing(f), standing(f).look())
	f.fail, f.failOp = nil, errNope
	b = hit(b, "enter")
	if !b.asking || !b.aboutStore() {
		t.Fatalf("the board stopped asking after the apply failed, and says %q", strip(b.foot()))
	}
	if !errors.Is(b.err, errNope) {
		t.Errorf("the board said %v, want what failed", b.err)
	}
	f.failOp = nil
	b = hit(b, "enter")
	if b.asking {
		t.Errorf("the question stands after the apply went through: %q", strip(b.foot()))
	}
}

func TestEscOnAStoreBehindLeavesTheBoardAndTouchesNothing(t *testing.T) {
	f := newFake()
	f.fail = engine.ErrStoreBehind
	b := pump(standing(f), standing(f).look())
	asked := len(f.asked)
	_, cmd := b.pressed("esc")
	if cmd == nil {
		t.Fatal("esc on the question asked for nothing")
	}
	if _, quits := cmd().(tea.QuitMsg); !quits {
		t.Errorf("esc asked for %T, not to quit", cmd())
	}
	if len(f.asked) != asked {
		t.Errorf("leaving asked the store %d more times", len(f.asked)-asked)
	}
}

func TestAStoreFromANewerCairnIsSaidAndOnlyLeftWithEsc(t *testing.T) {
	f := newFake()
	f.fail = engine.ErrNewerStore
	b := pump(standing(f), standing(f).look())
	if said := strip(b.foot()); !strings.Contains(said, "this store comes from a newer cairn") {
		t.Fatalf("the foot says %q", said)
	}
	asked := len(f.asked)
	after, cmd := b.pressed("enter")
	if cmd != nil {
		t.Errorf("enter on a store from a newer cairn asked for %T", cmd())
	}
	if len(f.asked) != asked || !after.(Board).asking {
		t.Errorf("enter went ahead: asking=%v, asked %d more", after.(Board).asking, len(f.asked)-asked)
	}
	_, cmd = b.pressed("esc")
	if cmd == nil {
		t.Fatal("esc asked for nothing")
	}
	if _, quits := cmd().(tea.QuitMsg); !quits {
		t.Errorf("esc asked for %T, not to quit", cmd())
	}
}
