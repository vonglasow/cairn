package gui

import (
	"testing"

	"cairn/internal/reading"
)

// The modal that adds and edits. What it asks about comes from reading; what is
// here is which operation a saved form means, and what the two buttons do.

func TestEditingStartsOnWhatTheTaskCarries(t *testing.T) {
	f := newFake()
	w := holding(f, reading.Card{Number: 1, Title: "Write the spec", State: "todo", Priority: "high"})

	w = w.Click(middleOf(w, 1, 0))
	w = w.Choose("Edit")

	if w.Screen() != OnForm {
		t.Fatalf("the window is on %v", w.Screen())
	}
	if got := w.Form().ValueOf("title"); got != "Write the spec" {
		t.Errorf("the title reads %q", got)
	}
	if got := w.Form().ValueOf("column"); got != "todo" {
		t.Errorf("the column reads %q", got)
	}
	if got := w.Form().ValueOf("nothing at all"); got != "" {
		t.Errorf("a field that is not there answered %q", got)
	}
}

// Nothing chosen opens nothing: there is no task to be a form about.
func TestEditingNothingOpensNothing(t *testing.T) {
	f := newFake()
	w := holding(f, reading.Card{Number: 1, Title: "Write the spec", State: "todo"})
	if w = w.Choose("Edit"); w.Screen() != OnTheBoard {
		t.Fatalf("editing nothing left the window on %v", w.Screen())
	}
}

func TestEditingAnObjectiveAsksTheObjectivesQuestions(t *testing.T) {
	f := newFake()
	w := New(f, 1200, 800)
	w.snap = reading.Snapshot{Goals: []reading.Goal{{Number: 3, Title: "Ship cairn", State: "active"}}}
	w.tab = Objectives
	w.chosen, w.chosenGoal = 3, true

	w = w.Edit()
	if !w.Form().ForObjective {
		t.Fatal("editing an objective opened a task's form")
	}
	for _, absent := range []string{"blockers", "parent", "objectives"} {
		if w.Form().ValueOf(absent) != "" {
			t.Errorf("an objective's form asks about %q", absent)
		}
	}
}

func TestSavingAFormWritesOnceAndClosesIt(t *testing.T) {
	f := newFake()
	w := holding(f, reading.Card{Number: 1, Title: "Write the spec", State: "todo"})

	w = w.Click(middleOf(w, 1, 0))
	w = w.Choose("Edit")
	w.form = w.form.Set("title", " Write the whole spec ")
	w = w.Choose("Save")

	if w.Screen() != OnTheBoard {
		t.Fatalf("saving left the window on %v", w.Screen())
	}
	if f.what("task", "edit") != 1 {
		t.Fatalf("saving asked for %d edits", f.what("task", "edit"))
	}
	in, _ := f.gave("task", "edit")
	if in.Given["title"] != "Write the whole spec" {
		t.Errorf("the title travelled as %q, and the space somebody left is not part of a name", in.Given["title"])
	}
	if in.Given["number"] != "1" {
		t.Errorf("the form saved task %q", in.Given["number"])
	}
}

func TestAddingSavesWithoutANumber(t *testing.T) {
	f := newFake()
	w := New(f, 1200, 800)

	w = w.Key("cmd+n")
	if w.Screen() != OnForm || !w.Form().Adding {
		t.Fatalf("cmd+n left the window on %v", w.Screen())
	}
	w.form = w.form.Set("title", "Write the spec")
	w = w.Choose("Save")

	if f.what("task", "add") != 1 {
		t.Fatalf("adding asked for %d adds", f.what("task", "add"))
	}
	in, _ := f.gave("task", "add")
	if _, named := in.Given["number"]; named {
		t.Error("a task being added was given a number it does not have yet")
	}
}

func TestCancellingWritesNothing(t *testing.T) {
	f := newFake()
	w := holding(f, reading.Card{Number: 1, Title: "Write the spec", State: "todo"})

	w = w.Click(middleOf(w, 1, 0))
	w = w.Choose("Edit")
	w.form = w.form.Set("title", "Something else")
	w = w.Choose("Cancel")

	if w.Screen() != OnTheBoard {
		t.Fatalf("cancelling left the window on %v", w.Screen())
	}
	if len(f.asked) != 0 {
		t.Errorf("cancelling asked the store %d things", len(f.asked))
	}
}

// Save and Cancel are buttons on a form, and a board with no form open has
// neither. Pressing one anyway does nothing rather than something surprising.
func TestTheButtonsDoNothingWithNoFormOpen(t *testing.T) {
	f := newFake()
	w := holding(f, reading.Card{Number: 1, Title: "Write the spec", State: "todo"})
	for _, name := range []string{"Save", "Cancel", "something nobody offers"} {
		if got := w.Choose(name); got.Screen() != OnTheBoard {
			t.Errorf("%q left the window on %v", name, got.Screen())
		}
	}
	if len(f.asked) != 0 {
		t.Errorf("the store was asked %d things", len(f.asked))
	}
}

