package gui

import (
	"cairn/internal/board"
	"cairn/internal/gui/design"
	"cairn/internal/gui/geom"
)

const listLeast = 240

// ChoosingFor is the panel of a list at a corner, as wide as given and as tall as a run of rows needs between its heading and its offers.
func ChoosingFor(x, y, wide, rows int, searched bool) geom.Box {
	top := ChoosingFrom(geom.Box{X: x, Y: y, W: wide, Unsearched: !searched}) - inset
	return geom.Box{X: x, Y: y, W: wide, H: top - y + rows*apart + padding + geom.BarHigh + padding, Unsearched: !searched}
}

// ChoosingBox is the panel of the list open on a form: under the field it answers, or above it when the window has no room below, and inside the panel the form is drawn in.
func (w Window) ChoosingBox() geom.Box {
	over := Over(w.width, w.height)
	searched := w.ChoosingSearched()
	most := max(1, (over.H-ChoosingFor(0, 0, 0, 0, searched).H)/apart)
	rows := min(max(1, len(w.listed())), most)
	wide := min(over.W, w.choosingWide())
	if slot, ok := w.chosenSlot(); ok && w.PickingLabels {
		return w.underTheCard(slot, wide, rows, searched)
	}
	high := ChoosingFor(0, 0, wide, rows, searched).H
	at := w.hung(wide, high)
	return ChoosingFor(at.X, at.Y, wide, rows, searched)
}

func (w Window) underTheCard(slot geom.Box, wide, rows int, searched bool) geom.Box {
	below := w.height - geom.Margin - (slot.Y + slot.H + between)
	above := slot.Y - between - geom.Margin
	bare := ChoosingFor(0, 0, wide, 0, searched).H
	room := max(below, above)
	rows = max(1, min(rows, (room-bare)/apart))
	high := ChoosingFor(0, 0, wide, rows, searched).H
	x := max(geom.Margin, min(slot.X, w.width-geom.Margin-wide))
	if below >= above {
		return ChoosingFor(x, slot.Y+slot.H+between, wide, rows, searched)
	}
	return ChoosingFor(x, slot.Y-between-high, wide, rows, searched)
}

// Hung is where a panel opened from a field goes: a gap under it when it ends inside the panel there, a gap above it otherwise.
func Hung(field, over geom.Box, wide, high int) geom.Point {
	y := field.Y + field.H + between
	if y+high > over.Y+over.H {
		y = field.Y - between - high
	}
	return geom.Point{X: min(field.X, over.X+over.W-wide), Y: max(over.Y, y)}
}

func (w Window) listed() []Choice {
	w.Filtering = board.Written{}
	return w.Choosing()
}

func (w Window) choosingWide() int {
	m := w.measured()
	text := ends(m, w.EmptyListSays(), reads)
	for _, one := range w.listed() {
		text = max(text, tickRoom+ends(m, w.RowSays(one), reads))
	}
	rows := text + 4*padding + barWide + between
	return max(listLeast, rows, w.headWide(), w.offersWide())
}

func (w Window) backupsOpens() geom.Box {
	said := make([]string, 0, len(w.Backups))
	for _, one := range w.Backups {
		said = append(said, backupSays(one))
	}
	return w.listBox(said)
}

func (w Window) listBox(said []string) geom.Box {
	over := Over(w.width, w.height)
	m := w.measured()
	text := 0
	for _, one := range said {
		text = max(text, ends(m, one, reads))
	}
	wide := min(over.W, max(listLeast, text+4*padding, w.headWide(), w.offersWide()))
	fixed := FirstRow(geom.Box{}) + between + padding + geom.BarHigh + padding
	rows := min(max(1, len(said)), max(1, (over.H-fixed)/menuRow))
	high := fixed + rows*menuRow
	return geom.Box{X: (w.width - wide) / 2, Y: over.Y + (over.H-high)/2, W: wide, H: high}
}

func (w Window) headWide() int {
	return ends(w.headMeasured(), Heads(w), design.HeadingSize) + 2*padding
}

func (w Window) offersWide() int {
	wide := 2 * padding
	for _, one := range w.offerWidths() {
		wide += one
	}
	return wide
}

// WordsOf is the part of a box its words take: from its start, as wide as they are written at that size, never wider than the box.
func WordsOf(m Measure, b geom.Box, said string, size int) geom.Box {
	return geom.Box{X: b.X, Y: b.Y, W: min(b.W, ends(m, said, size)), H: b.H}
}

func ends(m Measure, said string, size int) int {
	edges := m(said, size)
	return edges[len(edges)-1]
}

// BackupsBox is the panel of the list of backups: across the middle of the window, as long as the backups up to what the window has room for.
func (w Window) BackupsBox() geom.Box { return w.placedPanel(board.OnBackups, w.backupsOpens()) }
