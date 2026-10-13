package engine

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"cairn/internal/domain"
	"cairn/internal/task"
	"cairn/internal/trace"
)

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

func TestMayBlock_refusesARingNothingCouldEverStart(t *testing.T) {
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

func TestMayBlock_handsBackWhatTheStoreCouldNotAnswer(t *testing.T) {
	sulking := errors.New("the store cannot be read")
	err := mayBlock(1, []int{2}, func(int) ([]int, error) { return nil, sulking })
	if !errors.Is(err, sulking) {
		t.Errorf("err = %v, want the store's own", err)
	}
}

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

func TestEdit_putsATaskBackInTodoWhenTheLastThingGoes(t *testing.T) {
	e := openEngine(t)
	spec := add(t, e, "Write the spec")
	shipIt := add(t, e, "Ship it")
	blocking := factsOf(spec).number
	number := factsOf(shipIt).number

	one := []string{"1"}
	if _, err := e.edit(number, Changes{Blocker: &one}); err != nil {
		t.Fatalf("naming a blocker: %v", err)
	}
	stopped := string(domain.Blocked)
	if _, err := e.edit(number, Changes{State: &stopped}); err != nil {
		t.Fatalf("stopping it: %v", err)
	}

	working := string(domain.Current)
	if _, err := e.edit(number, Changes{State: &working}); !errors.Is(err, ErrStillBlocked) {
		t.Errorf("err = %v, want it to say the task is still blocked", err)
	}

	taken := []string{fmt.Sprintf("-%d", blocking)}
	freed, err := e.edit(number, Changes{Blocker: &taken})
	if err != nil {
		t.Fatalf("taking the last blocker away: %v", err)
	}
	if got := factsOf(freed).state; got != domain.Todo {
		t.Errorf("it ended up in %q, want todo", got)
	}

	acts, err := e.historyOf(number)
	if err != nil {
		t.Fatalf("History: %v", err)
	}
	if len(acts) == 0 || !strings.Contains(trace.List(acts[len(acts)-1:]), string(domain.Todo)) {
		t.Errorf("the last act does not say it went back to todo: %v", acts)
	}
}

func TestEdit_handsBackABlockerItCannotRead(t *testing.T) {
	e := openEngine(t)
	one := add(t, e, "Ship it")
	soon := []string{"soon"}
	if _, err := e.edit(factsOf(one).number, Changes{Blocker: &soon}); !errors.Is(err, ErrBlockerRequired) {
		t.Errorf("err = %v, want it to say a blocker is a task number", err)
	}
}

func TestEdit_writesDownWhichTaskWasInTheWay(t *testing.T) {
	e := openEngine(t)
	spec := add(t, e, "Write the spec")
	shipIt := add(t, e, "Ship it")
	blocking := fmt.Sprintf("#%d", factsOf(spec).number)

	named := []string{fmt.Sprint(factsOf(spec).number)}
	if _, err := e.edit(factsOf(shipIt).number, Changes{Blocker: &named}); err != nil {
		t.Fatalf("naming a blocker: %v", err)
	}
	acts, err := e.historyOf(factsOf(shipIt).number)
	if err != nil {
		t.Fatalf("History: %v", err)
	}
	if written := trace.List(acts); !strings.Contains(written, blocking) {
		t.Errorf("the history does not say %s was in the way:\n%s", blocking, written)
	}
}

func TestTheRefusalDoesNotSpeakInCommandLineFlags(t *testing.T) {
	e := openEngine(t)
	number := factsOf(add(t, e, "Waiting")).number
	blocked := string(domain.Blocked)
	because := "the lawyer"
	if _, err := e.edit(number, Changes{State: &blocked, Because: &because}); err != nil {
		t.Fatal(err)
	}
	working := string(domain.Current)
	_, err := e.edit(number, Changes{State: &working})
	if err == nil {
		t.Fatal("a blocked task left its column")
	}
	if strings.Contains(err.Error(), "--") {
		t.Fatalf("the refusal names a command line flag: %q", err)
	}
}
