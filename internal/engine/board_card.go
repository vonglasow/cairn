package engine

// What a card says about a task without knowing how a card is drawn. A
// terminal draws these facts in cells and a window draws them in pixels; that
// one of the seven states means "not moving", and that a deadline four days
// out reads "4d", is the same on both.

import (
	"fmt"
	"slices"
	"time"

	"cairn/internal/domain"
)

// soon is how many days ahead a deadline starts being worth a card's width.
// Beyond it every task would carry one and the ones that matter would not
// stand out.
const soon = 7

// Stopped says whether a state is one of the three that mean "not moving", and
// therefore whether the task belongs in the band rather than in a column.
func Stopped(state string) bool {
	return slices.Contains(Stalled, state)
}

// Deadline says how a day is written on a card, and whether it is written at
// all. Two answers rather than one string: a day that has passed is drawn in a
// different colour, and a colour is not something a string can carry.
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

// CardDeadline is what a card says of its day: Deadline's answer, and nothing
// once the task is done. A finished task is wanted on no day any more, and a
// countdown on it — drawn late, most often — would call for an act nobody has
// left to do.
func CardDeadline(c Card, now time.Time) (text string, show, late bool) {
	if Finished(c) {
		return "", false, false
	}
	return Deadline(c.Deadline, now)
}

// startOfDay throws away the time of day. A deadline is a day, so the answer to
// "how many days" must not depend on what o'clock it is when the board is
// opened.
func startOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// Serving is the work under one objective, in the order the listing gave it.
//
// It reads the snapshot both boards already hold rather than asking a question
// of its own: every card says what it serves, and asking twice could answer
// twice differently — the same reason the focus travels with the listing that
// marks it.
func Serving(cards []Card, objective int) []Card {
	var under []Card
	for _, c := range cards {
		if slices.Contains(c.Objectives, objective) {
			under = append(under, c)
		}
	}
	return under
}

// Hanging is the work under one task, in the order the listing gave it. The
// same reading of the same snapshot as Serving, asked of a parent rather than
// an objective: a board shows what is under a thing, and which thing it is
// changes nothing about where the answer comes from.
//
// Nought is nobody: a root task hangs under nothing, and answering every root
// task to a question about task nought would be answering a question nobody
// asked.
func Hanging(cards []Card, parent int) []Card {
	if parent == 0 {
		return nil
	}
	var left, done []Card
	for _, c := range cards {
		if c.Parent != parent {
			continue
		}
		if Finished(c) {
			done = append(done, c)
			continue
		}
		left = append(left, c)
	}
	// What is left first, and what is done after it, each in the order it was
	// put in. What is left is what is looked for, and a done part mixed in
	// among them was read as one still to do — the difference was the
	// brightness of the ink, and that was not enough to tell them apart.
	return append(left, done...)
}

// Tally counts what of the parts is left and what is done.
func Tally(parts []Card) (left, done int) {
	for _, c := range parts {
		if Finished(c) {
			done++
			continue
		}
		left++
	}
	return left, done
}

// Tallied is that count as a heading says it, and nothing when there is
// nothing to count.
func Tallied(parts []Card) string {
	if len(parts) == 0 {
		return ""
	}
	left, done := Tally(parts)
	return fmt.Sprintf("%d to do · %d done", left, done)
}

// ReachesCard says whether what was typed reaches a card. The rule is the
// engine's, over what the engine said a search looks in when the card was
// read: once it lived here three times over, and the copies drifted.
func ReachesCard(c Card, query string) bool {
	return Reaches(c.Said, query)
}

// Finished says whether a card is done. One line, and it earns being here
// rather than written out at each board: `done` is a word of the domain's, and
// a board spelling it for itself is a board deciding what finished means.
func Finished(c Card) bool { return c.State == string(domain.Done) }
