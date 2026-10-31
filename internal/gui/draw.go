package gui

import (
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

	"cairn/internal/domain"
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

// Skin is every colour the window uses, in one place, so a second theme is a second Skin and not a search through the drawing.
type Skin struct {
	Back, Panel, Edge  color.NRGBA
	Ground, GroundEdge color.NRGBA
	Card, Hover, Field color.NRGBA
	Ink, Faint, Low    color.NRGBA
	Chosen, Selected   color.NRGBA
	Tint, Pill         color.NRGBA
	Wrong, Danger      color.NRGBA
	Late, Done         color.NRGBA
	Ghost              color.NRGBA
	Veil               color.NRGBA
	Priority           map[string]color.NRGBA
	Stones             []color.NRGBA
}

// Dark is the one skin there is.
func Dark() Skin {
	return Skin{
		Back:       color.NRGBA{R: 12, G: 13, B: 15, A: 255},
		Panel:      color.NRGBA{R: 38, G: 40, B: 45, A: 255},
		Edge:       color.NRGBA{R: 96, G: 98, B: 106, A: 255},
		Ground:     color.NRGBA{R: 23, G: 24, B: 27, A: 255},
		GroundEdge: color.NRGBA{R: 40, G: 41, B: 46, A: 255},
		Card:       color.NRGBA{R: 31, G: 32, B: 36, A: 255},
		Hover:      color.NRGBA{R: 37, G: 38, B: 43, A: 255},
		Field:      color.NRGBA{R: 19, G: 20, B: 23, A: 255},
		Ink:        color.NRGBA{R: 231, G: 231, B: 234, A: 255},
		Faint:      color.NRGBA{R: 154, G: 155, B: 162, A: 255},
		Low:        color.NRGBA{R: 95, G: 96, B: 103, A: 255},
		Chosen:     color.NRGBA{R: 123, G: 147, B: 255, A: 255},
		Selected:   color.NRGBA{R: 141, G: 143, B: 156, A: 255},
		Tint:       color.NRGBA{R: 30, G: 36, B: 60, A: 255},
		Pill:       color.NRGBA{R: 48, G: 50, B: 56, A: 255},
		Wrong:      color.NRGBA{R: 255, G: 155, B: 152, A: 255},
		Danger:     color.NRGBA{R: 58, G: 31, B: 34, A: 255},
		Late:       color.NRGBA{R: 240, G: 186, B: 96, A: 255},
		Done:       color.NRGBA{R: 127, G: 182, B: 159, A: 255},
		Ghost:      color.NRGBA{R: 42, G: 51, B: 80, A: 255},
		Veil:       color.NRGBA{R: 6, G: 7, B: 9, A: 158},
		Stones: []color.NRGBA{
			{R: 86, G: 84, B: 92, A: 255},
			{R: 104, G: 101, B: 108, A: 255},
			{R: 122, G: 119, B: 124, A: 255},
		},
		Priority: map[string]color.NRGBA{
			string(domain.Critical): {R: 229, G: 96, B: 94, A: 255},
			string(domain.High):     {R: 229, G: 162, B: 74, A: 255},
			string(domain.Normal):   {R: 79, G: 174, B: 140, A: 255},
			string(domain.Low):      {R: 125, G: 130, B: 150, A: 255},
		},
	}
}

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
func Draw(gtx layout.Context, th *material.Theme, skin Skin, w Window) {
	scaled := inDp(gtx)
	defer scaled.Pop()

	gtx.Metric = unit.Metric{PxPerDp: 1, PxPerSp: 1}

	fill(gtx, Box{W: w.width, H: w.height}, skin.Back)
	if w.Screen() == TooSmallToDraw {
		say(gtx, th, skin.Ink, Message(w.width, w.height), w.Says(), heading)
		return
	}
	if w.Blank() {
		return
	}
	l := w.Layout()
	drawTabs(gtx, th, skin, l, w)
	if !w.native {
		for _, stroke := range MenuStrokes(MenuButton(l)) {
			fill(gtx, stroke, skin.Chosen)
		}
	}
	if w.Tab() == Objectives {
		drawGoals(gtx, th, skin, l, w)
	} else if w.Tab() == Time {
		drawWeek(gtx, th, skin, l, w)
	} else {
		drawColumns(gtx, th, skin, l, w)
		drawBand(gtx, th, skin, l, w)
		drawDragging(gtx, th, skin, w)
	}
	drawBar(gtx, th, skin, l, w)
	if said := w.FilteredBy(); said != "" && w.Tab() != Objectives {
		say(gtx, th, skin.Faint, FilteredSaid(l), "filtered by "+said, reads)
	}
	if w.Screen() == OnSearch {
		drawSearch(gtx, th, skin, l, w)
	}
	if w.formUnder() {
		drawForm(gtx, th, skin, w)
	}
	if w.Screen() == OnHelp {
		drawHelp(gtx, th, skin, w)
	}
	if w.Screen() == OnAsking {
		fill(gtx, Box{W: w.width, H: w.height}, skin.Veil)
		asked := w.QuestionBox()
		drawSaid(gtx, th, skin, w, w.Asks(), asked)
		outline(gtx, asked, skin.Chosen)
		outline(gtx, Box{X: asked.X + 1, Y: asked.Y + 1, W: asked.W - 2, H: asked.H - 2}, skin.Chosen)
	}
	if w.Screen() == OnChoosing {
		drawChoosing(gtx, th, skin, w)
	}
	if w.Screen() == OnCalendar {
		drawCalendar(gtx, th, skin, w)
	}
	if w.Screen() == OnMenu {
		drawMenu(gtx, th, skin, w)
	}
	if w.Screen() == OnBackups {
		drawBackups(gtx, th, skin, w)
	}
	if w.Screen() == OnLabels {
		drawLabels(gtx, th, skin, w)
	}
	if w.Screen() == OnFiltering {
		drawFilters(gtx, th, skin, w)
	}
	if w.Screen() == OnConfiguration {
		drawConfiguration(gtx, th, skin, w)
	}
	if bubble, said, ok := w.BubbleBox(); ok {
		raised(gtx, bubble, bubbleRadius, skin.Panel)
		for i, line := range said {
			say(gtx, th, skin.Ink, ExplainedLine(bubble, i), line, plain)
		}
	}
	if bubble, pills, ok := w.LabelsBubbleBox(); ok {
		raised(gtx, bubble, bubbleRadius, skin.Panel)
		for i, pill := range pills {
			drawPill(gtx, th, skin, pill, w.bubbled[i], skin.Pill)
		}
	}
}

func drawBackups(gtx layout.Context, th *material.Theme, skin Skin, w Window) {
	box := w.BackupsBox()
	drawPanel(gtx, th, skin, box, Heads(w))
	from := w.BackupsFrom()
	for i, one := range w.Backups()[from:min(len(w.Backups()), from+w.BackupsRoom())] {
		row := ListRowAt(box, i)
		ink := skin.Ink
		if from+i == w.BackupOn() {
			fill(gtx, row, skin.Back)
			ink = skin.Chosen
		}
		say(gtx, th, ink, Inside(row), backupSays(one), reads)
	}
	drawPanelOffers(gtx, th, skin, box, w)
}

func drawLabels(gtx layout.Context, th *material.Theme, skin Skin, w Window) {
	box := w.LabelsBox()
	drawPanel(gtx, th, skin, box, Heads(w))
	from := w.LabelsFrom()
	for i, one := range w.LabelsCarried()[from:min(len(w.LabelsCarried()), from+w.LabelsRoom())] {
		row := ListRowAt(box, i)
		ink := skin.Ink
		if from+i == w.LabelOn() {
			fill(gtx, row, skin.Back)
			ink = skin.Chosen
		}
		say(gtx, th, ink, Inside(row), LabelSays(one), reads)
	}
	drawPanelOffers(gtx, th, skin, box, w)
}

func drawConfiguration(gtx layout.Context, th *material.Theme, skin Skin, w Window) {
	box := w.ConfigurationBox()
	drawPanel(gtx, th, skin, box, Heads(w))
	for i, line := range w.Configured() {
		ink := skin.Ink
		if line.Key == "" {
			ink = skin.Faint
		}
		say(gtx, th, ink, Inside(ListRowAt(box, i)), ConfigurationSays(line), reads)
	}
	drawPanelOffers(gtx, th, skin, box, w)
}

func drawFilters(gtx layout.Context, th *material.Theme, skin Skin, w Window) {
	box := w.FiltersBox()
	drawPanel(gtx, th, skin, box, Heads(w))
	for i, f := range w.Filters() {
		row := ListRowAt(box, i)
		ink := skin.Ink
		if i == w.FilterOn() {
			fill(gtx, row, skin.Back)
			ink = skin.Chosen
		}
		say(gtx, th, ink, Inside(row), FilterSays(f), reads)
	}
	drawPanelOffers(gtx, th, skin, box, w)
}

func drawWeek(gtx layout.Context, th *material.Theme, skin Skin, l Layout, w Window) {
	measuring := gtx
	measuring.Constraints.Min = image.Point{}
	for _, chip := range TimeChips {
		drawButton(gtx, th, skin, TimeChip(l, chip), chip, lookOf(chip == w.ranging()))
	}
	if w.Monthly() {
		drawButton(gtx, th, skin, TimeChip(l, "the month"), w.week.Month().String(), lookOf(w.grid == viewMonths))
		drawButton(gtx, th, skin, TimeChip(l, "the year"), strconv.Itoa(w.week.Year()), lookOf(w.grid == viewYears))
	} else {
		say(gtx, th, skin.Ink, Inside(TimeSaid(l)), TimeSays(w), reads)
	}
	if w.Monthly() {
		drawMonth(gtx, th, skin, l, w)
		return
	}
	shown := w.TimeShown()
	longest := shown.Longest()
	if w.Daily() {
		drawDayAlone(gtx, th, skin, l, w, longest)
		return
	}
	for i, day := range w.WeekDays() {
		box := WeekDayAt(l, i)
		fill(gtx, box, skin.Panel)
		say(gtx, th, skin.Faint, WeekHeadAt(box), day.Format("Mon 02/01")+"   "+w.DayHeld(i).Says(), reads)
		clipped := clip.Rect(rect(box)).Push(gtx.Ops)
		for j, one := range w.DayShown(i) {
			row := WeekRowAt(box, j)
			drawCard(gtx, th, skin, DayCardAt(box, j), w.DayCard(one), w.Chosen() == one.Number && !w.chosenGoal, DayCardAt(box, j) == w.pointed)
			say(gtx, th, skin.Chosen, DaySpentAt(row), one.Spent(), small)
			fill(gtx, WeekBar(row, one.Minutes, longest), skin.Chosen)
		}
		clipped.Pop()
		if track, more := DayTrack(box, len(w.DayHeld(i).Tasks)); more {
			fill(gtx, track, skin.Back)
			fill(gtx, Thumb(track, len(w.DayHeld(i).Tasks), DayRoom(box), w.DayFrom(i)), skin.Edge)
		}
	}
}

func drawDayAlone(gtx layout.Context, th *material.Theme, skin Skin, l Layout, w Window, longest int) {
	box := DayAloneAt(l)
	held := w.DayHeld(0)
	fill(gtx, box, skin.Panel)
	say(gtx, th, skin.Faint, WeekHeadAt(box), held.Says(), reads)
	clipped := clip.Rect(rect(box)).Push(gtx.Ops)
	for j, one := range w.AloneShown() {
		drawCard(gtx, th, skin, AloneCardAt(box, j), w.DayCard(one), w.Chosen() == one.Number && !w.chosenGoal, AloneCardAt(box, j) == w.pointed)
		say(gtx, th, skin.Chosen, AloneSpentAt(box, j), one.Spent(), small)
		fill(gtx, AloneBar(box, j, one.Minutes, longest), skin.Chosen)
	}
	clipped.Pop()
	if track, more := DayTrack(box, len(held.Tasks)); more {
		fill(gtx, track, skin.Back)
		fill(gtx, Thumb(track, len(held.Tasks), DayRoom(box), w.DayFrom(0)), skin.Edge)
	}
}

func drawMonth(gtx layout.Context, th *material.Theme, skin Skin, l Layout, w Window) {
	if cells := w.TimeCells(); len(cells) > 0 {
		for i, cell := range cells {
			drawButton(gtx, th, skin, TimeCellAt(l, i), cell, lookOf(cell == gridMarked(w.grid, w.week)))
		}
		return
	}
	for i, name := range Weekdays[:w.pages.WeekDays()] {
		say(gtx, th, skin.Faint, MonthWeekdayAt(l, i), name, reads)
	}
	for i, day := range engine.MonthGrid(w.week) {
		cell := MonthCellAt(l, i)
		if cell == (Box{}) {
			continue
		}
		fill(gtx, cell, skin.Panel)
		if day.Equal(w.today()) {
			outline(gtx, cell, skin.Chosen)
		}
		ink := skin.Ink
		if !sameMonth(day, w.week) {
			ink = skin.Faint
		}
		held := w.MonthDay(i)
		say(gtx, th, ink, CellHeadAt(cell), CellHead(day, held), reads)
		clipped := clip.Rect(rect(cell)).Push(gtx.Ops)
		shown := cellShown(held.Tasks, CellRoom(cell))
		for j, line := range CellLines(held.Tasks, CellRoom(cell)) {
			at := CellLineAt(cell, j)
			if j < shown {
				c := w.DayCard(held.Tasks[j])
				fill(gtx, Box{X: at.X - inset, Y: at.Y, W: edgeWide, H: at.H}, skin.edgeOf(c))
				if w.Chosen() == c.Number && !w.chosenGoal {
					outline(gtx, at, skin.Chosen)
				}
				say(gtx, th, skin.Writing(c), Box{X: at.X + edgeWide, Y: at.Y, W: at.W - edgeWide, H: at.H}, line, small)
				continue
			}
			say(gtx, th, skin.Faint, at, line, small)
		}
		clipped.Pop()
	}
}

func drawMenu(gtx layout.Context, th *material.Theme, skin Skin, w Window) {
	box := w.MenuBox()
	fill(gtx, box, skin.Panel)
	outline(gtx, box, skin.Edge)
	for i, said := range w.Menu() {
		row := MenuRowAt(box, i)
		ink := skin.Ink
		if i == w.MenuOn() {
			fill(gtx, row, skin.Back)
			ink = skin.Chosen
		}
		say(gtx, th, ink, Inside(row), said, reads)
	}
}

func drawCalendar(gtx layout.Context, th *material.Theme, skin Skin, w Window) {
	box := w.CalendarBox()
	drawPanelBare(gtx, skin, box)
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
		drawButton(gtx, th, skin, chip.at, chip.says, lookOf(chip.on))
	}
	if cells := w.CalendarCells(); len(cells) > 0 {
		for i, cell := range cells {
			drawButton(gtx, th, skin, CalendarCell(box, i), cell, lookOf(cell == w.CalendarMarked()))
		}
		drawPanelOffers(gtx, th, skin, box, w)
		return
	}
	for i, name := range Weekdays {
		say(gtx, th, skin.Faint, WeekdayAt(box, i), name, reads)
	}
	today := w.today()
	for i, day := range engine.MonthGrid(w.day) {
		at, said := DayAt(box, i), strconv.Itoa(day.Day())
		switch {
		case day.Equal(w.day):
			drawButton(gtx, th, skin, at, said, buttonOn)
		case day.Equal(today):
			rounded(gtx, at, buttonRadius, skin.Selected)
			rounded(gtx, Box{X: at.X + 1, Y: at.Y + 1, W: at.W - 2, H: at.H - 2}, buttonRadius-1, skin.Panel)
			sayCentred(gtx, th, skin.Ink, ButtonSaid(at), said, reads)
		default:
			ink := skin.Ink
			if w.Faded(day) {
				ink = skin.Faint
			}
			sayCentred(gtx, th, ink, ButtonSaid(at), said, reads)
		}
	}
	drawPanelOffers(gtx, th, skin, box, w)
}

