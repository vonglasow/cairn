package test

import (
	"cairn/internal/domain"
	"cairn/internal/engine"
	"cairn/internal/objective"
	"cairn/internal/task"
)

// Searchable is a reading whose every card and objective says what the engine says it carries.
func Searchable(s engine.Snapshot) engine.Snapshot {
	cards := make([]engine.Card, 0, len(s.Cards))
	for _, c := range s.Cards {
		cards = append(cards, SearchableCard(c))
	}
	goals := make([]engine.Goal, 0, len(s.Goals))
	for _, g := range s.Goals {
		goals = append(goals, SearchableGoal(g))
	}
	s.Cards, s.Goals = cards, goals
	return s
}

// SearchableCard is a card saying what the engine says its task carries.
func SearchableCard(c engine.Card) engine.Card {
	var told task.AsTold
	told.Numbered(c.Number)
	told.Titled(c.Title)
	told.InColumn(domain.State(c.State))
	told.Carries(domain.Priority(c.Priority))
	told.Described(c.Description)
	for _, l := range c.Labels {
		told.Labelled(l)
	}
	told.LabelsTold()
	told.Because(c.Because)
	if !c.Deadline.IsZero() {
		told.Due(c.Deadline)
	}
	if d, err := domain.ParseEstimate(c.Estimate); err == nil {
		told.Estimated(d)
	}
	if d, err := domain.ParseEstimate(c.Spent); err == nil {
		told.Spent(d)
	}
	c.Said = task.Said(told.Task())
	return c
}

// SearchableGoal is an objective saying what the engine says it carries.
func SearchableGoal(g engine.Goal) engine.Goal {
	var told objective.AsTold
	told.Numbered(g.Number)
	told.Titled(g.Title)
	told.Standing(objective.State(g.State))
	told.Carries(domain.Priority(g.Priority))
	told.Described(g.Description)
	for _, l := range g.Labels {
		told.Labelled(l)
	}
	told.LabelsTold()
	if !g.Deadline.IsZero() {
		told.Due(g.Deadline)
	}
	if d, err := domain.ParseEstimate(g.Estimate); err == nil {
		told.Estimated(d)
	}
	g.Said = objective.Said(told.Objective())
	return g
}
