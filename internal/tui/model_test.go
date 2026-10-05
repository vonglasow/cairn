package tui

import (
	"errors"
	"strconv"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"

	"cairn/internal/engine"
	"cairn/internal/reading"
	"cairn/internal/trace"
)

func twoTasks() *fake {
	f := newFake()
	f.cards = []reading.Card{
		{Number: 1, Title: "First", State: "todo"},
		{Number: 2, Title: "Second", State: "todo"},
		{Number: 3, Title: "Waiting", State: "blocked"},
	}
	return f
}

func TestTakingTheFocusAndLettingItGo(t *testing.T) {
	f := twoTasks()
	b := hit(standing(f), "l", "i")
	if _, cmd := b.pressed("f"); cmd != nil {
		cmd()
	}
	if len(f.tookHold) != 1 || f.tookHold[0] != 1 {
		t.Fatalf("the board took hold of %v, want task 1", f.tookHold)
	}

	f.focus = []int{1}
	held := standing(f)
	held = hit(held, "l", "i")
	if _, cmd := held.pressed("f"); cmd != nil {
		cmd()
	}
	if f.letGo != 1 {
		t.Fatalf("the focus was let go %d times, want once", f.letGo)
	}
}

func TestATaskThatRefusesToBeDeletedIsSaid(t *testing.T) {
	f := twoTasks()
	f.failOp = errRefuse
	b := hit(standing(f), "l", "i", "d")
	_, cmd := b.pressed("enter")
	looked, ok := cmd().(Looked)
	if !ok || looked.Err == nil {
		t.Fatalf("a deletion that was refused answered %v", looked)
	}
	if !errors.Is(looked.Err, errRefuse) {
		t.Fatalf("the board reported %v rather than what the store refused with", looked.Err)
	}
}

func TestDeleting(t *testing.T) {
	f := twoTasks()
	b := hit(standing(f), "l", "i", "d")
	if _, cmd := b.pressed("enter"); cmd != nil {
		cmd()
	}
	if len(f.deleted) != 1 || f.deleted[0] != 1 {
		t.Fatalf("the board deleted %v, want task 1", f.deleted)
	}
}

func TestAFailureIsSaidRatherThanSwallowed(t *testing.T) {
	f := twoTasks()
	f.fail = errNope
	b := hit(standing(f), "l", "i", "d")
	_, cmd := b.pressed("enter")
	msg := cmd()
	looked, ok := msg.(Looked)
	if !ok || looked.Err == nil {
		t.Fatalf("a store that refused answered %v", msg)
	}
	after := b.took(looked)
	if !strings.Contains(after.Draw(), errNope.Error()) {
		t.Fatalf("the board says nothing about a store that refused:\n%s", after.Draw())
	}
}

func TestQuitting(t *testing.T) {
	b := standing(twoTasks())
	for _, word := range []string{"q", "quit"} {
		typed := hit(b, ":")
		for _, r := range word {
			typed = hit(typed, string(r))
		}
		_, cmd := typed.pressed("enter")
		if cmd == nil {
			t.Fatalf("%q did not ask to leave", word)
		}
		if _, ok := cmd().(tea.QuitMsg); !ok {
			t.Fatalf("%q asked for something other than leaving", word)
		}
	}
}

func TestTheOtherTab(t *testing.T) {
	f := twoTasks()
	f.goals = []reading.Goal{{Number: 7, Title: "Ship it", State: "active"}}
	b := hit(standing(f), "tab")
	if b.Where().Goal != 7 {
		t.Fatalf("the cursor landed on objective %d, want 7", b.Where().Goal)
	}
	b = hit(b, "tab")
	if b.tab != Tasks {
		t.Fatal("the second press did not come back to the tasks")
	}
}

func TestTheObjectivesTabWithNothingOnIt(t *testing.T) {
	b := hit(standing(twoTasks()), "tab")
	if b.Where().Goal != 0 {
		t.Fatalf("the cursor landed on objective %d where there are none", b.Where().Goal)
	}
	if b = hit(b, "enter"); b.Where().Showing == 0 {
		t.Fatal("opening nothing narrowed the board to nothing")
	}
}

func TestHelpOpensAndClosesFromAnywhere(t *testing.T) {
	for _, keys := range [][]string{{"l"}, {"l", "i"}, {"l", "i", "e"}, {"/"}} {
		b := hit(standing(twoTasks()), keys...)
		was := b.screen
		b = hit(b, "?")
		if b.screen != OnHelp {
			t.Fatalf("help did not answer after %v", keys)
		}
		b = hit(b, "?")
		if b.screen != was {
			t.Fatalf("closing help after %v left %v on screen, want %v", keys, b.screen, was)
		}
	}
}

