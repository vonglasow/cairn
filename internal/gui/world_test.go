package gui

// The world the window's scenarios live in. It reaches the engine for the
// givens — a scenario says what is in the store, not how it got there — and
// drives the window itself for everything else: the window is a value, so a
// gesture is a call and what it decided is a field, with no screen anywhere.

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cucumber/godog"

	"cairn/internal/engine"
	"cairn/internal/reading"
	"cairn/internal/task"
)

type world struct {
	t     *testing.T
	path  string
	held  *engine.Engine
	other *engine.Engine
	work  *refusing
	w     Window

	width, high int
	// scrolled is what each column stood at before the last gesture, so a
	// scenario can say which one moved.
	scrolled [Columns]int
	// was is where the history stood before a step moved it, so a step after
	// it can say whether it moved at all.
	was int
}

func newWorld(t *testing.T) *world { return &world{t: t, width: 1200, high: 800} }

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
	// window notices another connection's writes and cannot notice its own.
	other, err := engine.Open(w.path)
	if err != nil {
		w.t.Fatalf("opening the store a second time: %v", err)
	}
	w.other = other
	w.work = &refusing{held: held}
	w.w = Window{}
	w.width, w.high = 1200, 800
	w.scrolled = [Columns]int{}
}

// refusing is the store the window talks to. It is the real engine, with one
// thing a real engine cannot be asked for on demand: a refusal of everything
// that changes something.
type refusing struct {
	held   *engine.Engine
	refuse error
	reads  int
	// wrote is every operation that changes something, in order — which is
	// what a scenario asks about when it says the engine was asked, or was
	// not, or was asked once.
	wrote []wrote
}

// wrote is one thing the window told the store to change.
type wrote struct {
	thing, verb, number string
}

func (r *refusing) Ask(op engine.Operation, in engine.Input) (engine.Answer, error) {
	if op.Level != engine.Read {
		r.wrote = append(r.wrote, wrote{op.Thing, op.Verb, in.Given["number"]})
	}
	if op.Level != engine.Read && r.refuse != nil {
		return engine.Answer{}, r.refuse
	}
	if op.Thing == "task" && op.Verb == "list" {
		r.reads++
	}
	return r.held.Ask(op, in)
}

func (r *refusing) Version() (int64, error) { return r.held.Version() }

// ── what a scenario sets up ───────────────────────────────────────────────

func (w *world) aWindow(width, high int) error {
	w.width, w.high = width, high
	return nil
}

func (w *world) addTask(title, state, desc string) error {
	changes := engine.Changes{}
	if state != "" {
		changes.State = &state
	}
	if desc != "" {
		changes.Description = &desc
	}
	_, err := w.other.Add(title, changes)
	return err
}

func (w *world) addMany(n int, state string) error {
	for i := 1; i <= n; i++ {
		if err := w.addTask(fmt.Sprintf("Task %d", i), state, ""); err != nil {
			return err
		}
	}
	return nil
}

func (w *world) storeRefuses() error {
	w.work.refuse = fmt.Errorf("that cannot be done")
	return nil
}

// ── what a person does ────────────────────────────────────────────────────

func (w *world) open() error {
	w.w = New(w.work, w.width, w.high)
	w.w = w.w.Look()
	return w.w.Wrong()
}

func (w *world) resize(width, high int) error {
	w.w = w.w.Resized(width, high)
	return nil
}

// find is where on screen a card sits, by the title a scenario calls it.
func (w *world) find(title string) (Point, error) {
	l := w.w.Layout()
	dealt, stalled := reading.Deal(w.w.Reading().Cards)
	for column, held := range dealt {
		for at, c := range held.Cards {
			if c.Title != title {
				continue
			}
			box := l.CardAt(column, at, w.w.scroll[column])
			return Point{X: box.X + box.W/2, Y: box.Y + box.H/2}, nil
		}
	}
	for lane, kind := range Lanes {
		at := 0
		for _, c := range stalled {
			if c.State != kind {
				continue
			}
			if c.Title == title {
				box := l.LaneCardAt(lane, at, 0)
				return Point{X: box.X + box.W/2, Y: box.Y + box.H/2}, nil
			}
			at++
		}
	}
	return Point{}, fmt.Errorf("no card titled %q is on screen", title)
}

