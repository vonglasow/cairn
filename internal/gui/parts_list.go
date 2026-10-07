package gui

// The parts of a task in its form, as a list worked with the keyboard and the
// pointer: walked, ticked, renamed, struck out, dragged and added to. Written
// underneath as the field's lines — "#2 [x] Relire l'ADR" — which is what is
// saved, so the list is a way of writing those lines and not a second thing
// to keep true beside them.

import (
	"slices"
	"strconv"

	"cairn/internal/engine"
	"cairn/internal/reading"
)

// partsField is which of a form's fields holds its parts, if one does.
func (f Form) partsField() (int, bool) {
	at := slices.IndexFunc(f.fields, func(field reading.Field) bool { return field.Name == "parts" })
	return at, at != -1
}

// parts is what the parts field says now.
func (f Form) parts() []reading.Part {
	at, ok := f.partsField()
	if !ok {
		return nil
	}
	return reading.PartsIn(f.fields[at].Value)
}

// withParts writes parts back into the field.
func (f Form) withParts(parts []reading.Part) Form {
	if at, ok := f.partsField(); ok {
		f.fields = slices.Clone(f.fields)
		f.fields[at].Value = reading.PartsWritten(parts)
	}
	return f
}

// onParts says whether the keyboard is on the parts field.
func (f Form) onParts() bool {
	at, ok := f.partsField()
	return ok && f.at == at
}

// lines is how many lines the list has: the parts, and the one that adds.
func (f Form) lines() int { return len(f.parts()) + 1 }

// isStruck says whether a part is struck out.
func (f Form) isStruck(part reading.Part) bool {
	return part.Number != 0 && slices.Contains(f.struck, part.Number)
}

// withoutStruck takes what is struck out off the field, which is what saving
// it means.
func (f Form) withoutStruck() Form {
	if len(f.struck) == 0 {
		return f
	}
	f = f.withParts(slices.DeleteFunc(f.parts(), f.isStruck))
	f.struck = nil
	return f
}

// walkingParts moves the keyboard along the list; up from its first line is
// up to the field before it, the way the arrows leave every other field.
func (f Form) walkingParts(by int) Form {
	if f.partAt+by < 0 {
		return f.onto(f.at - 1)
	}
	f.partAt = min(f.lines()-1, f.partAt+by)
	return f
}

// ticking turns the part the keyboard is on done, or back. The keyboard is on
// a part, and one not struck out: tickingPart says so before asking.
func (f Form) ticking() Form {
	parts := f.parts()
	parts[f.partAt].Done = !parts[f.partAt].Done
	return f.withParts(parts)
}

// tickingPart ticks the part the keyboard is on, or unticks it, and says so to
// the engine at once: a box ticked on a part says it is done, and a part done
// only once its task's form was saved is a box that lied until then. What the
// form opened with is told the same, so a tick is not something the form asks
// about leaving. A part nobody has saved yet is only ticked in the list: it
// goes in, ticked, with the form.
func (w Window) tickingPart() Window {
	parts := w.form.parts()
	at := w.form.partAt
	if at >= len(parts) || w.form.isStruck(parts[at]) {
		return w
	}
	w.form = w.form.ticking()
	part := w.form.parts()[at]
	if part.Number == 0 {
		return w
	}
	before := w.form
	w = w.asking(engine.TaskEdit, engine.Input{Given: map[string]string{
		"number": strconv.Itoa(part.Number), "state": reading.PartState(part.Done)}})
	// Refused, the box stays as it was: it says what the store holds.
	if w.err != nil {
		w.form = before.ticking()
		return w
	}
	w.form = before.openedWith(part)
	return w
}

// openedWith tells what the form opened with that a part is done, or not.
func (f Form) openedWith(part reading.Part) Form {
	at, _ := f.partsField()
	opened := reading.PartsIn(f.opened[at])
	for i := range opened {
		if opened[i].Number == part.Number {
			opened[i].Done = part.Done
		}
	}
	f.opened = slices.Clone(f.opened)
	f.opened[at] = reading.PartsWritten(opened)
	return f
}

