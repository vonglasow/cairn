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
// padding puts a click on the day beside the one drawn. The days share the
// width of the panel between them.
func TestTheCalendarIsLaidOutToThePixel(t *testing.T) {
	box := CalendarFor(100, 50, 400)
	if want := (Box{X: 100, Y: 50, W: 400, H: 322}); box != want {
		t.Errorf("the calendar is %+v, want %+v", box, want)
	}
	if got, want := DayAt(box, 0), (Box{X: 112, Y: 124, W: 49, H: 28}); got != want {
		t.Errorf("the first day is at %+v, want %+v", got, want)
	}
	if got, want := DayAt(box, 9), (Box{X: 218, Y: 156, W: 49, H: 28}); got != want {
		t.Errorf("the tenth day is at %+v, want %+v", got, want)
	}
	if got, want := WeekdayAt(box, 2), (Box{X: 218, Y: 100, W: 49, H: 18}); got != want {
		t.Errorf("Wednesday is written at %+v, want %+v", got, want)
	}
	if got, want := MonthAfter(box), (Box{X: 460, Y: 62, W: monthWide, H: titleHi}); got != want {
		t.Errorf("the next month is offered at %+v, want %+v", got, want)
	}
	if got, want := MonthBefore(box), (Box{X: 460 - monthWide, Y: 62, W: monthWide, H: titleHi}); got != want {
		t.Errorf("the month before is offered at %+v, want %+v", got, want)
	}
	if got := len(MonthGrid(time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC))); got != 42 {
		t.Errorf("a month is shown as %d days, want six weeks", got)
	}
}

// The calendar hangs under the field it answers, the way a list does, as wide
// as its heading with the arrows beside it, its offers, or a week of days.
func TestTheCalendarHangsUnderItsField(t *testing.T) {
	w := holding(newFake(), reading.Card{Number: 1, Title: "A", State: "todo"})
	w = w.clicked(middleOf(w, 1, 0)).Choose("Edit")
	for w.Form().Fields()[w.Form().On()].Arg != "deadline" {
		w = w.Key("tab")
	}
	w = w.Key("enter")
	box, frame := w.CalendarBox(), fieldFrame(w)
	if box.X != frame.X || box.Y != frame.Y+frame.H+between {
		t.Errorf("the calendar is at %+v, under a field at %+v", box, frame)
	}
	september := w
	september.day = time.Date(2026, time.September, 30, 0, 0, 0, 0, time.UTC)
	want := max(september.headWide()+2*monthWide, w.offersWide())
	if box.W != want || box.H != CalendarFor(0, 0, want).H {
		t.Errorf("the calendar is %dx%d, want %d wide", box.W, box.H, want)
	}
}

// A click on the deadline opens the calendar, as one on a field a list
// answers opens the list; on any other field it only puts the keyboard there.
func TestAClickOnTheDeadlineOpensTheCalendar(t *testing.T) {
	w := holding(newFake(), reading.Card{Number: 1, Title: "A", State: "todo"})
	w = w.clicked(middleOf(w, 1, 0)).Choose("Edit")
	for _, c := range []struct {
		arg    string
		screen Screen
	}{{"deadline", OnCalendar}, {"estimate", OnForm}} {
		for i, field := range w.Form().Fields() {
			if field.Arg != c.arg {
				continue
			}
			_, value, _ := FormPlace(w.formBox(), w.Form().Fields(), i)
			if got := w.clicked(Point{X: value.X + 4, Y: value.Y + 4}); got.Screen() != c.screen || got.Form().On() != i {
				t.Errorf("a click on the %s left the window on %v, on field %d", c.arg, got.Screen(), got.Form().On())
			}
		}
	}
}

// The arrows in the heading walk a month either way, and so does the wheel
// over the calendar, a notch a month; off the calendar it does nothing.
func TestThePointerWalksTheMonths(t *testing.T) {
	w := Window{screen: OnCalendar, width: 1200, height: 800, calendarState: calendarState{day: time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)}}
	w.form = editing(reading.Card{Number: 1, Title: "A"}, nil)
	box := w.CalendarBox()
	middle := func(b Box) Point { return Point{X: b.X + b.W/2, Y: b.Y + b.H/2} }
	if got := domain.FormatDay(w.Click(middle(MonthAfter(box))).day); got != "2026-02-28" {
		t.Errorf("the next month is %s", got)
	}
	if got := domain.FormatDay(w.Click(middle(MonthBefore(box))).day); got != "2025-12-31" {
		t.Errorf("the month before is %s", got)
	}
	if got := w.Click(middle(MonthAfter(box))).Screen(); got != OnCalendar {
		t.Errorf("walking a month left the window on %v", got)
	}
	in := middle(box)
	if got := domain.FormatDay(w.Scroll(in, lineHigh).day); got != "2026-02-28" {
		t.Errorf("a notch down is %s", got)
	}
	if got := domain.FormatDay(w.Scroll(in, -2*lineHigh).day); got != "2025-11-30" {
		t.Errorf("two notches up is %s", got)
	}
	if got := w.Scroll(in, lineHigh/2); !got.day.Equal(w.day) || got.spareChoosing != lineHigh/2 {
		t.Errorf("half a notch walked to %s, keeping %d", domain.FormatDay(got.day), got.spareChoosing)
	}
	if got := w.Scroll(Point{X: 1, Y: 1}, lineHigh); !got.day.Equal(w.day) {
		t.Errorf("the wheel off the calendar walked to %s", domain.FormatDay(got.day))
	}
}

