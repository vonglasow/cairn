package engine

import (
	"cairn/internal/domain"
	"cairn/internal/objective"
	"cairn/internal/store"
	"cairn/internal/task"
)

// ErrNoSuchObjective is the name the engine answers with when an objective is asked for by a number nothing carries.
var ErrNoSuchObjective = store.ErrNoSuchObjective

// ErrObjectiveCarriesTasks is why an objective with work under it is not deleted: the tasks would be left pointing at a number nothing carries.
var ErrObjectiveCarriesTasks = objective.ErrObjectiveCarriesTasks

type objectives interface {
	KeepObjective(objective.Objective) (objective.Objective, error)
	FindObjective(int) (objective.Objective, error)
	AllObjectives() ([]objective.Objective, error)
	NarrowedObjectives(*domain.Narrowing) ([]objective.Objective, error)
	UpdateObjective(objective.Objective) error
	DeleteObjective(int) error
	TasksServing(int) ([]task.Task, error)
}

// ObjectiveChanges are the attributes an operation was given for an objective.
type ObjectiveChanges struct {
	Title       *string
	Description *string
	State       *string
	Priority    *string
	Labels      *[]string
	Deadline    *string
	Estimate    *string
}

func (c ObjectiveChanges) named() bool {
	return c.Title != nil || c.Description != nil || c.State != nil ||
		c.Priority != nil || c.Labels != nil || c.Deadline != nil || c.Estimate != nil
}

func (e *Engine) addObjective(title string, c ObjectiveChanges) (objective.Objective, error) {
	o, err := objective.NewObjective(title)
	if err != nil {
		return objective.Objective{}, err
	}
	if o, err = applyObjective(o, c, creating); err != nil {
		return objective.Objective{}, err
	}
	return e.objectives.KeepObjective(o)
}

func (e *Engine) listObjectives() ([]objective.Objective, error) {
	return e.objectives.AllObjectives()
}

func (e *Engine) narrowedObjectives(kept *domain.Narrowing) ([]objective.Objective, error) {
	return e.objectives.NarrowedObjectives(kept)
}

func (e *Engine) showObjective(number int) (objective.Objective, error) {
	return e.objectives.FindObjective(number)
}

func (e *Engine) editObjective(number int, c ObjectiveChanges) (objective.Objective, error) {
	if !c.named() {
		return objective.Objective{}, ErrNothingToChange
	}
	o, err := e.objectives.FindObjective(number)
	if err != nil {
		return objective.Objective{}, err
	}
	if o, err = applyObjective(o, c, changing); err != nil {
		return objective.Objective{}, err
	}
	if err := e.objectives.UpdateObjective(o); err != nil {
		return objective.Objective{}, err
	}
	return o, nil
}

func (e *Engine) deleteObjective(number int) error {
	serving, err := e.objectives.TasksServing(number)
	if err != nil {
		return err
	}
	if err := objective.MayDelete(number, len(serving)); err != nil {
		return err
	}
	return e.objectives.DeleteObjective(number)
}

func factsOfObjective(o objective.Objective) objective.Carried { return o.Carried() }

func applyObjective(o objective.Objective, c ObjectiveChanges, atCreation bool) (objective.Objective, error) {
	a := o.Carried()
	if err := applyCarried(carriedTo{&a.Title, &a.Description, &a.Priority, &a.Labels, &a.Deadline, &a.Estimate}, "an objective", carriedChanges{
		title: c.Title, description: c.Description, priority: c.Priority,
		labels: c.Labels, deadline: c.Deadline, estimate: c.Estimate,
	}, atCreation); err != nil {
		return objective.Objective{}, err
	}
	if c.State != nil {
		state, err := objective.ParseObjectiveState(*c.State)
		if err != nil {
			return objective.Objective{}, err
		}
		a.State = state
	}
	return objective.Made(a), nil
}

func (e *Engine) canServe(t task.Task) error {
	served := factsOf(t).Objectives
	if len(served) == 0 {
		return nil
	}
	for _, one := range served {
		if _, err := e.objectives.FindObjective(one); err != nil {
			return err
		}
	}
	return nil
}

func (e *Engine) tasksUnder(number int) ([]task.Task, error) {
	return e.objectives.TasksServing(number)
}

func (e *Engine) objectiveProgress(number int) (domain.Progress, error) {
	if _, err := e.objectives.FindObjective(number); err != nil {
		return domain.Progress{}, err
	}
	_, objectives, err := e.alongEvery(nil, false)
	if err != nil {
		return domain.Progress{}, err
	}
	return objectives[number], nil
}