// TickedMark says an objective is in.
const TickedMark = "✓"

func drawChoosing(gtx layout.Context, th *material.Theme, skin Skin, w Window) {
	box := w.ChoosingBox()
	drawPanelBare(gtx, skin, box)

	if !box.Unsearched {
		looking := ChoosingLooking(box)
		fill(gtx, looking, skin.Back)
		outline(gtx, looking, skin.Edge)
		writing(gtx, th, skin, w.measured(), Inside(looking), w.Filtering(), true, reads, 1, Follow)
	}

	at := ChoosingFrom(box)
	offered := w.Choosing()
	if len(offered) == 0 {
		say(gtx, th, skin.Faint, ChoosingLine(box, at), w.EmptyListSays(), reads)
		drawPanelOffers(gtx, th, skin, box, w)
		return
	}
	run := ChoosingRun(box)
	clipped := clip.Rect(rect(run)).Push(gtx.Ops)
	from := w.ChoosingFromRow()
	for i := from; i < Lasting(len(offered), from, w.ChoosingRoom()); i++ {
		one := offered[i]
		row := ChoosingRow(box, at)
		if i == w.ChoosingAt() {
			fill(gtx, row, skin.Back)
			outline(gtx, row, skin.Chosen)
		}
		line := ChoosingLine(box, at)
		say(gtx, th, skin.Ink, line, w.RowSays(one), reads)
		at += apart
	}
	clipped.Pop()

	track := ChoosingBar(box)
	fill(gtx, track, skin.Back)
	fill(gtx, Thumb(track, len(offered), w.ChoosingRoom(), from), skin.Edge)
	drawPanelOffers(gtx, th, skin, box, w)
}

