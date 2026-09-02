package board

import (
	"time"

	"cairn/internal/domain"
	"cairn/internal/engine"
	"cairn/internal/objective"
	"cairn/internal/task"
	"cairn/internal/trace"
)

// Work is what the board needs from whatever holds the work. Declared here, by
// the consumer, so a test answers for it without a database and the board never
// learns what is behind it.
//
// Everything that changes anything goes through the same operations the command
// line calls. A key pressed here and a command typed there reach the engine by
// the same door, or one of them would be a second place where a rule is kept.
type Work interface {
	List() ([]task.Task, error)
	ListObjectives() ([]objective.Objective, error)
	Focus() (int, error)
	TaskProgress(int) (domain.Progress, error)
	ObjectiveProgress(int) (domain.Progress, error)
	History(int) ([]trace.Act, error)
	Add(string, engine.Changes) (task.Task, error)
	Edit(int, engine.Changes) (task.Task, error)
	AddObjective(string, engine.ObjectiveChanges) (objective.Objective, error)
	EditObjective(int, engine.ObjectiveChanges) (objective.Objective, error)
	Delete(int) error
	DeleteObjective(int) error
	Move(int, engine.Place, int) error
	PutInFocus(int) (task.Task, error)
	LetGo() error
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
	tasks, err := w.List()
	if err != nil {
		return Snapshot{}, err
	}
	goals, err := w.ListObjectives()
	if err != nil {
		return Snapshot{}, err
	}
	held, err := w.Focus()
	if err != nil {
		return Snapshot{}, err
	}
	version, err := w.Version()
	if err != nil {
		return Snapshot{}, err
	}
	found := Snapshot{Focus: held, Version: version}
	for _, t := range tasks {
		card, err := cardOf(w, t, held)
		if err != nil {
			return Snapshot{}, err
		}
		found.Cards = append(found.Cards, card)
	}
	for _, o := range goals {
		goal, err := goalOf(w, o)
		if err != nil {
			return Snapshot{}, err
		}
		found.Goals = append(found.Goals, goal)
	}
	return found, nil
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
	along, err := w.TaskProgress(told.Number)
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
	along, err := w.ObjectiveProgress(told.Number)
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
