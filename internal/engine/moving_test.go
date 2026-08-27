package engine

// The order and the focus, asked of the engine: the rules it refuses on, and
// what it answers when it does not.

import (
	"errors"
	"strconv"
	"strings"
	"testing"

	"cairn/internal/store"
	"cairn/internal/task"
)

// titlesHeld reads back the titles the engine lists, in the one order there is.
func titlesHeld(t *testing.T, e *Engine) []string {
	t.Helper()
	all, err := e.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	titles := make([]string, 0, len(all))
	for _, one := range all {
		titles = append(titles, factsOf(one).title)
	}
	return titles
}

// wantHeld fails unless the engine lists exactly these titles, in this order.
func wantHeld(t *testing.T, e *Engine, want ...string) {
	t.Helper()
	if got := titlesHeld(t, e); strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("the order is %v, want %v", got, want)
	}
}

// three adds three tasks and hands back their numbers.
func three(t *testing.T, e *Engine) (int, int, int) {
	t.Helper()
	return factsOf(add(t, e, "First")).number,
		factsOf(add(t, e, "Second")).number,
		factsOf(add(t, e, "Third")).number
}

func TestMove_putsATaskWhereItWasTold(t *testing.T) {
	for _, tc := range []struct {
		what string
		at   Place
		want []string
	}{
		{"first", First, []string{"Third", "First", "Second"}},
		{"last", Last, []string{"First", "Second", "Third"}},
	} {
		e := openEngine(t)
		_, _, third := three(t, e)
		if err := e.Move(third, tc.at, 0); err != nil {
			t.Fatalf("Move %s: %v", tc.what, err)
		}
		wantHeld(t, e, tc.want...)
	}
}

func TestMove_putsATaskBehindTheOneNamed(t *testing.T) {
	e := openEngine(t)
	first, second, _ := three(t, e)
	if err := e.Move(first, Behind, second); err != nil {
		t.Fatalf("Move behind: %v", err)
	}
	wantHeld(t, e, "Second", "First", "Third")
}

// The rule lives in the task and is asked before the store is: every refusal
// here comes back from the notion rather than from a failed statement.
func TestMove_refusesWhatTheNotionRefuses(t *testing.T) {
	e := openEngine(t)
	first, _, _ := three(t, e)
	toDo, err := e.Edit(factsOf(add(t, e, "Chosen")).number, Changes{State: of("todo")})
	if err != nil {
		t.Fatalf("Edit: %v", err)
	}
	under, err := e.Add("A sister", Changes{Parent: of(first)})
	if err != nil {
		t.Fatalf("Add a sub-task: %v", err)
	}

	for _, tc := range []struct {
		what   string
		moving int
		target int
		want   error
	}{
		{"behind itself", first, first, ErrMovedBehindItself},
		{"behind one in another column", first, factsOf(toDo).number, ErrOtherColumn},
		{"a sub-task behind a root task", factsOf(under).number, first, ErrNotAmongItsSisters},
	} {
		if err := e.Move(tc.moving, Behind, tc.target); !errors.Is(err, tc.want) {
			t.Errorf("moving %s = %v, want %v", tc.what, err, tc.want)
		}
	}
}

func TestMove_refusesATaskThatIsNotThere(t *testing.T) {
	e := openEngine(t)
	first, _, _ := three(t, e)
	for _, tc := range []struct {
		what           string
		moving, target int
	}{
		{"the task moving", 99, first},
		{"the task named", first, 99},
	} {
		if err := e.Move(tc.moving, Behind, tc.target); !errors.Is(err, ErrNoSuchTask) {
			t.Errorf("%s is not there: Move = %v, want %v", tc.what, err, ErrNoSuchTask)
		}
	}
	if err := e.Move(99, First, 0); !errors.Is(err, ErrNoSuchTask) {
		t.Errorf("Move first on a task that is not there = %v, want %v", err, ErrNoSuchTask)
	}
}

func TestFocus_isNothingToBeginWith(t *testing.T) {
	e := openEngine(t)
	add(t, e, "The only one")
	held, number, err := e.InFocus()
	if err != nil {
		t.Fatalf("InFocus: %v", err)
	}
	if number != store.NoFocus {
		t.Errorf("InFocus = %d, want %d", number, store.NoFocus)
	}
	if got := factsOf(held).title; got != "" {
		t.Errorf("it handed back the task %q, want none", got)
	}
}

func TestPutInFocus_handsBackTheTaskBeingWorkedOn(t *testing.T) {
	e := openEngine(t)
	_, second, _ := three(t, e)
	put, err := e.PutInFocus(second)
	if err != nil {
		t.Fatalf("PutInFocus: %v", err)
	}
	if got := factsOf(put).title; got != "Second" {
		t.Errorf("PutInFocus handed back %q, want %q", got, "Second")
	}
	held, number, err := e.InFocus()
	if err != nil {
		t.Fatalf("InFocus: %v", err)
	}
	if number != second || factsOf(held).title != "Second" {
		t.Errorf("InFocus = %d (%q), want %d (%q)",
			number, factsOf(held).title, second, "Second")
	}
}

