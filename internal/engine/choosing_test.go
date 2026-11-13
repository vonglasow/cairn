package engine

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"cairn/internal/trace"
)

func futureDay(days int) string { return time.Now().AddDate(0, 0, days).Format(time.DateOnly) }

func choosingStore(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "cairn.db")
	for _, title := range []string{"Write the spec", "Ship it", "Announce it", "Thank the reviewers"} {
		did(t, path, TaskNew, ask("title", title))
	}
	return path
}

func rowOn(t *testing.T, rows []map[string]any, n int) map[string]any {
	t.Helper()
	for _, row := range rows {
		if row["number"] == float64(n) {
			return row
		}
	}
	t.Fatalf("no row of task %d in %v", n, rows)
	return nil
}

func listedOn(t *testing.T, path, day string, a request) []map[string]any {
	t.Helper()
	return rowsAsJSON(t, TaskList, doneOn(t, path, day, TaskList, a))
}

func TestTaskList_countsTheOpenTasksATaskIsInTheWayOf(t *testing.T) {
	path := choosingStore(t)
	did(t, path, TaskEdit, ask("number", "2").with("blocker", "1"))
	did(t, path, TaskEdit, ask("number", "3").with("blocker", "1", "2"))
	did(t, path, TaskEdit, ask("number", "4", "state", "done").with("blocker", "1"))
	rows := listedOn(t, path, "2026-11-13", ask())
	for n, want := range map[int]float64{1: 2, 2: 1, 3: 0, 4: 0} {
		if got := rowOn(t, rows, n)["blocks"]; got != want {
			t.Errorf("task %d blocks %v, want %v", n, got, want)
		}
	}
}

func TestTaskList_countsWhatATaskBlocksBeyondThePageOrTheNarrowing(t *testing.T) {
	path := choosingStore(t)
	did(t, path, TaskEdit, ask("number", "2", "state", "current").with("blocker", "1"))
	rows := listedOn(t, path, "2026-11-13", ask("state", "backlog"))
	if got := rowOn(t, rows, 1)["blocks"]; got != float64(1) {
		t.Errorf("task 1 blocks %v, want 1 though task 2 is not listed", got)
	}
}

func TestTaskList_aTaskArchivedIsInNobodysWay(t *testing.T) {
	path := choosingStore(t)
	did(t, path, TaskEdit, ask("number", "2").with("blocker", "1"))
	did(t, path, TaskEdit, ask("number", "2", "state", "done"))
	did(t, path, TaskArchive, ask("number", "2"))
	did(t, path, TaskEdit, ask("number", "3").with("blocker", "1"))
	rows := listedOn(t, path, "2026-11-13", ask())
	if got := rowOn(t, rows, 1)["blocks"]; got != float64(1) {
		t.Errorf("task 1 blocks %v, want 1", got)
	}
}

func TestTaskList_saysHowFarEachTaskIsFromItsDeadline(t *testing.T) {
	path := choosingStore(t)
	did(t, path, TaskEdit, ask("number", "1", "deadline", "2026-11-10"))
	did(t, path, TaskEdit, ask("number", "2", "deadline", "2026-11-13"))
	did(t, path, TaskEdit, ask("number", "3", "deadline", "2026-11-14"))
	rows := listedOn(t, path, "2026-11-13", ask())
	for n, want := range map[int][2]any{1: {float64(3), nil}, 2: {nil, float64(0)}, 3: {nil, float64(1)}, 4: {nil, nil}} {
		row := rowOn(t, rows, n)
		if row["overdue"] != want[0] || row["due_in"] != want[1] {
			t.Errorf("task %d is overdue %v and due in %v, want %v and %v", n, row["overdue"], row["due_in"], want[0], want[1])
		}
		if _, carried := row["overdue"]; !carried {
			t.Errorf("task %d carries no overdue at all", n)
		}
	}
}

func TestTaskList_aTaskDoneIsNeitherOverdueNorDue(t *testing.T) {
	path := choosingStore(t)
	did(t, path, TaskEdit, ask("number", "1", "deadline", "2026-11-10", "state", "done"))
	row := rowOn(t, listedOn(t, path, "2026-11-13", ask()), 1)
	if row["overdue"] != nil || row["due_in"] != nil {
		t.Errorf("a task done is overdue %v and due in %v, want neither", row["overdue"], row["due_in"])
	}
}

func TestTaskList_saysHowManyDaysEachTaskSatInItsColumn(t *testing.T) {
	path := choosingStore(t)
	rows := listedOn(t, path, futureDay(3), ask())
	if got := rowOn(t, rows, 1)["age"]; got != float64(3) {
		t.Errorf("task 1 is %v days in its column, want 3", got)
	}
}

func TestTaskList_aTaskMovedThisDayIsNoDaysInItsColumn(t *testing.T) {
	path := choosingStore(t)
	did(t, path, TaskEdit, ask("number", "1", "state", "todo"))
	rows := listedOn(t, path, futureDay(0), ask())
	if got := rowOn(t, rows, 1)["age"]; got != float64(0) {
		t.Errorf("task 1 is %v days in its column, want 0", got)
	}
}

