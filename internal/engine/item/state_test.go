package item

import (
	"slices"
	"strings"
	"testing"
)

func TestStates_areExactlyWhatParseStateAccepts(t *testing.T) {
	for _, name := range []string{"backlog", "todo", "current", "done"} {
		if !slices.Contains(States, State(name)) {
			t.Errorf("States is missing %q, though ParseState accepts it", name)
		}
	}
	for _, s := range States {
		if _, err := ParseState(string(s)); err != nil {
			t.Errorf("ParseState(%q): %v, want every entry of States accepted", s, err)
		}
	}
	if slices.Contains(States, State("someday")) {
		t.Error("States should not hold someday: ParseState refuses it")
	}
}

func TestPriorities_areExactlyWhatParsePriorityAccepts(t *testing.T) {
	for _, name := range []string{"low", "normal", "high", "critical"} {
		if !slices.Contains(Priorities, Priority(name)) {
			t.Errorf("Priorities is missing %q, though ParsePriority accepts it", name)
		}
	}
	for _, p := range Priorities {
		if _, err := ParsePriority(string(p)); err != nil {
			t.Errorf("ParsePriority(%q): %v, want every entry of Priorities accepted", p, err)
		}
	}
	if slices.Contains(Priorities, Priority("urgent")) {
		t.Error("Priorities should not hold urgent: ParsePriority refuses it")
	}
}

func TestParseState_acceptsTheFourColumns(t *testing.T) {
	for _, name := range []string{"backlog", "todo", "current", "done"} {
		got, err := ParseState(name)
		if err != nil {
			t.Errorf("ParseState(%q): %v", name, err)
			continue
		}
		if string(got) != name {
			t.Errorf("ParseState(%q) = %q", name, got)
		}
	}
}

func TestParseState_refusesAnythingElse(t *testing.T) {
	if _, err := ParseState("someday"); err == nil {
		t.Error("ParseState(someday) should fail: it is not one of the columns")
	}
}

func TestParseState_readsAnEmptyValueAsTheBacklog(t *testing.T) {
	got, err := ParseState("")
	if err != nil || got != Backlog {
		t.Errorf("ParseState(\"\") = %q, %v; want the backlog", got, err)
	}
}

func TestParsePriority_acceptsTheFourPriorities(t *testing.T) {
	for _, name := range []string{"low", "normal", "high", "critical"} {
		got, err := ParsePriority(name)
		if err != nil {
			t.Errorf("ParsePriority(%q): %v", name, err)
			continue
		}
		if string(got) != name {
			t.Errorf("ParsePriority(%q) = %q", name, got)
		}
	}
}

func TestParsePriority_refusesAnythingElse(t *testing.T) {
	if _, err := ParsePriority("urgent"); err == nil {
		t.Error("ParsePriority(urgent) should fail: it is not one of the priorities")
	}
}

func TestParsePriority_readsAnEmptyValueAsNormal(t *testing.T) {
	got, err := ParsePriority("")
	if err != nil || got != Normal {
		t.Errorf("ParsePriority(\"\") = %q, %v; want normal", got, err)
	}
}

func TestARefusalSaysWhatWouldHaveBeenTaken(t *testing.T) {
	for _, c := range []struct {
		what string
		err  error
		says []string
	}{
		{"state", refusedBy(ParseState("nope")), []string{"backlog", "todo", "current", "done"}},
		{"priority", refusedBy(ParsePriority("nope")), []string{"low", "normal", "high", "critical"}},
	} {
		t.Run(c.what, func(t *testing.T) {
			if c.err == nil {
				t.Fatal("nothing was refused")
			}
			for _, word := range c.says {
				if !strings.Contains(c.err.Error(), word) {
					t.Errorf("%q does not say %q is taken", c.err, word)
				}
			}
		})
	}
}

func refusedBy[T any](_ T, err error) error { return err }
