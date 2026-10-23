package engine

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"cairn/internal/config"
	"cairn/internal/trace"
)

func actAt(number int, did string, at time.Time) trace.Act {
	return trace.Made(trace.Carried{Task: number, Did: did, At: at})
}

func TestSession_isUpSinceTheFocusWasLastTakenOrRenewed(t *testing.T) {
	taken := time.Date(2026, 10, 23, 9, 0, 0, 0, time.UTC)
	acts := []trace.Act{
		actAt(4, trace.FocusTaken, taken),
		actAt(4, trace.FocusRenewed, taken.Add(time.Hour)),
		actAt(4, trace.Changed, taken.Add(2*time.Hour)),
	}
	got := sessionOf(4, acts, taken.Add(150*time.Minute), 0)
	if !got.Since.Equal(taken.Add(time.Hour)) || got.Up != "1h30m" || got.Ask != "" {
		t.Errorf("session = %+v, want up 1h30m since the renewal, asking nothing", got)
	}
}

func TestSession_asksOnlyOnceItsLengthIsReached(t *testing.T) {
	taken := time.Date(2026, 10, 23, 9, 0, 0, 0, time.UTC)
	acts := []trace.Act{actAt(4, trace.FocusTaken, taken)}
	if got := sessionOf(4, acts, taken.Add(59*time.Minute), time.Hour); got.Ask != "" {
		t.Errorf("within its length it asks %q", got.Ask)
	}
	want := "this focus session has been up 2h: renew it by focusing task 4 again, or end it by unfocusing task 4"
	if got := sessionOf(4, acts, taken.Add(2*time.Hour), time.Hour); got.Ask != want {
		t.Errorf("past its length it asks %q, want %q", got.Ask, want)
	}
}

func homeLasting(t *testing.T, path, length string, later time.Duration) Home {
	t.Helper()
	home := t.TempDir()
	dir := filepath.Join(home, ".config", "cairn")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.toml"), []byte("[focus]\nsession = \""+length+"\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	now := time.Now().Add(later).Format(time.RFC3339)
	return Home{Dir: home, Lookup: func(k string) (string, bool) {
		switch k {
		case "CAIRN_STORE":
			return path, true
		case NowVariable:
			return now, later != 0
		}
		return "", false
	}}
}

func renewals(t *testing.T, path string) int {
	t.Helper()
	e, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = e.Close() }()
	acts, err := e.history.ActsOn(1)
	if err != nil {
		t.Fatal(err)
	}
	n := 0
	for _, a := range acts {
		if a.Carried().Did == trace.FocusRenewed {
			n++
		}
	}
	return n
}

func TestFocus_againRenewsOnlyASessionPastItsLength(t *testing.T) {
	path := storeAt(t, "one")
	if _, err := DoFrom(TaskFocus, homeLasting(t, path, "1h", 0), given("number", "1"), Acting{}); err != nil {
		t.Fatal(err)
	}
	if _, err := DoFrom(TaskFocus, homeLasting(t, path, "1h", 30*time.Minute), given("number", "1"), Acting{}); err != nil {
		t.Fatal(err)
	}
	if n := renewals(t, path); n != 0 {
		t.Fatalf("focusing again within the session renewed it %d time(s)", n)
	}
	if _, err := DoFrom(TaskFocus, homeLasting(t, path, "1h", 2*time.Hour), given("number", "1"), Acting{}); err != nil {
		t.Fatal(err)
	}
	if n := renewals(t, path); n != 1 {
		t.Errorf("focusing again past the session renewed it %d time(s), want 1", n)
	}
}

func TestFocus_refusesASessionLengthThatIsNoDuration(t *testing.T) {
	path := storeAt(t, "one")
	_, err := DoFrom(TaskFocus, homeLasting(t, path, "an hour", 0), given("number", "1"), Acting{})
	if !errors.Is(err, config.ErrFocus) {
		t.Errorf("err = %v, want config.ErrFocus", err)
	}
}

func TestAnswer_asksToRenewWhateverWasCalled(t *testing.T) {
	path := storeAt(t, "one")
	if _, err := DoFrom(TaskFocus, homeLasting(t, path, "1h", 0), given("number", "1"), Acting{}); err != nil {
		t.Fatal(err)
	}
	answer, err := DoFrom(TaskList, homeLasting(t, path, "1h", 2*time.Hour), given(), Acting{})
	if err != nil {
		t.Fatal(err)
	}
	if len(answer.Asked) != 1 {
		t.Errorf("a listing past the session's length asked %v, want one question", answer.Asked)
	}
}