func TestHelpIsContextual(t *testing.T) {
	f := twoTasks()
	f.goals = []reading.Goal{{Number: 1, Title: "Ship it", State: "active"}}
	onTasks := bare(hit(standing(f), "?").Draw())
	onGoals := bare(hit(standing(f), "tab", "?").Draw())
	if !strings.Contains(onTasks, "INSERT") {
		t.Fatal("the tasks tab's help says nothing about insert")
	}
	if strings.Contains(onGoals, "INSERT") {
		t.Fatal("the objectives tab's help offers a mode it has not got")
	}
}

func TestAWindowThatChangesSize(t *testing.T) {
	b := standing(twoTasks())
	model, _ := b.Update(tea.WindowSizeMsg{Width: 60, Height: 20})
	if got := model.(Board).width; got != 60 {
		t.Fatalf("the board is %d wide after being told 60", got)
	}
}

func TestAMessageNobodySentIsIgnored(t *testing.T) {
	b := standing(twoTasks())
	model, cmd := b.Update("something else entirely")
	if cmd != nil || model.(Board).Draw() != b.Draw() {
		t.Fatal("a message the board knows nothing about changed it")
	}
}

func TestTheBeatAsksAgain(t *testing.T) {
	f := twoTasks()
	b := standing(f)
	_, cmd := b.Update(Beat(time.Now()))
	if cmd == nil {
		t.Fatal("the beat asked for nothing")
	}
}

func TestNothingIsReadWhileTheStoreHasNotMoved(t *testing.T) {
	f := twoTasks()
	f.version = 9
	b := standing(f)
	if msg := b.wonder()(); msg != nil {
		t.Fatalf("the board read the store again for nothing: %v", msg)
	}
	f.version = 10
	if msg := b.wonder()(); msg == nil {
		t.Fatal("the board did not read the store after it moved")
	}
	f.fail = errNope
	if msg := b.wonder()(); msg == nil {
		t.Fatal("a store that refused was passed over in silence")
	}
}

func TestTheDetailOfATaskThatWentAway(t *testing.T) {
	b := standing(twoTasks())
	b.screen = OnDetail
	if !strings.Contains(b.Draw(), "no such task") {
		t.Fatalf("the detail of nothing says:\n%s", b.Draw())
	}
}

func TestTheDetailShowsSubTasksAndHistory(t *testing.T) {
	f := newFake()
	f.cards = []reading.Card{
		{Number: 1, Title: "Parent", State: "todo", SubDone: 1, SubTotal: 2, Labels: []string{"spec"}, Estimate: "3h", Spent: "1h", Deadline: on("2026-09-02")},
		{Number: 2, Title: "Child", State: "done", Parent: 1},
	}
	var told trace.AsTold
	told.On(1)
	told.Did("created")
	told.Changing("title")
	told.WasNothing()
	told.Became("Parent")
	told.At(today)
	f.acts = []trace.Act{told.Act()}
	b := hit(standing(f), "l", "i", "enter")
	drawn := strip(b.Draw())
	for _, want := range []string{"SUB-TASKS", "Child", "HISTORY", "created", "title"} {
		if !strings.Contains(drawn, want) {
			t.Fatalf("the detail never says %q:\n%s", want, drawn)
		}
	}
}

func TestTheDetailSaysWhenAHistoryCannotBeRead(t *testing.T) {
	f := twoTasks()
	f.fail = errNope
	b := heardOf(standing(f), 1)
	b.mode, b.task, b.screen = Insert, 1, OnDetail
	if !strings.Contains(b.Draw(), errNope.Error()) {
		t.Fatalf("a history that could not be read was passed over:\n%s", b.Draw())
	}
}

func TestTooNarrowSaysWhatItNeeds(t *testing.T) {
	b := standing(twoTasks())
	b.width = 47
	drawn := bare(b.Draw())
	if !strings.Contains(drawn, "bigger window") || !strings.Contains(drawn, "60 by 26") || !strings.Contains(drawn, "is 47 by") {
		t.Fatalf("a window too small says:\n%s", drawn)
	}
}

func TestTheViewCarriesTheScreenAndTakesTheWholeTerminal(t *testing.T) {
	b := standing(twoTasks())
	view := b.View()
	if view.Content != b.Draw() {
		t.Fatal("what bubbletea is handed is not what the board drew")
	}
	if !view.AltScreen {
		t.Fatal("the board did not ask for the whole terminal")
	}
}

func TestInitAsksForAReadingAndABeat(t *testing.T) {
	if standing(twoTasks()).Init() == nil {
		t.Fatal("a board opened without asking for anything")
	}
}

func TestKeyNameSpellsAShiftedLetterOneWay(t *testing.T) {
	if got := keyName(tea.KeyPressMsg{Code: 'L', Text: "L"}); got != "shift+l" {
		t.Fatalf("a capital is called %q", got)
	}
	if got := keyName(tea.KeyPressMsg{Code: 'l', Text: "l"}); got != "l" {
		t.Fatalf("a plain letter is called %q", got)
	}
}

