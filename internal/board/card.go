package board

import (
	"cairn/internal/engine"
	"cairn/internal/engine/calendar"
	"cairn/internal/engine/item"
	"fmt"
	"slices"
	"time"
)

const soon = 7

// Stopped says whether a state is one of the three that mean "not moving", and therefore whether the task belongs in the band rather than in a column.
func Stopped(state string) bool {
	return slices.Contains(Stalled, state)
}

// deadline says how a day is written on a card, and whether it is written at all.
func deadline(due, now time.Time) (text string, show, late bool) {
	if due.IsZero() {
		return "", false, false
	}
	days := int(due.Sub(calendar.StartOfDay(now)).Hours() / 24)
	if days < 0 {
		return fmt.Sprintf("−%dd", -days), true, true
	}
	if days > soon {
		return "", false, false
	}
	return fmt.Sprintf("%dd", days), true, false
}

// CardDeadline is what a card says of its day: Deadline's answer, and nothing once the task is done.
func CardDeadline(c Card, now time.Time) (text string, show, late bool) {
	if Finished(c) {
		return "", false, false
	}
	return deadline(c.Deadline, now)
}

// Serving is the work under any of the objectives named, in the order the listing gave it.
func Serving(cards []Card, objectives ...int) []Card {
	var under []Card
	for _, c := range cards {
		if slices.ContainsFunc(c.Objectives, func(o int) bool { return slices.Contains(objectives, o) }) {
			under = append(under, c)
		}
	}
	return under
}

// Hanging is the steps of a task as a board shows them: what is left first, then what is done, each keeping the order it was written in, and each named by its place in the task.
func Hanging(c Card) []engine.Part {
	var left, done []engine.Part
	for _, step := range c.Steps {
		if step.Done {
			done = append(done, step)
			continue
		}
		left = append(left, step)
	}
	return append(left, done...)
}

// tally counts what of the steps is left and what is done.
func tally(parts []engine.Part) (left, done int) {
	for _, p := range parts {
		if p.Done {
			done++
			continue
		}
		left++
	}
	return left, done
}

// Tallied is that count as a heading says it, and nothing when there is nothing to count.
func Tallied(parts []engine.Part) string {
	if len(parts) == 0 {
		return ""
	}
	left, done := tally(parts)
	return fmt.Sprintf("%d left · %d done", left, done)
}

// ReachesCard says whether what was typed reaches a card.
func ReachesCard(c Card, query string) bool {
	return engine.Reaches(c.Said, query)
}

// Finished says whether a card is done.
func Finished(c Card) bool { return c.State == string(item.Done) }

// Cost is what a card says it cost against what it was reckoned to cost, and whether that is over, when the time is shown; nothing when it is hidden or off.
func Cost(p engine.Pages, c Card) (string, bool) {
	if !p.Shows(engine.PageTime) {
		return "", false
	}
	return c.Cost()
}
