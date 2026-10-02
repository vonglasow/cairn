package tui

// The world the board's scenarios live in. It reaches the engine for the
// givens — a scenario says what is in the store, not how it got there — and
// drives the board itself for everything else: the board is a value, so a key
// press is a call and the screen is a string, with no terminal anywhere.

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/cucumber/godog"

	"cairn/internal/engine"
	"cairn/internal/reading"
)

// stood is the day every scenario is run on, so a deadline reads the same
// today as it will next year.
var stood = time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC)

type world struct {
	t     *testing.T
	path  string
	held  *engine.Engine
	other *engine.Engine
	b     Board
	width int
	high  int
	asked bool
	// formAt is which field of the form the last step left the cursor on, so
	// the next one knows how far it still has to walk.
	formAt int
}

func newWorld(t *testing.T) *world {
	return &world{t: t, width: 100, high: 44}
}

func (w *world) reset() {
	if w.held != nil {
		_ = w.held.Close()
	}
	if w.other != nil {
		_ = w.other.Close()
	}
	w.path = filepath.Join(w.t.TempDir(), "cairn.db")
	held, err := engine.Open(w.path)
	if err != nil {
		w.t.Fatalf("opening the store: %v", err)
	}
	w.held = held
	// Everything a scenario writes goes through a second connection, because
	// that is what it is: somebody else, at a command line or over MCP. The
	// board notices another connection's writes and cannot notice its own, so
	// a scenario sharing the board's connection would be testing nothing.
	other, err := engine.Open(w.path)
	if err != nil {
		w.t.Fatalf("opening the store a second time: %v", err)
	}
	w.other = other
	w.b = Board{}
	w.asked = false
	w.formAt = 0
}

// open builds the board and gives it its first reading.
//
// It hands the reading over rather than running Init, and for a measured
// reason: Init asks for the beat as well, and running the beat's command
// sleeps for as long as the beat lasts. Thirty-eight scenarios each waiting two
// seconds was the whole of this suite's running time, and a suite that slow
// makes every mutant time out — which reads as a gate that passed while judging
// nothing. What the beat does is checked on its own, in model_test.go.
func (w *world) open() {
	w.b = New(w.held, GlyphsFor(Unicode), func() time.Time { return stood }, w.width, w.high)
	w.refresh()
}

// refresh hands the board a fresh reading, the way its own beat would once it
// noticed the store had moved.
func (w *world) refresh() {
	found, err := reading.Look(w.held)
	if err != nil {
		w.t.Fatalf("looking: %v", err)
	}
	model, cmd := w.b.Update(Looked{Snapshot: found})
	w.b = model.(Board)
	w.settle(cmd)
}

// settle runs a command and everything it leads to, so a scenario sees the
// board as it will be once the work has been done.
func (w *world) settle(cmd tea.Cmd) {
	for depth := 0; cmd != nil && depth < 8; depth++ {
		msg := cmd()
		if msg == nil {
			return
		}
		if batch, ok := msg.(tea.BatchMsg); ok {
			var next tea.Cmd
			for _, one := range batch {
				if one == nil {
					continue
				}
				inner := one()
				if inner == nil {
					continue
				}
				model, after := w.b.Update(inner)
				w.b = model.(Board)
				next = after
			}
			cmd = next
			continue
		}
		model, after := w.b.Update(msg)
		w.b = model.(Board)
		cmd = after
	}
}

