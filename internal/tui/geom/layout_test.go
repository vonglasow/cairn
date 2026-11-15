package geom

import (
	"testing"
)

func TestLayoutFor(t *testing.T) {
	for _, c := range []struct{ width, column, interior int }{
		{120, 29, 25},
		{100, 24, 20},
		{80, 19, 15},
		{60, 14, 10},
	} {
		l := LayoutFor(c.width, 40)
		if l.Column != c.column || l.Interior != c.interior {
			t.Fatalf("LayoutFor(%d) = %+v, want column %d interior %d",
				c.width, l, c.column, c.interior)
		}
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
	for height := Shortest; height <= 80; height++ {
		l := LayoutFor(100, height)
		if l.Rows < 1 {
			t.Fatalf("at %d rows the band was given %d", height, l.Rows)
		}
		if l.Top < 1 {
			t.Fatalf("at %d rows the columns were given %d", height, l.Top)
		}
		if l.Top+l.Rows+Fixed+1 != height {
			t.Fatalf("at %d rows the board takes %d", height, l.Top+l.Rows+Fixed+1)
		}
	}
}

func TestTheBandTakesTwoRowsOfCardsWhenThereIsRoom(t *testing.T) {
	if got := LayoutFor(100, 60).Rows; got != BandCards*CardHeight {
		t.Fatalf("the band got %d rows in a tall window, want two rows of cards", got)
	}
	if got := LayoutFor(100, 26).Rows; got != 8 {
		t.Fatalf("the band got %d rows in a short window, want half of what was left, less its lanes' heads", got)
	}
}

func TestShowing(t *testing.T) {
	for _, c := range []struct{ rows, want int }{{30, 5}, {6, 1}, {5, 0}, {0, 0}} {
		if got := c.rows / CardHeight; got != c.want {
			t.Fatalf("%d rows show %d cards, want %d", c.rows, got, c.want)
		}
	}
}

func TestScrolled(t *testing.T) {
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
	if got := LayoutFor(99, 40).Column; got != 23 {
		t.Fatalf("a column at 99 columns is %d wide, want 23", got)
	}
	if got := LayoutFor(101, 40).Column; got != 24 {
		t.Fatalf("a column at 101 columns is %d wide, want 24", got)
	}
}