// A refused save says so and keeps the form open, because the form holds what
// somebody typed and closing it would throw that away.
func TestASaveTheStoreRefusesKeepsTheForm(t *testing.T) {
	f := newFake()
	f.failOp = errRefuse
	w := holding(f, reading.Card{Number: 1, Title: "Write the spec", State: "todo"})

	w = w.Click(middleOf(w, 1, 0))
	w = w.Choose("Edit")
	w = w.Choose("Save")

	if w.Wrong() == nil {
		t.Fatal("a refused save said nothing")
	}
	if w.Screen() != OnForm {
		t.Fatalf("a refused save closed the form and lost what was typed")
	}
}

// The keyboard walks the fields and stops at both ends: a form that wraps is a
// form somebody falls off the bottom of.
func TestWalkingTheFieldsStopsAtBothEnds(t *testing.T) {
	form := adding(Tasks)
	if got := form.Walk(-1).On(); got != 0 {
		t.Errorf("walking up from the first field landed on %d", got)
	}
	last := len(form.Fields()) - 1
	if got := form.Walk(last + 5).On(); got != last {
		t.Errorf("walking past the last field landed on %d, want %d", got, last)
	}
	if got := form.Walk(2).On(); got != 2 {
		t.Errorf("walking two fields landed on %d", got)
	}
}

// Setting a field puts the keyboard on it, because that is where somebody
// typing is looking.
func TestSettingAFieldMovesTheKeyboardToIt(t *testing.T) {
	form := adding(Tasks).Set("priority", "high")
	if got := form.ValueOf("priority"); got != "high" {
		t.Fatalf("the field reads %q", got)
	}
	if form.Fields()[form.On()].Name != "priority" {
		t.Errorf("the keyboard is on %q", form.Fields()[form.On()].Name)
	}
	if same := form.Set("nothing at all", "x"); same.ValueOf("title") != "" {
		t.Error("setting a field that is not there wrote somewhere else")
	}
}

// Escape is the other way out of a form, and writes nothing.
func TestEscapeLeavesAFormAndWritesNothing(t *testing.T) {
	f := newFake()
	w := holding(f, reading.Card{Number: 1, Title: "Write the spec", State: "todo"})
	w = w.Click(middleOf(w, 1, 0))
	w = w.Choose("Edit")
	w = w.Key("esc")

	if w.Screen() != OnTheBoard {
		t.Fatalf("escape left the window on %v", w.Screen())
	}
	if len(f.asked) != 0 {
		t.Errorf("escape asked the store %d things", len(f.asked))
	}
}

// A key nobody bound does nothing, rather than something.
func TestAKeyNobodyBoundDoesNothing(t *testing.T) {
	f := newFake()
	w := holding(f, reading.Card{Number: 1, Title: "Write the spec", State: "todo"})
	if got := w.Key("q"); got.Screen() != OnTheBoard || len(f.asked) != 0 {
		t.Error("an unbound key did something")
	}
}

// Typing. Without it a form is a window that shows what a task carries and
// refuses to change any of it, which is what it was.

func TestTypingGoesIntoTheFieldTheKeyboardIsOn(t *testing.T) {
	f := newFake()
	w := holding(f, reading.Card{Number: 1, Title: "Write the spec", State: "todo"})
	w = w.Click(middleOf(w, 1, 0)).Choose("Edit")

	w = w.On("title").Types(" properly")
	if got := w.Form().ValueOf("title"); got != "Write the spec properly" {
		t.Fatalf("the title reads %q", got)
	}
	w = w.On("description").Types("two columns")
	if got := w.Form().ValueOf("description"); got != "two columns" {
		t.Errorf("the description reads %q", got)
	}
	if got := w.Form().ValueOf("title"); got != "Write the spec properly" {
		t.Errorf("typing elsewhere changed the title to %q", got)
	}
}

// Typing with no form open goes nowhere rather than somewhere.
func TestTypingWithNoFormOpenChangesNothing(t *testing.T) {
	f := newFake()
	w := holding(f, reading.Card{Number: 1, Title: "Write the spec", State: "todo"})
	if got := w.Types("hello"); got.Form().ValueOf("title") != "" {
		t.Error("typing on the board wrote into a form nobody opened")
	}
}

func TestBackspaceTakesTheLastCharacterBack(t *testing.T) {
	f := newFake()
	w := holding(f, reading.Card{Number: 1, Title: "Write the specx", State: "todo"})
	w = w.Click(middleOf(w, 1, 0)).Choose("Edit").On("title")

	w = w.Key("backspace")
	if got := w.Form().ValueOf("title"); got != "Write the spec" {
		t.Fatalf("the title reads %q", got)
	}
	// Taking back what is not there leaves nothing rather than failing.
	empty := w.On("because")
	for range 3 {
		empty = empty.Key("backspace")
	}
	if got := empty.Form().ValueOf("because"); got != "" {
		t.Errorf("an empty field reads %q", got)
	}
}

