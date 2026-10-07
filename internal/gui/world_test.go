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

	"gioui.org/io/key"

	"github.com/cucumber/godog"

	"cairn/internal/engine"
	"cairn/internal/reading"
	"cairn/internal/task"
	"cairn/test"
)

type world struct {
	t     *testing.T
	path  string
	held  *engine.Engine
	other *engine.Engine
	work  *refusing
	w     Window

	width, high int
	// clipboard is what the system's would hold, kept here because a scenario
	// has no system: the window asks and the loop does, so the world plays the
	// loop's part.
	clipboard string
	// scrolled is what each column stood at before the last gesture, so a
	// scenario can say which one moved.
	scrolled [Columns]int
	// laneScrolled is the same for the three lanes of the band.
	laneScrolled [3]int
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
	// The notes are named, as a surface finds them, and never made: cairn
	// only says where they are, so an absolute place nobody writes to is safe.
	held, err := engine.OpenAt(engine.Places{Store: w.path, Notes: notesRoot})
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
	w.laneScrolled = [3]int{}
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

// givens writes what a scenario says is in the store, as somebody else.
func (w *world) givens() test.Givens {
	return test.Givens{Other: func() test.Asker { return w.other }}
}

func (w *world) addTask(title, state, desc string) error {
	given := map[string]string{"title": title}
	if state != "" {
		given["state"] = state
	}
	if desc != "" {
		given["desc"] = desc
	}
	_, err := w.other.Ask(engine.TaskAdd, engine.Input{Given: given})
	return err
}

// inField is where the nth character of a field sits: where production reads
// a character from, not a geometry worked out beside it.
func (w *world) inField(name string) (func(int) Point, error) {
	box := w.w.formBox()
	fields := w.w.Form().Fields()
	for i, field := range fields {
		if field.Name != name {
			continue
		}
		value := formValue(box, fields, i)
		inside := Inside(value)
		return func(n int) Point {
			return Point{X: inside.X + n*glyphOf(FormSize(field)), Y: inside.Y + inside.H/2}
		}, nil
	}
	return nil, fmt.Errorf("the form has no field called %q", name)
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
	dealt, stalled := reading.Deal(w.w.Reading().Cards, w.w.archived)
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
				box := l.LaneCardAt(lane, at, w.w.lanes[lane])
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
	w.w = w.w.clicked(at)
	return nil
}

// doubleClick is the two presses and the two releases a mouse sends, not a
// call to DoubleClick: what a release does with what the press opened is the
// whole of what this is for.
func (w *world) doubleClick(title string) error {
	at, err := w.find(title)
	if err != nil {
		return err
	}
	// On the card's left edge and not its middle: a card is wider than the
	// panel that opens over it, so the middle of one lands inside the panel
	// and its edge does not. The release that follows the double click landed
	// outside and closed what the click had just opened, and a test aiming at
	// the middle could never see it.
	edge := Point{X: at.X - w.w.Layout().Column[0].W/2 + 2, Y: at.Y}
	w.w = w.w.PressAt(edge).LetGo(edge)
	w.w = w.w.doubleClicked(edge).LetGo(edge)
	return nil
}

func (w *world) clickTab(name string) error {
	which := map[string]Tab{"TASKS": Tasks, "OBJECTIVES": Objectives}
	tab, known := which[name]
	if !known {
		return fmt.Errorf("no tab called %q", name)
	}
	box := TabAt(w.w.Layout(), int(tab))
	w.w = w.w.clicked(Point{X: box.X + box.W/2, Y: box.Y + box.H/2})
	return nil
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
	w.w = w.w.clicked(at)
	return nil
}

// clickWith clicks what is titled with a key held down, the way a hand holds
// ⌘ or ⇧ while it clicks: down before the press, up after the release.
func (w *world) clickWith(held key.Modifiers, find func(string) (Point, error), title string) error {
	at, err := find(title)
	if err != nil {
		return err
	}
	w.w = w.w.KeysDown(held).clicked(at).KeysDown(0)
	return nil
}

// clickOffer presses one of the bar's offers, which is how a person reaches
// anything here without knowing a shortcut.
func (w *world) clickOffer(name string) error {
	for i, offer := range w.w.Offers() {
		if offer != name {
			continue
		}
		// Where it is drawn: on the panel when one is open, and on the bar
		// under the board otherwise. A person clicks what they are looking at.
		box := OfferAt(w.w.Layout(), i)
		if s := w.w.screen; s != OnTheBoard && s != OnSearch {
			box = PanelOfferAt(w.w.panel(), i)
		}
		w.w = w.w.clicked(Point{X: box.X + box.W/2, Y: box.Y + box.H/2})
		return w.w.Wrong()
	}
	return fmt.Errorf("the bar offers %v, and none of them is %q", w.w.Offers(), name)
}

// press sends a key the way a window receives one: through the conversion a
// real keyboard goes through. Driving Key directly let a scenario pass on a
// key no window could send — ⇧A among them, which is how the key that shows
// what was archived came to do nothing outside the tests.
func (w *world) press(name string) error {
	said, with := asGio(name)
	w.w = Typed(w.w, said, with)
	return w.clipping()
}

// clipping is what the loop does between frames: it puts on the clipboard what
// the window asked to copy, and hands back what is on it when the window has
// asked to paste.
func (w *world) clipping() error {
	if said, wants := w.w.Copying(); wants {
		w.clipboard = said
		w.w = w.w.Copied()
	}
	if w.w.Pasting() {
		w.w = w.w.Pasted(w.clipboard)
	}
	return nil
}

// asGio spells a key the way Gio hands it over: the name it gives the key, and
// the modifiers apart from it.
func asGio(name string) (string, key.Modifiers) {
	var with key.Modifiers
	for prefix, mod := range map[string]key.Modifiers{
		"shift+": key.ModShift, "cmd+": key.ModCommand, "ctrl+": key.ModCtrl,
	} {
		if rest, ok := strings.CutPrefix(name, prefix); ok {
			name, with = rest, mod
		}
	}
	switch name {
	case "esc":
		return string(key.NameEscape), with
	case "tab":
		return string(key.NameTab), with
	case "enter", "return":
		return string(key.NameReturn), with
	case "backspace":
		return string(key.NameDeleteBackward), with
	case "left":
		return string(key.NameLeftArrow), with
	case "right":
		return string(key.NameRightArrow), with
	case "up":
		return string(key.NameUpArrow), with
	case "down":
		return string(key.NameDownArrow), with
	}
	return strings.ToUpper(name), with
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
		dealt, _ := reading.Deal(w.w.Reading().Cards, w.w.archived)
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
	lane, err := laneNamed(kind)
	if err != nil {
		return err
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
	dealt, _ := reading.Deal(w.w.Reading().Cards, w.w.archived)
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

// laneNamed is which of the band's lanes holds the kind of stopped named.
func laneNamed(kind string) (int, error) {
	for i, name := range Lanes {
		if name == kind {
			return i, nil
		}
	}
	return 0, fmt.Errorf("no lane called %q", kind)
}

func (w *world) scrollOverLane(kind string) error {
	lane, err := laneNamed(kind)
	if err != nil {
		return err
	}
	w.laneScrolled = w.w.lanes
	box := w.w.Layout().Lane[lane]
	w.w = w.w.Scroll(Point{X: box.X + box.W/2, Y: box.Y + box.H/2}, CardHeight)
	return nil
}

// clickTopOfLane clicks just inside the top of a lane, where whichever card the
// lane has been scrolled to is drawn.
func (w *world) clickTopOfLane(kind string) error {
	lane, err := laneNamed(kind)
	if err != nil {
		return err
	}
	box := w.w.Layout().Lane[lane]
	w.w = w.w.Click(Point{X: box.X + box.W/2, Y: box.Y + 1})
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
	// The box production looks in, not one worked back out of the layout: a
	// helper that reckons its own geometry clicks where nothing is, and says
	// so by failing somewhere else entirely.
	box := w.w.formBox()
	fields := w.w.Form().Fields()
	for i, field := range fields {
		if field.Name != name {
			continue
		}
		row := formName(box, fields, i)
		w.w = w.w.clicked(Point{X: row.X + 10, Y: row.Y + row.H/2})
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
		value := formValue(box, fields, i)
		// Where the text itself begins, which is the frame's own inset and not
		// a card's padding: the two happened to be close enough to land on the
		// right glyph until the spacing scale was put straight.
		inside := Inside(value)
		at := Point{X: inside.X + characters*glyphOf(FormSize(field)), Y: inside.Y + inside.H/2}
		w.w = w.w.clicked(at)
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
	return w.givens().IAddedATaskTitled(title)
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

	w.givens().Register(sc)
	sc.Given(`^a window of (\d+) by (\d+)$`, w.aWindow)
	sc.Step(`^the store refuses what it is asked to change$`, w.storeRefuses)

	sc.When(`^I open the window$`, w.open)
	sc.Step(`^I resize the window to (\d+) by (\d+)$`, w.resize)
	sc.Step(`^I click the card "([^"]*)"$`, w.click)
	sc.Step(`^I click the tab "([^"]*)"$`, w.clickTab)
	sc.Step(`^I ⌘-click the card "([^"]*)"$`, func(title string) error {
		return w.clickWith(key.ModCommand, w.find, title)
	})
	sc.Step(`^I ⇧-click the card "([^"]*)"$`, func(title string) error {
		return w.clickWith(key.ModShift, w.find, title)
	})
	sc.Step(`^I ⌘-click the objective "([^"]*)"$`, func(title string) error {
		return w.clickWith(key.ModCommand, w.findGoal, title)
	})
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
	sc.Step(`^I double-click the card "([^"]*)"$`, w.doubleClick)
	// Pressed and let go, which is how a click reaches this window. Calling
	// Click straight misses everything a release decides on the way — and it
	// did: this passed while no click outside a panel closed anything at all.
	sc.Step(`^I click outside the panel$`, func() error {
		off := Point{X: 2, Y: 2}
		w.w = w.w.PressAt(off).LetGo(off)
		return nil
	})
	sc.Step(`^I click the search box$`, func() error {
		box := Searching(w.w.Layout())
		w.w = w.w.clicked(Point{X: box.X + box.W/2, Y: box.Y + box.H/2})
		return nil
	})
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
	sc.Step(`^I scroll down over the lane "([^"]*)"$`, w.scrollOverLane)
	sc.Step(`^I click the top of the lane "([^"]*)"$`, w.clickTopOfLane)
	sc.Step(`^it was worked on (\d+) times$`, w.workedOn)
	sc.Step(`^I scroll down over the description$`, func() error { return w.scrollSaying(1) })
	sc.Step(`^I scroll up over the description$`, func() error { return w.scrollSaying(-1) })
	sc.Step(`^I scroll down over the description (\d+) times$`, func(n int) error {
		for range n {
			if err := w.scrollSaying(1); err != nil {
				return err
			}
		}
		return nil
	})
	sc.Step(`^I added a task titled "([^"]*)" with a description of (\d+) lines$`,
		func(title string, lines int) error {
			said := make([]string, lines)
			for i := range said {
				said[i] = fmt.Sprintf("Ligne %d de ce que la tâche raconte", i+1)
			}
			return w.addTask(title, "", strings.Join(said, "\n"))
		})
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
	sc.Step(`^I drag the description's bar to the bottom$`, w.dragSayingBar)
	sc.Step(`^I scroll up over the description (\d+) times$`, func(n int) error {
		for range n {
			if err := w.scrollSaying(-1); err != nil {
				return err
			}
		}
		return nil
	})
	sc.Step(`^I added (\d+) tasks serving objective (\d+)$`, func(n, serving int) error {
		for i := range n {
			if err := w.givens().IAddedATaskServingObjective(fmt.Sprintf("Task %d", i+1), serving); err != nil {
				return err
			}
		}
		return nil
	})
	sc.Step(`^I drag the tile "([^"]*)" onto the tile "([^"]*)"$`, w.dragTile)
	sc.Step(`^I take hold of the tile "([^"]*)" and move it over the tile "([^"]*)"$`, func(what, onto string) error {
		from, err := w.tileAt(what)
		if err != nil {
			return err
		}
		to, err := w.tileAt(onto)
		if err != nil {
			return err
		}
		w.w = w.w.PressAt(from).MoveTo(to)
		return nil
	})
	sc.Step(`^I drag the tile "([^"]*)" out of its box$`, w.dragTileOut)
	sc.Step(`^I scroll down over the work under it$`, func() error { return w.scrollServing(1) })
	sc.Step(`^I scroll up over the work under it$`, func() error { return w.scrollServing(-1) })
	sc.Step(`^I scroll down over the work under it (\d+) times$`, func(n int) error {
		for range n {
			if err := w.scrollServing(1); err != nil {
				return err
			}
		}
		return nil
	})
	sc.Step(`^I choose "([^"]*)"$`, w.choose)
	sc.Step(`^I set the field "([^"]*)" to "([^"]*)"$`, w.setField)
	sc.Step(`^I put the keyboard on "([^"]*)"$`, w.keyboardOn)
	// Both halves of one keystroke, in the order a keyboard sends them and
	// through Gio's own router: the name first, then the text an input method
	// made of it. This is what Linux does and macOS does not, and driving it
	// here is how the difference is reached without a second machine.
	sc.Step(`^the keyboard says "([^"]*)" as both a key and as text$`, func(said string) error {
		w.w, _, _ = frames(w.t, w.w,
			key.Event{Name: key.Name(strings.ToUpper(said)), State: key.Press},
			key.EditEvent{Text: said})
		return nil
	})
	sc.Step(`^I type "([^"]*)"$`, w.types)
	sc.Step(`^I click the field "([^"]*)"$`, w.clickField)
	sc.Step(`^I drag across the field "([^"]*)" from (\d+) to (\d+)$`,
		func(name string, from, to int) error {
			at, err := w.inField(name)
			if err != nil {
				return err
			}
			w.w = w.w.PressAt(at(from)).MoveTo(at(to)).LetGo(at(to))
			return nil
		})
	sc.Step(`^I drag across the field "([^"]*)" from (\d+) to past its end$`,
		func(name string, from int) error {
			at, err := w.inField(name)
			if err != nil {
				return err
			}
			// Past the frame and past the panel, where no field is: the hand
			// overshoots, and what it holds should not stop where it left.
			past := Point{X: w.w.width - 2, Y: at(from).Y}
			w.w = w.w.PressAt(at(from)).MoveTo(past).LetGo(past)
			return nil
		})
	sc.Step(`^I double-click the field "([^"]*)" at (\d+) and drag to (\d+)$`,
		func(name string, from, to int) error {
			at, err := w.inField(name)
			if err != nil {
				return err
			}
			w.w = w.w.doubleDragged(at(from), at(to))
			return nil
		})
	sc.Step(`^I double-click the list row "([^"]*)"$`, func(title string) error {
		box := ChoosingOver(w.w.width, w.w.height)
		for i, one := range w.w.Choosing() {
			if one.Title != title {
				continue
			}
			row := ChoosingRow(box, ChoosingFrom(box)+(i-w.w.ChoosingFromRow())*apart)
			w.w = w.w.doubleClicked(Point{X: row.X + row.W/2, Y: row.Y + row.H/2})
			return nil
		}
		return fmt.Errorf("the list offers no row called %q", title)
	})
	sc.Step(`^I click the list row "([^"]*)"$`, func(title string) error {
		box := ChoosingOver(w.w.width, w.w.height)
		for i, one := range w.w.Choosing() {
			if one.Title != title {
				continue
			}
			row := ChoosingRow(box, ChoosingFrom(box)+i*apart)
			w.w = w.w.clicked(Point{X: row.X + row.W/2, Y: row.Y + row.H/2})
			return nil
		}
		return fmt.Errorf("the list offers no row called %q", title)
	})
	sc.Step(`^I added (\d+) tasks$`, func(n int) error {
		for i := 1; i <= n; i++ {
			if err := w.addTask(fmt.Sprintf("Une tâche de plus, la %d", i), "", ""); err != nil {
				return err
			}
		}
		return nil
	})
	sc.Step(`^I scroll down over the list$`, func() error { return w.scrollList(1) })
	sc.Step(`^I scroll down over the list (\d+) times$`, func(n int) error {
		for range n {
			if err := w.scrollList(1); err != nil {
				return err
			}
		}
		return nil
	})
	sc.Step(`^I click the line that narrows the list$`, func() error {
		look := ChoosingLooking(ChoosingOver(w.w.width, w.w.height))
		w.w = w.w.clicked(Point{X: look.X + 20, Y: look.Y + look.H/2})
		return nil
	})
	sc.Step(`^I click (\d+) characters into the field "([^"]*)"$`, w.clickInto)
	sc.Step(`^somebody else adds a task titled "([^"]*)"$`, w.somebodyElseAdds)
	sc.Step(`^the window looks again$`, w.looksAgain)

	w.expects(sc)
}

// workedOn edits the last task added over and over, so it carries more history
// than any panel could hold.
func (w *world) workedOn(times int) error {
	found, err := w.givens().Tasks()
	if err != nil {
		return err
	}
	if len(found) == 0 {
		return fmt.Errorf("nothing has been added to work on")
	}
	number := task.OfTask(found[len(found)-1]).Number
	for i := range times {
		if err := w.givens().IEditedTaskWithTheState(number, []string{"todo", "current", "backlog"}[i%3]); err != nil {
			return err
		}
	}
	return nil
}

// middleOfHistory is a point over the history's box, which is what the wheel
// has to be over for the history and not a column to move.
func (w *world) middleOfHistory() Point {
	box := BesideBox(Over(w.w.width, w.w.height))
	return Point{X: box.X + box.W/2, Y: box.Y + box.H/2}
}

// tileAt is the middle of the tile a part is drawn on.
func (w *world) tileAt(title string) (Point, error) {
	for at, c := range w.w.Under() {
		if c.Title == title {
			box := w.w.ServingCardAt(at)
			return Point{X: box.X + box.W/2, Y: box.Y + box.H/2}, nil
		}
	}
	return Point{}, fmt.Errorf("nothing under it is titled %q", title)
}

func (w *world) dragTile(what, onto string) error {
	from, err := w.tileAt(what)
	if err != nil {
		return err
	}
	to, err := w.tileAt(onto)
	if err != nil {
		return err
	}
	w.w = w.w.PressAt(from).MoveTo(to).LetGo(to)
	return nil
}

// dragTileOut lets a tile go past the bottom of its box, which is nowhere it
// can land: a part moves among its sisters and has no column to fall into.
func (w *world) dragTileOut(what string) error {
	from, err := w.tileAt(what)
	if err != nil {
		return err
	}
	box := BesideBox(Over(w.w.width, w.w.height))
	out := Point{X: from.X, Y: box.Y + box.H + CardHeight}
	w.w = w.w.PressAt(from).MoveTo(out).LetGo(out)
	return nil
}

func (w *world) scrollServing(by int) error {
	w.was = w.w.serving
	// One tile a notch, the way a column is scrolled by one card.
	w.w = w.w.Scroll(w.middleOfServing(), by*SubHeight)
	return nil
}

// middleOfServing is a point inside the box the work under an objective is
// drawn in, which is where a wheel has to be for it to scroll that box.
func (w *world) middleOfServing() Point {
	box := BesideBox(Over(w.w.width, w.w.height))
	return Point{X: box.X + box.W/2, Y: box.Y + box.H/2}
}

// middleOfSaying is a point over the description on an open task.
func (w *world) middleOfSaying() Point {
	b, _, _, _, _ := w.w.sayingOn()
	return Point{X: b.X + b.W/2, Y: b.Y + b.H/2}
}

// scrollList turns the wheel over an open list.
func (w *world) scrollList(by int) error {
	run := ChoosingRun(ChoosingOver(w.w.width, w.w.height))
	w.was = w.w.ChoosingFromRow()
	w.w = w.w.Scroll(Point{X: run.X + run.W/2, Y: run.Y + run.H/2}, by*lineHigh)
	return nil
}

func (w *world) scrollSaying(by int) error {
	w.was = w.w.saying
	w.w = w.w.Scroll(w.middleOfSaying(), by*lineHigh)
	return nil
}

func (w *world) scrollHistory(by int) error {
	w.was = w.w.history
	w.w = w.w.Scroll(w.middleOfHistory(), by*lineHigh)
	return nil
}

// dragSayingBar pulls the bar beside a description past the bottom of its
// track, the way the history's is pulled.
func (w *world) dragSayingBar() error {
	_, track, _, _, _ := w.w.sayingOn()
	thumb := Thumb(track, len(w.w.SaidLines()), w.w.SaidRows(), w.w.saying)
	w.was = w.w.saying
	w.w = w.w.PressAt(Point{X: thumb.X + thumb.W/2, Y: thumb.Y + thumb.H/2})
	w.w = w.w.MoveTo(Point{X: thumb.X + thumb.W/2, Y: track.Y + track.H + 999})
	w.w = w.w.LetGo(Point{X: thumb.X + thumb.W/2, Y: track.Y + track.H + 999})
	return nil
}

// dragHistoryBar takes hold of the bar beside the history and pulls it past
// the bottom of its track, which is as far as it goes.
func (w *world) dragHistoryBar() error {
	track := BesideBar(Over(w.w.width, w.w.height))
	thumb := Thumb(track, w.w.Told(), w.w.Rows(), w.w.history)
	w.was = w.w.history
	w.w = w.w.PressAt(Point{X: thumb.X + thumb.W/2, Y: thumb.Y + thumb.H/2})
	w.w = w.w.MoveTo(Point{X: thumb.X + thumb.W/2, Y: track.Y + track.H + 999})
	w.w = w.w.LetGo(Point{X: thumb.X + thumb.W/2, Y: track.Y + track.H + 999})
	return nil
}

func (w *world) addManyObjectives(n int) error {
	for i := 1; i <= n; i++ {
		if err := w.givens().IAddedAnObjectiveTitled(fmt.Sprintf("Objective %d", i)); err != nil {
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

// notesRoot is where a scenario's board is told the notes live.
const notesRoot = "/vault"
