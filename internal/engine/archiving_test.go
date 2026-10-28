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