// A character is one rune and not one byte: taking back the last of "réunion"
// leaves "réunio", not half of an n.
func TestBackspaceTakesBackOneCharacterAndNotOneByte(t *testing.T) {
	f := newFake()
	w := holding(f, reading.Card{Number: 1, Title: "réunion", State: "todo"})
	w = w.Click(middleOf(w, 1, 0)).Choose("Edit").On("title")

	w = w.Key("backspace")
	if got := w.Form().ValueOf("title"); got != "réunio" {
		t.Fatalf("the title reads %q", got)
	}
	for range 6 {
		w = w.Key("backspace")
	}
	if got := w.Form().ValueOf("title"); got != "" {
		t.Errorf("after taking it all back the title reads %q", got)
	}
}

func TestTabWalksTheFieldsAndShiftTabWalksBack(t *testing.T) {
	f := newFake()
	w := holding(f, reading.Card{Number: 1, Title: "Write the spec", State: "todo"})
	w = w.Click(middleOf(w, 1, 0)).Choose("Edit").On("title")

	w = w.Key("tab")
	if got := w.Form().Fields()[w.Form().On()].Name; got != "description" {
		t.Fatalf("tab landed on %q", got)
	}
	w = w.Key("shift+tab")
	if got := w.Form().Fields()[w.Form().On()].Name; got != "title" {
		t.Fatalf("shift+tab landed on %q", got)
	}
	// Down and up do the same, because a form is a column of fields.
	if got := w.Key("down").Form().On(); got != 1 {
		t.Errorf("down landed on field %d", got)
	}
	if got := w.Key("up").Form().On(); got != 0 {
		t.Errorf("up from the first field landed on %d", got)
	}
}

// A field with a closed set is changed in place, because there is nothing to
// type: seven columns and four priorities are values to cycle.
func TestAFieldWithAClosedSetIsCycledRatherThanTyped(t *testing.T) {
	f := newFake()
	w := holding(f, reading.Card{Number: 1, Title: "Write the spec", State: "backlog"})
	w = w.Click(middleOf(w, 0, 0)).Choose("Edit").On("column")

	if got := w.Key("right").Form().ValueOf("column"); got != "todo" {
		t.Errorf("right of backlog is %q", got)
	}
	if got := w.Key("left").Form().ValueOf("column"); got != "backlog" {
		t.Errorf("left of backlog is %q, and there is nothing before it", got)
	}
	last := w.On("column")
	for range 10 {
		last = last.Key("right")
	}
	if got := last.Form().ValueOf("column"); got != "done" {
		t.Errorf("past the last of the set is %q", got)
	}
	// A field with no closed set is not cycled by the same keys.
	typed := w.On("title").Key("right")
	if got := typed.Form().ValueOf("title"); got != "Write the spec" {
		t.Errorf("right in a typed field changed it to %q", got)
	}
}

// Clicking a field is how the keyboard gets there with a mouse.
func TestClickingAFieldPutsTheKeyboardOnIt(t *testing.T) {
	f := newFake()
	w := holding(f, reading.Card{Number: 1, Title: "Write the spec", State: "todo"})
	w = w.Click(middleOf(w, 1, 0)).Choose("Edit")

	box := Over(w.width, w.height)
	at := 3
	name, _ := FieldRow(box, FieldTop(box, w.Form().Fields(), at))
	w = w.Click(Point{X: name.X + 10, Y: name.Y + name.H/2})
	if got := w.Form().On(); got != at {
		t.Fatalf("clicking field %d put the keyboard on %d", at, got)
	}
}

// Enter saves, because a form is answered rather than dismissed.
func TestEnterSavesTheForm(t *testing.T) {
	f := newFake()
	w := holding(f, reading.Card{Number: 1, Title: "Write the spec", State: "todo"})
	w = w.Click(middleOf(w, 1, 0)).Choose("Edit")

	w = w.Key("enter")
	if w.Screen() != OnTheBoard {
		t.Fatalf("enter left the window on %v", w.Screen())
	}
	if f.what("task", "edit") != 1 {
		t.Errorf("enter asked for %d edits", f.what("task", "edit"))
	}
}

