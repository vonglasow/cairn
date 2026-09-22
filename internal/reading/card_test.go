package reading

import (
	"slices"
	"testing"
	"time"

	"cairn/internal/domain"
)

// today is the day these tests are read from, so "in two days" is a fact and
// not a moving target.
var today = time.Date(2026, 9, 1, 15, 30, 0, 0, time.UTC)

// on reads a day the way the domain writes one, rather than through a layout
// kept here: a second layout is a second thing to keep true.
func on(day string) time.Time {
	parsed, err := domain.ParseDeadline(day)
	if err != nil {
		panic(err)
	}
	return parsed
}

func TestStopped(t *testing.T) {
	for _, state := range []string{"blocked", "watching", "paused"} {
		if !Stopped(state) {
			t.Fatalf("%q means not moving, and was not read that way", state)
		}
	}
	for _, state := range []string{"backlog", "todo", "current", "done", ""} {
		if Stopped(state) {
			t.Fatalf("%q was read as not moving", state)
		}
	}
}

func TestDeadline(t *testing.T) {
	for _, c := range []struct {
		name       string
		due        time.Time
		text       string
		show, late bool
	}{
		{"none at all", time.Time{}, "", false, false},
		{"today", on("2026-09-01"), "0d", true, false},
		{"in two days", on("2026-09-03"), "2d", true, false},
		{"at the edge of soon", on("2026-09-08"), "7d", true, false},
		{"one day past soon", on("2026-09-09"), "", false, false},
		{"gone by three days", on("2026-08-29"), "−3d", true, true},
	} {
		t.Run(c.name, func(t *testing.T) {
			text, show, late := Deadline(c.due, today)
			if text != c.text || show != c.show || late != c.late {
				t.Fatalf("Deadline = %q %v %v, want %q %v %v", text, show, late, c.text, c.show, c.late)
			}
		})
	}
}

func TestDeadlineIgnoresTheTimeOfDay(t *testing.T) {
	// The answer to "how many days" must not depend on what o'clock it is when
	// the board is opened.
	morning := time.Date(2026, 9, 1, 0, 1, 0, 0, time.UTC)
	evening := time.Date(2026, 9, 1, 23, 59, 0, 0, time.UTC)
	first, _, _ := Deadline(on("2026-09-03"), morning)
	second, _, _ := Deadline(on("2026-09-03"), evening)
	if first != second {
		t.Fatalf("the same deadline read %q in the morning and %q at night", first, second)
	}
}

// MayLandAfter is the question a board asks before it draws where a held card
// would go. It is asked of both boards, so it lives here rather than in either.
func TestMayLandAfter_answersBeforeTheGestureIsOver(t *testing.T) {
	var (
		task   = Card{Number: 1}
		other  = Card{Number: 2}
		sub    = Card{Number: 3, Parent: 1}
		sister = Card{Number: 4, Parent: 1}
		cousin = Card{Number: 5, Parent: 2}
	)
	for _, tc := range []struct {
		what         string
		held, target Card
		want         bool
	}{
		{"a task behind a task", task, other, true},
		{"a sub-task behind its sister", sub, sister, true},
		{"a sub-task behind a task", sub, other, false},
		{"a task behind a sub-task", task, sub, false},
		{"a sub-task behind another parent's", sub, cousin, false},
		{"a card behind itself", task, task, false},
	} {
		if got := MayLandAfter(tc.held, tc.target); got != tc.want {
			t.Errorf("%s: MayLandAfter = %v, want %v", tc.what, got, tc.want)
		}
	}
}

// Serving is the work under one objective, taken from the listing both boards
// already hold. No question of its own: the snapshot carries what each task
// serves, and asking again could answer differently.
func TestServing_isTheWorkUnderOneObjective(t *testing.T) {
	cards := []Card{
		{Number: 1, Title: "Write the spec", Objectives: []int{1}},
		{Number: 2, Title: "Buy a lamp"},
		{Number: 3, Title: "Draw the band", Objectives: []int{2, 1}},
		{Number: 4, Title: "Something else", Objectives: []int{2}},
	}
	got := titlesOf(Serving(cards, 1))
	want := []string{"Write the spec", "Draw the band"}
	if !slices.Equal(got, want) {
		t.Errorf("Serving(1) = %v, want %v", got, want)
	}
	if under := Serving(cards, 9); len(under) != 0 {
		t.Errorf("Serving(9) = %v, want nothing under an objective nothing serves", titlesOf(under))
	}
}

func titlesOf(cards []Card) []string {
	var said []string
	for _, c := range cards {
		said = append(said, c.Title)
	}
	return said
}

// Hanging is the work under one task, read from the listing both boards
// already hold — the same shape as Serving, asked of a parent instead of an
// objective, because it is the same question put to a different thing.
func TestHanging_isTheWorkUnderOneTask(t *testing.T) {
	cards := []Card{
		{Number: 1, Title: "Ship it"},
		{Number: 2, Title: "Write the spec", Parent: 1},
		{Number: 3, Title: "Something else"},
		{Number: 4, Title: "Draw the band", Parent: 1},
		{Number: 5, Title: "Under another", Parent: 3},
	}
	got := titlesOf(Hanging(cards, 1))
	want := []string{"Write the spec", "Draw the band"}
	if !slices.Equal(got, want) {
		t.Errorf("Hanging(1) = %v, want %v", got, want)
	}
	if under := Hanging(cards, 2); len(under) != 0 {
		t.Errorf("Hanging(2) = %v, want nothing under a task nothing hangs under", titlesOf(under))
	}
	// Not the root tasks: they hang under nothing, and nothing is not a task.
	if under := Hanging(cards, 0); len(under) != 0 {
		t.Errorf("Hanging(0) = %v, want nothing — no task is numbered nought", titlesOf(under))
	}
}
