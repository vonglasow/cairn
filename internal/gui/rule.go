package gui

import (
	"slices"
	"strconv"
	"time"

	"cairn/internal/domain"
)

type ruleState struct {
	parts domain.RuleParts
	line  int
	day   int
	was   string
}

const (
	ruleEvery = iota
	ruleUnit
	ruleOn
	ruleAt
	ruleEnds
	ruleEnd
	ruleComes
)

var (
	endChoices  = []string{"never", "after", "on a day"}
	ruleAnchors = []string{"on the calendar", "after the last is done"}
)

const (
	ruleMost       = 99
	ruleTimesFirst = 10
	ruleWeeksAhead = 104
	ruleUntilFirst = 4
	ruleAnyTime    = "any time"
	ruleNoEnd      = "—"
)

func (w Window) openingRule() Window {
	field := w.form.Fields()[w.form.On()]
	w.ruling = ruleState{parts: domain.RuleParts{Every: 1, Unit: "days"}, was: field.Value}
	if rule, err := domain.ParseRule(field.Value); err == nil {
		w.ruling.parts = rule.Parts()
	}
	w.screen = OnRule
	return w
}

func (w Window) closingRule(keep bool) Window {
	if !keep {
		w.form = w.form.Writes(Ending(w.ruling.was))
	}
	w.screen = OnForm
	return w
}

func ruleChoices(line int, p domain.RuleParts, today time.Time) []string {
	switch line {
	case ruleEvery:
		return counted(1, ruleMost, "")
	case ruleUnit:
		return domain.RuleUnits
	case ruleOn:
		return Weekdays[:]
	case ruleAt:
		return timesOfDay(p.At)
	case ruleEnds:
		return endChoices
	case ruleEnd:
		return endsOf(p, today)
	}
	return ruleAnchors
}

func counted(from, to int, after string) []string {
	out := make([]string, 0, to-from+1)
	for n := from; n <= to; n++ {
		out = append(out, strconv.Itoa(n)+after)
	}
	return out
}

func timesOfDay(held string) []string {
	out := []string{ruleAnyTime}
	for half := range 48 {
		out = append(out, time.Date(0, 1, 1, half/2, half%2*30, 0, 0, time.UTC).Format("15:04"))
	}
	if held != "" && !slices.Contains(out, held) {
		out = append(out, held)
		slices.Sort(out[1:])
	}
	return out
}

func endsOf(p domain.RuleParts, today time.Time) []string {
	switch {
	case p.Times > 0:
		return counted(1, ruleMost, " times")
	case !p.Until.IsZero():
		return untilDays(p.Until, today)
	}
	return []string{ruleNoEnd}
}

func untilDays(held, today time.Time) []string {
	var out []string
	for week := 1; week <= ruleWeeksAhead; week++ {
		out = append(out, domain.FormatDay(domain.DayOf(today).AddDate(0, 0, 7*week)))
	}
	if said := domain.FormatDay(held); !slices.Contains(out, said) {
		out = append(out, said)
		slices.Sort(out)
	}
	return out
}

func ruleOnOf(line int, p domain.RuleParts, today time.Time) int {
	switch line {
	case ruleEvery:
		return p.Every - 1
	case ruleUnit:
		return slices.Index(domain.RuleUnits, p.Unit)
	case ruleOn:
		return -1
	case ruleAt:
		if p.At == "" {
			return 0
		}
		return slices.Index(timesOfDay(p.At), p.At)
	case ruleEnds:
		switch {
		case p.Times > 0:
			return 1
		case !p.Until.IsZero():
			return 2
		}
		return 0
	case ruleEnd:
		switch {
		case p.Times > 0:
			return p.Times - 1
		case !p.Until.IsZero():
			return slices.Index(untilDays(p.Until, today), domain.FormatDay(p.Until))
		}
		return 0
	}
	if p.AfterDone {
		return 1
	}
	return 0
}

func ticked(p domain.RuleParts) []bool {
	out := make([]bool, len(Weekdays))
	if p.Unit != "weeks" {
		return out
	}
	for _, d := range p.Days {
		out[(int(d)+6)%7] = true
	}
	return out
}