// Save and Cancel are drawn on the form and have to work where they are drawn.
// They were clickable only down in the bar, which is not where anybody looks.
func TestTheFormsOwnButtonsAnswerWhereTheyAreDrawn(t *testing.T) {
	f := newFake()
	w := holding(f, reading.Card{Number: 1, Title: "Write the spec", State: "todo"})
	opened := w.Click(middleOf(w, 1, 0)).Choose("Edit")

	box := Over(opened.width, opened.height)
	press := func(at int) Window {
		button := PanelOfferAt(box, at)
		return opened.Click(Point{X: button.X + button.W/2, Y: button.Y + button.H/2})
	}

	if got := press(0); got.Screen() != OnTheBoard || f.what("task", "edit") != 1 {
		t.Errorf("Save on the form left the window on %v having asked for %d edits",
			got.Screen(), f.what("task", "edit"))
	}
	f.asked = nil
	if got := press(1); got.Screen() != OnTheBoard || len(f.asked) != 0 {
		t.Errorf("Cancel on the form left the window on %v having asked %d things",
			got.Screen(), len(f.asked))
	}
}

// The same for the question asked before deleting, and for the help.
func TestEveryPanelsButtonsAnswerWhereTheyAreDrawn(t *testing.T) {
	f := newFake()
	w := holding(f, reading.Card{Number: 1, Title: "Write the spec", State: "todo"})
	w = w.Click(middleOf(w, 1, 0))

	box := Over(w.width, w.height)
	press := func(from Window, at int) Window {
		button := PanelOfferAt(box, at)
		return from.Click(Point{X: button.X + button.W/2, Y: button.Y + button.H/2})
	}

	asked := w.Choose("Delete")
	if got := press(asked, 0); f.what("task", "delete") != 1 {
		t.Errorf("Delete on the question asked for %d deletes, leaving the window on %v",
			f.what("task", "delete"), got.Screen())
	}
	if got := press(w.Choose("Help"), 0); got.Screen() != OnTheBoard {
		t.Errorf("Close on the help left the window on %v", got.Screen())
	}
	if got := press(w.DoubleClick(middleOf(w, 1, 0)), 1); got.Screen() != OnTheBoard {
		t.Errorf("Close on the detail left the window on %v", got.Screen())
	}
}

// A field that wants several rows is clicked anywhere in them, not only on its
// first line.
func TestAFieldOfSeveralRowsIsClickedAnywhereInThem(t *testing.T) {
	f := newFake()
	w := holding(f, reading.Card{Number: 1, Title: "Write the spec", State: "todo"})
	w = w.Click(middleOf(w, 1, 0)).Choose("Edit")

	box := Over(w.width, w.height)
	fields := w.Form().Fields()
	at := 1 // the description
	if !fields[at].Long {
		t.Fatalf("field %d is %q, which is not the long one", at, fields[at].Name)
	}
	_, value := FieldRow(box, FieldTop(box, fields, at))

	name, _ := FieldRow(box, FieldTop(box, fields, at))
	for _, row := range []int{0, RowsOf(fields[at]) - 1} {
		down := row*lineHigh + lineHigh/2
		if got := w.Click(Point{X: value.X + 10, Y: value.Y + down}).Form().On(); got != at {
			t.Errorf("clicking row %d of the description put the keyboard on %d", row, got)
		}
		// Its name answers over the same rows, or the left half of a tall
		// field would be dead.
		if got := w.Click(Point{X: name.X + 10, Y: name.Y + down}).Form().On(); got != at {
			t.Errorf("clicking row %d beside the description put the keyboard on %d", row, got)
		}
	}
}

// The keyboard lands on a field that exists, whatever it is asked for.
func TestTheKeyboardLandsOnAFieldThatExists(t *testing.T) {
	form := adding(Tasks)
	last := len(form.Fields()) - 1
	if got := form.onto(-5).On(); got != 0 {
		t.Errorf("before the first field is %d", got)
	}
	if got := form.onto(last + 5).On(); got != last {
		t.Errorf("past the last field is %d, want %d", got, last)
	}
	// And arriving at a field puts the caret at its end.
	arrived := form.Set("title", "spec").onto(0)
	if got := arrived.Writing().At; got != 4 {
		t.Errorf("arriving at a field of four characters put the caret at %d", got)
	}
}

// What a field says when it carries a number the snapshot has never heard of.
// A name is looked up and there is none: the number stands on its own rather
// than the field going blank, because a number that names nothing is still
// truer than an empty space where a name should be.
func TestFieldSays_keepsANumberItCanPutNoNameTo(t *testing.T) {
	f := newFake()
	w := New(f, 1200, 800)
	w.snap = reading.Snapshot{
		Goals: []reading.Goal{{Number: 1, Title: "Ship the desktop board"}},
	}
	field := reading.Field{Name: "objectives", Arg: "objective", Value: "1 7"}
	// Asked whole, and not for the two halves separately: a name and a number
	// both being in there says nothing about the name belonging to that
	// number. Measured — a test that asked that way passed with the lookup
	// inverted, handing each number the other one's name.
	want := "#1 Ship the desktop board, #7 "
	if got := w.FieldSays(field); got != want {
		t.Errorf("FieldSays = %q, want %q", got, want)
	}
}
