package trace

// The two renderings a person reads: one act on a line, a task's history, and a
// period across tasks.

import (
	"strings"
	"testing"
	"time"
)

// rendered is one act as a history shows it. Asserted through Text and not
// through a rendering of a single act: Text is what the product calls, and a
// test on anything else would prove something nobody uses.
func rendered(a Act) string { return strings.TrimRight(Text([]Act{a}), "\n") }

// acted builds an act, naming only what the rendering reads.
func acted(task int, did, field, was, became, through string, hour, minute int) Act {
	var a AsTold
	a.On(task)
	a.Did(did)
	if field == "" {
		a.NothingNamed()
	} else {
		a.Changing(field)
	}
	if was == "" {
		a.WasNothing()
	} else {
		a.Was(was)
	}
	if became == "" {
		a.BecameNothing()
	} else {
		a.Became(became)
	}
	if through == "" {
		a.ByHand()
	} else {
		a.Through(through)
	}
	a.At(time.Date(2026, 3, 3, hour, minute, 0, 0, time.UTC))
	return a.Act()
}

// A change says what it went from and to, and the attribute names the line: a
// line reading "changed  priority" would say the same thing twice.
func TestLine_namesTheAttributeAndBothItsValues(t *testing.T) {
	got := rendered(acted(1, Changed, "priority", "normal", "high", "", 9, 15))
	for _, want := range []string{"priority", "normal", went, "high"} {
		if !strings.Contains(got, want) {
			t.Errorf("the line does not carry %q: %q", want, got)
		}
	}
	if strings.Contains(got, Changed) {
		t.Errorf("the line says %q as well as the attribute: %q", Changed, got)
	}
}

// An attribute that held nothing before says so, because going from nothing to
// something is worth seeing.
func TestLine_showsThatAnAttributeHeldNothingBefore(t *testing.T) {
	got := rendered(acted(1, Changed, "deadline", "", "2026-09-01", "", 9, 15))
	if !strings.Contains(got, "— "+went) {
		t.Errorf("the line does not show it held nothing before: %q", got)
	}
}

// An act about no attribute has no before, so it says its detail plainly.
// Rendering it as "— → …" would claim a previous state that never existed.
func TestLine_claimsNoPreviousStateForAnActAboutNoAttribute(t *testing.T) {
	for _, tc := range []struct {
		what   string
		act    Act
		carry  string
		reject string
	}{
		{"a move", acted(1, Moved, "", "", "first", "", 9, 15), "first", went},
		{"the focus taken", acted(1, FocusTaken, "", "", "", "", 9, 15), FocusTaken, went},
		{"a task deleted", acted(1, Deleted, "", "", "", "", 9, 15), Deleted, went},
	} {
		got := rendered(tc.act)
		if !strings.Contains(got, tc.carry) {
			t.Errorf("%s: the line does not carry %q: %q", tc.what, tc.carry, got)
		}
		if strings.Contains(got, tc.reject) {
			t.Errorf("%s: the line claims a previous state: %q", tc.what, got)
		}
	}
}

// The surface shows only when it was not the terminal: silent in the ordinary
// case, visible in the one that matters.
func TestLine_namesTheSurfaceOnlyWhenItWasNotTheTerminal(t *testing.T) {
	byHand := rendered(acted(1, Changed, "priority", "normal", "high", "", 9, 15))
	if strings.Contains(byHand, "(") {
		t.Errorf("a line done at the terminal names a surface: %q", byHand)
	}
	byAgent := rendered(acted(1, Changed, "priority", "normal", "high", "mcp", 9, 15))
	if !strings.Contains(byAgent, "(mcp)") {
		t.Errorf("a line done by an agent does not name it: %q", byAgent)
	}
}

// What a task was born carrying is laid out under the creation it belongs to,
// indented and without a moment of its own — so the moment is not repeated once
// per attribute, and a reader can tell the birth from what came after.
func TestText_laysTheBirthUnderTheCreation(t *testing.T) {
	got := Text([]Act{
		acted(1, Created, "", "", "", "", 9, 0),
		acted(1, Created, "title", "", "Write the spec", "", 9, 0),
		acted(1, Created, "column", "", "backlog", "", 9, 0),
		acted(1, Changed, "priority", "normal", "high", "", 9, 30),
	})
	lines := strings.Split(strings.TrimRight(got, "\n"), "\n")
	if len(lines) != 4 {
		t.Fatalf("%d lines, want 4:\n%s", len(lines), got)
	}
	if strings.HasPrefix(lines[0], " ") {
		t.Errorf("the creation is indented: %q", lines[0])
	}
	for _, i := range []int{1, 2} {
		if !strings.HasPrefix(lines[i], " ") {
			t.Errorf("what it was born carrying is not indented: %q", lines[i])
		}
		if strings.Contains(lines[i], "09:00") {
			t.Errorf("a birth line repeats the creation's moment: %q", lines[i])
		}
	}
	if strings.HasPrefix(lines[3], " ") {
		t.Errorf("what came after the birth is indented: %q", lines[3])
	}
}

// A period names the task each act was on, since it reaches across them.
func TestList_namesTheTaskOfEveryAct(t *testing.T) {
	got := List([]Act{
		acted(3, FocusTaken, "", "", "", "", 9, 0),
		acted(3, FocusReleased, "", "", "to task 7", "", 9, 30),
		acted(7, FocusTaken, "", "", "", "", 9, 30),
	})
	for _, want := range []string{"#3", "#7", "to task 7"} {
		if !strings.Contains(got, want) {
			t.Errorf("the period does not carry %q:\n%s", want, got)
		}
	}
}

