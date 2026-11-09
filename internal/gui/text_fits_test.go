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
	return []struct {
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
	}
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
