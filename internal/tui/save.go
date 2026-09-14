package tui

import (
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"

	"cairn/internal/engine"
	"cairn/internal/reading"
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
	op, asked := b.saving(form, engine.TaskAdd, engine.TaskEdit)
	return func() tea.Msg {
		if _, err := b.work.Ask(op, asked); err != nil {
			return Looked{Err: err}
		}
		found, err := reading.Look(b.work)
		return Looked{Snapshot: found, Err: err}
	}
}

// saving says which operation a saved form means and what it carries. Adding
// and editing are one form, so they are one function here too: the difference
// is the operation and the number, and nothing else about a form knows which
// of the two it is.
//
// The title is trimmed on its way out and nowhere else: a title is what a task
// is called, and what somebody called it does not include the space they left
// after it.
func (b Board) saving(form Form, adding, editing engine.Operation) (engine.Operation, engine.Input) {
	op := editing
	if form.Adding {
		op = adding
	}
	asked := reading.Asked(form.fields, op)
	asked.Given["title"] = strings.TrimSpace(form.valueOf("title"))
	if !form.Adding {
		asked.Given["number"] = strconv.Itoa(form.Number)
	}
	return op, asked
}

// saveGoal adds or edits an objective from the form.
func (b Board) saveGoal(form Form) tea.Cmd {
	op, asked := b.saving(form, engine.ObjectiveAdd, engine.ObjectiveEdit)
	return func() tea.Msg {
		if _, err := b.work.Ask(op, asked); err != nil {
			return Looked{Err: err}
		}
		found, err := reading.Look(b.work)
		return Looked{Snapshot: found, Err: err}
	}
}
