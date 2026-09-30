package gui

// Drawing, and nothing else. Every rectangle comes from Layout and every word
// from the window's own state — this file decides nothing, and it is the only
// place Gio is spoken to.

import (
	"image"
	"image/color"
	"strconv"
	"strings"

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
		// Over the panel and not under it: what is in hand is in hand over
		// whatever is on screen, and the panel is what is on screen.
		drawDragging(gtx, th, skin, w)
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
	if w.Screen() == OnChoosing {
		// The form first, and the list on top of it: what is being filled in
		// stays on screen, because a choice made with the rest of the form out
		// of sight is a choice made from memory.
		drawForm(gtx, th, skin, w)
		drawChoosing(gtx, th, skin, w)
	}
}

// TickedMark says an objective is in. A third mark in a window that took two on
// purpose, and it earns its place the way they did: this list is about ticking,
// and a word in the column where every row carries one would be read as part of
// the name rather than as its state.
const TickedMark = "✓"

// drawChoosing is the list of objectives a task may serve, over the form. Drawn
// like the help and the question rather than like a board: it is read down, and
// what it offers is one line each.
func drawChoosing(gtx layout.Context, th *material.Theme, skin Skin, w Window) {
	box := ChoosingOver(w.width, w.height)
	drawPanel(gtx, th, skin, box, Heads(w))

	// The line that narrows the list, above it and inside the same panel.
	looking := ChoosingLooking(box)
	fill(gtx, looking, skin.Back)
	outline(gtx, looking, skin.Edge)
	writing(gtx, th, skin, Inside(looking), w.Filtering().Reads(), w.Filtering().At, true, reads, 1, 0)

	at := ChoosingFrom(box)
	offered := w.Choosing()
	if len(offered) == 0 {
		say(gtx, th, skin.Faint, Row(box, at, lineHigh), ChoosingSays, reads)
		drawPanelOffers(gtx, th, skin, box, w)
		return
	}
	for i, one := range offered {
		row := ChoosingRow(box, at)
		// The row the keyboard is on is lit, and the row is what is lit: a
		// list is read down, and a mark beside the names said nothing about
		// which line it belonged to.
		if i == w.ChoosingAt() {
			fill(gtx, row, skin.Back)
			outline(gtx, row, skin.Chosen)
		}
		said := "   "
		if w.Ticked(one.Number) {
			said = TickedMark + "  "
		}
		say(gtx, th, skin.Ink, Row(box, at, lineHigh),
			said+Numbered(reading.Card{Number: one.Number})+"  "+one.Title, reads)
		at += apart
	}
	drawPanelOffers(gtx, th, skin, box, w)
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
	writing(gtx, th, skin, into, w.Searched().Reads(), w.Searched().At, true, reads, 1, 0)
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
	dealt, _ := reading.Deal(w.Showing().Cards, w.archived)
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
	_, stalled := reading.Deal(w.Showing().Cards, w.archived)
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
	fill(gtx, parts.Edge, skin.edgeOf(c))
	if chosen {
		outline(gtx, parts.Panel, skin.Chosen)
	}
	say(gtx, th, skin.Faint, parts.Number, Numbered(c), small)
	if said := Carried(c, Marks()); said != "" {
		say(gtx, th, skin.Faint, CountedAt(parts.Number, said), said, small)
	}
	say(gtx, th, skin.Writing(c), parts.Title, TitleOf(c.Title, parts.Panel.W), plain)
	say(gtx, th, skin.Faint, parts.Labels, Tagged(c, parts.Labels.W), small)
}

// Writing is the ink a card's title is written in. Archived, it is faint: what
// is off the board is on screen because somebody asked, not because it is work.
func (s Skin) Writing(c reading.Card) color.NRGBA {
	if c.Archived {
		return s.Faint
	}
	return s.Ink
}

// edgeOf is the stripe down a card's side, which is its priority. Archived, it
// is the plain edge: a colour that shouts is a colour asking to be acted on.
func (s Skin) edgeOf(c reading.Card) color.NRGBA {
	if edge, known := s.Priority[c.Priority]; known && !c.Archived {
		return edge
	}
	return s.Edge
}

