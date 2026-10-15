package gui

import (
	"fmt"
	"slices"
	"testing"
	"time"

	"cairn/internal/domain"
	"cairn/internal/engine"
)

func many(day string, n int) engine.TimesheetDay {
	d := engine.TimesheetDay{Day: day}
	for i := 1; i <= n; i++ {
		d.Tasks = append(d.Tasks, engine.TimesheetTask{Number: i, Title: fmt.Sprintf("Task %d", i), Minutes: 15})
		d.Minutes += 15
	}
	return d
}

func TestTheTimePageIsLaidOutToThePixel(t *testing.T) {
	l := LayoutFor(1200, 800)
	left, right := l.Column[0].X, l.Column[engine.Columns-1].X+l.Column[engine.Columns-1].W
	strip := TimeStrip(l)
	if strip != (Box{X: left, Y: l.Head[0].Y, W: right - left, H: barHigh}) {
		t.Errorf("the strip is %+v", strip)
	}
	before, after := TimeChip(l, "‹"), TimeChip(l, "›")
	if before != (Box{X: left, Y: strip.Y, W: arrowWide, H: barHigh}) || after.X != before.X+arrowWide+between {
		t.Errorf("the arrows are %+v and %+v", before, after)
	}
	month, week, day, today := TimeChip(l, "Month"), TimeChip(l, "Week"), TimeChip(l, "Day"), TimeChip(l, "Today")
	if month.X+month.W != right || week.X+week.W+between != month.X || day.X+day.W+between != week.X || today.X+today.W+2*between != day.X || month.W != rangeWide {
		t.Errorf("the ranges are %+v, %+v, %+v, %+v", today, day, week, month)
	}
	if got := TimeChip(l, "nothing"); got != (Box{}) {
		t.Errorf("a chip nobody named is %+v", got)
	}
	if said := TimeSaid(l); said.X != after.X+arrowWide+padding || said.X+said.W != today.X-padding {
		t.Errorf("what is shown is said in %+v", said)
	}
	first := WeekDayAt(l, 0)
	if first.Y != strip.Y+strip.H+between || first.Y+first.H != l.Band.Y+l.Band.H {
		t.Errorf("the first day is %+v", first)
	}
	if got := DayRoom(Box{H: 2*lineHigh + padding + 3*weekRow + 5}); got != 3 {
		t.Errorf("a day with room for three tasks holds %d", got)
	}
	if got := DayRoom(Box{H: 10}); got != 1 {
		t.Errorf("a day with room for none holds %d", got)
	}
	cell, ninth := MonthCellAt(l, 0), MonthCellAt(l, 8)
	top := strip.Y + strip.H + between + lineHigh
	wide, high := (right-left-6*gap)/7, (l.Band.Y+l.Band.H-top-5*gap)/6
	if cell != (Box{X: left, Y: top, W: wide, H: high}) {
		t.Errorf("the first day of a month is %+v", cell)
	}
	if ninth != (Box{X: left + wide + gap, Y: top + high + gap, W: wide, H: high}) {
		t.Errorf("the ninth day of a month is %+v", ninth)
	}
	if name := MonthWeekdayAt(l, 2); name != (Box{X: left + 2*(wide+gap), Y: strip.Y + strip.H + between, W: wide, H: lineHigh}) {
		t.Errorf("Wednesday is written at %+v", name)
	}
	if got := CellRoom(Box{H: 4*lineHigh + padding + 3}); got != 3 {
		t.Errorf("a day of a month with room for three lines holds %d", got)
	}
}

