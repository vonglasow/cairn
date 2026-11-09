package gui

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"slices"
	"strconv"
	"strings"
	"sync"

	"gioui.org/f32"
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"golang.org/x/image/math/fixed"

	"cairn/internal/engine"
)

const (
	small   = 12
	reads   = 13
	plain   = 14
	heading = 16
)

// HeadingSize is the size every panel's heading is drawn at.
const HeadingSize = 18

// HeadingWeight is the weight every panel's heading is drawn in.
const HeadingWeight = font.Bold

// Theme is the shaper and the sizes Gio needs, built once.
func Theme() *material.Theme {
	th := material.NewTheme()
	th.Shaper = text.NewShaper(text.WithCollection(Faces()))
	th.Face = TextFace
	return th
}

// Measured is the Measure of a theme's face: each line shaped the way a label shapes it, at the same size and in the same dp, and an edge read after every cluster of characters the face draws as one.
func Measured(th *material.Theme) Measure { return measuredIn(th, font.Normal) }

// MeasuredHeadings is the Measure of the same face in the weight a panel's heading is drawn in.
func MeasuredHeadings(th *material.Theme) Measure { return measuredIn(th, HeadingWeight) }

func measuredIn(th *material.Theme, weight font.Weight) Measure {
	return func(line string, size int) []int {
		edges := []int{0}
		th.Shaper.LayoutString(text.Parameters{
			Font: font.Font{Typeface: th.Face, Weight: weight}, PxPerEm: fixed.I(size),
			MaxLines: 1, MaxWidth: math.MaxInt32,
		}, line)
		for g, ok := th.Shaper.NextGlyph(); ok; g, ok = th.Shaper.NextGlyph() {
			if g.Flags&text.FlagClusterBreak == 0 {
				continue
			}
			for range g.Runes {
				edges = append(edges, (g.X + g.Advance).Round())
			}
		}
		return edges
	}
}

// Draw puts the whole window on screen.
func Draw(k Kit, w Window) {
	scaled := inDp(k.gtx)
	defer scaled.Pop()

	k.gtx.Metric = unit.Metric{PxPerDp: 1, PxPerSp: 1}

	fill(k.gtx, Box{W: w.width, H: w.height}, k.skin.Colour.Surface.Back)
	if w.Screen() == TooSmallToDraw {
		sayCentred(k.gtx, k.th, k.skin.Colour.Ink.Faint, Message(w.width, w.height), w.Says(), reads)
		return
	}
	if w.Blank() {
		return
	}
	l := w.Layout()
	drawPlace(k, l, w)
	drawTabs(k, l, w)
	if !w.native {
		for _, stroke := range MenuStrokes(MenuButton(l)) {
			fill(k.gtx, stroke, k.skin.Colour.Accent.Base)
		}
	}
	if w.Tab() == Objectives {
		drawGoals(k, l, w)
	} else if w.Tab() == Time {
		drawWeek(k, l, w)
	} else {
		drawColumns(k, l, w)
		drawBand(k, l, w)
		drawDragging(k.on(LayerDrag), w)
	}
	drawBar(k.on(LayerBar), l, w)
	drawPills(k.on(LayerBar), w)
	panel, popover := k.on(LayerPanel), k.on(LayerPopover)
	if w.formUnder() {
		drawForm(panel, w)
	}
	if w.Screen() == OnHelp {
		drawHelp(panel, w)
	}
	if w.Screen() == OnAsking {
		fill(k.gtx, Box{W: w.width, H: w.height}, k.skin.Colour.Surface.Veil)
		asked := w.QuestionBox()
		drawSaid(panel, w, w.Asks(), asked)
	}
	if w.Screen() == OnChoosing {
		drawChoosing(panel, w)
	}
	if w.Screen() == OnCalendar {
		drawCalendar(panel, w)
	}
	if w.Screen() == OnMenu {
		drawMenu(popover, w)
	}
	if w.Screen() == OnBackups {
		drawBackups(panel, w)
	}
	if w.Screen() == OnLabels {
		drawLabels(panel, w)
	}
	if w.Screen() == OnFiltering {
		drawFilters(popover, w)
	}
	if w.Screen() == OnConfiguration {
		drawConfiguration(panel, w)
	}
	if w.Screen() == OnPalette {
		drawPalette(popover, w)
	}
	if w.Screen() == OnWorkspaces {
		drawSwitcher(popover, w)
	}
	for _, mark := range w.CopyMarks() {
		drawCopyMark(popover, mark.At)
	}
	if bubble, said, ok := w.BubbleBox(); ok {
		raised(popover, bubble, k.skin.Radius.Card, k.skin.Colour.Surface.Raised)
		for i, line := range said {
			say(k.gtx, k.th, k.skin.Colour.Ink.Text, ExplainedLine(bubble, i), line, plain)
		}
	}
	if bubble, pills, ok := w.LabelsBubbleBox(); ok {
		raised(popover, bubble, k.skin.Radius.Card, k.skin.Colour.Surface.Raised)
		for i, pill := range pills {
			drawLabelPill(popover, w, pill, w.bubbled[i])
		}
	}
}

func drawBackups(k Kit, w Window) {
	box := w.BackupsBox()
	drawPanel(k, box, Heads(w))
	from := w.BackupsFrom()
	for i, one := range w.Backups()[from:min(len(w.Backups()), from+w.BackupsRoom())] {
		drawListRow(k, ListRowAt(box, i), backupSays(one), from+i == w.BackupOn())
	}
	drawPanelOffers(k, box, w)
}

func drawLabels(k Kit, w Window) {
	box := w.LabelsBox()
	drawPanel(k, box, Heads(w))
	m := w.measured()
	drawLabelSearch(k, w, m, w.LabelSearchAt())
	rows := w.LabelRows()
	if creating, ok := w.LabelCreating(); ok {
		say(k.gtx, k.th, k.skin.Colour.Accent.Base, Inside(ListRowAt(box, 1)), CreatingSays(creating), reads)
	} else if len(rows) == 0 {
		say(k.gtx, k.th, k.skin.Colour.Ink.Faint, Inside(ListRowAt(box, 1)), NoLabelSays, reads)
	}
	for _, row := range rows {
		drawLabelRow(k, w, m, row)
	}
	drawPanelOffers(k, box, w)
	drawSwatches(k, w)
}

func drawLabelSearch(k Kit, w Window, m Measure, row Box) {
	field := Box{X: row.X, Y: row.Y + 1, W: row.W, H: row.H - 2}
	edge := k.skin.Colour.Line.Rule
	if w.Seeking() {
		edge = k.skin.Colour.Accent.Base
	}
	framed(k, field, k.skin.Radius.Control, edge, k.skin.Colour.Surface.Field)
	chip := Box{X: field.X + inset, Y: field.Y + (field.H-chipHigh)/2, W: KeyChipWide("/"), H: chipHigh}
	rounded(k.gtx, chip, k.skin.Radius.Chip, k.skin.Colour.Surface.Control)
	sayOnChip(k, k.skin.Colour.Ink.Faint, chip, "/")
	line := Inside(Box{X: chip.X + chip.W, Y: field.Y, W: field.X + field.W - chip.X - chip.W, H: field.H})
	if w.Sought().Empty() && !w.Seeking() {
		say(k.gtx, k.th, k.skin.Colour.Ink.Faint, line, LabelSearchSays, reads)
		return
	}
	writing(k, m, line, w.Sought(), w.Seeking(), reads, 1, Follow)
}

func drawLabelRow(k Kit, w Window, m Measure, row LabelRow) {
	if row.On {
		rounded(k.gtx, row.At, k.skin.Radius.Row, k.skin.Colour.Surface.Control)
	}
	line := Inside(row.At)
	if row.Typing {
		field := Box{X: line.X, Y: row.At.Y + 2, W: line.W / 2, H: row.At.H - 4}
		framed(k, field, k.skin.Radius.Control, k.skin.Colour.Accent.Base, k.skin.Colour.Surface.Field)
		writing(k, m, Inside(field), w.Renamed(), true, reads, 1, Follow)
		return
	}
	rounded(k.gtx, row.Dot, k.skin.Radius.Full(row.Dot), k.skin.Colour.Labels[w.LabelColour(row.Carried.Label)])
	drawLabelPill(k, w, row.Name, row.Carried.Label)
	counts := line
	if len(row.Actions) > 0 {
		counts.W = row.Actions[0].At.X - padding - counts.X
	}
	sayIn(k.gtx, k.th, k.skin.Colour.Ink.Faint, counts, carriedSays(row.Carried), small, font.Font{}, text.End)
	for _, action := range row.Actions {
		look := buttonLit
		if action.Offer == "Delete" {
			look = buttonDanger
		}
		drawKeyedButton(k, action.At, m, action.Offer, w.OfferKey(action.Offer), look)
	}
}

func drawSwatches(k Kit, w Window) {
	swatches := w.ColourSwatches()
	if len(swatches) == 0 {
		return
	}
	first, last := swatches[0].At, swatches[len(swatches)-1].At
	raised(k, Box{X: first.X - inset*2, Y: first.Y - inset*2, W: last.X + last.W - first.X + inset*4, H: first.H + inset*4}, k.skin.Radius.Card, k.skin.Colour.Surface.Raised)
	for i, swatch := range swatches {
		if i == w.ColourPick() {
			ring := Box{X: swatch.At.X - 2, Y: swatch.At.Y - 2, W: swatch.At.W + 4, H: swatch.At.H + 4}
			rounded(k.gtx, ring, k.skin.Radius.Full(ring), k.skin.Colour.Ink.Text)
		}
		rounded(k.gtx, swatch.At, k.skin.Radius.Full(swatch.At), k.skin.Colour.Labels[swatch.Colour])
	}
}

func drawConfiguration(k Kit, w Window) {
	box := w.ConfigurationBox()
	drawPanel(k, box, Heads(w))
	mono := k.skin.Type.Mono
	lines := w.ConfigLines()
	head := w.ConfigHead()
	for i, said := range head {
		sayIn(k.gtx, k.th, k.skin.Colour.Ink.Faint, Inside(ListRowAt(box, i)), said, small, mono, text.Start)
	}
	for _, line := range lines {
		sayIn(k.gtx, k.th, k.skin.Colour.Ink.Low, line.KeyAt, line.Key, small, mono, text.Start)
		if line.Swatch != "" {
			rounded(k.gtx, line.SwatchAt, k.skin.Radius.Full(line.SwatchAt), k.skin.Colour.Labels[line.Swatch])
		}
		if line.Key != "pages.time" {
			say(k.gtx, k.th, k.skin.Colour.Ink.Text, line.ValueAt, line.Value, reads)
		}
		sayIn(k.gtx, k.th, k.skin.Colour.Ink.Faint, line.FromAt, line.fromSaid(), small, mono, text.End)
	}
	drawTimeSegments(k, w)
	for i, said := range w.ConfigFoot() {
		row := Inside(ListRowAt(box, len(head)+len(lines)+i))
		if i == 0 {
			row = Box{X: row.X + copyRoom, Y: row.Y, W: row.W - copyRoom, H: row.H}
		}
		say(k.gtx, k.th, k.skin.Colour.Ink.Faint, row, said, reads)
	}
	drawPanelOffers(k, box, w)
}

