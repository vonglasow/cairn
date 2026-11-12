package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"cairn/internal/trace"
)

func midnight(at time.Time) time.Time {
	return time.Date(at.Year(), at.Month(), at.Day(), 0, 0, 0, 0, time.Local)
}

func archivedOnItsOwn(t *testing.T, e *Engine, number int) bool {
	t.Helper()
	acts, err := e.historyOf(number)
	if err != nil {
		t.Fatal(err)
	}
	for _, act := range acts {
		if c := act.Carried(); c.Field == "archived" && c.Became == "yes" {
			if c.Through != throughArchive || c.By != "sam" {
				t.Errorf("task %d was archived by %q through %q", number, c.By, c.Through)
			}
			return true
		}
	}
	return false
}

func TestWhatWentDoneBeforeThePeriodUnderWayIsArchived(t *testing.T) {
	e := openEngine(t)
	e.acting = Acting{Through: "cli", By: "sam"}
	add(t, e, "Migration Jenkins")
	add(t, e, "Spike Argo")
	if _, err := e.Ask(TaskEdit, given("number", "1", "state", "done")); err != nil {
		t.Fatal(err)
	}
	today := midnight(time.Now())
	if err := e.sweptByPeriod(today, 14, today.AddDate(0, 0, 5)); err != nil {
		t.Fatal(err)
	}
	if archivedOnItsOwn(t, e, 1) {
		t.Fatal("a task that went done in the period under way was archived")
	}
	if err := e.sweptByPeriod(today, 14, today.AddDate(0, 0, 20)); err != nil {
		t.Fatal(err)
	}
	if !archivedOnItsOwn(t, e, 1) {
		t.Error("a task that went done before the period under way is not archived")
	}
	if archivedOnItsOwn(t, e, 2) {
		t.Error("a task never done was archived")
	}
	if err := e.sweptByPeriod(today, 14, today.AddDate(0, 0, 40)); err != nil {
		t.Errorf("a task already archived stops the sweep: %v", err)
	}
	if e.acting.Through != "cli" {
		t.Errorf("the sweep left the engine acting through %q", e.acting.Through)
	}
}

func TestABoardOpeningArchivesOnItsOwn(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cairn.db")
	e, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	e.acting = Acting{Through: "cli", By: "sam"}
	add(t, e, "Migration Jenkins")
	if _, err := e.Ask(TaskEdit, given("number", "1", "state", "done")); err != nil {
		t.Fatal(err)
	}
	ago := time.Now().AddDate(0, 0, -10)
	if err := e.history.Record(trace.Made(trace.Carried{Task: 1, Did: trace.Changed, Field: "column", Was: "todo", Became: "done", At: ago})); err != nil {
		t.Fatal(err)
	}
	_ = e.Close()
	home := t.TempDir()
	dir := filepath.Join(home, ".config", "cairn")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	start := time.Now().AddDate(0, 0, -20).Format(time.DateOnly)
	if err := os.WriteFile(filepath.Join(dir, "config.toml"), fmt.Appendf(nil, "[archive]\nstart = %s\n", start), 0o600); err != nil {
		t.Fatal(err)
	}
	held, err := Hold(Home{Dir: home, Lookup: func(k string) (string, bool) { return path, k == "CAIRN_STORE" }}, Acting{Through: "tui", By: "sam"})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = held.Close() }()
	if !archivedOnItsOwn(t, held, 1) {
		t.Error("opening a board left a task done before the period under way on it")
	}
}

func TestConfigShowSaysWhatIsArchivedOnItsOwn(t *testing.T) {
	said := configText(Config{Archive: &Archive{Start: "2026-10-21", Days: 10}})
	if !strings.Contains(said, "[archive]\nstart = 2026-10-21\ndays = 10\n") {
		t.Errorf("config show says:\n%s", said)
	}
	if said := configText(Config{Archive: &Archive{Days: 14}}); !strings.Contains(said, "nothing is archived on its own") {
		t.Errorf("with no archiving, config show says:\n%s", said)
	}
}

func TestTodayIsTheDayTheEnvironmentNames(t *testing.T) {
	named := func(day string) Home {
		return Home{Lookup: func(k string) (string, bool) { return day, k == TodayVariable && day != "" }}
	}
	if got := today(named("2026-11-04")); got.Format(time.DateOnly) != "2026-11-04" || got.Location() != time.Local {
		t.Errorf("CAIRN_TODAY=2026-11-04 gives %v", got)
	}
	if got := today(named("")); midnight(got) != midnight(time.Now()) {
		t.Errorf("with no day named, today is %v", got)
	}
	if got := today(named("soon")); midnight(got) != midnight(time.Now()) {
		t.Errorf("with a day that is not one, today is %v", got)
	}
}

func TestHold_holdsAStoreBehind(t *testing.T) {
	path := behindStore(t)
	home := t.TempDir()
	dir := filepath.Join(home, ".config", "cairn")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	start := time.Now().AddDate(0, 0, -20).Format(time.DateOnly)
	if err := os.WriteFile(filepath.Join(dir, "config.toml"), fmt.Appendf(nil, "[archive]\nstart = %s\n", start), 0o600); err != nil {
		t.Fatal(err)
	}
	held, err := Hold(Home{Dir: home, Lookup: func(k string) (string, bool) { return path, k == "CAIRN_STORE" }}, Acting{Through: "tui", By: "sam"})
	if err != nil {
		t.Fatalf("Hold: %v", err)
	}
	defer func() { _ = held.Close() }()
	waiting, err := held.store.Waiting()
	if err != nil {
		t.Fatal(err)
	}
	if len(waiting) != 1 {
		t.Errorf("%d changes waiting after holding, want the 1 that was", len(waiting))
	}
}

