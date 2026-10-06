package engine

import (
	"time"

	"cairn/internal/domain"
	"cairn/internal/objective"
	"cairn/internal/store"
	"cairn/internal/task"
)

// ErrNoSuchObjective is the name the engine answers with when an objective is
// asked for by a number nothing carries. A surface may not import the store,
// so the name it recognises has to live here.
var ErrNoSuchObjective = store.ErrNoSuchObjective

// ErrSubTaskServesNoObjective is one rule seen from three sides: a task never
// carries a parent and an objective at the same time. Re-exported from
// domain, which is where a task's own rules live: a surface asks the engine
// and may not import domain, so the name it recognises has to live here too.
var ErrSubTaskServesNoObjective = task.ErrSubTaskServesNoObjective

// ErrObjectiveCarriesTasks is why an objective with work under it is not
// deleted: the tasks would be left pointing at a number nothing carries.
// Re-exported from domain, which is where an objective's own rules live: a
// surface asks the engine and may not import domain, so the name it
// recognises has to live here too.
var ErrObjectiveCarriesTasks = objective.ErrObjectiveCarriesTasks

// objectives is what the engine needs from whatever keeps its objectives.
// Declared here, by the consumer, for the same reason tasks is.
// It names only what is called today. An interface listing what a caller might
// want one day is a promise nothing checks, and the store already offers the
// rest — each arrives here with the operation that needs it.
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
// A nil field was not named at all; a field pointing at an empty value was
// named and emptied — which is how an attribute is taken away.
//
// It is not Changes: an objective has no parent, no objective of its own,
// nothing in its way, no words beside that and no time spent, and its state is
// not a task's. A shared struct would offer every surface five fields that mean
// nothing here; what the two carry alike is applied once, by applyCarried.
type ObjectiveChanges struct {
	Title       *string
	Description *string
	State       *string
	Priority    *string
	Labels      *[]string
	Deadline    *string
	Estimate    *string
}

// named says whether anything at all was given to change.
func (c ObjectiveChanges) named() bool {
	return c.Title != nil || c.Description != nil || c.State != nil ||
		c.Priority != nil || c.Labels != nil || c.Deadline != nil || c.Estimate != nil
}

// addObjective records an objective under the title given.
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

// listObjectives hands back every objective, in the order they were added.
func (e *Engine) listObjectives() ([]objective.Objective, error) {
	return e.objectives.AllObjectives()
}

// narrowedObjectives hands back the objectives a narrowing keeps. Told nothing,
// it keeps everything, so this is ListObjectives with something to say.
func (e *Engine) narrowedObjectives(kept *domain.Narrowing) ([]objective.Objective, error) {
	return e.objectives.NarrowedObjectives(kept)
}

// showObjective hands back the objective carrying number.
func (e *Engine) showObjective(number int) (objective.Objective, error) {
	return e.objectives.FindObjective(number)
}

// editObjective changes what was named and leaves the rest alone.
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

// deleteObjective removes an objective, refusing while tasks are still under
// it. Cascading would take work nobody named, and setting those tasks loose
// silently would leave somebody wondering where their objective went.
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

// objectiveFacts is an objective's membrane for what the engine, or a test on
// it, has to read back off it — everything else is a rule the objective
// already owns, or a fact only a builder writing a new objective needs told
// to it.
type objectiveFacts struct {
	number             int
	title, description string
	state              objective.State
	priority           domain.Priority
	labels             []string
	deadline           time.Time
	estimate           time.Duration
}

func (f *objectiveFacts) Numbered(n int)             { f.number = n }
func (f *objectiveFacts) Titled(s string)            { f.title = s }
func (f *objectiveFacts) Standing(s objective.State) { f.state = s }
func (f *objectiveFacts) Carries(p domain.Priority)  { f.priority = p }
func (f *objectiveFacts) Described(s string)         { f.description = s }
func (f *objectiveFacts) NotDescribed()              { f.description = "" }
func (f *objectiveFacts) Labelled(l string)          { f.labels = append(f.labels, l) }
func (f *objectiveFacts) NotLabelled()               { f.labels = nil }
func (f *objectiveFacts) LabelsTold()                {}
func (f *objectiveFacts) Due(t time.Time)            { f.deadline = t }
func (f *objectiveFacts) NotDue()                    { f.deadline = time.Time{} }
func (f *objectiveFacts) Estimated(d time.Duration)  { f.estimate = d }
func (f *objectiveFacts) NotEstimated()              { f.estimate = 0 }

