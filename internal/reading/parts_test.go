package reading

import (
	"errors"
	"testing"
)

// Saving a task's form settles what hangs under it with what the parts field
// says: a numbered line is that part, renamed and ticked as it reads; a line
// with no number is a new part; a part the field no longer names is gone. Both
// boards carried the same copy of this, down to its comment.
func TestSettleParts_asksForEachPartWhatTheFieldSays(t *testing.T) {
	f := newFake()
	had := []Card{{Number: 2, Parent: 1, State: "todo"}, {Number: 3, Parent: 1, State: "done"}}
	said := "#2 [x] Write it\nRead it first"
	if err := SettleParts(f, 1, said, had); err != nil {
		t.Fatal(err)
	}
	var did []string
	for _, a := range f.asked {
		did = append(did, a.op.Name()+" "+a.in.Given["number"]+a.in.Given["title"]+" "+a.in.Given["state"]+a.in.Given["parent"])
	}
	want := []string{"task_edit 2Write it done", "task_add Read it first todo1", "task_delete 3 "}
	if len(did) != len(want) {
		t.Fatalf("asked %v, want %v", did, want)
	}
	for i := range want {
		if did[i] != want[i] {
			t.Errorf("asked %q, want %q", did[i], want[i])
		}
	}
}

func TestSettleParts_stopsAtTheFirstRefusal(t *testing.T) {
	f := newFake()
	f.fail = errNope
	if err := SettleParts(f, 1, "Read it first", nil); !errors.Is(err, errNope) {
		t.Errorf("SettleParts = %v, want the refusal", err)
	}
	if err := SettleParts(f, 1, "#2 [ ] Read it", nil); !errors.Is(err, errNope) {
		t.Errorf("SettleParts = %v, want the refusal", err)
	}
	if err := SettleParts(f, 1, "", []Card{{Number: 2, Parent: 1}}); !errors.Is(err, errNope) {
		t.Errorf("SettleParts = %v, want the refusal", err)
	}
}

// Ticked or not is the one thing a part carries besides its title, and it is
// done or todo — said once, for both boards and for ticking one in the window.
func TestPartState_isDoneOrTodo(t *testing.T) {
	if PartState(true) != "done" || PartState(false) != "todo" {
		t.Errorf("PartState = %q, %q", PartState(true), PartState(false))
	}
}
