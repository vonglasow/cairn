package tui

import "testing"

func TestLayoutFor(t *testing.T) {
	for _, c := range []struct{ width, column, interior, band int }{
		{120, 29, 25, 89},
		{100, 24, 20, 74},
		{80, 19, 15, 59},
		{60, 14, 10, 44},
	} {
		l := LayoutFor(c.width, 40)
		if l.Column != c.column || l.Interior != c.interior || l.Band != c.band {
			t.Fatalf("LayoutFor(%d) = %+v, want column %d interior %d band %d",
				c.width, l, c.column, c.interior, c.band)
		}
	}
}

func TestLayoutBandSpansTheLastThreeColumns(t *testing.T) {
	l := LayoutFor(100, 40)
	// The band lies under three columns and the two gaps between them, and
	// nowhere else: the backlog keeps its own full height beside it.
	if want := l.Column*3 + 2; l.Band != want {
		t.Fatalf("the band is %d wide, want %d", l.Band, want)
	}
}

func TestTooSmall(t *testing.T) {
	if TooSmall(Narrowest, Shortest) {
		t.Fatal("the smallest window that fits was refused")
	}
	if !TooSmall(Narrowest-1, Shortest) {
		t.Fatal("a window one column too narrow was accepted")
	}
	if !TooSmall(Narrowest, Shortest-1) {
		t.Fatal("a window one row too short was accepted")
	}
}

func TestTheBandIsNeverGivenNothing(t *testing.T) {
	// Whatever the window, the band keeps a share of it: what is not moving is
	// the thing most worth seeing, and a band pushed off the bottom might as
	// well not be there.
	for height := Shortest; height <= 80; height++ {
		l := LayoutFor(100, height)
		if l.Rows < 1 {
			t.Fatalf("at %d rows the band was given %d", height, l.Rows)
		}
		if l.Top < 1 {
			t.Fatalf("at %d rows the columns were given %d", height, l.Top)
		}
		if l.Top+l.Rows+fixed != height {
			t.Fatalf("at %d rows the board takes %d", height, l.Top+l.Rows+fixed)
		}
	}
}

func TestTheBandTakesTwoRowsOfCardsWhenThereIsRoom(t *testing.T) {
	if got := LayoutFor(100, 60).Rows; got != bandCards*CardHeight {
		t.Fatalf("the band got %d rows in a tall window, want two rows of cards", got)
	}
	// And half of what is left when there is not.
	if got := LayoutFor(100, 26).Rows; got != 9 {
		t.Fatalf("the band got %d rows in a short window, want half of what was left", got)
	}
}

func TestShowing(t *testing.T) {
	for _, c := range []struct{ rows, want int }{{30, 5}, {6, 1}, {5, 0}, {0, 0}} {
		if got := Showing(c.rows); got != c.want {
			t.Fatalf("Showing(%d) = %d, want %d", c.rows, got, c.want)
		}
	}
}

func TestScrolled(t *testing.T) {
	// Thirty rows is five cards. Nothing moves while the one wanted is already
	// on screen; it moves by whole cards when it is not.
	for _, c := range []struct {
		name                   string
		offset, at, rows, want int
	}{
		{"already in view", 0, 2, 30, 0},
		{"just below the last one in view", 0, 5, 30, 6},
		{"far below", 0, 9, 30, 30},
		{"above what is shown", 30, 2, 30, 12},
		{"a window too short for a whole card", 0, 3, 4, 18},
	} {
		t.Run(c.name, func(t *testing.T) {
			if got := Scrolled(c.offset, c.at, c.rows); got != c.want {
				t.Fatalf("Scrolled(%d, %d, %d) = %d, want %d", c.offset, c.at, c.rows, got, c.want)
			}
		})
	}
}

func TestLayoutDividesWhatIsLeftAfterTheMarginAndTheGaps(t *testing.T) {
	// Ninety-nine, not a hundred: at a hundred the margin and the gaps happen
	// to divide out the same way whichever sign they carry, so a wrong one
	// would go unnoticed.
	if got := LayoutFor(99, 40).Column; got != 23 {
		t.Fatalf("a column at 99 columns is %d wide, want 23", got)
	}
	if got := LayoutFor(101, 40).Column; got != 24 {
		t.Fatalf("a column at 101 columns is %d wide, want 24", got)
	}
}