// factsOfObjective reads back what the engine needs from an objective,
// without reaching into fields that stopped being exported.
func factsOfObjective(o objective.Objective) objectiveFacts {
	var f objectiveFacts
	o.Tell(&f)
	return f
}

// applyObjective lays the changes over an objective, refusing any value it
// cannot make sense of. A field that was not named is left exactly as it was.
//
// It reads like apply for a task, and stays written out rather than shared:
// every line here is a call to a parser both already use, and the one piece
// with logic of its own — the labels — is changedLabels.
func applyObjective(o objective.Objective, c ObjectiveChanges, atCreation bool) (objective.Objective, error) {
	current := factsOfObjective(o)
	a := &objective.AsTold{}
	o.Tell(a)

	if err := applyCarried(a, "an objective", carriedChanges{
		title: c.Title, description: c.Description, priority: c.Priority,
		labels: c.Labels, deadline: c.Deadline, estimate: c.Estimate,
	}, current.labels, atCreation); err != nil {
		return objective.Objective{}, err
	}
	if c.State != nil {
		state, err := objective.ParseObjectiveState(*c.State)
		if err != nil {
			return objective.Objective{}, err
		}
		a.Standing(state)
	}
	return a.Objective(), nil
}

// canServe says whether a task may serve the objectives it was told to. Every
// one has to exist, and the task has to be one in its own right: what hangs
// under a task never serves an objective directly.
//
// One that does not exist turns down the whole set, and the task keeps what it
// had. Attaching the ones that do and dropping the rest would write a set
// nobody asked for, and say nothing about the one that went missing.
func (e *Engine) canServe(t task.Task) error {
	served := factsOf(t).objectives
	if len(served) == 0 {
		return nil
	}
	if err := t.MayServeAnObjective(); err != nil {
		// No number in front: at creation there is none yet, and "task 0" reads
		// as a task somebody could go and look at. On an edit the number was
		// just typed, so repeating it adds nothing either.
		return err
	}
	for _, one := range served {
		if _, err := e.objectives.FindObjective(one); err != nil {
			return err
		}
	}
	return nil
}

// tasksUnder hands back the tasks put directly under an objective. Sub-tasks
// are not among them: they are counted towards the objective, never listed as
// serving it.
func (e *Engine) tasksUnder(number int) ([]task.Task, error) {
	return e.objectives.TasksServing(number)
}

// objectiveProgress is how far along an objective is: the tasks put under it,
// and the sub-tasks hanging under those. The two are counted apart because
// they are not the same thing, and the objective has to exist to be asked.
func (e *Engine) objectiveProgress(number int) (domain.Progress, error) {
	if _, err := e.objectives.FindObjective(number); err != nil {
		return domain.Progress{}, err
	}
	// Counted the way every listing counts: one reading of every task, then
	// arithmetic. Asking for the parts of each task it serves read every link
	// of the store once per task — four queries a task, for one number.
	_, objectives, err := e.alongEvery(nil, false)
	if err != nil {
		return domain.Progress{}, err
	}
	return objectives[number], nil
}

// alongEvery is how far along every task and every objective is, counted from
// one reading of every task — on the board or off it, the way showing one
// counts them. Handed what was already read when that was everything.
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
		serves[f.number] = f.objectives
		for _, o := range f.objectives {
			p := objectives[o]
			p.Tasks = counted(p.Tasks, t)
			objectives[o] = p
		}
	}
	for _, t := range all {
		f := factsOf(t)
		if f.parent == task.NoParent {
			continue
		}
		p := tasks[f.parent]
		p.SubTasks = counted(p.SubTasks, t)
		tasks[f.parent] = p
		// A part counts towards an objective by way of the task it hangs
		// under, never on its own.
		for _, o := range serves[f.parent] {
			p := objectives[o]
			p.SubTasks = counted(p.SubTasks, t)
			objectives[o] = p
		}
	}
	return tasks, objectives, nil
}

