package board

import (
	"slices"
	"testing"
	"time"

	"cairn/internal/engine"
	"cairn/internal/engine/series"
)

func TestFactsOf_isEveryFactACardSaysInOrder(t *testing.T) {
	now := time.Date(2026, 11, 15, 10, 0, 0, 0, time.UTC)
	c := Card{Number: 1, State: "todo", Blocks: 2, Deadline: time.Date(2026, 11, 14, 0, 0, 0, 0, time.UTC),
		Spent: "2h", Estimate: "1h", Overrun: true, SubDone: 1, SubTotal: 2,
		Labels: []string{"api"}, Serves: []string{"Ship", "Tidy"}}
	var marks []Mark
	var tones []Tone
	for _, f := range FactsOf(c, now, engine.Pages{}) {
		marks, tones = append(marks, f.Mark), append(tones, f.Tone)
	}
	if want := []Mark{"", "", "", Steps, Label, Serves, Serves}; !slices.Equal(marks, want) {
		t.Errorf("marks %v, want %v", marks, want)
	}
	if want := []Tone{Normal, Late, Late, Faint, Says, Says, Says}; !slices.Equal(tones, want) {
		t.Errorf("tones %v, want %v", tones, want)
	}
	if got := FactsOf(c, now, engine.PagesOff(engine.PageTime)); len(got) != 6 {
		t.Errorf("with the time off, %d facts, want the cost left out", len(got))
	}
}

func TestFactsOf_aStoppedCardSaysWhatIsInItsWayOrThatItIsFree(t *testing.T) {
	now := time.Now()
	stopped := Card{Number: 1, State: "blocked", Blockers: []int{3}}
	if got := FactsOf(stopped, now, engine.Pages{}); len(got) == 0 || got[0].Mark != InTheWay || got[0].Tone != Faint {
		t.Errorf("a blocked card says %+v, want what is in its way first", got)
	}
	freed := Card{Number: 1, State: "blocked", Freed: true, Labels: []string{"api"}}
	if got := FactsOf(freed, now, engine.Pages{}); len(got) != 1 || got[0].Mark != Freed || got[0].Tone != Accent {
		t.Errorf("a freed card says %+v, want that nothing blocks it, alone", got)
	}
}

func TestFactsOf_anOccurrencePassedIsNotLate(t *testing.T) {
	now := time.Date(2026, 11, 15, 10, 0, 0, 0, time.UTC)
	c := Card{Number: 1, State: "todo", Series: &series.Occurs{Day: "2026-11-14"}}
	if got := FactsOf(c, now, engine.Pages{}); len(got) == 0 || (got[0].Tone != Late && got[0].Tone != Passed) {
		t.Errorf("an occurrence of yesterday says %+v", got)
	}
}

func TestMarksOf_isWhatACardWearsBesideItsNumber(t *testing.T) {
	c := Card{Number: 1, State: "paused", Archived: true, Series: &series.Occurs{}, Description: "why"}
	if got, want := MarksOf(c), []Mark{Archived, Repeats, Described, Paused}; !slices.Equal(got, want) {
		t.Errorf("marks %v, want %v", got, want)
	}
	if got := MarksOf(Card{State: "todo"}); len(got) != 0 {
		t.Errorf("a plain card wears %v", got)
	}
}

func TestCutFromStart_keepsTheEndRead(t *testing.T) {
	within := func(room int) func(string) bool {
		return func(s string) bool { return len([]rune(s)) <= room }
	}
	for _, c := range []struct {
		said string
		room int
		want string
	}{
		{"Home › Enclosure", 20, "Home › Enclosure"},
		{"Home › Enclosure", 12, "…› Enclosure"},
		{"Home › Enclosure", 2, "…e"},
		{"Home › Enclosure", 1, ""},
	} {
		if got := CutFromStart(c.said, within(c.room)); got != c.want {
			t.Errorf("CutFromStart(%q, %d) = %q, want %q", c.said, c.room, got, c.want)
		}
	}
}
