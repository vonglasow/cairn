package test

// The world every feature file shares, and the steps they have in common.
//
// A step vocabulary is worth more than the feature that first needed it:
// `I added a task titled "..."` reads the same whether the scenario is about
// tasks, about the store that holds them, or about an agent asking over MCP.
// Feature-specific steps live beside their feature; everything here is claimed
// by more than one surface.
//
// Each step is a named method rather than a closure inside Register. A closure
// can only be reached through godog, and a harness nothing can call directly is
// a harness nothing checks: a step that stopped asserting would show up as a
// scenario passing for the wrong reason, in whichever surface happened to run
// it. Register is then pure wiring, and holds no judgement of its own.

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cucumber/godog"

	"cairn/internal/engine"
)

// World carries what one scenario needs between its steps.
//
// Out, StorePath and Defaults are read by the steps a surface keeps to itself,
// so they are fields rather than accessors: a surface extending this world
// reads the same state the shared steps wrote.
type World struct {
	T *testing.T
	// Out is what the last command line answered.
	Out Result
	// StorePath is the store this scenario works against.
	StorePath string
	// Defaults is where the default place was moved to, somewhere a step may
	// look at.
	Defaults string

	run  Runner
	undo []func()
}

// NewWorld builds the state one scenario works against, reaching its surface
// through run.
func NewWorld(t *testing.T, run Runner) *World { return &World{T: t, run: run} }

// Setenv sets an environment variable for the length of the scenario. Not
// t.Setenv: that restores when the Go test ends, so the next scenario in the
// same run would start on this one's leftovers.
func (w *World) Setenv(key, value string) {
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

// OnEnd registers something to put back when the scenario ends. Anything a step
// changes outside its own world — an environment, a file mode — is undone here.
func (w *World) OnEnd(undo func()) { w.undo = append(w.undo, undo) }

// EndScenario puts back everything the scenario changed, most recent first.
func (w *World) EndScenario() {
	for i := len(w.undo) - 1; i >= 0; i-- {
		w.undo[i]()
	}
	w.undo = nil
}

// Run executes a command line the way a shell would have handed it over,
// dropping the leading `cairn` the scenario writes for the reader's benefit.
func (w *World) Run(line string) {
	args := SplitArgs(line)
	w.Out = w.run(w.T, args[1:]...)
}

// Seed puts cairn in a given state through the real command, so that what a
// precondition writes is what the command under test would have written. Going
// around the command would leave nothing proving `task add` writes at all.
//
// What the precondition answered is put back afterwards: a scenario asserts
// against the command under test, not against the one that set it up.
func (w *World) Seed(line string) error {
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

// AStoreOfMyOwn points cairn at a store nothing else will look at, and moves
// the default place somewhere a step may look at.
func (w *World) AStoreOfMyOwn() error {
	mine := w.T.TempDir()
	w.StorePath = filepath.Join(mine, "mine", "cairn.db")
	// The default place is moved somewhere else entirely, and somewhere a step
	// may look at. Pointing it at the store's own folder would make the two the
	// same place and empty the scenario of its meaning.
	w.Defaults = w.T.TempDir()
	w.Setenv(engine.StoreVariable, w.StorePath)
	w.Setenv("HOME", w.Defaults)
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
	return w.Seed(fmt.Sprintf(`cairn task add "%s"`, title))
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

// ItSaysHowToUseIt fails unless the answer points at the usage, on either
// stream: a surface may put it where it likes, as long as it is said.
func (w *World) ItSaysHowToUseIt() error {
	said := strings.ToLower(w.Out.Stderr + w.Out.Stdout)
	for _, hint := range []string{"usage", "try --help"} {
		if strings.Contains(said, hint) {
			return nil
		}
	}
	return fmt.Errorf("it does not say how to use it:\n%s", said)
}

// ItSaysWhatItExpectsInsteadOf fails unless the reason quotes the value cairn
// could not read.
//
// Quoting what was turned down is the whole of "nothing is guessed": a refusal
// that does not name the value leaves somebody who passed three flags with no
// idea which one cairn could not read.
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

// Register declares the steps more than one surface claims, and puts the world
// back when the scenario ends. It decides nothing: every judgement it wires up
// is a method above, reachable without godog.
func (w *World) Register(sc *godog.ScenarioContext) {
	sc.After(w.afterScenario)

	sc.Step(`^a store of my own, somewhere that is not the default place$`, w.AStoreOfMyOwn)
	sc.Step(`^nothing in it yet$`, w.NothingInItYet)
	sc.Step(`^I added a task titled "([^"]*)"$`, w.IAddedATaskTitled)
	sc.Step("^I run `([^`]*)`$", w.IRun)
	sc.Step(`^it succeeds$`, w.ItSucceeded)
	sc.Step(`^it fails$`, w.ItFailed)
	sc.Step(`^it says how to use it$`, w.ItSaysHowToUseIt)
	sc.Step(`^it says what it expects instead of "([^"]*)"$`, w.ItSaysWhatItExpectsInsteadOf)
	sc.Step("^`([^`]*)` lists (\\d+) tasks?$", w.LineListsTasks)
}

// afterScenario puts the world back, whatever the scenario did to it.
func (w *World) afterScenario(ctx context.Context, _ *godog.Scenario, err error) (context.Context, error) {
	w.EndScenario()
	return ctx, err
}