// press sends one key and settles whatever it set off.
//
// A press that has nothing to do with the form is settled rather than raced
// against a deadline, and this is the whole difference between a scenario that
// holds and one that holds nine times in ten. Measured: `⇧J` writes to a real
// store and reads the board back, and under the load of ten suites at once
// that crossed the five milliseconds pump allows — so the refresh was thrown
// away and the board answered from its old reading, exactly once in ten runs
// of `prek run --all-files`.
//
// What decides is where the board ends up, not where it was. The press that
// saves a form is a press the form no longer has the window for, and its
// command is the write — judging by the state beforehand would have raced
// exactly the one press that must not be raced.
//
// While the form does have the window, the deadline stays: it is the form that
// asks for a cursor blink, and a blink is a timer nobody can wait out.
func (w *world) press(name string) {
	was := w.b.screen
	model, cmd := w.b.Update(keyPress(name))
	w.b = model.(Board)
	if w.b.screen == OnForm {
		w.b = pump(w.b, cmd)
	} else {
		w.settle(cmd)
	}
	if was != OnForm && w.b.screen == OnForm {
		w.b = settleForm(w.b)
	}
}

// screen is what the board says, with the styling taken off: a scenario reads
// words, not escape codes.
func (w *world) screen() string { return strip(w.b.Draw()) }

// strip takes the colour out of a rendered screen.
func strip(s string) string {
	var out strings.Builder
	inEscape := false
	for _, r := range s {
		switch {
		case r == 0x1b:
			inEscape = true
		case inEscape && r == 'm':
			inEscape = false
		case !inEscape:
			out.WriteRune(r)
		}
	}
	return out.String()
}

