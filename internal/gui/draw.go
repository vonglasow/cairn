package gui

// Drawing, and nothing else. Every rectangle comes from Layout and every word
// from the window's own state — this file decides nothing, and it is the only
// place Gio is spoken to.

import (
	"image"
	"image/color"
	"math"
	"strconv"
	"strings"

	"gioui.org/f32"
	"gioui.org/font"
	"gioui.org/font/gofont"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"golang.org/x/image/math/fixed"

	"cairn/internal/domain"
	"cairn/internal/reading"
)

// The sizes anything is written at, named so a drawing never carries a number
// of its own.
const (
	small   = 12
	reads   = 13
	plain   = 14
	heading = 16
)

// Skin is every colour the window uses, in one place, so a second theme is a
// second Skin and not a search through the drawing.
type Skin struct {
	Back, Panel, Edge  color.NRGBA
	Ink, Faint, Chosen color.NRGBA
	Wrong              color.NRGBA
	// Late is what a card writes a day that has passed in, and an estimate
	// already spent. It is not Wrong: nothing has gone wrong, something is
	// asking to be looked at.
	Late color.NRGBA
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
		Panel:  color.NRGBA{R: 44, G: 44, B: 54, A: 255},
		Edge:   color.NRGBA{R: 104, G: 104, B: 122, A: 255},
		Ink:    color.NRGBA{R: 226, G: 226, B: 236, A: 255},
		Faint:  color.NRGBA{R: 160, G: 160, B: 178, A: 255},
		Chosen: color.NRGBA{R: 108, G: 158, B: 246, A: 255},
		Wrong:  color.NRGBA{R: 255, G: 110, B: 160, A: 255},
		Late:   color.NRGBA{R: 240, G: 186, B: 96, A: 255},
		Ghost:  color.NRGBA{R: 72, G: 88, B: 124, A: 255},
		// A priority is told from every other one, from the edge a card with
		// no priority wears, and from the blue that says what is chosen —
		// which "normal" used to be, at the bit.
		Priority: map[string]color.NRGBA{
			string(domain.Critical): {R: 236, G: 96, B: 96, A: 255},
			string(domain.High):     {R: 236, G: 168, B: 72, A: 255},
			string(domain.Normal):   {R: 96, G: 178, B: 146, A: 255},
			string(domain.Low):      {R: 140, G: 146, B: 172, A: 255},
		},
	}
}

// Contrast is how far two colours stand apart, by the measure a person reading
// a screen is held to: 21 at most, 1 when they are the same. Written here
// rather than judged by eye, because a palette nothing measures is a palette
// that drifts — this one had a card standing at 1.10 against its own board.
func Contrast(a, b color.NRGBA) float64 {
	one, two := lit(a), lit(b)
	return (max(one, two) + 0.05) / (min(one, two) + 0.05)
}

// lit is how much light a colour carries, which is not how bright it looks:
// the three channels are weighted the way an eye weighs them, and each is
// straightened out of the curve a screen applies.
func lit(c color.NRGBA) float64 {
	return 0.2126*straight(c.R) + 0.7152*straight(c.G) + 0.0722*straight(c.B)
}

// The curve turns at 0.04045 of full, which falls between the tenth and the
// eleventh of the 256 values a channel can hold. Written as that value and not
// as the fraction: a comparison whose two sides part company only where no
// input can reach is a comparison nothing can judge.
func straight(v uint8) float64 {
	f := float64(v) / 255
	if v <= 10 {
		return f / 12.92
	}
	return math.Pow((f+0.055)/1.055, 2.4)
}

// Theme is the shaper and the sizes Gio needs, built once.
func Theme() *material.Theme {
	th := material.NewTheme()
	th.Shaper = text.NewShaper(text.WithCollection(gofont.Collection()))
	return th
}

