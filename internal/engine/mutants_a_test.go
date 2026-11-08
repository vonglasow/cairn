package engine

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"

	"cairn/internal/debug"
	"cairn/internal/domain"
	"cairn/internal/task"
	"cairn/internal/upgrade"
)

func TestSummary_saysEachFieldSetOnceAndNothingElse(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cairn.db")
	made, err := DoAt(TaskNew, Places{Store: path}, Input{Given: map[string]string{"title": "Write the spec"}}, Acting{})
	if err != nil {
		t.Fatal(err)
	}
	total, written := 3, "/tmp/x"
	for _, c := range []struct {
		answer Answer
		want   string
	}{
		{Answer{}, ""},
		{Answer{Task: made.Task, Tasks: []task.Task{*made.Task}}, "task #1 backlog, tasks 1"},
		{Answer{Tasks: []task.Task{}}, ""},
		{Answer{Labels: &[]LabelCarried{{}, {}}}, "labels 2"},
		{Answer{Total: &total, Written: &written}, "total 3, written"},
		{Answer{Context: "here"}, "context"},
	} {
		if got := summary(c.answer); got != c.want {
			t.Errorf("summary of %+v = %q, want %q", c.answer, got, c.want)
		}
	}
}

func TestCameTo_namesEveryErrorAFailureWasWrappedIn(t *testing.T) {
	l := debugLog(t)
	got := cameTo(l, Answer{}, fmt.Errorf("wrapped: %w", errors.New("inner")))
	if want := "failed *fmt.wrapError→*errors.errorString " + l.Text("wrapped: inner"); got != want {
		t.Errorf("a failure came to %q, want %q", got, want)
	}
}

func TestGivenIn_writesAListedArgumentAsItsValues(t *testing.T) {
	l := debugLog(t)
	op := Operation{Takes: []Argument{{Name: "n", Kind: Number}}}
	if got := givenIn(l, op, Input{Listed: map[string][]string{"n": {"1", "2"}}}); got != " n=[1 2]" {
		t.Errorf("a listed argument reads %q", got)
	}
}

func TestShown_masksOnlyWhatAPersonWrote(t *testing.T) {
	l := debugLog(t)
	for _, c := range []struct {
		a     Argument
		value string
		plain bool
	}{
		{Argument{Kind: Number}, "abc", true},
		{Argument{Kind: Day}, "abc", true},
		{Argument{Kind: Duration}, "abc", true},
		{Argument{Kind: Bool}, "abc", true},
		{Argument{Kind: Text, OneOf: []string{"abc"}}, "abc", true},
		{Argument{Kind: Text}, "true", true},
		{Argument{Kind: Text}, "false", true},
		{Argument{Kind: Text}, "12", true},
		{Argument{Kind: Text}, "abc", false},
		{Argument{Kind: Text, OneOf: []string{"xyz"}}, "abc", false},
	} {
		want := l.Text(c.value)
		if c.plain {
			want = c.value
		}
		if got := shown(l, c.a, c.value); got != want {
			t.Errorf("%q given as %+v reads %q, want %q", c.value, c.a, got, want)
		}
	}
}

func TestConfigLines_saysEachLineOrThatItCannotBeRead(t *testing.T) {
	lines := configLines(Home{Dir: t.TempDir()})
	if len(lines) == 0 || slices.Contains(lines, "") || slices.Contains(lines, "cannot be read") {
		t.Errorf("a configuration reads %q", lines)
	}
	broken := t.TempDir()
	if err := os.MkdirAll(filepath.Join(broken, ".config", "cairn", "config.toml"), 0o700); err != nil {
		t.Fatal(err)
	}
	if got := configLines(Home{Dir: broken}); !slices.Equal(got, []string{"cannot be read"}) {
		t.Errorf("no configuration reads %q", got)
	}
}

func headerFacts(at Places) map[string]string {
	facts := map[string]string{}
	for _, f := range storeInNumbers(at) {
		facts[f[0]] = f[1]
	}
	return facts
}

func TestStoreInNumbers_countsTheArchivedAndTheChangesApplied(t *testing.T) {
	at := Places{Store: filepath.Join(t.TempDir(), "cairn.db")}
	for _, given := range []map[string]string{{"title": "Ship it", "state": "done"}, {"number": "1"}} {
		op := TaskNew
		if given["number"] != "" {
			op = TaskArchive
		}
		if _, err := DoAt(op, at, Input{Given: given}, Acting{}); err != nil {
			t.Fatal(err)
		}
	}
	facts := headerFacts(at)
	if facts["archived"] != "1" {
		t.Errorf("archived reads %q", facts["archived"])
	}
	m := regexp.MustCompile(`^(\d+) of (\d+) changes applied$`).FindStringSubmatch(facts["schema"])
	if m == nil || m[1] != m[2] || m[1] == "0" {
		t.Errorf("the schema reads %q", facts["schema"])
	}
}

func TestStoreInNumbers_saysAStoreThatCannotBeOpened(t *testing.T) {
	at := Places{Store: filepath.Join(t.TempDir(), "cairn.db")}
	if err := os.WriteFile(at.Store, []byte("not a store at all, not even close to one"), 0o600); err != nil {
		t.Fatal(err)
	}
	if got := headerFacts(at)["store"]; !strings.HasPrefix(got, "cannot be read: ") {
		t.Errorf("a store that cannot be opened reads %q", got)
	}
}