// drawGoals is the objectives tab: a list across the whole board, because an
// objective has no column to stand in.
func drawGoals(gtx layout.Context, th *material.Theme, skin Skin, l Layout, w Window) {
	// Cut to the run they belong in, the way a column's cards are. Scrolled,
	// they were drawn over the band of tabs above them.
	clipped := clip.Rect(rect(GoalsRun(l))).Push(gtx.Ops)
	defer clipped.Pop()
	for at, g := range reading.Sorted(w.Showing().Goals) {
		box := GoalAt(l, at, w.goals)
		drawCard(gtx, th, skin, box, reading.Card{
			Number: g.Number, Title: g.Title, State: g.State, Priority: g.Priority,
		}, g.Number == w.Chosen())
		// How much work it carries and how much is done, on the number line an
		// objective otherwise leaves half empty: it is the question an
		// objective exists to ask, and a list that answered it only on opening
		// each one would make somebody open all of them to find out.
		if said := Counted(g); said != "" {
			say(gtx, th, skin.Faint, CountedAt(PartsOf(box).Number, said), said, small)
		}
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
	body := ghost
	if w.Screen() == OnDetail {
		body = SubPanel(ghost)
	}
	fill(gtx, body, skin.Ghost)
	outline(gtx, body, skin.Chosen)
	for _, c := range w.Reading().Cards {
		if c.Number != w.held.Number {
			continue
		}
		// Written where the thing in hand writes it. A card puts its title 28
		// pixels down and a tile is 22 high, so a card's layout put the
		// writing past the bottom of the ghost and onto whatever was behind
		// it. What is carried is drawn the way it is drawn where it came from.
		if w.Screen() == OnDetail {
			line := SubWriting(ghost)
			say(gtx, th, skin.Ink, line, fits(gtx, th, SubLine(c), line.W, small), small)
			continue
		}
		say(gtx, th, skin.Ink, PartsOf(ghost).Title, TitleOf(c.Title, ghost.W), plain)
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
		if fields[i].Long {
			drawSaying(gtx, th, skin, box, fields, i, w)
			continue
		}
		drawField(gtx, th, skin, DetailFields(box), fields, i, false, 0, false, w.FieldSays(fields[i]))
	}
	drawUnder(gtx, th, skin, box, fields, w)
	drawHistory(gtx, th, skin, box, w)
	drawPanelOffers(gtx, th, skin, box, w)
}

// drawHistory is what was done to the task, beside what the task carries, in a
// box of its own that scrolls rather than one that runs off the window.
// drawServing is the work under an open objective, in the room the fields left
// at the foot of the left column. Only for an objective: a task is served by
// nothing, and a frame drawn over nothing is a frame that says there is
// something to read.
// drawSub is one part of the work: a strip saying whether it is done, then its
// number and its title on the one line it has.
//
// The strip carries done where a card's carries priority. It is the only fact
// a part has left that a glance should not have to read the words for, and a
// sub-task has no priority of its own to put there instead.
func drawSub(gtx layout.Context, th *material.Theme, skin Skin, box Box, c reading.Card, on bool) {
	panel := SubPanel(box)
	fill(gtx, panel, skin.Panel)
	edge, ink := skin.Edge, skin.Faint
	if reading.Finished(c) {
		edge, ink = skin.Chosen, skin.Ink
	}
	fill(gtx, Box{X: panel.X, Y: panel.Y, W: edgeWide, H: panel.H}, edge)
	// Outlined where the keyboard is, the way a form lights the field it is
	// on: one place to look for where you are, whatever is open.
	if on {
		outline(gtx, panel, skin.Chosen)
	}
	line := SubWriting(box)
	say(gtx, th, ink, line, fits(gtx, th, SubLine(c), line.W, small), small)
}

func drawUnder(gtx layout.Context, th *material.Theme, skin Skin, box Box, fields []reading.Field, w Window) {
	say(gtx, th, skin.Faint, ServingHead(box, fields), w.UnderHead(), small)
	frame := ServingBox(box, fields)
	fill(gtx, frame, skin.Back)
	outline(gtx, frame, skin.Edge)

	under := w.Under()
	if len(under) == 0 {
		// An objective says so; a task says nothing, because most tasks have
		// nothing under them and a line saying so on every one of them would
		// be a sentence nobody is reading by the third time.
		if w.ChosenGoal() {
			say(gtx, th, skin.Faint, LineAt(Inside(frame), 0, 0), ServingSays, small)
		}
		return
	}

	// Tiles of one line, cut to the frame the way a column's cards are. Not
	// the board's cards: those are 96 pixels and use 50, the rest being room
	// for facts a part of the work does not carry.
	clipped := clip.Rect(rect(frame)).Push(gtx.Ops)
	for at, c := range under {
		drawSub(gtx, th, skin, w.ServingCardAt(at), c, at == w.UnderAt())
	}
	clipped.Pop()

	// Drawn whether or not there is anything to scroll, for the same reason the
	// history's is: a bar that comes and goes is a bar somebody reaches for and
	// does not find.
	track := ServingBar(box, fields)
	fill(gtx, track, skin.Back)
	fill(gtx, Thumb(track, len(under)*SubHeight, frame.H, w.ServingAt()), skin.Edge)
}

func drawHistory(gtx layout.Context, th *material.Theme, skin Skin, box Box, w Window) {
	say(gtx, th, skin.Faint, HistoryHead(box), w.PanelHead(), small)
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
func drawField(gtx layout.Context, th *material.Theme, skin Skin, box Box, fields []reading.Field, i int, lit bool, caret int, carets bool, said string) {
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
	writing(gtx, th, skin, Inside(frame), said, caret, carets, reads, RowsOf(field), 0)
}

// drawSaying is the description on an open task: read rather than written, so
// what does not fit is reached by the wheel and a bar says where in it you are.
// It draws the pieces a field draws, in a box narrowed by the bar beside it.
func drawSaying(gtx layout.Context, th *material.Theme, skin Skin, box Box, fields []reading.Field, at int, w Window) {
	name, _ := FieldRow(DetailFields(box), FieldTop(DetailFields(box), fields, at))
	frame := SayingBox(box, fields, at)
	say(gtx, th, skin.Faint, name, fields[at].Name, small)
	fill(gtx, frame, skin.Back)
	outline(gtx, frame, skin.Edge)
	writing(gtx, th, skin, Inside(frame), w.FieldSays(fields[at]), 0, false, reads, w.SaidRows(), w.Saying())

	track := SayingBar(box, fields, at)
	fill(gtx, track, skin.Back)
	fill(gtx, Thumb(track, len(w.SaidLines()), w.SaidRows(), w.Saying()), skin.Edge)
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
		drawField(gtx, th, skin, box, fields, i, on, w.Form().Writing().At, on, w.FieldSays(fields[i]))
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
func writing(gtx layout.Context, th *material.Theme, skin Skin, b Box, said string, at int, carets bool, size, rows, scrolled int) {
	clipped := clip.Rect(rect(b)).Push(gtx.Ops)
	defer clipped.Pop()

	lines := Wrapped(said, room(b))
	line, column := lines.At(at)
	// A caret drags the window along with it; with no caret there is nothing
	// to follow, so what is in sight is what was scrolled to.
	from := scrolled
	if carets {
		from = lines.Window(rows, line)
	}

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

// fits cuts a piece of text to the room there is, measured rather than
// reckoned. Counting an average glyph width is close enough for a card, whose
// title has a line to itself; on a tile it left the text running to the
// frame's edge, where say clipped it mid-word against the bar beside it.
//
// Measured by drawing nowhere, which is what wide already does for the caret.
// Walked back one rune at a time rather than guessed at: the font is
// proportional, so no arithmetic on a count of runes answers this.
func fits(gtx layout.Context, th *material.Theme, said string, width, size int) string {
	// Loosened before measuring, the way the caret loosens it. A label laid out
	// under the window's own constraints answers the width it was allowed and
	// not the width it wants, so every string measured wider than any tile and
	// every title came out as a lone ellipsis. Seen in a screenshot: the pixel
	// tests all passed, because what they asked was whether writing reached
	// the frame's edge, and there was no writing left anywhere.
	loose := gtx
	loose.Constraints = layout.Constraints{Max: image.Pt(1<<20, lineHigh)}
	if wide(loose, th, said, size).Size.X <= width {
		return said
	}
	// Grown from the front rather than shrunk from the back: the longest that
	// still fits is kept, and the first that does not ends it.
	//
	// Walking down from the full length gave the same answer from any starting
	// point at or above it — every step above the answer produces a string too
	// wide to take, so where the walk began could not be read from what it
	// returned. Growing forward makes each step tell.
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
