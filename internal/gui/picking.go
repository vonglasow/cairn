package gui

// Text on an open task, held and copied.
//
// A window draws text as shapes, and nothing said where a shape's letters were:
// nothing written on an open task could be held, and nothing could be copied
// out of it. Every piece of text it writes is listed here once, laid out the
// way it is drawn, so that what a pointer lands on and what is painted held are
// worked out from the same thing.

import "strconv"

// Readable is one piece of text a screen writes: what it is, where its first
// line in sight is drawn, what it says, the lines it is laid out on, which of
// them is first in sight, and the size it is written at.
type Readable struct {
	Key   string
	Box   Box
	Said  string
	Lines Lines
	From  int
	Size  int
}

// one is a piece of text written on one line, whatever its length: the say
// every label goes through, which cuts rather than wraps.
func one(key string, box Box, said string, size int) Readable {
	return Readable{Key: key, Box: box, Said: said, Size: size,
		Lines: Lines{Said: []string{said}, From: []int{0}}}
}

// Measure says where the edges between the characters of one line fall when it
// is written at a size: one edge before the first character and one after each,
// in dp from where the line begins. What a pointer lands on is read off the
// same face the line is drawn in — reckoned at one width for every character, a
// run held with the mouse ended a word away from the pointer by the end of a
// line, because the face is proportional and an i is not a W.
type Measure func(line string, size int) []int

// Reckoned is a Measure with no face behind it: every character as wide as an
// average one. It is what a window measures with until it is given a face, so
// that what a test asks of it does not depend on a font.
func Reckoned(line string, size int) []int {
	edges := make([]int, len([]rune(line))+1)
	for i := range edges {
		edges[i] = i * glyphOf(size)
	}
	return edges
}

// Nearest is the edge between two characters nearest to a point so far across
// the line: a pointer just short of a letter's end is after it, not before.
func (m Measure) Nearest(line string, size, x int) int {
	edges := m(line, size)
	at := 0
	for i := 1; i < len(edges); i++ {
		if 2*x >= edges[i-1]+edges[i] {
			at = i
		}
	}
	return at
}

// glyphOf is how wide a character is reckoned to be at a size, where nothing
// measures it.
func glyphOf(size int) int {
	if size <= small {
		return perSmallGlyph
	}
	if size >= heading {
		return perGlyph + 1
	}
	return perGlyph
}

// At is the character of a piece of text under a point, measured with the face
// it is drawn in.
func (r Readable) At(p Point, m Measure) int {
	return r.Lines.Reaches(r.Box, p, r.From, m, r.Size)
}

// inSight is the room a piece of text's lines take on screen.
func (r Readable) inSight() Box {
	rows := len(r.Lines.Said) - r.From
	return Box{X: r.Box.X, Y: r.Box.Y, W: r.Box.W, H: max(r.Box.H, rows*lineHigh)}
}

// Readables is every piece of text an open task writes, and nothing on any
// other screen: the form holds its own text in its fields.
func (w Window) Readables() []Readable {
	if w.screen != OnDetail {
		return nil
	}
	box := Over(w.width, w.height)
	column := DetailFields(box)
	fields := w.Shown()
	out := []Readable{one("heading", Heading(box), Heads(w), heading)}
	for i, field := range fields {
		name, value := FieldRow(column, FieldTop(column, fields, i))
		out = append(out, one("name:"+field.Name, name, field.Name, small))
		said := w.FieldSays(field)
		inside, from := Inside(Framed(value, RowsOf(column, fields, i))), 0
		if field.Long {
			inside, from = Inside(SayingBox(box, fields, i)), w.Saying()
		}
		out = append(out, Readable{Key: "field:" + field.Name, Box: inside, Said: said,
			Lines: Wrapped(said, room(inside)), From: from, Size: reads})
	}
	out = append(out, one("under", ServingHead(box, fields), w.UnderHead(), small))
	frame := ServingBox(box, fields)
	for at, c := range w.Under() {
		line := SubWriting(w.ServingCardAt(at))
		if !Within(line, frame) {
			continue
		}
		out = append(out, one("part:"+strconv.Itoa(c.Number), line, SubLine(c), small))
	}
	out = append(out, w.ContextSays()...)
	out = append(out, one("history", HistoryHead(box), w.PanelHead(), small))
	inside := Inside(HistoryBox(box))
	lines, from := w.Lines(), w.History()
	for i := from; i < Lasting(len(lines), from, w.Rows()); i++ {
		out = append(out, one("history:"+strconv.Itoa(i), LineAt(inside, i, from), lines[i], small))
	}
	return out
}

// Measuring gives the window the face its text is drawn in, so a pointer lands
// on the character drawn under it.
func (w Window) Measuring(m Measure) Window {
	w.measure = m
	return w
}