func TestADayScrollsToItsLastTask(t *testing.T) {
	w := Window{work: newFake(), tab: Time, width: 1200, height: 800, timeState: timeState{week: dayNamed("2026-09-28"), sheet: engine.Timesheet{Days: []engine.TimesheetDay{many("2026-09-29", 20), many("2026-09-30", 2)}}}}
	l := w.Layout()
	tuesday := WeekDayAt(l, 1)
	room := DayRoom(tuesday)
	if room >= 20 {
		t.Fatalf("a day holds %d tasks: nothing here needs scrolling", room)
	}
	in := Point{X: tuesday.X + 4, Y: tuesday.Y + tuesday.H/2}
	ended := w.Scroll(in, 99*weekRow)
	if got := ended.DayFrom(1); got != 20-room {
		t.Errorf("turned past its end, the day starts at %d, want %d", got, 20-room)
	}
	shown := ended.DayShown(1)
	if len(shown) != room || shown[len(shown)-1].Title != "Task 20" {
		t.Errorf("the day shows %d ending with %+v", len(shown), shown[len(shown)-1])
	}
	if got := ended.DayFrom(2); got != 0 {
		t.Errorf("Wednesday moved to %d", got)
	}
	if got := ended.Scroll(in, -99*weekRow).DayFrom(1); got != 0 {
		t.Errorf("turned back past its start, the day starts at %d", got)
	}
	if got := w.Scroll(in, weekRow/2); got.DayFrom(1) != 0 || got.spareChoosing != weekRow/2 {
		t.Errorf("half a notch moved the day to %d, keeping %d", got.DayFrom(1), got.spareChoosing)
	}
	if got := w.Scroll(Point{X: 1, Y: 1}, 99*weekRow); got.DayFrom(1) != 0 {
		t.Error("the wheel off the days scrolled one")
	}
	if got := w.DayShown(2); len(got) != 2 {
		t.Errorf("a day of two tasks shows %d", len(got))
	}
	if got := ended.onWeek("left").DayFrom(1); got != 0 {
		t.Errorf("the week before starts Tuesday at %d", got)
	}
}

func TestTheTimePageShowsAWeekOrAMonth(t *testing.T) {
	f := newFake()
	w := Window{work: f, tab: Time, width: 1200, height: 800, now: dayNamed("2026-10-14"), timeState: timeState{week: dayNamed("2026-10-12")}}
	month := w.onWeek("m")
	if !month.Monthly() || month.week.Format("January 2006") != "October 2026" {
		t.Fatalf("m shows %v, monthly %v", month.week, month.Monthly())
	}
	in, _ := f.gave("timesheet", "show")
	grid := engine.MonthGrid(dayNamed("2026-10-01"))
	if in.Given["since"] != domain.FormatDay(grid[0]) || in.Given["until"] != "2026-10-14" {
		t.Errorf("the month asked for %v", in.Given)
	}
	if got := (Window{work: f, now: dayNamed("2026-10-14"), timeState: timeState{week: dayNamed("2026-09-28")}}).onWeek("m"); got.week.Format("January 2006") != "October 2026" {
		t.Errorf("the week of 28 September turned to %s", got.week.Format("January 2006"))
	}
	if got := (Window{work: f, now: dayNamed("2026-09-30"), timeState: timeState{week: dayNamed("2026-09-28")}}).onWeek("m"); got.week.Format("January 2006") != "September 2026" {
		t.Errorf("the week of 28 September, today the 30th, turned to %s", got.week.Format("January 2006"))
	}
	if got := month.onWeek("right").week.Format("January 2006"); got != "November 2026" {
		t.Errorf("right shows %s", got)
	}
	if got := month.onWeek("left").onWeek("t").week.Format("January 2006"); got != "October 2026" {
		t.Errorf("t shows %s", got)
	}
	if got := month.onWeek("m"); got.Monthly() || domain.FormatDay(got.week) != "2026-10-12" {
		t.Errorf("m again shows the week of %s, monthly %v", domain.FormatDay(got.week), got.Monthly())
	}
	later := month.onWeek("right").onWeek("right")
	if got := later.onWeek("m"); domain.FormatDay(got.week) != "2026-11-30" {
		t.Errorf("the week of December shown is %s", domain.FormatDay(got.week))
	}
	l := w.Layout()
	middle := func(b Box) Point { return Point{X: b.X + b.W/2, Y: b.Y + b.H/2} }
	for chip, want := range map[string]string{"‹": "2026-10-05", "›": "2026-10-19", "Today": "2026-10-12", "Week": "2026-10-12"} {
		if got := w.Click(middle(TimeChip(l, chip))); got.Monthly() || domain.FormatDay(got.week) != want {
			t.Errorf("%s shows the week of %s", chip, domain.FormatDay(got.week))
		}
	}
	if got := w.Click(middle(TimeChip(l, "Month"))); !got.Monthly() {
		t.Error("Month shows no month")
	}
	if got := month.Click(middle(TimeChip(l, "Month"))); !got.Monthly() {
		t.Error("Month on a month left it")
	}
	fourth := slices.IndexFunc(grid, func(d time.Time) bool { return d.Day() == 4 })
	if got := month.Click(middle(MonthCellAt(l, fourth))); got.Monthly() || domain.FormatDay(got.week) != "2026-09-28" {
		t.Errorf("the 4th clicked shows the week of %s, monthly %v", domain.FormatDay(got.week), got.Monthly())
	}
	if got := month.Click(Point{X: 1, Y: l.Band.Y + l.Band.H + 2}); !got.Monthly() {
		t.Error("a click under the month left it")
	}
	if got := TimeSays(w); got != "12 – 18 Oct 2026" {
		t.Errorf("the week says %q", got)
	}
	if got := TimeSays(month); got != "October 2026" {
		t.Errorf("the month says %q", got)
	}
}

