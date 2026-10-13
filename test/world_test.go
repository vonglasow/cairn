package test

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cucumber/godog"

	"cairn/internal/config"
)

func answering(t *testing.T, answer Result) (*World, *[]string) {
	t.Helper()
	var asked []string
	w := NewWorld(t, func(_ *testing.T, args ...string) Result {
		asked = append(asked, strings.Join(args, " "))
		return answer
	})
	t.Cleanup(w.EndScenario)
	return w, &asked
}

func TestRun_handsOverWhatAShellWouldHaveHandedOver(t *testing.T) {
	w, asked := answering(t, Result{Stdout: "done"})
	w.Run(`cairn task new "Write the spec"`)
	if want := []string{`task new Write the spec`}; len(*asked) != 1 || (*asked)[0] != want[0] {
		t.Errorf("asked %#v, want %#v — the leading `cairn` is the reader's, not the surface's", *asked, want)
	}
	if w.Out.Stdout != "done" {
		t.Errorf("Out.Stdout = %q, want what the surface answered", w.Out.Stdout)
	}
}

func TestSeed_putsBackWhatThePreconditionAnswered(t *testing.T) {
	w, _ := answering(t, Result{Stdout: "seeded"})
	w.Out = Result{Stdout: "what the scenario is asserting against"}
	if err := w.Seed(`cairn task new "One"`); err != nil {
		t.Fatalf("Seed = %v, want it to succeed", err)
	}
	if w.Out.Stdout != "what the scenario is asserting against" {
		t.Errorf("Out.Stdout = %q: the precondition overwrote what is under test", w.Out.Stdout)
	}
}

func TestSeed_reportsAPreconditionThatCouldNotBeMet(t *testing.T) {
	w, _ := answering(t, Result{Code: 1, Stderr: "a task needs a title"})
	w.Out = Result{Stdout: "untouched"}
	err := w.Seed(`cairn task new ""`)
	if err == nil {
		t.Fatal("Seed = nil: a precondition that failed would leave the scenario asserting nothing")
	}
	for _, want := range []string{"cairn task new", "exit 1", "a task needs a title"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("%q does not carry %q", err, want)
		}
	}
	if w.Out.Stdout != "untouched" {
		t.Errorf("Out.Stdout = %q, want the failed precondition put back", w.Out.Stdout)
	}
}

func TestSetenv_putsBackAVariableThatWasThere(t *testing.T) {
	const key = "CAIRN_A_VARIABLE_THAT_WAS_THERE"
	t.Setenv(key, "before")
	w, _ := answering(t, Result{})
	w.Setenv(key, "during")
	if got := os.Getenv(key); got != "during" {
		t.Fatalf("%s = %q during the scenario", key, got)
	}
	w.EndScenario()
	if got := os.Getenv(key); got != "before" {
		t.Errorf("%s = %q after the scenario, want it put back", key, got)
	}
}

func TestSetenv_takesAwayAVariableThatWasNotThere(t *testing.T) {
	const key = "CAIRN_A_VARIABLE_THAT_WAS_NOT_THERE"
	w, _ := answering(t, Result{})
	w.Setenv(key, "during")
	w.EndScenario()
	if _, had := os.LookupEnv(key); had {
		t.Errorf("%s is still set: the next scenario starts on this one's leftovers", key)
	}
}

func TestEndScenario_undoesInReverseAndOnlyOnce(t *testing.T) {
	w, _ := answering(t, Result{})
	var order []int
	for _, n := range []int{1, 2, 3} {
		w.OnEnd(func() { order = append(order, n) })
	}
	w.EndScenario()
	if want := []int{3, 2, 1}; fmt.Sprint(order) != fmt.Sprint(want) {
		t.Errorf("undone %v, want %v — the last change is the first put back", order, want)
	}
	w.EndScenario()
	if len(order) != 3 {
		t.Errorf("undone %v: a scenario's undoing was replayed", order)
	}
}

func TestAStoreOfMyOwn_pointsSomewhereNothingElseLooks(t *testing.T) {
	w, _ := answering(t, Result{})
	if err := w.AStoreOfMyOwn(); err != nil {
		t.Fatalf("AStoreOfMyOwn = %v", err)
	}
	if w.StorePath == "" || w.Defaults == "" {
		t.Fatalf("StorePath = %q, Defaults = %q, want both chosen", w.StorePath, w.Defaults)
	}
	if strings.HasPrefix(w.StorePath, w.Defaults) {
		t.Error("the store sits inside the default place: the two are the same place, and the scenario means nothing")
	}
	if err := w.NothingInItYet(); err != nil {
		t.Errorf("NothingInItYet = %v, want the store still to be created", err)
	}
}