func TestDirections(t *testing.T) {
	for key, want := range map[string]string{
		"h": goLeft, "left": goLeft,
		"j": goDown, "down": goDown,
		"k": goUp, "up": goUp,
		"l": goRight, "right": goRight,
	} {
		if got, ok := direction(key); !ok || got != want {
			t.Fatalf("%q points %q, want %q", key, got, want)
		}
	}
	if _, ok := direction("a"); ok {
		t.Fatal("a letter that points nowhere was said to point somewhere")
	}
	if _, ok := shifted("a"); ok {
		t.Fatal("an unshifted letter was read as shifted")
	}
	if got, ok := shifted("shift+l"); !ok || got != goRight {
		t.Fatalf("shift+l points %q", got)
	}
}

func TestTyped(t *testing.T) {
	for key, want := range map[string]string{"a": "a", "shift+a": "A", "space": " "} {
		got, ok := typed(key)
		if !ok || got != want {
			t.Fatalf("%q types %q, want %q", key, got, want)
		}
	}
	if _, ok := typed("f1"); ok {
		t.Fatal("a key that types nothing was said to type something")
	}
}

func TestUnknownNamesWhatWasTypedAndWhatIsTaken(t *testing.T) {
	err := unknown("zz")
	if !strings.Contains(err.Error(), "zz") || !strings.Contains(err.Error(), "backup") {
		t.Fatalf("the refusal says %q", err)
	}
}

// Every argument the operation takes has to be named, empty ones included: a
// form is the whole of a task, so a field left blank is a field somebody
// emptied. Read off the operation rather than listed here, so an argument added
// to the engine and forgotten on the form fails this test.
func TestSavingAFormNamesEveryArgumentTheOperationTakes(t *testing.T) {
	form := asking(reading.TaskFields(reading.Card{
		Number: 1, Title: "One", State: "todo", Priority: "high",
		Labels: []string{"spec"}, Blockers: []int{2}, Because: "waiting",
		Estimate: "3h", Spent: "1h", Parent: 4, Objectives: []int{5},
		Deadline: on("2026-09-02"), Description: "why",
	}), 80, 20, plainSkin)
	form.Number = 1

	_, in := (Board{}).saving(form, engine.TaskAdd, engine.TaskEdit)
	for _, argument := range engine.TaskEdit.Takes {
		if !in.Named(argument.Name) {
			t.Errorf("%s was not named, so emptying it would do nothing", argument.Name)
		}
	}
	if in.Given["state"] != "todo" || in.Given["priority"] != "high" || in.Given["parent"] != "4" {
		t.Fatalf("a field was read back wrong: %+v", in.Given)
	}
}

func TestValueOfAFieldThatIsNotThere(t *testing.T) {
	if got := (Form{}).valueOf("nothing"); got != "" {
		t.Fatalf("a field that does not exist answered %q", got)
	}
}

func TestEngineAnswersEverythingTheBoardAsksFor(t *testing.T) {
	// The board declares what it needs; this is what says the engine gives it.
	// Without it, a method renamed in the engine would only be found by opening
	// the board by hand.
	var _ reading.Work = (*engine.Engine)(nil)
}

func TestRemovingAnObjective(t *testing.T) {
	f := newFake()
	f.goals = []reading.Goal{{Number: 3, Title: "Ship it", State: "active"}}
	b := hit(standing(f), "tab", "d")
	if _, cmd := b.pressed("enter"); cmd != nil {
		cmd()
	}
	if len(f.deletedGoal) != 1 || f.deletedGoal[0] != 3 {
		t.Fatalf("the board deleted %v, want objective 3", f.deletedGoal)
	}
}

func TestRemovingNoObjectiveAtAll(t *testing.T) {
	f := newFake()
	b := hit(standing(f), "tab")
	if _, cmd := b.pressed("d"); cmd != nil {
		t.Fatal("delete asked for something with no objective under the cursor")
	}
}

func TestAnObjectiveThatRefusesToBeDeletedIsSaid(t *testing.T) {
	f := newFake()
	f.goals = []reading.Goal{{Number: 3, Title: "Ship it", State: "active"}}
	f.failOp = errRefuse
	b := hit(standing(f), "tab", "d")
	_, cmd := b.pressed("enter")
	if looked, ok := cmd().(Looked); !ok || looked.Err == nil {
		t.Fatal("an objective that could not be deleted was passed over")
	}
}

func TestAFocusThatRefusesIsSaid(t *testing.T) {
	f := twoTasks()
	f.failOp = errRefuse
	b := hit(standing(f), "l", "i")
	_, cmd := b.pressed("f")
	if looked, ok := cmd().(Looked); !ok || looked.Err == nil {
		t.Fatal("a focus that could not be taken was passed over")
	}
	f.focus = []int{1}
	held := hit(standing(f), "l", "i")
	_, cmd = held.pressed("f")
	if looked, ok := cmd().(Looked); !ok || looked.Err == nil {
		t.Fatal("a focus that could not be let go was passed over")
	}
}

