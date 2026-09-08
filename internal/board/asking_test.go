package board

import (
	"testing"

	"cairn/internal/engine"
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

// Letting the focus go is its own operation and names no task: the board has to
// ask `task unfocus`, not `task focus` with nothing in it.
func TestLettingTheFocusGoAsksTaskUnfocus(t *testing.T) {
	f := newFake()
	f.cards = []reading.Card{{Number: 1, Title: "One", State: "todo"}}
	f.focus = 1
	got := asks(t, f, []string{"l", "i"}, "f")
	if got.op.Thing != "task" || got.op.Verb != "unfocus" {
		t.Errorf("asked %q %q, want task unfocus", got.op.Thing, got.op.Verb)
	}
}

// A form is turned into what the engine takes by reading each field's own
// declaration of which argument it fills. These two say so: the engine's word
// is what travels, and the screen's word never does.
func TestAskedFromNamesEveryFieldByItsEngineArgument(t *testing.T) {
	form := asking(reading.TaskFields(reading.Card{
		Title:       "One",
		Description: "why it matters",
		State:       "todo",
		Priority:    "high",
		Labels:      []string{"api", "urgent"},
	}), 80, 20, plainSkin)

	in := askedFrom(form, engine.TaskEdit)
	if in.Given["desc"] != "why it matters" {
		t.Errorf("desc = %q, want the description", in.Given["desc"])
	}
	if in.Given["state"] != "todo" {
		t.Errorf("state = %q, want todo", in.Given["state"])
	}
	if got := in.Listed["label"]; len(got) != 2 || got[0] != "api" || got[1] != "urgent" {
		t.Errorf("label = %v, want the two labels, listed", got)
	}
	for _, screen := range []string{"description", "labels", "column", "time spent"} {
		if _, named := in.Given[screen]; named {
			t.Errorf("%q travelled: that is the screen's word, not the engine's", screen)
		}
	}
}

// A number nobody can read is not named at all, rather than sent for the
// engine to refuse: one unreadable field would otherwise lose the whole form.
func TestAParentThatIsNotANumberIsNotNamed(t *testing.T) {
	form := asking([]reading.Field{{Name: "parent", Arg: "parent", Value: "nine"}}, 80, 6, plainSkin)
	if in := askedFrom(form, engine.TaskEdit); in.Named("parent") {
		t.Fatal("a parent nobody can read was passed to the engine anyway")
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
