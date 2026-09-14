package gui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Every size the window draws from is worked out here and nowhere else, so a
// test can read the whole shape without a screen. What is pinned is numbers:
// pixels would depend on a font, a driver and a version of Gio, and would go
// red on another machine for something that is not a defect.

func TestTooSmallIsTheTwoSidesTogether(t *testing.T) {
	for _, c := range []struct {
		what          string
		width, height int
		small         bool
	}{
		{"room enough", Narrowest, Shortest, false},
		{"one short of the width", Narrowest - 1, Shortest, true},
		{"one short of the height", Narrowest, Shortest - 1, true},
		{"short of both", 320, 240, true},
		{"far more than enough", 1600, 1000, false},
	} {
		if got := TooSmall(c.width, c.height); got != c.small {
			t.Errorf("%s: TooSmall(%d, %d) = %v", c.what, c.width, c.height, got)
		}
	}
}

func TestTheFourColumnsShareTheWidthEvenly(t *testing.T) {
	l := LayoutFor(1200, 800)
	first := l.Column[0]
	for i, column := range l.Column {
		if column.W != first.W {
			t.Errorf("column %d is %d wide, the first is %d", i, column.W, first.W)
		}
		if i > 0 && column.X <= l.Column[i-1].X {
			t.Errorf("column %d starts at %d, not after column %d at %d", i, column.X, i-1, l.Column[i-1].X)
		}
	}
	if got := l.Column[Columns-1].X + first.W; got > 1200 {
		t.Errorf("the last column ends at %d, past a window of 1200", got)
	}
}

// The backlog is the heap everything lands in, and it holds more than the
// other three together — so it runs the whole height, past where the band
// stops the others.
func TestTheBacklogRunsPastWhereTheBandStopsTheOthers(t *testing.T) {
	l := LayoutFor(1200, 800)
	if l.Column[0].H <= l.Column[1].H {
		t.Fatalf("the backlog is %d tall and todo is %d", l.Column[0].H, l.Column[1].H)
	}
	if got, want := l.Column[0].Y+l.Column[0].H, l.Band.Y+l.Band.H; got != want {
		t.Errorf("the backlog ends at %d, the band at %d: they should end together", got, want)
	}
}

// The band lies under the last three columns and not under the backlog.
func TestTheBandLiesUnderTheLastThreeColumns(t *testing.T) {
	l := LayoutFor(1200, 800)
	if l.Band.X != l.Column[1].X {
		t.Errorf("the band starts at %d, todo at %d", l.Band.X, l.Column[1].X)
	}
	if got, want := l.Band.X+l.Band.W, l.Column[Columns-1].X+l.Column[Columns-1].W; got != want {
		t.Errorf("the band ends at %d, the last column at %d", got, want)
	}
	if l.Band.Y < l.Column[1].Y+l.Column[1].H {
		t.Error("the band overlaps the columns above it")
	}
}

// Three lanes, named, sharing the band's width: which lane a card sits in is
// what says which kind of stopped it is.
func TestTheBandIsThreeLanesSharingItsWidth(t *testing.T) {
	l := LayoutFor(1200, 800)
	if len(Lanes) != 3 {
		t.Fatalf("there are %d lanes, want 3", len(Lanes))
	}
	for i, lane := range l.Lane {
		if lane.W != l.Lane[0].W {
			t.Errorf("lane %d is %d wide, the first is %d", i, lane.W, l.Lane[0].W)
		}
		if lane.X < l.Band.X || lane.X+lane.W > l.Band.X+l.Band.W {
			t.Errorf("lane %d runs from %d to %d, outside the band", i, lane.X, lane.X+lane.W)
		}
	}
}

// A card is always the same height, which is what lets a column say which one
// is under the pointer by arithmetic rather than by measurement.
func TestWhichCardIsUnderAPoint(t *testing.T) {
	l := LayoutFor(1200, 800)
	body := l.Column[1]
	for _, at := range []int{0, 1, 2} {
		card := l.CardAt(1, at, 0)
		if card.H != CardHeight {
			t.Errorf("card %d is %d tall, want %d", at, card.H, CardHeight)
		}
		middle := Point{X: card.X + card.W/2, Y: card.Y + card.H/2}
		column, index, found := l.Under(middle, 0)
		if !found || column != 1 || index != at {
			t.Errorf("the middle of card %d answered (%d, %d, %v)", at, column, index, found)
		}
	}
	above := Point{X: body.X + 1, Y: body.Y - 1}
	if _, _, found := l.Under(above, 0); found {
		t.Error("a point above the column answered a card")
	}
}

// Scrolling moves what is drawn, so the same point answers a different card.
func TestScrollingChangesWhatIsUnderAPoint(t *testing.T) {
	l := LayoutFor(1200, 800)
	card := l.CardAt(1, 0, 0)
	middle := Point{X: card.X + card.W/2, Y: card.Y + card.H/2}
	_, first, _ := l.Under(middle, 0)
	_, scrolled, _ := l.Under(middle, CardHeight)
	if scrolled != first+1 {
		t.Errorf("scrolled by one card, the same point answered %d after %d", scrolled, first)
	}
}

func TestHowManyCardsAColumnShows(t *testing.T) {
	l := LayoutFor(1200, 800)
	if got := l.Showing(1); got != l.Column[1].H/CardHeight {
		t.Errorf("todo shows %d cards in %d of height", got, l.Column[1].H)
	}
	if l.Showing(0) <= l.Showing(1) {
		t.Error("the backlog is taller and should show more")
	}
}

// The whole shape at two sizes, written down. A sum changed by accident shows
// up here as a number that moved, which is the only way to notice one.
func TestTheShapeAtTwoSizes(t *testing.T) {
	for _, size := range []struct{ w, h int }{{1200, 800}, {Narrowest, Shortest}} {
		name := fmt.Sprintf("%dx%d", size.w, size.h)
		t.Run(name, func(t *testing.T) {
			got := LayoutFor(size.w, size.h).Written()
			path := filepath.Join("testdata", "layout-"+name+".txt")
			if os.Getenv("UPDATE_FRAMES") != "" {
				if err := os.MkdirAll("testdata", 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
					t.Fatal(err)
				}
				return
			}
			want, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("%v — run the suite with UPDATE_FRAMES=1 to write it", err)
			}
			if strings.TrimSpace(string(want)) != strings.TrimSpace(got) {
				t.Errorf("the shape moved:\n%s\nwant:\n%s", got, want)
			}
		})
	}
}