func TestPutInFocus_refusesATaskThatIsNotThere(t *testing.T) {
	if _, err := openEngine(t).PutInFocus(99); !errors.Is(err, ErrNoSuchTask) {
		t.Errorf("PutInFocus(99) = %v, want %v", err, ErrNoSuchTask)
	}
}

func TestLetGo_leavesNothingInFocus(t *testing.T) {
	e := openEngine(t)
	first, _, _ := three(t, e)
	if _, err := e.PutInFocus(first); err != nil {
		t.Fatalf("PutInFocus: %v", err)
	}
	if err := e.LetGo(); err != nil {
		t.Fatalf("LetGo: %v", err)
	}
	if number, err := e.Focus(); err != nil || number != store.NoFocus {
		t.Errorf("Focus = %d (%v), want %d", number, err, store.NoFocus)
	}
}

// The focus points at a task that may go away. It goes with it, and InFocus
// answers "none" rather than a task it could not read.
func TestInFocus_answersNoneOnceTheTaskIsDeleted(t *testing.T) {
	e := openEngine(t)
	first, _, _ := three(t, e)
	if _, err := e.PutInFocus(first); err != nil {
		t.Fatalf("PutInFocus: %v", err)
	}
	if err := e.Delete(first); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, number, err := e.InFocus(); err != nil || number != store.NoFocus {
		t.Errorf("InFocus = %d (%v), want %d", number, err, store.NoFocus)
	}
}

func TestPlaceFrom_readsWhereAMoveWasTold(t *testing.T) {
	for _, tc := range []struct {
		what   string
		in     Input
		at     Place
		target int
	}{
		{"first", Input{Given: map[string]string{"first": "true"}}, First, 0},
		{"last", Input{Given: map[string]string{"last": "true"}}, Last, 0},
		{"behind another", Input{Given: map[string]string{"after": "2"}}, Behind, 2},
	} {
		at, target, err := placeFrom(tc.in)
		if err != nil {
			t.Errorf("%s: placeFrom = %v", tc.what, err)
			continue
		}
		if at != tc.at || target != tc.target {
			t.Errorf("%s: placeFrom = %v behind %d, want %v behind %d",
				tc.what, at, target, tc.at, tc.target)
		}
	}
}

// A flag given as false was named, and reading "named" as a request would turn
// `--first=false` into moving a task first.
func TestPlaceFrom_refusesWhatNamesNoPlaceOrSeveral(t *testing.T) {
	for _, tc := range []struct {
		what string
		in   Input
		want error
	}{
		{"nothing at all", Input{Given: map[string]string{}}, ErrNowhereToMove},
		{"a flag given as false", Input{Given: map[string]string{"first": "false"}}, ErrNowhereToMove},
		{"first and last at once", Input{Given: map[string]string{"first": "true", "last": "true"}}, ErrMoreThanOnePlace},
		{"first and behind at once", Input{Given: map[string]string{"first": "true", "after": "2"}}, ErrMoreThanOnePlace},
	} {
		if _, _, err := placeFrom(tc.in); !errors.Is(err, tc.want) {
			t.Errorf("%s: placeFrom = %v, want %v", tc.what, err, tc.want)
		}
	}
}

func TestYes_readsOnlyAFlagGivenAsTrue(t *testing.T) {
	for _, tc := range []struct {
		given string
		want  bool
	}{{"true", true}, {"false", false}, {"", false}, {"yes", false}} {
		in := Input{Given: map[string]string{"first": tc.given}}
		if got := in.Yes("first"); got != tc.want {
			t.Errorf("Yes on %q = %v, want %v", tc.given, got, tc.want)
		}
	}
	if (Input{}).Yes("first") {
		t.Error("Yes on a flag nobody named = true, want false")
	}
}

// Both the listing and the focus are answered from the same two calls, so what
// they hand back is asserted where an operation would read it.
func TestListing_carriesTheOrderAndWhatIsInFocus(t *testing.T) {
	e := openEngine(t)
	_, second, _ := three(t, e)
	if _, err := e.PutInFocus(second); err != nil {
		t.Fatalf("PutInFocus: %v", err)
	}
	answer, err := listing(On{Engine: e})
	if err != nil {
		t.Fatalf("listing: %v", err)
	}
	if len(answer.Tasks) != 3 {
		t.Errorf("it listed %d tasks, want 3", len(answer.Tasks))
	}
	if answer.InFocus == nil || *answer.InFocus != second {
		t.Errorf("InFocus = %v, want %d", answer.InFocus, second)
	}
}

func TestWhatIsInFocus_answersTheTaskOrSaysThereIsNone(t *testing.T) {
	e := openEngine(t)
	first, _, _ := three(t, e)

	answer, err := whatIsInFocus(On{Engine: e})
	if err != nil {
		t.Fatalf("whatIsInFocus: %v", err)
	}
	if answer.Task != nil {
		t.Errorf("it answered with a task, want none: %v", factsOf(*answer.Task).title)
	}
	if answer.InFocus == nil || *answer.InFocus != store.NoFocus {
		t.Errorf("InFocus = %v, want %d", answer.InFocus, store.NoFocus)
	}

	if _, err := e.PutInFocus(first); err != nil {
		t.Fatalf("PutInFocus: %v", err)
	}
	answer, err = whatIsInFocus(On{Engine: e})
	if err != nil {
		t.Fatalf("whatIsInFocus: %v", err)
	}
	if answer.Task == nil || factsOf(*answer.Task).title != "First" {
		t.Errorf("it answered with %v, want the task titled %q", answer.Task, "First")
	}
}