func (w Window) ruled(line, at int) Window {
	choices := ruleChoices(line, w.ruling.parts, w.today())
	if at < 0 || at >= len(choices) {
		return w
	}
	p := w.ruling.parts
	switch line {
	case ruleEvery:
		p.Every = at + 1
	case ruleUnit:
		p.Unit = choices[at]
	case ruleOn:
		p = dayTicked(p, time.Weekday((at+1)%7))
		w.ruling.day = at
	case ruleAt:
		p.At = choices[at]
		if at == 0 {
			p.At = ""
		}
	case ruleEnds:
		p.Times, p.Until = 0, time.Time{}
		if at == 1 {
			p.Times = ruleTimesFirst
		}
		if at == 2 {
			p.Until = domain.DayOf(w.today()).AddDate(0, 0, 7*ruleUntilFirst)
		}
	case ruleEnd:
		p = endAt(p, choices[at], at)
	case ruleComes:
		p.AfterDone = at == 1
	}
	if p.Unit != "weeks" || p.AfterDone {
		p.Days = nil
	}
	w.ruling.parts, w.ruling.line = p, line
	w.form = w.form.Writes(Ending(p.Words()))
	return w
}

func dayTicked(p domain.RuleParts, day time.Weekday) domain.RuleParts {
	if p.Unit != "weeks" {
		p.Days = nil
	}
	p.Unit, p.AfterDone = "weeks", false
	if at := slices.Index(p.Days, day); at >= 0 {
		p.Days = slices.Delete(slices.Clone(p.Days), at, at+1)
		return p
	}
	p.Days = append(slices.Clone(p.Days), day)
	return p
}

func endAt(p domain.RuleParts, said string, at int) domain.RuleParts {
	switch {
	case p.Times > 0:
		p.Times = at + 1
	case !p.Until.IsZero():
		if day, err := domain.ParseDeadline(said); err == nil {
			p.Until = day
		}
	}
	return p
}

func (w Window) onRule(name string) Window {
	line := w.ruling.line
	switch name {
	case "down", "j":
		w.ruling.line = min(ruleComes, line+1)
	case "up", "k":
		w.ruling.line = max(0, line-1)
	case "left", "h", "right", "l":
		by := 1
		if name == "left" || name == "h" {
			by = -1
		}
		if line == ruleOn {
			w.ruling.day = (w.ruling.day + by + len(Weekdays)) % len(Weekdays)
			return w
		}
		return w.ruled(line, ruleOnOf(line, w.ruling.parts, w.today())+by)
	case "space":
		if line == ruleOn {
			return w.ruled(ruleOn, w.ruling.day)
		}
	}
	return w
}

func (w Window) clickingRule(p Point) Window {
	picker := w.RulePicked()
	if !picker.Box.In(p) {
		return w.clickedOff(p)
	}
	for i, line := range picker.Lines {
		switch {
		case line.Back.In(p):
			return w.ruled(i, line.On-1)
		case line.Next.In(p):
			return w.ruled(i, line.On+1)
		}
		for j, pick := range line.Picks {
			if pick.In(p) && !line.Compact {
				return w.ruled(i, j)
			}
		}
	}
	return w
}

// RulePicked is the picker as it stands: each part's choices and the one chosen, laid out under what the field reads back.
func (w Window) RulePicked() RulePicker {
	p, today := w.ruling.parts, w.today()
	lines := make([]RuleLine, ruleComes+1)
	for i := range lines {
		lines[i].Choices = ruleChoices(i, p, today)
		lines[i].On = ruleOnOf(i, p, today)
	}
	lines[ruleOn].Ticked = ticked(p)
	return RulePickerFor(w.ruleHung(), lines)
}

func (w Window) ruleHung() Point {
	field := fieldFrame(w)
	if read, ok := RepeatsReadAt(w.formBox(), w.form.Fields()); ok && read.Y > field.Y {
		field.H = read.Y + read.H - field.Y
	}
	return Hung(field, Over(w.width, w.height), ruleWide, RulePickerHigh())
}