func TestADayOfAMonthSaysHowManyMore(t *testing.T) {
	five := many("2026-10-01", 5).Tasks
	if got := CellLines(five, 5); len(got) != 5 || got[4] != five[4].Says() {
		t.Errorf("five in room for five read %q", got)
	}
	if got := CellLines(five, 3); !slices.Equal(got, []string{five[0].Says(), five[1].Says(), "+3 more"}) {
		t.Errorf("five in room for three read %q", got)
	}
	if got := CellLines(five[:2], 1); !slices.Equal(got, []string{"+2 more"}) {
		t.Errorf("two in room for one read %q", got)
	}
	if got := CellLines(nil, 3); len(got) != 0 {
		t.Errorf("nothing reads %q", got)
	}
	w := Window{tab: Time, width: 1200, height: 800, timeState: timeState{monthly: true, week: dayNamed("2026-10-01"), sheet: engine.Timesheet{Days: []engine.TimesheetDay{many("2026-10-01", 2)}}}}
	at := slices.IndexFunc(engine.MonthGrid(w.week), func(d time.Time) bool { return d.Day() == 1 })
	if got := w.MonthDay(at); got.Day != "2026-10-01" || len(got.Tasks) != 2 {
		t.Errorf("the 1st reads %+v", got)
	}
}

func TestTheTimePageAsksNothingOfTheFuture(t *testing.T) {
	f := newFake()
	w := Window{work: f, tab: Time, now: dayNamed("2026-10-14"), timeState: timeState{week: dayNamed("2026-10-12")}}
	w.readingWeek()
	if in, _ := f.gave("timesheet", "show"); in.Given["since"] != "2026-10-12" || in.Given["until"] != "2026-10-14" {
		t.Errorf("the week in progress asked for %v", in.Given)
	}
	asked := f.what("timesheet", "show")
	w.week = dayNamed("2026-10-19")
	if got := w.readingWeek(); f.what("timesheet", "show") != asked || len(got.sheet.Days) != 0 {
		t.Errorf("a week to come was asked for, and reads %+v", got.sheet)
	}
}

func TestTheRangeShownIsFilledAndTodayOutlined(t *testing.T) {
	w := Window{work: newFake(), tab: Time, width: 1200, height: 800, now: dayNamed("2026-09-30"), timeState: timeState{week: dayNamed("2026-09-28")}}
	l := w.Layout()
	outlined := func(v Window, b Box) bool { return painted(drawn(t, v), Edges(b)[0], Dark().Chosen) == b.W }
	filled := func(v Window, b Box) bool { return painted(drawn(t, v), b, Dark().Chosen) > b.W*b.H/2 }
	if !filled(w, TimeChip(l, "Week")) || filled(w, TimeChip(l, "Month")) {
		t.Error("a week shown does not fill Week alone")
	}
	m := w
	m.monthly, m.week = true, dayNamed("2026-09-01")
	if !filled(m, TimeChip(l, "Month")) || filled(m, TimeChip(l, "Week")) {
		t.Error("a month shown does not fill Month alone")
	}
	grid := engine.MonthGrid(m.week)
	today := slices.IndexFunc(grid, func(d time.Time) bool { return d.Equal(dayNamed("2026-09-30")) })
	if !outlined(m, MonthCellAt(l, today)) || outlined(m, MonthCellAt(l, today-1)) {
		t.Error("today is not the one day outlined")
	}
}

