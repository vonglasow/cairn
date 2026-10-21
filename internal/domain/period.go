package domain

import "time"

// PeriodUnderWay is the day the period under way at now began, periods of that many days following one another from start; nothing before start.
func PeriodUnderWay(start time.Time, days int, now time.Time) (time.Time, bool) {
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, start.Location())
	if days < 1 || today.Before(start) {
		return time.Time{}, false
	}
	from := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, time.UTC)
	to := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC)
	elapsed := int(to.Sub(from).Hours() / 24)
	return start.AddDate(0, 0, elapsed/days*days), true
}