func drawFilters(k Kit, w Window) {
	box := w.FiltersBox()
	raised(k, box, k.skin.Radius.Panel, k.skin.Colour.Surface.Raised)
	head := Box{X: box.X + padding, Y: box.Y + (dropHead-lineHigh)/2, W: box.W - 2*padding, H: lineHigh}
	foot := Box{X: box.X, Y: box.Y + box.H - dropFoot, W: box.W, H: dropFoot}
	fill(k.gtx, Box{X: box.X, Y: foot.Y, W: box.W, H: 1}, k.skin.Colour.Line.Raised)
	x := foot.X + padding
	for _, key := range w.FilterFoot() {
		chip := drawKeyChip(k, Box{X: x, Y: foot.Y, H: foot.H}, key[0])
		said := Box{X: chip.X + chip.W + hintSpace, Y: foot.Y + (foot.H-lineHigh)/2, W: ends(w.measured(), key[1], small), H: lineHigh}
		sayIn(k.gtx, k.th, k.skin.Colour.Ink.Faint, said, key[1], small, font.Font{}, text.Start)
		x = said.X + said.W + padding
	}
	if values, open := w.FilterList(); open {
		f := w.Filters()[w.FilterOn()]
		fill(k.gtx, Box{X: box.X, Y: box.Y + dropHead - 1, W: box.W, H: 1}, k.skin.Colour.Line.Raised)
		say(k.gtx, k.th, k.skin.Colour.Ink.Text, head, Titled(f.Name), reads)
		sayIn(k.gtx, k.th, k.skin.Colour.Ink.Faint, head, w.FilterHint(), small, font.Font{}, text.End)
		counts := w.FilterCounts()
		for i, value := range values {
			if value.H == 0 {
				continue
			}
			taken := w.FilterTaken(i)
			if taken {
				rounded(k.gtx, value, k.skin.Radius.Row, k.skin.Colour.Surface.Tint)
			}
			if i == w.FilterPick() {
				rounded(k.gtx, value, k.skin.Radius.Row, k.skin.Colour.Surface.Control)
			}
			line := Inside(value)
			if taken {
				say(k.gtx, k.th, k.skin.Colour.Accent.Base, line, TickedMark, reads)
			}
			named := Box{X: line.X + tickRoom, Y: line.Y, W: line.W - tickRoom, H: line.H}
			if f.Arg == "label" {
				dot := Box{X: named.X, Y: named.Y + (named.H-dotSide)/2, W: dotSide, H: dotSide}
				rounded(k.gtx, dot, k.skin.Radius.Full(dot), k.skin.Colour.Labels[w.LabelColour(f.Shown[i])])
				named.X, named.W = named.X+dotSide+between, named.W-dotSide-between
			}
			say(k.gtx, k.th, k.skin.Colour.Ink.Text, named, f.Shown[i], reads)
			if i < len(counts) {
				sayIn(k.gtx, k.th, k.skin.Colour.Ink.Faint, line, strconv.Itoa(counts[i]), small, k.skin.Type.Mono, text.End)
			}
		}
		return
	}
	sayIn(k.gtx, k.th, k.skin.Colour.Ink.Low, head, FilterMenuSays, small, k.skin.Type.MonoBold, text.Start)
	for i, f := range w.Filters() {
		row := w.FilterRowAt(i)
		if i == w.FilterOn() {
			rounded(k.gtx, row, k.skin.Radius.Row, k.skin.Colour.Surface.Control)
		}
		line := Inside(row)
		say(k.gtx, k.th, k.skin.Colour.Ink.Text, line, Titled(f.Name), reads)
		if len(f.Taken()) > 0 {
			at := ends(w.measured(), Titled(f.Name), reads) + between
			sayIn(k.gtx, k.th, k.skin.Colour.Accent.Base, Box{X: line.X + at, Y: line.Y, W: line.W - at - padding - small, H: line.H}, f.TakenSays(), small, font.Font{}, text.Start)
		}
		sayIn(k.gtx, k.th, k.skin.Colour.Ink.Faint, line, FilterKey(f), small, k.skin.Type.Mono, text.End)
	}
}

func drawWeek(k Kit, l Layout, w Window) {
	measuring := k.gtx
	measuring.Constraints.Min = image.Point{}
	control := RangesControl(l)
	rounded(k.gtx, control, k.skin.Radius.Control+segmentInset, k.skin.Colour.Surface.Control)
	for _, chip := range TimeChips {
		at := TimeChip(l, chip)
		switch chip {
		case "Day", "Week", "Month":
			ink := k.skin.Colour.Ink.Faint
			if chip == w.ranging() {
				rounded(k.gtx, Box{X: at.X, Y: at.Y + segmentInset, W: at.W, H: at.H - 2*segmentInset}, k.skin.Radius.Control, k.skin.Colour.Surface.ControlLit)
				ink = k.skin.Colour.Ink.Text
			}
			sayStyled(k.gtx, k.th, ink, ButtonSaid(at), chip, reads, font.Medium, 1, text.Middle)
		default:
			drawButton(k, at, chip, buttonPlain)
		}
	}
	if w.Monthly() {
		drawButton(k, TimeChip(l, "the month"), w.week.Month().String(), lookOf(w.grid == viewMonths))
		drawButton(k, TimeChip(l, "the year"), strconv.Itoa(w.week.Year()), lookOf(w.grid == viewYears))
		if total := w.PeriodSays(); total != "" {
			at := PeriodAt(w, l)
			say(k.gtx, k.th, k.skin.Colour.Ink.Text, Box{X: at.X, Y: at.Y + (at.H-lineHigh)/2, W: at.W, H: lineHigh}, "· "+total, reads)
		}
	} else {
		say(k.gtx, k.th, k.skin.Colour.Ink.Text, Inside(TimeSaid(l)), TimeSays(w), reads)
	}
	if w.Monthly() {
		drawMonth(k, l, w)
		return
	}
	shown := w.TimeShown()
	longest := shown.Longest()
	if w.Daily() {
		drawDayAlone(k, l, w, longest)
		return
	}
	for i, day := range w.WeekDays() {
		box := WeekDayAt(l, i)
		rounded(k.gtx, box, k.skin.Radius.Ground, k.skin.Colour.Surface.Raised)
		head := WeekHeadAt(box)
		ink := k.skin.Colour.Ink.Faint
		if day.Equal(w.today()) {
			rounded(k.gtx, TodayPill(head, w.measured()), k.skin.Radius.Today, k.skin.Colour.Accent.Base)
			ink = k.skin.Colour.Surface.Back
		}
		say(k.gtx, k.th, ink, head, day.Format(todayFormat), reads)
		sayIn(k.gtx, k.th, k.skin.Colour.Ink.Low, WeekTotalAt(box, w.measured()), w.DayHeld(i).Says(), small, k.skin.Type.Mono, text.End)
		clipped := clip.Rect(rect(box)).Push(k.gtx.Ops)
		for j, one := range w.DayShown(i) {
			row := WeekRowAt(box, j)
			drawCard(k, w, DayCardAt(box, j), w.DayCard(one), w.Chosen() == one.Number && !w.chosenGoal, DayCardAt(box, j) == w.pointed)
			say(k.gtx, k.th, k.skin.Colour.Accent.Base, DaySpentAt(row), one.Spent(), small)
			fill(k.gtx, WeekBar(row, one.Minutes, longest), k.skin.Colour.Accent.Base)
		}
		clipped.Pop()
		if track, more := DayTrack(box, len(w.DayHeld(i).Tasks)); more {
			fill(k.gtx, track, k.skin.Colour.Surface.Back)
			fill(k.gtx, Thumb(track, len(w.DayHeld(i).Tasks), DayRoom(box), w.DayFrom(i)), k.skin.Colour.Line.Edge)
		}
	}
}

func drawDayAlone(k Kit, l Layout, w Window, longest int) {
	box := DayAloneAt(l)
	held := w.DayHeld(0)
	rounded(k.gtx, box, k.skin.Radius.Ground, k.skin.Colour.Surface.Raised)
	say(k.gtx, k.th, k.skin.Colour.Ink.Faint, WeekHeadAt(box), held.Says(), reads)
	clipped := clip.Rect(rect(box)).Push(k.gtx.Ops)
	for j, one := range w.AloneShown() {
		drawCard(k, w, AloneCardAt(box, j), w.DayCard(one), w.Chosen() == one.Number && !w.chosenGoal, AloneCardAt(box, j) == w.pointed)
		say(k.gtx, k.th, k.skin.Colour.Accent.Base, AloneSpentAt(box, j), one.Spent(), small)
		fill(k.gtx, AloneBar(box, j, one.Minutes, longest), k.skin.Colour.Accent.Base)
	}
	clipped.Pop()
	if track, more := DayTrack(box, len(held.Tasks)); more {
		fill(k.gtx, track, k.skin.Colour.Surface.Back)
		fill(k.gtx, Thumb(track, len(held.Tasks), DayRoom(box), w.DayFrom(0)), k.skin.Colour.Line.Edge)
	}
}

func drawMonth(k Kit, l Layout, w Window) {
	if cells := w.TimeCells(); len(cells) > 0 {
		for i, cell := range cells {
			drawButton(k, TimeCellAt(l, i), cell, lookOf(cell == gridMarked(w.grid, w.week)))
		}
		return
	}
	for i, name := range Weekdays[:w.pages.WeekDays()] {
		say(k.gtx, k.th, k.skin.Colour.Ink.Faint, MonthWeekdayAt(l, i), name, reads)
	}
	for i, day := range engine.MonthGrid(w.week) {
		cell := MonthCellAt(l, i)
		if cell == (Box{}) {
			continue
		}
		if day.Equal(w.today()) {
			framed(k, cell, k.skin.Radius.Control, k.skin.Colour.Accent.Base, k.skin.Colour.Surface.Raised)
		} else {
			rounded(k.gtx, cell, k.skin.Radius.Control, k.skin.Colour.Surface.Raised)
		}
		ink := k.skin.Colour.Ink.Text
		if !sameMonth(day, w.week) {
			ink = k.skin.Colour.Ink.Faint
		}
		held := w.MonthDay(i)
		say(k.gtx, k.th, ink, CellHeadAt(cell), CellHead(day, held), reads)
		clipped := clip.Rect(rect(cell)).Push(k.gtx.Ops)
		shown := cellShown(held.Tasks, CellRoom(cell))
		for j, line := range CellLines(held.Tasks, CellRoom(cell)) {
			at := CellLineAt(cell, j)
			if j < shown {
				c := w.DayCard(held.Tasks[j])
				fill(k.gtx, Box{X: at.X - inset, Y: at.Y, W: edgeWide, H: at.H}, k.skin.edgeOf(c))
				if w.Chosen() == c.Number && !w.chosenGoal {
					outline(k.gtx, at, k.skin.Colour.Accent.Base)
				}
				say(k.gtx, k.th, k.skin.Writing(c), Box{X: at.X + edgeWide, Y: at.Y, W: at.W - edgeWide, H: at.H}, line, small)
				continue
			}
			say(k.gtx, k.th, k.skin.Colour.Ink.Faint, at, line, small)
		}
		clipped.Pop()
	}
}

