package gui

import (
	"cairn/internal/reading"
)

// Where every parcel of every panel goes.
//
// It is here and not in the drawing for a reason mutation found: arithmetic
// inside a drawing is arithmetic no assertion can judge. A title moved six
// pixels still looks like a title, and the only test that would notice is a
// pinned pixel frame — which goes red on another machine for a reason that is
// not a defect. Written as numbers, it is pinned like every other size.

// The room inside a card, and what each of its two lines takes.
const (
	edgeWide = 3  // the coloured strip down a card's left, for its priority
	numberAt = 6  // how far below a card's top its number sits
	titleAt  = 28 // and its title
	labelsAt = 54 // and the labels it carries, under the room the title has
	lineHigh = 18
	titleHi  = 22
)

// tabWide is how much of the strip one tab takes. Written in rather than
// measured: a tab is a word, and a word that changes length should not move the
// one beside it.
const tabWide = 140

// namesWide is how much of a form's row the name of a field takes, leaving the
// rest for what it holds.
const namesWide = 170

// Parts is where the pieces of a card go inside the box it was given.
type Parts struct {
	Panel, Edge, Number, Title, Labels Box
}

// PartsOf shares out a card's box. The gap at the bottom is what keeps two
// cards from touching; a card is CardHeight tall and draws in rather less.
func PartsOf(box Box) Parts {
	panel := Box{X: box.X, Y: box.Y, W: box.W, H: box.H - gap}
	inside := panel.W - padding - padding
	return Parts{
		Panel:  panel,
		Edge:   Box{X: panel.X, Y: panel.Y, W: edgeWide, H: panel.H},
		Number: Box{X: panel.X + padding, Y: panel.Y + numberAt, W: inside, H: lineHigh},
		Title:  Box{X: panel.X + padding, Y: panel.Y + titleAt, W: inside, H: titleHi},
		Labels: Box{X: panel.X + padding, Y: panel.Y + labelsAt, W: inside, H: lineHigh},
	}
}

// SubPanel is the body of one tile: its box less the gap that separates it
// from the next.
func SubPanel(b Box) Box {
	return Box{X: b.X, Y: b.Y, W: b.W, H: b.H - between}
}

// SubWriting is where a tile writes what it carries, which is the one line it
// has. Asked for by the tile and by the ghost of one in hand, so that what is
// carried is drawn where it would be drawn if it were put down — a ghost laid
// out any other way is a ghost that lies about where the thing will go.
func SubWriting(b Box) Box {
	panel := SubPanel(b)
	return Row(panel, panel.Y+inset, lineHigh)
}

// CountedAt is where an objective's count sits on its number line: hard right,
// so that it and the number are read as two separate things rather than as one
// run of digits.
func CountedAt(number Box, said string) Box {
	wide := len([]rune(said)) * perGlyph
	return Box{X: number.X + number.W - wide, Y: number.Y, W: wide, H: number.H}
}

// ChoosingOver is the panel the list of objectives is drawn in: inside the
// form's own, so the form shows around it. A choice made with the rest of the
// form out of sight is a choice made from memory.
func ChoosingOver(width, height int) Box {
	over := Over(width, height)
	return Box{
		X: over.X + over.W/8,
		Y: over.Y + over.H/8,
		W: over.W * 3 / 4,
		H: over.H * 3 / 4,
	}
}

// Over is the panel a detail or a form is drawn in: the middle of the window,
// with the board still showing around it.
//
// It takes rather more of the window than it used to. Once every field was
// drawn in a frame, a form of twelve fields ran off the bottom of the panel at
// the smallest window this board draws in — and a field off the bottom is a
// field nobody can fill in. Measured, not guessed: see the test that puts the
// last field of the longest form against what the panel offers.
func Over(width, height int) Box {
	return Box{X: width / 8, Y: height / 12, W: width * 3 / 4, H: height * 5 / 6}
}

// Row is one line inside a panel.
func Row(box Box, at, high int) Box {
	return Box{X: box.X + padding, Y: at, W: box.W - padding - padding, H: high}
}

// FieldRow is one line of a form: what the field is called, and what it holds.
func FieldRow(box Box, at int) (name, value Box) {
	return Box{X: box.X + padding, Y: at, W: namesWide - padding, H: lineHigh},
		Box{X: box.X + namesWide, Y: at, W: box.W - namesWide - padding, H: lineHigh}
}

