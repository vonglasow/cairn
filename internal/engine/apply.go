package engine

import (
	"cairn/internal/engine/calendar"
	"cairn/internal/engine/change"
	"cairn/internal/engine/item"
	"cairn/internal/engine/label"
	"errors"
	"time"

	"cairn/internal/engine/task"
)

// Changes are the attributes an operation was given.
type Changes struct {
	Title       *string
	Description *string
	Replace     *string
	By          *string
	State       *string
	Priority    *string
	Labels      *[]string
	Blocker     *[]string
	Link        *[]string
	Because     *string
	Deadline    *string
	Estimate    *string
	Spent       *string
	Objective   *[]string
	Steps       *[]task.Step
	Finishing   bool
}

func (c Changes) named() bool {
	return c.Title != nil || c.Description != nil || c.Replace != nil || c.By != nil || c.State != nil ||
		c.Priority != nil || c.Labels != nil || c.Deadline != nil ||
		c.Estimate != nil || c.Spent != nil || c.Steps != nil || c.Objective != nil ||
		c.Blocker != nil || c.Link != nil || c.Because != nil
}

const (
	creating = true
	changing = false
)

func changedLabels(given, current []string, atCreation bool) ([]string, error) {
	change, err := label.ParseLabelChange(given)
	if err != nil {
		return nil, err
	}
	if atCreation && change.Signed() {
		return nil, errLabelSignAtCreation
	}
	return change.Apply(current)
}

func apply(t task.Task, c Changes, atCreation bool) (task.Task, error) {
	a := t.Carried()
	if err := applyCarried(carriedTo{&a.Title, &a.Description, &a.Priority, &a.Labels, &a.Deadline, &a.Estimate}, "a task", carriedChanges{
		title: c.Title, description: c.Description, replace: c.Replace, by: c.By, priority: c.Priority,
		labels: c.Labels, deadline: c.Deadline, estimate: c.Estimate,
	}, atCreation); err != nil {
		return task.Task{}, err
	}
	if c.State != nil {
		state, err := item.ParseState(*c.State)
		if err != nil {
			return task.Task{}, err
		}
		a.State = state
	}
	if c.Blocker != nil {
		blockers, err := task.ChangedBlockers(*c.Blocker, a.Blockers, atCreation)
		if err != nil {
			return task.Task{}, err
		}
		a.Blockers = blockers
	}
	if c.Link != nil {
		linked, err := task.ChangedLinks(*c.Link, a.Linked, atCreation)
		if err != nil {
			return task.Task{}, err
		}
		a.Linked = linked
	}
	if c.Because != nil {
		said, err := item.Prose(*c.Because)
		if err != nil {
			return task.Task{}, err
		}
		a.Because = said
	}
	if c.Spent != nil {
		spent, err := change.ParseSpent(*c.Spent)
		if err != nil {
			return task.Task{}, err
		}
		a.Spent = spent.Apply(a.Spent)
	}
	if c.Steps != nil {
		a.Steps = *c.Steps
	}
	if c.Objective != nil {
		served, err := change.ChangedNumbers(*c.Objective, a.Objectives, atCreation, errObjectiveRequired,
			errObjectivesBothWays, errObjectiveSignAtBirth)
		if err != nil {
			return task.Task{}, err
		}
		a.Objectives = served
	}
	return task.Made(a), nil
}

type carriedTo struct {
	title, description *string
	priority           *item.Priority
	labels             *[]string
	deadline           *time.Time
	estimate           *time.Duration
}

type carriedChanges struct {
	title, description, replace, by, priority, deadline, estimate *string
	labels                                                        *[]string
}

// What a change to a piece of a description can be turned down for.
var (
	errReplaceAlone        = errors.New("--replace and --by go together: what to replace, and what to put instead")
	errDescriptionBothWays = errors.New("the description is changed whole or in part, not both")
)

func replaced(description string, c carriedChanges) (string, error) {
	if (c.replace == nil) != (c.by == nil) {
		return "", errReplaceAlone
	}
	if c.description != nil {
		return "", errDescriptionBothWays
	}
	said, err := change.ReplaceOnce(description, *c.replace, *c.by)
	if err != nil {
		return "", err
	}
	return item.Prose(said)
}

func applyCarried(to carriedTo, of string, c carriedChanges, atCreation bool) error {
	if c.title != nil {
		title, err := item.Title(of, *c.title)
		if err != nil {
			return err
		}
		*to.title = title
	}
	if c.description != nil {
		said, err := item.Prose(*c.description)
		if err != nil {
			return err
		}
		*to.description = said
	}
	if c.replace != nil || c.by != nil {
		said, err := replaced(*to.description, c)
		if err != nil {
			return err
		}
		*to.description = said
	}
	if c.priority != nil {
		priority, err := item.ParsePriority(*c.priority)
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
		deadline, err := calendar.ParseDeadline(*c.deadline)
		if err != nil {
			return err
		}
		*to.deadline = deadline
	}
	if c.estimate != nil {
		estimate, err := calendar.ParseEstimate(*c.estimate)
		if err != nil {
			return err
		}
		*to.estimate = estimate
	}
	return nil
}

// What a change to what a task serves can be turned down for.
var (
	errObjectiveRequired    = errors.New("an objective number is expected")
	errObjectivesBothWays   = errors.New("the objectives are either replaced or changed, not both")
	errObjectiveSignAtBirth = errors.New("an objective is added or taken away only when changing a task")
)
