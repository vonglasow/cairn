package engine

import (
	"archive/zip"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"cairn/internal/debug"
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