// register wires every step the board's scenarios use.
func (w *world) register(sc *godog.ScenarioContext) {
	sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
		w.reset()
		return ctx, nil
	})

	sc.Given(`^a store of my own, somewhere that is not the default place$`, func() error { return nil })
	sc.Given(`^nothing in it yet$`, func() error { return nil })
	sc.Given(`^a window of (\d+) by (\d+)$`, w.aWindow)

	sc.Step(`^I added a task titled "([^"]*)"$`, func(title string) error { return w.addTask(title, "", "", 0) })
	sc.Step(`^I added a task titled "([^"]*)" with the estimate "([^"]*)"$`, func(title, est string) error {
		_, err := w.other.Add(title, engine.Changes{Estimate: &est})
		w.afterAChange()
		return err
	})
	sc.Step(`^I archived task (\d+)$`, func(n int) error {
		if err := w.other.Archive(n); err != nil {
			return err
		}
		w.afterAChange()
		return nil
	})
	sc.Step(`^I added a task titled "([^"]*)" with the state "([^"]*)"$`, func(title, state string) error {
		return w.addTask(title, state, "", 0)
	})
	sc.Step(`^I added a task titled "([^"]*)" with the description "([^"]*)"$`, func(title, desc string) error {
		return w.addTask(title, "", desc, 0)
	})
	sc.Step(`^I added a task titled "([^"]*)" serving objective (\d+)$`, func(title string, n int) error {
		return w.addTask(title, "", "", n)
	})
	sc.Step(`^I added a task titled "([^"]*)" with the state "([^"]*)" serving objective (\d+)$`, func(title, state string, n int) error {
		return w.addTask(title, state, "", n)
	})
	sc.Step(`^I added an objective titled "([^"]*)"$`, w.addObjective)
	sc.Step(`^I edited task (\d+) with the blocker "([^"]*)"$`, w.blockTask)
	sc.Step(`^I edited task (\d+) with the state "([^"]*)"$`, w.stateTask)
	sc.Step(`^I edited task (\d+) with the labels "([^"]*)" and "([^"]*)"$`, w.labelTask)
	sc.Step(`^I deleted task (\d+)$`, w.deleteTask)
	sc.Step(`^I put task (\d+) in focus$`, w.focusTask)

	sc.When(`^I open the board$`, func() error { w.open(); return nil })
	sc.Step(`^I press (".*")$`, w.pressMany)
	sc.Step(`^I type "([^"]*)"$`, w.typeIn)
	sc.Step(`^I choose "([^"]*)"$`, w.choose)
	sc.Step(`^I set the field "([^"]*)" to "([^"]*)"$`, w.setField)
	sc.Step(`^I save the form$`, w.saveForm)
	sc.Step(`^the board looks again$`, w.lookAgain)

	sc.Then(`^the columns read (".*")$`, w.columnsRead)
	sc.Then(`^the band is named "([^"]*)"$`, w.bandNamed)
	sc.Then(`^the band carries (\d+) tasks?$`, w.bandCarries)
	sc.Then(`^"([^"]*)" carries "([^"]*)"$`, w.columnCarries)
	sc.Then(`^"([^"]*)" reads (".*")$`, w.columnReads)
	sc.Then(`^it says it needs (\d+) columns$`, w.saysItNeeds)
	sc.Then(`^it says the window is (\d+)$`, w.saysTheWindowIs)
	sc.Then(`^the mode is "([^"]*)"$`, w.modeIs)
	sc.Then(`^the tab is "([^"]*)"$`, w.tabIs)
	sc.Then(`^the cursor is on the column "([^"]*)"$`, w.cursorOnColumn)
	sc.Then(`^the cursor is on the task "([^"]*)"$`, w.cursorOnTask)
	sc.Then(`^task (\d+) is in the column "([^"]*)"$`, w.taskInColumn)
	sc.Then(`^task (\d+) is titled "([^"]*)"$`, w.taskTitled)
	sc.Then(`^it asks which of "([^"]*)", "([^"]*)", "([^"]*)"$`, w.asksWhichOf)
	sc.Then(`^it asks nothing$`, w.asksNothing)
	sc.Then(`^the bar offers "([^"]*)"$`, w.barOffers)
	sc.Then(`^the bar does not offer "([^"]*)"$`, w.barDoesNotOffer)
	sc.Then(`^the help is shown$`, w.helpIsShown)
	sc.Then(`^the board carries (\d+) tasks?$`, w.boardCarries)
	sc.Then(`^the board says nothing of "([^"]*)"$`, w.boardSaysNothingOf)
	sc.Step(`^I added a task titled "([^"]*)" hanging under task (\d+)$`, func(title string, under int) error {
		_, err := w.other.Add(title, engine.Changes{Parent: &under})
		w.afterAChange()
		return err
	})
	sc.Then(`^the card for task (\d+) says "([^"]*)"$`, w.cardSays)
	sc.Then(`^the card for objective (\d+) says "([^"]*)"$`, w.goalCardSays)
	sc.Then(`^the card for task (\d+) is marked as described$`, w.cardDescribed)
	sc.Then(`^the card for task (\d+) is marked as archived$`, w.cardArchived)
	sc.Then(`^the card for task (\d+) is drawn as the one in focus$`, w.cardInFocus)
	sc.Then(`^the board says "([^"]*)"$`, w.boardSays)
	sc.Then(`^the board does not say "([^"]*)"$`, w.boardDoesNotSay)
}

func (w *world) aWindow(width, high int) error {
	w.width, w.high = width, high
	return nil
}

func say(s string) *string { return &s }

func (w *world) addTask(title, state, desc string, serving int) error {
	changes := engine.Changes{}
	if state != "" {
		changes.State = say(state)
	}
	if desc != "" {
		changes.Description = say(desc)
	}
	if serving != 0 {
		changes.Objective = &[]string{fmt.Sprint(serving)}
	}
	_, err := w.other.Add(title, changes)
	w.afterAChange()
	return err
}

func (w *world) addObjective(title string) error {
	_, err := w.other.AddObjective(title, engine.ObjectiveChanges{})
	return err
}

func (w *world) blockTask(number int, blocker string) error {
	_, err := w.other.Edit(number, engine.Changes{Blocker: &[]string{blocker}})
	w.afterAChange()
	return err
}

func (w *world) stateTask(number int, state string) error {
	_, err := w.other.Edit(number, engine.Changes{State: say(state)})
	w.afterAChange()
	return err
}

func (w *world) deleteTask(number int) error {
	err := w.other.Delete(number)
	w.afterAChange()
	return err
}

