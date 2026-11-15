package calendar

import (
	"errors"
)

// ErrDayToCome is returned when a day has not happened yet: work cannot be said to have been done tomorrow.
var ErrDayToCome = errors.New("that has not happened yet")

// ErrNotATime is returned when time added by hand is not a length of time, or is none at all.
var ErrNotATime = errors.New("time added is an amount and a unit, as in 1h or 45m — less than nothing, as in -30m, takes some back")
