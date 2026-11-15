package tui

import (
	"cairn/internal/board"
	"testing"
)

func TestTheBlockersFieldSaysTheOtherTasksItReaches(t *testing.T) {
	b := Board{skin: plainSkin, snap: board.Snapshot{Cards: []board.Card{
		{Number: 1, Title: "Fix the login", Said: []string{"Fix the login"}},
		{Number: 2, Title: "Write the docs", Said: []string{"Write the docs", "login"}},
		{Number: 3, Title: "Plan it", Said: []string{"Plan it"}},
	}}}
	for _, c := range []struct{ typed, want string }{
		{"", ""},
		{"1", "[x] 1 Fix the login"},
		{"1 login", "[ ] 2 Write the docs\n[x] 1 Fix the login"},
		{"nothing", "no other task"},
	} {
		typed := c.typed
		if got := b.about(false, 3)(board.Field{Arg: "blocker"}, &typed)(); got != c.want {
			t.Errorf("typing %q in blockers says %q, want %q", c.typed, got, c.want)
		}
	}
}
