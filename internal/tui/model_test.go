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

func TestACommandNobodyKnowsIsSaid(t *testing.T) {
	b := hit(standing(twoTasks()), ":", "z", "enter")
	if !strings.Contains(bare(b.Draw()), "not a command") {
		t.Fatalf("a command nobody knows was swallowed:\n%s", b.Draw())
	}
}

func TestASearchIsKeptWhenItIsAccepted(t *testing.T) {
	b := hit(standing(twoTasks()), "/", "F", "i", "r", "enter")
	if got := b.Where().Showing; got != 1 {
		t.Fatalf("the board carries %d tasks after a search, want 1", got)
	}
}

func TestACapitalCanBeTyped(t *testing.T) {
	b := hit(standing(twoTasks()), "/", "shift+f")
	if !strings.Contains(bare(b.Draw()), "/F") {
		t.Fatalf("a capital never reached the prompt:\n%s", b.Draw())
	}
}

func TestBackspaceAtAPrompt(t *testing.T) {
	b := hit(standing(twoTasks()), "/", "F", "i", "backspace")
	if !strings.Contains(bare(b.Draw()), "/F▏") {
		t.Fatalf("backspace did not take the last letter off:\n%s", b.Draw())
	}
}

func TestAKeyThatMeansNothingAtAPromptIsIgnored(t *testing.T) {
	b := hit(standing(twoTasks()), "/", "F", "f1")
	if !strings.Contains(bare(b.Draw()), "/F▏") {
		t.Fatalf("an unknown key changed what was typed:\n%s", b.Draw())
	}
}

func TestWalkingAcrossToAnEmptyColumnLetsGo(t *testing.T) {
	b := hit(standing(twoTasks()), "l", "i", "l")
	if b.Where().Mode != Normal {
		t.Fatal("the cursor kept hold of a task on a column with none")
	}
	if b.Where().Task != 0 {
		t.Fatalf("the cursor still holds task %d", b.Where().Task)
	}
}

func TestWalkingAcrossKeepsHoldWhereThereIsSomething(t *testing.T) {
	f := twoTasks()
	f.cards = append(f.cards, reading.Card{Number: 4, Title: "Over there", State: "current"})
	b := hit(standing(f), "l", "i", "l")
	if b.Where().Task != 4 {
		t.Fatalf("the cursor holds %d, want 4", b.Where().Task)
	}
}

func TestWalkingAcrossStopsAtEitherEnd(t *testing.T) {
	f := twoTasks()
	b := hit(standing(f), "l", "i", "h", "h")
	if b.Where().Column != 0 {
		t.Fatalf("the cursor walked past the first column to %d", b.Where().Column)
	}
}

func TestTheBandIsOnePlaceHoweverWideItIs(t *testing.T) {
	b := hit(standing(twoTasks()), "l", "j")
	before := b.Where().Column
	b = hit(b, "l", "h")
	if b.Where().Column != before || !b.Where().InBand {
		t.Fatal("walking left and right inside the band moved between columns")
	}
}