func TestHold_sweepsTheArchiveOnceTheStoreIsCurrent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cairn.db")
	e, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	e.acting = Acting{Through: "cli", By: "sam"}
	add(t, e, "Migration Jenkins")
	if _, err := e.Ask(TaskEdit, given("number", "1", "state", "done")); err != nil {
		t.Fatal(err)
	}
	ago := time.Now().AddDate(0, 0, -10)
	if err := e.history.Record(trace.Made(trace.Carried{Task: 1, Did: trace.Changed, Field: "column", Was: "todo", Became: "done", At: ago})); err != nil {
		t.Fatal(err)
	}
	if _, err := e.rollBack(path); err != nil {
		t.Fatal(err)
	}
	_ = e.Close()
	home := t.TempDir()
	dir := filepath.Join(home, ".config", "cairn")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	start := time.Now().AddDate(0, 0, -20).Format(time.DateOnly)
	if err := os.WriteFile(filepath.Join(dir, "config.toml"), fmt.Appendf(nil, "[archive]\nstart = %s\n", start), 0o600); err != nil {
		t.Fatal(err)
	}
	held, err := Hold(Home{Dir: home, Lookup: func(k string) (string, bool) { return path, k == "CAIRN_STORE" }}, Acting{Through: "tui", By: "sam"})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = held.Close() }()
	if _, err := held.Ask(MigrationApply, Input{}); err != nil {
		t.Fatalf("apply: %v", err)
	}
	if _, err := held.Ask(TaskList, Input{}); err != nil {
		t.Fatalf("list: %v", err)
	}
	if !archivedOnItsOwn(t, held, 1) {
		t.Error("a held engine whose store became current left a task done before the period under way on the board")
	}
}

func TestArchive_aSweepCountsWhatItArchived(t *testing.T) {
	path := storeAt(t, "Write the spec", "Ship it", "Unrelated")
	for _, n := range []string{"1", "2"} {
		if _, err := Do(TaskEdit, path, given("number", n, "state", "done")); err != nil {
			t.Fatal(err)
		}
	}
	answer, err := Do(TaskArchive, path, confirmed(given("state", "done")))
	if err != nil {
		t.Fatal(err)
	}
	if answer.Archived == nil || *answer.Archived != 2 {
		t.Errorf("Archived = %v, want 2", answer.Archived)
	}
}

func archivingHome(t *testing.T) (string, Home) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "cairn.db")
	e, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	add(t, e, "Migration Jenkins")
	if _, err := e.Ask(TaskEdit, given("number", "1", "state", "done")); err != nil {
		t.Fatal(err)
	}
	ago := time.Now().AddDate(0, 0, -10)
	if err := e.history.Record(trace.Made(trace.Carried{Task: 1, Did: trace.Changed, Field: "column", Was: "todo", Became: "done", At: ago})); err != nil {
		t.Fatal(err)
	}
	_ = e.Close()
	home := t.TempDir()
	dir := filepath.Join(home, ".config", "cairn")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	start := time.Now().AddDate(0, 0, -20).Format(time.DateOnly)
	if err := os.WriteFile(filepath.Join(dir, "config.toml"), fmt.Appendf(nil, "[archive]\nstart = %s\n", start), 0o600); err != nil {
		t.Fatal(err)
	}
	return path, Home{Dir: home, Lookup: func(k string) (string, bool) { return path, k == "CAIRN_STORE" }}
}

func TestArchiving_onlyTheListingOfTasksSweeps(t *testing.T) {
	path, home := archivingHome(t)
	at := Places{Store: path, Home: home}
	for op, in := range map[*Operation]Input{&TaskShow: given("number", "1"), &ObjectiveList: given()} {
		if _, err := DoAt(*op, at, in, Acting{Through: "cli", By: "sam"}); err != nil {
			t.Fatal(err)
		}
		shown, err := Do(TaskShow, path, given("number", "1"))
		if err != nil {
			t.Fatal(err)
		}
		if factsOf(*shown.Task).Archived {
			t.Fatalf("%s archived on its own", op.Line())
		}
	}
	if _, err := DoAt(TaskList, at, given(), Acting{Through: "cli", By: "sam"}); err != nil {
		t.Fatal(err)
	}
	shown, err := Do(TaskShow, path, given("number", "1"))
	if err != nil {
		t.Fatal(err)
	}
	if !factsOf(*shown.Task).Archived {
		t.Error("task list left a task done before the period under way on the board")
	}
}

func TestTheSweepArchivesEveryTaskDoneBeforeThePeriodPastOneDoneInIt(t *testing.T) {
	e := openEngine(t)
	e.acting = Acting{Through: "cli", By: "sam"}
	for _, title := range []string{"Spike Argo", "Migration Jenkins", "Ship it"} {
		add(t, e, title)
	}
	if _, err := e.Ask(TaskEdit, given("number", "1", "state", "done")); err != nil {
		t.Fatal(err)
	}
	doneLongAgo(t, e, 2)
	doneLongAgo(t, e, 3)
	today := midnight(time.Now())
	if err := e.sweptByPeriod(today.AddDate(0, 0, -20), 14, today); err != nil {
		t.Fatal(err)
	}
	if archivedOnItsOwn(t, e, 1) {
		t.Error("a task done in the period under way was archived")
	}
	for _, n := range []int{2, 3} {
		if !archivedOnItsOwn(t, e, n) {
			t.Errorf("task %d, done before the period under way, is not archived", n)
		}
	}
}
