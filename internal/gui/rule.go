package gui

import (
	"cairn/internal/board"
	"cairn/internal/engine/calendar"
	"cairn/internal/engine/series"
	"cairn/internal/gui/geom"
	"slices"
	"strconv"
	"time"
)

type ruleState struct {
	parts series.RuleParts
	line  int
	day   int
	was   string
	from  string
	since string
}

const (
	ruleEvery = iota
	ruleUnit
	ruleOn
	ruleAt
	ruleFrom
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
	ruleDaysAhead  = 366
	ruleAnyTime    = "any time"
	ruleNoEnd      = "—"
)

func (w Window) openingRule() Window {
	field := w.Form.Fields[w.Form.On()]
	from := ""
	for _, f := range w.Form.Fields {
		if f.Name == board.FromField().Name {
			from = f.Value
		}
	}
	w.ruling = ruleState{parts: series.RuleParts{Every: 1, Unit: "days"}, was: field.Value, from: from, since: from}
	if rule, err := series.ParseRule(field.Value); err == nil {
		w.ruling.parts = rule.Parts()
	}
	w.Screen = board.OnRule
	return w
}

func (w Window) closingRule(keep bool) Window {
	if !keep {
		w.Form = w.Form.Writes(board.Ending(w.ruling.was))
		w = w.fromSet(w.ruling.since)
	}
	w.Screen = board.OnForm
	return w
}

func (w Window) fromSet(day string) Window {
	if !slices.ContainsFunc(w.Form.Fields, func(f board.Field) bool { return f.Name == board.FromField().Name }) {
		return w
	}
	at := w.Form.At
	w.Form = w.Form.Set(board.FromField().Name, day)
	w.Form.At = at
	return w
}

func fromDays(held string, today time.Time) []string {
	out := make([]string, 0, ruleDaysAhead+1)
	for n := range ruleDaysAhead {
		out = append(out, calendar.FormatDay(series.DayOf(today).AddDate(0, 0, n)))
	}
	if held != "" && !slices.Contains(out, held) {
		out = append(out, held)
		slices.Sort(out)
	}
	return out
}

func ruleChoices(line int, r ruleState, today time.Time) []string {
	p := r.parts
	switch line {
	case ruleFrom:
		return fromDays(r.from, today)
	case ruleEvery:
		return counted(1, ruleMost, "")
	case ruleUnit:
		return series.RuleUnits
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

func endsOf(p series.RuleParts, today time.Time) []string {
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
		out = append(out, calendar.FormatDay(series.DayOf(today).AddDate(0, 0, 7*week)))
	}
	if said := calendar.FormatDay(held); !slices.Contains(out, said) {
		out = append(out, said)
		slices.Sort(out)
	}
	return out
}

func ruleOnOf(line int, r ruleState, today time.Time) int {
	p := r.parts
	switch line {
	case ruleFrom:
		return max(0, slices.Index(fromDays(r.from, today), r.from))
	case ruleEvery:
		return p.Every - 1
	case ruleUnit:
		return slices.Index(series.RuleUnits, p.Unit)
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
			return slices.Index(untilDays(p.Until, today), calendar.FormatDay(p.Until))
		}
		return 0
	}
	if p.AfterDone {
		return 1
	}
	return 0
}

func ticked(p series.RuleParts) []bool {
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
	choices := ruleChoices(line, w.ruling, w.Today())
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
	case ruleFrom:
		w.ruling.from, w.ruling.line = choices[at], line
		return w.fromSet(choices[at])
	case ruleEnds:
		p.Times, p.Until = 0, time.Time{}
		if at == 1 {
			p.Times = ruleTimesFirst
		}
		if at == 2 {
			p.Until = series.DayOf(w.Today()).AddDate(0, 0, 7*ruleUntilFirst)
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
	w.Form = w.Form.Writes(board.Ending(p.Words()))
	return w
}

func dayTicked(p series.RuleParts, day time.Weekday) series.RuleParts {
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

func endAt(p series.RuleParts, said string, at int) series.RuleParts {
	switch {
	case p.Times > 0:
		p.Times = at + 1
	case !p.Until.IsZero():
		if day, err := calendar.ParseDeadline(said); err == nil {
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
		return w.ruled(line, ruleOnOf(line, w.ruling, w.Today())+by)
	case "space":
		if line == ruleOn {
			return w.ruled(ruleOn, w.ruling.day)
		}
	}
	return w
}

func (w Window) clickingRule(p geom.Point) Window {
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
	today := w.Today()
	lines := make([]RuleLine, ruleComes+1)
	for i := range lines {
		lines[i].Choices = ruleChoices(i, w.ruling, today)
		lines[i].On = ruleOnOf(i, w.ruling, today)
	}
	lines[ruleOn].Ticked = ticked(w.ruling.parts)
	return RulePickerFor(w.ruleHung(), lines)
}

func (w Window) ruleHung() geom.Point {
	field := fieldFrame(w)
	if read, ok := RepeatsReadAt(w.formBox(), w.Form.Fields); ok && read.Y > field.Y {
		field.H = read.Y + read.H - field.Y
	}
	return Hung(field, geom.Box{X: geom.Margin, Y: geom.Margin, W: w.width - 2*geom.Margin, H: w.height - 2*geom.Margin}, ruleWide, RulePickerHigh())
}
