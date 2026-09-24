package reading

// What a card says about a task without knowing how a card is drawn. A
// terminal draws these facts in cells and a window draws them in pixels; that
// one of the seven states means "not moving", and that a deadline four days
// out reads "4d", is the same on both.

import (
	"cairn/internal/domain"
	"fmt"
	"slices"
	"strconv"
	"strings"
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
	var under []Card
	for _, c := range cards {
		if c.Parent == parent {
			under = append(under, c)
		}
	}
	return under
}

// ReachesCard says whether a search reaches a card. It is what both boards ask,
// and it answers what `task search` answers: the same facts read the way a
// person reads them, case-insensitive, wherever among them the query sits.
//
// Its twin for objectives has been here all along, and both boards ask it.
// This one lived three times instead. The engine looked in six fields,
// the terminal board in three and the window in three, and nothing caught it —
// every check this repo runs measures one package against itself, and three
// tested copies of one rule score three perfect marks. depguard can forbid a
// board from reaching the other; it cannot make either reach for this.
//
// What it does not reach is what a card carries about other things — its
// parent, the objectives it serves — nor the closed sets a listing narrows by.
// Those are the same lines the command line draws, and the test beside this
// one holds the two answers together.
func ReachesCard(c Card, query string) bool {
	if query == "" {
		return false
	}
	query = strings.ToLower(query)
	said := []string{
		strconv.Itoa(c.Number),
		c.Title,
		c.Description,
		c.Because,
		c.Estimate,
		c.Spent,
	}
	if !c.Deadline.IsZero() {
		said = append(said, domain.FormatDay(c.Deadline))
	}
	for _, field := range said {
		if field != "" && strings.Contains(strings.ToLower(field), query) {
			return true
		}
	}
	return false
}