func (w *world) focusTask(number int) error {
	_, err := w.other.PutInFocus(number)
	w.afterAChange()
	return err
}

// afterAChange leaves the board alone. A change made while it is open is a
// change made by somebody else, and noticing it is the board's own business —
// which is what `the board looks again` is for.
func (w *world) afterAChange() {}

// pressMany takes a list of keys written the way a scenario writes one.
func (w *world) pressMany(list string) error {
	for _, name := range strings.Split(list, ",") {
		w.press(strings.Trim(strings.TrimSpace(name), `"`))
	}
	return nil
}

func (w *world) typeIn(text string) error {
	for _, r := range text {
		w.press(string(r))
	}
	return nil
}

func (w *world) choose(kind string) error {
	for range len(kinds) {
		if kinds[w.b.choice] == kind {
			w.press("enter")
			return nil
		}
		w.press("j")
	}
	return fmt.Errorf("%q was never offered", kind)
}

// setField walks to a field, clears it, and types what the scenario said. The
// form is Huh's, so moving on is a tab: enter would only add a line to the box
// the description is written in.
// labelTask puts two labels on a task, so a scenario can search for one.
func (w *world) labelTask(number int, one, two string) error {
	_, err := w.held.Ask(engine.TaskEdit, engine.Input{
		Given:  map[string]string{"number": strconv.Itoa(number)},
		Listed: map[string][]string{"label": {one, two}},
	})
	return err
}

func (w *world) setField(name, value string) error {
	at := -1
	for i, field := range w.b.form.fields {
		if field.Name == name {
			at = i
		}
	}
	if at < 0 {
		return fmt.Errorf("no field called %q", name)
	}
	for range at - w.formAt {
		w.press("tab")
	}
	w.formAt = at
	for range len([]rune(w.b.form.valueOf(name))) {
		if w.b.form.valueOf(name) == "" {
			break
		}
		w.press("backspace")
	}
	for _, r := range value {
		if r == ' ' {
			w.press("space")
			continue
		}
		w.press(string(r))
	}
	return nil
}

// saveForm walks whatever is left of the form and keeps what is in it. The last
// thing Huh asks is whether to save, and it is answered yes by default — so
// reaching it and pressing enter is what saving is.
func (w *world) saveForm() error {
	for range len(w.b.form.fields) - w.formAt {
		w.press("tab")
	}
	// Twice: the first answers the last question, and the form asks for one
	// more thing before it says it is done.
	w.press("enter")
	w.press("enter")
	return nil
}

func (w *world) lookAgain() error {
	w.refresh()
	return nil
}

// A form refused keeps what was typed. It closed before the write was tried,
// so a title left empty or a date mistyped put the refusal in the footer and
// the whole form in the bin — the window has always done the opposite.
func TestARefusedFormKeepsWhatWasTyped(t *testing.T) {
	w := newWorld(t)
	w.reset()
	if err := w.addTask("Write the spec", "", "", 0); err != nil {
		t.Fatal(err)
	}
	w.open()
	w.press("i")
	w.press("e")
	if w.b.screen != OnForm {
		t.Fatalf("the form did not open: %v", w.b.screen)
	}
	if err := w.setField("title", ""); err != nil {
		t.Fatal(err)
	}
	if err := w.saveForm(); err != nil {
		t.Fatal(err)
	}
	if w.b.screen != OnForm {
		t.Errorf("a refused form left the board on %v, losing what was typed", w.b.screen)
	}
	if w.b.err == nil {
		t.Error("a refused form said nothing about why")
	}
}