func (w *world) click(title string) error {
	at, err := w.find(title)
	if err != nil {
		return err
	}
	w.w = w.w.Click(at)
	return nil
}

func (w *world) doubleClick(title string) error {
	at, err := w.find(title)
	if err != nil {
		return err
	}
	w.w = w.w.DoubleClick(at)
	return nil
}

func (w *world) clickTab(name string) error {
	which := map[string]Tab{"TASKS": Tasks, "OBJECTIVES": Objectives}
	tab, known := which[name]
	if !known {
		return fmt.Errorf("no tab called %q", name)
	}
	box := TabAt(w.w.Layout(), int(tab))
	w.w = w.w.Click(Point{X: box.X + box.W/2, Y: box.Y + box.H/2})
	return nil
}

func (w *world) addObjective(title string) error {
	_, err := w.other.AddObjective(title, engine.ObjectiveChanges{})
	return err
}

// findGoal is where an objective sits in the list, by the title a scenario
// calls it.
func (w *world) findGoal(title string) (Point, error) {
	for i, g := range reading.Sorted(w.w.Reading().Goals) {
		if g.Title != title {
			continue
		}
		box := GoalAt(w.w.Layout(), i, 0)
		return Point{X: box.X + box.W/2, Y: box.Y + box.H/2}, nil
	}
	return Point{}, fmt.Errorf("no objective titled %q is on screen", title)
}

func (w *world) clickGoal(title string) error {
	at, err := w.findGoal(title)
	if err != nil {
		return err
	}
	w.w = w.w.Click(at)
	return nil
}

// clickOffer presses one of the bar's offers, which is how a person reaches
// anything here without knowing a shortcut.
func (w *world) clickOffer(name string) error {
	for i, offer := range w.w.Offers() {
		if offer != name {
			continue
		}
		box := OfferAt(w.w.Layout(), i)
		w.w = w.w.Click(Point{X: box.X + box.W/2, Y: box.Y + box.H/2})
		return w.w.Wrong()
	}
	return fmt.Errorf("the bar offers %v, and none of them is %q", w.w.Offers(), name)
}

func (w *world) press(key string) error {
	w.w = w.w.Key(key)
	return nil
}

func (w *world) nudge(title string, by int) error {
	at, err := w.find(title)
	if err != nil {
		return err
	}
	w.w = w.w.PressAt(at)
	w.w = w.w.MoveTo(Point{X: at.X + by, Y: at.Y})
	w.w = w.w.LetGo(Point{X: at.X + by, Y: at.Y})
	return nil
}

// slotIn is the point that drops a card into a column at a given place: the
// head of it, or the slot a named card occupies.
func (w *world) slotIn(column string, above string) (Point, error) {
	at := reading.IndexOf(strings.ToLower(column))
	if at < 0 {
		return Point{}, fmt.Errorf("no column called %q", column)
	}
	l := w.w.Layout()
	slot := 0
	if above != "" {
		dealt, _ := reading.Deal(w.w.Reading().Cards)
		found := false
		for i, c := range dealt[at].Cards {
			if c.Title == above {
				slot, found = i, true
			}
		}
		if !found {
			return Point{}, fmt.Errorf("no card titled %q in %q", above, column)
		}
	}
	box := l.CardAt(at, slot, w.w.scroll[at])
	return Point{X: box.X + box.W/2, Y: box.Y + box.H/2}, nil
}

func (w *world) dragOnto(title, column string) error { return w.drag(title, column, "") }

func (w *world) dragAbove(title, column, above string) error { return w.drag(title, column, above) }

