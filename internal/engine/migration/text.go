package migration

import (
	"cairn/internal/engine/backup"
	"cairn/internal/engine/calendar"
	"cairn/internal/engine/item"
	"fmt"
	"strings"
)

func textOf(m Migration) Carried { return m.Carried() }

func state(f Carried) string {
	if f.Applied {
		return "applied " + calendar.FormatMoment(f.Made)
	}
	return "waiting          "
}

// Text renders one change in full: what it is, what it does, and what undoing it does.
func Text(m Migration) string {
	f := textOf(m)
	var out strings.Builder
	fmt.Fprintf(&out, "#%d  %s  %s\n\n", f.Number, state(f), f.Name)
	fmt.Fprintf(&out, "what it does\n%s\n\n", item.OrNothing(strings.TrimSpace(f.Does)))
	fmt.Fprintf(&out, "what undoing it does\n%s\n\n", item.OrNothing(strings.TrimSpace(f.Undoes)))
	fmt.Fprintf(&out, "what undoing it loses\n%s\n", item.OrNothing(f.Loses))
	return out.String()
}

// List renders every change, applied or waiting: a list that only showed what was done could not tell you what to apply.
func List(all []Migration, held backup.Held) string {
	waiting := 0
	var out strings.Builder
	for _, m := range all {
		f := textOf(m)
		if !f.Applied {
			waiting++
		}
		fmt.Fprintf(&out, "#%d  %s  %s\n", f.Number, state(f), f.Name)
	}
	fmt.Fprintf(&out, "\nthe store %s, %d change(s) waiting\n", backup.HeldText(held), waiting)
	return out.String()
}

// RolledBackText names the change a rollback unmade, and what that cost.
func RolledBackText(m Migration) string {
	f := textOf(m)
	var out strings.Builder
	fmt.Fprintf(&out, "rolled back %q\n", f.Name)
	if f.Loses != "" {
		fmt.Fprintf(&out, "lost: %s — `cairn backup restore --yes` puts it back\n", f.Loses)
	}
	return out.String()
}