func (e *Engine) alongEvery(read []task.Task, everything bool) (tasks, objectives map[int]domain.Progress, err error) {
	all := read
	if !everything {
		every := domain.Narrowing{}
		every.WithArchived()
		if all, err = e.tasks.Narrowed(&every); err != nil {
			return nil, nil, err
		}
	}
	tasks, objectives = map[int]domain.Progress{}, map[int]domain.Progress{}
	serves := map[int][]int{}
	for _, t := range all {
		f := factsOf(t)
		serves[f.Number] = f.Objectives
		for _, o := range f.Objectives {
			p := objectives[o]
			p.Tasks = counted(p.Tasks, t)
			objectives[o] = p
		}
	}
	for _, t := range all {
		f := factsOf(t)
		steps := stepsTallied(f.Steps)
		if steps.Total == 0 {
			continue
		}
		tasks[f.Number] = domain.Progress{Steps: steps}
		for _, o := range serves[f.Number] {
			p := objectives[o]
			p.Steps.Total += steps.Total
			p.Steps.Done += steps.Done
			objectives[o] = p
		}
	}
	return tasks, objectives, nil
}

func stepsTallied(steps []task.Step) domain.Tally {
	tally := domain.Tally{Total: len(steps)}
	for _, s := range steps {
		if s.Done {
			tally.Done++
		}
	}
	return tally
}

func (e *Engine) taskProgress(number int) (domain.Progress, error) {
	t, err := e.tasks.Find(number)
	if err != nil {
		return domain.Progress{}, err
	}
	return domain.Progress{Steps: stepsTallied(factsOf(t).Steps)}, nil
}

func counted(tally domain.Tally, t task.Task) domain.Tally { return task.Count(tally, t) }

var objectiveAttributes = []Argument{
	{Name: "desc", Kind: Text, Means: "what the objective is about"},
	{Name: "state", Kind: Text, OneOf: oneOf(objective.ObjectiveStates), Outside: objective.ErrNotAnObjectiveState, Means: "where the objective stands"},
	{Name: "priority", Kind: Text, OneOf: oneOf(domain.Priorities), Outside: domain.ErrNotAPriority, Means: "the priority it carries"},
	{Name: "label", Kind: List, Means: "a label it carries"},
	{Name: "deadline", Kind: Day, Means: "the day it is due"},
	{Name: "estimate", Kind: Duration, Means: "how long it should take"},
}

func objectiveChanges(in Input) ObjectiveChanges {
	var c ObjectiveChanges
	c.Description = in.Change("desc")
	c.State = in.Change("state")
	c.Priority = in.Change("priority")
	c.Labels = in.Changed("label")
	c.Deadline = in.Change("deadline")
	c.Estimate = in.Change("estimate")
	return c
}

func readObjective(eng *Engine, n int) (objective.Objective, []task.Task, domain.Progress, error) {
	o, err := eng.showObjective(n)
	if err != nil {
		return objective.Objective{}, nil, domain.Progress{}, err
	}
	under, err := eng.tasksUnder(n)
	if err != nil {
		return objective.Objective{}, nil, domain.Progress{}, err
	}
	p, err := eng.objectiveProgress(n)
	if err != nil {
		return objective.Objective{}, nil, domain.Progress{}, err
	}
	return o, under, p, nil
}

// ObjectiveNew records a new objective.
var ObjectiveNew = Operation{
	Thing:   "objective",
	Verb:    "new",
	Short:   "Make a new objective",
	Does:    "Make a new objective, carrying whatever of its attributes were given.",
	Example: "cairn objective new \"Ship cairn 1.0\" --deadline 2026-12-01",
	Level:   Write,
	Needs:   AStore,
	Again:   DoesItAgain,
	Answers: Objective{},
	Takes: append([]Argument{
		{Name: "title", Kind: Text, Positional: true, Required: true, Means: "what the objective is called"},
	}, objectiveAttributes...),
	Run: func(on On, in Input) (Answer, error) {
		title := ""
		if given := in.Change("title"); given != nil {
			title = *given
		}
		o, err := on.Engine.addObjective(title, objectiveChanges(in))
		if err != nil {
			return Answer{}, withTypedTitle(err, title)
		}
		return Answer{Objective: &o, Progress: domain.Progress{}}, nil
	},
}

// ObjectiveList names every objective, or only the ones carrying what was named.
var ObjectiveList = Operation{
	Thing: "objective",
	Verb:  "list",
	Short: "List objectives, all of them or only some",
	Does:  "List every objective, or only the ones carrying everything named. An objective has to carry all of it, not one of it.",
	Example: `cairn objective list
cairn objective list --state reached`,
	Level:   Read,
	Needs:   AStore,
	Answers: Objectives{},
	Takes:   objectiveNarrowing,
	Run: func(on On, in Input) (Answer, error) {
		kept, err := keeping(in, objectiveNarrowing)
		if err != nil {
			return Answer{}, err
		}
		all, err := on.Engine.narrowedObjectives(kept)
		if err != nil {
			return Answer{}, err
		}
		_, along, err := on.Engine.alongEvery(nil, false)
		if err != nil {
			return Answer{}, err
		}
		if kept.Narrows() {
			return Answer{MatchedObjectives: all, Along: along}, nil
		}
		return Answer{Objectives: all, Along: along}, nil
	},
}