func (w *world) drag(title, column, above string) error {
	from, err := w.find(title)
	if err != nil {
		return err
	}
	onto, err := w.slotIn(column, above)
	if err != nil {
		return err
	}
	w.w = w.w.PressAt(from)
	w.w = w.w.MoveTo(onto)
	w.w = w.w.LetGo(onto)
	return nil
}

func (w *world) dragToLane(title, kind string) error {
	from, err := w.find(title)
	if err != nil {
		return err
	}
	lane := -1
	for i, name := range Lanes {
		if name == kind {
			lane = i
		}
	}
	if lane < 0 {
		return fmt.Errorf("no lane called %q", kind)
	}
	box := w.w.Layout().Lane[lane]
	onto := Point{X: box.X + box.W/2, Y: box.Y + box.H/2}
	w.w = w.w.PressAt(from)
	w.w = w.w.MoveTo(onto)
	w.w = w.w.LetGo(onto)
	return nil
}

// takeHold picks a card up and moves it over a column without letting go, so a
// scenario can look at what is shown while a hand is still holding one.
// takeHoldUnder picks a card up and holds it just below another, which is the
// place behind that card: what the engine is told is `after` the card the
// pointer sits under, so the step says the same thing the gesture does.
func (w *world) takeHoldUnder(title, below string) error {
	from, err := w.find(title)
	if err != nil {
		return err
	}
	under, err := w.under(below)
	if err != nil {
		return err
	}
	w.w = w.w.PressAt(from)
	w.w = w.w.MoveTo(under)
	return nil
}

// under is the point just inside the slot below a card.
func (w *world) under(title string) (Point, error) {
	l := w.w.Layout()
	dealt, _ := reading.Deal(w.w.Reading().Cards)
	for column, held := range dealt {
		for at, c := range held.Cards {
			if c.Title != title {
				continue
			}
			box := l.CardAt(column, at+1, w.w.scroll[column])
			return Point{X: box.X + box.W/2, Y: box.Y + box.H/2}, nil
		}
	}
	return Point{}, fmt.Errorf("no card titled %q", title)
}

func (w *world) takeHold(title, column string) error {
	from, err := w.find(title)
	if err != nil {
		return err
	}
	onto, err := w.slotIn(column, "")
	if err != nil {
		return err
	}
	w.w = w.w.PressAt(from)
	w.w = w.w.MoveTo(onto)
	return nil
}

func (w *world) scrollOver(column string) error {
	at := reading.IndexOf(strings.ToLower(column))
	if at < 0 {
		return fmt.Errorf("no column called %q", column)
	}
	w.scrolled = w.w.scroll
	box := w.w.Layout().Column[at]
	w.w = w.w.Scroll(Point{X: box.X + box.W/2, Y: box.Y + box.H/2}, CardHeight)
	return nil
}

func (w *world) choose(name string) error {
	w.w = w.w.Choose(name)
	return w.w.Wrong()
}

func (w *world) keyboardOn(name string) error {
	if w.w.Screen() != OnForm {
		return fmt.Errorf("no form is open")
	}
	w.w = w.w.On(name)
	return nil
}

func (w *world) types(text string) error {
	w.w = w.w.Types(text)
	return nil
}

// clickField is how a pointer puts the keyboard on a field.
func (w *world) clickField(name string) error {
	box := Over(w.w.Layout().Bar.W+2*margin, w.w.Layout().Bar.Y+barHigh+margin)
	fields := w.w.Form().Fields()
	for i, field := range fields {
		if field.Name != name {
			continue
		}
		row, _ := FieldRow(box, FieldTop(box, fields, i))
		w.w = w.w.Click(Point{X: row.X + 10, Y: row.Y + row.H/2})
		return nil
	}
	return fmt.Errorf("the form has no field called %q", name)
}

