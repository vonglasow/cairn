package engine

import (
	"testing"
	"time"
)

func TestADeadlineIsWrittenOnlyWhenItIsLateOrWithinAWeek(t *testing.T) {
	now := time.Date(2026, 6, 10, 15, 0, 0, 0, time.Local)
	on := func(days int) time.Time { return startOfDay(now).AddDate(0, 0, days) }
	for _, tc := range []struct {
		due        time.Time
		text       string
		show, late bool
	}{
		{time.Time{}, "", false, false},
		{on(-2), "−2d", true, true},
		{on(0), "0d", true, false},
		{on(3), "3d", true, false},
		{on(soon), "7d", true, false},
		{on(soon + 1), "", false, false},
	} {
		text, show, late := Deadline(tc.due, now)
		if text != tc.text || show != tc.show || late != tc.late {
			t.Errorf("Deadline(%v) = %q, %v, %v, want %q, %v, %v", tc.due, text, show, late, tc.text, tc.show, tc.late)
		}
	}
}

func TestACardDoneSaysNothingOfItsDay(t *testing.T) {
	now := time.Date(2026, 6, 10, 15, 0, 0, 0, time.Local)
	due := startOfDay(now).AddDate(0, 0, 2)
	if text, show, _ := CardDeadline(Card{State: "done", Deadline: due}, now); show || text != "" {
		t.Errorf("a card done says %q of its day", text)
	}
	if text, show, _ := CardDeadline(Card{State: "todo", Deadline: due}, now); !show || text != "2d" {
		t.Errorf("a card to do says %q of its day, shown %v", text, show)
	}
}

func TestStepsAreTalliedLeftAndDone(t *testing.T) {
	parts := []Part{{Title: "a", Done: true}, {Title: "b"}, {Title: "c"}}
	if left, done := Tally(parts); left != 2 || done != 1 {
		t.Errorf("Tally = %d left, %d done, want 2 and 1", left, done)
	}
	if got := Tallied(parts); got != "2 left · 1 done" {
		t.Errorf("Tallied = %q", got)
	}
	if got := Tallied(nil); got != "" {
		t.Errorf("Tallied with no step = %q, want nothing", got)
	}
}
