// Package migration holds the change: what it does to the shape of the
// store, what undoes it, what undoing it costs, and whether it has been
// applied yet — and the membrane anything uses to read or build one. It
// knows only domain, for what a store holds and for writing a moment the way
// a person reads one.
package migration

import "time"

// Migration is one change to the shape of the store. It carries the SQL that
// makes it and the SQL that unmakes it, so somebody can read what a tool is
// about to do to their data before letting it, and whether it has been
// applied, and when — a store's ledger is what settles that, not the change
// itself.
type Migration struct {
	number  int
	name    string
	does    string
	undoes  string
	loses   string
	applied bool
	made    time.Time
}

// Told is the membrane of a change: one method per fact it has to tell,
// absence included. Whoever wants to know something about a change
// implements this and is told it; nothing reads a change's insides, and a
// change knows none of them.
type Told interface {
	Numbered(int)
	Named(string)
	Does(string)
	Undoes(string)
	NotUndoable()
	Loses(string)
	NothingLost()
	Applied(time.Time)
	NotApplied()
}

// Tell says every fact this change carries to the given membrane, absence
// included, in the order a card lays them out.
func (m Migration) Tell(to Told) {
	to.Numbered(m.number)
	to.Named(m.name)
	to.Does(m.does)
	if m.undoes == "" {
		to.NotUndoable()
	} else {
		to.Undoes(m.undoes)
	}
	if m.loses == "" {
		to.NothingLost()
	} else {
		to.Loses(m.loses)
	}
	if m.applied {
		to.Applied(m.made)
	} else {
		to.NotApplied()
	}
}
