package test

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cucumber/godog"

	"cairn/internal/config"
	"cairn/internal/engine"
	"cairn/internal/migration"
)

// World carries what one scenario needs between its steps.
type World struct {
	T         *testing.T
	Out       Result
	StorePath string
	Defaults  string
	Line      string
	Here      string

	run  Runner
	env  map[string]string
	undo []func()
}

// NewWorld builds the state one scenario works against, reaching its surface through run.
func NewWorld(t *testing.T, run Runner) *World { return &World{T: t, run: run} }

// Setenv sets an environment variable for this scenario alone: what it runs sees it, the process and the scenarios running beside it do not.
func (w *World) Setenv(key, value string) {
	if w.env == nil {
		w.env = map[string]string{}
	}
	w.env[key] = value
}

// Lookup reads a variable the way what this scenario runs sees it: what the scenario set, else what the process has.
func (w *World) Lookup(key string) (string, bool) {
	if value, set := w.env[key]; set {
		return value, true
	}
	return os.LookupEnv(key)
}

// Home is the home what this scenario runs is in, read through the scenario's own environment.
func (w *World) Home() engine.Home {
	dir, _ := w.Lookup(homeVariable())
	if dir == "" {
		return engine.Home{Missing: fmt.Errorf("$%s is not defined", homeVariable()), Lookup: w.Lookup}
	}
	return engine.Home{Dir: dir, Lookup: w.Lookup, Here: w.Here}
}

func homeVariable() string {
	if runtime.GOOS == "windows" {
		return "USERPROFILE"
	}
	return "HOME"
}

// Environ is the environment a program this scenario starts is handed: the process's, with what the scenario set over it.
func (w *World) Environ() []string {
	var env []string
	for _, v := range os.Environ() {
		if key, _, _ := strings.Cut(v, "="); !w.setHere(key) {
			env = append(env, v)
		}
	}
	for key, value := range w.env {
		env = append(env, key+"="+value)
	}
	return env
}

func (w *World) setHere(key string) bool {
	_, set := w.env[key]
	return set
}

// SetProcessEnv sets a variable in the process itself, for what reads it there rather than through a home, and puts it back when the scenario ends; only a suite whose scenarios run one at a time may call it.
func (w *World) SetProcessEnv(key, value string) {
	previous, had := os.LookupEnv(key)
	w.OnEnd(func() {
		if had {
			_ = os.Setenv(key, previous)
			return
		}
		_ = os.Unsetenv(key)
	})
	_ = os.Setenv(key, value)
}

// OnEnd registers something to put back when the scenario ends.
func (w *World) OnEnd(undo func()) { w.undo = append(w.undo, undo) }

// EndScenario puts back everything the scenario changed, most recent first.
func (w *World) EndScenario() {
	for i := len(w.undo) - 1; i >= 0; i-- {
		w.undo[i]()
	}
	w.undo = nil
	w.env = nil
}

// Run executes a command line the way a shell would have handed it over, dropping the leading `cairn` the scenario writes for the reader's benefit.
func (w *World) Run(line string) {
	args := SplitArgs(line)
	w.Line = line
	w.Out = w.run(w.T, w.Home(), args[1:]...)
}

// Seed puts cairn in a given state through the real command, so that what a precondition writes is what the command under test would have written; a store the scenario points at and that is not there yet is first laid down from one this cairn made once for the whole run.
func (w *World) Seed(line string) error {
	if err := w.layDownTheStoreWhenThereIsNone(); err != nil {
		return err
	}
	restore := w.setAsideConfig()
	defer restore()

	before := w.Out
	w.Run(line)
	if w.Out.Code != ok {
		got := w.Out
		w.Out = before
		return fmt.Errorf("precondition %q failed: exit %d (%s)", line, got.Code, got.Stderr)
	}
	w.Out = before
	return nil
}

func (w *World) layDownTheStoreWhenThereIsNone() error {
	if at, _ := w.Lookup(config.StoreVariable); w.StorePath == "" || at != w.StorePath {
		return nil
	}
	if _, err := os.Stat(w.StorePath); !errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	laid, err := anEmptyStore()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(w.StorePath), 0o700); err != nil {
		return err
	}
	return os.WriteFile(w.StorePath, laid, 0o600)
}

