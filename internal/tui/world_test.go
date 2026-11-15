package tui

import (
	"cairn/internal/board"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"charm.land/lipgloss/v2"

	tea "charm.land/bubbletea/v2"
	"github.com/cucumber/godog"

	"cairn/internal/engine"
	"cairn/internal/engine/task"
	"cairn/test"
)

var stood = time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC)

type world struct {
	t      *testing.T
	path   string
	held   *engine.Engine
	other  *engine.Engine
	b      Board
	pages  engine.Pages
	width  int
	high   int
	asked  bool
	opened bool
	left   bool
	spaces *test.Spaces
	recurs bool
}

func newWorld(t *testing.T) *world {
	return &world{t: t, width: 100, high: 44, spaces: &test.Spaces{}}
}

func (w *world) reset() {
	if w.held != nil {
		_ = w.held.Close()
	}
	if w.other != nil {
		_ = w.other.Close()
	}
	w.path = filepath.Join(w.t.TempDir(), "cairn.db")
	held, err := engine.OpenAt(engine.Places{Store: w.path, Notes: notesRoot})
	if err != nil {
		w.t.Fatalf("opening the store: %v", err)
	}
	w.held = held
	other, err := engine.Open(w.path)
	if err != nil {
		w.t.Fatalf("opening the store a second time: %v", err)
	}
	w.other = other
	w.b = Board{}
	w.asked = false
	w.left = false
	*w.spaces = test.Spaces{Made: w.t.TempDir, Writing: w.storeAt, Now: func() time.Time { return stood }}
	w.pages = engine.Pages{}
	w.recurs = false
}

func (w *world) shownPages() engine.Pages {
	if w.recurs {
		return w.pages
	}
	return w.pages.WithoutRecurrence()
}

func (w *world) open() {
	w.b = New(w.held, GlyphsFor(Unicode), func() time.Time { return stood }, w.width, w.high).Paged(w.shownPages())
	if w.spaces.Dir != "" {
		worn, _ := engine.Wearing(w.homeOf(""))
		w.b = w.b.Wearing(worn)
	}
	w.opened = true
	w.refresh()
	if w.spaces.Dir != "" {
		w.settle(w.b.readingConfig())
	}
	if w.spaces.Dir != "" {
		held := w.held
		w.b = w.b.Switching(func(to engine.Elsewhere) (engine.Pages, error) { return engine.Switch(held, to) })
		w.settle(w.b.placing())
	}
}

func (w *world) refresh() {
	w.settle(w.b.look())
}