// A move answers with the order, not with the card of the task moved: the card
// would not show what moved.
func TestTaskMove_answersWithTheOrder(t *testing.T) {
	e := openEngine(t)
	_, _, third := three(t, e)
	answer, err := TaskMove.Run(On{Engine: e}, Input{
		Given: map[string]string{"number": itoaEngine(third), "first": "true"},
	})
	if err != nil {
		t.Fatalf("TaskMove.Run: %v", err)
	}
	if answer.Task != nil {
		t.Error("it answered with a task, want the order")
	}
	if len(answer.Tasks) != 3 || factsOf(answer.Tasks[0]).title != "Third" {
		t.Errorf("it answered with %v, want Third at the head", titlesOfTasks(answer.Tasks))
	}
}

// itoaEngine writes a number the way an argument arrives, since Input speaks
// one language for every kind.
func itoaEngine(n int) string { return strconv.Itoa(n) }

// titlesOfTasks reads the titles of a list of tasks, for a message.
func titlesOfTasks(all []task.Task) []string {
	titles := make([]string, 0, len(all))
	for _, one := range all {
		titles = append(titles, factsOf(one).title)
	}
	return titles
}

// The three operations are run here, not only their engine methods: a Run body
// is where an argument becomes a call, and the lines that check what came back
// are exercised by nothing else.
func TestTaskFocus_answersWithTheTaskOrSaysThereIsNone(t *testing.T) {
	e := openEngine(t)
	_, second, _ := three(t, e)

	put, err := TaskFocus.Run(On{Engine: e}, Input{
		Given: map[string]string{"number": itoaEngine(second)},
	})
	if err != nil {
		t.Fatalf("TaskFocus.Run: %v", err)
	}
	if put.Task == nil || factsOf(*put.Task).title != "Second" {
		t.Errorf("it answered with %v, want the task titled %q", put.Task, "Second")
	}
	if put.InFocus == nil || *put.InFocus != second {
		t.Errorf("InFocus = %v, want %d", put.InFocus, second)
	}

	asked, err := TaskFocus.Run(On{Engine: e}, Input{Given: map[string]string{}})
	if err != nil {
		t.Fatalf("TaskFocus.Run with no number: %v", err)
	}
	if asked.Task == nil || factsOf(*asked.Task).title != "Second" {
		t.Errorf("asking answered with %v, want the task in focus", asked.Task)
	}
}

func TestTaskFocus_refusesATaskThatIsNotThere(t *testing.T) {
	e := openEngine(t)
	add(t, e, "The only one")
	if _, err := TaskFocus.Run(On{Engine: e}, Input{
		Given: map[string]string{"number": "99"},
	}); !errors.Is(err, ErrNoSuchTask) {
		t.Errorf("TaskFocus.Run on a task that is not there = %v, want %v", err, ErrNoSuchTask)
	}
}

func TestTaskUnfocus_answersThatNothingIsInFocus(t *testing.T) {
	e := openEngine(t)
	first, _, _ := three(t, e)
	if _, err := e.PutInFocus(first); err != nil {
		t.Fatalf("PutInFocus: %v", err)
	}
	answer, err := TaskUnfocus.Run(On{Engine: e}, Input{})
	if err != nil {
		t.Fatalf("TaskUnfocus.Run: %v", err)
	}
	if answer.InFocus == nil || *answer.InFocus != store.NoFocus {
		t.Errorf("InFocus = %v, want %d", answer.InFocus, store.NoFocus)
	}
	if answer.Task != nil {
		t.Error("it answered with a task, want nothing in focus")
	}
}

// A move that names no place is refused before the store is touched.
func TestTaskMove_refusesAMoveThatNamesNoPlace(t *testing.T) {
	e := openEngine(t)
	first, _, _ := three(t, e)
	if _, err := TaskMove.Run(On{Engine: e}, Input{
		Given: map[string]string{"number": itoaEngine(first)},
	}); !errors.Is(err, ErrNowhereToMove) {
		t.Errorf("TaskMove.Run with no place = %v, want %v", err, ErrNowhereToMove)
	}
}

// And a move of a task that is not there is refused with the reason a person
// can act on.
func TestTaskMove_refusesATaskThatIsNotThere(t *testing.T) {
	if _, err := TaskMove.Run(On{Engine: openEngine(t)}, Input{
		Given: map[string]string{"number": "99", "first": "true"},
	}); !errors.Is(err, ErrNoSuchTask) {
		t.Errorf("TaskMove.Run on a task that is not there = %v, want %v", err, ErrNoSuchTask)
	}
}
