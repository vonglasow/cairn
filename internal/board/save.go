package board

import (
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"

	"cairn/internal/engine"
)

// Saving is the only moment the form touches anything. What it hands over is
// the same Changes the command line hands over, so a field set here and a flag
// typed there are read by one piece of code and refused by one piece of code.

// save writes what the form holds, and closes it.
func (b Board) save() (tea.Model, tea.Cmd) {
	form := b.form
	b.screen = OnTheBoard
	if form.ForObjective {
		return b, b.saveGoal(form)
	}
	return b, b.saveTask(form)
}

// saveTask adds or edits a task from the form.
func (b Board) saveTask(form Form) tea.Cmd {
	title := strings.TrimSpace(form.valueOf("title"))
	changes := changesFrom(form)
	return func() tea.Msg {
		var err error
		if form.Adding {
			_, err = b.work.Add(title, changes)
		} else {
			changes.Title = &title
			_, err = b.work.Edit(form.Number, changes)
		}
		if err != nil {
			return Looked{Err: err}
		}
		found, err := Look(b.work)
		return Looked{Snapshot: found, Err: err}
	}
}

// changesFrom turns the form's fields into what the engine takes. Every field
// is named, including the empty ones: a form is the whole of a task, so a field
// left blank is a field somebody emptied, not a field they said nothing about.
//
// A task being added is the exception on two of them. Its column is what was
// given, and its labels cannot carry a + or a -, so what is not there is not
// named rather than named empty.
func changesFrom(form Form) engine.Changes {
	state := form.valueOf("column")
	priority := form.valueOf("priority")
	description := form.valueOf("description")
	because := form.valueOf("because")
	deadline := form.valueOf("deadline")
	estimate := form.valueOf("estimate")
	spent := form.valueOf("time spent")
	labels := words(form.valueOf("labels"))
	blockers := words(form.valueOf("blockers"))
	objectives := words(form.valueOf("objectives"))
	changes := engine.Changes{
		State:       &state,
		Priority:    &priority,
		Description: &description,
		Because:     &because,
		Deadline:    &deadline,
		Estimate:    &estimate,
		Spent:       &spent,
		Labels:      &labels,
		Blocker:     &blockers,
		Objective:   &objectives,
	}
	if parent, ok := number(form.valueOf("parent")); ok {
		changes.Parent = &parent
	}
	return changes
}

// number reads a field back as a task number, and says when there was nothing
// to read. Nothing is not zero: a parent of zero is a task hanging under
// nothing, which is what emptying the field means, so both answer the same
// value and the caller cannot tell them apart — nor does it need to.
func number(s string) (int, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, true
	}
	parsed, err := strconv.Atoi(s)
	if err != nil {
		return 0, false
	}
	return parsed, true
}

// saveGoal adds or edits an objective from the form.
func (b Board) saveGoal(form Form) tea.Cmd {
	title := strings.TrimSpace(form.valueOf("title"))
	state := form.valueOf("standing")
	priority := form.valueOf("priority")
	description := form.valueOf("description")
	deadline := form.valueOf("deadline")
	estimate := form.valueOf("estimate")
	labels := words(form.valueOf("labels"))
	changes := engine.ObjectiveChanges{
		State:       &state,
		Priority:    &priority,
		Description: &description,
		Deadline:    &deadline,
		Estimate:    &estimate,
		Labels:      &labels,
	}
	return func() tea.Msg {
		var err error
		if form.Adding {
			_, err = b.work.AddObjective(title, changes)
		} else {
			changes.Title = &title
			_, err = b.work.EditObjective(form.Number, changes)
		}
		if err != nil {
			return Looked{Err: err}
		}
		found, err := Look(b.work)
		return Looked{Snapshot: found, Err: err}
	}
}

// words splits what was typed into a field the way a shell would have split
// it: on spaces, and nothing cleverer. A form is not a place to invent a
// syntax.
func words(s string) []string {
	found := strings.Fields(s)
	if found == nil {
		return []string{}
	}
	return found
}
