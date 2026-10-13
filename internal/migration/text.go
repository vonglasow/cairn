package migration

import (
	"fmt"
	"strings"
	"time"

	"cairn/internal/domain"
)

type textFacts struct {
	number                    int
	name, does, undoes, loses string
	applied                   bool
	made                      time.Time
}

func (f *textFacts) Numbered(n int)  { f.number = n }
func (f *textFacts) Named(s string)  { f.name = s }
func (f *textFacts) Does(s string)   { f.does = s }
func (f *textFacts) Undoes(s string) { f.undoes = s }
func (f *textFacts) NotUndoable()    { f.undoes = "" }
func (f *textFacts) Loses(s string)  { f.loses = s }
func (f *textFacts) NothingLost()    { f.loses = "" }
func (f *textFacts) Applied(t time.Time) {
	f.applied = true
	f.made = t
}
func (f *textFacts) NotApplied() { f.applied = false }

func textOf(m Migration) textFacts {
	var f textFacts
	m.Tell(&f)
	return f
}

func state(f textFacts) string {
	if f.applied {
		return "applied " + domain.FormatMoment(f.made)
	}
	return "waiting          "
}

// Text renders one change in full: what it is, what it does, and what undoing it does.
func Text(m Migration) string {
	f := textOf(m)
	var out strings.Builder
	fmt.Fprintf(&out, "#%d  %s  %s\n\n", f.number, state(f), f.name)
	fmt.Fprintf(&out, "what it does\n%s\n\n", domain.OrNothing(strings.TrimSpace(f.does)))
	fmt.Fprintf(&out, "what undoing it does\n%s\n\n", domain.OrNothing(strings.TrimSpace(f.undoes)))
	fmt.Fprintf(&out, "what undoing it loses\n%s\n", domain.OrNothing(f.loses))
	return out.String()
}

// List renders every change, applied or waiting: a list that only showed what was done could not tell you what to apply.
func List(all []Migration, held domain.Held) string {
	waiting := 0
	var out strings.Builder
	for _, m := range all {
		f := textOf(m)
		if !f.applied {
			waiting++
		}
		fmt.Fprintf(&out, "#%d  %s  %s\n", f.number, state(f), f.name)
	}
	fmt.Fprintf(&out, "\nthe store %s, %d change(s) waiting\n", domain.HeldText(held), waiting)
	return out.String()
}

// RolledBackText names the change a rollback unmade, and what that cost.
func RolledBackText(m Migration) string {
	f := textOf(m)
	var out strings.Builder
	fmt.Fprintf(&out, "rolled back %q\n", f.name)
	if f.loses != "" {
		fmt.Fprintf(&out, "lost: %s — `cairn backup restore --yes` puts it back\n", f.loses)
	}
	return out.String()
}
