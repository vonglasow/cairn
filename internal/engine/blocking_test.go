package engine

// What the rules about blocking are worth, asked of them directly. The
// scenarios exercise the same rules through the commands, and credit this
// package with nothing: coverage and mutation are counted per package, so a
// rule tested only from outside is a rule nothing here can vouch for.

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"cairn/internal/domain"
	"cairn/internal/task"
	"cairn/internal/trace"
)

// A plain number says what the set becomes; a sign says how it moves.
func TestChangedBlockers_readsTheSameGrammarAsTheLabels(t *testing.T) {
	for _, tc := range []struct {
		what    string
		given   []string
		current []int
		want    []int
	}{
		{"a plain number replaces", []string{"3"}, []int{1, 2}, []int{3}},
		{"several plain numbers replace", []string{"3", "4"}, []int{1}, []int{3, 4}},
		{"a plus adds", []string{"+3"}, []int{1}, []int{1, 3}},
		{"a minus takes away", []string{"-1"}, []int{1, 2}, []int{2}},
		{"both signs at once", []string{"+3", "-1"}, []int{1, 2}, []int{2, 3}},
		{"a minus of what was not there changes nothing", []string{"-9"}, []int{1}, []int{1}},
		{"nothing named clears the set", []string{""}, []int{1, 2}, []int{}},
		{"the same number twice is one blocker", []string{"3", "3"}, nil, []int{3}},
		{"a comma separates, the way labels are given", []string{"3,4"}, nil, []int{3, 4}},
		{"adding to nothing", []string{"+3"}, nil, []int{3}},
	} {
		got, err := changedBlockers(tc.given, tc.current, changing)
		if err != nil {
			t.Errorf("%s: changedBlockers: %v", tc.what, err)
			continue
		}
		if len(got) != len(tc.want) {
			t.Errorf("%s: got %v, want %v", tc.what, got, tc.want)
			continue
		}
		for i := range tc.want {
			if got[i] != tc.want[i] {
				t.Errorf("%s: got %v, want %v", tc.what, got, tc.want)
				break
			}
		}
	}
}

// What a blocking change is turned down for, before anything is written.
func TestChangedBlockers_refusesWhatItCannotRead(t *testing.T) {
	for _, tc := range []struct {
		what  string
		given []string
		at    bool
		want  error
	}{
		{"a word is not a task number", []string{"soon"}, changing, ErrBlockerRequired},
		{"zero is not a task number", []string{"0"}, changing, ErrBlockerRequired},
		{"a negative that is not a sign", []string{"-x"}, changing, ErrBlockerRequired},
		{"replacing and adjusting at once", []string{"3", "+4"}, changing, ErrBlockersBothWays},
		{"a sign at creation", []string{"+3"}, creating, ErrBlockerSignAtBirth},
	} {
		_, err := changedBlockers(tc.given, nil, tc.at)
		if !errors.Is(err, tc.want) {
			t.Errorf("%s: err = %v, want %v", tc.what, err, tc.want)
		}
	}
}

// A ring in which nothing could ever start is refused, however long the way
// round. The walk asks what blocks each task it reaches, which is the one thing
// here that cannot be read off the task in hand.
func TestMayBlock_refusesARingNothingCouldEverStart(t *testing.T) {
	// Task 2 waits on 3, and 3 waits on 4. Nothing waits on 1.
	blocking := map[int][]int{2: {3}, 3: {4}}
	asked := func(n int) ([]int, error) { return blocking[n], nil }

	for _, tc := range []struct {
		what     string
		number   int
		blockers []int
		want     error
	}{
		{"itself", 1, []int{1}, ErrBlockedByItself},
		{"the far end of the chain waits on it", 4, []int{2}, ErrBlockingCycle},
		{"a chain that closes on nothing", 1, []int{2}, nil},
		{"nothing named at all", 1, nil, nil},
	} {
		err := mayBlock(tc.number, tc.blockers, asked)
		if tc.want == nil && err != nil {
			t.Errorf("%s: err = %v, want none", tc.what, err)
		}
		if tc.want != nil && !errors.Is(err, tc.want) {
			t.Errorf("%s: err = %v, want %v", tc.what, err, tc.want)
		}
	}
}

// What blocks a task is asked of the store, so a store that cannot answer stops
// the walk rather than being read as "nothing blocks it".
func TestMayBlock_handsBackWhatTheStoreCouldNotAnswer(t *testing.T) {
	sulking := errors.New("the store cannot be read")
	err := mayBlock(1, []int{2}, func(int) ([]int, error) { return nil, sulking })
	if !errors.Is(err, sulking) {
		t.Errorf("err = %v, want the store's own", err)
	}
}

// blocked builds the facts mayLeave and freedFrom read, without a store.
func blocked(number int, state domain.State, blockers []int, because string) taskFacts {
	a := &task.AsTold{}
	a.Numbered(number)
	a.InColumn(state)
	for _, one := range blockers {
		a.BlockedBy(one)
	}
	a.BlockersTold()
	a.Because(because)
	return factsOf(a.Task())
}

