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

// Carried is everything an objective carries, as plain facts.
type Carried struct {
	Number             int
	Title, Description string
	State              State
	Priority           domain.Priority
	Labels             []string
	Deadline           time.Time
	Estimate           time.Duration
}

// Carried hands over what the objective carries.
func (o Objective) Carried() Carried {
	return Carried{
		Number: o.number, Title: o.title, Description: o.description, State: o.state,
		Priority: o.priority, Labels: slices.Clone(o.labels), Deadline: o.deadline, Estimate: o.estimate,
	}
}

// Made is the objective that carries what is given.
func Made(c Carried) Objective {
	return Objective{
		number: c.Number, title: c.Title, description: c.Description, state: c.State,
		priority: c.Priority, labels: c.Labels, deadline: c.Deadline, estimate: c.Estimate,
	}
}

// AsTold builds an objective one fact at a time.
type AsTold struct{ c Carried }

// From starts a builder on what a objective already carries.
func From(o Objective) AsTold { return AsTold{c: o.Carried()} }

// Numbered is told the objective's number.
func (a *AsTold) Numbered(n int) { a.c.Number = n }

// Titled is told the objective's title.
func (a *AsTold) Titled(s string) { a.c.Title = s }

// Standing is told where the objective stands.
func (a *AsTold) Standing(s State) { a.c.State = s }

// Carries is told the priority the objective carries.
func (a *AsTold) Carries(p domain.Priority) { a.c.Priority = p }

// Described is told the objective's description.
func (a *AsTold) Described(s string) { a.c.Description = s }

// Labelled is told one more label the objective carries.
func (a *AsTold) Labelled(s string) { a.c.Labels = append(a.c.Labels, s) }

// Due is told the objective's deadline.
func (a *AsTold) Due(t time.Time) { a.c.Deadline = t }

// Estimated is told the objective's estimate.
func (a *AsTold) Estimated(d time.Duration) { a.c.Estimate = d }

// Objective hands back the objective AsTold was told to build.
func (a *AsTold) Objective() Objective { return Made(a.c) }

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