// Measured is the Measure of a theme's face: each line shaped the way a label
// shapes it, at the same size and in the same dp, and an edge read after every
// cluster of characters the face draws as one. A cluster's characters all end
// where it ends — a pointer cannot land inside a glyph the face drew whole.
//
// The face accounts for every character it is given, whatever they are: an
// accent written apart from its letter, an emoji and its skin tone, bytes that
// are not text at all — measured, each comes back as many characters as went
// in. So there is one edge per character and nothing to make up.
func Measured(th *material.Theme) Measure {
	return func(line string, size int) []int {
		edges := []int{0}
		th.Shaper.LayoutString(text.Parameters{
			Font: font.Font{Typeface: th.Face}, PxPerEm: fixed.I(size),
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
	if w.Screen() == OnForm {
		drawForm(gtx, th, skin, w)
	}
	if w.Screen() == OnHelp {
		drawSaid(gtx, th, skin, w, w.Helping(), Over(w.width, w.height))
	}
	if w.Screen() == OnAsking {
		drawSaid(gtx, th, skin, w, w.Asks(), w.QuestionBox())
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
	writing(gtx, th, skin, w.measured(), Inside(looking), w.Filtering(), true, reads, 1, Follow)

	at := ChoosingFrom(box)
	offered := w.Choosing()
	if len(offered) == 0 {
		say(gtx, th, skin.Faint, ChoosingLine(box, at), w.EmptyListSays(), reads)
		drawPanelOffers(gtx, th, skin, box, w)
		return
	}
	// Cut to the run the panel has. A list of everything a task may hang under
	// is as long as the board, and drawn whole it ran past the panel, past the
	// window, and over everything behind both.
	run := ChoosingRun(box)
	clipped := clip.Rect(rect(run)).Push(gtx.Ops)
	from := w.ChoosingFromRow()
	for i := from; i < Lasting(len(offered), from, w.ChoosingRoom()); i++ {
		one := offered[i]
		row := ChoosingRow(box, at)
		// The row the keyboard is on is lit, and the row is what is lit: a
		// list is read down, and a mark beside the names said nothing about
		// which line it belonged to.
		if i == w.ChoosingAt() {
			fill(gtx, row, skin.Back)
			outline(gtx, row, skin.Chosen)
		}
		line := ChoosingLine(box, at)
		say(gtx, th, skin.Ink, line, w.RowSays(one), reads)
		// Where it stands, hard right and faint: a list of fifty rows that all
		// read alike is scanned by something other than the words.
		if one.State != "" {
			say(gtx, th, skin.Faint, CountedAt(line, one.State), one.State, small)
		}
		at += apart
	}
	clipped.Pop()

	track := ChoosingBar(box)
	fill(gtx, track, skin.Back)
	fill(gtx, Thumb(track, len(offered), w.ChoosingRoom(), from), skin.Edge)
	drawPanelOffers(gtx, th, skin, box, w)
}

// asking is the one question this window puts, written the way the help is so
// they are drawn by one piece of code.

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
	writing(gtx, th, skin, w.measured(), into, w.Searched(), true, reads, 1, Follow)
}

// drawSaid is a panel of lines: the help, and the question. Both are read
// rather than pointed at, and both are two columns of words — the left one a
// key, drawn on a chip so it reads as a key rather than as another word.
func drawSaid(gtx layout.Context, th *material.Theme, skin Skin, w Window, said []Said, box Box) {
	drawPanel(gtx, th, skin, box, Heads(w))
	at := FirstRow(box)
	for _, line := range said {
		// A line with no key is a sentence, and runs the width of the box.
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
			drawCard(gtx, th, skin, l.CardAt(i, at, w.scroll[i]), c, c.Number == w.Chosen() || w.isTaken(c.Number))
		}
		clipped.Pop()
	}
}

func drawBand(gtx layout.Context, th *material.Theme, skin Skin, l Layout, w Window) {
	for lane, kind := range Lanes {
		say(gtx, th, skin.Faint, LaneName(l, lane), w.LaneSays(kind), small)
		clipped := clip.Rect(rect(l.Lane[lane])).Push(gtx.Ops)
		for at, c := range w.inLane(lane) {
			drawCard(gtx, th, skin, l.LaneCardAt(lane, at, w.lanes[lane]), c, c.Number == w.Chosen() || w.isTaken(c.Number))
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
	// Every card has an edge; the chosen one wears the colour that says so.
	// Drawn only when chosen, the edge was the one thing telling a card from
	// the board — and a fill at 1.35 against it is a hint, not a boundary.
	outline(gtx, parts.Panel, skin.Edge)
	if chosen {
		outline(gtx, parts.Panel, skin.Chosen)
	}
	say(gtx, th, skin.Faint, parts.Number, Numbered(c), small)
	if said := Carried(c, Marks()); said != "" {
		say(gtx, th, skin.Faint, CountedAt(parts.Number, said), said, small)
	}
	say(gtx, th, skin.Writing(c), parts.Title, TitleOf(c.Title, parts.Panel.W), plain)
	// What is due, what it has cost and what it serves, hard right on the label
	// row; the labels take whatever width that leaves. The window used to know
	// nothing at all about time.
	said, warn := Facts(c, gtx.Now)
	ink := skin.Faint
	if warn {
		ink = skin.Late
	}
	say(gtx, th, ink, CountedAt(parts.Labels, said), said, small)
	say(gtx, th, skin.Faint, parts.Labels, Tagged(c, LabelsRoom(parts.Labels, said)), small)
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
		drawCard(gtx, th, skin, box, goalFace(g), g.Number == w.Chosen() || w.isTaken(g.Number))
		// How much work it carries and how much is done, on the number line an
		// objective otherwise leaves half empty: it is the question an
		// objective exists to ask, and a list that answered it only on opening
		// each one would make somebody open all of them to find out.
		if said := Counted(g); said != "" {
			say(gtx, th, skin.Faint, CountedAt(PartsOf(box).Number, said), said, small)
		}
	}
}

// goalFace is an objective in the frame a card is drawn in: its number, title
// and priority where a task's go, and where it stands in the row a task's labels
// take. Said once, here, rather than built inline wherever a goal is drawn — a
// card's frame is the one this board draws, and an objective has no other.
func goalFace(g reading.Goal) reading.Card {
	return reading.Card{Number: g.Number, Title: g.Title, State: g.State, Priority: g.Priority, Labels: Standings(g)}
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
		if c.Number != w.held.Number {
			continue
		}
		// Written where the thing in hand writes it, which is where it would
		// be written once it has landed.
		line := PartsOf(ghost).Title
		say(gtx, th, skin.Ink, line, fits(gtx, th, TitleOf(c.Title, ghost.W), line.W, plain), plain)
	}
}

// drawBar is what can be done here, each offer a place a pointer can land.
func drawBar(gtx layout.Context, th *material.Theme, skin Skin, l Layout, w Window) {
	if w.Wrong() != nil {
		// Read at the size everything else is offered at, over two lines.
		// The one thing in the window somebody has to read in full was the
		// smallest text in it, on one line, cut wherever it ran out. On a
		// ground of its own, since it lies over the foot of the board.
		fill(gtx, WrongBox(l), skin.Back)
		sayOver(gtx, th, skin.Wrong, WrongBox(l), w.Wrong().Error(), reads, 2)
	}
	if s := w.Screen(); s != OnTheBoard && s != OnSearch {
		for i, offer := range w.BoardOffers() {
			say(gtx, th, skin.Faint, OfferAt(l, i), said(offer), reads)
		}
		return
	}
	for i, offer := range w.Offers() {
		drawOffer(gtx, th, skin, OfferAt(l, i), offer, w)
	}
}

// drawOffer writes one offer, lit when the pointer is resting on it. Lighting
// it is the one thing that says a word can be clicked, and the window said it
// nowhere: every offer read exactly like a label until it was pressed.
func drawOffer(gtx layout.Context, th *material.Theme, skin Skin, box Box, offer string, w Window) {
	if w.Lighting(offer) {
		fill(gtx, box, skin.Panel)
		say(gtx, th, skin.Ink, box, said(offer), reads)
		return
	}
	say(gtx, th, skin.Chosen, box, said(offer), reads)
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
func drawSub(gtx layout.Context, th *material.Theme, skin Skin, box Box, c reading.Card, on bool, held Written) {
	panel := SubPanel(box)
	fill(gtx, panel, skin.Panel)
	// Outlined where the keyboard is, the way a form lights the field it is
	// on: one place to look for where you are, whatever is open.
	if on {
		outline(gtx, panel, skin.Chosen)
	}
	// The box says whether it is done, by its shape: empty, or filled and
	// ticked. The ink and the line through the name say it again, the way
	// round a list is read — what is left stands out, what is done steps
	// back. It was the other way round, and brightness was the only
	// difference.
	check := SubCheck(box)
	done := reading.Finished(c)
	if done {
		fill(gtx, check, skin.Faint)
		say(gtx, th, skin.Panel, TickAt(check), TickedMark, small)
	} else {
		outline(gtx, check, skin.Ink)
	}
	line := SubWriting(box)
	said := fits(gtx, th, SubLine(c), line.W, small)
	held.Value = said
	sayHolding(gtx, th, skin, SubInk(skin, c), line, held, small)
	if done {
		loose := gtx
		loose.Constraints = layout.Constraints{Max: image.Pt(line.W, lineHigh)}
		fill(gtx, StruckAt(line, wide(loose, th, said, small).Size.X), skin.Faint)
	}
}

// SubInk is the ink a part's name is written in: the reading ink for what is
// left, the faint one for what is done.
func SubInk(skin Skin, c reading.Card) color.NRGBA {
	if reading.Finished(c) {
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
		drawSub(gtx, th, skin, w.ServingCardAt(at), c, false,
			w.Holding("part:"+strconv.Itoa(c.Number), SubLine(c)))
	}
	clipped.Pop()

	// Drawn whether or not there is anything to scroll, for the same reason the
	// history's is: a bar that comes and goes is a bar somebody reaches for and
	// does not find.
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
	lines, from, rows := w.Lines(), w.History(), w.Rows()
	for i := from; i < Lasting(len(lines), from, rows); i++ {
		drawTold(gtx, th, skin, LineAt(inside, i, from), w.Holding("history:"+strconv.Itoa(i), lines[i]))
	}
	clipped.Pop()

	// The bar is drawn whether or not there is anything to scroll: a bar that
	// comes and goes is a bar somebody reaches for and does not find.
	track := BesideBar(box)
	fill(gtx, track, skin.Back)
	fill(gtx, Thumb(track, len(lines), rows, w.History()), skin.Edge)
}

// drawTold writes one line of a history: the time it happened in the ink kept
// for the things beside what is read, and what happened in the ink that is
// read. It used to be all one faint grey — the densest, palest region of the
// window, and the one thing in it somebody opened a task to read.
func drawTold(gtx layout.Context, th *material.Theme, skin Skin, box Box, held Written) {
	said := held.Value
	// What is held of the line is painted under the whole of it, before its two
	// inks are written over it.
	paintHeld(gtx, th, skin, box, held, small)
	stamp, rest, ok := Stamped(said)
	if !ok {
		say(gtx, th, skin.Ink, box, said, small)
		return
	}
	say(gtx, th, skin.Faint, box, stamp, small)
	say(gtx, th, skin.Ink, PastStamp(box, stamp), rest, small)
}

// sayHolding writes one line with whatever run of it is held painted under it,
// in the ground the form paints a held run with: every piece of text written
// on an open task can be held, and what is held is shown.
func sayHolding(gtx layout.Context, th *material.Theme, skin Skin, ink color.NRGBA, box Box, said Written, size int) {
	paintHeld(gtx, th, skin, box, said, size)
	say(gtx, th, ink, box, said.Value, size)
}

// paintHeld paints what is held of one line under where it is written,
// measured with the face it is written in. Kept to what the line shows: a line
// cut to its box holds no more than it shows.
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

// drawTicked is the parts of a task in its form: a tile for each, ticked or
// not, and one struck out for each taken away until the form is saved; the
// line the keyboard is on outlined; a line being written drawn as it is
// written; and last the line that adds another.
func drawTicked(gtx layout.Context, th *material.Theme, skin Skin, name, frame Box, w Window) {
	f := w.Form()
	sayHolding(gtx, th, skin, skin.Faint, name, Written{Value: "parts"}, small)
	fill(gtx, frame, skin.Back)
	outline(gtx, frame, skin.Edge)
	parts, from := w.PartsTicked(), w.partsFrom()
	on := f.onParts() && w.screen == OnForm
	places := w.partTiles()
	for k, tile := range places[:len(places)-1] {
		at := from + k
		if f.isStruck(reading.Part{Number: parts[at].Number}) {
			drawStruck(gtx, th, skin, tile, parts[at], on && f.partAt == at)
			continue
		}
		if on && f.writingPart && f.partAt == at {
			fill(gtx, SubPanel(tile), skin.Panel)
			outline(gtx, SubPanel(tile), skin.Chosen)
			writing(gtx, th, skin, w.measured(), SubWriting(tile), f.named, true, small, 1, Follow)
			continue
		}
		drawSub(gtx, th, skin, tile, parts[at], on && f.partAt == at, Written{})
	}
	// The bar beside the list, in the margin every long field keeps: how much
	// of it is in sight, and which part.
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
	// What is in hand, over the list: the part itself under the pointer,
	// written the way its tile writes it, and over it where it would land —
	// drawn under the part, the line was hidden by the very thing it aims.
	if ghost, ok := w.PartGhost(); ok {
		fill(gtx, ghost, skin.Ghost)
		outline(gtx, ghost, skin.Chosen)
		line := SubWriting(ghost)
		say(gtx, th, skin.Ink, line, fits(gtx, th, SubLine(parts[f.carried-1]), line.W, small), small)
	}
	if line, ok := w.PartLanding(); ok {
		fill(gtx, line, skin.Chosen)
	}
}

// drawStruck is a part struck out of the form: faint, crossed through, and
// waiting for the form to be saved before it goes — or for backspace again.
func drawStruck(gtx layout.Context, th *material.Theme, skin Skin, box Box, part reading.Card, on bool) {
	panel := SubPanel(box)
	fill(gtx, panel, skin.Panel)
	if on {
		outline(gtx, panel, skin.Chosen)
	}
	line := SubWriting(box)
	said := fits(gtx, th, SubLine(part), line.W, small)
	say(gtx, th, skin.Faint, line, said, small)
	loose := gtx
	loose.Constraints = layout.Constraints{Max: image.Pt(line.W, lineHigh)}
	fill(gtx, StruckAt(line, wide(loose, th, said, small).Size.X), skin.Faint)
}

// drawFieldAt is drawField where the field is put and where its text starts in
// sight are given. A long field is drawn with the bar the description carries
// on an open task.
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
	drawPanel(gtx, th, skin, panel, HeadsOn(w, OnForm))
	// Every field is a box, whether or not the keyboard is on it, and the one
	// it is on is the one whose frame is lit. A place somebody types should
	// look like one before they have typed in it.
	fields := w.Form().Fields()
	for i := range fields {
		on := i == w.Form().On()
		// The field the keyboard is on is drawn from what is being typed, with
		// its caret and whatever run is held; the rest from what they say. A
		// field a list answers is never typed into, so it says its names even
		// while the keyboard is on it — the numbers underneath are for the
		// engine, not for reading.
		typed := on && !chosenFromAList(fields[i])
		said := Written{Value: w.FieldSays(fields[i])}
		if typed {
			said = w.Form().Writing()
		}
		scrolled := w.scrolledOf(i)
		name, value, rows := FormPlace(box, fields, i)
		if fields[i].Name == "parts" {
			drawTicked(gtx, th, skin, name, Framed(value, rows), w)
			continue
		}
		drawFieldAt(gtx, th, skin, w.measured(), name, value, rows, FormSize(fields[i]), fields[i].Long, on, typed, said, Written{Value: fields[i].Name}, scrolled)
	}
	for _, head := range FormHeads(box, fields) {
		say(gtx, th, skin.Faint, head.Box, head.Said, head.Size)
	}
	fill(gtx, FormRule(panel), skin.Edge)
	drawPanelOffers(gtx, th, skin, panel, w)
	// Beside a form opened on something already there: what was done to a
	// task, and where what was written about it elsewhere lives; or the work
	// serving an objective.
	if !w.Form().Adding {
		if w.ChosenGoal() {
			drawUnder(gtx, th, skin, panel, w)
		} else {
			drawHistory(gtx, th, skin, panel, w)
		}
		for _, r := range w.ContextSays() {
			sayHolding(gtx, th, skin, skin.Faint, r.Box, w.Holding(r.Key, r.Said), r.Size)
		}
	}
}

// drawPanelOffers writes what can be done on a panel where it can be pressed:
// the same words the bar says, in the panel's own row of places.
func drawPanelOffers(gtx layout.Context, th *material.Theme, skin Skin, box Box, w Window) {
	for i, offer := range w.Offers() {
		drawOffer(gtx, th, skin, PanelOfferAt(box, i), offer, w)
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
func writing(gtx layout.Context, th *material.Theme, skin Skin, m Measure, b Box, said Written, carets bool, size, rows, scrolled int) {
	clipped := clip.Rect(rect(b)).Push(gtx.Ops)
	defer clipped.Pop()

	at := said.At
	lines := wrapIn(m, said.Reads(), b, size)
	line, column := lines.At(at)
	from := StartsAt(lines, rows, line, scrolled)

	// The held run first, so the words are read over it rather than under.
	// Measured with the font, the way the caret is: a run measured at so many
	// pixels a glyph ended a word away from its own caret.
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
	// Measured with room to grow and not with the box's exact size: asked for
	// an exact size, a label answers that size whatever it wrote, and a caret
	// placed from that lands at the far edge of the box every time.
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
	sayOver(gtx, th, ink, b, said, size, 1)
}

// sayOver writes over however many lines it is given. One line is what nearly
// everything wants — a card, a field, a row of a list, all of which keep their
// shape whatever they hold — and a refusal is the exception.
func sayOver(gtx layout.Context, th *material.Theme, ink color.NRGBA, b Box, said string, size, lines int) {
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
	label.MaxLines = lines
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
