package gui

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"slices"
	"strconv"
	"strings"

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
	k = k.measuring(w.measured())
	scaled := inDp(k.gtx)
	defer scaled.Pop()

	k.gtx.Metric = unit.Metric{PxPerDp: 1, PxPerSp: 1}

	Backdrop{}.Draw(k.on(LayerBack), Box{W: w.width, H: w.height})
	if w.Screen() == TooSmallToDraw {
		sayCentred(k.gtx, k.th, k.skin.Colour.Ink.Faint, Message(w.width, w.height), w.Says(), reads)
		return
	}
	if w.Blank() {
		drawBlank(k, w)
		return
	}
	l := w.Layout()
	scene, _ := w.Scene()
	window := Box{W: w.width, H: w.height}
	filmed := k.posed(scene.Camera, window)
	top := k.posed(scene.Top, window)
	drawPlace(k, l, w)
	drawTabs(k, l, w)
	if !w.native {
		MenuMark{}.Draw(k.inside("tabs"), MenuButton(l))
	}
	top()
	if w.Tab() == Objectives {
		drawGoals(k, l, w)
		drawGoalDragging(k.on(LayerDrag), w)
	} else if w.Tab() == Time {
		drawWeek(k, l, w)
	} else if w.Tab() == Series {
		drawSeries(k, l, w)
	} else {
		drawColumns(k, l, w, scene)
		drawBand(k, l, w, scene)
		drawArrivals(k.on(LayerDrag), w, scene)
		drawDragging(k.on(LayerDrag), w)
	}
	bar := k.posed(scene.Bar, window)
	drawBar(k.on(LayerBar), l, w)
	drawPills(k.on(LayerBar), w)
	bar()
	filmed()
	drawBoardBanners(k.on(LayerBar), w)
	panel, popover := k.on(LayerPanel), k.on(LayerPopover)
	if w.formUnder() {
		drawForm(panel, w)
	}
	if w.Screen() == OnHelp {
		drawHelp(panel, w)
	}
	if w.Screen() == OnAsking {
		Veil{}.Draw(panel, Box{W: w.width, H: w.height})
		drawSaid(panel.inside("veil"), w, w.Asks(), w.QuestionBox())
	}
	if w.Screen() == OnChoosing {
		drawChoosing(popover, w)
	}
	if w.Screen() == OnFates {
		drawFates(panel, w)
	}
	if w.Screen() == OnSeriesFates {
		drawSeriesFates(panel, w)
	}
	if w.Screen() == OnCalendar {
		drawCalendar(popover, w)
	}
	if w.Screen() == OnRule {
		drawRule(popover, w)
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
	if w.Screen() == OnKeeping {
		drawKeeping(popover, w)
	}
	if w.Screen() == OnViews {
		drawViews(panel, w)
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
		mark.Draw(popover, mark.At)
	}
	if bubble, said, ok := w.BubbleBox(); ok {
		Popover{}.Draw(popover, bubble)
		for i, line := range said {
			say(k.gtx, k.th, k.skin.Colour.Ink.Text, ExplainedLine(bubble, i), line, plain)
		}
	}
	if bubble, pills, ok := w.LabelsBubbleBox(); ok {
		Popover{}.Draw(popover, bubble)
		for i, pill := range pills {
			LabelChip{Label: w.bubbled[i], Colour: w.LabelColour(w.bubbled[i])}.Draw(popover.inside("popover"), pill)
		}
	}
}

func drawBackups(k Kit, w Window) {
	box := w.BackupsBox()
	Panel{Head: Heads(w)}.Draw(k, box)
	k = k.inside("panel")
	from := w.BackupsFrom()
	for i, one := range w.Backups()[from:min(len(w.Backups()), from+w.BackupsRoom())] {
		Row{Said: backupSays(one), Picked: from+i == w.BackupOn()}.Draw(k, ListRowAt(box, i))
	}
	drawPanelOffers(k, box, w)
}

func drawLabels(k Kit, w Window) {
	box := w.LabelsBox()
	Panel{Head: Heads(w)}.Draw(k, box)
	m := w.measured()
	drawLabelSearch(k.inside("panel"), w, m, w.LabelSearchAt())
	rows := w.LabelRows()
	if creating, ok := w.LabelCreating(); ok {
		say(k.gtx, k.th, k.skin.Colour.Accent.Base, Inside(ListRowAt(box, 1)), CreatingSays(creating), reads)
	} else if len(rows) == 0 {
		say(k.gtx, k.th, k.skin.Colour.Ink.Faint, Inside(ListRowAt(box, 1)), NoLabelSays, reads)
	}
	for _, row := range rows {
		drawLabelRow(k.inside("panel"), w, m, row)
	}
	drawPanelOffers(k.inside("panel"), box, w)
	drawSwatches(k.on(LayerPopover), w)
}

func drawLabelSearch(k Kit, w Window, m Measure, row Box) {
	field := Field{Lit: w.Seeking()}.Draw(k, Box{X: row.X, Y: row.Y + 1, W: row.W, H: row.H - 2})
	chip := KeyChip{Key: "/", Quiet: true}.Draw(k.inside("field"), Box{X: field.X + inset, Y: field.Y, H: field.H})
	line := Inside(Box{X: chip.X + chip.W, Y: field.Y, W: field.X + field.W - chip.X - chip.W, H: field.H})
	if w.Sought().Empty() && !w.Seeking() {
		say(k.gtx, k.th, k.skin.Colour.Ink.Faint, line, LabelSearchSays, reads)
		return
	}
	writing(k, m, line, w.Sought(), w.Seeking(), reads, 1, Follow)
}

func drawLabelRow(k Kit, w Window, m Measure, row LabelRow) {
	Row{Picked: row.On}.Draw(k, row.At)
	k = k.inside("row")
	line := Inside(row.At)
	if row.Typing {
		field := Field{Lit: true}.Draw(k, Box{X: line.X, Y: row.At.Y + 2, W: line.W / 2, H: row.At.H - 4})
		writing(k, m, Inside(field), w.Renamed(), true, reads, 1, Follow)
		return
	}
	Dot{Colour: w.LabelColour(row.Carried.Label)}.Draw(k, row.Dot)
	LabelChip{Label: row.Carried.Label, Colour: w.LabelColour(row.Carried.Label)}.Draw(k, row.Name)
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
		Button{Said: action.Offer, Key: w.OfferKey(action.Offer), Look: look, Keyed: true}.Draw(k, action.At)
	}
}

func drawSwatches(k Kit, w Window) {
	swatches := w.ColourSwatches()
	if len(swatches) == 0 {
		return
	}
	first, last := swatches[0].At, swatches[len(swatches)-1].At
	Popover{}.Draw(k, Box{X: first.X - inset*2, Y: first.Y - inset*2, W: last.X + last.W - first.X + inset*4, H: first.H + inset*4})
	for i, swatch := range swatches {
		Swatch{Colour: swatch.Colour, Picked: i == w.ColourPick()}.Draw(k.inside("popover"), swatch.At)
	}
}

