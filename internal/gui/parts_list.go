package gui

import (
	"cairn/internal/board"
	"cairn/internal/gui/geom"
	"slices"
	"strconv"

	"cairn/internal/engine"
)

func (w Window) tickingPart() Window {
	parts := w.Form.Parts()
	if w.Form.PartAt >= len(parts) {
		return w
	}
	together := w.Form.TickedTogether()
	done := slices.ContainsFunc(together, func(at int) bool { return !parts[at].Done })
	for _, at := range together {
		if w = w.tickingOne(at, done); w.Err != nil {
			return w
		}
	}
	return w
}

func (w Window) tickingOne(at int, done bool) Window {
	before := w.Form
	w.Form = w.Form.Ticking(at, done)
	part := w.Form.Parts()[at]
	if part.Number == 0 {
		return w
	}
	ticked := "not-done"
	if done {
		ticked = "done"
	}
	w = w.asking(engine.StepEdit, engine.Input{Given: map[string]string{
		"number": strconv.Itoa(w.Form.Number), "step": strconv.Itoa(part.Number), ticked: "true"}})
	if w.Err != nil {
		w.Form = before
		return w
	}
	w.Form = before.Ticking(at, done).OpenedWith(part)
	return w
}

func (w Window) askingToDelete() Window {
	switch taken := w.Form.SelectedParts(); len(taken) {
	case 0:
	case 1:
		w.Form.PartAt = taken[0]
		w.Screen, w.Question = board.OnAsking, board.AboutDeletingPart
	default:
		w.Screen, w.Question = board.OnAsking, board.AboutDeletingParts
	}
	return w
}

func (w Window) deletingPart() Window {
	w.Screen, w.Question = board.OnForm, board.AboutDeleting
	parts := w.Form.Parts()
	at := w.Form.PartAt
	if at >= len(parts) {
		return w
	}
	gone := parts[at].Number
	rest := slices.Delete(parts, at, at+1)
	if gone == 0 {
		w.Form = w.Form.WithParts(rest).TakingTheOneOn()
		return w
	}
	before := w.Form
	field, _ := w.Form.PartsField()
	w.Form = w.Form.WithParts(board.Without(rest, gone)).TakingTheOneOn()
	w.Form.Opened = slices.Clone(w.Form.Opened)
	w.Form.Opened[field] = board.PartsWritten(board.Without(engine.PartsIn(w.Form.Opened[field]), gone))
	w.Err = nil
	w = w.asking(engine.StepDelete, engine.Input{Given: map[string]string{
		"number": strconv.Itoa(w.Form.Number), "step": strconv.Itoa(gone)}})
	if w.Err != nil {
		w.Form = before
	}
	return w
}

func (w Window) lettingGoOfAPart(p geom.Point) Window {
	from, moved := w.Form.Carried-1, w.held.Moved
	w.Form.Carried, w.held = 0, Held{}
	if moved {
		if to, on := w.partUnder(p); on && slices.Contains(w.Form.TakenParts(), from) {
			w.Form = w.Form.CarryingSeveral(to)
		} else if on {
			w.Form = w.Form.Carrying(from, to).TakingTheOneOn()
		}
		return w
	}
	if !w.gathering && !w.stretching {
		w.Form = w.Form.TakingTheOneOn()
	}
	if from == w.Form.Lines()-1 {
		w.Form = w.Form.Naming()
		return w
	}
	tiles := w.partTiles()
	if slices.ContainsFunc(tiles[:len(tiles)-1], func(tile geom.Box) bool { return SubCheck(tile).In(p) }) {
		return w.tickingPart()
	}
	return w
}

func (w Window) inHand() (int, bool) {
	from := w.Form.Carried - 1
	return from, from >= 0 && from < len(w.Form.Parts()) && w.held.Moved
}

// PartGhost is the step in hand: its own tile's size, moving only up and down, held where it was taken.
func (w Window) PartGhost() (geom.Box, bool) {
	from, ok := w.inHand()
	if !ok {
		return geom.Box{}, false
	}
	tiles := w.partTiles()
	tile := tiles[0]
	if at := from - w.partsFrom(); at >= 0 && at < len(tiles)-1 {
		tile = tiles[at]
		return geom.Box{X: tile.X, Y: tile.Y + w.held.At.Y - w.held.From.Y, W: tile.W, H: tile.H}, true
	}
	return geom.Box{X: tile.X, Y: w.held.At.Y - geom.SubHeight/2, W: tile.W, H: geom.SubHeight}, true
}