// ChoosingRow is one line of a list: the whole width of the panel, not a form's
// name column beside its value column. A list is read down, so what says where
// the keyboard is has to be the line itself — lit in the name column, it sat a
// hundred and fifty pixels from the row it marked.
func ChoosingRow(box Box, at int) Box {
	return Box{X: box.X + padding, Y: at - inset, W: box.W - padding - padding, H: lineHigh + inset + inset}
}

// ChoosingLooking is the line at the top of a list that narrows it, above the
// first row and inside the same panel.
func ChoosingLooking(box Box) Box {
	return Box{X: box.X + padding, Y: FirstRow(box) - inset, W: box.W - padding - padding, H: lineHigh + inset + inset}
}

// ChoosingRun is the room a list's rows have: from under the line that narrows
// them to above the offers along the panel's foot. A list is not as long as
// what it offers — it is as long as the panel, and the rest is scrolled to.
func ChoosingRun(box Box) Box {
	// Begun where a row's box begins and not where its text does: a row is
	// drawn an inset above its line, so a run measured from the line left the
	// first row's own box hanging over the top of it.
	top := ChoosingFrom(box) - inset
	return Box{X: box.X + padding, Y: top, W: box.W - padding - padding - barWide - between,
		H: max(apart, PanelOfferAt(box, 0).Y-padding-top)}
}

// ChoosingRows is how many rows fit in that run.
func ChoosingRows(box Box) int { return max(1, ChoosingRun(box).H/apart) }

// ChoosingBar is the track of the bar beside a list, drawn whether or not
// there is anything to scroll — the same rule the history goes by.
func ChoosingBar(box Box) Box {
	run := ChoosingRun(box)
	return Box{X: run.X + run.W + between, Y: run.Y, W: barWide, H: run.H}
}

// ChoosingAtPoint says which row of a list a point lands on, and whether it
// lands on one at all.
func ChoosingAtPoint(box Box, rows, from int, p Point) (int, bool) {
	for i := from; i < min(rows, from+ChoosingRows(box)); i++ {
		if ChoosingRow(box, ChoosingFrom(box)+(i-from)*apart).In(p) {
			return i, true
		}
	}
	return 0, false
}

// ChoosingFrom is where a list's own rows begin, under the line that narrows
// them.
func ChoosingFrom(box Box) int {
	looking := ChoosingLooking(box)
	return looking.Y + looking.H + between
}

// PanelOfferAt is where one of a panel's own offers sits: along its bottom,
// each in its own place so a pointer can land on one of them.
//
// A panel draws what can be done on it, and what is drawn has to work where it
// is drawn — words that look like buttons and answer only somewhere else are
// worse than no words at all.
func PanelOfferAt(box Box, at int) Box {
	return Box{
		X: box.X + padding + at*offerWide,
		Y: box.Y + box.H - padding - lineHigh,
		W: offerWide,
		H: lineHigh,
	}
}

// TabAt is where one of the two tabs sits. It stands a little below the top of
// the band, so the band is something the tabs sit in rather than something
// they fill.
func TabAt(l Layout, at int) Box {
	return Box{X: l.Tabs.X + at*tabWide, Y: l.Tabs.Y + between, W: tabWide, H: l.Tabs.H - between}
}

// The tabs are folders, not two words one of which is coloured.
//
// A band closed by a line runs the width of the window, and the tab somebody
// is on is a block whose bottom breaks that line — it does not stand out by
// its colour, it stands out by opening onto the page. That is what says "you
// are on this page" rather than "this word is selected", and a colour contrast
// is what it replaced: the two words sat in the same emptiness and neither of
// them looked like a page.

// TabsLine is the line closing the band, run the width of the window.
func TabsLine(l Layout) Box {
	return Box{X: l.Tabs.X, Y: l.Tabs.Y + l.Tabs.H - 1, W: l.Tabs.W, H: 1}
}

// TabsLineAround is that line in the two pieces it is drawn in: the tab
// somebody is on breaks it, which is the whole of the folder.
func TabsLineAround(l Layout, at int) [2]Box {
	line, tab := TabsLine(l), TabAt(l, at)
	return [2]Box{
		{X: line.X, Y: line.Y, W: max(0, tab.X-line.X), H: line.H},
		{X: tab.X + tab.W, Y: line.Y, W: max(0, line.X+line.W-tab.X-tab.W), H: line.H},
	}
}

// TabSides is the three sides a tab is drawn with: a folder is open at the
// bottom, which is where the page it opens onto begins.
func TabSides(b Box) [3]Box {
	return [3]Box{
		{X: b.X, Y: b.Y, W: b.W, H: 1},
		{X: b.X, Y: b.Y, W: 1, H: b.H},
		{X: b.X + b.W - 1, Y: b.Y, W: 1, H: b.H},
	}
}

