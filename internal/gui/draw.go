package gui

// Drawing, and nothing else. Every rectangle comes from Layout and every word
// from the window's own state — this file decides nothing, and it is the only
// place Gio is spoken to.

import (
	"image"
	"image/color"
	"strconv"

	"gioui.org/f32"
	"gioui.org/font/gofont"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget/material"

	"cairn/internal/reading"
)

// The sizes anything is written at, named so a drawing never carries a number
// of its own.
const (
	small   = 12
	reads   = 13
	plain   = 14
	heading = 16
	// apart is how far one line of a panel is from the next.
	apart = 26
)

// Skin is every colour the window uses, in one place, so a second theme is a
// second Skin and not a search through the drawing.
type Skin struct {
	Back, Panel, Edge  color.NRGBA
	Ink, Faint, Chosen color.NRGBA
	Wrong              color.NRGBA
	// Ghost is what a card in mid-air is drawn in, which has to be told from
	// both the board behind it and the card it will land on.
	Ghost    color.NRGBA
	Priority map[string]color.NRGBA
}

// Dark is the one skin there is. Named rather than assumed, so the day a second
// one is wanted there is somewhere for it to go.
func Dark() Skin {
	return Skin{
		Back:   color.NRGBA{R: 18, G: 18, B: 24, A: 255},
		Panel:  color.NRGBA{R: 28, G: 28, B: 38, A: 255},
		Edge:   color.NRGBA{R: 56, G: 56, B: 72, A: 255},
		Ink:    color.NRGBA{R: 226, G: 226, B: 236, A: 255},
		Faint:  color.NRGBA{R: 138, G: 138, B: 156, A: 255},
		Chosen: color.NRGBA{R: 108, G: 158, B: 246, A: 255},
		Wrong:  color.NRGBA{R: 236, G: 112, B: 112, A: 255},
		Ghost:  color.NRGBA{R: 64, G: 78, B: 108, A: 255},
		Priority: map[string]color.NRGBA{
			"critical": {R: 236, G: 96, B: 96, A: 255},
			"high":     {R: 236, G: 168, B: 72, A: 255},
			"normal":   {R: 108, G: 158, B: 246, A: 255},
			"low":      {R: 120, G: 128, B: 148, A: 255},
		},
	}
}

// Theme is the shaper and the sizes Gio needs, built once.
func Theme() *material.Theme {
	th := material.NewTheme()
	th.Shaper = text.NewShaper(text.WithCollection(gofont.Collection()))
	return th
}

// Draw puts the whole window on screen.
//
// Everything below this line is in dp. A window reports itself in pixels, and
// on a screen with two of them to the dp a board drawn at one to one comes out
// half the size it should be — measured, on a Retina screen, where every card
// was too short to read. So the whole drawing is scaled once, here, and nothing
// under it ever hears about a pixel.
func Draw(gtx layout.Context, th *material.Theme, skin Skin, w Window) {
	scaled := op.Affine(f32.Affine2D{}.Scale(f32.Pt(0, 0), f32.Pt(gtx.Metric.PxPerDp, gtx.Metric.PxPerDp))).Push(gtx.Ops)
	defer scaled.Pop()

	// And the ratio is spent: everything below is in dp, and a label told the
	// ratio again would be scaled by it twice. Measured on a Retina screen —
	// text four times too large, cut off at the bottom of its row.
	gtx.Metric = unit.Metric{PxPerDp: 1, PxPerSp: 1}

	fill(gtx, Box{W: w.width, H: w.height}, skin.Back)
	if w.Screen() == TooSmallToDraw {
		say(gtx, th, skin.Ink, Message(w.width, w.height), w.Says(), heading)
		return
	}
	l := w.Layout()
	drawTabs(gtx, th, skin, l, w)
	if w.Tab() == Objectives {
		drawGoals(gtx, th, skin, l, w)
	} else {
		drawColumns(gtx, th, skin, l, w)
		drawBand(gtx, th, skin, l, w)
		drawDragging(gtx, th, skin, w)
	}
	drawBar(gtx, th, skin, l, w)
	if w.Screen() == OnSearch {
		drawSearch(gtx, th, skin, l, w)
	}
	if w.Screen() == OnDetail {
		drawDetail(gtx, th, skin, w)
	}
	if w.Screen() == OnForm {
		drawForm(gtx, th, skin, w)
	}
	if w.Screen() == OnHelp {
		drawSaid(gtx, th, skin, w, w.Helping())
	}
	if w.Screen() == OnAsking {
		drawSaid(gtx, th, skin, w, asking(w))
	}
}

// asking is the one question this window puts, written the way the help is so
// they are drawn by one piece of code.
func asking(w Window) []reading.Field {
	what := "task"
	if w.ChosenGoal() {
		what = "objective"
	}
	return []reading.Field{
		{Name: "", Value: "This " + what + " cannot be brought back."},
		{Name: "", Value: "Delete again to agree, or Cancel."},
	}
}

