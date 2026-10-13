package gui

// HelpRows is the help as the lines it takes: each key on the first line of what it does, and what does not fit in the column on lines of its own under it, with no key.
func (w Window) HelpRows() []Said {
	room := helpValueWide(Over(w.width, w.height))
	var rows []Said
	for _, line := range w.Helping() {
		for i, piece := range w.measured().Wrapped(line.Means, room, reads).Said {
			key := ""
			if i == 0 {
				key = line.Key
			}
			rows = append(rows, Said{Key: key, Means: piece})
		}
	}
	return rows
}

// HelpBox is the panel the help is drawn in: as wide as every panel over the board, and as tall as what it says, up to the room there is.
func (w Window) HelpBox() Box {
	over := Over(w.width, w.height)
	fixed := FirstRow(over) - over.Y + padding + barHigh + padding
	over.H = min(over.H, fixed+len(w.HelpRows())*apart)
	return over
}

// HelpRoom is how many lines of the help are in sight at once.
func (w Window) HelpRoom() int {
	box := w.HelpBox()
	fits := (PanelOfferAt(box, 0).Y - padding - FirstRow(box)) / apart
	return min(len(w.HelpRows()), max(1, fits))
}

// HelpFrom is the first line of the help in sight.
func (w Window) HelpFrom() int { return w.settledHelp(w.helpFrom) }

func (w Window) settledHelp(want int) int {
	return min(max(0, len(w.HelpRows())-w.HelpRoom()), max(0, want))
}

// HelpRowAt is where the i-th line in sight of the help is written: its key, and what it does.
func HelpRowAt(box Box, i int) (name, value Box) {
	name, value = FieldRow(box, FirstRow(box)+i*apart)
	value.W = helpValueWide(box)
	return name, value
}

func helpValueWide(box Box) int {
	_, value := FieldRow(box, 0)
	return value.W - between - barWide
}

// HelpBar is the track of the bar beside the help, and whether there is one: there is when the help holds more lines than are in sight.
func (w Window) HelpBar() (Box, bool) {
	box, room := w.HelpBox(), w.HelpRoom()
	if room >= len(w.HelpRows()) {
		return Box{}, false
	}
	return Box{
		X: box.X + box.W - padding - barWide,
		Y: FirstRow(box),
		W: barWide,
		H: (room-1)*apart + lineHigh,
	}, true
}

func (w Window) inHelp(name string) Window {
	switch name {
	case "down", "j":
		w.helpFrom = w.settledHelp(w.HelpFrom() + 1)
	case "up", "k":
		w.helpFrom = w.settledHelp(w.HelpFrom() - 1)
	}
	return w
}
