package gui

import (
	"cairn/internal/build"
	"cairn/internal/engine"
	"slices"
	"testing"
)

func TestTheSystemOpensAFileWithItsOwnCommand(t *testing.T) {
	for _, c := range []struct {
		goos string
		name string
		args []string
	}{
		{"darwin", "open", []string{"-t", "/h/config.toml"}},
		{"linux", "xdg-open", []string{"/h/config.toml"}},
		{"windows", "cmd", []string{"/c", "start", "", "/h/config.toml"}},
	} {
		name, args := OpenCommand(c.goos, "/h/config.toml")
		if name != c.name || !slices.Equal(args, c.args) {
			t.Errorf("on %s the file is opened by %q %q, want %q %q", c.goos, name, args, c.name, c.args)
		}
	}
}

func TestAPathIsCopiedByTheMarkBeforeIt(t *testing.T) {
	w := holding(newFake())
	w.configured = engine.Config{File: "/h/.config/cairn/config.toml",
		Store: engine.Emplacement{At: "/h/.cairn/cairn.db", By: "nothing"},
		Notes: engine.Emplacement{At: "/h/.cairn/notes", By: "nothing"}, Plugin: engine.Emplacement{At: "/h/.cairn/plugin", By: "nothing"}}
	w.configHere, w.screen = true, OnConfiguration
	marks := w.CopyMarks()
	if len(marks) != 4 {
		t.Fatalf("the configuration offers %d paths to copy: %+v", len(marks), marks)
	}
	store := marks[0]
	if got, _ := w.clicked(Point{X: store.At.X + store.At.W/2, Y: store.At.Y + store.At.H/2}).Copying(); got != "/h/.cairn/cairn.db" {
		t.Errorf("the store's mark copies %q", got)
	}
	for _, line := range w.ConfigLines() {
		if line.Key == "store" && line.ValueAt.X < store.At.X+store.At.W {
			t.Errorf("the store's path %+v is written over its mark %+v", line.ValueAt, store.At)
		}
	}
	if painted(drawn(t, w), store.At, Dark().Low) == 0 {
		t.Error("the mark to copy the store's path is not drawn")
	}
}

func TestTheAboutBoxCopiesWhereCairnIsKept(t *testing.T) {
	w := holding(newFake())
	w.screen, w.question = OnAsking, aboutCairn
	marks := w.CopyMarks()
	if len(marks) != 1 || marks[0].Said != build.Home {
		t.Fatalf("About offers to copy %+v", marks)
	}
}

func TestAnOpenTasksContextIsCopiedByItsMark(t *testing.T) {
	w := holding(newFake(), engine.Card{Number: 1, Title: "A", State: "todo", Context: "/h/notes/1"})
	w = w.clicked(middleOf(w, 1, 0)).Key("enter").Key("e")
	marks := w.CopyMarks()
	if len(marks) != 1 || marks[0].Said != "/h/notes/1" {
		t.Fatalf("an open task offers to copy %+v", marks)
	}
	if said := w.ContextSays(); len(said) != 2 || said[1].Box.X < marks[0].At.X+marks[0].At.W {
		t.Errorf("the context %+v is written over its mark %+v", said, marks[0].At)
	}
}

func TestAKeyTheStoresSettingsWriteIsSaidToComeFromThem(t *testing.T) {
	w := holding(newFake())
	w.configured = engine.Config{File: "/h/.config/cairn/config.toml",
		Pages: &engine.PagesConfig{}, Archive: &engine.Archive{}, Focus: &engine.FocusConfig{}, Time: &engine.TimeConfig{}, Window: &engine.WindowConfig{},
		Written: []string{"focus.session", "time.weekends"}, WrittenHere: []string{"time.weekends"}}
	from := map[string]string{}
	for _, line := range w.ConfigLines() {
		from[line.Key] = line.From
	}
	if from["time.weekends"] != "settings" || from["focus.session"] != "file" || from["time.deleted"] != "default" {
		t.Errorf("the keys come from %v", from)
	}
}
