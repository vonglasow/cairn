package gui

import (
	"cairn/internal/engine"
	"image/color"
	"slices"
	"testing"
	"time"
)

func TestTheMenuBarHoldsEveryAct(t *testing.T) {
	want := []MenuGroup{
		{Name: "File", Items: []string{"New", "New backup", "Backups", "Labels", "Export", "Import", "Configuration…"}},
		{Name: "Edit", Items: []string{"Edit", "Focus", "Archive", "Delete", "Search"}},
		{Name: "View", Items: []string{"Tasks", "Objectives", "Time", "Archived", "Filter", "Reset sizes"}},
		{Name: "Help", Items: []string{"Help", "About cairn"}},
	}
	got := MenuBar(engine.Pages{})
	if len(got) != len(want) {
		t.Fatalf("the menu bar is %v", got)
	}
	for i := range want {
		if got[i].Name != want[i].Name || !slices.Equal(got[i].Items, want[i].Items) {
			t.Errorf("menu %d is %v, want %v", i, got[i], want[i])
		}
	}
}

func TestALinePickedActsFromTheBoard(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "A", State: "todo"})
	if got := w.Picked("New"); got.Screen() != OnForm || !got.Form().Adding {
		t.Errorf("New left the window on %v", got.Screen())
	}
	if got := w.Picked("Objectives"); got.Tab() != Objectives {
		t.Errorf("Objectives left the window on %v", got.Tab())
	}
	if got := w.Picked("Time"); got.Tab() != Time {
		t.Errorf("Time left the window on %v", got.Tab())
	}
	if got := w.onPage(Time).Picked("Tasks"); got.Tab() != Tasks {
		t.Errorf("Tasks left the window on %v", got.Tab())
	}
	chosen := w.clicked(middleOf(w, 1, 0))
	if chosen.Picked("Focus"); f.what("task", "focus") != 1 {
		t.Error("Take up took nothing up")
	}
	if got := chosen.Picked("Edit"); got.Screen() != OnForm || got.Form().Adding {
		t.Errorf("Edit left the window on %v", got.Screen())
	}
	open := chosen.Picked("Edit")
	if got := open.Picked("Objectives"); got.Tab() != Tasks || got.Screen() != OnForm {
		t.Errorf("picked over a form, the window went to %v on %v", got.Tab(), got.Screen())
	}
}

func TestTheMenuButtonStandsInForAMenuBar(t *testing.T) {
	w := holding(newFake(), engine.Card{Number: 1, Title: "A", State: "todo"})
	button := MenuButton(w.Layout())
	if tabs := w.Layout().Tabs; button.X+button.W != tabs.X+tabs.W || button.Y < tabs.Y || button.Y+button.H > tabs.Y+tabs.H {
		t.Errorf("the button %+v is not at the end of the tabs %+v", button, tabs)
	}
	opened := w.clicked(Point{X: button.X + button.W/2, Y: button.Y + button.H/2})
	var all []string
	for _, group := range MenuBar(engine.Pages{}) {
		for _, line := range group.Items {
			if w.Doable(line) {
				all = append(all, line)
			}
		}
	}
	if slices.Contains(all, "Edit") || !slices.Contains(all, "New") {
		t.Fatalf("with nothing chosen the button's menu would offer %v", all)
	}
	if opened.Screen() != OnMenu || !slices.Equal(opened.Menu(), all) {
		t.Fatalf("the button opened %v offering %v", opened.Screen(), opened.Menu())
	}
	if got := opened.fromMenu(slices.Index(all, "Objectives")); got.Tab() != Objectives {
		t.Errorf("Objectives from the button's menu left the window on %v", got.Tab())
	}
	box := opened.MenuBox()
	if box.X+box.W != button.X+button.W || box.Y != button.Y+button.H {
		t.Errorf("the menu %+v runs out of the window, or over its button %+v", box, button)
	}
	drawnOn := func(v Window) int {
		shot := drawn(t, v)
		at := shot.RGBAAt(button.X+1, button.Y+1)
		return inked(shot, Inside(button), color.NRGBA{R: at.R, G: at.G, B: at.B, A: 255})
	}
	if drawnOn(w) == 0 {
		t.Error("the button is not drawn")
	}
	native := w.Native(true)
	if got := native.clicked(Point{X: button.X + button.W/2, Y: button.Y + button.H/2}); got.Screen() == OnMenu {
		t.Error("with a menu bar of the system's, the button still opens a menu")
	}
	if drawnOn(native) != 0 {
		t.Error("with a menu bar of the system's, the button is still drawn")
	}
}

