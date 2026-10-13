package domain

import (
	"errors"
	"strings"
	"testing"
)

// A card is a number and a title, then one fact a line, its values in one
// column and an absence shown as one. A task's card and an objective's are
// laid out by this and nothing else; each used to lay its own the same way.
func TestCard_laysEveryFactOutInOneColumn(t *testing.T) {
	got := Card(4, "Ship it", []Fact{{"state", "todo"}, {"description", ""}, {"labels", "spec"}})
	want := "#4  Ship it\n\nstate        todo\ndescription  " + Absent + "\nlabels       spec\n"
	if got != want {
		t.Errorf("Card =\n%q\nwant\n%q", got, want)
	}
}

// How far along something is, said of whatever it is, and nothing at all when
// nothing is under it.
func TestProgress_saysHowFarAlongWhatItIsAbout(t *testing.T) {
	for _, tc := range []struct {
		p    Progress
		want string
	}{
		{Progress{}, ""},
		{Progress{SubTasks: Tally{Total: 4, Done: 1}}, "1 of 4 sub-tasks are done — the task is 25% done"},
		{Progress{Tasks: Tally{Total: 2, Done: 1}, SubTasks: Tally{Total: 2, Done: 1}},
			"1 of 2 tasks and 1 of 2 sub-tasks are done — the task is 50% done"},
	} {
		if got := tc.p.Line("the task"); got != tc.want {
			t.Errorf("Line = %q, want %q", got, tc.want)
		}
	}
	if got := (Progress{Tasks: Tally{Total: 1}}).Line("the objective"); got != "0 of 1 tasks are done — the objective is 0% done" {
		t.Errorf("Line = %q, want it about the objective", got)
	}
}

// A title is trimmed, held to one line, and refused when it says nothing — the
// same rule for whatever carries it, named by what carries it.
func TestTitle_isOneLineOfTextAboutWhatCarriesIt(t *testing.T) {
	if got, err := Title("a task", "  Ship it  "); err != nil || got != "Ship it" {
		t.Errorf("Title = %q, %v", got, err)
	}
	if _, err := Title("an objective", "   "); !errors.Is(err, ErrTitleRequired) || !strings.Contains(err.Error(), "an objective") {
		t.Errorf("Title of nothing = %v, want ErrTitleRequired naming an objective", err)
	}
	if _, err := Title("a task", "two\nlines"); !errors.Is(err, ErrTitleNotOneLine) {
		t.Errorf("Title of two lines = %v, want ErrTitleNotOneLine", err)
	}
}

// Looking for things in what something carries: each named thing found once,
// anywhere, without regard to case; a thing named as nothing never found.
func TestLooking_findsEveryThingNamedAnywhere(t *testing.T) {
	l := LookingFor([]string{"Spec", "team"})
	l.Consider("Write the SPEC")
	if l.All() {
		t.Error("all found with one of two seen")
	}
	l.Consider("for the team")
	if !l.All() {
		t.Error("not all found with both seen")
	}
	empty := LookingFor([]string{""})
	empty.Consider("anything at all")
	if empty.All() {
		t.Error("a thing named as nothing was found")
	}
}

// A number written where a person reads it as naming something: #4. Written
// four times over before, in three packages.
func TestRef_writesANumberAsANameFor(t *testing.T) {
	if got := Ref(4); got != "#4" {
		t.Errorf("Ref(4) = %q, want #4", got)
	}
}

// Every named thing is looked for in every text considered, whatever was found
// before it: a search that stopped at the first thing already found would
// leave the rest unlooked-for, depending on the order a map is walked in.
func TestLooking_looksForEverythingInEveryText(t *testing.T) {
	named := []string{"a1", "b2", "c3", "d4", "e5", "f6", "g7", "h8"}
	l := LookingFor(named)
	for _, n := range named {
		l.Consider("here is " + n)
	}
	if !l.All() {
		t.Error("a thing named was left unfound after a text holding it was considered")
	}
}

// Reaches is the one rule a search keeps: every thing named sits somewhere in
// what was said, case aside, and nothing named reaches nothing.
func TestReaches_isEveryThingNamedSomewhereInWhatIsSaid(t *testing.T) {
	said := []string{"Write the SPEC", "team"}
	if !Reaches(said, []string{"spec", "Team"}) {
		t.Error("two things said do not reach")
	}
	if Reaches(said, []string{"spec", "lamp"}) {
		t.Error("a thing not said reached")
	}
	if Reaches(said, nil) {
		t.Error("nothing named reached")
	}
}