// taskProgress is how far along a task is, counted over what hangs under it.
// A task never counts itself: two sub-tasks of which one is done is one of
// two, and nobody reads it as one of three.
func (e *Engine) taskProgress(number int) (domain.Progress, error) {
	under, err := e.tasks.Under(number)
	if err != nil {
		return domain.Progress{}, err
	}
	var p domain.Progress
	for _, sub := range under {
		p.SubTasks = counted(p.SubTasks, sub)
	}
	return p, nil
}

// counted adds one task to a tally by telling it every fact it carries and
// keeping only what a tally needs: the column it sits in.
func counted(tally domain.Tally, t task.Task) domain.Tally {
	c := task.NewCounter(tally)
	t.Tell(c)
	return c.Tally()
}

// objectiveAttributes are the seven attributes an objective carries, minus
// title, which is taken as a positional argument where it can be given at
// all.
//
// It is not taskAttributes with fields left out: an objective has no column
// to sit in the way a task does, nothing to hang under and no time spent on
// it, and a shared set would offer three flags that could only ever be
// refused.
var objectiveAttributes = []Argument{
	{Name: "desc", Kind: Text, Means: "what the objective is about"},
	{Name: "state", Kind: Text, OneOf: oneOf(objective.ObjectiveStates), Outside: objective.ErrNotAnObjectiveState, Means: "where the objective stands"},
	{Name: "priority", Kind: Text, OneOf: oneOf(domain.Priorities), Outside: domain.ErrNotAPriority, Means: "the priority it carries"},
	{Name: "label", Kind: List, Means: "a label it carries"},
	{Name: "deadline", Kind: Day, Means: "the day it is due"},
	{Name: "estimate", Kind: Duration, Means: "how long it should take"},
}

// objectiveChanges turns what was named in in into the changes
// addObjective and editObjective take.
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

// readObjective asks for everything one objective is shown with. Three calls
// that always travel together, so ObjectiveShow and ObjectiveEdit cannot each
// remember a different two of them.
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

// ObjectiveAdd records a new objective.
var ObjectiveAdd = Operation{
	Thing:   "objective",
	Verb:    "add",
	Short:   "Add an objective",
	Does:    "Add an objective, carrying whatever of its attributes were given.",
	Example: "cairn objective add \"Ship cairn 1.0\" --deadline 2026-12-01",
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
		// Nothing hangs under an objective that has just been created, so the
		// progress is asked of nobody rather than of the store.
		return Answer{Objective: &o, Progress: domain.Progress{}}, nil
	},
}

// ObjectiveList names every objective, or only the ones carrying what was
// named.
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
		// A narrowed listing is answered flat, the same way a narrowed listing
		// of tasks is. Nothing hangs under an objective in a listing, so the
		// two read alike either way — but the fields are kept apart so a reader
		// of the answer can still tell which question was asked.
		if kept.Narrows() {
			return Answer{MatchedObjectives: all, Along: along}, nil
		}
		return Answer{Objectives: all, Along: along}, nil
	},
}

// objectiveNarrowing are the facts a listing of objectives can be kept down to.
// Not what it serves, since nothing puts an objective under another.
var objectiveNarrowing = narrowedBy(objectiveAttributes, map[string]string{
	"state":    "keep only the objectives standing here",
	"priority": "keep only the objectives carrying this priority",
	"label":    "keep only the objectives carrying this label; named twice, only those carrying both",
})

// ObjectiveShow says what an objective carries, what hangs under it and how
// far along it is.
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

// ObjectiveSearch names every objective carrying everything named, wherever
// among its facts each one happens to sit.
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
	},
	Run: func(on On, in Input) (Answer, error) {
		n, _ := in.Whole("number")
		if err := on.Engine.deleteObjective(n); err != nil {
			return Answer{}, err
		}
		return Answer{Deleted: &n}, nil
	},
}
