// Package migration holds the change: what it does to the shape of the store, what undoes it, what undoing it costs, and whether it has been applied yet — and the membrane anything uses to read or build one.
package migration

import (
	"errors"
	"time"
)

// Migration is one change to the shape of the store.
type Migration struct {
	number  int
	name    string
	does    string
	undoes  string
	loses   string
	applied bool
	made    time.Time
}

// Carried is everything a change carries, as plain facts.
type Carried struct {
	Number                    int
	Name, Does, Undoes, Loses string
	Applied                   bool
	Made                      time.Time
}

// Carried hands over what the change carries.
func (m Migration) Carried() Carried {
	return Carried{Number: m.number, Name: m.name, Does: m.does, Undoes: m.undoes, Loses: m.loses, Applied: m.applied, Made: m.made}
}

// Made is the change that carries what is given.
func Made(c Carried) Migration {
	return Migration{number: c.Number, name: c.Name, does: c.Does, undoes: c.Undoes, loses: c.Loses, applied: c.Applied, made: c.Made}
}

// What can be wrong with the shape of a store.
var (
	// ErrNewerStore: an older binary cannot know what a later one wrote, and writing into it anyway is how a store gets quietly wrecked.
	ErrNewerStore = errors.New("the store comes from a newer cairn")

	// ErrStoreBehind: the store is missing a change this cairn needs.
	ErrStoreBehind = errors.New("the store is behind this cairn")

	// ErrNoSuchChange: a number that names no change.
	ErrNoSuchChange = errors.New("there is no such change")

	// ErrNothingWaiting: everything this cairn knows is already applied.
	ErrNothingWaiting = errors.New("there is nothing waiting to be applied")

	// ErrAlreadyApplied: a change named by number that the store already has.
	ErrAlreadyApplied = errors.New("the change is already applied")

	// ErrNothingToRollBack: only the shape a store starts with is left, and that one is not a change on top of something else — undoing it would mean throwing the store away, which is what deleting the file is for.
	ErrNothingToRollBack = errors.New("there is nothing left to roll back")

	// ErrUndoneMeanwhile: the change read as the last applied was undone by another cairn between the reading and the rollback.
	ErrUndoneMeanwhile = errors.New("the change was rolled back by somebody else meanwhile")

	// ErrEarlierWaiting: a change was asked for while one before it is still waiting.
	ErrEarlierWaiting = errors.New("an earlier change is still waiting")
)