func drawSearch(gtx layout.Context, th *material.Theme, skin Skin, l Layout, w Window) {
	box := Searching(l)
	fill(gtx, box, skin.Panel)
	outline(gtx, box, skin.Chosen)
	say(gtx, th, skin.Faint, SearchWord(box), "SEARCH", small)
	into := SearchInto(box)
	writing(gtx, th, skin, w.measured(), into, w.Searched(), true, reads, 1, Follow)
}

func drawSaid(gtx layout.Context, th *material.Theme, skin Skin, w Window, said []Said, box Box) {
	drawPanel(gtx, th, skin, box, Heads(w))
	at := FirstRow(box)
	for _, line := range said {
		if line.Key == "" {
			say(gtx, th, skin.Ink, Row(box, at, lineHigh), line.Means, reads)
			at += apart
			continue
		}
		name, value := FieldRow(box, at)
		if line.Key != "" {
			fill(gtx, Chip(name), skin.Back)
		}
		say(gtx, th, skin.Chosen, name, line.Key, reads)
		say(gtx, th, skin.Ink, value, line.Means, reads)
		at += apart
	}
	drawPanelOffers(gtx, th, skin, box, w)
}

func drawHelp(gtx layout.Context, th *material.Theme, skin Skin, w Window) {
	box, rows, from := w.HelpBox(), w.HelpRows(), w.HelpFrom()
	drawPanel(gtx, th, skin, box, Heads(w))
	for i := range Lasting(len(rows), from, w.HelpRoom()) - from {
		line := rows[from+i]
		name, value := HelpRowAt(box, i)
		if line.Key != "" {
			fill(gtx, Chip(name), skin.Back)
			say(gtx, th, skin.Chosen, name, line.Key, reads)
		}
		say(gtx, th, skin.Ink, value, line.Means, reads)
	}
	if bar, ok := w.HelpBar(); ok {
		fill(gtx, bar, skin.Back)
		fill(gtx, Thumb(bar, len(rows), w.HelpRoom(), from), skin.Edge)
	}
	drawPanelOffers(gtx, th, skin, box, w)
}

