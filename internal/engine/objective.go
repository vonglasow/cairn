package engine

import (
	"cairn/internal/engine/change"
	"cairn/internal/engine/filter"
	"cairn/internal/engine/item"
	"cairn/internal/engine/label"
	"cairn/internal/engine/objective"
	"cairn/internal/engine/task"
)

// ObjectiveChanges are the attributes an operation was given for an objective.
type ObjectiveChanges struct {
	Title       *string
	Description *string
	Replace     *string
	By          *string
	State       *string
	Priority    *string
	Labels      *[]string
	Deadline    *string
	Estimate    *string
	Parent      *string
	WaitsOn     *[]string
}

func (c ObjectiveChanges) named() bool {
	return c.Title != nil || c.Description != nil || c.Replace != nil || c.By != nil || c.State != nil ||
		c.Priority != nil || c.Labels != nil || c.Deadline != nil || c.Estimate != nil || c.Parent != nil ||
		c.WaitsOn != nil
}

func (e *Engine) addObjective(title string, c ObjectiveChanges) (objective.Objective, error) {
	o, err := objective.NewObjective(title)
	if err != nil {
		return objective.Objective{}, err
	}
	if o, err = applyObjective(o, c, creating); err != nil {
		return objective.Objective{}, err
	}
	if err := e.inScope(nil, o.Carried().Labels, label.OnObjectives); err != nil {
		return objective.Objective{}, err
	}
	if err := e.mayHang(o); err != nil {
		return objective.Objective{}, err
	}
	if err := e.mayWait(o); err != nil {
		return objective.Objective{}, err
	}
	return e.objectives.KeepObjective(o)
}

func (e *Engine) listObjectives() ([]objective.Objective, error) {
	return e.objectives.AllObjectives()
}

func (e *Engine) narrowedObjectives(kept *filter.Narrowing) ([]objective.Objective, error) {
	return e.objectives.NarrowedObjectives(kept)
}

func (e *Engine) showObjective(number int) (objective.Objective, error) {
	return e.objectives.FindObjective(number)
}

func (e *Engine) editObjective(number int, c ObjectiveChanges) (objective.Objective, error) {
	if !c.named() {
		return objective.Objective{}, errNothingToChange
	}
	was, err := e.objectives.FindObjective(number)
	if err != nil {
		return objective.Objective{}, err
	}
	o, err := applyObjective(was, c, changing)
	if err != nil {
		return objective.Objective{}, err
	}
	if err := e.inScope(was.Carried().Labels, o.Carried().Labels, label.OnObjectives); err != nil {
		return objective.Objective{}, err
	}
	if err := e.mayHang(o); err != nil {
		return objective.Objective{}, err
	}
	if err := e.mayWait(o); err != nil {
		return objective.Objective{}, err
	}
	if err := e.objectives.UpdateObjective(o); err != nil {
		return objective.Objective{}, err
	}
	if now := o.Carried().Parent; now != was.Carried().Parent {
		if err := e.objectives.PlaceObjective(number, now, -1); err != nil {
			return objective.Objective{}, err
		}
	}
	return o, nil
}

