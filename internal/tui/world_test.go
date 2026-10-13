package tui

// The world the board's scenarios live in. It reaches the engine for the
// givens — a scenario says what is in the store, not how it got there — and
// drives the board itself for everything else: the board is a value, so a key
// press is a call and the screen is a string, with no terminal anywhere.

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/cucumber/godog"

	"cairn/internal/engine"
	"cairn/internal/task"
	"cairn/test"
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
	// opened says a board stands, so a step after it can draw it.
	opened bool
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
	// The notes are named, as a surface finds them, and never made: cairn
	// only says where they are, so an absolute place nobody writes to is safe.
	held, err := engine.OpenAt(engine.Places{Store: w.path, Notes: notesRoot})
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

// open builds the board and gives it its first engine.
//
// It hands the reading over rather than running Init, and for a measured
// reason: Init asks for the beat as well, and running the beat's command
// sleeps for as long as the beat lasts. Thirty-eight scenarios each waiting two
// seconds was the whole of this suite's running time, and a suite that slow
// makes every mutant time out — which reads as a gate that passed while judging
// nothing. What the beat does is checked on its own, in model_test.go.
func (w *world) open() {
	w.b = New(w.held, GlyphsFor(Unicode), func() time.Time { return stood }, w.width, w.high)
	w.opened = true
	w.refresh()
}

// refresh hands the board a fresh reading, the way its own beat would once it
// noticed the store had moved.
func (w *world) refresh() {
	found, err := engine.Look(w.held)
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
		w.opened = false
		return ctx, nil
	})

	// Every step leaves a board that can be drawn — the window's scenarios
	// learnt it from a list opened over a form that panicked in somebody's
	// hands while every step leading there passed. Drawn after each step, a
	// state that cannot be drawn fails whichever scenario reached it.
	sc.StepContext().After(func(ctx context.Context, _ *godog.Step, _ godog.StepResultStatus, err error) (context.Context, error) {
		if err != nil || !w.opened {
			return ctx, err
		}
		return ctx, drawsWithoutFault(w.b)
	})

	w.givens().Register(sc)
	sc.Given(`^a window of (\d+) by (\d+)$`, w.aWindow)

	sc.When(`^I open the board$`, func() error { w.open(); return nil })
	sc.Step(`^I press (".*")$`, w.pressMany)
	sc.Step(`^I press "([^"]*)" (\d+) times$`, func(name string, times int) error {
		for range times {
			w.press(name)
		}
		return nil
	})
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
	sc.Then(`^the card for task (\d+) says "([^"]*)"$`, w.cardSays)
	sc.Then(`^the card for objective (\d+) says "([^"]*)"$`, w.goalCardSays)
	sc.Then(`^the card for task (\d+) is marked as described$`, w.cardDescribed)
	sc.Then(`^the card for task (\d+) is marked as archived$`, w.cardArchived)
	sc.Then(`^the card for task (\d+) is drawn as the one in focus$`, w.cardInFocus)
	sc.Then(`^the board says "([^"]*)"$`, w.boardSays)
	sc.Then(`^the open task says where the context of task (\d+) lives$`, func(n int) error {
		if err := w.mustSay("CONTEXT"); err != nil {
			return err
		}
		return w.mustSay(notesRoot + "/" + strconv.Itoa(n))
	})
	sc.Then(`^the board does not say "([^"]*)"$`, w.boardDoesNotSay)
}

func (w *world) aWindow(width, high int) error {
	w.width, w.high = width, high
	return nil
}

// givens writes what a scenario says is in the store, as somebody else.
func (w *world) givens() test.Givens {
	return test.Givens{Other: func() test.Asker { return w.other }}
}

// addTask adds a task carrying what a test named about it, as somebody else.
func (w *world) addTask(title, state, desc string, serving int) error {
	given := map[string]string{"title": title}
	if state != "" {
		given["state"] = state
	}
	if desc != "" {
		given["desc"] = desc
	}
	in := engine.Input{Given: given}
	if serving != 0 {
		in.Listed = map[string][]string{"objective": {fmt.Sprint(serving)}}
	}
	_, err := w.other.Ask(engine.TaskNew, in)
	return err
}

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