// measured is what the window measures text with: the face it was given, or the
// reckoning when it was given none.
func (w Window) measured() Measure {
	if w.measure == nil {
		return Reckoned
	}
	return w.measure
}

// readableAt is the piece of text a point lands on.
func (w Window) readableAt(p Point) (Readable, bool) {
	for _, r := range w.Readables() {
		if r.inSight().In(p) {
			return r, true
		}
	}
	return Readable{}, false
}

// readable is the piece of text written under a key, if it is still written.
func (w Window) readable(key string) (Readable, bool) {
	for _, r := range w.Readables() {
		if r.Key == key {
			return r, true
		}
	}
	return Readable{}, false
}

// Picked is a run of text held on an open task: which piece of text, and the
// characters it runs from and to. From and To the same is nothing held.
type Picked struct {
	Key      string
	From, To int
}

// Holding is a piece of text as it is drawn: with the run held in it, when the
// run held is in it, and plain otherwise. One answer for the drawing and for
// what ⌘C copies, so what is shown held is what is copied.
func (w Window) Holding(key, said string) Written {
	if w.picked.Key != key || w.picked.From == w.picked.To {
		return Written{Value: said}
	}
	return Written{Value: said, At: w.picked.To, Held: w.picked.From - w.picked.To}
}

// Unit is what a run held with the pointer grows by: a character for a press,
// a word for a double click, a paragraph for a triple — dragged on, each grows
// by whole ones of what it began with, the way text does everywhere else.
type Unit int

// The three a run grows by.
const (
	Characters Unit = iota
	Words
	Paragraphs
)

// runOf is the run of a unit around a character.
func runOf(said string, by Unit, at int) (from, to int) {
	switch by {
	case Words:
		return Written{Value: said}.Word(at).Run()
	case Paragraphs:
		return Written{Value: said}.Paragraph(at).Run()
	}
	return at, at
}

// Stretched is a run held from an anchor to where the pointer is, grown to
// whole units: the end that stays is the anchor's far side from the pointer,
// and the end that moves is the far side of the unit the pointer is in. The
// first is where the run is fixed, the second where it is going.
func Stretched(said string, by Unit, anchor Picked, at int) (fixed, moving int) {
	from, to := runOf(said, by, at)
	if from < anchor.From {
		return anchor.To, from
	}
	return anchor.From, max(to, anchor.To)
}

// picking starts a run where a press lands on a piece of text, and lets go of
// whatever was held when it lands on none.
func (w Window) picking(p Point) Window {
	r, on := w.readableAt(p)
	if !on {
		w.picked, w.pickingText = Picked{}, false
		return w
	}
	return w.pickingBy(r, p, Characters)
}

// pickingBy holds the unit of a piece of text a press lands on, and keeps it as
// the anchor a drag grows from.
func (w Window) pickingBy(r Readable, p Point, by Unit) Window {
	from, to := runOf(r.Said, by, r.At(p, w.measured()))
	w.picked, w.pickingText = Picked{Key: r.Key, From: from, To: to}, true
	w.growing, w.anchor = by, w.picked
	return w
}

// dragging runs the held text to where the pointer is, inside the piece of
// text the press began in and no other.
func (w Window) draggingText(p Point) Window {
	if r, on := w.readable(w.picked.Key); on {
		w.picked.From, w.picked.To = Stretched(r.Said, w.growing, w.anchor, r.At(p, w.measured()))
	}
	return w
}

// word holds the word a double click lands on: the run of it between spaces.
func (w Window) word(p Point) Window {
	if r, on := w.readableAt(p); on {
		return w.pickingBy(r, p, Words)
	}
	return w
}

// paragraph holds the paragraph a triple click lands on.
func (w Window) paragraph(p Point) Window {
	if r, on := w.readableAt(p); on {
		return w.pickingBy(r, p, Paragraphs)
	}
	return w
}

// all holds the whole of the piece of text last pressed in.
func (w Window) all() Window {
	if r, on := w.readable(w.picked.Key); on {
		w.picked.From, w.picked.To = 0, len([]rune(r.Said))
	}
	return w
}

// copyingPicked puts the run held on the clipboard.
func (w Window) copyingPicked() Window {
	if r, on := w.readable(w.picked.Key); on {
		w.copying = w.Holding(r.Key, r.Said).Holds()
	}
	return w
}

// Within says whether the whole of a line is inside a box, its edges included:
// a part half out of sight is not one a pointer can be said to land on.
func Within(line, frame Box) bool {
	return line.Y >= frame.Y && line.Y+line.H <= frame.Y+frame.H
}

// HeldRun is where a held run is painted on its line: as far in as what comes
// before it, as wide as it is.
func HeldRun(box Box, left, wide int) Box {
	return Box{X: box.X + left, Y: box.Y, W: wide, H: box.H}
}
