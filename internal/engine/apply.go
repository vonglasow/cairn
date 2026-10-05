package engine

import (
	"time"

	"cairn/internal/domain"
	"cairn/internal/task"
)

// Applying what a caller named to a task or an objective: the changes it was
// given, read the way each attribute reads, and the set grammar for labels.

// Changes are the attributes an operation was given. A nil field was not
// named at all; a field pointing at an empty value was named and emptied —
// which is how an attribute is taken away.
type Changes struct {
	Title       *string
	Description *string
	State       *string
	Priority    *string
	Labels      *[]string
	Blocker     *[]string
	Because     *string
	Deadline    *string
	Estimate    *string
	Spent       *string
	Parent      *int
	Objective   *[]string
}

// named says whether anything at all was given to change.
func (c Changes) named() bool {
	return c.Title != nil || c.Description != nil || c.State != nil ||
		c.Priority != nil || c.Labels != nil || c.Deadline != nil ||
		c.Estimate != nil || c.Spent != nil || c.Parent != nil || c.Objective != nil ||
		c.Blocker != nil || c.Because != nil
}

// Whether a task is being created or changed. Labels are the only attribute
// that cares: a sign says how an existing set moves, and there is no set to
// move at creation.
const (
	creating = true
	changing = false
)

// changedLabels works out the set that a change leaves behind. A sign says how
// an existing set moves, and there is no set to move at creation, so one given
// there is refused rather than read as a label named "+urgent".
//
// Shared by tasks and objectives: both carry a set and both move it the same
// way, and this is the one part of applying a change that is more than a call
// to a parser.
func changedLabels(given, current []string, atCreation bool) ([]string, error) {
	change, err := domain.ParseLabelChange(given)
	if err != nil {
		return nil, err
	}
	if atCreation && change.Adjusts() {
		return nil, ErrLabelSignAtCreation
	}
	return change.Apply(current)
}

// apply lays the changes over a task, refusing any value it cannot make sense
// of. A field that was not named is left exactly as it was.
func apply(t task.Task, c Changes, atCreation bool) (task.Task, error) {
	current := factsOf(t)
	a := &task.AsTold{}
	t.Tell(a)

	if err := applyCarried(a, "a task", carriedChanges{
		title: c.Title, description: c.Description, priority: c.Priority,
		labels: c.Labels, deadline: c.Deadline, estimate: c.Estimate,
	}, current.labels, atCreation); err != nil {
		return task.Task{}, err
	}
	if c.State != nil {
		state, err := domain.ParseState(*c.State)
		if err != nil {
			return task.Task{}, err
		}
		a.InColumn(state)
	}
	if c.Blocker != nil {
		blockers, err := changedBlockers(*c.Blocker, current.blockers, atCreation)
		if err != nil {
			return task.Task{}, err
		}
		a.NotBlocked()
		for _, blocker := range blockers {
			a.BlockedBy(blocker)
		}
		a.BlockersTold()
	}
	if c.Because != nil {
		a.Because(*c.Because)
	}
	if c.Spent != nil {
		spent, err := domain.ParseSpent(*c.Spent)
		if err != nil {
			return task.Task{}, err
		}
		a.Spent(spent.Apply(current.timeSpent))
	}
	if c.Parent != nil {
		a.Under(*c.Parent)
	}
	if c.Objective != nil {
		served, err := changedNumbers(*c.Objective, current.objectives, atCreation, ErrObjectiveRequired,
			ErrObjectivesBothWays, ErrObjectiveSignAtBirth)
		if err != nil {
			return task.Task{}, err
		}
		a.ServesNothing()
		for _, one := range served {
			a.Serves(one)
		}
		a.ObjectivesTold()
	}
	return a.Task(), nil
}

// carried is what both a task and an objective are told about what they carry
// alike — the half of AsTold the two share.
type carried interface {
	Titled(string)
	Described(string)
	Carries(domain.Priority)
	NotLabelled()
	Labelled(string)
	LabelsTold()
	Due(time.Time)
	Estimated(time.Duration)
}

// carriedChanges are the changes to what a task and an objective carry alike.
type carriedChanges struct {
	title, description, priority, deadline, estimate *string
	labels                                           *[]string
}

// applyCarried tells a task or an objective what changed among what both
// carry, read the same way for both. It was written once for each, and the two
// copies differed only in what they were building.
func applyCarried(a carried, of string, c carriedChanges, labels []string, atCreation bool) error {
	if c.title != nil {
		title, err := domain.Title(of, *c.title)
		if err != nil {
			return err
		}
		a.Titled(title)
	}
	if c.description != nil {
		a.Described(*c.description)
	}
	if c.priority != nil {
		priority, err := domain.ParsePriority(*c.priority)
		if err != nil {
			return err
		}
		a.Carries(priority)
	}
	if c.labels != nil {
		changed, err := changedLabels(*c.labels, labels, atCreation)
		if err != nil {
			return err
		}
		a.NotLabelled()
		for _, label := range changed {
			a.Labelled(label)
		}
		a.LabelsTold()
	}
	if c.deadline != nil {
		deadline, err := domain.ParseDeadline(*c.deadline)
		if err != nil {
			return err
		}
		a.Due(deadline)
	}
	if c.estimate != nil {
		estimate, err := domain.ParseEstimate(*c.estimate)
		if err != nil {
			return err
		}
		a.Estimated(estimate)
	}
	return nil
}
