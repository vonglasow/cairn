package engine

// The order and the focus, asked of the engine: the rules it refuses on, and
// what it answers when it does not.

import (
	"errors"
	"slices"
	"strconv"
	"strings"
	"testing"

	"cairn/internal/task"
	"cairn/internal/trace"
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
	held, numbers, err := e.InFocus()
	if err != nil {
		t.Fatalf("InFocus: %v", err)
	}
	if len(numbers) != 0 || len(held) != 0 {
		t.Errorf("InFocus = %v (%d tasks), want nothing", numbers, len(held))
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
	held, numbers, err := e.InFocus()
	if err != nil {
		t.Fatalf("InFocus: %v", err)
	}
	if !slices.Equal(numbers, []int{second}) {
		t.Errorf("InFocus = %v, want %v", numbers, []int{second})
	}
	if len(held) != 1 || factsOf(held[0]).title != "Second" {
		t.Errorf("it handed back %d tasks, want the one titled %q", len(held), "Second")
	}
}

// Several at once, in the order they were taken up. One at a time is what a
// person does; an agent session is started per task, and several run together.
func TestPutInFocus_holdsSeveralAtOnce(t *testing.T) {
	e := openEngine(t)
	first, second, third := three(t, e)
	for _, n := range []int{third, first} {
		if _, err := e.PutInFocus(n); err != nil {
			t.Fatalf("PutInFocus(%d): %v", n, err)
		}
	}
	if got, err := e.Focus(); err != nil || !slices.Equal(got, []int{third, first}) {
		t.Errorf("Focus = %v (%v), want %v", got, err, []int{third, first})
	}
	// And taking up what is already held changes nothing at all.
	if _, err := e.PutInFocus(third); err != nil {
		t.Fatalf("PutInFocus again: %v", err)
	}
	if got, err := e.Focus(); err != nil || !slices.Equal(got, []int{third, first}) {
		t.Errorf("Focus = %v (%v) after taking up what was held, want %v", got, err, []int{third, first})
	}
	_ = second
}

func TestPutInFocus_refusesATaskThatIsNotThere(t *testing.T) {
	if _, err := openEngine(t).PutInFocus(99); !errors.Is(err, ErrNoSuchTask) {
		t.Errorf("PutInFocus(99) = %v, want %v", err, ErrNoSuchTask)
	}
}

// Letting go of one leaves the rest held. Letting go of one that was never
// held is not a failure: the state asked for already holds.
func TestLetGoOf_releasesOneAndKeepsTheRest(t *testing.T) {
	e := openEngine(t)
	first, second, _ := three(t, e)
	for _, n := range []int{first, second} {
		if _, err := e.PutInFocus(n); err != nil {
			t.Fatalf("PutInFocus(%d): %v", n, err)
		}
	}
	if err := e.LetGoOf(first); err != nil {
		t.Fatalf("LetGoOf: %v", err)
	}
	if got, err := e.Focus(); err != nil || !slices.Equal(got, []int{second}) {
		t.Errorf("Focus = %v (%v), want %v", got, err, []int{second})
	}
	if err := e.LetGoOf(first); err != nil {
		t.Errorf("LetGoOf a task that was not held = %v, want nothing wrong", err)
	}
	if got, err := e.Focus(); err != nil || !slices.Equal(got, []int{second}) {
		t.Errorf("Focus = %v (%v), want %v", got, err, []int{second})
	}
}

func TestLetGo_leavesNothingInFocus(t *testing.T) {
	e := openEngine(t)
	first, second, _ := three(t, e)
	for _, n := range []int{first, second} {
		if _, err := e.PutInFocus(n); err != nil {
			t.Fatalf("PutInFocus(%d): %v", n, err)
		}
	}
	if err := e.LetGo(); err != nil {
		t.Fatalf("LetGo: %v", err)
	}
	if got, err := e.Focus(); err != nil || len(got) != 0 {
		t.Errorf("Focus = %v (%v), want nothing", got, err)
	}
}

// The focus points at tasks that may go away. One goes with it, and InFocus
// answers with the rest rather than a task it could not read.
func TestInFocus_dropsATaskOnceItIsDeleted(t *testing.T) {
	e := openEngine(t)
	first, second, _ := three(t, e)
	for _, n := range []int{first, second} {
		if _, err := e.PutInFocus(n); err != nil {
			t.Fatalf("PutInFocus(%d): %v", n, err)
		}
	}
	if err := e.Delete(first); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	held, numbers, err := e.InFocus()
	if err != nil {
		t.Fatalf("InFocus: %v", err)
	}
	if !slices.Equal(numbers, []int{second}) {
		t.Errorf("InFocus = %v, want %v once the first was deleted", numbers, []int{second})
	}
	if len(held) != 1 {
		t.Errorf("it handed back %d tasks, want one", len(held))
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
	answer, err := listing(On{Engine: e}, Input{})
	if err != nil {
		t.Fatalf("listing: %v", err)
	}
	if len(answer.Tasks) != 3 {
		t.Errorf("it listed %d tasks, want 3", len(answer.Tasks))
	}
	if answer.InFocus == nil || !slices.Equal(*answer.InFocus, []int{second}) {
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
	if len(answer.Held) != 0 {
		t.Errorf("it answered with %d tasks, want none", len(answer.Held))
	}
	if answer.InFocus == nil || len(*answer.InFocus) != 0 {
		t.Errorf("InFocus = %v, want nothing", answer.InFocus)
	}

	if _, err := e.PutInFocus(first); err != nil {
		t.Fatalf("PutInFocus: %v", err)
	}
	answer, err = whatIsInFocus(On{Engine: e})
	if err != nil {
		t.Fatalf("whatIsInFocus: %v", err)
	}
	if len(answer.Held) != 1 || factsOf(answer.Held[0]).title != "First" {
		t.Errorf("it answered with %d tasks, want the one titled %q", len(answer.Held), "First")
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
	if put.InFocus == nil || !slices.Equal(*put.InFocus, []int{second}) {
		t.Errorf("InFocus = %v, want %d", put.InFocus, second)
	}

	asked, err := TaskFocus.Run(On{Engine: e}, Input{Given: map[string]string{}})
	if err != nil {
		t.Fatalf("TaskFocus.Run with no number: %v", err)
	}
	if len(asked.Held) != 1 || factsOf(asked.Held[0]).title != "Second" {
		t.Errorf("asking answered with %d tasks, want the one in focus", len(asked.Held))
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
	if answer.InFocus == nil || len(*answer.InFocus) != 0 {
		t.Errorf("InFocus = %v, want nothing", answer.InFocus)
	}
	if len(answer.Held) != 0 {
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

// Letting go of everything writes the release on every task that was held, and
// not only on the first: a release that named one would leave the others'
// stretches open for ever.
func TestLetGo_writesTheReleaseOnEveryTaskHeld(t *testing.T) {
	e := openEngine(t)
	first, second, _ := three(t, e)
	for _, n := range []int{first, second} {
		if _, err := e.PutInFocus(n); err != nil {
			t.Fatalf("PutInFocus(%d): %v", n, err)
		}
	}
	if err := e.LetGo(); err != nil {
		t.Fatalf("LetGo: %v", err)
	}
	for _, n := range []int{first, second} {
		released := 0
		for _, a := range actsOf(t, e, n) {
			if did, _, _, _ := didOf(a); did == trace.FocusReleased {
				released++
			}
		}
		if released != 1 {
			t.Errorf("task %d carries %d releases, want one", n, released)
		}
	}
}

// Letting go of one task through the operation names it, and leaves the rest.
func TestTaskUnfocus_letsGoOfTheTaskItWasGiven(t *testing.T) {
	e := openEngine(t)
	first, second, _ := three(t, e)
	for _, n := range []int{first, second} {
		if _, err := e.PutInFocus(n); err != nil {
			t.Fatalf("PutInFocus(%d): %v", n, err)
		}
	}
	answer, err := TaskUnfocus.Run(On{Engine: e}, Input{
		Given: map[string]string{"number": itoaEngine(first)},
	})
	if err != nil {
		t.Fatalf("TaskUnfocus.Run: %v", err)
	}
	if answer.InFocus == nil || !slices.Equal(*answer.InFocus, []int{second}) {
		t.Errorf("InFocus = %v, want %v", answer.InFocus, []int{second})
	}
	if len(answer.Held) != 1 || factsOf(answer.Held[0]).title != "Second" {
		t.Errorf("it answered with %d tasks, want the one still held", len(answer.Held))
	}
}
