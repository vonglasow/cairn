package cli

import (
	"encoding/json"
	"strings"
	"testing"

	"cairn/internal/domain"
	"cairn/internal/engine"
	"cairn/internal/task"
)

// spec builds a numbered task with a title, for the tests below.
func spec(t *testing.T) task.Task {
	t.Helper()
	built, err := task.NewTask("Write the spec")
	if err != nil {
		t.Fatalf("NewTask: %v", err)
	}
	var a task.AsTold
	built.Tell(&a)
	a.Numbered(1)
	return a.Task()
}

// A machine reading the answer has to loop over the labels. A task carrying
// none must therefore hand back an empty list, never a null: null is not a
// list, and a reader that trusted the field would break on the first task
// nobody labelled.
func TestForWire_writesAnEmptyListRatherThanNothingAtAll(t *testing.T) {
	encoded, err := json.Marshal(engine.OfTask(spec(t), domain.Progress{}))
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if !strings.Contains(string(encoded), `"labels":[]`) {
		t.Errorf("labels should be an empty list, got %s", encoded)
	}
}

func TestForWire_writesTheLabelsATaskCarries(t *testing.T) {
	built := spec(t)
	var a task.AsTold
	built.Tell(&a)
	a.Labelled("spec")
	a.Labelled("team")
	a.LabelsTold()
	built = a.Task()

	encoded, err := json.Marshal(engine.OfTask(built, domain.Progress{}))
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if !strings.Contains(string(encoded), `"labels":["spec","team"]`) {
		t.Errorf("labels should carry what the task carries, got %s", encoded)
	}
}

func TestForWire_writesEveryAttributeAsText(t *testing.T) {
	built := spec(t)
	deadline, err := domain.ParseDeadline("2026-09-01")
	if err != nil {
		t.Fatalf("ParseDeadline: %v", err)
	}
	estimate, err := domain.ParseEstimate("90m")
	if err != nil {
		t.Fatalf("ParseEstimate: %v", err)
	}
	var a task.AsTold
	built.Tell(&a)
	a.Due(deadline)
	a.Estimated(estimate)
	built = a.Task()

	wire := engine.OfTask(built, domain.Progress{})
	if wire.Deadline != "2026-09-01" {
		t.Errorf("Deadline = %q, want 2026-09-01", wire.Deadline)
	}
	if wire.Estimate != "1h30m" {
		t.Errorf("Estimate = %q, want 1h30m", wire.Estimate)
	}
}

// The two places are padded to the longer of them, whichever that is.
//
// Only scenarios reached this, and they expand ~ into whatever temporary home
// the run was given — so which of the two paths was longer changed from one
// run to the next, and the comparison was killed or survived by luck. The
// mutation gate caught it as a flake: the same command answered 44/0 and 42/2
// on the same code, minutes apart.
func TestRenderConfig_padsBothPlacesToTheLongerOfThem(t *testing.T) {
	for _, tc := range []struct {
		what         string
		store, notes string
	}{
		{"the notes path is longer", "/s/db", "/notes/are/longer"},
		{"the store path is longer", "/store/is/longer/db", "/n"},
		{"both the same length", "/same/len", "/same/len"},
	} {
		said := renderConfig(engine.Config{
			File:  "/c.toml",
			Store: engine.Emplacement{At: tc.store, By: "the file"},
			Notes: engine.Emplacement{At: tc.notes, By: "nothing"},
		})
		var at []int
		for _, line := range strings.Split(said, "\n") {
			if i := strings.Index(line, "named by"); i >= 0 {
				at = append(at, i)
			}
		}
		if len(at) != 2 {
			t.Fatalf("%s: %q names what named it %d times, want twice", tc.what, said, len(at))
		}
		if at[0] != at[1] {
			t.Errorf("%s: the second column starts at %d and %d:\n%s", tc.what, at[0], at[1], said)
		}
		wide := max(len(tc.store), len(tc.notes))
		if got := at[0] - len("store   "); got < wide {
			t.Errorf("%s: the column is %d wide and the longer path is %d:\n%s", tc.what, got, wide, said)
		}
	}
}