// drawSearch is the box along the bottom that says what is being looked for.
//
// A box and not a line of text: a search used to be drawn as the key that
// opened it followed by what had been typed, so the first thing on screen was
// a character nobody had typed, in a place that looked like nothing.
func drawSearch(gtx layout.Context, th *material.Theme, skin Skin, l Layout, w Window) {
	box := Searching(l)
	fill(gtx, box, skin.Panel)
	outline(gtx, box, skin.Chosen)
	say(gtx, th, skin.Faint, SearchWord(box), "SEARCH", small)
	into := SearchInto(box)
	writing(gtx, th, skin, into, w.Searched().Reads(), w.Searched().At, true, reads, 1)
}

// drawSaid is a panel of lines: the help, and the question. Both are read
// rather than pointed at, and both are two columns of words — the left one a
// key, drawn on a chip so it reads as a key rather than as another word.
func drawSaid(gtx layout.Context, th *material.Theme, skin Skin, w Window, said []reading.Field) {
	box := Over(w.width, w.height)
	drawPanel(gtx, th, skin, box, Heads(w))
	at := FirstRow(box)
	for _, line := range said {
		name, value := FieldRow(box, at)
		if line.Name != "" {
			fill(gtx, Chip(name), skin.Back)
		}
		say(gtx, th, skin.Chosen, name, line.Name, reads)
		say(gtx, th, skin.Ink, value, line.Value, reads)
		at += apart
	}
	drawPanelOffers(gtx, th, skin, box, w)
}

// drawPanel is the ground every panel stands on, and the name it goes by.
func drawPanel(gtx layout.Context, th *material.Theme, skin Skin, box Box, head string) {
	fill(gtx, box, skin.Panel)
	outline(gtx, box, skin.Edge)
	say(gtx, th, skin.Ink, Heading(box), head, heading)
	fill(gtx, Ruled(box), skin.Edge)
}

// drawTabs is the band of folders across the top. The band is its own ground,
// closed by a line the open tab breaks: a tab does not say it is open by its
// colour, it says it by opening onto the page.
func drawTabs(gtx layout.Context, th *material.Theme, skin Skin, l Layout, w Window) {
	fill(gtx, l.Tabs, skin.Panel)
	for i, name := range Pages {
		tab := TabAt(l, i)
		ink := skin.Faint
		if int(w.Tab()) == i {
			// The open tab is the page's own ground, run up into the band, so
			// there is nothing between it and what it opens onto.
			fill(gtx, tab, skin.Back)
			for _, side := range TabSides(tab) {
				fill(gtx, side, skin.Edge)
			}
			ink = skin.Ink
		}
		say(gtx, th, ink, TabName(tab), name+"   "+strconv.Itoa(w.Holds(Tab(i))), plain)
	}
	// The line the open tab breaks, drawn in the two pieces either side of it.
	for _, piece := range TabsLineAround(l, int(w.Tab())) {
		fill(gtx, piece, skin.Edge)
	}
}

func drawColumns(gtx layout.Context, th *material.Theme, skin Skin, l Layout, w Window) {
	dealt, _ := reading.Deal(w.Showing().Cards)
	for i, column := range dealt {
		say(gtx, th, skin.Faint, l.Head[i], column.Name+"  "+strconv.Itoa(len(column.Cards)), reads)
		clipped := clip.Rect(rect(l.Column[i])).Push(gtx.Ops)
		for at, c := range column.Cards {
			drawCard(gtx, th, skin, l.CardAt(i, at, w.scroll[i]), c, c.Number == w.Chosen())
		}
		clipped.Pop()
	}
}

func drawBand(gtx layout.Context, th *material.Theme, skin Skin, l Layout, w Window) {
	_, stalled := reading.Deal(w.Showing().Cards)
	for lane, kind := range Lanes {
		say(gtx, th, skin.Faint, LaneName(l, lane), kind, small)
		clipped := clip.Rect(rect(l.Lane[lane])).Push(gtx.Ops)
		at := 0
		for _, c := range stalled {
			if c.State != kind {
				continue
			}
			drawCard(gtx, th, skin, l.LaneCardAt(lane, at, 0), c, c.Number == w.Chosen())
			at++
		}
		clipped.Pop()
	}
}

