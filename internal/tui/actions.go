// Package tui draws cairn on a screen that stays open.
package tui

import (
	"slices"
	"strconv"
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

var pageOfTab = map[Tab]string{Tasks: engine.PageTasks, Objectives: engine.PageObjectives, Series: engine.PageTasks, Time: engine.PageTime}

// Paged answers the board showing only the pages the configuration leaves on, on the first of them when the one it was on is off.
func (b Board) Paged(p engine.Pages) Board {
	b.pages = p
	b.snap = b.pages.Recurring(b.snap)
	if shown := b.shown(); len(shown) > 0 && !slices.Contains(shown, b.tab) {
		return b.onTab(shown[0])
	}
	return b
}

func (b Board) shown() []Tab {
	var shown []Tab
	for _, tab := range []Tab{Tasks, Objectives, Time, Series} {
		if tab == Series && !b.pages.Recurs() {
			continue
		}
		if b.pages.Shows(pageOfTab[tab]) {
			shown = append(shown, tab)
		}
	}
	return shown
}

func (b Board) otherTab() Board {
	shown := b.shown()
	if len(shown) == 0 {
		return b
	}
	return b.onTab(shown[(max(0, slices.Index(shown, b.tab))+1)%len(shown)])
}

func (b Board) onTab(tab Tab) Board {
	b.mode = Normal
	switch tab {
	case Objectives:
		b.tab = Objectives
		b.task = 0
		b.goal = b.firstGoal()
		return b.follow()
	case Time:
		b.tab, b.inside = Time, 0
		if b.week.IsZero() {
			b.week = engine.Monday(b.clock())
		}
		return b
	case Series:
		return b.onSeries(b.seriesOn)
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
	branches := engine.Branches(engine.FoundGoals(b.snap.Goals, b.query))
	out := make([]engine.Goal, 0, len(branches))
	for _, br := range branches {
		if br.Under && slices.Contains(b.folded, br.Goal.Parent) {
			continue
		}
		out = append(out, br.Goal)
	}
	return out
}

func (b Board) folding() Board {
	tree := b.snap.TreeOf(b.goal)
	if g, found := b.snap.Goal(tree); !found || len(g.Children) == 0 {
		return b
	}
	if at := slices.Index(b.folded, tree); at >= 0 {
		b.folded = slices.Delete(slices.Clone(b.folded), at, at+1)
	} else {
		b.folded = append(slices.Clone(b.folded), tree)
	}
	b.goal = tree
	return b.follow()
}

func (b Board) segments() []engine.Segment {
	if b.inside != 0 {
		return b.snap.Segments(b.inside, b.unopened())
	}
	if value, ok := b.objectiveShown(); ok && b.tab == Tasks {
		root, _ := strconv.Atoi(strings.TrimSuffix(value, " itself"))
		return b.snap.Segments(b.snap.TreeOf(root), engine.OnTheBoard(b.snap.Cards, b.archived))
	}
	return nil
}

func (b Board) objectiveShown() (string, bool) {
	for _, f := range b.filters {
		if taken := f.Taken(); f.Arg == "objective" && len(taken) == 1 {
			return taken[0], true
		}
	}
	return "", false
}

func (b Board) segmentOn() int {
	if b.inside != 0 {
		return b.segment
	}
	value, _ := b.objectiveShown()
	for i, seg := range b.segments() {
		if seg.Value == value {
			return i
		}
	}
	return 0
}

func (b Board) walkingSegments(by int) Board {
	segments := b.segments()
	if len(segments) == 0 {
		return b
	}
	b.segment = ((b.segmentOn()+by)%len(segments) + len(segments)) % len(segments)
	if b.inside == 0 {
		b.filters = b.filters.On("objective", segments[b.segment].Value)
		b.segment = 0
	}
	b.mode, b.task, b.inBand = Normal, 0, false
	b.offsets = [engine.Columns]int{}
	return b
}

func (b Board) showingItsTasks() Board {
	if b.goal == 0 {
		return b
	}
	b.filters = engine.BoardFilters(b.snap, engine.DayOf(b.clock())).Keeping(b.filters).On("objective", strconv.Itoa(b.goal))
	b = b.onTab(Tasks)
	b.column, b.inBand, b.inside = 0, false, 0
	return b
}

func (b Board) openObjective() Board {
	if b.goal == 0 {
		return b
	}
	b.inside = b.goal
	b.segment = 0
	b.tab = Tasks
	b.mode = Normal
	b.column = 0
	b.inBand = false
	return b
}

func (b Board) leaveObjective() Board {
	b.goal = b.inside
	b.inside, b.segment = 0, 0
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
	if model, cmd, took := b.viewCommand(words); took {
		return model, cmd
	}
	op, found := b.pages.Called(words[0], words[1])
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
		answer, err := b.work.Ask(op, in)
		if err != nil {
			if agreeing, asks := engine.AgreeingTo(op, in, err); asks {
				return agreeing
			}
			return Looked{Err: err}
		}
		found, err := engine.Look(b.work)
		return Looked{Snapshot: found, Err: err, Warned: answer.Warnings}
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