func drawPanelBare(gtx layout.Context, skin Skin, box Box) {
	raised(gtx, box, panelRadius, skin.Panel)
}

const panelRadius = 12

func drawPanel(gtx layout.Context, th *material.Theme, skin Skin, box Box, head string) {
	drawPanelHeaded(gtx, th, skin, box, FormHeading(box), head)
}

func drawPanelHeaded(gtx layout.Context, th *material.Theme, skin Skin, box, heading Box, head string) {
	drawPanelBare(gtx, skin, box)
	sayStyled(gtx, th, skin.Ink, heading, head, HeadingSize, HeadingWeight, 1, text.Start)
}

func drawTabs(gtx layout.Context, th *material.Theme, skin Skin, l Layout, w Window) {
	for i, page := range w.Shown() {
		tab := TabAt(l, i)
		ink := skin.Faint
		if w.Tab() == page {
			rounded(gtx, tab, tabRadius, skin.Pill)
			ink = skin.Ink
		}
		room, name := TabName(tab), Titled(Pages[page])
		sayStyled(gtx, th, ink, room, name, reads, font.Medium, 1, text.Start)
		counted := ends(w.measured(), name, reads) + countGap
		sayIn(gtx, th, skin.Low, Box{X: room.X + counted, Y: room.Y, W: max(0, room.W-counted), H: room.H}, strconv.Itoa(w.Holds(page)), small, font.Font{Typeface: MonoFace}, text.Start)
	}
}

const (
	tabRadius = 6
	countGap  = 6
)

func drawColumns(gtx layout.Context, th *material.Theme, skin Skin, l Layout, w Window) {
	dealt, _ := engine.Deal(w.Showing().Cards, w.archived)
	for i, column := range dealt {
		ground(gtx, skin, Box{X: l.Head[i].X, Y: l.Head[i].Y, W: l.Head[i].W, H: l.Head[i].H + l.Column[i].H})
		drawHead(gtx, th, skin, NameIn(l.Head[i]), Titled(column.Name), len(column.Cards))
		if said := w.EmptySays(i); said != "" {
			drawEmpty(gtx, th, skin, EmptyAt(l.CardAt(i, 0, 0)), said)
		}
	}
}

func drawBand(gtx layout.Context, th *material.Theme, skin Skin, l Layout, w Window) {
	for lane, kind := range engine.Stalled {
		ground(gtx, skin, LaneGround(l, lane))
		held := len(w.inLane(lane))
		drawHead(gtx, th, skin, LaneName(l, lane), Titled(kind), held)
		if held == 0 {
			drawEmpty(gtx, th, skin, EmptyAt(l.LaneCardAt(lane, 0, 0)), nothingHere)
		}
	}
	drawPlaced(gtx, th, skin, w)
}

func drawHead(gtx layout.Context, th *material.Theme, skin Skin, room Box, name string, held int) {
	sayStyled(gtx, th, skin.Faint, room, name, small, font.Medium, 1, text.Start)
	sayIn(gtx, th, skin.Low, room, strconv.Itoa(held), small, font.Font{Typeface: MonoFace}, text.End)
}

func drawEmpty(gtx layout.Context, th *material.Theme, skin Skin, at Box, said string) {
	dashed(gtx, at, skin.Edge)
	sayCentred(gtx, th, skin.Low, Box{X: at.X, Y: at.Y + (at.H-lineHigh)/2, W: at.W, H: lineHigh}, said, small)
}

func drawPlaced(gtx layout.Context, th *material.Theme, skin Skin, w Window) {
	progress, building := w.Intro()
	var seen []Placed
	for _, p := range w.Placed() {
		if p.seen() {
			seen = append(seen, p)
		}
	}
	for k, p := range seen {
		chosen := p.Card.Number == w.Chosen() || w.isTaken(p.Card.Number)
		clipped := clip.Rect(rect(p.Within)).Push(gtx.Ops)
		if !building {
			if Dimmed(p.Card) {
				half := paint.PushOpacity(gtx.Ops, 0.5)
				drawCard(gtx, th, skin, p.Slot, p.Card, chosen, p.Slot == w.pointed)
				half.Pop()
			} else {
				drawCard(gtx, th, skin, p.Slot, p.Card, chosen, p.Slot == w.pointed)
			}
			clipped.Pop()
			continue
		}
		_, fade := Arriving(progress, k, len(seen))
		faded := paint.PushOpacity(gtx.Ops, float32(fade))
		drawCard(gtx, th, skin, p.Slot, p.Card, chosen, p.Slot == w.pointed)
		faded.Pop()
		clipped.Pop()
	}
	if !building {
		return
	}
	for k, p := range seen {
		fall, fade := Arriving(progress, k, len(seen))
		panel := PartsOf(p.Slot).Panel
		faded := paint.PushOpacity(gtx.Ops, float32(1-fade))
		for i, stone := range StonesOf(panel, p.Card.Number) {
			drawStone(gtx, Fallen(stone, Dropped(panel, fall)), skin.stoneOf(i))
		}
		faded.Pop()
	}
}