// drawCard is the one shape this board is made of: a panel, a coloured edge for
// what it carries, its number, and as much of its title as fits.
//
// A card in a lane says nothing about which kind of stopped it is — the lane it
// sits in says that, which is what a window has room for and a terminal has not.
func drawCard(gtx layout.Context, th *material.Theme, skin Skin, box Box, c reading.Card, chosen bool) {
	parts := PartsOf(box)
	fill(gtx, parts.Panel, skin.Panel)
	edge, known := skin.Priority[c.Priority]
	if !known {
		edge = skin.Edge
	}
	fill(gtx, parts.Edge, edge)
	if chosen {
		outline(gtx, parts.Panel, skin.Chosen)
	}
	say(gtx, th, skin.Faint, parts.Number, Numbered(c), small)
	say(gtx, th, skin.Ink, parts.Title, TitleOf(c.Title, parts.Panel.W), plain)
}

// drawGoals is the objectives tab: a list across the whole board, because an
// objective has no column to stand in.
func drawGoals(gtx layout.Context, th *material.Theme, skin Skin, l Layout, w Window) {
	// Cut to the run they belong in, the way a column's cards are. Scrolled,
	// they were drawn over the band of tabs above them.
	clipped := clip.Rect(rect(GoalsRun(l))).Push(gtx.Ops)
	defer clipped.Pop()
	for at, g := range reading.Sorted(w.Showing().Goals) {
		drawCard(gtx, th, skin, GoalAt(l, at, w.goals), reading.Card{
			Number: g.Number, Title: g.Title, State: g.State, Priority: g.Priority,
		}, g.Number == w.Chosen())
	}
}

// drawDragging is the card in mid-air and the line it would land on. Without
// them a drop is aimed by hope.
func drawDragging(gtx layout.Context, th *material.Theme, skin Skin, w Window) {
	if landing, somewhere := w.Landing(); somewhere {
		fill(gtx, landing, skin.Chosen)
	}
	ghost, dragging := w.Ghost()
	if !dragging {
		return
	}
	fill(gtx, ghost, skin.Ghost)
	outline(gtx, ghost, skin.Chosen)
	for _, c := range w.Reading().Cards {
		if c.Number == w.held.Number {
			say(gtx, th, skin.Ink, PartsOf(ghost).Title, TitleOf(c.Title, ghost.W), plain)
		}
	}
}

// drawBar is what can be done here, each offer a place a pointer can land.
func drawBar(gtx layout.Context, th *material.Theme, skin Skin, l Layout, w Window) {
	if w.Wrong() != nil {
		say(gtx, th, skin.Wrong, l.Bar, w.Wrong().Error(), small)
		return
	}
	for i, offer := range w.Offers() {
		say(gtx, th, skin.Chosen, OfferAt(l, i), said(offer), reads)
	}
}

// drawDetail and drawForm are panels over the board, which stays drawn behind
// them so nothing is lost from sight.
func drawDetail(gtx layout.Context, th *material.Theme, skin Skin, w Window) {
	box := Over(w.width, w.height)
	drawPanel(gtx, th, skin, box, Heads(w))

	// Everything the task carries, drawn in the boxes a form writes into —
	// the same task looks the same whether somebody is reading it or changing
	// it, and only the lit frame says which of the two is happening.
	fields := w.Shown()
	for i := range fields {
		drawField(gtx, th, skin, DetailFields(box), fields, i, false, 0, false)
	}
	drawHistory(gtx, th, skin, box, w)
	drawPanelOffers(gtx, th, skin, box, w)
}

// drawHistory is what was done to the task, beside what the task carries, in a
// box of its own that scrolls rather than one that runs off the window.
func drawHistory(gtx layout.Context, th *material.Theme, skin Skin, box Box, w Window) {
	say(gtx, th, skin.Faint, HistoryHead(box), "HISTORY", small)
	frame := HistoryBox(box)
	fill(gtx, frame, skin.Back)
	outline(gtx, frame, skin.Edge)

	inside := Inside(frame)
	clipped := clip.Rect(rect(inside)).Push(gtx.Ops)
	lines, from, rows := w.Lines(), w.History(), w.Rows()
	for i := from; i < Lasting(len(lines), from, rows); i++ {
		say(gtx, th, skin.Faint, LineAt(inside, i, from), lines[i], small)
	}
	clipped.Pop()

	// The bar is drawn whether or not there is anything to scroll: a bar that
	// comes and goes is a bar somebody reaches for and does not find.
	track := HistoryBar(box)
	fill(gtx, track, skin.Back)
	fill(gtx, Thumb(track, len(lines), rows, w.History()), skin.Edge)
}

// drawField draws one field of a panel: its name beside it, what it holds
// inside a box of its own, and the frame lit when the keyboard is on it.
func drawField(gtx layout.Context, th *material.Theme, skin Skin, box Box, fields []reading.Field, i int, lit bool, caret int, carets bool) {
	field := fields[i]
	name, value := FieldRow(box, FieldTop(box, fields, i))
	frame := Framed(value, RowsOf(field))
	ink, edge := skin.Faint, skin.Edge
	if lit {
		ink, edge = skin.Chosen, skin.Chosen
	}
	say(gtx, th, ink, name, field.Name, small)
	fill(gtx, frame, skin.Back)
	outline(gtx, frame, edge)
	writing(gtx, th, skin, Inside(frame), field.Value, caret, carets, reads, RowsOf(field))
}

