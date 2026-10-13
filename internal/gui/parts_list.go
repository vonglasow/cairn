package gui

import (
	"slices"
	"strconv"

	"cairn/internal/engine"
)

func (f Form) partsField() (int, bool) {
	at := slices.IndexFunc(f.fields, func(field engine.Field) bool { return field.Name == "parts" })
	return at, at != -1
}

func (f Form) parts() []engine.Part {
	at, ok := f.partsField()
	if !ok {
		return nil
	}
	return engine.PartsIn(f.fields[at].Value)
}

func (f Form) withParts(parts []engine.Part) Form {
	if at, ok := f.partsField(); ok {
		f.fields = slices.Clone(f.fields)
		f.fields[at].Value = engine.PartsWritten(parts)
	}
	return f
}

func (f Form) onParts() bool {
	at, ok := f.partsField()
	return ok && f.at == at
}

func (f Form) lines() int { return len(f.parts()) + 1 }

func (f Form) isStruck(part engine.Part) bool {
	return part.Number != 0 && slices.Contains(f.struck, part.Number)
}

func (f Form) withoutStruck() Form {
	if len(f.struck) == 0 {
		return f
	}
	f = f.withParts(slices.DeleteFunc(f.parts(), f.isStruck))
	f.struck = nil
	return f
}

func (f Form) walkingParts(by int) Form {
	if f.partAt+by < 0 {
		return f.onto(f.at - 1)
	}
	f.partAt = min(f.lines()-1, f.partAt+by)
	return f
}

func (f Form) ticking() Form {
	parts := f.parts()
	parts[f.partAt].Done = !parts[f.partAt].Done
	return f.withParts(parts)
}

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
		"number": strconv.Itoa(part.Number), "state": engine.PartState(part.Done)}})
	if w.err != nil {
		w.form = before.ticking()
		return w
	}
	w.form = before.openedWith(part)
	return w
}

func (f Form) openedWith(part engine.Part) Form {
	at, _ := f.partsField()
	opened := engine.PartsIn(f.opened[at])
	for i := range opened {
		if opened[i].Number == part.Number {
			opened[i].Done = part.Done
		}
	}
	f.opened = slices.Clone(f.opened)
	f.opened[at] = engine.PartsWritten(opened)
	return f
}

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
	f = f.withParts(append(parts, engine.Part{Title: f.named.Reads()}))
	f.partAt, f.writingPart, f.named = f.lines()-1, true, Written{}
	return f
}

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
	tiles := w.partTiles()
	if slices.ContainsFunc(tiles[:len(tiles)-1], func(tile Box) bool { return SubCheck(tile).In(p) }) {
		return w.tickingPart()
	}
	return w
}

func (w Window) inHand() (int, bool) {
	from := w.form.carried - 1
	return from, from >= 0 && from < len(w.form.parts()) && w.held.Moved
}

// PartGhost is the part in hand, drawn under the pointer the way a card in hand is: as wide as a tile and as high.
func (w Window) PartGhost() (Box, bool) {
	if _, ok := w.inHand(); !ok {
		return Box{}, false
	}
	tile := w.partTiles()[0]
	return Box{X: tile.X, Y: w.held.At.Y - SubHeight/2, W: tile.W, H: SubHeight}, true
}

// PartLanding is the line saying where the part in hand would land: over the tile it would take the place of when carried up or held over its own, under it when carried down — which is where it ends up, the others moving to make room — and over the line that adds when it would go last.
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

func (w Window) partTiles() []Box {
	at, ok := w.form.partsField()
	if !ok {
		return nil
	}
	_, value, rows := FormPlace(w.formBox(), w.form.Fields(), at)
	tiles, adding := TickedPlaces(Framed(value, rows), w.form.lines()-1, w.partsFrom())
	return append(tiles, adding)
}

func (w Window) partsFrame() (Box, bool) {
	at, ok := w.form.partsField()
	if !ok {
		return Box{}, false
	}
	_, value, rows := FormPlace(w.formBox(), w.form.Fields(), at)
	return Framed(value, rows), true
}

func (w Window) partsShown() int {
	frame, _ := w.partsFrame()
	return max(1, ticksFit(frame))
}

func (w Window) partsFrom() int { return w.partsWithin(w.form.partsTop) }

func (w Window) partsWithin(top int) int {
	return max(0, min(top, len(w.form.parts())-w.partsShown()))
}

func (w Window) keepingPartInSight() Window {
	line := w.form.partAt
	w.form.partsTop = w.partsWithin(min(line, max(w.partsFrom(), line-w.partsShown()+1)))
	return w
}

func (w Window) scrollingParts(by int) Window {
	total := w.form.spareParts + by
	w.form.partsTop = w.partsWithin(w.partsFrom() + total/SubHeight)
	w.form.spareParts = total % SubHeight
	return w
}

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
