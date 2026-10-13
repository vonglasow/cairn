// Package migration holds the change: what it does to the shape of the store, what undoes it, what undoing it costs, and whether it has been applied yet — and the membrane anything uses to read or build one.
package migration

import "time"

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
