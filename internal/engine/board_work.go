package engine

import (
	"slices"
	"strconv"
	"time"

	"cairn/internal/domain"
	"cairn/internal/objective"
	"cairn/internal/task"
)

// Work is what the board needs from whatever holds the work. Declared here, by
// the consumer, so a test answers for it without a database and the board never
// learns what is behind it.
//
// Everything that changes anything goes through the same operations the command
// line calls. A key pressed here and a command typed there reach the engine by
// the same door, or one of them would be a second place where a rule is kept.
type Work interface {
	// Ask carries out one of the operations the engine describes. The board
	// names no operation of its own: an operation added to the engine reaches
	// this surface the way it reaches the command line and the agent, without
	// a line written here.
	Ask(Operation, Input) (Answer, error)

	Version() (int64, error)
}

// Card is one task, flattened into what a board draws. It is read off the
// task's own membrane rather than out of its fields, which is the only way in:
// a task tells its facts, it does not hand them over.
type Card struct {
	Number int
	Parent int
	Title  string
	State  string
	// Archived says the task is off the board, drawn only when it is asked for.
	Archived bool
	Priority string
	Labels   []string
	Blockers []int
	Because  string
	Deadline time.Time
	Estimate string
	Spent    string
	// Overrun says the time spent has gone past the estimate — read off the
	// lengths, so no board reads it back off how a card writes them.
	Overrun     bool
	Objectives  []int
	Description string
	SubDone     int
	SubTotal    int
	InFocus     bool
	// Freed says this blocked task is waiting on nothing any more: every task
	// it named is done. It is the one thing about blocking that has to be acted
	// on, and it used to reach the command line's listing and nowhere else.
	Freed bool
	// Context is where what was written about the task elsewhere lives, as
	// showing it answered; empty when the board was not told where the notes
	// are. See docs/context.md.
	Context string
	// Said is what a search looks in, as the engine says it.
	Said []string
}

// Described says whether anything is written about the task.
func (c Card) Described() bool { return c.Description != "" }

// Goal is one objective, flattened the same way.
type Goal struct {
	Number      int
	Title       string
	State       string
	Priority    string
	Description string
	Labels      []string
	Deadline    time.Time
	Estimate    string
	Done        int
	Total       int
	// Said is what a search looks in, as the engine says it.
	Said []string
}

// Described says whether anything is written about the objective.
func (g Goal) Described() bool { return g.Description != "" }

// Snapshot is everything the board draws, read in one go. The board holds one
// and redraws from it; nothing is fetched while a key is being handled, so what
// is on screen is always one consistent reading rather than several.
type Snapshot struct {
	Cards []Card
	Goals []Goal
	// Focus is every task being worked on, in the order they were taken up.
	Focus   []int
	Version int64
}

// Look reads the whole of it. A failure to read anything at all is a failure to
// draw anything at all: half a board would be a board that lies about what is
// there.
func Look(w Work) (Snapshot, error) {
	// Everything, archived included: which of them a board draws is the
	// board's to decide.
	listed, err := w.Ask(TaskList, Input{
		Given: map[string]string{"with-archived": "true"},
	})
	if err != nil {
		return Snapshot{}, err
	}
	goals, err := w.Ask(ObjectiveList, Input{})
	if err != nil {
		return Snapshot{}, err
	}
	version, err := w.Version()
	if err != nil {
		return Snapshot{}, err
	}
	// The focus comes with the listing that marks it rather than through a
	// question of its own: every listing carries it, and asking twice could
	// answer twice differently.
	var held []int
	if listed.InFocus != nil {
		held = *listed.InFocus
	}
	found := Snapshot{Focus: held, Version: version}
	// How far along each one is and where its context lives come with the
	// listings, so drawing two hundred cards is two questions and not two
	// hundred.
	for _, t := range listed.Tasks {
		card, err := cardOf(t, held, listed.NothingBlocks, listed.Along, listed.Contexts)
		if err != nil {
			return Snapshot{}, err
		}
		found.Cards = append(found.Cards, card)
	}
	for _, o := range goals.Objectives {
		goal, err := goalOf(o, goals.Along)
		if err != nil {
			return Snapshot{}, err
		}
		found.Goals = append(found.Goals, goal)
	}
	return found, nil
}

// Numbered is what every operation about one task or one objective is given.
// Its own function because it is written five times otherwise, and a map built
// by hand five times is a map spelt wrong once.
func Numbered(n int) Input {
	return Input{Given: map[string]string{"number": strconv.Itoa(n)}}
}

// shownAll is every label as it is safe to draw. What was written before
// cairn refused escapes, or comes back with an older backup, is shown and
// never obeyed — the terminal board draws these strings straight to a screen.
func shownAll(kept []string) []string {
	shown := make([]string, 0, len(kept))
	for _, k := range kept {
		shown = append(shown, domain.Shown(k))
	}
	return shown
}

// cardOf turns one task into what a card draws. What it says is shown, never
// obeyed: see shownAll.
func cardOf(t task.Task, held, freed []int, along map[int]domain.Progress, contexts map[int]string) (Card, error) {
	told := task.OfTask(t)
	// The deadline crosses the membrane written the way a person reads it, and
	// the board needs the day itself to say "in two days". Read back with the
	// same parser that wrote it, rather than a second layout kept here — a
	// second layout is a second thing to keep true.
	due, err := domain.ParseDeadline(told.Deadline)
	if err != nil {
		return Card{}, err
	}
	sofar := along[told.Number]
	return Card{
		Number:      told.Number,
		Parent:      told.Parent,
		Archived:    told.Archived,
		Title:       domain.Shown(told.Title),
		State:       told.State,
		Priority:    told.Priority,
		Description: domain.Shown(told.Description),
		Labels:      shownAll(told.Labels),
		Blockers:    told.BlockedBy,
		Because:     domain.Shown(told.Because),
		Deadline:    due,
		Estimate:    told.Estimate,
		Spent:       told.TimeSpent,
		Overrun:     overrun(told.Estimate, told.TimeSpent),
		Objectives:  told.Objectives,
		SubDone:     sofar.Done(),
		SubTotal:    sofar.Total(),
		InFocus:     slices.Contains(held, told.Number),
		Freed:       slices.Contains(freed, told.Number),
		Context:     contexts[told.Number],
		Said:        task.Said(t),
	}, nil
}

// goalOf turns one objective into what a card draws.
func goalOf(o objective.Objective, along map[int]domain.Progress) (Goal, error) {
	told := objective.OfObjective(o)
	due, err := domain.ParseDeadline(told.Deadline)
	if err != nil {
		return Goal{}, err
	}
	sofar := along[told.Number]
	return Goal{
		Number:      told.Number,
		Title:       domain.Shown(told.Title),
		State:       told.State,
		Priority:    told.Priority,
		Description: domain.Shown(told.Description),
		Labels:      shownAll(told.Labels),
		Deadline:    due,
		Estimate:    told.Estimate,
		Done:        sofar.Done(),
		Total:       sofar.Total(),
		Said:        objective.Said(o),
	}, nil
}
