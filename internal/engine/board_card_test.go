package engine

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

var today = time.Date(2026, 9, 1, 15, 30, 0, 0, time.UTC)

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

func TestCardDeadline_saysNothingOnceTheTaskIsDone(t *testing.T) {
	gone := on("2026-08-29")
	if text, show, late := CardDeadline(Card{State: "done", Deadline: gone}, today); text != "" || show || late {
		t.Errorf("a done task's deadline reads %q %v %v, want nothing", text, show, late)
	}
	for _, state := range []string{"backlog", "todo", "current", "blocked"} {
		if text, show, late := CardDeadline(Card{State: state, Deadline: gone}, today); text != "−3d" || !show || !late {
			t.Errorf("a task in %s reads %q %v %v, want its day gone by three", state, text, show, late)
		}
	}
}

func TestDeadlineIgnoresTheTimeOfDay(t *testing.T) {
	morning := time.Date(2026, 9, 1, 0, 1, 0, 0, time.UTC)
	evening := time.Date(2026, 9, 1, 23, 59, 0, 0, time.UTC)
	first, _, _ := Deadline(on("2026-09-03"), morning)
	second, _, _ := Deadline(on("2026-09-03"), evening)
	if first != second {
		t.Fatalf("the same deadline read %q in the morning and %q at night", first, second)
	}
}

func TestServing_isTheWorkUnderOneObjective(t *testing.T) {
	cards := []Card{
		{Number: 1, Title: "Write the spec", Objectives: []int{1}},
		{Number: 2, Title: "Buy a lamp"},
		{Number: 3, Title: "Draw the band", Objectives: []int{2, 1}},
		{Number: 4, Title: "Something else", Objectives: []int{2}},
	}
	got := cardTitles(Serving(cards, 1))
	want := []string{"Write the spec", "Draw the band"}
	if !slices.Equal(got, want) {
		t.Errorf("Serving(1) = %v, want %v", got, want)
	}
	if under := Serving(cards, 9); len(under) != 0 {
		t.Errorf("Serving(9) = %v, want nothing under an objective nothing serves", cardTitles(under))
	}
}

func cardTitles(cards []Card) []string {
	var said []string
	for _, c := range cards {
		said = append(said, c.Title)
	}
	return said
}

func TestHanging_isTheWorkUnderOneTask(t *testing.T) {
	cards := []Card{
		{Number: 1, Title: "Ship it"},
		{Number: 2, Title: "Write the spec", Parent: 1},
		{Number: 3, Title: "Something else"},
		{Number: 4, Title: "Draw the band", Parent: 1},
		{Number: 5, Title: "Under another", Parent: 3},
	}
	got := cardTitles(Hanging(cards, 1))
	want := []string{"Write the spec", "Draw the band"}
	if !slices.Equal(got, want) {
		t.Errorf("Hanging(1) = %v, want %v", got, want)
	}
	if under := Hanging(cards, 2); len(under) != 0 {
		t.Errorf("Hanging(2) = %v, want nothing under a task nothing hangs under", cardTitles(under))
	}
	if under := Hanging(cards, 0); len(under) != 0 {
		t.Errorf("Hanging(0) = %v, want nothing — no task is numbered nought", cardTitles(under))
	}
}

func TestACardIsReachedByWhatTheEngineSaysItCarries(t *testing.T) {
	c := Card{Number: 12, Title: "Write the spec", Said: []string{"backend", "high"}}
	for _, typed := range []string{"backend", "BACKEND high"} {
		if !ReachesCard(c, typed) {
			t.Errorf("%q does not reach a card that says it", typed)
		}
	}
	if ReachesCard(c, "spec") {
		t.Error("a word the engine did not say reached the card")
	}
}

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