func drawMenu(k Kit, w Window) {
	box := w.MenuBox()
	raised(k, box, k.skin.Radius.Ground, k.skin.Colour.Surface.Raised)
	for i, said := range w.Menu() {
		row := MenuRowAt(box, i)
		ink := k.skin.Colour.Ink.Text
		if said == "Delete" {
			ink = k.skin.Colour.State.Wrong
		}
		if i == w.MenuOn() {
			rounded(k.gtx, row, k.skin.Radius.Row, k.skin.Colour.Surface.Control)
		}
		say(k.gtx, k.th, ink, Inside(row), said, reads)
		if key := MenuKey(said); key != "" {
			spelled := w.keyboard.Spelled(Pretty(key))
			drawKeyChip(k, Box{X: row.X + row.W - KeyChipWide(spelled) - inset*2, Y: row.Y, H: row.H}, spelled)
		}
	}
}

func drawCalendar(k Kit, w Window) {
	box := w.CalendarBox()
	drawPanelBare(k, box)
	for _, chip := range []struct {
		at   Box
		says string
		on   bool
	}{
		{CalendarMonth(box), w.day.Month().String(), w.view == viewMonths},
		{CalendarYear(box), strconv.Itoa(w.day.Year()), w.view == viewYears},
		{MonthBefore(box), "‹", false},
		{MonthAfter(box), "›", false},
	} {
		drawButton(k, chip.at, chip.says, lookOf(chip.on))
	}
	if cells := w.CalendarCells(); len(cells) > 0 {
		for i, cell := range cells {
			drawButton(k, CalendarCell(box, i), cell, lookOf(cell == w.CalendarMarked()))
		}
		drawPanelOffers(k, box, w)
		return
	}
	for i, name := range Weekdays {
		say(k.gtx, k.th, k.skin.Colour.Ink.Faint, WeekdayAt(box, i), name, reads)
	}
	today := w.today()
	for i, day := range engine.MonthGrid(w.day) {
		at, said := DayAt(box, i), strconv.Itoa(day.Day())
		switch {
		case day.Equal(w.day):
			drawButton(k, at, said, buttonOn)
		case day.Equal(today):
			framed(k, at, k.skin.Radius.Control, k.skin.Colour.Line.Selected, k.skin.Colour.Surface.Raised)
			sayCentred(k.gtx, k.th, k.skin.Colour.Ink.Text, ButtonSaid(at), said, reads)
		default:
			ink := k.skin.Colour.Ink.Text
			if w.Faded(day) {
				ink = k.skin.Colour.Ink.Faint
			}
			sayCentred(k.gtx, k.th, ink, ButtonSaid(at), said, reads)
		}
	}
	drawPanelOffers(k, box, w)
}

// TickedMark says an objective is in.
const TickedMark = "✓"

func drawChoosing(k Kit, w Window) {
	box := w.ChoosingBox()
	drawPanelBare(k, box)

	if !box.Unsearched {
		looking := ChoosingLooking(box)
		fill(k.gtx, looking, k.skin.Colour.Surface.Back)
		outline(k.gtx, looking, k.skin.Colour.Line.Edge)
		typed := Inside(looking)
		for _, taken := range w.LabelsTaken() {
			drawLabelPill(k, w, Box{X: taken.Pill.X, Y: taken.Pill.Y, W: taken.Pill.W - crossWide, H: taken.Pill.H}, taken.Label)
			ground, ink := k.skin.LabelPill(w.LabelColour(taken.Label))
			rounded(k.gtx, taken.Cross, k.skin.Radius.Full(taken.Cross), ground)
			sayCentred(k.gtx, k.th, ink, pillLine(k.th, taken.Cross), "×", small)
			typed.W -= taken.Pill.X + taken.Pill.W + pillGap - typed.X
			typed.X = taken.Pill.X + taken.Pill.W + pillGap
		}
		writing(k, w.measured(), typed, w.Filtering(), true, reads, 1, Follow)
	}

	at := ChoosingFrom(box)
	offered := w.Choosing()
	if len(offered) == 0 {
		say(k.gtx, k.th, k.skin.Colour.Ink.Faint, ChoosingLine(box, at), w.EmptyListSays(), reads)
		drawPanelOffers(k, box, w)
		return
	}
	run := ChoosingRun(box)
	clipped := clip.Rect(rect(run)).Push(k.gtx.Ops)
	from := w.ChoosingFromRow()
	for i := from; i < Lasting(len(offered), from, w.ChoosingRoom()); i++ {
		one := offered[i]
		row := ChoosingRow(box, at)
		if i == w.ChoosingAt() {
			rounded(k.gtx, row, k.skin.Radius.Row, k.skin.Colour.Surface.Control)
		}
		line := ChoosingLine(box, at)
		if one.Creates != "" {
			said := "Create"
			say(k.gtx, k.th, k.skin.Colour.Ink.Faint, line, said, reads)
			from := line.X + ends(w.measured(), said, reads) + pillGap*2
			ground := k.skin.Colour.Surface.Control
			if i == w.ChoosingAt() {
				ground = k.skin.Colour.Surface.Raised
			}
			drawPill(k, Box{X: from, Y: line.Y + (line.H-pillHigh)/2, W: ends(w.measured(), one.Creates, small) + 2*pillPad, H: pillHigh}, one.Creates, ground)
			at += apart
			continue
		}
		if w.choosingFor == "label" {
			if w.Ticked(one.Number) {
				say(k.gtx, k.th, k.skin.Colour.Accent.Base, line, TickedMark, reads)
			}
			drawLabelPill(k, w, LabelPillAt(line, ends(w.measured(), one.Title, small)), one.Title)
			sayIn(k.gtx, k.th, k.skin.Colour.Ink.Faint, line, strconv.Itoa(one.Carried), small, k.skin.Type.Mono, text.End)
			at += apart
			continue
		}
		say(k.gtx, k.th, k.skin.Colour.Ink.Text, line, w.RowSays(one), reads)
		if stripe, ok := w.RowStripe(one, line); ok {
			rounded(k.gtx, stripe, k.skin.Radius.Mark, k.skin.Colour.Priority[one.Title])
		}
		at += apart
	}
	clipped.Pop()

	track := ChoosingBar(box)
	fill(k.gtx, track, k.skin.Colour.Surface.Back)
	fill(k.gtx, Thumb(track, len(offered), w.ChoosingRoom(), from), k.skin.Colour.Line.Edge)
	drawPanelOffers(k, box, w)
}

func drawSaid(k Kit, w Window, said []Said, box Box) {
	drawPanel(k, box, Heads(w))
	at := FirstRow(box)
	for _, line := range said {
		row := Row(box, at, lineHigh)
		for _, mark := range w.CopyMarks() {
			if mark.Said == line.Means {
				row = Box{X: row.X + copyRoom, Y: row.Y, W: row.W - copyRoom, H: row.H}
			}
		}
		say(k.gtx, k.th, k.skin.Colour.Ink.Low, row, line.Means, reads)
		at += apart
	}
	drawPanelOffers(k, box, w)
}

func drawHelp(k Kit, w Window) {
	box := w.HelpBox()
	drawPanel(k, box, Heads(w))
	for _, line := range w.HelpLines() {
		if line.Head != "" {
			sayStyled(k.gtx, k.th, k.skin.Colour.Ink.Faint, line.At, line.Head, small, font.SemiBold, 1, text.Start)
			continue
		}
		rounded(k.gtx, line.KeyAt, k.skin.Radius.Chip, k.skin.Colour.Surface.Control)
		sayOnChip(k, k.skin.Colour.Ink.Text, line.KeyAt, line.Key)
		say(k.gtx, k.th, k.skin.Colour.Ink.Low, line.MeansAt, line.Means, reads)
	}
	drawPanelOffers(k, box, w)
}

func drawPanelBare(k Kit, box Box) {
	raised(k, box, k.skin.Radius.Panel, k.skin.Colour.Surface.Raised)
}

func drawPanel(k Kit, box Box, head string) {
	drawPanelHeaded(k, box, FormHeading(box), head)
}

func drawPanelHeaded(k Kit, box, heading Box, head string) {
	drawPanelBare(k, box)
	sayStyled(k.gtx, k.th, k.skin.Colour.Ink.Text, heading, head, HeadingSize, HeadingWeight, 1, text.Start)
}

func drawTabs(k Kit, l Layout, w Window) {
	for i, page := range w.Shown() {
		tab := TabAt(l, i)
		ink := k.skin.Colour.Ink.Faint
		if w.Tab() == page {
			rounded(k.gtx, tab, k.skin.Radius.Row, k.skin.Colour.Surface.Control)
			ink = k.skin.Colour.Ink.Text
		}
		room, name := TabName(tab), Titled(Pages[page])
		sayStyled(k.gtx, k.th, ink, room, name, reads, font.Medium, 1, text.Start)
		counted := ends(w.measured(), name, reads) + countGap
		sayIn(k.gtx, k.th, k.skin.Colour.Ink.Low, Box{X: room.X + counted, Y: room.Y, W: max(0, room.W-counted), H: room.H}, w.TabCount(page), small, k.skin.Type.Mono, text.Start)
	}
}

const countGap = 6

func drawColumns(k Kit, l Layout, w Window) {
	dealt, _ := engine.Deal(w.Showing().Cards, w.archived)
	for i, column := range dealt {
		ground(k, Box{X: l.Head[i].X, Y: l.Head[i].Y, W: l.Head[i].W, H: l.Head[i].H + l.Column[i].H})
		drawHead(k, NameIn(l.Head[i]), Titled(column.Name), len(column.Cards))
		if said := w.EmptySays(i); said != "" {
			drawEmpty(k, EmptyAt(l.CardAt(i, 0, 0)), said)
		}
	}
}

func drawBand(k Kit, l Layout, w Window) {
	for lane, kind := range engine.Stalled {
		ground(k, LaneGround(l, lane))
		held := len(w.inLane(lane))
		drawHead(k, LaneName(l, lane), Titled(kind), held)
		if held == 0 {
			drawEmpty(k, EmptyAt(l.LaneCardAt(lane, 0, 0)), nothingHere)
		}
	}
	drawPlaced(k, w)
}

func drawHead(k Kit, room Box, name string, held int) {
	sayStyled(k.gtx, k.th, k.skin.Colour.Ink.Faint, room, name, small, font.Medium, 1, text.Start)
	sayIn(k.gtx, k.th, k.skin.Colour.Ink.Low, room, strconv.Itoa(held), small, k.skin.Type.Mono, text.End)
}

