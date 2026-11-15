// Package calendar holds days, deadlines, estimates, durations and periods.
package calendar

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

const dayLayout = "2006-01-02"

// What a day or a duration given can be turned down for.
var (
	ErrNotADay          = errors.New("not a date")
	ErrNotADuration     = errors.New("not a duration")
	ErrNegativeAmount   = errors.New("a duration cannot run backwards")
	ErrFinerThanASecond = errors.New("a duration is kept to the second, not finer")
)

// ParseDeadline reads a day.
func ParseDeadline(day string) (time.Time, error) {
	if day == "" {
		return time.Time{}, nil
	}
	parsed, err := time.Parse(dayLayout, day)
	if err != nil {
		return time.Time{}, fmt.Errorf("%q: %w — a day written %s", day, ErrNotADay, dayLayout)
	}
	return parsed, nil
}

// ParseEstimate reads a duration.
func ParseEstimate(amount string) (time.Duration, error) {
	if amount == "" {
		return 0, nil
	}
	parsed, err := time.ParseDuration(amount)
	if err != nil {
		return 0, fmt.Errorf("%q: %w — an amount and a unit, as in 2h or 1h30m", amount, ErrNotADuration)
	}
	if parsed < 0 {
		return 0, fmt.Errorf("%q: %w", amount, ErrNegativeAmount)
	}
	if parsed%time.Second != 0 {
		return 0, fmt.Errorf("%q: %w", amount, ErrFinerThanASecond)
	}
	return parsed, nil
}

// FormatDuration writes a duration the way it was meant to be read: 90 minutes comes back as 1h30m, not as 1h30m0s.
func FormatDuration(d time.Duration) string {
	if d == 0 {
		return ""
	}
	var out strings.Builder
	if h := int(d.Hours()); h > 0 {
		fmt.Fprintf(&out, "%dh", h)
		d -= time.Duration(h) * time.Hour
	}
	if m := int(d.Minutes()); m > 0 {
		fmt.Fprintf(&out, "%dm", m)
		d -= time.Duration(m) * time.Minute
	}
	if s := int(d.Seconds()); s > 0 {
		fmt.Fprintf(&out, "%ds", s)
	}
	return out.String()
}

// FormatDay writes a deadline back as the day it was given.
func FormatDay(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(dayLayout)
}

// FormatMoment writes a moment the way a person reads one: local time, to the minute.
func FormatMoment(t time.Time) string {
	return t.Local().Format(momentLayout)
}

const momentLayout = "2006-01-02 15:04"

// ParseMoment reads back a moment FormatMoment wrote, in local time.
func ParseMoment(s string) (time.Time, error) {
	return time.ParseInLocation(momentLayout, s, time.Local)
}