var anEmptyStore = sync.OnceValues(func() ([]byte, error) {
	dir, err := os.MkdirTemp("", "cairn-empty-store-")
	if err != nil {
		return nil, err
	}
	defer func() { _ = os.RemoveAll(dir) }()
	path := filepath.Join(dir, "cairn.db")
	made, err := engine.Open(path)
	if err != nil {
		return nil, err
	}
	if err := made.Close(); err != nil {
		return nil, err
	}
	return os.ReadFile(path)
})

// ConfigPath is where the configuration cairn reads sits, for this scenario's home — the one place a precondition needs to know it, so it is spelled here rather than a second time wherever else asks.
func (w *World) ConfigPath() string {
	if w.Defaults == "" {
		return ""
	}
	return filepath.Join(w.Defaults, ".config", "cairn", "config.toml")
}

// WriteConfig poses a configuration file by hand, the way a person would have — never through cairn, so a scenario asserting what cairn does with it is not also asserting that cairn wrote it; the file is replaced whole, so a server reading it meanwhile never reads it half-written.
func (w *World) WriteConfig(content string) error {
	path := w.ConfigPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	beside, err := os.CreateTemp(filepath.Dir(path), "config-*.toml")
	if err != nil {
		return err
	}
	if _, err := beside.WriteString(content); err != nil {
		_ = beside.Close()
		return err
	}
	if err := beside.Close(); err != nil {
		return err
	}
	return os.Rename(beside.Name(), path)
}

// Archiving poses a configuration that archives on its own, counting periods of that length from that many days from today, a negative count being in the past; a length of 0 leaves it unsaid.
func (w *World) Archiving(length, from int) error {
	start := time.Now().AddDate(0, 0, from).Format(time.DateOnly)
	if length == 0 {
		return w.WriteConfig(fmt.Sprintf("[archive]\nstart = %s\n", start))
	}
	return w.WriteConfig(fmt.Sprintf("[archive]\nstart = %s\ndays = %d\n", start, length))
}

// AStoreFromAnOlderCairn turns the store into one an older cairn would have written, holding one task: the task added and the last change to the shape undone, through the real commands.
func (w *World) AStoreFromAnOlderCairn(title string) error {
	if err := w.Seed(fmt.Sprintf(`cairn task new "%s"`, title)); err != nil {
		return err
	}
	if err := w.Seed(`cairn migration rollback --yes`); err != nil {
		return err
	}
	return w.ForgetBackups()
}

// ANewerCairn is a cairn that knows one change more than this one, UnknownChange, holding the store at path open whatever shape it is in; a store it applies that change to comes from a cairn newer than this one. Every other cairn in the process goes on knowing what it knew.
func ANewerCairn(path string) (*engine.Engine, error) {
	return engine.OpenBehindKnowing(path, append(slices.Clone(migration.Known), migration.Made(migration.Carried{
		Name: UnknownChange, Does: "SELECT 1", Undoes: "SELECT 1",
	})))
}

// UnknownChange is the name of the change a newer cairn knows and this one does not.
const UnknownChange = "never heard of"

// ForgetBackups removes every backup beside the store, so a scenario counts only what it caused.
func (w *World) ForgetBackups() error {
	return os.RemoveAll(filepath.Join(filepath.Dir(w.StorePath), "backups"))
}

func (w *World) setAsideConfig() func() {
	path := w.ConfigPath()
	if path == "" {
		return func() {}
	}
	aside := path + ".seed-aside"
	if err := os.Rename(path, aside); err != nil {
		return func() {}
	}
	return func() {
		_ = os.Rename(aside, path)
	}
}

// AHomeOfMyOwn points HOME at a fresh directory this run owns, and nothing else — no store pointed at, so a scenario about where a place falls when nothing else names it is left free to say so.
func (w *World) AHomeOfMyOwn() error {
	w.Defaults = w.T.TempDir()
	w.Setenv("HOME", w.Defaults)
	w.Setenv(engine.StateVariable, "")
	w.Setenv(engine.WorkspaceVariable, "")
	return nil
}

