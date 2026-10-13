// Package objective holds the objective: what it carries, the rules that keep it valid, and the membrane anything uses to read or build one.
package objective

import (
	"errors"
	"fmt"
	"slices"
	"time"

	"cairn/internal/domain"
)

// ErrNotAnObjectiveState is returned when something that is not one of the four is offered as an objective's state.
var ErrNotAnObjectiveState = errors.New("not one of the objective states")

// State is where an objective stands.
type State string

// ObjectiveStates holds the four states, in the order they are declared above.
var ObjectiveStates = []State{Active, OnHold, Reached, Dropped}

// The four states an objective can be in.
const (
	Active  State = "active"
	OnHold  State = "on-hold"
	Reached State = "reached"
	Dropped State = "dropped"
)

// ErrObjectiveCarriesTasks is returned when an objective serving tasks is asked to be deleted.
var ErrObjectiveCarriesTasks = errors.New("the objective carries tasks")

// MayDelete says whether the objective carrying number may be deleted, given how many tasks serve it.
func MayDelete(number int, serving int) error {
	if serving > 0 {
		return fmt.Errorf("objective %d: %w", number, ErrObjectiveCarriesTasks)
	}
	return nil
}

// Objective is something worth doing, that tasks are put under.
type Objective struct {
	number      int
	title       string
	description string
	state       State
	priority    domain.Priority
	labels      []string
	deadline    time.Time
	estimate    time.Duration
}

// Told is the membrane of an objective: one method per fact it has to tell, absence included.
type Told interface {
	Numbered(int)
	Titled(string)
	Standing(State)
	Carries(domain.Priority)
	Described(string)
	NotDescribed()
	Labelled(string)
	NotLabelled()
	LabelsTold()
	Due(time.Time)
	NotDue()
	Estimated(time.Duration)
	NotEstimated()
}

// Tell says every fact this objective carries to the given membrane, absence included, in the order a card lays them out.
func (o Objective) Tell(to Told) {
	to.Numbered(o.number)
	to.Titled(o.title)
	to.Standing(o.state)
	to.Carries(o.priority)
	domain.TellOrNot(o.description, to.Described, to.NotDescribed)
	domain.TellAll(o.labels, to.Labelled, to.NotLabelled, to.LabelsTold)
	domain.TellOrNot(o.deadline, to.Due, to.NotDue)
	domain.TellOrNot(o.estimate, to.Estimated, to.NotEstimated)
}

// AsTold builds an objective from what it is told.
type AsTold struct {
	objective Objective
	labels    []string
}

// Numbered is told the objective's number.
func (a *AsTold) Numbered(n int) { a.objective.number = n }

// Titled is told the objective's title.
func (a *AsTold) Titled(s string) { a.objective.title = s }

// Standing is told where the objective stands.
func (a *AsTold) Standing(s State) { a.objective.state = s }

// Carries is told the priority the objective carries.
func (a *AsTold) Carries(p domain.Priority) { a.objective.priority = p }

// Described is told the objective's description.
func (a *AsTold) Described(s string) { a.objective.description = s }

// NotDescribed is told the objective carries no description.
func (a *AsTold) NotDescribed() { a.objective.description = "" }

// Labelled is told one more label the objective carries.
func (a *AsTold) Labelled(s string) { a.labels = append(a.labels, s) }

// NotLabelled is told the objective carries no labels at all.
func (a *AsTold) NotLabelled() { a.labels = nil }

// LabelsTold is told there are no more labels coming, and closes the set.
func (a *AsTold) LabelsTold() { a.objective.labels = a.labels }

// Due is told the objective's deadline.
func (a *AsTold) Due(t time.Time) { a.objective.deadline = t }

// NotDue is told the objective carries no deadline.
func (a *AsTold) NotDue() { a.objective.deadline = time.Time{} }

// Estimated is told the objective's estimate.
func (a *AsTold) Estimated(d time.Duration) { a.objective.estimate = d }

// NotEstimated is told the objective carries no estimate.
func (a *AsTold) NotEstimated() { a.objective.estimate = 0 }

// Objective hands back the objective AsTold was told to build.
func (a *AsTold) Objective() Objective { return a.objective }

// NewObjective builds an objective from a title, refusing one that carries no text.
func NewObjective(raw string) (Objective, error) {
	title, err := domain.Title("an objective", raw)
	if err != nil {
		return Objective{}, err
	}
	return Objective{title: title, state: Active, priority: domain.Normal}, nil
}

// ParseObjectiveState reads the name of a state an objective can be in.
func ParseObjectiveState(name string) (State, error) {
	if name == "" {
		return Active, nil
	}
	if slices.Contains(ObjectiveStates, State(name)) {
		return State(name), nil
	}
	return "", fmt.Errorf("%q: %w", name, ErrNotAnObjectiveState)
}
