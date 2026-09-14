// Package board draws cairn on a screen that stays open. It is a surface like
// the command line and the MCP server: it reads a key, calls one thing in
// internal/engine, and draws what comes back — it decides nothing.
//
// Nothing here touches a terminal. A board is a value, a key press turns one
// board into another, and drawing reads a board without changing it, which is
// what lets a scenario press a key and read the screen with no tty anywhere.
package board

import (
	"slices"
	"strconv"

	tea "charm.land/bubbletea/v2"

	"cairn/internal/engine"
	"cairn/internal/reading"
)

// What a key changes, and how. Every one of these goes back through the same
// operation the command line calls: a board that wrote its own way to the store
// would be a second place where a rule about a task is kept, and the two would
// drift the first time one of them was fixed.

// kinds are the three the band holds, in the order the popup offers them.
var kinds = []string{"blocked", "watching", "paused"}

// setState asks the engine to move a task to a column.
func (b Board) setState(state string) tea.Cmd {
	number := b.task
	return func() tea.Msg {
		asked := engine.Input{Given: map[string]string{
			"number": strconv.Itoa(number),
			"state":  state,
		}}
		if _, err := b.work.Ask(engine.TaskEdit, asked); err != nil {
			return Looked{Err: err}
		}
		found, err := reading.Look(b.work)
		return Looked{Snapshot: found, Err: err}
	}
}

// moveBehind puts the task straight after another among the ones it sits with.
func (b Board) moveBehind(target int) tea.Cmd {
	return b.moving(map[string]string{"after": strconv.Itoa(target)})
}

// moveFirst puts it at the head of them.
func (b Board) moveFirst() tea.Cmd {
	return b.moving(map[string]string{"first": "true"})
}

// moving asks the engine to change a task's place. It never changes the
// column: that is an edit, and one act with two meanings is an act nobody can
// undo confidently — the engine says so, and this respects it.
func (b Board) moving(where map[string]string) tea.Cmd {
	number := b.task
	return func() tea.Msg {
		given := map[string]string{"number": strconv.Itoa(number)}
		for name, value := range where {
			given[name] = value
		}
		if _, err := b.work.Ask(engine.TaskMove, engine.Input{Given: given}); err != nil {
			return Looked{Err: err}
		}
		found, err := reading.Look(b.work)
		return Looked{Snapshot: found, Err: err}
	}
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
		if way == goDown && b.choice < len(kinds)-1 {
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
		return b, b.setState(kinds[b.choice])
	}
	return b, nil
}

// toggleFocus takes the task in hand, or lets it go when it is already the one
// in hand. One key for both, because it is one question.
func (b Board) toggleFocus() tea.Cmd {
	number := b.task
	held := b.snap.Focus
	return func() tea.Msg {
		// Both name the task. Several can be in focus at once, so letting go
		// of nothing in particular would let go of every one of them — which
		// is what bare `task unfocus` is for, and not what pressing f on one
		// card means.
		op := engine.TaskFocus
		if slices.Contains(held, number) {
			op = engine.TaskUnfocus
		}
		given := map[string]string{"number": strconv.Itoa(number)}
		if _, err := b.work.Ask(op, engine.Input{Given: given}); err != nil {
			return Looked{Err: err}
		}
		found, err := reading.Look(b.work)
		return Looked{Snapshot: found, Err: err}
	}
}

// remove deletes the task the cursor is holding.
func (b Board) remove() tea.Cmd {
	number := b.task
	return func() tea.Msg {
		asked := engine.Input{Given: map[string]string{"number": strconv.Itoa(number)}}
		if _, err := b.work.Ask(engine.TaskDelete, asked); err != nil {
			return Looked{Err: err}
		}
		found, err := reading.Look(b.work)
		return Looked{Snapshot: found, Err: err}
	}
}

// removeGoal deletes the objective the cursor is on.
func (b Board) removeGoal() tea.Cmd {
	number := b.goal
	if number == 0 {
		return nil
	}
	return func() tea.Msg {
		asked := engine.Input{Given: map[string]string{"number": strconv.Itoa(number)}}
		if _, err := b.work.Ask(engine.ObjectiveDelete, asked); err != nil {
			return Looked{Err: err}
		}
		found, err := reading.Look(b.work)
		return Looked{Snapshot: found, Err: err}
	}
}

// otherTab moves between the two tabs, and lets go of whatever the cursor was
// holding on the one being left — a task number means nothing on a list of
// objectives.
func (b Board) otherTab() Board {
	if b.tab == Tasks {
		b.tab = Objectives
		b.mode = Normal
		b.task = 0
		b.goal = b.firstGoal()
		return b.follow()
	}
	b.tab = Tasks
	b.mode = Normal
	b.inside = 0
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
func (b Board) goals() []reading.Goal {
	var kept []reading.Goal
	for _, g := range b.snap.Goals {
		if b.query != "" && !reading.Reaches(g, b.query) {
			continue
		}
		kept = append(kept, g)
	}
	return reading.Sorted(kept)
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
func (b Board) run(line string) (tea.Model, tea.Cmd) {
	switch line {
	case "q", "quit":
		return b, tea.Quit
	}
	b.err = unknown(line)
	return b, nil
}
