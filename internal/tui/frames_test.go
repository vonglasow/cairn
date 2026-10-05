package tui

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"cairn/internal/reading"
	"cairn/internal/trace"
)

// The frames, pinned to the character. Nothing else can hold a drawing
// honestly: a width off by one, a gap that moved, a column that lost its
// header — none of it shows up in a test that only looks for a word, and all of
// it shows up here. Run with -update to write the files again after a change
// that was meant.
//
// What is stored is the screen with the styling taken off, because a colour is
// checked where it is decided and a frame is checked where it is drawn.

var update = os.Getenv("UPDATE_FRAMES") != ""

func fullBoard() *fake {
	f := newFake()
	f.cards = []reading.Card{
		{Number: 1, Title: "Buy a lamp", State: "backlog", Priority: "normal"},
		{Number: 2, Title: "Read the ADR", State: "backlog", Priority: "normal"},
		{Number: 3, Title: "Ask legal about the retention clause", State: "backlog", Priority: "critical", Deadline: on("2026-08-29")},
		{Number: 4, Title: "Write the spec", State: "todo", Priority: "high", Description: "The client wants two columns, not three.", Deadline: on("2026-09-03"), SubDone: 1, SubTotal: 2, Labels: []string{"spec"}},
		{Number: 11, Title: "Draft the columns", State: "done", Parent: 4},
		{Number: 12, Title: "Get the client's nod", State: "todo", Parent: 4},
		{Number: 5, Title: "Fix the hook", State: "todo", Priority: "normal"},
		{Number: 6, Title: "Spec review", State: "current", Priority: "high", Description: "something written", InFocus: true, Deadline: on("2026-09-02"), SubDone: 2, SubTotal: 4, Labels: []string{"spec"}},
		{Number: 7, Title: "Migrate the store", State: "done", Priority: "normal"},
		{Number: 8, Title: "Client's answer", State: "blocked", Priority: "critical", Blockers: []int{4}, Objectives: []int{1}, Description: "the one that is in the way"},
		{Number: 9, Title: "Handed to Marie", State: "watching", Priority: "normal", Description: "something written", Because: "she confirms first"},
		{Number: 10, Title: "Refactor migrate", State: "paused", Priority: "normal", Because: "your call"},
		{Number: 13, Title: "Waiting on the lawyer", State: "blocked", Priority: "normal", Because: "the fourth, so the band wraps"},
	}
	f.goals = []reading.Goal{
		{Number: 1, Title: "Ship the desktop board", State: "active", Priority: "high", Description: "something written", Done: 7, Total: 12, Labels: []string{"product"}, Deadline: on("2026-10-16")},
		{Number: 2, Title: "Cut the onboarding time in half", State: "on-hold", Priority: "normal", Done: 1, Total: 8},
	}
	f.focus = []int{6}
	// A life, so the history panel has something in it, and enough of one that
	// it has to scroll.
	for i := range 40 {
		var told trace.AsTold
		told.On(4)
		told.Did("changed")
		told.Changing("priority")
		told.Was("normal")
		told.Became("high")
		told.At(today.Add(time.Duration(i) * time.Hour))
		f.acts = append(f.acts, told.Act())
	}
	return f
}

// crowded is a board with more in it than the window can hold, so what scrolls
// has something to scroll.
func crowded() *fake {
	f := newFake()
	for i := 1; i <= 12; i++ {
		f.cards = append(f.cards, reading.Card{Number: i, Title: "Task " + strconv.Itoa(i), State: "todo", Priority: "normal"})
	}
	for i := 20; i < 28; i++ {
		f.cards = append(f.cards, reading.Card{Number: i, Title: "Stalled " + strconv.Itoa(i), State: "blocked", Because: "waiting"})
	}
	for i := 1; i <= 9; i++ {
		f.goals = append(f.goals, reading.Goal{Number: i, Title: "Objective " + strconv.Itoa(i), State: "active", Done: i, Total: 9})
	}
	// One with nothing under it yet, so a card with no count to show is drawn
	// as well as one with.
	f.goals = append(f.goals, reading.Goal{Number: 10, Title: "Nothing serves this yet", State: "active"})
	// And the tasks of the first objective, so its own board has something on
	// it and something in its band.
	for i := 1; i <= 12; i++ {
		f.cards[i-1].Objectives = []int{1}
	}
	return f
}

// The frames were pinned where they were written, at UTC+2 in September, and
// a moment is drawn in the zone of the machine drawing it: run anywhere else
// — a Linux container keeps UTC — every line of a history moved by two hours
// and every frame that shows one went red for nothing. The zone they were
// pinned in is the zone the whole package's tests draw in.
func init() { time.Local = time.FixedZone("UTC+2", 2*60*60) }

