package migration

import "time"

// AsTold builds a change from what it is told. It is the same membrane the
// other way round: whoever has a change's facts — the raw list, or the
// ledger a store keeps beside it — tells them, and is handed the change they
// make. Nothing has to reach inside a change to put one together, which is
// what lets a change keep its insides to itself.
type AsTold struct {
	migration Migration
}

// Numbered is told the change's number.
func (a *AsTold) Numbered(n int) { a.migration.number = n }

// Named is told the change's name.
func (a *AsTold) Named(s string) { a.migration.name = s }

// Does is told the SQL that makes the change.
func (a *AsTold) Does(s string) { a.migration.does = s }

// Undoes is told the SQL that unmakes the change.
func (a *AsTold) Undoes(s string) { a.migration.undoes = s }

// NotUndoable is told the change cannot be undone.
func (a *AsTold) NotUndoable() { a.migration.undoes = "" }

// Loses is told what undoing the change costs.
func (a *AsTold) Loses(s string) { a.migration.loses = s }

// NothingLost is told that undoing the change costs nothing.
func (a *AsTold) NothingLost() { a.migration.loses = "" }

// Applied is told the change has been applied, and when.
func (a *AsTold) Applied(t time.Time) {
	a.migration.applied = true
	a.migration.made = t
}

// NotApplied is told the change has not been applied.
func (a *AsTold) NotApplied() {
	a.migration.applied = false
	a.migration.made = time.Time{}
}

// Migration hands back the change AsTold was told to build.
func (a *AsTold) Migration() Migration { return a.migration }