func TestALineOfTheMenuBarSaysItsKey(t *testing.T) {
	mac, pc := KeyboardOf("darwin"), KeyboardOf("windows")
	for _, c := range []struct {
		item, mac, pc string
	}{
		{"New", "⌘N", "Ctrl+N"},
		{"Export", "⌘E", "Ctrl+E"},
		{"Archived", "⇧A", "Shift+A"},
		{"Edit", "e", "e"},
		{"Tasks", "", ""},
	} {
		if got := mac.KeyOf(c.item); got != c.mac {
			t.Errorf("%s is under %q on a Mac", c.item, got)
		}
		if got := pc.KeyOf(c.item); got != c.pc {
			t.Errorf("%s is under %q on Windows", c.item, got)
		}
	}
}

func TestWhatTheSystemsMenuBarsNeedDecided(t *testing.T) {
	if made(0) || !made(1) {
		t.Error("a handle of nought names something, or one of 1 nothing")
	}
	lines := []string{"New", "New backup"}
	for at, want := range map[int]bool{-1: false, 0: true, 1: true, 2: false} {
		if _, ok := lineAt(lines, at); ok != want {
			t.Errorf("line %d found: %v", at, ok)
		}
	}
	if line, _ := lineAt(lines, 1); line != "New backup" {
		t.Errorf("line 1 is %q", line)
	}
	if commandOf(0) != 1 || commandOf(4) != 5 {
		t.Error("a command is not one past its line")
	}
	for _, c := range []struct {
		msg, wParam uintptr
		line        int
		ok          bool
	}{
		{0x0111, 1, 0, true},
		{0x0111, 5, 4, true},
		{0x0111, 1<<16 | 1, 0, false},
		{0x0110, 1, 0, false},
		{0x0112, 1, 0, false},
	} {
		if line, ok := commandLine(c.msg, c.wParam); ok != c.ok || (ok && line != c.line) {
			t.Errorf("message %#x, %#x reads line %d, %v", c.msg, c.wParam, line, ok)
		}
	}
	if winTitle("New") != "New\tCtrl+N" || winTitle("Tasks") != "Tasks" {
		t.Errorf("Win32 titles read %q and %q", winTitle("New"), winTitle("Tasks"))
	}
	if winMenu("File") != "&File" {
		t.Errorf("a menu reads %q", winMenu("File"))
	}
	for item, want := range map[string]string{"New": "n", "Export": "e", "Archived": "", "Edit": "", "Tasks": ""} {
		if got := macKey(item); got != want {
			t.Errorf("%s answers %q on a Mac", item, got)
		}
	}
}

func TestTheMenuButtonIsThreeStrokes(t *testing.T) {
	got := MenuStrokes(Box{X: 100, Y: 10, W: 40, H: 30})
	want := [3]Box{{X: 110, Y: 19, W: 20, H: 2}, {X: 110, Y: 24, W: 20, H: 2}, {X: 110, Y: 29, W: 20, H: 2}}
	if got != want {
		t.Errorf("the strokes are %+v", got)
	}
}

func TestTheMenuButtonIsDrawnAsItsStrokes(t *testing.T) {
	w := New(newFake(), 1200, 800)
	shot := drawn(t, w)
	for _, stroke := range MenuStrokes(MenuButton(w.Layout())) {
		if got := painted(shot, stroke, Dark().Chosen); got != stroke.W*stroke.H {
			t.Errorf("the stroke %+v has %d of its %d pixels", stroke, got, stroke.W*stroke.H)
		}
	}
}

func TestALinePickedIsHandedOverWithoutWaiting(t *testing.T) {
	picks := make(chan string, 1)
	release, woken := make(chan struct{}), make(chan struct{})
	wake := func() { <-release; close(woken) }
	returned := make(chan struct{})
	go func() { handedOver("Quit", picks, wake); close(returned) }()
	select {
	case <-returned:
	case <-time.After(time.Second):
		t.Fatal("handing a line over waited on the frame it asked for")
	}
	if got := <-picks; got != "Quit" {
		t.Errorf("the loop was handed %q", got)
	}
	close(release)
	select {
	case <-woken:
	case <-time.After(time.Second):
		t.Fatal("no frame was asked for")
	}
	full := make(chan string)
	done := make(chan struct{})
	go func() { handedOver("Quit", full, func() {}); close(done) }()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("handing a line to a full loop waited")
	}
}

func TestTheNativeBarsReadWhatCanBeDoneNow(t *testing.T) {
	w := holding(newFake(), engine.Card{Number: 1, Title: "A", State: "todo"})
	w.tellingTheBars()
	if lineDoable("Edit") || !lineDoable("New") {
		t.Fatalf("with nothing chosen the bars read Edit %v and New %v", lineDoable("Edit"), lineDoable("New"))
	}
	w.clicked(middleOf(w, 1, 0)).tellingTheBars()
	if !lineDoable("Edit") {
		t.Error("with a card chosen the bars still grey Edit")
	}
	if lineDoable("a line nobody has") {
		t.Error("a line no bar holds reads as doable")
	}
}