func drawEmpty(k Kit, at Box, said string) {
	dashed(k, at, k.skin.Colour.Line.Edge)
	sayCentred(k.gtx, k.th, k.skin.Colour.Ink.Low, Box{X: at.X, Y: at.Y + (at.H-lineHigh)/2, W: at.W, H: lineHigh}, said, small)
}

func drawPlaced(k Kit, w Window) {
	progress, building := w.Intro()
	var seen []Placed
	for _, p := range w.Placed() {
		if p.seen() {
			seen = append(seen, p)
		}
	}
	for n, p := range seen {
		chosen := p.Card.Number == w.Chosen() || w.isTaken(p.Card.Number)
		clipped := clip.Rect(rect(p.Within)).Push(k.gtx.Ops)
		if !building {
			if Dimmed(p.Card) {
				half := paint.PushOpacity(k.gtx.Ops, 0.5)
				drawCard(k, w, p.Slot, p.Card, chosen, p.Slot == w.pointed)
				half.Pop()
			} else {
				drawCard(k, w, p.Slot, p.Card, chosen, p.Slot == w.pointed)
			}
			clipped.Pop()
			continue
		}
		_, fade := Arriving(progress, n, len(seen))
		faded := paint.PushOpacity(k.gtx.Ops, float32(fade))
		drawCard(k, w, p.Slot, p.Card, chosen, p.Slot == w.pointed)
		faded.Pop()
		clipped.Pop()
	}
	if !building {
		return
	}
	for n, p := range seen {
		fall, fade := Arriving(progress, n, len(seen))
		panel := PartsOf(p.Slot).Panel
		faded := paint.PushOpacity(k.gtx.Ops, float32(1-fade))
		for i, stone := range StonesOf(panel, p.Card.Number) {
			drawStone(k.gtx, Fallen(stone, Dropped(panel, fall)), k.skin.stoneOf(i))
		}
		faded.Pop()
	}
}

func drawStone(gtx layout.Context, stone []Point, shade Shade) {
	var path clip.Path
	path.Begin(gtx.Ops)
	top, bottom, left := stone[0].Y, stone[0].Y, stone[0].X
	for i, p := range stone {
		at := f32.Pt(float32(p.X), float32(p.Y))
		top, bottom, left = min(top, p.Y), max(bottom, p.Y), min(left, p.X)
		if i == 0 {
			path.MoveTo(at)
			continue
		}
		path.LineTo(at)
	}
	path.Close()
	outlined := clip.Outline{Path: path.End()}.Op().Push(gtx.Ops)
	paint.LinearGradientOp{
		Stop1: f32.Pt(0, float32(top)), Color1: shade.Light,
		Stop2: f32.Pt(0, float32(bottom)), Color2: shade.Dark,
	}.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)
	moved := op.Offset(image.Pt(left, top)).Push(gtx.Ops)
	grainOp().Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)
	moved.Pop()
	outlined.Pop()
}

var grainOp = sync.OnceValue(func() paint.ImageOp { return paint.NewImageOp(StoneGrain()) })

func drawCross(k Kit, cross Box, ink color.NRGBA) {
	reach := float32(cross.W) / 4
	x, y := float32(cross.X)+float32(cross.W)/2, float32(cross.Y)+float32(cross.H)/2
	var strokes clip.Path
	strokes.Begin(k.gtx.Ops)
	strokes.MoveTo(f32.Pt(x-reach, y-reach))
	strokes.LineTo(f32.Pt(x+reach, y+reach))
	strokes.MoveTo(f32.Pt(x+reach, y-reach))
	strokes.LineTo(f32.Pt(x-reach, y+reach))
	paint.FillShape(k.gtx.Ops, ink, clip.Stroke{Path: strokes.End(), Width: k.skin.Size.Stroke}.Op())
}

func drawGrip(k Kit, grip Box, ink color.NRGBA) {
	x, y := float32(grip.X+grip.W), float32(grip.Y+grip.H)
	var strokes clip.Path
	strokes.Begin(k.gtx.Ops)
	for _, reach := range []float32{float32(grip.W), float32(grip.W) / 2} {
		strokes.MoveTo(f32.Pt(x-reach, y))
		strokes.LineTo(f32.Pt(x, y-reach))
	}
	paint.FillShape(k.gtx.Ops, ink, clip.Stroke{Path: strokes.End(), Width: k.skin.Size.Stroke}.Op())
}

func drawChevron(k Kit, mark Box, ink color.NRGBA) {
	x, y := float32(mark.X), float32(mark.Y)+float32(mark.H)*0.3
	w, h := float32(mark.W), float32(mark.H)*0.4
	var strokes clip.Path
	strokes.Begin(k.gtx.Ops)
	strokes.MoveTo(f32.Pt(x+1, y))
	strokes.LineTo(f32.Pt(x+w/2, y+h))
	strokes.LineTo(f32.Pt(x+w-1, y))
	paint.FillShape(k.gtx.Ops, ink, clip.Stroke{Path: strokes.End(), Width: k.skin.Size.Stroke}.Op())
}

func drawCalendarMark(gtx layout.Context, mark Box, ink color.NRGBA) {
	outline(gtx, Box{X: mark.X, Y: mark.Y + 1, W: mark.W, H: mark.H - 1}, ink)
	fill(gtx, Box{X: mark.X, Y: mark.Y + 1, W: mark.W, H: 3}, ink)
}

func drawFocusMark(gtx layout.Context, mark Box, dot, halo color.NRGBA) {
	glow := GlowOf(mark)
	disc(gtx, glow.X, glow.Y, glow.Halo, halo)
	disc(gtx, glow.X, glow.Y, glow.Dot, dot)
}

func disc(gtx layout.Context, x, y, r float32, ink color.NRGBA) {
	round := clip.Ellipse{Min: image.Pt(int(x-r+0.5), int(y-r+0.5)), Max: image.Pt(int(x+r+0.5), int(y+r+0.5))}
	paint.FillShape(gtx.Ops, ink, round.Op(gtx.Ops))
}

func drawCard(k Kit, w Window, box Box, c engine.Card, chosen, hovered bool) {
	parts := PartsOf(box)
	ground := k.skin.Colour.Surface.Card
	switch {
	case c.InFocus:
		ground = k.skin.Colour.Surface.Tint
	case hovered:
		ground = k.skin.Colour.Surface.Hover
	}
	if chosen {
		framed(k, parts.Panel, k.skin.Radius.Card, k.skin.Colour.Line.Selected, ground)
	} else {
		rounded(k.gtx, parts.Panel, k.skin.Radius.Card, ground)
	}
	stripe := clip.RRect{Rect: rect(parts.Edge), NE: k.skin.Radius.Stripe, SE: k.skin.Radius.Stripe}
	paint.FillShape(k.gtx.Ops, k.skin.edgeOf(c), stripe.Op(k.gtx.Ops))

	m := Measured(k.th)
	wideAt := func(size int) func(string) int { return func(said string) int { return ends(m, said, size) } }
	first, rest := TitleRoom(c, parts.Title)
	lines := TitleLines(c.Title, first, rest, wideAt(reads))
	for i, line := range lines {
		sayStyled(k.gtx, k.th, k.skin.Writing(c), Box{X: parts.Title.X, Y: parts.Title.Y + i*lineHigh, W: parts.Title.W, H: lineHigh}, line, reads, font.Medium, 1, text.Start)
	}
	if mark, ok := FocusMark(c, parts.Title); ok {
		drawFocusMark(k.gtx, mark, k.skin.Colour.Accent.Base, k.skin.Colour.Accent.Glow)
	}

	mono := k.skin.Type.Mono
	laid := CardRowOf(c, parts.Number, m, k.gtx.Now, w.pages)
	sayIn(k.gtx, k.th, k.skin.Colour.Ink.Low, parts.Number, laid.Number, small, mono, text.Start)
	if laid.Facts != "" {
		ink := k.skin.Colour.Ink.Low
		if laid.Late {
			ink = k.skin.Colour.State.Late
		}
		sayIn(k.gtx, k.th, ink, laid.FactsAt, laid.Facts, small, mono, text.End)
	}
	for i, pill := range laid.Pills {
		drawLabelPill(k, w, pill, laid.Shown[i])
	}
	if laid.Count != "" {
		rounded(k.gtx, laid.More, k.skin.Radius.Full(laid.More), k.skin.Colour.Line.Edge)
		drawPill(k, Box{X: laid.More.X + 1, Y: laid.More.Y + 1, W: laid.More.W - 2, H: laid.More.H - 2}, laid.Count, ground)
	}

	if c.SubTotal > 0 {
		drawProgress(k, parts.Steps, c.SubDone, c.SubTotal, strconv.Itoa(c.SubDone)+"/"+strconv.Itoa(c.SubTotal))
	}
}

func drawProgress(k Kit, at Box, done, total int, said string) {
	counted := monoWide(said, small)
	track := Box{X: at.X, Y: at.Y + (at.H-k.skin.Size.Track)/2, W: max(0, at.W-counted-pillGap*2), H: k.skin.Size.Track}
	rounded(k.gtx, track, k.skin.Size.Track, k.skin.Colour.Surface.Control)
	if done > 0 {
		rounded(k.gtx, Box{X: track.X, Y: track.Y, W: max(k.skin.Size.Track, track.W*done/total), H: track.H}, k.skin.Size.Track, k.skin.Colour.State.Done)
	}
	sayIn(k.gtx, k.th, k.skin.Colour.Ink.Low, Box{X: at.X + at.W - counted, Y: at.Y + (at.H-lineHigh)/2, W: counted, H: lineHigh}, said, small, k.skin.Type.Mono, text.End)
}

func drawPill(k Kit, pill Box, said string, ground color.NRGBA) {
	rounded(k.gtx, pill, k.skin.Radius.Full(pill), ground)
	sayCentred(k.gtx, k.th, k.skin.Colour.Ink.Faint, pillLine(k.th, pill), said, small)
}

func pillLine(th *material.Theme, pill Box) Box {
	top, bottom := inkOf(th, font.Font{Typeface: th.Face}, pillReference)
	return Box{X: pill.X, Y: pill.Y + (pill.H-top-bottom)/2, W: pill.W, H: lineHigh}
}

func drawLabelPill(k Kit, w Window, pill Box, label string) {
	ground, ink := k.skin.LabelPill(w.LabelColour(label))
	rounded(k.gtx, pill, k.skin.Radius.Full(pill), ground)
	sayCentred(k.gtx, k.th, ink, pillLine(k.th, pill), label, small)
}