var objectiveNarrowing = narrowedBy(objectiveAttributes, map[string]string{
	"state":    "keep only the objectives standing here",
	"priority": "keep only the objectives carrying this priority",
	"label":    "keep only the objectives carrying this label; named twice, only those carrying both",
})

// ObjectiveShow says what an objective carries, the tasks serving it and how far along it is.
var ObjectiveShow = Operation{
	Thing:   "objective",
	Verb:    "show",
	Short:   "Show what an objective carries, and how far along it is",
	Does:    "Show what an objective carries, what tasks serve it and how far along it is.",
	Example: "cairn objective show 1",
	Level:   Read,
	Needs:   AStore,
	Answers: Objective{},
	Takes: []Argument{
		{Name: "number", Kind: Number, Positional: true, Required: true, Of: "an objective", Means: "the number of the objective"},
	},
	Run: func(on On, in Input) (Answer, error) {
		n, _ := in.Whole("number")
		o, under, p, err := readObjective(on.Engine, n)
		if err != nil {
			return Answer{}, err
		}
		return Answer{Objective: &o, Under: under, Progress: p}, nil
	},
}

// ObjectiveEdit changes what an objective carries.
var ObjectiveEdit = Operation{
	Thing:      "objective",
	Verb:       "edit",
	Short:      "Change what an objective carries",
	Does:       "Change what an objective carries.",
	Example:    "cairn objective edit 1 --state reached",
	Level:      Write,
	Needs:      AStore,
	Again:      ChangesNothingMore,
	Overwrites: true,
	Answers:    Objective{},
	Takes: append([]Argument{
		{Name: "number", Kind: Number, Positional: true, Required: true, Of: "an objective", Means: "the number of the objective"},
		{Name: "title", Kind: Text, Means: "what the objective is called"},
	}, objectiveAttributes...),
	Run: func(on On, in Input) (Answer, error) {
		n, _ := in.Whole("number")
		changes := objectiveChanges(in)
		changes.Title = in.Change("title")
		if _, err := on.Engine.editObjective(n, changes); err != nil {
			title := ""
			if changes.Title != nil {
				title = *changes.Title
			}
			return Answer{}, withTypedTitle(editReason(err), title)
		}
		o, under, p, err := readObjective(on.Engine, n)
		if err != nil {
			return Answer{}, err
		}
		return Answer{Objective: &o, Under: under, Progress: p}, nil
	},
}

// ObjectiveSearch names every objective carrying everything named, wherever among its facts each one happens to sit.
var ObjectiveSearch = Operation{
	Thing:   "objective",
	Verb:    "search",
	Short:   "Look for an objective by anything it carries",
	Does:    "List every objective carrying everything named, wherever among its facts each one sits.",
	Example: "cairn objective search ship",
	Level:   Read,
	Needs:   AStore,
	Answers: Objectives{},
	Takes: []Argument{
		{Name: "named", Kind: List, Positional: true, Required: true, Means: "what to look for, read the way it is shown"},
	},
	Run: func(on On, in Input) (Answer, error) {
		named := in.Changed("named")
		all, err := on.Engine.listObjectives()
		if err != nil {
			return Answer{}, err
		}
		found := []objective.Objective{}
		for _, o := range all {
			if named != nil && objective.Matching(o, *named) {
				found = append(found, o)
			}
		}
		return Answer{MatchedObjectives: found}, nil
	},
}

// ObjectiveDelete removes an objective.
var ObjectiveDelete = Operation{
	Thing:   "objective",
	Verb:    "delete",
	Short:   "Delete an objective",
	Does:    "Delete an objective.",
	Example: "cairn objective delete 1 --yes",
	Level:   Full,
	Needs:   AStore,
	Again:   ChangesNothingMore,
	Answers: Deleted{},
	Takes: []Argument{
		{Name: "number", Kind: Number, Positional: true, Required: true, Of: "an objective", Means: "the number of the objective"},
		alsoOf("an objective"),
	},
	Run: func(on On, in Input) (Answer, error) {
		numbers, err := together(in)
		if err != nil {
			return Answer{}, err
		}
		for _, n := range numbers {
			if err := on.Engine.deleteObjective(n); err != nil {
				return Answer{}, err
			}
		}
		return Answer{Deleted: &numbers[0], Also: numbers[1:]}, nil
	},
}
