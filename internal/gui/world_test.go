package gui

import (
	"cairn/internal/board"
	"cairn/internal/cli/render"
	"cairn/internal/engine/calendar"
	"cairn/internal/gui/design"
	"cairn/internal/gui/geom"
	"cmp"
	"context"
	"fmt"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"gioui.org/io/key"
	"gioui.org/widget/material"

	"github.com/cucumber/godog"

	"cairn/internal/engine"
	"cairn/internal/engine/task"
	"cairn/internal/store/files"
	"cairn/test"
)

type world struct {
	t      *testing.T
	path   string
	home   string
	held   *engine.Engine
	other  *engine.Engine
	work   *refusing
	w      Window
	opened geom.Box
	pages  engine.Pages
	recurs bool

	width, high  int
	today        time.Time
	system       string
	clipboard    string
	file         string
	scrolled     [board.Columns]int
	laneScrolled [3]int
	was          int
	theme        *material.Theme
	keepSizes    bool
	lessMotion   bool
	spaces       *test.Spaces
	apart        elsewhere
}

var drawingTheme = sync.OnceValue(Theme)

var themes = sync.Pool{New: func() any { return Theme() }}

func newWorld(t *testing.T) *world {
	return &world{t: t, width: 1200, high: 800, spaces: &test.Spaces{}}
}

