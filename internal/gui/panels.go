package gui

// PanelLeast is the smallest a panel is resized to: room for a heading, a few lines and the buttons at its foot.
var PanelLeast = Point{X: 360, Y: 240}

type panelPlace struct {
	by   Point
	size [2]int
}

func movable(s Screen) bool {
	return s == OnConfiguration || s == OnHelp || s == OnLabels || s == OnBackups || s == OnViews
}

// PanelBox is the panel open on top of the board, wherever it was moved and however it was resized.
func (w Window) PanelBox() Box { return w.panel() }

func (w Window) placedPanel(s Screen, opens Box) Box {
	at := w.panels[s]
	if at.size != ([2]int{}) {
		opens.W = max(min(PanelLeast.X, opens.W), min(at.size[0], w.width))
		opens.H = max(min(PanelLeast.Y, opens.H), min(at.size[1], w.height))
	}
	return Shifted(opens, at.by, w.width, w.height)
}

func (w Window) placingPanel(s Screen, at panelPlace) Window {
	kept := make(map[Screen]panelPlace, len(w.panels)+1)
	for screen, place := range w.panels {
		kept[screen] = place
	}
	kept[s] = at
	w.panels = kept
	return w
}

func (w Window) takingPanel(p Point) (Window, bool) {
	if !movable(w.screen) {
		return w, false
	}
	box := w.panel()
	switch {
	case FormGrip(box).In(p):
		w.resizingPanel, w.resizedFrom, w.held = true, box, Held{From: p, At: p}
	case Box{X: box.X, Y: box.Y, W: box.W, H: dropHead}.In(p):
		w.movingPanel, w.dragged, w.formGrip, w.held = true, false, p, Held{From: p, At: p}
	default:
		return w, false
	}
	return w, true
}

func (w Window) carryingPanel(p Point) Window {
	at := w.panels[w.screen]
	if w.resizingPanel {
		at.size = [2]int{max(PanelLeast.X, w.resizedFrom.W+p.X-w.held.From.X), max(PanelLeast.Y, w.resizedFrom.H+p.Y-w.held.From.Y)}
		return w.placingPanel(w.screen, at)
	}
	box := w.panel()
	moved := Shifted(box, Point{X: p.X - w.formGrip.X, Y: p.Y - w.formGrip.Y}, w.width, w.height)
	at.by = Point{X: at.by.X + moved.X - box.X, Y: at.by.Y + moved.Y - box.Y}
	w.formGrip = p
	if abs(p.X-w.held.From.X) > Wobble || abs(p.Y-w.held.From.Y) > Wobble {
		w.dragged = true
	}
	return w.placingPanel(w.screen, at)
}

func (w Window) lettingGoOfPanel(p Point) Window {
	moving := w.movingPanel
	w.movingPanel, w.resizingPanel, w.held, w.pressed = false, false, Held{}, false
	if moving && !w.dragged {
		return w.Click(p)
	}
	return w
}
