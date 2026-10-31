package gui

import (
	"testing"
	"time"

	"cairn/internal/domain"
	"cairn/internal/engine"
)

func TestTheCalendarIsLaidOutToThePixel(t *testing.T) {
	box := CalendarFor(100, 50, 400)
	if want := (Box{X: 100, Y: 50, W: 400, H: 358}); box != want {
		t.Errorf("the calendar is %+v, want %+v", box, want)
	}
	if got, want := DayAt(box, 0), (Box{X: 112, Y: 124, W: 47, H: 32}); got != want {
		t.Errorf("the first day is at %+v, want %+v", got, want)
	}
	if got, want := DayAt(box, 9), (Box{X: 218, Y: 162, W: 47, H: 32}); got != want {
		t.Errorf("the tenth day is at %+v, want %+v", got, want)
	}
	if got, want := WeekdayAt(box, 2), (Box{X: 218, Y: 100, W: 47, H: 18}); got != want {
		t.Errorf("Wednesday is written at %+v, want %+v", got, want)
	}
	if got, want := MonthAfter(box), (Box{X: 456, Y: 62, W: monthWide, H: calendarHead}); got != want {
		t.Errorf("the next month is offered at %+v, want %+v", got, want)
	}
	if got, want := CalendarMonth(box), (Box{X: 112, Y: 62, W: nameWide, H: calendarHead}); got != want {
		t.Errorf("the month is offered at %+v, want %+v", got, want)
	}
	if got, want := MonthBefore(box), (Box{X: 456 - monthWide - inset, Y: 62, W: monthWide, H: calendarHead}); got != want {
		t.Errorf("the month before is offered at %+v, want %+v", got, want)
	}
	if got := len(engine.MonthGrid(time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC))); got != 42 {
		t.Errorf("a month is shown as %d days, want six weeks", got)
	}
}

func TestTheCalendarHangsUnderItsField(t *testing.T) {
	w := holding(newFake(), engine.Card{Number: 1, Title: "A", State: "todo"})
	w = w.clicked(middleOf(w, 1, 0)).Choose("Edit")
	for w.Form().Fields()[w.Form().On()].Arg != "deadline" {
		w = w.Key("tab")
	}
	w = w.Key("enter")
	box, frame := w.CalendarBox(), fieldFrame(w)
	if box.X != frame.X || box.Y != frame.Y+frame.H+between {
		t.Errorf("the calendar is at %+v, under a field at %+v", box, frame)
	}
	want := max(w.headWide()+nameWide+yearWide+2*monthWide, w.offersWide())
	if box.W != want || box.H != CalendarFor(0, 0, want).H {
		t.Errorf("the calendar is %dx%d, want %d wide", box.W, box.H, want)
	}
}

func TestAClickOnTheDeadlineOpensTheCalendar(t *testing.T) {
	w := holding(newFake(), engine.Card{Number: 1, Title: "A", State: "todo"})
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

func TestThePointerWalksTheMonths(t *testing.T) {
	w := Window{screen: OnCalendar, width: 1200, height: 800, calendarState: calendarState{day: time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)}}
	w.form = editing(engine.Card{Number: 1, Title: "A"})
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

func TestTheCalendarIsDrawn(t *testing.T) {
	w := Window{screen: OnCalendar, width: 1200, height: 800, calendarState: calendarState{day: time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC)}}
	w.form = editing(engine.Card{Number: 1, Title: "Write the spec"})
	shot := drawn(t, w)
	for i, day := range engine.MonthGrid(w.day) {
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

func TestAPanelsOffersAreAsWideAsWhatTheySay(t *testing.T) {
	w := holding(newFake(), engine.Card{Number: 1, Title: "A", State: "todo"})
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
	no := boxes[1]
	if got := w.Click(Point{X: no.X + no.W - 2, Y: no.Y + no.H/2}); got.Screen() != OnForm || got.Form().Fields()[got.Form().On()].Value != "" {
		t.Errorf("a click at the end of %q left the window on %v", offers[1], got.Screen())
	}
}

func TestTheFormUnderAPanelOffersNothing(t *testing.T) {
	w := holding(newFake(), engine.Card{Number: 1, Title: "A", State: "todo"})
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

func TestTheContextLineStartsAfterTheOffersDrawn(t *testing.T) {
	w := holding(newFake(), engine.Card{Number: 1, Title: "A", State: "todo", Context: "/notes/1"})
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

func TestACalendarWithNoFieldIsInTheMiddle(t *testing.T) {
	w := Window{screen: OnCalendar, width: 1200, height: 800, calendarState: calendarState{day: time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC)}}
	box := w.CalendarBox()
	if box.X != (1200-box.W)/2 || box.Y != (800-box.H)/2 {
		t.Errorf("the calendar is at %+v", box)
	}
}

func TestTheMonthArrowsAreOnChips(t *testing.T) {
	w := Window{screen: OnCalendar, width: 1200, height: 800, calendarState: calendarState{day: time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC)}}
	w.form = editing(engine.Card{Number: 1, Title: "A"})
	shot, skin := drawn(t, w), Dark()
	box := w.CalendarBox()
	for _, chip := range []Box{CalendarMonth(box), CalendarYear(box), MonthBefore(box), MonthAfter(box)} {
		if got := painted(shot, Edges(chip)[0], skin.Edge); got != chip.W {
			t.Errorf("the chip at %+v is edged on %d of %d pixels", chip, got, chip.W)
		}
	}
}

func TestAClickOffTheCalendarClosesIt(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "A", State: "todo"})
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
	if left := typed.Click(Point{X: 2, Y: 2}); left.Screen() != OnAsking || f.what("task", "edit") != 0 {
		t.Errorf("a click off a form typed into left the window on %v, %d edits asked", left.Screen(), f.what("task", "edit"))
	}
}