// AStoreOfMyOwn points cairn at a store nothing else will look at, and moves the default place somewhere a step may look at.
func (w *World) AStoreOfMyOwn() error {
	mine := w.T.TempDir()
	w.StorePath = filepath.Join(mine, "mine", "cairn.db")
	w.Defaults = w.T.TempDir()
	w.Setenv(config.StoreVariable, w.StorePath)
	w.Setenv("HOME", w.Defaults)
	w.Setenv(engine.StateVariable, "")
	w.Setenv(engine.WorkspaceVariable, "")
	return nil
}

// NothingInItYet fails unless the store has still to be created.
func (w *World) NothingInItYet() error {
	if _, err := os.Stat(w.StorePath); err == nil {
		return fmt.Errorf("the store at %s already exists", w.StorePath)
	}
	return nil
}

// IAddedATaskTitled puts a task in the store through the real command.
func (w *World) IAddedATaskTitled(title string) error {
	return w.Seed(fmt.Sprintf(`cairn task new "%s"`, title))
}

// IAddedAnObjectiveTitled seeds the store with one objective, standing on its own.
func (w *World) IAddedAnObjectiveTitled(title string) error {
	return w.Seed(fmt.Sprintf(`cairn objective new "%s"`, title))
}

// IAddedAnObjectiveUnder seeds the store with one objective, put under the one numbered parent.
func (w *World) IAddedAnObjectiveUnder(title string, parent int) error {
	return w.Seed(fmt.Sprintf(`cairn objective new "%s" --parent %d`, title, parent))
}

// IRun hands a command line over and keeps what came back.
func (w *World) IRun(line string) error {
	w.Run(line)
	return nil
}

// ItSucceeded fails unless the last command line reported nothing wrong.
func (w *World) ItSucceeded() error {
	if w.Out.Code != ok {
		return fmt.Errorf("exit %d, want %d (%s)", w.Out.Code, ok, w.Out.Stderr)
	}
	return nil
}

// ItFailed fails unless the last command line reported something wrong.
func (w *World) ItFailed() error {
	if w.Out.Code == ok {
		return fmt.Errorf("it succeeded, want a failure:\n%s", w.Out.Stdout)
	}
	return nil
}

// ItSaysHowToUseIt fails unless the answer points at the usage, on either stream: a surface may put it where it likes, as long as it is said.
func (w *World) ItSaysHowToUseIt() error {
	said := strings.ToLower(w.Out.Stderr + w.Out.Stdout)
	for _, hint := range []string{"usage", "try --help"} {
		if strings.Contains(said, hint) {
			return nil
		}
	}
	return fmt.Errorf("it does not say how to use it:\n%s", said)
}

// ItSaysWhatItExpectsInsteadOf fails unless the reason quotes the value cairn could not read.
func (w *World) ItSaysWhatItExpectsInsteadOf(value string) error {
	if w.Out.Stderr == "" {
		return fmt.Errorf("stderr is empty, want what it expects instead")
	}
	return WantContains(w.Out.Stderr, value, "the reason")
}

// LineListsTasks runs a command line and counts the tasks it listed.
func (w *World) LineListsTasks(line string, n int) error {
	w.Run(line)
	return CountsTasks(w.Out.Stdout, n)
}

// CommandSays runs a command line and fails unless its answer mentions what is wanted.
func (w *World) CommandSays(command, want string) error {
	w.Run(command)
	return WantContains(w.Out.Stdout, want, "the answer of "+command)
}

// FocusSessionsLast poses a configuration in which a focus session asks to be renewed once it has been up that long.
func (w *World) FocusSessionsLast(length string) error {
	return w.WriteConfig(fmt.Sprintf("[focus]\nsession = %q\n", length))
}

// Recurrence turns recurrence on or off in the settings beside the store, the way a person would write them — never through cairn.
func (w *World) Recurrence(said string) error {
	path := config.SettingsBeside(w.StorePath)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	return os.WriteFile(path, fmt.Appendf(nil, "[series]\non = %t\n", said == "on"), 0o600)
}

// ItIsLater moves the moment cairn takes for now that far ahead of the clock.
func (w *World) ItIsLater(count int, unit string) error {
	later := time.Duration(count) * time.Minute
	if unit == "hours" {
		later = time.Duration(count) * time.Hour
	}
	w.Setenv(engine.NowVariable, time.Now().Add(later).Format(time.RFC3339))
	return nil
}

