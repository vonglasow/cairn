// Package objective holds the objective: what it carries, the rules that
// keep it valid, and the membrane anything uses to read or build one. It
// knows only domain, for the attributes an objective carries and the title
// rule it shares with a task.
package objective

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"cairn/internal/domain"
)

// ErrNotAnObjectiveState is returned when something that is not one of the
// four is offered as an objective's state. It is kept apart from
// domain.ErrNotAState on purpose: both are called a state, but a task's is a
// column it sits in and an objective's is not, so the two sets never overlap.
// A parser shared between them would accept "todo" for an objective without a
// word.
var ErrNotAnObjectiveState = errors.New("not one of the objective states")

// State is where an objective stands. There are four, and nothing else is
// one. It is a type of its own rather than a domain.State: they share the
// word because that is what both are called, but a task's state is a column
// it gets moved between and an objective's is not, so neither set of values
// is ever valid for the other.
type State string

// ObjectiveStates holds the four states, in the order they are declared
// above. It is what ParseObjectiveState consults, so the two cannot say
// different things.
var ObjectiveStates = []State{Active, OnHold, Reached, Dropped}

// The four states an objective can be in. One with none stated is active.
const (
	Active  State = "active"
	OnHold  State = "on-hold"
	Reached State = "reached"
	Dropped State = "dropped"
)

// ErrObjectiveCarriesTasks is returned when an objective serving tasks is
// asked to be deleted. What serves it would be left pointing at a number
// nothing carries, and a deletion that takes more than the number typed is a
// deletion nobody asked for.
var ErrObjectiveCarriesTasks = errors.New("the objective carries tasks")

// MayDelete says whether the objective carrying number may be deleted, given
// how many tasks serve it.
func MayDelete(number int, serving int) error {
	if serving > 0 {
		return fmt.Errorf("objective %d: %w", number, ErrObjectiveCarriesTasks)
	}
	return nil
}

// Objective is something worth doing, that tasks are put under. It carries a
// title and a number, and shares every other attribute's type with a task —
// but not its struct: the six fields they have in common say the same thing
// about two different things, and folding them together would make every
// future field of one a field of the other.
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

// Told is the membrane of an objective: one method per fact it has to tell,
// absence included. Whoever wants to know something about an objective
// implements this and is told it; nothing reads an objective's insides, and
// an objective knows none of them.
//
// It carries thirteen methods, not the task's nineteen: an objective has no
// parent, no objective it serves, and no time spent, so there is nothing to
// tell about any of the three.
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

// Tell says every fact this objective carries to the given membrane, absence
// included, in the order a card lays them out.
func (o Objective) Tell(to Told) {
	to.Numbered(o.number)
	to.Titled(o.title)
	to.Standing(o.state)
	to.Carries(o.priority)
	if o.description == "" {
		to.NotDescribed()
	} else {
		to.Described(o.description)
	}
	if len(o.labels) == 0 {
		to.NotLabelled()
	} else {
		for _, label := range o.labels {
			to.Labelled(label)
		}
	}
	to.LabelsTold()
	if o.deadline.IsZero() {
		to.NotDue()
	} else {
		to.Due(o.deadline)
	}
	if o.estimate == 0 {
		to.NotEstimated()
	} else {
		to.Estimated(o.estimate)
	}
}

// AsTold builds an objective from what it is told. It is the same membrane
// the other way round: whoever has an objective's facts — a row out of the
// store — tells them, and is handed the objective they make. Nothing has to
// reach inside an objective to put one together, which is what lets an
// objective keep its insides to itself.
//
// Labels are trusted rather than re-parsed: whoever tells AsTold a label is
// retelling one an objective already carried, made lower case, deduplicated
// and sorted once already on the way into the store.
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

// NewObjective builds an objective from a title, refusing one that carries no
// text. What surrounds the title is not part of it; what sits inside it is.
func NewObjective(title string) (Objective, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return Objective{}, fmt.Errorf("an objective %w", domain.ErrTitleRequired)
	}
	if err := domain.OneLine(title); err != nil {
		return Objective{}, err
	}
	return Objective{title: title, state: Active, priority: domain.Normal}, nil
}

// ParseObjectiveState reads the name of a state an objective can be in.
// Nothing given returns active, which is where an objective starts and what
// clearing the state returns to.
func ParseObjectiveState(name string) (State, error) {
	if name == "" {
		return Active, nil
	}
	if slices.Contains(ObjectiveStates, State(name)) {
		return State(name), nil
	}
	return "", fmt.Errorf("%q: %w", name, ErrNotAnObjectiveState)
}
