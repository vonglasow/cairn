package tui

import (
	"cairn/internal/board"
	"cairn/internal/tui/geom"
	"strconv"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"

	"cairn/internal/engine"
	"cairn/internal/engine/act"
)

func twoTasks() *fake {
	f := newFake()
	f.cards = []board.Card{
		{Number: 1, Title: "First", State: "todo"},
		{Number: 2, Title: "Second", State: "todo"},
		{Number: 3, Title: "Waiting", State: "blocked"},
	}
	return f
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

func TestTheObjectivesTabWithNothingOnIt(t *testing.T) {
	b := hit(standing(twoTasks()), "tab")
	if b.Where().Goal != 0 {
		t.Fatalf("the cursor landed on objective %d where there are none", b.Where().Goal)
	}
	if b = hit(b, "enter"); b.Where().Showing == 0 {
		t.Fatal("opening nothing narrowed the board to nothing")
	}
}

func TestAMessageNobodySentIsIgnored(t *testing.T) {
	b := standing(twoTasks())
	model, cmd := b.Update("something else entirely")
	if cmd != nil || model.(Board).Draw() != b.Draw() {
		t.Fatal("a message the board knows nothing about changed it")
	}
}

func TestNothingIsReadWhileTheStoreHasNotMoved(t *testing.T) {
	f := twoTasks()
	f.version = 9
	b := standing(f)
	if msg := b.wonder(false)(); msg != nil {
		t.Fatalf("the board read the store again for nothing: %v", msg)
	}
	f.version = 10
	if msg := b.wonder(false)(); msg == nil {
		t.Fatal("the board did not read the store after it moved")
	}
	f.fail = errNope
	if msg := b.wonder(false)(); msg == nil {
		t.Fatal("a store that refused was passed over in silence")
	}
}

func TestATaskInFocusIsReadAgainEachMinute(t *testing.T) {
	f := twoTasks()
	f.version = 9
	b := standing(f)
	start := time.Date(2026, 11, 1, 9, 0, 0, 0, time.UTC)
	b.lookedAt = start
	if b.focusAged(start.Add(time.Hour)) {
		t.Fatal("a board with no task in focus reads the store again for its time")
	}
	b.snap.Focus = []int{1}
	if b.focusAged(start.Add(30 * time.Second)) {
		t.Error("the time in focus was read again before the minute shown had passed")
	}
	if !b.focusAged(start.Add(time.Minute)) {
		t.Error("the time in focus stood still past a minute")
	}
	if msg := b.wonder(true)(); msg == nil {
		t.Error("a board whose time in focus aged did not read the store")
	}
	next, _ := b.Update(Beat(start.Add(time.Minute)))
	if next.(Board).lookedAt != start.Add(time.Minute) {
		t.Error("the beat that read the time in focus again did not note when")
	}
}

func TestTheDetailOfATaskThatWentAway(t *testing.T) {
	b := standing(twoTasks())
	b.screen = board.OnDetail
	if !strings.Contains(b.Draw(), "no such task") {
		t.Fatalf("the detail of nothing says:\n%s", b.Draw())
	}
}

func TestTheDetailShowsStepsAndHistory(t *testing.T) {
	f := newFake()
	f.cards = []board.Card{
		{Number: 1, Title: "Parent", State: "todo", SubDone: 1, SubTotal: 1, Labels: []string{"spec"}, Estimate: "3h", Spent: "1h", Deadline: on("2026-09-02"),
			Steps: []engine.Part{{Number: 1, Title: "Child", Done: true}}},
	}
	var told act.AsTold
	told.On(1)
	told.Did("created")
	told.Changing("title")
	told.Became("Parent")
	told.At(today)
	f.acts = []act.Act{told.Act()}
	b := hit(standing(f), "ctrl+l", "enter")
	drawn := strip(b.Draw())
	for _, want := range []string{"STEPS", "Child", "TRACE", "created", "title"} {
		if !strings.Contains(drawn, want) {
			t.Fatalf("the detail never says %q:\n%s", want, drawn)
		}
	}
}

func TestTheDetailSaysWhenAHistoryCannotBeRead(t *testing.T) {
	f := twoTasks()
	f.fail = errNope
	b := heardOf(standing(f), 1)
	b.task, b.screen = 1, board.OnDetail
	if !strings.Contains(b.Draw(), errNope.Error()) {
		t.Fatalf("a history that could not be read was passed over:\n%s", b.Draw())
	}
}

func manyGoals() *fake {
	f := newFake()
	for i := 1; i <= 9; i++ {
		f.goals = append(f.goals, board.Goal{Number: i, Title: "Objective " + strconv.Itoa(i), State: "active"})
	}
	return f
}

func TestAWindowOfOneRowStillDrawsARow(t *testing.T) {
	if got := through("one\ntwo\nthree", 10, 1, 0); got != "one       " {
		t.Fatalf("a window one row tall drew %q", got)
	}
	if got := through("one", 10, 0, 0); got != "" {
		t.Fatalf("a window no rows tall drew %q", got)
	}
}

func TestTheHistoryShowsTheLatestAndScrollsForTheRest(t *testing.T) {
	f := newFake()
	f.cards = []board.Card{{Number: 1, Title: "One", State: "todo"}}
	for i := range 9 {
		f.acts = append(f.acts, aChange("title", "act-"+strconv.Itoa(i)))
	}
	b := heardOf(standing(f), 1)
	b.task = 1
	got := b.history(1, 60, 3)
	if len(got) != 3 {
		t.Fatalf("a window three rows tall showed %d acts", len(got))
	}
	if !strings.Contains(strip(got[2]), "act-8") {
		t.Fatalf("the last act shown is %q, want the latest", got[2])
	}
	if all := b.history(1, 40, 0); len(all) != 9 {
		t.Fatalf("with no window %d acts came back, want all nine", len(all))
	}
	if one := b.history(1, 40, 1); len(one) != 1 {
		t.Fatalf("a window one row tall showed %d acts", len(one))
	}
}

func TestEveryScreenFitsTheWindowItIsDrawnIn(t *testing.T) {
	f := fullBoard()
	for _, keys := range [][]string{
		nil, {"l"}, {"l", "i"}, {"?"}, {"l", "i", "enter"}, {"l", "i", "e"},
		{"l", "a"}, {"tab"}, {"tab", "?"}, {"tab", "e"}, {"tab", "a"},
		{"tab", "enter"}, {"/", "s"}, {":", "b"}, {"l", "j"},
	} {
		for _, size := range [][2]int{{geom.Narrowest, geom.Shortest}, {72, 30}, {100, 44}, {160, 60}} {
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

func TestTheHelpCanBeReadPastTheBottomOfTheWindow(t *testing.T) {
	b := standing(fullBoard())
	b.width, b.height = geom.Narrowest, geom.Shortest
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

func TestAHistoryLineSaysWhoDidItWhenItWasNotYou(t *testing.T) {
	b := standing(newFake())
	var told act.AsTold
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
