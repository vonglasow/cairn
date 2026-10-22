package engine

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"cairn/internal/config"
	"cairn/internal/trace"
)

func atTime(hour, minute int) time.Time {
	return time.Date(2026, 10, 21, hour, minute, 0, 0, time.Local)
}

func TestFocusLimits_sayWhenAFocusIsLetGo(t *testing.T) {
	end, err := readFocusLimits("", "20:00")
	if err != nil {
		t.Fatal(err)
	}
	if !end.passed(atTime(19, 59), atTime(20, 1)) {
		t.Error("a focus last used at 19:59 is kept at 20:01, after the day ended")
	}
	if end.passed(atTime(20, 1), atTime(20, 2)) {
		t.Error("a focus last used at 20:01 is let go at 20:02, before the next day end")
	}
	if end.passed(atTime(19, 58), atTime(19, 59)) {
		t.Error("a focus is let go before the day ended")
	}
	if !end.passed(atTime(20, 1), atTime(20, 1).AddDate(0, 0, 1)) {
		t.Error("a focus last used at 20:01 is kept past the next day end")
	}
	idle, err := readFocusLimits("2h", "")
	if err != nil {
		t.Fatal(err)
	}
	if !idle.passed(atTime(9, 0), atTime(11, 1)) || idle.passed(atTime(9, 0), atTime(10, 59)) {
		t.Error("a focus is not let go exactly when it has been unused for longer than the idle limit")
	}
	if none, _ := readFocusLimits("", ""); none.passed(atTime(9, 0), atTime(9, 0).AddDate(0, 0, 9)) {
		t.Error("a focus is let go when nothing says so")
	}
	for _, refused := range []struct{ idle, dayEnds, names string }{
		{"two hours", "", "focus.idle"}, {"-1h", "", "focus.idle"}, {"", "8pm", "focus.day_ends"}, {"", "25:00", "focus.day_ends"},
	} {
		if _, err := readFocusLimits(refused.idle, refused.dayEnds); !errors.Is(err, config.ErrFocus) || !strings.Contains(err.Error(), refused.names) {
			t.Errorf("%+v gives %v, want ErrFocus naming %s", refused, err, refused.names)
		}
	}
}

func TestAFocusIsLetGoWhereItWasLastUsedWhenABoardOpens(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cairn.db")
	e, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	e.acting = Acting{Through: "cli", By: "alexis"}
	add(t, e, "Ship it")
	if _, err := e.Ask(TaskFocus, given("number", "1")); err != nil {
		t.Fatal(err)
	}
	_ = e.Close()
	home := t.TempDir()
	dir := filepath.Join(home, ".config", "cairn")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.toml"), []byte("[focus]\nidle = \"2h\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	began := time.Now()
	opened := func(later time.Duration) *Engine {
		t.Helper()
		held, err := Hold(Home{Dir: home, Lookup: func(k string) (string, bool) {
			if k == NowVariable {
				return began.Add(later).Format(time.RFC3339), true
			}
			return path, k == "CAIRN_STORE"
		}}, Acting{Through: "tui", By: "alexis"})
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = held.Close() })
		return held
	}
	for _, later := range []time.Duration{time.Hour, 150 * time.Minute} {
		if held, _ := opened(later).focus(); len(held) != 1 {
			t.Fatalf("a focus used %v before is let go: %v", later, held)
		}
	}
	held := opened(5 * time.Hour)
	if now, _ := held.focus(); len(now) != 0 {
		t.Fatalf("a focus unused for 150 minutes is kept: %v", now)
	}
	acts, err := held.historyOf(1)
	if err != nil {
		t.Fatal(err)
	}
	var released time.Time
	for _, act := range acts {
		if c := act.Carried(); c.Did == trace.FocusReleased {
			released = c.At
			if c.Through != "focus" || c.By != "alexis" {
				t.Errorf("the focus was let go by %q through %q", c.By, c.Through)
			}
		}
	}
	if want := began.Add(150 * time.Minute); released.Sub(want).Abs() > 2*time.Second {
		t.Errorf("the focus was let go at %v, want where it was last used, %v", released, want)
	}
}

func TestAHeldEngineLetsGoOfAStaleFocusBeforeItWrites(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cairn.db")
	e, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	add(t, e, "Ship it")
	if _, err := e.Ask(TaskFocus, given("number", "1")); err != nil {
		t.Fatal(err)
	}
	_ = e.Close()
	home := t.TempDir()
	dir := filepath.Join(home, ".config", "cairn")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.toml"), []byte("[focus]\nidle = \"2h\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	later := time.Duration(0)
	held, err := Hold(Home{Dir: home, Lookup: func(k string) (string, bool) {
		if k == NowVariable {
			return time.Now().Add(later).Format(time.RFC3339), true
		}
		return path, k == "CAIRN_STORE"
	}}, Acting{Through: "tui", By: "alexis"})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = held.Close() }()
	later = 5 * time.Hour
	if _, err := held.Ask(TaskList, Input{}); err != nil {
		t.Fatal(err)
	}
	if now, _ := held.focus(); len(now) != 1 {
		t.Fatalf("a read lets go of a focus: %v", now)
	}
	if _, err := held.Ask(TaskNew, given("title", "Another")); err != nil {
		t.Fatal(err)
	}
	if now, _ := held.focus(); len(now) != 0 {
		t.Errorf("a write on a board left open leaves %v in focus", now)
	}
}

func TestConfigShowSaysWhenAFocusIsLetGoOnItsOwn(t *testing.T) {
	for _, said := range []struct {
		focus *FocusLimits
		want  string
	}{
		{&FocusLimits{Idle: "2h", DayEnds: "20:00"}, "focus       let go after 2h unused, and at 20:00\n"},
		{&FocusLimits{Idle: "2h"}, "focus       let go after 2h unused\n"},
		{&FocusLimits{DayEnds: "20:00"}, "focus       let go at 20:00\n"},
		{nil, "nothing is let go on its own\n"},
	} {
		if got := configText(Config{Focus: said.focus}); !strings.Contains(got, said.want) {
			t.Errorf("config show with %+v says:\n%s", said.focus, got)
		}
	}
}
