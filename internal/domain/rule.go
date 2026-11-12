package domain

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"
)

// ErrNotARule is what a rule cairn cannot read is refused with.
var ErrNotARule = errors.New("not a rule cairn reads")

const ruleWords = "every N days, weeks or months, a day of the week, weekday, a time as 10:30, after done, N times, until 2026-12-31"

type unit int

const (
	noUnit unit = iota
	inDays
	inWeeks
	inMonths
)

var unitNames = map[unit][2]string{inDays: {"day", "days"}, inWeeks: {"week", "weeks"}, inMonths: {"month", "months"}}

var weekdays = map[string]time.Weekday{
	"mon": time.Monday, "monday": time.Monday, "mondays": time.Monday,
	"tue": time.Tuesday, "tues": time.Tuesday, "tuesday": time.Tuesday, "tuesdays": time.Tuesday,
	"wed": time.Wednesday, "weds": time.Wednesday, "wednesday": time.Wednesday, "wednesdays": time.Wednesday,
	"thu": time.Thursday, "thur": time.Thursday, "thurs": time.Thursday, "thursday": time.Thursday, "thursdays": time.Thursday,
	"fri": time.Friday, "friday": time.Friday, "fridays": time.Friday,
	"sat": time.Saturday, "saturday": time.Saturday, "saturdays": time.Saturday,
	"sun": time.Sunday, "sunday": time.Sunday, "sundays": time.Sunday,
}

var workingWeek = []time.Weekday{time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday}

// Rule is how often something comes back: every so many days, weeks or months, on days of the week, at a time, counted on the calendar or from the day the last was done, and when it ends.
type Rule struct {
	every     int
	unit      unit
	days      []time.Weekday
	at        string
	afterDone bool
	times     int
	until     time.Time
}

type reading struct {
	words []string
	at    int
	rule  Rule
}

func (r *reading) next() (string, bool) {
	if r.at+1 >= len(r.words) {
		return "", false
	}
	r.at++
	return r.words[r.at], true
}

func refused(format string, args ...any) error {
	return fmt.Errorf(format+": %w — "+ruleWords, append(args, ErrNotARule)...)
}

// ParseRule reads a rule written in plain words: every weekday 10:30, every other thursday 14:00, every 3 days after done.
func ParseRule(typed string) (Rule, error) {
	r := &reading{words: strings.Fields(strings.ToLower(strings.ReplaceAll(typed, ",", " "))), at: -1, rule: Rule{every: 1}}
	for word, more := r.next(); more; word, more = r.next() {
		if err := r.read(word); err != nil {
			return Rule{}, err
		}
	}
	return r.rule.checked()
}

func (r *reading) read(word string) error {
	if day, isDay := weekdays[word]; isDay {
		r.rule.days = append(r.rule.days, day)
		return nil
	}
	if n, ok := whole(word); ok {
		return r.number(word, n)
	}
	if hours, minutes, ok := strings.Cut(word, ":"); ok {
		return r.time(word, hours, minutes)
	}
	switch word {
	case "every", "on", "at", "and", "for":
		return nil
	case "other":
		r.rule.every = 2
	case "day", "days":
		r.rule.unit = inDays
	case "week", "weeks":
		r.rule.unit = inWeeks
	case "month", "months":
		r.rule.unit = inMonths
	case "weekday", "weekdays":
		r.rule.days = append(r.rule.days, workingWeek...)
	case "weekend", "weekends":
		r.rule.days = append(r.rule.days, time.Saturday, time.Sunday)
	case "after":
		return r.afterDone()
	case "until":
		return r.until()
	default:
		return refused("%q", word)
	}
	return nil
}

func (r *reading) number(word string, n int) error {
	if r.at+1 < len(r.words) && (r.words[r.at+1] == "times" || r.words[r.at+1] == "time") {
		r.next()
		r.rule.times = n
		if n < 1 || n > most {
			return refused("%q: a rule ends after 1 to 999 times", word)
		}
		return nil
	}
	if n < 1 || n > most {
		return refused("%q: a rule comes back every 1 to 999", word)
	}
	r.rule.every = n
	return nil
}

