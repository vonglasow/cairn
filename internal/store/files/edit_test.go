package files

import (
	"errors"
	"strings"
	"testing"
)

const handWritten = `# home: what this board does differently
[pages]
# no objectives at home
objectives = false

[window]
theme = "nord"   # told apart from work at a glance

[workspace]
colour = "green"
`

func TestAKeyWrittenCommentedIsTakenOutOfItsCommentInPlace(t *testing.T) {
	got, err := WithKey([]byte(Example), "focus.session", `"45m"`)
	if err != nil {
		t.Fatal(err)
	}
	if want := strings.Replace(Example, "[focus]\n# session = \"1h\"\n", "[focus]\nsession = \"45m\"\n", 1); string(got) != want {
		t.Errorf("the example became\n%s", got)
	}
}

func TestAFileInAShapeTheEditorDoesNotKnowIsLeftForTheHand(t *testing.T) {
	for _, data := range []string{
		"pages = { time = \"on\" }\n",
		"[pages]\ntime = \"\"\"\non\"\"\"\n",
		"not toml [[[\n",
	} {
		if _, err := WithKey([]byte(data), "pages.time", `"off"`); !errors.Is(err, ErrNotEdited) {
			t.Errorf("%q was written over: %v", data, err)
		}
	}
}

func TestAFileReadsOrSaysWhy(t *testing.T) {
	if err := Readable([]byte(handWritten), "/h", true); err != nil {
		t.Errorf("the settings do not read: %v", err)
	}
	for _, c := range []struct {
		data     string
		settings bool
		said     string
	}{
		{"[pages]\ntime = \"hiden\"\n", true, "pages.time"},
		{"store = \"~/s.db\"\n", true, "store"},
		{"store = \"s.db\"\n", false, "must be absolute"},
		{"[[[\n", false, "make sense of"},
		{"[workspaces]\nwork = \"w\"\n", false, "workspaces.work"},
	} {
		if err := Readable([]byte(c.data), "/h", c.settings); err == nil || !strings.Contains(err.Error(), c.said) {
			t.Errorf("%q reads with %v, want %q", c.data, err, c.said)
		}
	}
}

func TestTheLineAKeyIsWrittenOnIsFound(t *testing.T) {
	for key, want := range map[string]int{"pages.objectives": 4, "window.theme": 7, "pages.time": 0, "workspace.colour": 10} {
		if got := LineOf([]byte(handWritten), key); got != want {
			t.Errorf("%s is on line %d, want %d", key, got, want)
		}
	}
}

func TestAStringThatNeverEndsIsLeftForTheHand(t *testing.T) {
	for _, data := range []string{"[approach]\nlabel = \"abc\n", "[approach]\nlabel = 'abc\n", "[approach]\nlabel = \"ab\\\n"} {
		if _, err := WithKey([]byte(data), "approach.label", `"c"`); !errors.Is(err, ErrNotEdited) {
			t.Errorf("%q was written over: %v", data, err)
		}
	}
}

func TestAKeyJoinsItsTableWhereTheTableEnds(t *testing.T) {
	for _, c := range []struct{ name, was, key, literal, want string }{
		{"an empty table with another after it", "[pages]\n[window]\ntheme = \"nord\"\n", "pages.time", `"on"`,
			"[pages]\ntime = \"on\"\n[window]\ntheme = \"nord\"\n"},
		{"before the blank and the comment heading the next table", "[pages]\ntime = \"on\"\n\n# the window\n[window]\n", "pages.objectives", "false",
			"[pages]\ntime = \"on\"\nobjectives = false\n\n# the window\n[window]\n"},
		{"after a comment touching the next table", "[pages]\ntime = \"on\"\n# the window\n[window]\n", "pages.objectives", "false",
			"[pages]\ntime = \"on\"\n# the window\nobjectives = false\n[window]\n"},
		{"after a comment ending the file", "[pages]\ntime = \"on\"\n\n# the end\n", "pages.objectives", "false",
			"[pages]\ntime = \"on\"\n\n# the end\nobjectives = false\n"},
		{"the second key of a table changed", "[pages]\ntime = \"on\"\nobjectives = false\n", "pages.objectives", "true",
			"[pages]\ntime = \"on\"\nobjectives = true\n"},
		{"a file of one blank line", "\n", "plugin", `"~/p"`, "plugin = \"~/p\"\n\n"},
	} {
		got, err := WithKey([]byte(c.was), c.key, c.literal)
		if err != nil || string(got) != c.want {
			t.Errorf("%s: %q became %q, %v; want %q", c.name, c.was, got, err, c.want)
		}
	}
}

func TestAGhostIsAPercentageFromTheLeastToTheMost(t *testing.T) {
	ghost, _ := KeyNamed("window.ghost")
	for _, taken := range []string{"20", "90"} {
		if _, err := ghost.Literal(taken); err != nil {
			t.Errorf("%s is refused: %v", taken, err)
		}
		if err := Readable([]byte("[window]\nghost = "+taken+"\n"), "/h", false); err != nil {
			t.Errorf("a file saying %s is refused: %v", taken, err)
		}
	}
	for _, refused := range []string{"19", "91", "much"} {
		if _, err := ghost.Literal(refused); !errors.Is(err, ErrValue) {
			t.Errorf("%s is taken: %v", refused, err)
		}
	}
	for _, refused := range []string{"19", "91"} {
		if err := Readable([]byte("[window]\nghost = "+refused+"\n"), "/h", false); err == nil {
			t.Errorf("a file saying %s is read", refused)
		}
	}
}
