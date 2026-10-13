package engine

import (
	"maps"
	"slices"
	"testing"

	"cairn/internal/domain"
)

func column(state string, numbers ...int) []Card {
	cards := make([]Card, 0, len(numbers))
	for _, n := range numbers {
		cards = append(cards, Card{Number: n, State: state})
	}
	return cards
}

func TestStep_movesBehindWhatWillBeBeforeIt(t *testing.T) {
	here := column("todo", 1, 2, 3, 4)
	for _, tc := range []struct {
		what   string
		number int
		by     int
		want   map[string]string
	}{
		{"down from the head", 1, 1, map[string]string{"after": "2"}},
		{"down from the middle", 2, 1, map[string]string{"after": "3"}},
		{"up from second", 2, -1, map[string]string{"first": "true"}},
		{"up from third", 3, -1, map[string]string{"after": "1"}},
		{"up from last", 4, -1, map[string]string{"after": "2"}},
	} {
		got, ok := Step(here, tc.number, tc.by)
		if !ok || !maps.Equal(got, tc.want) {
			t.Errorf("%s: Step = %v, %v, want %v", tc.what, got, ok, tc.want)
		}
	}
	for _, tc := range []struct {
		what       string
		number, by int
	}{
		{"up from the head", 1, -1},
		{"down from the last", 4, 1},
		{"a card that is not there", 9, 1},
		{"nowhere at all", 2, 0},
	} {
		if got, ok := Step(here, tc.number, tc.by); ok {
			t.Errorf("%s: Step = %v, want nowhere to go", tc.what, got)
		}
	}
}

func TestBeside_isTheColumnOneOver(t *testing.T) {
	for _, tc := range []struct {
		what   string
		column int
		inBand bool
		by     int
		want   string
		ok     bool
	}{
		{"right of the backlog", 0, false, 1, "todo", true},
		{"left of done", 3, false, -1, "current", true},
		{"left of the backlog", 0, false, -1, "", false},
		{"left of todo", 1, false, -1, "backlog", true},
		{"not over at all", 1, false, 0, "", false},
		{"not over at all, in the band", 2, true, 0, "", false},
		{"right of done", 3, false, 1, "", false},
		{"left of the band", 2, true, -1, "backlog", true},
		{"right of the band", 2, true, 1, "", false},
	} {
		got, ok := Beside(tc.column, tc.inBand, tc.by)
		if got != tc.want || ok != tc.ok {
			t.Errorf("%s: Beside = %q, %v, want %q, %v", tc.what, got, ok, tc.want, tc.ok)
		}
	}
}

func TestWhere_findsACardInAColumnOrInTheBand(t *testing.T) {
	cards := append(column("todo", 1, 2), Card{Number: 3, State: "blocked"}, Card{Number: 4, State: "paused"})
	if column, inBand, at, ok := Where(cards, false, 2); !ok || column != 1 || inBand || at != 1 {
		t.Errorf("task 2 is at %d, band %v, place %d, %v", column, inBand, at, ok)
	}
	if _, inBand, at, ok := Where(cards, false, 4); !ok || !inBand || at != 1 {
		t.Errorf("task 4 is in the band %v at %d, %v", inBand, at, ok)
	}
	if _, _, _, ok := Where(cards, false, 9); ok {
		t.Error("a card that is not there was found")
	}
}

func TestTheBoardIsLaidOutFromDomainsStates(t *testing.T) {
	if len(columns) != Columns {
		t.Errorf("%d columns are listed and Columns says %d", len(columns), Columns)
	}
	for _, state := range domain.States {
		inColumn := IndexOf(string(state)) >= 0
		if inColumn == Stopped(string(state)) {
			t.Errorf("%s is in a column: %v, and stopped: %v — want exactly one", state, inColumn, Stopped(string(state)))
		}
	}
	if len(Stalled) != 3 {
		t.Errorf("the band holds %v, want the three stopped states", Stalled)
	}
}

func TestInLanes_dealsTheBandIntoItsThreeLanes(t *testing.T) {
	band := []Card{{Number: 1, State: "paused"}, {Number: 2, State: "blocked"}, {Number: 3, State: "blocked"}}
	lanes := InLanes(band)
	if len(lanes) != len(Stalled) {
		t.Fatalf("%d lanes, want %d", len(lanes), len(Stalled))
	}
	got := [][]int{numbersOfCards(lanes[0]), numbersOfCards(lanes[1]), numbersOfCards(lanes[2])}
	if !slices.Equal(got[0], []int{2, 3}) || len(got[1]) != 0 || !slices.Equal(got[2], []int{1}) {
		t.Errorf("lanes hold %v", got)
	}
}
