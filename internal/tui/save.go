package tui

import (
	tea "charm.land/bubbletea/v2"

	"cairn/internal/engine"
)

// Saving is the only moment the form touches anything. What it hands over is
// the same Changes the command line hands over, so a field set here and a flag
// typed there are read by one piece of code and refused by one piece of code.

// save writes what the form holds. It gives the window back at once and keeps
// what was typed: a refusal puts the form back with everything still in it,
// which is what a form closed before the write was tried could not do — a date
// mistyped or a title left empty took the whole form with it.
//
// Handing the window back now and not when the answer arrives is not only
// about what is drawn. While the form has the window its commands are the
// form's, and one of them is a cursor blink — a timer nobody can wait out, so
// they are raced against a deadline. A write held under that rule is a write
// that loses the race on a loaded machine and is dropped: measured, the board
// answered from a reading the task had never been in.
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
	op, asked := b.saving(form, engine.TaskNew, engine.TaskEdit)
	return func() tea.Msg {
		if _, err := b.work.Ask(op, asked); err != nil {
			return Looked{Err: err}
		}
		if err := b.savingParts(form); err != nil {
			return Looked{Err: err}
		}
		found, err := engine.Look(b.work)
		return Looked{Snapshot: found, Err: err}
	}
}

// savingParts settles what hangs under the task with what the form's parts
// field says — see engine.SettleParts.
func (b Board) savingParts(form Form) error {
	said, holds := form.held["parts"]
	if !holds || form.Adding {
		return nil
	}
	return engine.SettleParts(b.work, form.Number, *said, b.snap.Cards)
}

// saving says which operation a saved form means and what it carries — see
// engine.Saved.
func (b Board) saving(form Form, adding, editing engine.Operation) (engine.Operation, engine.Input) {
	return engine.Saved(form.typed(), form.Adding, form.Number, adding, editing)
}

// saveGoal adds or edits an objective from the form.
func (b Board) saveGoal(form Form) tea.Cmd {
	return b.carry(b.saving(form, engine.ObjectiveNew, engine.ObjectiveEdit))
}
