package board

import (
	"path/filepath"
	"strconv"
	"testing"

	"cairn/internal/engine"
	"cairn/internal/engine/task"
)

func openEngine(t *testing.T) *engine.Engine {
	t.Helper()
	e, err := engine.Open(filepath.Join(t.TempDir(), "cairn.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = e.Close() })
	return e
}

func asked(t *testing.T, e *engine.Engine, op engine.Operation, in engine.Input) {
	t.Helper()
	if _, err := e.Ask(op, in); err != nil {
		t.Fatalf("%s: %v", op.Line(), err)
	}
}

func addObjective(t *testing.T, e *engine.Engine, title string) {
	t.Helper()
	asked(t, e, engine.ObjectiveNew, engine.Input{Given: map[string]string{"title": title}})
}

func under(t *testing.T, e *engine.Engine, title, parent string) {
	t.Helper()
	asked(t, e, engine.ObjectiveNew, engine.Input{Given: map[string]string{"title": title, "parent": parent}})
}

func servingAll(t *testing.T, e *engine.Engine, title string, numbers ...int) {
	t.Helper()
	var serving []string
	for _, n := range numbers {
		serving = append(serving, strconv.Itoa(n))
	}
	asked(t, e, engine.TaskNew, engine.Input{Given: map[string]string{"title": title}, Listed: map[string][]string{"objective": serving}})
}

type request struct {
	given  map[string]string
	listed map[string][]string
}

func ask(pairs ...string) request {
	a := request{given: map[string]string{}, listed: map[string][]string{}}
	for i := 0; i+1 < len(pairs); i += 2 {
		a.given[pairs[i]] = pairs[i+1]
	}
	return a
}

func (a request) with(name string, values ...string) request {
	a.listed[name] = values
	return a
}

func (a request) input() engine.Input {
	return engine.Input{Given: a.given, Listed: a.listed, Confirmed: true}
}

func builtTask(t *testing.T, number int, title string) task.Task {
	t.Helper()
	built, err := task.NewTask(title)
	if err != nil {
		t.Fatalf("NewTask(%q): %v", title, err)
	}
	a := &task.AsTold{}
	*a = task.From(built)
	a.Numbered(number)
	return a.Task()
}