func (w *world) reset() {
	if w.held != nil {
		_ = w.held.Close()
	}
	if w.other != nil {
		_ = w.other.Close()
	}
	w.path = filepath.Join(w.t.TempDir(), "cairn.db")
	w.home = w.t.TempDir()
	held, err := engine.OpenAt(engine.Places{Store: w.path, Notes: notesRoot, Home: engine.Home{Dir: w.home}})
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
	w.recurs = false
	w.system = "darwin"
	w.today = moment
	w.scrolled = [board.Columns]int{}
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
	if op.Thing == engine.WindowKeep.Thing {
		return r.held.Ask(op, in)
	}
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
	return test.Givens{Other: func() test.Asker { return w.other }}
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

func (w *world) aChangeThisCairnDoesNotKnow() error {
	newer, err := test.ANewerCairn(w.path)
	if err != nil {
		return err
	}
	if _, err := newer.Ask(engine.MigrationApply, engine.Input{}); err != nil {
		_ = newer.Close()
		return err
	}
	_ = w.other.Close()
	w.other = newer
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
		return w.path, name == files.StoreVariable
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

func (w *world) inField(name string) (func(int) geom.Point, error) {
	box := w.w.formBox()
	fields := w.w.Form.Fields
	for i, field := range fields {
		if field.Name != name {
			continue
		}
		value := formValue(box, fields, i)
		inside := Inside(value)
		return func(n int) geom.Point {
			return geom.Point{X: inside.X + n*glyphOf(FormSize(field)), Y: inside.Y + inside.H/2}
		}, nil
	}
	return nil, fmt.Errorf("the form has no field called %q", name)
}

func (w *world) storeRefuses() error {
	w.work.refuse = fmt.Errorf("that cannot be done")
	return nil
}

func (w *world) open() error {
	w.w = New(w.work, w.width, w.high).OnSystem(w.system).At(w.today).Paged(w.shownPages()).
		Drafting(DraftsOf(w.work))
	if w.keepSizes {
		if err := os.MkdirAll(filepath.Dir(files.FileIn(w.home)), 0o700); err != nil {
			return err
		}
		if err := os.WriteFile(files.FileIn(w.home), []byte("[window]\nkeep_sizes = true\n"), 0o600); err != nil {
			return err
		}
	}
	w.w = w.w.KeepingSizes(SizesOf(w.work))
	w.w = w.w.Look().ReadingConfiguration().LessMotion(w.lessMotion)
	return w.w.Wrong()
}

func (w *world) find(title string) (geom.Point, error) {
	l := w.w.Layout()
	dealt, stalled := board.Deal(w.w.Reading().Cards, w.w.Archived)
	for column, held := range dealt {
		for at, c := range held.Cards {
			if c.Title != title {
				continue
			}
			box := l.CardAt(column, at, w.w.scroll[column])
			return geom.Point{X: box.X + box.W/2, Y: box.Y + box.H/2}, nil
		}
	}
	for lane, kind := range board.Stalled {
		at := 0
		for _, c := range stalled {
			if c.State != kind {
				continue
			}
			if c.Title == title {
				box := l.LaneCardAt(lane, at, w.w.lanes[lane])
				return geom.Point{X: box.X + box.W/2, Y: box.Y + box.H/2}, nil
			}
			at++
		}
	}
	return geom.Point{}, fmt.Errorf("no card titled %q is on screen", title)
}

func (w *world) click(title string) error {
	at, err := w.find(title)
	if err != nil {
		return err
	}
	w.w = w.w.clicked(at)
	return nil
}

func (w *world) between(above, below string) (geom.Point, error) {
	top, err := w.find(above)
	if err != nil {
		return geom.Point{}, err
	}
	under, err := w.find(below)
	if err != nil {
		return geom.Point{}, err
	}
	if under.X != top.X || under.Y-top.Y != geom.CardHeight {
		return geom.Point{}, fmt.Errorf("%q does not sit straight under %q", below, above)
	}
	return geom.Point{X: top.X, Y: top.Y + geom.CardHeight/2 - geom.Gap/2}, nil
}

func (w *world) doubleClick(title string) error {
	at, err := w.find(title)
	if err != nil {
		return err
	}
	edge := geom.Point{X: at.X - w.w.Layout().Column[0].W/2 + 2, Y: at.Y}
	w.w = w.w.PressAt(edge).LetGo(edge)
	w.w = w.w.doubleClicked(edge).LetGo(edge)
	return nil
}

func (w *world) clickTab(name string) error {
	tab := slices.Index(Pages, name)
	if tab < 0 {
		return fmt.Errorf("no tab called %q", name)
	}
	box := w.w.TabsAt()[slices.Index(w.w.Shown(), board.Tab(tab))]
	w.w = w.w.clicked(geom.Point{X: box.X + box.W/2, Y: box.Y + box.H/2})
	return nil
}

func (w *world) findGoal(title string) (geom.Point, error) {
	for i, b := range w.w.Tree() {
		if b.Goal.Title != title {
			continue
		}
		box := ItemOf(GoalAt(w.w.Layout(), i, w.w.goals), depthOf(b))
		return geom.Point{X: box.X + box.W/2, Y: box.Y + box.H/2}, nil
	}
	return geom.Point{}, fmt.Errorf("no objective titled %q is on screen", title)
}

func (w *world) goalBox(title string) (geom.Box, error) {
	for i, b := range w.w.Tree() {
		if b.Goal.Title == title {
			return ItemOf(GoalAt(w.w.Layout(), i, w.w.goals), depthOf(b)), nil
		}
	}
	return geom.Box{}, fmt.Errorf("no objective titled %q is on screen", title)
}

func (w *world) clickGoal(title string) error {
	at, err := w.findGoal(title)
	if err != nil {
		return err
	}
	w.w = w.w.clicked(at)
	return nil
}

func (w *world) clickWith(held key.Modifiers, find func(string) (geom.Point, error), title string) error {
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
		box := w.w.BarOffers()[i]
		if s := w.w.Screen; s != board.OnTheBoard && s != board.OnSearch {
			box = w.w.PanelOffers()[i]
		}
		w.w = w.w.clicked(geom.Point{X: box.X + box.W/2, Y: box.Y + box.H/2})
		return w.w.Wrong()
	}
	if edit, ok := w.w.OccurrenceEdit(); ok && name == EditTheSeries {
		w.w = w.w.clicked(geom.Point{X: edit.X + edit.W/2, Y: edit.Y + edit.H/2})
		return w.w.Wrong()
	}
	if d := w.w.SeriesDetail(); w.w.Tab == board.Series && w.w.Screen == board.OnTheBoard {
		for i, offer := range d.Offers {
			if offer == name {
				w.w = w.w.clicked(geom.Point{X: d.Buttons[i].X + d.Buttons[i].W/2, Y: d.Buttons[i].Y + d.Buttons[i].H/2})
				return w.w.Wrong()
			}
		}
	}
	return fmt.Errorf("the bar offers %v, and none of them is %q", w.w.Offers(), name)
}

func (w *world) press(name string) error {
	said, with := asGio(name)
	w.w = Typed(w.w, said, with)
	return w.clipping()
}

func (w *world) clipping() error {
	w.handingApart()
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

func (w *world) slotIn(column string, above string) (geom.Point, error) {
	at := board.IndexOf(strings.ToLower(column))
	if at < 0 {
		return geom.Point{}, fmt.Errorf("no column called %q", column)
	}
	l := w.w.Layout()
	slot := 0
	if above != "" {
		dealt, _ := board.Deal(w.w.Reading().Cards, w.w.Archived)
		found := false
		for i, c := range dealt[at].Cards {
			if c.Title == above {
				slot, found = i, true
			}
		}
		if !found {
			return geom.Point{}, fmt.Errorf("no card titled %q in %q", above, column)
		}
	}
	box := l.CardAt(at, slot, w.w.scroll[at])
	return geom.Point{X: box.X + box.W/2, Y: box.Y + box.H/4}, nil
}

func (w *world) footOf(column string) (geom.Point, error) {
	at := board.IndexOf(strings.ToLower(column))
	if at < 0 {
		return geom.Point{}, fmt.Errorf("no column called %q", column)
	}
	box := w.w.Layout().Column[at]
	return geom.Point{X: box.X + box.W/2, Y: box.Y + box.H - 4}, nil
}

func (w *world) dragOnto(title, column string) error { return w.drag(title, column, "") }

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

func laneNamed(kind string) (int, error) {
	for i, name := range board.Stalled {
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
	w.w = w.w.Scroll(geom.Point{X: box.X + box.W/2, Y: box.Y + box.H/2}, geom.CardHeight)
	return nil
}

func (w *world) pullLaneBar(kind string) error {
	track, found, err := w.laneBar(kind)
	if err != nil || !found {
		return cmp.Or(err, fmt.Errorf("the lane %q has no bar to pull", kind))
	}
	w.laneScrolled = w.w.lanes
	top := geom.Point{X: track.X + track.W/2, Y: track.Y + 2}
	foot := geom.Point{X: top.X, Y: track.Y + track.H}
	w.w = w.w.PressAt(top).MoveTo(foot).LetGo(foot)
	return nil
}

func (w *world) choose(name string) error {
	w.w = w.w.Choose(name)
	return w.w.Wrong()
}

func (w *world) keyboardOn(name string) error {
	if w.w.Screen != board.OnForm {
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
	fields := w.w.Form.Fields
	for i, field := range fields {
		if field.Name != name {
			continue
		}
		row := formName(box, fields, i)
		w.w = w.w.clicked(geom.Point{X: row.X + 10, Y: row.Y + row.H/2})
		return nil
	}
	return fmt.Errorf("the form has no field called %q", name)
}

func (w *world) setField(name, value string) error {
	if w.w.Screen != board.OnForm {
		return fmt.Errorf("no form is open")
	}
	w.w.Form = w.w.Form.Set(name, value)
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
		w.theme = themes.Get().(*material.Theme)
		return ctx, nil
	})
	sc.After(func(ctx context.Context, _ *godog.Scenario, err error) (context.Context, error) {
		themes.Put(w.theme)
		return ctx, err
	})

	sc.StepContext().After(func(ctx context.Context, _ *godog.Step, _ godog.StepResultStatus, err error) (context.Context, error) {
		if err != nil || w.w.width == 0 {
			return ctx, err
		}
		return ctx, drawsWithoutFault(w.theme, w.w)
	})

	w.givens().Register(sc)
	w.registerSeries(sc)
	sc.Given(`^a window of (\d+) by (\d+)$`, w.aWindow)
	sc.Step(`^the store refuses what it is asked to change$`, w.storeRefuses)

	sc.Given(`^a store from an older cairn, holding a task titled "([^"]*)"$`, w.aStoreFromAnOlderCairn)
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
	sc.Then(`^no change is waiting$`, func() error {
		if n, err := w.changesWaiting(); err != nil || n != 0 {
			return fmt.Errorf("%d changes are waiting, want none: %w", n, err)
		}
		return nil
	})
	sc.When(`^I open the window$`, w.open)
	sc.Step(`^the window is closed and opened again$`, w.open)
	sc.Then(`^the window asks to pick up the task left unsaved$`, func() error {
		if w.w.Screen != board.OnAsking || w.w.Question != board.AboutDraft {
			return fmt.Errorf("the window is on %v, asking %v", w.w.Screen, w.w.Question)
		}
		return nil
	})
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
		parsed, err := calendar.ParseDeadline(day)
		w.today = parsed.Add(12 * time.Hour)
		return err
	})
	sc.Then(`^the calendar is open on "([^"]*)"$`, func(day string) error {
		if w.w.Screen != board.OnCalendar {
			return fmt.Errorf("the window is on %v, not on the calendar", w.w.Screen)
		}
		if got := calendar.FormatDay(w.w.Day); got != day {
			return fmt.Errorf("the calendar is on %s, want %s", got, day)
		}
		return nil
	})
	sc.Step(`^I click the day "([^"]*)" on the calendar$`, func(day string) error {
		for i, shown := range board.MonthGrid(w.w.Day) {
			if calendar.FormatDay(shown) == day {
				box := DayAt(w.w.CalendarBox(), i)
				w.w = w.w.clicked(geom.Point{X: box.X + box.W/2, Y: box.Y + box.H/2})
				return nil
			}
		}
		return fmt.Errorf("the calendar does not show %s", day)
	})
	sc.Step(`^I click the calendar's month$`, func() error {
		box := CalendarMonth(w.w.CalendarBox())
		w.w = w.w.clicked(geom.Point{X: box.X + box.W/2, Y: box.Y + box.H/2})
		return nil
	})
	sc.Then(`^the calendar offers the twelve months$`, func() error {
		want := []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}
		if got := w.w.CalendarCells(); !slices.Equal(got, want) {
			return fmt.Errorf("the calendar offers %v, want %v", got, want)
		}
		return nil
	})
	sc.Step(`^I click the month "([^"]*)"$`, func(name string) error {
		i := slices.Index(w.w.CalendarCells(), name)
		if i < 0 {
			return fmt.Errorf("the calendar offers no %q", name)
		}
		box := CalendarCell(w.w.CalendarBox(), i)
		w.w = w.w.clicked(geom.Point{X: box.X + box.W/2, Y: box.Y + box.H/2})
		return nil
	})
	sc.Then(`^the calendar shows ([A-Z][a-z]+)$`, func(month string) error {
		if w.w.Screen != board.OnCalendar || len(w.w.CalendarCells()) != 0 {
			return fmt.Errorf("the calendar is not showing its days")
		}
		if got := w.w.Day.Month().String(); got != month {
			return fmt.Errorf("the calendar shows %s, want %s", got, month)
		}
		return nil
	})
	sc.Step(`^I click the calendar's year$`, func() error {
		box := CalendarYear(w.w.CalendarBox())
		w.w = w.w.clicked(geom.Point{X: box.X + box.W/2, Y: box.Y + box.H/2})
		return nil
	})
	sc.Then(`^the calendar offers the years around the one shown$`, func() error {
		shown := strconv.Itoa(w.w.Day.Year())
		got := w.w.CalendarCells()
		if len(got) != 12 || !slices.Contains(got, shown) {
			return fmt.Errorf("the calendar offers %v, want twelve years around %s", got, shown)
		}
		for i := 1; i < len(got); i++ {
			if got[i-1] >= got[i] {
				return fmt.Errorf("the years %v are not in order", got)
			}
		}
		return nil
	})
	sc.Step(`^I click the year "([^"]*)"$`, func(year string) error {
		i := slices.Index(w.w.CalendarCells(), year)
		if i < 0 {
			return fmt.Errorf("the calendar offers no %q", year)
		}
		box := CalendarCell(w.w.CalendarBox(), i)
		w.w = w.w.clicked(geom.Point{X: box.X + box.W/2, Y: box.Y + box.H/2})
		return nil
	})
	sc.Then(`^the calendar offers the twelve months of (\d+)$`, func(year string) error {
		if got := strconv.Itoa(w.w.Day.Year()); got != year {
			return fmt.Errorf("the calendar is in %s, want %s", got, year)
		}
		if got := w.w.CalendarCells(); len(got) != 12 || got[0] != "Jan" {
			return fmt.Errorf("the calendar offers %v, want the twelve months", got)
		}
		return nil
	})
	sc.Step(`^I pick that file in the dialog$`, func() error { return w.answerDialog(dialog{open: w.file}) })
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
	sc.Then(`^task (\d+) carries the priority "([^"]*)"$`, func(n int, priority string) error {
		shown, err := w.other.Ask(engine.TaskShow, engine.Input{Given: map[string]string{"number": strconv.Itoa(n)}})
		if err != nil {
			return err
		}
		if said := render.Text(engine.TaskShow, shown); !strings.Contains(said, "priority     "+priority) {
			return fmt.Errorf("task %d reads:\n%s", n, said)
		}
		return nil
	})
	sc.Step(`^I pick "([^"]*)" in the menu bar$`, func(item string) error {
		w.w = w.w.Picked(item)
		return w.w.Wrong()
	})
	menuLines := func() []string {
		var lines []string
		for _, group := range MenuBar("windows", w.w.Pages) {
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
	sc.Step(`^I pick "([^"]*)" from the menu bar$`, func(line string) error {
		w.w = w.w.Picked(line)
		return nil
	})
	sc.Given(`^the weekends are hidden$`, func() error {
		w.pages = w.pages.WithoutWeekends()
		return nil
	})
	sc.Given(`^the pages turned off are "([^"]*)"$`, func(list string) error {
		w.pages = engine.PagesOff(strings.Split(list, ", ")...)
		return nil
	})
	sc.Then(`^every column's name stands at the left of its head$`, func() error {
		return w.everyColumnsName(func(head, name geom.Box) error {
			if left := name.X - head.X; left < 6 || left > 16 {
				return fmt.Errorf("the name starts %dpx into its head, want 6 to 16", left)
			}
			return nil
		})
	})
	sc.Then(`^the column heads read (".*")$`, func(list string) error {
		if got := w.w.ColumnHeads(); strings.Join(got, "|") != strings.Join(quoted(list), "|") {
			return fmt.Errorf("the column heads read %v", got)
		}
		return nil
	})
	sc.Then(`^every column's name stands clear of the edge above it$`, func() error {
		return w.everyColumnsName(func(head, name geom.Box) error {
			if clear := name.Y - head.Y; clear < 6 {
				return fmt.Errorf("the name stands %dpx under the edge, want 6px at least", clear)
			}
			return nil
		})
	})
	sc.Then(`^the window says every page is turned off$`, func() error {
		if got := w.w.BlankSays(); got != BlankSays {
			return fmt.Errorf("the window says %q, want %q", got, BlankSays)
		}
		return nil
	})
	sc.When(`^I click the offer to open the settings$`, func() error {
		b := BlankOffer(w.w.width, w.w.height)
		w.w = w.w.clicked(geom.Point{X: b.X + b.W/2, Y: b.Y + b.H/2})
		return nil
	})
	sc.Then(`^the window is on the page "([^"]*)"$`, func(page string) error {
		if got := Pages[w.w.Tab]; got != page {
			return fmt.Errorf("the window is on %q", got)
		}
		return nil
	})
	sc.Step(`^I click the card "([^"]*)"$`, w.click)
	sc.Step(`^I click the tab "([^"]*)"$`, w.clickTab)
	sc.Step(`^I ⌘-click the card "([^"]*)"$`, func(title string) error {
		return w.clickWith(key.ModShortcut, w.find, title)
	})
	sc.Step(`^I ⌘-click the objective "([^"]*)"$`, func(title string) error {
		return w.clickWith(key.ModShortcut, w.findGoal, title)
	})
	sc.Step(`^I click the objective "([^"]*)"$`, func(title string) error {
		if err := w.clickGoal(title); err != nil {
			return err
		}
		for i, g := range board.Sorted(w.w.Showing().Goals) {
			if g.Number == w.w.Chosen {
				w.was = i
			}
		}
		return nil
	})
	sc.Step(`^I drag the objective "([^"]*)" (onto|above|below) "([^"]*)"$`, func(title, how, target string) error {
		from, err := w.findGoal(title)
		if err != nil {
			return err
		}
		box, err := w.goalBox(target)
		if err != nil {
			return err
		}
		to := geom.Point{X: box.X + box.W/2, Y: box.Y + box.H/2}
		switch how {
		case "above":
			to.Y = box.Y + 2
		case "below":
			to.Y = box.Y + box.H - 2
		}
		w.w = w.w.PressAt(from).MoveTo(geom.Point{X: from.X, Y: from.Y + Wobble*3}).MoveTo(to).LetGo(to)
		return nil
	})
	sc.Step(`^I hold the objective "([^"]*)" over "([^"]*)" for (\d+) ms$`, func(title, target string, ms int) error {
		from, err := w.findGoal(title)
		if err != nil {
			return err
		}
		box, err := w.goalBox(target)
		if err != nil {
			return err
		}
		to := geom.Point{X: box.X + box.W/2, Y: box.Y + box.H/2}
		w.w = w.w.PressAt(from).MoveTo(geom.Point{X: from.X, Y: from.Y + Wobble*3}).MoveTo(to)
		w.w = w.w.At(w.w.Now.Add(time.Duration(ms) * time.Millisecond))
		return nil
	})
	sc.Step(`^I hold the objective "([^"]*)" below "([^"]*)", left of its indent$`, func(title, target string) error {
		from, err := w.findGoal(title)
		if err != nil {
			return err
		}
		box, err := w.goalBox(target)
		if err != nil {
			return err
		}
		to := geom.Point{X: box.X - 4, Y: box.Y + box.H - 2}
		w.w = w.w.PressAt(from).MoveTo(geom.Point{X: from.X, Y: from.Y + Wobble*3}).MoveTo(to)
		return nil
	})
	sc.Step(`^I let the objective go below "([^"]*)"$`, func(target string) error {
		box, err := w.goalBox(target)
		if err != nil {
			return err
		}
		to := geom.Point{X: box.X + box.W/2, Y: box.Y + box.H - 2}
		w.w = w.w.MoveTo(to).LetGo(to)
		return nil
	})
	sc.Step(`^the landing says "([^"]*)"$`, func(want string) error {
		if got := w.w.GoalLandingSays(); got != want {
			return fmt.Errorf("the landing says %q, want %q", got, want)
		}
		return nil
	})
	sc.Step(`^I click the task "([^"]*)" in the detail$`, func(title string) error {
		d := w.w.Detail()
		for _, line := range append(append([]TaskLine{}, d.Own...), d.InSubs...) {
			if line.Card.Title == title {
				w.w = w.w.clicked(geom.Point{X: line.At.X + line.At.W/2, Y: line.At.Y + line.At.H/2})
				return nil
			}
		}
		return fmt.Errorf("the detail lists no task titled %q", title)
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
	column := func(name string) (geom.Point, error) {
		for i, heading := range []string{"BACKLOG", "TODO", "CURRENT", "DONE"} {
			if heading == name {
				box := w.w.Layout().Column[i]
				return geom.Point{X: box.X + box.W/2, Y: box.Y + box.H - 10}, nil
			}
		}
		return geom.Point{}, fmt.Errorf("no column is headed %q", name)
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
				w.w = w.w.clicked(geom.Point{X: box.X + box.W/2, Y: box.Y + box.H/2})
				return nil
			}
		}
		return fmt.Errorf("the menu offers %q, and not %q", w.w.Menu(), said)
	})
	sc.Then(`^the menu offers (".*")$`, func(list string) error {
		var want []string
		for _, quoted := range strings.Split(list, ", ") {
			want = append(want, strings.Trim(quoted, `"`))
		}
		if w.w.Screen != board.OnMenu {
			return fmt.Errorf("no menu is open: the window is on %v", w.w.Screen)
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
	sc.Step(`^I double-click between the cards "([^"]*)" and "([^"]*)"$`, func(above, below string) error {
		at, err := w.between(above, below)
		w.w = w.w.clicked(at).DoubleClick(at)
		return err
	})
	sc.Step(`^I click outside the panel$`, func() error {
		off := geom.Point{X: 2, Y: 2}
		w.w = w.w.PressAt(off).LetGo(off)
		return nil
	})
	sc.Step(`^I click the search box$`, func() error {
		box := Searching(w.w.Layout())
		w.w = w.w.clicked(geom.Point{X: box.X + box.W/2, Y: box.Y + box.H/2})
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
	sc.Step(`^I drag the card "([^"]*)" onto "([^"]*)"$`, w.dragOnto)
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
	sc.Step(`^I scroll down over the lane "([^"]*)"$`, w.scrollOverLane)
	sc.Step(`^I pull the bar of the lane "([^"]*)" to its foot$`, w.pullLaneBar)
	sc.Step(`^it was worked on (\d+) times$`, w.workedOn)
	sc.Step(`^I drag the form from a place that holds no text (\d+) to the right and (\d+) down$`, func(right, down int) error {
		box := w.w.formPanel()
		from := geom.Point{X: box.X + 3, Y: box.Y + 3}
		if w.w.textAt(from) {
			return fmt.Errorf("the point %v of the form holds a text", from)
		}
		return w.dragForm(from, right, down)
	})
	sc.Step(`^I click the panel's cross$`, func() error {
		cross, ok := w.w.PanelClose()
		if !ok {
			return fmt.Errorf("the panel on %v has no cross", w.w.Screen)
		}
		w.w = w.w.clicked(geom.Point{X: cross.X + cross.W/2, Y: cross.Y + cross.H/2})
		return w.w.Wrong()
	})
	sc.Step(`^I drag the form from its heading (\d+) to the right and (\d+) down$`, func(right, down int) error {
		head := HeadingAt(w.w.formPanel())
		return w.dragForm(geom.Point{X: head.X + 4, Y: head.Y + head.H/2}, right, down)
	})
	sc.Step(`^the form stands (\d+) to the right and (\d+) lower than it opened$`, func(right, down int) error {
		return w.formStandsBy(right, down)
	})
	sc.Step(`^the form stands where it opened$`, func() error { return w.formStandsBy(0, 0) })
	sc.Step(`^I choose "([^"]*)"$`, w.choose)
	sc.Step(`^I set the field "([^"]*)" to "([^"]*)"$`, w.setField)
	sc.Step(`^I put the keyboard on "([^"]*)"$`, w.keyboardOn)
	sc.Step(`^I drag the step "([^"]*)" onto the step "([^"]*)"$`, func(from, onto string) error {
		at := func(title string) (geom.Point, error) {
			for i, c := range w.w.PartsTicked() {
				if c.Title == title {
					tile := w.w.partTiles()[i]
					return geom.Point{X: tile.X + tile.W/2, Y: tile.Y + tile.H/2}, nil
				}
			}
			return geom.Point{}, fmt.Errorf("no part is titled %q", title)
		}
		a, err := at(from)
		if err != nil {
			return err
		}
		b, err := at(onto)
		if err != nil {
			return err
		}
		w.w = w.w.PressAt(a).MoveTo(geom.Point{X: a.X, Y: a.Y + Wobble + 1}).MoveTo(b).LetGo(b)
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
	sc.Step(`^I click the field "([^"]*)"$`, w.clickField)
	sc.Step(`^I drag across the field "([^"]*)" from line (\d+) character (\d+) to line (\d+) character (\d+)$`,
		func(name string, fromLine, fromColumn, toLine, toColumn int) error {
			fields := w.w.Form.Fields
			for i, field := range fields {
				if field.Name != name {
					continue
				}
				_, value, rows := FormPlace(w.w.formBox(), fields, i)
				inside := Inside(Framed(value, rows))
				at := func(line, column int) geom.Point {
					row := LineAt(inside, line, 0)
					return geom.Point{X: row.X + column*glyphOf(FormSize(field)), Y: row.Y + row.H/2}
				}
				w.w = w.w.PressAt(at(fromLine, fromColumn)).MoveTo(at(toLine, toColumn)).LetGo(at(toLine, toColumn))
				return nil
			}
			return fmt.Errorf("the form has no field called %q", name)
		})
	sc.Step(`^I double-click the list row "([^"]*)"$`, func(title string) error {
		box := w.w.ChoosingBox()
		for i, one := range w.w.Choosing() {
			if one.Title != title {
				continue
			}
			row := ChoosingRow(box, ChoosingFrom(box)+(i-w.w.ChoosingFromRow())*apart)
			w.w = w.w.doubleClicked(geom.Point{X: row.X + row.W/2, Y: row.Y + row.H/2})
			return nil
		}
		return fmt.Errorf("the list offers no row called %q", title)
	})
	sc.Step(`^I click the line that narrows the list$`, func() error {
		look := ChoosingLooking(w.w.ChoosingBox())
		w.w = w.w.clicked(geom.Point{X: look.X + 20, Y: look.Y + look.H/2})
		return nil
	})
	sc.Step(`^somebody else adds a task titled "([^"]*)"$`, w.somebodyElseAdds)
	sc.Step(`^the window looks again$`, w.looksAgain)
	sc.Step(`^somebody else adds the step "([^"]*)" to task (\d+)$`, func(title string, n int) error {
		return w.givens().IAddedTheStepToTask(title, n, false)
	})

	w.expects(sc)
	w.gutters(sc)
	w.filtersSteps(sc)
	w.viewsSteps(sc)
	w.configurationSteps(sc)
	w.workspaceSteps(sc)
	w.componentsSteps(sc)
	w.objectivesSteps(sc)
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

func (w *world) addManyObjectives(n int) error {
	for i := 1; i <= n; i++ {
		if err := w.givens().IAddedAnObjectiveTitled(fmt.Sprintf("Objective %d", i)); err != nil {
			return err
		}
	}
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

// headClear is how far inside a head its name is looked for: past the rounding of its ground's corners, which is not ink.
const headClear = 5

func (w *world) everyColumnsName(holds func(head, name geom.Box) error) error {
	shot, err := drawnInAStep(w.theme, w.w)
	if err != nil {
		return err
	}
	for i, head := range w.w.Layout().Head {
		inside := geom.Box{X: head.X + headClear, Y: head.Y + headClear, W: head.W - 2*headClear, H: head.H - headClear}
		name, ok := inkedIn(shot, inside, design.Dark().Colour.Surface.Ground)
		if !ok {
			return fmt.Errorf("column %d: its head carries no name", i)
		}
		if err := holds(head, name); err != nil {
			return fmt.Errorf("column %d: %w", i, err)
		}
	}
	return nil
}

func inkedIn(shot *image.RGBA, b geom.Box, ground color.NRGBA) (geom.Box, bool) {
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
		return geom.Box{}, false
	}
	return geom.Box{X: left, Y: top, W: right - left + 1, H: bottom - top + 1}, true
}

func (w *world) dragForm(from geom.Point, right, down int) error {
	w.opened = w.w.formPanel()
	to := geom.Point{X: from.X + right, Y: from.Y + down}
	w.w = w.w.PressAt(from)
	w.w = w.w.MoveTo(geom.Point{X: from.X + right/2, Y: from.Y + down/2})
	w.w = w.w.MoveTo(to)
	w.w = w.w.LetGo(to)
	return nil
}

func (w *world) formStandsBy(right, down int) error {
	if w.w.Screen != board.OnForm {
		return fmt.Errorf("the form is no longer open: the window is on %v", w.w.Screen)
	}
	now := w.w.formPanel()
	if got := (geom.Point{X: now.X - w.opened.X, Y: now.Y - w.opened.Y}); got != (geom.Point{X: right, Y: down}) {
		return fmt.Errorf("the form moved by %v, want %d right and %d down", got, right, down)
	}
	return nil
}