func TestFramesOfACrowdedBoard(t *testing.T) {
	f := crowded()
	for _, c := range []struct {
		name          string
		width, height int
		keys          []string
	}{
		{"crowded", 100, 40, nil},
		{"crowded-scrolled", 100, 40, []string{"l", "i", "j", "j", "j", "j", "j", "j"}},
		{"crowded-band", 100, 40, []string{"l", "i", "j", "j", "j", "j", "j", "j", "j", "j", "j", "j", "j", "j", "j"}},
		{"crowded-objectives", 100, 40, []string{"tab", "j", "j", "j", "j", "j", "j"}},
		{"crowded-inside", 100, 40, []string{"tab", "enter", "l", "i", "j", "j", "j", "j"}},
	} {
		t.Run(c.name, func(t *testing.T) {
			b := standing(f)
			b.width, b.height = c.width, c.height
			pinned(t, c.name, hit(b, c.keys...).Draw())
		})
	}
}

func TestFrames(t *testing.T) {
	f := fullBoard()
	for _, c := range []struct {
		name          string
		width, height int
		keys          []string
	}{
		{"board", 100, 40, nil},
		{"board-short", 100, 26, nil},
		{"board-narrow", 60, 40, nil},
		{"too-narrow", 47, 20, nil},
		{"insert", 100, 40, []string{"l", "i"}},
		{"search", 100, 40, []string{"/", "s", "p", "e", "c"}},
		{"command", 100, 40, []string{":", "b", "a"}},
		{"choosing", 100, 40, []string{"l", "i", "shift+j"}},
		{"help", 100, 40, []string{"?"}},
		{"detail", 100, 40, []string{"l", "i", "enter"}},
		{"form", 100, 40, []string{"l", "i", "e"}},
		{"objectives", 100, 40, []string{"tab"}},
		{"objectives-help", 100, 40, []string{"tab", "?"}},
		{"inside", 100, 40, []string{"tab", "enter"}},
		{"band", 100, 40, []string{"l", "j"}},
		{"band-insert", 100, 40, []string{"l", "j", "i"}},
		// By tab and not by down: inside the description the arrows walk its
		// lines, so down stays there, and tab is what leaves it.
		{"form-on-a-choice", 100, 40, []string{"l", "i", "e", "tab", "tab"}},
		{"form-changed-a-choice", 100, 40, []string{"l", "i", "e", "tab", "tab", "right"}},
		{"objectives-form", 100, 40, []string{"tab", "e"}},
		{"detail-stalled", 100, 40, []string{"l", "j", "i", "enter"}},
		{"help-narrow", 72, 30, []string{"?"}},
		{"detail-narrow", 72, 28, []string{"l", "i", "enter"}},
		{"detail-stalled-short", 100, 28, []string{"l", "j", "i", "enter"}},
		{"form-narrow", 72, 28, []string{"l", "i", "e"}},
		{"objectives-narrow", 72, 30, []string{"tab"}},
		{"board-tall", 100, 52, nil},
	} {
		t.Run(c.name, func(t *testing.T) {
			b := standing(f)
			b.width, b.height = c.width, c.height
			b = hit(b, c.keys...)
			pinned(t, c.name, b.Draw())
		})
	}
}

// pinned compares a drawing with what was pinned, or writes it when asked.
func pinned(t *testing.T, name, got string) {
	t.Helper()
	got = bare(got)
	path := filepath.Join("testdata", name+".txt")
	if update {
		if err := os.WriteFile(path, []byte(got), 0o600); err != nil {
			t.Fatalf("writing %s: %v", path, err)
		}
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %s: %v — run with UPDATE_FRAMES=1 to write it", path, err)
	}
	if got != string(want) {
		t.Fatalf("the frame changed.\n--- pinned ---\n%s\n--- drawn ---\n%s", want, got)
	}
}

// bare takes the styling out of a drawing, and the trailing blanks off every
// line: a run of spaces at the end of a row is invisible on a screen and
// impossible to read in a file.
func bare(s string) string {
	var out strings.Builder
	inEscape := false
	for _, r := range s {
		switch {
		case r == 0x1b:
			inEscape = true
		case inEscape && r == 'm':
			inEscape = false
		case !inEscape:
			out.WriteRune(r)
		}
	}
	lines := strings.Split(out.String(), "\n")
	for i := range lines {
		lines[i] = strings.TrimRight(lines[i], " ")
	}
	return strings.Join(lines, "\n") + "\n"
}
