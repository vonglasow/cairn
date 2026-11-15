package test

import (
	"cairn/internal/board"
	"cairn/internal/engine/calendar"
	"cairn/internal/engine/item"
	"cairn/internal/engine/objective"
	"cairn/internal/engine/task"
)

// Searchable is a reading whose every card and objective says what the engine says it carries.
func Searchable(s board.Snapshot) board.Snapshot {
	cards := make([]board.Card, 0, len(s.Cards))
	for _, c := range s.Cards {
		cards = append(cards, SearchableCard(c))
	}
	goals := make([]board.Goal, 0, len(s.Goals))
	for _, g := range s.Goals {
		goals = append(goals, SearchableGoal(g))
	}
	s.Cards, s.Goals = cards, goals
	return s
}

// SearchableCard is a card saying what the engine says its task carries.
func SearchableCard(c board.Card) board.Card {
	estimate, _ := calendar.ParseEstimate(c.Estimate)
	spent, _ := calendar.ParseEstimate(c.Spent)
	c.Said = task.Said(task.Made(task.Carried{
		Number: c.Number, Title: c.Title, State: item.State(c.State), Priority: item.Priority(c.Priority),
		Description: c.Description, Labels: c.Labels, Because: c.Because, Deadline: c.Deadline,
		Estimate: estimate, Spent: spent,
	}))
	return c
}

// SearchableGoal is an objective saying what the engine says it carries.
func SearchableGoal(g board.Goal) board.Goal {
	estimate, _ := calendar.ParseEstimate(g.Estimate)
	g.Said = objective.Said(objective.Made(objective.Carried{
		Number: g.Number, Title: g.Title, State: objective.State(g.State), Priority: item.Priority(g.Priority),
		Description: g.Description, Labels: g.Labels, Deadline: g.Deadline, Estimate: estimate,
	}))
	return g
}