func drawStone(gtx layout.Context, stone []Point, ink color.NRGBA) {
	var path clip.Path
	path.Begin(gtx.Ops)
	for i, p := range stone {
		at := f32.Pt(float32(p.X), float32(p.Y))
		if i == 0 {
			path.MoveTo(at)
			continue
		}
		path.LineTo(at)
	}
	path.Close()
	paint.FillShape(gtx.Ops, ink, clip.Outline{Path: path.End()}.Op())
}

func drawCross(gtx layout.Context, cross Box, ink color.NRGBA) {
	reach := float32(cross.W) / 4
	x, y := float32(cross.X)+float32(cross.W)/2, float32(cross.Y)+float32(cross.H)/2
	var strokes clip.Path
	strokes.Begin(gtx.Ops)
	strokes.MoveTo(f32.Pt(x-reach, y-reach))
	strokes.LineTo(f32.Pt(x+reach, y+reach))
	strokes.MoveTo(f32.Pt(x+reach, y-reach))
	strokes.LineTo(f32.Pt(x-reach, y+reach))
	paint.FillShape(gtx.Ops, ink, clip.Stroke{Path: strokes.End(), Width: 1.5}.Op())
}

func drawCairn(gtx layout.Context, mark Box, ink color.NRGBA) {
	scale := float32(mark.W) / 14
	for _, stone := range [][4]float32{{7, 11.4, 5.6, 2.1}, {7.3, 7.3, 4, 1.8}, {6.8, 3.6, 2.5, 1.5}} {
		x, y := float32(mark.X)+stone[0]*scale, float32(mark.Y)+stone[1]*scale
		rx, ry := stone[2]*scale, stone[3]*scale
		oval := clip.Ellipse{Min: image.Pt(int(x-rx), int(y-ry)), Max: image.Pt(int(x+rx+0.5), int(y+ry+0.5))}
		paint.FillShape(gtx.Ops, ink, oval.Op(gtx.Ops))
	}
}

func drawCard(gtx layout.Context, th *material.Theme, skin Skin, box Box, c engine.Card, chosen, hovered bool) {
	parts := PartsOf(box)
	ground := skin.Card
	switch {
	case c.InFocus:
		ground = skin.Tint
	case hovered:
		ground = skin.Hover
	}
	if chosen {
		rounded(gtx, parts.Panel, cardRadius, skin.Selected)
		rounded(gtx, Box{X: parts.Panel.X + 1, Y: parts.Panel.Y + 1, W: parts.Panel.W - 2, H: parts.Panel.H - 2}, cardRadius-1, ground)
	} else {
		rounded(gtx, parts.Panel, cardRadius, ground)
	}
	stripe := clip.RRect{Rect: rect(parts.Edge), NE: stripeRadius, SE: stripeRadius}
	paint.FillShape(gtx.Ops, skin.edgeOf(c), stripe.Op(gtx.Ops))

	m := Measured(th)
	wideAt := func(size int) func(string) int { return func(said string) int { return ends(m, said, size) } }
	lines := TitleLines(c.Title, TitleRoom(c, parts.Title), wideAt(reads))
	for i, line := range lines {
		sayStyled(gtx, th, skin.Writing(c), Box{X: parts.Title.X, Y: parts.Title.Y + i*lineHigh, W: parts.Title.W, H: lineHigh}, line, reads, font.Medium, 1, text.Start)
	}
	if mark, ok := FocusMark(c, parts.Title, len(lines), ends(m, lines[len(lines)-1], reads)); ok {
		drawCairn(gtx, mark, skin.Chosen)
	}

	mono := font.Font{Typeface: MonoFace}
	laid := CardRowOf(c, parts.Number, m, gtx.Now)
	sayIn(gtx, th, skin.Low, parts.Number, laid.Number, small, mono, text.Start)
	if laid.Facts != "" {
		ink := skin.Low
		if laid.Late {
			ink = skin.Late
		}
		sayIn(gtx, th, ink, laid.FactsAt, laid.Facts, small, mono, text.End)
	}
	for i, pill := range laid.Pills {
		drawPill(gtx, th, skin, pill, laid.Shown[i], skin.Pill)
	}
	if laid.Count != "" {
		rounded(gtx, laid.More, pillHigh/2, skin.Edge)
		drawPill(gtx, th, skin, Box{X: laid.More.X + 1, Y: laid.More.Y + 1, W: laid.More.W - 2, H: laid.More.H - 2}, laid.Count, ground)
	}

	if c.SubTotal > 0 {
		done := strconv.Itoa(c.SubDone) + "/" + strconv.Itoa(c.SubTotal)
		counted := monoWide(done, small)
		track := Box{X: parts.Steps.X, Y: parts.Steps.Y + (parts.Steps.H-trackHigh)/2, W: parts.Steps.W - counted - pillGap*2, H: trackHigh}
		rounded(gtx, track, trackHigh, skin.Pill)
		if c.SubDone > 0 {
			rounded(gtx, Box{X: track.X, Y: track.Y, W: max(trackHigh, track.W*c.SubDone/c.SubTotal), H: track.H}, trackHigh, skin.Done)
		}
		sayIn(gtx, th, skin.Low, Box{X: parts.Steps.X + parts.Steps.W - counted, Y: parts.Steps.Y + (parts.Steps.H-lineHigh)/2, W: counted, H: lineHigh}, done, small, mono, text.End)
	}
}