// clickInto clicks a given number of characters into a field's value, which is
// how a pointer puts the caret in the middle of a word.
func (w *world) clickInto(characters int, name string) error {
	box := Over(w.w.Layout().Bar.W+2*margin, w.w.Layout().Bar.Y+barHigh+margin)
	fields := w.w.Form().Fields()
	for i, field := range fields {
		if field.Name != name {
			continue
		}
		_, value := FieldRow(box, FieldTop(box, fields, i))
		at := Point{X: value.X + padding + characters*perGlyph, Y: value.Y + value.H/2}
		w.w = w.w.Click(at)
		return nil
	}
	return fmt.Errorf("the form has no field called %q", name)
}

func (w *world) setField(name, value string) error {
	if w.w.Screen() != OnForm {
		return fmt.Errorf("no form is open")
	}
	w.w.form = w.w.form.Set(name, value)
	return nil
}

func (w *world) somebodyElseAdds(title string) error {
	_, err := w.other.Add(title, engine.Changes{})
	return err
}

func (w *world) looksAgain() error {
	w.w = w.w.Wonder()
	return nil
}

func (w *world) register(sc *godog.ScenarioContext) {
	sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
		w.reset()
		return ctx, nil
	})

	sc.Given(`^a store of my own, somewhere that is not the default place$`, func() error { return nil })
	sc.Given(`^nothing in it yet$`, func() error { return nil })
	sc.Given(`^a window of (\d+) by (\d+)$`, w.aWindow)
	sc.Step(`^I added a task titled "([^"]*)"$`, func(t string) error { return w.addTask(t, "", "") })
	sc.Step(`^I added a task titled "([^"]*)" with the state "([^"]*)"$`, func(t, s string) error {
		return w.addTask(t, s, "")
	})
	sc.Step(`^I added a task titled "([^"]*)" with the description "([^"]*)"$`, func(t, d string) error {
		return w.addTask(t, "", d)
	})
	sc.Step(`^I added (\d+) tasks to "([^"]*)"$`, w.addMany)
	sc.Step(`^I added a task titled "([^"]*)" hanging under task (\d+)$`, func(title string, under int) error {
		_, err := w.other.Add(title, engine.Changes{Parent: &under})
		return err
	})
	sc.Step(`^I put task (\d+) in focus$`, func(n int) error {
		_, err := w.other.PutInFocus(n)
		return err
	})
	sc.Step(`^the store refuses what it is asked to change$`, w.storeRefuses)

	sc.When(`^I open the window$`, w.open)
	sc.Step(`^I resize the window to (\d+) by (\d+)$`, w.resize)
	sc.Step(`^I click the card "([^"]*)"$`, w.click)
	sc.Step(`^I click the tab "([^"]*)"$`, w.clickTab)
	sc.Step(`^I click the objective "([^"]*)"$`, func(title string) error {
		if err := w.clickGoal(title); err != nil {
			return err
		}
		for i, g := range reading.Sorted(w.w.Showing().Goals) {
			if g.Number == w.w.Chosen() {
				w.was = i
			}
		}
		return nil
	})
	sc.Step(`^I click "([^"]*)"$`, w.clickOffer)
	sc.Step(`^I added an objective titled "([^"]*)"$`, w.addObjective)
	sc.Step(`^I double-click the card "([^"]*)"$`, w.doubleClick)
	sc.Step(`^I press "([^"]*)"$`, w.press)
	sc.Step(`^I press "([^"]*)" (\d+) times$`, func(key string, n int) error {
		for range n {
			if err := w.press(key); err != nil {
				return err
			}
		}
		return nil
	})
	sc.Step(`^I added (\d+) objectives$`, w.addManyObjectives)
	sc.Step(`^I scroll down over the objectives$`, w.scrollObjectives)
	sc.Step(`^I press the card "([^"]*)" and move it (\d+) pixels and let go$`, w.nudge)
	sc.Step(`^I take hold of the card "([^"]*)" and move it over "([^"]*)"$`, w.takeHold)
	sc.Step(`^I take hold of the card "([^"]*)" and move it under the card "([^"]*)"$`, w.takeHoldUnder)
	sc.Step(`^I drag the card "([^"]*)" onto "([^"]*)"$`, w.dragOnto)
	sc.Step(`^I drag the card "([^"]*)" onto "([^"]*)" above "([^"]*)"$`, w.dragAbove)
	sc.Step(`^I drag the card "([^"]*)" onto the lane "([^"]*)"$`, w.dragToLane)
	sc.Step(`^I scroll down over "([^"]*)"$`, w.scrollOver)
	sc.Step(`^it was worked on (\d+) times$`, w.workedOn)
	sc.Step(`^I scroll down over the history$`, func() error { return w.scrollHistory(1) })
	sc.Step(`^I scroll up over the history$`, func() error { return w.scrollHistory(-1) })
	sc.Step(`^I scroll down over the history (\d+) times$`, func(n int) error {
		for range n {
			if err := w.scrollHistory(1); err != nil {
				return err
			}
		}
		return nil
	})
	sc.Step(`^I drag the history's bar to the bottom$`, w.dragHistoryBar)
	sc.Step(`^I choose "([^"]*)"$`, w.choose)
	sc.Step(`^I set the field "([^"]*)" to "([^"]*)"$`, w.setField)
	sc.Step(`^I put the keyboard on "([^"]*)"$`, w.keyboardOn)
	sc.Step(`^I type "([^"]*)"$`, w.types)
	sc.Step(`^I click the field "([^"]*)"$`, w.clickField)
	sc.Step(`^I click (\d+) characters into the field "([^"]*)"$`, w.clickInto)
	sc.Step(`^somebody else adds a task titled "([^"]*)"$`, w.somebodyElseAdds)
	sc.Step(`^the window looks again$`, w.looksAgain)

	w.expects(sc)
}