func drawGoals(k Kit, l Layout, w Window) {
	clipped := clip.Rect(rect(GoalsRun(l))).Push(k.gtx.Ops)
	defer clipped.Pop()
	for at, g := range engine.Sorted(w.Showing().Goals) {
		box := GoalAt(l, at, w.goals)
		face := goalFace(g)
		drawCard(k, w, box, face, g.Number == w.Chosen() || w.isTaken(g.Number), box == w.pointed)
		if inked, ok := k.skin.Colour.Standing[g.State]; ok {
			if pill := StandingPill(face, box, Measured(k.th)); pill != (Box{}) {
				rounded(k.gtx, pill, k.skin.Radius.Full(pill), inked.Back)
				sayCentred(k.gtx, k.th, inked.Ink, Box{X: pill.X, Y: pill.Y + (pill.H-lineHigh)/2, W: pill.W, H: lineHigh}, face.Labels[0], small)
			}
		}
		if g.Total > 0 {
			drawProgress(k, PartsOf(box).Steps, g.Done, g.Total, Counted(g))
		}
	}
}

func goalFace(g engine.Goal) engine.Card {
	return engine.Card{Number: g.Number, Title: g.Title, State: g.State, Priority: g.Priority, Labels: Standings(g),
		Deadline: g.Deadline}
}

func drawDragging(k Kit, w Window) {
	if landing, somewhere := w.Landing(); somewhere {
		outline(k.gtx, PartsOf(landing).Panel, k.skin.Colour.Accent.Base)
	}
	ghost, dragging := w.Ghost()
	if !dragging {
		return
	}
	raised(k, ghost, k.skin.Radius.Card, k.skin.Colour.Surface.Ghost)
	for _, c := range w.Reading().Cards {
		if c.Number != w.held.Number {
			continue
		}
		line := PartsOf(ghost).Title
		say(k.gtx, k.th, k.skin.Colour.Ink.Text, line, fits(k.gtx, k.th, c.Title, line.W, plain), plain)
	}
}

func drawBar(k Kit, l Layout, w Window) {
	if w.Wrong() != nil {
		fill(k.gtx, WrongBox(l), k.skin.Colour.Surface.Back)
		sayOver(k.gtx, k.th, k.skin.Colour.State.Wrong, WrongBox(l), w.Wrong().Error(), reads, 2)
	} else if w.Note() != "" {
		fill(k.gtx, WrongBox(l), k.skin.Colour.Surface.Back)
		sayOver(k.gtx, k.th, k.skin.Colour.Ink.Text, WrongBox(l), w.Note(), reads, 2)
	}
	bar := l.Bar
	framed(k, bar, k.skin.Radius.Ground, k.skin.Colour.Line.Rule, k.skin.Colour.Surface.Raised)
	offers, faint := w.Offers(), false
	if s := w.Screen(); s != OnTheBoard && s != OnSearch {
		offers, faint = w.BoardOffers(), true
	}
	along, worded := w.barLaid(offers)
	if w.Screen() == OnSearch {
		drawKeyChip(k, Box{X: bar.X + barInset, Y: bar.Y, H: bar.H}, "/")
		writing(k, w.measured(), SearchInto(bar), w.Searched(), true, reads, 1, Follow)
	} else {
		chip := drawKeyChip(k, Box{X: bar.X + barInset, Y: bar.Y, W: 0, H: bar.H}, w.keyboard.Spelled(Pretty("cmd+k")))
		end := bar.X + bar.W
		if len(along) > 0 {
			end = along[0].X
		}
		from := chip.X + chip.W + hintSpace
		sayIn(k.gtx, k.th, k.skin.Colour.Ink.Low, Box{X: from, Y: bar.Y + (bar.H-lineHigh)/2, W: max(0, end-hintMargin-from), H: lineHigh}, w.CommandSays(), reads, font.Font{}, text.Start)
	}
	for i, offer := range offers {
		if along[i].W == 0 {
			continue
		}
		word := offer
		if !worded {
			word = ""
		}
		drawHint(k, along[i], w.BarKey(offer), word, faint, !faint && w.Lighting(offer))
	}
}

func drawKeyChip(k Kit, row Box, key string) Box {
	chip := Box{X: row.X, Y: row.Y + (row.H-chipHigh)/2, W: KeyChipWide(key), H: chipHigh}
	if key == "" {
		return Box{X: row.X, Y: chip.Y, H: chipHigh}
	}
	rounded(k.gtx, chip, k.skin.Radius.Chip, k.skin.Colour.Surface.Control)
	sayOnChip(k, k.skin.Colour.Ink.Text, chip, key)
	return chip
}

const chipHigh = 18

func sayOnChip(k Kit, ink color.NRGBA, chip Box, key string) {
	top, bottom := inkOf(k.th, k.skin.Type.MonoBold, key)
	at := chip.Y + chip.H/2 - (top+bottom)/2
	sayIn(k.gtx, k.th, ink, Box{X: chip.X, Y: at, W: chip.W, H: lineHigh}, key, small, k.skin.Type.MonoBold, text.Middle)
}

// pillReference is what a pill's name is centred by: letters that rise and stand on the line, so every pill's name sits on the same line whatever it hangs below it.
const pillReference = "Hl"

func inkOf(th *material.Theme, face font.Font, key string) (top, bottom int) {
	th.Shaper.LayoutString(text.Parameters{Font: face, PxPerEm: fixed.I(small), MaxLines: 1, MaxWidth: math.MaxInt32}, key)
	top, bottom = lineHigh, 0
	for g, ok := th.Shaper.NextGlyph(); ok; g, ok = th.Shaper.NextGlyph() {
		top = min(top, int(g.Y)+g.Bounds.Min.Y.Floor())
		bottom = max(bottom, int(g.Y)+g.Bounds.Max.Y.Ceil())
	}
	if top > bottom {
		return 0, lineHigh
	}
	return top, bottom
}

func drawHint(k Kit, at Box, key, word string, faint, lit bool) {
	if lit {
		rounded(k.gtx, Box{X: at.X, Y: at.Y + 3, W: at.W - hintSpace, H: at.H - 6}, k.skin.Radius.Row, k.skin.Colour.Surface.Hover)
	}
	chip := drawKeyChip(k, Box{X: at.X + hintSpace/2, Y: at.Y, H: at.H}, key)
	ink := k.skin.Colour.Ink.Text
	if faint {
		ink = k.skin.Colour.Ink.Faint
	}
	from := chip.X + chip.W + hintSpace
	sayIn(k.gtx, k.th, ink, Box{X: from, Y: at.Y + (at.H-lineHigh)/2, W: at.X + at.W - from, H: lineHigh}, word, reads, font.Font{}, text.Start)
}

type buttonLook int

func litLook(look buttonLook) buttonLook {
	switch look {
	case buttonOn:
		return buttonOnLit
	case buttonDanger:
		return buttonDangerLit
	}
	return buttonLit
}

func lookOf(on bool) buttonLook {
	if on {
		return buttonOn
	}
	return buttonPlain
}

const (
	buttonPlain buttonLook = iota
	buttonOn
	buttonLit
	buttonFaint
	buttonDanger
	buttonOnLit
	buttonDangerLit
)

func drawButton(k Kit, button Box, said string, look buttonLook) {
	back, ink, _ := buttonInks(k.skin, look)
	rounded(k.gtx, button, k.skin.Radius.Control, back)
	sayStyled(k.gtx, k.th, ink, ButtonSaid(button), said, reads, font.Medium, 1, text.Middle)
}

func buttonInks(skin Skin, look buttonLook) (back, ink, chip color.NRGBA) {
	c := skin.Colour
	switch look {
	case buttonOn:
		return c.Accent.Base, c.Ink.OnAccent, c.Accent.Chip
	case buttonOnLit:
		return c.Accent.Lit, c.Ink.OnAccent, c.Accent.LitChip
	case buttonLit:
		return c.Surface.ControlLit, c.Ink.Text, c.Surface.ChipOnControlLit
	case buttonFaint:
		return c.Surface.Control, c.Ink.Faint, c.Surface.ChipOnControl
	case buttonDanger:
		return c.State.Danger, c.State.Wrong, c.State.DangerChip
	case buttonDangerLit:
		return c.State.DangerLit, c.State.Wrong, c.State.DangerLitChip
	}
	return c.Surface.Control, c.Ink.Text, c.Surface.ChipOnControl
}

func drawKeyedButton(k Kit, button Box, m Measure, word, key string, look buttonLook) {
	back, ink, under := buttonInks(k.skin, look)
	rounded(k.gtx, button, k.skin.Radius.Control, back)
	said, chip := ButtonParts(button, m, word, key)
	sayStyled(k.gtx, k.th, ink, said, word, reads, font.Medium, 1, text.Start)
	if key == "" {
		return
	}
	rounded(k.gtx, chip, k.skin.Radius.Chip, under)
	sayOnChip(k, ink, chip, key)
}

func drawSub(k Kit, box Box, c engine.Card, line string, on, picked bool, held Written) {
	panel := SubPanel(box)
	ground := k.skin.Colour.Surface.Raised
	if picked {
		ground = k.skin.Colour.Surface.Ghost
	}
	fill(k.gtx, panel, ground)
	if on {
		outline(k.gtx, panel, k.skin.Colour.Accent.Base)
	}
	check := SubCheck(box)
	done := engine.Finished(c)
	if done {
		fill(k.gtx, check, k.skin.Colour.Ink.Faint)
		say(k.gtx, k.th, k.skin.Colour.Surface.Raised, TickAt(check), TickedMark, small)
	} else {
		outline(k.gtx, check, k.skin.Colour.Ink.Text)
	}
	writ := SubWriting(box)
	said := fits(k.gtx, k.th, line, writ.W, small)
	held.Value = said
	sayHolding(k, SubInk(k.skin, c), writ, held, small)
	if done {
		loose := k.gtx
		loose.Constraints = layout.Constraints{Max: image.Pt(writ.W, lineHigh)}
		fill(k.gtx, StruckAt(writ, wide(loose, k.th, said, small).Size.X), k.skin.Colour.Ink.Faint)
	}
}

// SubInk is the ink a step's name, or a serving task's, is written in: the reading ink for what is left, the faint one for what is done.
func SubInk(skin Skin, c engine.Card) color.NRGBA {
	if engine.Finished(c) {
		return skin.Colour.Ink.Faint
	}
	return skin.Colour.Ink.Text
}

func drawBesideTabs(k Kit, panel Box, w Window) {
	tabs, says := BesideTabs(panel), w.BesideTabsSay()
	rounded(k.gtx, Box{X: tabs[0].X, Y: tabs[0].Y, W: tabs[1].X + tabs[1].W - tabs[0].X, H: tabs[0].H}, k.skin.Radius.Control, k.skin.Colour.Surface.Control)
	for i, tab := range tabs {
		ink := k.skin.Colour.Ink.Faint
		if (i == 1) == w.Tracing() {
			rounded(k.gtx, Box{X: tab.X + segmentInset, Y: tab.Y + segmentInset, W: tab.W - 2*segmentInset, H: tab.H - 2*segmentInset}, k.skin.Radius.Control, k.skin.Colour.Surface.ControlLit)
			ink = k.skin.Colour.Ink.Text
		}
		sayStyled(k.gtx, k.th, ink, ButtonSaid(tab), says[i], reads, font.Medium, 1, text.Middle)
	}
}