const (
	bubbleRadius = 8
	stripeRadius = 3
	trackHigh    = 3
)

func drawPill(gtx layout.Context, th *material.Theme, skin Skin, pill Box, said string, ground color.NRGBA) {
	rounded(gtx, pill, pill.H/2, ground)
	sayCentred(gtx, th, skin.Faint, Box{X: pill.X, Y: pill.Y + (pill.H-lineHigh)/2, W: pill.W, H: lineHigh}, said, small)
}

// Writing is the ink a card's title is written in.
func (s Skin) Writing(c engine.Card) color.NRGBA {
	if c.Archived {
		return s.Faint
	}
	return s.Ink
}

func (s Skin) stoneOf(i int) color.NRGBA { return s.Stones[i%len(s.Stones)] }

func (s Skin) edgeOf(c engine.Card) color.NRGBA {
	if edge, known := s.Priority[c.Priority]; known && !c.Archived {
		return edge
	}
	return s.Edge
}

func drawGoals(gtx layout.Context, th *material.Theme, skin Skin, l Layout, w Window) {
	clipped := clip.Rect(rect(GoalsRun(l))).Push(gtx.Ops)
	defer clipped.Pop()
	for at, g := range engine.Sorted(w.Showing().Goals) {
		box := GoalAt(l, at, w.goals)
		drawCard(gtx, th, skin, box, goalFace(g), g.Number == w.Chosen() || w.isTaken(g.Number), box == w.pointed)
		if said := Counted(g); said != "" {
			say(gtx, th, skin.Faint, CountedAt(PartsOf(box).Number, said), said, small)
		}
	}
}

func goalFace(g engine.Goal) engine.Card {
	return engine.Card{Number: g.Number, Title: g.Title, State: g.State, Priority: g.Priority, Labels: Standings(g)}
}

func drawDragging(gtx layout.Context, th *material.Theme, skin Skin, w Window) {
	if landing, somewhere := w.Landing(); somewhere {
		outline(gtx, PartsOf(landing).Panel, skin.Chosen)
	}
	ghost, dragging := w.Ghost()
	if !dragging {
		return
	}
	fill(gtx, ghost, skin.Ghost)
	outline(gtx, ghost, skin.Chosen)
	for _, c := range w.Reading().Cards {
		if c.Number != w.held.Number {
			continue
		}
		line := PartsOf(ghost).Title
		say(gtx, th, skin.Ink, line, fits(gtx, th, c.Title, line.W, plain), plain)
	}
}

func drawBar(gtx layout.Context, th *material.Theme, skin Skin, l Layout, w Window) {
	if w.Wrong() != nil {
		fill(gtx, WrongBox(l), skin.Back)
		sayOver(gtx, th, skin.Wrong, WrongBox(l), w.Wrong().Error(), reads, 2)
	} else if w.Note() != "" {
		fill(gtx, WrongBox(l), skin.Back)
		sayOver(gtx, th, skin.Ink, WrongBox(l), w.Note(), reads, 2)
	}
	if s := w.Screen(); s != OnTheBoard && s != OnSearch {
		under := w.barAlong(w.BoardOffers())
		for i, offer := range w.BoardOffers() {
			drawButton(gtx, th, skin, Button(under[i]), w.Said(offer), buttonFaint)
		}
		return
	}
	bar := w.BarOffers()
	for i, offer := range w.Offers() {
		drawOffer(gtx, th, skin, bar[i], offer, w)
	}
}