func applyObjective(o objective.Objective, c ObjectiveChanges, atCreation bool) (objective.Objective, error) {
	a := o.Carried()
	if err := applyCarried(carriedTo{&a.Title, &a.Description, &a.Priority, &a.Labels, &a.Deadline, &a.Estimate}, "an objective", carriedChanges{
		title: c.Title, description: c.Description, replace: c.Replace, by: c.By, priority: c.Priority,
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
	if c.Parent != nil {
		parent, err := parentNamed(*c.Parent)
		if err != nil {
			return objective.Objective{}, err
		}
		a.Parent = parent
	}
	if c.WaitsOn != nil {
		on, err := change.ChangedNumbers(*c.WaitsOn, a.WaitsOn, atCreation, objective.ErrWaitsOnRequired, objective.ErrWaitsOnBothWays, objective.ErrWaitsOnSignAtBirth)
		if err != nil {
			return objective.Objective{}, err
		}
		a.WaitsOn = on
	}
	return objective.Made(a), nil
}

func (e *Engine) canServe(t task.Task) error {
	served := t.Carried().Objectives
	if len(served) == 0 {
		return nil
	}
	for _, one := range served {
		if _, err := e.objectives.FindObjective(one); err != nil {
			return err
		}
	}
	return e.countedOnce(t.Carried().Number, served, nil)
}

func (e *Engine) tasksUnder(number int) ([]task.Task, error) {
	t, err := e.tree()
	if err != nil {
		return nil, err
	}
	return spentTasks{history: e.history}.told(e.store.TasksServing(t.Reach(number)...))
}

func (e *Engine) objectiveProgress(number int) (item.Progress, error) {
	if _, err := e.objectives.FindObjective(number); err != nil {
		return item.Progress{}, err
	}
	_, objectives, err := e.alongEvery(nil, false)
	if err != nil {
		return item.Progress{}, err
	}
	return objectives[number], nil
}

func (e *Engine) alongEvery(read []task.Task, everything bool) (tasks, objectives map[int]item.Progress, err error) {
	all := read
	if !everything {
		every := filter.Narrowing{}
		every.WithArchived()
		if all, err = e.tasks.Narrowed(&every); err != nil {
			return nil, nil, err
		}
	}
	tr, err := e.tree()
	if err != nil {
		return nil, nil, err
	}
	tasks, objectives = map[int]item.Progress{}, map[int]item.Progress{}
	serves := map[int][]int{}
	for _, t := range all {
		f := t.Carried()
		serves[f.Number] = tr.Counted(f.Objectives)
		for _, o := range serves[f.Number] {
			p := objectives[o]
			p.Tasks = counted(p.Tasks, t)
			objectives[o] = p
		}
	}
	for _, t := range all {
		f := t.Carried()
		steps := stepsTallied(f.Steps)
		if steps.Total == 0 {
			continue
		}
		tasks[f.Number] = item.Progress{Steps: steps}
		for _, o := range serves[f.Number] {
			p := objectives[o]
			p.Steps.Total += steps.Total
			p.Steps.Done += steps.Done
			objectives[o] = p
		}
	}
	return tasks, objectives, nil
}

func stepsTallied(steps []task.Step) item.Tally {
	tally := item.Tally{Total: len(steps)}
	for _, s := range steps {
		if s.Done {
			tally.Done++
		}
	}
	return tally
}

func (e *Engine) taskProgress(number int) (item.Progress, error) {
	t, err := e.tasks.Find(number)
	if err != nil {
		return item.Progress{}, err
	}
	return item.Progress{Steps: stepsTallied(t.Carried().Steps)}, nil
}

func counted(tally item.Tally, t task.Task) item.Tally { return task.Count(tally, t) }

var objectiveAttributes = []Argument{
	{Name: "desc", Kind: Text, Means: "what the objective is about"},
	{Name: "state", Kind: Text, OneOf: oneOf(objective.ObjectiveStates), Outside: objective.ErrNotAnObjectiveState, Means: "where the objective stands"},
	{Name: "priority", Kind: Text, OneOf: oneOf(item.Priorities), Outside: item.ErrNotAPriority, Means: "the priority it carries"},
	{Name: "label", Kind: List, Means: "a label it carries; when changing, x or +x adds one, -x takes one off, =x replaces them all"},
	{Name: "deadline", Kind: day, Means: "the day it is due"},
	{Name: "estimate", Kind: duration, Means: "how long it should take"},
	{Name: "parent", Kind: Text, Of: "an objective", Means: "the number of the objective it sits under, two levels deep at most; none to stand on its own"},
	{Name: "waits-on", Kind: List, Of: "an objective", Means: "the number of an objective it waits on, which warns when work starts on it and never blocks; when changing, 6 or +6 adds one, -6 takes one away, =6 replaces them all"},
}

func objectiveChanges(in Input) ObjectiveChanges {
	var c ObjectiveChanges
	c.Description = in.change("desc")
	c.State = in.change("state")
	c.Priority = in.change("priority")
	c.Labels = in.changed("label")
	c.Deadline = in.change("deadline")
	c.Estimate = in.change("estimate")
	c.Parent = in.change("parent")
	c.WaitsOn = in.changed("waits-on")
	return c
}

func readObjective(eng *Engine, n int) (Answer, error) {
	o, err := eng.showObjective(n)
	if err != nil {
		return Answer{}, err
	}
	under, err := eng.tasksUnder(n)
	if err != nil {
		return Answer{}, err
	}
	p, err := eng.objectiveProgress(n)
	if err != nil {
		return Answer{}, err
	}
	f, err := eng.familyOf(n)
	if err != nil {
		return Answer{}, err
	}
	tree, waiters, err := eng.seating()
	if err != nil {
		return Answer{}, err
	}
	return Answer{Objective: &o, Under: under, Progress: p, Family: &f, Tree: tree, Waiters: waiters}, nil
}

// ObjectiveNew records a new objective.
var ObjectiveNew = Operation{
	Thing:   "objective",
	Verb:    "new",
	Short:   "Make a new objective",
	Does:    "Make a new objective, carrying whatever of its attributes were given.",
	Example: "cairn objective new \"Ship cairn 1.0\" --deadline 2026-12-01",
	Level:   write,
	Needs:   aStore,
	Again:   doesItAgain,
	Answers: objectiveRow{},
	Takes: append([]Argument{
		{Name: "title", Kind: Text, Positional: true, Required: true, Means: "what the objective is called"},
	}, objectiveAttributes...),
	Run: func(on workedOn, in Input) (Answer, error) {
		title := ""
		if given := in.change("title"); given != nil {
			title = *given
		}
		o, err := on.Engine.addObjective(title, objectiveChanges(in))
		if err != nil {
			return Answer{}, task.WithTypedTitle(err, title)
		}
		read, err := readObjective(on.Engine, o.Carried().Number)
		if err != nil {
			return Answer{}, err
		}
		read.Proposes, err = on.Engine.proposedBy(nil, o)
		return read, err
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
	Needs:   aStore,
	Answers: objectivesListed{},
	Takes:   objectiveNarrowing,
	Run: func(on workedOn, in Input) (Answer, error) {
		kept, err := keeping(in, objectiveNarrowing, nil)
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
		spent, err := on.Engine.spentEvery()
		if err != nil {
			return Answer{}, err
		}
		tree, waiters, err := on.Engine.seating()
		if err != nil {
			return Answer{}, err
		}
		if kept.Narrows() {
			return Answer{MatchedObjectives: all, Along: along, Spent: spent, Tree: tree, Waiters: waiters}, nil
		}
		return Answer{Objectives: all, Along: along, Spent: spent, Tree: tree, Waiters: waiters}, nil
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
	Needs:   aStore,
	Answers: objectiveRow{},
	Takes: []Argument{
		{Name: "number", Kind: Number, Positional: true, Required: true, Of: "an objective", Means: "the number of the objective"},
	},
	Run: func(on workedOn, in Input) (Answer, error) {
		n, _ := in.whole("number")
		return readObjective(on.Engine, n)
	},
}

// ObjectiveEdit changes what an objective carries.
var ObjectiveEdit = Operation{
	Thing:      "objective",
	Verb:       "edit",
	Short:      "Change what an objective carries",
	Does:       "Change what an objective carries.",
	Example:    "cairn objective edit 1 --state reached",
	Level:      write,
	Needs:      aStore,
	Again:      ChangesNothingMore,
	Overwrites: true,
	Answers:    objectiveRow{},
	Takes: append([]Argument{
		{Name: "number", Kind: Number, Positional: true, Required: true, Of: "an objective", Means: "the number of the objective"},
		{Name: "title", Kind: Text, Means: "what the objective is called"},
	}, append(objectiveAttributes, replaceArguments...)...),
	Run: func(on workedOn, in Input) (Answer, error) {
		n, _ := in.whole("number")
		changes := objectiveChanges(in)
		changes.Replace, changes.By = in.change("replace"), in.change("by")
		changes.Title = in.change("title")
		was, _ := on.Engine.showObjective(n)
		edited, err := on.Engine.editObjective(n, changes)
		if err != nil {
			title := ""
			if changes.Title != nil {
				title = *changes.Title
			}
			return Answer{}, task.WithTypedTitle(editReason(err), title)
		}
		read, err := readObjective(on.Engine, n)
		if err != nil {
			return Answer{}, err
		}
		read.Moved = movedOfObjective(was, edited)
		read.Proposes, err = on.Engine.proposedBy(&was, edited)
		return read, err
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
	Needs:   aStore,
	Answers: objectivesListed{},
	Takes: []Argument{
		{Name: "named", Kind: List, Positional: true, Required: true, Means: "what to look for, read the way it is shown"},
	},
	Run: func(on workedOn, in Input) (Answer, error) {
		named := in.changed("named")
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
		tree, waiters, err := on.Engine.seating()
		return Answer{MatchedObjectives: found, Tree: tree, Waiters: waiters}, err
	},
}

// ObjectiveDelete removes an objective, and what was said to go with it.
var ObjectiveDelete = Operation{
	Thing: "objective",
	Verb:  "delete",
	Short: "Delete an objective",
	Does:  "Delete an objective. One holding sub-objectives or carrying tasks is refused, the refusal listing them, until it is said what becomes of each: a sub-objective is kept, standing on its own, or deleted; a task is detached or deleted. Nothing is decided for the person.",
	Example: `cairn objective delete 1 --yes
cairn objective delete 4 --sub-objectives keep --tasks detach --yes
cairn objective delete 4 --keep 5,7 --delete 6 --tasks delete --yes`,
	Level:   Full,
	Needs:   aStore,
	Again:   ChangesNothingMore,
	Answers: Deleted{},
	Takes: append([]Argument{
		{Name: "number", Kind: Number, Positional: true, Required: true, Of: "an objective", Means: "the number of the objective"},
		alsoOf("an objective"),
	}, deletionChoices...),
	Run: func(on workedOn, in Input) (Answer, error) {
		numbers, err := together(in)
		if err != nil {
			return Answer{}, err
		}
		d, err := deletionOf(in)
		if err != nil {
			return Answer{}, err
		}
		went, err := on.Engine.deleteObjectives(numbers, d)
		if err != nil {
			return Answer{}, err
		}
		return Answer{Deleted: &numbers[0], Also: numbers[1:], Went: &went}, nil
	},
}

func reviewOf(under []task.Task) objective.Review {
	var r objective.Review
	for _, t := range under {
		c := t.Carried()
		r.Spent += c.Spent
		r.Left = r.Left.Counting(c.Number, c.State, c.Archived, c.Deadline)
	}
	return r
}

func leftOf(under []task.Task) []int {
	left := []int{}
	for _, t := range under {
		if c := t.Carried(); !c.Archived && c.State != item.Done {
			left = append(left, c.Number)
		}
	}
	return left
}
