package engine

import (
	"cairn/internal/engine/calendar"
	"testing"
	"time"

	"cairn/internal/engine/act"
)

func TestSpentByDay_reachesBackToTheDayTimeWasAddedFor(t *testing.T) {
	e := openEngine(t)
	day := calendar.FormatDay(time.Now().AddDate(0, 0, -10))
	if err := e.record(1, act.TimeAdded, "", day, "2h"); err != nil {
		t.Fatal(err)
	}
	spent, err := e.spentByDay()
	if err != nil {
		t.Fatal(err)
	}
	if got := spent[1]; got[day] != "2h" || len(got) != 1 {
		t.Errorf("spent %v, want 2h on %s", got, day)
	}
}