func TestWhatTheTimePageIsDrawnOn(t *testing.T) {
	day := Box{X: 100, Y: 50, W: 200, H: 2*lineHigh + padding + 3*weekRow}
	if _, more := DayTrack(day, 3); more {
		t.Error("a day showing its three tasks draws a bar")
	}
	if got, more := DayTrack(day, 4); !more || got != (Box{X: 290, Y: WeekRowAt(day, 0).Y, W: barWide, H: 3 * weekRow}) {
		t.Errorf("a day of four in room for three draws %+v, %v", got, more)
	}
	if got := DayRoom(Box{H: 2*lineHigh + padding + 3*weekRow - 1}); got != 2 {
		t.Errorf("a pixel short of three, a day holds %d", got)
	}
	if got := CellHead(dayNamed("2026-10-04"), engine.TimesheetDay{}); got != "4" {
		t.Errorf("a day nothing was spent on reads %q", got)
	}
	if got := CellHead(dayNamed("2026-10-04"), engine.TimesheetDay{Minutes: 90}); got != "4   1h30m" {
		t.Errorf("a day of an hour and a half reads %q", got)
	}
	cell := Box{X: 10, Y: 20, W: 100, H: 90}
	if got := CellHeadAt(cell); got != Row(cell, 20+inset, lineHigh) {
		t.Errorf("the head is at %+v", got)
	}
	if got := CellLineAt(cell, 1); got != Row(cell, 20+inset+2*lineHigh, lineHigh) {
		t.Errorf("the second line is at %+v", got)
	}
	if !sameMonth(dayNamed("2026-10-01"), dayNamed("2026-10-31")) || sameMonth(dayNamed("2026-10-01"), dayNamed("2025-10-01")) || sameMonth(dayNamed("2026-10-01"), dayNamed("2026-11-01")) {
		t.Error("sameMonth is wrong")
	}
	for week, want := range map[string]string{"2026-09-28": "28 Sep – 4 Oct 2026", "2026-12-28": "28 Dec 2026 – 3 Jan 2027", "2026-10-12": "12 – 18 Oct 2026"} {
		if got := TimeSays(Window{timeState: timeState{week: dayNamed(week)}}); got != want {
			t.Errorf("the week of %s says %q", week, got)
		}
	}
	m := Window{work: newFake(), now: dayNamed("2026-10-14"), timeState: timeState{monthly: true, week: dayNamed("2026-10-01")}}
	if got := m.onWeek("right").week; domain.FormatDay(got) != "2026-11-01" {
		t.Errorf("right from October shows %s", domain.FormatDay(got))
	}
	if got := m.onWeek("left").week; domain.FormatDay(got) != "2026-09-01" {
		t.Errorf("left from October shows %s", domain.FormatDay(got))
	}
}

func TestADayOfTheWeekLaysItsTasksOutAsCards(t *testing.T) {
	day := Box{X: 10, Y: 20, W: 150, H: 600}
	first, second := DayCardAt(day, 0), DayCardAt(day, 1)
	if first.H != CardHeight || first.W != day.W-2*padding || second.Y != first.Y+weekRow {
		t.Errorf("the cards of a day are %+v then %+v", first, second)
	}
	if bar := WeekBar(WeekRowAt(day, 0), 30, 60); bar.Y < first.Y+first.H-gap {
		t.Errorf("the bar %+v is drawn over the card %+v", bar, first)
	}
}