func drawDropFoot(k Kit, w Window, box Box, keys [][2]string) {
	foot := Box{X: box.X, Y: box.Y + box.H - dropFoot, W: box.W, H: dropFoot}
	Rule{}.Draw(k, Box{X: box.X, Y: foot.Y, W: box.W})
	x := foot.X + padding
	for _, key := range keys {
		if x+KeyChipWide(key[0])+hintSpace+ends(w.measured(), key[1], small) > box.X+box.W-padding {
			return
		}
		chip := KeyChip{Key: key[0]}.Draw(k, Box{X: x, Y: foot.Y, H: foot.H})
		said := Box{X: chip.X + chip.W + hintSpace, Y: foot.Y + (foot.H-lineHigh)/2, W: ends(w.measured(), key[1], small), H: lineHigh}
		sayIn(k.gtx, k.th, k.skin.Colour.Ink.Faint, said, key[1], small, font.Font{}, text.Start)
		x = said.X + said.W + padding
	}
}

func drawKeeping(k Kit, w Window) {
	box := w.KeepingBox()
	Panel{}.Draw(k, box)
	k = k.inside("panel")
	head := Box{X: box.X + padding, Y: box.Y + (dropHead-lineHigh)/2, W: box.W - 2*padding, H: lineHigh}
	Head{Said: w.KeepingHead()}.Draw(k, head)
	Rule{}.Draw(k, Box{X: box.X, Y: box.Y + dropHead - 1, W: box.W})
	if !w.ofView {
		field := Field{Lit: true}.Draw(k, w.KeepingField())
		writing(k.inside("field"), w.measured(), Inside(field), w.Naming(), true, reads, 1, Follow)
	}
	w.keepingFacts().Draw(k, w.KeepingFactsAt())
	drawDropFoot(k, w, box, w.KeepingFoot())
}

func drawViews(k Kit, w Window) {
	box := w.ViewsBox()
	Panel{Head: Heads(w)}.Draw(k, box)
	k = k.inside("panel")
	rows := w.ViewsPanelRows()
	if len(rows) == 0 {
		say(k.gtx, k.th, k.skin.Colour.Ink.Faint, Inside(ListRowAt(box, 0)), ViewsNoneSays, reads)
	}
	for _, row := range rows {
		drawViewRow(k, w, row, true)
	}
	if said, at, ok := w.ViewsWarning(); ok {
		Banner{Said: said, Then: viewsThenSays, Warn: true}.Draw(k, at)
	}
	drawPanelOffers(k, box, w)
}

func drawViewRow(k Kit, w Window, row ViewRow, holds bool) {
	r := Row{Picked: row.On, Taken: row.Taken}
	r.Draw(k, row.At)
	line := Inside(row.At)
	key := w.keyboard.Spelled(Pretty(row.Key))
	chip := Box{X: line.X + line.W - KeyChipWide(key), Y: line.Y, W: KeyChipWide(key), H: line.H}
	if key != "" {
		r.Chip(k, KeyChip{Key: key, Quiet: true}, chip)
	}
	count := strconv.Itoa(row.Keeps)
	counted := Box{X: chip.X - padding - monoWide(count, small), Y: line.Y, W: monoWide(count, small), H: line.H}
	sayIn(k.gtx, k.th, k.skin.Colour.Ink.Low, counted, count, small, k.skin.Type.Mono, text.End)
	name := Box{X: line.X, Y: line.Y, W: counted.X - padding - line.X, H: line.H}
	if row.Typing {
		field := Field{Lit: true}.Draw(k.inside("row"), Box{X: name.X, Y: row.At.Y + 2, W: name.W / 2, H: row.At.H - 4})
		writing(k.inside("row").inside("field"), w.measured(), Inside(field), w.viewRenamed, true, reads, 1, Follow)
		return
	}
	said := ends(w.measured(), row.Name, reads)
	if !holds || said+padding >= name.W {
		say(k.gtx, k.th, k.skin.Colour.Ink.Text, name, row.Name, reads)
		return
	}
	say(k.gtx, k.th, k.skin.Colour.Ink.Text, Box{X: name.X, Y: name.Y, W: said + padding/2, H: name.H}, row.Name, reads)
	ink := k.skin.Colour.Ink.Faint
	if row.Late {
		ink = k.skin.Colour.State.Late
	}
	sayIn(k.gtx, k.th, ink, Box{X: name.X + said + padding, Y: name.Y, W: name.W - said - padding, H: name.H}, row.Holds, small, font.Font{}, text.Start)
}

func drawFilters(k Kit, w Window) {
	box := w.FiltersBox()
	Panel{}.Draw(k, box)
	k = k.inside("panel")
	drawDropFoot(k, w, box, w.FilterFoot())
	columns, stub := w.FilterColumns()
	if stub.W > 0 {
		say(k.gtx, k.th, k.skin.Colour.Ink.Faint, Box{X: stub.X + inset, Y: stub.Y + (dropHead-lineHigh)/2, W: stub.W - inset, H: lineHigh}, "‹", reads)
	}
	for i, column := range columns {
		if i > 0 || stub.W > 0 {
			Rule{Upright: true}.Draw(k, Box{X: column.Box.X, Y: column.Box.Y, W: 1, H: column.Box.H - dropFoot})
		}
		head := Box{X: column.Box.X + padding, Y: column.Box.Y + (dropHead-lineHigh)/2, W: column.Box.W - 2*padding, H: lineHigh}
		if column.At == 0 {
			drawFilterNames(k, w, head)
			continue
		}
		drawFilterValues(k, w, column, head, i == len(columns)-1)
	}
}

func drawFilterValues(k Kit, w Window, column FilterColumn, head Box, last bool) {
	f := w.Filters()[w.FilterOn()]
	Rule{}.Draw(k, Box{X: column.Box.X, Y: column.Box.Y + dropHead - 1, W: column.Box.W})
	if column.At == 1 {
		say(k.gtx, k.th, k.skin.Colour.Ink.Text, head, Titled(f.Name), reads)
		sayIn(k.gtx, k.th, k.skin.Colour.Ink.Faint, head, w.FilterHint(), small, font.Font{}, text.End)
	} else if g, found := w.snap.Goal(w.filterTrail[column.At-2]); found {
		say(k.gtx, k.th, k.skin.Colour.Ink.Text, head, "#"+strconv.Itoa(g.Number)+" "+g.Title, reads)
	}
	counts := w.FilterCounts()
	values, at := w.ValueRows(column)
	for j, i := range values {
		taken := w.FilterTaken(i)
		opened := !last && column.At-1 < len(w.filterTrail) && w.goalOpened(i, column.At-1)
		row := Row{Taken: taken, Picked: last && i == w.FilterPick(), Hover: opened}
		row.Draw(k, at[j])
		line := Inside(at[j])
		row.Check(k, Check{Taken: taken}, CheckAt(line))
		named, count := FilterValueParts(at[j])
		if f.Arg == "label" {
			dot := Box{X: named.X, Y: named.Y + (named.H-dotSide)/2, W: dotSide, H: dotSide}
			Dot{Colour: w.LabelColour(f.Shown[i])}.Draw(k.inside("row"), dot)
			named.X, named.W = named.X+dotSide+between, named.W-dotSide-between
		}
		say(k.gtx, k.th, k.skin.Colour.Ink.Text, named, f.Shown[i], reads)
		said := ""
		if i < len(counts) {
			said = strconv.Itoa(counts[i])
		}
		if w.opensAColumn(i, column.At-1) {
			said += " ›"
		}
		sayIn(k.gtx, k.th, k.skin.Colour.Ink.Faint, count, said, small, k.skin.Type.Mono, text.End)
	}
}

