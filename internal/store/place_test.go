package store

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"cairn/internal/domain"
	"cairn/internal/task"
)

func titlesInOrder(t *testing.T, s *Store) []string {
	t.Helper()
	all, err := s.All()
	if err != nil {
		t.Fatalf("All: %v", err)
	}
	titles := make([]string, 0, len(all))
	for _, one := range all {
		titles = append(titles, factsOf(one).title)
	}
	return titles
}

func wantOrder(t *testing.T, s *Store, want ...string) {
	t.Helper()
	if got := titlesInOrder(t, s); strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("the order is %v, want %v", got, want)
	}
}

func keepInColumn(t *testing.T, s *Store, title string, state domain.State) task.Task {
	t.Helper()
	kept, err := s.Keep(changed(mustBuild(t, title), func(a *task.AsTold) {
		a.InColumn(state)
	}))
	if err != nil {
		t.Fatalf("Keep: %v", err)
	}
	return kept
}

func keepUnder(t *testing.T, s *Store, title string, parent int) task.Task {
	t.Helper()
	kept, err := s.Keep(changed(mustBuild(t, title), func(a *task.AsTold) {
		a.Under(parent)
	}))
	if err != nil {
		t.Fatalf("Keep: %v", err)
	}
	return kept
}

func mustBuild(t *testing.T, title string) task.Task {
	t.Helper()
	built, err := task.NewTask(title)
	if err != nil {
		t.Fatalf("NewTask: %v", err)
	}
	return built
}

func wantEveryGroupRunsOneToN(t *testing.T, s *Store) {
	t.Helper()
	rows, err := s.db.Query(
		`SELECT state, parent, position FROM tasks ORDER BY state, parent, position`)
	if err != nil {
		t.Fatalf("reading the places: %v", err)
	}
	defer func() { _ = rows.Close() }()

	seen := map[string][]int{}
	for rows.Next() {
		var state string
		var parent, place int
		if err := rows.Scan(&state, &parent, &place); err != nil {
			t.Fatalf("reading a place: %v", err)
		}
		seen[fmt.Sprintf("%s under %d", state, parent)] = append(seen[fmt.Sprintf("%s under %d", state, parent)], place)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("reading the places: %v", err)
	}
	for group, places := range seen {
		for i, place := range places {
			if place != i+1 {
				t.Errorf("the places in %s are %v, want 1..%d with no hole",
					group, places, len(places))
				break
			}
		}
	}
}

func TestThePlaces_runOneToNAfterEveryAct(t *testing.T) {
	s := openStore(t)
	first := keep(t, s, "First")
	second := keep(t, s, "Second")
	third := keep(t, s, "Third")
	sister := keepUnder(t, s, "A sister", factsOf(first).number)
	keepUnder(t, s, "Another sister", factsOf(first).number)
	other := keep(t, s, "Another parent")
	wantEveryGroupRunsOneToN(t, s)

	for _, act := range []struct {
		what string
		do   func() error
	}{
		{"moving a task to the head", func() error { return onlyErr(s.MoveFirst(factsOf(second).number)) }},
		{"moving a task to the tail", func() error { return onlyErr(s.MoveLast(factsOf(first).number)) }},
		{"moving a task behind another", func() error {
			return onlyErr(s.MoveAfter(factsOf(second).number, factsOf(first).number))
		}},
		{"changing a task's column", func() error {
			return s.Update(changed(second, func(a *task.AsTold) { a.InColumn(domain.Todo) }))
		}},
		{"giving a sub-task another parent", func() error {
			return s.Update(changed(sister, func(a *task.AsTold) { a.Under(factsOf(other).number) }))
		}},
		{"deleting a task from the middle of its group", func() error {
			return s.Delete(factsOf(third).number)
		}},
		{"deleting the last task of a group", func() error { return s.Delete(factsOf(second).number) }},
	} {
		if err := act.do(); err != nil {
			t.Fatalf("%s: %v", act.what, err)
		}
		t.Run(act.what, func(t *testing.T) { wantEveryGroupRunsOneToN(t, s) })
	}
}

func TestKeep_putsATaskBehindTheOnesItSitsWith(t *testing.T) {
	s := openStore(t)
	keep(t, s, "First")
	keep(t, s, "Second")
	keep(t, s, "Third")
	wantOrder(t, s, "First", "Second", "Third")
}

func TestAll_readsTheColumnsInTheOrderTheyAreRead(t *testing.T) {
	s := openStore(t)
	keepInColumn(t, s, "Finished", domain.Done)
	keepInColumn(t, s, "Under way", domain.Current)
	keepInColumn(t, s, "Chosen", domain.Todo)
	keepInColumn(t, s, "Waiting", domain.Backlog)
	wantOrder(t, s, "Waiting", "Chosen", "Under way", "Finished")
}

