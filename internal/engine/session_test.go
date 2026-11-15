package engine

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"cairn/internal/engine/act"
	"cairn/internal/store/files"
)

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
		if a.Carried().Did == act.FocusRenewed {
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
	if !errors.Is(err, files.ErrFocus) {
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
