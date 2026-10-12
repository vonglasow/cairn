package gui

// A panel is the size of what it holds. The list a field answers with took
// three quarters of the form whatever it offered: four priorities sat in a
// panel made for fifty tasks. A list now hangs under its field, as wide as
// what it says and as tall as its rows, up to what the window has room for,
// and scrolls past that; the list of backups is as long as the backups.

// listLeast is the narrowest a list is drawn: room to read what is typed into
// the line that narrows it.
const listLeast = 240

// ChoosingFor is the panel of a list at a corner, as wide as given and as tall
// as a run of rows needs between its heading and its offers.
func ChoosingFor(x, y, wide, rows int) Box {
	top := ChoosingFrom(Box{X: x, Y: y, W: wide}) - inset
	return Box{X: x, Y: y, W: wide, H: top - y + rows*apart + padding + barHigh + padding}
}

// ChoosingBox is the panel of the list open on a form: under the field it
// answers, or above it when the window has no room below, and inside the
// panel the form is drawn in.
func (w Window) ChoosingBox() Box {
	over := Over(w.width, w.height)
	most := max(1, (over.H-ChoosingFor(0, 0, 0, 0).H)/apart)
	rows := min(max(1, len(w.listed())), most)
	wide := min(over.W, w.choosingWide())
	high := ChoosingFor(0, 0, wide, rows).H
	at := w.hung(wide, high)
	return ChoosingFor(at.X, at.Y, wide, rows)
}

// Hung is where a panel opened from a field goes: a gap under it when it ends
// inside the panel there, a gap above it otherwise. Above, it may start over
// the panel's top, and is held there. The field starts inside the panel and
// what hangs from it is never wider than the panel, so only its right end can
// run out.
func Hung(field, over Box, wide, high int) Point {
	y := field.Y + field.H + between
	if y+high > over.Y+over.H {
		y = field.Y - between - high
	}
	return Point{X: min(field.X, over.X+over.W-wide), Y: max(over.Y, y)}
}

// listed is what the list offers before anything narrows it. The panel is
// sized on it, so it does not shrink under the typing and move every row and
// offer from under the pointer.
func (w Window) listed() []Choice {
	w.filtering = Written{}
	return w.Choosing()
}

// choosingWide is how wide a list must be to say all it holds: its widest row
// with the state written beside it, its heading, and its offers.
func (w Window) choosingWide() int {
	m := w.measured()
	text := ends(m, w.EmptyListSays(), reads)
	for _, one := range w.listed() {
		wide := ends(m, w.RowSays(one), reads)
		if one.State != "" {
			wide += padding + len([]rune(one.State))*perGlyph
		}
		text = max(text, wide)
	}
	// A row's words sit a padding inside the run, which sits a padding inside
	// the panel, with the bar beside it.
	rows := text + 4*padding + barWide + between
	return max(listLeast, rows, w.headWide(), w.offersWide())
}

// BackupsBox is the panel of the list of backups: across the middle of the
// window, as long as the backups up to what the window has room for.
func (w Window) BackupsBox() Box {
	over := Over(w.width, w.height)
	m := w.measured()
	text := 0
	for _, one := range w.backups {
		text = max(text, ends(m, backupSays(one), reads))
	}
	wide := min(over.W, max(listLeast, text+4*padding, w.headWide(), w.offersWide()))
	fixed := FirstRow(Box{}) + between + padding + barHigh + padding
	rows := min(max(1, len(w.backups)), max(1, (over.H-fixed)/menuRow))
	high := fixed + rows*menuRow
	return Box{X: (w.width - wide) / 2, Y: over.Y + (over.H-high)/2, W: wide, H: high}
}

// headWide is how wide a panel must be to write its heading whole.
func (w Window) headWide() int { return ends(w.measured(), Heads(w), heading) + 2*padding }

// offersWide is how wide a panel must be to lay its offers along its foot.
func (w Window) offersWide() int {
	wide := 2 * padding
	for _, one := range w.offerWidths() {
		wide += one
	}
	return wide
}

// ends is how wide a line is written at a size: where its last character ends.
func ends(m Measure, said string, size int) int {
	edges := m(said, size)
	return edges[len(edges)-1]
}