// What a task was born carrying is left out of a period: those acts belong to
// that task's own history, and a day is unreadable with four lines of birth for
// every task created in it.
func TestList_leavesTheBirthOutOfAPeriod(t *testing.T) {
	got := List([]Act{
		acted(1, Created, "", "", "", "", 9, 0),
		acted(1, Created, "title", "", "Write the spec", "", 9, 0),
		acted(1, Created, "column", "", "backlog", "", 9, 0),
	})
	lines := strings.Split(strings.TrimRight(got, "\n"), "\n")
	if len(lines) != 1 {
		t.Errorf("%d lines, want just the creation:\n%s", len(lines), got)
	}
	if strings.Contains(got, "Write the spec") {
		t.Errorf("a period carries what a task was born with:\n%s", got)
	}
}

func TestList_saysNothingWasDoneWhenNothingWas(t *testing.T) {
	if got := List(nil); !strings.Contains(got, "nothing was done") {
		t.Errorf("List(nil) = %q, want it to say nothing was done", got)
	}
}

// A period holding nothing but a birth has nothing to show either: the birth
// belongs to a history, so leaving it out leaves nothing.
func TestList_saysNothingWasDoneWhenOnlyABirthIsLeftOut(t *testing.T) {
	got := List([]Act{acted(1, Created, "title", "", "Write the spec", "", 9, 0)})
	if !strings.Contains(got, "nothing was done") {
		t.Errorf("List = %q, want it to say nothing was done", got)
	}
}

// An attribute act that carries neither value shows the attribute and nothing
// else. It cannot arise from an edit — giving an attribute the value it already
// holds is not a change — but the rendering has to say something, and saying
// "— → —" would claim two states where there is no information at all.
func TestLine_showsOnlyTheAttributeWhenNeitherValueIsThere(t *testing.T) {
	got := rendered(acted(1, Changed, "priority", "", "", "", 9, 15))
	if !strings.Contains(got, "priority") {
		t.Errorf("the line does not name the attribute: %q", got)
	}
	if strings.Contains(got, went) || strings.Contains(got, "—") {
		t.Errorf("the line claims a value it does not have: %q", got)
	}
}

// A deletion carries what the task held, and a history that printed only the
// word "deleted" kept it where nobody could read it.
func TestAHistorySaysWhatADeletionTookWithIt(t *testing.T) {
	var a AsTold
	a.On(1)
	a.Did(Deleted)
	a.NothingNamed()
	a.Was("title: Write the spec; priority: high")
	a.BecameNothing()
	a.ByHand()
	a.At(time.Date(2026, 10, 1, 9, 0, 0, 0, time.UTC))
	said := Text([]Act{a.Act()})
	for _, want := range []string{"deleted", "Write the spec", "high"} {
		if !strings.Contains(said, want) {
			t.Errorf("the history does not say %q:\n%s", want, said)
		}
	}
}

// A value written over several lines, changed by an agent, leaves the columns
// of every act after it where they were. The surface used to be a column of its
// own, so the last line of such a value ended in a tab, became a cell, and
// widened the first column of every line after it — measured on a real store:
// the moments of a whole afternoon pushed seventy characters to the right.
func TestList_aValueOverSeveralLinesLeavesTheColumnsAfterIt(t *testing.T) {
	said := "first line\nthe last line of the description, far longer than any moment is"
	for name, render := range map[string]func([]Act) string{"a period": List, "a history": Text} {
		got := render([]Act{
			acted(3, Changed, "description", "", said, "mcp", 9, 0),
			acted(3, Changed, "priority", "normal", "high", "", 9, 30),
		})
		lines := strings.Split(got, "\n")
		last := lines[len(lines)-2]
		if !strings.Contains(got, "(mcp)") {
			t.Errorf("%s no longer says the change came through the server:\n%s", name, got)
		}
		if gap := strings.Index(last, "  ") + 2; strings.TrimLeft(last[gap:], " ") != last[gap:] {
			t.Errorf("%s pushed its next line to the right:\n%s", name, got)
		}
	}
}

// Time added by hand says how much, and for which day: the day it was added on
// is the moment of the line, and is not the day it counts for.
func TestText_saysTimeAddedForTheDayItCountsFor(t *testing.T) {
	got := rendered(acted(1, TimeAdded, "", "2026-09-29", "1h", "", 9, 0))
	if !strings.Contains(got, "time added") || !strings.Contains(got, "1h for 2026-09-29") {
		t.Errorf("time added reads %q", got)
	}
}

// A sheet writes each day, then each task's time before its title — so a title
// holding "1h" is never read as one — then what the day came to; and says when
// nothing was in focus at all.
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

// What a task was born with is left out, and what was done to it afterwards
// is not: the listing reads on past a birth.
func TestList_readsOnPastABirth(t *testing.T) {
	got := List([]Act{
		acted(1, Created, "", "", "", "", 9, 0),
		acted(1, Created, "title", "", "Write the spec", "", 9, 0),
		acted(1, Changed, "priority", "normal", "high", "", 10, 0),
	})
	if !strings.Contains(got, "high") {
		t.Errorf("what was done after the birth is gone:\n%s", got)
	}
}
