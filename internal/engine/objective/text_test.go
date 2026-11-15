package objective

import (
	"cairn/internal/engine/calendar"
	"cairn/internal/engine/item"
	"strings"
	"testing"
	"time"
)

func plan(t *testing.T) Objective {
	t.Helper()
	built, err := NewObjective("Ship the release")
	if err != nil {
		t.Fatalf("NewObjective: %v", err)
	}
	a := From(built)
	a.Numbered(1)
	return a.Objective()
}

func TestText_showsWhatTheObjectiveCarries(t *testing.T) {
	built := plan(t)
	deadline, err := calendar.ParseDeadline("2026-10-01")
	if err != nil {
		t.Fatalf("ParseDeadline: %v", err)
	}
	estimate, err := calendar.ParseEstimate("5h")
	if err != nil {
		t.Fatalf("ParseEstimate: %v", err)
	}
	a := From(built)
	a.Described("the whole thing")
	a.Labelled("launch")
	a.Labelled("team")
	a.Due(deadline)
	a.Estimated(estimate)
	built = a.Objective()

	got := Text(built, item.Progress{}, Review{}, Family{})
	for _, want := range []string{"the whole thing", "launch, team", "2026-10-01", "5h"} {
		if !strings.Contains(got, want) {
			t.Errorf("text is missing %q:\n%s", want, got)
		}
	}
}

func TestText_saysTheTimeSpentAndWhatIsLeft(t *testing.T) {
	got := Text(plan(t), item.Progress{}, Review{Spent: 2*time.Hour + 15*time.Minute,
		Left: Left{Tasks: 2, Due: time.Date(2026, 11, 5, 0, 0, 0, 0, time.UTC), On: 3}}, Family{})
	for _, want := range []string{"time spent   2h15m", "left         2 tasks, the next deadline 2026-11-05 (#3)"} {
		if !strings.Contains(got, want) {
			t.Errorf("text is missing %q:\n%s", want, got)
		}
	}
}

func TestReport_saysThePeriodTheMonthsAndWhatWasDone(t *testing.T) {
	day := func(m time.Month, d int) time.Time { return time.Date(2026, m, d, 0, 0, 0, 0, time.UTC) }
	in := Period{
		Since:  day(8, 1),
		Until:  day(9, 30),
		Months: []Month{{Month: day(8, 1), Spent: 90 * time.Minute}, {Month: day(9, 1), Spent: 45 * time.Minute}},
		Done:   []Finished{{Number: 3, Title: "Write the spec", On: day(9, 14)}},
	}
	got := Report(plan(t), item.Progress{}, Review{Spent: 4*time.Hour + 15*time.Minute}, in)
	for _, want := range []string{
		"Ship the release — from 2026-08-01 to 2026-09-30",
		"time spent  2h15m in the period, 4h15m in all",
		"by month\n  2026-08  1h30m\n  2026-09  45m\n",
		"done in the period\n  #3  Write the spec  · 2026-09-14\n",
		"left        nothing",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("the report is missing %q:\n%s", want, got)
		}
	}
	empty := Report(plan(t), item.Progress{}, Review{}, Period{Since: day(8, 1), Until: day(8, 31)})
	for _, want := range []string{"time spent  — in the period, — in all", "nothing was spent in the period", "nothing was done in the period"} {
		if !strings.Contains(empty, want) {
			t.Errorf("an empty report is missing %q:\n%s", want, empty)
		}
	}
}