// The help gives back whatever it was asked over. Leaving it put the board
// back on the board, so asking for help from a form threw the form away —
// which docs/tui.md promises it does not.
func TestTheHelpGivesBackWhatItWasAskedOver(t *testing.T) {
	w := newWorld(t)
	w.reset()
	if err := w.addTask("Write the spec", "", "", 0); err != nil {
		t.Fatal(err)
	}
	w.open()
	w.press("i")
	w.press("e")
	if w.b.screen != OnForm {
		t.Fatalf("the form did not open: %v", w.b.screen)
	}
	w.press("?")
	w.press("esc")
	if w.b.screen != OnForm {
		t.Errorf("leaving the help by esc put the board on %v", w.b.screen)
	}
}

// ctrl+c closes the board. :q was the only way out, and nothing said so —
// whoever had not read the documentation was shut inside the alt-screen.
func TestCtrlCClosesTheBoard(t *testing.T) {
	w := newWorld(t)
	w.reset()
	w.open()
	_, cmd := w.b.Update(keyPress("ctrl+c"))
	if cmd == nil {
		t.Fatal("ctrl+c asked for nothing")
	}
	if _, quits := cmd().(tea.QuitMsg); !quits {
		t.Errorf("ctrl+c asked for %T, not to quit", cmd())
	}
}

// Deleting is asked about first. The command line wants --yes, the window puts
// a question, the agent is told the act destroys — the terminal alone did it on
// one key, and nothing anywhere can undo it.
func TestDeletingIsAskedAboutFirst(t *testing.T) {
	w := newWorld(t)
	w.reset()
	if err := w.addTask("Write the spec", "", "", 0); err != nil {
		t.Fatal(err)
	}
	w.open()
	w.press("i")
	w.press("d")
	if !w.b.Asking() {
		t.Fatal("d deleted without asking")
	}
	w.press("esc")
	if w.b.Asking() {
		t.Error("esc left the question open")
	}
	if w.b.Where().Showing == 0 {
		t.Error("the task went even though the question was backed out of")
	}
	w.press("d")
	w.press("enter")
	if err := w.lookAgain(); err != nil {
		t.Fatal(err)
	}
	if w.b.Where().Showing != 0 {
		t.Error("agreeing to the question did not delete the task")
	}
}

// A search that was accepted can be cleared. esc cleared it only while it was
// being typed; once accepted the board stayed narrowed, the mode still read
// NORMAL, and nothing said a filter was on.
func TestAnAcceptedSearchIsCleared(t *testing.T) {
	w := newWorld(t)
	w.reset()
	for _, title := range []string{"Write the spec", "Ship it"} {
		if err := w.addTask(title, "", "", 0); err != nil {
			t.Fatal(err)
		}
	}
	w.open()
	w.press("/")
	for _, r := range "spec" {
		w.press(string(r))
	}
	w.press("enter")
	if w.b.query != "spec" {
		t.Fatalf("the board is narrowed to %q", w.b.query)
	}
	w.press("esc")
	if w.b.query != "" {
		t.Errorf("esc left the board narrowed to %q", w.b.query)
	}
}

// The board guards a deletion of two words and threw ten minutes of typing away
// on an esc. A form somebody typed into is asked about; one they only looked at
// is not.
func TestLeavingAFilledFormAsksFirst(t *testing.T) {
	w := newWorld(t)
	w.reset()
	if err := w.addTask("Write the spec", "", "", 0); err != nil {
		t.Fatal(err)
	}
	w.open()
	w.press("i")
	w.press("e")
	if w.b.screen != OnForm {
		t.Fatalf("the form did not open: %v", w.b.screen)
	}
	w.b.form = w.b.form.typedInto("title", "Write the spec, differently")
	w.press("esc")
	if !w.b.Asking() {
		t.Fatal("a filled form was thrown away without a question")
	}
	w.press("esc")
	if w.b.screen != OnForm {
		t.Fatalf("backing out of the question left the board on %v", w.b.screen)
	}
	w.press("esc")
	w.press("enter")
	if w.b.screen == OnForm {
		t.Fatal("agreeing to lose what was typed left the form open")
	}
}

