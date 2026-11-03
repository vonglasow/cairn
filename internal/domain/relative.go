package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// ErrNotADayFromToday is returned when a day counted from today is not written +3d, -1d or +2w.
var ErrNotADayFromToday = errors.New("a day counted from today is written +3d, -1d or +2w")

// FromToday writes a day counted from today, +3d or -2w, as the day it is; anything else is handed back as it was given.
func FromToday(said string, today time.Time) (string, error) {
	counted := strings.HasPrefix(said, "+") || strings.HasPrefix(said, "-") || strings.HasSuffix(said, "d") || strings.HasSuffix(said, "w")
	if !counted {
		return said, nil
	}
	days, ok := daysCounted(said)
	if !ok {
		return "", fmt.Errorf("%q: %w", said, ErrNotADayFromToday)
	}
	return FormatDay(today.AddDate(0, 0, days)), nil
}

func daysCounted(said string) (int, bool) {
	if len(said) < 3 {
		return 0, false
	}
	sign, digits, unit := said[0], said[1:len(said)-1], said[len(said)-1]
	if sign != '+' && sign != '-' || unit != 'd' && unit != 'w' {
		return 0, false
	}
	n := 0
	for _, d := range digits {
		if d < '0' || d > '9' {
			return 0, false
		}
		n = n*10 + int(d-'0')
	}
	if unit == 'w' {
		n *= 7
	}
	if sign == '-' {
		n = -n
	}
	return n, true
}