func TestSteppingUpAndDownAColumn(t *testing.T) {
	b := hit(standing(twoTasks()), "l", "i")
	if b.Where().Task != 1 {
		t.Fatalf("the cursor took hold of %d, want 1", b.Where().Task)
	}
	b = hit(b, "j")
	if b.Where().Task != 2 {
		t.Fatalf("j moved to %d, want 2", b.Where().Task)
	}
	// Off the bottom of a column the cursor carries on into the band, which is
	// what is under it on screen.
	b = hit(b, "j")
	if !b.Where().InBand || b.Where().Task != 3 {
		t.Fatalf("j off the last task landed on %d, band %v", b.Where().Task, b.Where().InBand)
	}
	b = hit(b, "k")
	if b.Where().InBand || b.Where().Task != 2 {
		t.Fatalf("k out of the band landed on %d, band %v", b.Where().Task, b.Where().InBand)
	}
	b = hit(b, "k", "k")
	if b.Where().Task != 1 {
		t.Fatal("k walked past the first task")
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

func TestTheKindPopupWalksAndCancels(t *testing.T) {
	// From the last task in the column: from any other, pushing down moves it
	// one place lower rather than out of the column.
	b := hit(standing(twoTasks()), "l", "i", "j", "shift+j")
	if !strings.Contains(bare(b.Draw()), "which kind of stopped") {
		t.Fatal("pushing the last task down asked nothing")
	}
	b = hit(b, "j", "j", "j")
	if b.choice != len(kinds)-1 {
		t.Fatalf("the popup walked to %d, want its last answer", b.choice)
	}
	b = hit(b, "k", "k", "k")
	if b.choice != 0 {
		t.Fatalf("the popup walked to %d, want its first answer", b.choice)
	}
	b = hit(b, "esc")
	if strings.Contains(b.Draw(), "which kind of stopped") {
		t.Fatal("escape did not close the popup")
	}
}

func TestPushingIntoTheBandWritesTheKindChosen(t *testing.T) {
	f := twoTasks()
	b := hit(standing(f), "l", "i", "j", "shift+j", "j")
	_, cmd := b.pressed("enter")
	cmd()
	if got := f.edited[2].State; got == nil || *got != "watching" {
		t.Fatalf("the board wrote %v, want watching", got)
	}
}

func TestPushingDownFromTheBandDoesNothing(t *testing.T) {
	f := twoTasks()
	b := hit(standing(f), "l", "j", "i")
	b = hit(b, "shift+j")
	if strings.Contains(b.Draw(), "which kind of stopped") {
		t.Fatal("pushing down out of the band asked which kind")
	}
}

func TestPushingUpFromAColumnDoesNothing(t *testing.T) {
	f := twoTasks()
	b := hit(standing(f), "l", "i", "shift+k")
	if len(f.edited) != 0 {
		t.Fatalf("pushing up from a column wrote %v", f.edited)
	}
	_ = b
}

func TestPushingRightOutOfTheBandDoesNothing(t *testing.T) {
	f := twoTasks()
	b := hit(standing(f), "l", "j", "i", "shift+l")
	if len(f.edited) != 0 {
		t.Fatalf("pushing right out of the band wrote %v", f.edited)
	}
	_ = b
}

func TestPushingPastTheLastColumnDoesNothing(t *testing.T) {
	f := newFake()
	f.cards = []reading.Card{{Number: 1, Title: "Done", State: "done"}}
	b := hit(standing(f), "l", "l", "l", "i", "shift+l")
	if len(f.edited) != 0 {
		t.Fatalf("pushing past the last column wrote %v", f.edited)
	}
	_ = b
}

func TestPushingWithNothingHeldDoesNothing(t *testing.T) {
	f := newFake()
	b := standing(f)
	b.mode = Insert
	if _, cmd := b.push(goRight); cmd != nil {
		t.Fatal("a push with no task under the cursor still asked for a change")
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

func TestTheCursorFallsToTheNextOneAlong(t *testing.T) {
	f := twoTasks()
	b := hit(standing(f), "l", "i")
	fewer := reading.Snapshot{Cards: []reading.Card{{Number: 2, Title: "Second", State: "todo"}}}
	after := b.took(Looked{Snapshot: fewer})
	if after.Where().Task != 2 {
		t.Fatalf("the cursor fell to %d, want 2", after.Where().Task)
	}
}

func TestTheCursorFallsBackToTheOneBefore(t *testing.T) {
	f := twoTasks()
	b := hit(standing(f), "l", "i", "j")
	fewer := reading.Snapshot{Cards: []reading.Card{{Number: 1, Title: "First", State: "todo"}}}
	after := b.took(Looked{Snapshot: fewer})
	if after.Where().Task != 1 {
		t.Fatalf("the cursor fell to %d, want 1", after.Where().Task)
	}
}

func TestTheCursorLetsGoWhenTheColumnEmpties(t *testing.T) {
	f := twoTasks()
	b := hit(standing(f), "l", "i")
	after := b.took(Looked{Snapshot: reading.Snapshot{}})
	if after.Where().Mode != Normal || after.Where().Task != 0 {
		t.Fatalf("the cursor still holds %d in mode %v", after.Where().Task, after.Where().Mode)
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
	drawn := b.Draw()
	for _, want := range []string{"SUB-TASKS", "Child", "HISTORY", "created", "title"} {
		if !strings.Contains(drawn, want) {
			t.Fatalf("the detail never says %q:\n%s", want, drawn)
		}
	}
}

func TestTheDetailSaysWhenAHistoryCannotBeRead(t *testing.T) {
	f := twoTasks()
	f.fail = errNope
	b := standing(f)
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

func TestSearchingReachesTheDescriptionAndTheReason(t *testing.T) {
	f := newFake()
	f.cards = []reading.Card{
		{Number: 1, Title: "One", State: "todo", Description: "about the spec"},
		{Number: 2, Title: "Two", State: "todo", Because: "the client"},
		{Number: 3, Title: "Three", State: "todo"},
	}
	b := hit(standing(f), "/", "s", "p", "e", "c")
	if got := b.Where().Showing; got != 1 {
		t.Fatalf("a search of the description reached %d tasks, want 1", got)
	}
	b = hit(standing(f), "/", "c", "l", "i")
	if got := b.Where().Showing; got != 1 {
		t.Fatalf("a search of the reason reached %d tasks, want 1", got)
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

func TestSearchingTheObjectives(t *testing.T) {
	f := newFake()
	f.goals = []reading.Goal{
		{Number: 1, Title: "Ship the desktop board", State: "active"},
		{Number: 2, Title: "Cut the onboarding time", State: "active"},
	}
	b := hit(standing(f), "tab", "/", "s", "h", "i", "p")
	if !strings.Contains(b.Draw(), "Ship the desktop") {
		t.Fatalf("the search hid what it should have found:\n%s", b.Draw())
	}
	if strings.Contains(b.Draw(), "onboarding") {
		t.Fatalf("the search kept what it should have dropped:\n%s", b.Draw())
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

func TestWalkingRightToTheLastColumnAndNoFurther(t *testing.T) {
	f := newFake()
	f.cards = []reading.Card{
		{Number: 1, Title: "One", State: "backlog"},
		{Number: 2, Title: "Four", State: "done"},
	}
	b := hit(standing(f), "l", "l", "l")
	if b.Where().Column != lastColumn() {
		t.Fatalf("walking right stopped at %d, want %d", b.Where().Column, lastColumn())
	}
	b = hit(b, "l")
	if b.Where().Column != lastColumn() {
		t.Fatal("walking right went past the last column")
	}
	// And with a task in every column, the same walk in insert reaches the last
	// one and stops there rather than walking off the end.
	full := newFake()
	full.cards = []reading.Card{
		{Number: 1, Title: "One", State: "backlog"},
		{Number: 2, Title: "Two", State: "todo"},
		{Number: 3, Title: "Three", State: "current"},
		{Number: 4, Title: "Four", State: "done"},
	}
	held := hit(standing(full), "i", "l", "l", "l")
	if held.Where().Task != 4 {
		t.Fatalf("walking right in insert landed on %d, want the task in done", held.Where().Task)
	}
	held = hit(held, "l")
	if held.Where().Column != lastColumn() || held.Where().Task != 4 {
		t.Fatalf("walking right in insert went past the last column, to %d", held.Where().Column)
	}
}

func TestTheCursorFallsToWhateverStandsInItsPlace(t *testing.T) {
	f := newFake()
	f.cards = []reading.Card{
		{Number: 5, Title: "First", State: "todo"},
		{Number: 3, Title: "Second", State: "todo"},
		{Number: 9, Title: "Third", State: "todo"},
	}
	// The list is not in order of number, so a cursor that fell to the next
	// number would land somewhere else entirely.
	b := hit(standing(f), "l", "i", "j")
	fewer := reading.Snapshot{Cards: []reading.Card{
		{Number: 5, Title: "First", State: "todo"},
		{Number: 9, Title: "Third", State: "todo"},
	}}
	after := b.took(Looked{Snapshot: fewer})
	if after.Where().Task != 9 {
		t.Fatalf("the cursor fell to %d, want 9 — the one now standing where it was", after.Where().Task)
	}
}

func TestTheBeatIsTwoSeconds(t *testing.T) {
	if got := beat(); got != 2*time.Second {
		t.Fatalf("the board asks again every %v", got)
	}
}

func TestSteppingWithATaskThatIsNotInTheColumn(t *testing.T) {
	// Three tasks, so that "nowhere in the list" and "second in the list" lead
	// somewhere different: with two, stepping on from either lands nowhere and
	// the two cannot be told apart.
	f := newFake()
	f.cards = []reading.Card{
		{Number: 1, Title: "One", State: "todo"},
		{Number: 2, Title: "Two", State: "todo"},
		{Number: 3, Title: "Three", State: "todo"},
	}
	b := standing(f)
	b.mode, b.column, b.task = Insert, 1, 99
	if after := b.stepBy(goDown); after.Where().Task != 99 {
		t.Fatalf("stepping from a task that is not there landed on %d", after.Where().Task)
	}
}

func TestPushingLeftAndRightLandsInTheRightColumn(t *testing.T) {
	f := newFake()
	f.cards = []reading.Card{{Number: 1, Title: "One", State: "current"}}
	// current is the third column: left is todo, right is done. A push that
	// went the other way, or one column too far, would land somewhere else.
	left := hit(standing(f), "l", "l", "i")
	if _, cmd := left.pressed("shift+h"); cmd != nil {
		cmd()
	}
	if got := f.edited[1].State; got == nil || *got != "todo" {
		t.Fatalf("pushing left wrote %v, want todo", got)
	}
	right := hit(standing(f), "l", "l", "i")
	if _, cmd := right.pressed("shift+l"); cmd != nil {
		cmd()
	}
	if got := f.edited[1].State; got == nil || *got != "done" {
		t.Fatalf("pushing right wrote %v, want done", got)
	}
}

func TestPushingLeftReachesTheFirstColumn(t *testing.T) {
	f := newFake()
	f.cards = []reading.Card{{Number: 1, Title: "One", State: "todo"}}
	b := hit(standing(f), "l", "i")
	if _, cmd := b.pressed("shift+h"); cmd != nil {
		cmd()
	}
	if got := f.edited[1].State; got == nil || *got != "backlog" {
		t.Fatalf("pushing left off todo wrote %v, want backlog", got)
	}
}

func TestPushingRightReachesTheLastColumn(t *testing.T) {
	f := newFake()
	f.cards = []reading.Card{{Number: 1, Title: "One", State: "current"}}
	b := hit(standing(f), "l", "l", "i")
	if _, cmd := b.pressed("shift+l"); cmd != nil {
		cmd()
	}
	if got := f.edited[1].State; got == nil || *got != "done" {
		t.Fatalf("pushing right off current wrote %v, want done", got)
	}
}

func TestTheCursorReachesTheFirstColumnAndTheLast(t *testing.T) {
	b := standing(twoTasks())
	b = hit(b, "l", "l", "l")
	if b.Where().Column != lastColumn() {
		t.Fatalf("walking right stopped at %d", b.Where().Column)
	}
	b = hit(b, "h", "h", "h")
	if b.Where().Column != 0 {
		t.Fatalf("walking left stopped at %d", b.Where().Column)
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

func TestWalkingTheObjectives(t *testing.T) {
	b := hit(standing(manyGoals()), "tab")
	if b.Where().Goal != 1 {
		t.Fatalf("the cursor landed on %d, want the first", b.Where().Goal)
	}
	b = hit(b, "j")
	if b.Where().Goal != 2 {
		t.Fatalf("j moved to %d, want 2", b.Where().Goal)
	}
	b = hit(b, "k")
	if b.Where().Goal != 1 {
		t.Fatalf("k moved to %d, want 1", b.Where().Goal)
	}
	b = hit(b, "k")
	if b.Where().Goal != 1 {
		t.Fatal("k walked past the first objective")
	}
	for range 12 {
		b = hit(b, "j")
	}
	if b.Where().Goal != 9 {
		t.Fatalf("j walked to %d, want the last", b.Where().Goal)
	}
}

func TestSidewaysDoesNothingOnTheObjectives(t *testing.T) {
	// The list is one column. A key that pointed sideways and moved something
	// would be a key that looked broken.
	b := hit(standing(manyGoals()), "tab", "j")
	if after := hit(b, "l", "h"); after.Where().Goal != b.Where().Goal {
		t.Fatalf("sideways moved the cursor to %d", after.Where().Goal)
	}
}

func TestWalkingAnEmptyObjectivesList(t *testing.T) {
	b := hit(standing(newFake()), "tab", "j", "k")
	if b.Where().Goal != 0 {
		t.Fatalf("the cursor landed on %d where there are none", b.Where().Goal)
	}
}

func TestTheObjectivesListScrollsToTheCursor(t *testing.T) {
	b := hit(standing(manyGoals()), "tab")
	if b.goalAt != 0 {
		t.Fatalf("the list opened scrolled to %d", b.goalAt)
	}
	for range 8 {
		b = hit(b, "j")
	}
	if b.goalAt == 0 {
		t.Fatal("the list never scrolled, so the last objective is off screen")
	}
	for range 8 {
		b = hit(b, "k")
	}
	if b.goalAt != 0 {
		t.Fatalf("coming back up left the list scrolled to %d", b.goalAt)
	}
}

func TestAColumnScrollsToTheCursor(t *testing.T) {
	f := newFake()
	for i := 1; i <= 12; i++ {
		f.cards = append(f.cards, reading.Card{Number: i, Title: "Task " + strconv.Itoa(i), State: "todo"})
	}
	b := hit(standing(f), "l", "i")
	if b.offsets[1] != 0 {
		t.Fatalf("the column opened scrolled to %d", b.offsets[1])
	}
	for range 11 {
		b = hit(b, "j")
	}
	if b.offsets[1] == 0 {
		t.Fatal("the column never scrolled, so the last task is off screen")
	}
	// And every card is reachable, which is what scrolling is for.
	if b.Where().Task != 12 {
		t.Fatalf("the cursor reached %d, want the last of twelve", b.Where().Task)
	}
	for range 11 {
		b = hit(b, "k")
	}
	if b.offsets[1] != 0 {
		t.Fatalf("coming back up left the column scrolled to %d", b.offsets[1])
	}
}

func TestTheBandScrollsToTheCursor(t *testing.T) {
	f := newFake()
	f.cards = append(f.cards, reading.Card{Number: 1, Title: "One", State: "todo"})
	for i := 20; i < 32; i++ {
		f.cards = append(f.cards, reading.Card{Number: i, Title: "Stalled " + strconv.Itoa(i), State: "blocked"})
	}
	b := hit(standing(f), "l", "i", "j")
	if !b.Where().InBand {
		t.Fatal("j off the last task did not reach the band")
	}
	for range 11 {
		b = hit(b, "j")
	}
	if b.bandAt == 0 {
		t.Fatal("the band never scrolled, so what is at its bottom is unreachable")
	}
}

func TestTheCursorFollowsInsideAnObjective(t *testing.T) {
	f := newFake()
	f.goals = []reading.Goal{{Number: 1, Title: "Ship it", State: "active"}}
	for i := 1; i <= 8; i++ {
		f.cards = append(f.cards, reading.Card{Number: i, Title: "Task " + strconv.Itoa(i), State: "todo", Objectives: []int{1}})
	}
	b := hit(standing(f), "tab", "enter", "l", "i")
	for range 7 {
		b = hit(b, "j")
	}
	if b.Where().Task != 8 {
		t.Fatalf("the cursor reached %d inside the objective, want the last", b.Where().Task)
	}
	if b.offsets[1] == 0 {
		t.Fatal("the column inside an objective never scrolled")
	}
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

func TestThePopupMarksTheOneItIsOn(t *testing.T) {
	b := standing(twoTasks())
	b.skin = SkinOf(Default)
	b.choosing = true
	first := b.kindPopup()
	b.choice = 1
	second := b.kindPopup()
	if first == second {
		t.Fatal("the popup looks the same whichever answer it is on")
	}
	if !strings.Contains(first, b.skin.Picked.Render(" "+kinds[0]+" ")) {
		t.Fatalf("the first answer is not the one picked out:\n%q", first)
	}
	if !strings.Contains(second, b.skin.Picked.Render(" "+kinds[1]+" ")) {
		t.Fatalf("the second answer is not the one picked out:\n%q", second)
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

func TestWalkingAlongTheBandsFirstRow(t *testing.T) {
	f := newFake()
	f.cards = []reading.Card{
		{Number: 1, Title: "One", State: "todo"},
		{Number: 2, Title: "Blocked one", State: "blocked"},
		{Number: 3, Title: "Blocked two", State: "blocked"},
	}
	// Two stalled cards make one row, so along it is right and left — down
	// would be a row that is not there.
	b := hit(standing(f), "l", "i", "j", "l")
	if !b.Where().InBand || b.Where().Task != 3 {
		t.Fatalf("the cursor is on %d, band %v, want the second stalled card", b.Where().Task, b.Where().InBand)
	}
	b = hit(b, "h")
	if !b.Where().InBand || b.Where().Task != 2 {
		t.Fatalf("left inside the band left it: task %d, band %v", b.Where().Task, b.Where().InBand)
	}
	b = hit(b, "k")
	if b.Where().InBand || b.Where().Task != 1 {
		t.Fatalf("up off the top row landed on %d, band %v", b.Where().Task, b.Where().InBand)
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
	b := standing(f)
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

func TestTheBandScrollsByTheRowACardIsIn(t *testing.T) {
	// Three cards to a row, so the seventh is in the third row. A band that
	// counted rows any other way would scroll to the wrong place.
	f := newFake()
	f.cards = append(f.cards, reading.Card{Number: 1, Title: "One", State: "todo"})
	for i := 20; i < 32; i++ {
		f.cards = append(f.cards, reading.Card{Number: i, Title: "Stalled", State: "blocked"})
	}
	// Down twice from the first card is two rows down, which is the seventh
	// card: three to a row.
	b := hit(standing(f), "l", "i", "j", "j", "j")
	if b.Where().Task != 26 {
		t.Fatalf("two rows down landed on %d, want the seventh stalled card", b.Where().Task)
	}
	l := LayoutFor(b.width, b.height)
	if want := Scrolled(0, 2, l.Rows); b.bandAt != want {
		t.Fatalf("the band scrolled to %d for the third row, want %d", b.bandAt, want)
	}
}

func TestTheHistoryIsGivenWhatTheWindowLeavesIt(t *testing.T) {
	f := newFake()
	f.cards = []reading.Card{{Number: 1, Title: "One", State: "todo"}}
	for i := range 40 {
		f.acts = append(f.acts, anAct("changed", strconv.Itoa(i)))
	}
	for _, height := range []int{28, 40} {
		b := standing(f)
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
	got := standing(f).history(1, 40, 3)
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

func TestPushingDownMovesTheTaskOnePlaceLower(t *testing.T) {
	f := newFake()
	f.cards = []reading.Card{
		{Number: 1, Title: "One", State: "todo"},
		{Number: 2, Title: "Two", State: "todo"},
		{Number: 3, Title: "Three", State: "todo"},
	}
	b := hit(standing(f), "l", "i")
	if _, cmd := b.pressed("shift+j"); cmd != nil {
		cmd()
	}
	if len(f.moved) != 1 {
		t.Fatalf("pushing down asked for %v", f.moved)
	}
	if got := f.moved[0]; got.number != 1 || got.at != engine.Behind || got.target != 2 {
		t.Fatalf("pushing the first down asked for %+v, want task 1 behind task 2", got)
	}
	if len(f.edited) != 0 {
		t.Fatalf("pushing down changed the column: %v", f.edited)
	}
}

func TestPushingUpMovesTheTaskOnePlaceHigher(t *testing.T) {
	f := newFake()
	f.cards = []reading.Card{
		{Number: 1, Title: "One", State: "todo"},
		{Number: 2, Title: "Two", State: "todo"},
		{Number: 3, Title: "Three", State: "todo"},
	}
	// The third: it goes behind the first, which is one place higher.
	b := hit(standing(f), "l", "i", "j", "j")
	if _, cmd := b.pressed("shift+k"); cmd != nil {
		cmd()
	}
	if got := f.moved[0]; got.number != 3 || got.at != engine.Behind || got.target != 1 {
		t.Fatalf("pushing the third up asked for %+v, want task 3 behind task 1", got)
	}
	// And the second goes to the head, because there is nothing to sit behind.
	f.moved = nil
	second := hit(standing(f), "l", "i", "j")
	if _, cmd := second.pressed("shift+k"); cmd != nil {
		cmd()
	}
	if got := f.moved[0]; got.number != 2 || got.at != engine.First {
		t.Fatalf("pushing the second up asked for %+v, want task 2 first", got)
	}
}

func TestPushingTheFirstTaskUpDoesNothing(t *testing.T) {
	f := twoTasks()
	b := hit(standing(f), "l", "i")
	if _, cmd := b.pressed("shift+k"); cmd != nil {
		t.Fatal("pushing the first task up asked for something")
	}
	_ = b
}

func TestAMoveThatRefusesIsSaid(t *testing.T) {
	f := twoTasks()
	f.failOp = errRefuse
	b := hit(standing(f), "l", "i")
	_, cmd := b.pressed("shift+j")
	if looked, ok := cmd().(Looked); !ok || looked.Err == nil {
		t.Fatal("a move that was refused was passed over")
	}
}

func TestTheBandIsWalkedAsAGrid(t *testing.T) {
	f := newFake()
	f.cards = append(f.cards, reading.Card{Number: 1, Title: "One", State: "todo"})
	for i := 20; i < 27; i++ {
		f.cards = append(f.cards, reading.Card{Number: i, Title: "Stalled", State: "blocked"})
	}
	// Seven stalled cards, three to a row: 20 21 22 / 23 24 25 / 26.
	b := hit(standing(f), "l", "i", "j")
	if b.Where().Task != 20 {
		t.Fatalf("the band opened on %d, want the first", b.Where().Task)
	}
	b = hit(b, "l")
	if b.Where().Task != 21 {
		t.Fatalf("right moved to %d, want along the row", b.Where().Task)
	}
	b = hit(b, "j")
	if b.Where().Task != 24 {
		t.Fatalf("down moved to %d, want down a row", b.Where().Task)
	}
	b = hit(b, "h")
	if b.Where().Task != 23 {
		t.Fatalf("left moved to %d, want back along the row", b.Where().Task)
	}
	b = hit(b, "l", "l")
	if b.Where().Task != 25 {
		t.Fatal("right walked off the end of the row")
	}
	b = hit(b, "k")
	if b.Where().Task != 22 {
		t.Fatalf("up moved to %d, want up a row", b.Where().Task)
	}
	b = hit(b, "k")
	if b.Where().InBand {
		t.Fatal("up off the top row stayed in the band")
	}
}

func TestTheBandStopsAtTheEndOfWhatIsInIt(t *testing.T) {
	f := newFake()
	f.cards = append(f.cards, reading.Card{Number: 1, Title: "One", State: "todo"})
	for i := 20; i < 24; i++ {
		f.cards = append(f.cards, reading.Card{Number: i, Title: "Stalled", State: "blocked"})
	}
	// Four stalled: 20 21 22 / 23. Down from 22 has no card under it.
	b := hit(standing(f), "l", "i", "j", "l", "l")
	if b.Where().Task != 22 {
		t.Fatalf("the cursor is on %d, want the last of the first row", b.Where().Task)
	}
	if after := hit(b, "j"); after.Where().Task != 22 {
		t.Fatalf("down with nothing under it moved to %d", after.Where().Task)
	}
	if after := hit(b, "l"); after.Where().Task != 22 {
		t.Fatal("right walked off the end of the row")
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

func TestTheBandsRowsAreNotWalkedIntoOneAnother(t *testing.T) {
	f := newFake()
	f.cards = append(f.cards, reading.Card{Number: 1, Title: "One", State: "todo"})
	for i := 20; i < 26; i++ {
		f.cards = append(f.cards, reading.Card{Number: i, Title: "Stalled", State: "blocked"})
	}
	// Six stalled: 20 21 22 / 23 24 25. Left from the head of the second row is
	// the end of the first, and walking there would move the cursor up a row
	// while pointing sideways.
	b := hit(standing(f), "l", "i", "j", "j")
	if b.Where().Task != 23 {
		t.Fatalf("down a row landed on %d, want the head of the second row", b.Where().Task)
	}
	if after := hit(b, "h"); after.Where().Task != 23 {
		t.Fatalf("left off the head of a row moved to %d", after.Where().Task)
	}
	// And up from the head of the second row is the head of the first, not out
	// of the band.
	up := hit(b, "k")
	if !up.Where().InBand || up.Where().Task != 20 {
		t.Fatalf("up from the second row landed on %d, band %v", up.Where().Task, up.Where().InBand)
	}
}

func aColumnOf(n int) *fake {
	f := newFake()
	for i := 1; i <= n; i++ {
		f.cards = append(f.cards, reading.Card{Number: i, Title: "Task " + strconv.Itoa(i), State: "todo"})
	}
	return f
}

func TestWalkingSidewaysKeepsTheLevel(t *testing.T) {
	f := aColumnOf(4)
	f.cards = append(f.cards,
		reading.Card{Number: 11, Title: "Also one", State: "current"},
		reading.Card{Number: 12, Title: "Also two", State: "current"},
		reading.Card{Number: 13, Title: "Also three", State: "current"},
	)
	b := hit(standing(f), "l", "i", "j", "j")
	if b.Where().Task != 3 {
		t.Fatalf("the cursor is on %d, want the third of todo", b.Where().Task)
	}
	b = hit(b, "l")
	if b.Where().Task != 13 {
		t.Fatalf("walking across landed on %d, want the third of current", b.Where().Task)
	}
	// And back, at the same level again.
	if back := hit(b, "h"); back.Where().Task != 3 {
		t.Fatalf("walking back landed on %d, want the third of todo", back.Where().Task)
	}
}

func TestWalkingIntoAShorterColumnStopsAtItsEnd(t *testing.T) {
	f := aColumnOf(5)
	f.cards = append(f.cards, reading.Card{Number: 11, Title: "Only one", State: "current"})
	b := hit(standing(f), "l", "i", "j", "j", "j", "j")
	if b.Where().Task != 5 {
		t.Fatalf("the cursor is on %d, want the fifth", b.Where().Task)
	}
	b = hit(b, "l")
	if b.Where().Task != 11 {
		t.Fatalf("walking into a column of one landed on %d", b.Where().Task)
	}
}

func TestGoesToTheEndOfAColumnAndBackToItsStart(t *testing.T) {
	b := hit(standing(aColumnOf(6)), "l", "i")
	b = hit(b, "shift+g")
	if b.Where().Task != 6 {
		t.Fatalf("G landed on %d, want the last", b.Where().Task)
	}
	b = hit(b, "g", "g")
	if b.Where().Task != 1 {
		t.Fatalf("gg landed on %d, want the first", b.Where().Task)
	}
}

func TestOneGOnItsOwnDoesNothingAndIsForgotten(t *testing.T) {
	b := hit(standing(aColumnOf(6)), "l", "i")
	after := hit(b, "g")
	if after.Where().Task != 1 || after.pending != "g" {
		t.Fatalf("one g moved to %d, pending %q", after.Where().Task, after.pending)
	}
	// Something else clears it, so the next g is a first g and not a second.
	after = hit(after, "j")
	if after.pending != "" {
		t.Fatalf("a key after a g left it pending as %q", after.pending)
	}
	if after.Where().Task != 2 {
		t.Fatalf("the key after a g was swallowed: the cursor is on %d", after.Where().Task)
	}
	if again := hit(after, "g"); again.Where().Task != 2 {
		t.Fatalf("a g after that jumped to %d", again.Where().Task)
	}
}

func TestGoesToTheEndsOfTheBandToo(t *testing.T) {
	f := newFake()
	f.cards = append(f.cards, reading.Card{Number: 1, Title: "One", State: "todo"})
	for i := 20; i < 26; i++ {
		f.cards = append(f.cards, reading.Card{Number: i, Title: "Stalled", State: "blocked"})
	}
	b := hit(standing(f), "l", "i", "j", "shift+g")
	if b.Where().Task != 25 {
		t.Fatalf("G in the band landed on %d, want the last", b.Where().Task)
	}
	if start := hit(b, "g", "g"); start.Where().Task != 20 {
		t.Fatalf("gg in the band landed on %d, want the first", start.Where().Task)
	}
}

func TestGoesToTheEndsOfTheObjectives(t *testing.T) {
	b := hit(standing(manyGoals()), "tab", "shift+g")
	if b.Where().Goal != 9 {
		t.Fatalf("G landed on objective %d, want the last", b.Where().Goal)
	}
	if start := hit(b, "g", "g"); start.Where().Goal != 1 {
		t.Fatalf("gg landed on objective %d, want the first", start.Where().Goal)
	}
}

func TestGIsALetterWhereALetterIsWhatItMeans(t *testing.T) {
	// At a prompt, in a form, and on a column where nothing is held: a g there
	// is a g, and jumping has no business reading it.
	typed := hit(standing(aColumnOf(3)), "/", "g")
	if !strings.Contains(bare(typed.Draw()), "/g") {
		t.Fatalf("a g at the prompt did not reach it:\n%s", bare(typed.Draw()))
	}
	onColumn := hit(standing(aColumnOf(3)), "shift+g")
	if onColumn.Where().Task != 0 {
		t.Fatalf("G on a column took hold of task %d", onColumn.Where().Task)
	}
	inForm := hit(standing(aColumnOf(3)), "l", "i", "e", "g")
	if inForm.screen != OnForm {
		t.Fatal("a g in the form closed it")
	}
	if !strings.Contains(inForm.form.valueOf("title"), "g") {
		t.Fatalf("a g in the form did not reach the field: %q", inForm.form.valueOf("title"))
	}
	choosing := hit(standing(aColumnOf(1)), "l", "i", "shift+j")
	if after := hit(choosing, "shift+g"); !after.choosing {
		t.Fatal("G closed the popup")
	}
}

func TestGoesNowhereWhenThereIsNothingToGoTo(t *testing.T) {
	f := newFake()
	b := standing(f)
	b.mode = Insert
	if after := hit(b, "shift+g"); after.Where().Task != 0 {
		t.Fatalf("G on an empty column took hold of %d", after.Where().Task)
	}
	goals := hit(standing(newFake()), "tab")
	if after := hit(goals, "shift+g"); after.Where().Goal != 0 {
		t.Fatalf("G on an empty list took hold of %d", after.Where().Goal)
	}
}

// The bar offers ":" and the prompt knew two words out of the whole engine. A
// command typed there was answered "unknown" however well the command line knew it.
func TestTheCommandPromptKnowsWhatTheEngineKnows(t *testing.T) {
	f := twoTasks()
	b := hit(standing(f), ":")
	for _, r := range "task focus 1" {
		b = hit(b, string(r))
	}
	if _, cmd := b.pressed("enter"); cmd != nil {
		cmd()
	}
	if len(f.asked) == 0 || f.asked[0].op.Verb != "focus" {
		t.Fatalf("the prompt asked %v, want a task focus", f.asked)
	}
}

// What the prompt refuses, and where it refuses it: a word where no argument
// goes, and a required one left out. Both are said rather than run.
func TestTheCommandPromptRefusesWhatTheEngineWouldNotUnderstand(t *testing.T) {
	for _, c := range []struct {
		what string
		line string
	}{
		{"a word past the last argument", "task show 1 2"},
		{"a flag with nothing after it", "task show --number"},
		{"an argument that does not exist", "task show --nope 1"},
		{"a required argument left out", "task add"},
	} {
		t.Run(c.what, func(t *testing.T) {
			f := twoTasks()
			b := hit(standing(f), ":")
			for _, r := range c.line {
				if r == ' ' {
					b = hit(b, "space")
					continue
				}
				b = hit(b, string(r))
			}
			if _, cmd := b.pressed("enter"); cmd != nil {
				cmd()
			}
			if len(f.asked) != 0 {
				t.Fatalf("%q was run: %v", c.line, f.asked)
			}
		})
	}
}

// A sweep the store refuses says so rather than redrawing as though a column
// had gone off the board.
func TestASweepTheStoreRefusesIsSaid(t *testing.T) {
	f := twoTasks()
	f.failOp = errRefuse
	b := hit(standing(f), "shift+x")
	_, cmd := b.pressed("enter")
	if cmd == nil {
		t.Fatal("agreeing to the sweep asked for nothing")
	}
	looked, ok := cmd().(Looked)
	if !ok || looked.Err == nil {
		t.Fatalf("a sweep that was refused answered %v", looked)
	}
}

// Two words are a command: the thing and the verb, with nothing after them.
func TestACommandOfTwoWordsRuns(t *testing.T) {
	f := twoTasks()
	b := hit(standing(f), ":")
	for _, r := range "task list" {
		if r == ' ' {
			b = hit(b, "space")
			continue
		}
		b = hit(b, string(r))
	}
	if _, cmd := b.pressed("enter"); cmd != nil {
		cmd()
	}
	if len(f.asked) == 0 || f.asked[0].op.Verb != "list" {
		t.Fatalf("the prompt asked %v, want a task list", f.asked)
	}
}

// What the prompt runs and the store refuses is said, the way every other
// gesture's refusal is.
func TestACommandTheStoreRefusesIsSaid(t *testing.T) {
	f := twoTasks()
	f.failOp = errRefuse
	b := hit(standing(f), ":")
	for _, r := range "task focus 1" {
		if r == ' ' {
			b = hit(b, "space")
			continue
		}
		b = hit(b, string(r))
	}
	_, cmd := b.pressed("enter")
	if cmd == nil {
		t.Fatal("the prompt asked for nothing")
	}
	looked, ok := cmd().(Looked)
	if !ok || looked.Err == nil {
		t.Fatalf("a command that was refused answered %v", looked)
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