func drawUnder(k Kit, w Window) {
	under := w.Under()
	done := 0
	for _, c := range under {
		if engine.Finished(c) {
			done++
		}
	}
	name := w.WorkName()
	say(k.gtx, k.th, k.skin.Colour.Ink.Faint, name, ServingHead, small)
	sayIn(k.gtx, k.th, k.skin.Colour.Ink.Faint, name, fmt.Sprintf("%s · %d done", countOf(len(under), "task"), done), small, k.skin.Type.Mono, text.End)
	frame := w.workFrame()
	fill(k.gtx, frame, k.skin.Colour.Surface.Field)
	clipped := clip.Rect(rect(frame)).Push(k.gtx.Ops)
	on, walking := w.WorkOn()
	for _, tile := range w.WorkTiles() {
		drawWorkTile(k, w, tile)
		if walking && tile == on {
			panel := SubPanel(tile.At)
			outline(k.gtx, Box{X: panel.X + inset, Y: panel.Y + 1, W: panel.W - 2*inset, H: panel.H - 2}, k.skin.Colour.Accent.Base)
		}
	}
	clipped.Pop()
	track := w.workBar()
	fill(k.gtx, track, k.skin.Colour.Surface.Back)
	fill(k.gtx, Thumb(track, (len(under)+1)*SubHeight, frame.H, w.ServingAt()), k.skin.Colour.Line.Edge)
}

func drawWorkTile(k Kit, w Window, tile WorkTile) {
	line := SubWriting(tile.At)
	if tile.Number == 0 {
		say(k.gtx, k.th, k.skin.Colour.Ink.Faint, Box{X: tile.At.X + padding, Y: line.Y, W: tile.At.W - 2*padding, H: line.H}, tile.Said, small)
		return
	}
	panel := SubPanel(tile.At)
	rounded(k.gtx, Box{X: panel.X + inset, Y: panel.Y + 1, W: panel.W - 2*inset, H: panel.H - 2}, k.skin.Radius.Row, k.skin.Colour.Surface.Raised)
	c, _ := w.snap.Card(tile.Number)
	x := tile.At.X + padding
	number := "#" + strconv.Itoa(c.Number)
	sayIn(k.gtx, k.th, k.skin.Colour.Ink.Faint, Box{X: x, Y: line.Y, W: monoWide(number, small), H: line.H}, number, small, k.skin.Type.Mono, text.Start)
	x += monoWide(number, small) + padding
	spent := c.Spent
	if spent == "" {
		spent = "—"
	}
	spentAt := Box{X: tile.At.X + tile.At.W - padding - monoWide("00h00m", small), Y: line.Y, W: monoWide("00h00m", small), H: line.H}
	sayIn(k.gtx, k.th, k.skin.Colour.Ink.Faint, spentAt, spent, small, k.skin.Type.Mono, text.End)
	state := c.State
	pill := Box{X: spentAt.X - padding - ends(w.measured(), state, small) - 2*pillPad, Y: line.Y + (line.H-pillHigh)/2, W: ends(w.measured(), state, small) + 2*pillPad, H: pillHigh}
	drawPill(k, pill, state, k.skin.Colour.Surface.Control)
	ink := k.skin.Colour.Ink.Text
	if engine.Finished(c) {
		ink = k.skin.Colour.Ink.Faint
	}
	title := Box{X: x, Y: line.Y, W: max(0, pill.X-padding-x), H: line.H}
	said := fits(k.gtx, k.th, c.Title, title.W, small)
	say(k.gtx, k.th, ink, title, said, small)
	if engine.Finished(c) {
		loose := k.gtx
		loose.Constraints = layout.Constraints{Max: image.Pt(title.W, lineHigh)}
		fill(k.gtx, StruckAt(title, wide(loose, k.th, said, small).Size.X), k.skin.Colour.Ink.Faint)
	}
}

func drawProgressBeside(k Kit, w Window) {
	g, ok := w.snap.Goal(w.form.Number)
	if !ok {
		return
	}
	at := w.ProgressBox()
	say(k.gtx, k.th, k.skin.Colour.Ink.Faint, Box{X: at.X, Y: at.Y, W: at.W, H: lineHigh}, ProgressHead, small)
	card := Box{X: at.X, Y: at.Y + lineHigh + between, W: at.W, H: at.H - lineHigh - between}
	rounded(k.gtx, card, k.skin.Radius.Ground, k.skin.Colour.Surface.Card)
	inside := Box{X: card.X + padding, Y: card.Y + padding/2, W: card.W - 2*padding, H: card.H - padding}
	sayStyled(k.gtx, k.th, k.skin.Colour.Ink.Text, Box{X: inside.X, Y: inside.Y, W: inside.W, H: titleHi + 4}, strconv.Itoa(g.Percent())+"%", k.skin.Type.Figure, font.SemiBold, 1, text.Start)
	sayIn(k.gtx, k.th, k.skin.Colour.Ink.Low, Box{X: inside.X, Y: inside.Y + 4, W: inside.W, H: lineHigh}, fmt.Sprintf("%d of %d tasks done", g.Done, g.Total), reads, font.Font{}, text.End)
	track := Box{X: inside.X, Y: inside.Y + titleHi + 12, W: inside.W, H: k.skin.Size.Track + 1}
	rounded(k.gtx, track, k.skin.Size.Track, k.skin.Colour.Surface.Control)
	if g.Total > 0 && g.Done > 0 {
		rounded(k.gtx, Box{X: track.X, Y: track.Y, W: max(k.skin.Size.Track, track.W*g.Done/g.Total), H: track.H}, k.skin.Size.Track, k.skin.Colour.State.Done)
	}
	facts := ProgressFacts(g, w.now)
	third := inside.W / len(facts)
	for i, fact := range facts {
		cell := Box{X: inside.X + i*third, Y: track.Y + track.H + between, W: third, H: lineHigh}
		sayIn(k.gtx, k.th, k.skin.Colour.Ink.Text, cell, fact[0], small, k.skin.Type.MonoBold, text.Start)
		say(k.gtx, k.th, k.skin.Colour.Ink.Faint, Box{X: cell.X, Y: cell.Y + lineHigh, W: cell.W, H: lineHigh}, fact[1], small)
	}
}

func drawHistory(k Kit, box Box, w Window) {
	frame := BesideBox(box)
	fill(k.gtx, frame, k.skin.Colour.Surface.Back)
	outline(k.gtx, frame, k.skin.Colour.Line.Edge)

	inside := Inside(frame)
	clipped := clip.Rect(rect(inside)).Push(k.gtx.Ops)
	told, from, rows := w.traced(), w.History(), w.Rows()
	for i := from; i < Lasting(len(told.Said), from, rows); i++ {
		drawTold(k, LineAt(inside, i, from), w.HoldingFrom("trace", told.Said[i], told.From[i]))
	}
	clipped.Pop()

	track := BesideBar(box)
	fill(k.gtx, track, k.skin.Colour.Surface.Back)
	fill(k.gtx, Thumb(track, len(told.Said), rows, w.History()), k.skin.Colour.Line.Edge)
}

func drawTold(k Kit, box Box, held Written) {
	said := held.Value
	paintHeld(k, box, held, small)
	stamp, rest, ok := Stamped(said)
	if !ok {
		say(k.gtx, k.th, k.skin.Colour.Ink.Text, box, said, small)
		return
	}
	say(k.gtx, k.th, k.skin.Colour.Ink.Faint, box, stamp, small)
	say(k.gtx, k.th, k.skin.Colour.Ink.Text, PastStamp(box, stamp), rest, small)
}

func sayHolding(k Kit, ink color.NRGBA, box Box, said Written, size int) {
	paintHeld(k, box, said, size)
	say(k.gtx, k.th, ink, box, said.Value, size)
}

func paintHeld(k Kit, box Box, said Written, size int) {
	if !said.Holding() {
		return
	}
	kept := []rune(said.Value)
	from, to := said.Run()
	from, to = min(from, len(kept)), min(to, len(kept))
	loose := k.gtx
	loose.Constraints = layout.Constraints{Max: image.Pt(box.W, box.H)}
	fill(k.gtx, HeldRun(box, wide(loose, k.th, string(kept[:from]), size).Size.X,
		wide(loose, k.th, string(kept[from:to]), size).Size.X), k.skin.Colour.Surface.Ghost)
}

func drawTicked(k Kit, name, frame Box, w Window) {
	f := w.Form()
	sayHolding(k, k.skin.Colour.Ink.Faint, name, Written{Value: "steps"}, small)
	fill(k.gtx, frame, k.skin.Colour.Surface.Back)
	outline(k.gtx, frame, k.skin.Colour.Line.Edge)
	parts, from := w.PartsTicked(), w.partsFrom()
	on := f.onParts() && w.screen == OnForm
	places := w.partTiles()
	for n, tile := range places[:len(places)-1] {
		at := from + n
		if on && f.writingPart && f.partAt == at {
			fill(k.gtx, SubPanel(tile), k.skin.Colour.Surface.Raised)
			outline(k.gtx, SubPanel(tile), k.skin.Colour.Accent.Base)
			writing(k, w.measured(), SubWriting(tile), f.named, true, small, 1, Follow)
			continue
		}
		drawSub(k, tile, parts[at], StepLine(parts[at]), on && f.partAt == at, on && slices.Contains(f.selectedParts(), at), Written{})
	}
	if field, ok := f.partsField(); ok {
		_, value, rows := FormPlace(w.formBox(), f.Fields(), field)
		_, bar := Saying(value, rows)
		fill(k.gtx, bar, k.skin.Colour.Surface.Back)
		fill(k.gtx, Thumb(bar, len(parts), w.partsShown(), from), k.skin.Colour.Line.Edge)
	}
	adding := places[len(places)-1]
	onAdding := on && f.partAt == f.lines()-1
	if onAdding && f.writingPart {
		writing(k, w.measured(), adding, f.named, true, small, 1, Follow)
	} else {
		ink := k.skin.Colour.Ink.Faint
		if onAdding {
			ink = k.skin.Colour.Accent.Base
		}
		say(k.gtx, k.th, ink, adding, AddingAPart, small)
	}
	if ghost, ok := w.PartGhost(); ok {
		fill(k.gtx, ghost, k.skin.Colour.Surface.Ghost)
		outline(k.gtx, ghost, k.skin.Colour.Accent.Base)
		line := SubWriting(ghost)
		say(k.gtx, k.th, k.skin.Colour.Ink.Text, line, fits(k.gtx, k.th, StepLine(parts[f.carried-1]), line.W, small), small)
	}
	if line, ok := w.PartLanding(); ok {
		fill(k.gtx, line, k.skin.Colour.Accent.Base)
	}
}