func TestNothingInItYet_refusesAStoreThatIsAlreadyThere(t *testing.T) {
	w, _ := answering(t, Result{})
	w.StorePath = filepath.Join(t.TempDir(), "cairn.db")
	if err := os.WriteFile(w.StorePath, []byte("a store"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := w.NothingInItYet(); err == nil {
		t.Error("NothingInItYet = nil, want it to report the store it found")
	}
}

func TestIAddedATaskTitled_goesThroughTheRealCommand(t *testing.T) {
	w, asked := answering(t, Result{})
	if err := w.IAddedATaskTitled("Write the spec"); err != nil {
		t.Fatalf("IAddedATaskTitled = %v", err)
	}
	if len(*asked) != 1 || (*asked)[0] != "task new Write the spec" {
		t.Errorf("asked %#v, want the task added through the command", *asked)
	}
}

func TestIRun_keepsWhatTheSurfaceAnswered(t *testing.T) {
	w, _ := answering(t, Result{Stdout: "out", Stderr: "err", Code: 2})
	if err := w.IRun("cairn nope"); err != nil {
		t.Fatalf("IRun = %v", err)
	}
	if w.Out != (Result{Stdout: "out", Stderr: "err", Code: 2}) {
		t.Errorf("Out = %#v, want the three of them kept apart", w.Out)
	}
}

func TestItSucceeded_readsTheCode(t *testing.T) {
	w, _ := answering(t, Result{})
	w.Out = Result{Code: 0}
	if err := w.ItSucceeded(); err != nil {
		t.Errorf("ItSucceeded = %v on code 0", err)
	}
	w.Out = Result{Code: 1, Stderr: "the reason"}
	err := w.ItSucceeded()
	if err == nil {
		t.Fatal("ItSucceeded = nil on code 1")
	}
	if !strings.Contains(err.Error(), "the reason") {
		t.Errorf("%q does not carry the reason", err)
	}
}

func TestItFailed_readsTheCode(t *testing.T) {
	w, _ := answering(t, Result{})
	w.Out = Result{Code: 1}
	if err := w.ItFailed(); err != nil {
		t.Errorf("ItFailed = %v on code 1", err)
	}
	w.Out = Result{Code: 0, Stdout: "it worked"}
	err := w.ItFailed()
	if err == nil {
		t.Fatal("ItFailed = nil on code 0")
	}
	if !strings.Contains(err.Error(), "it worked") {
		t.Errorf("%q does not show what it answered instead", err)
	}
}

func TestItSaysHowToUseIt_looksAtBothStreams(t *testing.T) {
	w, _ := answering(t, Result{})
	for _, said := range []Result{
		{Stderr: "USAGE\n  cairn mcp serve"},
		{Stdout: "Usage:\n  cairn"},
		{Stderr: "try --help for what it can do"},
	} {
		w.Out = said
		if err := w.ItSaysHowToUseIt(); err != nil {
			t.Errorf("ItSaysHowToUseIt = %v for %#v", err, said)
		}
	}
	w.Out = Result{Stderr: "a task needs a title"}
	if err := w.ItSaysHowToUseIt(); err == nil {
		t.Error("ItSaysHowToUseIt = nil for an answer that says nothing of the sort")
	}
}

func TestItSaysWhatItExpectsInsteadOf_wantsTheValueQuotedBack(t *testing.T) {
	w, _ := answering(t, Result{})
	w.Out = Result{Stderr: `"sideways" is not an access: read, write or full`}
	if err := w.ItSaysWhatItExpectsInsteadOf("sideways"); err != nil {
		t.Errorf("ItSaysWhatItExpectsInsteadOf = %v", err)
	}
	w.Out = Result{Stderr: "that is not an access"}
	if err := w.ItSaysWhatItExpectsInsteadOf("sideways"); err == nil {
		t.Error("ItSaysWhatItExpectsInsteadOf = nil for a refusal that names nothing")
	}
	w.Out = Result{Stdout: `"sideways" is not an access`}
	if err := w.ItSaysWhatItExpectsInsteadOf("sideways"); err == nil {
		t.Error("ItSaysWhatItExpectsInsteadOf = nil: a reason on standard output is a reason nobody piping the answer will read")
	}
}

func TestAHomeOfMyOwn_doesNotPointAtAStore(t *testing.T) {
	t.Setenv(config.StoreVariable, "a store nothing here should touch")
	w, _ := answering(t, Result{})
	if err := w.AHomeOfMyOwn(); err != nil {
		t.Fatalf("AHomeOfMyOwn = %v", err)
	}
	if w.Defaults == "" || os.Getenv("HOME") != w.Defaults {
		t.Fatalf("HOME = %q, Defaults = %q, want HOME pointed at a fresh Defaults", os.Getenv("HOME"), w.Defaults)
	}
	if got := os.Getenv(config.StoreVariable); got != "a store nothing here should touch" {
		t.Errorf("%s = %q, want it left untouched — a home of my own names no store", config.StoreVariable, got)
	}
}

func TestConfigPath_composesUnderTheHomeOrIsEmptyWithoutOne(t *testing.T) {
	w, _ := answering(t, Result{})
	if got := w.ConfigPath(); got != "" {
		t.Errorf("ConfigPath = %q, want empty with no home posed", got)
	}
	w.Defaults = t.TempDir()
	if want, got := filepath.Join(w.Defaults, ".config", "cairn", "config.toml"), w.ConfigPath(); got != want {
		t.Errorf("ConfigPath = %q, want %q", got, want)
	}
}

func TestSetAsideConfig_movesAFileAsideAndBackModeIncluded(t *testing.T) {
	w, _ := answering(t, Result{})
	w.Defaults = t.TempDir()
	path := w.ConfigPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	const mode = 0o640
	if err := os.WriteFile(path, []byte("a configuration"), mode); err != nil {
		t.Fatal(err)
	}
	restore := w.setAsideConfig()
	if _, err := os.Stat(path); err == nil {
		t.Fatal("the configuration is still there, want it set aside")
	}
	restore()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat(%s) = %v, want the configuration put back", path, err)
	}
	if info.Mode().Perm() != mode {
		t.Errorf("mode = %o, want %o carried along with the file", info.Mode().Perm(), mode)
	}
}

func TestSetAsideConfig_isANoOpWhenNothingIsThere(t *testing.T) {
	w, _ := answering(t, Result{})
	w.Defaults = t.TempDir()
	path := w.ConfigPath()
	restore := w.setAsideConfig()
	restore()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("Stat(%s) = %v, want nothing created where there was nothing", path, err)
	}
}