func TestAStateThatRefusesIsSaid(t *testing.T) {
	f := twoTasks()
	f.failOp = errRefuse
	b := hit(standing(f), "l", "i")
	_, cmd := b.pressed("shift+l")
	if looked, ok := cmd().(Looked); !ok || looked.Err == nil {
		t.Fatal("a move that could not be written was passed over")
	}
}

func TestASaveThatRefusesIsSaid(t *testing.T) {
	f := newFake()
	f.cards = []reading.Card{{Number: 1, Title: "First", State: "todo"}}
	f.failOp = errRefuse
	if b := finish(hit(standing(f), "l", "i", "e"), true); b.err == nil {
		t.Fatal("a save that could not be written was passed over")
	}
	if b := finish(hit(standing(f), "l", "a"), true); b.err == nil {
		t.Fatal("an add that could not be written was passed over")
	}
}

func TestAnObjectiveSaveThatRefusesIsSaid(t *testing.T) {
	f := newFake()
	f.goals = []reading.Goal{{Number: 3, Title: "Ship it", State: "active"}}
	f.failOp = errRefuse
	if b := finish(hit(standing(f), "tab", "e"), true); b.err == nil {
		t.Fatal("an objective that could not be saved was passed over")
	}
	if b := finish(hit(standing(f), "tab", "a"), true); b.err == nil {
		t.Fatal("an objective that could not be added was passed over")
	}
}

func TestTheBeatIsTwoSeconds(t *testing.T) {
	if got := beat(); got != 2*time.Second {
		t.Fatalf("the board asks again every %v", got)
	}
}

func TestChosenIsOnlyTheOneHeldInInsert(t *testing.T) {
	b := standing(twoTasks())
	card := reading.Card{Number: 1}
	if b.chosen(card) {
		t.Fatal("a card was picked out while the cursor was on a column")
	}
	b.mode, b.task = Insert, 1
	if !b.chosen(card) {
		t.Fatal("the card the cursor holds was not picked out")
	}
	if b.chosen(reading.Card{Number: 2}) {
		t.Fatal("a card the cursor does not hold was picked out")
	}
}

func TestTheObjectiveUnderTheCursorIsDrawnApart(t *testing.T) {
	f := newFake()
	f.goals = []reading.Goal{
		{Number: 1, Title: "One", State: "active"},
		{Number: 2, Title: "Two", State: "active"},
	}
	b := hit(standing(f), "tab")
	// The difference is a colour, so the drawing is read before the styling is
	// taken off it — and it is the row of the objective the cursor is on that
	// has to change, not merely some row.
	// Compared against a card drawn as the one under the cursor and one drawn
	// as any other, so the test says which of the two is which rather than only
	// that they differ.
	width := b.width - 2*margin - frame
	held := b
	held.goal = 1
	chosen := strings.TrimRight(rowOf(t, held.goalCard(f.goals[0], width), "One"), " ")
	elsewhere := b
	elsewhere.goal = 2
	other := strings.TrimRight(rowOf(t, elsewhere.goalCard(f.goals[0], width), "One"), " ")
	if chosen == other {
		t.Fatal("a card under the cursor is drawn exactly like any other")
	}
	b.goal = 1
	if got := onTheBoard(t, b, "One"); got != chosen {
		t.Fatalf("the objective under the cursor is not drawn as the one under it:\n%q\n%q", got, chosen)
	}
	b.goal = 2
	if got := onTheBoard(t, b, "One"); got != other {
		t.Fatalf("an objective the cursor is not on is drawn as though it were:\n%q\n%q", got, other)
	}
}

// onTheBoard finds the row a title is drawn on, as the board draws it: without
// the margin in front of it, and without the blank the window pads it out with.
func onTheBoard(t *testing.T, b Board, title string) string {
	t.Helper()
	return strings.TrimRight(strings.TrimPrefix(rowOf(t, b.Draw(), title), " "), " ")
}

// rowOf finds the drawn line carrying a title, styling and all.
func rowOf(t *testing.T, drawn, title string) string {
	t.Helper()
	for _, row := range strings.Split(drawn, "\n") {
		if strings.Contains(row, title) {
			return row
		}
	}
	t.Fatalf("no row carries %q:\n%s", title, drawn)
	return ""
}

func TestTheBarIsCutToTheScreenAndNoWider(t *testing.T) {
	// Forty-eight columns is exactly where "/ search" starts to fit. Pinned at
	// the character, because a bar one column too generous wraps and pushes the
	// board up out of the window.
	b := standing(twoTasks())
	b.skin = plainSkin
	b.width = 47
	if got := b.foot(); got != " hjkl ←↓↑→ columns   i insert   a add" {
		t.Fatalf("at 47 the bar reads %q", got)
	}
	b.width = 48
	if got := b.foot(); got != " hjkl ←↓↑→ columns   i insert   a add   / search" {
		t.Fatalf("at 48 the bar reads %q", got)
	}
}

