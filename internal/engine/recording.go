package engine

// What the engine writes down. Every act on a task already passes through here,
// which is what makes one place enough to record them all — a surface records
// nothing, and could not, since it never changes anything itself.

import (
	"fmt"
	"strings"
	"time"

	"cairn/internal/domain"
	"cairn/internal/task"
	"cairn/internal/trace"
)

// history is what the engine needs from whatever keeps what happened. Declared
// here, by the consumer, for the same reason tasks and places are.
type history interface {
	Record(trace.Act) error
	ActsOn(int) ([]trace.Act, error)
	ActsSince(time.Time) ([]trace.Act, error)
}

// theAttributes are the twelve a task carries, in the order a card lays them out,
// paired with how to read each off a task. Named once: recording a change and
// recording what a task started as both walk this list, and two lists would
// drift the day an attribute is added.
var theAttributes = []struct {
	named string
	value func(taskFacts) string
}{
	{"title", func(f taskFacts) string { return f.title }},
	{"description", func(f taskFacts) string { return f.description }},
	{"column", func(f taskFacts) string { return string(f.state) }},
	{"priority", func(f taskFacts) string { return string(f.priority) }},
	{"labels", func(f taskFacts) string { return strings.Join(f.labels, ", ") }},
	{"blockers", func(f taskFacts) string { return numbersOrNothing(f.blockers) }},
	{"because", func(f taskFacts) string { return f.because }},
	{"deadline", func(f taskFacts) string { return dayOrNothing(f.deadline) }},
	{"estimate", func(f taskFacts) string { return spanOrNothing(f.estimate) }},
	{"time spent", func(f taskFacts) string { return spanOrNothing(f.timeSpent) }},
	{"parent", func(f taskFacts) string { return numberOrNothing(f.parent) }},
	{"objective", func(f taskFacts) string { return numberOrNothing(f.objective) }},
}

// History hands back everything done to a task, oldest first, and refuses a
// task nothing was ever written down about.
//
// Nothing written down means nothing was ever there: a task always carries at
// least its creation. And it cannot be mistaken for a task since deleted, since
// a deletion takes nothing away from a history — which is what lets this be the
// signal rather than asking whether the task is still around.
func (e *Engine) History(number int) ([]trace.Act, error) {
	acts, err := e.history.ActsOn(number)
	if err != nil {
		return nil, err
	}
	if len(acts) == 0 {
		return nil, fmt.Errorf("task %d: %w", number, ErrNoSuchTask)
	}
	return acts, nil
}

// Since hands back everything done from a day onward, across every task.
func (e *Engine) Since(day time.Time) ([]trace.Act, error) {
	return e.history.ActsSince(day)
}

// recordCreated writes down a task coming into being, and what it started as:
// the act itself, then one act per attribute the task carries.
//
// One act per attribute and not one act carrying them all. Crammed into a
// single field they cannot be read back apart — measured: a description accepts
// tabs and newlines, so no separator is safe, and `title=Écrire la spéc, enfin`
// splits into a pair and a word. Each value alone in its own act is exact
// without a format to invent, and without one to get wrong.
//
// The values are written down rather than left to be inferred from the defaults
// of the day: the day a default changes, every older history would replay wrong
// and do it in silence.
func (e *Engine) recordCreated(t task.Task) error {
	f := factsOf(t)
	if err := e.record(f.number, trace.Created, "", "", ""); err != nil {
		return err
	}
	for _, a := range theAttributes {
		value := a.value(f)
		if value == "" {
			continue
		}
		if err := e.record(f.number, trace.Created, a.named, "", value); err != nil {
			return err
		}
	}
	return nil
}

// recordChanges writes down one act per attribute that actually took another
// value. An attribute named in an edit but given what it already held is not a
// change, and recording it would fill a history with acts nobody performed.
func (e *Engine) recordChanges(was, became task.Task) error {
	before, after := factsOf(was), factsOf(became)
	for _, a := range theAttributes {
		from, to := a.value(before), a.value(after)
		if from == to {
			continue
		}
		if err := e.record(after.number, trace.Changed, a.named, from, to); err != nil {
			return err
		}
	}
	return nil
}

// record writes one act down, stamped now and with the surface that asked.
func (e *Engine) record(number int, did, field, was, became string) error {
	var a trace.AsTold
	a.On(number)
	a.Did(did)
	trace.TellOrNot(field, a.Changing, a.NothingNamed)
	trace.TellOrNot(was, a.Was, a.WasNothing)
	trace.TellOrNot(became, a.Became, a.BecameNothing)
	trace.TellOrNot(e.through, a.Through, a.ByHand)
	a.At(time.Now())
	return e.history.Record(a.Act())
}

// dayOrNothing writes a deadline, or nothing when there is none.
func dayOrNothing(day time.Time) string {
	if day.IsZero() {
		return ""
	}
	return domain.FormatDay(day)
}

// spanOrNothing writes a length of time, or nothing when there is none.
func spanOrNothing(span time.Duration) string {
	if span == 0 {
		return ""
	}
	return domain.FormatDuration(span)
}

// numberOrNothing writes the number a task points at, or nothing when it points
// at none.
func numberOrNothing(number int) string {
	if number == 0 {
		return ""
	}
	return fmt.Sprint(number)
}

// numbersOrNothing reads a set of task numbers the way a card writes them, and
// nothing at all when the set is empty — so a history says a blocker went from
// nothing to #1 rather than from an empty string to one.
func numbersOrNothing(numbers []int) string {
	if len(numbers) == 0 {
		return ""
	}
	written := make([]string, 0, len(numbers))
	for _, one := range numbers {
		written = append(written, fmt.Sprintf("#%d", one))
	}
	return strings.Join(written, ", ")
}
