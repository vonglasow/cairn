package gui

import (
	"cairn/internal/board"
	"cairn/internal/gui/design"
	"cairn/internal/gui/geom"
	"slices"
	"strings"
	"testing"

	"cairn/internal/engine"
	"cairn/test"
)

func TestAddingSavesWithoutANumber(t *testing.T) {
	f := newFake()
	w := New(f, 1200, 800)

	w = w.Key("cmd+n")
	if w.Screen != board.OnForm || !w.Form.Adding {
		t.Fatalf("cmd+n left the window on %v", w.Screen)
	}
	w.Form = w.Form.Set("title", "Write the spec")
	w = w.Choose("Save")

	if f.what("task", "new") != 1 {
		t.Fatalf("adding asked for %d adds", f.what("task", "new"))
	}
	in, _ := f.gave("task", "new")
	if _, named := in.Given["number"]; named {
		t.Error("a task being added was given a number it does not have yet")
	}
}

func TestTheButtonsDoNothingWithNoFormOpen(t *testing.T) {
	f := newFake()
	w := holding(f, board.Card{Number: 1, Title: "Write the spec", State: "todo"})
	for _, name := range []string{"Save", "Cancel", "something nobody offers"} {
		if got := w.Choose(name); got.Screen != board.OnTheBoard {
			t.Errorf("%q left the window on %v", name, got.Screen)
		}
	}
	if len(f.wrote()) != 0 {
		t.Errorf("the store was asked %d things", len(f.wrote()))
	}
}

func TestASaveTheStoreRefusesKeepsTheForm(t *testing.T) {
	f := newFake()
	f.failOp = errRefuse
	w := holding(f, board.Card{Number: 1, Title: "Write the spec", State: "todo"})

	w = w.clicked(middleOf(w, 1, 0))
	w = w.Choose("Edit")
	w = w.Choose("Save")

	if w.Wrong() == nil {
		t.Fatal("a refused save said nothing")
	}
	if w.Screen != board.OnForm {
		t.Fatalf("a refused save closed the form and lost what was typed")
	}
}

func TestTabWalksTheFieldsAndShiftTabWalksBack(t *testing.T) {
	f := newFake()
	w := holding(f, board.Card{Number: 1, Title: "Write the spec", State: "todo"})
	w = w.clicked(middleOf(w, 1, 0)).Choose("Edit").On("title")

	w = w.Key("tab")
	if got := w.Form.Fields[w.Form.On()].Name; got != "description" {
		t.Fatalf("tab landed on %q", got)
	}
	w = w.Key("shift+tab")
	if got := w.Form.Fields[w.Form.On()].Name; got != "title" {
		t.Fatalf("shift+tab landed on %q", got)
	}
	if got := w.Key("down").Form.On(); got != 1 {
		t.Errorf("down landed on field %d", got)
	}
	if got := w.Key("up").Form.On(); got != 0 {
		t.Errorf("up from the first field landed on %d", got)
	}
}

func TestAFieldWithAClosedSetIsCycledRatherThanTyped(t *testing.T) {
	f := newFake()
	w := holding(f, board.Card{Number: 1, Title: "Write the spec", State: "backlog"})
	w = w.clicked(middleOf(w, 0, 0)).Choose("Edit").On("column")

	if got := valueOf(w.Key("right").Form, "column"); got != "todo" {
		t.Errorf("right of backlog is %q", got)
	}
	if got := valueOf(w.Key("left").Form, "column"); got != "backlog" {
		t.Errorf("left of backlog is %q, and there is nothing before it", got)
	}
	last := w.On("column")
	for range 10 {
		last = last.Key("right")
	}
	if got := valueOf(last.Form, "column"); got != "done" {
		t.Errorf("past the last of the set is %q", got)
	}
	typed := w.On("title").Key("right")
	if got := valueOf(typed.Form, "title"); got != "Write the spec" {
		t.Errorf("right in a typed field changed it to %q", got)
	}
}

func TestChoosingOutOfAnEmptyListDoesNothing(t *testing.T) {
	f := newFake()
	w := New(f, 1200, 800)
	w.Snap = test.Searchable(board.Snapshot{Cards: []board.Card{{Number: 1, Title: "Write the spec", State: "todo"}}})
	w = w.clicked(middleOf(w, 1, 0)).Choose("Edit")
	for w.Form.Fields[w.Form.On()].Arg != "objective" {
		w = w.Key("tab")
	}
	w = w.Key("enter")
	if len(w.Choosing()) != 0 {
		t.Fatalf("the store has no objectives and the list offers %d", len(w.Choosing()))
	}
	if got := w.Choose("Take"); got.Screen != board.OnChoosing {
		t.Errorf("choosing out of an empty list left the window on %v", got.Screen)
	}
}

