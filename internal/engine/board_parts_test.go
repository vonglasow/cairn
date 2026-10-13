package engine

import (
	"errors"
	"slices"
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
	want := []string{"task_edit 2Write it done", "task_new Read it first todo1", "task_delete 3 "}
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

// Parts are settled under a task, and number 0 is no task: every task with no
// parent would read as a part the field no longer names, and be deleted. So a
// parent of nothing asks the engine for nothing at all, whatever the field says.
func TestSettleParts_underNoTaskAsksNothing(t *testing.T) {
	f := newFake()
	had := []Card{{Number: 1, Title: "Ship it"}, {Number: 2, Title: "Write the spec"}}
	if err := SettleParts(f, 0, "A part", had); err == nil || len(f.asked) != 0 {
		t.Errorf("settling under no task answered %v and asked %d things", err, len(f.asked))
	}
}

// Parts written down are read back the same: a number when the part has one,
// its box, its title — and a part nobody has saved yet is written without a
// number, the way it is typed.
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

// Saving puts the parts in the order the field says, by moving the ones
// already there — and moves nothing when the order is the one they had, so a
// save that changed nothing about the order writes nothing about it.
func TestSettleParts_putsThePartsInTheOrderTheFieldSays(t *testing.T) {
	had := []Card{{Number: 2, Parent: 1, State: "todo"}, {Number: 3, Parent: 1, State: "todo"}, {Number: 4, Parent: 1, State: "todo"}}
	moves := func(said string) []string {
		f := newFake()
		if err := SettleParts(f, 1, said, had); err != nil {
			t.Fatal(err)
		}
		var did []string
		for _, a := range f.asked {
			if a.op.Name() == "task_move" {
				did = append(did, a.in.Given["number"]+":"+a.in.Given["first"]+a.in.Given["after"])
			}
		}
		return did
	}
	if got := moves("#2 [ ] a\n#3 [ ] b\n#4 [ ] c"); len(got) != 0 {
		t.Errorf("the order they had moved %v", got)
	}
	if got, want := moves("#4 [ ] c\n#2 [ ] a\n#3 [ ] b"), []string{"4:true", "2:4", "3:2"}; !slices.Equal(got, want) {
		t.Errorf("a new order moved %v, want %v", got, want)
	}
}

// Parts are put in order among those in the same column only: a task moves
// among the ones it sits with, and one done behind one to do is a move the
// engine refuses. Saved with five parts, done and to do in turn, the order
// asked for each column is kept and nothing is moved across the two.
func TestSettleParts_ordersEachColumnApart(t *testing.T) {
	had := []Card{
		{Number: 2, Parent: 1, State: "done"}, {Number: 3, Parent: 1, State: "done"},
		{Number: 4, Parent: 1, State: "todo"}, {Number: 5, Parent: 1, State: "done"},
		{Number: 6, Parent: 1, State: "todo"},
	}
	asked := func(said string) []string {
		f := newFake()
		if err := SettleParts(f, 1, said, had); err != nil {
			t.Fatal(err)
		}
		var did []string
		for _, a := range f.asked {
			if a.op.Name() == "task_move" {
				did = append(did, a.in.Given["number"]+":"+a.in.Given["first"]+a.in.Given["after"])
			}
		}
		return did
	}
	// As it is read back: what is left first, what is done after.
	if got := asked("#4 [ ] d\n#6 [ ] f\n#2 [x] b\n#3 [x] c\n#5 [x] e"); len(got) != 0 {
		t.Errorf("the order they were read in moved %v", got)
	}
	// The same, in the order they were put in: nothing to move either.
	if got := asked("#2 [x] b\n#3 [x] c\n#4 [ ] d\n#5 [x] e\n#6 [ ] f"); len(got) != 0 {
		t.Errorf("the order they were put in moved %v", got)
	}
	// Two to do swapped: those two are put in order, and nothing done moves.
	if got, want := asked("#6 [ ] f\n#2 [x] b\n#4 [ ] d\n#3 [x] c\n#5 [x] e"), []string{"6:true", "4:6"}; !slices.Equal(got, want) {
		t.Errorf("two to do swapped moved %v, want %v", got, want)
	}
	// A part ticked in the form is ordered among the done ones it joins.
	if got, want := asked("#4 [x] d\n#2 [x] b\n#3 [x] c\n#5 [x] e\n#6 [ ] f"), []string{"4:true", "2:4", "3:2", "5:3"}; !slices.Equal(got, want) {
		t.Errorf("a part ticked and put first moved %v, want %v", got, want)
	}
}

// A move the engine refuses is said, and nothing after it is asked.
func TestSettleParts_saysAMoveRefused(t *testing.T) {
	f := newFake()
	f.refuse = map[string]error{"task move": errNope}
	had := []Card{{Number: 2, Parent: 1, State: "todo"}, {Number: 3, Parent: 1, State: "todo"}}
	if err := SettleParts(f, 1, "#3 [ ] b\n#2 [ ] a", had); !errors.Is(err, errNope) {
		t.Errorf("a refused move answered %v", err)
	}
	if last := f.asked[len(f.asked)-1]; last.op.Name() != "task_move" || last.in.Given["number"] != "3" {
		t.Errorf("after the refusal it went on to ask %s %v", last.op.Name(), last.in.Given)
	}
}