func TestMoveFirst_putsATaskAtTheHeadOfItsColumn(t *testing.T) {
	s := openStore(t)
	keep(t, s, "First")
	keep(t, s, "Second")
	third := keep(t, s, "Third")
	if _, err := s.MoveFirst(factsOf(third).number); err != nil {
		t.Fatalf("MoveFirst: %v", err)
	}
	wantOrder(t, s, "Third", "First", "Second")
}

func TestMoveLast_putsATaskBehindTheRest(t *testing.T) {
	s := openStore(t)
	first := keep(t, s, "First")
	keep(t, s, "Second")
	keep(t, s, "Third")
	if _, err := s.MoveLast(factsOf(first).number); err != nil {
		t.Fatalf("MoveLast: %v", err)
	}
	wantOrder(t, s, "Second", "Third", "First")
}

func TestMoveAfter_landsBehindTheTargetFromEitherSide(t *testing.T) {
	t.Run("coming from in front", func(t *testing.T) {
		s := openStore(t)
		first := keep(t, s, "First")
		second := keep(t, s, "Second")
		keep(t, s, "Third")
		if _, err := s.MoveAfter(factsOf(first).number, factsOf(second).number); err != nil {
			t.Fatalf("MoveAfter: %v", err)
		}
		wantOrder(t, s, "Second", "First", "Third")
	})
	t.Run("coming from behind", func(t *testing.T) {
		s := openStore(t)
		first := keep(t, s, "First")
		keep(t, s, "Second")
		third := keep(t, s, "Third")
		if _, err := s.MoveAfter(factsOf(third).number, factsOf(first).number); err != nil {
			t.Fatalf("MoveAfter: %v", err)
		}
		wantOrder(t, s, "First", "Third", "Second")
	})
}

func onlyErr(_ bool, err error) error { return err }

func TestMove_saysWhetherTheTaskWentAnywhere(t *testing.T) {
	s := openStore(t)
	first := keep(t, s, "First")
	second := keep(t, s, "Second")
	for _, tc := range []struct {
		what string
		move func() (bool, error)
		want bool
	}{
		{"the head to the head", func() (bool, error) { return s.MoveFirst(factsOf(first).number) }, false},
		{"the tail to the tail", func() (bool, error) { return s.MoveLast(factsOf(second).number) }, false},
		{"behind the one it is behind", func() (bool, error) {
			return s.MoveAfter(factsOf(second).number, factsOf(first).number)
		}, false},
		{"the tail to the head", func() (bool, error) { return s.MoveFirst(factsOf(second).number) }, true},
	} {
		moved, err := tc.move()
		if err != nil {
			t.Fatalf("moving %s: %v", tc.what, err)
		}
		if moved != tc.want {
			t.Errorf("moving %s says it moved = %v, want %v", tc.what, moved, tc.want)
		}
	}
}

func TestMove_toWhereItAlreadyIsChangesNothing(t *testing.T) {
	s := openStore(t)
	first := keep(t, s, "First")
	keep(t, s, "Second")
	if _, err := s.MoveFirst(factsOf(first).number); err != nil {
		t.Fatalf("MoveFirst: %v", err)
	}
	wantOrder(t, s, "First", "Second")
}

func TestMoveAfter_behindItselfLeavesTheOrderAlone(t *testing.T) {
	s := openStore(t)
	first := keep(t, s, "First")
	keep(t, s, "Second")
	keep(t, s, "Third")
	if _, err := s.MoveAfter(factsOf(first).number, factsOf(first).number); err != nil {
		t.Fatalf("MoveAfter behind itself: %v", err)
	}
	wantOrder(t, s, "First", "Second", "Third")
}

func TestMove_refusesATaskThatIsNotThere(t *testing.T) {
	s := openStore(t)
	keep(t, s, "The only one")
	for _, tc := range []struct {
		what string
		move func() error
	}{
		{"first", func() error { return onlyErr(s.MoveFirst(99)) }},
		{"last", func() error { return onlyErr(s.MoveLast(99)) }},
		{"after another", func() error { return onlyErr(s.MoveAfter(99, 1)) }},
		{"after one that is not there", func() error { return onlyErr(s.MoveAfter(1, 99)) }},
	} {
		if err := tc.move(); !errors.Is(err, ErrNoSuchTask) {
			t.Errorf("moving %s = %v, want %v", tc.what, err, ErrNoSuchTask)
		}
	}
}