// striking strikes the part the keyboard is on out, in its place until the
// form is saved — or, struck out already, brings it back. One never saved is
// simply taken away: there is nothing to strike out.
func (f Form) striking() Form {
	parts := f.parts()
	if f.partAt >= len(parts) {
		return f
	}
	part := parts[f.partAt]
	if part.Number == 0 {
		return f.withParts(slices.Delete(parts, f.partAt, f.partAt+1))
	}
	if at := slices.Index(f.struck, part.Number); at >= 0 {
		f.struck = slices.Delete(slices.Clone(f.struck), at, at+1)
		return f
	}
	f.struck = append(slices.Clone(f.struck), part.Number)
	return f
}

// naming begins writing on the line the keyboard is on: the title of a part,
// or a new one on the line that adds. A part struck out is not written on.
func (f Form) naming() Form {
	parts := f.parts()
	if f.partAt < len(parts) && f.isStruck(parts[f.partAt]) {
		return f
	}
	f.writingPart, f.named = true, Written{}
	if f.partAt < len(parts) {
		title := parts[f.partAt].Title
		f.named = Written{Value: title, At: len([]rune(title))}
	}
	return f
}

// committing puts what was written where it was written: a part renamed, or a
// new one added — and then the next is waiting on the line that adds. A name
// left empty changes nothing.
func (f Form) committing() Form {
	if !f.writingPart {
		return f
	}
	f.writingPart = false
	if f.named.Empty() {
		return f
	}
	parts := f.parts()
	if f.partAt < len(parts) {
		parts[f.partAt].Title = f.named.Reads()
		return f.withParts(parts)
	}
	f = f.withParts(append(parts, reading.Part{Title: f.named.Reads()}))
	f.partAt, f.writingPart, f.named = f.lines()-1, true, Written{}
	return f
}

// carrying moves a part from one place in the list to another.
func (f Form) carrying(from, to int) Form {
	parts := f.parts()
	if from >= len(parts) {
		return f
	}
	to = min(len(parts)-1, to)
	moved := parts[from]
	f.partAt = to
	return f.withParts(slices.Insert(slices.Delete(parts, from, from+1), to, moved))
}

// lettingGoOfAPart ends a press on the list of parts: carried to another
// place, or — not carried — a click, which ticks a part when it lands on its
// box and begins a new one on the line that adds.
func (w Window) lettingGoOfAPart(p Point) Window {
	from, moved := w.form.carried-1, w.held.Moved
	w.form.carried, w.held = 0, Held{}
	if moved {
		if to, on := w.partUnder(p); on {
			w.form = w.form.carrying(from, to)
		}
		return w
	}
	if from == w.form.lines()-1 {
		w.form = w.form.naming()
		return w
	}
	// Let go where it was pressed, so the box under the pointer is the box
	// of the part the press took hold of.
	tiles := w.partTiles()
	if slices.ContainsFunc(tiles[:len(tiles)-1], func(tile Box) bool { return SubCheck(tile).In(p) }) {
		return w.tickingPart()
	}
	return w
}

// inHand is the part being carried, and whether one is: pressed, and moved
// past the grip. The line that adds is pressed and never carried.
func (w Window) inHand() (int, bool) {
	from := w.form.carried - 1
	return from, from >= 0 && from < len(w.form.parts()) && w.held.Moved
}

// PartGhost is the part in hand, drawn under the pointer the way a card in
// hand is: as wide as a tile and as high. A drop with nothing shown moving is
// a drop nobody can aim.
func (w Window) PartGhost() (Box, bool) {
	if _, ok := w.inHand(); !ok {
		return Box{}, false
	}
	tile := w.partTiles()[0]
	return Box{X: tile.X, Y: w.held.At.Y - SubHeight/2, W: tile.W, H: SubHeight}, true
}

// PartLanding is the line saying where the part in hand would land: over the
// tile it would take the place of when carried up or held over its own, under
// it when carried down — which is where it ends up, the others moving to make
// room — and over the line that adds when it would go last. Nothing off the
// list.
func (w Window) PartLanding() (Box, bool) {
	from, ok := w.inHand()
	to, on := w.partUnder(w.held.At)
	if !ok || !on {
		return Box{}, false
	}
	tiles := w.partTiles()
	if to >= len(w.form.parts()) {
		adding := tiles[len(tiles)-1]
		return Box{X: tiles[0].X, Y: adding.Y, W: tiles[0].W, H: landingHigh}, true
	}
	tile := tiles[to-w.partsFrom()]
	if to > from {
		return Box{X: tile.X, Y: tile.Y + tile.H - landingHigh, W: tile.W, H: landingHigh}, true
	}
	return Box{X: tile.X, Y: tile.Y, W: tile.W, H: landingHigh}, true
}

