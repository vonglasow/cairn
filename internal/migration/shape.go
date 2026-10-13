package migration

import "cairn/internal/domain"

// Shape is how a change crosses to a machine.
type Shape struct {
	Number  int    `json:"number"`
	Name    string `json:"name"`
	Applied bool   `json:"applied"`
	Made    string `json:"made,omitempty"`
}

// OfMigration turns a change into what crosses.
func OfMigration(m Migration) Shape {
	f := textOf(m)
	made := ""
	if f.Applied {
		made = domain.FormatMoment(f.Made)
	}
	return Shape{Number: f.Number, Name: f.Name, Applied: f.Applied, Made: made}
}

// RolledBack names the change a rollback unmade, and what that cost.
type RolledBack struct {
	RolledBack string `json:"rolled_back"`
	Lost       string `json:"lost"`
}

// OfRolledBack turns the change a rollback unmade into what crosses.
func OfRolledBack(m Migration) RolledBack {
	f := textOf(m)
	return RolledBack{RolledBack: f.Name, Lost: f.Loses}
}
