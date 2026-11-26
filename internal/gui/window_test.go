package gui

import (
	"cairn/internal/board"
	"cairn/internal/gui/geom"
	"fmt"

	"strconv"
	"strings"
	"testing"

	"gioui.org/io/pointer"

	"cairn/internal/engine"
)

func TestAClickOnNothingChoosesNothing(t *testing.T) {
	f := newFake()
	w := holding(f, card(1, "Write the spec", "todo"))
	w = w.clicked(middleOf(w, 1, 0))
	w = w.clicked(geom.Point{X: 0, Y: 0})
	if w.Chosen != 0 {
		t.Fatalf("clicking the margin left %d chosen", w.Chosen)
	}
}

func TestAWindowTooSmallSaysSo(t *testing.T) {
	f := newFake()
	w := New(f, 320, 240)
	if w.Screen != board.TooSmallToDraw {
		t.Fatalf("the window is on %v", w.Screen)
	}
	w = w.Resized(1200, 800)
	if w.Screen != board.OnTheBoard {
		t.Fatalf("growing back left the window on %v", w.Screen)
	}
}

func TestDeletingAnObjective(t *testing.T) {
	f := newFake()
	w := New(f, 1200, 800)
	w.Snap = board.Snapshot{Goals: []board.Goal{{Number: 3, Title: "Ship cairn", State: "active"}}}
	w.Tab, w.Chosen, w.ChosenGoal = board.Objectives, 3, true

	w = w.Choose("Delete").Choose("Delete")
	if f.what("objective", "delete") != 1 {
		t.Errorf("it asked for %d objective deletes", f.what("objective", "delete"))
	}
}

func TestWhatEachPanelSaysItIs(t *testing.T) {
	f := newFake()
	w := holding(f, board.Card{Number: 7, Title: "Write the spec", State: "todo"})
	if got := Heads(w); got != "" {
		t.Errorf("the board is headed %q, and it is its own heading", got)
	}
	on := w.clicked(middleOf(w, 1, 0))
	if got := Heads(on.Key("enter").Types("x")); got != "task 7 · Write the spec" {
		t.Errorf("a task being edited is headed %q", got)
	}
	if got := Heads(on.Choose("Delete")); got != "Delete task 7?" {
		t.Errorf("the question is headed %q", got)
	}
	if got := Heads(w.Key("?")); got != "Help" {
		t.Errorf("the help is headed %q", got)
	}
	for screen, want := range map[board.Screen]string{board.OnFiltering: "Filters", board.OnLabels: "Labels"} {
		if got := HeadsOn(w, screen); got != want {
			t.Errorf("a panel is headed %q, not its name, %q", got, want)
		}
	}
	if got := Heads(w.Key("cmd+n")); got != "New task" {
		t.Errorf("adding a task is headed %q", got)
	}
	second := w.TabsAt()[1]
	goals := w.clicked(geom.Point{X: second.X + second.W/2, Y: second.Y + second.H/2})
	if got := Heads(goals.Key("cmd+n")); got != "New objective" {
		t.Errorf("adding an objective is headed %q", got)
	}
}

func TestTheWheelMovesAHistoryByWholeLines(t *testing.T) {
	f := newFake()
	for range 40 {
		f.acts = append(f.acts, anAct("moved", "column", "todo", "current"))
	}
	w := holding(f, board.Card{Number: 1, Title: "Write the spec", State: "todo"})
	w = w.clicked(middleOf(w, 1, 0)).Key("enter").Key("cmd+t")
	over := BesideBox(w.formPanel())
	at := geom.Point{X: over.X + over.W/2, Y: over.Y + over.H/2}

	if got := w.Scroll(at, 3*lineHigh).History(); got != 3 {
		t.Errorf("turned by three lines, the history is at %d", got)
	}
	if got := w.Scroll(at, lineHigh-1).History(); got != 0 {
		t.Errorf("turned by less than a line, the history is at %d", got)
	}
	if got := w.Scroll(geom.Point{X: 1, Y: 1}, 9*lineHigh).History(); got != 0 {
		t.Errorf("turned away from the history, it is at %d", got)
	}
}