// The arrows walk back as they walk on: a day, and a week.
func TestTheCalendarWalksBack(t *testing.T) {
	w := Window{screen: OnCalendar, calendarState: calendarState{day: time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC)}}
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
	w := Window{screen: OnCalendar, calendarState: calendarState{day: time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)}}
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
	w := Window{screen: OnCalendar, width: 1200, height: 800, calendarState: calendarState{day: time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC)}}
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
	w := Window{screen: OnCalendar, calendarState: calendarState{day: time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC)}}
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

// With no form to hang from, the calendar is across the middle of the window.
func TestACalendarWithNoFieldIsInTheMiddle(t *testing.T) {
	w := Window{screen: OnCalendar, width: 1200, height: 800, calendarState: calendarState{day: time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC)}}
	box := w.CalendarBox()
	if box.X != (1200-box.W)/2 || box.Y != (800-box.H)/2 {
		t.Errorf("the calendar is at %+v", box)
	}
}

// Each arrow of the heading sits on a chip two pixels inside its room, so the
// two read as two things to press and not one bar.
func TestTheMonthArrowsAreOnChips(t *testing.T) {
	w := Window{screen: OnCalendar, width: 1200, height: 800, calendarState: calendarState{day: time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC)}}
	w.form = editing(reading.Card{Number: 1, Title: "A"}, nil)
	shot, skin := drawn(t, w), Dark()
	for _, arrow := range []Box{MonthBefore(w.CalendarBox()), MonthAfter(w.CalendarBox())} {
		chip := Box{X: arrow.X + 2, Y: arrow.Y, W: arrow.W - 4, H: 1}
		if got := painted(shot, chip, skin.Back); got != chip.W {
			t.Errorf("the chip at %+v has %d of %d pixels", chip, got, chip.W)
		}
		for _, edge := range []Box{{X: arrow.X + 1, Y: arrow.Y, W: 1, H: 1}, {X: arrow.X + arrow.W - 2, Y: arrow.Y, W: 1, H: 1}} {
			if painted(shot, edge, skin.Back) != 0 {
				t.Errorf("the chip reaches %+v, past its room", edge)
			}
		}
	}
}

// A click off the calendar closes it: on the form, back to the form with the
// day as it was; off the form as well, the form goes too, what was typed in
// it saved, as a click off any panel puts it away.
func TestAClickOffTheCalendarClosesIt(t *testing.T) {
	f := newFake()
	w := holding(f, reading.Card{Number: 1, Title: "A", State: "todo"})
	w = w.clicked(middleOf(w, 1, 0)).Choose("Edit")
	for w.Form().Fields()[w.Form().On()].Arg != "deadline" {
		w = w.Key("tab")
	}
	w = w.Key("enter")
	if w.Screen() != OnCalendar {
		t.Fatalf("enter left the window on %v", w.Screen())
	}
	title := formValue(w.formBox(), w.Form().Fields(), 0)
	back := w.Click(Point{X: title.X + 4, Y: title.Y + 4})
	if back.Screen() != OnForm || back.Form().Fields()[back.Form().On()].Value != "" {
		t.Errorf("a click on the form left the window on %v", back.Screen())
	}
	if gone := w.Click(Point{X: 2, Y: 2}); gone.Screen() != OnTheBoard {
		t.Errorf("a click off the form left the window on %v", gone.Screen())
	}
	if inside := w.CalendarBox(); w.Click(Point{X: inside.X + inside.W - 2, Y: inside.Y + inside.H - 2}).Screen() != OnCalendar {
		t.Error("a click on the calendar's own ground closed it")
	}
	typed := w.Key("esc").On("title").Types("x")
	if left := typed.Click(Point{X: 2, Y: 2}); left.Screen() != OnTheBoard || f.what("task", "edit") != 1 {
		t.Errorf("a click off a form typed into left the window on %v, %d edits asked", left.Screen(), f.what("task", "edit"))
	}
}