func TestABandOfExactlyOneRowDrawsOneRow(t *testing.T) {
	f := newFake()
	f.cards = []reading.Card{
		{Number: 1, State: "blocked", Title: "A"},
		{Number: 2, State: "watching", Title: "B"},
		{Number: 3, State: "paused", Title: "C"},
	}
	b := standing(f)
	drawn := b.laidOut(f.cards, LayoutFor(100, 40))
	if got := len(strings.Split(drawn, "\n")); got != CardHeight {
		t.Fatalf("three cards across drew %d rows, want %d:\n%s", got, CardHeight, drawn)
	}
}

func manyGoals() *fake {
	f := newFake()
	for i := 1; i <= 9; i++ {
		f.goals = append(f.goals, reading.Goal{Number: i, Title: "Objective " + strconv.Itoa(i), State: "active"})
	}
	return f
}

func TestLeavingAnObjectiveGoesBackToTheList(t *testing.T) {
	f := newFake()
	f.goals = []reading.Goal{{Number: 3, Title: "Ship it", State: "active"}}
	f.cards = []reading.Card{{Number: 1, Title: "One", State: "todo", Objectives: []int{3}}}
	b := hit(standing(f), "tab", "enter")
	if b.inside != 3 {
		t.Fatalf("opening the objective left inside at %d", b.inside)
	}
	b = hit(b, "esc")
	if b.inside != 0 || b.tab != Objectives || b.Where().Goal != 3 {
		t.Fatalf("escape left inside=%d tab=%v goal=%d", b.inside, b.tab, b.Where().Goal)
	}
}

func TestEscapeOnTheBoardItselfDoesNothing(t *testing.T) {
	b := standing(twoTasks())
	if after := hit(b, "esc"); after.Draw() != b.Draw() {
		t.Fatal("escape changed a board there is nothing above")
	}
}

func TestACardThatIsNotStalledIsNotMarkedAsThoughItWere(t *testing.T) {
	// The state's mark is put on in colour, and only when there is one. A card
	// in a column has none, and nothing at all should be added to its first row
	// — not even an empty piece of styling, which is invisible until somebody
	// measures the row.
	skin := SkinOf(Default)
	g := GlyphsFor(Unicode)
	todo := Rows(reading.Card{Number: 1, Title: "One", State: "todo"}, g, skin, today, 20)
	if got, want := todo[0], Spread(skin.Faint.Render("1"), "", 20); got != want {
		t.Fatalf("a card in a column has %q on its first row, want %q", got, want)
	}
	stalled := Rows(reading.Card{Number: 1, Title: "One", State: "blocked"}, g, skin, today, 20)
	if stalled[0] == todo[0] {
		t.Fatal("a stalled card is marked exactly like one in a column")
	}
}

func TestAnObjectiveThatIsNotStandingSomewhereKnownIsNotMarked(t *testing.T) {
	b := standing(newFake())
	b.skin = SkinOf(Default)
	known := b.goalRows(reading.Goal{Number: 1, Title: "One", State: "active"}, 40)
	unknown := b.goalRows(reading.Goal{Number: 1, Title: "One", State: "invented"}, 40)
	if known[0] == unknown[0] {
		t.Fatal("a standing nobody planned for is marked like one that is known")
	}
	if got, want := unknown[0], Spread(b.skin.Faint.Render("1"), "", 40); got != want {
		t.Fatalf("an unknown standing put %q on the row, want %q", got, want)
	}
}

func TestTheObjectiveDrawnAsTheOneUnderTheCursor(t *testing.T) {
	f := newFake()
	f.goals = []reading.Goal{
		{Number: 1, Title: "One", State: "active"},
		{Number: 2, Title: "Two", State: "active"},
	}
	b := hit(standing(f), "tab")
	width := b.width - 2*margin - frame
	b.goal = 1
	// Compared with the frame drawn on purpose as the one under the cursor, so
	// the test says which of the two is which rather than only that they differ.
	want := b.skin.Card(width+frame, "", false, true).Render(strings.Join(b.goalRows(f.goals[0], width), "\n"))
	if got := b.goalCard(f.goals[0], width); got != want {
		t.Fatal("the objective under the cursor is not drawn as the one under it")
	}
	b.goal = 2
	elsewhere := b.skin.Card(width+frame, "", false, false).Render(strings.Join(b.goalRows(f.goals[0], width), "\n"))
	if got := b.goalCard(f.goals[0], width); got != elsewhere {
		t.Fatal("an objective the cursor is not on is drawn as though it were")
	}
}

func TestAWindowOfOneRowStillDrawsARow(t *testing.T) {
	if got := through("one\ntwo\nthree", 10, 1, 0); got != "one       " {
		t.Fatalf("a window one row tall drew %q", got)
	}
	if got := through("one", 10, 0, 0); got != "" {
		t.Fatalf("a window no rows tall drew %q", got)
	}
}

