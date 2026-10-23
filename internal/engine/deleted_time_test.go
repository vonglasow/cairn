package engine

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"cairn/internal/trace"
)

func homeSaying(t *testing.T, path, toml string) Home {
	t.Helper()
	home := t.TempDir()
	dir := filepath.Join(home, ".config", "cairn")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.toml"), []byte(toml), 0o600); err != nil {
		t.Fatal(err)
	}
	return Home{Dir: home, Lookup: func(k string) (string, bool) { return path, k == "CAIRN_STORE" }}
}

func TestDelete_takesItsTimeOnlyWhenTheConfigurationSaysSo(t *testing.T) {
	for _, c := range []struct {
		toml  string
		taken bool
	}{{"", false}, {"[time]\ndeleted = \"taken\"\n", true}} {
		path := storeAt(t, "one")
		if _, err := DoFrom(TaskDelete, homeSaying(t, path, c.toml), Agreed(given("number", "1")), Acting{}); err != nil {
			t.Fatal(err)
		}
		e, err := Open(path)
		if err != nil {
			t.Fatal(err)
		}
		acts, err := e.history.ActsOn(1)
		_ = e.Close()
		if err != nil {
			t.Fatal(err)
		}
		went := acts[len(acts)-1].Carried().Was
		if strings.HasSuffix(went, trace.WentWithItsTime) != c.taken {
			t.Errorf("with %q, the deletion took %q", c.toml, went)
		}
	}
}

func TestTimesheet_saysATaskDeletedWithItsTimeKeptIsGone(t *testing.T) {
	path := storeAt(t, "Write the spec")
	if _, err := Do(TimesheetAdd, path, given("number", "1", "time", "1h")); err != nil {
		t.Fatal(err)
	}
	if _, err := Do(TaskDelete, path, Agreed(given("number", "1"))); err != nil {
		t.Fatal(err)
	}
	answer, err := Do(TimesheetShow, path, given())
	if err != nil {
		t.Fatal(err)
	}
	sheet, _ := Of(TimesheetShow, answer).(Timesheet)
	for _, day := range sheet.Days {
		for _, one := range day.Tasks {
			if one.Number == 1 && one.Title == "Write the spec" && one.Deleted {
				return
			}
		}
	}
	t.Errorf("the sheet does not say task 1 was deleted: %+v", sheet)
}
