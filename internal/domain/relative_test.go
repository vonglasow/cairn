package domain

import (
	"errors"
	"testing"
	"time"
)

func TestFromToday_readsADayCountedFromToday(t *testing.T) {
	today := time.Date(2026, 11, 3, 15, 0, 0, 0, time.Local)
	for said, want := range map[string]string{
		"+3d": "2026-11-06", "-1d": "2026-11-02", "+2w": "2026-11-17", "-1w": "2026-10-27", "+0d": "2026-11-03",
		"2026-10-01": "2026-10-01", "": "",
	} {
		if got, err := FromToday(said, today); err != nil || got != want {
			t.Errorf("FromToday(%q) = %q, %v; want %q", said, got, err, want)
		}
	}
}

func TestFromToday_refusesACountItCannotRead(t *testing.T) {
	for _, said := range []string{"+3x", "+d", "3d", "+3", "-w", "+1.5d"} {
		if _, err := FromToday(said, time.Now()); !errors.Is(err, ErrNotADayFromToday) {
			t.Errorf("FromToday(%q) = %v, want ErrNotADayFromToday", said, err)
		}
	}
}
