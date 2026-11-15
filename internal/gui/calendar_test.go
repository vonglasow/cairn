package gui

import (
	"cairn/internal/board"
	"cairn/internal/engine/calendar"
	"cairn/internal/gui/geom"
	"testing"
	"time"
)

func TestThePointerWalksTheMonths(t *testing.T) {
	w := Window{Board: board.Board{Screen: board.OnCalendar, CalendarState: board.CalendarState{Day: time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)}}, width: 1200, height: 800}
	w.Form = Window{}.editing(board.Card{Number: 1, Title: "A"})
	box := w.CalendarBox()
	middle := func(b geom.Box) geom.Point { return geom.Point{X: b.X + b.W/2, Y: b.Y + b.H/2} }
	if got := calendar.FormatDay(w.Click(middle(MonthAfter(box))).Day); got != "2026-02-28" {
		t.Errorf("the next month is %s", got)
	}
	if got := calendar.FormatDay(w.Click(middle(MonthBefore(box))).Day); got != "2025-12-31" {
		t.Errorf("the month before is %s", got)
	}
	if got := w.Click(middle(MonthAfter(box))).Screen; got != board.OnCalendar {
		t.Errorf("walking a month left the window on %v", got)
	}
	in := middle(box)
	if got := calendar.FormatDay(w.Scroll(in, lineHigh).Day); got != "2026-02-28" {
		t.Errorf("a notch down is %s", got)
	}
	if got := calendar.FormatDay(w.Scroll(in, -2*lineHigh).Day); got != "2025-11-30" {
		t.Errorf("two notches up is %s", got)
	}
	if got := w.Scroll(in, lineHigh/2); !got.Day.Equal(w.Day) || got.spareChoosing != lineHigh/2 {
		t.Errorf("half a notch walked to %s, keeping %d", calendar.FormatDay(got.Day), got.spareChoosing)
	}
	if got := w.Scroll(geom.Point{X: 1, Y: 1}, lineHigh); !got.Day.Equal(w.Day) {
		t.Errorf("the wheel off the calendar walked to %s", calendar.FormatDay(got.Day))
	}
}

func TestTheCalendarWalksBack(t *testing.T) {
	w := Window{Board: board.Board{Screen: board.OnCalendar, CalendarState: board.CalendarState{Day: time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC)}}}
	if got := calendar.FormatDay(w.onCalendar("left").Day); got != "2026-10-11" {
		t.Errorf("a day back is %s", got)
	}
	if got := calendar.FormatDay(w.onCalendar("up").Day); got != "2026-10-05" {
		t.Errorf("a week back is %s", got)
	}
	if !w.Faded(time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC)) || w.Faded(time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)) {
		t.Error("only the days of the months either side are faint")
	}
}

func TestTheCalendarWalksMonthsWithoutSpillingOver(t *testing.T) {
	w := Window{Board: board.Board{Screen: board.OnCalendar, CalendarState: board.CalendarState{Day: time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)}}}
	if got := calendar.FormatDay(w.onCalendar("shift+right").Day); got != "2026-02-28" {
		t.Errorf("a month on from 2026-01-31 is %s, want 2026-02-28", got)
	}
	w.Day = time.Date(2026, 3, 31, 0, 0, 0, 0, time.UTC)
	if got := calendar.FormatDay(w.onCalendar("shift+left").Day); got != "2026-02-28" {
		t.Errorf("a month back from 2026-03-31 is %s, want 2026-02-28", got)
	}
}

func TestAPanelsOffersAreAsWideAsWhatTheySay(t *testing.T) {
	w := holding(newFake(), board.Card{Number: 1, Title: "A", State: "todo"})
	w = w.clicked(middleOf(w, 1, 0)).Choose("Edit")
	for w.Form.Fields[w.Form.On()].Arg != "deadline" {
		w = w.Key("tab")
	}
	w = w.Key("enter")
	if w.Screen != board.OnCalendar {
		t.Fatalf("enter on the deadline left the window on %v", w.Screen)
	}
	boxes, offers := w.PanelOffers(), w.Offers()
	if len(boxes) != len(offers) {
		t.Fatalf("%d boxes for %d offers", len(boxes), len(offers))
	}
	if right := w.CalendarBox().X + w.CalendarBox().W - padding; boxes[0].X+boxes[0].W != right {
		t.Errorf("the first offer ends at %d, not at the panel's %d", boxes[0].X+boxes[0].W, right)
	}
	for i, offer := range offers {
		said := Reckoned(w.Said(offer), reads)
		if wide := said[len(said)-1]; boxes[i].W < wide {
			t.Errorf("%q needs %d and has %d", w.Said(offer), wide, boxes[i].W)
		}
		if boxes[i].W < offerWide {
			t.Errorf("%q is narrower than an offer is ever drawn: %d", offer, boxes[i].W)
		}
		if i > 0 && boxes[i].X+boxes[i].W > boxes[i-1].X {
			t.Errorf("%q ends at %d, inside the one before it", offer, boxes[i].X+boxes[i].W)
		}
		if boxes[i] != (geom.Box{X: boxes[i].X, Y: PanelOfferAt(w.CalendarBox(), i).Y, W: boxes[i].W, H: geom.BarHigh}) {
			t.Errorf("%q is not on the panel's row of offers: %+v", offer, boxes[i])
		}
	}
	no := boxes[1]
	if got := w.Click(geom.Point{X: no.X + no.W - 2, Y: no.Y + no.H/2}); got.Screen != board.OnForm || got.Form.Fields[got.Form.On()].Value != "" {
		t.Errorf("a click at the end of %q left the window on %v", offers[1], got.Screen)
	}
}

