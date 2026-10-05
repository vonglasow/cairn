package gui

// Where the parts of the list a field answers with go: its rows, the line
// it filters by, and its bar.

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

// Reaches says which character of a wrapped text a point lands on: the line it
// is down to, counting the lines scrolled out of sight above the first, then
// the edge between two characters of that line nearest to it. Reading only how
// far across it was put every click on the first line, so the third line of a
// description could not be reached with a pointer at all.
func (l Lines) Reaches(inside Box, p Point, from int, m Measure, size int) int {
	line := min(len(l.Said)-1, max(0, from+(p.Y-inside.Y)/lineHigh))
	return l.From[line] + m.Nearest(l.Said[line], size, p.X-inside.X)
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
	run := ChoosingRun(box)
	return Box{X: run.X, Y: at - inset, W: run.W, H: lineHigh + inset + inset}
}

// ChoosingLine is where a row's words go: the same width, so what is written
// hard right of it is not cut off by the bar beside the run.
func ChoosingLine(box Box, at int) Box {
	run := ChoosingRun(box)
	return Box{X: run.X + padding, Y: at, W: run.W - padding - padding, H: lineHigh}
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