// The two columns that claim the work is yours are refused while something is
// still in the way; every other column stays reachable.
func TestMayLeave_refusesOnlyTheColumnsThatClaimTheWork(t *testing.T) {
	for _, tc := range []struct {
		what    string
		was     taskFacts
		becomes taskFacts
		refused bool
	}{
		{"into todo, still blocked",
			blocked(1, domain.Blocked, []int{2}, ""),
			blocked(1, domain.Todo, []int{2}, ""), true},
		{"into current, blocked by words",
			blocked(1, domain.Blocked, nil, "the client"),
			blocked(1, domain.Current, nil, "the client"), true},
		{"into done, still blocked",
			blocked(1, domain.Blocked, []int{2}, ""),
			blocked(1, domain.Done, []int{2}, ""), false},
		{"into todo with nothing recorded",
			blocked(1, domain.Blocked, nil, ""),
			blocked(1, domain.Todo, nil, ""), false},
		{"into todo from somewhere that was never blocked",
			blocked(1, domain.Backlog, []int{2}, ""),
			blocked(1, domain.Todo, []int{2}, ""), false},
	} {
		err := mayLeave(tc.was, tc.becomes)
		if tc.refused && err == nil {
			t.Errorf("%s: it was allowed", tc.what)
		}
		if !tc.refused && err != nil {
			t.Errorf("%s: it was refused: %v", tc.what, err)
		}
	}
}

// A refusal is an instruction: it names what is in the way and the flags that
// would take each one away. Without that, somebody holding three blockers has
// to work out which is which.
func TestMayLeave_saysWhatIsInTheWayAndHowToTakeItAway(t *testing.T) {
	err := mayLeave(
		blocked(1, domain.Blocked, []int{2, 3}, "the client's answer"),
		blocked(1, domain.Todo, []int{2, 3}, "the client's answer"))
	if err == nil {
		t.Fatal("it was allowed")
	}
	for _, said := range []string{"#2", "#3", "the client's answer",
		"task 2", "task 3", "the reason it is waiting"} {
		if !strings.Contains(err.Error(), said) {
			t.Errorf("the refusal does not say %q: %v", said, err)
		}
	}
}

// Taking the last thing away puts a task back in todo. Taking one of two does
// not, and neither does anything happening to a task nobody said was blocked.
func TestFreedFrom_putsATaskBackOnlyWhenTheLastThingGoes(t *testing.T) {
	for _, tc := range []struct {
		what    string
		was     taskFacts
		becomes taskFacts
		freed   bool
	}{
		{"the last blocker",
			blocked(1, domain.Blocked, []int{2}, ""),
			blocked(1, domain.Blocked, nil, ""), true},
		{"the last words",
			blocked(1, domain.Blocked, nil, "the client"),
			blocked(1, domain.Blocked, nil, ""), true},
		{"one of two",
			blocked(1, domain.Blocked, []int{2, 3}, ""),
			blocked(1, domain.Blocked, []int{3}, ""), false},
		{"the blocker, with words left",
			blocked(1, domain.Blocked, []int{2}, "the client"),
			blocked(1, domain.Blocked, nil, "the client"), false},
		{"a task that was not blocked",
			blocked(1, domain.Todo, []int{2}, ""),
			blocked(1, domain.Todo, nil, ""), false},
		{"a task leaving the column itself",
			blocked(1, domain.Blocked, []int{2}, ""),
			blocked(1, domain.Done, nil, ""), false},
	} {
		freed, put := freedFrom(tc.was, tc.becomes)
		if put != tc.freed {
			t.Errorf("%s: freed = %v, want %v", tc.what, put, tc.freed)
			continue
		}
		if put && freed != domain.Todo {
			t.Errorf("%s: it landed in %q, want todo", tc.what, freed)
		}
	}
}

// The mark and the return read the words differently, and the difference is who
// is doing the reading.
//
// Cairn never concludes anything from words it cannot check, so a task blocked
// by words is never marked as having nothing in its way. But somebody erasing
// those words has said the thing is out of the way, so that erasure takes the
// task back to todo like any other last thing. One is a guess cairn is not
// allowed to make; the other is a sentence somebody typed.
func TestFreedFrom_actsOnTheWordsBeingErasedAndNeverOnTheirMeaning(t *testing.T) {
	if _, put := freedFrom(
		blocked(1, domain.Blocked, nil, "the client"),
		blocked(1, domain.Blocked, nil, "the client")); put {
		t.Error("a task moved with its words untouched")
	}
	if _, put := freedFrom(
		blocked(1, domain.Blocked, nil, "the client"),
		blocked(1, domain.Blocked, nil, "")); !put {
		t.Error("erasing the last thing in its way left it blocked")
	}
}

