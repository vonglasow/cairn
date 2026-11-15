package tui

import (
	"cairn/internal/board"
	"cairn/internal/theme"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"cairn/internal/engine"
	"cairn/internal/engine/act"
)

var update = os.Getenv("UPDATE_FRAMES") != ""

func fullBoard() *fake {
	f := newFake()
	f.cards = []board.Card{
		{Number: 1, Title: "Buy a lamp", State: "backlog", Priority: "normal"},
		{Number: 2, Title: "Read the ADR", State: "backlog", Priority: "normal"},
		{Number: 3, Title: "Ask legal about the retention clause", State: "backlog", Priority: "critical", Deadline: on("2026-08-29")},
		{Number: 4, Title: "Write the spec", State: "todo", Priority: "high", Description: "The client wants two columns, not three.", Deadline: on("2026-09-03"), SubDone: 1, SubTotal: 2, Labels: []string{"spec"},
			Steps: []engine.Part{{Number: 1, Title: "Draft the columns", Done: true}, {Number: 2, Title: "Get the client's nod"}}},
		{Number: 5, Title: "Fix the hook", State: "todo", Priority: "normal"},
		{Number: 6, Title: "Spec review", State: "current", Priority: "high", Description: "something written", InFocus: true, Deadline: on("2026-09-02"), SubDone: 2, SubTotal: 4, Labels: []string{"spec"}},
		{Number: 7, Title: "Migrate the store", State: "done", Priority: "normal"},
		{Number: 8, Title: "Client's answer", State: "blocked", Priority: "critical", Blockers: []int{4}, Objectives: []int{1}, Description: "the one that is in the way"},
		{Number: 9, Title: "Handed to Marie", State: "watching", Priority: "normal", Description: "something written", Because: "she confirms first"},
		{Number: 10, Title: "Refactor migrate", State: "paused", Priority: "normal", Because: "your call"},
		{Number: 13, Title: "Waiting on the lawyer", State: "blocked", Priority: "normal", Because: "the fourth, so the band wraps"},
	}
	f.goals = []board.Goal{
		{Number: 1, Title: "Ship the desktop board", State: "active", Priority: "high", Description: "something written", Done: 7, Total: 12, Labels: []string{"product"}, Deadline: on("2026-10-16")},
		{Number: 2, Title: "Cut the onboarding time in half", State: "on-hold", Priority: "normal", Done: 1, Total: 8},
	}
	f.focus = []int{6}
	for i := range 40 {
		var told act.AsTold
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

func init() { time.Local = time.FixedZone("UTC+2", 2*60*60) }

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
		{"holding", 100, 40, []string{"ctrl+l"}},
		{"search", 100, 40, []string{"/", "s", "p", "e", "c"}},
		{"command", 100, 40, []string{":", "b", "a"}},
		{"choosing", 100, 40, []string{"ctrl+l", "shift+j"}},
		{"help", 100, 40, []string{"?"}},
		{"detail", 100, 40, []string{"ctrl+l", "enter"}},
		{"form", 100, 40, []string{"ctrl+l", "e"}},
		{"objectives", 100, 40, []string{"tab"}},
		{"objectives-help", 100, 40, []string{"tab", "?"}},
		{"inside", 100, 40, []string{"tab", "enter"}},
		{"band", 100, 40, []string{"ctrl+l", "ctrl+j"}},
		{"form-on-a-choice", 100, 40, []string{"ctrl+l", "e", "tab", "tab"}},
		{"form-changed-a-choice", 100, 40, []string{"ctrl+l", "e", "tab", "tab", "right"}},
		{"objectives-form", 100, 40, []string{"tab", "e"}},
		{"detail-stalled", 100, 40, []string{"ctrl+l", "ctrl+j", "enter"}},
		{"help-narrow", 72, 30, []string{"?"}},
		{"detail-narrow", 72, 28, []string{"ctrl+l", "enter"}},
		{"detail-stalled-short", 100, 28, []string{"ctrl+l", "ctrl+j", "enter"}},
		{"form-narrow", 72, 28, []string{"ctrl+l", "e"}},
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

func pinned(t *testing.T, name, got string) {
	t.Helper()
	if found := outsideTheScheme(got, theme.Default()); len(found) > 0 {
		t.Errorf("the frame %s wears colours the scheme does not have: %q", name, found)
	}
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
