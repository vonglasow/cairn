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
	estimate, _ := domain.ParseEstimate(c.Estimate)
	spent, _ := domain.ParseEstimate(c.Spent)
	c.Said = task.Said(task.Made(task.Carried{
		Number: c.Number, Title: c.Title, State: domain.State(c.State), Priority: domain.Priority(c.Priority),
		Description: c.Description, Labels: c.Labels, Because: c.Because, Deadline: c.Deadline,
		Estimate: estimate, Spent: spent,
	}))
	return c
}

// SearchableGoal is an objective saying what the engine says it carries.
func SearchableGoal(g engine.Goal) engine.Goal {
	estimate, _ := domain.ParseEstimate(g.Estimate)
	g.Said = objective.Said(objective.Made(objective.Carried{
		Number: g.Number, Title: g.Title, State: objective.State(g.State), Priority: domain.Priority(g.Priority),
		Description: g.Description, Labels: g.Labels, Deadline: g.Deadline, Estimate: estimate,
	}))
	return g
}
