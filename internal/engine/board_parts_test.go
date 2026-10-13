package engine

import (
	"cairn/internal/trace"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"testing"
)

func TestPartState_isDoneOrTodo(t *testing.T) {
	if PartState(true) != "done" || PartState(false) != "todo" {
		t.Errorf("PartState = %q, %q", PartState(true), PartState(false))
	}
}

func TestPartsWritten_isWhatPartsInReads(t *testing.T) {
	parts := []Part{{Number: 2, Title: "Write it", Done: true}, {Title: "Read it first"}}
	said := PartsWritten(parts)
	if said != "#2 [x] Write it\n[ ] Read it first" {
		t.Errorf("PartsWritten = %q", said)
	}
	if back := PartsIn(said); len(back) != 2 || back[0] != parts[0] || back[1] != parts[1] {
		t.Errorf("read back as %+v", back)
	}
}

func TestTaskEdit_settlesThePartsItIsGiven(t *testing.T) {
	e := openEngine(t)
	ship := addTitled(t, e, "Ship it", nil)
	spec := addTitled(t, e, "Write the spec", map[string]string{"parent": strconv.Itoa(ship)})
	kept := addTitled(t, e, "Kept", map[string]string{"parent": strconv.Itoa(ship)})
	if _, err := e.Ask(TaskEdit, Input{Given: map[string]string{
		"number": strconv.Itoa(ship),
		"parts":  "[ ] Ship the docs\n#" + strconv.Itoa(spec) + " [x] Write it",
	}}); err != nil {
		t.Fatalf("saving the parts: %v", err)
	}
	under, err := e.under(ship)
	if err != nil {
		t.Fatal(err)
	}
	said := map[string]string{}
	for _, u := range under {
		f := factsOf(u)
		said[f.title] = string(f.state)
	}
	if said["Write it"] != "done" || said["Ship the docs"] != "todo" || said["Kept"] != "todo" || len(said) != 3 {
		t.Errorf("the parts read %v, want Write it done, Ship the docs and Kept to do", said)
	}
	_ = kept
}

func TestTaskEdit_writesNoPartWhenOneIsRefused(t *testing.T) {
	e := openEngine(t)
	ship := addTitled(t, e, "Ship it", nil)
	if _, err := e.Ask(TaskEdit, Input{Given: map[string]string{
		"number": strconv.Itoa(ship), "parts": "[ ] Fine\n#99 [ ] Nowhere",
	}}); err == nil {
		t.Fatal("a part that is not there was taken")
	}
	if under, _ := e.under(ship); len(under) != 0 {
		t.Errorf("%d parts were written by an act that was refused", len(under))
	}
}

func TestTaskNew_writesThePartsItIsGiven(t *testing.T) {
	e := openEngine(t)
	answer, err := e.Ask(TaskNew, Input{Given: map[string]string{"title": "Ship it", "parts": "[ ] One\n[ ] Two"}})
	if err != nil {
		t.Fatal(err)
	}
	under, _ := e.under(factsOf(*answer.Task).number)
	if got := titlesOfTasks(under); !slices.Equal(got, []string{"One", "Two"}) {
		t.Errorf("the parts read %v", got)
	}
}

func addTitled(t *testing.T, e *Engine, title string, given map[string]string) int {
	t.Helper()
	in := Input{Given: map[string]string{"title": title}}
	for k, v := range given {
		in.Given[k] = v
	}
	answer, err := e.Ask(TaskNew, in)
	if err != nil {
		t.Fatalf("adding %q: %v", title, err)
	}
	return factsOf(*answer.Task).number
}

func TestTaskEdit_ordersThePartsColumnByColumn(t *testing.T) {
	e := openEngine(t)
	ship := addTitled(t, e, "Ship it", nil)
	under := map[string]string{"parent": strconv.Itoa(ship)}
	a, b := addTitled(t, e, "a", under), addTitled(t, e, "b", under)
	c := addTitled(t, e, "c", map[string]string{"parent": strconv.Itoa(ship), "state": "done"})
	d := addTitled(t, e, "d", map[string]string{"parent": strconv.Itoa(ship), "state": "done"})
	line := func(n int, done bool, title string) string {
		mark := " "
		if done {
			mark = "x"
		}
		return "#" + strconv.Itoa(n) + " [" + mark + "] " + title
	}
	moved := func() int {
		answer, err := e.Ask(TraceList, Input{Given: map[string]string{"since": "2000-01-01"}})
		if err != nil {
			t.Fatal(err)
		}
		count := 0
		for _, act := range answer.Acts {
			if strings.Contains(fmt.Sprint(trace.OfAct(act)), "moved") {
				count++
			}
		}
		return count
	}
	save := func(lines ...string) {
		if _, err := e.Ask(TaskEdit, Input{Given: map[string]string{"number": strconv.Itoa(ship), "parts": strings.Join(lines, "\n")}}); err != nil {
			t.Fatal(err)
		}
	}
	save(line(a, false, "a"), line(b, false, "b"), line(c, true, "c"), line(d, true, "d"))
	if got := moved(); got != 0 {
		t.Errorf("the order they had wrote %d moves", got)
	}
	save(line(d, true, "d"), line(b, false, "b"), line(c, true, "c"), line(a, false, "a"))
	parts, _ := e.under(ship)
	if got := titlesOfTasks(parts); !slices.Equal(got, []string{"b", "a", "d", "c"}) {
		t.Errorf("the parts read %v, want b a to do and d c done", got)
	}
}
