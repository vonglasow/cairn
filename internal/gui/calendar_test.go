package gui

import (
	"cairn/internal/board"
	"cairn/internal/engine/calendar"
	"cairn/internal/gui/design"
	"cairn/internal/gui/geom"
	"testing"
	"time"

	"cairn/internal/engine"
)

func TestTheCalendarIsLaidOutToThePixel(t *testing.T) {
	box := CalendarFor(100, 50, 400)
	if want := (geom.Box{X: 100, Y: 50, W: 400, H: 358}); box != want {
		t.Errorf("the calendar is %+v, want %+v", box, want)
	}
	if got, want := DayAt(box, 0), (geom.Box{X: 112, Y: 124, W: 47, H: 32}); got != want {
		t.Errorf("the first day is at %+v, want %+v", got, want)
	}
	if got, want := DayAt(box, 9), (geom.Box{X: 218, Y: 162, W: 47, H: 32}); got != want {
		t.Errorf("the tenth day is at %+v, want %+v", got, want)
	}
	if got, want := WeekdayAt(box, 2), (geom.Box{X: 218, Y: 100, W: 47, H: 18}); got != want {
		t.Errorf("Wednesday is written at %+v, want %+v", got, want)
	}
	if got, want := MonthAfter(box), (geom.Box{X: 456, Y: 62, W: monthWide, H: calendarHead}); got != want {
		t.Errorf("the next month is offered at %+v, want %+v", got, want)
	}
	if got, want := CalendarMonth(box), (geom.Box{X: 112, Y: 62, W: nameWide, H: calendarHead}); got != want {
		t.Errorf("the month is offered at %+v, want %+v", got, want)
	}
	if got, want := MonthBefore(box), (geom.Box{X: 456 - monthWide - inset, Y: 62, W: monthWide, H: calendarHead}); got != want {
		t.Errorf("the month before is offered at %+v, want %+v", got, want)
	}
	if got := len(board.MonthGrid(time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC))); got != 42 {
		t.Errorf("a month is shown as %d days, want six weeks", got)
	}
}

func TestTheCalendarHangsUnderItsField(t *testing.T) {
	w := holding(newFake(), board.Card{Number: 1, Title: "A", State: "todo"})
	w = w.clicked(middleOf(w, 1, 0)).Choose("Edit")
	for w.Form.Fields[w.Form.On()].Arg != "deadline" {
		w = w.Key("tab")
	}
	w = w.Key("enter")
	box, frame := w.CalendarBox(), fieldFrame(w)
	if box.X > frame.X || box.X+box.W < frame.X+frame.W || box.Y != frame.Y+frame.H+between {
		t.Errorf("the calendar is at %+v, under a field at %+v", box, frame)
	}
	want := max(w.headWide()+nameWide+yearWide+2*monthWide, w.offersWide())
	if box.W != want || box.H != CalendarFor(0, 0, want).H {
		t.Errorf("the calendar is %dx%d, want %d wide", box.W, box.H, want)
	}
}

func TestAClickOnTheDeadlineOpensTheCalendar(t *testing.T) {
	w := holding(newFake(), board.Card{Number: 1, Title: "A", State: "todo"})
	w = w.clicked(middleOf(w, 1, 0)).Choose("Edit")
	for _, c := range []struct {
		arg    string
		screen board.Screen
	}{{"deadline", board.OnCalendar}, {"estimate", board.OnForm}} {
		for i, field := range w.Form.Fields {
			if field.Arg != c.arg {
				continue
			}
			_, value, _ := FormPlace(w.formBox(), w.Form.Fields, i)
			if got := w.clicked(geom.Point{X: value.X + 4, Y: value.Y + 4}); got.Screen != c.screen || got.Form.On() != i {
				t.Errorf("a click on the %s left the window on %v, on field %d", c.arg, got.Screen, got.Form.On())
			}
		}
	}
}

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

