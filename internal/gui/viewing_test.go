package gui

import (
	"slices"
	"strings"
	"testing"

	"cairn/internal/board"
	"cairn/internal/gui/geom"
)

func TestTheTasksTabNamesTheViewInForceAndItsArrowOpensTheViews(t *testing.T) {
	w := viewedHome(1200, 800)
	if said := w.TabSaid(board.Tasks); said != "Tasks" {
		t.Errorf("on the whole board the tab says %q", said)
	}
	w = w.Key("cmd+1")
	if said := w.TabSaid(board.Tasks); said != "Tasks · Home" {
		t.Errorf("with the view Home in force the tab says %q", said)
	}
	arrow := w.TabArrow()
	if !w.TabsAt()[0].In(centre(arrow)) {
		t.Fatalf("the arrow %v is not in the tab %v", arrow, w.TabsAt()[0])
	}
	if opened := w.clicked(centre(arrow)); opened.Screen != board.OnViewing {
		t.Errorf("a click on the arrow leaves the window on %v", opened.Screen)
	}
	if opened := w.Key("cmd+shift+v"); opened.Screen != board.OnViewing {
		t.Errorf("⌘⇧V leaves the window on %v", opened.Screen)
	}
	if stays := w.clicked(centre(TabName(w.TabsAt()[0]))); stays.Screen != board.OnTheBoard {
		t.Errorf("a click on the tab's name leaves the window on %v", stays.Screen)
	}
}

func namesOf(lines []ViewingLine) []string {
	var names []string
	for _, line := range lines {
		names = append(names, line.Name)
	}
	return names
}

func TestTheViewsDroppedUnderTheTabFindAndOpenOne(t *testing.T) {
	w := viewedHome(1200, 800).Key("cmd+shift+v")
	tab := w.TabsAt()[0]
	if box := w.ViewingBox(); box.Y < tab.Y+tab.H || box.X > tab.X+tab.W {
		t.Errorf("the views drop at %v, not under the tab %v", box, tab)
	}
	lines := w.ViewingLines()
	if got := namesOf(lines); !slices.Equal(got, []string{WholeBoardSays, "Home", "Done"}) {
		t.Fatalf("the views listed are %q", got)
	}
	if lines[0].Key != "cmd+0" || lines[1].Key != "cmd+1" || lines[1].Count != "3" {
		t.Errorf("the lines say %+v", lines)
	}
	if opened := w.clicked(centre(lines[2].At)); opened.Viewing != "Done" || opened.Screen != board.OnTheBoard {
		t.Errorf("a click on Done opens %q on %v", opened.Viewing, opened.Screen)
	}
	sought := w.Edits(0, 0, "do")
	if got := namesOf(sought.ViewingLines()); !slices.Equal(got, []string{"Done"}) {
		t.Errorf("typing \"do\" leaves %q", got)
	}
	if opened := sought.Key("enter"); opened.Viewing != "Done" {
		t.Errorf("enter opens %q", opened.Viewing)
	}
	if walked := w.Key("down").Key("enter"); walked.Viewing != "Home" {
		t.Errorf("down then enter opens %q", walked.Viewing)
	}
	if closed := w.Key("esc"); closed.Screen != board.OnTheBoard {
		t.Errorf("esc leaves the window on %v", closed.Screen)
	}
	if closed := w.clicked(geom.Point{X: 1100, Y: 700}); closed.Screen != board.OnTheBoard {
		t.Errorf("a click outside leaves the window on %v", closed.Screen)
	}
}

func saidOf(acts []ViewingAct) []string {
	var said []string
	for _, act := range acts {
		said = append(said, act.Said)
	}
	return said
}

func TestTheViewsFootSavesTheFiltersAndOpensThePanel(t *testing.T) {
	bare := viewedHome(1200, 800).Key("cmd+shift+v")
	acts := bare.ViewingActs()
	if got := saidOf(acts); !slices.Equal(got, []string{KeepAsViewSays, EditViewsSays}) {
		t.Fatalf("with nothing on top, the foot says %q", got)
	}
	if !acts[0].Off || acts[0].Tag == "" {
		t.Errorf("with no filter on, saving is not greyed with what to do first: %+v", acts[0])
	}
	if stays := bare.clicked(centre(acts[0].At)); stays.Screen != board.OnViewing {
		t.Errorf("a click on a greyed line leaves the window on %v", stays.Screen)
	}
	if panel := bare.clicked(centre(acts[1].At)); panel.Screen != board.OnViews {
		t.Errorf("Edit the views… leaves the window on %v", panel.Screen)
	}
	searched := viewedHome(1200, 800)
	searched.Looking = board.Ending("spec")
	searched = searched.Key("cmd+shift+v")
	if keeping := searched.clicked(centre(searched.ViewingActs()[0].At)); keeping.Screen != board.OnKeeping {
		t.Errorf("saving a search as a view leaves the window on %v", keeping.Screen)
	}
	changed := viewedHome(1200, 800).Key("cmd+1")
	changed.Looking = board.Ending("spec")
	changed = changed.Key("cmd+shift+v")
	if got := saidOf(changed.ViewingActs()); !slices.Equal(got, []string{SaveToViewSays, KeepAsViewSays, EditViewsSays}) {
		t.Fatalf("with a search on top of a view, the foot says %q", got)
	}
	if saved := changed.clicked(centre(changed.ViewingActs()[0].At)); saved.Screen != board.OnTheBoard || !saved.Looking.Empty() {
		t.Errorf("saving to the view leaves the window on %v, searching %q", saved.Screen, saved.Looking.Reads())
	}
}

func TestTheViewInForceLivesInTheTabAloneAndItsChangeIsMarkedThere(t *testing.T) {
	w := viewedHome(1200, 800).Key("cmd+1")
	if w.TabChanged(board.Tasks) {
		t.Error("the tab marks the view changed with nothing on top")
	}
	w.Looking = board.Ending("spec")
	if !w.TabChanged(board.Tasks) || w.TabChanged(board.Objectives) {
		t.Error("only the Tasks tab marks the view changed under a search")
	}
	for _, pill := range w.Pills() {
		if strings.HasPrefix(pill.Said, "view ") {
			t.Errorf("a pill still stands for the view: %q", pill.Said)
		}
	}
	menu := w.Key("shift+f")
	if first := menu.FilterRowAt(0); first != DropRowAt(menu.filtersColumn(), 0) {
		t.Errorf("the filter menu opens on %v, not on its first filter at its head", first)
	}
}
