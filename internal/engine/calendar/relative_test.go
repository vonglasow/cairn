package calendar

import (
	"errors"
	"testing"
	"time"
)

func TestFromToday_refusesACountItCannotRead(t *testing.T) {
	for _, said := range []string{"+3x", "+d", "3d", "+3", "-w", "+1.5d"} {
		if _, err := FromToday(said, time.Now()); !errors.Is(err, ErrNotADayFromToday) {
			t.Errorf("FromToday(%q) = %v, want ErrNotADayFromToday", said, err)
		}
	}
}