func TestATaskOfTheTimePageIsTheCardTheBoardHolds(t *testing.T) {
	w := Window{snap: engine.Snapshot{Cards: []engine.Card{{Number: 1, Title: "Write the spec", Estimate: "3h"}}}}
	if c := w.DayCard(engine.TimesheetTask{Number: 1, Title: "Write the spec"}); c.Estimate != "3h" {
		t.Errorf("the time page's card is %+v", c)
	}
	if c := w.DayCard(engine.TimesheetTask{Number: 9, Title: "Gone"}); c.Number != 9 || c.Title != "Gone" {
		t.Errorf("a task the board no longer holds is %+v", c)
	}
}

func TestTheRangesOfTheTimePageAreButtons(t *testing.T) {
	w := Window{work: newFake(), tab: Time, width: 1200, height: 800, timeState: timeState{week: dayNamed("2026-09-28")}}
	l, skin := w.Layout(), Dark()
	shot := drawn(t, w)
	for _, chip := range []string{"Today", "Day", "Month"} {
		box := TimeChip(l, chip)
		if painted(shot, box, skin.Panel) < box.W*box.H/2 {
			t.Errorf("%s is not drawn as a button: too little of it is filled", chip)
		}
	}
	week := TimeChip(l, "Week")
	if painted(shot, week, skin.Chosen) < week.W*week.H/2 {
		t.Errorf("the range shown is not filled as the one chosen")
	}
}

func TestAChipSaysItsNameInItsMiddle(t *testing.T) {
	chip := Box{X: 100, Y: 10, W: 72, H: 30}
	at := ChipSaid(chip, 32)
	if at.X != 120 || at.W != 32 || at.Y != chip.Y+(chip.H-lineHigh)/2 || at.H != lineHigh {
		t.Errorf("a name 32 wide is said at %+v in %+v", at, chip)
	}
	if wide := ChipSaid(chip, 90); wide.X != chip.X || wide.W != chip.W {
		t.Errorf("a name wider than its chip is said at %+v", wide)
	}
}

func TestANarrowCardLeavesOutWhatItHasNoRoomFor(t *testing.T) {
	c := engine.Card{Labels: []string{"feature", "release-26.1.6", "ado", "77711"}}
	if got := Tagged(c, 10); got != "" {
		t.Errorf("in room for nothing the labels say %q", got)
	}
	if got := Tagged(c, 3*perSmallGlyph); got != "+4" {
		t.Errorf("in room for their count the labels say %q", got)
	}
	number := Box{X: 0, Y: 0, W: 80, H: lineHigh}
	if !Crowded(number, 60, "▪ 2/5") {
		t.Errorf("a count written over what the number line already holds is not crowded")
	}
	if Crowded(number, 20, "▪ 2/5") {
		t.Errorf("a count with room to spare is crowded")
	}
	labels := Box{X: 10, Y: 0, W: 100, H: lineHigh}
	if at := LabelsAt(labels, "12h30m/4h"); at.X != labels.X || at.W != LabelsRoom(labels, "12h30m/4h") {
		t.Errorf("the labels are written in %+v, past the facts' room", at)
	}
}

func TestADayAloneHoldsItsTasksAcrossTheBoard(t *testing.T) {
	l := LayoutFor(1200, 800)
	strip, alone := TimeStrip(l), DayAloneAt(l)
	if alone.X != strip.X || alone.W != strip.W || alone.Y != WeekDayAt(l, 0).Y || alone.Y+alone.H != l.Band.Y+l.Band.H {
		t.Errorf("a day alone is %+v", alone)
	}
	card := AloneCardAt(alone, 1)
	if card.X != alone.X+padding || card.W != alone.W/3-padding || card.Y != WeekRowAt(alone, 1).Y || card.H != CardHeight {
		t.Errorf("its second card is %+v", card)
	}
	bar := AloneBar(alone, 1, 30, 60)
	if bar.X <= card.X+card.W || bar.X+bar.W > alone.X+alone.W-padding {
		t.Errorf("its second bar is %+v, beside a card at %+v", bar, card)
	}
}