// Register declares the steps more than one surface claims, and puts the world back when the scenario ends.
func (w *World) Register(sc *godog.ScenarioContext) {
	sc.After(w.afterScenario)

	sc.Step(`^a home of my own, made for this run$`, w.AHomeOfMyOwn)
	sc.Step(`^a store of my own, somewhere that is not the default place$`, w.AStoreOfMyOwn)
	sc.Step(`^nothing in it yet$`, w.NothingInItYet)
	sc.Step(`^I added a task titled "([^"]*)"$`, w.IAddedATaskTitled)
	sc.Step("^I run `([^`]*)`$", w.IRun)
	sc.Step(`^it succeeds$`, w.ItSucceeded)
	sc.Step(`^it fails$`, w.ItFailed)
	sc.Step(`^it says how to use it$`, w.ItSaysHowToUseIt)
	sc.Step(`^it says what it expects instead of "([^"]*)"$`, w.ItSaysWhatItExpectsInsteadOf)
	sc.Step("^`([^`]*)` says \"([^\"]*)\"$", w.CommandSays)
	sc.Step("^`([^`]*)` lists (\\d+) tasks?$", w.LineListsTasks)
	sc.Step("^`([^`]*)` lists (\\d+) objectives?$", w.LineListsTasks)
	sc.Step(`^I added an objective titled "([^"]*)"$`, w.IAddedAnObjectiveTitled)
	sc.Step(`^I added an objective titled "([^"]*)" under objective (\d+)$`, w.IAddedAnObjectiveUnder)
	sc.Step(`^I put task (\d+) in focus$`, func(n int) error {
		return w.Seed(fmt.Sprintf(`cairn task focus %d`, n))
	})
	sc.Step(`^I archived task (\d+)$`, func(n int) error {
		return w.Seed(fmt.Sprintf("cairn task archive %d", n))
	})
	sc.Step(`^focus sessions last "([^"]*)"$`, w.FocusSessionsLast)
	sc.Step(`^the window keeps its sizes$`, func() error { return w.WriteConfig("[window]\nkeep_sizes = true\n") })
	sc.Step(`^the window's keep_sizes is "([^"]*)"$`, func(said string) error {
		return w.WriteConfig(fmt.Sprintf("[window]\nkeep_sizes = %q\n", said))
	})
	sc.Step(`^the board arrives by "([^"]*)"$`, func(said string) error {
		return w.WriteConfig(fmt.Sprintf("[window]\nintro = %q\n", said))
	})
	sc.Step(`^a deleted task takes its time with it$`, func() error { return w.WriteConfig("[time]\ndeleted = \"taken\"\n") })
	sc.Step(`^the time of a deleted task is "([^"]*)"$`, func(said string) error {
		return w.WriteConfig(fmt.Sprintf("[time]\ndeleted = %q\n", said))
	})
	sc.Step(`^a configuration that is not TOML$`, func() error { return w.WriteConfig("this is not toml: [[[\n") })
	sc.Step(`^the weekends are hidden$`, func() error { return w.WriteConfig("[time]\nweekends = \"hidden\"\n") })
	sc.Step(`^the weekends are "([^"]*)"$`, func(said string) error {
		return w.WriteConfig(fmt.Sprintf("[time]\nweekends = %q\n", said))
	})
	sc.Step(`^it is (\d+) (minutes|hours) later$`, w.ItIsLater)
	sc.Step(`^recurrence is (on|off) for this store$`, w.Recurrence)
	sc.Step(`^today is "(\d{4}-\d{2}-\d{2})"$`, func(day string) error {
		w.Setenv(engine.TodayVariable, day)
		return nil
	})
	sc.Step(`^it is (\d+) days from now$`, func(days int) error {
		w.Setenv("CAIRN_TODAY", time.Now().AddDate(0, 0, days).Format(time.DateOnly))
		return nil
	})
	sc.Step(`^archiving every (\d+) days, counted from (\d+) days ago$`, func(length, ago int) error {
		return w.Archiving(length, -ago)
	})
	sc.Step(`^archiving every (\d+) days, counted from (\d+) days from now$`, w.Archiving)
	sc.Step(`^archiving counted from (\d+) days ago, every so many days not said$`, func(ago int) error {
		return w.Archiving(0, -ago)
	})
}

func (w *World) afterScenario(ctx context.Context, _ *godog.Scenario, err error) (context.Context, error) {
	w.EndScenario()
	return ctx, err
}
