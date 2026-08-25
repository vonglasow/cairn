package migration

import "cairn/internal/domain"

// Shape is how a change crosses to a machine. The field names are the wire,
// so they are written here once rather than derived from the entity — an
// entity that knew its wire format would drag a protocol into every test.
type Shape struct {
	Number  int    `json:"number"`
	Name    string `json:"name"`
	Applied bool   `json:"applied"`
	Made    string `json:"made,omitempty"`
}

// OfMigration turns a change into what crosses. Made is left out until it is
// applied: a change waiting has no moment to report yet.
func OfMigration(m Migration) Shape {
	f := textOf(m)
	made := ""
	if f.applied {
		made = domain.FormatMoment(f.made)
	}
	return Shape{Number: f.number, Name: f.name, Applied: f.applied, Made: made}
}

// RolledBack names the change a rollback unmade, and what that cost. The
// cost is said here rather than left to be worked out from the SQL: somebody
// who rolls back to try something is exactly somebody who has not read the
// schema.
type RolledBack struct {
	RolledBack string `json:"rolled_back"`
	Lost       string `json:"lost"`
}

// OfRolledBack turns the change a rollback unmade into what crosses.
func OfRolledBack(m Migration) RolledBack {
	f := textOf(m)
	return RolledBack{RolledBack: f.name, Lost: f.loses}
}
