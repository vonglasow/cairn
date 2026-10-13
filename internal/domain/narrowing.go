package domain

// Narrowing is what a read was told to keep: facts named one at a time, and a fact never named narrows nothing — so a Narrowing nobody named anything on reads everything.
type Narrowing struct {
	told    []func(NarrowingTold)
	narrows bool
}

// NarrowingTold is what a narrowing tells whoever reads by it.
type NarrowingTold interface {
	InColumn(state string)
	Carrying(priority string)
	DueBefore(day string)
	Serving(objective int)
	Labelled(labels []string)
	Archived()
	WithArchived()
}

func (n *Narrowing) name(narrows bool, tell func(NarrowingTold)) {
	n.told = append(n.told, tell)
	n.narrows = n.narrows || narrows
}

// InColumn keeps what sits in one column.
func (n *Narrowing) InColumn(state string) {
	n.name(true, func(to NarrowingTold) { to.InColumn(state) })
}

// Carrying keeps what carries one priority.
func (n *Narrowing) Carrying(priority string) {
	n.name(true, func(to NarrowingTold) { to.Carrying(priority) })
}

// DueBefore keeps what is due on or before a day.
func (n *Narrowing) DueBefore(day string) {
	n.name(true, func(to NarrowingTold) { to.DueBefore(day) })
}

// Serving keeps what serves one objective.
func (n *Narrowing) Serving(objective int) {
	n.name(true, func(to NarrowingTold) { to.Serving(objective) })
}

// Labelled keeps what carries every label named.
func (n *Narrowing) Labelled(labels []string) {
	n.name(true, func(to NarrowingTold) { to.Labelled(labels) })
}

// Archived keeps what was taken off the board, instead of what is on it.
func (n *Narrowing) Archived() {
	n.name(true, func(to NarrowingTold) { to.Archived() })
}

// WithArchived keeps both sides of the board, which is what a board asks for: it draws the two apart rather than reading twice.
func (n *Narrowing) WithArchived() {
	n.name(false, func(to NarrowingTold) { to.WithArchived() })
}

// Narrows says whether anything was named that keeps less than everything.
func (n Narrowing) Narrows() bool { return n.narrows }

// Tell says every fact named, in the order it was named.
func (n Narrowing) Tell(to NarrowingTold) {
	for _, tell := range n.told {
		tell(to)
	}
}
