package engine

import (
	"slices"
	"testing"
	"time"
)

func TestLook_readsTheSeriesAndMarksTheirOccurrences(t *testing.T) {
	e := openEngine(t)
	if _, err := e.Ask(TaskNew, Input{Given: map[string]string{"title": "Stand-up", "every": "every day"}}); err != nil {
		t.Fatal(err)
	}
	if _, err := e.Ask(TaskSkip, Numbered(1)); err != nil {
		t.Fatal(err)
	}
	if _, err := e.Ask(TaskNew, Input{Given: map[string]string{"title": "Write the spec"}}); err != nil {
		t.Fatal(err)
	}
	s := lookAt(t, e)
	one, found := s.SeriesOf(1)
	if !found || one.Title != "Stand-up" || len(one.Occurrences) != 2 || !slices.Equal(one.Open, []int{2}) {
		t.Fatalf("the series read %+v (%v)", one, found)
	}
	if _, found := s.SeriesOf(9); found {
		t.Error("a series that is not there was found")
	}
	for _, c := range s.Cards {
		switch c.Number {
		case 1, 2:
			if c.Series == nil || c.Series.Series != 1 || c.Series.Rule != "every day" {
				t.Errorf("card %d carries %+v", c.Number, c.Series)
			}
		default:
			if c.Series != nil {
				t.Errorf("card %d carries a series", c.Number)
			}
		}
	}
}

func TestLook_readsNoSeriesWhileRecurrenceIsOff(t *testing.T) {
	f := newFake()
	f.refuse = map[string]error{"series list": ErrRecurrenceOff}
	found, err := Look(f)
	if err != nil || found.Series != nil {
		t.Errorf("Look read %v (%v), want no series and no refusal", found.Series, err)
	}
	f.refuse = map[string]error{"series list": errNope}
	if _, err := Look(f); err == nil {
		t.Error("a store that does not answer was passed over")
	}
}

func TestOccurrenceDay_saysTheDayOfAnOccurrenceWithinAWeek(t *testing.T) {
	now := time.Date(2026, 11, 9, 11, 0, 0, 0, time.Local)
	for _, c := range []struct {
		day, at, state string
		text           string
		show, late     bool
		passed         bool
	}{
		{"2026-11-09", "10:30", "todo", "10:30", true, false, true},
		{"2026-11-09", "14:00", "todo", "14:00", true, false, false},
		{"2026-11-09", "", "todo", "today", true, false, false},
		{"2026-11-12", "", "todo", "Thu 12", true, false, false},
		{"2026-11-16", "", "todo", "Mon 16", true, false, false},
		{"2026-11-17", "", "todo", "", false, false, false},
		{"2026-11-06", "", "todo", "−3d", true, true, false},
		{"2026-11-09", "10:30", "done", "", false, false, false},
	} {
		card := Card{State: c.state, Series: &Occurs{Day: c.day, At: c.at}}
		text, show, late, passed := OccurrenceDay(card, now)
		if text != c.text || show != c.show || late != c.late || passed != c.passed {
			t.Errorf("%s %s %s reads %q %v %v %v, want %q %v %v %v", c.day, c.at, c.state, text, show, late, passed, c.text, c.show, c.late, c.passed)
		}
	}
	if _, show, _, _ := OccurrenceDay(Card{State: "todo"}, now); show {
		t.Error("a card of no series says a day")
	}
}

func TestSeriesFields_areWhatTheNextIsMadeFromAndItsRule(t *testing.T) {
	s := Series{Number: 3, Title: "Stand-up", Rule: "every weekday 10:30", From: "2026-11-09",
		Template: Template{Title: "Stand-up", Priority: "high", Labels: []string{"team", "work"}, Steps: []string{"Yesterday", "Today"}, Estimate: "15m", Objectives: []int{2}}}
	fields := SeriesFields(s)
	got := map[string]string{}
	for _, f := range fields {
		got[f.Name] = f.Value
		if f.Value != f.Was {
			t.Errorf("%s opens changed", f.Name)
		}
	}
	for name, want := range map[string]string{"title": "Stand-up", "priority": "high", "labels": "team work", "estimate": "15m",
		"objectives": "2", "repeats": "every weekday 10:30", "from": "2026-11-09", "steps": "[ ] Yesterday\n[ ] Today"} {
		if got[name] != want {
			t.Errorf("the field %s holds %q, want %q", name, got[name], want)
		}
	}
	op, in := Saved(fields, false, 3, TaskNew, SeriesEdit)
	if op.Thing != "series" || in.Given["number"] != "3" || in.Given["title"] != "Stand-up" {
		t.Errorf("an untouched form asks %s %v", op.Line(), in.Given)
	}
	if _, named := in.Given["every"]; named {
		t.Error("an untouched rule was sent")
	}
}

func TestSaved_sendsNoRuleLeftEmpty(t *testing.T) {
	fields := append(AddingFields("todo"), RepeatsField())
	for i := range fields {
		if fields[i].Name == "title" {
			fields[i].Value = "Water"
		}
	}
	_, in := Saved(fields, true, 0, TaskNew, TaskEdit)
	if _, named := in.Given["every"]; named {
		t.Errorf("an empty rule was sent: %v", in.Given)
	}
	fields[len(fields)-1].Value = "every 3 days"
	if _, in := Saved(fields, true, 0, TaskNew, TaskEdit); in.Given["every"] != "every 3 days" {
		t.Errorf("the rule typed was not sent: %v", in.Given)
	}
	if f := RepeatsField(); f.Name != "repeats" || f.Arg != "every" || f.Means == "" {
		t.Errorf("the field reads %+v", f)
	}
}

func TestFoundSeries_keepsWhatCarriesEveryWord(t *testing.T) {
	all := []Series{{Number: 1, Title: "Stand-up", Rule: "every day"}, {Number: 2, Title: "Expense report", Rule: "every 4 weeks on mon"}}
	if got := FoundSeries(all, "expense"); len(got) != 1 || got[0].Number != 2 {
		t.Errorf("expense found %v", got)
	}
	if got := FoundSeries(all, ""); len(got) != 2 {
		t.Errorf("nothing typed found %v", got)
	}
	if got := FoundSeries(all, "↻1"); len(got) != 1 || got[0].Number != 1 {
		t.Errorf("↻1 found %v", got)
	}
}

func TestPages_withoutRecurrenceOffersNoSeries(t *testing.T) {
	p := Pages{}.WithoutRecurrence()
	if p.Recurs() || p.Offers(SeriesList) {
		t.Error("recurrence turned off still offers the series")
	}
}