func onTheObjectives(w Window) Window {
	w = w.clicked(middleOf(w, 1, 0)).Choose("Edit")
	for w.Form.Fields[w.Form.On()].Arg != "objective" {
		w = w.Key("tab")
	}
	return w.Key("enter")
}

func manyGoals(n int) []board.Goal {
	goals := make([]board.Goal, 0, n)
	for i := 1; i <= n; i++ {
		goals = append(goals, board.Goal{Number: i, Title: "One more", State: "active"})
	}
	return goals
}

func TestAClickAfterScrollingLandsOnTheRowItIsOver(t *testing.T) {
	f := newFake()
	w := New(f, 1200, 800)
	w.Snap = test.Searchable(board.Snapshot{
		Cards: []board.Card{{Number: 1, Title: "Write the spec", State: "todo"}},
		Goals: manyGoals(44),
	})
	w = onTheObjectives(w)
	box := w.ChoosingBox()
	run := ChoosingRun(box)
	w = w.Scroll(geom.Point{X: run.X + run.W/2, Y: run.Y + run.H/2}, 6*lineHigh)
	if w.ChoosingFromRow() != 6 {
		t.Fatalf("the list starts at row %d, want 6", w.ChoosingFromRow())
	}

	row := ChoosingRow(box, ChoosingFrom(box)+apart)
	if got := w.clicked(geom.Point{X: row.X + row.W/2, Y: row.Y + row.H/2}); got.ChoosingOn != 7 {
		t.Errorf("clicking the second row drawn put the keyboard on row %d, want 7", got.ChoosingOn)
	}
}

func TestLeavingAFilledFormAsksFirst(t *testing.T) {
	w := holding(newFake(), board.Card{Number: 1, Title: "Write the spec", State: "todo"})
	w = w.clicked(middleOf(w, 1, 0)).Choose("Edit")
	if w.Screen != board.OnForm {
		t.Fatalf("the form did not open: %v", w.Screen)
	}
	if left := w.Choose("Cancel"); left.Screen != board.OnTheBoard {
		t.Fatalf("an untouched form asked about being closed: %v", left.Screen)
	}
	w = w.Types("x")
	asked := w.Choose("Cancel")
	if asked.Screen != board.OnAsking {
		t.Fatalf("a filled form was thrown away without a question: %v", asked.Screen)
	}
	if !strings.Contains(Heads(asked), "typed") {
		t.Errorf("the question does not say what is being lost: %q", Heads(asked))
	}
	if back := asked.Choose("Back"); back.Screen != board.OnForm {
		t.Fatalf("backing out of the question did not go back to the form: %v", back.Screen)
	}
	if gone := asked.Choose("Agree"); gone.Screen != board.OnTheBoard {
		t.Fatalf("agreeing to lose it left %v", gone.Screen)
	}
}

func TestEnterMovesOnOrBreaksALineAndNeverSaves(t *testing.T) {
	f := newFake()
	w := holding(f, board.Card{Number: 1, Title: "Write the spec", State: "todo"})
	w = w.clicked(middleOf(w, 1, 0)).Choose("Edit")

	w.Form = w.Form.Onto(0)
	if after := w.Key("enter"); after.Form.On() != 1 || after.Screen != board.OnForm {
		t.Fatalf("enter on the title went to field %d on %v", after.Form.On(), after.Screen)
	}
	for i, field := range w.Form.Fields {
		if field.Name == "description" {
			w.Form = w.Form.Onto(i)
		}
	}
	w = w.Types("Two columns").Key("enter").Types("and a band")
	if said := w.Form.Writing().Reads(); said != "Two columns\nand a band" {
		t.Fatalf("the description reads %q", said)
	}
	if w.Screen != board.OnForm || len(f.wrote()) != 0 {
		t.Fatalf("enter left the form on %v having asked %v", w.Screen, f.asked)
	}
	for _, chord := range []string{"cmd+enter", "cmd+s"} {
		if after := w.Key(chord); after.Screen == board.OnForm {
			t.Errorf("%s did not save the form", chord)
		}
	}
}