func drawFilterNames(k Kit, w Window, head Box) {
	v, viewed := w.inView()
	if rows := w.ViewRows(); len(rows) > 0 {
		Head{Said: ViewsSays}.Draw(k, head)
		for _, row := range rows {
			drawViewRow(k, w, row, false)
		}
		Row{Said: EditViewsSays}.Draw(k, w.EditViewsAt())
		head = Inside(w.FilterHeadAt())
	}
	Head{Said: FilterMenuSays}.Draw(k, head)
	open := w.Screen() == OnFiltering && w.filterOpen
	for i, f := range w.Filters() {
		row := w.FilterRowAt(i)
		Row{Said: Titled(f.Name), Picked: !open && i == w.FilterOn() && w.ViewWalked() == 0, Hover: open && i == w.FilterOn()}.Draw(k, row)
		line := Inside(row)
		if len(f.Taken()) == 0 && viewed && v.Holds.Narrows(f.Arg) {
			at := ends(w.measured(), Titled(f.Name), reads) + between
			sayIn(k.gtx, k.th, k.skin.Colour.Ink.Faint, Box{X: line.X + at, Y: line.Y, W: line.W - at - padding - 2*small, H: line.H}, InTheViewSays, small, font.Font{}, text.Start)
		}
		if len(f.Taken()) > 0 {
			at := ends(w.measured(), Titled(f.Name), reads) + between
			sayIn(k.gtx, k.th, k.skin.Colour.Accent.Base, Box{X: line.X + at, Y: line.Y, W: line.W - at - padding - 2*small, H: line.H}, f.TakenSays(), small, font.Font{}, text.Start)
		}
		sayIn(k.gtx, k.th, k.skin.Colour.Ink.Faint, line, FilterKey(f)+" ›", small, k.skin.Type.Mono, text.End)
	}
}

func drawWeek(k Kit, l Layout, w Window) {
	measuring := k.gtx
	measuring.Constraints.Min = image.Point{}
	ranges := []string{"Day", "Week", "Month"}
	var segments []Box
	for _, chip := range ranges {
		segments = append(segments, TimeChip(l, chip))
	}
	Segmented{Said: ranges, At: segments, On: slices.Index(ranges, w.ranging())}.Draw(k, RangesControl(l))
	for _, chip := range TimeChips {
		if !slices.Contains(ranges, chip) {
			Button{Said: chip, Look: buttonPlain}.Draw(k, TimeChip(l, chip))
		}
	}
	if w.Monthly() {
		Button{Said: w.week.Month().String(), Look: lookOf(w.grid == viewMonths)}.Draw(k, TimeChip(l, "the month"))
		Button{Said: strconv.Itoa(w.week.Year()), Look: lookOf(w.grid == viewYears)}.Draw(k, TimeChip(l, "the year"))
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
		Ground{Raised: true}.Draw(k, box)
		on := k.inside("ground")
		head := WeekHeadAt(box)
		ink := k.skin.Colour.Ink.Faint
		if day.Equal(w.today()) {
			TodayMark{Radius: k.skin.Radius.Today}.Draw(on, TodayPill(head, w.measured()))
			ink = k.skin.Colour.Ink.Text
		}
		say(k.gtx, k.th, ink, head, day.Format(todayFormat), reads)
		sayIn(k.gtx, k.th, k.skin.Colour.Ink.Low, WeekTotalAt(box, w.measured()), w.DayHeld(i).Says(), small, k.skin.Type.Mono, text.End)
		in, clipped := on.within(box)
		for j, one := range w.DayShown(i) {
			row := WeekRowAt(box, j)
			card := DayCardAt(box, j, len(w.DayHeld(i).Tasks))
			w.cardOf(w.DayCard(one), w.Chosen() == one.Number && !w.chosenGoal, card == w.pointed).Draw(in, card)
			say(k.gtx, k.th, k.skin.Colour.Accent.Base, DaySpentAt(row), one.Spent(), small)
			Spent{}.Draw(in, WeekBar(row, one.Minutes, longest))
		}
		clipped.Pop()
		if track, more := DayTrack(box, len(w.DayHeld(i).Tasks)); more {
			Scrollbar{Lines: len(w.DayHeld(i).Tasks), Rows: DayRoom(box), At: w.DayFrom(i)}.Draw(on, track)
		}
	}
}

func drawDayAlone(k Kit, l Layout, w Window, longest int) {
	box := DayAloneAt(l)
	held := w.DayHeld(0)
	Ground{Raised: true}.Draw(k, box)
	k = k.inside("ground")
	say(k.gtx, k.th, k.skin.Colour.Ink.Faint, WeekHeadAt(box), held.Says(), reads)
	in, clipped := k.within(box)
	for j, one := range w.AloneShown() {
		w.cardOf(w.DayCard(one), w.Chosen() == one.Number && !w.chosenGoal, AloneCardAt(box, j) == w.pointed).Draw(in, AloneCardAt(box, j))
		say(k.gtx, k.th, k.skin.Colour.Accent.Base, AloneSpentAt(box, j), one.Spent(), small)
		Spent{}.Draw(in, AloneBar(box, j, one.Minutes, longest))
	}
	clipped.Pop()
	if track, more := DayTrack(box, len(held.Tasks)); more {
		Scrollbar{Lines: len(held.Tasks), Rows: DayRoom(box), At: w.DayFrom(0)}.Draw(k, track)
	}
}