func TestACalendarOverATaskWithStepsIsDrawn(t *testing.T) {
	w := holding(newFake(), engine.Card{Number: 1, Title: "Ship it", State: "todo",
		Steps: []engine.Part{{Number: 1, Title: "Draft the columns"}}})
	w.chosen = 1
	w = w.Edit()
	for i, field := range w.Form().Fields() {
		if field.Arg == "deadline" {
			_, value, _ := FormPlace(w.formBox(), w.Form().Fields(), i)
			w = w.clicked(Point{X: value.X + 4, Y: value.Y + 4})
		}
	}
	if w.Screen() != OnCalendar {
		t.Fatalf("the deadline left the window on %v", w.Screen())
	}
	if parts := w.PartsTicked(); len(parts) != 1 {
		t.Fatalf("under the calendar the steps are drawn as %+v", parts)
	}
	drawn(t, w)
}

func TestTheCalendarsDaysHaveRoomAndItStaysInTheSmallestWindow(t *testing.T) {
	w := Window{screen: OnCalendar, width: 1024, height: 640, calendarState: calendarState{day: time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC)}}
	w.form = editing(engine.Card{Number: 1, Title: "A"})
	box := w.CalendarBox()
	day, next := DayAt(box, 0), DayAt(box, 1)
	if day.H < 32 {
		t.Errorf("a day is %d high, want at least 32", day.H)
	}
	if gap := next.X - day.X - day.W; gap < 6 {
		t.Errorf("days are %d apart, want at least 6", gap)
	}
	if over := Over(w.width, w.height); box.Y < over.Y || box.Y+box.H > over.Y+over.H || box.X < 0 || box.X+box.W > w.width {
		t.Errorf("the calendar at %+v leaves the window %+v", box, over)
	}
}

func TestTheCalendarsGridsAreWalkedAndLeftByTheKeyboard(t *testing.T) {
	w := Window{screen: OnCalendar, width: 1200, height: 800, calendarState: calendarState{day: time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC)}}
	w.form = editing(engine.Card{Number: 1, Title: "A"})
	middle := func(b Box) Point { return Point{X: b.X + b.W/2, Y: b.Y + b.H/2} }
	months := w.Key("m")
	if got := months.CalendarCells(); len(got) != 12 || got[0] != "Jan" {
		t.Fatalf("m offers %v, want the twelve months", got)
	}
	if got := domain.FormatDay(months.Key("left").Key("down").day); got != "2026-12-12" {
		t.Errorf("left then down in the months reach %s, want 2026-12-12", got)
	}
	if got := domain.FormatDay(months.Key("up").day); got != "2026-07-12" {
		t.Errorf("up in the months reaches %s, want 2026-07-12", got)
	}
	if got := domain.FormatDay(months.clicked(middle(MonthAfter(months.CalendarBox()))).day); got != "2027-10-12" {
		t.Errorf("the arrow in the months reaches %s, want a year on", got)
	}
	if got := months.Key("enter"); got.view != viewDays || got.Screen() != OnCalendar {
		t.Errorf("enter on a month leaves the calendar on view %d, screen %v", got.view, got.Screen())
	}
	if got := months.Key("esc"); got.view != viewDays || got.Screen() != OnCalendar || domain.FormatDay(got.day) != "2026-10-12" {
		t.Errorf("esc in the months leaves view %d, screen %v", got.view, got.Screen())
	}
	years := w.Key("y")
	if got := domain.FormatDay(years.Key("right").Key("down").day); got != "2030-10-12" {
		t.Errorf("right then down in the years reach %s, want 2030-10-12", got)
	}
	if got := domain.FormatDay(years.clicked(middle(MonthBefore(years.CalendarBox()))).day); got != "2014-10-12" {
		t.Errorf("the arrow in the years reaches %s, want twelve years back", got)
	}
	if got := years.Key("enter"); got.view != viewMonths {
		t.Errorf("enter on a year leaves view %d, want the months", got.view)
	}
	if got := years.Key("esc"); got.view != viewDays || got.Screen() != OnCalendar {
		t.Errorf("esc in the years leaves view %d, screen %v", got.view, got.Screen())
	}
	if got := w.Key("esc"); got.Screen() != OnForm {
		t.Errorf("esc on the days leaves screen %v, want the form", got.Screen())
	}
}