func drawFieldAt(k Kit, m Measure, name, value Box, rows, size int, long, lit, carets bool, said, named Written, scrolled int) {
	frame := Framed(value, rows)
	var bar Box
	if long {
		frame, bar = Saying(value, rows)
	}
	ink, edge := k.skin.Colour.Ink.Low, k.skin.Colour.Line.Rule
	if lit {
		ink, edge = k.skin.Colour.Accent.Base, k.skin.Colour.Accent.Base
		halo := k.skin.Space.Halo
		rounded(k.gtx, Box{X: frame.X - halo, Y: frame.Y - halo, W: frame.W + 2*halo, H: frame.H + 2*halo}, k.skin.Radius.Control+halo, k.skin.Colour.Accent.Halo)
	}
	if name.W > 0 {
		sayHolding(k, ink, name, named, small)
	}
	framed(k, frame, k.skin.Radius.Control, edge, k.skin.Colour.Surface.Field)
	writing(k, m, Inside(frame), said, carets, size, rows, scrolled)
	if long {
		lines := wrapIn(m, said.Reads(), Inside(frame), size)
		line, _ := lines.At(said.At)
		fill(k.gtx, bar, k.skin.Colour.Surface.Back)
		fill(k.gtx, Thumb(bar, len(lines.Said), rows, StartsAt(lines, rows, line, scrolled)), k.skin.Colour.Line.Edge)
	}
}

func drawForm(k Kit, w Window) {
	panel, box := w.formPanel(), w.formBox()
	drawPanelHeaded(k, panel, FormHeading(panel), HeadsOn(w, OnForm))
	drawCross(k, FormClose(panel), k.skin.Colour.Ink.Faint)
	fields := w.Form().Fields()
	for i := range fields {
		if w.hidden(fields[i]) {
			continue
		}
		on := i == w.Form().On() && !w.form.onWork
		typed := on && !chosenFromAList(fields[i])
		said := Written{Value: w.FieldSays(fields[i])}
		if typed {
			said = w.Form().Writing()
		}
		scrolled := w.scrolledOf(i)
		name, value, rows := FormPlace(box, fields, i)
		if fields[i].Name == "steps" {
			drawTicked(k, name, Framed(value, rows), w)
			continue
		}
		drawFieldAt(k, w.measured(), name, value, rows, FormSize(fields[i]), fields[i].Long, on, typed, said, Written{Value: fields[i].Name}, scrolled)
		frame := Framed(value, rows)
		switch fields[i].Arg {
		case "priority":
			if stripe, known := k.skin.Colour.Priority[said.Reads()]; known {
				rounded(k.gtx, Box{X: FieldMark(frame).X - padding/2 - edgeWide, Y: frame.Y + (frame.H-stripeHigh)/2, W: edgeWide, H: stripeHigh}, k.skin.Radius.Mark, stripe)
			}
		case "label":
			if !on {
				drawFieldPills(k, w, w.measured(), frame, strings.FieldsFunc(said.Reads(), func(r rune) bool { return r == ',' || r == ' ' }))
			}
		}
		switch {
		case on:
		case chosenOnACalendar(fields[i]):
			drawCalendarMark(k.gtx, FieldMark(frame), k.skin.Colour.Ink.Low)
		case chosenFromAList(fields[i]):
			drawChevron(k, FieldMark(frame), k.skin.Colour.Ink.Low)
		}
	}
	if !w.Tracing() {
		for _, head := range FormHeads(box, fields) {
			say(k.gtx, k.th, k.skin.Colour.Ink.Faint, head.Box, head.Said, head.Size)
		}
	}
	drawGrip(k, FormGrip(panel), k.skin.Colour.Ink.Faint)
	if w.screen == OnForm {
		drawPanelOffers(k, panel, w)
		if note := w.FormNote(); note != "" && len(w.ContextSays()) == 0 {
			say(k.gtx, k.th, k.skin.Colour.Ink.Low, ContextName(panel), note, small)
		}
	}
	if tip, said, ok := w.Explanation(); ok {
		raised(k, tip, k.skin.Radius.Card, k.skin.Colour.Surface.Raised)
		for i, line := range said {
			ink := k.skin.Colour.Ink.Text
			if strings.HasPrefix(line, ExampleSays) && i == len(said)-1 && w.ExplainsBy() != "" {
				ink = k.skin.Colour.Ink.Faint
			}
			say(k.gtx, k.th, ink, ExplainedLine(tip, i), line, small)
		}
	}
	switch {
	case w.TabsBeside():
		drawBesideTabs(k, panel, w)
	case w.besideShown():
		say(k.gtx, k.th, k.skin.Colour.Ink.Faint, BesideHead(panel), DetailsHead, small)
		drawUnder(k, w)
		drawProgressBeside(k, w)
	}
	if w.Tracing() {
		drawHistory(k, panel, w)
	}
	for _, r := range w.ContextSays() {
		sayHolding(k, k.skin.Colour.Ink.Faint, r.Box, w.Holding(r.Key, r.Said), r.Size)
	}
}

func drawPanelOffers(k Kit, box Box, w Window) {
	along := w.offersAlong(box)
	if len(along) > 0 {
		fill(k.gtx, Box{X: box.X, Y: along[0].Y - between, W: box.W, H: 1}, k.skin.Colour.Line.Raised)
	}
	for i, offer := range w.Offers() {
		look := w.OfferLook(i, offer)
		if w.Lighting(offer) {
			look = litLook(look)
		}
		drawKeyedButton(k, Button(along[i]), Measured(k.th), offer, w.OfferKey(offer), look)
	}
	if cross, ok := w.PanelClose(); ok && w.screen != OnForm {
		drawCross(k, cross, k.skin.Colour.Ink.Faint)
	}
}

func inDp(gtx layout.Context) op.TransformStack {
	return op.Affine(f32.Affine2D{}.Scale(f32.Pt(0, 0), f32.Pt(gtx.Metric.PxPerDp, gtx.Metric.PxPerDp))).Push(gtx.Ops)
}

func rect(b Box) image.Rectangle {
	return image.Rect(b.X, b.Y, b.X+b.W, b.Y+b.H)
}

func fill(gtx layout.Context, b Box, with color.NRGBA) {
	paint.FillShape(gtx.Ops, with, clip.Rect(rect(b)).Op())
}

func ground(k Kit, b Box) { rounded(k.gtx, b, k.skin.Radius.Ground, k.skin.Colour.Surface.Ground) }

func dashed(k Kit, b Box, with color.NRGBA) {
	gtx, dash, dashGap := k.gtx, k.skin.Size.Dash, k.skin.Space.DashGap
	for x := b.X; x < b.X+b.W; x += dash + dashGap {
		long := min(dash, b.X+b.W-x)
		fill(gtx, Box{X: x, Y: b.Y, W: long, H: 1}, with)
		fill(gtx, Box{X: x, Y: b.Y + b.H - 1, W: long, H: 1}, with)
	}
	for y := b.Y; y < b.Y+b.H; y += dash + dashGap {
		long := min(dash, b.Y+b.H-y)
		fill(gtx, Box{X: b.X, Y: y, W: 1, H: long}, with)
		fill(gtx, Box{X: b.X + b.W - 1, Y: y, W: 1, H: long}, with)
	}
}

func dashedPill(k Kit, b Box, with color.NRGBA) {
	gtx, dash, dashGap := k.gtx, k.skin.Size.Dash, k.skin.Space.DashGap
	r := k.skin.Radius.Full(b)
	for _, end := range []Box{{X: b.X, Y: b.Y, W: r, H: b.H}, {X: b.X + b.W - r, Y: b.Y, W: r, H: b.H}} {
		cut := clip.Rect(rect(end)).Push(gtx.Ops)
		paint.FillShape(gtx.Ops, with, clip.Stroke{Path: clip.UniformRRect(rect(Box{X: b.X, Y: b.Y, W: b.W, H: b.H}), r).Path(gtx.Ops), Width: float32(k.skin.Size.Rule)}.Op())
		cut.Pop()
	}
	for x := b.X + r + dashGap; x < b.X+b.W-r; x += dash + dashGap {
		long := min(dash, b.X+b.W-r-x)
		fill(gtx, Box{X: x, Y: b.Y, W: long, H: 1}, with)
		fill(gtx, Box{X: x, Y: b.Y + b.H - 1, W: long, H: 1}, with)
	}
}

func rounded(gtx layout.Context, b Box, radius int, with color.NRGBA) {
	paint.FillShape(gtx.Ops, with, clip.UniformRRect(rect(b), radius).Op(gtx.Ops))
}

func raised(k Kit, b Box, radius int, with color.NRGBA) {
	for _, s := range k.skin.Shadow {
		under := Box{X: b.X - s.Spread, Y: b.Y + s.Down - s.Spread, W: b.W + 2*s.Spread, H: b.H + 2*s.Spread}
		rounded(k.gtx, under, radius+s.Spread, s.Ink)
	}
	rounded(k.gtx, b, radius, with)
}

func framed(k Kit, b Box, radius int, edge, inside color.NRGBA) {
	rounded(k.gtx, b, radius, edge)
	lining(k, b, radius, inside)
}

func lining(k Kit, b Box, radius int, with color.NRGBA) {
	r := k.skin.Size.Rule
	rounded(k.gtx, Box{X: b.X + r, Y: b.Y + r, W: b.W - 2*r, H: b.H - 2*r}, radius-r, with)
}

func drawCopyMark(k Kit, at Box) {
	back := Box{X: at.X + 3, Y: at.Y, W: at.W - 3, H: at.H - 3}
	front := Box{X: at.X, Y: at.Y + 3, W: at.W - 3, H: at.H - 3}
	outline(k.gtx, back, k.skin.Colour.Ink.Low)
	fill(k.gtx, front, k.skin.Colour.Surface.Raised)
	outline(k.gtx, front, k.skin.Colour.Ink.Low)
}

func outline(gtx layout.Context, b Box, with color.NRGBA) {
	for _, edge := range Edges(b) {
		fill(gtx, edge, with)
	}
}

func writing(k Kit, m Measure, b Box, said Written, carets bool, size, rows, scrolled int) {
	clipped := clip.Rect(rect(b)).Push(k.gtx.Ops)
	defer clipped.Pop()

	at := said.At
	lines := wrapIn(m, said.Reads(), b, size)
	line, column := lines.At(at)
	from := StartsAt(lines, rows, line, scrolled)

	begins, ends := said.Run()
	loose := k.gtx
	loose.Constraints = layout.Constraints{Max: image.Pt(b.W, lineHigh)}
	for i := from; i < Lasting(len(lines.Said), from, rows); i++ {
		if a, z, on := HeldOn(lines, begins, ends, i); said.Holding() && on {
			kept := []rune(lines.Said[i])
			at := LineAt(b, i, from)
			left := wide(loose, k.th, string(kept[:a]), size).Size.X
			fill(k.gtx, Box{X: at.X + left, Y: at.Y, H: at.H,
				W: wide(loose, k.th, string(kept[a:z]), size).Size.X}, k.skin.Colour.Surface.Ghost)
		}
		say(k.gtx, k.th, k.skin.Colour.Ink.Text, LineAt(b, i, from), lines.Said[i], size)
	}
	if !carets {
		return
	}
	kept := []rune(lines.Said[line])
	before := wide(loose, k.th, string(kept[:min(len(kept), column)]), size)

	fill(k.gtx, CaretAt(b, line, from, before.Size.X), k.skin.Colour.Accent.Base)
}

