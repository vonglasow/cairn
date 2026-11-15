package gui

import "cairn/internal/gui/geom"

// Reaches says which character of a wrapped text a point lands on: the line it is down to, counting the lines scrolled out of sight above the first, then the edge between two characters of that line nearest to it.
func (l Lines) Reaches(inside geom.Box, p geom.Point, from int, m Measure, size int) int {
	line := min(len(l.Said)-1, max(0, from+(p.Y-inside.Y)/lineHigh))
	return l.From[line] + m.Nearest(l.Said[line], size, p.X-inside.X)
}

// FieldRow is one line of a form: what the field is called, and what it holds.
func FieldRow(box geom.Box, at int) (name, value geom.Box) {
	return geom.Box{X: box.X + padding, Y: at, W: namesWide - padding, H: lineHigh},
		geom.Box{X: box.X + namesWide, Y: at, W: box.W - namesWide - padding, H: lineHigh}
}

// ChoosingRow is one line of a list: the whole width of the panel, not a form's name column beside its value column.
func ChoosingRow(box geom.Box, at int) geom.Box {
	run := ChoosingRun(box)
	return geom.Box{X: run.X, Y: at - inset, W: run.W, H: lineHigh + inset + inset}
}

// ChoosingLine is where a row's words go: the same width, so what is written hard right of it is not cut off by the bar beside the run.
func ChoosingLine(box geom.Box, at int) geom.Box {
	run := ChoosingRun(box)
	return geom.Box{X: run.X + padding, Y: at, W: run.W - padding - padding, H: lineHigh}
}

// ChoosingLooking is the line at the top of a list that narrows it, above the first row and inside the same panel, which has no heading.
func ChoosingLooking(box geom.Box) geom.Box {
	return geom.Box{X: box.X + padding, Y: box.Y + padding, W: box.W - padding - padding, H: lineHigh + inset + inset}
}

// ChoosingRun is the room a list's rows have: from under the line that narrows them to above the offers along the panel's foot.
func ChoosingRun(box geom.Box) geom.Box {
	top := ChoosingFrom(box) - inset
	return geom.Box{X: box.X + padding, Y: top, W: box.W - padding - padding - barWide - between,
		H: max(apart, PanelOfferAt(box, 0).Y-padding-top)}
}

// ChoosingRows is how many rows fit in that run.
func ChoosingRows(box geom.Box) int { return max(1, ChoosingRun(box).H/apart) }

// ChoosingBar is the track of the bar beside a list, drawn whether or not there is anything to scroll — the same rule the history goes by.
func ChoosingBar(box geom.Box) geom.Box {
	run := ChoosingRun(box)
	return geom.Box{X: run.X + run.W + between, Y: run.Y, W: barWide, H: run.H}
}

// ChoosingAtPoint says which row of a list a point lands on, and whether it lands on one at all.
func ChoosingAtPoint(box geom.Box, rows, from int, p geom.Point) (int, bool) {
	for i := from; i < min(rows, from+ChoosingRows(box)); i++ {
		if ChoosingRow(box, ChoosingFrom(box)+(i-from)*apart).In(p) {
			return i, true
		}
	}
	return 0, false
}

// ChoosingFrom is where a list's own rows begin, under the line that narrows them.
func ChoosingFrom(box geom.Box) int {
	if box.Unsearched {
		return box.Y + padding + inset
	}
	looking := ChoosingLooking(box)
	return looking.Y + looking.H + between
}

// LabelPillAt is where a list of labels draws one as a pill: past the room its tick takes, as wide as its name written small.
func LabelPillAt(line geom.Box, wide int) geom.Box {
	return geom.Box{X: PastCheck(line).X, Y: line.Y + (line.H-pillHigh)/2, W: wide + 2*pillPad, H: pillHigh}
}

const tickRoom = 20

// PastCheck is what is left of a line once its box has taken its room.
func PastCheck(line geom.Box) geom.Box {
	return geom.Box{X: line.X + tickRoom, Y: line.Y, W: line.W - tickRoom, H: line.H}
}