func drawOffer(gtx layout.Context, th *material.Theme, skin Skin, box Box, offer string, w Window) {
	look := buttonPlain
	if w.Lighting(offer) {
		look = buttonLit
	}
	drawButton(gtx, th, skin, Button(box), w.Said(offer), look)
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

const buttonRadius = 7

func drawButton(gtx layout.Context, th *material.Theme, skin Skin, button Box, said string, look buttonLook) {
	back, ink := skin.Pill, skin.Ink
	switch look {
	case buttonOn:
		back, ink = skin.Chosen, skin.Back
	case buttonOnLit:
		back, ink = lighter(skin.Chosen), skin.Back
	case buttonLit:
		back = lighter(skin.Pill)
	case buttonFaint:
		ink = skin.Faint
	case buttonDanger:
		back, ink = skin.Danger, skin.Wrong
	case buttonDangerLit:
		back, ink = lighter(skin.Danger), skin.Wrong
	}
	rounded(gtx, button, buttonRadius, back)
	sayStyled(gtx, th, ink, ButtonSaid(button), said, reads, font.Medium, 1, text.Middle)
}

func drawSub(gtx layout.Context, th *material.Theme, skin Skin, box Box, c engine.Card, line string, on, picked bool, held Written) {
	panel := SubPanel(box)
	ground := skin.Panel
	if picked {
		ground = skin.Ghost
	}
	fill(gtx, panel, ground)
	if on {
		outline(gtx, panel, skin.Chosen)
	}
	check := SubCheck(box)
	done := engine.Finished(c)
	if done {
		fill(gtx, check, skin.Faint)
		say(gtx, th, skin.Panel, TickAt(check), TickedMark, small)
	} else {
		outline(gtx, check, skin.Ink)
	}
	writ := SubWriting(box)
	said := fits(gtx, th, line, writ.W, small)
	held.Value = said
	sayHolding(gtx, th, skin, SubInk(skin, c), writ, held, small)
	if done {
		loose := gtx
		loose.Constraints = layout.Constraints{Max: image.Pt(writ.W, lineHigh)}
		fill(gtx, StruckAt(writ, wide(loose, th, said, small).Size.X), skin.Faint)
	}
}

// SubInk is the ink a step's name, or a serving task's, is written in: the reading ink for what is left, the faint one for what is done.
func SubInk(skin Skin, c engine.Card) color.NRGBA {
	if engine.Finished(c) {
		return skin.Faint
	}
	return skin.Ink
}

func drawUnder(gtx layout.Context, th *material.Theme, skin Skin, box Box, w Window) {
	sayHolding(gtx, th, skin, skin.Faint, BesideHead(box), w.Holding("under", w.UnderHead()), small)
	frame := BesideBox(box)
	fill(gtx, frame, skin.Back)
	outline(gtx, frame, skin.Edge)

	under := w.Under()
	if len(under) == 0 {
		if w.ChosenGoal() {
			say(gtx, th, skin.Faint, LineAt(Inside(frame), 0, 0), ServingSays, small)
		}
		return
	}

	clipped := clip.Rect(rect(frame)).Push(gtx.Ops)
	for at, c := range under {
		drawSub(gtx, th, skin, w.ServingCardAt(at), c, SubLine(c), false, false,
			w.Holding("part:"+strconv.Itoa(c.Number), SubLine(c)))
	}
	clipped.Pop()

	track := BesideBar(box)
	fill(gtx, track, skin.Back)
	fill(gtx, Thumb(track, len(under)*SubHeight, frame.H, w.ServingAt()), skin.Edge)
}

func drawHistory(gtx layout.Context, th *material.Theme, skin Skin, box Box, w Window) {
	sayHolding(gtx, th, skin, skin.Faint, BesideHead(box), w.Holding("history", w.PanelHead()), small)
	frame := BesideBox(box)
	fill(gtx, frame, skin.Back)
	outline(gtx, frame, skin.Edge)

	inside := Inside(frame)
	clipped := clip.Rect(rect(inside)).Push(gtx.Ops)
	told, from, rows := w.traced(), w.History(), w.Rows()
	for i := from; i < Lasting(len(told.Said), from, rows); i++ {
		drawTold(gtx, th, skin, LineAt(inside, i, from), w.HoldingFrom("trace", told.Said[i], told.From[i]))
	}
	clipped.Pop()

	track := BesideBar(box)
	fill(gtx, track, skin.Back)
	fill(gtx, Thumb(track, len(told.Said), rows, w.History()), skin.Edge)
}

func drawTold(gtx layout.Context, th *material.Theme, skin Skin, box Box, held Written) {
	said := held.Value
	paintHeld(gtx, th, skin, box, held, small)
	stamp, rest, ok := Stamped(said)
	if !ok {
		say(gtx, th, skin.Ink, box, said, small)
		return
	}
	say(gtx, th, skin.Faint, box, stamp, small)
	say(gtx, th, skin.Ink, PastStamp(box, stamp), rest, small)
}

func sayHolding(gtx layout.Context, th *material.Theme, skin Skin, ink color.NRGBA, box Box, said Written, size int) {
	paintHeld(gtx, th, skin, box, said, size)
	say(gtx, th, ink, box, said.Value, size)
}

func paintHeld(gtx layout.Context, th *material.Theme, skin Skin, box Box, said Written, size int) {
	if !said.Holding() {
		return
	}
	kept := []rune(said.Value)
	from, to := said.Run()
	from, to = min(from, len(kept)), min(to, len(kept))
	loose := gtx
	loose.Constraints = layout.Constraints{Max: image.Pt(box.W, box.H)}
	fill(gtx, HeldRun(box, wide(loose, th, string(kept[:from]), size).Size.X,
		wide(loose, th, string(kept[from:to]), size).Size.X), skin.Ghost)
}

func drawTicked(gtx layout.Context, th *material.Theme, skin Skin, name, frame Box, w Window) {
	f := w.Form()
	sayHolding(gtx, th, skin, skin.Faint, name, Written{Value: "steps"}, small)
	fill(gtx, frame, skin.Back)
	outline(gtx, frame, skin.Edge)
	parts, from := w.PartsTicked(), w.partsFrom()
	on := f.onParts() && w.screen == OnForm
	places := w.partTiles()
	for k, tile := range places[:len(places)-1] {
		at := from + k
		if on && f.writingPart && f.partAt == at {
			fill(gtx, SubPanel(tile), skin.Panel)
			outline(gtx, SubPanel(tile), skin.Chosen)
			writing(gtx, th, skin, w.measured(), SubWriting(tile), f.named, true, small, 1, Follow)
			continue
		}
		drawSub(gtx, th, skin, tile, parts[at], StepLine(parts[at]), on && f.partAt == at, on && slices.Contains(f.selectedParts(), at), Written{})
	}
	if field, ok := f.partsField(); ok {
		_, value, rows := FormPlace(w.formBox(), f.Fields(), field)
		_, bar := Saying(value, rows)
		fill(gtx, bar, skin.Back)
		fill(gtx, Thumb(bar, len(parts), w.partsShown(), from), skin.Edge)
	}
	adding := places[len(places)-1]
	onAdding := on && f.partAt == f.lines()-1
	if onAdding && f.writingPart {
		writing(gtx, th, skin, w.measured(), adding, f.named, true, small, 1, Follow)
	} else {
		ink := skin.Faint
		if onAdding {
			ink = skin.Chosen
		}
		say(gtx, th, ink, adding, AddingAPart, small)
	}
	if ghost, ok := w.PartGhost(); ok {
		fill(gtx, ghost, skin.Ghost)
		outline(gtx, ghost, skin.Chosen)
		line := SubWriting(ghost)
		say(gtx, th, skin.Ink, line, fits(gtx, th, StepLine(parts[f.carried-1]), line.W, small), small)
	}
	if line, ok := w.PartLanding(); ok {
		fill(gtx, line, skin.Chosen)
	}
}

func drawFieldAt(gtx layout.Context, th *material.Theme, skin Skin, m Measure, name, value Box, rows, size int, long, lit, carets bool, said, named Written, scrolled int) {
	frame := Framed(value, rows)
	var bar Box
	if long {
		frame, bar = Saying(value, rows)
	}
	ink, edge := skin.Faint, skin.Edge
	if lit {
		ink, edge = skin.Chosen, skin.Chosen
	}
	sayHolding(gtx, th, skin, ink, name, named, small)
	fill(gtx, frame, skin.Back)
	outline(gtx, frame, edge)
	writing(gtx, th, skin, m, Inside(frame), said, carets, size, rows, scrolled)
	if long {
		lines := wrapIn(m, said.Reads(), Inside(frame), size)
		line, _ := lines.At(said.At)
		fill(gtx, bar, skin.Back)
		fill(gtx, Thumb(bar, len(lines.Said), rows, StartsAt(lines, rows, line, scrolled)), skin.Edge)
	}
}

func drawForm(gtx layout.Context, th *material.Theme, skin Skin, w Window) {
	panel, box := w.formPanel(), w.formBox()
	drawPanelHeaded(gtx, th, skin, panel, FormHeading(panel), HeadsOn(w, OnForm))
	drawCross(gtx, FormClose(panel), skin.Faint)
	fields := w.Form().Fields()
	for i := range fields {
		on := i == w.Form().On()
		typed := on && !chosenFromAList(fields[i])
		said := Written{Value: w.FieldSays(fields[i])}
		if typed {
			said = w.Form().Writing()
		}
		scrolled := w.scrolledOf(i)
		name, value, rows := FormPlace(box, fields, i)
		if fields[i].Name == "steps" {
			drawTicked(gtx, th, skin, name, Framed(value, rows), w)
			continue
		}
		drawFieldAt(gtx, th, skin, w.measured(), name, value, rows, FormSize(fields[i]), fields[i].Long, on, typed, said, Written{Value: fields[i].Name}, scrolled)
	}
	for _, head := range FormHeads(box, fields) {
		say(gtx, th, skin.Faint, head.Box, head.Said, head.Size)
	}
	fill(gtx, FormRule(panel), skin.Edge)
	if w.screen == OnForm {
		drawPanelOffers(gtx, th, skin, panel, w)
	}
	if tip, said, ok := w.Explanation(); ok {
		fill(gtx, tip, skin.Back)
		outline(gtx, tip, skin.Edge)
		for i, line := range said {
			ink := skin.Ink
			if strings.HasPrefix(line, ExampleSays) && i == len(said)-1 && w.ExplainsBy() != "" {
				ink = skin.Faint
			}
			say(gtx, th, ink, ExplainedLine(tip, i), line, small)
		}
	}
	switch {
	case !w.besideShown():
	case w.Form().ForObjective:
		drawUnder(gtx, th, skin, panel, w)
	default:
		drawHistory(gtx, th, skin, panel, w)
	}
	for _, r := range w.ContextSays() {
		sayHolding(gtx, th, skin, skin.Faint, r.Box, w.Holding(r.Key, r.Said), r.Size)
	}
}

func drawPanelOffers(gtx layout.Context, th *material.Theme, skin Skin, box Box, w Window) {
	along := w.offersAlong(box)
	if len(along) > 0 {
		fill(gtx, Box{X: box.X, Y: along[0].Y - between, W: box.W, H: 1}, skin.GroundEdge)
	}
	for i, offer := range w.Offers() {
		look := w.OfferLook(i, offer)
		if w.Lighting(offer) {
			look = litLook(look)
		}
		drawButton(gtx, th, skin, Button(along[i]), w.Said(offer), look)
	}
	if cross, ok := w.PanelClose(); ok && w.screen != OnForm {
		drawCross(gtx, cross, skin.Faint)
	}
}

func lighter(c color.NRGBA) color.NRGBA {
	up := func(v uint8) uint8 { return uint8(min(255, int(v)+18)) }
	return color.NRGBA{R: up(c.R), G: up(c.G), B: up(c.B), A: c.A}
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

func ground(gtx layout.Context, skin Skin, b Box) { rounded(gtx, b, groundRadius, skin.Ground) }

const (
	groundRadius = 10
	dash         = 4
	dashGap      = 3
)

func dashed(gtx layout.Context, b Box, with color.NRGBA) {
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

func rounded(gtx layout.Context, b Box, radius int, with color.NRGBA) {
	paint.FillShape(gtx.Ops, with, clip.UniformRRect(rect(b), radius).Op(gtx.Ops))
}

var shadows = []struct {
	down, spread int
	alpha        uint8
}{{18, 10, 18}, {10, 4, 35}, {4, 0, 60}}

func raised(gtx layout.Context, b Box, radius int, with color.NRGBA) {
	for _, s := range shadows {
		under := Box{X: b.X - s.spread, Y: b.Y + s.down - s.spread, W: b.W + 2*s.spread, H: b.H + 2*s.spread}
		rounded(gtx, under, radius+s.spread, color.NRGBA{A: s.alpha})
	}
	rounded(gtx, b, radius, with)
}

func outline(gtx layout.Context, b Box, with color.NRGBA) {
	for _, edge := range Edges(b) {
		fill(gtx, edge, with)
	}
}

func writing(gtx layout.Context, th *material.Theme, skin Skin, m Measure, b Box, said Written, carets bool, size, rows, scrolled int) {
	clipped := clip.Rect(rect(b)).Push(gtx.Ops)
	defer clipped.Pop()

	at := said.At
	lines := wrapIn(m, said.Reads(), b, size)
	line, column := lines.At(at)
	from := StartsAt(lines, rows, line, scrolled)

	begins, ends := said.Run()
	loose := gtx
	loose.Constraints = layout.Constraints{Max: image.Pt(b.W, lineHigh)}
	for i := from; i < Lasting(len(lines.Said), from, rows); i++ {
		if a, z, on := HeldOn(lines, begins, ends, i); said.Holding() && on {
			kept := []rune(lines.Said[i])
			at := LineAt(b, i, from)
			left := wide(loose, th, string(kept[:a]), size).Size.X
			fill(gtx, Box{X: at.X + left, Y: at.Y, H: at.H,
				W: wide(loose, th, string(kept[a:z]), size).Size.X}, skin.Ghost)
		}
		say(gtx, th, skin.Ink, LineAt(b, i, from), lines.Said[i], size)
	}
	if !carets {
		return
	}
	kept := []rune(lines.Said[line])
	before := wide(loose, th, string(kept[:min(len(kept), column)]), size)

	fill(gtx, CaretAt(b, line, from, before.Size.X), skin.Chosen)
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

func sayWith(gtx layout.Context, th *material.Theme, ink color.NRGBA, b Box, said string, size int, face font.Font, lines int, along text.Alignment) {
	if said == "" {
		return
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