// PartLanding is the line saying where the step in hand would land: over the tile it would take the place of when carried up or held over its own, under it when carried down — which is where it ends up, the others moving to make room — and over the line that adds when it would go last.
func (w Window) PartLanding() (geom.Box, bool) {
	from, ok := w.inHand()
	to, on := w.partUnder(w.held.At)
	if !ok || !on {
		return geom.Box{}, false
	}
	tiles := w.partTiles()
	if to >= len(w.Form.Parts()) {
		adding := tiles[len(tiles)-1]
		return geom.Box{X: tiles[0].X, Y: adding.Y, W: tiles[0].W, H: landingHigh}, true
	}
	if to == from {
		return geom.Box{}, false
	}
	tile := tiles[to-w.partsFrom()]
	if to > from {
		return geom.Box{X: tile.X, Y: tile.Y + tile.H + 1 - landingHigh/2, W: tile.W, H: landingHigh}, true
	}
	return geom.Box{X: tile.X, Y: tile.Y - 1 - landingHigh/2, W: tile.W, H: landingHigh}, true
}

func (w Window) onParts(name string) (Window, bool) {
	if w.Form.WritingPart {
		switch name {
		case "enter":
			w.Form = w.Form.Committing()
		case "backspace":
			w.Form.Named = w.Form.Named.Back()
		case "left":
			w.Form.Named = w.Form.Named.Left()
		case "right":
			w.Form.Named = w.Form.Named.Right()
		case "cmd+a":
			w.Form.Named = w.Form.Named.All()
		default:
			return w, false
		}
		return w, true
	}
	switch name {
	case "down":
		w.Form = w.Form.WalkingParts(1)
	case "up":
		w.Form = w.Form.WalkingParts(-1)
	case "space":
		w = w.tickingPart()
	case "enter":
		w.Form = w.Form.Naming()
	case "backspace", "delete-forward":
		w = w.askingToDelete()
	case "cmd+a":
		w.Form = w.Form.TakingAll()
	default:
		return w, false
	}
	return w, true
}

func (w Window) partTiles() []geom.Box {
	at, ok := w.Form.PartsField()
	if !ok {
		return nil
	}
	_, value, rows := FormPlace(w.formBox(), w.Form.Fields, at)
	tiles, adding := TickedPlaces(Framed(value, rows), w.Form.Lines()-1, w.partsFrom())
	return append(tiles, adding)
}

func (w Window) partsFrame() (geom.Box, bool) {
	at, ok := w.Form.PartsField()
	if !ok {
		return geom.Box{}, false
	}
	_, value, rows := FormPlace(w.formBox(), w.Form.Fields, at)
	return Framed(value, rows), true
}

func (w Window) partsShown() int {
	frame, _ := w.partsFrame()
	return max(1, ticksFit(frame))
}

func (w Window) partsFrom() int { return w.partsWithin(w.Form.PartsTop) }

func (w Window) partsWithin(top int) int {
	return max(0, min(top, len(w.Form.Parts())-w.partsShown()))
}

func (w Window) keepingPartInSight() Window {
	line := w.Form.PartAt
	w.Form.PartsTop = w.partsWithin(min(line, max(w.partsFrom(), line-w.partsShown()+1)))
	return w
}

func (w Window) scrollingParts(by int) Window {
	total := w.Form.SpareParts + by
	w.Form.PartsTop = w.partsWithin(w.partsFrom() + total/geom.SubHeight)
	w.Form.SpareParts = total % geom.SubHeight
	return w
}

func (w Window) partUnder(p geom.Point) (int, bool) {
	tiles := w.partTiles()
	for i, tile := range tiles {
		if !tile.In(p) {
			continue
		}
		if i == len(tiles)-1 {
			return w.Form.Lines() - 1, true
		}
		return i + w.partsFrom(), true
	}
	return 0, false
}

func (w Window) deletingParts() Window {
	w.Screen, w.Question = board.OnForm, board.AboutDeleting
	taken, parts := w.Form.TakenParts(), w.Form.Parts()
	var gone []int
	var rest []engine.Part
	for i, p := range parts {
		if !slices.Contains(taken, i) {
			rest = append(rest, p)
		} else if p.Number != 0 {
			gone = append(gone, p.Number)
		}
	}
	renumbered := func(ps []engine.Part) []engine.Part {
		for k := len(gone) - 1; k >= 0; k-- {
			ps = board.Without(ps, gone[k])
		}
		return ps
	}
	before := w.Form
	w.Form = w.Form.WithParts(renumbered(rest))
	w.Form.PartAt = min(taken[0], len(rest))
	w.Form = w.Form.TakingTheOneOn()
	if len(gone) == 0 {
		return w
	}
	field, _ := w.Form.PartsField()
	kept := board.PartsWritten(renumbered(engine.PartsIn(before.Opened[field])))
	w.Form.Opened = slices.Clone(w.Form.Opened)
	w.Form.Opened[field] = kept
	w.Err = nil
	w = w.asking(engine.TaskEdit, engine.Input{Given: map[string]string{"number": strconv.Itoa(w.Form.Number), "steps": kept}})
	if w.Err != nil {
		w.Form = before
	}
	return w
}
