package tui

import (
	"cairn/internal/engine"
	"strings"
	"testing"
)

func TestHeader(t *testing.T) {
	got := Board{skin: plainSkin}.header("TODO", 12, 20, false)
	if !strings.HasPrefix(got, "TODO") {
		t.Fatalf("the name is not first: %q", got)
	}
	if !strings.HasSuffix(got, "12") {
		t.Fatalf("the count is not last: %q", got)
	}
	if len([]rune(got)) != 20 {
		t.Fatalf("the header is %d wide, want 20", len([]rune(got)))
	}
}

func TestTheColumnUnderTheCursorSaysSo(t *testing.T) {
	b := Board{skin: plainSkin}
	plainHead := b.header("TODO", 3, 20, false)
	under := b.header("TODO", 3, 20, true)
	if plainHead == under {
		t.Fatal("the column under the cursor has the same header as the others")
	}
	if len([]rune(under)) != 20 {
		t.Fatalf("the marked header is %d wide, want 20", len([]rune(under)))
	}
	if !strings.Contains(under, "▌TODO") {
		t.Fatalf("the mark is not in front of the name: %q", under)
	}
}

func TestTheRuleUnderTheCursorIsHeavier(t *testing.T) {
	b := Board{skin: plainSkin}
	if got := b.rule(4, false); got != "────" {
		t.Fatalf("an ordinary rule is %q", got)
	}
	if got := b.rule(4, true); got != "━━━━" {
		t.Fatalf("the rule under the cursor is %q", got)
	}
}

func TestUnderTheBand(t *testing.T) {
	if underTheBand(0) {
		t.Fatal("the band was said to lie under the backlog, and it does not")
	}
	for i := 1; i < engine.Columns; i++ {
		if !underTheBand(i) {
			t.Fatalf("the band was said not to lie under column %d", i)
		}
	}
}

func TestColumnsAreJoinedByARuleOnEveryRow(t *testing.T) {
	b := Board{skin: plainSkin}
	got := b.apart([]string{"a\nb", "c\nd", "e\nf"})
	if got != "a│c│e\nb│d│f" {
		t.Fatalf("the columns are joined as %q", got)
	}
}

func TestTheRuleStopsWhereTheBandStarts(t *testing.T) {
	b := standing(newFake())
	b.width, b.height = 100, 40
	l := LayoutFor(b.width, b.height)
	ruled := 0
	for _, line := range strings.Split(bare(b.Draw()), "\n") {
		if strings.Contains(line, "│") {
			ruled++
		}
		if strings.Contains(line, engine.Band) && strings.Contains(line, "│") {
			t.Fatalf("the rule runs through the band: %q", line)
		}
	}
	if want := 2 + l.Top; ruled != want {
		t.Fatalf("the rule stands on %d rows, want the header, the rule and %d of cards", ruled, l.Top)
	}
}
