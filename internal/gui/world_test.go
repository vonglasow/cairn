package gui

import (
	"context"
	"errors"
	"fmt"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"gioui.org/io/key"
	"gioui.org/widget/material"

	"github.com/cucumber/godog"

	"cairn/internal/build"
	"cairn/internal/config"
	"cairn/internal/domain"
	"cairn/internal/engine"
	"cairn/internal/task"
	"cairn/test"
)

type world struct {
	t      *testing.T
	path   string
	held   *engine.Engine
	other  *engine.Engine
	work   *refusing
	w      Window
	opened Box
	pages  engine.Pages

	width, high  int
	today        time.Time
	system       string
	clipboard    string
	file         string
	scrolled     [engine.Columns]int
	laneScrolled [3]int
	was          int
	pointer      Point
	theme        *material.Theme
	keepSizes    bool
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
	w.work = &refusing{held: held}
	w.w = Window{}
	w.width, w.high = 1200, 800
	w.keepSizes = false
	w.system = "darwin"
	w.today = moment
	w.scrolled = [engine.Columns]int{}
	w.laneScrolled = [3]int{}
}

type refusing struct {
	held        *engine.Engine
	refuse      error
	refuseParts bool
	reads       int
	wrote       []wrote
}

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
	if r.refuseParts && op.Thing == "task" && addsAPart(in) {
		return engine.Answer{}, fmt.Errorf("that part cannot be added")
	}
	if op.Thing == "task" && op.Verb == "list" {
		r.reads++
	}
	return r.held.Ask(op, in)
}

func (r *refusing) Version() (int64, error) { return r.held.Version() }

func addsAPart(in engine.Input) bool {
	for _, part := range engine.PartsIn(in.Given["steps"]) {
		if part.Number == 0 {
			return true
		}
	}
	return false
}

func (w *world) aWindow(width, high int) error {
	w.width, w.high = width, high
	return nil
}

func (w *world) givens() test.Givens {
	return test.Givens{Other: func() test.Asker { return standing{w} }}
}

type standing struct{ w *world }

func (s standing) Ask(op engine.Operation, in engine.Input) (engine.Answer, error) {
	answer, err := s.w.other.Ask(op, in)
	if !errors.Is(err, engine.ErrNewerStore) {
		return answer, err
	}
	err = test.AsANewerCairn(unknownChange, func() error {
		answer, err = s.w.other.Ask(op, in)
		return err
	})
	return answer, err
}