// TabName is where a tab's name and what it holds are written inside it.
func TabName(b Box) Box {
	return Box{X: b.X + padding, Y: b.Y + (b.H-lineHigh)/2, W: b.W - padding - padding, H: lineHigh}
}

// LaneName is where a lane's name is written, above the cards in it.
func LaneName(l Layout, lane int) Box {
	return Box{X: l.Lane[lane].X, Y: l.Band.Y, W: l.Lane[lane].W, H: laneHigh}
}

// Message is where a window too small says so: across the middle of it.
func Message(width, height int) Box {
	return Box{X: margin, Y: height/2 - lineHigh, W: width - margin - margin, H: titleHi}
}

// Edges is the four sides of a box, each one row or one column thick. Written
// here rather than inside the drawing so the arithmetic that finds the far side
// of a box is arithmetic a test can pin.
func Edges(b Box) [4]Box {
	return [4]Box{
		{X: b.X, Y: b.Y, W: b.W, H: 1},
		{X: b.X, Y: b.Y + b.H - 1, W: b.W, H: 1},
		{X: b.X, Y: b.Y, W: 1, H: b.H},
		{X: b.X + b.W - 1, Y: b.Y, W: 1, H: b.H},
	}
}

// FirstRow is where the first line inside a panel begins: under its heading.
func FirstRow(box Box) int {
	ruled := Ruled(box)
	return ruled.Y + ruled.H + padding
}

// offerWide is how much of the bar one offer takes. Wide enough for the longest
// word the bar says, so an offer never moves because another changed length.
const offerWide = 110

// OfferAt is where one of the bar's offers sits. They are places a pointer can
// land, which is what makes anything at all reachable without a shortcut.
func OfferAt(l Layout, at int) Box {
	return Box{X: l.Bar.X + at*offerWide, Y: l.Bar.Y, W: offerWide, H: l.Bar.H}
}

// GoalsRun is the whole of the board the objectives are laid out in: every
// column's width, and the height the first of them runs to.
//
// Written once and used three times — to cut the grid to, to say how far it
// scrolls, and to bring what is chosen into sight. Cut to one column's width
// instead, three quarters of the grid went missing.
func GoalsRun(l Layout) Box {
	last := l.Column[Columns-1]
	return Box{
		X: l.Column[0].X,
		Y: l.Column[0].Y,
		W: last.X + last.W - l.Column[0].X,
		H: l.Column[0].H,
	}
}

// GoalsAcross is how many objectives stand side by side. The board's own
// columns, so an objective is the width of a card and the grid reflows with
// the window without a sum of its own.
func GoalsAcross() int { return Columns }

// GoalAt is where the n-th objective sits in the grid, filled across and then
// down.
//
// A grid and not a list. One to a row, each the width of the whole board, an
// objective was a card's height stretched over four columns of nothing — and
// a dozen of them ran off the bottom of the window, so the ones that mattered
// least were the only ones in sight. Across, a dozen fit without scrolling at
// the smallest window this board draws in.
func GoalAt(l Layout, at, scrolled int) Box {
	across := GoalsAcross()
	column := l.Column[at%across]
	return Box{
		X: column.X,
		Y: l.Column[0].Y + (at/across)*CardHeight - scrolled,
		W: column.W,
		H: CardHeight,
	}
}

// GoalRows is how many rows a given number of objectives comes to.
func GoalRows(goals int) int {
	across := GoalsAcross()
	return (goals + across - 1) / across
}

// Searching is where what is being looked for is written: a line along the
// bottom, taking the rest of the bar beside the one offer that closes it.
//
// Along the bottom and not across the middle, which is where a window says it
// is too small — a place nobody types, and a place that covers the very board
// a search is narrowing.
func Searching(l Layout) Box { return l.Bar }

// searchWide is the room the word "Search" takes at the left of that box.
const searchWide = 76

// SearchWord is where the box says what it is, and SearchInto where what is
// being looked for is typed.
//
// Said in a word and not in a slash: the key that opens a search used to be
// drawn where the text goes, so the first thing anybody saw was a character
// they had not typed, in a place that looked like nothing in particular.
func SearchWord(b Box) Box {
	return Box{X: b.X + padding, Y: b.Y + (b.H-lineHigh)/2, W: searchWide, H: lineHigh}
}

