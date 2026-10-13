// Package tui draws cairn on a screen that stays open. It is a surface like
// the command line and the MCP server: it reads a key, calls one thing in
// internal/engine, and draws what comes back — it decides nothing.
//
// Nothing here touches a terminal. A board is a value, a key press turns one
// board into another, and drawing reads a board without changing it, which is
// what lets a scenario press a key and read the screen with no tty anywhere.
package tui

import (
	"strings"

	tea "charm.land/bubbletea/v2"

	"cairn/internal/engine"
)

// What a key changes, and how. Every one of these goes back through the same
// operation the command line calls: a board that wrote its own way to the store
// would be a second place where a rule about a task is kept, and the two would
// drift the first time one of them was fixed.

// setState asks the engine to move a task to a column.
func (b Board) setState(state string) tea.Cmd {
	in := engine.Numbered(b.task)
	in.Given["state"] = state
	return b.carry(engine.TaskEdit, in)
}

// moving asks the engine to change a task's place. It never changes the
// column: that is an edit, and one act with two meanings is an act nobody can
// undo confidently — the engine says so, and this respects it.
func (b Board) moving(where map[string]string) tea.Cmd {
	in := engine.Numbered(b.task)
	for name, value := range where {
		in.Given[name] = value
	}
	return b.carry(engine.TaskMove, in)
}

// choosingKind is what a key does while the popup is asking which of the three
// kinds of stopped a task is going into.
func (b Board) choosingKind(key string) (tea.Model, tea.Cmd) {
	// Ifs rather than a tagless switch: a switch with no expression puts its
	// conditions where the mutation gate cannot reach them.
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

// toggleFocus takes the task in hand, or lets it go when it is already the one
// in hand. One key for both, because it is one question — see engine.FocusOp.
func (b Board) toggleFocus() tea.Cmd {
	card, holding := b.snap.Card(b.task)
	if !holding {
		return nil
	}
	return b.carry(engine.FocusOp(card), engine.Numbered(card.Number))
}

// remove deletes the task the cursor is holding.
func (b Board) remove() tea.Cmd {
	return b.carry(engine.TaskDelete, engine.Agreed(engine.Numbered(b.task)))
}

// puttingAway takes the task the cursor is on off the board, or brings it back
// when it is already off. One key both ways: the board is showing which of the
// two it is, so a second key would only be a second thing to remember.
func (b Board) puttingAway() tea.Cmd {
	card, holding := b.held()
	if !holding {
		return nil
	}
	return b.carry(engine.AwayOp(card), engine.Numbered(card.Number))
}

// sweeping takes every task in the column the cursor is on off the board. The
// engine has swept by a narrowing all along and only the command line ever
// asked it to: archiving twelve finished tasks was one gesture there and
// twenty-four here.
func (b Board) sweepingColumn() tea.Cmd {
	return b.carry(engine.TaskArchive, engine.Agreed(engine.Input{Given: map[string]string{"state": engine.StateAt(b.column)}}))
}

// removeGoal deletes the objective the cursor is on.
func (b Board) removeGoal() tea.Cmd {
	if b.goal == 0 {
		return nil
	}
	return b.carry(engine.ObjectiveDelete, engine.Agreed(engine.Numbered(b.goal)))
}

// otherTab moves between the two tabs, and lets go of whatever the cursor was
// holding on the one being left — a task number means nothing on a list of
// objectives.
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

// firstGoal is the objective the cursor lands on when the tab is opened.
func (b Board) firstGoal() int {
	goals := b.goals()
	if len(goals) == 0 {
		return 0
	}
	return goals[0].Number
}

// goals is every objective the board is drawing, in the order it draws them.
func (b Board) goals() []engine.Goal {
	return engine.FoundGoals(b.snap.Goals, b.query)
}

// openObjective opens the board of the objective the cursor is on: the same
// board, holding only what serves it.
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

// leaveObjective closes an objective's own board and goes back to the list it
// was opened from, with the cursor still on it.
func (b Board) leaveObjective() Board {
	b.goal = b.inside
	b.inside = 0
	b.tab = Objectives
	b.mode = Normal
	b.task = 0
	b.inBand = false
	return b.follow()
}

// accept is what a prompt does when it is told to go ahead.
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

// run carries out what was typed at the command prompt. The words are the
// command line's words, so what is understood here is what is understood
// there — and what is not understood is said rather than swallowed.
//
// The vocabulary is engine.All() and nothing else: the prompt knew two words
// out of thirty, which made a bar that offers ":" an offer of almost nothing.
// An operation added to the engine is understood here the day it is added.
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

// carry runs an operation the prompt named and reads the board back, which is
// what every other gesture does once it has written.
func (b Board) carry(op engine.Operation, in engine.Input) tea.Cmd {
	return func() tea.Msg {
		if _, err := b.work.Ask(op, in); err != nil {
			return Looked{Err: err}
		}
		found, err := engine.Look(b.work)
		return Looked{Snapshot: found, Err: err}
	}
}

// telling runs an operation somebody asked for by name — at the prompt, or
// by a key that stands for one — and says what it answered, the way the
// command line would. A gesture on the board says nothing: what it did is
// on the board.
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