func (w *world) aStoreFromAnOlderCairn(title string) error {
	if err := w.addTask(title, "", ""); err != nil {
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
	w.held, w.other, w.work = held, other, &refusing{held: held}
	return nil
}

func (w *world) anotherCairnRollsTheLastChangeBack() error {
	if _, err := w.other.Ask(engine.MigrationRollback, engine.Agreed(engine.Input{})); err != nil {
		return err
	}
	return w.forgetBackups()
}

const unknownChange = "never heard of"

func (w *world) aChangeThisCairnDoesNotKnow() error {
	if err := test.AsANewerCairn(unknownChange, func() error {
		_, err := w.other.Ask(engine.MigrationApply, engine.Input{})
		return err
	}); err != nil {
		return err
	}
	return w.forgetBackups()
}

func (w *world) aStoreANewerCairnChanged(title string) error {
	if err := w.addTask(title, "", ""); err != nil {
		return err
	}
	if err := w.aChangeThisCairnDoesNotKnow(); err != nil {
		return err
	}
	_ = w.held.Close()
	held, err := engine.Hold(engine.Home{Dir: filepath.Dir(w.path), Lookup: func(name string) (string, bool) {
		return w.path, name == config.StoreVariable
	}}, engine.Through("gui", engine.Home{}))
	if err != nil {
		return err
	}
	w.held, w.work = held, &refusing{held: held}
	return nil
}

func (w *world) backupsKept() (int, error) {
	kept, err := w.work.Ask(engine.BackupList, engine.Input{})
	return len(kept.Backups), err
}

func (w *world) changesWaiting() (int, error) {
	listed, err := w.work.Ask(engine.MigrationList, engine.Input{})
	waiting := 0
	for _, c := range listed.Changes {
		if !c.Carried().Applied {
			waiting++
		}
	}
	return waiting, err
}

func (w *world) addTask(title, state, desc string) error {
	given := map[string]string{"title": title}
	if state != "" {
		given["state"] = state
	}
	if desc != "" {
		given["desc"] = desc
	}
	_, err := w.other.Ask(engine.TaskNew, engine.Input{Given: given})
	return err
}

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

func (w *world) open() error {
	w.w = New(w.work, w.width, w.high).OnSystem(w.system).At(w.today).Paged(w.pages).
		Drafting(DraftsAt(engine.Places{Store: w.path}.Draft()))
	if w.keepSizes {
		w.w = w.w.KeepingSizes(SizesAt(engine.Places{Store: w.path}.Sizes()))
	}
	w.w = w.w.Look()
	return w.w.Wrong()
}

func (w *world) resize(width, high int) error {
	w.w = w.w.Resized(width, high)
	return nil
}

func (w *world) find(title string) (Point, error) {
	l := w.w.Layout()
	dealt, stalled := engine.Deal(w.w.Reading().Cards, w.w.archived)
	for column, held := range dealt {
		for at, c := range held.Cards {
			if c.Title != title {
				continue
			}
			box := l.CardAt(column, at, w.w.scroll[column])
			return Point{X: box.X + box.W/2, Y: box.Y + box.H/2}, nil
		}
	}
	for lane, kind := range engine.Stalled {
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

func (w *world) between(above, below string) (Point, error) {
	top, err := w.find(above)
	if err != nil {
		return Point{}, err
	}
	under, err := w.find(below)
	if err != nil {
		return Point{}, err
	}
	if under.X != top.X || under.Y-top.Y != CardHeight {
		return Point{}, fmt.Errorf("%q does not sit straight under %q", below, above)
	}
	return Point{X: top.X, Y: top.Y + CardHeight/2 - gap/2}, nil
}

func (w *world) doubleClick(title string) error {
	at, err := w.find(title)
	if err != nil {
		return err
	}
	edge := Point{X: at.X - w.w.Layout().Column[0].W/2 + 2, Y: at.Y}
	w.w = w.w.PressAt(edge).LetGo(edge)
	w.w = w.w.doubleClicked(edge).LetGo(edge)
	return nil
}

func (w *world) clickTab(name string) error {
	tab := slices.Index(Pages, name)
	if tab < 0 {
		return fmt.Errorf("no tab called %q", name)
	}
	box := TabAt(w.w.Layout(), int(tab))
	w.w = w.w.clicked(Point{X: box.X + box.W/2, Y: box.Y + box.H/2})
	return nil
}

func (w *world) findGoal(title string) (Point, error) {
	for i, g := range engine.Sorted(w.w.Reading().Goals) {
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

func (w *world) scrollParts(times int) error {
	tiles := w.w.partTiles()
	if len(tiles) < 2 {
		return fmt.Errorf("the form shows no part")
	}
	w.was = w.w.partsFrom()
	over := Point{X: tiles[0].X + tiles[0].W/2, Y: tiles[0].Y + tiles[0].H/2}
	for range times {
		w.w = w.w.Scroll(over, SubHeight)
	}
	return nil
}

func (w *world) partPoint(title string) (Point, error) {
	for i, c := range w.w.PartsTicked() {
		if c.Title == title {
			tile := w.w.partTiles()[i-w.w.partsFrom()]
			return Point{X: tile.X + tile.W/2, Y: tile.Y + tile.H/2}, nil
		}
	}
	return Point{}, fmt.Errorf("no part is titled %q", title)
}

func (w *world) clickWith(held key.Modifiers, find func(string) (Point, error), title string) error {
	at, err := find(title)
	if err != nil {
		return err
	}
	w.w = w.w.KeysDown(held).clicked(at).KeysDown(0)
	return nil
}

func (w *world) clickOffer(name string) error {
	for i, offer := range w.w.Offers() {
		if offer != name {
			continue
		}
		box := OfferAt(w.w.Layout(), i)
		if s := w.w.screen; s != OnTheBoard && s != OnSearch {
			box = w.w.PanelOffers()[i]
		}
		w.w = w.w.clicked(Point{X: box.X + box.W/2, Y: box.Y + box.H/2})
		return w.w.Wrong()
	}
	return fmt.Errorf("the bar offers %v, and none of them is %q", w.w.Offers(), name)
}

func (w *world) press(name string) error {
	said, with := asGio(name)
	w.w = Typed(w.w, said, with)
	return w.clipping()
}

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

func asGio(name string) (string, key.Modifiers) {
	var with key.Modifiers
	for found := true; found; {
		found = false
		for prefix, mod := range map[string]key.Modifiers{
			"shift+": key.ModShift, "cmd+": key.ModCommand, "ctrl+": key.ModCtrl, "alt+": key.ModAlt,
		} {
			if rest, ok := strings.CutPrefix(name, prefix); ok {
				name, with, found = rest, with|mod, true
			}
		}
	}
	switch name {
	case "home":
		return string(key.NameHome), with
	case "end":
		return string(key.NameEnd), with
	case "delete-forward", "delete":
		return string(key.NameDeleteForward), with
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

func (w *world) slotIn(column string, above string) (Point, error) {
	at := engine.IndexOf(strings.ToLower(column))
	if at < 0 {
		return Point{}, fmt.Errorf("no column called %q", column)
	}
	l := w.w.Layout()
	slot := 0
	if above != "" {
		dealt, _ := engine.Deal(w.w.Reading().Cards, w.w.archived)
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
	return Point{X: box.X + box.W/2, Y: box.Y + box.H/4}, nil
}

func (w *world) footOf(column string) (Point, error) {
	at := engine.IndexOf(strings.ToLower(column))
	if at < 0 {
		return Point{}, fmt.Errorf("no column called %q", column)
	}
	box := w.w.Layout().Column[at]
	return Point{X: box.X + box.W/2, Y: box.Y + box.H - 4}, nil
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

func (w *world) dragToLaneAbove(title, kind, above string) error {
	from, err := w.find(title)
	if err != nil {
		return err
	}
	lane, err := laneNamed(kind)
	if err != nil {
		return err
	}
	_, stalled := engine.Deal(w.w.Reading().Cards, w.w.archived)
	for at, c := range engine.InLanes(stalled)[lane] {
		if c.Title != above {
			continue
		}
		box := w.w.Layout().LaneCardAt(lane, at, w.w.lanes[lane])
		onto := Point{X: box.X + box.W/2, Y: box.Y + box.H/4}
		w.w = w.w.PressAt(from)
		w.w = w.w.MoveTo(onto)
		w.w = w.w.LetGo(onto)
		return nil
	}
	return fmt.Errorf("no card titled %q in the lane %q", above, kind)
}

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

func (w *world) under(title string) (Point, error) {
	l := w.w.Layout()
	dealt, _ := engine.Deal(w.w.Reading().Cards, w.w.archived)
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
	at := engine.IndexOf(strings.ToLower(column))
	if at < 0 {
		return fmt.Errorf("no column called %q", column)
	}
	w.scrolled = w.w.scroll
	box := w.w.Layout().Column[at]
	w.w = w.w.Scroll(Point{X: box.X + box.W/2, Y: box.Y + box.H/2}, CardHeight)
	return nil
}

func laneNamed(kind string) (int, error) {
	for i, name := range engine.Stalled {
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

func (w *world) clickField(name string) error {
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

func (w *world) clickInto(characters int, name string) error {
	box := Over(w.w.Layout().Bar.W+2*margin, w.w.Layout().Bar.Y+barHigh+margin)
	fields := w.w.Form().Fields()
	for i, field := range fields {
		if field.Name != name {
			continue
		}
		value := formValue(box, fields, i)
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

	sc.StepContext().After(func(ctx context.Context, _ *godog.Step, _ godog.StepResultStatus, err error) (context.Context, error) {
		if err != nil || w.w.width == 0 {
			return ctx, err
		}
		if w.theme == nil {
			w.theme = Theme()
		}
		return ctx, drawsWithoutFault(w.theme, w.w)
	})

	w.givens().Register(sc)
	sc.Given(`^a window of (\d+) by (\d+)$`, w.aWindow)
	sc.Step(`^the store refuses what it is asked to change$`, w.storeRefuses)
	sc.Step(`^the store refuses to add a step$`, func() error {
		w.work.refuseParts = true
		return nil
	})

	sc.Given(`^a store from an older cairn, holding a task titled "([^"]*)"$`, w.aStoreFromAnOlderCairn)
	sc.When(`^another cairn rolls the last change back$`, w.anotherCairnRollsTheLastChangeBack)
	sc.When(`^another cairn applies a change this cairn does not know$`, w.aChangeThisCairnDoesNotKnow)
	sc.Given(`^a store a newer cairn changed, holding a task titled "([^"]*)"$`, w.aStoreANewerCairnChanged)
	sc.Then(`^the only answer it offers is "([^"]*)"$`, func(answer string) error {
		if got := w.w.Offers(); len(got) != 1 || got[0] != answer {
			return fmt.Errorf("the window offers %q, want only %q", got, answer)
		}
		return nil
	})
	sc.Then(`^a backup was set aside$`, func() error {
		if n, err := w.backupsKept(); err != nil || n == 0 {
			return fmt.Errorf("%d backups were set aside, want one: %w", n, err)
		}
		return nil
	})
	sc.Then(`^no backup was set aside$`, func() error {
		if n, err := w.backupsKept(); err != nil || n != 0 {
			return fmt.Errorf("%d backups were set aside, want none: %w", n, err)
		}
		return nil
	})
	sc.Then(`^no change is waiting$`, func() error {
		if n, err := w.changesWaiting(); err != nil || n != 0 {
			return fmt.Errorf("%d changes are waiting, want none: %w", n, err)
		}
		return nil
	})
	sc.Then(`^(\d+) changes? (?:is|are) still waiting$`, func(want int) error {
		if n, err := w.changesWaiting(); err != nil || n != want {
			return fmt.Errorf("%d changes are waiting, want %d: %w", n, want, err)
		}
		return nil
	})
	sc.When(`^I open the window$`, w.open)
	sc.Step(`^the window is closed and opened again$`, w.open)
	sc.Then(`^the window asks to pick up the task left unsaved$`, func() error {
		if w.w.Screen() != OnAsking || w.w.question != aboutDraft {
			return fmt.Errorf("the window is on %v, asking %v", w.w.Screen(), w.w.question)
		}
		return nil
	})
	sc.Step(`^I resize the window to (\d+) by (\d+)$`, w.resize)
	sc.Step(`^the window draws its first frame$`, func() error {
		w.w = w.w.Born(w.today)
		return nil
	})
	sc.Step(`^the window draws a frame (\d+) seconds? later$`, func(n int) error {
		w.w = w.w.At(w.today.Add(time.Duration(n) * time.Second))
		return nil
	})
	sc.Then(`^a cairn is being built$`, func() error {
		if _, on := w.w.Intro(); !on {
			return fmt.Errorf("no cairn is being built")
		}
		return nil
	})
	sc.Then(`^no cairn is being built$`, func() error {
		if progress, on := w.w.Intro(); on {
			return fmt.Errorf("a cairn is still being built, %.0f%% of the way", progress*100)
		}
		return nil
	})
	sc.Given(`^today is "([^"]*)"$`, func(day string) error {
		parsed, err := domain.ParseDeadline(day)
		w.today = parsed.Add(12 * time.Hour)
		return err
	})
	sc.Then(`^the calendar is open on "([^"]*)"$`, func(day string) error {
		if w.w.Screen() != OnCalendar {
			return fmt.Errorf("the window is on %v, not on the calendar", w.w.Screen())
		}
		if got := domain.FormatDay(w.w.day); got != day {
			return fmt.Errorf("the calendar is on %s, want %s", got, day)
		}
		return nil
	})
	sc.Step(`^I click the day "([^"]*)" on the calendar$`, func(day string) error {
		for i, shown := range engine.MonthGrid(w.w.day) {
			if domain.FormatDay(shown) == day {
				box := DayAt(w.w.CalendarBox(), i)
				w.w = w.w.clicked(Point{X: box.X + box.W/2, Y: box.Y + box.H/2})
				return nil
			}
		}
		return fmt.Errorf("the calendar does not show %s", day)
	})
	sc.Then(`^the calendar hangs under the field "([^"]*)"$`, func(name string) error {
		box := w.w.CalendarBox()
		fields := w.w.Form().Fields()
		for i, field := range fields {
			if field.Name != name {
				continue
			}
			_, value, rows := FormPlace(w.w.formBox(), fields, i)
			frame := Framed(value, rows)
			if box.X != frame.X || box.Y != frame.Y+frame.H+between {
				return fmt.Errorf("the calendar is at %d,%d and the field %q ends at %d,%d", box.X, box.Y, name, frame.X, frame.Y+frame.H)
			}
			return nil
		}
		return fmt.Errorf("the form has no field %q", name)
	})
	sc.Step(`^I click the next month on the calendar$`, func() error {
		box := MonthAfter(w.w.CalendarBox())
		w.w = w.w.clicked(Point{X: box.X + box.W/2, Y: box.Y + box.H/2})
		return nil
	})
	sc.Step(`^I click the month before on the calendar$`, func() error {
		box := MonthBefore(w.w.CalendarBox())
		w.w = w.w.clicked(Point{X: box.X + box.W/2, Y: box.Y + box.H/2})
		return nil
	})
	sc.Step(`^I scroll down over the calendar$`, func() error {
		box := w.w.CalendarBox()
		w.w = w.w.Scroll(Point{X: box.X + box.W/2, Y: box.Y + box.H/2}, lineHigh)
		return nil
	})
	sc.Step(`^I save it in the dialog as a file of my own$`, func() error {
		w.file = filepath.Join(w.t.TempDir(), "mine.json")
		return w.answerDialog(dialog{save: w.file})
	})
	sc.Step(`^I pick that file in the dialog$`, func() error { return w.answerDialog(dialog{open: w.file}) })
	sc.Step(`^I close the dialog$`, func() error { return w.answerDialog(dialog{closed: true}) })
	sc.Given(`^a file of my own exported from a store with a task titled "([^"]*)"$`, func(title string) error {
		dir := w.t.TempDir()
		other, err := engine.Open(filepath.Join(dir, "other.db"))
		if err != nil {
			return err
		}
		defer func() { _ = other.Close() }()
		if _, err := other.Ask(engine.TaskNew, engine.Input{Given: map[string]string{"title": title}}); err != nil {
			return err
		}
		w.file = filepath.Join(dir, "tasks.json")
		_, err = other.Ask(engine.TaskExport, engine.Input{Given: map[string]string{"to": w.file}})
		return err
	})
	sc.Then(`^that file holds the task "([^"]*)"$`, func(title string) error {
		held, err := os.ReadFile(w.file)
		if err != nil {
			return err
		}
		if !strings.Contains(string(held), `"title": "`+title+`"`) {
			return fmt.Errorf("the file holds %s", held)
		}
		return nil
	})
	sc.Then(`^task (\d+) carries the priority "([^"]*)"$`, func(n int, priority string) error {
		shown, err := w.other.Ask(engine.TaskShow, engine.Input{Given: map[string]string{"number": strconv.Itoa(n)}})
		if err != nil {
			return err
		}
		if said := engine.RenderText(engine.TaskShow, shown); !strings.Contains(said, "priority     "+priority) {
			return fmt.Errorf("task %d reads:\n%s", n, said)
		}
		return nil
	})
	sc.Step(`^I pick "([^"]*)" in the menu bar$`, func(item string) error {
		w.w = w.w.Picked(item)
		return w.w.Wrong()
	})
	sc.Then(`^the window says the version this build reports$`, func() error {
		said := Heads(w.w)
		for _, line := range w.w.Asks() {
			said += "\n" + line.Means
		}
		if want := build.Read().Version; !strings.Contains(said, "cairn "+want) {
			return fmt.Errorf("the window says %q, not cairn %s", said, want)
		}
		return nil
	})
	menuLines := func() []string {
		var lines []string
		for _, group := range MenuBar(w.w.pages) {
			lines = append(lines, group.Items...)
		}
		return lines
	}
	sc.Then(`^the menu bar offers no (".*")$`, func(list string) error {
		for _, line := range quoted(list) {
			if slices.Contains(menuLines(), line) {
				return fmt.Errorf("the menu bar offers %q: %v", line, menuLines())
			}
		}
		return nil
	})
	sc.Then(`^the menu bar offers (".*")$`, func(list string) error {
		for _, line := range quoted(list) {
			if !slices.Contains(menuLines(), line) {
				return fmt.Errorf("the menu bar does not offer %q: %v", line, menuLines())
			}
		}
		return nil
	})
	sc.Then(`^the menu bar offers nothing$`, func() error {
		if lines := menuLines(); len(lines) > 0 {
			return fmt.Errorf("the menu bar offers %v", lines)
		}
		return nil
	})
	sc.Step(`^I pick "([^"]*)" from the menu bar$`, func(line string) error {
		w.w = w.w.Picked(line)
		return nil
	})
	sc.Given(`^the pages turned off are "([^"]*)"$`, func(list string) error {
		w.pages = engine.PagesOff(strings.Split(list, ", ")...)
		return nil
	})
	sc.Then(`^every column stands on a ground of its own$`, func() error {
		l := w.w.Layout()
		for i := range l.Column {
			head, body := l.Head[i], l.Column[i]
			if err := w.standsApart(Box{X: head.X, Y: head.Y, W: head.W, H: head.H + body.H}); err != nil {
				return fmt.Errorf("column %d: %w", i, err)
			}
		}
		return nil
	})
	sc.Then(`^every column's name is centred in its head$`, func() error {
		return w.everyColumnsName(func(head, name Box) error {
			left, right := name.X-head.X, head.X+head.W-(name.X+name.W)
			if left-right > 2 || right-left > 2 {
				return fmt.Errorf("the name stands %dpx from the left and %dpx from the right", left, right)
			}
			return nil
		})
	})
	sc.Then(`^every column's name stands clear of the edge above it$`, func() error {
		return w.everyColumnsName(func(head, name Box) error {
			if clear := name.Y - head.Y; clear < 6 {
				return fmt.Errorf("the name stands %dpx under the edge, want 6px at least", clear)
			}
			return nil
		})
	})
	sc.Then(`^every lane's name stands as clear of its edges as a column's name does$`, func() error {
		shot, err := drawnInAStep(w.w)
		if err != nil {
			return err
		}
		l := w.w.Layout()
		head := l.Head[0]
		column, ok := inkedIn(shot, Box{X: head.X + 1, Y: head.Y + 1, W: head.W - 2, H: head.H - 1}, Dark().Ground)
		if !ok {
			return fmt.Errorf("a column's head carries no name")
		}
		want := 2*(column.Y-head.Y) + column.H
		for i, lane := range l.Lane {
			top := Box{X: lane.X, Y: l.Band.Y, W: lane.W, H: lane.Y - l.Band.Y}
			name, ok := inkedIn(shot, Box{X: top.X + 1, Y: top.Y + 1, W: top.W - 2, H: top.H - 1}, Dark().Ground)
			if !ok {
				return fmt.Errorf("lane %d: carries no name", i)
			}
			if clear := name.Y - top.Y; clear < 6 {
				return fmt.Errorf("lane %d: the name stands %dpx under the edge, want 6px at least", i, clear)
			}
			if mid := 2*(name.Y-top.Y) + name.H; mid-want > 1 || want-mid > 1 {
				return fmt.Errorf("lane %d: the name is centred %d half-pixels under the edge, a column's %d", i, mid, want)
			}
			if left, right := name.X-top.X, top.X+top.W-(name.X+name.W); left-right > 2 || right-left > 2 {
				return fmt.Errorf("lane %d: the name stands %dpx from the left and %dpx from the right", i, left, right)
			}
			if below := top.Y + top.H - (name.Y + name.H); below < 1 {
				return fmt.Errorf("lane %d: the name is cut by the lane's body", i)
			}
		}
		return nil
	})
	sc.Then(`^the band stands on a ground of its own$`, func() error {
		return w.standsApart(w.w.Layout().Band)
	})
	sc.Then(`^the window draws nothing but its ground$`, func() error {
		shot, err := drawnInAStep(w.w)
		if err != nil {
			return err
		}
		if all, ground := w.w.width*w.w.height, painted(shot, Box{W: w.w.width, H: w.w.height}, Dark().Back); ground != all {
			return fmt.Errorf("%d of %d pixels are something other than the ground", all-ground, all)
		}
		return nil
	})
	sc.Step(`^I click the menu button$`, func() error {
		button := MenuButton(w.w.Layout())
		w.w = w.w.clicked(Point{X: button.X + button.W/2, Y: button.Y + button.H/2})
		return nil
	})
	sc.Then(`^the window is on the page "([^"]*)"$`, func(page string) error {
		if got := Pages[w.w.Tab()]; got != page {
			return fmt.Errorf("the window is on %q", got)
		}
		return nil
	})
	sc.Given(`^the window runs on "([^"]*)"$`, func(system string) error {
		w.system = system
		return nil
	})
	sc.Step(`^I click the card "([^"]*)"$`, w.click)
	sc.Step(`^I click the tab "([^"]*)"$`, w.clickTab)
	sc.Step(`^I ⌘-click the card "([^"]*)"$`, func(title string) error {
		return w.clickWith(key.ModShortcut, w.find, title)
	})
	sc.Step(`^I ⇧-click the card "([^"]*)"$`, func(title string) error {
		return w.clickWith(key.ModShift, w.find, title)
	})
	sc.Step(`^I ⌘-click the objective "([^"]*)"$`, func(title string) error {
		return w.clickWith(key.ModShortcut, w.findGoal, title)
	})
	sc.Step(`^I click the objective "([^"]*)"$`, func(title string) error {
		if err := w.clickGoal(title); err != nil {
			return err
		}
		for i, g := range engine.Sorted(w.w.Showing().Goals) {
			if g.Number == w.w.Chosen() {
				w.was = i
			}
		}
		return nil
	})
	sc.Step(`^I click "([^"]*)"$`, w.clickOffer)
	sc.Step(`^I right-click the card "([^"]*)"$`, func(title string) error {
		at, err := w.find(title)
		if err != nil {
			return err
		}
		w.w = w.w.ContextAt(at)
		return nil
	})
	column := func(name string) (Point, error) {
		for i, heading := range []string{"BACKLOG", "TODO", "CURRENT", "DONE"} {
			if heading == name {
				box := w.w.Layout().Column[i]
				return Point{X: box.X + box.W/2, Y: box.Y + box.H - 10}, nil
			}
		}
		return Point{}, fmt.Errorf("no column is headed %q", name)
	}
	sc.Step(`^I right-click the column "([^"]*)"$`, func(name string) error {
		at, err := column(name)
		w.w = w.w.ContextAt(at)
		return err
	})
	sc.Step(`^I double-click the column "([^"]*)"$`, func(name string) error {
		at, err := column(name)
		w.w = w.w.clicked(at).DoubleClick(at)
		return err
	})
	sc.Step(`^I click the column "([^"]*)"$`, func(name string) error {
		at, err := column(name)
		w.w = w.w.clicked(at)
		return err
	})
	sc.Step(`^I click "([^"]*)" in the menu$`, func(said string) error {
		for i, line := range w.w.Menu() {
			if line == said {
				box := MenuRowAt(w.w.MenuBox(), i)
				w.w = w.w.clicked(Point{X: box.X + box.W/2, Y: box.Y + box.H/2})
				return nil
			}
		}
		return fmt.Errorf("the menu offers %q, and not %q", w.w.Menu(), said)
	})
	sc.Then(`^the menu bar can do (".*")$`, func(list string) error {
		for _, line := range quoted(list) {
			if !w.w.Doable(line) {
				return fmt.Errorf("the menu bar cannot do %q on %v", line, w.w.Screen())
			}
		}
		return nil
	})
	sc.Then(`^the menu bar cannot do (".*")$`, func(list string) error {
		for _, line := range quoted(list) {
			if w.w.Doable(line) {
				return fmt.Errorf("the menu bar can do %q on %v", line, w.w.Screen())
			}
		}
		return nil
	})
	sc.Step(`^I open the menu from its button$`, func() error {
		w.w = w.w.openingMenuBar()
		return nil
	})
	sc.Then(`^the menu offers no (".*")$`, func(list string) error {
		for _, line := range quoted(list) {
			if slices.Contains(w.w.Menu(), line) {
				return fmt.Errorf("the menu offers %q: %v", line, w.w.Menu())
			}
		}
		return nil
	})
	sc.Then(`^the menu also offers (".*")$`, func(list string) error {
		for _, line := range quoted(list) {
			if !slices.Contains(w.w.Menu(), line) {
				return fmt.Errorf("the menu does not offer %q: %v", line, w.w.Menu())
			}
		}
		return nil
	})
	sc.Then(`^the menu offers (".*")$`, func(list string) error {
		var want []string
		for _, quoted := range strings.Split(list, ", ") {
			want = append(want, strings.Trim(quoted, `"`))
		}
		if w.w.Screen() != OnMenu {
			return fmt.Errorf("no menu is open: the window is on %v", w.w.Screen())
		}
		if got := w.w.Menu(); strings.Join(got, "|") != strings.Join(want, "|") {
			return fmt.Errorf("the menu offers %q, want %q", got, want)
		}
		return nil
	})
	sc.Then(`^task (\d+) is in focus$`, func(n int) error {
		c, found := w.w.Look().Reading().Card(n)
		if !found || !c.InFocus {
			return fmt.Errorf("task %d is not in focus", n)
		}
		return nil
	})
	sc.Step(`^I double-click the card "([^"]*)"$`, w.doubleClick)
	sc.Step(`^I click between the cards "([^"]*)" and "([^"]*)"$`, func(above, below string) error {
		at, err := w.between(above, below)
		w.w = w.w.clicked(at)
		return err
	})
	sc.Step(`^I double-click between the cards "([^"]*)" and "([^"]*)"$`, func(above, below string) error {
		at, err := w.between(above, below)
		w.w = w.w.clicked(at).DoubleClick(at)
		return err
	})
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
	sc.Step(`^I press "([^"]*)" holding shift$`, func(said string) error {
		named := map[string]key.Name{"backspace": key.NameDeleteBackward, "delete": key.NameDeleteForward}[said]
		w.w = Typed(w.w, string(named), key.ModShift)
		return nil
	})
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
	sc.Step(`^I drag the card "([^"]*)" onto the lane "([^"]*)" above "([^"]*)"$`, w.dragToLaneAbove)
	sc.Step(`^I drag the card "([^"]*)" to the foot of "([^"]*)"$`, func(title, column string) error {
		from, err := w.find(title)
		if err != nil {
			return err
		}
		foot, err := w.footOf(column)
		if err != nil {
			return err
		}
		w.w = w.w.PressAt(from).MoveTo(foot).LetGo(foot)
		return nil
	})
	sc.Step(`^I scroll down over "([^"]*)" (\d+) times$`, func(column string, n int) error {
		for range n {
			if err := w.scrollOver(column); err != nil {
				return err
			}
		}
		return nil
	})
	sc.Step(`^I let go at the foot of "([^"]*)"$`, func(column string) error {
		foot, err := w.footOf(column)
		if err != nil {
			return err
		}
		w.w = w.w.MoveTo(foot).LetGo(foot)
		return nil
	})
	sc.Then(`^the last card of "([^"]*)" is "([^"]*)"$`, func(column, title string) error {
		here, err := w.cardsOf(column)
		if err != nil {
			return err
		}
		if len(here) == 0 || here[len(here)-1].Title != title {
			names := make([]string, 0, len(here))
			for _, c := range here {
				names = append(names, c.Title)
			}
			return fmt.Errorf("%q reads %q", column, names)
		}
		return nil
	})
	sc.Step(`^I scroll down over "([^"]*)"$`, w.scrollOver)
	sc.Step(`^I scroll down over the lane "([^"]*)"$`, w.scrollOverLane)
	sc.Step(`^I click the top of the lane "([^"]*)"$`, w.clickTopOfLane)
	sc.Step(`^it was worked on (\d+) times$`, w.workedOn)
	sc.Step(`^I scroll down over the description$`, func() error { return w.scrollSaying(1) })
	sc.Step(`^I added (\d+) steps to task (\d+)$`, func(n, number int) error {
		for i := 1; i <= n; i++ {
			if err := w.givens().IAddedTheStepToTask(fmt.Sprintf("Step %d", i), number, false); err != nil {
				return err
			}
		}
		return nil
	})
	sc.Step(`^I scroll down over the steps$`, func() error { return w.scrollParts(1) })
	sc.Step(`^I scroll down over the steps (\d+) times$`, w.scrollParts)
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
	sc.Step(`^I scroll down over the trace$`, func() error { return w.scrollHistory(1) })
	sc.Step(`^I scroll up over the trace$`, func() error { return w.scrollHistory(-1) })
	sc.Step(`^I scroll down over the trace (\d+) times$`, func(n int) error {
		for range n {
			if err := w.scrollHistory(1); err != nil {
				return err
			}
		}
		return nil
	})
	sc.Step(`^I drag the trace's bar to the bottom$`, w.dragHistoryBar)
	sc.Step(`^I drag the form from a place that holds no text (\d+) to the right and (\d+) down$`, func(right, down int) error {
		box := w.w.formPanel()
		from := Point{X: box.X + 3, Y: box.Y + 3}
		if w.w.textAt(from) {
			return fmt.Errorf("the point %v of the form holds a text", from)
		}
		return w.dragForm(from, right, down)
	})
	sc.Step(`^I click the form's cross$`, func() error {
		cross := FormClose(w.w.formPanel())
		w.w = w.w.clicked(Point{X: cross.X + cross.W/2, Y: cross.Y + cross.H/2})
		return nil
	})
	sc.Step(`^I drag the form from its heading (\d+) to the right and (\d+) down$`, func(right, down int) error {
		head := Heading(w.w.formPanel())
		return w.dragForm(Point{X: head.X + 4, Y: head.Y + head.H/2}, right, down)
	})
	sc.Step(`^the form stands (\d+) to the right and (\d+) lower than it opened$`, func(right, down int) error {
		return w.formStandsBy(right, down)
	})
	sc.Step(`^I drag the form from its heading's line past its words (\d+) to the right and (\d+) down$`, func(right, down int) error {
		head := Heading(w.w.formPanel())
		from := Point{X: head.X + ends(w.w.measured(), HeadsOn(w.w, OnForm), heading) + 20, Y: head.Y + head.H/2}
		if _, on := w.w.offerAt(from); on {
			return fmt.Errorf("the point %v past the heading is an offer", from)
		}
		return w.dragForm(from, right, down)
	})
	sc.Step(`^I drag the form from the line naming "([^"]*)" past its name (\d+) to the right and (\d+) down$`, func(name string, right, down int) error {
		fields := w.w.form.Fields()
		for i, field := range fields {
			if field.Name != name {
				continue
			}
			named, _, _ := FormPlace(w.w.formBox(), fields, i)
			from := Point{X: named.X + named.W - 4, Y: named.Y + named.H/2}
			if _, on := w.w.offerAt(from); on {
				return fmt.Errorf("the point %v past the name %q is an offer", from, name)
			}
			return w.dragForm(from, right, down)
		}
		return fmt.Errorf("no field of the form is named %q", name)
	})
	sc.Step(`^the form stands where it opened$`, func() error { return w.formStandsBy(0, 0) })
	sc.Step(`^the form stands where a form opens$`, func() error {
		if w.w.Screen() != OnForm {
			return fmt.Errorf("no form is open: the window is on %v", w.w.Screen())
		}
		if now, opens := w.w.formPanel(), w.w.formOpensIn(); now != opens {
			return fmt.Errorf("the form stands at %v, a form opens at %v", now, opens)
		}
		return nil
	})
	sc.Step(`^the form stands wholly inside the window$`, func() error {
		box := w.w.formPanel()
		if box.X < 0 || box.Y < 0 || box.X+box.W > w.w.width || box.Y+box.H > w.w.height {
			return fmt.Errorf("the form stands at %v in a window of %d by %d", box, w.w.width, w.w.height)
		}
		return nil
	})
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
	sc.Step(`^I take hold of the step "([^"]*)" and hold it over the step "([^"]*)"$`, func(from, onto string) error {
		a, err := w.partPoint(from)
		if err != nil {
			return err
		}
		b, err := w.partPoint(onto)
		if err != nil {
			return err
		}
		w.pointer = b
		w.w = w.w.PressAt(a).MoveTo(Point{X: a.X, Y: a.Y + Grip + 1}).MoveTo(b)
		return nil
	})
	sc.Step(`^I drag the step "([^"]*)" onto the step "([^"]*)"$`, func(from, onto string) error {
		at := func(title string) (Point, error) {
			for i, c := range w.w.PartsTicked() {
				if c.Title == title {
					tile := w.w.partTiles()[i]
					return Point{X: tile.X + tile.W/2, Y: tile.Y + tile.H/2}, nil
				}
			}
			return Point{}, fmt.Errorf("no part is titled %q", title)
		}
		a, err := at(from)
		if err != nil {
			return err
		}
		b, err := at(onto)
		if err != nil {
			return err
		}
		w.w = w.w.PressAt(a).MoveTo(Point{X: a.X, Y: a.Y + Grip + 1}).MoveTo(b).LetGo(b)
		return nil
	})
	sc.Step(`^the keyboard says "([^"]*)" as both a key and as text$`, func(said string) error {
		on, _ := w.w.Typed()
		w.w, _, _ = frames(w.t, w.w,
			key.Event{Name: key.Name(strings.ToUpper(said)), State: key.Press},
			key.EditEvent{Range: key.Range{Start: on.At, End: on.At + on.Held}, Text: said})
		return nil
	})
	sc.Step(`^I type "([^"]*)"$`, w.types)
	sc.Step(`^I compose "([^"]*)" from the dead key "([^"]*)" and the letter "([^"]*)"$`,
		func(made, dead, letter string) error {
			w.w = composing(w.w, made, dead, letter)
			return nil
		})
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
	sc.Step(`^I drag across the field "([^"]*)" from line (\d+) character (\d+) to line (\d+) character (\d+)$`,
		func(name string, fromLine, fromColumn, toLine, toColumn int) error {
			fields := w.w.Form().Fields()
			for i, field := range fields {
				if field.Name != name {
					continue
				}
				_, value, rows := FormPlace(w.w.formBox(), fields, i)
				inside := Inside(Framed(value, rows))
				at := func(line, column int) Point {
					row := LineAt(inside, line, 0)
					return Point{X: row.X + column*glyphOf(FormSize(field)), Y: row.Y + row.H/2}
				}
				w.w = w.w.PressAt(at(fromLine, fromColumn)).MoveTo(at(toLine, toColumn)).LetGo(at(toLine, toColumn))
				return nil
			}
			return fmt.Errorf("the form has no field called %q", name)
		})
	sc.Step(`^I drag across the field "([^"]*)" from (\d+) to past its end$`,
		func(name string, from int) error {
			at, err := w.inField(name)
			if err != nil {
				return err
			}
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
		box := w.w.ChoosingBox()
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
		box := w.w.ChoosingBox()
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
		look := ChoosingLooking(w.w.ChoosingBox())
		w.w = w.w.clicked(Point{X: look.X + 20, Y: look.Y + look.H/2})
		return nil
	})
	sc.Step(`^I click (\d+) characters into the field "([^"]*)"$`, w.clickInto)
	sc.Step(`^somebody else adds a task titled "([^"]*)"$`, w.somebodyElseAdds)
	sc.Step(`^the window looks again$`, w.looksAgain)
	sc.Step(`^somebody else adds the step "([^"]*)" to task (\d+)$`, func(title string, n int) error {
		return w.givens().IAddedTheStepToTask(title, n, false)
	})
	sc.Step(`^somebody else renames task (\d+) "([^"]*)"$`, func(n int, title string) error {
		_, err := w.other.Ask(engine.TaskEdit, engine.Input{Given: map[string]string{"number": strconv.Itoa(n), "title": title}})
		return err
	})

	w.expects(sc)
	w.gutters(sc)
	w.filtersSteps(sc)
}

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

func (w *world) middleOfHistory() Point {
	box := BesideBox(w.w.formPanel())
	return Point{X: box.X + box.W/2, Y: box.Y + box.H/2}
}

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

func (w *world) dragTileOut(what string) error {
	from, err := w.tileAt(what)
	if err != nil {
		return err
	}
	box := BesideBox(w.w.formPanel())
	out := Point{X: from.X, Y: box.Y + box.H + CardHeight}
	w.w = w.w.PressAt(from).MoveTo(out).LetGo(out)
	return nil
}

func (w *world) scrollServing(by int) error {
	w.was = w.w.serving
	w.w = w.w.Scroll(w.middleOfServing(), by*SubHeight)
	return nil
}

func (w *world) middleOfServing() Point {
	box := BesideBox(w.w.formPanel())
	return Point{X: box.X + box.W/2, Y: box.Y + box.H/2}
}

func (w *world) middleOfSaying() Point {
	b, _, _, _, _ := w.w.sayingOn()
	return Point{X: b.X + b.W/2, Y: b.Y + b.H/2}
}

func (w *world) scrollList(by int) error {
	run := ChoosingRun(w.w.ChoosingBox())
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

func (w *world) dragSayingBar() error {
	_, track, _, _, _ := w.w.sayingOn()
	thumb := Thumb(track, len(w.w.SaidLines()), w.w.SaidRows(), w.w.saying)
	w.was = w.w.saying
	w.w = w.w.PressAt(Point{X: thumb.X + thumb.W/2, Y: thumb.Y + thumb.H/2})
	w.w = w.w.MoveTo(Point{X: thumb.X + thumb.W/2, Y: track.Y + track.H + 999})
	w.w = w.w.LetGo(Point{X: thumb.X + thumb.W/2, Y: track.Y + track.H + 999})
	return nil
}

func (w *world) dragHistoryBar() error {
	track := BesideBar(w.w.formPanel())
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

const notesRoot = "/vault"

func (w *world) answerDialog(d dialog) error {
	ask, asked := w.w.FileAsked()
	if !asked {
		return fmt.Errorf("the window asked for no dialog")
	}
	w.w = w.w.FileTaken().Filed(Filing(ask, d))
	return w.w.Wrong()
}

func (w *world) standsApart(b Box) error {
	shot, err := drawnInAStep(w.w)
	if err != nil {
		return err
	}
	skin := Dark()
	for _, c := range []struct {
		what string
		at   Point
		want color.NRGBA
	}{
		{"top, inside", Point{X: b.X + b.W - 3, Y: b.Y + 3}, skin.Ground},
		{"bottom, inside", Point{X: b.X + b.W - 3, Y: b.Y + b.H - 3}, skin.Ground},
		{"left edge", Point{X: b.X, Y: b.Y + b.H/2}, skin.GroundEdge},
		{"just outside on the left", Point{X: b.X - 1, Y: b.Y + b.H/2}, skin.Back},
		{"just outside above", Point{X: b.X + b.W/2, Y: b.Y - 1}, skin.Back},
	} {
		got := shot.RGBAAt(c.at.X, c.at.Y)
		if got.R != c.want.R || got.G != c.want.G || got.B != c.want.B {
			return fmt.Errorf("%s at %d,%d is %d,%d,%d, want %d,%d,%d", c.what, c.at.X, c.at.Y, got.R, got.G, got.B, c.want.R, c.want.G, c.want.B)
		}
	}
	return nil
}

func (w *world) everyColumnsName(holds func(head, name Box) error) error {
	shot, err := drawnInAStep(w.w)
	if err != nil {
		return err
	}
	for i, head := range w.w.Layout().Head {
		inside := Box{X: head.X + 1, Y: head.Y + 1, W: head.W - 2, H: head.H - 1}
		name, ok := inkedIn(shot, inside, Dark().Ground)
		if !ok {
			return fmt.Errorf("column %d: its head carries no name", i)
		}
		if err := holds(head, name); err != nil {
			return fmt.Errorf("column %d: %w", i, err)
		}
	}
	return nil
}

func inkedIn(shot *image.RGBA, b Box, ground color.NRGBA) (Box, bool) {
	left, top, right, bottom := b.X+b.W, b.Y+b.H, b.X-1, b.Y-1
	for y := b.Y; y < b.Y+b.H; y++ {
		for x := b.X; x < b.X+b.W; x++ {
			if got := shot.RGBAAt(x, y); got.R == ground.R && got.G == ground.G && got.B == ground.B {
				continue
			}
			left, right, top, bottom = min(left, x), max(right, x), min(top, y), max(bottom, y)
		}
	}
	if right < left {
		return Box{}, false
	}
	return Box{X: left, Y: top, W: right - left + 1, H: bottom - top + 1}, true
}

func (w *world) dragForm(from Point, right, down int) error {
	w.opened = w.w.formPanel()
	to := Point{X: from.X + right, Y: from.Y + down}
	w.w = w.w.PressAt(from)
	w.w = w.w.MoveTo(Point{X: from.X + right/2, Y: from.Y + down/2})
	w.w = w.w.MoveTo(to)
	w.w = w.w.LetGo(to)
	return nil
}

func (w *world) formStandsBy(right, down int) error {
	if w.w.Screen() != OnForm {
		return fmt.Errorf("the form is no longer open: the window is on %v", w.w.Screen())
	}
	now := w.w.formPanel()
	if got := (Point{X: now.X - w.opened.X, Y: now.Y - w.opened.Y}); got != (Point{X: right, Y: down}) {
		return fmt.Errorf("the form moved by %v, want %d right and %d down", got, right, down)
	}
	return nil
}