func drawMonth(k Kit, l Layout, w Window) {
	if cells := w.TimeCells(); len(cells) > 0 {
		for i, cell := range cells {
			Button{Said: cell, Look: lookOf(cell == gridMarked(w.grid, w.week))}.Draw(k, TimeCellAt(l, i))
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
		DayCell{Today: day.Equal(w.today()), Raised: true}.Draw(k, cell)
		on := k.inside("day")
		ink := k.skin.Colour.Ink.Text
		if !sameMonth(day, w.week) {
			ink = k.skin.Colour.Ink.Faint
		}
		held := w.MonthDay(i)
		say(k.gtx, k.th, ink, CellHeadAt(cell), CellHead(day, held), reads)
		in, clipped := on.within(cell)
		shown := cellShown(held.Tasks, CellRoom(cell))
		for j, line := range CellLines(held.Tasks, CellRoom(cell)) {
			at := CellLineAt(cell, j)
			if j < shown {
				c := w.DayCard(held.Tasks[j])
				Stripe{Priority: priorityOf(c), Free: true}.Draw(in, Box{X: at.X - inset, Y: at.Y, W: edgeWide, H: at.H})
				if w.Chosen() == c.Number && !w.chosenGoal {
					Ring{Radius: k.skin.Radius.Row}.Draw(in, at)
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
	Popover{}.Draw(k, box)
	k = k.inside("popover")
	for i, said := range w.Menu() {
		at := MenuRowAt(box, i)
		row := Row{Said: said, Danger: said == "Delete", Picked: i == w.MenuOn()}
		row.Draw(k, at)
		if key := MenuKey(said); key != "" {
			spelled := w.keyboard.Spelled(Pretty(key))
			row.Chip(k, KeyChip{Key: spelled}, Box{X: at.X + at.W - KeyChipWide(spelled) - inset*2, Y: at.Y, H: at.H})
		}
	}
}

func drawCalendar(k Kit, w Window) {
	box := w.CalendarBox()
	Panel{}.Draw(k, box)
	k = k.inside("panel")
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
		Button{Said: chip.says, Look: lookOf(chip.on)}.Draw(k, chip.at)
	}
	if cells := w.CalendarCells(); len(cells) > 0 {
		for i, cell := range cells {
			Button{Said: cell, Look: lookOf(cell == w.CalendarMarked())}.Draw(k, CalendarCell(box, i))
		}
		drawPanelOffers(k, box, w)
		return
	}
	for i, name := range Weekdays {
		say(k.gtx, k.th, k.skin.Colour.Ink.Faint, WeekdayAt(box, i), name, reads)
	}
	today := w.today()
	for i, day := range engine.MonthGrid(w.day) {
		DayCell{Said: strconv.Itoa(day.Day()), Picked: day.Equal(w.day), Today: day.Equal(today), Faded: w.Faded(day)}.Draw(k, DayAt(box, i))
	}
	drawPanelOffers(k, box, w)
}

func drawChoosing(k Kit, w Window) {
	box := w.ChoosingBox()
	Panel{}.Draw(k, box)
	k = k.inside("panel")

	if !box.Unsearched {
		looking := Field{}.Draw(k, ChoosingLooking(box))
		typed := Inside(looking)
		for _, taken := range w.LabelsTaken() {
			LabelChip{Label: taken.Label, Colour: w.LabelColour(taken.Label), Removable: true}.Draw(k.inside("field"), taken.Pill)
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
	in, clipped := k.within(run)
	from := w.ChoosingFromRow()
	for i := from; i < Lasting(len(offered), from, w.ChoosingRoom()); i++ {
		one := offered[i]
		row := Row{Picked: i == w.ChoosingAt()}
		row.Draw(in, ChoosingRow(box, at))
		line := ChoosingLine(box, at)
		if one.Creates != "" {
			said := "Create"
			say(k.gtx, k.th, k.skin.Colour.Ink.Faint, line, said, reads)
			from := line.X + ends(w.measured(), said, reads) + pillGap*2
			row.Pill(in, Pill{Said: one.Creates}, Box{X: from, Y: line.Y + (line.H-pillHigh)/2, W: ends(w.measured(), one.Creates, small) + 2*pillPad, H: pillHigh})
			at += apart
			continue
		}
		if one.Depth > 0 {
			top := ChoosingRow(box, at).Y
			gutter := apart - ChoosingRow(box, at).H
			row.Connector(in, Connector{Last: one.Last, Elbow: line.Y + line.H/2 - top + gutter}, Box{X: line.X + checkSide/2, Y: top - gutter, W: indent - checkSide/2 - inset, H: apart})
			line.X, line.W = line.X+indent, line.W-indent
		}
		row.Check(in, Check{Taken: w.Ticked(one.Number)}, CheckAt(line))
		if w.choosingFor == "label" {
			LabelChip{Label: one.Title, Colour: w.LabelColour(one.Title)}.Draw(in.inside("row"), LabelPillAt(line, ends(w.measured(), one.Title, small)))
			sayIn(k.gtx, k.th, k.skin.Colour.Ink.Faint, line, strconv.Itoa(one.Carried), small, k.skin.Type.Mono, text.End)
			at += apart
			continue
		}
		ink := k.skin.Colour.Ink.Text
		if aside := w.RowAside(one); aside != "" {
			ink = k.skin.Colour.Ink.Faint
			sayIn(k.gtx, k.th, k.skin.Colour.Ink.Faint, line, aside, small, font.Font{}, text.End)
		}
		say(k.gtx, k.th, ink, PastCheck(line), w.RowSays(one), reads)
		if stripe, ok := w.RowStripe(one, line); ok {
			Stripe{Priority: one.Title, Free: true}.Draw(in.inside("row"), stripe)
		}
		at += apart
	}
	clipped.Pop()

	Scrollbar{Lines: len(offered), Rows: w.ChoosingRoom(), At: from}.Draw(k, ChoosingBar(box))
	drawPanelOffers(k, box, w)
	if refused, at, ok := w.RefusalBox(); ok {
		Banner{Said: refused, Then: w.ChoosingThen()}.Draw(k.on(LayerPopover), at)
	}
}

func drawSaid(k Kit, w Window, said []Said, box Box) {
	Panel{Head: Heads(w)}.Draw(k, box)
	k = k.inside("panel")
	at := FirstRow(box)
	for _, line := range said {
		row := RowIn(box, at, lineHigh)
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
	Panel{Head: Heads(w)}.Draw(k, box)
	k = k.inside("panel")
	for _, line := range w.HelpLines() {
		if line.Head != "" {
			Head{Said: line.Head}.Draw(k, line.At)
			continue
		}
		KeyChip{Key: line.Key}.Draw(k, line.KeyAt)
		say(k.gtx, k.th, k.skin.Colour.Ink.Low, line.MeansAt, line.Means, reads)
	}
	drawPanelOffers(k, box, w)
}

func drawTabs(k Kit, l Layout, w Window) {
	tabs := Tabs{On: -1}
	for i, page := range w.Shown() {
		tabs.Said = append(tabs.Said, Titled(Pages[page]))
		tabs.Counts = append(tabs.Counts, w.TabCount(page))
		tabs.At = append(tabs.At, TabAt(l, i))
		if w.Tab() == page {
			tabs.On = i
		}
	}
	tabs.Draw(k, l.Tabs)
}

func drawColumns(k Kit, l Layout, w Window, scene Scene) {
	dealt, _ := engine.Deal(w.Showing().Cards, w.archived)
	for i, column := range dealt {
		ground := Box{X: l.Head[i].X, Y: l.Head[i].Y, W: l.Head[i].W, H: l.Head[i].H + l.Column[i].H}
		pose := scene.ground(i)
		done := k.posed(pose, ground)
		drawGround(k, ground, pose)
		on := k.inside("ground")
		drawHead(on, NameIn(l.Head[i]), Titled(column.Name), len(column.Cards))
		if said := w.EmptySays(i); said != "" {
			Empty{Said: said}.Draw(on, EmptyAt(l.CardAt(i, 0, 0)))
		}
		drawPlaced(on, w, l.Column[i], scene)
		drawReferences(on, w, l.Column[i])
		if pose.Shade > 0 {
			Wash{Dark: true, Ground: true, By: pose.Shade}.Draw(on, ground)
		}
		done()
	}
}

func drawGround(k Kit, at Box, pose Pose) {
	if pose.Draft == nil {
		Ground{}.Draw(k, at)
		return
	}
	filled := k.faded(float32(pose.Draft.Fill))
	Ground{}.Draw(k, at)
	filled.Pop()
	Drafted{Ground: true, By: 1 - pose.Draft.Fill}.Draw(k, at)
}

func (sc Scene) ground(i int) Pose {
	if i < len(sc.Grounds) {
		return sc.Grounds[i]
	}
	return AtRest()
}

func (sc Scene) card(n int) Pose {
	if n < len(sc.Cards) {
		return sc.Cards[n]
	}
	return AtRest()
}

// InTheViewSays is what a filter of the menu says when the view in force holds it.
const InTheViewSays = "in the view"

func drawReferences(k Kit, w Window, within Box) {
	for _, p := range w.References() {
		if p.Within != within || !p.seen() {
			continue
		}
		in, clipped := k.within(p.Within)
		card := w.cardOf(p.Card, false, false)
		card.Reference, card.Why = true, p.Why()
		card.Draw(in, p.Slot)
		clipped.Pop()
	}
}

func drawBand(k Kit, l Layout, w Window, scene Scene) {
	for lane, kind := range engine.Stalled {
		ground := LaneGround(l, lane)
		pose := scene.ground(len(l.Column) + lane)
		done := k.posed(pose, ground)
		drawGround(k, ground, pose)
		on := k.inside("ground")
		held := len(w.inLane(lane))
		drawHead(on, LaneName(l, lane), Titled(kind), held)
		if held == 0 {
			Empty{Said: nothingHere}.Draw(on, EmptyAt(l.LaneCardAt(lane, 0, 0)))
		}
		drawPlaced(on, w, l.Lane[lane], scene)
		if pose.Shade > 0 {
			Wash{Dark: true, Ground: true, By: pose.Shade}.Draw(on, ground)
		}
		done()
	}
	drawArriving(k, w, scene)
}

func drawHead(k Kit, room Box, name string, held int) {
	Head{Said: name, Count: strconv.Itoa(held)}.Draw(k, room)
}

func drawPlaced(k Kit, w Window, within Box, scene Scene) {
	for n, p := range w.seenPlaced() {
		pose := scene.card(n)
		if p.Within != within || pose.Free {
			continue
		}
		in, clipped := k.within(p.Within)
		drawPosed(in, w, p, pose, scene.Beat != nil)
		clipped.Pop()
	}
}

func drawPosed(k Kit, w Window, p Placed, pose Pose, beating bool) {
	chosen := p.Card.Number == w.Chosen() || w.isTaken(p.Card.Number)
	if w.Dragging() && (p.Card.Number == w.held.Number || w.isTaken(w.held.Number) && w.isTaken(p.Card.Number)) {
		origin := k.faded(OriginShows)
		defer origin.Pop()
	}
	card := w.cardOf(p.Card, chosen, p.Slot == w.pointed)
	card.Dimmed, card.Undotted, card.Draft = Dimmed(p.Card), beating && p.Card.InFocus, pose.Draft
	if pose.Draft != nil && pose.Draft.Typed >= 0 {
		card.Face.Title = string([]rune(card.Face.Title)[:min(pose.Draft.Typed, len([]rune(card.Face.Title)))])
	}
	panel := PartsOf(p.Slot).Panel
	done := k.posed(pose, panel)
	defer done()
	if pose.Lift > 0 {
		Lifted{By: pose.Lift}.Draw(k, panel)
	}
	if pose.Soft > 0 {
		for _, by := range [][2]float64{{-1, -1}, {1, -1}, {-1, 1}, {1, 1}} {
			ghost := k.posed(Pose{Seen: 0.3, DX: by[0] * pose.Soft / 2, DY: by[1] * pose.Soft / 2, Grow: 1, Wide: 1, High: 1}, panel)
			card.Draw(k, p.Slot)
			ghost()
		}
	}
	card.Draw(k, p.Slot)
	if pose.Glare > 0 {
		Wash{By: pose.Glare}.Draw(k, panel)
	}
	if pose.Shade > 0 {
		Wash{Dark: true, By: pose.Shade}.Draw(k, panel)
	}
}

func drawArrivals(k Kit, w Window, scene Scene) {
	seen := w.seenPlaced()
	for _, n := range scene.Order {
		if n < len(seen) {
			drawPosed(k, w, seen[n], scene.card(n), scene.Beat != nil)
		}
	}
	window := Box{W: w.width, H: w.height}
	if scene.Wave != nil {
		scene.Wave.Draw(k, window)
	}
	if scene.Sweep != nil {
		scene.Sweep.Draw(k, window)
	}
	if len(scene.Grains) > 0 {
		scene.Grains.Draw(k, window)
	}
	if scene.Beat != nil {
		scene.Beat.Draw(k, scene.Beat.Mark)
	}
}

func (w Window) seenPlaced() []Placed {
	var seen []Placed
	for _, p := range w.Placed() {
		if p.seen() {
			seen = append(seen, p)
		}
	}
	return seen
}

func drawArriving(k Kit, w Window, scene Scene) {
	if !scene.Stones {
		return
	}
	progress := scene.Progress
	seen := w.seenPlaced()
	for n, p := range seen {
		fall, fade := Arriving(progress, n, len(seen))
		panel := PartsOf(p.Slot).Panel
		faded := k.faded(float32(1 - fade))
		for i, stone := range StonesOf(panel, p.Card.Number) {
			Stone{Shape: Fallen(stone, Dropped(panel, fall)), Shade: k.skin.stoneOf(i)}.Draw(k, panel)
		}
		faded.Pop()
	}
}

func disc(gtx layout.Context, x, y, r float32, ink color.NRGBA) {
	round := clip.Ellipse{Min: image.Pt(int(x-r+0.5), int(y-r+0.5)), Max: image.Pt(int(x+r+0.5), int(y+r+0.5))}
	paint.FillShape(gtx.Ops, ink, round.Op(gtx.Ops))
}

func (w Window) cardOf(c engine.Card, chosen, hovered bool) Card {
	return Card{Face: c, Chosen: chosen, Hovered: hovered, Pages: w.pages, Colour: w.LabelColour}
}

func goalFace(g engine.Goal) engine.Card {
	return engine.Card{Number: g.Number, Title: g.Title, State: g.State, Priority: g.Priority, Labels: Standings(g),
		Deadline: g.Deadline}
}

func drawBlank(k Kit, w Window) {
	if w.Screen() == OnConfiguration {
		drawConfiguration(k.on(LayerPanel), w)
		return
	}
	offer := BlankOffer(w.width, w.height)
	Empty{Said: w.BlankSays()}.Draw(k, Box{X: offer.X - 60, Y: offer.Y - 3*lineHigh, W: offer.W + 120, H: 2 * lineHigh})
	Button{Said: "Open settings", Key: Pretty("cmd+,"), Keyed: true}.Draw(k, offer)
}

func drawGoalDragging(k Kit, w Window) {
	if landing, somewhere := w.GoalLanding(); somewhere && landing.H == landingHigh {
		Landing{}.Draw(k, landing)
	} else if somewhere {
		Ring{Target: true, Radius: k.skin.Radius.Card}.Draw(k, landing)
	}
	if landing, somewhere := w.GoalLanding(); somewhere {
		said := w.GoalLandingSays()
		wide := 7*len(said) + 16
		Pill{Said: said, Tone: ToneHere}.Draw(k, Box{X: landing.X + landing.W - wide - 8, Y: landing.Y - 10, W: wide, H: 20})
	}
	if !w.GoalDragging() {
		return
	}
	ghost, found := w.GoalGhost()
	if !found {
		return
	}
	l := w.Layout()
	for at, b := range w.Tree() {
		if b.Goal.Number != w.grabbedGoal {
			continue
		}
		slot := GoalAt(l, at, w.goals)
		row := ItemOf(slot, depthOf(b))
		seen := k.faded(w.GhostShows())
		Lifted{By: 1}.Draw(k, ghost)
		w.goalItem(b, slot).Draw(k, Box{X: slot.X + ghost.X - row.X, Y: slot.Y + ghost.Y - row.Y, W: slot.W, H: slot.H})
		seen.Pop()
	}
}

func drawDragging(k Kit, w Window) {
	landing, somewhere := w.Landing()
	slot, dragging := w.ghostSlot()
	if !dragging {
		return
	}
	for _, c := range w.Reading().Cards {
		if c.Number != w.held.Number {
			continue
		}
		shadow := k.faded(w.GhostShows())
		Lifted{By: 1}.Draw(k, PartsOf(slot).Panel)
		shadow.Pop()
		if somewhere {
			Landing{}.Draw(k, landing)
		}
		seen := k.faded(w.GhostShows())
		w.cardOf(c, false, false).Draw(k, slot)
		seen.Pop()
	}
}

func drawBar(k Kit, l Layout, w Window) {
	if w.Wrong() != nil {
		Backdrop{}.Draw(k, WrongBox(l))
		sayOver(k.gtx, k.th, k.skin.Colour.State.Wrong, WrongBox(l), w.Wrong().Error(), reads, 2)
	} else if w.Note() != "" {
		Backdrop{}.Draw(k, WrongBox(l))
		sayOver(k.gtx, k.th, k.skin.Colour.Ink.Text, WrongBox(l), w.Note(), reads, 2)
	}
	bar := Bar{}.Draw(k, l.Bar)
	k = k.inside("bar")
	offers, faint := w.Offers(), false
	if s := w.Screen(); s != OnTheBoard && s != OnSearch {
		offers, faint = w.BoardOffers(), true
	}
	along, worded := w.barLaid(offers)
	if w.Screen() == OnSearch {
		KeyChip{Key: "/"}.Draw(k, Box{X: bar.X + barInset, Y: bar.Y, H: bar.H})
		writing(k, w.measured(), SearchInto(bar), w.Searched(), true, reads, 1, Follow)
	} else {
		chip := KeyChip{Key: w.keyboard.Spelled(Pretty("cmd+k"))}.Draw(k, Box{X: bar.X + barInset, Y: bar.Y, H: bar.H})
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
		Hint{Key: w.BarKey(offer), Word: word, Faint: faint, Lit: !faint && w.Lighting(offer)}.Draw(k, along[i])
	}
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

func drawSub(k Kit, box Box, c engine.Card, line string, on, picked bool, held Written) {
	panel := Tile{Picked: picked}.Draw(k, SubPanel(box))
	k = k.inside("tile")
	if on {
		Ring{Radius: k.skin.Radius.Row}.Draw(k, panel)
	}
	done := engine.Finished(c)
	Check{Done: done}.Draw(k, SubCheck(box))
	writ := SubWriting(box)
	said := fits(k.gtx, k.th, line, writ.W, small)
	held.Value = said
	sayHolding(k, SubInk(k.skin, c), writ, held, small)
	if done {
		loose := k.gtx
		loose.Constraints = layout.Constraints{Max: image.Pt(writ.W, lineHigh)}
		Struck{}.Draw(k, StruckAt(writ, wide(loose, k.th, said, small).Size.X))
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
	on := 0
	if w.Tracing() {
		on = 1
	}
	Segmented{Said: says[:], At: tabs[:], On: on}.Draw(k, Box{X: tabs[0].X, Y: tabs[0].Y, W: tabs[1].X + tabs[1].W - tabs[0].X, H: tabs[0].H})
}

func drawUnder(k Kit, w Window) {
	under := w.Under()
	done := 0
	for _, c := range under {
		if engine.Finished(c) {
			done++
		}
	}
	Head{Said: ServingHead, Count: fmt.Sprintf("%s · %d done", countOf(len(under), "task"), done)}.Draw(k, w.WorkName())
	frame := Field{}.Draw(k, w.workFrame())
	in, clipped := k.inside("field").within(frame)
	on, walking := w.WorkOn()
	for _, tile := range w.WorkTiles() {
		drawWorkTile(in, w, tile)
		if walking && tile == on {
			Ring{Radius: k.skin.Radius.Row}.Draw(in.inside("tile"), workTileAt(tile))
		}
	}
	clipped.Pop()
	Scrollbar{Lines: (len(under) + 1) * SubHeight, Rows: frame.H, At: w.ServingAt()}.Draw(k, w.workBar())
}

func drawWorkTile(k Kit, w Window, tile WorkTile) {
	line := SubWriting(tile.At)
	if tile.Number == 0 {
		say(k.gtx, k.th, k.skin.Colour.Ink.Faint, Box{X: tile.At.X + padding, Y: line.Y, W: tile.At.W - 2*padding, H: line.H}, tile.Said, small)
		return
	}
	Tile{}.Draw(k, workTileAt(tile))
	k = k.inside("tile")
	c, _ := w.snap.Card(tile.Number)
	x := tile.At.X + padding
	number := "#" + strconv.Itoa(c.Number)
	sayIn(k.gtx, k.th, k.skin.Colour.Ink.Faint, Box{X: x, Y: line.Y, W: monoWide(number, small), H: line.H}, number, small, k.skin.Type.Mono, text.Start)
	x += monoWide(number, small) + padding
	spent := w.spentSaid(c.Spent)
	spentAt := Box{X: tile.At.X + tile.At.W - padding - monoWide(widestTime, small), Y: line.Y, W: monoWide(widestTime, small), H: line.H}
	sayIn(k.gtx, k.th, k.skin.Colour.Ink.Faint, spentAt, spent, small, k.skin.Type.Mono, text.End)
	state := c.State
	pill := Box{X: spentAt.X - padding - ends(w.measured(), state, small) - 2*pillPad, Y: line.Y + (line.H-pillHigh)/2, W: ends(w.measured(), state, small) + 2*pillPad, H: pillHigh}
	Pill{Said: state}.Draw(k, pill)
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
		Struck{}.Draw(k, StruckAt(title, wide(loose, k.th, said, small).Size.X))
	}
}

func drawProgressBeside(k Kit, w Window) {
	g, ok := w.snap.Goal(w.form.Number)
	if !ok {
		return
	}
	at := w.ProgressBox()
	card := InfoBox{}.Draw(k, at)
	k = k.inside("info")
	inside := Box{X: card.X + padding, Y: card.Y + padding, W: card.W - 2*padding, H: card.H - 2*padding}
	Head{Said: ProgressHead, Count: strconv.Itoa(g.Percent()) + "%"}.Draw(k, Box{X: inside.X, Y: inside.Y, W: inside.W, H: lineHigh})
	track := Progress{Done: g.Done, Total: g.Total}.Draw(k, Box{X: inside.X, Y: inside.Y + lineHigh + between, W: inside.W, H: k.skin.Size.Track})
	facts := w.GoalProgress(g).Within(at.H - progressHead)
	facts.Draw(k, Box{X: inside.X, Y: track.Y + track.H + padding, W: inside.W, H: facts.High()})
}

func drawHistory(k Kit, box Box, w Window) {
	frame := Field{}.Draw(k, BesideBox(box))

	inside := Inside(frame)
	in, clipped := k.inside("field").within(inside)
	told, from, rows := w.traced(), w.History(), w.Rows()
	for i := from; i < Lasting(len(told.Said), from, rows); i++ {
		drawTold(in, LineAt(inside, i, from), w.HoldingFrom("trace", told.Said[i], told.From[i]))
	}
	clipped.Pop()

	Scrollbar{Lines: len(told.Said), Rows: rows, At: w.History()}.Draw(k, BesideBar(box))
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

func drawTicked(k Kit, name, frame Box, w Window) {
	f := w.Form()
	sayHolding(k, k.skin.Colour.Ink.Faint, name, Written{Value: "steps"}, small)
	Field{}.Draw(k, frame)
	k = k.inside("field")
	parts, from := w.PartsTicked(), w.partsFrom()
	on := f.onParts() && w.screen == OnForm
	places := w.partTiles()
	for n, tile := range places[:len(places)-1] {
		at := from + n
		if on && f.writingPart && f.partAt == at {
			Ring{Target: true, Radius: k.skin.Radius.Row}.Draw(k.inside("tile"), Tile{}.Draw(k, SubPanel(tile)))
			writing(k, w.measured(), SubWriting(tile), f.named, true, small, 1, Follow)
			continue
		}
		if _, carried := w.PartGhost(); carried && f.carried-1 == at {
			origin := k.faded(OriginShows)
			drawSub(k, tile, parts[at], StepLine(parts[at]), false, false, Written{})
			origin.Pop()
			continue
		}
		drawSub(k, tile, parts[at], StepLine(parts[at]), on && f.partAt == at, on && slices.Contains(f.selectedParts(), at), Written{})
	}
	if field, ok := f.partsField(); ok {
		_, value, rows := FormPlace(w.formBox(), f.Fields(), field)
		_, bar := Saying(value, rows)
		Scrollbar{Lines: len(parts), Rows: w.partsShown(), At: from}.Draw(k, bar)
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
		seen := k.faded(w.GhostShows())
		Lifted{By: 1}.Draw(k, ghost)
		Tile{}.Draw(k, ghost)
		line := SubWriting(ghost)
		say(k.gtx, k.th, k.skin.Colour.Ink.Text, line, fits(k.gtx, k.th, StepLine(parts[f.carried-1]), line.W, small), small)
		seen.Pop()
	}
	if line, ok := w.PartLanding(); ok {
		Landing{}.Draw(k, line)
	}
}

func drawFieldAt(k Kit, m Measure, name, value Box, rows, size int, long, lit, carets bool, said, named Written, scrolled int) {
	frame := Framed(value, rows)
	var bar Box
	if long {
		frame, bar = Saying(value, rows)
	}
	ink := k.skin.Colour.Ink.Low
	if lit {
		ink = k.skin.Colour.Accent.Base
	}
	Field{Lit: lit}.Draw(k, frame)
	k = k.inside("field")
	if name.W > 0 {
		sayHolding(k, ink, name, named, small)
	}
	writing(k, m, Inside(frame), said, carets, size, rows, scrolled)
	if long {
		lines := wrapIn(m, said.Reads(), Inside(frame), size)
		line, _ := lines.At(said.At)
		Scrollbar{Lines: len(lines.Said), Rows: rows, At: StartsAt(lines, rows, line, scrolled)}.Draw(k, bar)
	}
}

func drawForm(k Kit, w Window) {
	panel, box := w.formPanel(), w.formBox()
	head, occurs := w.OccurrenceHeading()
	if occurs {
		Panel{}.Draw(k, panel)
	} else {
		Panel{Head: HeadsOn(w, OnForm)}.Draw(k, panel)
	}
	k = k.inside("panel")
	if occurs {
		loop, title, under := OccurrenceHeadAt(FormHeading(panel))
		LoopMark{Ink: k.skin.Colour.Ink.Text}.Draw(k, loop)
		Heading{Said: head.Title, Number: head.Number}.Draw(k, title)
		sayIn(k.gtx, k.th, k.skin.Colour.Ink.Faint, under, fits(k.gtx, k.th, head.Under, under.W, small), small, font.Font{}, text.End)
	}
	Cross{Ink: k.skin.Colour.Ink.Faint}.Draw(k, FormClose(panel))
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
		labels := strings.FieldsFunc(said.Reads(), func(r rune) bool { return r == ',' || r == ' ' })
		written := said
		pilled := fields[i].Arg == "label" && !on && len(labels) > 0
		if pilled {
			written = Written{}
		}
		drawFieldAt(k, w.measured(), name, value, rows, FormSize(fields[i]), fields[i].Long, on, typed, written, Written{Value: fields[i].Name}, scrolled)
		frame := Framed(value, rows)
		if _, known := k.skin.Colour.Priority[said.Reads()]; known && fields[i].Arg == "priority" {
			Stripe{Priority: said.Reads(), Free: true}.Draw(k.inside("field"), Box{X: FieldMark(frame).X - padding/2 - edgeWide, Y: frame.Y + (frame.H-stripeHigh)/2, W: edgeWide, H: stripeHigh})
		}
		if pilled {
			drawFieldPills(k.inside("field"), w, w.measured(), frame, labels)
		}
		switch {
		case on:
		case chosenOnACalendar(fields[i]):
			CalendarMark{Ink: k.skin.Colour.Ink.Low}.Draw(k.inside("field"), FieldMark(frame))
		case chosenFromAList(fields[i]), chosenFromRules(fields[i]):
			Chevron{Ink: k.skin.Colour.Ink.Low}.Draw(k.inside("field"), FieldMark(frame))
		}
	}
	if !w.Tracing() {
		for _, head := range FormHeads(box, fields) {
			say(k.gtx, k.th, k.skin.Colour.Ink.Faint, head.Box, head.Said, head.Size)
		}
		drawReadBack(k, w)
	}
	Grip{Ink: k.skin.Colour.Ink.Faint}.Draw(k, FormGrip(panel))
	if w.screen == OnForm {
		drawPanelOffers(k, panel, w)
		if note := w.FormNote(); note != "" && len(w.ContextSays()) == 0 {
			say(k.gtx, k.th, k.skin.Colour.Ink.Low, ContextName(panel), note, small)
		}
	}
	if tip, said, ok := w.Explanation(); ok {
		Popover{}.Draw(k.on(LayerPopover), tip)
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
	case w.form.ForSeries:
		Head{Said: DetailsHead}.Draw(k, BesideHead(panel))
	case w.besideShown():
		Head{Said: DetailsHead}.Draw(k, BesideHead(panel))
		drawUnder(k, w)
		drawProgressBeside(k, w)
	}
	if w.Tracing() {
		drawHistory(k, panel, w)
	}
	drawOccurrenceBeside(k, w)
	for _, r := range w.ContextSays() {
		sayHolding(k, k.skin.Colour.Ink.Faint, r.Box, w.Holding(r.Key, r.Said), r.Size)
	}
}

func drawPanelOffers(k Kit, box Box, w Window) {
	along := w.offersAlong(box)
	if len(along) > 0 {
		Rule{}.Draw(k, Box{X: box.X, Y: along[0].Y - between, W: box.W})
	}
	for i, offer := range w.Offers() {
		look := w.OfferLook(i, offer)
		if w.Lighting(offer) {
			look = litLook(look)
		}
		Button{Said: offer, Key: w.OfferKey(offer), Look: look, Keyed: true}.Draw(k, ButtonIn(along[i]))
	}
	if cross, ok := w.PanelClose(); ok && w.screen != OnForm {
		Cross{Ink: k.skin.Colour.Ink.Faint}.Draw(k, cross)
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

func ringed(k Kit, b Box, radius int, ink color.NRGBA) {
	line, r := k.skin.Size.Rule, min(radius, b.W/2, b.H/2)
	for _, edge := range []Box{{X: b.X + r, Y: b.Y, W: b.W - 2*r, H: line}, {X: b.X + r, Y: b.Y + b.H - line, W: b.W - 2*r, H: line}, {X: b.X, Y: b.Y + r, W: line, H: b.H - 2*r}, {X: b.X + b.W - line, Y: b.Y + r, W: line, H: b.H - 2*r}} {
		fill(k.gtx, edge, ink)
	}
	if r == 0 {
		return
	}
	right, foot := b.X+b.W-r, b.Y+b.H-r
	for _, corner := range []Box{{X: b.X, Y: b.Y, W: r, H: r}, {X: right, Y: b.Y, W: r, H: r}, {X: b.X, Y: foot, W: r, H: r}, {X: right, Y: foot, W: r, H: r}} {
		var path clip.Path
		path.Begin(k.gtx.Ops)
		around(&path, b, float32(r), false)
		around(&path, Box{X: b.X + line, Y: b.Y + line, W: b.W - 2*line, H: b.H - 2*line}, float32(max(0, r-line)), true)
		shape := clip.Outline{Path: path.End()}.Op()
		cut := clip.Rect(rect(corner)).Push(k.gtx.Ops)
		paint.FillShape(k.gtx.Ops, ink, shape)
		cut.Pop()
	}
}

func around(path *clip.Path, b Box, r float32, backwards bool) {
	const iq = 1 - 4*(math.Sqrt2-1)/3
	w, n, e, s := float32(b.X), float32(b.Y), float32(b.X+b.W), float32(b.Y+b.H)
	path.MoveTo(f32.Pt(w+r, n))
	if backwards {
		path.CubeTo(f32.Pt(w+r*iq, n), f32.Pt(w, n+r*iq), f32.Pt(w, n+r))
		path.LineTo(f32.Pt(w, s-r))
		path.CubeTo(f32.Pt(w, s-r*iq), f32.Pt(w+r*iq, s), f32.Pt(w+r, s))
		path.LineTo(f32.Pt(e-r, s))
		path.CubeTo(f32.Pt(e-r*iq, s), f32.Pt(e, s-r*iq), f32.Pt(e, s-r))
		path.LineTo(f32.Pt(e, n+r))
		path.CubeTo(f32.Pt(e, n+r*iq), f32.Pt(e-r*iq, n), f32.Pt(e-r, n))
		path.Close()
		return
	}
	path.LineTo(f32.Pt(e-r, n))
	path.CubeTo(f32.Pt(e-r*iq, n), f32.Pt(e, n+r*iq), f32.Pt(e, n+r))
	path.LineTo(f32.Pt(e, s-r))
	path.CubeTo(f32.Pt(e, s-r*iq), f32.Pt(e-r*iq, s), f32.Pt(e-r, s))
	path.LineTo(f32.Pt(w+r, s))
	path.CubeTo(f32.Pt(w+r*iq, s), f32.Pt(w, s-r*iq), f32.Pt(w, s-r))
	path.LineTo(f32.Pt(w, n+r))
	path.CubeTo(f32.Pt(w, n+r*iq), f32.Pt(w+r*iq, n), f32.Pt(w+r, n))
	path.Close()
}

func outline(gtx layout.Context, b Box, with color.NRGBA) {
	for _, edge := range Edges(b) {
		fill(gtx, edge, with)
	}
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
	Veil{Light: true}.Draw(k, Box{W: w.width, H: w.height})
	k = k.inside("veil")
	rise, _ := w.PaletteRise()
	risen := k.shifted(-int((1 - rise) * paletteRiseBy))
	defer risen.Pop()
	faded := k.faded(float32(rise))
	defer faded.Pop()
	box := w.PaletteBox()
	Panel{}.Draw(k, box)
	k = k.inside("panel")
	writing(k, w.measured(), PaletteTyped(box), w.palette.typed, true, reads, 1, Follow)
	shut := w.keyboard.Spelled(Pretty("esc"))
	closeAt := Box{X: box.X + box.W - padding - ends(w.measured(), "close", small), Y: PaletteTyped(box).Y, W: ends(w.measured(), "close", small), H: lineHigh}
	sayIn(k.gtx, k.th, k.skin.Colour.Ink.Low, closeAt, "close", small, font.Font{}, text.End)
	KeyChip{Key: shut}.Draw(k, Box{X: closeAt.X - hintSpace - KeyChipWide(shut), Y: closeAt.Y, H: lineHigh})
	Rule{}.Draw(k, Box{X: box.X, Y: box.Y + paletteHead - 1, W: box.W})
	in, clipped := k.within(Box{X: box.X, Y: box.Y + paletteHead, W: box.W, H: box.H - paletteHead})
	defer clipped.Pop()
	if at, ok := w.PaletteBannerAt(); ok {
		said, _ := w.PaletteBanner()
		first, then := paletteBannerSaid(said)
		Banner{Said: first, Then: then, Warn: true}.Draw(in, at)
	}
	rows := w.PaletteRows()
	for _, line := range w.PaletteLines() {
		if line.Row < 0 {
			Head{Said: line.Group}.Draw(in, Box{X: line.Box.X, Y: line.Box.Y + line.Box.H - lineHigh, W: line.Box.W, H: lineHigh})
			continue
		}
		row := Row{Picked: line.Row == w.palette.row}
		row.Draw(in, line.Box)
		sayIn(k.gtx, k.th, k.skin.Colour.Ink.Text, Box{X: line.Box.X + padding/2, Y: line.Box.Y + (line.Box.H-lineHigh)/2, W: line.Box.W - padding, H: lineHigh}, rows[line.Row].Says, reads, font.Font{}, text.Start)
		end := line.Box.X + line.Box.W - padding/2
		if row.Picked {
			enter := w.keyboard.Spelled(Pretty("enter"))
			chip := row.Chip(in, KeyChip{Key: enter}, PaletteKeyAt(line.Box, enter))
			end = chip.X - hintSpace
		}
		if key := rows[line.Row].Key; key != "" {
			spelled := w.keyboard.Spelled(Pretty(key))
			chip := row.Chip(in, KeyChip{Key: spelled, Quiet: true}, PaletteKeyAt(Box{X: line.Box.X, Y: line.Box.Y, W: end - line.Box.X + padding/2, H: line.Box.H}, spelled))
			end = chip.X - hintSpace
		}
		if ago := rows[line.Row].Ago(w.now); ago != "" {
			sayIn(k.gtx, k.th, k.skin.Colour.Ink.Faint, Box{X: line.Box.X, Y: line.Box.Y + (line.Box.H-lineHigh)/2, W: end - line.Box.X, H: lineHigh}, ago, small, font.Font{}, text.End)
		}
	}
}

func drawPills(k Kit, w Window) {
	for _, pill := range w.Pills() {
		pill.Draw(k, pill.Box)
	}
	drawNarrowing(k, w)
}

func drawFieldPills(k Kit, w Window, m Measure, frame Box, labels []string) {
	room := Inside(frame)
	room.W -= markWide + padding
	x, top := room.X, frame.Y+(frame.H-pillHigh)/2
	wide := func(said string) int { return ends(m, said, small) }
	shown, more := Pills(labels, room.W, wide)
	for _, label := range shown {
		pill := Box{X: x, Y: top, W: wide(label) + 2*pillPad, H: pillHigh}
		LabelChip{Label: label, Colour: w.LabelColour(label)}.Draw(k, pill)
		x += pill.W + pillGap
	}
	if more > 0 {
		count := "+" + strconv.Itoa(more)
		Pill{Said: count}.Draw(k, Box{X: x, Y: top, W: wide(count) + 2*pillPad, H: pillHigh})
	}
}