// SearchInto is the rest of that box, which is what is typed into.
func SearchInto(b Box) Box {
	return Box{X: b.X + padding + searchWide, Y: b.Y + (b.H-lineHigh)/2, W: b.W - padding - padding - searchWide, H: lineHigh}
}

// Chip is the patch drawn behind a key so it reads as a key and not as a word.
// It stands a little off the line it is on, because a chip flush with the text
// is a chip nobody sees.
func Chip(name Box) Box {
	return Box{X: name.X - chipAt, Y: name.Y - chipAt, W: name.W, H: name.H + chipAt + chipAt}
}

// chipAt is how far a chip stands off the line it sits on.
const chipAt = 3

// Heading is where a panel says what it is, above everything it holds. Every
// panel has one: a panel that opens with no name is a panel somebody has to
// read through before knowing what they are looking at.
func Heading(box Box) Box {
	return Box{X: box.X + padding, Y: box.Y + padding, W: box.W - padding - padding, H: titleHi}
}

// Ruled is the line under that heading, which is what makes it a heading
// rather than the first row.
func Ruled(box Box) Box {
	head := Heading(box)
	return Box{X: head.X, Y: head.Y + head.H + chipAt, W: head.W, H: 1}
}

// inset is what a frame keeps between its edge and what it holds. A field is
// drawn in a box the way the search is drawn in a box: a place somebody types
// should look the same wherever it is, and a value floating on the panel's own
// ground looked like something written rather than something to write in.
const inset = 4

// Framed is the box drawn around what a field holds, tall enough for every
// line that field is given.
func Framed(value Box, rows int) Box {
	return Box{X: value.X, Y: value.Y - inset, W: value.W, H: rows*lineHigh + inset + inset}
}

// The detail panel is two columns: what the task carries on the left, what was
// done to it on the right.
//
// Two columns and not one above the other, because the history is the part
// that never fits. Stacked, twelve framed fields take nearly the whole panel
// at the smallest window this board draws in, and the history gets what is
// left — which is nothing. Beside them it gets the panel's full height at
// every size.
const (
	// fieldsPart is how much of the panel's width the fields take.
	fieldsPart = 3
	ofPanel    = 5
	// barWide is the bar beside the history, which says how much of it is
	// showing and where in it somebody is.
	barWide = 10
	// leastThumb is the smallest the bar's grip is drawn. A grip in proportion
	// to a very long history would be a few pixels tall: too small to see, and
	// too small to take hold of.
	leastThumb = 24
)

// DetailFields is the left column of the detail panel.
func DetailFields(box Box) Box {
	return Box{X: box.X, Y: box.Y, W: box.W * fieldsPart / ofPanel, H: box.H}
}

// DetailHistory is the right column, which the history has to itself.
func DetailHistory(box Box) Box {
	left := DetailFields(box)
	return Box{X: left.X + left.W, Y: box.Y, W: box.W - left.W, H: box.H}
}

// ServingHead is where the box under the fields says what it is: the left
// column, below the last field, in the room the fields did not take.
func ServingHead(box Box, fields []reading.Field) Box {
	left := DetailFields(box)
	return Box{
		X: left.X + padding,
		Y: FieldTop(box, fields, len(fields)),
		W: left.W - padding - padding,
		H: lineHigh,
	}
}

// ServingBox is the frame the work under an objective is drawn in: from under
// its own heading down to just above what the panel offers, so it reaches
// neither. Never smaller than one line, the way the history's frame is not —
// a frame that collapses is a frame somebody reads as missing.
func ServingBox(box Box, fields []reading.Field) Box {
	head := ServingHead(box, fields)
	top := head.Y + head.H + between
	return Box{
		X: head.X,
		Y: top,
		W: head.W - barWide - between,
		H: max(lineHigh, PanelOfferAt(box, 0).Y-padding-top),
	}
}

// ServingBar is the track beside that box, in the room its frame gave up for
// it — the same shape the history's bar has.
func ServingBar(box Box, fields []reading.Field) Box {
	frame := ServingBox(box, fields)
	return Box{X: frame.X + frame.W + between, Y: frame.Y, W: barWide, H: frame.H}
}

// HistoryHead is where that column says what it is.
func HistoryHead(box Box) Box {
	right := DetailHistory(box)
	return Box{X: right.X + padding, Y: FirstRow(box), W: right.W - padding - padding, H: lineHigh}
}