func TestTaskList_saysWhenAnObjectiveATaskServesIsLate(t *testing.T) {
	path := choosingStore(t)
	did(t, path, ObjectiveNew, ask("title", "Launch", "deadline", "2026-11-10"))
	did(t, path, ObjectiveNew, ask("title", "Polish", "deadline", "2026-11-13"))
	did(t, path, ObjectiveNew, ask("title", "Someday"))
	did(t, path, ObjectiveNew, ask("title", "Shipped", "deadline", "2026-11-01"))
	did(t, path, ObjectiveEdit, ask("number", "4", "state", "reached"))
	did(t, path, ObjectiveNew, ask("title", "Gave up", "deadline", "2026-11-01"))
	did(t, path, ObjectiveEdit, ask("number", "5", "state", "dropped"))
	did(t, path, TaskEdit, ask("number", "1").with("objective", "2", "1"))
	did(t, path, TaskEdit, ask("number", "2").with("objective", "2", "3"))
	did(t, path, TaskEdit, ask("number", "3").with("objective", "4", "5"))
	rows := listedOn(t, path, "2026-11-13", ask())
	for n, want := range map[int]bool{1: true, 2: false, 3: false, 4: false} {
		if got := rowOn(t, rows, n)["objective_late"]; got != want {
			t.Errorf("task %d objective_late %v, want %v", n, got, want)
		}
	}
}

func TestTaskList_keepsTheTasksOneIsInTheWayOf(t *testing.T) {
	path := choosingStore(t)
	did(t, path, TaskEdit, ask("number", "2").with("blocker", "1"))
	did(t, path, TaskEdit, ask("number", "4").with("blocker", "3", "1"))
	did(t, path, TaskEdit, ask("number", "3").with("blocker", "2"))
	rows := listedOn(t, path, "2026-11-13", ask("blocker", "1"))
	if len(rows) != 2 || rows[0]["number"] != float64(2) || rows[1]["number"] != float64(4) {
		t.Errorf("listed %v, want tasks 2 and 4", rows)
	}
}

func TestTaskList_aBlockerThatIsNoTaskIsRefused(t *testing.T) {
	path := choosingStore(t)
	for _, said := range []string{"x", "0", "-2"} {
		if _, err := Do(TaskList, path, ask("blocker", said).input()); err == nil || !strings.Contains(err.Error(), said) {
			t.Errorf("--blocker %s: %v, want it refused naming it", said, err)
		}
	}
}

func TestTaskList_aBlockerKeepsTheListingNarrowed(t *testing.T) {
	path := choosingStore(t)
	answer := doneOn(t, path, "2026-11-13", TaskList, ask("blocker", "1"))
	if answer.Matched == nil || answer.Tasks != nil {
		t.Errorf("a listing by blocker answered %v as the whole board", answer)
	}
}

func TestTaskShow_namesTheTasksItIsInTheWayOf(t *testing.T) {
	path := choosingStore(t)
	did(t, path, TaskEdit, ask("number", "3").with("blocker", "1"))
	did(t, path, TaskEdit, ask("number", "2").with("blocker", "1"))
	answer := did(t, path, TaskShow, ask("number", "1"))
	if said := RenderText(TaskShow, answer); !strings.Contains(said, "blocks       ⊳2  #2, #3\n") {
		t.Errorf("task show said:\n%s\nwant it to name tasks 2 and 3", said)
	}
	if said := RenderText(TaskShow, did(t, path, TaskShow, ask("number", "2"))); strings.Contains(said, "⊳") {
		t.Errorf("task show said:\n%s\nwant nothing of what it blocks", said)
	}
}

func TestTaskList_marksHowManyTasksEachIsInTheWayOf(t *testing.T) {
	path := choosingStore(t)
	did(t, path, TaskEdit, ask("number", "2").with("blocker", "1"))
	said := RenderText(TaskList, did(t, path, TaskList, ask()))
	if !strings.Contains(said, "#1  Write the spec  · backlog  · normal  · ⊳1\n") || strings.Count(said, "⊳") != 1 {
		t.Errorf("task list said:\n%s\nwant task 1 alone marked ⊳1", said)
	}
}

func TestTaskSearch_givesTheFactsThatHelpChooseToo(t *testing.T) {
	path := choosingStore(t)
	did(t, path, TaskEdit, ask("number", "2").with("blocker", "1"))
	rows := rowsAsJSON(t, TaskSearch, doneOn(t, path, futureDay(2), TaskSearch, ask().with("named", "spec")))
	if row := rowOn(t, rows, 1); row["blocks"] != float64(1) || row["age"] != float64(2) {
		t.Errorf("the search row of task 1 blocks %v and is %v days old, want 1 and 2", row["blocks"], row["age"])
	}
}

func TestTaskList_aChangeOtherThanItsColumnLeavesATasksAgeAlone(t *testing.T) {
	e, path := engineOn(t)
	add(t, e, "Write the spec")
	later := time.Now().AddDate(0, 0, 2)
	if err := e.history.Record(trace.Made(trace.Carried{Task: 1, Did: trace.Changed, Field: "title", Was: "Write the spec", Became: "Write the whole spec", At: later})); err != nil {
		t.Fatal(err)
	}
	_ = e.Close()
	rows := listedOn(t, path, futureDay(3), ask())
	if got := rowOn(t, rows, 1)["age"]; got != float64(3) {
		t.Errorf("task 1 is %v days in its column, want 3: only its title changed since", got)
	}
}
