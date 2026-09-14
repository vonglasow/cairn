package board

import (
	"testing"

	"cairn/internal/reading"
)

// The board asks the engine through the operation the engine describes, not
// through a method it named itself. This is the test that says so.
func TestPushingIntoAColumnAsksTaskEdit(t *testing.T) {
	f := newFake()
	f.cards = []reading.Card{{Number: 1, Title: "One", State: "current"}}
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

// asks drives a gesture and answers the one operation it asked for.
func asks(t *testing.T, f *fake, keys []string, press string) asked {
	t.Helper()
	b := hit(standing(f), keys...)
	if _, cmd := b.pressed(press); cmd != nil {
		cmd()
	}
	// What follows is the refresh: a gesture writes, then reads the board back.
	// Only the first is the gesture's own doing.
	if len(f.asked) == 0 {
		t.Fatal("the board asked for nothing")
	}
	return f.asked[0]
}

func TestEveryGestureAsksTheOperationItMeans(t *testing.T) {
	two := []reading.Card{{Number: 1, Title: "One", State: "todo"}, {Number: 2, Title: "Two", State: "todo"}}
	for _, c := range []struct {
		what   string
		cards  []reading.Card
		goals  []reading.Goal
		keys   []string
		press  string
		thing  string
		verb   string
		number string
	}{
		{what: "taking the focus", cards: two[:1], keys: []string{"l", "i"}, press: "f", thing: "task", verb: "focus", number: "1"},
		{what: "deleting a task", cards: two[:1], keys: []string{"l", "i"}, press: "d", thing: "task", verb: "delete", number: "1"},
		{what: "pushing a task down", cards: two, keys: []string{"l", "i"}, press: "shift+j", thing: "task", verb: "move", number: "1"},
		{
			what: "deleting an objective", goals: []reading.Goal{{Number: 3, Title: "Ship"}},
			keys: []string{"tab"}, press: "d", thing: "objective", verb: "delete", number: "3",
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

// Letting the focus go is its own operation, and it names the task: several
// can be in focus at once, so letting go of nothing in particular would let go
// of every one of them.
func TestLettingTheFocusGoAsksTaskUnfocusByNumber(t *testing.T) {
	f := newFake()
	f.cards = []reading.Card{{Number: 1, Title: "One", State: "todo", InFocus: true}}
	f.focus = []int{1}
	got := asks(t, f, []string{"l", "i"}, "f")
	if got.op.Thing != "task" || got.op.Verb != "unfocus" {
		t.Errorf("asked %q %q, want task unfocus", got.op.Thing, got.op.Verb)
	}
	if got.in.Given["number"] != "1" {
		t.Errorf("it let go of %q, want task 1", got.in.Given["number"])
	}
}

// And a task that is not in focus is taken up alongside whatever else is,
// rather than replacing it.
func TestTakingUpATaskAsksTaskFocusByNumber(t *testing.T) {
	f := newFake()
	f.cards = []reading.Card{
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

func TestTheHistoryPanelAsksTraceShow(t *testing.T) {
	f := newFake()
	b := standing(f)
	b.history(14, 80, 5)
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
