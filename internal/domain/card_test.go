package domain

import (
	"errors"
	"strings"
	"testing"
)

func TestCard_laysEveryFactOutInOneColumn(t *testing.T) {
	got := Card(4, "Ship it", []Fact{{"state", "todo"}, {"description", ""}, {"labels", "spec"}})
	want := "#4  Ship it\n\nstate        todo\ndescription  " + Absent + "\nlabels       spec\n"
	if got != want {
		t.Errorf("Card =\n%q\nwant\n%q", got, want)
	}
}

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

func TestRef_writesANumberAsANameFor(t *testing.T) {
	if got := Ref(4); got != "#4" {
		t.Errorf("Ref(4) = %q, want #4", got)
	}
}

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