func TestTheBandDrawsEveryRowItIsGiven(t *testing.T) {
	f := newFake()
	for i := 1; i <= 3; i++ {
		f.cards = append(f.cards, reading.Card{Number: i, Title: "One", State: "blocked"})
	}
	b := standing(f)
	// Three cards is exactly one row of three. A fourth row of nothing would be
	// a blank line the band pays for.
	if got := len(strings.Split(b.laidOut(f.cards, LayoutFor(100, 40)), "\n")); got != CardHeight {
		t.Fatalf("three cards across drew %d rows, want %d", got, CardHeight)
	}
	f.cards = append(f.cards, reading.Card{Number: 4, Title: "Four", State: "blocked"})
	if got := len(strings.Split(b.laidOut(f.cards, LayoutFor(100, 40)), "\n")); got != 2*CardHeight {
		t.Fatalf("four cards across drew %d rows, want two rows of cards", got)
	}
}

func TestTheHistoryShowsTheLatestAndScrollsForTheRest(t *testing.T) {
	f := newFake()
	f.cards = []reading.Card{{Number: 1, Title: "One", State: "todo"}}
	for i := range 9 {
		f.acts = append(f.acts, anAct("changed", strconv.Itoa(i)))
	}
	b := heardOf(standing(f), 1)
	b.mode, b.task = Insert, 1
	// A window with room for three: the three that matter are the last three,
	// because a history is read from its end.
	got := b.history(1, 40, 3)
	if len(got) != 3 {
		t.Fatalf("a window three rows tall showed %d acts", len(got))
	}
	if !strings.Contains(got[2], "8") {
		t.Fatalf("the last act shown is %q, want the latest", got[2])
	}
	// And no window at all shows everything, for whoever is counting rather
	// than looking.
	if all := b.history(1, 40, 0); len(all) != 9 {
		t.Fatalf("with no window %d acts came back, want all nine", len(all))
	}
	if one := b.history(1, 40, 1); len(one) != 1 {
		t.Fatalf("a window one row tall showed %d acts", len(one))
	}
}

func TestAnObjectiveWithNothingUnderItShowsNoCount(t *testing.T) {
	b := standing(newFake())
	b.skin = plainSkin
	if got := b.goalFacts(reading.Goal{Number: 1}, 60); strings.Contains(got, "0/0") {
		t.Fatalf("an objective nothing serves claimed a count: %q", got)
	}
	if got := b.goalFacts(reading.Goal{Number: 1, Done: 0, Total: 3}, 60); !strings.Contains(got, "0/3") {
		t.Fatalf("an objective nothing is done on lost its count: %q", got)
	}
}

func TestTheFactsBesideAnObjectiveBarAreCutToWhatIsLeft(t *testing.T) {
	b := standing(newFake())
	b.skin = plainSkin
	g := reading.Goal{Number: 1, Done: 1, Total: 2, Labels: []string{"ab"}}
	// "1/2  ⚑ ab" is nine columns, and what is left for it is the width less
	// the bar and the two spaces beside it.
	const room = barWidth + 2 + 9
	if got := b.goalFacts(g, room); !strings.Contains(got, "⚑ ab") {
		t.Fatalf("a row with room for the label reads %q", got)
	}
	if got := b.goalFacts(g, room-1); strings.Contains(got, "⚑ ab") {
		t.Fatalf("a row one column too narrow kept the label: %q", got)
	}
}

func TestAnObjectiveCardIsAsWideAsItIsGiven(t *testing.T) {
	b := standing(newFake())
	drawn := b.goalCard(reading.Goal{Number: 1, Title: "One"}, 40)
	for _, row := range strings.Split(strings.TrimRight(bare(drawn), "\n"), "\n") {
		if got := len([]rune(row)); got != 40+frame {
			t.Fatalf("a card given 40 drew a row %d wide, want %d", got, 40+frame)
		}
	}
}

func TestEveryScreensBarIsCutToTheSameWidth(t *testing.T) {
	b := standing(twoTasks())
	b.skin = plainSkin
	b.width = 30
	if got := b.bar(b.offer("a", "one"), b.offer("b", "two"), b.offer("c", "three")); got != " a one   b two   c three" {
		t.Fatalf("a bar that fits reads %q", got)
	}
	b.width = 17
	if got := b.bar(b.offer("a", "one"), b.offer("b", "two"), b.offer("c", "three")); got != " a one   b two" {
		t.Fatalf("a bar one column too narrow for the last reads %q", got)
	}
}

func TestTheHistoryIsGivenWhatTheWindowLeavesIt(t *testing.T) {
	f := newFake()
	f.cards = []reading.Card{{Number: 1, Title: "One", State: "todo"}}
	for i := range 40 {
		f.acts = append(f.acts, anAct("changed", strconv.Itoa(i)))
	}
	for _, height := range []int{28, 40} {
		b := heardOf(standing(f), 1)
		b.width, b.height = 100, height
		b.mode, b.task, b.screen = Insert, 1, OnDetail
		// The panel gets what the window leaves after the tabs, the band and
		// the bar, and it fills it: a history longer than that scrolls.
		if got := len(b.history(1, 40, height-10)); got != height-10 {
			t.Fatalf("in a window %d tall the history drew %d rows, want %d", height, got, height-10)
		}
	}
}

