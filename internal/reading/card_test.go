package reading

import (
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"

	"cairn/internal/trace"

	"cairn/internal/task"

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

// ReachesCard is what a search reaches, and it must reach exactly what the command
// line reaches. This pins the two together: the same facts, told to the domain
// and drawn as a card, have to answer the same of every query.
//
// It exists because they had drifted. The engine looked in six fields and each
// board looked in three, and nothing noticed — every check this repo runs
// measures one package against itself, and three tested copies of one rule
// score three perfect marks.
func TestReaches_answersWhatTheCommandLineAnswers(t *testing.T) {
	spent := 90 * time.Minute
	due := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	// Built through the membrane a store builds one through, so this compares
	// the real rule against the real rule and not against a fixture.
	var told task.AsTold
	told.Numbered(228)
	told.Titled("Write the spec")
	told.Described("two columns")
	told.Because("the client's answer")
	told.Due(due)
	told.Estimated(spent)
	told.Spent(spent)
	built := told.Task()
	card := Card{
		Number: 228, Title: "Write the spec", Description: "two columns",
		Because: "the client's answer", Deadline: due,
		Estimate: domain.FormatDuration(spent), Spent: domain.FormatDuration(spent),
	}
	for _, q := range []string{
		"spec", "SPEC", "columns", "client", "228", "22", "1h30m", "2026-09",
		"lamp", "9", "", "todo", "high",
	} {
		want := task.Matching(built, []string{q})
		if got := ReachesCard(card, q); got != want {
			t.Errorf("ReachesCard(%q) = %v and the command line answers %v", q, got, want)
		}
	}
}

// Whether a part of the work is done. Both boards ask it and the answer is one
// word of the domain's, so it is not each board's to spell.
func TestFinished_isWhatDoneMeansOnACard(t *testing.T) {
	for _, tc := range []struct {
		state string
		want  bool
	}{
		{"done", true},
		{"todo", false},
		{"current", false},
		{"", false},
	} {
		if got := Finished(Card{State: tc.state}); got != tc.want {
			t.Errorf("Finished(%q) = %v, want %v", tc.state, got, tc.want)
		}
	}
}

// The boards searched six fields with one word; the engine searches every fact
// a task carries, with as many words as were typed, all of them. Two meanings
// for one word: searching by label was impossible on either board.
func TestACardIsSearchedTheWayTheEngineSearches(t *testing.T) {
	c := Card{
		Number: 12, Title: "Write the spec", State: "current", Priority: "high",
		Labels: []string{"backend", "sprint-172"},
	}
	for _, word := range []string{"backend", "sprint-172", "current", "high"} {
		if !ReachesCard(c, word) {
			t.Errorf("a search for %q does not reach the card", word)
		}
	}
	if !ReachesCard(c, "spec backend") {
		t.Error("two words the card carries do not reach it")
	}
	if ReachesCard(c, "spec frontend") {
		t.Error("a word the card does not carry still reached it")
	}
}

// What hangs under a task comes what is left first and what is done after, in
// the order it was put in inside each: what is left is what is looked for, and
// a done part mixed in among them is a part read as one still to do.
func TestWhatHangsUnderATaskComesWhatIsLeftFirst(t *testing.T) {
	cards := []Card{
		{Number: 1, Title: "Ship it", State: "todo"},
		{Number: 2, Title: "A", State: "done", Parent: 1},
		{Number: 3, Title: "B", State: "todo", Parent: 1},
		{Number: 4, Title: "C", State: "done", Parent: 1},
		{Number: 5, Title: "D", State: "todo", Parent: 1},
	}
	var got []int
	for _, c := range Hanging(cards, 1) {
		got = append(got, c.Number)
	}
	if fmt.Sprint(got) != "[3 5 2 4]" {
		t.Fatalf("what hangs under it comes as %v, want what is left first: [3 5 2 4]", got)
	}
	if left, done := Tally(Hanging(cards, 1)); left != 2 || done != 2 {
		t.Fatalf("tallied %d left and %d done, want 2 and 2", left, done)
	}
	if said := Tallied(Hanging(cards, 1)); said != "2 to do · 2 done" {
		t.Fatalf("the tally reads %q", said)
	}
	if said := Tallied(nil); said != "" {
		t.Fatalf("nothing under it reads %q", said)
	}
}

// What was kept before cairn refused escapes reaches a board shown, never
// obeyed: the terminal board draws these strings straight to the screen.
func TestCardOf_showsWhatATerminalWouldObey(t *testing.T) {
	var told task.AsTold
	told.Numbered(1)
	told.Titled("look \x1b[8m here")
	told.Described("bell\x07")
	told.Because("\u202Eevil")
	told.Labelled("odd\x1b")
	told.LabelsTold()
	card, err := cardOf(told.Task(), nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	for what, said := range map[string]string{"title": card.Title, "description": card.Description, "because": card.Because, "label": card.Labels[0]} {
		if strings.ContainsAny(said, "\x1b\x07\u202E") {
			t.Errorf("the %s reaches the board as %q", what, said)
		}
	}
}

// A line of history read before the refusals is shown the same way.
func TestSaidOf_showsWhatATerminalWouldObey(t *testing.T) {
	var a trace.AsTold
	a.On(1)
	a.Did(trace.Changed)
	a.Changing("description")
	a.Was("old\x1b]52;c;aGk=\x07")
	a.Became("new")
	a.ByHand()
	a.At(time.Now())
	if line := SaidOf(a.Act()).Line(); strings.ContainsAny(line, "\x1b\x07") {
		t.Errorf("the history line reaches the board as %q", line)
	}
}