func TestEveryBarCanBePulled(t *testing.T) {
	f := newFake()
	for range 40 {
		f.acts = append(f.acts, anAct("moved", "column", "todo", "current"))
	}
	cards := []board.Card{{Number: 1, Title: "Ship it", State: "todo",
		Description: strings.Repeat("one more line of description ", 40)}}
	for i := range 6 {
		cards[0].Steps = append(cards[0].Steps, engine.Part{Number: i + 1, Title: fmt.Sprintf("Step %d", i+1)})
	}
	w := New(f, 1200, 800)
	w.Snap = board.Snapshot{Cards: cards}
	w = w.clicked(middleOf(w, 1, 0)).Key("enter").Key("cmd+t")

	box := w.formPanel()
	_, saying, _, _, _ := w.sayingOn()
	for _, c := range []struct {
		what  string
		track geom.Box
		read  func(Window) int
	}{
		{"the history", BesideBar(box), func(w Window) int { return w.History() }},
		{"the description", saying, func(w Window) int { return w.Saying() }},
	} {
		middle := geom.Point{X: c.track.X + c.track.W/2, Y: c.track.Y + 4}
		far := geom.Point{X: middle.X, Y: c.track.Y + c.track.H + 999}
		pulled := w.PressAt(middle).MoveTo(far).LetGo(far)
		if c.read(pulled) == c.read(w) {
			t.Errorf("pulling the bar beside %s left it at %d", c.what, c.read(w))
		}
	}
}

func TestTheWindowKnowsWhatThePointerIsOver(t *testing.T) {
	w := New(newFake(), 1200, 800).Wonder()
	if _, over := w.Hovered(); over {
		t.Fatal("something is hovered before the pointer has moved")
	}
	first := w.BarOffers()[0]
	w = Pointed(w, pointer.Move, geom.Point{X: first.X + first.W/2, Y: first.Y + first.H/2}, 0)
	said, over := w.Hovered()
	if !over || said != w.Offers()[0] {
		t.Fatalf("the pointer is over %q %v, want the first offer", said, over)
	}
	w = Pointed(w, pointer.Leave, geom.Point{}, 0)
	if _, over := w.Hovered(); over {
		t.Fatal("the pointer left the window and something is still hovered")
	}
}

func TestTheListsBarCanBePulled(t *testing.T) {
	f := newFake()
	var goals []board.Goal
	for i := 1; i <= 40; i++ {
		goals = append(goals, board.Goal{Number: i, Title: "Objective " + strconv.Itoa(i)})
	}
	w := New(f, 1200, 800)
	w.Snap = board.Snapshot{Cards: []board.Card{{Number: 1, Title: "One", State: "todo"}}, Goals: goals}
	w = w.clicked(middleOf(w, 1, 0)).Choose("Edit")
	for i, field := range w.Form.Fields {
		if field.Arg == "objective" {
			w.Form = w.Form.Onto(i)
		}
	}
	w = w.Choose("Choose")
	if w.Screen != board.OnChoosing {
		t.Fatalf("the list did not open: %v", w.Screen)
	}
	box := w.ChoosingBox()
	track := ChoosingBar(box)
	w = w.PressAt(geom.Point{X: track.X + track.W/2, Y: track.Y + 2}).
		MoveTo(geom.Point{X: track.X + track.W/2, Y: track.Y + track.H - 2})
	if w.ChoosingFromRow() == 0 {
		t.Fatal("pulling the list's bar to the bottom left it at the top")
	}
}

func TestTheHelpNamesEveryKeyTheBarTeaches(t *testing.T) {
	for _, system := range []string{"darwin", "linux", "windows"} {
		var said strings.Builder
		w := holding(newFake())
		w.keyboard = KeyboardOf(system)
		for _, f := range w.Helping() {
			said.WriteString(f.Key + " ")
		}
		for _, offer := range []string{"new", "search", "help", "archived", "archive", "take", "keep",
			"edit", "delete", "agree", "save", "done", "close"} {
			key := w.keyboard.Spelled(Pretty(Shortcut(offer)))
			if !strings.Contains(said.String(), key) {
				t.Errorf("on %s, %s is %q on the bar and the help never names it", system, offer, key)
			}
		}
	}
}
