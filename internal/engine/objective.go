package engine

import (
	"errors"
	"fmt"

	"cairn/internal/domain"
	"cairn/internal/store"
)

// ErrNoSuchObjective is the name the engine answers with when an objective is
// asked for by a number nothing carries. A surface may not import the store,
// so the name it recognises has to live here.
var ErrNoSuchObjective = store.ErrNoSuchObjective

// ErrSubTaskServesNoObjective is one rule seen from three sides: a task never
// carries a parent and an objective at the same time. A sub-task lives with
// the task it hangs under, counts towards that task's objective through it,
// and is never given one of its own — an objective it served alone would be
// one its parent might have nothing to do with.
var ErrSubTaskServesNoObjective = errors.New("a sub-task serves no objective")

// ErrObjectiveCarriesTasks is why an objective with work under it is not
// deleted: the tasks would be left pointing at a number nothing carries.
var ErrObjectiveCarriesTasks = errors.New("the objective carries tasks")

// objectives is what the engine needs from whatever keeps its objectives.
// Declared here, by the consumer, for the same reason tasks is.
// It names only what is called today. An interface listing what a caller might
// want one day is a promise nothing checks, and the store already offers the
// rest — each arrives here with the operation that needs it.
type objectives interface {
	KeepObjective(domain.Objective) (domain.Objective, error)
	FindObjective(int) (domain.Objective, error)
	AllObjectives() ([]domain.Objective, error)
	UpdateObjective(domain.Objective) error
	DeleteObjective(int) error
	TasksServing(int) ([]domain.Task, error)
}

// ObjectiveChanges are the attributes an operation was given for an objective.
// A nil field was not named at all; a field pointing at an empty value was
// named and emptied — which is how an attribute is taken away.
//
// It is not Changes: an objective has no parent, no objective of its own and
// no time spent, and its state is not a task's. A shared struct would offer
// every surface four fields that mean nothing here.
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

// AddObjective records an objective under the title given.
func (e *Engine) AddObjective(title string, c ObjectiveChanges) (domain.Objective, error) {
	o, err := domain.NewObjective(title)
	if err != nil {
		return domain.Objective{}, err
	}
	if o, err = applyObjective(o, c, creating); err != nil {
		return domain.Objective{}, err
	}
	return e.objectives.KeepObjective(o)
}

// ListObjectives hands back every objective, in the order they were added.
func (e *Engine) ListObjectives() ([]domain.Objective, error) {
	return e.objectives.AllObjectives()
}

// ShowObjective hands back the objective carrying number.
func (e *Engine) ShowObjective(number int) (domain.Objective, error) {
	return e.objectives.FindObjective(number)
}

// EditObjective changes what was named and leaves the rest alone.
func (e *Engine) EditObjective(number int, c ObjectiveChanges) (domain.Objective, error) {
	if !c.named() {
		return domain.Objective{}, ErrNothingToChange
	}
	o, err := e.objectives.FindObjective(number)
	if err != nil {
		return domain.Objective{}, err
	}
	if o, err = applyObjective(o, c, changing); err != nil {
		return domain.Objective{}, err
	}
	if err := e.objectives.UpdateObjective(o); err != nil {
		return domain.Objective{}, err
	}
	return o, nil
}

// DeleteObjective removes an objective, refusing while tasks are still under
// it. Cascading would take work nobody named, and setting those tasks loose
// silently would leave somebody wondering where their objective went.
func (e *Engine) DeleteObjective(number int) error {
	serving, err := e.objectives.TasksServing(number)
	if err != nil {
		return err
	}
	if len(serving) > 0 {
		return fmt.Errorf("objective %d: %w", number, ErrObjectiveCarriesTasks)
	}
	return e.objectives.DeleteObjective(number)
}

// applyObjective lays the changes over an objective, refusing any value it
// cannot make sense of. A field that was not named is left exactly as it was.
//
// It reads like apply for a task, and stays written out rather than shared:
// every line here is a call to a parser both already use, and the one piece
// with logic of its own — the labels — is changedLabels.
func applyObjective(o domain.Objective, c ObjectiveChanges, atCreation bool) (domain.Objective, error) {
	if c.Title != nil {
		retitled, err := domain.NewObjective(*c.Title)
		if err != nil {
			return domain.Objective{}, err
		}
		o.Title = retitled.Title
	}
	if c.Description != nil {
		o.Description = *c.Description
	}
	if c.State != nil {
		state, err := domain.ParseObjectiveState(*c.State)
		if err != nil {
			return domain.Objective{}, err
		}
		o.State = state
	}
	if c.Priority != nil {
		priority, err := domain.ParsePriority(*c.Priority)
		if err != nil {
			return domain.Objective{}, err
		}
		o.Priority = priority
	}
	if c.Labels != nil {
		labels, err := changedLabels(*c.Labels, o.Labels, atCreation)
		if err != nil {
			return domain.Objective{}, err
		}
		o.Labels = labels
	}
	if c.Deadline != nil {
		deadline, err := domain.ParseDeadline(*c.Deadline)
		if err != nil {
			return domain.Objective{}, err
		}
		o.Deadline = deadline
	}
	if c.Estimate != nil {
		estimate, err := domain.ParseEstimate(*c.Estimate)
		if err != nil {
			return domain.Objective{}, err
		}
		o.Estimate = estimate
	}
	return o, nil
}

// canServe says whether a task may serve the objective it was told to. The
// objective has to exist, and the task has to be one in its own right: what
// hangs under a task never serves an objective directly.
func (e *Engine) canServe(task domain.Task) error {
	if task.Objective == domain.NoObjective {
		return nil
	}
	if task.Parent != domain.NoParent {
		// No number in front: at creation there is none yet, and "task 0" reads
		// as a task somebody could go and look at. On an edit the number was
		// just typed, so repeating it adds nothing either.
		return ErrSubTaskServesNoObjective
	}
	if _, err := e.objectives.FindObjective(task.Objective); err != nil {
		return err
	}
	return nil
}

// TasksUnder hands back the tasks put directly under an objective. Sub-tasks
// are not among them: they are counted towards the objective, never listed as
// serving it.
func (e *Engine) TasksUnder(number int) ([]domain.Task, error) {
	return e.objectives.TasksServing(number)
}

// ObjectiveProgress is how far along an objective is: the tasks put under it,
// and the sub-tasks hanging under those. The two are counted apart because
// they are not the same thing, and the objective has to exist to be asked.
func (e *Engine) ObjectiveProgress(number int) (domain.Progress, error) {
	if _, err := e.objectives.FindObjective(number); err != nil {
		return domain.Progress{}, err
	}
	serving, err := e.objectives.TasksServing(number)
	if err != nil {
		return domain.Progress{}, err
	}
	var p domain.Progress
	for _, task := range serving {
		p.Tasks = counted(p.Tasks, task)
		under, err := e.tasks.Under(task.Number)
		if err != nil {
			return domain.Progress{}, err
		}
		for _, sub := range under {
			p.SubTasks = counted(p.SubTasks, sub)
		}
	}
	return p, nil
}

// TaskProgress is how far along a task is, counted over what hangs under it.
// A task never counts itself: two sub-tasks of which one is done is one of
// two, and nobody reads it as one of three.
func (e *Engine) TaskProgress(number int) (domain.Progress, error) {
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

// counted adds one task to a tally.
func counted(t domain.Tally, task domain.Task) domain.Tally {
	t.Total++
	if task.State == domain.Done {
		t.Done++
	}
	return t
}