// A blocker naming a task the store does not hold is refused rather than kept:
// a link to nothing is a fact that was never true.
func TestCanBeBlocked_refusesATaskTheStoreDoesNotHold(t *testing.T) {
	e := openEngine(t)
	one := add(t, e, "Ship it")

	a := &task.AsTold{}
	one.Tell(a)
	a.BlockedBy(9)
	a.BlockersTold()
	if err := e.canBeBlocked(a.Task()); err == nil {
		t.Error("a blocker nobody added was accepted")
	}

	if err := e.canBeBlocked(one); err != nil {
		t.Errorf("a task blocked by nothing was refused: %v", err)
	}
}

// The whole way round, through the engine: a task blocked, then the last thing
// taken away. Walked here as well as in the scenarios, because the column cairn
// puts it back in is written by this package and nothing in this package
// otherwise reaches it.
func TestEdit_putsATaskBackInTodoWhenTheLastThingGoes(t *testing.T) {
	e := openEngine(t)
	spec := add(t, e, "Write the spec")
	shipIt := add(t, e, "Ship it")
	blocking := factsOf(spec).number
	number := factsOf(shipIt).number

	one := []string{"1"}
	if _, err := e.Edit(number, Changes{Blocker: &one}); err != nil {
		t.Fatalf("naming a blocker: %v", err)
	}
	stopped := string(domain.Blocked)
	if _, err := e.Edit(number, Changes{State: &stopped}); err != nil {
		t.Fatalf("stopping it: %v", err)
	}

	// Still blocked, so the two columns that claim the work is yours are shut.
	working := string(domain.Current)
	if _, err := e.Edit(number, Changes{State: &working}); !errors.Is(err, ErrStillBlocked) {
		t.Errorf("err = %v, want it to say the task is still blocked", err)
	}

	taken := []string{fmt.Sprintf("-%d", blocking)}
	freed, err := e.Edit(number, Changes{Blocker: &taken})
	if err != nil {
		t.Fatalf("taking the last blocker away: %v", err)
	}
	if got := factsOf(freed).state; got != domain.Todo {
		t.Errorf("it ended up in %q, want todo", got)
	}

	// And the column cairn chose is written down, so a history says how the
	// task got there rather than leaving a reader to guess.
	acts, err := e.History(number)
	if err != nil {
		t.Fatalf("History: %v", err)
	}
	if len(acts) == 0 || !strings.Contains(trace.List(acts[len(acts)-1:]), string(domain.Todo)) {
		t.Errorf("the last act does not say it went back to todo: %v", acts)
	}
}

// A blocker cairn cannot read stops the edit, and the reason comes back rather
// than a task quietly carrying no blocker at all.
func TestEdit_handsBackABlockerItCannotRead(t *testing.T) {
	e := openEngine(t)
	one := add(t, e, "Ship it")
	soon := []string{"soon"}
	if _, err := e.Edit(factsOf(one).number, Changes{Blocker: &soon}); !errors.Is(err, ErrBlockerRequired) {
		t.Errorf("err = %v, want it to say a blocker is a task number", err)
	}
}

// A history says which task was in the way, not merely that one was. Written
// the way a card writes it, so a number read here can be taken straight to
// `task show`.
func TestEdit_writesDownWhichTaskWasInTheWay(t *testing.T) {
	e := openEngine(t)
	spec := add(t, e, "Write the spec")
	shipIt := add(t, e, "Ship it")
	blocking := fmt.Sprintf("#%d", factsOf(spec).number)

	named := []string{fmt.Sprint(factsOf(spec).number)}
	if _, err := e.Edit(factsOf(shipIt).number, Changes{Blocker: &named}); err != nil {
		t.Fatalf("naming a blocker: %v", err)
	}
	acts, err := e.History(factsOf(shipIt).number)
	if err != nil {
		t.Fatalf("History: %v", err)
	}
	if written := trace.List(acts); !strings.Contains(written, blocking) {
		t.Errorf("the history does not say %s was in the way:\n%s", blocking, written)
	}
}

// The engine is behind four surfaces and the refusal spoke to one of them. A
// person on the board or in the window was told to type --blocker -4, which is
// not something either of them can be told.
func TestTheRefusalDoesNotSpeakInCommandLineFlags(t *testing.T) {
	e := openEngine(t)
	number := factsOf(add(t, e, "Waiting")).number
	blocked := string(domain.Blocked)
	because := "the lawyer"
	if _, err := e.Edit(number, Changes{State: &blocked, Because: &because}); err != nil {
		t.Fatal(err)
	}
	working := string(domain.Current)
	_, err := e.Edit(number, Changes{State: &working})
	if err == nil {
		t.Fatal("a blocked task left its column")
	}
	if strings.Contains(err.Error(), "--") {
		t.Fatalf("the refusal names a command line flag: %q", err)
	}
}
