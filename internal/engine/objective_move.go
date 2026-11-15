package engine

import (
	"fmt"

	"cairn/internal/engine/objective"
)

// ObjectiveMove changes where an objective sits: its place among its siblings, and the parent it sits under, in one act.
var ObjectiveMove = Operation{
	Thing: "objective",
	Verb:  "move",
	Short: "Move an objective in the order, or under another parent",
	Does:  "Move an objective to another place among the ones it sits with, or under another parent and a place there, in one act; a parent moves with its sub-objectives. Behind another objective, it goes under that one's parent. Under a parent it would sit three levels deep, or with a task counted twice there, it is refused and nothing moves.",
	Example: `cairn objective move 3 --first
cairn objective move 3 --after 1
cairn objective move 10 --parent 1 --last
cairn objective move 13 --parent none`,
	Level:   write,
	Needs:   aStore,
	Again:   ChangesNothingMore,
	Answers: objectivesListed{},
	Takes: []Argument{
		{Name: "number", Kind: Number, Positional: true, Required: true, Of: "an objective", Means: "the number of the objective"},
		{Name: "first", Kind: Bool, Means: "put it at the head of the ones it sits with"},
		{Name: "after", Kind: Number, Of: "an objective", Means: "put it directly behind this objective, under this objective's parent"},
		{Name: "last", Kind: Bool, Means: "put it behind all the ones it sits with"},
		{Name: "parent", Kind: Text, Of: "an objective", Means: "the objective to sit under, none to stand on its own; last there unless a place is named"},
	},
	Run: func(on workedOn, in Input) (Answer, error) {
		n, _ := in.whole("number")
		if err := on.Engine.moveObjective(n, in); err != nil {
			return Answer{}, err
		}
		return objectiveList.Run(on, Input{})
	},
}

func (e *Engine) moveObjective(n int, in Input) error {
	was, err := e.objectives.FindObjective(n)
	if err != nil {
		return err
	}
	parent := was.Carried().Parent
	if named := in.change("parent"); named != nil {
		if parent, err = parentNamed(*named); err != nil {
			return err
		}
	}
	at := -1
	if in.yes("first") {
		at = 0
	}
	if after, given := in.whole("after"); given {
		if after == n {
			return fmt.Errorf("objective %d: %w", n, objective.ErrObjectiveNowhereToGo)
		}
		behind, err := e.objectives.FindObjective(after)
		if err != nil {
			return err
		}
		its := behind.Carried().Parent
		if in.change("parent") != nil && its != parent {
			return fmt.Errorf("objective %d behind %d: %w", n, after, objective.ErrObjectiveParentAndAfter)
		}
		parent, at = its, after
	}
	if parent != was.Carried().Parent {
		c := was.Carried()
		c.Parent = parent
		if err := e.mayHang(objective.Made(c)); err != nil {
			return err
		}
	}
	return e.objectives.PlaceObjective(n, parent, at)
}