func drawForm(gtx layout.Context, th *material.Theme, skin Skin, w Window) {
	box := Over(w.width, w.height)
	drawPanel(gtx, th, skin, box, Heads(w))
	// Every field is a box, whether or not the keyboard is on it, and the one
	// it is on is the one whose frame is lit. A place somebody types should
	// look like one before they have typed in it.
	fields := w.Form().Fields()
	for i := range fields {
		on := i == w.Form().On()
		drawField(gtx, th, skin, box, fields, i, on, w.Form().Writing().At, on)
	}
	drawPanelOffers(gtx, th, skin, box, w)
}

// drawPanelOffers writes what can be done on a panel where it can be pressed:
// the same words the bar says, in the panel's own row of places.
func drawPanelOffers(gtx layout.Context, th *material.Theme, skin Skin, box Box, w Window) {
	for i, offer := range w.Offers() {
		say(gtx, th, skin.Chosen, PanelOfferAt(box, i), said(offer), reads)
	}
}

// ── the only three things here that touch a pixel ─────────────────────────

func rect(b Box) image.Rectangle {
	return image.Rect(b.X, b.Y, b.X+b.W, b.Y+b.H)
}

func fill(gtx layout.Context, b Box, with color.NRGBA) {
	paint.FillShape(gtx.Ops, with, clip.Rect(rect(b)).Op())
}

func outline(gtx layout.Context, b Box, with color.NRGBA) {
	for _, edge := range Edges(b) {
		fill(gtx, edge, with)
	}
}

// writing draws a text with a caret standing between two of its letters.
//
// The caret is a shape and not a character, and where it stands is measured
// rather than reckoned: a proportional face gives "iii" and "WWW" very
// different widths, so counting characters would put the caret in the wrong
// place in every word that is not all the same letter. What is measured is the
// text before it, drawn nowhere and thrown away.
// writing draws a text in its box, one line at a time, and the caret between
// two letters rather than among them.
//
// Line by line and not as one label with a line count: the lines are worked
// out here, so the caret can be put on the line it actually fell on. A caret
// placed from the height of a whole wrapped label lands on the last line
// whatever was typed.
func writing(gtx layout.Context, th *material.Theme, skin Skin, b Box, said string, at int, carets bool, size, rows int) {
	clipped := clip.Rect(rect(b)).Push(gtx.Ops)
	defer clipped.Pop()

	lines := Wrapped(said, room(b))
	line, column := lines.At(at)
	from := lines.Window(rows, line)

	for i := from; i < Lasting(len(lines.Said), from, rows); i++ {
		say(gtx, th, skin.Ink, LineAt(b, i, from), lines.Said[i], size)
	}
	if !carets {
		return
	}
	// Measured with room to grow and not with the box's exact size: asked for
	// an exact size, a label answers that size whatever it wrote, and a caret
	// placed from that lands at the far edge of the box every time.
	loose := gtx
	loose.Constraints = layout.Constraints{Max: image.Pt(b.W, lineHigh)}
	kept := []rune(lines.Said[line])
	before := wide(loose, th, string(kept[:min(len(kept), column)]), size)

	fill(gtx, CaretAt(b, line, from, before.Size.X), skin.Chosen)
}

// wide is how much room a piece of text takes, measured by drawing it nowhere.
func wide(gtx layout.Context, th *material.Theme, said string, size int) layout.Dimensions {
	recorded := op.Record(gtx.Ops)
	label := material.Label(th, unit.Sp(float32(size)), said)
	label.MaxLines = 1
	dims := label.Layout(gtx)
	recorded.Stop()
	return dims
}

// say writes one line inside a box, clipped to it so nothing spills into what
// is beside it. One line and not more: a card is a fixed height, and a title
// that wrapped would change the shape of the board.
func say(gtx layout.Context, th *material.Theme, ink color.NRGBA, b Box, said string, size int) {
	if said == "" {
		return
	}
	clipped := clip.Rect(rect(b)).Push(gtx.Ops)
	defer clipped.Pop()
	offset := op.Offset(image.Pt(b.X, b.Y)).Push(gtx.Ops)
	defer offset.Pop()

	inner := gtx
	inner.Constraints = layout.Exact(image.Pt(b.W, b.H))
	label := material.Label(th, unit.Sp(float32(size)), said)
	label.Color = ink
	label.MaxLines = 1
	label.Layout(inner)
}

// said is an offer as the bar writes it: the key beside the word. An offer
// with no key is written alone rather than with a gap in front of it.
func said(offer string) string {
	key := Shortcut(offer)
	if key == "" {
		return offer
	}
	return Pretty(key) + "  " + offer
}
