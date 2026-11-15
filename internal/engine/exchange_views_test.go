package engine

import (
	"testing"
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
