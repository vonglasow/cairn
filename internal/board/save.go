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
	op, asked := b.saving(form, engine.TaskAdd, engine.TaskEdit)
	return func() tea.Msg {
		if _, err := b.work.Ask(op, asked); err != nil {
			return Looked{Err: err}
		}
		found, err := Look(b.work)
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
	asked := askedFrom(form, op)
	asked.Given["title"] = strings.TrimSpace(form.valueOf("title"))
	if !form.Adding {
		asked.Given["number"] = strconv.Itoa(form.Number)
	}
	return op, asked
}

// askedFrom turns the form's fields into what the operation takes. Every field
// is named, including the empty ones: a form is the whole of a task, so a field
// left blank is a field somebody emptied, not a field they said nothing about.
//
// Nothing here lists the fields or pairs them up. Each field says which
// argument it fills, and the operation says what kind that argument is — so a
// field added to the form, or an argument added to the engine, arrives without
// a line written here. That is the whole reason Field carries Arg.
func askedFrom(form Form, op engine.Operation) engine.Input {
	kinds := map[string]engine.Kind{}
	for _, argument := range op.Takes {
		kinds[argument.Name] = argument.Kind
	}
	in := engine.Input{Given: map[string]string{}, Listed: map[string][]string{}}
	for _, field := range form.fields {
		switch kinds[field.Arg] {
		case engine.List:
			in.Listed[field.Arg] = words(field.Value)
		case engine.Number:
			// A number nobody can read is not a change anybody asked for. Empty
			// is not that: it reads as zero, which is how such a field is
			// cleared, and the engine reads it back the same way.
			if _, ok := number(field.Value); ok {
				in.Given[field.Arg] = strings.TrimSpace(field.Value)
			}
		default:
			in.Given[field.Arg] = field.Value
		}
	}
	return in
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
	op, asked := b.saving(form, engine.ObjectiveAdd, engine.ObjectiveEdit)
	return func() tea.Msg {
		if _, err := b.work.Ask(op, asked); err != nil {
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