const most = 999

func whole(word string) (int, bool) {
	n := 0
	for _, r := range word {
		if r < '0' || r > '9' {
			return 0, false
		}
		n = min(n*10+int(r-'0'), most+1)
	}
	return n, word != ""
}

func (r *reading) time(word, h, m string) error {
	hours, isHours := whole(h)
	minutes, isMinutes := whole(m)
	if !isHours || !isMinutes || len(h) > 2 || len(m) != 2 || hours > 23 || minutes > 59 {
		return refused("%q: a time is written 10:30", word)
	}
	r.rule.at = fmt.Sprintf("%02d:%02d", hours, minutes)
	return nil
}

func (r *reading) afterDone() error {
	for word, more := r.next(); more; word, more = r.next() {
		if word == "done" {
			r.rule.afterDone = true
			return nil
		}
		if !slices.Contains([]string{"the", "last", "one", "is"}, word) {
			break
		}
	}
	return refused("after: after done is how it is said")
}

func (r *reading) until() error {
	word, more := r.next()
	if !more {
		return refused("until: until a day is how it is said")
	}
	day, err := ParseDeadline(word)
	if err != nil {
		return refused("%q: until a day written 2026-12-31", word)
	}
	r.rule.until = day
	return nil
}

func (r Rule) checked() (Rule, error) {
	if len(r.days) > 0 {
		if r.unit == inDays || r.unit == inMonths {
			return Rule{}, refused("%s: a day of the week goes with weeks", unitNames[r.unit][1])
		}
		r.unit = inWeeks
		slices.SortFunc(r.days, func(a, b time.Weekday) int { return sinceMonday(a) - sinceMonday(b) })
		r.days = slices.Compact(r.days)
	}
	if r.unit == noUnit {
		return Rule{}, refused("it says how often: every N days, weeks or months, or a day of the week")
	}
	if r.times > 0 && !r.until.IsZero() {
		return Rule{}, refused("an end is a number of times or a day, never both")
	}
	if r.afterDone && len(r.days) > 0 {
		return Rule{}, refused("after done counts days, weeks or months from the day the last is done, and names no day of the week")
	}
	return r, nil
}

func (r Rule) weekdaysOnly() bool {
	return r.every == 1 && slices.Equal(r.days, workingWeek)
}

// Words is the rule written the one way cairn keeps it, read back by ParseRule as the same rule.
func (r Rule) Words() string {
	said := []string{"every"}
	switch {
	case r.weekdaysOnly():
		said = append(said, "weekday")
	default:
		said = append(said, r.howOften())
		if len(r.days) > 0 {
			said = append(said, "on")
			for _, d := range r.days {
				said = append(said, strings.ToLower(d.String()[:3]))
			}
		}
	}
	if r.at != "" {
		said = append(said, r.at)
	}
	if r.afterDone {
		said = append(said, "after done")
	}
	if r.times > 0 {
		said = append(said, fmt.Sprintf("%d times", r.times))
	}
	if !r.until.IsZero() {
		said = append(said, "until "+FormatDay(r.until))
	}
	return strings.Join(said, " ")
}

func (r Rule) howOften() string {
	if r.every == 1 {
		return unitNames[r.unit][0]
	}
	return fmt.Sprintf("%d %s", r.every, unitNames[r.unit][1])
}

// Often is how often it comes back, the way a person reads it: every weekday at 10:30, every 2 weeks on Thursday.
func (r Rule) Often() string {
	said := "every " + r.howOften()
	if r.weekdaysOnly() {
		said = "every weekday"
	} else if len(r.days) > 0 {
		names := make([]string, 0, len(r.days))
		for _, d := range r.days {
			names = append(names, d.String())
		}
		said += " on " + andJoined(names)
	}
	if r.at != "" {
		said += " at " + r.at
	}
	return said
}