func fits(gtx layout.Context, th *material.Theme, said string, width, size int) string {
	loose := gtx
	loose.Constraints = layout.Constraints{Max: image.Pt(1<<20, lineHigh)}
	if wide(loose, th, said, size).Size.X <= width {
		return said
	}
	best := "…"
	kept := []rune(said)
	for n := range kept {
		cut := strings.TrimRight(string(kept[:n+1]), " ") + "…"
		if wide(loose, th, cut, size).Size.X > width {
			break
		}
		best = cut
	}
	return best
}

func wide(gtx layout.Context, th *material.Theme, said string, size int) layout.Dimensions {
	recorded := op.Record(gtx.Ops)
	label := material.Label(th, unit.Sp(float32(size)), said)
	label.MaxLines = 1
	dims := label.Layout(gtx)
	recorded.Stop()
	return dims
}

func say(gtx layout.Context, th *material.Theme, ink color.NRGBA, b Box, said string, size int) {
	sayOver(gtx, th, ink, b, said, size, 1)
}

func sayCentred(gtx layout.Context, th *material.Theme, ink color.NRGBA, b Box, said string, size int) {
	sayAligned(gtx, th, ink, b, said, size, 1, text.Middle)
}

func sayOver(gtx layout.Context, th *material.Theme, ink color.NRGBA, b Box, said string, size, lines int) {
	sayAligned(gtx, th, ink, b, said, size, lines, text.Start)
}

func sayAligned(gtx layout.Context, th *material.Theme, ink color.NRGBA, b Box, said string, size, lines int, along text.Alignment) {
	sayStyled(gtx, th, ink, b, said, size, font.Normal, lines, along)
}

// textSlack is how much taller than its box a line may be laid out, the box still clipping it: Figtree stands a point taller than the box a line of Go was given.
const textSlack = 4

func sayStyled(gtx layout.Context, th *material.Theme, ink color.NRGBA, b Box, said string, size int, weight font.Weight, lines int, along text.Alignment) {
	sayWith(gtx, th, ink, b, said, size, font.Font{Weight: weight}, lines, along)
}

func sayIn(gtx layout.Context, th *material.Theme, ink color.NRGBA, b Box, said string, size int, face font.Font, along text.Alignment) {
	sayWith(gtx, th, ink, b, said, size, face, 1, along)
}

var noteSaid func(gtx layout.Context, th *material.Theme, b Box, said string, size int, face font.Font, lines int, along text.Alignment)

func sayWith(gtx layout.Context, th *material.Theme, ink color.NRGBA, b Box, said string, size int, face font.Font, lines int, along text.Alignment) {
	if said == "" {
		return
	}
	if noteSaid != nil {
		noteSaid(gtx, th, b, said, size, face, lines, along)
	}
	clipped := clip.Rect(rect(b)).Push(gtx.Ops)
	defer clipped.Pop()
	offset := op.Offset(image.Pt(b.X, b.Y)).Push(gtx.Ops)
	defer offset.Pop()

	inner := gtx
	inner.Constraints = layout.Constraints{Min: image.Pt(b.W, 0), Max: image.Pt(b.W, b.H+textSlack)}
	label := material.Label(th, unit.Sp(float32(size)), said)
	label.Color = ink
	label.Font.Weight = face.Weight
	if face.Typeface != "" {
		label.Font.Typeface = face.Typeface
	}
	label.MaxLines = lines
	label.Alignment = along
	label.Layout(inner)
}

func drawPalette(k Kit, w Window) {
	fill(k.gtx, Box{W: w.width, H: w.height}, k.skin.Colour.Surface.VeilLight)
	rise, _ := w.PaletteRise()
	risen := op.Offset(image.Pt(0, -int((1-rise)*paletteRiseBy))).Push(k.gtx.Ops)
	defer risen.Pop()
	faded := paint.PushOpacity(k.gtx.Ops, float32(rise))
	defer faded.Pop()
	box := w.PaletteBox()
	raised(k, box, k.skin.Radius.Panel, k.skin.Colour.Surface.Raised)
	writing(k, w.measured(), PaletteTyped(box), w.palette.typed, true, reads, 1, Follow)
	shut := w.keyboard.Spelled(Pretty("esc"))
	closeAt := Box{X: box.X + box.W - padding - ends(w.measured(), "close", small), Y: PaletteTyped(box).Y, W: ends(w.measured(), "close", small), H: lineHigh}
	sayIn(k.gtx, k.th, k.skin.Colour.Ink.Low, closeAt, "close", small, font.Font{}, text.End)
	drawKeyChip(k, Box{X: closeAt.X - hintSpace - KeyChipWide(shut), Y: closeAt.Y, H: lineHigh}, shut)
	fill(k.gtx, Box{X: box.X, Y: box.Y + paletteHead - 1, W: box.W, H: 1}, k.skin.Colour.Line.Raised)
	clipped := clip.Rect(rect(Box{X: box.X, Y: box.Y + paletteHead, W: box.W, H: box.H - paletteHead})).Push(k.gtx.Ops)
	defer clipped.Pop()
	rows := w.PaletteRows()
	for _, line := range w.PaletteLines() {
		if line.Row < 0 {
			sayIn(k.gtx, k.th, k.skin.Colour.Ink.Low, Box{X: line.Box.X, Y: line.Box.Y + line.Box.H - lineHigh, W: line.Box.W, H: lineHigh}, strings.ToUpper(line.Group), small, k.skin.Type.MonoBold, text.Start)
			continue
		}
		ink := k.skin.Colour.Ink.Faint
		if line.Row == w.palette.row {
			rounded(k.gtx, line.Box, k.skin.Radius.Row, k.skin.Colour.Surface.Control)
			ink = k.skin.Colour.Ink.Text
		}
		sayIn(k.gtx, k.th, ink, Box{X: line.Box.X + padding/2, Y: line.Box.Y + (line.Box.H-lineHigh)/2, W: line.Box.W - padding, H: lineHigh}, rows[line.Row].Says, reads, font.Font{}, text.Start)
		end := line.Box.X + line.Box.W - padding/2
		if line.Row == w.palette.row {
			enter := w.keyboard.Spelled(Pretty("enter"))
			chip := PaletteKeyAt(line.Box, enter)
			rounded(k.gtx, chip, k.skin.Radius.Chip, k.skin.Colour.Surface.ChipOnControl)
			sayOnChip(k, k.skin.Colour.Ink.Text, chip, enter)
			end = chip.X - hintSpace
		}
		if key := rows[line.Row].Key; key != "" {
			spelled := w.keyboard.Spelled(Pretty(key))
			chip := PaletteKeyAt(Box{X: line.Box.X, Y: line.Box.Y, W: end - line.Box.X + padding/2, H: line.Box.H}, spelled)
			sayOnChip(k, k.skin.Colour.Ink.Faint, chip, spelled)
			end = chip.X - hintSpace
		}
		if ago := rows[line.Row].Ago(w.now); ago != "" {
			sayIn(k.gtx, k.th, k.skin.Colour.Ink.Faint, Box{X: line.Box.X, Y: line.Box.Y + (line.Box.H-lineHigh)/2, W: end - line.Box.X, H: lineHigh}, ago, small, font.Font{}, text.End)
		}
	}
}

func drawPills(k Kit, w Window) {
	for _, pill := range w.Pills() {
		line := Box{X: pill.Box.X, Y: pill.Box.Y + (pill.Box.H-lineHigh)/2, W: pill.Box.W, H: lineHigh}
		switch pill.Kind {
		case PillSearch, PillFilter:
			drawFilterPill(k, w, pill)
		case PillAdd:
			dashedPill(k, pill.Box, k.skin.Colour.Line.Edge)
			sayIn(k.gtx, k.th, k.skin.Colour.Ink.Text, line, pill.Said, small, font.Font{}, text.Middle)
		case PillFolded:
			rounded(k.gtx, pill.Box, k.skin.Radius.Full(pill.Box), k.skin.Colour.Surface.Tint)
			sayIn(k.gtx, k.th, k.skin.Colour.Ink.Text, line, pill.Said, small, font.Font{}, text.Middle)
		case PillClear:
			sayIn(k.gtx, k.th, k.skin.Colour.Ink.Faint, line, pill.Said, small, font.Font{}, text.Middle)
		}
	}
}

func drawFilterPill(k Kit, w Window, pill Pill) {
	name, value := k.skin.Colour.Ink.Faint, k.skin.Colour.Ink.OnTint
	if pill.Faint {
		name, value = k.skin.Colour.Ink.Low, k.skin.Colour.Ink.Faint
	}
	rounded(k.gtx, pill.Box, k.skin.Radius.Full(pill.Box), k.skin.Colour.Surface.Tint)
	line := Box{X: pill.Box.X + pillPad, Y: pill.Box.Y + (pill.Box.H-lineHigh)/2, W: pill.Box.W - pillPad - pillCross, H: lineHigh}
	sayIn(k.gtx, k.th, name, line, pill.Name, small, font.Font{}, text.Start)
	at := ends(w.measured(), pill.Name, small) + nameGap
	sayIn(k.gtx, k.th, value, Box{X: line.X + at, Y: line.Y, W: line.W - at, H: line.H}, strings.TrimPrefix(pill.Said, pill.Name+" "), small, font.Font{Weight: font.Medium}, text.Start)
	drawCross(k, Box{X: pill.Box.X + pill.Box.W - pillCross, Y: pill.Box.Y + (pill.Box.H-pillCross)/2, W: pillCross - 4, H: pillCross}, name)
}

func drawListRow(k Kit, row Box, said string, on bool) {
	if on {
		rounded(k.gtx, row, k.skin.Radius.Row, k.skin.Colour.Surface.Control)
	}
	say(k.gtx, k.th, k.skin.Colour.Ink.Text, Inside(row), said, reads)
}

func drawFieldPills(k Kit, w Window, m Measure, frame Box, labels []string) {
	if len(labels) == 0 {
		return
	}
	room := Inside(frame)
	room.W -= markWide + padding
	lining(k, frame, k.skin.Radius.Control, k.skin.Colour.Surface.Field)
	x, top := room.X, frame.Y+(frame.H-pillHigh)/2
	wide := func(said string) int { return ends(m, said, small) }
	shown, more := Pills(labels, room.W, wide)
	for _, label := range shown {
		pill := Box{X: x, Y: top, W: wide(label) + 2*pillPad, H: pillHigh}
		drawLabelPill(k, w, pill, label)
		x += pill.W + pillGap
	}
	if more > 0 {
		count := "+" + strconv.Itoa(more)
		drawPill(k, Box{X: x, Y: top, W: wide(count) + 2*pillPad, H: pillHigh}, count, k.skin.Colour.Surface.Control)
	}
}