func TestAClickOffTheCalendarClosesIt(t *testing.T) {
	f := newFake()
	w := holding(f, board.Card{Number: 1, Title: "A", State: "todo"})
	w = w.clicked(middleOf(w, 1, 0)).Choose("Edit")
	for w.Form.Fields[w.Form.On()].Arg != "deadline" {
		w = w.Key("tab")
	}
	w = w.Key("enter")
	if w.Screen != board.OnCalendar {
		t.Fatalf("enter left the window on %v", w.Screen)
	}
	title := formValue(w.formBox(), w.Form.Fields, 0)
	back := w.Click(geom.Point{X: title.X + 4, Y: title.Y + 4})
	if back.Screen != board.OnForm || back.Form.Fields[back.Form.On()].Value != "" {
		t.Errorf("a click on the form left the window on %v", back.Screen)
	}
	if gone := w.Click(geom.Point{X: 2, Y: 2}); gone.Screen != board.OnTheBoard {
		t.Errorf("a click off the form left the window on %v", gone.Screen)
	}
	if inside := w.CalendarBox(); w.Click(geom.Point{X: inside.X + inside.W - 2, Y: inside.Y + inside.H - 2}).Screen != board.OnCalendar {
		t.Error("a click on the calendar's own ground closed it")
	}
	typed := w.Key("esc").On("title").Types("x")
	if left := typed.Click(geom.Point{X: 2, Y: 2}); left.Screen != board.OnAsking || f.what("task", "edit") != 0 {
		t.Errorf("a click off a form typed into left the window on %v, %d edits asked", left.Screen, f.what("task", "edit"))
	}
}

func TestTheCalendarsGridsAreWalkedAndLeftByTheKeyboard(t *testing.T) {
	w := Window{Board: board.Board{Screen: board.OnCalendar, CalendarState: board.CalendarState{Day: time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC)}}, width: 1200, height: 800}
	w.Form = Window{}.editing(board.Card{Number: 1, Title: "A"})
	middle := func(b geom.Box) geom.Point { return geom.Point{X: b.X + b.W/2, Y: b.Y + b.H/2} }
	months := w.Key("m")
	if got := months.CalendarCells(); len(got) != 12 || got[0] != "Jan" {
		t.Fatalf("m offers %v, want the twelve months", got)
	}
	if got := calendar.FormatDay(months.Key("left").Key("down").Day); got != "2026-12-12" {
		t.Errorf("left then down in the months reach %s, want 2026-12-12", got)
	}
	if got := calendar.FormatDay(months.Key("up").Day); got != "2026-07-12" {
		t.Errorf("up in the months reaches %s, want 2026-07-12", got)
	}
	if got := calendar.FormatDay(months.clicked(middle(MonthAfter(months.CalendarBox()))).Day); got != "2027-10-12" {
		t.Errorf("the arrow in the months reaches %s, want a year on", got)
	}
	if got := months.Key("enter"); got.CalendarView != board.ViewDays || got.Screen != board.OnCalendar {
		t.Errorf("enter on a month leaves the calendar on view %d, screen %v", got.CalendarView, got.Screen)
	}
	if got := months.Key("esc"); got.CalendarView != board.ViewDays || got.Screen != board.OnCalendar || calendar.FormatDay(got.Day) != "2026-10-12" {
		t.Errorf("esc in the months leaves view %d, screen %v", got.CalendarView, got.Screen)
	}
	years := w.Key("y")
	if got := calendar.FormatDay(years.Key("right").Key("down").Day); got != "2030-10-12" {
		t.Errorf("right then down in the years reach %s, want 2030-10-12", got)
	}
	if got := calendar.FormatDay(years.clicked(middle(MonthBefore(years.CalendarBox()))).Day); got != "2014-10-12" {
		t.Errorf("the arrow in the years reaches %s, want twelve years back", got)
	}
	if got := years.Key("enter"); got.CalendarView != board.ViewMonths {
		t.Errorf("enter on a year leaves view %d, want the months", got.CalendarView)
	}
	if got := years.Key("esc"); got.CalendarView != board.ViewDays || got.Screen != board.OnCalendar {
		t.Errorf("esc in the years leaves view %d, screen %v", got.CalendarView, got.Screen)
	}
	if got := w.Key("esc"); got.Screen != board.OnForm {
		t.Errorf("esc on the days leaves screen %v, want the form", got.Screen)
	}
}
