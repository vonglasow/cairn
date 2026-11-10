package gui

import (
	"fmt"
	"image"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget/material"

	"cairn/internal/engine"
)

type saidAt struct {
	box  Box
	ink  Box
	said string
	high int
}

func onTheFixtures(t *testing.T) engine.Work {
	t.Helper()
	store := filepath.Join(t.TempDir(), "cairn.db")
	from := filepath.Join(t.TempDir(), "fixtures.json")
	if err := os.WriteFile(from, engine.FixturesFile(), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := engine.Do(engine.TaskImport, store, engine.Input{Given: map[string]string{"from": from}}); err != nil {
		t.Fatal(err)
	}
	for _, kept := range []struct {
		op engine.Operation
		in engine.Input
	}{
		{engine.ViewNew, engine.Input{Given: map[string]string{"name": "Home automation"}, Listed: map[string][]string{"objective": {"8"}}}},
		{engine.ViewNew, engine.Input{Given: map[string]string{"name": "Backend, due soon", "priority": "high"}, Listed: map[string][]string{"label": {"backend"}}}},
		{engine.ViewNew, engine.Input{Given: map[string]string{"name": "Gone"}, Listed: map[string][]string{"label": {"long-gone"}}}},
		{engine.LabelDelete, engine.Input{Given: map[string]string{"label": "long-gone"}, Confirmed: true}},
	} {
		if _, err := engine.Do(kept.op, store, kept.in); err != nil {
			t.Fatal(err)
		}
	}
	held, err := engine.OpenAt(engine.Places{Store: store, Notes: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = held.Close() })
	return held
}

func aFixturesBoard(t *testing.T) Window {
	t.Helper()
	return New(onTheFixtures(t), 1200, 800).At(time.Now()).Look()
}

func textDrawn(w Window) []saidAt {
	var noted []saidAt
	was := noteSaid
	noteSaid = func(gtx layout.Context, th *material.Theme, b Box, said string, size int, face font.Font, lines int, along text.Alignment) {
		boxed := gtx
		boxed.Constraints = layout.Constraints{Max: image.Pt(max(b.W, 1), 1<<20)}
		recorded := op.Record(boxed.Ops)
		label := material.Label(th, unit.Sp(float32(size)), said)
		label.Font.Weight = face.Weight
		label.MaxLines = lines
		laid := label.Layout(boxed).Size
		recorded.Stop()
		ink := Box{X: b.X, Y: b.Y, W: min(laid.X, b.W), H: min(laid.Y, b.H)}
		switch along {
		case text.End:
			ink.X = b.X + b.W - ink.W
		case text.Middle:
			ink.X = b.X + (b.W-ink.W)/2
		}
		noted = append(noted, saidAt{box: b, ink: ink, said: said, high: laid.Y})
	}
	defer func() { noteSaid = was }()
	var ops op.Ops
	gtx := layout.Context{Ops: &ops, Constraints: layout.Exact(image.Pt(w.width, w.height)), Metric: unit.Metric{PxPerDp: 1, PxPerSp: 1}}
	Draw(NewKit(gtx, Theme(), Dark(), &Drawn{}), w)
	return noted
}

var pinnedSizes = []image.Point{{Narrowest, Shortest}, {1200, 800}}

func (w Window) choosingFirstList() Window {
	for at, f := range w.form.Fields() {
		if chosenFromAList(f) {
			w.form = w.form.onto(at)
			return w.choosingList()
		}
	}
	return w
}

func everyScreen(w Window) []struct {
	name string
	w    Window
} {
	card := w.clicked(middleOf(w, 1, 0))
	parent, sub, refused := objectiveScreens(w)
	viewed, changed, keeping := viewScreens(w)
	return append([]struct {
		name string
		w    Window
	}{
		{"the board", w},
		{"a chosen card", card},
		{"a task's form", card.Edit()},
		{"the help", func() Window { h := w; h.screen = OnHelp; return h }()},
		{"the objectives", w.onPage(Objectives)},
		{"the time", w.onPage(Time)},
		{"the labels", w.listingLabels()},
		{"the backups", w.backingUp().listingBackups()},
		{"the filters", w.filteringBoard()},
		{"the configuration", w.openingConfiguration()},
		{"the command bar", w.openingPalette()},
		{"a card's menu", card.Key("m")},
		{"a card's labels", card.Key("l")},
		{"a card's calendar", card.Key("t")},
		{"a list chosen from", card.Edit().choosingFirstList()},
		{"a question", card.Key("shift+x")},
		{"the switcher", w.openingSwitcher()},
		{"a chosen objective", w.onPage(Objectives).Key("right")},
		{"an objective's form", w.onPage(Objectives).Key("right").Edit()},
		{"a task's trace", card.Edit().togglingTrace()},
		{"the week", w.onPage(Time)},
		{"the week before", w.onPage(Time).Key("left")},
		{"a day", w.onPage(Time).ranged("Day")},
		{"the month", w.onPage(Time).ranged("Month")},
		{"the months", func() Window { m := w.onPage(Time).ranged("Month"); m.grid = viewMonths; return m }()},
		{"the years", func() Window { m := w.onPage(Time).ranged("Month"); m.grid = viewYears; return m }()},
		{"an objective holding sub-objectives", parent},
		{"its tasks shown", parent.Key("t")},
		{"its tasks narrowed", parent.Key("t").Key("]")},
		{"its deletion asked", parent.Delete()},
		{"a sub-objective's form", sub.Edit()},
		{"the parents offered", onField(sub.Edit(), "parent")},
		{"what a task serves, refused", refused},
		{"a view in force", viewed},
		{"a view changed on top", changed},
		{"the view's popover", changed.openingViewPopover()},
		{"a view kept", keeping},
		{"the views", w.listingViews()},
		{"a view gone", w.listingViews().Key("down").Key("down")},
		{"a view renamed", w.listingViews().Choose("Rename")},
		{"a sub-objective added under it", parent.Key("shift+n")},
		{"a narrowed part walked to", parent.Key("t").Key("]").Key("]")},
		{"the views walked to", w.filteringBoard().Key("up")},
		{"a view narrowed to a part", viewed.Key("]")},
	}, seriesScreens(w)...)
}

func seriesScreens(w Window) []struct {
	name string
	w    Window
} {
	r := w.Paged(engine.Pages{}).Look()
	tab := r.onPage(Series)
	stopped := tab
	for _, s := range tab.SeriesListed() {
		if s.State == "stopped" {
			stopped.seriesOn = s.Number
		}
	}
	occurrence := r
	for _, c := range r.snap.Cards {
		if c.Series != nil && !c.Archived && occurrence.chosen == 0 {
			occurrence.chosen = c.Number
		}
	}
	asked := occurrence.Edit()
	asked.form = asked.form.Set("title", "Daily stand-up, again")
	asked = asked.Save()
	off := w.openingPalette()
	off.palette.typed, off.palette.opened = Written{Value: "series", At: 6}, off.now.Add(-time.Second)
	return []struct {
		name string
		w    Window
	}{
		{"the series", tab},
		{"a stopped series", stopped},
		{"the series walked down", tab.Key("down")},
		{"a series' form", tab.Key("e")},
		{"a series' deletion asked", tab.Key("backspace")},
		{"every task of a series' deletion asked", tab.Key("backspace").Key("down").Key("down").Key("enter")},
		{"a new series", tab.Key("n")},
		{"the rules offered", tab.Key("n").choosingList()},
		{"the board marking its series", r},
		{"a chosen occurrence", occurrence},
		{"the help, with the series", func() Window { h := r; h.screen = OnHelp; return h }()},
		{"an occurrence's form", occurrence.Edit()},
		{"saving an occurrence asked", asked},
		{"the command bar, recurrence off", off},
	}
}

func viewScreens(w Window) (viewed, changed, keeping Window) {
	viewed = w.Key("cmd+1")
	changed = w.Key("cmd+2").filteringBoard()
	changed.filters = changed.filters.On("state", "todo")
	changed.screen = OnTheBoard
	keeping = w.filteringBoard()
	keeping.filters = keeping.filters.On("priority", "high").On("state", "current")
	return viewed, changed, keeping.Key("v")
}

func objectiveScreens(w Window) (parent, sub, refused Window) {
	for _, g := range w.snap.Goals {
		if len(g.Children) > 0 && parent.width == 0 {
			parent = w.onPage(Objectives)
			parent.chosen, parent.chosenGoal = g.Number, true
			sub = parent
			sub.chosen = g.Children[0]
		}
	}
	for _, c := range w.snap.Cards {
		if g, ok := w.snap.Goal(sub.chosen); ok && slices.Contains(c.Objectives, g.Number) && refused.width == 0 {
			refused = w
			refused.chosen = c.Number
			refused = onField(refused.Edit(), "objective")
			for i, one := range refused.Choosing() {
				if one.Number == g.Parent {
					refused.choosing = i
				}
			}
			refused = refused.taking()
		}
	}
	return parent, sub, refused
}

func TestTheFixturesDrawEveryTextWhereItFits(t *testing.T) {
	work := onTheFixtures(t)
	for _, size := range pinnedSizes {
		var wrong []string
		for _, s := range everyScreen(New(work, size.X, size.Y).At(time.Now()).Look()) {
			drawn := textDrawn(s.w)
			for _, one := range drawn {
				if one.box.W <= 0 || one.box.H <= 0 {
					wrong = append(wrong, fmt.Sprintf("%s: %q drawn in an empty box %+v", s.name, one.said, one.box))
					continue
				}
				if one.high > one.box.H+textSlack {
					wrong = append(wrong, fmt.Sprintf("%s: %q is %d high in a box of %d", s.name, one.said, one.high, one.box.H))
				}
			}
		}
		if len(wrong) > 0 {
			t.Errorf("at %dx%d, %d problem(s):\n%s", size.X, size.Y, len(wrong), strings.Join(wrong, "\n"))
		}
	}
}

func TestBarWidths_giveWordsUpThenTheFirstOffers(t *testing.T) {
	full, keys := []int{100, 100, 100}, []int{30, 30, 30}
	for _, c := range []struct {
		room   int
		want   []int
		worded bool
	}{
		{300, []int{100, 100, 100}, true},
		{299, []int{30, 30, 30}, false},
		{90, []int{30, 30, 30}, false},
		{89, []int{0, 30, 30}, false},
		{30, []int{0, 0, 30}, false},
		{0, []int{0, 0, 0}, false},
	} {
		got, worded := barWidths(full, keys, c.room)
		if fmt.Sprint(got) != fmt.Sprint(c.want) || worded != c.worded {
			t.Errorf("room %d: %v %v, want %v %v", c.room, got, worded, c.want, c.worded)
		}
	}
}

func TestAViewsNameHasTheRoomLeftOfItsCount(t *testing.T) {
	w := aFixturesBoard(t).filteringBoard()
	rows := w.ViewRows()
	if len(rows) == 0 {
		t.Fatal("the filter menu lists no view")
	}
	for _, one := range textDrawn(w) {
		if one.said == rows[0].Name && one.box.W < rows[0].At.W/2 {
			t.Errorf("the view %q is written in %+v, of a row %+v", one.said, one.box, rows[0].At)
		}
	}
}
