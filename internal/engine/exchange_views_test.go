package engine

import (
	"testing"

	"cairn/internal/store"
)

func TestImport_aViewKeepsTheFocusInOrOutAndNothingElse(t *testing.T) {
	for focus, taken := range map[string]bool{"in": true, "out": true, "sideways": false} {
		file := `{"cairn": "tasks", "version": 7, "objectives": [], "tasks": [{"number": 1, "title": "Ship it", "state": "todo", "priority": "normal"}], "views": [{"name": "Home", "focus": "` + focus + `"}]}`
		_, err := importInto(t, openEngine(t), file)
		if (err == nil) != taken {
			t.Errorf("a view keeping the focus %s: %v", focus, err)
		}
	}
}

func TestExport_passesOverAViewNarrowingNothingAndWritesTheNext(t *testing.T) {
	e := openEngine(t)
	if err := e.store.KeepViews([]store.View{{Name: "Gone", Objectives: []int{9}}, {Name: "Todo", State: "todo"}}); err != nil {
		t.Fatal(err)
	}
	tr, err := e.tree()
	if err != nil {
		t.Fatal(err)
	}
	written, err := e.viewsWritten(tr)
	if err != nil || len(written) != 1 || written[0].Name != "Todo" {
		t.Errorf("written %+v, %v", written, err)
	}
}