// HistoryBox is the frame the history is drawn in: from under its own heading
// down to just above what the panel offers, so it never reaches either.
func HistoryBox(box Box) Box {
	head := HistoryHead(box)
	top := head.Y + head.H + between
	return Box{
		X: head.X,
		Y: top,
		W: head.W - barWide - between,
		H: max(lineHigh, PanelOfferAt(box, 0).Y-padding-top),
	}
}

// SayingBox is where a task's description is read on the detail, narrowed by
// the room its bar takes. Its own function and not a field's plain box: on the
// detail the description is read rather than written, so nothing follows a
// caret and what does not fit has to be reachable another way.
func SayingBox(box Box, fields []reading.Field, at int) Box {
	_, value := FieldRow(DetailFields(box), FieldTop(DetailFields(box), fields, at))
	frame := Framed(value, longRows)
	return Box{X: frame.X, Y: frame.Y, W: frame.W - barWide - between, H: frame.H}
}

// SayingBar is the track of the bar beside it, drawn whether or not there is
// anything to scroll — the same rule the history goes by.
func SayingBar(box Box, fields []reading.Field, at int) Box {
	frame := SayingBox(box, fields, at)
	return Box{X: frame.X + frame.W + between, Y: frame.Y, W: barWide, H: frame.H}
}

// LongField says which of a panel's fields is the one read in that box, and
// whether there is one at all.
func LongField(fields []reading.Field) (int, bool) {
	for i, field := range fields {
		if field.Long {
			return i, true
		}
	}
	return 0, false
}

// HistoryBar is the track of the bar beside it.
func HistoryBar(box Box) Box {
	frame := HistoryBox(box)
	return Box{X: frame.X + frame.W + between, Y: frame.Y, W: barWide, H: frame.H}
}

// Thumb is where that bar's grip sits: how tall it is says how much of the
// history is showing, and where it sits says which part.
//
// A history that fits gets a grip filling the whole track — the bar is still
// drawn, because a bar that comes and goes is a bar somebody reaches for and
// does not find.
func Thumb(track Box, lines, rows, at int) Box {
	high := thumbHigh(track, lines, rows)
	room := track.H - high
	if room < 1 {
		return track
	}
	down := room * min(at, lines-rows) / (lines - rows)
	return Box{X: track.X, Y: track.Y + down, W: track.W, H: high}
}

// thumbHigh is how tall that grip is: in proportion to how much of the history
// is showing, never smaller than can be taken hold of, and the whole track
// when there is nothing to scroll.
//
// One height and not two. Thumb and Reached each worked it out, and worked it
// out differently — a grip drawn one size and dragged as though it were
// another is a grip that jumps under the hand.
func thumbHigh(track Box, lines, rows int) int {
	if lines <= rows {
		return track.H
	}
	return max(leastThumb, track.H*rows/lines)
}

// Reached is which line of the history a point on the track means, so that
// taking hold of the grip and pulling it moves the history with it.
func Reached(track Box, lines, rows int, p Point) int {
	high := thumbHigh(track, lines, rows)
	room := track.H - high
	if room < 1 {
		return 0
	}
	// Past that guard there is more history than room, so the division below
	// has something to divide by.
	return (p.Y - track.Y - high/2) * (lines - rows) / room
}

// inSight moves a run so that a row is inside it, and answers where the run
// then stands.
//
// One arithmetic and not two: the columns and the objectives each had a copy,
// and the second copy of a boundary is where a boundary goes wrong unnoticed.
// Written without branches for a reason the mutation gate found: a run moved
// to exactly the row's top, and a row whose bottom lands exactly on the run's
// edge, both answer the same whichever branch is taken — so the two boundaries
// went out unjudged. As two clamps there are no boundaries to get wrong.
func inSight(at, high, run, where int) int {
	top := at * high
	return min(top, max(where, top+high-run))
}

// caretWide is how thick the caret is drawn. Two, because one disappears on a
// screen with two pixels to the dp.
const caretWide = 2

// LineAt is where one line of a wrapped text is drawn, given which line is the
// first one showing.
func LineAt(inside Box, i, from int) Box {
	return Box{X: inside.X, Y: inside.Y + (i-from)*lineHigh, W: inside.W, H: lineHigh}
}

// Lasting is the line after the last one drawn: the last there is, or the last
// there is room for.
func Lasting(lines, from, rows int) int { return min(lines, from+rows) }

// CaretAt is where the caret stands: on the line it fell on, after however
// much room what comes before it takes.
func CaretAt(inside Box, line, from, wide int) Box {
	at := LineAt(inside, line, from)
	return Box{X: at.X + wide, Y: at.Y, W: caretWide, H: at.H}
}

