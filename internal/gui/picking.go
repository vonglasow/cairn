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

// glyphOf is how wide a character is reckoned to be at a size. Reckoned, the
// way the form reckons where a click lands: the face is proportional, and what
// is painted held is measured with it.
func glyphOf(size int) int {
	if size <= small {
		return perSmallGlyph
	}
	if size >= heading {
		return perGlyph + 1
	}
	return perGlyph
}

// At is the character of a piece of text under a point: the line it is down
// to, counting what is scrolled out of sight above it, then how far across.
func (r Readable) At(p Point) int {
	line := min(len(r.Lines.Said)-1, max(0, r.From+(p.Y-r.Box.Y)/lineHigh))
	across := max(0, (p.X-r.Box.X)/glyphOf(r.Size))
	return r.Lines.From[line] + min(across, len([]rune(r.Lines.Said[line])))
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
	out = append(out, one("history", HistoryHead(box), w.PanelHead(), small))
	inside := Inside(HistoryBox(box))
	lines, from := w.Lines(), w.History()
	for i := from; i < Lasting(len(lines), from, w.Rows()); i++ {
		out = append(out, one("history:"+strconv.Itoa(i), LineAt(inside, i, from), lines[i], small))
	}
	return out
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

// picking starts a run where a press lands on a piece of text, and lets go of
// whatever was held when it lands on none.
func (w Window) picking(p Point) Window {
	r, on := w.readableAt(p)
	if !on {
		w.picked, w.pickingText = Picked{}, false
		return w
	}
	at := r.At(p)
	w.picked, w.pickingText = Picked{Key: r.Key, From: at, To: at}, true
	return w
}

// dragging runs the held text to where the pointer is, inside the piece of
// text the press began in and no other.
func (w Window) draggingText(p Point) Window {
	if r, on := w.readable(w.picked.Key); on {
		w.picked.To = r.At(p)
	}
	return w
}

// word holds the word a double click lands on: the run of it between spaces.
func (w Window) word(p Point) Window {
	r, on := w.readableAt(p)
	if !on {
		return w
	}
	from, to := Written{Value: r.Said}.Word(r.At(p)).Run()
	w.picked = Picked{Key: r.Key, From: from, To: to}
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
