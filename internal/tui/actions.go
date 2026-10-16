// Package tui draws cairn on a screen that stays open.
package tui

import (
	"errors"
	"strings"

	tea "charm.land/bubbletea/v2"

	"cairn/internal/engine"
)

func (b Board) setState(state string) tea.Cmd {
	in := engine.Numbered(b.task)
	in.Given["state"] = state
	return b.carry(engine.TaskEdit, in)
}

func (b Board) moving(where map[string]string) tea.Cmd {
	in := engine.Numbered(b.task)
	for name, value := range where {
		in.Given[name] = value
	}
	return b.carry(engine.TaskMove, in)
}

func (b Board) choosingKind(key string) (tea.Model, tea.Cmd) {
	if way, ok := direction(key); ok {
		if way == goUp && b.choice > 0 {
			b.choice--
		}
		if way == goDown && b.choice < len(engine.Stalled)-1 {
			b.choice++
		}
		return b, nil
	}
	if key == "esc" {
		b.choosing = false
		return b, nil
	}
	if key == "enter" {
		b.choosing = false
		b.cameFrom = b.column
		b.inBand = true
		return b, b.setState(engine.Stalled[b.choice])
	}
	return b, nil
}

func (b Board) toggleFocus() tea.Cmd {
	card, holding := b.snap.Card(b.task)
	if !holding {
		return nil
	}
	return b.carry(engine.FocusOp(card), engine.Numbered(card.Number))
}

func (b Board) remove() tea.Cmd {
	return b.carry(engine.TaskDelete, engine.Agreed(engine.Numbered(b.task)))
}

func (b Board) puttingAway() tea.Cmd {
	card, holding := b.held()
	if !holding {
		return nil
	}
	return b.carry(engine.AwayOp(card), engine.Numbered(card.Number))
}

func (b Board) sweepingColumn() tea.Cmd {
	return b.carry(engine.TaskArchive, engine.Agreed(engine.Input{Given: map[string]string{"state": engine.StateAt(b.column)}}))
}

func (b Board) removeGoal() tea.Cmd {
	if b.goal == 0 {
		return nil
	}
	return b.carry(engine.ObjectiveDelete, engine.Agreed(engine.Numbered(b.goal)))
}

func (b Board) otherTab() Board {
	b.mode = Normal
	switch b.tab {
	case Tasks:
		b.tab = Objectives
		b.task = 0
		b.goal = b.firstGoal()
		return b.follow()
	case Objectives:
		b.tab, b.inside = Time, 0
		if b.week.IsZero() {
			b.week = engine.Monday(b.clock())
		}
		return b
	}
	b.tab = Tasks
	return b
}

func (b Board) firstGoal() int {
	goals := b.goals()
	if len(goals) == 0 {
		return 0
	}
	return goals[0].Number
}

func (b Board) goals() []engine.Goal {
	return engine.FoundGoals(b.snap.Goals, b.query)
}

func (b Board) openObjective() Board {
	if b.goal == 0 {
		return b
	}
	b.inside = b.goal
	b.tab = Tasks
	b.mode = Normal
	b.column = 0
	b.inBand = false
	return b
}

func (b Board) leaveObjective() Board {
	b.goal = b.inside
	b.inside = 0
	b.tab = Objectives
	b.mode = Normal
	b.task = 0
	b.inBand = false
	return b.follow()
}

func (b Board) accept() (tea.Model, tea.Cmd) {
	typed := b.typed
	was := b.prompt
	b.prompt, b.typed = NoPrompt, ""
	if was == Searching {
		b.query = typed
		return b, nil
	}
	return b.run(typed)
}

func (b Board) run(line string) (tea.Model, tea.Cmd) {
	words := strings.Fields(line)
	if len(words) == 1 && (words[0] == "q" || words[0] == "quit") {
		return b, tea.Quit
	}
	if len(words) < 2 {
		b.err = unknown(line)
		return b, nil
	}
	op, found := engine.Called(words[0], words[1])
	if !found {
		b.err = unknown(line)
		return b, nil
	}
	in, err := op.Read(words[2:])
	if err != nil {
		b.err = err
		return b, nil
	}
	return b, b.telling(op, in)
}

func (b Board) carry(op engine.Operation, in engine.Input) tea.Cmd {
	return func() tea.Msg {
		if _, err := b.work.Ask(op, in); err != nil {
			if op.Finishes && !in.Confirmed && errors.Is(err, engine.ErrSubTaskStillOpen) {
				return Finishing{Op: op, In: in}
			}
			return Looked{Err: err}
		}
		found, err := engine.Look(b.work)
		return Looked{Snapshot: found, Err: err}
	}
}

func (b Board) telling(op engine.Operation, in engine.Input) tea.Cmd {
	return func() tea.Msg {
		answer, err := b.work.Ask(op, in)
		if err != nil {
			return Looked{Err: err}
		}
		found, err := engine.Look(b.work)
		return Answered{Text: engine.RenderText(op, answer), Looked: Looked{Snapshot: found, Err: err}}
	}
}
