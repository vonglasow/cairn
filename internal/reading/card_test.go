package reading

import (
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
