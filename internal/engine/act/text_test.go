package act

import (
	"strings"
	"testing"
	"time"
)

func acted(task int, did, field, was, became, through string, hour, minute int) Act {
	var a AsTold
	a.On(task)
	a.Did(did)
	if field == "" {
	} else {
		a.Changing(field)
	}
	if was == "" {
	} else {
		a.Was(was)
	}
	if became == "" {
	} else {
		a.Became(became)
	}
	if through == "" {
	} else {
		a.Through(through)
	}
	a.At(time.Date(2026, 3, 3, hour, minute, 0, 0, time.UTC))
	return a.Act()
}

func TestList_saysNothingWasDoneWhenOnlyABirthIsLeftOut(t *testing.T) {
	got := List([]Act{acted(1, Created, "title", "", "Write the spec", "", 9, 0)})
	if !strings.Contains(got, "nothing was done") {
		t.Errorf("List = %q, want it to say nothing was done", got)
	}
}

func TestAHistorySaysWhatADeletionTookWithIt(t *testing.T) {
	var a AsTold
	a.On(1)
	a.Did(Deleted)
	a.Was("title: Write the spec; priority: high")
	a.At(time.Date(2026, 10, 1, 9, 0, 0, 0, time.UTC))
	said := Text([]Act{a.Act()})
	for _, want := range []string{"deleted", "Write the spec", "high"} {
		if !strings.Contains(said, want) {
			t.Errorf("the history does not say %q:\n%s", want, said)
		}
	}
}

func TestList_aValueOverSeveralLinesLeavesTheColumnsAfterIt(t *testing.T) {
	said := "first line\nthe last line of the description, far longer than any moment is"
	for name, render := range map[string]func([]Act) string{"a period": List, "a history": Text} {
		got := render([]Act{
			acted(3, Changed, "description", "", said, "mcp", 9, 0),
			acted(3, Changed, "priority", "normal", "high", "", 9, 30),
		})
		lines := strings.Split(got, "\n")
		last := lines[len(lines)-2]
		if !strings.Contains(got, "(through mcp)") {
			t.Errorf("%s no longer says the change came through the server:\n%s", name, got)
		}
		if gap := strings.Index(last, "  ") + 2; strings.TrimLeft(last[gap:], " ") != last[gap:] {
			t.Errorf("%s pushed its next line to the right:\n%s", name, got)
		}
	}
}

func TestSheetText_writesEachDayTaskByTask(t *testing.T) {
	day := time.Date(2026, 9, 29, 0, 0, 0, 0, time.UTC)
	got := SheetText(Sheet{Days: []Day{{Day: day, Total: 8*time.Hour + 30*time.Minute, Tasks: []Worked{
		{Task: 1, Took: 90 * time.Minute},
		{Task: 19, Took: 7 * time.Hour, LeftOpen: true},
	}}}}, map[int]string{1: "Write the spec", 19: "Stabilise"})
	for _, want := range []string{"2026-09-29\n", "  #1   1h30  Write the spec\n",
		"  #19  7h    Stabilise — still in focus at midnight\n", "  total  8h30\n"} {
		if !strings.Contains(got, want) {
			t.Errorf("the sheet does not read %q:\n%s", want, got)
		}
	}
	two := SheetText(Sheet{Days: []Day{{Day: day}, {Day: day.AddDate(0, 0, 1)}}}, nil)
	if want := "2026-09-29\n  total  0m\n\n2026-09-30\n  total  0m\n"; two != want {
		t.Errorf("two days read %q, want %q", two, want)
	}
	if got := SheetText(Sheet{}, nil); got != "nothing was in focus\n" {
		t.Errorf("an empty sheet reads %q", got)
	}
}

func TestText_saysTimeAddedWithNoDayWithoutOne(t *testing.T) {
	line := Text([]Act{added(1, "", "2h30m", at(9, 0))})
	if !strings.Contains(line, "time added  2h30m") || strings.Contains(line, " for") {
		t.Errorf("time added with no day reads %q", line)
	}
}
