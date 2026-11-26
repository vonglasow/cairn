package files

import (
	"strings"
	"testing"
)

func TestAValueIsWrittenAsTheKeyTakesIt(t *testing.T) {
	for _, c := range []struct{ key, value, literal string }{
		{"pages.time", "hidden", `"hidden"`},
		{"pages.tasks", "false", "false"},
		{"series.on", "true", "true"},
		{"focus.session", "45m", `"45m"`},
		{"agents.silent", "25m", `"25m"`},
		{"archive.start", "2026-10-21", "2026-10-21"},
		{"archive.days", "14", "14"},
		{"approach.label", "discuss", `"discuss"`},
		{"plugin", "~/plugins", `"~/plugins"`},
		{"window.theme", "nord", `"nord"`},
		{"window.intro", "flap", `"flap"`},
		{"workspace.colour", "green", `"green"`},
	} {
		key, _ := KeyNamed(c.key)
		got, err := key.Literal(c.value)
		if err != nil || got != c.literal {
			t.Errorf("%s = %q is written %q, %v; want %q", c.key, c.value, got, err, c.literal)
		}
	}
}

func TestAValueTheKeyDoesNotTakeIsRefusedNamingIt(t *testing.T) {
	for _, c := range []struct{ key, value, said string }{
		{"pages.time", "hiden", `"hiden" is not a word pages.time takes (on, hidden, off)`},
		{"series.on", "yes", `series.on: "yes" is true or false`},
		{"archive.days", "0", `archive.days: "0" is not a number of days`},
		{"archive.days", "x", `archive.days: "x" is not a number of days`},
		{"archive.start", "monday", `archive.start: "monday" is not a day`},
		{"plugin", "plugins", `"plugins"`},
		{"approach.label", "", "approach.label"},
		{"window.intro", "confetti", `"confetti" is not a word window.intro takes`},
	} {
		key, _ := KeyNamed(c.key)
		_, err := key.Literal(c.value)
		if err == nil || !strings.Contains(err.Error(), c.said) {
			t.Errorf("%s = %q is refused with %v, want %q", c.key, c.value, err, c.said)
		}
	}
}
