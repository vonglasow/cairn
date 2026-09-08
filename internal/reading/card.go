package reading

// What a card says about a task without knowing how a card is drawn. A
// terminal draws these facts in cells and a window draws them in pixels; that
// one of the seven states means "not moving", and that a deadline four days
// out reads "4d", is the same on both.

import (
	"fmt"
	"time"
)

// soon is how many days ahead a deadline starts being worth a card's width.
// Beyond it every task would carry one and the ones that matter would not
// stand out.
const soon = 7

// Stopped says whether a state is one of the three that mean "not moving", and
// therefore whether the task belongs in the band rather than in a column.
func Stopped(state string) bool {
	return state == "blocked" || state == "watching" || state == "paused"
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

// startOfDay throws away the time of day. A deadline is a day, so the answer to
// "how many days" must not depend on what o'clock it is when the board is
// opened.
func startOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}
