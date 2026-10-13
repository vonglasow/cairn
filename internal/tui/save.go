package tui

import (
	tea "charm.land/bubbletea/v2"

	"cairn/internal/engine"
)

func (b Board) save() (tea.Model, tea.Cmd) {
	form := b.form
	b.screen = OnTheBoard
	if form.ForObjective {
		return b, b.saveGoal(form)
	}
	return b, b.saveTask(form)
}

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

func (b Board) savingParts(form Form) error {
	said, holds := form.held["parts"]
	if !holds || form.Adding {
		return nil
	}
	for _, gone := range engine.PartsGone(*said, b.snap.Cards, form.Number) {
		if _, err := b.work.Ask(engine.TaskDelete, engine.Agreed(engine.Numbered(gone))); err != nil {
			return err
		}
	}
	return nil
}

func (b Board) saving(form Form, adding, editing engine.Operation) (engine.Operation, engine.Input) {
	return engine.Saved(form.typed(), form.Adding, form.Number, adding, editing)
}

func (b Board) saveGoal(form Form) tea.Cmd {
	return b.carry(b.saving(form, engine.ObjectiveNew, engine.ObjectiveEdit))
}