func TestTheCalendarIsDrawn(t *testing.T) {
	w := Window{Board: board.Board{Screen: board.OnCalendar, CalendarState: board.CalendarState{Day: time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC)}}, width: 1200, height: 800}
	w.Form = Window{}.editing(board.Card{Number: 1, Title: "Write the spec"})
	shot := drawn(t, w)
	for i, day := range board.MonthGrid(w.Day) {
		edge := Edges(DayAt(w.CalendarBox(), i))[0]
		outlined := ink(shot, edge, design.Dark().Colour.Surface.Raised) > 0
		if day.Equal(w.Day) && !outlined {
			t.Error("the day the keyboard is on is not outlined")
		}
		if !day.Equal(w.Day) && outlined {
			t.Errorf("%s is outlined, and the keyboard is not on it", calendar.FormatDay(day))
		}
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

func TestTheFormUnderAPanelOffersNothing(t *testing.T) {
	w := holding(newFake(), board.Card{Number: 1, Title: "A", State: "todo"})
	w = w.clicked(middleOf(w, 1, 0)).Choose("Edit")
	foot := ButtonIn(w.PanelOffers()[0])
	skin := design.Dark()
	if painted(drawn(t, w), foot, skin.Colour.Accent.Base) == 0 {
		t.Fatal("the open form draws no offer at its foot")
	}
	for w.Form.Fields[w.Form.On()].Arg != "deadline" {
		w = w.Key("tab")
	}
	got := w.Key("enter")
	calendar := got.CalendarBox()
	if left := (geom.Box{X: foot.X, Y: foot.Y, W: max(0, min(foot.W, calendar.X-foot.X)), H: foot.H}); left.W == 0 {
		foot = geom.Box{X: calendar.X + calendar.W, Y: foot.Y, W: max(0, foot.X+foot.W-calendar.X-calendar.W), H: foot.H}
	} else {
		foot = left
	}
	if foot.W > 0 && painted(drawn(t, got), foot, skin.Colour.Accent.Base) != 0 {
		t.Error("the form under the calendar still draws offers at its foot")
	}
}

func TestTheContextLineEndsBeforeTheOffersDrawn(t *testing.T) {
	w := holding(newFake(), board.Card{Number: 1, Title: "A", State: "todo", Context: "/notes/1"})
	w.keyboard = KeyboardOf("windows")
	w = w.clicked(middleOf(w, 1, 0)).Choose("Edit")
	offers := w.PanelOffers()
	last := offers[len(offers)-1]
	said := w.ContextSays()
	if len(said) != 2 {
		t.Fatalf("the context says %d things", len(said))
	}
	name, path := said[0].inSight(), said[1].inSight()
	if name.X != w.formPanel().X+padding {
		t.Errorf("the context's name starts at %d, not at the panel's left", name.X)
	}
	if path.X != name.X+name.W+copyRoom || path.X+path.W != last.X-padding {
		t.Errorf("the path runs %+v after a name %+v, and the leftmost offer starts at %d", path, name, last.X)
	}
}

func TestACalendarWithNoFieldIsInTheMiddle(t *testing.T) {
	w := Window{Board: board.Board{Screen: board.OnCalendar, CalendarState: board.CalendarState{Day: time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC)}}, width: 1200, height: 800}
	box := w.CalendarBox()
	if box.X != (1200-box.W)/2 || box.Y != (800-box.H)/2 {
		t.Errorf("the calendar is at %+v", box)
	}
}

func TestTheMonthArrowsAreOnChips(t *testing.T) {
	w := Window{Board: board.Board{Screen: board.OnCalendar, CalendarState: board.CalendarState{Day: time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC)}}, width: 1200, height: 800}
	w.Form = Window{}.editing(board.Card{Number: 1, Title: "A"})
	shot, skin := drawn(t, w), design.Dark()
	box := w.CalendarBox()
	for _, chip := range []geom.Box{CalendarMonth(box), CalendarYear(box), MonthBefore(box), MonthAfter(box)} {
		row := geom.Box{X: chip.X + design.Dark().Radius.Control, Y: chip.Y + 2, W: chip.W - 2*design.Dark().Radius.Control, H: 1}
		if got := painted(shot, row, skin.Colour.Surface.Control) + painted(shot, row, skin.Colour.Accent.Base); got != row.W {
			t.Errorf("the chip at %+v is filled on %d of %d pixels", chip, got, row.W)
		}
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

func TestACalendarOverATaskWithStepsIsDrawn(t *testing.T) {
	w := holding(newFake(), board.Card{Number: 1, Title: "Ship it", State: "todo",
		Steps: []engine.Part{{Number: 1, Title: "Draft the columns"}}})
	w.Chosen = 1
	w = w.Edit()
	for i, field := range w.Form.Fields {
		if field.Arg == "deadline" {
			_, value, _ := FormPlace(w.formBox(), w.Form.Fields, i)
			w = w.clicked(geom.Point{X: value.X + 4, Y: value.Y + 4})
		}
	}
	if w.Screen != board.OnCalendar {
		t.Fatalf("the deadline left the window on %v", w.Screen)
	}
	if parts := w.PartsTicked(); len(parts) != 1 {
		t.Fatalf("under the calendar the steps are drawn as %+v", parts)
	}
	drawn(t, w)
}

func TestTheCalendarsDaysHaveRoomAndItStaysInTheSmallestWindow(t *testing.T) {
	w := Window{Board: board.Board{Screen: board.OnCalendar, CalendarState: board.CalendarState{Day: time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC)}}, width: 1024, height: 640}
	w.Form = Window{}.editing(board.Card{Number: 1, Title: "A"})
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
