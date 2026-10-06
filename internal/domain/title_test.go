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

// What is written about a thing runs over several lines and may line things up
// with a tab, but it is printed straight to a terminal: an escape among it is
// an instruction to that terminal, not text \u2014 hidden text, a link that lies
// about where it goes, a write to the clipboard.
func TestProse_refusesWhatATerminalWouldReadAsAnInstruction(t *testing.T) {
	for what, said := range map[string]string{
		"an escape":              "look \x1b]52;c;aGk=\x07 here",
		"a bell":                 "ring\x07",
		"a lone carriage return": "over\rwritten",
		"a backspace":            "erased\b",
		"a C1 control":           "two\u009blines",
		"a NUL":                  "cut\x00short",
	} {
		_, err := Prose(said)
		if !errors.Is(err, ErrNotText) {
			t.Errorf("%s: Prose(%q) = %v, want it refused", what, said, err)
		}
	}
}

func TestProse_keepsLinesAndTabs(t *testing.T) {
	for said, want := range map[string]string{
		"two\nparagraphs":          "two\nparagraphs",
		"lined\tup":                "lined\tup",
		"written on\r\nwindows":    "written on\nwindows",
		"\u00C9crire \u2014 enfin": "\u00C9crire \u2014 enfin",
		"":                         "",
	} {
		got, err := Prose(said)
		if err != nil || got != want {
			t.Errorf("Prose(%q) = %q, %v; want %q", said, got, err, want)
		}
	}
}

// A label is shown on one line among others, so it is held to the title's rule.
func TestParseLabels_refusesALabelThatIsNotOneLine(t *testing.T) {
	for _, label := range []string{"two\nlines", "esc\x1b[2J", "tab\tbed"} {
		if _, err := ParseLabels([]string{label}); !errors.Is(err, ErrLabelNotOneLine) {
			t.Errorf("ParseLabels(%q) = %v, want it refused", label, err)
		}
		if _, err := ParseLabelChange([]string{"+" + label}); !errors.Is(err, ErrLabelNotOneLine) {
			t.Errorf("ParseLabelChange(+%q) = %v, want it refused", label, err)
		}
	}
}

// A character that turns text around or hides in it says one thing on screen
// and keeps another: a title shown backwards is a title somebody acts on
// believing it says what it does not.
func TestOneLineAndProse_refuseWhatTurnsTextAroundOrHides(t *testing.T) {
	for what, said := range map[string]string{
		"a right-to-left override": "pay \u202Eevil\u202C now",
		"an isolate":               "a \u2066b\u2069",
		"a zero-width space":       "spec\u200Bial",
		"a byte-order mark":        "\uFEFFtitle",
	} {
		if err := OneLine(said); !errors.Is(err, ErrTitleNotOneLine) {
			t.Errorf("%s: OneLine = %v, want it refused", what, err)
		}
		if _, err := Prose(said); !errors.Is(err, ErrNotText) {
			t.Errorf("%s: Prose = %v, want it refused", what, err)
		}
	}
}

// What was kept before the refusals existed is shown with what a terminal
// would obey replaced, and everything else as it was.
func TestShown_replacesWhatATerminalWouldObey(t *testing.T) {
	for kept, want := range map[string]string{
		"look \x1b]52;c;aGk=\x07 here": "look \uFFFD]52;c;aGk=\uFFFD here",
		"pay \u202Eevil\u202C now":     "pay \uFFFDevil\uFFFD now",
		"two\nparagraphs\twith a tab":  "two\nparagraphs\twith a tab",
		"\u00C9crire \u2014 enfin":     "\u00C9crire \u2014 enfin",
	} {
		if got := Shown(kept); got != want {
			t.Errorf("Shown(%q) = %q, want %q", kept, got, want)
		}
	}
}

// Every character Unicode says formats text rather than shows it hides: the
// bidirectional marks and overrides, the zero-width ones, the soft hyphen,
// the invisible operators, and the tag characters a person cannot see and a
// model reads. A list of the ones somebody thought of left the rest through.
func TestHides_takesEveryFormatCharacter(t *testing.T) {
	for r, want := range map[rune]bool{
		'\u2029': false, '\u202A': true, '\u202E': true, '\u202F': false,
		'\u2066': true, '\u2069': true, '\u206A': true,
		'\u200A': false, '\u200B': true, '\u200D': true, '\u200E': true, '\u200F': true,
		'\u061C': true, '\u00AD': true, '\u2060': true, '\u2064': true,
		'\U000E0041': true, '\U000E007F': true,
		'\uFEFE': false, '\uFEFF': true, 'a': false, '\u00E9': false,
	} {
		if got := hides(r); got != want {
			t.Errorf("hides(%U) = %v, want %v", r, got, want)
		}
	}
}
