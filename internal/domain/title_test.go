package domain

import (
	"errors"
	"strings"
	"testing"
)

// The rule lives here because a task and an objective are shown the same way
// and break the same way, so its test lives here too. It moved here without
// them once, and the mutation gate is what caught it: the two Unicode
// separators were only ever exercised from another package, whose coverage
// does not count for this one.
func TestOneLine_refusesEverythingThatWouldEndTheLine(t *testing.T) {
	for what, title := range map[string]string{
		"a newline":             "two\nlines",
		"a carriage return":     "two\rlines",
		"a tab":                 "two\tcolumns",
		"a vertical tab":        "two\vlines",
		"a form feed":           "two\flines",
		"a line separator":      "two\u2028lines",
		"a paragraph separator": "two\u2029paragraphs",
		"a C1 control":          "two\u0085lines",
	} {
		err := OneLine(title)
		if !errors.Is(err, ErrTitleNotOneLine) {
			t.Errorf("%s: OneLine(%q) = %v, want it refused", what, title, err)
		}
		if err != nil && !strings.Contains(err.Error(), "two") {
			t.Errorf("%s: the reason does not quote what was given: %v", what, err)
		}
	}
}

func TestOneLine_acceptsATitleThatIsText(t *testing.T) {
	for _, title := range []string{
		"Write the spec",
		"\u00c9crire la sp\u00e9c \u2014 la premi\u00e8re",
		"",
		"   ",
		"a title with an emoji \U0001f5ff",
	} {
		if err := OneLine(title); err != nil {
			t.Errorf("OneLine(%q) = %v, want it accepted", title, err)
		}
	}
}

func TestOrNothing_showsAnAbsenceAsAnAbsence(t *testing.T) {
	if got := OrNothing(""); got != Absent {
		t.Errorf("OrNothing(%q) = %q, want %q", "", got, Absent)
	}
	if got := OrNothing("something"); got != "something" {
		t.Errorf("OrNothing = %q, want it untouched", got)
	}
}
