package tui

import (
	"strings"
	"testing"

	"cairn/internal/engine"
)

// The third tab is the time spent, a week at a time; tab walks the three
// round, and the arrows walk the weeks, t coming back to today's.
func TestTheTimeTabWalksTheWeeks(t *testing.T) {
	f := newFake()
	b := hit(standing(f), "tab", "tab")
	if b.tab != Time {
		t.Fatalf("two tabs left the board on %v", b.tab)
	}
	if got := hit(b, "tab").tab; got != Tasks {
		t.Errorf("a third tab left the board on %v", got)
	}
	monday := engine.Monday(today)
	if !b.week.Equal(monday) {
		t.Errorf("the week shown starts on %v, want %v", b.week, monday)
	}
	if got := hit(b, "left").week; !got.Equal(monday.AddDate(0, 0, -7)) {
		t.Errorf("left shows the week of %v", got)
	}
	if got := hit(b, "right", "right", "t").week; !got.Equal(monday) {
		t.Errorf("t shows the week of %v", got)
	}
	asked := 0
	for _, a := range f.asked {
		if a.op.Name() == engine.TimesheetShow.Name() {
			asked++
		}
	}
	if asked < 2 {
		t.Errorf("the timesheet was asked for %d times", asked)
	}
	if said := b.Draw(); !strings.Contains(said, "TIME") || !strings.Contains(said, monday.Format("Mon 02/01")) {
		t.Errorf("the time tab draws:\n%s", said)
	}
}

// b backs the store up and B lists the backups, each answer said until esc.
func TestBackupsAreReachedFromTheBoard(t *testing.T) {
	f := newFake()
	b := hit(standing(f), "b")
	added := false
	for _, a := range f.asked {
		added = added || a.op.Name() == engine.BackupNew.Name()
	}
	if !added {
		t.Error("b made no backup")
	}
	if b.screen != OnSaid {
		t.Errorf("b left the board on %v", b.screen)
	}
	if got := hit(b, "esc"); got.screen != OnTheBoard {
		t.Errorf("esc left the board on %v", got.screen)
	}
	hit(standing(f), "B")
	found := false
	for _, a := range f.asked {
		found = found || a.op.Name() == engine.BackupList.Name()
	}
	if !found {
		t.Error("B listed no backup")
	}
}