func TestTheHistoryShowsItsEndAndNotItsBeginning(t *testing.T) {
	f := newFake()
	f.cards = []reading.Card{{Number: 1, Title: "One", State: "todo"}}
	for i := range 9 {
		f.acts = append(f.acts, anAct("changed", strconv.Itoa(i)))
	}
	got := heardOf(standing(f), 1).history(1, 40, 3)
	if !strings.Contains(got[0], "6") {
		t.Fatalf("the first row shown is %q, want the seventh act", got[0])
	}
	if !strings.Contains(got[2], "8") {
		t.Fatalf("the last row shown is %q, want the ninth act", got[2])
	}
}

func TestEveryScreenFitsTheWindowItIsDrawnIn(t *testing.T) {
	// The one thing that has to hold of every screen: it fits. A screen one row
	// too tall pushes the tabs off the top and the bar off the bottom, and
	// neither comes back.
	f := fullBoard()
	for _, keys := range [][]string{
		nil, {"l"}, {"l", "i"}, {"?"}, {"l", "i", "enter"}, {"l", "i", "e"},
		{"l", "a"}, {"tab"}, {"tab", "?"}, {"tab", "e"}, {"tab", "a"},
		{"tab", "enter"}, {"/", "s"}, {":", "b"}, {"l", "j"},
	} {
		for _, size := range [][2]int{{Narrowest, Shortest}, {72, 30}, {100, 44}, {160, 60}} {
			b := standing(f)
			b.width, b.height = size[0], size[1]
			b = hit(b, keys...)
			if got := len(strings.Split(b.Draw(), "\n")); got > b.height {
				t.Fatalf("after %v in a window %dx%d the screen is %d rows, more than the window",
					keys, size[0], size[1], got)
			}
		}
	}
}

func TestEscapeClosesTheForm(t *testing.T) {
	f := newFake()
	f.cards = []reading.Card{{Number: 1, Title: "One", State: "todo", Priority: "normal"}}
	b := hit(standing(f), "l", "i", "e")
	if b.screen != OnForm {
		t.Fatal("the form did not open")
	}
	b = hit(b, "esc")
	if b.screen != OnTheBoard {
		t.Fatal("escape did not close the form")
	}
	if len(f.edited) != 0 {
		t.Fatalf("escape still wrote %v", f.edited)
	}
}

func TestEscapeClosesAnObjectivesForm(t *testing.T) {
	f := newFake()
	f.goals = []reading.Goal{{Number: 3, Title: "Ship it", State: "active"}}
	b := hit(standing(f), "tab", "e")
	if b.screen != OnForm {
		t.Fatal("the objective's form did not open")
	}
	b = hit(b, "esc")
	if b.screen != OnTheBoard || b.tab != Objectives {
		t.Fatalf("escape left screen=%v tab=%v", b.screen, b.tab)
	}
	if len(f.editedGoal) != 0 {
		t.Fatalf("escape still wrote %v", f.editedGoal)
	}
}

func TestTheHelpCanBeReadPastTheBottomOfTheWindow(t *testing.T) {
	b := standing(fullBoard())
	b.width, b.height = Narrowest, Shortest
	b = hit(b, "?")
	if b.helpAt != 0 {
		t.Fatalf("the help opened scrolled to %d", b.helpAt)
	}
	b = hit(b, "j", "j", "j")
	if b.helpAt != 3 {
		t.Fatalf("reading on left it at %d, want three lines down", b.helpAt)
	}
	b = hit(b, "k", "k", "k", "k")
	if b.helpAt != 0 {
		t.Fatalf("reading back up left it at %d, want the top", b.helpAt)
	}
}

func TestTheFormHeadSaysWhatIsBeingChanged(t *testing.T) {
	f := newFake()
	f.cards = []reading.Card{{Number: 7, Title: "Write the spec", State: "blocked", Priority: "critical"}}
	b := hit(standing(f), "l", "j", "i", "e").with(plainSkin)
	head := b.formHead()
	for _, want := range []string{"7", "Write the spec", "blocked", "critical"} {
		if !strings.Contains(head, want) {
			t.Fatalf("the form's head %q is missing %q", head, want)
		}
	}
	adding := hit(standing(f), "l", "a").with(plainSkin)
	if got := adding.formHead(); !strings.Contains(got, "new") || !strings.Contains(got, "untitled") {
		t.Fatalf("a new task's head reads %q", got)
	}
	goals := newFake()
	goals.goals = []reading.Goal{{Number: 2, Title: "Ship it", State: "on-hold", Priority: "low"}}
	objective := hit(standing(goals), "tab", "e").with(plainSkin)
	if got := objective.formHead(); !strings.Contains(got, "on-hold") {
		t.Fatalf("an objective's head reads %q", got)
	}
}

