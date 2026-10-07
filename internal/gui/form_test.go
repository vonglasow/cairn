package gui

import (
	"slices"
	"strconv"
	"strings"
	"testing"

	"cairn/internal/engine"
	"cairn/internal/reading"
)

// The modal that adds and edits. What it asks about comes from reading; what is
// here is which operation a saved form means, and what the two buttons do.

func TestEditingStartsOnWhatTheTaskCarries(t *testing.T) {
	f := newFake()
	w := holding(f, reading.Card{Number: 1, Title: "Write the spec", State: "todo", Priority: "high"})

	w = w.clicked(middleOf(w, 1, 0))
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

	w = w.clicked(middleOf(w, 1, 0))
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

	w = w.clicked(middleOf(w, 1, 0))
	w = w.Choose("Edit")
	w.form = w.form.Set("title", "Something else")
	// Twice: once for the question a changed form now asks, once to agree to
	// losing what was typed.
	w = w.Choose("Cancel").Choose("Agree")

	if w.Screen() != OnTheBoard {
		t.Fatalf("cancelling left the window on %v", w.Screen())
	}
	if len(f.wrote()) != 0 {
		t.Errorf("cancelling asked the store %d things", len(f.wrote()))
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
	if len(f.wrote()) != 0 {
		t.Errorf("the store was asked %d things", len(f.wrote()))
	}
}

// A refused save says so and keeps the form open, because the form holds what
// somebody typed and closing it would throw that away.
func TestASaveTheStoreRefusesKeepsTheForm(t *testing.T) {
	f := newFake()
	f.failOp = errRefuse
	w := holding(f, reading.Card{Number: 1, Title: "Write the spec", State: "todo"})

	w = w.clicked(middleOf(w, 1, 0))
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
	form := adding(Tasks, "backlog")
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
	form := adding(Tasks, "backlog").Set("priority", "high")
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
	w = w.clicked(middleOf(w, 1, 0))
	w = w.Choose("Edit")
	w = w.Key("esc")

	if w.Screen() != OnTheBoard {
		t.Fatalf("escape left the window on %v", w.Screen())
	}
	if len(f.wrote()) != 0 {
		t.Errorf("escape asked the store %d things", len(f.wrote()))
	}
}

// A key nobody bound does nothing, rather than something.
func TestAKeyNobodyBoundDoesNothing(t *testing.T) {
	f := newFake()
	w := holding(f, reading.Card{Number: 1, Title: "Write the spec", State: "todo"})
	if got := w.Key("q"); got.Screen() != OnTheBoard || len(f.wrote()) != 0 {
		t.Error("an unbound key did something")
	}
}

// Typing. Without it a form is a window that shows what a task carries and
// refuses to change any of it, which is what it was.

func TestTypingGoesIntoTheFieldTheKeyboardIsOn(t *testing.T) {
	f := newFake()
	w := holding(f, reading.Card{Number: 1, Title: "Write the spec", State: "todo"})
	w = w.clicked(middleOf(w, 1, 0)).Choose("Edit")

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
	w = w.clicked(middleOf(w, 1, 0)).Choose("Edit").On("title")

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
	w = w.clicked(middleOf(w, 1, 0)).Choose("Edit").On("title")

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
	w = w.clicked(middleOf(w, 1, 0)).Choose("Edit").On("title")

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
	w = w.clicked(middleOf(w, 0, 0)).Choose("Edit").On("column")

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
	w = w.clicked(middleOf(w, 1, 0)).Choose("Edit")

	box := Over(w.width, w.height)
	at := 3
	name := formName(box, w.Form().Fields(), at)
	w = w.clicked(Point{X: name.X + 10, Y: name.Y + name.H/2})
	if got := w.Form().On(); got != at {
		t.Fatalf("clicking field %d put the keyboard on %d", at, got)
	}
}

// Enter saves, because a form is answered rather than dismissed.
func TestTheKeyTheBarOffersSavesTheForm(t *testing.T) {
	f := newFake()
	w := holding(f, reading.Card{Number: 1, Title: "Write the spec", State: "todo"})
	w = w.clicked(middleOf(w, 1, 0)).Choose("Edit")

	w = w.Key("cmd+enter")
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
	opened := w.clicked(middleOf(w, 1, 0)).Choose("Edit")

	box := opened.formPanel()
	press := func(at int) Window {
		button := PanelOfferAt(box, at)
		return opened.clicked(Point{X: button.X + button.W/2, Y: button.Y + button.H/2})
	}

	if got := press(0); got.Screen() != OnTheBoard || f.what("task", "edit") != 1 {
		t.Errorf("Save on the form left the window on %v having asked for %d edits",
			got.Screen(), f.what("task", "edit"))
	}
	f.asked = nil
	if got := press(1); got.Screen() != OnTheBoard || len(f.wrote()) != 0 {
		t.Errorf("Cancel on the form left the window on %v having asked %d things",
			got.Screen(), len(f.wrote()))
	}
}

// The same for the question asked before deleting, and for the help.
func TestEveryPanelsButtonsAnswerWhereTheyAreDrawn(t *testing.T) {
	f := newFake()
	w := holding(f, reading.Card{Number: 1, Title: "Write the spec", State: "todo"})
	w = w.clicked(middleOf(w, 1, 0))

	press := func(from Window, at int) Window {
		button := PanelOfferAt(from.panel(), at)
		return from.clicked(Point{X: button.X + button.W/2, Y: button.Y + button.H/2})
	}

	asked := w.Choose("Delete")
	if got := press(asked, 0); f.what("task", "delete") != 1 {
		t.Errorf("Delete on the question asked for %d deletes, leaving the window on %v",
			f.what("task", "delete"), got.Screen())
	}
	if got := press(w.Choose("Help"), 0); got.Screen() != OnTheBoard {
		t.Errorf("Close on the help left the window on %v", got.Screen())
	}
	if got := press(w.doubleClicked(middleOf(w, 1, 0)), 1); got.Screen() != OnTheBoard {
		t.Errorf("Close on the detail left the window on %v", got.Screen())
	}
}

// A field that wants several rows is clicked anywhere in them, not only on its
// first line.
func TestAFieldOfSeveralRowsIsClickedAnywhereInThem(t *testing.T) {
	f := newFake()
	w := holding(f, reading.Card{Number: 1, Title: "Write the spec", State: "todo"})
	w = w.clicked(middleOf(w, 1, 0)).Choose("Edit")

	box := Over(w.width, w.height)
	fields := w.Form().Fields()
	at := 1 // the description
	if !fields[at].Long {
		t.Fatalf("field %d is %q, which is not the long one", at, fields[at].Name)
	}
	value := formValue(box, fields, at)

	name := formName(box, fields, at)
	for _, row := range []int{0, formRows(box, fields, at) - 1} {
		down := row*lineHigh + lineHigh/2
		if got := w.clicked(Point{X: value.X + 10, Y: value.Y + down}).Form().On(); got != at {
			t.Errorf("clicking row %d of the description put the keyboard on %d", row, got)
		}
		// Its name answers over the same rows, or the left half of a tall
		// field would be dead.
		if got := w.clicked(Point{X: name.X + 10, Y: name.Y + down}).Form().On(); got != at {
			t.Errorf("clicking row %d beside the description put the keyboard on %d", row, got)
		}
	}
}

// The keyboard lands on a field that exists, whatever it is asked for.
func TestTheKeyboardLandsOnAFieldThatExists(t *testing.T) {
	form := adding(Tasks, "backlog")
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

// The objectives list narrows as it is typed into, the way the parents list
// does. One list, two fields: a filter that answered only one of them would be
// a list behaving differently depending on where it was opened from.
func TestTheObjectivesListNarrowsAsItIsTypedInto(t *testing.T) {
	f := newFake()
	w := New(f, 1200, 800)
	w.snap = reading.Snapshot{
		Cards: []reading.Card{{Number: 1, Title: "Write the spec", State: "todo"}},
		Goals: []reading.Goal{
			{Number: 1, Title: "Modularité / Helm"},
			{Number: 2, Title: "Stabiliser les environnements"},
		},
	}
	w = w.clicked(middleOf(w, 1, 0)).Choose("Edit")
	for w.Form().Fields()[w.Form().On()].Arg != "objective" {
		w = w.Key("tab")
	}
	w = w.Key("enter")
	if got := w.Choosing(); len(got) != 2 {
		t.Fatalf("the list offers %d objectives, want both", len(got))
	}
	narrowed := w.Types("helm")
	if got := narrowed.Choosing(); len(got) != 1 || got[0].Number != 1 {
		t.Errorf("narrowed to \"helm\", the list offers %v", goalTitles(got))
	}
	// And it does not care how it was typed, the way looking does not.
	if got := w.Types("HELM").Choosing(); len(got) != 1 {
		t.Errorf("narrowed to \"HELM\", the list offers %v", goalTitles(got))
	}
}

// Choosing out of an empty list does nothing rather than reaching past its end.
func TestChoosingOutOfAnEmptyListDoesNothing(t *testing.T) {
	f := newFake()
	w := New(f, 1200, 800)
	w.snap = reading.Snapshot{Cards: []reading.Card{{Number: 1, Title: "Write the spec", State: "todo"}}}
	w = w.clicked(middleOf(w, 1, 0)).Choose("Edit")
	for w.Form().Fields()[w.Form().On()].Arg != "objective" {
		w = w.Key("tab")
	}
	w = w.Key("enter")
	if len(w.Choosing()) != 0 {
		t.Fatalf("the store has no objectives and the list offers %d", len(w.Choosing()))
	}
	if got := w.Choose("Take"); got.Screen() != OnChoosing {
		t.Errorf("choosing out of an empty list left the window on %v", got.Screen())
	}
}

// A click lands on the row it is over, and not always on the first one. The
// walk down the rows is arithmetic, and arithmetic only exercised at nought
// answers the same whatever it does.
func TestAClickLandsOnTheRowItIsOver(t *testing.T) {
	f := newFake()
	w := New(f, 1200, 800)
	w.snap = reading.Snapshot{Cards: []reading.Card{
		{Number: 1, Title: "Ship it", State: "todo"},
		{Number: 2, Title: "Write the spec", State: "todo"},
		{Number: 3, Title: "Rewrite the tests", State: "todo"},
	}}
	w = w.clicked(middleOf(w, 1, 1)).Choose("Edit")
	for w.Form().Fields()[w.Form().On()].Arg != "parent" {
		w = w.Key("tab")
	}
	w = w.Key("enter")
	offered := w.Choosing()
	if len(offered) != 2 {
		t.Fatalf("the list offers %v, want the two tasks this one may hang under", goalTitles(offered))
	}

	box := ChoosingOver(w.width, w.height)
	for at, one := range offered {
		row := ChoosingRow(box, ChoosingFrom(box)+at*apart)
		// A click says which row; space takes it.
		got := w.clicked(Point{X: row.X + row.W/2, Y: row.Y + row.H/2}).Key("enter")
		if got.Form().ValueOf("parent") != strconv.Itoa(one.Number) {
			t.Errorf("clicking row %d took %q, want task %d (%s)",
				at, got.Form().ValueOf("parent"), one.Number, one.Title)
		}
	}
	// And a point below the last row lands on no row at all.
	below := ChoosingRow(box, ChoosingFrom(box)+len(offered)*apart)
	if _, on := ChoosingAtPoint(box, len(offered), 0, Point{X: below.X + 10, Y: below.Y + below.H/2}); on {
		t.Error("a point past the last row was read as a row")
	}
}

// Walking down a long list brings each row into sight, and brings in exactly
// the one asked for: the run rarely divides into whole rows, and an answer
// rounded the other way leaves the row the keyboard is on half under the edge.
func TestWalkingALongListBringsInExactlyTheRowAskedFor(t *testing.T) {
	f := newFake()
	w := New(f, 1200, 800)
	cards := []reading.Card{{Number: 1, Title: "Write the spec", State: "todo"}}
	for i := 2; i <= 45; i++ {
		cards = append(cards, reading.Card{Number: i, Title: "Une de plus", State: "backlog"})
	}
	w.snap = reading.Snapshot{Cards: cards}
	w = w.clicked(middleOf(w, 1, 0)).Choose("Edit")
	for w.Form().Fields()[w.Form().On()].Arg != "parent" {
		w = w.Key("tab")
	}
	w = w.Key("enter")
	room := w.ChoosingRoom()

	for at := range len(w.Choosing()) {
		if at > 0 {
			w = w.Key("down")
		}
		from := w.ChoosingFromRow()
		if at < room {
			if from != 0 {
				t.Fatalf("on row %d of %d in sight, the list moved to %d", at, room, from)
			}
			continue
		}
		if want := at - room + 1; from != want {
			t.Fatalf("on row %d, the list starts at %d, want %d", at, from, want)
		}
	}
}

// A click lands on the row it is over after the list has been scrolled, not on
// the one that was there before.
func TestAClickAfterScrollingLandsOnTheRowItIsOver(t *testing.T) {
	f := newFake()
	w := New(f, 1200, 800)
	cards := []reading.Card{{Number: 1, Title: "Write the spec", State: "todo"}}
	for i := 2; i <= 45; i++ {
		cards = append(cards, reading.Card{Number: i, Title: "Une de plus", State: "backlog"})
	}
	w.snap = reading.Snapshot{Cards: cards}
	w = w.clicked(middleOf(w, 1, 0)).Choose("Edit")
	for w.Form().Fields()[w.Form().On()].Arg != "parent" {
		w = w.Key("tab")
	}
	w = w.Key("enter")
	box := ChoosingOver(w.width, w.height)
	run := ChoosingRun(box)
	w = w.Scroll(Point{X: run.X + run.W/2, Y: run.Y + run.H/2}, 6*lineHigh)
	if w.ChoosingFromRow() != 6 {
		t.Fatalf("the list starts at row %d, want 6", w.ChoosingFromRow())
	}

	// The second row drawn is the eighth of the list once six are above it.
	row := ChoosingRow(box, ChoosingFrom(box)+apart)
	got := w.clicked(Point{X: row.X + row.W/2, Y: row.Y + row.H/2}).Key("enter")
	want := w.Choosing()[7].Number
	if got.Form().ValueOf("parent") != strconv.Itoa(want) {
		t.Errorf("clicking the second row drawn took %q, want task %d",
			got.Form().ValueOf("parent"), want)
	}
}

// Shift+enter writes a line break where there is room for one, and nowhere
// else. A title is one line whatever is typed into it, and a break written
// into it would be a character nothing can draw and nobody asked for.
func TestShiftEnterWritesALineBreakOnlyWhereThereIsRoom(t *testing.T) {
	f := newFake()
	w := holding(f, reading.Card{Number: 1, Title: "Write the spec", State: "todo"})
	w = w.clicked(middleOf(w, 1, 0)).Choose("Edit")

	on := w.On("title").Key("shift+enter")
	if got := on.Form().ValueOf("title"); got != "Write the spec" {
		t.Errorf("the title reads %q, and a title is one line", got)
	}
	long := w.On("description").Types("un").Key("shift+enter").Types("deux")
	if got := long.Form().ValueOf("description"); got != "un\ndeux" {
		t.Errorf("the description reads %q", got)
	}
}

// What a list of tasks offers, told something to look for and told nothing.
// The two answers are different sets, and the line between them is the empty
// filter: told nothing, a list offers everything rather than nothing.
func TestChoices_offersEverythingUntilItIsToldWhatToLookFor(t *testing.T) {
	w := New(newFake(), 1200, 800)
	w.snap = reading.Snapshot{Cards: []reading.Card{
		{Number: 1, Title: "Write the spec", State: "todo"},
		{Number: 2, Title: "Ship it", State: "current"},
		{Number: 3, Title: "Rewrite the tests", State: "todo"},
	}}
	w = w.clicked(middleOf(w, 1, 0)).Choose("Edit")
	for w.Form().Fields()[w.Form().On()].Arg != "parent" {
		w = w.Key("tab")
	}
	w = w.Key("enter")
	if got := w.Choosing(); len(got) != 2 {
		t.Fatalf("told nothing, the list offers %v, want both of the others", goalTitles(got))
	}
	if got := w.Types("ship").Choosing(); len(got) != 1 || got[0].Number != 2 {
		t.Errorf("told \"ship\", the list offers %v", goalTitles(got))
	}
	if got := w.Types("zzz").Choosing(); len(got) != 0 {
		t.Errorf("told something nothing carries, the list offers %v", goalTitles(got))
	}
}

// Both surfaces guard a deletion of two words and neither guarded ten minutes
// of typing: Cancel emptied a filled form without a question.
func TestLeavingAFilledFormAsksFirst(t *testing.T) {
	w := holding(newFake(), reading.Card{Number: 1, Title: "Write the spec", State: "todo"})
	w = w.clicked(middleOf(w, 1, 0)).Choose("Edit")
	if w.Screen() != OnForm {
		t.Fatalf("the form did not open: %v", w.Screen())
	}
	// Nothing typed: there is nothing to lose, and a question would be noise.
	if left := w.Choose("Cancel"); left.Screen() != OnTheBoard {
		t.Fatalf("an untouched form asked about being closed: %v", left.Screen())
	}
	w = w.Types("x")
	asked := w.Choose("Cancel")
	if asked.Screen() != OnAsking {
		t.Fatalf("a filled form was thrown away without a question: %v", asked.Screen())
	}
	if !strings.Contains(Heads(asked), "typed") {
		t.Errorf("the question does not say what is being lost: %q", Heads(asked))
	}
	if back := asked.Choose("Back"); back.Screen() != OnForm {
		t.Fatalf("backing out of the question did not go back to the form: %v", back.Screen())
	}
	if gone := asked.Choose("Agree"); gone.Screen() != OnTheBoard {
		t.Fatalf("agreeing to lose it left %v", gone.Screen())
	}
}

// into() takes the box the text is drawn in, and went on subtracting a card's
// padding from it as well: the caret landed a glyph and a half from the click.
func TestClickingIntoAFieldLandsWhereItWasClicked(t *testing.T) {
	box := Box{X: 100, Y: 0, W: 200, H: lineHigh}
	if got := into(box, Point{X: box.X}, "abcdef", Reckoned); got != 0 {
		t.Errorf("a click on the first character lands at %d", got)
	}
	if got := into(box, Point{X: box.X + 3*perGlyph}, "abcdef", Reckoned); got != 3 {
		t.Errorf("a click three glyphs in lands at %d", got)
	}
}

// A key does what the bar offers, where the bar offers it. The bar offers Save
// under ⌘⏎ and enter saved as well, which is a key doing what the bar does not
// say — and it is why writing a description in paragraphs needed ⇧⏎.
func TestEnterDoesNotSaveTheForm(t *testing.T) {
	f := newFake()
	w := holding(f, reading.Card{Number: 1, Title: "Write the spec", State: "todo"})
	w = w.clicked(middleOf(w, 1, 0)).Choose("Edit").Types("x")

	if after := w.Key("enter"); after.Screen() != OnForm {
		t.Fatalf("enter left the form: %v", after.Screen())
	}
	if len(f.wrote()) != 0 {
		t.Fatalf("enter wrote to the store: %v", f.wrote())
	}
	if after := w.Key("cmd+enter"); after.Screen() == OnForm {
		t.Fatal("⌘⏎ did not save the form")
	}
}

// In a field with room for several lines, enter is what a line break always
// was everywhere else: the one key a hand reaches for.
func TestShiftEnterIsALineBreakWithoutLeavingTheBox(t *testing.T) {
	f := newFake()
	w := holding(f, reading.Card{Number: 1, Title: "Write the spec", State: "todo"})
	w = w.clicked(middleOf(w, 1, 0)).Choose("Edit")
	for i, field := range w.Form().Fields() {
		if field.Name == "description" {
			w.form = w.form.onto(i)
		}
	}
	w = w.Types("Deux colonnes").Key("shift+enter").Types("et une bande")
	if said := w.Form().Writing().Reads(); said != "Deux colonnes\net une bande" {
		t.Fatalf("the description reads %q", said)
	}
	// A field of one line is not a place for a break: ⇧⏎ there moves on, the
	// way ⏎ does, and writes nothing into what it leaves.
	w.form = w.form.onto(0)
	before := w.Form().ValueOf("title")
	if after := w.Key("shift+enter"); after.Form().ValueOf("title") != before || after.Form().On() != 1 {
		t.Fatalf("⇧⏎ on the title left it %q and moved to field %d",
			after.Form().ValueOf("title"), after.Form().On())
	}
}

func TestThePartsOfATaskAreEditedInItsForm(t *testing.T) {
	f := newFake()
	w := holding(f,
		reading.Card{Number: 1, Title: "Ship it", State: "todo"},
		reading.Card{Number: 2, Title: "Relire l'ADR", State: "todo", Parent: 1},
		reading.Card{Number: 3, Title: "Trancher", State: "done", Parent: 1})
	w = w.clicked(middleOf(w, 1, 0)).Choose("Edit")

	if said := w.Form().ValueOf("parts"); said != "#2 [ ] Relire l'ADR\n#3 [x] Trancher" {
		t.Fatalf("the form opened holding %q", said)
	}

	// Rename one, tick it, drop the other, and name a new one.
	w.form = w.form.Set("parts", "#2 [x] Relire l'ADR 0014\n[ ] Rédiger la section Flux")
	w = w.Save()

	var added, edited, deleted []engine.Input
	for _, one := range f.asked {
		switch {
		case one.op.Thing == "task" && one.op.Verb == "add":
			added = append(added, one.in)
		case one.op.Thing == "task" && one.op.Verb == "edit" && one.in.Given["number"] != "1":
			edited = append(edited, one.in)
		case one.op.Thing == "task" && one.op.Verb == "delete":
			deleted = append(deleted, one.in)
		}
	}
	if len(added) != 1 || added[0].Given["title"] != "Rédiger la section Flux" ||
		added[0].Given["parent"] != "1" {
		t.Errorf("what was added: %v", added)
	}
	if len(edited) != 1 || edited[0].Given["title"] != "Relire l'ADR 0014" ||
		edited[0].Given["state"] != "done" {
		t.Errorf("what was changed: %v", edited)
	}
	if len(deleted) != 1 || deleted[0].Given["number"] != "3" {
		t.Errorf("what was taken away: %v", deleted)
	}
}

// A task with nothing under it opens on an empty field rather than none: it is
// where a first part is written.
func TestATaskWithNoPartsStillHasTheField(t *testing.T) {
	w := holding(newFake(), reading.Card{Number: 1, Title: "Ship it", State: "todo"})
	w = w.clicked(middleOf(w, 1, 0)).Choose("Edit")
	if said := w.Form().ValueOf("parts"); said != "" {
		t.Fatalf("a task with nothing under it opened holding %q", said)
	}
}

// A part the store will not let go of leaves the window saying so, rather than
// closing the form as though the list had been settled.
func TestAPartTheStoreWillNotDeleteIsSaid(t *testing.T) {
	f := newFake()
	w := holding(f,
		reading.Card{Number: 1, Title: "Ship it", State: "todo"},
		reading.Card{Number: 2, Title: "Relire l'ADR", State: "todo", Parent: 1})
	w = w.clicked(middleOf(w, 1, 0)).Choose("Edit")
	w.form = w.form.Set("parts", "")
	// Only the deletion: the task's own edit has to go through for the part to
	// be reached at all.
	f.refuseDelete = errRefuse
	w = w.Save()
	if w.Wrong() == nil {
		t.Fatal("a part that could not be deleted was passed over")
	}
}

// One rule for a form, on both surfaces: enter writes a line where a line can
// be written and moves on everywhere else, and it never saves. Saving is a
// chord of its own, so it cannot happen by reaching for the key that breaks a
// line.
func TestEnterMovesOnOrBreaksALineAndNeverSaves(t *testing.T) {
	f := newFake()
	w := holding(f, reading.Card{Number: 1, Title: "Write the spec", State: "todo"})
	w = w.clicked(middleOf(w, 1, 0)).Choose("Edit")

	// A one-line field: on to the next one.
	w.form = w.form.onto(0)
	if after := w.Key("enter"); after.Form().On() != 1 || after.Screen() != OnForm {
		t.Fatalf("enter on the title went to field %d on %v", after.Form().On(), after.Screen())
	}
	// The description: a line, and the form stays where it is.
	for i, field := range w.Form().Fields() {
		if field.Name == "description" {
			w.form = w.form.onto(i)
		}
	}
	w = w.Types("Deux colonnes").Key("enter").Types("et une bande")
	if said := w.Form().Writing().Reads(); said != "Deux colonnes\net une bande" {
		t.Fatalf("the description reads %q", said)
	}
	if w.Screen() != OnForm || len(f.wrote()) != 0 {
		t.Fatalf("enter left the form on %v having asked %v", w.Screen(), f.asked)
	}
	// Both chords save, from anywhere in it.
	for _, chord := range []string{"cmd+enter", "cmd+s"} {
		if after := w.Key(chord); after.Screen() == OnForm {
			t.Errorf("%s did not save the form", chord)
		}
	}
}

// Inside the description the arrows walk its lines, so every one of them can
// be reached — down used to leave for the next field, and the only way through
// a long text was a character at a time. Tab is what leaves.
func TestTheArrowsWalkTheLinesOfTheDescription(t *testing.T) {
	w := holding(newFake(), reading.Card{Number: 1, Title: "Write the spec", State: "todo",
		Description: "one\ntwo\nthree"})
	w = w.clicked(middleOf(w, 1, 0)).Choose("Edit")
	at := -1
	for i, field := range w.Form().Fields() {
		if field.Name == "description" {
			at = i
		}
	}
	w.form = w.form.onto(at)
	w.form = w.form.Writes(w.form.Writing().CaretAt(0))
	w = w.Key("down")
	if w.Form().On() != at {
		t.Fatalf("down left the description for field %d", w.Form().On())
	}
	if got := w.Form().Writing().At; got != 4 {
		t.Fatalf("down from the start of the first line put the caret at %d, want 4", got)
	}
	if got := w.Key("down").Form().Writing().At; got != 8 {
		t.Fatalf("down again put the caret at %d, want 8", got)
	}
	if got := w.Key("down").Key("up").Form().Writing().At; got != 4 {
		t.Fatalf("up from the last line put the caret at %d, want 4", got)
	}
	if got := w.Key("tab").Form().On(); got != at+1 {
		t.Fatalf("tab out of the description went to field %d", got)
	}
	// Past the last line it stays where it is: leaving is what tab is for.
	if got := w.Key("down").Key("down").Key("down").Form().Writing().At; got != 8 {
		t.Fatalf("down past the last line put the caret at %d, want it kept at 8", got)
	}
	// Into a shorter line, as far in as that line goes and no further: four
	// letters into "three", up onto "two", lands at the end of "two".
	w.form = w.form.Writes(w.form.Writing().CaretAt(12))
	if got := w.Key("up").Form().Writing().At; got != 7 {
		t.Fatalf("up from four letters into the last line put the caret at %d, want 7", got)
	}
	// And as far in as it was, onto a line long enough to take it.
	w.form = w.form.Writes(w.form.Writing().CaretAt(6))
	if got := w.Key("down").Form().Writing().At; got != 10 {
		t.Fatalf("down from two letters into the second line put the caret at %d, want 10", got)
	}
}

// The description in the form scrolls, carries the bar the detail's carries,
// and its end can be reached — the same box on both screens.
func TestTheDescriptionInTheFormScrollsToItsEnd(t *testing.T) {
	text := strings.Repeat("une phrase de description assez longue pour replier ", 60) + "LA FIN"
	w := holding(newFake(), reading.Card{Number: 1, Title: "Write the spec", State: "todo",
		Description: text})
	w = w.clicked(middleOf(w, 1, 0)).Choose("Edit")
	for i, field := range w.Form().Fields() {
		if field.Name == "description" {
			w.form = w.form.onto(i)
		}
	}
	frame, bar, lines, rows, ok := w.sayingOn()
	if !ok {
		t.Fatal("the description in the form has no box of its own")
	}
	if bar.X < frame.X+frame.W {
		t.Fatalf("the bar at %d sits over the text, which ends at %d", bar.X, frame.X+frame.W)
	}
	// The wheel takes it to the end, and the end is the text's last line.
	for range len(lines.Said) {
		w = w.Scroll(Point{X: frame.X + 4, Y: frame.Y + 4}, lineHigh)
	}
	if last := lines.Said[min(len(lines.Said)-1, w.saying+rows-1)]; !strings.HasSuffix(last, "FIN") {
		t.Fatalf("scrolled to the end, the last line in sight is %q", last)
	}
	// And the wheel over it does not move the keyboard off it.
	if w.Form().Fields()[w.Form().On()].Name != "description" {
		t.Fatal("scrolling the description put the keyboard somewhere else")
	}
}

// What is in sight follows the caret: arriving at the description shows where
// the caret is, and typing after the wheel moved the box away brings it back.
// The wheel is left alone otherwise — that is what reading back through a text
// is.
func TestTheDescriptionFollowsTheCaretWhenItMoves(t *testing.T) {
	text := strings.Repeat("une phrase de description assez longue pour replier ", 60)
	w := holding(newFake(), reading.Card{Number: 1, Title: "Write the spec", State: "todo",
		Description: text})
	w = w.clicked(middleOf(w, 1, 0)).Choose("Edit")
	w.form = w.form.onto(0)
	w = w.Key("tab") // onto the description, with the caret at its end
	if w.Form().Fields()[w.Form().On()].Name != "description" {
		t.Fatalf("tab went to %q", w.Form().Fields()[w.Form().On()].Name)
	}
	_, _, lines, rows, _ := w.sayingOn()
	end := len(lines.Said) - rows
	if w.saying != end {
		t.Fatalf("arriving at the description shows line %d, want the caret's at %d", w.saying, end)
	}
	frame, _, _, _, _ := w.sayingOn()
	for range len(lines.Said) {
		w = w.Scroll(Point{X: frame.X + 4, Y: frame.Y + 4}, -lineHigh)
	}
	if w.saying != 0 {
		t.Fatalf("the wheel left the description at %d, want its beginning", w.saying)
	}
	if w = w.Types("x"); w.saying != end {
		t.Fatalf("typing at the end left the description at %d, want %d", w.saying, end)
	}
}

// The bar beside the description in the form is pulled the way the detail's
// is, and a press on it is not a press in the text.
func TestTheBarBesideTheDescriptionInTheFormCanBePulled(t *testing.T) {
	text := strings.Repeat("une phrase de description assez longue pour replier ", 60) + "LA FIN"
	w := holding(newFake(), reading.Card{Number: 1, Title: "Write the spec", State: "todo",
		Description: text})
	w = w.clicked(middleOf(w, 1, 0)).Choose("Edit")
	for i, field := range w.Form().Fields() {
		if field.Name == "description" {
			w.form = w.form.onto(i)
		}
	}
	w.form = w.form.Writes(w.form.Writing().CaretAt(0))
	w = w.inSightOfTheCaret()
	_, bar, lines, rows, _ := w.sayingOn()
	before := w.Form().Writing().At
	w = w.PressAt(Point{X: bar.X + bar.W/2, Y: bar.Y + 1}).
		MoveTo(Point{X: bar.X + bar.W/2, Y: bar.Y + bar.H - 1})
	if w.saying != len(lines.Said)-rows {
		t.Fatalf("pulling the bar to the bottom left the description at %d, want %d",
			w.saying, len(lines.Said)-rows)
	}
	if w.Form().Writing().At != before {
		t.Fatalf("a press on the bar moved the caret from %d to %d", before, w.Form().Writing().At)
	}
}

// A row of a list says whether it is taken, and a task or an objective by its
// number too; a column is known by its name alone.
func TestARowSaysWhatItIsKnownBy(t *testing.T) {
	f := newFake()
	w := holding(f, card(1, "Write the spec", "todo"))
	w = w.clicked(middleOf(w, 1, 0)).Choose("Edit")
	for i, field := range w.Form().Fields() {
		if field.Name == "column" {
			w.form = w.form.onto(i)
		}
	}
	w = w.choosingList()
	got := []string{w.RowSays(w.Choosing()[0]), w.RowSays(w.Choosing()[1])}
	if want := []string{"   backlog", TickedMark + "  todo"}; got[0] != want[0] || got[1] != want[1] {
		t.Errorf("the column list says %q, want %q", got, want)
	}
	w = w.leaving()
	for i, field := range w.Form().Fields() {
		if field.Name == "parent" {
			w.form = w.form.onto(i)
		}
	}
	w = w.choosingList()
	if got := w.RowSays(Choice{Number: 7, Title: "Ship it"}); got != "   7  Ship it" {
		t.Errorf("a task on the parent list would say %q", got)
	}
}

// A task added in the window starts on the column of what is chosen, at normal
// priority — the way the terminal starts one. It used to start with neither.
func TestANewTaskStartsOnTheColumnOfWhatIsChosen(t *testing.T) {
	f := newFake()
	w := holding(f, reading.Card{Number: 1, Title: "In todo", State: "todo"})
	w = w.clicked(middleOf(w, 1, 0))
	form := adding(Tasks, w.addingOn())
	if got := form.ValueOf("column"); got != "todo" {
		t.Errorf("a task added on todo starts in %q", got)
	}
	if got := form.ValueOf("priority"); got != "normal" {
		t.Errorf("a new task starts at %q, want normal", got)
	}
	if got := adding(Tasks, holding(newFake()).addingOn()).ValueOf("column"); got != "backlog" {
		t.Errorf("a task added with nothing chosen starts in %q, want backlog", got)
	}
	band := holding(newFake(),
		reading.Card{Number: 1, Title: "In todo", State: "todo"},
		reading.Card{Number: 2, Title: "Blocked", State: "blocked"})
	band = band.clicked(middleOf(band, 1, 0)).Key("down")
	if got := adding(Tasks, band.addingOn()).ValueOf("column"); got != "blocked" {
		t.Errorf("a task added on the band starts in %q, want the first stopped state", got)
	}
}

// The parts of a task being edited are drawn as tiles to tick, read from what
// the field says now — a part typed and not yet saved among them, under its
// title alone — wherever the keyboard is.
func TestThePartsAreTicksWhereverTheKeyboardIs(t *testing.T) {
	f := newFake()
	w := holding(f, card(1, "Ship it", "todo"),
		reading.Card{Number: 2, Title: "Draft the columns", State: "done", Parent: 1})
	w.chosen = 1
	w = w.Edit()
	fields := w.form.Fields()
	at := len(fields) - 1
	w.form.fields[at].Value += "\nGet the nod"
	got := w.PartsTicked()
	if len(got) != 2 || got[0].Number != 2 || !reading.Finished(got[0]) || got[1].Title != "Get the nod" || reading.Finished(got[1]) {
		t.Fatalf("the parts are drawn as %+v", got)
	}
	if line := SubLine(got[1]); line != "Get the nod" {
		t.Errorf("a part not yet saved reads %q", line)
	}
	w.form = w.form.onto(at)
	if again := w.PartsTicked(); len(again) != 2 {
		t.Errorf("with the keyboard on them the parts are drawn as %+v", again)
	}
}

// The description is wrapped with the face it is drawn in, wherever its lines
// are counted: in the form and on the open task alike. Wrapped by a reckoned
// width, a line of narrow letters stopped two thirds of the way across its box.
func TestADescriptionIsWrappedWithTheFaceItIsDrawnIn(t *testing.T) {
	text := strings.Repeat("iiii ", 120)
	w := holding(newFake(), reading.Card{Number: 1, Title: "Write the spec", State: "todo", Description: text}).Measuring(narrow)
	w = w.clicked(middleOf(w, 1, 0))
	for _, open := range []Window{w.DoubleClick(middleOf(w, 1, 0)), w.Choose("Edit").On("description")} {
		frame, _, lines, _, ok := open.sayingOn()
		if !ok {
			t.Fatalf("on %v the description has no box", open.Screen())
		}
		wide := Inside(frame).W - padding - padding
		reckoned := Wrapped(text, room(Inside(frame)))
		if len(lines.Said) >= len(reckoned.Said) {
			t.Errorf("on %v the description takes %d lines, as many as the reckoning's %d", open.Screen(), len(lines.Said), len(reckoned.Said))
		}
		for _, line := range lines.Said {
			if edges := narrow(line, reads); edges[len(edges)-1] > wide {
				t.Errorf("on %v %q is %d wide in a room of %d", open.Screen(), line, edges[len(edges)-1], wide)
			}
		}
	}
}

// The history is wrapped with the face it is drawn in, the way the description
// is: an act of narrow letters runs across the whole of its box.
func TestTheHistoryIsWrappedWithTheFaceItIsDrawnIn(t *testing.T) {
	act := strings.Repeat("iiii ", 120)
	w := holding(newFake(), card(1, "Write the spec", "todo")).Measuring(narrow)
	w.chosen, w.screen, w.acts = 1, OnForm, []string{act}
	reckoned := Wrapped(act, room(Inside(BesideBox(Over(w.width, w.height)))))
	if got := w.Lines(); len(got) >= len(reckoned.Said) {
		t.Errorf("the history takes %d lines, as many as the reckoning's %d", len(got), len(reckoned.Said))
	}
}

// What a form set aside is offered back on the form it came from, and on no
// other: not on another task's, not on a new one, and not once something has
// been typed into the form since, which bringing it back would write over.
func TestWhatWasSetAsideIsOfferedOnlyWhereItCameFrom(t *testing.T) {
	w := holding(newFake(), card(1, "Write the spec", "todo"), card(2, "Ship it", "todo"))
	w = w.clicked(middleOf(w, 1, 0)).Choose("Edit").Types(" again").Key("esc").Key("esc")
	if w.Screen() != OnTheBoard {
		t.Fatalf("esc twice left the window on %v", w.Screen())
	}
	offered := func(v Window) bool { return slices.Contains(v.Offers(), "Recover") }
	if again := w.Choose("Edit"); !offered(again) {
		t.Errorf("the form it came from offers %v", again.Offers())
	} else if typed := again.Types("x"); offered(typed) {
		t.Errorf("a form typed into since still offers %v", typed.Offers())
	}
	if other := w.clicked(middleOf(w, 1, 1)).Choose("Edit"); offered(other) {
		t.Errorf("another task's form offers %v", other.Offers())
	}
	if fresh := w.Choose("New"); offered(fresh) {
		t.Errorf("a new task's form offers %v", fresh.Offers())
	}
	// Answered by a click off it, the question sets aside the same way.
	clickedOff := w.Choose("Edit").Types(" once more").Key("esc").clicked(Point{X: 1, Y: 1})
	recovered := clickedOff.Choose("Edit").Key("cmd+z").Form().Fields()[0].Value
	if clickedOff.Screen() != OnTheBoard || !strings.Contains(recovered, "once more") {
		t.Errorf("a click off the question left %v, and the form brought back %q", clickedOff.Screen(), recovered)
	}
	if says := w.Choose("Edit").Types("x").Key("esc").Asks(); len(says) != 1 || !strings.Contains(says[0].Means, "brings it back") {
		t.Errorf("the question about leaving says %+v", says)
	}
}

// The description shows where the wheel left it while the keyboard is on
// another field, and every other long field shows its beginning.
func TestTheDescriptionIsShownWhereTheWheelLeftIt(t *testing.T) {
	text := strings.Repeat("une phrase de description assez longue pour replier ", 60)
	w := holding(newFake(), reading.Card{Number: 1, Title: "Ship it", State: "todo", Description: text},
		reading.Card{Number: 2, Title: "Part", State: "todo", Parent: 1})
	w = w.clicked(middleOf(w, 1, 0)).DoubleClick(middleOf(w, 1, 0))
	frame, _, _, _, _ := w.sayingOn()
	w = w.Scroll(Point{X: frame.X + 4, Y: frame.Y + 4}, 3*lineHigh)
	fields := w.Form().Fields()
	if w.Saying() == 0 || w.scrolledOf(1) != w.Saying() || w.scrolledOf(len(fields)-1) != 0 || w.scrolledOf(0) != Follow {
		t.Errorf("scrolled to %d, the fields show from %d, %d and %d",
			w.Saying(), w.scrolledOf(0), w.scrolledOf(1), w.scrolledOf(len(fields)-1))
	}
}

// Beside an open task, a triple click holds the whole of a line of history,
// the way it holds a paragraph in a field.
func TestATripleClickBesideTheFormHoldsTheWholeLine(t *testing.T) {
	w, _ := openTask(t)
	r := readable(t, w, "history:0")
	if got, _ := w.tripleClicked(on(r, 3)).Key("cmd+c").Copying(); got != r.Said {
		t.Errorf("a triple click on a line of history copied %q, want %q", got, r.Said)
	}
}

// Beside an open objective, the work serving it is held and copied like any
// text, and its bar is pulled like the history's.
func TestTheWorkBesideAnObjectiveIsHeldAndItsBarPulled(t *testing.T) {
	var cards []reading.Card
	for i := range 30 {
		cards = append(cards, reading.Card{Number: i + 1, Title: "Task", State: "todo", Objectives: []int{1}})
	}
	w := New(newFake(), 1200, 800)
	w.snap = reading.Snapshot{Cards: cards, Goals: []reading.Goal{{Number: 1, Title: "Ship it", State: "active"}}}
	w = w.Key("tab").clicked(goalMiddle(w.Key("tab"), 1)).Key("enter")
	var parts []string
	for _, r := range w.Readables() {
		if strings.HasPrefix(r.Key, "part:") {
			parts = append(parts, r.Key)
		}
	}
	if len(parts) == 0 || len(parts) >= len(cards) {
		t.Errorf("beside the objective %d of %d tiles can be held", len(parts), len(cards))
	}
	track := BesideBar(w.formPanel())
	middle := Point{X: track.X + track.W/2, Y: track.Y + 4}
	far := Point{X: middle.X, Y: track.Y + track.H + 999}
	if pulled := w.PressAt(middle).MoveTo(far).LetGo(far); pulled.ServingAt() == 0 {
		t.Error("pulling the bar beside the work serving an objective moved nothing")
	}
}
