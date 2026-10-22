package tui

import (
	"strings"
	"testing"

	"cairn/internal/engine"
)

func TestTheWholeTitleIsGivenOnlyWhenCut(t *testing.T) {
	long := "A title far longer than any column could ever hope to show in full, even over two lines of a card"
	f := newFake()
	f.cards = []engine.Card{
		{Number: 1, Title: long, State: "todo"},
		{Number: 2, Title: "Short", State: "todo"},
	}
	b := hit(standing(f), "l", "i")
	if got := b.wholeTitle(); got != long {
		t.Fatalf("the cut title was given as %q, want %q", got, long)
	}
	if got := hit(b, "j").wholeTitle(); got != "" {
		t.Fatalf("a title that fits was given as %q, want nothing", got)
	}
	if got := hit(b, "esc").wholeTitle(); got != "" {
		t.Fatalf("outside insert the title was given as %q, want nothing", got)
	}
}

func TestTheWholeTitleIsDrawnAboveTheBarWithinTheWindow(t *testing.T) {
	long := "A title far longer than any column could ever hope to show in full, even over two lines of a card"
	f := newFake()
	f.cards = []engine.Card{{Number: 1, Title: long, State: "todo"}}
	b := standing(f)
	b.width, b.height = 100, 40
	plain := hit(b, "l", "i")
	lines := strings.Split(bare(plain.Draw()), "\n")
	bar := len(lines) - 2
	if got := strings.TrimSpace(lines[bar-1]); got != long {
		t.Fatalf("the line above the bar is %q, want the whole title", got)
	}
	if want := strings.Count(bare(hit(b, "l").Draw()), "\n"); strings.Count(bare(plain.Draw()), "\n") != want {
		t.Fatalf("the board grew when the title was written")
	}
	narrow := plain
	narrow.width = 60
	rows := strings.Split(bare(narrow.Draw()), "\n")
	if got := rows[len(rows)-3]; got != " "+Clip(long, 58) || !strings.HasSuffix(got, ellipsis) {
		t.Fatalf("the title in a narrow window is %q, want it cut at the width with an ellipsis", got)
	}
	pinned(t, "insert-long-title", plain.Draw())
}
