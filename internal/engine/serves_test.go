package engine

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"
)

func TestATaskNamesEachObjectiveItServesUnderItsParent(t *testing.T) {
	path := storeAt(t, "Pin map", "Lamp")
	for _, made := range []map[string]string{{"title": "Home automation"}, {"title": "Enclosure", "parent": "1"}} {
		if _, err := Do(ObjectiveNew, path, Input{Given: made}); err != nil {
			t.Fatal(err)
		}
	}
	for task, goal := range map[string]string{"1": "+2", "2": "+1"} {
		if _, err := Do(TaskEdit, path, Input{Given: map[string]string{"number": task}, Listed: map[string][]string{"objective": {goal}}}); err != nil {
			t.Fatal(err)
		}
	}
	shown, err := Do(TaskShow, path, given("number", "1"))
	if err != nil {
		t.Fatal(err)
	}
	if got := shown.Choosing[1].Serves; !slices.Equal(got, []string{"Home automation › Enclosure"}) {
		t.Errorf("task show serves %q", got)
	}
	listed, err := Do(TaskList, path, Input{})
	if err != nil {
		t.Fatal(err)
	}
	if got := listed.Choosing[2].Serves; !slices.Equal(got, []string{"Home automation"}) {
		t.Errorf("a task serving a top objective serves %q", got)
	}
	out, _ := json.Marshal(Of(TaskShow, shown))
	if !strings.Contains(string(out), `"Home automation › Enclosure"`) {
		t.Errorf("task show crosses as %s, naming no path", out)
	}
}