func TestLineListsTasks_runsTheLineAndCounts(t *testing.T) {
	w, asked := answering(t, Result{Stdout: "#1  One\n#2  Two\n"})
	if err := w.LineListsTasks("cairn task list", 2); err != nil {
		t.Errorf("LineListsTasks = %v", err)
	}
	if len(*asked) != 1 || (*asked)[0] != "task list" {
		t.Errorf("asked %#v, want the line run", *asked)
	}
	if err := w.LineListsTasks("cairn task list", 3); err == nil {
		t.Error("LineListsTasks = nil for a count that does not match")
	}
}

func TestRegister_wiresTheStepsEverySurfaceShares(t *testing.T) {
	w, asked := answering(t, Result{Code: 0, Stdout: "#1  Write the spec\n"})
	here, had := os.LookupEnv(config.StoreVariable)

	suite := godog.TestSuite{
		ScenarioInitializer: w.Register,
		Options: &godog.Options{
			Format: "pretty",
			Output: io.Discard,
			Strict: true,
			FeatureContents: []godog.Feature{{
				Name: "the vocabulary every surface shares",
				Contents: []byte(`Feature: the vocabulary every surface shares
  Scenario: the shared steps are wired and the world is put back
    Given a store of my own, somewhere that is not the default place
    And nothing in it yet
    And I added a task titled "Write the spec"
    When I run ` + "`cairn task list`" + `
    Then it succeeds
    And ` + "`cairn task list`" + ` lists 1 task
`),
			}},
		},
	}
	if code := suite.Run(); code != 0 {
		t.Fatalf("the shared steps ran to %d, want 0 — something is not wired", code)
	}
	if len(*asked) == 0 {
		t.Error("nothing reached the surface: the steps were wired to nothing")
	}
	if now, still := os.LookupEnv(config.StoreVariable); still != had || now != here {
		t.Errorf("the store variable is %q (set: %v), and it was %q (set: %v) before",
			now, still, here, had)
	}
}

func TestAStoreFromAnOlderCairn_isMadeThroughTheCommands(t *testing.T) {
	w, asked := answering(t, Result{})
	w.StorePath = filepath.Join(t.TempDir(), "cairn.db")
	backups := filepath.Join(filepath.Dir(w.StorePath), "backups")
	if err := os.MkdirAll(backups, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := w.AStoreFromAnOlderCairn("Write the spec"); err != nil {
		t.Fatal(err)
	}
	if strings.Join(*asked, "; ") != "task new Write the spec; migration rollback --yes" {
		t.Errorf("asked %#v", *asked)
	}
	if _, err := os.Stat(backups); !os.IsNotExist(err) {
		t.Error("the backup undoing set aside is still there")
	}
	refused, _ := answering(t, Result{Code: 1})
	if err := refused.AStoreFromAnOlderCairn("Write the spec"); err == nil {
		t.Error("a store that could not be made was reported as made")
	}
}

func TestWriteConfig_posesTheFileWhereCairnReadsIt(t *testing.T) {
	w, _ := answering(t, Result{})
	w.Defaults = t.TempDir()
	if err := w.WriteConfig("notes = \"~/vault\"\n"); err != nil {
		t.Fatal(err)
	}
	if read, err := os.ReadFile(w.ConfigPath()); err != nil || string(read) != "notes = \"~/vault\"\n" {
		t.Errorf("the configuration reads %q, %v", read, err)
	}
}
