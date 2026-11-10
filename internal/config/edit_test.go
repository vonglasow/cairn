package config

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

func TestAKeySetJoinsTheEndOfItsTable(t *testing.T) {
	got, err := WithKey([]byte(handWritten), "pages.time", `"hidden"`)
	if err != nil {
		t.Fatal(err)
	}
	want := strings.Replace(handWritten, "objectives = false\n", "objectives = false\ntime = \"hidden\"\n", 1)
	if string(got) != want {
		t.Errorf("the file became\n%s", got)
	}
}

func TestAValueChangedKeepsItsLineAndTheCommentAfterIt(t *testing.T) {
	got, err := WithKey([]byte(handWritten), "window.theme", `"gruvbox-dark"`)
	if err != nil {
		t.Fatal(err)
	}
	want := strings.Replace(handWritten, `theme = "nord"   #`, `theme = "gruvbox-dark"   #`, 1)
	if string(got) != want {
		t.Errorf("the file became\n%s", got)
	}
}

func TestAKeyInATableNotThereJoinsATableOfItsOwnAtTheEnd(t *testing.T) {
	got, err := WithKey([]byte(handWritten), "focus.session", `"1h"`)
	if err != nil {
		t.Fatal(err)
	}
	if want := handWritten + "\n[focus]\nsession = \"1h\"\n"; string(got) != want {
		t.Errorf("the file became\n%s", got)
	}
}

func TestAKeyInAnEmptyFileIsItsTableAlone(t *testing.T) {
	got, err := WithKey(nil, "recurrence.on", "true")
	if err != nil || string(got) != "[recurrence]\non = true\n" {
		t.Errorf("an empty file became %q, %v", got, err)
	}
}

func TestAKeyWrittenCommentedIsTakenOutOfItsCommentInPlace(t *testing.T) {
	got, err := WithKey([]byte(Example), "focus.session", `"45m"`)
	if err != nil {
		t.Fatal(err)
	}
	if want := strings.Replace(Example, "[focus]\n# session = \"1h\"\n", "[focus]\nsession = \"45m\"\n", 1); string(got) != want {
		t.Errorf("the example became\n%s", got)
	}
	got, err = WithKey([]byte(Example), "store", `"~/elsewhere/cairn.db"`)
	if err != nil || !strings.Contains(string(got), "# Absolute, or under ~. The default workspace, below, comes before it.\nstore = \"~/elsewhere/cairn.db\"\n") {
		t.Errorf("the store became %v\n%s", err, got)
	}
}

func TestAKeyBeforeAnyTableJoinsTheHeadOfTheFile(t *testing.T) {
	got, err := WithKey([]byte("[pages]\ntime = \"off\"\n"), "plugin", `"~/p"`)
	if err != nil || string(got) != "plugin = \"~/p\"\n\n[pages]\ntime = \"off\"\n" {
		t.Errorf("the file became %q, %v", got, err)
	}
	got, err = WithKey([]byte("# mine\nstore = \"~/s.db\"\n\n[pages]\n"), "plugin", `"~/p"`)
	if err != nil || string(got) != "# mine\nstore = \"~/s.db\"\nplugin = \"~/p\"\n\n[pages]\n" {
		t.Errorf("the file became %q, %v", got, err)
	}
}

func TestAValueQuotedWithAHashInItIsReplacedWhole(t *testing.T) {
	got, err := WithKey([]byte("[approach]\nlabel = \"a # b\" # why\n"), "approach.label", `"c"`)
	if err != nil || string(got) != "[approach]\nlabel = \"c\" # why\n" {
		t.Errorf("the file became %q, %v", got, err)
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

func TestAKeyTakenOutLosesItsLineAndNothingElse(t *testing.T) {
	got, err := WithoutKey([]byte(handWritten), "pages.objectives")
	if err != nil {
		t.Fatal(err)
	}
	if want := strings.Replace(handWritten, "objectives = false\n", "", 1); string(got) != want {
		t.Errorf("the file became\n%s", got)
	}
	got, err = WithoutKey([]byte("plugin = \"~/p\"\n[pages]\n"), "plugin")
	if err != nil || string(got) != "[pages]\n" {
		t.Errorf("the file became %q, %v", got, err)
	}
}

func TestAKeyNotSaidIsNotTakenOut(t *testing.T) {
	for _, key := range []string{"pages.time", "focus.session", "store"} {
		if _, err := WithoutKey([]byte(handWritten), key); !errors.Is(err, ErrNotSet) {
			t.Errorf("%s was taken out of a file that does not say it: %v", key, err)
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
