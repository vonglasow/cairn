package engine

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"cairn/internal/store"
)

func TestTaskExport_writesTheViewsAndTheObjectivesTheyName(t *testing.T) {
	read := exportedFrom(t,
		asked(ObjectiveNew, "title", "Named by a view alone"),
		asked(ObjectiveNew, "title", "Served by nothing"),
		asked(ViewNew, "name", "Home", "objective", "1", "itself", "true", "state", "todo"),
		asked(ViewNew, "name", "Shopping", "label", "shopping", "priority", "high", "due-before", "2026-12-01", "search", "milk", "with-archived", "true"),
		asked(ViewNew, "name", "Attic", "archived", "true"),
	)
	if read.Version != 7 {
		t.Errorf("version %d, want 7", read.Version)
	}
	if len(read.Views) != 3 {
		t.Fatalf("the file holds %d views, want 3: %+v", len(read.Views), read.Views)
	}
	home, shopping, attic := read.Views[0], read.Views[1], read.Views[2]
	if home.Name != "Home" || !slices.Equal(home.Objective, []int{1}) || !home.Itself || home.State != "todo" {
		t.Errorf("the first view reads %+v", home)
	}
	if shopping.Name != "Shopping" || !slices.Equal(shopping.Label, []string{"shopping"}) || shopping.Priority != "high" ||
		shopping.DueBefore != "2026-12-01" || shopping.Search != "milk" || !shopping.WithArchived {
		t.Errorf("the second view reads %+v", shopping)
	}
	if attic.Name != "Attic" || !attic.Archived {
		t.Errorf("the third view reads %+v", attic)
	}
	if len(read.Objectives) != 2 || read.Objectives[0].Title != "Named by a view alone" || read.Objectives[1].Title != "Served by nothing" {
		t.Errorf("the objectives written are %+v, want the one the view names and the one nothing names", read.Objectives)
	}
}

func TestTaskExport_leavesOutAnObjectiveGoneAndAViewLeftHoldingNothing(t *testing.T) {
	e := openEngine(t)
	if _, err := e.addObjective("Still here", ObjectiveChanges{}); err != nil {
		t.Fatal(err)
	}
	if err := e.store.KeepViews([]store.View{
		{Name: "Half gone", Objectives: []int{1, 9}},
		{Name: "All gone", Objectives: []int{9}},
	}); err != nil {
		t.Fatal(err)
	}
	read, err := e.exporting()
	if err != nil {
		t.Fatal(err)
	}
	if len(read.Views) != 1 || read.Views[0].Name != "Half gone" || !slices.Equal(read.Views[0].Objective, []int{1}) {
		t.Errorf("the views written are %+v, want the one still holding objective 1 alone", read.Views)
	}
}

func TestTaskImport_bringsTheViewsBackAfterTheOnesHereAtTheNewNumbers(t *testing.T) {
	e := openEngine(t)
	if _, err := e.addObjective("Already here", ObjectiveChanges{}); err != nil {
		t.Fatal(err)
	}
	if err := e.store.KeepViews([]store.View{{Name: "Shopping", State: "current"}}); err != nil {
		t.Fatal(err)
	}
	from := filepath.Join(t.TempDir(), "held.json")
	if err := os.WriteFile(from, []byte(`{"cairn": "tasks", "version": 7,
		"objectives": [{"number": 4, "title": "Home automation", "state": "active", "priority": "normal"}],
		"tasks": [],
		"views": [
		  {"name": "Home", "objective": [4], "itself": true, "state": "todo"},
		  {"name": "shopping", "label": ["food"]},
		  {"name": "Errands", "label": ["errand"], "priority": "high", "due_before": "2026-12-01", "search": "milk", "with_archived": true},
		  {"name": "Attic", "archived": true}]}`), 0o600); err != nil {
		t.Fatal(err)
	}
	read, err := readExport(from)
	if err == nil {
		_, err = e.importing(read)
	}
	if err != nil {
		t.Fatal(err)
	}
	views, err := e.store.Views()
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, v := range views {
		names = append(names, v.Name)
	}
	if !slices.Equal(names, []string{"Shopping", "Home", "Errands", "Attic"}) {
		t.Fatalf("the views are %v, want the one here first, then the file's, the name taken left out", names)
	}
	if views[0].State != "current" || len(views[0].Labels) != 0 {
		t.Errorf("the view here became %+v", views[0])
	}
	if home := views[1]; !slices.Equal(home.Objectives, []int{2}) || !home.Itself || home.State != "todo" {
		t.Errorf("Home came back as %+v, want objective 2", home)
	}
	if errands := views[2]; !slices.Equal(errands.Labels, []string{"errand"}) || errands.Priority != "high" ||
		errands.DueBefore != "2026-12-01" || errands.Search != "milk" || !errands.WithArchived {
		t.Errorf("Errands came back as %+v", errands)
	}
	if !views[3].Archived {
		t.Errorf("Attic came back as %+v", views[3])
	}
}

func TestTaskImport_refusesAViewItCannotReadBeforeAnythingIsAdded(t *testing.T) {
	for _, c := range []struct {
		view, says string
		outside    bool
	}{
		{view: `{"name": "Home", "objective": [9]}`, says: "view Home names objective 9 — objective 9 is not in the file", outside: true},
		{view: `{"name": "Home", "state": "todo"}, {"name": "home", "state": "done"}`, says: `view "home" is in the file twice`},
		{view: `{"name": "", "state": "todo"}`, says: "view"},
		{view: `{"name": "Home"}`, says: "view Home"},
		{view: `{"name": "Home", "state": "someday"}`, says: "view Home"},
		{view: `{"name": "Home", "priority": "urgent"}`, says: "view Home"},
		{view: `{"name": "Home", "due_before": "someday"}`, says: "view Home"},
		{view: `{"name": "Home", "label": ["a\nb"]}`, says: "view Home"},
	} {
		from := filepath.Join(t.TempDir(), "held.json")
		file := `{"cairn": "tasks", "version": 7, "objectives": [], "tasks": [{"number": 1, "title": "Ship it", "state": "todo", "priority": "normal"}], "views": [` + c.view + `]}`
		if err := os.WriteFile(from, []byte(file), 0o600); err != nil {
			t.Fatal(err)
		}
		e := openEngine(t)
		read, err := readExport(from)
		if err == nil {
			_, err = e.importing(read)
		}
		if err == nil || !strings.Contains(err.Error(), c.says) {
			t.Errorf("%s: err = %v, want it to say %q", c.view, err, c.says)
		}
		if c.outside && !errors.Is(err, ErrExportPointsOutside) {
			t.Errorf("%s: err = %v, want ErrExportPointsOutside", c.view, err)
		}
		if tasks, _ := e.store.All(); len(tasks) != 0 {
			t.Errorf("%s: %d tasks were added", c.view, len(tasks))
		}
		if views, _ := e.store.Views(); len(views) != 0 {
			t.Errorf("%s: %d views were added", c.view, len(views))
		}
	}
}