// workedOn edits the last task added over and over, so it carries more history
// than any panel could hold.
func (w *world) workedOn(times int) error {
	found, err := w.other.List()
	if err != nil {
		return err
	}
	if len(found) == 0 {
		return fmt.Errorf("nothing has been added to work on")
	}
	number := task.OfTask(found[len(found)-1]).Number
	for i := range times {
		state := []string{"todo", "current", "backlog"}[i%3]
		if _, err := w.other.Edit(number, engine.Changes{State: &state}); err != nil {
			return err
		}
	}
	return nil
}

// middleOfHistory is a point over the history's box, which is what the wheel
// has to be over for the history and not a column to move.
func (w *world) middleOfHistory() Point {
	box := HistoryBox(Over(w.w.width, w.w.height))
	return Point{X: box.X + box.W/2, Y: box.Y + box.H/2}
}

func (w *world) scrollHistory(by int) error {
	w.was = w.w.history
	w.w = w.w.Scroll(w.middleOfHistory(), by*lineHigh)
	return nil
}

// dragHistoryBar takes hold of the bar beside the history and pulls it past
// the bottom of its track, which is as far as it goes.
func (w *world) dragHistoryBar() error {
	track := HistoryBar(Over(w.w.width, w.w.height))
	thumb := Thumb(track, w.w.Told(), w.w.Rows(), w.w.history)
	w.was = w.w.history
	w.w = w.w.PressAt(Point{X: thumb.X + thumb.W/2, Y: thumb.Y + thumb.H/2})
	w.w = w.w.MoveTo(Point{X: thumb.X + thumb.W/2, Y: track.Y + track.H + 999})
	w.w = w.w.LetGo(Point{X: thumb.X + thumb.W/2, Y: track.Y + track.H + 999})
	return nil
}

func (w *world) addManyObjectives(n int) error {
	for i := 1; i <= n; i++ {
		if err := w.addObjective(fmt.Sprintf("Objective %d", i)); err != nil {
			return err
		}
	}
	return nil
}

func (w *world) scrollObjectives() error {
	w.was = w.w.goals
	box := w.w.Layout().Column[0]
	w.w = w.w.Scroll(Point{X: box.X + box.W/2, Y: box.Y + box.H/2}, CardHeight)
	return nil
}