func TestTheArrowsWalkTheLinesOfTheDescription(t *testing.T) {
	w := holding(newFake(), board.Card{Number: 1, Title: "Write the spec", State: "todo",
		Description: "one\ntwo\nthree"})
	w = w.clicked(middleOf(w, 1, 0)).Choose("Edit")
	at := -1
	for i, field := range w.Form.Fields {
		if field.Name == "description" {
			at = i
		}
	}
	w.Form = w.Form.Onto(at)
	w.Form = w.Form.Writes(w.Form.Writing().CaretAt(0))
	w = w.Key("down")
	if w.Form.On() != at {
		t.Fatalf("down left the description for field %d", w.Form.On())
	}
	if got := w.Form.Writing().At; got != 4 {
		t.Fatalf("down from the start of the first line put the caret at %d, want 4", got)
	}
	if got := w.Key("down").Form.Writing().At; got != 8 {
		t.Fatalf("down again put the caret at %d, want 8", got)
	}
	if got := w.Key("down").Key("up").Form.Writing().At; got != 4 {
		t.Fatalf("up from the last line put the caret at %d, want 4", got)
	}
	order := FormOrder(w.formBox(), w.Form.Fields)
	if got, want := w.Key("tab").Form.On(), order[slices.Index(order, at)+1]; got != want {
		t.Fatalf("tab out of the description went to field %d, want %d", got, want)
	}
	if got := w.Key("down").Key("down").Key("down").Form.Writing().At; got != 8 {
		t.Fatalf("down past the last line put the caret at %d, want it kept at 8", got)
	}
	w.Form = w.Form.Writes(w.Form.Writing().CaretAt(12))
	if got := w.Key("up").Form.Writing().At; got != 7 {
		t.Fatalf("up from four letters into the last line put the caret at %d, want 7", got)
	}
	w.Form = w.Form.Writes(w.Form.Writing().CaretAt(6))
	if got := w.Key("down").Form.Writing().At; got != 10 {
		t.Fatalf("down from two letters into the second line put the caret at %d, want 10", got)
	}
}

func TestANewTaskStartsOnTheColumnOfWhatIsChosen(t *testing.T) {
	f := newFake()
	w := holding(f, board.Card{Number: 1, Title: "In todo", State: "todo"})
	w = w.clicked(middleOf(w, 1, 0))
	form := Window{}.adding(board.Tasks, w.addingOn())
	if got := valueOf(form, "column"); got != "todo" {
		t.Errorf("a task added on todo starts in %q", got)
	}
	if got := valueOf(form, "priority"); got != "normal" {
		t.Errorf("a new task starts at %q, want normal", got)
	}
	if got := valueOf((Window{}).adding(board.Tasks, holding(newFake()).addingOn()), "column"); got != "backlog" {
		t.Errorf("a task added with nothing chosen starts in %q, want backlog", got)
	}
	band := holding(newFake(),
		board.Card{Number: 1, Title: "In todo", State: "todo"},
		board.Card{Number: 2, Title: "Blocked", State: "blocked"})
	band = band.clicked(middleOf(band, 1, 0)).Key("down")
	if got := valueOf((Window{}).adding(board.Tasks, band.addingOn()), "column"); got != "blocked" {
		t.Errorf("a task added on the band starts in %q, want the first stopped state", got)
	}
}

func TestTheStepsAreTicksWhereverTheKeyboardIs(t *testing.T) {
	f := newFake()
	w := holding(f, board.Card{Number: 1, Title: "Ship it", State: "todo",
		Steps: []engine.Part{{Number: 2, Title: "Draft the columns", Done: true}}})
	w.Chosen = 1
	w = w.Edit()
	fields := w.Form.Fields
	at := len(fields) - 1
	w.Form.Fields[at].Value += "\nGet the nod"
	got := w.PartsTicked()
	if len(got) != 2 || got[0].Number != 2 || !board.Finished(got[0]) || got[1].Title != "Get the nod" || board.Finished(got[1]) {
		t.Fatalf("the steps are drawn as %+v", got)
	}
	if line := SubLine(got[1]); line != "Get the nod" {
		t.Errorf("a step not yet saved reads %q", line)
	}
	w.Form = w.Form.Onto(at)
	if again := w.PartsTicked(); len(again) != 2 {
		t.Errorf("with the keyboard on them the steps are drawn as %+v", again)
	}
}

