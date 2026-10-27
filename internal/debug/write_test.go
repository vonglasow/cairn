package debug

import (
	"bytes"
	"errors"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"
)

func opened(t *testing.T) *Log {
	t.Helper()
	l, err := Open(t.TempDir(), "task list", moment)
	if err != nil {
		t.Fatal(err)
	}
	l.now = func() time.Time { return moment.Add(512 * time.Millisecond) }
	return l
}

func written(t *testing.T, l *Log) string {
	t.Helper()
	_ = l.Close()
	read, err := os.ReadFile(l.Path())
	if err != nil {
		t.Fatal(err)
	}
	return string(read)
}

func TestLog_saysALineWithTheMomentAndItsKind(t *testing.T) {
	l := opened(t)
	l.Say("engine", "task list → ok 3ms")
	l.Fact("system", "darwin/arm64")
	got := written(t, l)
	if !strings.Contains(got, "14:03:04.512 engine task list → ok 3ms\n") {
		t.Errorf("the file reads:\n%s", got)
	}
	if !strings.Contains(got, "# system           darwin/arm64\n") {
		t.Errorf("the fact reads:\n%s", got)
	}
}

func TestText_isAFingerprintAndALengthTheSameAcrossOneFile(t *testing.T) {
	l := opened(t)
	one, again, other := l.Text("Write the spec"), l.Text("Write the spec"), l.Text("Ship it")
	if !regexp.MustCompile(`^‹text [0-9a-f]{4}·14›$`).MatchString(one) {
		t.Errorf("a title is masked as %q", one)
	}
	if one != again || one == other {
		t.Errorf("the same text reads %q and %q, another %q", one, again, other)
	}
	if strings.Contains(one, "Write") {
		t.Errorf("the mask holds the text: %q", one)
	}
	elsewhere := opened(t)
	if elsewhere.Text("Write the spec") == one && elsewhere.Text("Ship it") == other {
		t.Error("two files fingerprint alike: the salt is not their own")
	}
	if got := l.Text(""); got != "‹text 0›" {
		t.Errorf("nothing is masked as %q", got)
	}
}

func TestArgs_keepsTheCommandAndTheFlagsAndMasksWhatWasWritten(t *testing.T) {
	l := opened(t)
	got := l.Args([]string{"task", "new"}, []string{"task", "new", "Write the spec", "--priority", "high", "--deadline=2026-12-01", "4"})
	if !strings.HasPrefix(got, "task new ‹text ") || !strings.Contains(got, " --priority ‹text ") ||
		!strings.Contains(got, " --deadline=‹text ") || !strings.HasSuffix(got, " 4") {
		t.Errorf("the arguments read %q", got)
	}
	if strings.Contains(got, "Write") || strings.Contains(got, "high") {
		t.Errorf("a value written by the person is plain: %q", got)
	}
}

func TestPanicked_writesTheValueAndTheStack(t *testing.T) {
	l := opened(t)
	l.Panicked("index out of range", []byte("goroutine 1 [running]:\nmain.main()\n"))
	got := written(t, l)
	if !strings.Contains(got, "panic index out of range\n") || !strings.Contains(got, "goroutine 1 [running]:") {
		t.Errorf("the panic reads:\n%s", got)
	}
}

func TestLeave_closesTheFileAndSaysWhereOnce(t *testing.T) {
	l := opened(t)
	var said bytes.Buffer
	l.Leave(&said)
	l.Leave(&said)
	if got := said.String(); got != "debug: written down in "+l.Path()+"\n" {
		t.Errorf("leaving says %q", got)
	}
	if _, err := l.file.WriteString("more"); !errors.Is(err, os.ErrClosed) {
		t.Errorf("the file is still open: %v", err)
	}
}

func TestMachine_saysWhatTheSystemIsAndWhatIsMissing(t *testing.T) {
	env := map[string]string{"LANG": "fr_FR.UTF-8", "TERM": "xterm-256color", "XDG_SESSION_TYPE": "wayland"}
	ran := func(name string, _ ...string) string {
		if name == "setxkbmap" {
			return ""
		}
		return name + " said"
	}
	facts := map[string]string{}
	for _, f := range Machine("linux", "amd64", func(k string) string { return env[k] }, ran) {
		facts[f[0]] = f[1]
	}
	for name, want := range map[string]string{
		"system":   "linux/amd64",
		"session":  "wayland",
		"desktop":  "unset",
		"language": "fr_FR.UTF-8",
		"terminal": "xterm-256color",
		"keyboard": "not found",
	} {
		if facts[name] != want {
			t.Errorf("%s reads %q, want %q", name, facts[name], want)
		}
	}
	mac := map[string]string{}
	for _, f := range Machine("darwin", "arm64", func(string) string { return "" }, ran) {
		mac[f[0]] = f[1]
	}
	if mac["macos"] != "sw_vers said" || mac["language"] != "unset" {
		t.Errorf("a mac reads %v", mac)
	}
}

func TestTyped_isALengthAndNothingElse(t *testing.T) {
	l := opened(t)
	if got := l.Typed("é"); got != "‹typed 1›" {
		t.Errorf("a letter typed reads %q", got)
	}
	if l.Typed("a") != l.Typed("b") {
		t.Error("two letters typed read differently: a file of them could be read back")
	}
}
