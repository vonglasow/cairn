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
	got, err := WithKey(nil, "series.on", "true")
	if err != nil || string(got) != "[series]\non = true\n" {
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

func TestAKeyNotWrittenJoinsItsTableAsACommentAndAWrittenOneStaysAsItIs(t *testing.T) {
	got := WithCommented([]byte("[pages]\n# none at home\nobjectives = false\n"), "pages.time", `"on"`, "by default")
	if want := "[pages]\n# none at home\nobjectives = false\n# time = \"on\"   # by default\n"; string(got) != want {
		t.Errorf("the file became\n%s", got)
	}
	if again := WithCommented(got, "pages.time", `"hidden"`, "from the global configuration"); string(again) != string(got) {
		t.Errorf("a key already there, as a comment, changed:\n%s", again)
	}
	if kept := WithCommented(got, "pages.objectives", "true", "by default"); string(kept) != string(got) {
		t.Errorf("a key written changed:\n%s", kept)
	}
	if fresh := WithCommented(nil, "window.ghost", "65", "by default"); string(fresh) != "[window]\n# ghost = 65   # by default\n" {
		t.Errorf("an empty file became\n%s", fresh)
	}
}

func TestAValueIsReplacedUpToWhereTOMLEndsIt(t *testing.T) {
	for _, c := range []struct{ was, key, literal, want string }{
		{"[approach]\nlabel = \"a \\\" # b\" # why\n", "approach.label", `"c"`, "[approach]\nlabel = \"c\" # why\n"},
		{"[approach]\nlabel = 'a # b' # why\n", "approach.label", `"c"`, "[approach]\nlabel = \"c\" # why\n"},
		{"[approach]\nlabel = '' # why\n", "approach.label", `"c"`, "[approach]\nlabel = \"c\" # why\n"},
		{"[window]\nghost = 65 # mine\n", "window.ghost", "70", "[window]\nghost = 70 # mine\n"},
		{"[window]\nghost = 65\n", "window.ghost", "70", "[window]\nghost = 70\n"},
	} {
		got, err := WithKey([]byte(c.was), c.key, c.literal)
		if err != nil || string(got) != c.want {
			t.Errorf("%q became %q, %v; want %q", c.was, got, err, c.want)
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

func TestAKeyNotWrittenInATableNotThereJoinsATableOfItsOwn(t *testing.T) {
	got := WithCommented([]byte("[pages]\n"), "window.ghost", "65", "by default")
	if want := "[pages]\n\n[window]\n# ghost = 65   # by default\n"; string(got) != want {
		t.Errorf("the file became %q", got)
	}
}

func TestEveryPlaceTheFileSaysIsChecked(t *testing.T) {
	if err := Readable([]byte("notes = \"relative\"\n"), "/h", false); err == nil {
		t.Error("a relative notes folder, said without a store, was read")
	}
}

func TestTheLineOfAKeyBeforeAnyTableIsFound(t *testing.T) {
	if got := LineOf([]byte("store = \"~/s.db\"\n[pages]\n"), "store"); got != 1 {
		t.Errorf("store is on line %d, want 1", got)
	}
}

func TestTheFirstKeyOfTheCatalogueIsFoundByItsName(t *testing.T) {
	first := Keys()[0]
	if got, ok := KeyNamed(first.Name); !ok || got.Name != first.Name {
		t.Errorf("%s is not found by its name: %v, %v", first.Name, got, ok)
	}
}

func TestAValueRefusedIsAValueError(t *testing.T) {
	refused := Refused("%s", "nope")
	if !errors.Is(refused, ErrValue) {
		t.Error("a value refused is not ErrValue")
	}
	if errors.Is(refused, ErrEmpty) {
		t.Error("a value refused is ErrEmpty")
	}
}

func TestANumberOfDaysIsOneOrMore(t *testing.T) {
	days, _ := KeyNamed("archive.days")
	if _, err := days.Literal("1"); err != nil {
		t.Errorf("one day is refused: %v", err)
	}
	if _, err := days.Literal("0"); !errors.Is(err, ErrValue) {
		t.Errorf("no day is taken: %v", err)
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

func TestAWorkspaceJoinsAnEmptyTableFollowedByAnother(t *testing.T) {
	got, err := WithWorkspace([]byte("[workspaces]\n[pages]\n"), "a", "~/a")
	if err != nil || string(got) != "[workspaces]\na = \"~/a\"\n[pages]\n" {
		t.Errorf("the file became %q, %v", got, err)
	}
}

func TestTheDefaultSaidFirstIsChangedInPlace(t *testing.T) {
	got, err := WithDefault([]byte("[workspaces]\ndefault = \"a\"\na = \"~/a\"\nb = \"~/b\"\n"), "b")
	if err != nil || string(got) != "[workspaces]\ndefault = \"b\"\na = \"~/a\"\nb = \"~/b\"\n" {
		t.Errorf("the file became %q, %v", got, err)
	}
}

func TestTheFirstWorkspaceOfAFileStartingWithTheTableIsTakenOut(t *testing.T) {
	got, err := WithoutWorkspace([]byte("[workspaces]\na = \"~/a\"\nb = \"~/b\"\n"), "a")
	if err != nil || string(got) != "[workspaces]\nb = \"~/b\"\n" {
		t.Errorf("the file became %q, %v", got, err)
	}
}

func TestAWorkspaceWrittenOtherThanAskedIsRefused(t *testing.T) {
	if _, err := checked([]byte("[workspaces]\nhome = \"~/a\"\n"), nil, "home", "~/b"); !errors.Is(err, ErrNotWritten) {
		t.Errorf("a workspace written to another folder passed: %v", err)
	}
}
