package gui

const listLeast = 240

// ChoosingFor is the panel of a list at a corner, as wide as given and as tall as a run of rows needs between its heading and its offers.
func ChoosingFor(x, y, wide, rows int) Box {
	top := ChoosingFrom(Box{X: x, Y: y, W: wide}) - inset
	return Box{X: x, Y: y, W: wide, H: top - y + rows*apart + padding + barHigh + padding}
}

// ChoosingBox is the panel of the list open on a form: under the field it answers, or above it when the window has no room below, and inside the panel the form is drawn in.
func (w Window) ChoosingBox() Box {
	over := Over(w.width, w.height)
	most := max(1, (over.H-ChoosingFor(0, 0, 0, 0).H)/apart)
	rows := min(max(1, len(w.listed())), most)
	wide := min(over.W, w.choosingWide())
	high := ChoosingFor(0, 0, wide, rows).H
	at := w.hung(wide, high)
	return ChoosingFor(at.X, at.Y, wide, rows)
}

// Hung is where a panel opened from a field goes: a gap under it when it ends inside the panel there, a gap above it otherwise.
func Hung(field, over Box, wide, high int) Point {
	y := field.Y + field.H + between
	if y+high > over.Y+over.H {
		y = field.Y - between - high
	}
	return Point{X: min(field.X, over.X+over.W-wide), Y: max(over.Y, y)}
}

func (w Window) listed() []Choice {
	w.filtering = Written{}
	return w.Choosing()
}

func (w Window) choosingWide() int {
	m := w.measured()
	text := ends(m, w.EmptyListSays(), reads)
	for _, one := range w.listed() {
		text = max(text, ends(m, w.RowSays(one), reads))
	}
	rows := text + 4*padding + barWide + between
	return max(listLeast, rows, w.headWide(), w.offersWide())
}

// BackupsBox is the panel of the list of backups: across the middle of the window, as long as the backups up to what the window has room for.
func (w Window) BackupsBox() Box {
	said := make([]string, 0, len(w.backups))
	for _, one := range w.backups {
		said = append(said, backupSays(one))
	}
	return w.listBox(said)
}

func (w Window) listBox(said []string) Box {
	over := Over(w.width, w.height)
	m := w.measured()
	text := 0
	for _, one := range said {
		text = max(text, ends(m, one, reads))
	}
	wide := min(over.W, max(listLeast, text+4*padding, w.headWide(), w.offersWide()))
	fixed := FirstRow(Box{}) + between + padding + barHigh + padding
	rows := min(max(1, len(said)), max(1, (over.H-fixed)/menuRow))
	high := fixed + rows*menuRow
	return Box{X: (w.width - wide) / 2, Y: over.Y + (over.H-high)/2, W: wide, H: high}
}

func (w Window) headWide() int { return ends(w.measured(), Heads(w), heading) + 2*padding }

func (w Window) offersWide() int {
	wide := 2 * padding
	for _, one := range w.offerWidths() {
		wide += one
	}
	return wide
}

// WordsOf is the part of a box its words take: from its start, as wide as they are written at that size, never wider than the box.
func WordsOf(m Measure, b Box, said string, size int) Box {
	return Box{X: b.X, Y: b.Y, W: min(b.W, ends(m, said, size)), H: b.H}
}

func ends(m Measure, said string, size int) int {
	edges := m(said, size)
	return edges[len(edges)-1]
}