func TestMove_countsASubTaskAmongItsSisters(t *testing.T) {
	s := openStore(t)
	parent := keep(t, s, "The parent")
	keepUnder(t, s, "First sister", factsOf(parent).number)
	second := keepUnder(t, s, "Second sister", factsOf(parent).number)
	other := keep(t, s, "Another root task")

	if _, err := s.MoveFirst(factsOf(second).number); err != nil {
		t.Fatalf("MoveFirst on a sub-task: %v", err)
	}
	under, err := s.Under(factsOf(parent).number)
	if err != nil {
		t.Fatalf("Under: %v", err)
	}
	if got := factsOf(under[0]).title; got != "Second sister" {
		t.Errorf("the first sister is %q, want %q", got, "Second sister")
	}

	if _, err := s.MoveFirst(factsOf(other).number); err != nil {
		t.Fatalf("MoveFirst on a root task: %v", err)
	}
	if got := factsOf(under[0]).title; got != "Second sister" {
		t.Errorf("moving a root task changed the sisters: first is %q", got)
	}
}

func TestUpdate_landsATaskBehindTheOnesInItsNewColumn(t *testing.T) {
	s := openStore(t)
	coming := keep(t, s, "Coming from the backlog")
	keepInColumn(t, s, "Already to do", domain.Todo)

	if err := s.Update(changed(coming, func(a *task.AsTold) {
		a.InColumn(domain.Todo)
	})); err != nil {
		t.Fatalf("Update: %v", err)
	}
	wantOrder(t, s, "Already to do", "Coming from the backlog")
}

func TestUpdate_leavesThePlaceAloneWhenTheColumnDoesNot(t *testing.T) {
	s := openStore(t)
	first := keep(t, s, "First")
	keep(t, s, "Second")

	if err := s.Update(changed(first, func(a *task.AsTold) {
		a.Titled("First, renamed")
	})); err != nil {
		t.Fatalf("Update: %v", err)
	}
	wantOrder(t, s, "First, renamed", "Second")
}

func TestUpdate_landsATaskAmongItsNewSisters(t *testing.T) {
	s := openStore(t)
	parent := keep(t, s, "The parent")
	keepUnder(t, s, "The first sister", factsOf(parent).number)
	joining := keep(t, s, "Joining them")

	if err := s.Update(changed(joining, func(a *task.AsTold) {
		a.Under(factsOf(parent).number)
	})); err != nil {
		t.Fatalf("Update: %v", err)
	}
	under, err := s.Under(factsOf(parent).number)
	if err != nil {
		t.Fatalf("Under: %v", err)
	}
	if len(under) != 2 || factsOf(under[1]).title != "Joining them" {
		t.Errorf("the sisters are %v, want the newcomer behind the first",
			titlesOf(under))
	}
}

func titlesOf(all []task.Task) []string {
	titles := make([]string, 0, len(all))
	for _, one := range all {
		titles = append(titles, factsOf(one).title)
	}
	return titles
}

func TestUpdate_changesTheLabelsOfATaskThatAlsoChangedColumn(t *testing.T) {
	s := openStore(t)
	kept, err := s.Keep(changed(mustBuild(t, "Moving and relabelled"), func(a *task.AsTold) {
		a.Labelled("spec")
		a.LabelsTold()
	}))
	if err != nil {
		t.Fatalf("Keep: %v", err)
	}
	keepInColumn(t, s, "Already to do", domain.Todo)

	if err := s.Update(changed(kept, func(a *task.AsTold) {
		a.InColumn(domain.Todo)
		a.NotLabelled()
		a.Labelled("team")
		a.LabelsTold()
	})); err != nil {
		t.Fatalf("Update: %v", err)
	}

	found, err := s.Find(factsOf(kept).number)
	if err != nil {
		t.Fatalf("Find: %v", err)
	}
	if got := factsOf(found).labels; len(got) != 1 || got[0] != "team" {
		t.Errorf("the labels are %v, want just %q", got, "team")
	}
	wantOrder(t, s, "Already to do", "Moving and relabelled")
}

func rankTheColumnsSay() string {
	var out strings.Builder
	out.WriteString("CASE state")
	for rank, state := range domain.States {
		_, _ = fmt.Fprintf(&out, " WHEN '%s' THEN %d", state, rank)
	}
	_, _ = fmt.Fprintf(&out, " ELSE %d END", len(domain.States))
	return out.String()
}

func TestTaskOrder_agreesWithTheColumnsCairnKnows(t *testing.T) {
	flat := strings.Join(strings.Fields(taskOrder), " ")
	if want := "ORDER BY " + rankTheColumnsSay() + ", position, number"; flat != want {
		t.Errorf("the order is\n  %s\nwant\n  %s", flat, want)
	}
}
