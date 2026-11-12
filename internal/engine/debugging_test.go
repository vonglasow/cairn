package engine

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"

	"cairn/internal/debug"
	"cairn/internal/task"
)

func debugLog(t *testing.T) *debug.Log {
	t.Helper()
	l, err := debug.Open(t.TempDir(), "test", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	NoteTo(l)
	t.Cleanup(func() { NoteTo(nil) })
	return l
}

func noted(t *testing.T, l *debug.Log) string {
	t.Helper()
	NoteTo(nil)
	_ = l.Close()
	read, err := os.ReadFile(l.Path())
	if err != nil {
		t.Fatal(err)
	}
	return string(read)
}

func TestNoteTo_writesAnOperationOnceWithWhatItWasGivenMasked(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cairn.db")
	l := debugLog(t)
	if _, err := DoAt(TaskNew, Places{Store: path}, Input{Given: map[string]string{"title": "Write the spec", "priority": "high"}}, Acting{}); err != nil {
		t.Fatal(err)
	}
	_, _ = DoAt(TaskShow, Places{Store: path}, Input{Given: map[string]string{"number": "99"}}, Acting{})
	said := noted(t, l)
	if strings.Count(said, "engine task new") != 1 || !strings.Contains(said, "priority=high") || !strings.Contains(said, "title=‹text ") {
		t.Errorf("the file reads:\n%s", said)
	}
	if strings.Contains(said, "Write the spec") {
		t.Errorf("the title reached the file:\n%s", said)
	}
	if !strings.Contains(said, "→ ok task #1 backlog") {
		t.Errorf("what it came to is not said:\n%s", said)
	}
	if !strings.Contains(said, "engine task show number=99 → refused") {
		t.Errorf("the refusal is not said:\n%s", said)
	}
}

func TestNoteTo_writesTheVersionABoardReadsOnlyWhenItChanges(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cairn.db")
	held, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = held.Close() }()
	l := debugLog(t)
	for range 3 {
		if _, err := held.Version(); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := DoAt(TaskNew, Places{Store: path}, Input{Given: map[string]string{"title": "Ship it"}}, Acting{}); err != nil {
		t.Fatal(err)
	}
	_, _ = held.Version()
	_, _ = held.Version()
	if got := strings.Count(noted(t, l), "engine version "); got != 2 {
		t.Errorf("the version is written %d times, want twice", got)
	}
}

func TestDebugHeader_saysTheBuildTheConfigurationAndTheStoreInNumbers(t *testing.T) {
	home := t.TempDir()
	at := Places{Store: filepath.Join(home, ".cairn", "cairn.db"), Home: Home{Dir: home}}
	if _, err := DoAt(TaskNew, at, Input{Given: map[string]string{"title": "Write the spec", "state": "todo"}}, Acting{}); err != nil {
		t.Fatal(err)
	}
	facts := map[string][]string{}
	for _, f := range DebugHeader(at, "task list") {
		facts[f[0]] = append(facts[f[0]], f[1])
	}
	if len(facts["cairn"]) != 1 || len(facts["system"]) != 1 || len(facts["config"]) == 0 {
		t.Errorf("the header holds %v", facts)
	}
	if got := strings.Join(facts["store"], "|"); got != "1 todo" {
		t.Errorf("the store reads %q", got)
	}
	if got := strings.Join(facts["command"], "|"); got != "task list" {
		t.Errorf("the command reads %q", got)
	}
	for _, line := range facts["config"] {
		if strings.Contains(line, home) {
			t.Errorf("the home is written whole: %q", line)
		}
	}
}

func TestDebugReport_packsEveryFileKeptAndAHeader(t *testing.T) {
	home := t.TempDir()
	store := filepath.Join(home, ".cairn", "cairn.db")
	for range 2 {
		l, err := debug.Open(Places{Store: store}.Debug(), "task list", time.Now())
		if err != nil {
			t.Fatal(err)
		}
		_ = l.Close()
	}
	to := t.TempDir()
	answer, err := DoAt(DebugReport, Places{Store: store, Home: Home{Dir: home}}, Input{Given: map[string]string{"to": to}}, Acting{})
	if err != nil {
		t.Fatal(err)
	}
	if answer.Written == nil || filepath.Dir(*answer.Written) != to || answer.Packed == nil || *answer.Packed != 2 {
		t.Fatalf("the report answered %+v", answer)
	}
	packed, err := zip.OpenReader(*answer.Written)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = packed.Close() }()
	var names []string
	for _, f := range packed.File {
		names = append(names, f.Name)
	}
	if len(names) != 3 || names[0] != "header.txt" || !strings.HasPrefix(names[1], "debug/") {
		t.Errorf("the archive holds %v", names)
	}
}

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