// What went wrong took the bar's place and kept it until the next good read,
// so every offer was gone while it stood. It has been read by the time a key
// is pressed: that key gives the bar back, and still does what it does.
func TestAKeyGivesTheBarBackOverWhatWentWrong(t *testing.T) {
	w := newWorld(t)
	w.reset()
	if err := w.addTask("Write the spec", "", "", 0); err != nil {
		t.Fatal(err)
	}
	w.open()
	w.b.err = errors.New("that cannot be done")
	if !strings.Contains(w.b.foot(), "that cannot be done") {
		t.Fatalf("the foot does not say what went wrong: %q", w.b.foot())
	}
	w.press("l")
	if w.b.err != nil || strings.Contains(w.b.foot(), "that cannot be done") {
		t.Errorf("a key left what went wrong in the bar's place: %q", w.b.foot())
	}
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
	if err := w.givens().IAddedATaskHangingUnderTask("Relire l'ADR", 1); err != nil {
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
	parts := engine.Hanging(w.b.snap.Cards, 1)
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

// One rule for a form on both boards: enter writes a line in the description
// and moves on everywhere else, the arrows walk the description's lines, tab
// leaves it, and ^S saves from anywhere — enter never does.
func TestTheBoardsFormFollowsTheSameRuleAsTheWindows(t *testing.T) {
	w := newWorld(t)
	w.reset()
	if err := w.addTask("Write the spec", "", "", 0); err != nil {
		t.Fatal(err)
	}
	w.open()
	w.press("i")
	w.press("e")
	w.press("tab") // onto the description
	for _, r := range "un" {
		w.press(string(r))
	}
	w.press("enter")
	for _, r := range "deux" {
		w.press(string(r))
	}
	if said := w.b.form.valueOf("description"); said != "un\ndeux" {
		t.Fatalf("the description reads %q", said)
	}
	if w.b.screen != OnForm {
		t.Fatalf("enter in the description left the form on %v", w.b.screen)
	}
	// Up stays inside the description, on its first line.
	w.press("up")
	w.press("x")
	if said := w.b.form.valueOf("description"); !strings.Contains(said, "x") || !strings.Contains(said, "\n") {
		t.Fatalf("after up and a letter the description reads %q", said)
	}
	w.press("ctrl+s")
	if err := w.lookAgain(); err != nil {
		t.Fatal(err)
	}
	if w.b.screen == OnForm {
		t.Fatal("^S did not save the form")
	}
	found, err := w.givens().Task(1)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(task.OfTask(found).Description, "deux") {
		t.Fatalf("what was saved reads %q", task.OfTask(found).Description)
	}
}

// What was typed into the form is what is saved, every field of it. Only the
// title was: the rest was read back from what the form had opened with, so a
// label, a date or a description changed on the board was dropped on save
// without a word — since the board first had a form.
func TestTheFormSavesEveryFieldThatWasTyped(t *testing.T) {
	w := newWorld(t)
	w.reset()
	if err := w.addTask("Write the spec", "", "", 0); err != nil {
		t.Fatal(err)
	}
	w.open()
	w.press("i")
	w.press("e")
	for field, value := range map[string]string{
		"labels":      "backend",
		"deadline":    "2026-10-09",
		"description": "deux colonnes",
	} {
		w.b.form = w.b.form.typedInto(field, value)
	}
	model, cmd := w.b.save()
	w.b = model.(Board)
	w.settle(cmd)
	found, err := w.givens().Task(1)
	if err != nil {
		t.Fatal(err)
	}
	told := task.OfTask(found)
	if len(told.Labels) != 1 || told.Labels[0] != "backend" {
		t.Errorf("the labels saved are %v", told.Labels)
	}
	if told.Deadline != "2026-10-09" {
		t.Errorf("the deadline saved is %q", told.Deadline)
	}
	if told.Description != "deux colonnes" {
		t.Errorf("the description saved is %q", told.Description)
	}
}

// The description on an open task is a box that scrolls, with its bar, the way
// the window's is — it used to be drawn whole, and a long one ran off the
// bottom of the terminal with everything under it. ^D and ^U walk it, half a
// box at a time, and the last line can be reached.
func TestTheDetailsDescriptionScrollsToItsEnd(t *testing.T) {
	w := newWorld(t)
	w.reset()
	text := strings.Repeat("une phrase de description assez longue pour replier ", 40) + "LA FIN"
	if err := w.addTask("Write the spec", "", text, 0); err != nil {
		t.Fatal(err)
	}
	w.open()
	w.press("i")
	w.press("enter")
	drawn := w.b.Draw()
	if got := strings.Count(drawn, "\n") + 1; got > w.high {
		t.Fatalf("the detail is %d lines tall in a terminal of %d", got, w.high)
	}
	if !strings.Contains(drawn, Track) && !strings.Contains(drawn, Thumb) {
		t.Fatal("the description carries no bar")
	}
	if strings.Contains(drawn, "LA FIN") {
		t.Fatal("this proves nothing: the whole description fits")
	}
	for range 60 {
		w.press("ctrl+d")
	}
	if !strings.Contains(w.b.Draw(), "LA FIN") {
		t.Fatal("^D to the end does not bring the last line into sight")
	}
	for range 60 {
		w.press("ctrl+u")
	}
	if !strings.Contains(w.b.Draw(), "une phrase") {
		t.Fatal("^U back up does not bring the beginning into sight")
	}
}

// The description on an open task, to the line: how many rows its box holds in
// a terminal of a known height, how far ^D and ^U move it, and where both ends
// stop. "The end can be reached" says nothing about a box one line too short.
func TestTheDetailsDescriptionToTheLine(t *testing.T) {
	w := newWorld(t)
	w.reset()
	lines := make([]string, 80)
	for i := range lines {
		lines[i] = fmt.Sprintf("line %d", i+1)
	}
	if err := w.addTask("Write the spec", "", strings.Join(lines, "\n"), 0); err != nil {
		t.Fatal(err)
	}
	w.open()
	w.press("i")
	w.press("enter")
	held, _ := w.b.held()
	laid, rows := w.b.saidOn(held, w.b.detailWidth(), w.b.detailLeftUnder(held, w.b.detailWidth()))
	// Forty-four rows, less the seven around the detail, the three of the
	// panel's own frame and name, and the four of the one under it saying
	// where the task's context lives: its frame, its name and the place.
	if rows != 30 {
		t.Fatalf("the description's box holds %d rows in a terminal of %d, want 30", rows, w.high)
	}
	if len(laid) != 80 {
		t.Fatalf("eighty short lines laid out as %d", len(laid))
	}
	w.press("ctrl+d")
	if w.b.saying != 15 {
		t.Fatalf("^D moved the description to %d, want half its box at 15", w.b.saying)
	}
	w.press("ctrl+u")
	if w.b.saying != 0 {
		t.Fatalf("^U moved it back to %d, want 0", w.b.saying)
	}
	for range 10 {
		w.press("ctrl+d")
	}
	if w.b.saying != 50 {
		t.Fatalf("^D to the end stopped at %d, want the last box of it at 50", w.b.saying)
	}
	w.b.saying = 1000
	if got := w.b.scrolledSaying(held, laid, rows); got != 50 {
		t.Fatalf("a scroll past the end reads as %d, want 50", got)
	}
}

// The panels under the description take their rows out of its box: a task with
// parts shows fewer lines of its description than one without, by exactly the
// height of what is under it.
func TestThePanelsUnderTheDescriptionTakeTheirRows(t *testing.T) {
	w := newWorld(t)
	w.reset()
	if err := w.addTask("Write the spec", "", "une description", 0); err != nil {
		t.Fatal(err)
	}
	w.open()
	w.press("i")
	held, _ := w.b.held()
	_, alone := w.b.saidOn(held, w.b.detailWidth(), nil)
	under := []string{"one\ntwo\nthree\nfour"}
	_, shared := w.b.saidOn(held, w.b.detailWidth(), under)
	if alone-shared != 4 {
		t.Fatalf("four rows under the description took %d of its own", alone-shared)
	}
}

// notesRoot is where a scenario's board is told the notes live.
const notesRoot = "/vault"

// drawsWithoutFault renders a board and says whether rendering it panicked.
func drawsWithoutFault(b Board) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("drawing the board panicked: %v", r)
		}
	}()
	_ = b.View()
	return nil
}