// PartOn is the title of the part the keyboard is on, and whether it is on one.
func (w Window) PartOn() (string, bool) {
	parts := w.form.parts()
	if w.screen != OnForm || !w.form.onParts() || w.form.partAt >= len(parts) {
		return "", false
	}
	return parts[w.form.partAt].Title, true
}

// PartsGone is what is struck out, until the form is saved.
func (w Window) PartsGone() []reading.Part {
	return slices.DeleteFunc(w.form.parts(), func(part reading.Part) bool { return !w.form.isStruck(part) })
}

// onParts is what a key does with the keyboard on the list of parts, and
// whether it did anything: what it does not answer goes to the form.
func (w Window) onParts(name string) (Window, bool) {
	if w.form.writingPart {
		switch name {
		case "enter":
			w.form = w.form.committing()
		case "backspace":
			w.form.named = w.form.named.Back()
		case "left":
			w.form.named = w.form.named.Left()
		case "right":
			w.form.named = w.form.named.Right()
		case "cmd+a":
			w.form.named = w.form.named.All()
		default:
			return w, false
		}
		return w, true
	}
	switch name {
	case "down":
		w.form = w.form.walkingParts(1)
	case "up":
		w.form = w.form.walkingParts(-1)
	case "space":
		w = w.tickingPart()
	case "enter":
		w.form = w.form.naming()
	case "backspace":
		w.form = w.form.striking()
	default:
		return w, false
	}
	return w, true
}

// partTiles is where the lines of the list in sight are drawn — a tile for
// each part and each one struck out, from the first in sight — and, last, the
// line that adds.
func (w Window) partTiles() []Box {
	at, ok := w.form.partsField()
	if !ok {
		return nil
	}
	_, value, rows := FormPlace(w.formBox(), w.form.Fields(), at)
	tiles, adding := TickedPlaces(Framed(value, rows), w.form.lines()-1, w.partsFrom())
	return append(tiles, adding)
}

// partsFrame is the frame the list of parts is drawn in, and whether the form
// has one.
func (w Window) partsFrame() (Box, bool) {
	at, ok := w.form.partsField()
	if !ok {
		return Box{}, false
	}
	_, value, rows := FormPlace(w.formBox(), w.form.Fields(), at)
	return Framed(value, rows), true
}

// partsShown is how many parts the frame shows at once: one at the least.
func (w Window) partsShown() int {
	frame, _ := w.partsFrame()
	return max(1, ticksFit(frame))
}

// partsFrom is the first part in sight: where the wheel or the keyboard left
// the list, kept between its two ends — the way a column is scrolled. It once
// followed the keyboard and nothing else: with the keyboard elsewhere the
// list stood at its top and the wheel could not move it.
func (w Window) partsFrom() int { return w.partsWithin(w.form.partsTop) }

// partsWithin keeps where the list starts between its two ends: never above
// its first part, never past the last screenful of it.
func (w Window) partsWithin(top int) int {
	return max(0, min(top, len(w.form.parts())-w.partsShown()))
}

// keepingPartInSight moves the list as far as brings the line the keyboard is
// on into sight, and no further — the way a column follows the arrows.
func (w Window) keepingPartInSight() Window {
	line := w.form.partAt
	w.form.partsTop = w.partsWithin(min(line, max(w.partsFrom(), line-w.partsShown()+1)))
	return w
}

// scrollingParts turns the list of parts by the wheel, a tile at a time and
// keeping what did not come to one for the next turn.
func (w Window) scrollingParts(by int) Window {
	total := w.form.spareParts + by
	w.form.partsTop = w.partsWithin(w.partsFrom() + total/SubHeight)
	w.form.spareParts = total % SubHeight
	return w
}

// partUnder is which line of the list a point is on, if any.
func (w Window) partUnder(p Point) (int, bool) {
	tiles := w.partTiles()
	for i, tile := range tiles {
		if !tile.In(p) {
			continue
		}
		if i == len(tiles)-1 {
			return w.form.lines() - 1, true
		}
		return i + w.partsFrom(), true
	}
	return 0, false
}