// Listed is how a row of a list writes the rule, the way a person reads it at a glance: every weekday · 10:30, every 4 weeks · monday, every 3 days, after the last is done.
func (r Rule) Listed() string {
	names := make([]string, 0, len(r.days))
	for _, d := range r.days {
		names = append(names, strings.ToLower(d.String()))
	}
	said := "every " + r.howOften()
	if r.weekdaysOnly() {
		said = "every weekday"
	} else if len(r.days) > 0 && r.every == 1 {
		said = "every " + strings.Join(names, ", ")
	} else if len(r.days) > 0 {
		said += " · " + strings.Join(names, ", ")
	}
	if r.at != "" {
		said += " · " + r.at
	}
	if r.afterDone {
		said += ", after the last is done"
	}
	return said
}

func andJoined(names []string) string {
	if len(names) == 1 {
		return names[0]
	}
	return strings.Join(names[:len(names)-1], ", ") + " and " + names[len(names)-1]
}

// Short is how often it comes back and, when it counts from the last one done, that it does.
func (r Rule) Short() string {
	if r.afterDone {
		return r.Often() + ", after the last is done"
	}
	return r.Often()
}

// Sentence is the whole rule the way a person reads it: how often, what it counts from, from which day, and when it ends.
func (r Rule) Sentence(from time.Time) string {
	end := r.end()
	if end == "" {
		end = "never ending"
	}
	return fmt.Sprintf("%s, %s, from %s, %s", r.Often(), r.anchor(), LongDay(from), end)
}

// Said is the rule read back before it has a day to count from: how often, what it counts from, and when it ends if it does.
func (r Rule) Said() string {
	said := r.Often() + ", " + r.anchor()
	if end := r.end(); end != "" {
		said += ", " + end
	}
	return said
}

func (r Rule) anchor() string {
	if r.afterDone {
		return "after the last is done"
	}
	return "on the calendar"
}

func (r Rule) end() string {
	if r.times == 1 {
		return "once"
	}
	if r.times > 0 {
		return fmt.Sprintf("%d times", r.times)
	}
	if !r.until.IsZero() {
		return "until " + LongDay(r.until)
	}
	return ""
}

// Then is what comes after an occurrence of a rule counted from the day the last is done: so long after it is done.
func (r Rule) Then() string {
	return r.howOftenCounted() + " after it is done"
}

func (r Rule) howOftenCounted() string {
	if r.every == 1 {
		return "1 " + unitNames[r.unit][0]
	}
	return fmt.Sprintf("%d %s", r.every, unitNames[r.unit][1])
}

// RuleParts is a rule taken apart the way a picker builds it: every so many days, weeks or months, the days of the week, the time, what it counts from and when it ends.
type RuleParts struct {
	Every     int
	Unit      string
	Days      []time.Weekday
	At        string
	AfterDone bool
	Times     int
	Until     time.Time
}

// RuleUnits are the units a rule counts in, as a picker offers them.
var RuleUnits = []string{"days", "weeks", "months"}

// Parts is the rule taken apart.
func (r Rule) Parts() RuleParts {
	return RuleParts{Every: r.every, Unit: unitNames[r.unit][1], Days: slices.Clone(r.days), At: r.at, AfterDone: r.afterDone, Times: r.times, Until: r.until}
}

// Words is the rule the parts make, written as cairn keeps it: days of the week only counted in weeks and on the calendar, a number of times rather than a day when both are said.
func (p RuleParts) Words() string {
	r := Rule{every: max(1, p.Every), unit: inDays, at: p.At, times: p.Times, until: p.Until, afterDone: p.AfterDone}
	for u, names := range unitNames {
		if names[1] == p.Unit {
			r.unit = u
		}
	}
	if r.unit == inWeeks && len(p.Days) > 0 {
		r.days, r.afterDone = slices.Clone(p.Days), false
	}
	if r.times > 0 {
		r.until = time.Time{}
	}
	if checked, err := r.checked(); err == nil {
		r = checked
	}
	return r.Words()
}

// AfterDone says the rule counts from the day the last one was done, rather than on the calendar.
func (r Rule) AfterDone() bool { return r.afterDone }

// Times is how many times it comes back before it ends, 0 when no number of times ends it.
func (r Rule) Times() int { return r.times }