func (w *world) settle(cmd tea.Cmd) {
	for depth := 0; cmd != nil && depth < 8; depth++ {
		msg := cmd()
		if msg == nil {
			return
		}
		w.noticeQuit(msg)
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
				w.noticeQuit(inner)
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

func (w *world) noticeQuit(msg tea.Msg) {
	if _, quits := msg.(tea.QuitMsg); quits {
		w.left = true
	}
}

func (w *world) press(name string) {
	was := w.b.screen
	model, cmd := w.b.Update(keyPress(name))
	w.b = model.(Board)
	if w.b.screen == board.OnForm {
		w.b = pump(w.b, cmd)
	} else {
		w.settle(cmd)
	}
	if was != board.OnForm && w.b.screen == board.OnForm {
		w.b = settleForm(w.b)
	}
}

func (w *world) screen() string { return strip(w.b.Draw()) }

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

func (w *world) register(sc *godog.ScenarioContext) {
	sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
		w.reset()
		w.opened = false
		return ctx, nil
	})

	sc.StepContext().After(func(ctx context.Context, _ *godog.Step, _ godog.StepResultStatus, err error) (context.Context, error) {
		if err != nil || !w.opened {
			return ctx, err
		}
		return ctx, drawsWithoutFault(w.b)
	})

	w.givens().Register(sc)
	w.registerWorkspaces(sc)
	w.registerColours(sc)
	w.registerTree(sc)
	w.registerConfiguring(sc)
	w.registerSeries(sc)
	sc.Given(`^a window of (\d+) by (\d+)$`, w.aWindow)

	sc.When(`^I open the board$`, func() error {
		if w.spaces.Dir != "" {
			if err := w.holdThroughHome(); err != nil {
				return err
			}
		}
		w.open()
		return nil
	})
	sc.Then(`^no screen runs past the terminal$`, w.noScreenRunsPastTheTerminal)
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
	sc.Given(`^a store from an older cairn, holding a task titled "([^"]*)"$`, w.aStoreFromAnOlderCairn)
	sc.Then(`^a backup was set aside$`, func() error {
		if n, err := w.backupsKept(); err != nil || n == 0 {
			return fmt.Errorf("%d backups were set aside, want one: %w", n, err)
		}
		return nil
	})
	sc.Then(`^no change is waiting$`, func() error {
		if n, err := w.changesWaiting(); err != nil || n != 0 {
			return fmt.Errorf("%d changes are waiting, want none: %w", n, err)
		}
		return nil
	})
	sc.Step(`^somebody else adds the step "([^"]*)" to task (\d+)$`, func(title string, n int) error {
		return w.givens().IAddedTheStepToTask(title, n, false)
	})

	sc.Then(`^the columns read (".*")$`, w.columnsRead)
	sc.Then(`^the band reads (".*")$`, w.bandReads)
	sc.Then(`^"([^"]*)" carries "([^"]*)"$`, w.columnCarries)
	sc.Then(`^"([^"]*)" reads (".*")$`, w.columnReads)
	sc.Then(`^it says it needs (\d+) columns$`, w.saysItNeeds)
	sc.Then(`^it says the window is (\d+)$`, w.saysTheWindowIs)
	sc.Then(`^the mode is "([^"]*)"$`, w.modeIs)
	sc.Then(`^the tab is "([^"]*)"$`, w.tabIs)
	sc.Then(`^the cursor is on the task "([^"]*)"$`, w.cursorOnTask)
	sc.Then(`^task (\d+) is in the column "([^"]*)"$`, w.taskInColumn)
	sc.Then(`^task (\d+) is titled "([^"]*)"$`, w.taskTitled)
	sc.Then(`^the bar offers "([^"]*)"$`, w.barOffers)
	sc.Then(`^the bar does not offer "([^"]*)"$`, w.barDoesNotOffer)
	sc.Then(`^the help is shown$`, w.helpIsShown)
	sc.Then(`^the board carries (\d+) tasks?$`, w.boardCarries)
	sc.Then(`^the board says nothing of "([^"]*)"$`, w.boardSaysNothingOf)
	sc.Then(`^the card for task (\d+) says "([^"]*)"$`, w.cardSays)
	sc.Then(`^the card for task (\d+) is drawn as the one in focus$`, w.cardInFocus)
	sc.Then(`^the board says "([^"]*)"$`, w.boardSays)
	sc.Then(`^the board says "([^"]*)" before "([^"]*)"$`, w.boardSaysBefore)
	sc.Given(`^the weekends are hidden$`, func() error {
		w.pages = w.pages.WithoutWeekends()
		return nil
	})
	sc.Given(`^the pages turned off are "([^"]*)"$`, func(list string) error {
		w.pages = engine.PagesOff(strings.Split(list, ", ")...)
		return nil
	})
	sc.Then(`^the screen holds nothing$`, func() error {
		if said := strings.TrimSpace(w.b.Draw()); said != "" {
			return fmt.Errorf("the screen holds:\n%s", said)
		}
		return nil
	})
	sc.Then(`^the board does not say "([^"]*)"$`, w.boardDoesNotSay)
}

func (w *world) aWindow(width, high int) error {
	w.width, w.high = width, high
	return nil
}

func (w *world) givens() test.Givens {
	return test.Givens{Other: func() test.Asker { return w.other }}
}

