package engine

import (
	"slices"
	"strconv"
	"time"

	"cairn/internal/domain"
	"cairn/internal/objective"
	"cairn/internal/task"
)

// Work is what the board needs from whatever holds the work.
type Work interface {
	Ask(Operation, Input) (Answer, error)

	Version() (int64, error)
}

// Card is one task, flattened into what a board draws.
type Card struct {
	Number      int
	Title       string
	State       string
	Archived    bool
	Priority    string
	Labels      []string
	Blockers    []int
	Because     string
	Deadline    time.Time
	Estimate    string
	Spent       string
	Overrun     bool
	Objectives  []int
	Description string
	Steps       []Part
	SubDone     int
	SubTotal    int
	InFocus     bool
	Freed       bool
	Context     string
	Said        []string
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
	Said        []string
}

// Described says whether anything is written about the objective.
func (g Goal) Described() bool { return g.Description != "" }

// Snapshot is everything the board draws, read in one go.
type Snapshot struct {
	Cards   []Card
	Goals   []Goal
	Focus   []int
	Version int64
}

// Look reads the whole of it.
func Look(w Work) (Snapshot, error) {
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
	var held []int
	if listed.InFocus != nil {
		held = *listed.InFocus
	}
	found := Snapshot{Focus: held, Version: version}
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
func Numbered(n int) Input {
	return Input{Given: map[string]string{"number": strconv.Itoa(n)}}
}

func shownAll(kept []string) []string {
	shown := make([]string, 0, len(kept))
	for _, k := range kept {
		shown = append(shown, domain.Shown(k))
	}
	return shown
}

func cardOf(t task.Task, held, freed []int, along map[int]domain.Progress, contexts map[int]string) (Card, error) {
	told := task.OfTask(t)
	due, err := domain.ParseDeadline(told.Deadline)
	if err != nil {
		return Card{}, err
	}
	sofar := along[told.Number]
	return Card{
		Number:      told.Number,
		Steps:       partsOf(told.Steps),
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

func partsOf(steps []task.Step) []Part {
	if len(steps) == 0 {
		return nil
	}
	parts := make([]Part, 0, len(steps))
	for i, s := range steps {
		parts = append(parts, Part{Number: i + 1, Title: domain.Shown(s.Title), Done: s.Done})
	}
	return parts
}

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