// At is the time of day it is for, written 10:30; nothing when it names none.
func (r Rule) At() string { return r.at }

// First is the first date of the rule from the day it starts: that day itself when it counts from the last one done.
func (r Rule) First(from time.Time) (time.Time, bool) {
	if r.afterDone {
		return from, r.within(from)
	}
	return r.Next(from.AddDate(0, 0, -1), from)
}

// Next is the date that follows a day: on the calendar, the first date of the rule from its start that falls after it; counted from the last one done, so long after that day.
func (r Rule) Next(after, from time.Time) (time.Time, bool) {
	if r.afterDone {
		next := r.add(after)
		return next, r.within(next)
	}
	start := after.AddDate(0, 0, 1)
	if start.Before(from) {
		start = from
	}
	next := r.onOrAfter(start, from)
	return next, r.within(next)
}

// Dates is the next n dates on the calendar after a day; none for a rule counted from the last one done, whose dates are not known before it is.
func (r Rule) Dates(after, from time.Time, n int) []time.Time {
	if r.afterDone {
		return nil
	}
	var out []time.Time
	for len(out) < n {
		next, ok := r.Next(after, from)
		if !ok {
			return out
		}
		out, after = append(out, next), next
	}
	return out
}

func (r Rule) within(d time.Time) bool {
	return r.until.IsZero() || !d.After(r.until)
}

func (r Rule) add(d time.Time) time.Time {
	if r.unit == inWeeks {
		return d.AddDate(0, 0, 7*r.every)
	}
	if r.unit == inMonths {
		return inMonth(d, d.Day(), r.every)
	}
	return d.AddDate(0, 0, r.every)
}

func inMonth(start time.Time, day, months int) time.Time {
	first := time.Date(start.Year(), start.Month()+time.Month(months), 1, 0, 0, 0, 0, time.UTC)
	last := first.AddDate(0, 1, -1).Day()
	return first.AddDate(0, 0, min(day, last)-1)
}

func (r Rule) onOrAfter(start, from time.Time) time.Time {
	if r.unit == inMonths {
		months := upTo(monthsBetween(from, start), r.every)
		if d := inMonth(from, from.Day(), months); !d.Before(start) {
			return d
		}
		return inMonth(from, from.Day(), months+r.every)
	}
	if r.unit == inWeeks && len(r.days) > 0 {
		return r.weekdayOnOrAfter(start, monday(from))
	}
	period := r.every
	if r.unit == inWeeks {
		period = 7 * r.every
	}
	return from.AddDate(0, 0, upTo(daysBetween(from, start), period))
}

func (r Rule) weekdayOnOrAfter(start, base time.Time) time.Time {
	week := upTo(daysBetween(base, monday(start))/7, r.every)
	for _, day := range r.days {
		if d := base.AddDate(0, 0, 7*week+sinceMonday(day)); !d.Before(start) {
			return d
		}
	}
	return base.AddDate(0, 0, 7*(week+r.every)+sinceMonday(r.days[0]))
}

func upTo(n, every int) int {
	return (n + every - 1) / every * every
}

func daysBetween(from, to time.Time) int {
	return int(to.Sub(from).Hours()+12) / 24
}

func monthsBetween(from, to time.Time) int {
	return (to.Year()-from.Year())*12 + int(to.Month()) - int(from.Month())
}

func monday(d time.Time) time.Time {
	return d.AddDate(0, 0, -sinceMonday(d.Weekday()))
}

func sinceMonday(d time.Weekday) int {
	return (int(d) + 6) % 7
}

// DayOf is the calendar day a moment falls on where it was taken, as a day is kept.
func DayOf(moment time.Time) time.Time {
	return time.Date(moment.Year(), moment.Month(), moment.Day(), 0, 0, 0, 0, time.UTC)
}

// ShortDay writes a day the way a person reads one close by: Mon 9 Nov.
func ShortDay(d time.Time) string { return d.Format("Mon 2 Jan") }

// LongDay writes a day with its year: Mon 9 Nov 2026.
func LongDay(d time.Time) string { return d.Format("Mon 2 Jan 2006") }
