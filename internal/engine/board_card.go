package engine

import (
	"fmt"
	"slices"
	"time"

	"cairn/internal/domain"
)

const soon = 7

// Stopped says whether a state is one of the three that mean "not moving", and therefore whether the task belongs in the band rather than in a column.
func Stopped(state string) bool {
	return slices.Contains(Stalled, state)
}

// Deadline says how a day is written on a card, and whether it is written at all.
func Deadline(due, now time.Time) (text string, show, late bool) {
	if due.IsZero() {
		return "", false, false
	}
	days := int(due.Sub(startOfDay(now)).Hours() / 24)
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
	return Deadline(c.Deadline, now)
}

func startOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// Serving is the work under one objective, in the order the listing gave it.
func Serving(cards []Card, objective int) []Card {
	var under []Card
	for _, c := range cards {
		if slices.Contains(c.Objectives, objective) {
			under = append(under, c)
		}
	}
	return under
}

// Hanging is the steps of a task as a board shows them: what is left first, then what is done, each keeping the order it was written in, and each named by its place in the task.
func Hanging(c Card) []Part {
	var left, done []Part
	for _, step := range c.Steps {
		if step.Done {
			done = append(done, step)
			continue
		}
		left = append(left, step)
	}
	return append(left, done...)
}

// Tally counts what of the steps is left and what is done.
func Tally(parts []Part) (left, done int) {
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
func Tallied(parts []Part) string {
	if len(parts) == 0 {
		return ""
	}
	left, done := Tally(parts)
	return fmt.Sprintf("%d to do · %d done", left, done)
}

// ReachesCard says whether what was typed reaches a card.
func ReachesCard(c Card, query string) bool {
	return Reaches(c.Said, query)
}

// Finished says whether a card is done.
func Finished(c Card) bool { return c.State == string(domain.Done) }
