package domain

import (
	"testing"
	"time"
)

func TestThePeriodUnderWayBeganOnTheLastOfItsDays(t *testing.T) {
	day := func(s string) time.Time {
		at, err := time.ParseInLocation(time.DateOnly, s, time.Local)
		if err != nil {
			t.Fatal(err)
		}
		return at
	}
	start := day("2026-10-21")
	for _, c := range []struct {
		now, want string
		days      int
		begun     bool
	}{
		{"2026-10-20", "", 14, false},
		{"2026-10-21", "2026-10-21", 14, true},
		{"2026-11-03", "2026-10-21", 14, true},
		{"2026-11-04", "2026-11-04", 14, true},
		{"2026-12-02", "2026-12-02", 14, true},
		{"2026-10-25", "2026-10-25", 1, true},
		{"2027-04-01", "2027-03-24", 14, true},
	} {
		got, begun := PeriodUnderWay(start, c.days, day(c.now).Add(15*time.Hour))
		if begun != c.begun || begun && !got.Equal(day(c.want)) {
			t.Errorf("on %s with periods of %d days, the period under way began on %v (%v), want %s", c.now, c.days, got, begun, c.want)
		}
	}
}