// Archiving twelve finished tasks was one gesture on the command line and
// twenty-four on the board. The column the cursor is on goes in one.
func TestTheWholeColumnCanBeArchivedAtOnce(t *testing.T) {
	w := newWorld(t)
	w.reset()
	for _, title := range []string{"One", "Two"} {
		if err := w.addTask(title, "done", "", 0); err != nil {
			t.Fatal(err)
		}
	}
	w.open()
	for i := 0; i < 3; i++ {
		w.press("l")
	}
	w.press("shift+x")
	if !w.b.Asking() {
		t.Fatal("a whole column went without a question")
	}
	w.press("enter")
	if err := w.lookAgain(); err != nil {
		t.Fatal(err)
	}
	if w.b.Where().Showing != 0 {
		t.Fatalf("%d tasks are still on the board", w.b.Where().Showing)
	}
}

// Saving hands the window back at once, and a refusal brings the form back
// with everything still in it. Keeping the window while the write was in
// flight made the write one of the form's own commands, and the form's
// commands are raced against a deadline — so under load the write was dropped
// and the board answered from a reading that never had the task in it.
func TestSavingHandsTheWindowBackAndARefusalGivesItBack(t *testing.T) {
	w := newWorld(t)
	w.reset()
	if err := w.addTask("Write the spec", "", "", 0); err != nil {
		t.Fatal(err)
	}
	w.open()
	w.press("i")
	w.press("e")
	w.b.form = w.b.form.typedInto("title", "Something else")
	model, _ := w.b.save()
	w.b = model.(Board)
	if w.b.screen != OnTheBoard {
		t.Fatalf("saving left the window on %v", w.b.screen)
	}
	if !w.b.form.Changed() {
		t.Fatal("what was typed went with the window")
	}
	after := w.b.took(Looked{Err: errNope})
	if after.screen != OnForm {
		t.Fatalf("a refused write left the board on %v", after.screen)
	}
	if after.form.valueOf("title") != "Something else" {
		t.Fatalf("the form came back holding %q", after.form.valueOf("title"))
	}
}

// The parts of a task are edited in the form of the task they belong to, here
// as in the window: they are off the board, so a task carrying them could
// neither be changed nor deleted once they were.
func TestThePartsOfATaskAreEditedInItsForm(t *testing.T) {
	w := newWorld(t)
	w.reset()
	if err := w.addTask("Ship it", "", "", 0); err != nil {
		t.Fatal(err)
	}
	under := 1
	if _, err := w.other.Add("Relire l'ADR", engine.Changes{Parent: &under}); err != nil {
		t.Fatal(err)
	}
	w.open()
	w.press("i")
	w.press("e")
	if said := w.b.form.valueOf("parts"); said != "#2 [ ] Relire l'ADR" {
		t.Fatalf("the form opened holding %q", said)
	}
	w.b.form = w.b.form.typedInto("parts", "#2 [x] Relire l'ADR 0014\n[ ] Rédiger la section Flux")
	model, cmd := w.b.save()
	w.b = model.(Board)
	if cmd != nil {
		w.settle(cmd)
	}
	if err := w.lookAgain(); err != nil {
		t.Fatal(err)
	}
	parts := reading.Hanging(w.b.snap.Cards, 1)
	if len(parts) != 2 {
		t.Fatalf("the task carries %d parts, want 2", len(parts))
	}
	// By what they are called and not by the order they come back in: what a
	// part is called and whether it is done is what the field says, and where
	// it sits among its sisters is the board's own ordering.
	states := map[string]string{}
	for _, part := range parts {
		states[part.Title] = part.State
	}
	if states["Relire l'ADR 0014"] != "done" {
		t.Errorf("the part that was ticked is %q", states["Relire l'ADR 0014"])
	}
	if states["Rédiger la section Flux"] != "todo" {
		t.Errorf("the part that was named is %q", states["Rédiger la section Flux"])
	}
}