func (w *world) aStoreFromAnOlderCairn(title string) error {
	if err := w.addTask(title, "", "", 0); err != nil {
		return err
	}
	if _, err := w.held.Ask(engine.MigrationRollback, engine.Agreed(engine.Input{})); err != nil {
		return err
	}
	if err := w.forgetBackups(); err != nil {
		return err
	}
	return w.reopen()
}

func (w *world) forgetBackups() error {
	return os.RemoveAll(filepath.Join(filepath.Dir(w.path), "backups"))
}

func (w *world) reopen() error {
	_ = w.held.Close()
	_ = w.other.Close()
	held, err := engine.OpenAt(engine.Places{Store: w.path, Notes: notesRoot})
	if err != nil {
		return err
	}
	other, err := engine.OpenBehind(w.path)
	if err != nil {
		return err
	}
	w.held, w.other = held, other
	return nil
}

func (w *world) backupsKept() (int, error) {
	kept, err := w.held.Ask(engine.BackupList, engine.Input{})
	return len(kept.Backups), err
}

func (w *world) changesWaiting() (int, error) {
	listed, err := w.held.Ask(engine.MigrationList, engine.Input{})
	waiting := 0
	for _, c := range listed.Changes {
		if !c.Carried().Applied {
			waiting++
		}
	}
	return waiting, err
}

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
	for range len(board.Stalled) {
		if board.Stalled[w.b.choice] == kind {
			w.press("enter")
			return nil
		}
		w.press("j")
	}
	return fmt.Errorf("%q was never offered", kind)
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
	for from := w.fieldAt(); from < at; from++ {
		w.press("tab")
	}
	for from := w.fieldAt(); from > at; from-- {
		w.press("shift+tab")
	}
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

func (w *world) fieldAt() int {
	focused := w.b.form.asked.GetFocusedField().GetKey()
	for i, field := range w.b.form.fields {
		if field.Name == focused {
			return i
		}
	}
	return len(w.b.form.fields)
}

func (w *world) saveForm() error {
	for range len(w.b.form.fields) - w.fieldAt() {
		w.press("tab")
	}
	w.press("enter")
	w.press("enter")
	return nil
}

func (w *world) lookAgain() error {
	w.refresh()
	return nil
}

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

func TestDeletingIsAskedAboutFirst(t *testing.T) {
	w := newWorld(t)
	w.reset()
	if err := w.addTask("Write the spec", "", "", 0); err != nil {
		t.Fatal(err)
	}
	w.open()
	w.press("i")
	w.press("D")
	if !w.b.Asking() {
		t.Fatal("D deleted without asking")
	}
	w.press("esc")
	if w.b.Asking() {
		t.Error("esc left the question open")
	}
	if w.b.Where().Showing == 0 {
		t.Error("the task went even though the question was backed out of")
	}
	w.press("D")
	w.press("enter")
	if err := w.lookAgain(); err != nil {
		t.Fatal(err)
	}
	if w.b.Where().Showing != 0 {
		t.Error("agreeing to the question did not delete the task")
	}
}

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

func TestLeavingAFilledFormAsksFirst(t *testing.T) {
	w := newWorld(t)
	w.reset()
	if err := w.addTask("Write the spec", "", "", 0); err != nil {
		t.Fatal(err)
	}
	w.open()
	w.press("i")
	w.press("e")
	if w.b.screen != board.OnForm {
		t.Fatalf("the form did not open: %v", w.b.screen)
	}
	w.b.form = w.b.form.typedInto("title", "Write the spec, differently")
	w.press("esc")
	if !w.b.Asking() {
		t.Fatal("a filled form was thrown away without a question")
	}
	w.press("esc")
	if w.b.screen != board.OnForm {
		t.Fatalf("backing out of the question left the board on %v", w.b.screen)
	}
	w.press("esc")
	w.press("enter")
	if w.b.screen == board.OnForm {
		t.Fatal("agreeing to lose what was typed left the form open")
	}
}

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

