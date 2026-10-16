package engine

import (
	"time"

	"cairn/internal/domain"
	"cairn/internal/task"
)

// Changes are the attributes an operation was given.
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
	Objective   *[]string
	Steps       *[]task.Step
	Finishing   bool
}

func (c Changes) named() bool {
	return c.Title != nil || c.Description != nil || c.State != nil ||
		c.Priority != nil || c.Labels != nil || c.Deadline != nil ||
		c.Estimate != nil || c.Spent != nil || c.Steps != nil || c.Objective != nil ||
		c.Blocker != nil || c.Because != nil
}

const (
	creating = true
	changing = false
)

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

func apply(t task.Task, c Changes, atCreation bool) (task.Task, error) {
	a := t.Carried()
	if err := applyCarried(carriedTo{&a.Title, &a.Description, &a.Priority, &a.Labels, &a.Deadline, &a.Estimate}, "a task", carriedChanges{
		title: c.Title, description: c.Description, priority: c.Priority,
		labels: c.Labels, deadline: c.Deadline, estimate: c.Estimate,
	}, atCreation); err != nil {
		return task.Task{}, err
	}
	if c.State != nil {
		state, err := domain.ParseState(*c.State)
		if err != nil {
			return task.Task{}, err
		}
		a.State = state
	}
	if c.Blocker != nil {
		blockers, err := changedBlockers(*c.Blocker, a.Blockers, atCreation)
		if err != nil {
			return task.Task{}, err
		}
		a.Blockers = blockers
	}
	if c.Because != nil {
		said, err := domain.Prose(*c.Because)
		if err != nil {
			return task.Task{}, err
		}
		a.Because = said
	}
	if c.Spent != nil {
		spent, err := domain.ParseSpent(*c.Spent)
		if err != nil {
			return task.Task{}, err
		}
		a.Spent = spent.Apply(a.Spent)
	}
	if c.Steps != nil {
		a.Steps = *c.Steps
	}
	if c.Objective != nil {
		served, err := changedNumbers(*c.Objective, a.Objectives, atCreation, ErrObjectiveRequired,
			ErrObjectivesBothWays, ErrObjectiveSignAtBirth)
		if err != nil {
			return task.Task{}, err
		}
		a.Objectives = served
	}
	return task.Made(a), nil
}

type carriedTo struct {
	title, description *string
	priority           *domain.Priority
	labels             *[]string
	deadline           *time.Time
	estimate           *time.Duration
}

type carriedChanges struct {
	title, description, priority, deadline, estimate *string
	labels                                           *[]string
}

func applyCarried(to carriedTo, of string, c carriedChanges, atCreation bool) error {
	if c.title != nil {
		title, err := domain.Title(of, *c.title)
		if err != nil {
			return err
		}
		*to.title = title
	}
	if c.description != nil {
		said, err := domain.Prose(*c.description)
		if err != nil {
			return err
		}
		*to.description = said
	}
	if c.priority != nil {
		priority, err := domain.ParsePriority(*c.priority)
		if err != nil {
			return err
		}
		*to.priority = priority
	}
	if c.labels != nil {
		changed, err := changedLabels(*c.labels, *to.labels, atCreation)
		if err != nil {
			return err
		}
		*to.labels = changed
	}
	if c.deadline != nil {
		deadline, err := domain.ParseDeadline(*c.deadline)
		if err != nil {
			return err
		}
		*to.deadline = deadline
	}
	if c.estimate != nil {
		estimate, err := domain.ParseEstimate(*c.estimate)
		if err != nil {
			return err
		}
		*to.estimate = estimate
	}
	return nil
}