// A write that went through is what closes the form, and nothing else is
// closed by a reading coming back: a task open on its detail stays open when
// the board hears from the store again.
func TestAReadingComingBackClosesTheFormAndNothingElse(t *testing.T) {
	f := twoTasks()
	b := hit(standing(f), "l", "i", "e")
	if b.screen != OnForm {
		t.Fatalf("the form did not open: %v", b.screen)
	}
	if after := b.took(Looked{Snapshot: reading.Snapshot{Cards: f.cards}}); after.screen != OnTheBoard {
		t.Fatalf("a write that went through left the board on %v", after.screen)
	}

	onDetail := hit(standing(f), "l", "i", "enter")
	if onDetail.screen != OnDetail {
		t.Fatalf("the detail did not open: %v", onDetail.screen)
	}
	if after := onDetail.took(Looked{Snapshot: reading.Snapshot{Cards: f.cards}}); after.screen != OnDetail {
		t.Fatalf("a reading coming back closed the detail: %v", after.screen)
	}
}

// Settling the parts of a task is several writes in one gesture, and any of
// them may be refused. What the store refuses is said rather than swallowed,
// whichever of them it was.
func TestAPartTheStoreRefusesIsSaid(t *testing.T) {
	for _, verb := range []string{"add", "delete"} {
		t.Run("a part the store will not "+verb, func(t *testing.T) {
			f := newFake()
			f.cards = []reading.Card{
				{Number: 1, Title: "Ship it", State: "todo"},
				{Number: 2, Title: "Relire l'ADR", State: "todo", Parent: 1},
			}
			b := hit(standing(f), "l", "i", "e")
			if b.screen != OnForm {
				t.Fatalf("the form did not open: %v", b.screen)
			}
			// Naming one and dropping the other, so both writes are asked for.
			b.form = b.form.typedInto("parts", "[ ] Rédiger la section Flux")
			f.refuseVerb = verb
			cmd := b.saveTask(b.form)
			if cmd == nil {
				t.Fatal("saving asked for nothing")
			}
			looked, ok := cmd().(Looked)
			if !ok || looked.Err == nil {
				t.Fatalf("a part the store refused to %s answered %v", verb, looked)
			}
		})
	}
}

// The history of an open task is read once, when it opens, and again when the
// board is read again — never while it is being drawn. Drawing asked the store
// every time, so every redraw of the open task was a round trip, against the
// promise that nothing is fetched while a key is being handled.
func TestTheHistoryIsReadWhenTheTaskOpensAndNotWhenItIsDrawn(t *testing.T) {
	f := twoTasks()
	b := hit(standing(f), "l", "i")
	b = hit(b, "enter")
	if b.screen != OnDetail {
		t.Fatalf("enter left the board on %v", b.screen)
	}
	heard := 0
	for _, a := range f.asked {
		if a.op.Name() == "trace_show" {
			heard++
		}
	}
	if heard != 1 {
		t.Fatalf("opening the task read its history %d times, want once", heard)
	}
	for range 3 {
		_ = b.Draw()
	}
	after := 0
	for _, a := range f.asked {
		if a.op.Name() == "trace_show" {
			after++
		}
	}
	if after != heard {
		t.Errorf("drawing three times read the history %d more times", after-heard)
	}
	// Stepping to the next task reads that one's.
	_ = hit(b, "j")
	last := f.asked[len(f.asked)-1]
	if last.op.Name() != "trace_show" || last.in.Given["number"] != "2" {
		t.Errorf("stepping to task 2 asked %v", last)
	}
}

// A line of history says which surface did it when it was not the person
// reading — and says nothing of the kind when it was.
func TestAHistoryLineSaysWhoDidItWhenItWasNotYou(t *testing.T) {
	b := standing(newFake())
	var told trace.AsTold
	told.On(1)
	told.Did("changed")
	told.Changing("priority")
	told.Was("normal")
	told.Became("high")
	told.Through("mcp")
	told.At(today)
	if got := strip(b.line(told.Act())); !strings.HasSuffix(got, "(mcp)") || !strings.Contains(got, "normal → high") {
		t.Errorf("an agent's change reads %q", got)
	}
	if got := strip(b.line(anAct("created", ""))); strings.Contains(got, "(") {
		t.Errorf("a change made by hand reads %q", got)
	}
}

// Reading the board again reads the open task's history again, since what was
// just written is in it; with no task open there is no history to read.
func TestReadingTheBoardAgainReadsTheOpenHistoryAgain(t *testing.T) {
	f := twoTasks()
	b := hit(standing(f), "l", "i", "enter")
	if _, cmd := b.Update(Looked{Snapshot: b.snap}); cmd == nil {
		t.Fatal("reading the board again with a task open read nothing more")
	}
	if _, cmd := standing(f).Update(Looked{Snapshot: b.snap}); cmd != nil {
		t.Error("reading the board again with no task open asked for more")
	}
}
