package gui

import (
	"testing"
	"time"

	"cairn/internal/domain"
	"cairn/internal/reading"
)

// A month is shown as six weeks from the Monday on or before its first day:
// every month fits, and the rows do not move from one month to the next.
func TestMonthGrid_startsOnTheMondayBeforeTheFirst(t *testing.T) {
	grid := MonthGrid(time.Date(2026, 10, 12, 15, 0, 0, 0, time.UTC))
	if got := domain.FormatDay(grid[0]); got != "2026-09-28" {
		t.Errorf("October 2026 starts its grid on %s, want the Monday 2026-09-28", got)
	}
	if got := domain.FormatDay(grid[len(grid)-1]); got != "2026-11-08" {
		t.Errorf("it ends on %s, want 2026-11-08", got)
	}
	if got := domain.FormatDay(MonthGrid(time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC))[0]); got != "2026-06-01" {
		t.Errorf("a month starting on a Monday starts its grid on %s", got)
	}
	if got := domain.FormatDay(MonthGrid(time.Date(2026, 11, 3, 0, 0, 0, 0, time.UTC))[0]); got != "2026-10-26" {
		t.Errorf("a month starting on a Sunday starts its grid on %s, want the Monday before", got)
	}
}

// The calendar's panel and its days, to the pixel: a sum gone wrong by a
// padding puts a click on the day beside the one drawn.
func TestTheCalendarIsLaidOutToThePixel(t *testing.T) {
	box := CalendarOver(1200, 800)
	if want := (Box{X: 378, Y: 191, W: 444, H: 418}); box != want {
		t.Errorf("the calendar is %+v, want %+v", box, want)
	}
	if got, want := DayAt(box, 0), (Box{X: 390, Y: 265, W: 56, H: 44}); got != want {
		t.Errorf("the first day is at %+v, want %+v", got, want)
	}
	if got, want := DayAt(box, 9), (Box{X: 510, Y: 313, W: 56, H: 44}); got != want {
		t.Errorf("the tenth day is at %+v, want %+v", got, want)
	}
	if got, want := WeekdayAt(box, 2), (Box{X: 510, Y: 241, W: 56, H: 18}); got != want {
		t.Errorf("Wednesday is written at %+v, want %+v", got, want)
	}
	if got := len(MonthGrid(time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC))); got != 42 {
		t.Errorf("a month is shown as %d days, want six weeks", got)
	}
}

// The arrows walk back as they walk on: a day, and a week.
func TestTheCalendarWalksBack(t *testing.T) {
	w := Window{screen: OnCalendar, day: time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC)}
	if got := domain.FormatDay(w.onCalendar("left").day); got != "2026-10-11" {
		t.Errorf("a day back is %s", got)
	}
	if got := domain.FormatDay(w.onCalendar("up").day); got != "2026-10-05" {
		t.Errorf("a week back is %s", got)
	}
	if !w.Faded(time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC)) || w.Faded(time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)) {
		t.Error("only the days of the months either side are faint")
	}
}

// A month on from the last day of a longer month is the last day of the
// shorter one, never a day into the month after it.
func TestTheCalendarWalksMonthsWithoutSpillingOver(t *testing.T) {
	w := Window{screen: OnCalendar, day: time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)}
	if got := domain.FormatDay(w.onCalendar("shift+right").day); got != "2026-02-28" {
		t.Errorf("a month on from 2026-01-31 is %s, want 2026-02-28", got)
	}
	w.day = time.Date(2026, 3, 31, 0, 0, 0, 0, time.UTC)
	if got := domain.FormatDay(w.onCalendar("shift+left").day); got != "2026-02-28" {
		t.Errorf("a month back from 2026-03-31 is %s, want 2026-02-28", got)
	}
}

// The calendar is drawn over the form: the day the keyboard is on carries the
// outline the chosen card does.
func TestTheCalendarIsDrawn(t *testing.T) {
	w := Window{screen: OnCalendar, width: 1200, height: 800, day: time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC)}
	w.form = editing(reading.Card{Number: 1, Title: "Write the spec"}, nil)
	shot := drawn(t, w)
	for i, day := range MonthGrid(w.day) {
		edge := Edges(DayAt(w.CalendarBox(), i))[0]
		outlined := ink(shot, edge, Dark().Panel) > 0
		if day.Equal(w.day) && !outlined {
			t.Error("the day the keyboard is on is not outlined")
		}
		if !day.Equal(w.day) && outlined {
			t.Errorf("%s is outlined, and the keyboard is not on it", domain.FormatDay(day))
		}
	}
}