func TestWhatWasSetAsideIsOfferedOnlyWhereItCameFrom(t *testing.T) {
	w := holding(newFake(), card(1, "Write the spec", "todo"), card(2, "Ship it", "todo"))
	w = w.clicked(middleOf(w, 1, 0)).Choose("Edit").Types(" again").Key("esc").Key("esc")
	if w.Screen != board.OnTheBoard {
		t.Fatalf("esc twice left the window on %v", w.Screen)
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
	clickedOff := w.Choose("Edit").Types(" once more").Key("esc").clicked(geom.Point{X: 1, Y: 1})
	recovered := clickedOff.Choose("Edit").Key("cmd+z").Form.Fields[0].Value
	if clickedOff.Screen != board.OnTheBoard || !strings.Contains(recovered, "once more") {
		t.Errorf("a click off the question left %v, and the form brought back %q", clickedOff.Screen, recovered)
	}
	if says := w.Choose("Edit").Types("x").Key("esc").Asks(); len(says) != 1 || !strings.Contains(says[0].Means, "brings it back") {
		t.Errorf("the question about leaving says %+v", says)
	}
}

func TestATripleClickBesideTheFormHoldsTheWholeLine(t *testing.T) {
	w, _ := openTask(t)
	r := readable(t, w, "trace")
	if got, _ := w.tripleClicked(on(r, 3)).Key("cmd+c").Copying(); got != r.Lines.Said[0] {
		t.Errorf("a triple click on a line of history copied %q, want %q", got, r.Lines.Said[0])
	}
}

func TestTheWorkUnderAnObjectiveHasItsBarPulled(t *testing.T) {
	var cards []board.Card
	for i := range 30 {
		cards = append(cards, board.Card{Number: i + 1, Title: "Task", State: "todo", Objectives: []int{1}})
	}
	w := New(newFake(), 1200, 800)
	w.Snap = test.Searchable(board.Snapshot{Cards: cards, Goals: []board.Goal{{Number: 1, Title: "Ship it", State: "active"}}})
	w = w.Key("tab").clicked(goalMiddle(w.Key("tab"), 1)).Key("enter")
	track := w.workBar()
	middle := geom.Point{X: track.X + track.W/2, Y: track.Y + 4}
	far := geom.Point{X: middle.X, Y: track.Y + track.H + 999}
	if pulled := w.PressAt(middle).MoveTo(far).LetGo(far); pulled.ServingAt() == 0 {
		t.Error("pulling the bar beside the work serving an objective moved nothing")
	}
}

func TestTypingOnAClosedSetOpensItsList(t *testing.T) {
	w := holding(newFake(), board.Card{Number: 1, Title: "A", State: "todo", Priority: "normal"})
	w = w.clicked(middleOf(w, 1, 0)).Choose("Edit")
	for w.Form.Fields[w.Form.On()].Arg != "priority" {
		w = w.Key("tab")
	}
	w = w.Types("h")
	if w.Screen != board.OnChoosing || w.Filtering.Value != "" || len(w.Choosing()) != 4 {
		t.Fatalf("typing h left the window on %v narrowed by %q to %d", w.Screen, w.Filtering.Value, len(w.Choosing()))
	}
	if got := w.Key("esc").Form.Fields[w.Form.On()].Value; got != "normal" {
		t.Errorf("the field reads %q once the list is put away", got)
	}
	w = w.Key("esc")
	for _, name := range []string{"backspace", "cmd+x", "alt+backspace"} {
		if got := w.Key(name).Form.Fields[w.Form.On()].Value; got != "normal" {
			t.Errorf("%s left the field reading %q", name, got)
		}
	}
	if got := w.Key("cmd+v").Pasted("crit"); got.Screen != board.OnChoosing || got.Filtering.Value != "" || len(got.Choosing()) != 4 {
		t.Errorf("pasting left the window on %v narrowed by %q to %d", got.Screen, got.Filtering.Value, len(got.Choosing()))
	}
	for w.Form.Fields[w.Form.On()].Arg != "estimate" {
		w = w.Key("tab")
	}
	if got := w.Types("2h"); got.Screen != board.OnForm || got.Form.Fields[got.Form.On()].Value != "2h" {
		t.Errorf("the estimate took %q", got.Form.Fields[got.Form.On()].Value)
	}
}

func TestClosingTheWindowKeepsWhatWasJustTypedIntoANewTask(t *testing.T) {
	drafts := &heldDrafts{}
	w := holding(newFake()).Drafting(drafts).Key("cmd+n").Types("Plan the week")
	w.Closing()
	held, err := drafts.Held()
	if err != nil || held["title"] != "Plan the week" {
		t.Fatalf("closed straight after typing, the draft holds %v, %v", held, err)
	}
}

func TestAnObjectivesNumberIsWrittenWithItsHash(t *testing.T) {
	w := Window{Board: board.Board{ChoosingFor: "objective"}}
	if got := w.RowSays(Choice{Number: 7, Title: "Ship it"}); !strings.Contains(got, "#7  Ship it") {
		t.Errorf("an objective's row says %q", got)
	}
}

func TestAFieldsExplanationIsABubbleWrappedEvenly(t *testing.T) {
	w := Window{Board: board.Board{Screen: board.OnForm, Form: (Window{}).editing(board.Card{Number: 1, Title: "Ship it"})}, work: newFake(), width: 1200, height: 800}
	at := slices.IndexFunc(w.Form.Fields, func(f board.Field) bool { return f.Name == "time spent" })
	name, _, _ := FormPlace(w.formBox(), w.Form.Fields, at)
	w = w.Hovering(geom.Point{X: name.X + 1, Y: name.Y + 1})
	tip, lines, ok := w.Explanation()
	if !ok || len(lines) < 2 {
		t.Fatalf("time spent explains %q", lines)
	}
	m := w.measured()
	for i, line := range lines {
		if ends(m, line, small) > tipWide-2*padding {
			t.Errorf("line %d, %q, runs wider than a bubble", i, line)
		}
	}
	means := lines[:len(lines)-1]
	if last, first := ends(m, means[len(means)-1], small), ends(m, means[0], small); len(means) > 1 && last*3 < first {
		t.Errorf("the explanation ends on a stub: %q", means)
	}
	if painted(drawn(t, w), geom.Box{X: tip.X + tip.W/2, Y: tip.Y + 1, W: 1, H: 1}, design.Dark().Colour.Surface.Raised) != 1 {
		t.Error("the explanation is not drawn as the window's bubble")
	}
}

func TestAFieldsValueCutShort(t *testing.T) {
	long := "the time we spent getting the whole board ready before anybody used it"
	w := Window{Board: board.Board{Screen: board.OnForm, Form: (Window{}).editing(board.Card{Number: 1, Title: "Ship it", Because: long})}, work: newFake(), width: 1200, height: 800}
	at := slices.IndexFunc(w.Form.Fields, func(f board.Field) bool { return f.Name == "because" })
	_, value, _ := FormPlace(w.formBox(), w.Form.Fields, at)
	if got := w.Hovering(geom.Point{X: value.X + value.W/2, Y: value.Y + value.H/2}).Bubble(); got != long {
		t.Errorf("a value cut short opens a bubble of %q", got)
	}
	at = slices.IndexFunc(w.Form.Fields, func(f board.Field) bool { return f.Name == "title" })
	_, value, _ = FormPlace(w.formBox(), w.Form.Fields, at)
	if got := w.Hovering(geom.Point{X: value.X + value.W/2, Y: value.Y + value.H/2}).Bubble(); got != "" {
		t.Errorf("a value that fits opens a bubble of %q", got)
	}
}

func TestAPressOnTheDetailsTabComesBackFromTheHistory(t *testing.T) {
	for _, goal := range []bool{false, true} {
		w := holding(newFake(), board.Card{Number: 1, Title: "A", State: "todo", Objectives: []int{1}})
		w.Snap.Goals = []board.Goal{{Number: 1, Title: "Ship it", State: "active"}}
		if goal {
			w = w.Key("tab").clicked(goalMiddle(w.Key("tab"), 0)).Key("enter")
		} else {
			w = w.clicked(middleOf(w, 1, 0)).Key("enter")
		}
		w = w.Key("cmd+t")
		tab := BesideTabs(w.formPanel())[0]
		p := geom.Point{X: tab.X + tab.W/2, Y: tab.Y + tab.H/2}
		if got := w.PressAt(p).LetGo(p); got.Tracing() {
			t.Errorf("an objective %v: a press on Details left the history shown", goal)
		}
	}
}

func TestTheFormIsResizedByItsCornerAndKeepsTheSize(t *testing.T) {
	w := holding(newFake(), card(1, "A", "todo"), card(2, "B", "todo"))
	opened := w.clicked(middleOf(w, 1, 0)).Key("enter")
	before := opened.formPanel()
	grip := FormGrip(before)
	from := geom.Point{X: grip.X + grip.W/2, Y: grip.Y + grip.H/2}
	to := geom.Point{X: from.X - 100, Y: from.Y - 60}
	resized := opened.PressAt(from).MoveTo(to).LetGo(to)
	after := resized.formPanel()
	if after.X != before.X || after.Y != before.Y || after.W != before.W-100 || after.H != before.H-60 {
		t.Fatalf("dragged by -100,-60 the form went from %+v to %+v", before, after)
	}
	again := resized.Key("esc").clicked(middleOf(w, 1, 1)).Key("enter")
	if got := again.formPanel(); got.W != after.W || got.H != after.H {
		t.Errorf("the next form opens at %+v, not at the size left %+v", got, after)
	}
	if resized.Sizes().Form == ([2]int{}) {
		t.Error("the size chosen is not kept with the window's other sizes")
	}
	tiny := opened.PressAt(from).MoveTo(geom.Point{X: 0, Y: 0}).LetGo(geom.Point{X: 0, Y: 0}).formPanel()
	if tiny.W < FormLeast.X || tiny.H < FormLeast.Y {
		t.Errorf("dragged far the form shrank to %+v", tiny)
	}
}

func TestTabWalksTheFieldsInTheOrderTheyAreDrawnAndRollsOver(t *testing.T) {
	for _, wide := range []int{900, 1600} {
		f := newFake()
		w := holding(f, board.Card{Number: 1, Title: "Write the spec", State: "todo"})
		w.width = wide
		w = w.clicked(middleOf(w, 1, 0)).Choose("Edit").On("title")
		fields := w.Form.Fields
		box := w.formBox()
		drawn := make([]int, len(fields))
		for i := range drawn {
			drawn[i] = i
		}
		slices.SortStableFunc(drawn, func(a, b int) int {
			_, va, _ := FormPlace(box, fields, a)
			_, vb, _ := FormPlace(box, fields, b)
			ca, cb := va.X >= box.X+box.W, vb.X >= box.X+box.W
			switch {
			case ca != cb && ca:
				return 1
			case ca != cb:
				return -1
			case va.Y != vb.Y:
				return va.Y - vb.Y
			}
			return va.X - vb.X
		})
		var walked []int
		for range fields {
			walked = append(walked, w.Form.On())
			w = w.Key("tab")
		}
		if !slices.Equal(walked, drawn) {
			t.Errorf("%d wide, tab walks %v, the fields are drawn %v", wide, walked, drawn)
		}
		if got := fields[w.Form.On()].Name; got != "title" {
			t.Errorf("%d wide, tab after the last field lands on %q", wide, got)
		}
		if got := w.Key("shift+tab").Form.On(); got != drawn[len(drawn)-1] {
			t.Errorf("%d wide, shift+tab from the title lands on %d, want %d", wide, got, drawn[len(drawn)-1])
		}
	}
}

func TestTheHelpNamesTheKeyThatShowsTheTrace(t *testing.T) {
	w := holding(newFake()).Key("?")
	for _, line := range w.HelpLines() {
		if line.Key == "⇥ ⌘T" && line.Means == "next field, trace" {
			return
		}
	}
	t.Error("the help names no key for the trace")
}

func TestAnOpenTaskInFocusSaysSoBesideItsHeading(t *testing.T) {
	f := newFake()
	w := holding(f, board.Card{Number: 1, Title: "Write the spec", State: "current", InFocus: true},
		board.Card{Number: 2, Title: "Read the spec", State: "todo"})
	focused := w.clicked(middleOf(w, 2, 0)).Choose("Edit")
	pill, ok := focused.FocusPill()
	head := FormHeading(focused.formPanel())
	if !ok || pill.X <= head.X || pill.X+pill.W > head.X+head.W {
		t.Errorf("the task in focus says it at %+v (%v), its heading at %+v", pill, ok, head)
	}
	if _, ok := w.clicked(middleOf(w, 1, 0)).Choose("Edit").FocusPill(); ok {
		t.Error("a task not in focus says it is")
	}
}
