package board

import (
	"strconv"
	"time"

	"cairn/internal/domain"
	"cairn/internal/engine"
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
	Ask(engine.Operation, engine.Input) (engine.Answer, error)

	Version() (int64, error)
}

// Card is one task, flattened into what a board draws. It is read off the
// task's own membrane rather than out of its fields, which is the only way in:
// a task tells its facts, it does not hand them over.
type Card struct {
	Number      int
	Parent      int
	Title       string
	State       string
	Priority    string
	Described   bool
	Labels      []string
	Blockers    []int
	Because     string
	Deadline    time.Time
	Estimate    string
	Spent       string
	Objectives  []int
	Description string
	SubDone     int
	SubTotal    int
	InFocus     bool
}

// Goal is one objective, flattened the same way.
type Goal struct {
	Number      int
	Title       string
	State       string
	Priority    string
	Described   bool
	Description string
	Labels      []string
	Deadline    time.Time
	Estimate    string
	Done        int
	Total       int
}

// Snapshot is everything the board draws, read in one go. The board holds one
// and redraws from it; nothing is fetched while a key is being handled, so what
// is on screen is always one consistent reading rather than several.
type Snapshot struct {
	Cards   []Card
	Goals   []Goal
	Focus   int
	Version int64
}

// Look reads the whole of it. A failure to read anything at all is a failure to
// draw anything at all: half a board would be a board that lies about what is
// there.
func Look(w Work) (Snapshot, error) {
	listed, err := w.Ask(engine.TaskList, engine.Input{})
	if err != nil {
		return Snapshot{}, err
	}
	goals, err := w.Ask(engine.ObjectiveList, engine.Input{})
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
	held := 0
	if listed.InFocus != nil {
		held = *listed.InFocus
	}
	found := Snapshot{Focus: held, Version: version}
	for _, t := range listed.Tasks {
		card, err := cardOf(w, t, held)
		if err != nil {
			return Snapshot{}, err
		}
		found.Cards = append(found.Cards, card)
	}
	for _, o := range goals.Objectives {
		goal, err := goalOf(w, o)
		if err != nil {
			return Snapshot{}, err
		}
		found.Goals = append(found.Goals, goal)
	}
	return found, nil
}

// numbered is what every operation about one task or one objective is given.
// Its own function because it is written five times otherwise, and a map built
// by hand five times is a map spelt wrong once.
func numbered(n int) engine.Input {
	return engine.Input{Given: map[string]string{"number": strconv.Itoa(n)}}
}

// cardOf turns one task into what a card draws.
func cardOf(w Work, t task.Task, held int) (Card, error) {
	told := task.OfTask(t)
	// The deadline crosses the membrane written the way a person reads it, and
	// the board needs the day itself to say "in two days". Read back with the
	// same parser that wrote it, rather than a second layout kept here — a
	// second layout is a second thing to keep true.
	due, err := domain.ParseDeadline(told.Deadline)
	if err != nil {
		return Card{}, err
	}
	shown, err := w.Ask(engine.TaskShow, numbered(told.Number))
	along := shown.Progress
	if err != nil {
		return Card{}, err
	}
	return Card{
		Number:      told.Number,
		Parent:      told.Parent,
		Title:       told.Title,
		State:       told.State,
		Priority:    told.Priority,
		Described:   told.Description != "",
		Description: told.Description,
		Labels:      told.Labels,
		Blockers:    told.BlockedBy,
		Because:     told.Because,
		Deadline:    due,
		Estimate:    told.Estimate,
		Spent:       told.TimeSpent,
		Objectives:  told.Objectives,
		SubDone:     along.Done(),
		SubTotal:    along.Total(),
		InFocus:     told.Number == held,
	}, nil
}

// goalOf turns one objective into what a card draws.
func goalOf(w Work, o objective.Objective) (Goal, error) {
	told := objective.OfObjective(o)
	due, err := domain.ParseDeadline(told.Deadline)
	if err != nil {
		return Goal{}, err
	}
	shown, err := w.Ask(engine.ObjectiveShow, numbered(told.Number))
	along := shown.Progress
	if err != nil {
		return Goal{}, err
	}
	return Goal{
		Number:      told.Number,
		Title:       told.Title,
		State:       told.State,
		Priority:    told.Priority,
		Described:   told.Description != "",
		Description: told.Description,
		Labels:      told.Labels,
		Deadline:    due,
		Estimate:    told.Estimate,
		Done:        along.Done(),
		Total:       along.Total(),
	}, nil
}
