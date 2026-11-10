package domain

import (
	"errors"
	"slices"
	"strings"
	"testing"
	"time"
)

func day(t *testing.T, said string) time.Time {
	t.Helper()
	d, err := ParseDeadline(said)
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func days(dates []time.Time) string {
	said := make([]string, 0, len(dates))
	for _, d := range dates {
		said = append(said, FormatDay(d))
	}
	return strings.Join(said, " ")
}

func TestParseRuleReadsPlainWords(t *testing.T) {
	for _, c := range []struct{ typed, words, often string }{
		{"every weekday 10:30", "every weekday 10:30", "every weekday at 10:30"},
		{"weekday at 10:30", "every weekday 10:30", "every weekday at 10:30"},
		{"every other thursday 14:00", "every 2 weeks on thu 14:00", "every 2 weeks on Thursday at 14:00"},
		{"every 3 days after done", "every 3 days after done", "every 3 days"},
		{"every 3 days after the last is done", "every 3 days after done", "every 3 days"},
		{"every day", "every day", "every day"},
		{"every week", "every week", "every week"},
		{"Every Monday, Thursday", "every week on mon thu", "every week on Monday and Thursday"},
		{"every mon and wed and fri", "every week on mon wed fri", "every week on Monday, Wednesday and Friday"},
		{"every weekend", "every week on sat sun", "every week on Saturday and Sunday"},
		{"every 2 months", "every 2 months", "every 2 months"},
		{"every month", "every month", "every month"},
		{"every day 10 times", "every day 10 times", "every day"},
		{"every day for 1 time", "every day 1 times", "every day"},
		{"every day until 2026-12-01", "every day until 2026-12-01", "every day"},
		{"every 4 weeks on monday", "every 4 weeks on mon", "every 4 weeks on Monday"},
		{"every tues", "every week on tue", "every week on Tuesday"},
		{"every sundays 9:05", "every week on sun 09:05", "every week on Sunday at 09:05"},
	} {
		r, err := ParseRule(c.typed)
		if err != nil {
			t.Fatalf("%q: %v", c.typed, err)
		}
		if r.Words() != c.words {
			t.Errorf("%q reads as %q, want %q", c.typed, r.Words(), c.words)
		}
		if r.Often() != c.often {
			t.Errorf("%q says %q, want %q", c.typed, r.Often(), c.often)
		}
		again, err := ParseRule(r.Words())
		if err != nil || again.Words() != r.Words() {
			t.Errorf("%q read back as %q (%v)", r.Words(), again.Words(), err)
		}
	}
}

func TestParseRuleRefusesWhatItCannotRead(t *testing.T) {
	for _, c := range []struct{ typed, says string }{
		{"every blue moon", `"blue"`},
		{"", "how often"},
		{"10:30", "how often"},
		{"every day 3 times until 2026-12-01", "an end is a number of times or a day, never both"},
		{"every monday after done", "names no day of the week"},
		{"every 3 days on monday", "a day of the week goes with weeks"},
		{"every month on friday", "a day of the week goes with weeks"},
		{"every 0 days", "1 to 999"},
		{"every day 0 times", "1 to 999"},
		{"every day 25:00", `"25:00"`},
		{"every day 10:61", `"10:61"`},
		{"every day until tomorrow", "tomorrow"},
		{"every day until", "until a day"},
		{"every day after", "after done"},
		{"every day after the", "after done"},
		{"every 3", "how often"},
		{"every 1000 days", "1 to 999"},
		{"every 99999999999999999999999 months", "1 to 999"},
		{"every day 1000 times", "1 to 999"},
		{"every day 1:5", `"1:5"`},
		{"every day a:30", `"a:30"`},
	} {
		_, err := ParseRule(c.typed)
		if !errors.Is(err, ErrNotARule) {
			t.Fatalf("%q: %v, want it refused as not a rule", c.typed, err)
		}
		if !strings.Contains(err.Error(), c.says) {
			t.Errorf("%q refused with %q, want it to say %q", c.typed, err, c.says)
		}
	}
}

func TestRuleSaysItsAnchorAndItsEnd(t *testing.T) {
	from := day(t, "2026-11-09")
	for _, c := range []struct{ typed, said string }{
		{"every weekday 10:30", "every weekday at 10:30, on the calendar, from Mon 9 Nov 2026, never ending"},
		{"every 3 days after done", "every 3 days, after the last is done, from Mon 9 Nov 2026, never ending"},
		{"every monday 3 times", "every week on Monday, on the calendar, from Mon 9 Nov 2026, 3 times"},
		{"every day until 2026-12-01", "every day, on the calendar, from Mon 9 Nov 2026, until Tue 1 Dec 2026"},
		{"every day 1 times", "every day, on the calendar, from Mon 9 Nov 2026, once"},
	} {
		r, err := ParseRule(c.typed)
		if err != nil {
			t.Fatal(err)
		}
		if got := r.Sentence(from); got != c.said {
			t.Errorf("%q says %q, want %q", c.typed, got, c.said)
		}
	}
}

func TestRuleKnowsItsAnchorAndHowManyTimes(t *testing.T) {
	r, _ := ParseRule("every 3 days after done 4 times")
	if !r.AfterDone() || r.Times() != 4 || r.At() != "" {
		t.Errorf("%q: after done %v, times %d, at %q", r.Words(), r.AfterDone(), r.Times(), r.At())
	}
	r, _ = ParseRule("every weekday 9:00")
	if r.AfterDone() || r.Times() != 0 || r.At() != "09:00" {
		t.Errorf("%q: after done %v, times %d, at %q", r.Words(), r.AfterDone(), r.Times(), r.At())
	}
	if r.Short() != "every weekday at 09:00" {
		t.Errorf("short reads %q", r.Short())
	}
	r, _ = ParseRule("every 3 days after done")
	if r.Short() != "every 3 days, after the last is done" {
		t.Errorf("short reads %q", r.Short())
	}
}

func TestRuleDatesOnTheCalendar(t *testing.T) {
	from := day(t, "2026-11-09")
	for _, c := range []struct{ typed, after, want string }{
		{"every weekday", "2026-11-08", "2026-11-09 2026-11-10 2026-11-11 2026-11-12 2026-11-13 2026-11-16"},
		{"every other thursday", "2026-11-08", "2026-11-12 2026-11-26 2026-12-10"},
		{"every 2 weeks on mon fri", "2026-11-08", "2026-11-09 2026-11-13 2026-11-23 2026-11-27"},
		{"every 3 days", "2026-11-08", "2026-11-09 2026-11-12 2026-11-15"},
		{"every week", "2026-11-09", "2026-11-16 2026-11-23"},
		{"every 2 weeks", "2026-11-08", "2026-11-09 2026-11-23"},
		{"every month", "2026-11-09", "2026-12-09 2027-01-09"},
		{"every day until 2026-11-11", "2026-11-08", "2026-11-09 2026-11-10 2026-11-11"},
		{"every friday", "2026-11-13", "2026-11-20 2026-11-27"},
	} {
		r, err := ParseRule(c.typed)
		if err != nil {
			t.Fatal(err)
		}
		n := len(strings.Fields(c.want))
		if got := days(r.Dates(day(t, c.after), from, n+2)); !strings.HasPrefix(got, c.want) || strings.Contains(c.typed, "until") && got != c.want {
			t.Errorf("%q after %s: %s, want %s", c.typed, c.after, got, c.want)
		}
	}
}

func TestRuleMonthsKeepTheDayOrTheLastOne(t *testing.T) {
	r, _ := ParseRule("every month")
	got := days(r.Dates(day(t, "2026-01-31"), day(t, "2026-01-31"), 3))
	if got != "2026-02-28 2026-03-31 2026-04-30" {
		t.Errorf("monthly from the 31st: %s", got)
	}
}

func TestRuleFirstIsTheFirstDateFromItsStart(t *testing.T) {
	r, _ := ParseRule("every thursday")
	first, ok := r.First(day(t, "2026-11-09"))
	if !ok || FormatDay(first) != "2026-11-12" {
		t.Errorf("first thursday from a monday: %s %v", FormatDay(first), ok)
	}
	r, _ = ParseRule("every 3 days after done")
	first, ok = r.First(day(t, "2026-11-09"))
	if !ok || FormatDay(first) != "2026-11-09" {
		t.Errorf("first after done: %s %v", FormatDay(first), ok)
	}
	r, _ = ParseRule("every day until 2026-11-01")
	if _, ok := r.First(day(t, "2026-11-09")); ok {
		t.Error("a rule ended before its start has a first date")
	}
}

func TestRuleAfterDoneCountsFromTheDayItWasDone(t *testing.T) {
	for _, c := range []struct{ typed, want string }{
		{"every 3 days after done", "2026-11-13"},
		{"every 2 weeks after done", "2026-11-24"},
		{"every month after done", "2026-12-10"},
	} {
		r, _ := ParseRule(c.typed)
		next, ok := r.Next(day(t, "2026-11-10"), day(t, "2026-11-01"))
		if !ok || FormatDay(next) != c.want {
			t.Errorf("%q done on the 10th: %s %v, want %s", c.typed, FormatDay(next), ok, c.want)
		}
	}
	r, _ := ParseRule("every 3 days after done until 2026-11-12")
	if _, ok := r.Next(day(t, "2026-11-10"), day(t, "2026-11-01")); ok {
		t.Error("a date past its end was given")
	}
	if then := r.Then(); then != "3 days after it is done" {
		t.Errorf("then reads %q", then)
	}
	r, _ = ParseRule("every month after done")
	if then := r.Then(); then != "1 month after it is done" {
		t.Errorf("then reads %q", then)
	}
	if got := r.Dates(day(t, "2026-11-10"), day(t, "2026-11-01"), 3); got != nil {
		t.Errorf("after done has no dates to come: %v", got)
	}
}

func TestRuleSaysADayTheWayAPersonReadsIt(t *testing.T) {
	if got := ShortDay(day(t, "2026-11-09")); got != "Mon 9 Nov" {
		t.Errorf("short day %q", got)
	}
	if got := LongDay(day(t, "2026-11-09")); got != "Mon 9 Nov 2026" {
		t.Errorf("long day %q", got)
	}
}

func TestRuleDayOfAMoment(t *testing.T) {
	moment := time.Date(2026, 11, 9, 23, 30, 0, 0, time.FixedZone("far", 10*3600))
	if got := DayOf(moment); !got.Equal(day(t, "2026-11-09")) {
		t.Errorf("day of %v is %v", moment, got)
	}
}

func TestRuleNeverFallsBeforeItsStart(t *testing.T) {
	r, _ := ParseRule("every day")
	if got := days(r.Dates(day(t, "2026-11-01"), day(t, "2026-11-09"), 2)); got != "2026-11-09 2026-11-10" {
		t.Errorf("from the 9th, after the 1st: %s", got)
	}
	if _, err := ParseRule("every day after blue"); !errors.Is(err, ErrNotARule) {
		t.Errorf("after followed by another word: %v", err)
	}
}

func TestRuleTakenApartIsWrittenBackAsTheSameRule(t *testing.T) {
	for _, typed := range []string{"every weekday 10:30", "every 2 weeks on thu 14:00", "every 3 days after done", "every month 5 times", "every day until 2026-12-31"} {
		rule, err := ParseRule(typed)
		if err != nil {
			t.Fatal(err)
		}
		if got := rule.Parts().Words(); got != rule.Words() {
			t.Errorf("%q taken apart is written %q, want %q", typed, got, rule.Words())
		}
	}
	parts, _ := ParseRule("every other thursday 14:00 3 times")
	if p := parts.Parts(); p.Every != 2 || p.Unit != "weeks" || !slices.Equal(p.Days, []time.Weekday{time.Thursday}) || p.At != "14:00" || p.Times != 3 || p.AfterDone {
		t.Errorf("the parts read %+v", p)
	}
	built := RuleParts{Every: 1, Unit: "days", Days: []time.Weekday{time.Monday}, AfterDone: true, Times: 4, Until: time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC)}
	if got := built.Words(); got != "every day after done 4 times" {
		t.Errorf("days on a rule counted in days, or an end said twice, are written %q", got)
	}
	weekly := RuleParts{Every: 1, Unit: "weeks", Days: []time.Weekday{time.Friday, time.Monday}, AfterDone: true, Until: time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC)}
	if got := weekly.Words(); got != "every week on mon fri until 2026-12-01" {
		t.Errorf("weekdays and after done are written %q", got)
	}
	if got := (RuleParts{}).Words(); got != "every day" {
		t.Errorf("nothing chosen is written %q", got)
	}
}

func TestRuleSaidLeavesTheStartOut(t *testing.T) {
	for typed, want := range map[string]string{
		"every weekday 10:30":            "every weekday at 10:30, on the calendar",
		"every 3 days after done 1 time": "every 3 days, after the last is done, once",
		"every day until 2026-12-31":     "every day, on the calendar, until Thu 31 Dec 2026",
	} {
		rule, err := ParseRule(typed)
		if err != nil {
			t.Fatal(err)
		}
		if got := rule.Said(); got != want {
			t.Errorf("%q is said %q, want %q", typed, got, want)
		}
	}
}