func TestDebugReport_writesInTheCurrentFolderWhenNoneIsSaid(t *testing.T) {
	here := t.TempDir()
	t.Chdir(here)
	answer, err := DoAt(DebugReport, Places{Store: filepath.Join(t.TempDir(), "cairn.db"), Home: Home{Dir: t.TempDir()}}, Input{}, Acting{})
	if err != nil {
		t.Fatal(err)
	}
	if answer.Written == nil {
		t.Fatalf("the report answered %+v", answer)
	}
	if got, _ := filepath.EvalSymlinks(filepath.Dir(*answer.Written)); got != mustReal(t, here) {
		t.Errorf("the archive is written in %s, want %s", got, here)
	}
}

func mustReal(t *testing.T, path string) string {
	t.Helper()
	real, err := filepath.EvalSymlinks(path)
	if err != nil {
		t.Fatal(err)
	}
	return real
}

func TestDebugReport_saysADebugFolderThatCannotBeRead(t *testing.T) {
	store := filepath.Join(t.TempDir(), "cairn.db")
	if err := os.WriteFile(Places{Store: store}.Debug(), []byte("a file"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := DoAt(DebugReport, Places{Store: store, Home: Home{Dir: t.TempDir()}}, Input{Given: map[string]string{"to": t.TempDir()}}, Acting{}); err == nil {
		t.Error("a debug folder that is a file is packed without a word")
	}
}

func TestDebugReport_packsTheLogFilesWholeAndNothingElse(t *testing.T) {
	store := filepath.Join(t.TempDir(), "cairn.db")
	kept := Places{Store: store}.Debug()
	if err := os.MkdirAll(filepath.Join(kept, "b.log"), 0o700); err != nil {
		t.Fatal(err)
	}
	for name, said := range map[string]string{"a.txt": "other", "c.log": "first", "d.log": "second"} {
		if err := os.WriteFile(filepath.Join(kept, name), []byte(said), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	answer, err := DoAt(DebugReport, Places{Store: store, Home: Home{Dir: t.TempDir()}}, Input{Given: map[string]string{"to": t.TempDir()}}, Acting{})
	if err != nil {
		t.Fatal(err)
	}
	if answer.Packed == nil || *answer.Packed != 2 {
		t.Fatalf("the report answered %+v", answer)
	}
	packed, err := zip.OpenReader(*answer.Written)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = packed.Close() }()
	held := map[string]string{}
	for _, f := range packed.File {
		read, err := f.Open()
		if err != nil {
			t.Fatal(err)
		}
		said, _ := io.ReadAll(read)
		_ = read.Close()
		held[f.Name] = string(said)
	}
	if held["debug/c.log"] != "first" || held["debug/d.log"] != "second" || len(held) != 3 {
		t.Errorf("the archive holds %v", held)
	}
}

func releasesSaying(t *testing.T, tag string) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprintf(w, `{"tag_name":%q,"assets":[]}`, tag)
	}))
	t.Cleanup(server.Close)
	was := upgrade.Latest
	upgrade.Latest = server.URL + "/latest"
	t.Cleanup(func() { upgrade.Latest = was })
}

func TestUpgrade_leavesTheNewestAsItIs(t *testing.T) {
	releasesSaying(t, "v0.10.0")
	binary := filepath.Join(t.TempDir(), "cairn")
	got, err := Upgrade("v0.10.0", binary, "darwin/arm64")
	if err != nil || got != (Upgraded{From: "v0.10.0", To: "v0.10.0", Binary: binary}) {
		t.Errorf("the newest already: %+v, %v", got, err)
	}
}

func TestUpgrade_saysAReleaseWithNothingForThisSystem(t *testing.T) {
	releasesSaying(t, "v0.11.0")
	if _, err := Upgrade("v0.10.0", filepath.Join(t.TempDir(), "cairn"), "darwin/arm64"); !errors.Is(err, upgrade.ErrNoBinaryHere) {
		t.Errorf("a release with no binary here: %v", err)
	}
}

func TestTaskShow_saysSinceWhenTheTaskStands(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cairn.db")
	if _, err := DoAt(TaskNew, Places{Store: path}, Input{Given: map[string]string{"title": "Ship it"}}, Acting{}); err != nil {
		t.Fatal(err)
	}
	shown, err := DoAt(TaskShow, Places{Store: path}, Input{Given: map[string]string{"number": "1"}}, Acting{})
	if err != nil {
		t.Fatal(err)
	}
	if shown.StoodSince.IsZero() {
		t.Error("a task shown says nothing of since when it stands")
	}
}

func TestSinceWhen_readsTheDaySaidOrTakesToday(t *testing.T) {
	got, err := sinceWhen(Input{Given: map[string]string{"since": "2026-01-02"}})
	want, _ := domain.ParseDeadline("2026-01-02")
	if err != nil || !got.Equal(want) {
		t.Errorf("since 2026-01-02 reads %v, %v", got, err)
	}
	now := time.Now()
	got, err = sinceWhen(Input{})
	if err != nil || !got.Equal(time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())) {
		t.Errorf("no day said reads %v, %v", got, err)
	}
}

func TestDebugging_isOffForAVariableSetEmpty(t *testing.T) {
	for said, on := range map[string]bool{"": false, "1": true} {
		h := Home{Lookup: func(name string) (string, bool) { return said, name == debug.Variable }}
		if got := Debugging(h); got != on {
			t.Errorf("%s=%q: debugging %v, want %v", debug.Variable, said, got, on)
		}
	}
}

func TestRender_writesTheContextAndMarksTheDeleted(t *testing.T) {
	if got := contextText("/notes/1"); got != "\ncontext  /notes/1\n" {
		t.Errorf("a context reads %q", got)
	}
	got := namedAsShown(map[int]string{1: "kept", 2: "gone"}, map[int]bool{2: true})
	if got[1] != "kept" || got[2] != "gone"+deletedMark {
		t.Errorf("the names read %v", got)
	}
}
