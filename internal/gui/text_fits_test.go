package gui

import (
	"cairn/internal/board"
	"cairn/internal/gui/design"
	"cairn/internal/gui/geom"
	"image"
	"os"
	"path/filepath"
	"slices"
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
	box  geom.Box
	ink  geom.Box
	said string
	high int
}

func onTheFixtures(t *testing.T) board.Work {
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

func textDrawn(w Window) []saidAt { return textDrawnWith(drawingTheme(), w) }

func noting(th *material.Theme, note sayNote) func() {
	notesSaid.Store(th, note)
	return func() { notesSaid.Delete(th) }
}

func textDrawnWith(th *material.Theme, w Window) []saidAt {
	var noted []saidAt
	stop := noting(th, func(gtx layout.Context, th *material.Theme, b geom.Box, said string, size int, face font.Font, lines int, along text.Alignment) {
		boxed := gtx
		boxed.Constraints = layout.Constraints{Max: image.Pt(max(b.W, 1), 1<<20)}
		recorded := op.Record(boxed.Ops)
		label := material.Label(th, unit.Sp(float32(size)), said)
		label.Font.Weight = face.Weight
		label.MaxLines = lines
		laid := label.Layout(boxed).Size
		recorded.Stop()
		ink := geom.Box{X: b.X, Y: b.Y, W: min(laid.X, b.W), H: min(laid.Y, b.H)}
		switch along {
		case text.End:
			ink.X = b.X + b.W - ink.W
		case text.Middle:
			ink.X = b.X + (b.W-ink.W)/2
		}
		noted = append(noted, saidAt{box: b, ink: ink, said: said, high: laid.Y})
	})
	defer stop()
	var ops op.Ops
	gtx := layout.Context{Ops: &ops, Constraints: layout.Exact(image.Pt(w.width, w.height)), Metric: unit.Metric{PxPerDp: 1, PxPerSp: 1}}
	Draw(NewKit(gtx, th, design.Dark(), &Drawn{}), w)
	return noted
}

var pinnedSizes = []image.Point{{geom.Narrowest, geom.Shortest}, {1200, 800}}

func (w Window) choosingFirstList() Window {
	for at, f := range w.Form.Fields {
		if chosenFromAList(f) {
			w.Form = w.Form.Onto(at)
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
		{"the help", func() Window { h := w; h.Screen = board.OnHelp; return h }()},
		{"the objectives", w.onPage(board.Objectives)},
		{"the time", w.onPage(board.Time)},
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
		{"the views under the tab", viewed.openingViews()},
		{"the views, a view changed", changed.openingViews()},
		{"a chosen objective", w.onPage(board.Objectives).Key("right")},
		{"an objective's form", w.onPage(board.Objectives).Key("right").Edit()},
		{"a task's trace", card.Edit().togglingTrace()},
		{"the week", w.onPage(board.Time)},
		{"the week before", w.onPage(board.Time).Key("left")},
		{"a day", w.onPage(board.Time).ranged("Day")},
		{"the month", w.onPage(board.Time).ranged("Month")},
		{"the months", func() Window { m := w.onPage(board.Time).ranged("Month"); m.Grid = board.ViewMonths; return m }()},
		{"the years", func() Window { m := w.onPage(board.Time).ranged("Month"); m.Grid = board.ViewYears; return m }()},
		{"an objective holding sub-objectives", parent},
		{"its tasks shown", parent.Key("t")},
		{"its tasks narrowed", parent.Key("t").Key("]")},
		{"its deletion asked", parent.Delete()},
		{"a sub-objective's form", sub.Edit()},
		{"the parents offered", onField(sub.Edit(), "parent")},
		{"what a task serves, refused", refused},
		{"a view in force", viewed},
		{"a view changed on top", changed},
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
	tab := r.onPage(board.Series)
	stopped := tab
	for _, s := range tab.SeriesListed() {
		if s.State == "stopped" {
			stopped.SeriesOn = s.Number
		}
	}
	occurrence := r
	for _, c := range r.Snap.Cards {
		if c.Series != nil && !c.Archived && occurrence.Chosen == 0 {
			occurrence.Chosen = c.Number
		}
	}
	asked := occurrence.Edit()
	asked.Form = asked.Form.Set("title", "Daily stand-up, again")
	asked = asked.Save()
	off := w.openingPalette()
	off.palette.typed, off.palette.opened = board.Written{Value: "series", At: 6}, off.Now.Add(-time.Second)
	return []struct {
		name string
		w    Window
	}{
		{"the series", tab},
		{"a stopped series", stopped},
		{"the series narrowed to a state", stopped.walkingSeriesState().walkingSeriesState()},
		{"the series walked down", tab.Key("down")},
		{"a series' form", tab.Key("e")},
		{"a series' deletion asked", tab.Key("backspace")},
		{"every task of a series' deletion asked", tab.Key("backspace").Key("down").Key("down").Key("enter")},
		{"a new series", tab.Key("n")},
		{"the rules offered", tab.Key("n").openingRule()},
		{"a rule built by its parts", tab.Key("n").openingRule().ruled(ruleUnit, 1).ruled(ruleOn, 3).ruled(ruleEnds, 2).Key("down")},
		{"a rule read back", func() Window {
			n := tab.Key("n")
			n.Form = n.Form.Writes(board.Ending("every other thursday 14:00"))
			return n
		}()},
		{"the board marking its series", r},
		{"a chosen occurrence", occurrence},
		{"the help, with the series", func() Window { h := r; h.Screen = board.OnHelp; return h }()},
		{"an occurrence's form", occurrence.Edit()},
		{"saving an occurrence asked", asked},
		{"the command bar, recurrence off", off},
	}
}

func viewScreens(w Window) (viewed, changed, keeping Window) {
	viewed = w.Key("cmd+1")
	changed = w.Key("cmd+2").filteringBoard()
	changed.Filters = changed.Filters.On("state", "todo")
	changed.Screen = board.OnTheBoard
	keeping = w.filteringBoard()
	keeping.Filters = keeping.Filters.On("priority", "high").On("state", "current")
	return viewed, changed, keeping.Key("v")
}

func objectiveScreens(w Window) (parent, sub, refused Window) {
	for _, g := range w.Snap.Goals {
		if len(g.Children) > 0 && parent.width == 0 {
			parent = w.onPage(board.Objectives)
			parent.Chosen, parent.ChosenGoal = g.Number, true
			sub = parent
			sub.Chosen = g.Children[0]
		}
	}
	for _, c := range w.Snap.Cards {
		if g, ok := w.Snap.Goal(sub.Chosen); ok && slices.Contains(c.Objectives, g.Number) && refused.width == 0 {
			refused = w
			refused.Chosen = c.Number
			refused = onField(refused.Edit(), "objective")
			for i, one := range refused.Choosing() {
				if one.Number == g.Parent {
					refused.ChoosingOn = i
				}
			}
			refused = refused.taking()
		}
	}
	return parent, sub, refused
}