// A form adding names what it adds as new, never as number nought: the list
// and the calendar it opens say so in their heading.
func TestAFormAddingIsNamedNew(t *testing.T) {
	w := Window{screen: OnCalendar, day: time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC)}
	w.form = adding(Tasks, "backlog")
	if got := Heads(w); got != "a deadline for the new task — October 2026" {
		t.Errorf("the calendar of a task being added is headed %q", got)
	}
	w.form = editing(reading.Card{Number: 4, Title: "Ship"}, nil)
	if got := Heads(w); got != "a deadline for task 4 — October 2026" {
		t.Errorf("the calendar of task 4 is headed %q", got)
	}
}

// Each offer of a panel is as wide as what it says, and the next one starts
// after it: "backspace  No day" ran into "esc  Cancel" when every offer had the
// same width whatever it said.
func TestAPanelsOffersAreAsWideAsWhatTheySay(t *testing.T) {
	w := holding(newFake(), reading.Card{Number: 1, Title: "A", State: "todo"})
	w = w.clicked(middleOf(w, 1, 0)).Choose("Edit")
	for w.Form().Fields()[w.Form().On()].Arg != "deadline" {
		w = w.Key("tab")
	}
	w = w.Key("enter")
	if w.Screen() != OnCalendar {
		t.Fatalf("enter on the deadline left the window on %v", w.Screen())
	}
	boxes, offers := w.PanelOffers(), w.Offers()
	if len(boxes) != len(offers) {
		t.Fatalf("%d boxes for %d offers", len(boxes), len(offers))
	}
	if first := PanelOfferAt(w.CalendarBox(), 0); boxes[0].X != first.X {
		t.Errorf("the first offer starts at %d, not at the panel's %d", boxes[0].X, first.X)
	}
	for i, offer := range offers {
		said := Reckoned(w.Said(offer), reads)
		if wide := said[len(said)-1]; boxes[i].W < wide {
			t.Errorf("%q needs %d and has %d", w.Said(offer), wide, boxes[i].W)
		}
		if boxes[i].W < offerWide {
			t.Errorf("%q is narrower than an offer is ever drawn: %d", offer, boxes[i].W)
		}
		if i > 0 && boxes[i].X < boxes[i-1].X+boxes[i-1].W {
			t.Errorf("%q starts at %d, inside the one before it", offer, boxes[i].X)
		}
		if boxes[i] != (Box{X: boxes[i].X, Y: PanelOfferAt(w.CalendarBox(), i).Y, W: boxes[i].W, H: barHigh}) {
			t.Errorf("%q is not on the panel's row of offers: %+v", offer, boxes[i])
		}
	}
	// And a click lands on the offer drawn where it is.
	no := boxes[1]
	if got := w.Click(Point{X: no.X + no.W - 2, Y: no.Y + no.H/2}); got.Screen() != OnForm || got.Form().Fields()[got.Form().On()].Value != "" {
		t.Errorf("a click at the end of %q left the window on %v", offers[1], got.Screen())
	}
}

// Under the calendar the form draws none of its offers: they would be the
// calendar's, drawn a second time at the form's foot.
func TestTheFormUnderAPanelOffersNothing(t *testing.T) {
	w := holding(newFake(), reading.Card{Number: 1, Title: "A", State: "todo"})
	w = w.clicked(middleOf(w, 1, 0)).Choose("Edit")
	foot := PanelOfferAt(w.formPanel(), 0)
	skin := Dark()
	if painted(drawn(t, w), foot, skin.Chosen) == 0 {
		t.Fatal("the open form draws no offer at its foot")
	}
	for w.Form().Fields()[w.Form().On()].Arg != "deadline" {
		w = w.Key("tab")
	}
	if got := w.Key("enter"); painted(drawn(t, got), foot, skin.Chosen) != 0 {
		t.Error("the form under the calendar still draws offers at its foot")
	}
}

// The line saying where a task's context lives starts after the offers as
// they are drawn: spelled for Windows, "Ctrl+Enter  Save" is wider than an
// offer's least width, and the line used to start under it.
func TestTheContextLineStartsAfterTheOffersDrawn(t *testing.T) {
	w := holding(newFake(), reading.Card{Number: 1, Title: "A", State: "todo", Context: "/notes/1"})
	w.keyboard = KeyboardOf("windows")
	w = w.clicked(middleOf(w, 1, 0)).Choose("Edit")
	offers := w.PanelOffers()
	last := offers[len(offers)-1]
	said := w.ContextSays()
	if len(said) != 2 {
		t.Fatalf("the context says %d things", len(said))
	}
	name, path := said[0].inSight(), said[1].inSight()
	if name.X < last.X+last.W+padding {
		t.Errorf("the context's name starts at %d, the last offer ends at %d", name.X, last.X+last.W)
	}
	if path.X != name.X+name.W || path.X+path.W != w.formPanel().X+w.formPanel().W-padding {
		t.Errorf("the path runs %+v after a name %+v", path, name)
	}
}
