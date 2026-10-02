package tui

import (
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"

	"cairn/internal/domain"
	"cairn/internal/engine"
	"cairn/internal/reading"
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
	op, asked := b.saving(form, engine.TaskAdd, engine.TaskEdit)
	return func() tea.Msg {
		if _, err := b.work.Ask(op, asked); err != nil {
			return Looked{Err: err}
		}
		if err := b.savingParts(form); err != nil {
			return Looked{Err: err}
		}
		found, err := reading.Look(b.work)
		return Looked{Snapshot: found, Err: err}
	}
}

// savingParts settles what hangs under the task with what the form's own field
// says: a line carrying a number is that part, named and ticked as it reads; a
// line carrying none is a part nobody has written down yet; a number the field
// no longer carries is a part that has gone.
//
// One operation each, through the engine like everything else. The engine has
// no argument for a set of parts and should not grow one: what it is asked for
// here is adding a task, changing a task and deleting a task.
func (b Board) savingParts(form Form) error {
	said, holds := form.held["parts"]
	if !holds || form.Adding {
		return nil
	}
	parent := strconv.Itoa(form.Number)
	kept := map[int]bool{}
	for _, part := range reading.PartsIn(*said) {
		state := string(domain.Todo)
		if part.Done {
			state = string(domain.Done)
		}
		if part.Number == 0 {
			if _, err := b.work.Ask(engine.TaskAdd, engine.Input{Given: map[string]string{
				"title": part.Title, "parent": parent, "state": state,
			}}); err != nil {
				return err
			}
			continue
		}
		kept[part.Number] = true
		if _, err := b.work.Ask(engine.TaskEdit, engine.Input{Given: map[string]string{
			"number": strconv.Itoa(part.Number), "title": part.Title, "state": state,
		}}); err != nil {
			return err
		}
	}
	for _, was := range reading.Hanging(b.snap.Cards, form.Number) {
		if kept[was.Number] {
			continue
		}
		if _, err := b.work.Ask(engine.TaskDelete, reading.Numbered(was.Number)); err != nil {
			return err
		}
	}
	return nil
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
	asked := reading.Asked(form.typed(), op)
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
