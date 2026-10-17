package engine

import (
	"errors"
	"slices"
	"strings"
	"testing"
	"time"

	"cairn/internal/domain"
	"cairn/internal/task"
)

type failingFind struct {
	tasks
	err error
}

func (f failingFind) Find(int) (task.Task, error) { return task.Task{}, f.err }

func TestTimed_saysAStoreItCouldNotRead(t *testing.T) {
	e := openEngine(t)
	add(t, e, "one")
	if _, err := TimesheetAdd.Run(On{Engine: e}, given("number", "1", "time", "1h", "on", "2026-09-29")); err != nil {
		t.Fatal(err)
	}
	broken := errors.New("disk on fire")
	e.tasks = failingFind{tasks: e.tasks, err: broken}
	day := time.Date(2026, 9, 29, 0, 0, 0, 0, time.Local)
	if _, _, err := e.timed(day, day); !errors.Is(err, broken) {
		t.Errorf("Timed = %v, want the store's failure", err)
	}
}

type failingClose struct {
	tasks
	err error
}

func (f failingClose) Close() error { return f.err }

func TestDoWith_saysAStoreItCouldNotClose(t *testing.T) {
	broken := errors.New("could not write back")
	open := func(path string) (*Engine, error) {
		e, err := Open(path)
		if err != nil {
			return nil, err
		}
		real := e.tasks
		t.Cleanup(func() { _ = real.Close() })
		e.tasks = failingClose{tasks: real, err: broken}
		return e, nil
	}
	op := Operation{Needs: AStore, Run: func(On, Input) (Answer, error) { return Answer{}, nil }}
	if _, err := doWith(open, op, Places{Store: storeAt(t)}, Input{}, Acting{}); !errors.Is(err, broken) {
		t.Errorf("doWith = %v, want the failure to close", err)
	}
}

func TestCheck_refusesARequiredNumberGivenAsNothing(t *testing.T) {
	if err := TaskShow.Check(given("number", "")); err == nil {
		t.Error("task show accepted an empty number")
	}
	if err := TaskMove.Check(given("number", "1", "after", "")); err != nil {
		t.Errorf("task move refused a number given as nothing where none is required: %v", err)
	}
}

func TestDo_keepsNothingOfAnOperationThatFailedHalfway(t *testing.T) {
	path := storeAt(t)
	failed := errors.New("halfway")
	op := Operation{Thing: "task", Verb: "half", Level: Write, Needs: AStore,
		Run: func(on On, _ Input) (Answer, error) {
			if _, err := on.Engine.add("Write the spec", Changes{}); err != nil {
				return Answer{}, err
			}
			return Answer{}, failed
		}}
	if _, err := Do(op, path, Input{}); !errors.Is(err, failed) {
		t.Fatalf("Do = %v, want the operation's own failure", err)
	}
	answer, err := Do(TaskList, path, Input{})
	if err != nil {
		t.Fatal(err)
	}
	if len(answer.Tasks) != 0 {
		t.Errorf("%d tasks kept from an operation that failed, want none", len(answer.Tasks))
	}
	if _, err := Do(TraceShow, path, given("number", "1")); !errors.Is(err, ErrNoSuchTask) {
		t.Errorf("the history of what was undone = %v, want nothing written down", err)
	}
}

func TestAsk_keepsNothingOfAnOperationThatFailedHalfway(t *testing.T) {
	e := openEngine(t)
	failed := errors.New("halfway")
	op := Operation{Thing: "task", Verb: "half", Level: Write, Needs: AStore,
		Run: func(on On, _ Input) (Answer, error) {
			if _, err := on.Engine.add("Write the spec", Changes{}); err != nil {
				return Answer{}, err
			}
			return Answer{}, failed
		}}
	if _, err := e.Ask(op, Input{}); !errors.Is(err, failed) {
		t.Fatalf("Ask = %v, want the operation's own failure", err)
	}
	all, err := e.tasks.All()
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 0 {
		t.Errorf("%d tasks kept, want none", len(all))
	}
}

type keptBy struct {
	tasks
	told []string
}

func (k *keptBy) Narrowed(n *domain.Narrowing) ([]task.Task, error) {
	n.Tell(k)
	return []task.Task{}, nil
}
func (k *keptBy) InColumn(s string)   { k.told = append(k.told, "column "+s) }
func (k *keptBy) Carrying(p string)   { k.told = append(k.told, "priority "+p) }
func (k *keptBy) DueBefore(string)    {}
func (k *keptBy) Serving(int)         {}
func (k *keptBy) Labelled(l []string) { k.told = append(k.told, "labels "+strings.Join(l, ",")) }
func (k *keptBy) Archived()           {}
func (k *keptBy) WithArchived()       {}

func TestNarrowed_isSaidToWhateverKeepsTheTasks(t *testing.T) {
	kept, err := keeping(Input{
		Given:  map[string]string{"state": "done", "priority": "high"},
		Listed: map[string][]string{"label": {"spec"}},
	}, taskNarrowing)
	if err != nil {
		t.Fatal(err)
	}
	k := &keptBy{}
	e := &Engine{tasks: k}
	if _, err := e.narrowed(kept); err != nil {
		t.Fatal(err)
	}
	if want := []string{"column done", "priority high", "labels spec"}; !slices.Equal(k.told, want) {
		t.Errorf("told %v, want %v", k.told, want)
	}
}