func TestTheBoardsFormFollowsTheSameRuleAsTheWindows(t *testing.T) {
	w := newWorld(t)
	w.reset()
	if err := w.addTask("Write the spec", "", "", 0); err != nil {
		t.Fatal(err)
	}
	w.open()
	w.press("i")
	w.press("e")
	w.press("tab")
	for _, r := range "one" {
		w.press(string(r))
	}
	w.press("enter")
	for _, r := range "two" {
		w.press(string(r))
	}
	if said := w.b.form.valueOf("description"); said != "one\ntwo" {
		t.Fatalf("the description reads %q", said)
	}
	if w.b.screen != board.OnForm {
		t.Fatalf("enter in the description left the form on %v", w.b.screen)
	}
	w.press("up")
	w.press("x")
	if said := w.b.form.valueOf("description"); !strings.Contains(said, "x") || !strings.Contains(said, "\n") {
		t.Fatalf("after up and a letter the description reads %q", said)
	}
	w.press("ctrl+s")
	if err := w.lookAgain(); err != nil {
		t.Fatal(err)
	}
	if w.b.screen == board.OnForm {
		t.Fatal("^S did not save the form")
	}
	found, err := w.givens().Task(1)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(task.OfTask(found).Description, "two") {
		t.Fatalf("what was saved reads %q", task.OfTask(found).Description)
	}
}

func TestTheDetailsDescriptionScrollsToItsEnd(t *testing.T) {
	w := newWorld(t)
	w.reset()
	text := strings.Repeat("a sentence of description long enough to wrap ", 40) + "THE END"
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
	if strings.Contains(drawn, "THE END") {
		t.Fatal("this proves nothing: the whole description fits")
	}
	for range 60 {
		w.press("ctrl+d")
	}
	if !strings.Contains(w.b.Draw(), "THE END") {
		t.Fatal("^D to the end does not bring the last line into sight")
	}
	for range 60 {
		w.press("ctrl+u")
	}
	if !strings.Contains(w.b.Draw(), "a sentence") {
		t.Fatal("^U back up does not bring the beginning into sight")
	}
}

const notesRoot = "/vault"

func drawsWithoutFault(b Board) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("drawing the board panicked: %v", r)
		}
	}()
	_ = b.View()
	return nil
}

func (w *world) everyScreen(look func(name string)) {
	look("the board")
	for _, step := range []struct{ key, name string }{
		{"enter", "a task's detail"}, {"esc", "the board again"}, {"e", "a task's form"}, {"esc", "the board again"},
		{"?", "the help"}, {"?", "the board again"},
		{"tab", "the objectives"}, {"enter", "an objective's detail"}, {"esc", "the objectives again"},
		{"tab", "the time"},
	} {
		w.press(step.key)
		look(step.name)
	}
	if !w.recurs {
		return
	}
	for _, step := range []struct{ key, name string }{
		{"tab", "the series"}, {"enter", "a series' occurrences"}, {"esc", "the series again"}, {"e", "a series' form"}, {"esc", "the series again"},
	} {
		w.press(step.key)
		look(step.name)
	}
}

func (w *world) noScreenRunsPastTheTerminal() error {
	var problems []string
	look := func(name string) {
		lines := strings.Split(strings.TrimRight(w.screen(), "\n"), "\n")
		if len(lines) > w.high {
			problems = append(problems, fmt.Sprintf("%s: %d lines, the terminal holds %d", name, len(lines), w.high))
		}
		for i, line := range lines {
			if wide := lipgloss.Width(line); wide > w.width {
				problems = append(problems, fmt.Sprintf("%s: line %d is %d wide, the terminal %d: %q", name, i+1, wide, w.width, line))
			}
		}
	}
	if strings.Contains(w.screen(), "needs a bigger window") {
		return fmt.Errorf("the board is not drawn at %d by %d, so nothing was checked", w.width, w.high)
	}
	w.everyScreen(look)
	if len(problems) > 0 {
		return fmt.Errorf("%d problem(s):\n%s", len(problems), strings.Join(problems, "\n"))
	}
	return nil
}