// HistoryRows is how many lines of history the frame can show at once.
func HistoryRows(box Box) int { return max(1, Inside(HistoryBox(box)).H/lineHigh) }

// Told is the history as the lines it takes in the room there is, one act
// after another with a blank line between two of them — an act that wrapped
// onto three lines and the next act beginning straight after it read as one
// act nobody could tell the end of.
func Told(acts []string, room int) []string {
	var lines []string
	for i, act := range acts {
		if i > 0 {
			lines = append(lines, "")
		}
		lines = append(lines, Wrapped(act, room).Said...)
	}
	return lines
}

// between is what one framed field leaves before the next. Small, because a
// frame already says where a field ends — the room a line needed when nothing
// was drawn around it is room a form no longer has to spend.
const between = 6

// fieldHigh is what one field takes from the top of its frame to the top of
// the next: its lines, the room its frame keeps, and what it leaves after it.
func fieldHigh(field reading.Field) int {
	return RowsOf(field)*lineHigh + inset + inset + between
}

// Inside is the room for the text within that frame.
func Inside(b Box) Box {
	return Box{X: b.X + inset, Y: b.Y + inset, W: b.W - inset - inset, H: b.H - inset - inset}
}

// Lines is a text broken into the lines it is drawn on, and where each of them
// began in the text — which is what puts the caret on the line it fell on,
// since the caret is counted in the whole text and not in one line of it.
type Lines struct {
	Said []string
	From []int
}

// Wrapped breaks a text into the lines it takes in a given room, at a space
// when there is one to break at, and inside a word when the word is longer
// than the room.
//
// Broken rather than cut: a description that runs out of its box is a
// description nobody can read back, and cutting it says nothing about what was
// lost. A word longer than the room is broken inside itself, because the only
// other answer is a line that runs out of the box — the very thing this
// prevents.
func Wrapped(said string, room int) Lines {
	kept := []rune(said)
	// No room is answered whole rather than one character at a time, and
	// nothing is answered as one empty line — a text of no length would
	// otherwise come back as no lines at all, and the caret has to fall
	// somewhere.
	if room < 1 || len(kept) == 0 {
		return Lines{Said: []string{said}, From: []int{0}}
	}
	lines := Lines{}
	from := 0
	for from < len(kept) {
		end := min(len(kept), from+room)
		if end < len(kept) {
			// Back to the last space there is room for — but never as far as
			// the line's own beginning, which would give an empty line and a
			// text that never gets any shorter.
			for i := end - 1; i > from; i-- {
				if kept[i] == ' ' {
					end = i
					break
				}
			}
		}
		lines.Said = append(lines.Said, string(kept[from:end]))
		lines.From = append(lines.From, from)
		from = end
		// The space a line was broken at belongs to neither line.
		if from < len(kept) && kept[from] == ' ' {
			from++
		}
	}
	return lines
}

// At says which line the caret fell on and how far into that line it is.
func (l Lines) At(caret int) (line, column int) {
	for i, from := range l.From {
		if caret >= from {
			line = i
		}
	}
	column = min(len([]rune(l.Said[line])), max(0, caret-l.From[line]))
	return line, column
}

// Window is the first of the lines shown when there are more of them than
// there is room for: the last screenful ending on the caret's line, because
// what is being typed is what has to stay in sight.
//
// It was also clamped against the lines there are, and the mutation gate
// showed that clamp could never win — the caret's line is always inside the
// text, so the run ending on it always ends inside the text too. Arithmetic
// nothing can reach is arithmetic nothing can judge.
func (l Lines) Window(rows, line int) int {
	return max(0, line-rows+1)
}

// longRows is how many rows a field that wants more than one gets. Four: a
// description is a paragraph, and a paragraph in one line is a slot the size of
// a word.
const longRows = 4

// RowsOf is how many rows a field takes.
func RowsOf(field reading.Field) int {
	if field.Long {
		return longRows
	}
	return 1
}

// FieldTop is where a field begins, counting what the fields above it took. A
// form is not a grid of equal rows once one of them wants four.
func FieldTop(box Box, fields []reading.Field, at int) int {
	top := FirstRow(box)
	// One guard and not two: a second condition on the same loop has a
	// boundary nothing can reach, since what is counted is only ever the
	// fields that exist.
	for _, field := range fields[:min(at, len(fields))] {
		top += fieldHigh(field)
	}
	return top
}

// room is how many characters a box can show, reckoned the way a title is.
func room(b Box) int { return (b.W - padding - padding) / perGlyph }
